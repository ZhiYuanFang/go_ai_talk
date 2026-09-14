## ADDED Requirements

### Requirement: cash-service MUST 提供 VIP 商品 Admin 读与写

cash-service MUST 提供 `GET /cash/admin/api/vip/product` 与 `POST /cash/admin/api/vip/product`。鉴权 MUST 与既有 cash Admin 一致（网关注入 `X-Admin-Password`）。GET MUST 返回一期 VIP 商品（`productCode` 为 `vip_monthly_19`）的 `title`、`priceFen`、`originalPriceFen`、`durationDays`、`appleProductId`、`status`。POST MUST 更新该行对应字段并持久化到 `ai_voice_cash.vip_product`；MUST NOT 创建第二行 VIP SKU，MUST NOT 修改 `product_code`。`priceFen` MUST 为非负整数（分）；`originalPriceFen=0` 表示无划线价。gateway-app MUST 经既有 `/cash/admin/api/*` 反代到达，MUST NOT 新增顶层前缀。本路径为运维通道，MUST NOT 计入 App usage 统计，MUST NOT 为此修改 `maintenance_skip.go`。

#### Scenario: 管理员读取 VIP 商品

- **WHEN** 已鉴权管理员请求 `GET /cash/admin/api/vip/product`
- **THEN** 系统 MUST 返回 `vip_monthly_19` 的现价、原价、时长与 Apple 商品 ID

#### Scenario: 管理员更新现价与 Apple 商品 ID

- **WHEN** 已鉴权管理员 POST 合法的 `priceFen` 与 `appleProductId`
- **THEN** 系统 MUST 写入 `vip_product`，且后续 App 读价与建单 MUST 使用新现价，Apple 建单/验单 MUST 使用新 `appleProductId`

#### Scenario: 未鉴权拒绝写

- **WHEN** 请求未携带有效 Admin 凭证
- **THEN** 系统 MUST 拒绝 POST 且 MUST NOT 修改 `vip_product`

### Requirement: VIP 与功能建单 MUST 只认库内 Apple 商品 ID

`GetActiveProduct`、VIP Apple 验单以及 EnsureSchema 种子 MUST NOT 读取 `CASH_APPLE_PRODUCT_ID` 或 yaml `cash.appleProductId`。VIP 的 Apple 商品 ID 真相源 MUST 仅为 `vip_product.apple_product_id`。当该字段为空时，VIP Apple 建单 MUST NOT 伪造 productId；响应可带说明需在开通功能管理配置的 `payTip`。部署模板（`.env.example`、cash-service compose 环境注入、cash-service 配置注释）MUST 删除 `CASH_APPLE_PRODUCT_ID`。`CASH_APPLE_BUNDLE_ID` MUST 保留（JWS bundle 校验）。功能 SKU 的 Apple ID 仍在 `feature_product`，MUST NOT 改回 env。

#### Scenario: 库中已配置 Apple 商品 ID

- **WHEN** `vip_product.apple_product_id` 为非空且进程环境未设置 `CASH_APPLE_PRODUCT_ID`
- **THEN** `GET /cash/app/api/vip/product` 与 VIP Apple 建单 MUST 返回该库值

#### Scenario: 库中未配置 Apple 商品 ID

- **WHEN** `vip_product.apple_product_id` 为空
- **THEN** 系统 MUST NOT 从环境变量填入 Apple 商品 ID

#### Scenario: 部署模板无 CASH_APPLE_PRODUCT_ID

- **WHEN** 运维按 `.env.example` 与 compose 部署 cash-service
- **THEN** 模板与 compose 注入 MUST NOT 再出现 `CASH_APPLE_PRODUCT_ID`

### Requirement: 开通功能管理页 MUST 可编辑 VIP 套餐

`cash-feature-admin.html` MUST 提供独立于功能定义列表的 VIP 套餐表单，字段至少含现价（分）、原价（分）、时长、Apple 商品 ID；保存 MUST 调用 VIP 商品 Admin 写接口。页面 MUST NOT 把 VIP 渲染为一条可点选的 `feature_def`。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 保存 VIP 价格

- **WHEN** 管理员在开通功能管理填写 VIP 现价并保存
- **THEN** 页面 MUST 经 adminFetch POST Admin VIP 商品接口且成功后主内容反映新值

#### Scenario: VIP 不出现在功能定义列表

- **WHEN** 管理员查看开通功能管理的功能定义列表
- **THEN** 列表 MUST NOT 将 `vip` 或 `vip_monthly_19` 作为功能编号行展示
