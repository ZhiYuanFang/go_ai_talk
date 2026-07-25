# confirm-ws-adaptation 任务清单

## 任务组 1：结构体字段适配

### 1.1 修改 `internal/services/voice/python_ai_client.go`
- [x] 在 `AnalyzeIntentResponse` 结构体中新增 3 个字段：`NeedConfirm`、`ConfirmMessage`、`ConversationID`
- [x] 新增 `ConfirmIntentRequest` 结构体（`ConversationID` + `UserFeedback`）
- [x] 新增 `ConfirmIntentResponse` 类型别名（= `AnalyzeIntentResponse`）

### 1.2 修改 `internal/services/voice/voice_chat_understanding.go`
- [x] 在 `deepSeekUnifiedIntent` 结构体中新增 3 个字段：`NeedConfirm`、`ConfirmMessage`、`ConversationID`

## 任务组 2：ConfirmIntent 方法

### 2.1 在 `internal/services/voice/python_ai_client.go` 中新增 `ConfirmIntent` 方法
- [x] 方法签名：`func (c *PythonAIClient) ConfirmIntent(ctx context.Context, req *ConfirmIntentRequest) (*ConfirmIntentResponse, error)`
- [x] 调用 `POST {baseURL}/v1/analyze/intent/confirm`
- [x] 处理状态码非 200 的情况
- [x] 解析 JSON 响应并返回
- [x] 添加中文业务注释

### 2.2 在 `callDeepSeekUnifiedIntent` 中映射新字段
- [x] 在 Python 响应映射到 `deepSeekUnifiedIntent` 时，传递 `NeedConfirm`、`ConfirmMessage`、`ConversationID`

## 任务组 3：pending 状态管理

### 3.1 新建 `internal/services/voice/voice_confirm_pending.go`
- [x] 定义 `pendingConfirmEntry` 结构体（`ConversationID`、`EventName`、`Action`、`CreatedAt`）
- [x] 定义 `pendingConfirmStateStruct` 结构体（`sync.RWMutex` + `map[string]*pendingConfirmEntry`）
- [x] 定义全局变量 `pendingConfirmState`
- [x] 实现 `set` 方法（加锁写入）
- [x] 实现 `get` 方法（读锁读取 + 60 秒超时检查）
- [x] 实现 `clear` 方法（加锁删除）
- [x] 实现 `parseConfirmFeedback` 函数（肯定词/否定词匹配）
- [x] 所有方法与函数添加中文业务注释

## 任务组 4：chatWithResult 分支

### 4.1 在 `chatWithResult` 开头检查 pending 状态
- [x] 在函数开头（normalizeAndValidateChatText 之后）检查 `pendingConfirmState.get(deviceNo)`
- [x] 若有 pending，调用 `parseConfirmFeedback(transcript)` 解析用户反馈
- [x] 解析为非空字符串：调用 `ConfirmIntent`，清理 pending，映射到 `deepSeekUnifiedIntent`，复用 `handleUnifiedIntentAction`
- [x] 解析为空字符串：清理 pending，落入常规流程

### 4.2 在 `callDeepSeekUnifiedIntent` 返回后检查 `NeedConfirm`
- [x] 检查 `intent.NeedConfirm` 是否为 true
- [x] 若 true：调用 `pendingConfirmState.set` 保存上下文，返回 `intent.ConfirmMessage` 作为回复

## 任务组 5：验证

### 5.1 编译验证
- [x] 运行 `go build ./...` 确认无编译错误

## 任务组 6：OpenSpec 文档

### 6.1 文档完善
- [x] 填充 `proposal.md` 内容（概述、背景、目标、范围、设计决策、非目标）
- [x] 创建 `design.md`（WS confirm 流程图、字段映射表、pending 状态管理设计、关键词列表）
- [x] 创建 `tasks.md`（本文件）
