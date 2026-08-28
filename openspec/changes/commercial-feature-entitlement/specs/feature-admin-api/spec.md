## ADDED Requirements

### Requirement: cash-service MUST 提供付费功能管理 Admin CRUD API

cash-service MUST 在 `/cash/admin/api/feature/**`（或等价前缀）提供管理端 API，支持功能定义（`feature_def`）、解锁方式、功能 SKU（`feature_product`）、预测数量授予相关配置的创建/读取/更新/停用。写操作 MUST 校验 Admin 鉴权（与现网 cash Admin 一致）。Admin API MUST NOT 改变 App usage 统计 denylist 的默认假设（管理端前缀）。

#### Scenario: 创建功能定义

- **WHEN** 管理员提交合法的新功能定义
- **THEN** 系统 MUST 持久化到 `ai_voice_cash` 并在后续目录/管理列表中可见

#### Scenario: 配置功能 SKU

- **WHEN** 管理员创建或更新功能 `product_code` 与价格/授予语义
- **THEN** 系统 MUST 持久化，且支付建单 MUST 能引用该 SKU

#### Scenario: 未鉴权拒绝

- **WHEN** 请求未携带有效 Admin 凭证
- **THEN** 系统 MUST 拒绝写操作

### Requirement: cash-service MUST 提供邀请码管理 Admin API 含兑换明细

系统 MUST 提供邀请码 Admin API（建议 `/cash/admin/api/invite-code/**`）：支持创建/读取/更新/停用邀请码；创建时 MUST 绑定 `owner_wx_id`；一期 MUST 强制一 `owner_wx_id` 仅一条有效码；MUST 可配置有效期与可开通功能集合。MUST 提供按码查询兑换明细：至少含 `redeemer_wx_id`、`device_no`、`featureId`、`redeemed_at`，供追踪销售推广。

#### Scenario: 创建邀请码绑定 owner

- **WHEN** 管理员为某 `owner_wx_id` 创建邀请码并指定有效期与可开功能
- **THEN** 系统 MUST 保存码记录；若该 owner 已有码则 MUST 拒绝或按产品约定停用旧码（一期一人一码）

#### Scenario: 查看某码被哪些人使用

- **WHEN** 管理员查询某邀请码的兑换明细
- **THEN** 系统 MUST 返回使用该码成功兑换的记录列表，含兑换人 `wx_id` 与时间

### Requirement: Admin API MUST 经 gateway-app 反代且禁止他域直查

`/cash/admin/api/feature/*` 与邀请码 Admin 路径 MUST 由 gateway-app cash 反代到达 cash-service。其他非 cash 进程 MUST NOT 直查 `ai_voice_cash` 功能表。

#### Scenario: 反代可达

- **WHEN** 已登录 Hub 的管理员调用功能或邀请码 Admin API
- **THEN** 请求 MUST 经 gateway 到达 cash-service 并成功鉴权（与 VIP Admin 同模式）
