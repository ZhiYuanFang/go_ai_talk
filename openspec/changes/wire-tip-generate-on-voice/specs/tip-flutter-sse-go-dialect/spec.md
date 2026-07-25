## ADDED Requirements

### Requirement: tip 请求路径为 /device/tip/generate

Flutter tip 客户端 SHALL 向网关基址下的 `/device/tip/generate` 发起 POST SSE 请求（不得使用缺少 `/device` 前缀的 `/tip/generate`）。

#### Scenario: URL 含 device tip 前缀

- **WHEN** `TipRepository.streamTip` 构建请求 URL
- **THEN** 路径 MUST 以 `/device/tip/generate` 结尾（相对 `apiBaseUrl`）

### Requirement: tip SSE 解析对齐 Go 方言

Flutter tip SSE 解析 SHALL 读取 `event:` 确定事件类型，`data:` 为纯文本内容；MUST 支持 thinking/answer/done/error，并在流结束识别 `data: [DONE]`。

#### Scenario: thinking/answer 增量

- **WHEN** 收到 `event: thinking`（或 `answer`）及随后 `data: <text>`
- **THEN** 客户端 MUST 将 data 纯文本作为该类型增量内容累积

#### Scenario: done 解析 answerId

- **WHEN** 收到 `event: done` 且 data 为含 `answerId` 的 JSON
- **THEN** 客户端 MUST 解析并保存 `answerId` 字符串，供后续反馈字段对齐使用

### Requirement: 状态模型使用 answerId

tip models / provider SHALL 使用 `String? answerId` 标识本次生成结果，MUST NOT 依赖从 SSE 流写入错误的 `tipId int`。

#### Scenario: 流结束后持有 answerId

- **WHEN** SSE done 事件成功解析 `answerId`
- **THEN** `TipContent`（或等价状态）MUST 更新为该 `answerId`，反馈按钮可用性可依据其非空判断

#### Scenario: 反馈 submit 字段对齐（可不完整飞轮）

- **WHEN** 调用方提交 tip 反馈
- **THEN** 请求体 SHOULD 使用 `answerId` 字符串字段对齐；完整 UI/后端反馈闭环属包 C，本包不要求反馈接口一定可用
