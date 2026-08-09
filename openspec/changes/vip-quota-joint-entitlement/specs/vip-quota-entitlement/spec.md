## ADDED Requirements

### Requirement: 统一 premium 权益判定

系统 MUST 通过公共方法（或等价单一出口）计算业务 LLM 是否走 premium 路径：`isPremium = isVip(wxId) OR hardwarePrivilege OR quota.Allowed(feature)`。VIP 判定 MUST 经 cash-service 内部契约按 `wxId` 读取；查询失败 MUST 视为非 VIP 并记录 Warning，MUST NOT 因此单独使无关主流程硬失败（调用方仍可走非 premium）。硬件特权 MUST 仅由显式入口标记赋予，MUST NOT 仅因 `wxId=0` 推断为硬件特权。

#### Scenario: VIP 账号为 premium

- **WHEN** cash 返回该 `wxId` 为 VIP 且 privilege 为 Account
- **THEN** `isPremium` MUST 为 true，无论该 feature 月度 used 是否已达 limit

#### Scenario: 非 VIP 但额度允许

- **WHEN** 非 VIP、非硬件特权，且对应 feature 的额度快照 `Allowed=true`
- **THEN** `isPremium` MUST 为 true

#### Scenario: 非 VIP 且额度用尽

- **WHEN** 非 VIP、非硬件特权，且额度 `Allowed=false`
- **THEN** `isPremium` MUST 为 false

#### Scenario: 硬件特权为 premium

- **WHEN** 请求携带硬件特权标记（如 `/voice/chat/ws` 或 MCP/internal text chat 入口所置）
- **THEN** `isPremium` MUST 为 true，MUST NOT 要求 `wxId>0` 或 cash VIP

#### Scenario: 未登录无硬件特权

- **WHEN** privilege 为 Account 且 `wxId<=0`（无硬件标记）
- **THEN** `isPremium` MUST 为 false（放行主流程时 MUST 走非 premium 选模），MUST NOT 仅因未登录返回 40301 而阻止已约定可匿名/设备通道以外的、本策略允许继续的 LLM 路径（各业务原有强制登录要求仍按其自身 Requirement）

### Requirement: VIP 与硬件不计次

当本次请求因 VIP 或硬件特权而 premium 时，系统 MUST NOT 对该 feature 执行成功计次（consume MUST 为 no-op）。额度读接口对 VIP 用户 MUST 将该 feature 视为有额度（`Allowed=true`，且 MUST NOT 因 used>=limit 单独返回 40302 阻断）。

#### Scenario: VIP 成功调用不扣次

- **WHEN** VIP 用户完成一次需计次的 LLM 成功路径
- **THEN** 系统 MUST NOT 增加该 feature 的月用量

#### Scenario: 硬件特权成功不扣次

- **WHEN** 带硬件特权的语音对话成功完成 LLM
- **THEN** 系统 MUST NOT 因该次成功增加 `voice_ai`（或其它）月用量

#### Scenario: VIP 读额度 API

- **WHEN** VIP 用户请求 App `ai-quota`（voice 或 ucg 域）
- **THEN** 响应中对应 feature MUST `allowed=true`（即使 used 已达历史 limit）

### Requirement: 原子选模出口

所有向 Python 传递模型配置的业务路径，以及 Go 直调 aimodel 且受本策略约束的路径（含 polish），MUST 经公共选模出口决策：premium 时使用该 lane 正式 `provider/model`；非 premium 时使用该 lane 的 free 配置；free 为空时 MUST omit model（Python 路径不出现 model/model_cfg 字段；Go 直调路径 MUST NOT 使用已废除的硬编码 Degraded* 作为真相源）。调用方 MUST NOT 在出口之外自行根据 VIP 或额度拼装另一套模型。

#### Scenario: premium 传正式模

- **WHEN** `isPremium=true` 且 lane 已配置正式模型
- **THEN** 下游请求 MUST 携带该正式 provider/model（或等价 ModelCfg）

#### Scenario: 非 premium 且 free 已配置

- **WHEN** `isPremium=false` 且 lane freeProvider/freeModel 均非空（或约定的有效 free 配置）
- **THEN** 下游 MUST 使用 free 配置模型

#### Scenario: 非 premium 且 free 为空

- **WHEN** `isPremium=false` 且 free 配置为空
- **THEN** Go→Python 请求 MUST 省略 model（及 care-alert 的 model_cfg 中的模型字段），由 Python 自行选择免费模型
