## ADDED Requirements

### Requirement: VIP 商业与判定路径 MUST 被移除

系统 MUST 删除 VIP 支付履约分流、VIP Admin/App/internal 判定入口，以及 voice/cash 中以 VIP 布尔驱动的选模或能力 OR 逻辑。新请求 MUST NOT 依赖 VIP 表或 `RemoteIsVip*` 完成业务。预测 Pro（`predict_capability_pro`）支付/试用/邀请壳 MUST 保持可用。

#### Scenario: 支付回调不再走 VIP 履约

- **WHEN** 渠道回调到达且订单非预测 Pro（或其它仍保留的非 VIP 功能单）
- **THEN** 系统 MUST NOT 执行 VIP 授予路径；VIP 专用代码路径 MUST 已删除或不可达

#### Scenario: Pro 壳仍可开通

- **WHEN** 用户对 `predict_capability_pro` 完成支付或试用或邀请开通
- **THEN** cash-service MUST 仍按既有 Pro 壳契约履约，MUST NOT 因删除 VIP 而失败

### Requirement: 值得留意 care-alert MUST 全栈下线

系统 MUST 删除 care-alert 的 device App API、cash feature/eligibility/access、voice 生成服务、对 Python care-alert 的调用、相关 Hub 入口与额度/lane 依赖。客户端调用已删 path MUST 得到未注册或网关未反代结果，MUST NOT 再生成日缓存分析。

#### Scenario: 日分析接口不可用

- **WHEN** 客户端请求原 `/device/api/care-alert/daily`（或 stream/item/feedback）路径
- **THEN** 服务端 MUST NOT 再绑定并执行 care-alert 业务逻辑

### Requirement: 7 日成长轨迹 growth-trajectory MUST 全栈下线

系统 MUST 删除 growth-trajectory 的 device App API、cash access/feature 种子与试用入口、voice 服务与 Python turn 调用。MUST NOT 再写入成长轨迹最新态业务表（若表残留，代码 MUST NOT 读写）。

#### Scenario: latest/turn 不可用

- **WHEN** 客户端请求原 `/device/api/growth-trajectory/latest` 或 `/turn`
- **THEN** 服务端 MUST NOT 再绑定并执行成长轨迹业务逻辑

### Requirement: 删除不得误伤预测 Pro 壳与喂养 eligibility 引擎

删除 VIP/care/growth 时，系统 MUST 保留 `predict_capability_pro`、`predict_pro_entry` 喂养前置、广告清理与 catalog `status` 行为。通用 feeding-eligibility 引擎若仍服务 `ucg_entry` / `predict_pro_entry`，MUST NOT 被整删。

#### Scenario: Pro eligibility 仍可用

- **WHEN** App 调用预测 Pro 喂养 eligibility API
- **THEN** 响应 MUST 仍按既有契约返回，MUST NOT 因本变更删除 care/growth 场景而 5xx
