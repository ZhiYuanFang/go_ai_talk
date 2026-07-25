# confirm-ws-adaptation 设计文档

## 1. 修复后的 WS confirm 流程图

```
用户语音输入
    │
    ▼
chatWithResult()
    │
    ├─[1] 检查 pendingConfirmState.get(deviceNo)
    │       │
    │       ├─ 有 pending（未超时）
    │       │   │
    │       │   ├─ parseConfirmFeedback(transcript)
    │       │   │   │
    │       │   │   ├─ "confirm" ──► 调用 ConfirmIntent(user_feedback="confirm")
    │       │   │   ├─ "reject"  ──► 调用 ConfirmIntent(user_feedback="reject")
    │       │   │   └─ ""        ──► 视为常规输入，清理 pending，走常规流程
    │       │   │
    │       │   ├─ clear(deviceNo)
    │       │   │
    │       │   ├─ 调用成功：映射到 deepSeekUnifiedIntent → handleUnifiedIntentAction
    │       │   └─ 调用失败：返回错误提示话术
    │       │
    │       └─ 无 pending（或已超时）
    │           │
    │           ▼
    ├─[2] 走原有 chatWithResult 流程
    │       │
    │       ├─ 模式切换 / 模式查询
    │       ├─ 闲聊模式（casual）
    │       ├─ 待选子事件（pendingChild）
    │       ├─ 命中预设动作
    │       └─ callDeepSeekUnifiedIntent()
    │               │
    │               ▼
    └─[3] 检查 intent.NeedConfirm
            │
            ├─ true  ──► pendingConfirmState.set(deviceNo, {ConversationID, EventName, Action, CreatedAt})
            │           返回 intent.ConfirmMessage 给用户
            │
            └─ false ──► 走原有 handleUnifiedIntentAction 流程
```

## 2. 字段映射表

### 2.1 Python `/v1/analyze/intent` 响应 → Go `AnalyzeIntentResponse`

| Python 字段 | Go 字段 | 类型 | 说明 |
|-------------|---------|------|------|
| `target_type` | `TargetType` | string | 目标类型 |
| `action` | `Action` | string | 动作类型 |
| `event_name` | `EventName` | string | 事件名称 |
| `keywords` | `Keywords` | []string | 关键词列表 |
| `content` | `Content` | string | 回答内容 |
| `need_confirm` | `NeedConfirm` | bool | **新增** 是否需要确认 |
| `confirm_message` | `ConfirmMessage` | string | **新增** 确认话术 |
| `conversation_id` | `ConversationID` | string | **新增** 会话 ID |

### 2.2 `AnalyzeIntentResponse` → `deepSeekUnifiedIntent`

| AnalyzeIntentResponse | deepSeekUnifiedIntent | 处理 |
|------------------------|----------------------|------|
| `TargetType` | `TargetType` | `strings.TrimSpace(strings.ToLower(...))` |
| `Action` | `Action` | `strings.TrimSpace(...)` |
| `EventName` | `EventName` | `strings.TrimSpace(...)` |
| `Content` | `Reply` | `sanitizeModelReplyText(...)` |
| `NeedConfirm` | `NeedConfirm` | 直接传递 |
| `ConfirmMessage` | `ConfirmMessage` | 直接传递 |
| `ConversationID` | `ConversationID` | 直接传递 |
| - | `NeedUserReply` | 默认 true |

## 3. pending 状态管理设计

### 3.1 数据结构

```go
// pendingConfirmEntry 待确认意图条目
type pendingConfirmEntry struct {
    ConversationID string    // 会话 ID（用于恢复 Python 图执行）
    EventName      string    // 事件名称
    Action         string    // 动作类型
    CreatedAt      time.Time // 创建时间（用于超时清理）
}

// pendingConfirmStateStruct 待确认意图状态管理
type pendingConfirmStateStruct struct {
    mu      sync.RWMutex
    entries map[string]*pendingConfirmEntry // key: deviceNo
}
```

### 3.2 操作语义

| 方法 | 行为 |
|------|------|
| `set(deviceNo, entry)` | 加锁写入，覆盖旧条目 |
| `get(deviceNo)` | 读锁读取；若超时（>60s）返回 nil |
| `clear(deviceNo)` | 加锁删除条目 |

### 3.3 全局实例

```go
var pendingConfirmState = &pendingConfirmStateStruct{
    entries: make(map[string]*pendingConfirmEntry),
}
```

与 `pendingChild` / `pendingQuantity` 的差异：
- `pendingChild` / `pendingQuantity` 挂在 `VoiceService` 上（`s.pendingChild`）
- `pendingConfirmState` 采用全局变量（包级单例），原因：
  - confirm 流程与设备模式无关，纯上下文缓存
  - 简化调用路径，避免在 chatWithResult 中传递 s 实例
  - 与 `pythonAIClientFromCfg()` 全局工厂模式一致

## 4. confirm/reject 关键词列表

### 4.1 肯定词（→ "confirm"）

```go
confirmWords := []string{
    "确认", "是的", "对", "好的", "没错",
    "嗯", "ok", "yes", "是", "对的", "可以",
}
```

### 4.2 否定词（→ "reject"）

```go
rejectWords := []string{
    "取消", "不是", "错", "不对", "没有",
    "no", "nope", "不", "错的", "不对的",
}
```

### 4.3 匹配规则

1. 文本先 `strings.ToLower(strings.TrimSpace(text))` 归一化
2. 顺序遍历肯定词列表，命中任一返回 `"confirm"`
3. 再遍历否定词列表，命中任一返回 `"reject"`
4. 都不命中返回 `""`（视为常规输入）

### 4.4 注意事项

- 否定词 `"不"` 在 `"不是"` 之前会因 `"不是"` 也包含 `"不"` 而被先匹配；
  但肯定词列表先于否定词执行，因此 `"是的"` 等不会被误判为否定
- 用户说 `"不要"` 时，肯定词列表均不命中，进入否定词列表，`"不"` 命中返回 `"reject"`
- 无法识别的输入（如 `"那个"`）返回空字符串，由调用方决定是否回退到常规流程

## 5. chatWithResult 中的分支位置

```
chatWithResult()
├─ [新增] 检查 pendingConfirmState.get(deviceNo)        ← 最优先
├─ 模式查询命令（isModeQueryCommand）
├─ 模式切换命令（isModeSwitchCommand）
├─ 闲聊模式（ChatModeCasual）
├─ 待选子事件（pendingChild）                            ← 与 pendingConfirm 平级
├─ 命中预设动作
├─ callDeepSeekUnifiedIntent()
│   └─ [新增] 检查 intent.NeedConfirm                   ← 调用后立即检查
└─ handleUnifiedIntentAction()
```

设计要点：
- pendingConfirm 检查放在最前面：一旦有 pending，无论用户输入什么都先尝试解析为 confirm/reject
- 若解析为空（用户说无关内容），清理 pending 后落入常规流程，保证用户能继续操作
- pendingConfirm 与 pendingChild 互斥：理论上不会同时存在（confirm 仅在意图分析后产生，而 pendingChild 在事件解析时产生）
