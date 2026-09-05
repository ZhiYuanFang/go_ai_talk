## ADDED Requirements

### Requirement: 系统 MUST 提供按主体与通道入参驱动的功能开通原子入口

cash-service MUST 提供统一开通入口（名称以实现为准），调用方 MUST 传入至少：`featureId`、激活主体类型（`device` 或 `user`）、主体键、开通通道（`payment`、`invite_code` 或 `ad`）、通道引用与操作者 `wxId`（邀请场景）。支付回调、邀请码兑换成功路径、广告完成开通 MUST 经该入口写入权益或数量，MUST NOT 旁路直接 upsert 表。一期对主体类型 `user` MUST 拒绝或返回明确错误（本变更不落地用户维权益表）。预测与值得留意等设备共享功能 MUST 使用主体 `device`。

#### Scenario: 支付汇入原子入口

- **WHEN** 功能订单支付履约成功
- **THEN** 系统 MUST 调用原子入口且 Channel 为 `payment`，按 SKU 授予效果写入对应设备权益或条数

#### Scenario: 邀请汇入原子入口

- **WHEN** 邀请码兑换某功能成功校验通过后授予
- **THEN** 系统 MUST 调用原子入口且 Channel 为 `invite_code`，MUST NOT 在兑换函数内绕过入口直接写 `feature_entitlement` / `feature_allowed_count`

#### Scenario: 广告汇入原子入口

- **WHEN** 广告完成开通申报通过校验
- **THEN** 系统 MUST 调用原子入口且 Channel 为 `ad`

#### Scenario: user 主体一期拒绝

- **WHEN** 调用方以 SubjectType=`user` 请求开通
- **THEN** 系统 MUST 拒绝，MUST NOT 写入设备权益行冒充成功

### Requirement: 邀请码与广告的授予效果 MUST 同源解析

对同一 `featureId`，通道 `invite_code` 与 `ad` MUST 使用同一套效果解析规则。对权益型功能，授予天数 MUST 取自该功能定义上的邀请/广告授予天数配置（`feature_def.duration_days` 或等价字段）：`0` 表示永久，`>0` 表示自授予时刻起对应自然日秒数。对预测数量类功能，邀请与广告 MUST 继续按永久条数增量语义授予，MUST NOT 因该天数配置改变条数语义。支付通道 MUST 继续以商品 SKU 的 `duration_days` / `grant_kind` 为准，MUST NOT 强制与邀请天数相同。

#### Scenario: 权益型邀请与广告同天数

- **WHEN** 某权益型功能配置邀请/广告授予天数为 7，用户分别用邀请码与广告开通该功能
- **THEN** 两次授予的到期语义 MUST 同为约 7 天（续期规则与既有 entitlement upsert 一致）

#### Scenario: 预测邀请广告仍为条数

- **WHEN** 用户对 `prediction_unlock` 完成邀请或广告开通
- **THEN** 系统 MUST 永久 `allowed_count` +1（或等价增量），MUST NOT 仅因定义天数改为限时全开

### Requirement: Admin MUST 可配置功能的邀请/广告授予天数

开通功能管理中的功能定义编辑 MUST 提供「邀请/广告授予天数」表单字段并持久化到功能定义。`0` MUST 表示永久。该字段 MUST NOT 被解释为支付套餐时长（支付仍以 SKU 为准）。预测类功能可展示该字段但运行时条数授予 MUST 忽略之。

#### Scenario: 运维修改邀请天数后兑码

- **WHEN** 管理员将某权益型功能的邀请/广告天数从 7 改为 3 并保存，之后用户首次邀请开通该功能
- **THEN** 新授予 MUST 按 3 天计算到期

#### Scenario: 支付套餐独立

- **WHEN** 功能定义邀请天数为 7 且某支付 SKU `duration_days=0`
- **THEN** 支付开通 MUST 为永久，MUST NOT 变成 7 天
