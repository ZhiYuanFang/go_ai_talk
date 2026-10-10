## ADDED Requirements

### Requirement: wx 维永久小智 MCP 能力权益

系统 SHALL 以微信账号（`wx_id`）为开通主体持久化小智 MCP 绑定能力权益。权益一经有效开通 MUST 视为永久（无到期续期模型）。开通来源 MUST 支持支付履约与 Admin 手工授。

#### Scenario: 支付开通后永久有效

- **WHEN** 某 wx 的小智 MCP 订单支付履约成功
- **THEN** 系统为该 `wx_id` 写入（或保持）有效永久权益
- **AND** MUST NOT 写入限时 `expires_at` 依赖 `ActivateFeature` 天数模型

#### Scenario: 未开通无权益

- **WHEN** 某 `wx_id` 从未支付成功且未被手工授
- **THEN** 该账号 MUST 视为未开通小智 MCP 能力

### Requirement: 支付通道复用与 MCP 类型履约

系统 SHALL 复用现有功能开通支付通道（建单调起、支付宝/Apple 回调与验单）处理小智 MCP 买断。建单金额与 Apple 商品 ID MUST 取自小智 MCP 专用 SKU 表。履约时 MUST 识别 MCP 商品类型并授予永久权益，MUST NOT 调用要求「授予天数≥1」的通用功能开通路径。

#### Scenario: MCP 订单履约

- **WHEN** 支付回调或验单命中小智 MCP 的 `product_code` 且订单可履约
- **THEN** 系统标记订单已支付并为订单 `wx_id` 授予永久小智 MCP 能力
- **AND** MUST NOT 调用 `ActivateFeature` 对 care/growth 等功能写权益

#### Scenario: 非 MCP 功能订单不受影响

- **WHEN** 订单商品属于既有 `feature_product` 功能 SKU
- **THEN** 履约行为 MUST 保持原功能开通语义（限时等）

### Requirement: App 开通态与可购 SKU

系统 SHALL 提供经 gateway-app 鉴权的 App 接口，返回当前登录用户是否已开通小智 MCP 能力，以及可售永久 SKU（含 `productCode`、价格、`appleProductId` 等）。该接口 MUST NOT 依赖 `GET /cash/app/api/feature/catalog`。

#### Scenario: 已开通

- **WHEN** 当前 wx 已有有效永久权益并请求开通态
- **THEN** 响应 `unlocked=true`（或等价字段）

#### Scenario: 未开通且 SKU 上架

- **WHEN** 当前 wx 未开通且专用 SKU `status` 为上架
- **THEN** 响应 `unlocked=false` 且包含可建单所需 SKU 信息

### Requirement: App 建单

系统 SHALL 提供 App 建单接口，使用与功能开通一致的支付渠道参数创建小智 MCP 订单，并返回调起支付所需信息。订单 MUST 关联当前登录 `wx_id` 与 MCP `product_code`。

#### Scenario: 成功建单

- **WHEN** 用户已登录、SKU 上架且渠道合法
- **THEN** 系统创建待支付订单并返回调起参数

#### Scenario: 已开通再建单

- **WHEN** 用户已拥有有效永久权益仍请求建单
- **THEN** 系统 MUST 拒绝或幂等提示已开通（实现选一并在错误文案中明确），MUST NOT 重复收费履约出第二份冲突权益语义

### Requirement: 添加绑定须已开通

device 域添加小智 MCP 绑定（App `POST` 添加）MUST 在写入绑定前确认当前 `wx_id` 已具备有效小智 MCP 永久能力。校验 MUST 经 cash 服务接口契约完成，MUST NOT 在 device 进程直查 cash 库表。List、改备注、删除 MUST NOT 因未开通而拒绝。

#### Scenario: 未开通拒绝添加

- **WHEN** 当前 wx 未开通且请求添加绑定
- **THEN** 添加 MUST 失败并返回明确业务错误
- **AND** MUST NOT 写入绑定行

#### Scenario: 已开通允许添加

- **WHEN** 当前 wx 已开通且其它绑定校验通过
- **THEN** 添加 MUST 按既有绑定规则成功（含 speaker_mac / token 唯一等）

#### Scenario: 未开通仍可列表与删除

- **WHEN** 当前 wx 未开通但已有历史绑定行
- **THEN** 列表、改备注、删除 MUST 仍可按既有归属规则执行

### Requirement: 支付退款撤销能力

当小智 MCP 支付订单发生渠道退款时，系统 MUST 撤销对应 wx 的小智 MCP 永久能力。已存在的绑定行 MAY 保留，但后续 Add MUST 因未开通而失败。

#### Scenario: 退款后不可再添加

- **WHEN** MCP 订单退款处理完成
- **THEN** 该 wx 开通态为未开通
- **AND** 再次 Add MUST 失败
