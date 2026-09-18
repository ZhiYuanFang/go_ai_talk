## ADDED Requirements

### Requirement: vip_order 与 feature_order SHALL 具备 grant_reason 列

`ai_voice_cash` 表 `vip_order` 与 `feature_order` MUST 各具备 `grant_reason` 列（字符串，建议 `VARCHAR(256) NOT NULL DEFAULT ''`）。支付渠道创建的订单 MUST 将该列置为空串。Admin 手工授予写入的订单 MUST 写入非空授权理由。cash-service 启动 `EnsureSchema`（或等价迁移）MUST 幂等补齐该列。

#### Scenario: 新库建表含列

- **WHEN** 空库执行 EnsureSchema
- **THEN** `vip_order` 与 `feature_order` MUST 包含 `grant_reason` 列

#### Scenario: 已有库补列

- **WHEN** 旧库缺少该列且服务启动 EnsureSchema
- **THEN** 系统 MUST 添加该列且 MUST NOT 因重复添加而 Fatal（幂等）

### Requirement: cash-service SHALL 提供 Admin 手工授 VIP API

cash-service MUST 提供 `POST /cash/admin/api/vip/entitlements/grant`，鉴权 MUST 校验网关注入的 `X-Admin-Password`。请求体 MUST 含 `wxId`（>0）、`durationDays`（整数 ≥1）、`reason`（trim 后非空，长度不得超过列上限）。成功时 MUST：

1. 插入一行 `vip_order`：`channel=admin`、`amount_fen=0`、`status=paid`、`paid_at` 为当前时间、`grant_reason` 为理由、`product_code` 为默认 VIP 商品码（与一期商品一致，如 `vip_monthly_19`）；
2. 按与支付相同的续期语义调用权益续期：`new_expire = max(now, current_expire) + durationDays * 86400`。

订单插入与权益写入 MUST 在同一成功路径内完成（推荐同一事务）；任一步失败 MUST 向调用方返回错误且 MUST NOT 声称授予成功。外部支付宝/Apple 回调 MUST NOT 被设计为创建 `channel=admin` 订单的正常路径。

#### Scenario: 合法授 VIP

- **WHEN** 管理员以正确口令提交有效 `wxId`、`durationDays=30`、非空 `reason`
- **THEN** 系统 MUST 写入带 `grant_reason` 的 admin `vip_order`，且该 `wxId` 的 `vip_entitlement.expire_at` MUST 按续期公式延长

#### Scenario: 理由为空拒绝

- **WHEN** 请求 `reason` 为空或仅空白
- **THEN** 系统 MUST 拒绝授予，MUST NOT 插入订单，MUST NOT 修改权益

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝且 MUST NOT 写入订单或权益

### Requirement: cash-service SHALL 提供 Admin 手工授功能 API

cash-service MUST 提供 `POST /cash/admin/api/feature/grants`（路径以实现为准，须落在 `/cash/admin/api/` 下），鉴权同 VIP Admin。请求 MUST 含 `featureId`、非空 `reason`，并按该功能 `activation_subject`：

- `user`：MUST 要求 `wxId` >0，写入账号维权益（或该功能既定 Grant 路径）；
- `device`：MUST 要求非空 `deviceNo`，写入设备维权益或预测数量权威。

`durationDays` 与 `grantQuantity` 按该功能 `grant_kind` / 产品语义使用：权益类 MUST 支持续期天数（与支付 Grant 一致；`durationDays=0` 表示永久仅当该功能既有 Grant 允许且请求显式传递——默认实现可要求权益类 `durationDays≥1`，预测 `allowed_count_delta` 则 MUST 要求 `grantQuantity≥1`）。成功时 MUST 插入 `feature_order`：`channel=admin`、`amount_fen=0`、`status=paid`、`grant_reason` 非空，并完成对应 Grant。

#### Scenario: 授 user 主体功能

- **WHEN** 管理员对 `activation_subject=user` 的功能提交有效 `wxId`、`durationDays≥1`、非空 `reason`
- **THEN** 系统 MUST 写入 admin `feature_order` 与账号维权益续期/授予

#### Scenario: 授 device 主体功能缺 deviceNo

- **WHEN** 功能为 `device` 主体且请求未提供 `deviceNo`
- **THEN** 系统 MUST 拒绝，MUST NOT 落库订单或权益

#### Scenario: 功能授理由必填

- **WHEN** `reason` 为空
- **THEN** 系统 MUST 拒绝授予

### Requirement: Hub VIP 权益页 SHALL 支持手工授 VIP

静态页 `cash-vip-admin.html` MUST 提供手工授 VIP 表单：至少 `wxId`、`durationDays`、`reason`（必填）。提交 MUST 调用授 VIP Admin API（经 `AdminCommon.adminFetch`）。成功后 MUST 刷新权益列表。页面 MUST NOT 在浏览器发送 `X-Admin-Password`。

#### Scenario: 运维从 VIP 页授 VIP

- **WHEN** 已登录 Hub 的运维填写有效表单并提交
- **THEN** 系统 MUST 完成授予，且刷新后列表 MUST 能反映该 `wxId` 的有效权益（或更新后的到期时间）

### Requirement: Hub 开通功能页 SHALL 支持手工授功能

静态页 `cash-feature-admin.html` MUST 提供手工授功能表单：至少 `featureId`、按主体的目标标识、期限或数量、`reason`（必填）。提交 MUST 调用授功能 Admin API。成功后 MUST 可刷新开通快照（若页内已有快照区）。

#### Scenario: 运维从开通功能页授功能

- **WHEN** 已登录运维提交合法功能授表单
- **THEN** 系统 MUST 完成授予且 MUST 落库带 `grant_reason` 的 admin `feature_order`

### Requirement: Admin 手工授 API SHALL 不计入 App usage 统计

`/cash/admin/api/vip/entitlements/grant` 与功能授 Admin 路径 MUST 视为运维管理通道，MUST NOT 按 App 对外接口计入 usage；实现 MUST NOT 仅为这些路径修改 App 侧 `maintenance_skip` 假定。

#### Scenario: 管理端手工授

- **WHEN** 运维通过 Hub 调用手工授 API
- **THEN** 系统 MUST NOT 将其记为需计入的 App 功能使用
