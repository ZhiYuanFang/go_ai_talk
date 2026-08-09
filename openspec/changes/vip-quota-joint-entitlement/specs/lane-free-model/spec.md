## ADDED Requirements

### Requirement: 业务 lane 可配置额度不足模型

对 lane `voiceUnderstanding`、`clinic`、`careAlert`、`polish`，Admin 配置 MUST 支持可选的免费/额度不足模型字段（`freeProvider`、`freeModel`，可为空）。非 premium 选模 MUST 读取上述字段：非空则作为下游模型；为空则 Go MUST omit 模型配置并由 Python（或 polish 侧约定的无模型覆盖语义）自行决定免费模型。Sim 四条 lane（`simText`/`simVision`/`simImageGen`/`simVideoGen`）MUST NOT 要求或持久化 free 字段。

#### Scenario: 保存 clinic free 模型

- **WHEN** 管理员 PUT llm-lanes 且 `clinic.freeProvider`/`freeModel` 为允许列表内有效值
- **THEN** 系统 MUST 持久化该 free 配置，供后续非 premium 诊所/tip 路径使用

#### Scenario: 清空 free 表示交 Python

- **WHEN** 管理员将某业务 lane 的 freeProvider 与 freeModel 存为空
- **THEN** 非 premium 且走 Python 的请求 MUST 省略 model 字段

#### Scenario: Sim 无 free

- **WHEN** 管理员 GET/PUT sim llm-lanes
- **THEN** API MUST NOT 要求客户端提交 free 字段；即便误传也 MUST NOT 改变 Sim 运行时选模契约（本变更不引入 Sim 额度策略）

### Requirement: 正式模型与 free 模型分离

lane 的正式 `provider`/`model` MUST 仅用于 premium 路径；MUST NOT 在非 premium 路径回退使用正式模型（除非 free 显式配置为相同值）。硬编码 `DegradedVoiceUnderstandingProfile` / `DegradedClinicProfile` / `DegradedPolishProfile` MUST NOT 再作为生产选模真相源。

#### Scenario: 额尽不再使用代码种子智谱

- **WHEN** 非 VIP 用户某 feature 额度用尽且该 lane free 配置指向运维设定的模型 A
- **THEN** 下游 MUST 使用模型 A，MUST NOT 忽略 free 而改用代码内 Degraded* 常量
