## Context

当前 Go 侧通过 `PythonAIClient`（[python_ai_client.go](file:///d:/work/go_ai_talk/internal/services/voice/python_ai_client.go)）调用 Python 微服务的 5 类接口：
- 意图分析（非流式）：`POST /v1/analyze/intent` ✅ 对齐
- 意图确认：`POST /v1/analyze/intent/confirm` ✅ 对齐
- 诊疗流式：`POST /v1/clinic/stream` ✅ 对齐
- 小贴士流式：`POST /v1/tip/stream` ⚠️ 字段类型不匹配
- 反馈（诊疗/小贴士）：`POST /v1/clinic|tip/feedback` ❌ 参数传递方式不匹配

Python 侧已实现流式意图分析接口 `/v1/analyze/intent/stream`（SSE 格式），但 Go 侧缺少客户端封装和服务层入口。

TTS 语音场景下百度 TTS 需要全量文案才能生成音频，因此必须使用非流式接口；MCP 服务和纯文字场景可以利用流式接口获得更好的体验。

## Goals / Non-Goals

**Goals:**
1. 修复 feedback 接口参数传递方式（Python 侧从 Query 改为 JSON Body）
2. 修复 TipStreamRequest 字段类型对齐（event_id int64, current_time int64 Unix秒）
3. Go 侧新增流式意图分析客户端封装（AnalyzeIntentStream）
4. Go 侧 voice service 新增流式意图分析入口，与 TTS 非流式路径分离

**Non-Goals:**
- 不修改现有 TTS 语音场景的非流式意图分析流程
- 不修改 Python 侧流式意图分析接口的 SSE 格式
- 不修改 intent 接口已有字段语义
- 不引入新的外部依赖

## Decisions

### 决策1：Python feedback 接口改为接收 JSON Body
- **方案**：Python 侧 `clinic_feedback` 和 `tip_feedback` 函数参数从裸 Query 参数改为接收 Pydantic Body 模型
- **理由**：Go 侧已经在发 JSON Body，修改 Python 侧成本更低；Body 方式比 Query 更适合未来扩展（如添加更多反馈字段）
- **备选**：Go 侧改发 Query 参数 — 排除，因为 Go 侧已有较多调用走 Body 模式，保持一致性更好

### 决策2：TipStreamRequest EventID 改为 int64，CurrentTime 改为 int64（Unix 秒）
- **方案**：Go 侧 `TipStreamRequest.EventID` 从 `string` → `int64`，`CurrentTime` 从 `string` → `int64`
- **理由**：Python 侧 [tip.py](file:///d:/work/python_ai_talk/app/tip/schemas/tip.py) 明确定义为 `int`；`current_time` 在 Python 侧作为 Unix 秒填入 LLM 提示词（[tip_answer.py:93](file:///d:/work/python_ai_talk/app/tip/graphs/nodes/prompts/tip_answer.py#L93)）
- **备选**：Python 侧改类型 — 排除，因为 Python 侧已经是正确的语义（event_id 是数据库主键 int，current_time 是 Unix 秒）

### 决策3：流式意图分析走独立服务层入口，不修改现有 TTS 路径
- **方案**：Go 侧 voice service 新增 `HandleTranscriptForIntentStream`（或类似命名）方法，内部调用 `PythonAIClient.AnalyzeIntentStream`；现有 `HandleTranscriptForStreaming`（TTS 场景）保持不变
- **理由**：TTS 场景（百度 TTS）需要全量文案才能合成音频，流式对其无价值；MCP 和纯文字场景需要流式获得更好体验
- **备选**：修改现有方法内部根据配置切换流式/非流式 — 排除，因为调用场景差异大，分离更清晰

### 决策4：流式意图分析 SSE 事件解析复用现有 ClinicStream 模式
- **方案**：`AnalyzeIntentStream` 使用与 `ClinicStream` 相同的 SSE 逐行解析 + 回调模式（`OnThinking` / `OnAnswer` / `OnDone`）
- **理由**：代码模式一致，降低维护成本；Python 侧 SSE 格式也与诊疗一致（type + content）

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| Python feedback 改 Body 可能影响其他调用方 | 在 Python 侧做兼容（同时支持 Query 和 Body），或确认无其他调用方 |
| TipStreamRequest 类型变更可能遗漏调用点 | 编译时类型检查兜底；grep 所有 `TipStreamRequest` 构造点 |
| 流式意图分析与 TTS 路径分离导致代码重复 | 共用底层 `chatWithResult` 的核心逻辑，仅调用 Python 方法不同 |
