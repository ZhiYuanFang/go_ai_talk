## ADDED Requirements

### Requirement: 功能支付履约 MUST 授予 device_no 权益或递增 allowedCount

cash-service MUST 支持独立功能 SKU（`feature_product`）：建单与支付宝/Apple IAP 验签 MUST 与 VIP 支付栈一致（服务端验签）。支付成功后 MUST 按商品配置向订单关联的 `device_no` 授予功能权益或递增 `allowedCount`。功能订单 MUST 使用独立 `feature_order` 表。支付宝异步通知与 Apple 验单入口 MUST 与 VIP **共用**，按订单表分流履约，MUST NOT 污染 VIP 订单语义。

#### Scenario: 功能 SKU 支付成功写权益

- **WHEN** 某功能商品支付回调/验单成功且 `grant_kind=entitlement`
- **THEN** 系统 MUST 幂等为该 `device_no` 写入/续期对应 `feature_id` 权益，`unlockMethod` MUST 为 `payment`

#### Scenario: 功能 SKU 支付成功增加 allowedCount

- **WHEN** 某功能商品支付成功且 `grant_kind` 为预测数量增量
- **THEN** 系统 MUST 幂等将该 `device_no` 的 `allowedCount` 增加配置的 `grant_quantity`

#### Scenario: 共用回调按订单分流

- **WHEN** 支付宝通知到达且 `out_trade_no` 属于 `feature_order`
- **THEN** 系统 MUST 执行功能履约且 MUST NOT 续期 `vip_entitlement`

#### Scenario: 重复回调不重复授予

- **WHEN** 同一渠道交易号或已 paid 订单再次通知
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误时长或数量

### Requirement: VIP 购买 MUST NOT 写入功能权益

VIP 商品履约 MUST 仅续期 `vip_entitlement`（`wx_id`）。MUST NOT 为任何 `device_no` 写入 `feature_entitlement`，MUST NOT 修改 `feature_allowed_count`。既有 VIP status API MUST 保持可用，供客户端 V1 覆盖功能列表（不得覆盖 UCG 入场）。

#### Scenario: VIP 履约不影响功能表

- **WHEN** 用户成功购买或续期月 VIP
- **THEN** 系统 MUST 更新 VIP 到期时间，且 MUST NOT 新增或更新任何功能权益行 / allowedCount

### Requirement: 邀请码兑换 MUST 按单功能开通并遵守一家绑定与人维去重

cash-service MUST 提供 `POST /cash/app/api/feature/invite-codes/redeem`（计入 usage）。请求 MUST 含邀请码与目标 `featureId`。主体：`device_no` 与 `wx_id` 均只信网关注入头；`wx_id` MUST > 0。

规则 MUST 包括：

- 一码可覆盖其配置的（或所有支持邀请码的）多个功能，但单次请求 MUST 只开通所请求的一个 `featureId`，MUST NOT 自动开通其余功能。
- 同一 `wx_id` 对同一 `featureId` 仅能通过邀请码开通一次。
- 同一 `wx_id` 仅能使用一家 `owner_wx_id` 的邀请码；**仅当成功开通某一功能后**才写入该绑定；失败兑换 MUST NOT 绑定。
- `redeemer_wx_id` MUST NOT 等于码的 `owner_wx_id`（不可自用）。
- 成功后 MUST 写设备权益或 `allowedCount`，`unlockMethod` MUST 为 `invite_code`，并写入兑换流水（含 owner/redeemer/device/feature/time）。

#### Scenario: 指定功能兑换成功

- **WHEN** 合法用户提交有效码与尚未码开的 `featureId` 且未违反一家/自用规则
- **THEN** 系统 MUST 仅开通该功能（或增加对应数量），并占用相应兑换记录

#### Scenario: 同码再次开通另一功能

- **WHEN** 用户已用某码成功开通功能 A，再次提交同一码与功能 B（B 尚未码开且码支持 B）
- **THEN** 系统 MUST 允许开通 B，且 MUST NOT 因已开 A 而自动开其它功能

#### Scenario: 换一家邀请码被拒绝

- **WHEN** 用户已成功使用 owner=X 的码开通过任一功能，再提交 owner=Y 的码
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 失败兑换不绑定一家

- **WHEN** 用户首次兑码因码无效/过期/功能不支持等原因失败
- **THEN** 系统 MUST NOT 写入 redeemer→owner 绑定

#### Scenario: 不可自用

- **WHEN** 兑换人 `wx_id` 等于码的 `owner_wx_id`
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 同一功能不可再次码开

- **WHEN** 某 `wx_id` 已通过邀请码开通过某 `featureId`
- **THEN** 再次用任意邀请码开通同一 `featureId` MUST 被拒绝

### Requirement: 广告完成开通 MVP MUST 信任客户端申报

cash-service MUST 提供 `POST /cash/app/api/feature/ad/complete`（计入 usage）。MVP MUST 接受客户端完成申报并为 `device_no` 授予对应功能或数量，`unlockMethod` MUST 为 `ad`。MUST 提供基础防刷（短窗幂等或设备日限额）。MUST NOT 强制接入第三方广告服务端验真。

#### Scenario: 客户端申报广告完成获得开通

- **WHEN** 合法请求申报某 `featureId` 广告完成且未触发防刷拒绝
- **THEN** 系统 MUST 授予对应权益或数量增量

#### Scenario: 重复申报幂等

- **WHEN** 同一幂等键在短窗内重复提交广告完成
- **THEN** 系统 MUST NOT 重复叠加授予
