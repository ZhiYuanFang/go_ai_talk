## 1. Python 侧 feedback 接口改为 JSON Body

- [x] 1.1 修改 `app/api/routes/clinic.py` 的 `clinic_feedback` 函数，参数从 Query 改为接收 Pydantic Body 模型（answer_id + feedback）
- [x] 1.2 修改 `app/api/routes/tip.py` 的 `tip_feedback` 函数，参数从 Query 改为接收 Pydantic Body 模型（answer_id + feedback）
- [x] 1.3 验证 Python 侧语法检查通过（无 import/syntax 错误）

## 2. Go 侧 TipStreamRequest 字段类型对齐

- [x] 2.1 修改 `internal/services/voice/python_ai_client.go` 中 `TipStreamRequest.EventID` 从 `string` 改为 `int64`
- [x] 2.2 修改 `internal/services/voice/python_ai_client.go` 中 `TipStreamRequest.CurrentTime` 从 `string` 改为 `int64`
- [x] 2.3 搜索所有构造 `TipStreamRequest` 的调用点，修正 `EventID` 和 `CurrentTime` 的传参类型
- [x] 2.4 验证 Go 侧编译通过（`go build` 或 `go vet`）

## 3. Go 侧流式意图分析客户端封装

- [x] 3.1 在 `internal/services/voice/python_ai_client.go` 中新增 `AnalyzeIntentStreamRequest` 结构体（与非流式请求字段一致）
- [x] 3.2 新增 `AnalyzeIntentStreamCallback` 结构体（OnThinking / OnAnswer / OnDone 回调）
- [x] 3.3 新增 `AnalyzeIntentStreamResponse` 结构体（完整思考内容 + 完整意图结果 JSON + 解析后的 AnalyzeIntentResponse）
- [x] 3.4 实现 `AnalyzeIntentStream` 方法：POST `/v1/analyze/intent/stream`，SSE 逐行解析，触发回调，累积结果
- [x] 3.5 验证 Go 侧编译通过

## 4. Go 侧流式意图分析服务层入口

- [x] 4.1 在 `internal/services/contracts/runtime_contracts.go` 的 `VoiceContract` 接口中新增流式意图方法签名
- [x] 4.2 在 `internal/services/voice/` 中实现流式意图分析服务层方法（独立于 TTS 非流式路径）
- [x] 4.3 在 voice service 的 adapter（local/remote/canary）中补齐流式方法的适配
- [x] 4.4 验证 Go 侧编译通过和 `go vet` 无告警
