## Why

Go 侧 `PythonAIClient` 与 Python 侧接口存在 3 处不匹配，导致调用失败或参数丢失：
1. 反馈接口（clinic/tip feedback）Go 侧发 JSON Body，但 Python 侧按 Query 参数解析，参数完全丢失。
2. TipStreamRequest 的 `event_id` 类型 Go 为 `string`，Python 为 `int`，Pydantic 校验失败。
3. TipStreamRequest 的 `current_time` 类型 Go 为 `string`（ISO格式），Python 为 `int`（Unix秒），用途是作为时间上下文填入 LLM 提示词。

同时，Python 侧已有流式意图分析接口 `/v1/analyze/intent/stream`，Go 侧缺少客户端封装和服务层调用入口，MCP 服务和纯文字场景无法利用流式能力。

## What Changes

- **Python 侧 clinic/tip feedback 接口**：从 Query 参数改为接收 JSON Body（Pydantic 模型）
- **Go 侧 TipStreamRequest**：`EventID` 从 `string` 改为 `int64`，`CurrentTime` 从 `string` 改为 `int64`（Unix 秒）
- **Go 侧 PythonAIClient**：新增 `AnalyzeIntentStream` 方法，支持 SSE 流式意图分析
- **Go 侧 voice service**：新增流式意图分析的服务层入口（V2 方法），供 MCP 服务和纯文字场景调用；现有 TTS 语音场景的非流式入口保持不变
- **BREAKING**：Python 侧 feedback 接口从 Query 参数改为 JSON Body，若有其他调用方需同步更新

## Capabilities

### New Capabilities

- `streaming-intent-client`：Go 侧 PythonAIClient 流式意图分析封装（SSE 解析 + 回调）
- `streaming-intent-service`：Go 侧 voice service 流式意图分析入口（独立于 TTS 非流式路径）

### Modified Capabilities

- `python-feedback-api`：Python 侧 clinic/tip feedback 接口从 Query 改为 JSON Body
- `tip-request-types`：Go 侧 TipStreamRequest 字段类型对齐（event_id int64, current_time int64）

## Impact

- **Python 侧**：`app/api/routes/clinic.py` feedback 接口、`app/api/routes/tip.py` feedback 接口
- **Go 侧**：`internal/services/voice/python_ai_client.go`（TipStreamRequest 类型修正 + 新增流式意图方法）
- **Go 侧**：`internal/services/voice/voice_chat.go` 或相关 service 文件（新增流式意图入口）
- **Go 侧**：所有构造 TipStreamRequest 的调用点（event_id 传 int64，current_time 传 Unix 秒）
- **配置**：无新增依赖
