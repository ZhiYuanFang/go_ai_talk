## ADDED Requirements

### Requirement: cash-service MUST 提供 Apple ASN V2 异步通知入口

cash-service MUST 提供 `POST /cash/app/api/vip/apple/notifications`：gateway-app MUST 将该精确 path 列入 Bearer 匿名白名单（与支付宝 notify 同模式）。请求体 MUST 按 App Store Server Notifications V2 解析；系统 MUST 校验 `signedPayload` 密码学签名。验签失败 MUST NOT 履约。校验通过后 MUST 在同一请求路径内同步履约或处理退款（MUST NOT 依赖 RabbitMQ 或扫表）。对 Apple 的成功应答 MUST 为 HTTP 2xx，以便停止无意义重试；业务上无法映射订单但仍验签成功的可忽略类型 MUST 返回 2xx 并记日志，避免毒害重试队列。

#### Scenario: 合法购买通知开通

- **WHEN** Apple 投递验签通过的购买成功类通知，且 `appAccountToken` 对应未支付的 VIP 或功能订单
- **THEN** 系统 MUST 将订单置 `paid`（若尚未），并按该订单类型开通 VIP 或功能权益

#### Scenario: 验签失败

- **WHEN** `signedPayload` 签名非法
- **THEN** 系统 MUST NOT 开通权益，MUST NOT 将订单标为 paid

#### Scenario: 重复通知幂等

- **WHEN** 同一 Apple `transactionId`（渠道交易号）对已 paid 订单再次通知
- **THEN** 系统 MUST 幂等成功且 MUST NOT 错误地重复叠加权益时长或数量

### Requirement: Apple 建单 MUST 生成并返回 appAccountToken UUID

当 `channel=apple_iap` 创建 VIP 或功能订单时，系统 MUST 生成 UUID 写入订单的 `app_account_token`，并在建单响应中返回 `appAccountToken`。客户端 MUST 将该值作为 StoreKit 购买的 `appAccountToken`。ASN 履约 MUST 使用通知中的 `appAccountToken` 查找订单；缺失 token 或找不到订单时 MUST NOT 凭空开通（可记日志）。

#### Scenario: Apple 建单返回 token

- **WHEN** 用户以 `channel=apple_iap` 成功建单
- **THEN** 响应 MUST 含非空 UUID 格式的 `appAccountToken`，且库中对应订单行 MUST 存有相同值

#### Scenario: 无 token 无法 ASN 开通

- **WHEN** 通知验签通过但 `appAccountToken` 为空或无匹配订单
- **THEN** 系统 MUST NOT 开通权益

### Requirement: REFUND 通知 MUST 收回对应开通

当验签通过的通知类型为退款（`REFUND`）时，系统 MUST 将该渠道交易关联订单标记为退款语义，并撤销或失效该笔支付所授予的 VIP 续期或功能权益（实现须可测、幂等）。MUST NOT 因退款通知开通新权益。

#### Scenario: 退款收回 VIP

- **WHEN** 某已履约 VIP 订单的 `transactionId` 收到合法 `REFUND`
- **THEN** 该账号因该笔支付获得的权益 MUST 被撤销或调整为不再有效（按实现约定），且重复 REFUND MUST 幂等

### Requirement: gateway 与 usage 对 ASN 的约定

gateway-app MUST 反代 `/cash/app/api/*` 使 ASN path 可达。`POST /cash/app/api/vip/apple/notifications` MUST NOT 计入 App usage 统计，MUST 写入 `maintenance_skip`（或等价精确排除）。

#### Scenario: 匿名可达

- **WHEN** 无用户 Bearer 的 Apple 服务器 POST 该 notifications path
- **THEN** gateway MUST 放行至 cash-service（由 cash 验签）

#### Scenario: 不计入 usage

- **WHEN** ASN 回调成功处理
- **THEN** 系统 MUST NOT 将其记为需展示的 App 功能使用次数
