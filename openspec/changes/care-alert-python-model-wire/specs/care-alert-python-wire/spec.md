## ADDED Requirements

### Requirement: care-alert 分析请求以 model 字段传首选模型

voice-service 调用 Python `POST /v1/care-alert/analyze` 时，MUST 将 `ResolveLaneModel` 得到的非空模型配置序列化为 JSON 字段 **`model`**（对象形状：`provider`、`name`、`max_in_flight`，与 clinic/tip/intent 的 `PythonModelCfg` 一致）。MUST NOT 依赖 JSON 字段 `model_cfg` 作为 Python 可读契约。当选模结果为 omit（指针 nil）时，请求体 MUST NOT 出现 `model` 字段，以便 Python 走免费保底序。

#### Scenario: premium 或 free 有配置时传 model 对象

- **WHEN** 日缓存未命中且 `ResolveLaneModel` 返回非 nil 的 `*PythonModelCfg`
- **THEN** 发往 Python 的 JSON MUST 包含 `model` 对象，且其 `provider`/`name` 与该配置一致
- **AND** 请求体 MUST NOT 以 `model_cfg` 作为唯一模型载体（不得仅发 `model_cfg` 而省略 `model`）

#### Scenario: omit 时不传 model

- **WHEN** 日缓存未命中且选模结果为 nil（非 premium 且 lane free 为空）
- **THEN** 发往 Python 的 JSON MUST 省略 `model` 字段
- **AND** 服务 MUST 仍可调用分析接口（由 Python 保底序处理）

#### Scenario: 与兄弟仓字段名对齐

- **WHEN** 对照 `python_ai_talk` 的 `CareAlertAnalyzeRequest`
- **THEN** Go 序列化字段名 MUST 为 `model`（可选），MUST NOT 要求 Python 解析 `model_cfg`
