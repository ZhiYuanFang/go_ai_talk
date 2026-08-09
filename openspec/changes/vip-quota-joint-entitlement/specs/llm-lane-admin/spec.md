## ADDED Requirements

### Requirement: voice llm-lanes 包含 careAlert 与 free 字段

`GET/PUT /voice/admin/api/llm-lanes` MUST 包含三个业务子对象：`voiceUnderstanding`、`clinic`、`careAlert`。每一子对象 MUST 含正式 `provider`/`model`/`maxInFlight`/`maxWaiters` 以及可选 `freeProvider`/`freeModel`（允许空字符串）。PUT MUST 校验正式模型 allowlist；当 free 非空时 MUST 同样校验 allowlist；free 全空 MUST 合法。

#### Scenario: PUT 写入 careAlert 与 free

- **WHEN** 管理员提交含 `careAlert` 正式模与空 free 的合法 PUT
- **THEN** 系统 MUST 持久化成功，后续 GET MUST 读出一致字段

#### Scenario: 非法 free 模型拒绝

- **WHEN** freeProvider/freeModel 非空但不在 allowlist
- **THEN** PUT MUST 失败且 MUST NOT 部分写入该 lane

### Requirement: ucg polish lane 支持 free 字段

ucg 侧 polish lane Admin 读写 MUST 支持与 voice 业务 lane 同语义的 `freeProvider`/`freeModel`（可空）。

#### Scenario: 保存 polish free

- **WHEN** 管理员保存 polish free 为允许列表内模型
- **THEN** 后续非 premium 润笔 MUST 能读到该配置
