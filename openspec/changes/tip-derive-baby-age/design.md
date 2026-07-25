## Context

跨仓 tip 生成链路现状：

```
Flutter (babyAgeMonths + currentTime camel)
  → Go DeviceTipGenerateReq / TipStream
    → Python TipRequest (baby_age_months + current_time)
      → TipState → tip_answer 提示词
```

问题：

1. Go `api/v1/device_tip_http.go` 注释写「`babyAgeMonths` 为 0 则服务端按生日推算」，但 `VoiceService.TipStream` **并未**按生日推算，只是把客户端值原样转给 Python；`currentTime<=0` 时 Go 用 `time.Now().Unix()` 补，与「应由 Python 写 prompt 时生成」的目标重复且易不一致。
2. Python 已在图中调用 `fetch_baby_profile`（history `birthday`），具备自算月龄的数据源。
3. 前置 change：`python_ai_talk/openspec/changes/fix-python-request-aliases` 保证 Go snake body 可过校验；本变更在其之上删除 tip 冗余字段。

约束（explore 锁定，必须遵守）：

- Go↔Python 内部 API 保持 snake_case；**不**改 `PythonAIClient` 为 camel。
- 月龄由 Python 自算；**不需要** Go/Flutter 传 `baby_age_months` / `babyAgeMonths`。
- `current_time` 由 Python 写提示词时生成；即时 tip 路径删除/废弃请求字段；**不保留为必填**。
- **不动 Chat 宿主**（仍走 history SSE）。
- App 对外 Flutter↔Go 仍可 camel；本变更焦点是 tip 契约瘦身 + Python 派生上下文。

## Goals / Non-Goals

**Goals:**

- Tip 请求契约仅保留生成所需：设备号、事件信息、模型配置（及现有必要字段）；去掉月龄与当前时间入参。
- Python 在画像可用时按 birthday 计算月龄；无生日时提示词明确「未知」，**不得**伪装成 0 个月。
- 提示词时间上下文由 Python 在 **Asia/Shanghai** 时区生成（见决策）。
- 修正 Go/文档中「0 则服务端推算」假注释与误导性默认值逻辑。

**Non-Goals:**

- 不改 Chat / history SSE 聊天宿主。
- 不改 clinic 请求体字段（仅时区约定可与 tip 对齐说明）。
- 不把月龄计算下沉到 Flutter 或 Go。
- 不新增自动化测试文件（仓库约定）。
- 不在本变更重做知识库算法（仅规定无月龄时的检索兜底语义）。

## Decisions

### 决策 0（依赖）：先完成 fix-python-request-aliases

- **方案**：apply 顺序：先 Python alias 修复，再本变更。
- **理由**：否则 Go snake tip body 在字段删除前仍可能 422，联调噪声大。

### 决策 1（锁定）：月龄由 Python 在 fetch_baby_profile 之后自算

- **方案**：在 tip 图中，`fetch_baby_profile` 成功拿到含 `birthday` 的 profile 后，用「当前日期（Asia/Shanghai）− birthday」计算整月月龄，写入 `TipState.baby_age_months`（或等价内部字段）供提示词与（若有）知识检索使用。
- **计算规则（拍板）**：按日历月差：`(now.year - birth.year) * 12 + (now.month - birth.month)`，若 `now.day < birth.day` 则减 1；结果 `< 0` 时钳为 0（生日在未来视为 0 个月，与「未知」区分）。
- **备选**：继续由 Flutter/Go 传入 — **排除**（锁定）。

### 决策 2（锁定）：无生日兜底 =「未知」，不是 0 个月

| 情况 | `baby_age_months` 内部表示 | 提示词文案 |
|------|---------------------------|------------|
| 有合法 birthday | 非负整数月龄 | `宝宝月龄：{n} 个月` |
| profile 缺失 / 无 birthday / 无法解析 | **`None` / 缺省**（不得填 0） | `宝宝月龄：未知` |
| birthday 在未来 | `0` | `宝宝月龄：0 个月`（明确是计算出的 0，不是未知） |

- **知识检索**：若现有逻辑用月龄过滤，无月龄（未知）时 SHALL 跳过月龄过滤或走「不限月龄」宽检索，MUST NOT 用 0 冒充新生儿月龄去过滤。
- **备选**：未知时传 0 — **排除**（会误导 LLM/检索）。

### 决策 3（锁定）：current now 由 Python 写提示词时生成；请求字段删除

- **方案**：从 `TipRequest`、Go `TipStreamRequest`、Go App `DeviceTipGenerateReq`、Flutter body **删除** `current_time` / `currentTime`（及 Go `TipStream` 参数）。在构建 tip 提示词时生成时间上下文。
- **格式（拍板）**：提示词同时给出：
  1. Asia/Shanghai 本地可读时间（如 `%Y-%m-%d %H:%M:%S`）；
  2. 对应 Unix 秒（`int(time.time())`）可选一行，便于模型理解绝对时间。
- **时区（拍板，写死）**：**`Asia/Shanghai`**。与 clinic 提示词若未显式时区，本变更 tip **统一上海时区**；clinic 对齐改造不在范围，但新增 tip 代码不得使用「服务器随意本地时区」而不写死。
- **备选**：保留可选 `current_time` 覆盖 — **排除**（不保留为必填，亦不保留覆盖通道，避免时钟漂移争议）。

### 决策 4：Go / Flutter 契约瘦身与 v1 结构约定

- **方案**：在同一 `POST /device/tip/generate` 上删除 `babyAgeMonths`、`currentTime` 字段；同步改 `TipStream` 签名与 `TipStreamRequest`；Flutter 停止计算/传递。
- **理由**：该接口仍处于接线 change（`wire-tip-generate-on-voice` / `tip-chat-streaming`）推进期，尚未作为长期冻结的生产契约；假注释证明字段语义本未落地。
- **与 AGENTS「v1 结构不可改」关系**：本接口视为接线期可修正的未冻结契约；若 apply 前已确认有生产客户端强依赖这两字段，则改为新增 `api/v2` tip generate（path 另定）且 v1 忽略字段——**默认按删除现字段实施**。
- **必做**：删除/改写「0 或负数表示由服务端根据生日推算」「0 表示使用服务端当前时间」等 dc/注释；Go 侧删除 `currentTime<=0` 补 `time.Now()` 再转 Python 的逻辑（时间改由 Python 生成）。

### 决策 5（锁定）：不动 Chat 宿主

- tip 仍走 voice tip SSE；Chat 继续 history SSE；本变更不交叉改 chat 流。

### 决策 6：Go↔Python 命名

- 内部 JSON 保持 snake_case（依赖 change 1）；App Flutter↔Go 剩余字段仍可 camel（`deviceNo`、`eventId`、`eventName`）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| birthday 字段名/格式与 history 返回不一致 | 实现时对照 history birthday API 实际 JSON；解析失败走「未知」 |
| 旧 Flutter 仍传月龄/时间 | 服务端忽略多余字段（Pydantic 默认忽略）或短暂双读后删除；Flutter 同步发版 |
| 未知月龄导致知识检索变差 | 宽检索 + 提示词标明未知，避免错误 0 月过滤 |
| 时区与服务器 UTC 容器不一致 | 写死 `ZoneInfo("Asia/Shanghai")`，不依赖主机 localtime |

## Migration Plan

1. 确认 `fix-python-request-aliases` 已部署或同批次先合并。
2. 先发 Python（可先兼容忽略入参、内部自算），再发 Go 去字段，再发 Flutter。
3. 回滚：恢复请求字段与旧传参；Python 可再读可选字段作 fallback（仅回滚期，非常态）。

## Open Questions

- 无阻塞。若负责人裁定 tip generate 已冻结必须 v+1，将 DeviceTipGenerate 改为 v2 path，tasks 增补一条即可。
