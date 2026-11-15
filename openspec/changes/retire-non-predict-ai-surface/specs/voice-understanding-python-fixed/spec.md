## ADDED Requirements

### Requirement: 喂养语音理解 MUST 直调 Python 且不传选模

voice-service 在喂养语音理解（统一意图 / AnalyzeIntent 及 stream）路径上 MUST 调用 Python `POST /v1/analyze/intent` 或 `/v1/analyze/intent/stream`，请求体 MUST NOT 包含 `model`（或等价 provider/name 选模对象）。Go MUST NOT 再经 `llm_lane_config` / `ResolveLaneModel` / VIP / AI 额度决定理解所用模型。

#### Scenario: intent 请求无 model 字段

- **WHEN** 已登录用户触发喂养理解并调用 Python intent
- **THEN** 发往 Python 的 JSON MUST 省略 `model`，且 Python 侧 MUST 仍能完成意图分析（定死 LLM）

### Requirement: 理解路径 MUST NOT 依赖 AI 额度门禁

理解路径 MUST NOT 因 `voice_ai` 月度额度用尽而返回 40302，MUST NOT 在调用前对理解做 premium/free omit 分轨。登录/设备鉴权失败等既有非额度错误语义 MAY 保留。

#### Scenario: 无额度表时理解仍可用

- **WHEN** voice AI 额度表与检查逻辑已删除且用户合法进入聊天
- **THEN** 理解调用 MUST 仍可到达 Python intent，MUST NOT 因缺失额度行而失败

### Requirement: 理解编排与 ASR/TTS 契约 MUST 保留

系统 MUST 保留聊天 WS/MCP 入口、理解编排（澄清/确认/reply 透传）及理解所依赖的 ASR/TTS/历史上下文路径。删除选模 MUST NOT 删除 `PythonAIClient` 的 intent 调用能力。

#### Scenario: 聊天主路径仍返回 reply

- **WHEN** 用户通过既有喂养聊天通道发送可理解文本
- **THEN** 系统 MUST 仍能返回基于 Python intent 的 reply（或既有规则/确认流），MUST NOT 因删除选模而整段聊天不可用
