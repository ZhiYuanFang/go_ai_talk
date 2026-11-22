# care-commerce

## Purpose

护理留意点、现金 VIP、功能解锁、试用/额度与商业履约的行为规格。 Care-alert gates plus cash/vip/feature catalog commerce.

<!-- migrated-from: apple-iap-cross-wx-reject, apple-server-notifications, care-alert-access-gate, care-alert-cash-vip, care-alert-feature-unlock, care-alert-feeding-eligibility, care-alert-force-refresh, care-alert-python-wire, care-alert-quota-lane, care-alert-unlock-client, care-alert-vip-selection, cash-admin-grant-revoke, cash-admin-manual-grant, cash-admin-revenue, cash-duration-days, cash-vip-admin-api, catalog-activatable-total, feature-activation-subject, feature-activation-toolkit, feature-ad-unlock-remove, feature-branding-client, feature-catalog-entitlement, feature-catalog-sku-embed, feature-order-query, feature-sku-update, feature-unlock-fulfillment, feature-visual-branding, retire-vip-care-growth, vip-entitlement, vip-overlay-entitlement, vip-payment, vip-product-pricing, vip-quota-entitlement, wx-account-vip -->

## Requirements

### Requirement: Apple IAP 已履约交易号异 wx MUST 拒绝授权

当 `channel=apple_iap` 且存在同 `channel_txn_id`（Apple `transactionId`）的已 `paid` 订单时，cash-service 在 VIP（`vip_order`）、功能与小智 MCP（`feature_order`）履约路径上 MUST 比较该已履约订单的 `wx_id` 与本次履约目标订单的 `wx_id`：

- 二者相同：MUST 幂等成功，MUST NOT 重复叠加权益或重复授永久能力。 - 二者不同：MUST 拒绝履约，MUST NOT 将目标订单标为 `paid`，MUST NOT 向目标 `wx_id` 授予权益； 错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」。 本规则 MUST 适用于 `FulfillPaid`、`FulfillFeaturePaid`、`FulfillXiaozhiMcpPaid`（含经 `VerifyAppleIAP` 与 `DispatchFulfillPaid` 的调用）。 `channel` 非.

#### Scenario: 换号复用同一 Apple transactionId

- **WHEN** 账号 A 的订单已因某 `transactionId` 处于 `paid`，账号 B 以新订单对该 `transactionId` 发起 Apple 履约（含 `POST /cash/app/api/vip/apple/verify`）
- **THEN** 系统 MUST 返回业务错误，文案为「一个 Apple 账号仅支持开通一个胖宝账号」，且 MUST NOT 为 B 写入权益或将 B 的订单标为 `paid`

#### Scenario: 同一 wx 重复 verify 或 ASN 后再 verify

- **WHEN** 同一 `wx_id` 对已履约的同一 `transactionId` 再次走 verify 或 ASN 履约
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加权益

#### Scenario: 非 Apple 渠道不受本规则影响

- **WHEN** `channel=alipay` 且同 `channel_txn_id` 已对应一笔 `paid` 订单再次回调
- **THEN** 系统 MUST 按既有支付宝幂等成功处理，MUST NOT 仅因本能力返回 Apple 专用文案

---

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

---

### Requirement: cash-service MUST 提供值得留意可看合成内部契约

cash-service MUST 提供经内部密钥保护的只读接口（路径以实现为准，如 `GET /cash/internal/api/care-alert/access`），入参 MUST 含 `deviceNo` 与 `wxId`。 响应 MUST 至少包含：`allowed`（bool）、`feedingQualified`（bool）、`featureActive`（bool）。 合成规则 MUST 为：`feedingQualified` 取自与 App `care-alert/eligibility` 同构的 `care_alert_entry` 判定； VIP / 功能开通 MUST NOT 将 `feedingQualified` 置 true。 voice-service MUST NOT 直查 `ai_voice_cash`。

#### Scenario: 双门槛均满足

- **WHEN** 设备喂养资格合格且（有未过期值得留意权益或账号 VIP）
- **THEN** 响应 MUST `allowed=true`、`feedingQualified=true`、`featureActive=true`

#### Scenario: 仅喂养合格未开通非 VIP

- **WHEN** `feedingQualified=true` 且无有效权益且非 VIP
- **THEN** 响应 MUST `allowed=false`、`featureActive=false`

#### Scenario: 已开通但喂养不合格

- **WHEN** 设备有永久值得留意权益但喂养资格不合格
- **THEN** 响应 MUST `allowed=false`、`feedingQualified=false`、`featureActive=true`

#### Scenario: VIP 覆盖开通但不覆盖喂养

- **WHEN** 用户 VIP 有效、无设备权益、喂养不合格
- **THEN** 响应 MUST `feedingQualified=false`、`featureActive=true`、`allowed=false`

### Requirement: voice CareAlert 业务接口 MUST 经 access 双门禁

voice-service 在处理值得留意日列表（`CareAlertDaily`，含强制刷新与缓存命中路径）时 MUST 先调用 cash 值得留意 access 契约；仅当 `allowed=true` 时 MAY 返回缓存内容或触发生成。当 cash 不可达、超时或返回错误时，系统 MUST fail-closed 拒绝该次业务（MUST NOT 当作已开通放行），并记录 Warning。删除单条与反馈接口 MUST 同样调用 access 且仅在 `allowed=true` 时执行变更（与 Daily 一致，防止过期后改缓存）。

#### Scenario: 未开通拒绝 daily

- **WHEN** access 返回 `allowed=false`（例如合格未开通）
- **THEN** CareAlertDaily MUST 返回错误或空拒绝语义，MUST NOT 返回日列表内容，MUST NOT 调用 Python 生成

#### Scenario: 缓存命中仍校验

- **WHEN** 当日日缓存存在但 access 返回 `allowed=false`（如邀请已过期）
- **THEN** 系统 MUST NOT 向客户端返回该缓存内容

#### Scenario: cash 故障 fail-closed

- **WHEN** voice 调用 cash access 失败
- **THEN** CareAlertDaily MUST 失败关闭，MUST NOT 降级为跳过开通检查

---

### Requirement: care-alert 经 cash-service 判定触发者 VIP

voice-service 在护理留意日缓存未命中、需要选模时，MUST 使用当前请求触发者的 `wxId`，经 HTTP 调用 cash-service 内部 VIP 接口判定是否 VIP：VIP MUST 选用 DeepSeek；非 VIP MUST 选用 Zhipu。MUST NOT 查询 `wx.is_vip`，MUST NOT 调用 device-service 的 VIP 接口，MUST NOT 使用 `deviceNo` 反查作为 VIP 或登录依据。日缓存键 MUST 仍为 `deviceNo + Asia/Shanghai` 自然日；缓存命中 MUST NOT 重跑 LLM。本路径 MUST NOT 扣减 clinic AI 配额。

#### Scenario: VIP 触发者首次生成

- **WHEN** 当日缓存未命中且 cash 返回该触发者 `isVip=true`
- **THEN** 服务 MUST 以 DeepSeek 配置调用 Python 分析并写入宝宝日缓存

#### Scenario: 非 VIP 触发者首次生成

- **WHEN** 当日缓存未命中且 cash 返回 `isVip=false`
- **THEN** 服务 MUST 以 Zhipu 配置调用 Python 分析并写入宝宝日缓存

#### Scenario: cash VIP 查询失败降级

- **WHEN** 调用 cash 内部 VIP 接口超时或失败
- **THEN** 服务 MUST 按非 VIP（Zhipu）继续生成（其它步骤成功时），MUST 打 Warning，MUST NOT 因 VIP 查询失败单独使整个 daily 失败

### Requirement: care-alert 必须携带有效 wxId

`GET/DELETE/POST` 之 `/device/api/care-alert/*` MUST 要求 `X-Internal-Wx-Id` 解析后 `wxId>0`；纯设备会话 MUST 被拒绝。

#### Scenario: 缺少 wxId

- **WHEN** care-alert 请求缺少有效 `X-Internal-Wx-Id`
- **THEN** 服务 MUST 返回错误且 MUST NOT 完成需登录的生成/删除/飞轮成功语义

### Requirement: 拆除 device 域 wx.is_vip 路径

本变更 MUST 移除（或停止注册）下列由 `care-alert-vip-by-wx` 引入的路径，使 VIP 真相源唯一落在 cash-service：`wx` 表 `is_vip` 模型字段与 DDL 执行指引、`GET /device/app/api/user/internal/vip-by-wx-id`、voice 经 device 的 `RemoteIsVipByWxID`。`hack/ddl_wx_is_vip.sql` MUST NOT 作为发布必执行项。

#### Scenario: device 不再提供 VIP 内部接口

- **WHEN** 客户端或 voice 请求已删除的 device `vip-by-wx-id` 路径
- **THEN** 该接口 MUST 不可用（404 或未注册），调用方 MUST 改用 cash internal VIP 接口

#### Scenario: 发布文档不再要求执行 is_vip DDL

- **WHEN** 按本变更更新后的 runbook 部署 VIP 能力
- **THEN** 文档 MUST 指引部署 `cash-service` / `ai_voice_cash`，MUST NOT 要求执行 `ddl_wx_is_vip.sql` 作为 VIP 前置

---

### Requirement: 系统 MUST 提供值得留意智能提醒功能定义与设备维开通

cash-service MUST 种子（或等价确保存在）稳定功能编号 `care_alert_smart_remind`（标题面向「值得留意智能提醒」），启用开通方式 MUST 包含 `payment`、`invite_code`、`ad`。权益主体 MUST 为 `device_no`（全家共享）。付费履约 MUST 授予永久 entitlement（SKU `duration_days=0` 或等价）。邀请码与广告开通 MUST 按功能定义上的邀请/广告授予天数限时授予（种子默认 7，Admin 可改）。VIP 购买/续期 MUST NOT 写入该功能的 `feature_entitlement` 行。

#### Scenario: 付费永久开通

- **WHEN** 用户支付值得留意开通商品成功
- **THEN** 该 `device_no` 对应 `care_alert_smart_remind` entitlement 的 `expires_at` MUST 为 0（永久）或等价永久语义

#### Scenario: 邀请限时开通

- **WHEN** 用户用邀请码兑换 `care_alert_smart_remind` 且定义天数为 7
- **THEN** 该设备 entitlement MUST 为约 7 天有效，且 unlock_method 反映邀请码

#### Scenario: VIP 不写功能表

- **WHEN** 账号 VIP 履约成功
- **THEN** 系统 MUST NOT 仅为值得留意插入或更新 `feature_entitlement`

### Requirement: 值得留意邀请开通 MUST 按设备仅成功一次且原力记用户

对 `care_alert_smart_remind`，系统在邀请码兑换成功时 MUST 以 `(device_no, feature_id)` 约束该设备仅能成功邀请开通一次（任意邀请码、任意兑换账号）。重复兑码 MUST 拒绝且 MUST NOT 延长权益或再次发放获客原力。获客原力与兑换流水 MUST 仍按邀请码主人/兑换用户维度记录，MUST NOT 因设备主体而改记到「设备账号」。人×码×功能去重 MUST 继续生效。本约束 MUST NOT 应用于 `prediction_unlock`。

#### Scenario: 同设备第二次邀请拒绝

- **WHEN** 设备 D 已成功用任意码开通 `care_alert_smart_remind`，同设备另一账号再次用其它码兑换该功能
- **THEN** 系统 MUST 拒绝，MUST NOT 更新 entitlement，MUST NOT 再次给码主人加原力

#### Scenario: 预测不受设备一次限制

- **WHEN** 同一设备依次用两个不同好友码兑换 `prediction_unlock`（且同设备互兑规则通过）
- **THEN** 两次均可成功且永久条数各 +1

#### Scenario: 原力仍记用户

- **WHEN** 用户 U 兑码为设备 D 开通值得留意成功
- **THEN** 获客原力 MUST 记到码主人用户侧，兑换流水 MUST 含 U 的 `wx_id` 与 D 的 `device_no`

### Requirement: 目录与 VIP 覆盖语义 MUST 对齐预测槽模式

合成功能目录 MUST 返回 `care_alert_smart_remind` 的真实设备 `unlocked` / `expiresAt`，MUST NOT 因请求者 `isVip` 将 `unlocked` 改写为 true。业务使用态（含 access 合成）MUST 在「设备权益未过期」与「当前 isVip」之间按 OR 放行。UCG 与值得留意**喂养**资格 MUST 继续不受 isVip / 功能开通影响。

#### Scenario: VIP 使用态可看但 catalog 真实

- **WHEN** 用户 isVip=true 且设备无值得留意 entitlement
- **THEN** catalog 该项 `unlocked` MUST 为 false，但业务可看判定 MUST 为允许（在喂养合格前提下）

#### Scenario: 过期后非 VIP 不可用

- **WHEN** 邀请授予已过期且用户非 VIP
- **THEN** catalog `unlocked` MUST 为 false，业务可看判定中 featureActive MUST 为 false

---

### Requirement: 值得留意喂养资格 App API MUST 由 cash-service 提供且与 UCG 同构

**cash-service** MUST 提供登录+绑机可达的 `GET /cash/app/api/care-alert/eligibility`（或等价路径），按 `care_alert_entry` 配置在 cash 内合成并返回至少：`qualified`、`requiredDays`、`effectiveDays`、`remainingDays`，并 MAY 返回 `message`。`device_no` MUST 只信网关内部头。该路径 MUST NOT 加入 Bearer 匿名白名单。资格结果 MUST 可按日 Redis 缓存且 MUST NOT 落 MySQL 权威。history 失败 MUST fail-closed。device-service MUST NOT 作为该喂养天数资格的权威计算宿主。

#### Scenario: 连续满配置天数合格

- **WHEN** `care_alert_entry.requiredDays=2` 且连续有效日 ≥ 2
- **THEN** 响应 MUST `qualified=true`、`remainingDays=0`

#### Scenario: 不足返回剩余天数

- **WHEN** `requiredDays=2` 且 `effectiveDays=1`
- **THEN** 响应 MUST `qualified=false`、`remainingDays=1`

#### Scenario: 资格 API 归属 cash

- **WHEN** 客户端查询值得留意喂养是否达标
- **THEN** 请求 MUST 打到 cash App eligibility 路径，MUST NOT 依赖 device care-alert 接口返回喂养连续天数资格

### Requirement: 客户端 MUST 以 cash 服务端资格替代「昨日有发生」闸门

Flutter 值得留意展示流 MUST 先请求 **cash** care-alert eligibility。 当 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 仍展示「值得留意」卡片，并 MUST 使用 eligibility 返回的 `effectiveDays` / `requiredDays` / `remainingDays` **由客户端拼接**进度文案，告知用户已累计的有效日（及/或剩余所需日），且 MUST NOT 调用值得留意生成或日列表数据接口。 当 `qualified=true` 时，客户端 MUST 再按原逻辑请求值得留意接口并展示内容。 客户端 MUST NOT 再以「上海昨日有发生」作为是否可拉取值得留意数据的前提，MUST NOT 用本地 history 自行判定权威 `qualified`。 客户端 MUST NOT 被要求向用户声明「今日不计入」； 服务端可选 `message` MUST NOT 作为进度数字的权威来源。 本资格门槛的产品目的为促活.

#### Scenario: 未合格展示已累计有效日

- **WHEN** eligibility 返回 `qualified=false`、`effectiveDays=1`、`requiredDays=2`、`remainingDays=1`
- **THEN** UI MUST 展示值得留意卡片，并由客户端拼出反映已累计有效日与所需/剩余日的进度提示，且 MUST NOT 请求 care-alert daily/生成接口

#### Scenario: 合格后按原逻辑拉数

- **WHEN** eligibility 返回 `qualified=true`
- **THEN** 客户端 MUST 允许按既有逻辑请求值得留意数据接口并展示结果

#### Scenario: 资格失败 fail-closed

- **WHEN** eligibility HTTP 失败
- **THEN** 客户端 MUST NOT 当作已合格去请求生成接口

#### Scenario: 不强制解释今日排除

- **WHEN** 展示未合格进度文案
- **THEN** 文案 MUST NOT 被规格要求包含「今日不计入」或等价规则说明书表述

### Requirement: UCG 进度文案同样由客户端拼字段

UCG 入场资格展示（若展示未合格进度）MUST 使用 cash UCG eligibility 的 `effectiveDays` / `requiredDays` / `remainingDays` 由客户端拼接；MUST NOT 依赖服务端 `message` 作为已累计有效日数字的唯一来源；MUST NOT 强制向用户声明「今日不计入」。

#### Scenario: UCG 未合格展示有效日进度

- **WHEN** UCG eligibility 返回 `qualified=false` 且 `effectiveDays=3`、`requiredDays=7`
- **THEN** 客户端进度展示 MUST 能体现已累计有效日为 3（或等价「3/7」「还差 4」等由字段可推导的表述）

---

### Requirement: care-alert daily 支持 force 强刷

`GET /device/api/care-alert/daily` MUST 支持强制刷新参数（query 名实现为 `force=1` 或等价布尔约定）。当 force 为真时，在通过既有鉴权（App Bearer，网关注入 `X-Internal-Wx-Id>0`）之后，服务端 MUST 删除该 `deviceNo` 当日（Asia/Shanghai）care-alert 日缓存键（经 `cachekit` 既有 `CareAlertDailyKey` builder），再执行与未命中缓存相同的生成/single-flight 流程并写回缓存。无 force 或 force 为假时，行为 MUST 与现网一致（命中日缓存则直接返回且不重跑 LLM）。

#### Scenario: 无 force 命中缓存不重生

- **WHEN** 当日缓存已存在且请求未带 force
- **THEN** 响应 MUST 返回缓存 items，MUST NOT 再次调用 Python 生成

#### Scenario: force 清缓存后重生

- **WHEN** 当日缓存已存在且请求带 force=1（或等价真值），且 wxId>0
- **THEN** 服务端 MUST 删除当日日缓存后重新生成并返回新 items，MUST 写回日缓存

#### Scenario: force 仍要登录

- **WHEN** 请求带 force 但缺少有效 App 鉴权导致 wxId 无效
- **THEN** 服务端 MUST 拒绝请求，MUST NOT 因 force 而跳过鉴权

### Requirement: force 不引入鉴权旁路

care-alert 强刷 MUST NOT 依赖 Admin JWT、MUST NOT 新增仅 Admin 可调的「伪造 wx」接口作为唯一强刷路径。运维调试台 MUST 使用与 App 相同的 daily+force 契约。

#### Scenario: Admin token 不能代替 App 强刷

- **WHEN** 仅携带 Admin JWT 调用 care-alert daily（含 force）
- **THEN** 网关/服务 MUST 按 App 用户接口规则拒绝（与现网 Admin 不可访问 App 接口一致）

---

### Requirement: care-alert 分析请求以 model 字段传首选模型

voice-service 调用 Python `POST /v1/care-alert/analyze` 时，MUST 将 `ResolveLaneModel` 得到的非空模型配置序列化为 JSON 字段 **`model`**（对象形状：`provider`、`name`、`max_in_flight`，与 clinic/tip/intent 的 `PythonModelCfg` 一致）。MUST NOT 依赖 JSON 字段 `model_cfg` 作为 Python 可读契约。当选模结果为 omit（指针 nil）时，请求体 MUST NOT 出现 `model` 字段，以便 Python 走免费保底序。

#### Scenario: premium 或 free 有配置时传 model 对象

- **WHEN** 日缓存未命中且 `ResolveLaneModel` 返回非 nil 的 `*PythonModelCfg`
- **THEN** 发往 Python 的 JSON MUST 包含 `model` 对象，且其 `provider`/`name` 与该配置一致
- **AND** 请求体 MUST NOT 以 `model_cfg` 作为唯一模型载体（不得仅发 `model_cfg` 而省略 `model`）

#### Scenario: omit 时不传 model

- **WHEN** 日缓存未命中且选模结果为 nil（非 premium 且 lane free 为空）
- **THEN** 发往 Python 的 JSON MUST 省略 `model` 字段
- **AND** 服务 MUST 仍可调用分析接口（由 Python 保底序处理）

#### Scenario: 与兄弟仓字段名对齐

- **WHEN** 对照 `python_ai_talk` 的 `CareAlertAnalyzeRequest`
- **THEN** Go 序列化字段名 MUST 为 `model`（可选），MUST NOT 要求 Python 解析 `model_cfg`

---

### Requirement: care-alert 独立额度 feature

voice-service MUST 提供独立 AI 额度 feature **`care_alert`**（与 `voice_ai` / `clinic_ai` 并列），含全局默认上限、per-wxId override、按月用量与 App/Admin 可读状态。care-alert 日生成路径 MUST 使用该 feature 参与统一 premium 判定，MUST NOT 再用 `clinic_ai` 代替 care-alert 额度，MUST NOT 仅根据 VIP 布尔硬切 DeepSeek/Zhipu。

#### Scenario: 非 VIP 有 care_alert 额度走正式模

- **WHEN** 日缓存未命中，触发者非 VIP，且 `care_alert` 额度 Allowed=true
- **THEN** 服务 MUST 按 `careAlert` lane 正式模型调用 Python 分析

#### Scenario: 非 VIP 且 care_alert 用尽走 free/omit

- **WHEN** 日缓存未命中，触发者非 VIP，且 `care_alert` 额度 Allowed=false
- **THEN** 服务 MUST 按 `careAlert` lane 的 free 配置选模；free 为空 MUST omit model 交 Python

#### Scenario: VIP 走正式模

- **WHEN** 日缓存未命中且触发者为 VIP
- **THEN** 服务 MUST 按 `careAlert` lane 正式模型调用 Python，MUST NOT 因 used 达旧 limit 改走 free

#### Scenario: Hub 可配置 care_alert 月度额度

- **WHEN** 管理员打开 `/device/admin/voice-admin.html` 或调用 voice Admin `ai-quota` API
- **THEN** 全局默认 MUST 可读/写 `careAlertMonthlyLimit`；用户列表 MUST 展示 `careAlert` 已用与上限，且 PUT 用户 override MUST 可提交 `careAlertMonthlyLimit`

### Requirement: care-alert 成功时非 VIP 计次

护理留意日分析在 Python 成功并写入（或准备写入）日缓存的成功路径上，若触发者 **非 VIP**，系统 MUST 对 `care_alert` 执行 consume；若触发者为 VIP，MUST NOT consume。缓存命中直接返回时 MUST NOT check/consume 额度，MUST NOT 重新选模。

#### Scenario: 非 VIP 首次生成成功扣次

- **WHEN** 非 VIP 触发者缓存未命中且生成成功
- **THEN** `care_alert` 月用量 MUST 增加 1（在 limit 允许的成功语义下）

#### Scenario: VIP 首次生成不计次

- **WHEN** VIP 触发者缓存未命中且生成成功
- **THEN** `care_alert` 月用量 MUST NOT 因本次成功增加

#### Scenario: 缓存命中不计次

- **WHEN** 当日缓存已存在
- **THEN** 服务 MUST 直接返回列表，MUST NOT consume `care_alert`

### Requirement: careAlert LLM lane 与并发

系统 MUST 提供独立 lane **`careAlert`**，具备与其它 voice lane 相同的 `provider`/`model`/`maxInFlight`/`maxWaiters` 及 free 字段，并经 aimodel 闸门 `Acquire` 控制并发（premium 或已配置 free 时按 design 约定 Acquire）。care-alert 生成 MUST 使用该 lane，MUST NOT 再借用 clinic 闸门数值并硬编码供应商切换。

#### Scenario: Admin 可配置 careAlert

- **WHEN** 管理员 GET/PUT `/voice/admin/api/llm-lanes`
- **THEN** 响应/请求 MUST 包含 `careAlert` 子对象（含正式模与 free 字段）

#### Scenario: 生成走 careAlert 闸门

- **WHEN** premium 路径调用 Python care-alert 分析
- **THEN** 服务 MUST 在调用前对该 lane 执行 Acquire（队列满语义与其它 lane 一致）

### Requirement: care-alert 仍要求有效 wxId

care-alert HTTP 接口 MUST 继续要求 `wxId>0`；MUST NOT 赋予硬件特权；MUST NOT 用 deviceNo 反查替代登录。

#### Scenario: 缺少 wxId

- **WHEN** care-alert 请求无有效 `X-Internal-Wx-Id`
- **THEN** 服务 MUST 拒绝，MUST NOT 生成或计次

---

### Requirement: Flutter 值得留意展示 MUST 区分喂养门槛与开通门槛

Flutter 值得留意卡片流 MUST 继续先请求 cash 喂养资格。当喂养 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 展示进度或失败提示，MUST NOT 请求值得留意 daily，MUST NOT 仅因未开通跳转开通中心。当喂养合格且设备该功能未有效开通且账号非 VIP 时，客户端 MUST 展示可点击的开通引导态，用户点击后 MUST 进入功能开通中心（宜定位到 `care_alert_smart_remind`）。当喂养合格且（目录显示已开通未过期 **或** 当前 isVip）时，客户端 MUST 按既有逻辑请求值得留意 daily 并展示。

#### Scenario: 未喂养合格不进开通中心

- **WHEN** eligibility `qualified=false`
- **THEN** UI MUST 展示喂养进度类文案，点击 MUST NOT 作为开通中心的主路径要求

#### Scenario: 合格未开通点击进开通中心

- **WHEN** eligibility `qualified=true` 且非 VIP 且 catalog 中 `care_alert_smart_remind` 未有效开通
- **THEN** 用户点击值得留意引导区域 MUST 导航至开通中心

#### Scenario: VIP 或已开通拉 daily

- **WHEN** eligibility `qualified=true` 且（isVip 或功能已开通未过期）
- **THEN** 客户端 MUST 允许请求 daily 并展示结果

### Requirement: 开通中心 MUST 展示值得留意项且 VIP 覆盖与预测模式一致

功能目录拉取后，开通中心 MUST 展示 `care_alert_smart_remind`（功能启用时）。有效开通展示 MUST 使用「设备 `unlocked` 未过期 **或** isVip」的 OR 语义（与非预测项既有 `isFeatureEffectivelyUnlocked` 一致）。catalog 的 `unlocked` MUST 仍反映设备真实权益。客户端 MUST NOT 用 isVip 绕过喂养资格。

#### Scenario: 非 VIP 未开通显示可开通

- **WHEN** 非 VIP 用户打开开通中心且设备未开通值得留意
- **THEN** 该项 MUST 展示为未开通并可走支付/邀请（及若启用的广告）开通

#### Scenario: VIP 覆盖显示

- **WHEN** VIP 用户打开开通中心且设备无值得留意权益行
- **THEN** 有效开通态 MUST 视为可用，且 MUST NOT 要求客户端把 catalog `unlocked` 改写成服务端已开通

---

### Requirement: care-alert 必须携带有效 wxId (variant 2)

voice-service 上经 gateway 暴露的 care-alert HTTP（`GET /device/api/care-alert/daily`、`DELETE /device/api/care-alert/daily/item`、`POST /device/api/care-alert/feedback`）MUST 要求请求携带有效 `X-Internal-Wx-Id`（解析后 `wxId > 0`）。纯设备会话（`wxId=0` 或缺失头）MUST 被拒绝。鉴权路径 MUST NOT 使用按 `deviceNo` 反查 `wxId` 作为登录替代。

#### Scenario: 缺少或无效 wxId

- **WHEN** 客户端调用任一 care-alert 接口且 `X-Internal-Wx-Id` 缺失、非正整数或为 `0`
- **THEN** 服务 MUST 返回错误且 MUST NOT 读写日缓存生成路径中的 LLM，也 MUST NOT 完成需登录的删除/飞轮成功语义

#### Scenario: 有效 wxId 与 deviceNo

- **WHEN** 请求同时具备 `wxId > 0` 与合法 `deviceNo`
- **THEN** 服务 MAY 继续执行宝宝维度的缓存/忽略/飞轮逻辑；VIP 判定主键 MUST 为该 `wxId`

### Requirement: 按触发者 VIP 选模且与日缓存维度分离

护理留意首次日生成（缓存未命中）时，voice-service MUST 按**当前请求触发者**的 `wxId` 查询账号 VIP：VIP 为真 MUST 选用 DeepSeek；非 VIP MUST 选用 Zhipu。日缓存键 MUST 仍为 `deviceNo + Asia/Shanghai` 自然日；缓存命中 MUST NOT 因读缓存再次查询 VIP 或重跑 LLM。本路径 MUST NOT 扣减 clinic AI 配额。

#### Scenario: VIP 触发者首次生成

- **WHEN** 某 `deviceNo` 当日缓存未命中，且触发者 `wxId` 经 device 契约判定为 VIP
- **THEN** 服务 MUST 以 DeepSeek 配置调用 Python 分析，并将结果写入该 `deviceNo` 当日缓存

#### Scenario: 非 VIP 触发者首次生成

- **WHEN** 某 `deviceNo` 当日缓存未命中，且触发者判定为非 VIP
- **THEN** 服务 MUST 以 Zhipu 配置调用 Python 分析并写入当日缓存

#### Scenario: 他看护命中缓存

- **WHEN** 同 `deviceNo` 同日已有缓存，另一 `wxId`（无论是否 VIP）再次 GET
- **THEN** 服务 MUST 直接返回缓存列表，MUST NOT 按后者 VIP 重新选模或重跑 LLM

### Requirement: VIP 查询失败降级为非 VIP

voice-service 在 care-alert 生成路径查询账号 VIP 时，若 device 契约超时、不可达、或返回不可用错误，MUST 将本次生成视为非 VIP（Zhipu），MUST NOT 因此使整个 daily 生成失败（设备校验与 Python 分析等其它错误仍按原语义失败）。实现 MUST 记录可观测 Warning。

#### Scenario: device VIP 接口失败

- **WHEN** 缓存未命中且查询 `vip-by-wx-id`（或等价契约）失败
- **THEN** 服务 MUST 选用 Zhipu 继续生成（若其它步骤成功），且 MUST 打出 Warning 级日志表明 VIP 降级

---

### Requirement: 系统仅在最近一笔为手工授且仍有效时允许撤销 VIP

cash-service MUST 提供 `POST /cash/admin/api/vip/entitlements/revoke`，鉴权 MUST 校验 `X-Admin-Password`。请求体 MUST 含 `wxId`（>0）。系统 MUST 仅在同时满足以下条件时撤销：

1. 该 `wxId` 的 `vip_entitlement.expire_at` 大于当前时间；
2. 该 `wxId` 最近一笔 `vip_order.status=paid` 的 `channel` 为 `admin`。

否则 MUST 拒绝，MUST NOT 修改权益或订单。

#### Scenario: 最近一笔为手工授且未过期

- **WHEN** 管理员对仍有效、且最近 paid 订单 `channel=admin` 的 `wxId` 提交撤销
- **THEN** 系统 MUST 将该 `expire_at` 设为当前时间，且该笔订单 `status` MUST 变为 `revoked`

#### Scenario: 最近一笔为付费渠道

- **WHEN** 最近 paid 订单 `channel` 为 `alipay` 或 `apple_iap`
- **THEN** 系统 MUST 拒绝撤销，MUST NOT 修改 `expire_at`

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝且 MUST NOT 写库

### Requirement: 撤销 VIP MUST 直接过期且 MUST NOT 按天数回退

撤销成功时系统 MUST 将 `vip_entitlement.expire_at` 设为不大于当前时间，使 `IsVip` 立即为 false。系统 MUST NOT 使用 `ShrinkEntitlement` 或按 `durationDays` 扣减。系统 MUST 使用订单状态 `revoked`，MUST NOT 把该订单标为 `refunded`。

#### Scenario: 付费剩余被一并清掉

- **WHEN** 账号先有付费剩余、随后手工续期，最近 paid 为 `admin`，管理员确认撤销
- **THEN** 系统 MUST 将到期设为当前时间，MUST NOT 只扣手工授的天数

#### Scenario: 重复撤销

- **WHEN** 目标 admin 订单已是 `revoked`，或权益已过期
- **THEN** 系统 MUST 返回错误，MUST NOT 再改写其它 paid 订单

### Requirement: 功能手工授撤销 MUST 按主体校验最近 admin 订单

cash-service MUST 提供 `POST /cash/admin/api/feature/grants/revoke`。请求 MUST 含 `featureId`，并按 `activation_subject` 要求 `wxId` 或 `deviceNo`。系统 MUST 找到该功能该主体最近一笔 `feature_order.status=paid`；仅当其 `channel=admin` 且对应权益仍有效时，MUST 将权益到期设为当前时间，并将该订单标为 `revoked`。`prediction_unlock` 与数量累加类 MUST NOT 提供此撤销。

#### Scenario: 撤销 user 主体功能

- **WHEN** 管理员对仍有效、最近 paid 为 `admin` 的账号维功能提交 `wxId` 与 `featureId`
- **THEN** 系统 MUST 使该账号该功能立即过期，并标记该 admin 订单 `revoked`

#### Scenario: 主体标识缺失

- **WHEN** 功能为 `device` 主体且请求没有 `deviceNo`
- **THEN** 系统 MUST 拒绝，MUST NOT 改权益

---

### Requirement: vip_order 与 feature_order SHALL 具备 grant_reason 列

`ai_voice_cash` 表 `vip_order` 与 `feature_order` MUST 各具备 `grant_reason` 列（字符串，建议 `VARCHAR(256) NOT NULL DEFAULT ''`）。支付渠道创建的订单 MUST 将该列置为空串。Admin 手工授予写入的订单 MUST 写入非空授权理由。cash-service 启动 `EnsureSchema`（或等价迁移）MUST 幂等补齐该列。

#### Scenario: 新库建表含列

- **WHEN** 空库执行 EnsureSchema
- **THEN** `vip_order` 与 `feature_order` MUST 包含 `grant_reason` 列

#### Scenario: 已有库补列

- **WHEN** 旧库缺少该列且服务启动 EnsureSchema
- **THEN** 系统 MUST 添加该列且 MUST NOT 因重复添加而 Fatal（幂等）

### Requirement: cash-service SHALL 提供 Admin 手工授 VIP API

cash-service MUST 提供 `POST /cash/admin/api/vip/entitlements/grant`，鉴权 MUST 校验网关注入的 `X-Admin-Password`。 请求体 MUST 含 `wxId`（>0）、`durationDays`（整数 ≥1）、`reason`（trim 后非空，长度不得超过列上限）。 成功时 MUST：

1. 订单插入与权益写入 MUST 在同一成功路径内完成（推荐同一事务）； 任一步失败 MUST 向调用方返回错误且 MUST NOT 声称授予成功。 外部支付宝/Apple 回调 MUST NOT 被设计为创建 `channel=admin` 订单的正常路径。

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

cash-service MUST 提供 `POST /cash/admin/api/feature/grants`（路径以实现为准，须落在 `/cash/admin/api/` 下），鉴权同 VIP Admin。 请求 MUST 含 `featureId`、非空 `reason`，并按该功能 `activation_subject`：

- `user`：MUST 要求 `wxId` >0，写入账号维权益（或该功能既定 Grant 路径）； - `device`：MUST 要求非空 `deviceNo`，写入设备维权益或预测数量权威。 `durationDays` 与 `grantQuantity` 按该功能 `grant_kind` / 产品语义使用：权益类 MUST 支持续期天数（与支付 Grant 一致； `durationDays=0` 表示永久仅当该功能既有 Grant 允许且请求显式传递——默认实现可要求权益类 `durationDays≥1`，预测 `allowed_count_delta` 则 MUST 要求.

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

---

### Requirement: 管理端 MUST 提供手填进账与出账

cash-service MUST 在与 `vip_order` 同库的新表中保存手填流水。每条 MUST 包含方向（`in` 进账或 `out` 出账）、非空名称、大于等于 1 的金额（分）和发生时间（unix 秒）。名称 MUST NOT 要求唯一。Admin MUST 能新增、按 id 修改上述字段、按 id 物理删除、并按发生时间倒序列出。方向、空名称或金额小于 1 MUST 拒绝且 MUST NOT 写库。未传发生时间时 MUST 使用写入时刻。

#### Scenario: 新增一笔出账

- **WHEN** 管理员提交方向 `out`、名称「模型费用」、金额 5000 分
- **THEN** 系统 MUST 写入该行，且列表中能按名称与金额读回

#### Scenario: 金额为 0

- **WHEN** 管理员提交金额 0
- **THEN** 系统 MUST 拒绝，且 MUST NOT 插入行

### Requirement: 收益合计 MUST 只统计实付并按功能名拆开

管理端收益合计 MUST 只把 `status=paid` 且 `channel` 为 `alipay` 或 `apple_iap` 的订单 `amount_fen` 算作付费。`channel=admin`、`refunded` 以及其他非 paid 状态 MUST NOT 计入。VIP 付费 MUST 为一个合计。功能付费 MUST 按功能拆开：展示名为功能标题，标题为空时 MUST 用功能编号；对不上功能商品的订单 MUST 单独成组并以商品编码为名称。合计 MUST 在读取时对订单求和，MUST NOT 把订单金额写入手填表。

#### Scenario: 功能付费按名称列出

- **WHEN** 两个不同功能各有一笔已支付的支付宝或 Apple 订单
- **THEN** 合计 MUST 返回两行，每行含该功能名称与对应金额之和

#### Scenario: 手工授与退款不计入

- **WHEN** 存在 `channel=admin` 的 paid 订单，以及 `status=refunded` 的订单
- **THEN** 这两类订单的金额 MUST NOT 出现在 VIP 或功能付费合计中

### Requirement: 差额 MUST 用手填与付费一起算出

在同一时间范围内，系统 MUST 计算差额 = VIP 付费 + 各功能付费之和 + 手填进账之和 − 手填出账之和。查询参数 `start`、`end`（unix 秒）均可省略。省略两者时 MUST 统计全部。给出 `start` 时订单 MUST 满足 `paid_at >= start`、手填 MUST 满足发生时间 `>= start`。给出 `end` 时对应时间 MUST `< end`。订单过滤与手填过滤 MUST 使用同一组参数。

#### Scenario: 对比成本与收益

- **WHEN** VIP 实付 1900 分、某功能实付 600 分、手填进账 0、手填出账 5000 分，且未传时间范围
- **THEN** 差额 MUST 为 1900 + 600 − 5000

#### Scenario: 时间范围同时作用于两边

- **WHEN** 管理员传入 `start` 与 `end`
- **THEN** 付费合计 MUST 只含该窗口内 `paid_at` 的实付，手填合计 MUST 只含该窗口内发生的流水

### Requirement: 开通功能管理 MUST 提供收益统计面板

`cash-feature-admin.html` 顶栏 MUST 在现有入口之外提供「收益统计」。选中后 MUST 只显示该面板，并展示 VIP 合计、按功能名的付费列表、手填进账合计、手填出账合计、差额，以及手填流水的增删改查。页面 MUST 说明付费金额是用户支付的标价，渠道抽成与运营成本须记为出账。金额输入 MUST 以分为单位。

#### Scenario: 打开收益统计

- **WHEN** 管理员在开通功能管理选中「收益统计」
- **THEN** 页面 MUST 展示上述合计与手填列表，且 MUST NOT 改动订单或权益数据

---

### Requirement: 启动种子 MUST NOT 覆盖已有 VIP 天数

cash-service `EnsureSchema` 在 `vip_product` 行已存在时 MUST NOT 更新 `duration_days`。首次插入该商品时 MUST 仍写入默认 30。`product_code` MUST 保持 `vip_monthly_19`。

#### Scenario: 重启不改运维天数

- **WHEN** `vip_product.duration_days` 已被改为 7，且 cash-service 再次执行 EnsureSchema
- **THEN** 该列 MUST 仍为 7

#### Scenario: 空库首次种子

- **WHEN** 不存在 `vip_monthly_19` 且执行 EnsureSchema
- **THEN** 系统 MUST 插入 `duration_days=30` 的该商品行

### Requirement: VIP 天数 MUST 大于等于 1

Admin 更新 VIP 商品时 `durationDays` MUST ≥1，否则 MUST 拒绝且 MUST NOT 写库。支付履约调用续期时，若商品天数 `<1`，系统 MUST 拒绝续期，MUST NOT 把天数替换为 30，MUST NOT 把 `expire_at` 写成 0。

#### Scenario: 后台保存 0 天

- **WHEN** 管理员将 VIP 有效天数提交为 0
- **THEN** 系统 MUST 拒绝，且库中原天数 MUST 不变

#### Scenario: 履约遇到非正天数

- **WHEN** 一笔 VIP 订单支付成功，但商品 `duration_days` 小于 1
- **THEN** 系统 MUST NOT 延长 `vip_entitlement`，MUST NOT 把该次续期当成 30 天

### Requirement: 功能授予天数 MUST 大于等于 1

功能 SKU 的 `duration_days`、功能定义的邀请授予天数与广告授予天数，在 Admin 保存时 MUST 均 ≥1。支付履约、邀请码兑换、广告开通 MUST NOT 在天数为 0 时写入永久权益（`expires_at=0`）。天数 `<1` 时 MUST 拒绝开通。

#### Scenario: 保存功能套餐天数为 0

- **WHEN** 管理员把某功能 SKU 的有效天数设为 0
- **THEN** 系统 MUST 拒绝保存

#### Scenario: 邀请天数配置为 0 时兑码

- **WHEN** 功能定义的邀请授予天数为 0，用户兑换邀请码
- **THEN** 系统 MUST 拒绝开通，MUST NOT 把权益 `expires_at` 写成 0

### Requirement: Hub MUST NOT 再把 0 标成永久

`cash-feature-admin.html` 中 VIP 有效天数、功能 SKU 有效天数、邀请/广告授予天数的输入 MUST 要求至少 1。页面 MUST NOT 再显示「0=永久」。

#### Scenario: 打开 VIP 天数输入

- **WHEN** 运维打开 VIP 套餐面板
- **THEN** 有效天数控件 MUST 不允许小于 1，且 MUST NOT 出现「0=永久」说明

### Requirement: 历史永久行 MUST NOT 被本变更批量改写

本变更 MUST NOT 执行把已有 `expires_at=0` 批量改为当前时间的迁移。已存在的此类权益行 MUST 仍按变更前的读规则判断是否有效。

#### Scenario: 上线时已有永久权益

- **WHEN** 某账号功能权益在本变更部署前 `expires_at=0`
- **THEN** 部署本身 MUST NOT 把该行改写为已过期

---

### Requirement: cash-service SHALL 提供分页只读 VIP 权益 Admin API

cash-service MUST 提供 `GET /cash/admin/api/vip/entitlements`，鉴权 MUST 校验网关注入的 `X-Admin-Password`（与 cash Admin 口令配置一致）。该接口 MUST 为只读，MUST NOT 修改 `vip_entitlement`、`vip_order` 或 `vip_product`。Query MUST 支持 `page`（默认 1）、`pageSize`（默认 20，最大 200），以及可选精确过滤 `wxId`。响应 MUST 含 `list`、`page`、`pageSize`、`total`。

#### Scenario: 合法管理员分页拉取

- **WHEN** 请求携带正确 `X-Admin-Password` 且未指定 `wxId`
- **THEN** cash-service MUST 返回分页后的权益列表与 `total`

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** cash-service MUST 拒绝该请求且 MUST NOT 返回权益数据

#### Scenario: 按 wxId 过滤

- **WHEN** 请求携带合法正整数 `wxId` 与正确口令
- **THEN** 响应 `list` MUST 仅包含该 `wxId` 的权益行（0 或 1 条）

### Requirement: 权益列表范围 SHALL 包含已过期 VIP

列表主数据源 MUST 为 `vip_entitlement` 的全部行，MUST NOT 默认过滤为仅 `expire_at > now`。每行 MUST 提供当前是否仍有效的布尔字段（如 `isVip`：当且仅当 `expire_at` 晚于当前时间），以及 `expireAt`（unix 秒）与可派生的剩余秒数（如 `remainingSeconds`，过期时可为 0 或负数）。

#### Scenario: 已过期权益仍出现在列表

- **WHEN** 某 `wx_id` 存在权益且 `expire_at` 已不晚于当前时间
- **THEN** 该行 MUST 仍可出现在列表中，且有效布尔字段 MUST 为 false

#### Scenario: 未过期权益标记有效

- **WHEN** 某权益 `expire_at` 晚于当前时间
- **THEN** 该行有效布尔字段 MUST 为 true

### Requirement: 激活金额 SHALL 取最近一次已支付订单金额

每条权益行的激活金额（如 `lastPaidAmountFen`）MUST 取该 `wx_id` 在 `vip_order` 中 `status=paid` 的订单，按 `paid_at DESC`（相等时可按 `id DESC`）的最近一条之 `amount_fen`。同条 MUST 附带该订单的 `channel` 与 `paidAt`（字段名以实现为准，语义一致；无 paid 订单时渠道与支付时间 MUST 为空或 0）。MUST NOT 使用未支付订单或商品现价冒充激活金额。若无任何 paid 订单，金额 MUST 为 0 或空语义，且 MUST NOT 伪造渠道支付时间。

#### Scenario: 多次支付取最近一次

- **WHEN** 同一 `wx_id` 存在多笔 `status=paid` 订单且 `paid_at` 不同
- **THEN** 激活金额与渠道、支付时间 MUST 对应 `paid_at` 最新的那一笔

#### Scenario: 无已支付订单

- **WHEN** 某权益行对应 `wx_id` 不存在 `status=paid` 的订单
- **THEN** 激活金额 MUST 不以未支付订单填充，响应 MUST 可区分「无实付」

### Requirement: gateway-app SHALL 反代并鉴权 cash Admin API

gateway-app-server MUST 将 `/cash/admin/api/*` 反代至 `CASH_SERVICE_URL` 指向的 cash-service。`IsGatewayAdminAPIPath` MUST 将 `/cash/admin/api/` 前缀视为须 Admin JWT 的管理 API。通过 Admin JWT 后，网关 MUST 向下游注入 `X-Admin-Password`（优先 `CASH_ADMIN_PASSWORD`，否则回退 `GATEWAY_APP_ADMIN_PASSWORD`）。浏览器客户端 MUST NOT 被要求自带 `X-Admin-Password`。device-service 及其它非 cash 进程 MUST NOT 直查 `ai_voice_cash` 以提供本列表。

#### Scenario: Hub 已登录调用 Admin API

- **WHEN** 运维持有效 Admin JWT 请求 `GET /cash/admin/api/vip/entitlements`
- **THEN** gateway-app MUST 反代至 cash-service 并注入口令头，cash-service MUST 可完成鉴权

#### Scenario: 无 Admin JWT

- **WHEN** 未携带有效 Admin JWT 请求 `/cash/admin/api/vip/entitlements`
- **THEN** gateway-app MUST 拒绝该请求（不得当作普通 App Bearer 用户接口放行）

### Requirement: cash Admin API SHALL 不计入 App usage 统计

`/cash/admin/api/*` MUST 视为运维管理通道，MUST NOT 按 App 对外接口计入 usage 统计策略；实现 MUST NOT 仅为该 Admin 路径修改 App 侧 `maintenance_skip` 假定。

#### Scenario: 管理端拉取列表

- **WHEN** 运维通过 Hub 调用 VIP 权益 Admin API
- **THEN** 系统 MUST NOT 将其记为需计入的 App 功能使用（与既有 voice/sim admin 管理通道一致）

---

### Requirement: catalog 聚合 totalActivatableCount

cash 合成功能目录时，对 `prediction_unlock` 项 MUST 调用 device **一级根**事件计数并写入响应字段 `totalActivatableCount`。永久合成 `allowedCount` MUST 仍为 defaultFree+permanentDelta（MUST NOT 因 VIP 或邀请改为 -1）。当计数契约失败时，系统 MUST NOT 返回误导性的正数天花板冒充成功（宜省略或 0），并记录日志。客户端预测锁槽位与「已全部激活」判断 MUST 与一级根集合对齐（本仓外 Flutter 同步）。

#### Scenario: catalog 含一级根天花板

- **WHEN** App 拉取 feature catalog 且 device 一级根计数成功返回 N
- **THEN** `prediction_unlock.totalActivatableCount` MUST 等于 N

#### Scenario: 无子一级根计入天花板

- **WHEN** 字典含无子一级根「洗澡」且 device 计数成功
- **THEN** 「洗澡」MUST 反映在 `totalActivatableCount` 中（即计入 N）

### Requirement: device 提供一级根事件计数

device 服务 MUST 提供可被 cash 调用的契约，返回事件字典中**一级根**事件数量（`parent_id = 0` 的事件行数，MUST 包含无子节点的一级根）。计数权威 MUST 为 device 事件字典，MUST NOT 要求 cash 直查 device 库表。该计数 MUST NOT 再等同于「至少有一个子事件的节点数」。

#### Scenario: 统计一级根含无子根

- **WHEN** 事件字典存在 8 个 `parent_id = 0` 的事件，其中部分无子节点
- **THEN** 返回的 count MUST 为 8

#### Scenario: 子事件不计入

- **WHEN** 某一级根下存在若干子事件
- **THEN** 这些子事件 MUST NOT 增加 count；仅该一级根本身计 1

---

### Requirement: 功能定义 MUST 声明开通主体 device 或 user

cash-service MUST 在 `feature_def` 上持久化 `activation_subject`，取值仅为 `device` 或 `user`，缺省 MUST 为 `device`。EnsureSchema MUST 幂等加列并对已有行回填 `device`。种子 `growth_trajectory_predict` MUST 为 `user`；`care_alert_smart_remind` 与 `prediction_unlock` MUST 为 `device`。

#### Scenario: 成长轨迹种子为人

- **WHEN** cash EnsureSchema 首次或幂等执行且写入/确认成长轨迹功能定义
- **THEN** 该功能 `activation_subject` MUST 为 `user`

#### Scenario: 值得留意仍为机

- **WHEN** 读取 `care_alert_smart_remind` 功能定义
- **THEN** `activation_subject` MUST 为 `device`

### Requirement: ActivateFeature MUST 按主体写入对应权益表

系统 MUST 支持 `ActivationSubjectDevice` 与 `ActivationSubjectUser`。当主体为 `device` 时 MUST 写入（或续期）`feature_entitlement`（按 `device_no`）。当主体为 `user` 时 MUST 写入（或续期）`feature_user_entitlement`（按 `wx_id`），MUST NOT 为该次履约写入成长轨迹的设备权益行。`prediction_unlock` 与 `allowed_count_delta` MUST 仅允许 device 主体。

#### Scenario: 用户主体支付履约

- **WHEN** 成长轨迹付费订单支付成功且功能定义为 `user`
- **THEN** 系统 MUST 以订单 `wx_id` 为键续期 `feature_user_entitlement`，时长取 SKU `duration_days`（种子 30）

#### Scenario: 拒绝预测写到账号维

- **WHEN** 调用 Activate 且 feature 为预测条数类且 SubjectType=user
- **THEN** 系统 MUST 拒绝且 MUST NOT 修改 `feature_allowed_count`

### Requirement: 支付邀请广告 MUST 按 activation_subject 选择 SubjectKey

支付履约、邀请码兑换、广告完成在授予非预测功能时 MUST 读取功能定义的 `activation_subject`：为 `user` 时 SubjectKey MUST 为触发账号 `wx_id`；为 `device` 时 MUST 为请求/订单中的 `device_no`。广告完成在主体为 `user` 时 MUST 要求有效 `wxId`，否则 MUST 拒绝。

#### Scenario: 邀请开通成长轨迹写人

- **WHEN** 用户兑码开通 `growth_trajectory_predict` 成功
- **THEN** 权益 MUST 落在兑换人 `wx_id` 的 user 权益表，且邀请天数取自 `feature_def.invite_duration_days`

### Requirement: 目录合成 MUST 按主体合并开通态

`GetFeatureCatalog`（或等价 App 目录）MUST 接受设备号，并在调用方提供登录 `wxId` 时：对 `activation_subject=device` 的功能用设备权益判断 `unlocked`；对 `user` 用该 `wxId` 的 user 权益判断。未提供有效 `wxId` 时，user 类功能 MUST 视为未解锁（`unlocked=false`），MUST NOT 误用设备权益冒充。

#### Scenario: 同设备另一账号目录未解锁

- **WHEN** 设备 D 上账号 A 已开通成长轨迹，账号 B（非 VIP）请求目录且绑定 D
- **THEN** 成长轨迹项对 B MUST 为未解锁

---

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

开通功能管理中的功能定义编辑 MUST 提供「邀请授予天数」表单字段并持久化到功能定义的邀请天数字段。天数 MUST ≥1（不支持永久邀请授予）。该字段 MUST NOT 被解释为支付套餐时长（支付仍以 SKU 为准）。Admin MUST NOT 再将广告授予天数作为可配置开通通道参数；保存时 MUST NOT 接受或持久化将 `ad` 纳入 `unlock_methods` 的配置。

#### Scenario: 运维修改邀请天数后兑码

- **WHEN** 管理员将某权益型功能的邀请天数从 7 改为 3 并保存，之后用户首次邀请开通该功能
- **THEN** 新授予 MUST 按 3 天计算到期

#### Scenario: 支付套餐独立

- **WHEN** 功能定义邀请天数为 7 且某支付 SKU `duration_days=30`
- **THEN** 支付履约 MUST 仍按 SKU 30 天授予，MUST NOT 被邀请天数覆盖

---

### Requirement: 本变更 MUST NOT 回退广告开通清理

功能开通域广告通道拒绝、Hub 无广告勾选、catalog 不依赖广告字段等行为 MUST 保持；本变更回退推送门禁与基础版时 MUST NOT 恢复 `channel=ad` 授予或 Hub 广告开通配置入口。

#### Scenario: Activate ad 仍拒绝

- **WHEN** 调用方以 Channel=`ad` 请求 ActivateFeature
- **THEN** 系统 MUST 仍返回错误且 MUST NOT 变更权益

### Requirement: 功能开通域 MUST 拒绝广告开通通道

系统 MUST NOT 再通过广告完成功能权益授予。`POST /cash/app/api/feature/ad/complete`（若仍注册）MUST 返回明确业务错误。`ActivateFeature` 在 `channel=ad` 时 MUST 拒绝。Admin 保存功能定义时，若 `unlockMethods` 含 `ad` MUST 拒绝。EnsureSchema MUST 将仍含 `ad` 的 care/growth/预测相关种子或强制回写路径剥离为 `payment,invite_code`（与现网 care/growth 一致并覆盖新预测功能）。

#### Scenario: App 广告 complete 失败

- **WHEN** 旧客户端调用功能广告 complete
- **THEN** 系统 MUST 失败且 MUST NOT 写入或续期任何功能权益

#### Scenario: Admin 勾选广告被拒

- **WHEN** 运维保存功能定义且 unlockMethods 含 ad
- **THEN** 系统 MUST 拒绝保存

### Requirement: Hub 与 App 目录 MUST 去掉广告开通展示依赖

`cash-feature-admin.html` MUST NOT 再提供「广告 ad」开通方式勾选，邀请天数文案 MUST NOT 再捆绑「广告授予天数」作为开通方式配置入口。App 合成 catalog 项 MUST NOT 再依赖 `adDurationDays` 作为开通能力展示（字段省略或恒为无意义值且文档标明废弃；推荐省略）。

#### Scenario: Hub 无广告勾选

- **WHEN** 运维打开开通功能管理编辑任一功能
- **THEN** 页面 MUST NOT 展示可勾选的广告开通方式

---

### Requirement: 客户端 MUST 从功能目录解析 logo 与 color

Flutter App MUST 在功能 catalog 模型（如 `FeatureCatalogItem`）中解析并保存每项的 `logo`（CDN URL 字符串）与 `color`（hex 字符串）。解析失败或缺字段时，`logo`/`color` MUST 视为空串，MUST NOT 导致目录整表解析失败。

#### Scenario: 解析含视觉字段的目录项

- **WHEN** catalog 返回某项含非空 `logo` 与 `color`
- **THEN** 客户端模型 MUST 暴露对应值供 UI 使用

### Requirement: 指定界面 MUST 在标题左侧展示功能 logo

下列界面在展示对应功能标题时，MUST 在标题左侧展示该功能 logo（网络图）；当 `logo` 为空时 MUST 使用本地占位，MUST NOT 留白导致标题错位：

1. 功能开通界面（各功能卡）
2. AI 分析界面（喂养分析入口与成长轨迹入口各自对应功能）
3. 功能开通弹窗
4. 喂养记录分析详情页
5. 成长轨迹详情页

#### Scenario: 开通中心卡片标题带 logo

- **WHEN** 用户打开功能开通界面且某功能已配置 logo
- **THEN** 该功能卡标题左侧 MUST 显示其 logo

#### Scenario: 无 logo 时占位

- **WHEN** 某功能 `logo` 为空
- **THEN** 上述界面 MUST 显示占位图/图标且标题仍可读

### Requirement: 指定界面 MUST 使用功能 color 作为玻璃拟态主色

上述界面中，对应功能所在的页面或对话框背景/玻璃面板 MUST 使用该功能 `color` 作为 accent（混入方式对齐喂养事件 sheet：`HistoryEditGlassPanel` / `AppModalGlassPanel` / `panelGlassGradient` 等既有 accent 混合）。AI 分析界面若同时展示喂养分析与成长轨迹入口，MUST 分别使用各自功能的 color，MUST NOT 整页强制单一功能色。

当 `color` 为空时，MUST 回退应用主题 primary（或等价），MUST NOT 崩溃。

#### Scenario: 开通弹窗随功能着色

- **WHEN** 用户打开某功能的开通弹窗且该功能 color 为非空 hex
- **THEN** 对话框玻璃拟态 MUST 呈现该主色倾向

#### Scenario: 成长轨迹详情随功能着色

- **WHEN** 用户进入成长轨迹详情页且 `growth_trajectory_predict` 的 color 非空
- **THEN** 页面/玻璃背景 MUST 使用该 color 作为 accent

#### Scenario: AI 分析双卡分色

- **WHEN** 用户打开 AI 分析界面
- **THEN** 喂养分析入口与成长轨迹入口 MUST 分别使用 care 与 growth 功能的 color（若已配置）

---

### Requirement: 系统 MUST 提供绑机可读的合成功能目录 API

cash-service MUST 提供 `GET /cash/app/api/feature/catalog`：须登录且 `X-Internal-Device-No` 非空； `device_no` MUST 仅取自该内部头，MUST NOT 采信 query/body 覆盖。 响应 MUST 一次返回功能目录项（无分页），**MUST 包含上架与下架（`status=0`）的功能定义项**（硬编码已废止的 `prediction_unlock` MUST 仍排除）。 每项 MUST 含稳定 `featureId`、展示标题、解锁方式提示、**`status`（0=下架，1=上架）**，以及按功能 `activation_subject` 解析的开通态：`unlocked`（bool）； 已开通项 MUST 含 `unlockMethod` 与 `expiresAt`（unix 秒； UCG 入场资格 MUST NOT 作为目录项出现。 该接口 MUST NOT 计入 App usage 统计。 该路径 MUST NOT 加入 Bearer.

#### Scenario: 一次返回完整目录且含开通态与 status

- **WHEN** 已登录且已绑机的客户端请求功能目录
- **THEN** 响应 MUST 包含功能项，每项含 `unlocked` 与 `status`，且 MUST NOT 要求 page/pageSize

#### Scenario: 未绑机拒绝

- **WHEN** 请求缺少有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝，MUST NOT 返回目录数据

#### Scenario: 下架功能仍返回并带 status=0

- **WHEN** 某功能在 Admin 被停用（status=0）且客户端请求 catalog
- **THEN** 目录 API MUST 仍返回该项，且该项 `status` MUST 为 0；客户端可自行决定如何展示

#### Scenario: UCG 不在目录中

- **WHEN** 客户端请求功能目录
- **THEN** 响应 MUST NOT 包含 UCG 入场资格项或用 `unlocked` 表示喂养资格

### Requirement: 过期权益 MUST NOT 视为已开通

当某权益 `expiresAt > 0` 且已早于当前时间，对应目录项 `unlocked` MUST 为 false（或等价不可用）。

#### Scenario: 过期后 unlocked 为 false

- **WHEN** 设备曾开通某功能但权益已过期
- **THEN** 目录中该项 `unlocked` MUST 为 false

### Requirement: 功能权益权威数据 MUST 落 MySQL 且按 device_no 维度

功能定义与设备权益 MUST 持久化在 `ai_voice_cash`。权益主体 MUST 为 `device_no`。其他服务 MUST NOT 直查该库。VIP 账号权益 MUST 使用独立 VIP 表，MUST NOT 因 VIP 购买写入功能权益行。

#### Scenario: 开通写入本域库

- **WHEN** 支付/邀请码/广告成功开通某功能
- **THEN** cash-service MUST 在 `ai_voice_cash` 写入或更新对应 `device_no` 权益行（或更新 `allowedCount`）

### Requirement: App MUST NOT 依赖独立 entitlements 列表接口作为主读模型

本变更 App 主读模型 MUST 为合成 catalog。系统 MUST NOT 将独立 `GET .../feature/entitlements` 作为客户端必调接口（可省略该 App 路径）。

#### Scenario: 客户端仅调 catalog 即可展示开通态

- **WHEN** 客户端已调用合成 catalog
- **THEN** 响应 MUST 足以按列表展示各功能是否已开通，无需再调独立权益列表

### Requirement: 本变更 MUST NOT 回退 catalog 下架 status 语义

`GET /cash/app/api/feature/catalog` MUST 继续返回含 `status` 的项（含下架 `status=0`），由前端决定展示；本变更 MUST NOT 恢复「停用功能从目录消失且无 status 字段」的旧行为。

#### Scenario: 下架项仍带 status=0

- **WHEN** 某功能 status=0 且客户端请求 catalog
- **THEN** 响应 MUST 仍含该项且 `status` 为 0

### Requirement: 下架功能 MUST 拒绝新开通

当功能定义 `status≠1` 时，点试用、创建功能订单、邀请兑换开通该功能 MUST 被拒绝。已存在的未过期权益在目录中 MUST 仍可反映 `unlocked=true`（若未过期），业务使用（含 imminent）MUST NOT 仅因下架而失效（停售不停用）。

#### Scenario: 下架后无法建单

- **WHEN** 功能已下架用户尝试创建该功能订单
- **THEN** 系统 MUST 拒绝建单

#### Scenario: 下架前已开通仍显示已解锁

- **WHEN** 用户在下架前已获未过期人权益且功能现为下架
- **THEN** catalog 该项 `status` MUST 为 0 且 `unlocked` MUST 为 true

---

### Requirement: App 合成功能目录项 MUST 嵌套可售 products 列表

`GET /cash/app/api/feature/catalog` 返回的每一功能项 MUST 包含 `products` 数组（可为空）。数组元素 MUST 为该 `featureId` 下状态启用的功能 SKU，且至少含：`productCode`、`priceFen`、`originalPriceFen`、`durationDays`、`grantKind`、`grantQuantity`；MAY 含 `appleProductId`。客户端 MUST 能仅凭本接口完成支付展示，并用所选 `productCode` 调用功能建单接口，MUST NOT 依赖硬编码商品码。系统 MUST NOT 为此新增独立的 App `GET /cash/app/api/feature/products` 作为必调接口。

#### Scenario: 启用 SKU 出现在对应功能项下

- **WHEN** Admin 为 `prediction_unlock` 配置了启用中的功能 SKU，且客户端请求合成 catalog
- **THEN** 该项 `products` MUST 含该 SKU 的 `productCode` 与 `priceFen` 等字段

#### Scenario: 无 payment 或无启用 SKU 时 products 为空

- **WHEN** 某功能无任何启用中的 `feature_product`
- **THEN** 该项 `products` MUST 为空数组（或等价空列表），MUST NOT 省略导致客户端解析失败（实现可用 `[]`）

#### Scenario: 停用 SKU 不下发

- **WHEN** 某 SKU `status` 非启用
- **THEN** catalog 的 `products` MUST NOT 包含该 SKU

#### Scenario: 不新增独立 App products 接口为必调

- **WHEN** 客户端需要展示功能价格并建单
- **THEN** 合成 catalog MUST 足够，MUST NOT 要求再调独立 App products 列表接口

---

### Requirement: 功能订单按交易号查空 MUST 不阻断履约

`FulfillFeaturePaid`（及小智 MCP 的 `FulfillXiaozhiMcpPaid`）在带有渠道交易号时 MUST 先查是否已有同渠道同交易号的已支付单。查无行时 MUST 视为尚未履约，MUST 继续按 `order_no` 加载订单并履约。MUST NOT 因该次空查询向调用方返回 `sql.ErrNoRows` 或导致支付宝回调失败。

当已存在同 `channel` 与 `channel_txn_id` 的 `paid` 功能/MCP 订单时：

- 若 `channel=apple_iap` 且已履约单的 `wx_id` 与本次目标订单 `wx_id` **不同**：MUST 拒绝履约，错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」，MUST NOT 将目标单标为 `paid`，MUST NOT 向目标 `wx` 授功能或 MCP 能力。
- 否则（同 `wx_id`，或非 `apple_iap`）：MUST 幂等成功且 MUST NOT 重复开通。

#### Scenario: 首次支付宝回调功能单

- **WHEN** 功能订单状态为 `created`，支付宝以新的 `trade_no` 回调，且库中尚无该 `channel_txn_id`
- **THEN** 系统 MUST 将该单置为 `paid` 并完成功能开通，MUST NOT 因「按交易号查无」失败

#### Scenario: 同交易号重复回调（同 wx 或非 Apple）

- **WHEN** 同一 `channel` 与 `channel_txn_id` 已对应一笔 `paid` 功能订单，且（`channel` 非 `apple_iap`，或目标订单 `wx_id` 与已履约单相同）
- **THEN** 系统 MUST 幂等成功且 MUST NOT 重复开通

#### Scenario: Apple 同交易号异 wx

- **WHEN** `channel=apple_iap` 且同一 `channel_txn_id` 已对应账号 A 的 `paid` 功能或小智 MCP 订单，账号 B 的订单对该交易号发起履约
- **THEN** 系统 MUST 拒绝，错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」，MUST NOT 为 B 开通

### Requirement: 功能订单与 SKU 读空 MUST 使用明确语义

按 `order_no` 查功能订单无行时，系统 MUST 返回明确的业务错误（订单不存在），MUST NOT 返回裸的 `sql.ErrNoRows`。按 `app_account_token` 查无行时 MUST 视为未找到（与 VIP 同语义）。按商品编码或 Apple 商品 ID 查启用 SKU 无行时，系统 MUST 返回明确的业务错误（商品不存在或未映射），MUST NOT 返回裸的 `sql.ErrNoRows`。

#### Scenario: 不存在的功能订单号

- **WHEN** 履约请求的 `order_no` 在 `feature_order` 中不存在
- **THEN** 错误信息 MUST 表明订单不存在，且 MUST NOT 仅为 `sql: no rows in result set`

---

### Requirement: 保存功能套餐 MUST NOT 新建商品编码

Admin 更新功能 SKU 时，商品编码去空白后为空，或 `feature_product` 中不存在该编码，系统 MUST 拒绝且 MUST NOT 插入新行。编码已存在时 MUST 只更新该行的价格、天数、Apple 商品 ID、上下架及其他非主键字段，MUST NOT 修改 `product_code`。

#### Scenario: 空编码保存

- **WHEN** 管理员提交的商品编码为空
- **THEN** 系统 MUST 拒绝，且 `feature_product` 行数 MUST 不变

#### Scenario: 已有编码改价

- **WHEN** 管理员提交已存在的商品编码，并把价格改为新的分值
- **THEN** 该编码的价格 MUST 更新，且 MUST NOT 出现新的商品编码

### Requirement: 打开功能时 MUST 填入已有套餐

开通功能管理选中某个功能时，售卖套餐表单 MUST 填入该功能已有套餐的商品编码。若存在上架且编码不以 `fp_` 开头的行，MUST 优先填入该行；否则填入上架行；再否则填入列表中的第一条。该功能没有任何套餐时，编码 MUST 保持为空，且页面 MUST NOT 生成新编码。页面 MUST NOT 再提供清空编码以新建套餐的操作。

#### Scenario: 打开已有种子套餐的功能

- **WHEN** 管理员打开一个已有非 `fp_` 上架套餐的功能
- **THEN** 表单商品编码 MUST 为该套餐编码

#### Scenario: 没有新建入口

- **WHEN** 管理员查看售卖套餐区域
- **THEN** 页面 MUST NOT 显示「新建套餐」

---

### Requirement: 功能支付履约 MUST 授予 device_no 权益或递增 allowedCount

cash-service MUST 支持独立功能 SKU（`feature_product`）：建单与支付宝/Apple IAP 验签 MUST 与 VIP 支付栈一致（服务端验签）。支付成功后 MUST 按商品配置向订单关联的主体授予功能权益或递增 `allowedCount`（主体按功能 `activation_subject`：device 或 user）。功能订单 MUST 使用独立 `feature_order` 表。支付宝异步通知、Apple ASN 与 Apple 客户端验单入口 MUST 与 VIP **共用分流**（按订单表区分），MUST NOT 污染 VIP 订单语义。`channel=apple_iap` 的功能建单 MUST 生成并返回 UUID `appAccountToken`，供 StoreKit 与 ASN 映射。

#### Scenario: 功能 SKU 支付成功写权益

- **WHEN** 某功能商品支付回调/验单成功且 `grant_kind=entitlement`
- **THEN** 系统 MUST 幂等为订单主体写入/续期对应 `feature_id` 权益，`unlockMethod` MUST 为 `payment`

#### Scenario: 功能 SKU 支付成功增加 allowedCount

- **WHEN** 某功能商品支付成功且 `grant_kind` 为预测数量增量
- **THEN** 系统 MUST 幂等将该主体的 `allowedCount` 增加配置的 `grant_quantity`

#### Scenario: 共用回调按订单分流

- **WHEN** 支付宝通知或 Apple ASN 到达且订单号/token 属于 `feature_order`
- **THEN** 系统 MUST 执行功能履约且 MUST NOT 续期 `vip_entitlement`

#### Scenario: 重复回调不重复授予

- **WHEN** 同一渠道交易号或已 paid 订单再次通知
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误时长或数量

#### Scenario: 功能 Apple 建单返回 appAccountToken

- **WHEN** 用户以 `channel=apple_iap` 创建功能订单
- **THEN** 响应 MUST 含 UUID `appAccountToken` 且落库可被 ASN 反查

### Requirement: VIP 购买 MUST NOT 写入功能权益

VIP 商品履约 MUST 仅续期 `vip_entitlement`（`wx_id`）。MUST NOT 为任何 `device_no` 写入 `feature_entitlement`，MUST NOT 修改 `feature_allowed_count`。既有 VIP status API MUST 保持可用，供客户端 V1 覆盖功能列表（不得覆盖 UCG 入场）。

#### Scenario: VIP 履约不影响功能表

- **WHEN** 用户成功购买或续期月 VIP
- **THEN** 系统 MUST 更新 VIP 到期时间，且 MUST NOT 新增或更新任何功能权益行 / allowedCount

### Requirement: 邀请码兑换 MUST 按单功能开通并遵守一家绑定与人维去重

cash-service MUST 提供 `POST /cash/app/api/feature/invite-codes/redeem`（计入 usage）。 请求 MUST 含邀请码与目标 `featureId`。 `wx_id` MUST > 0。 规则 MUST 包括：

- 一码可覆盖其配置的（或所有支持邀请码的）多个功能，但单次请求 MUST 只开通所请求的一个 `featureId`，MUST NOT 自动开通其余功能。 - 系统 MUST NOT 再执行「一家锁定」（redeemer 绑定单一 owner）； MUST NOT 因已兑过其他 owner 的码而拒绝新 owner（其它功能级 InviteOncePerUser 等约束除外）。 - 系统 MUST NOT 仅因 `redeemer_wx_id` 等于 `owner_wx_id` 而拒绝（已取消自用禁兑）。 - 系统 MUST 拒绝码主人当前绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换（同宝宝禁兑）。 - 成功后 MUST.

#### Scenario: 指定功能兑换成功

- **WHEN** 合法用户提交有效码与尚未码开的 `featureId` 且未违反同宝宝及其余人维规则
- **THEN** 系统 MUST 仅开通该功能（或增加对应数量），并占用相应兑换记录

#### Scenario: 同码再次开通另一功能

- **WHEN** 用户已用某码成功开通功能 A，再次提交同一码与功能 B（B 尚未码开且码支持 B）
- **THEN** 系统 MUST 允许开通 B，且 MUST NOT 因已开 A 而自动开其它功能

#### Scenario: 不同 owner 码在未触达功能级一次限制时可兑

- **WHEN** 用户已成功使用 owner=X 的码开通某功能后，再提交 owner=Y 的码开通**另一**仍允许邀请开通的功能，且未违反同宝宝与 InviteOncePerUser 等
- **THEN** 系统 MUST NOT 仅因「换一家」而拒绝

#### Scenario: 同宝宝拒兑

- **WHEN** 兑换者 `device_no` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 同人异机可兑自己的码

- **WHEN** 兑换人 `wx_id` 等于码的 `owner_wx_id`，且主人当前 `device_no` 与兑换者不同或主人未绑机，且其它校验通过
- **THEN** 系统 MUST NOT 因自用拒绝

#### Scenario: 同一功能不可再次码开（功能级一次时）

- **WHEN** 某 `wx_id` 已通过邀请码开通过某适用 InviteOncePerUser 的 `featureId`
- **THEN** 再次用任意邀请码开通同一 `featureId` MUST 被拒绝

### Requirement: ActivateFeature MUST 拒绝广告通道

统一开通入口在 `channel=ad` 时 MUST 拒绝并 MUST NOT 写入权益。允许的通道 MUST 包含 `payment`、`invite_code`、`trial`、`admin`（以实现为准）。

#### Scenario: ad 通道被拒

- **WHEN** 调用方以 Channel=`ad` 请求 ActivateFeature
- **THEN** 系统 MUST 返回错误且 MUST NOT 变更权益表

---

### Requirement: feature_def MUST 持久化 logo 与 color

`ai_voice_cash.feature_def` MUST 包含 `logo` 与 `color` 列。`logo` MUST 存储 OSS objectKey（建议前缀 `feature/`），MUST NOT 在库内持久化完整 CDN URL 作为权威值。`color` MUST 为 `#RGB` 或 `#RRGGBB`（大小写不敏感）或空串。cash-service 以外进程 MUST NOT 直查该表。

#### Scenario: 写入 objectKey 与 hex

- **WHEN** 管理员保存某功能的 logo 与主色
- **THEN** 系统 MUST 将 objectKey 与校验通过的 color 持久化到该 `feature_id` 行

#### Scenario: 非法 color 拒绝

- **WHEN** 提交的 color 不是空串且不符合 `#RGB`/`#RRGGBB`
- **THEN** 系统 MUST 拒绝保存并返回错误

### Requirement: 系统 MUST 为三个内置功能种子默认 color

EnsureSchema/种子 MUST 为下列功能写入默认 `color`（logo 默认为空串）：

| feature_id | color |
|------------|-------|
| `prediction_unlock` | `#3B82F6` |
| `care_alert_smart_remind` | `#0D9488` |
| `growth_trajectory_predict` | `#EA580C` |

当行已存在且 `color` 非空时，种子 MUST NOT 覆盖该 color。

#### Scenario: 新库种子带默认色

- **WHEN** 全新库执行功能定义种子
- **THEN** 上述三个 `feature_id` 的 `color` MUST 分别为表中默认值且 `logo` 为空

#### Scenario: 运维已改色不被种子覆盖

- **WHEN** 某功能 `color` 已被运维改为非空自定义值后再次跑种子
- **THEN** 该行 `color` MUST 保持运维值

### Requirement: App 功能目录 MUST 下发 logo 与 color

`GET /cash/app/api/feature/catalog` 每项 MUST 包含 `logo` 与 `color`。HTTP 边界上 `logo` MUST 为可访问的 CDN 绝对 URL；若库内 objectKey 为空，则 `logo` MUST 为空串。`color` MUST 原样下发库内值（种子后内置功能通常非空）。该扩展属于 v1 结构变更；本能力不要求保留无字段旧响应。

#### Scenario: 已配置视觉的目录项

- **WHEN** 某启用功能已配置 logo objectKey 与 color，客户端请求 catalog
- **THEN** 该项 MUST 含非空 CDN `logo` 与对应 `color`，且仍含既有 `unlocked` 等开通态字段

#### Scenario: 未上传 logo

- **WHEN** 某功能 `logo` 库内为空
- **THEN** catalog 该项 `logo` MUST 为空串，`color` 仍 MUST 返回（含种子默认色）

### Requirement: Admin 功能定义 API MUST 读写 logo 与 color

功能定义列表与更新 API MUST 支持 `logo` 与 `color`。列表/读出时 `logo` MUST 映射为 CDN URL（与 App 一致）。更新时客户端可提交 objectKey（上传接口返回值）与 color；未更换 logo 时 MUST 允许保留原 objectKey。

#### Scenario: 更新主色

- **WHEN** 管理员仅修改 color 并保存
- **THEN** 系统 MUST 更新 color，且 MUST NOT 清空已有 logo

### Requirement: 系统 MUST 支持功能 logo 上传至 OSS

系统 MUST 提供 Admin 可调用的功能 logo 上传能力（建议 `POST /cash/admin/api/feature/defs/logo` 或等价），经 ucg 契约写入 OSS，返回 objectKey（及可选 CDN URL）。objectKey MUST 使用 `feature/` 前缀（或与实现约定的等价 feature 命名空间）。实现 MUST 经 `clients/ucg`（或 platform 封装），MUST NOT 从 cash 业务包 import device 服务实现。

#### Scenario: 上传成功返回 objectKey

- **WHEN** 管理员上传合法 png/jpg/jpeg/webp logo
- **THEN** 系统 MUST 返回可写入 `feature_def.logo` 的 objectKey

#### Scenario: 超大或非法类型拒绝

- **WHEN** 文件超过大小上限或扩展名不支持
- **THEN** 系统 MUST 拒绝上传且 MUST NOT 写入 `feature_def`

---

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

---

### Requirement: VIP 权益以 ai_voice_cash 表为准

账号是否 VIP MUST 由 `ai_voice_cash` 中的权益记录判定，主键为 `wx.id`（`wx_id`）。MUST NOT 使用 `ai_voice_device.wx.is_vip` 或 `device_no` 作为 VIP 真相源。当且仅当权益 `expire_at` 晚于当前时间时，该账号 MUST 被视为 VIP。

#### Scenario: 无权益记录

- **WHEN** 查询某 `wxId` 且不存在权益行或 `expire_at` 已过期
- **THEN** 系统 MUST 判定 `isVip=false`

#### Scenario: 未过期权益

- **WHEN** 查询某 `wxId` 且权益 `expire_at` 大于当前时间
- **THEN** 系统 MUST 判定 `isVip=true`

### Requirement: 内部按 wxId 查询 VIP

cash-service MUST 提供经内部密钥鉴权的 HTTP 接口 `GET /cash/internal/api/vip/by-wx-id`（query `wxId`），响应至少包含 `wxId` 与 `isVip`（宜含 `expireAt`）。voice-service 等跨域调用方 MUST 经此契约读取 VIP，MUST NOT 直查 `ai_voice_cash`。

#### Scenario: 合法内部查询

- **WHEN** 调用方携带正确内部密钥请求正整数 `wxId`
- **THEN** cash-service MUST 返回该账号当前 VIP 布尔结果

#### Scenario: 缺少内部密钥

- **WHEN** 请求未携带正确内部密钥
- **THEN** cash-service MUST 拒绝该请求

#### Scenario: 无效 wxId

- **WHEN** `wxId` 缺失或 ≤0
- **THEN** cash-service MUST 返回参数错误，MUST NOT 视为 VIP

---

### Requirement: VIP 与永久单事件开通写路径分离

VIP 支付/续期履约 MUST 只更新 VIP 权益（含 `expireAt`），MUST NOT 修改 `feature_allowed_count` 或把 VIP 写成永久 feature entitlement 来「买断」预测条数。单事件支付/邀请/广告 MUST 只写入对应永久条数或 entitlement，MUST NOT 修改 VIP `expireAt`。

#### Scenario: 开通月卡

- **WHEN** 用户 VIP 履约成功
- **THEN** VIP 有效期延长且预测永久 `allowed_count` 不变

#### Scenario: 单次预测支付

- **WHEN** 用户支付预测 +1 成功
- **THEN** 永久条数 +1 且 VIP 状态不变

### Requirement: 读模型叠加使用权

对需开通的功能（含预测锁），业务使用态 MUST 在「永久已开通/条数内」与「当前 isVip」之间按 OR 放行。catalog 返回的预测 `allowedCount` MUST 反映永久合成值，MUST NOT 仅因 isVip 改为全开哨兵。UCG 入场资格 MUST 继续不受 isVip 影响。

#### Scenario: VIP 未买断条数

- **WHEN** 用户 isVip=true 且永久预测条数小于 totalActivatableCount
- **THEN** 使用态可查看预测事项，但 catalog `allowedCount` 仍为永久值

---

### Requirement: 一期单档月会员商品

cash-service MUST 提供一期唯一可售 VIP 商品：`product_code=vip_monthly_19`，标价 **1900 分人民币**，时长 **30 天**（续期按日叠加）。App MUST 能通过 `GET /cash/app/api/vip/product`（需登录，`X-Internal-Wx-Id>0`）读取该商品信息。

#### Scenario: 查询一期商品

- **WHEN** 已登录用户请求 VIP 商品接口
- **THEN** 响应 MUST 包含 `vip_monthly_19`、月费语义与 19 元标价（或等价分单位字段）

### Requirement: 创建 VIP 订单

cash-service MUST 提供 `POST /cash/app/api/vip/orders`：从 Header 读取 `X-Internal-Wx-Id`（>0），body 含 `productCode` 与 `channel`（`alipay` 或 `apple_iap`）。成功时 MUST 创建 `created` 状态订单并返回客户端调起支付所需参数（支付宝为调起串/参数；Apple 为 orderNo、预期 `appleProductId`，以及 UUID 格式的 `appAccountToken`）。金额 MUST 以服务端商品表为准。

#### Scenario: 支付宝建单

- **WHEN** 合法 `wxId` 以 `channel=alipay` 且 `productCode=vip_monthly_19` 建单
- **THEN** 系统 MUST 落库订单且状态为 `created`，并返回可用于调起支付宝的参数

#### Scenario: Apple IAP 建单

- **WHEN** 合法 `wxId` 以 `channel=apple_iap` 建单
- **THEN** 系统 MUST 落库订单且返回与 ASC 商品映射一致的 `appleProductId`（或等价字段）、订单号，以及非空 UUID `appAccountToken`

#### Scenario: 未登录建单

- **WHEN** 缺少有效 `X-Internal-Wx-Id`
- **THEN** 系统 MUST 拒绝建单

### Requirement: 支付宝异步通知开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/alipay/notify`：网关可匿名到达，但 MUST 校验支付宝签章与订单金额/商户订单号。校验通过后 MUST 将订单置为 `paid`（若尚未支付），并按商品时长续期该 `wx_id` 的 `vip_entitlement`（`new_expire = max(now, current_expire) + 30d`）。重复通知 MUST 幂等，MUST NOT 重复叠加错误时长。

#### Scenario: 首次合法 notify

- **WHEN** 支付宝对某 `created` 订单首次投递合法成功通知
- **THEN** 订单 MUST 变为 `paid`，且对应账号 VIP `expire_at` MUST 相对支付前至少延长 30 天

#### Scenario: 重复 notify

- **WHEN** 同一渠道交易对已 `paid` 订单再次通知
- **THEN** 系统 MUST 返回成功应答语义且 MUST NOT 再次延长权益（或仅保持幂等结果）

#### Scenario: 验签失败

- **WHEN** notify 签章非法
- **THEN** 系统 MUST NOT 将订单标为 paid，MUST NOT 开通权益

### Requirement: Apple IAP 验单开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/apple/verify`（需登录）：接受 App 提交的 IAP 交易凭证（JWS 或团队选定的等价载荷）。 该路径 MAY 用于支付后即时开通（加速 UX），但 **Apple 扣款后的权威开通来源 MUST 为 App Store Server Notifications**（见能力 `apple-server-notifications`）。 verify 与 ASN 履约 MUST 共用幂等键（`channel=apple_iap` 与 `channel_txn_id=transactionId`），MUST NOT 因双路径重复叠加权益。 校验通过且 `productId` 映射正确时，MUST 幂等将关联订单置 `paid`（或绑定已有订单）并续期 entitlement。 本路径 MUST 使用 **Apple IAP**，MUST NOT 使用 Apple Pay（PassKit）作为数字会员开通方式。 当同一.

#### Scenario: 合法 IAP 验单

- **WHEN** 已登录用户提交可验证的 IAP 交易且商品映射正确
- **THEN** 系统 MUST 开通或续期该 `wxId` 的月会员权益

#### Scenario: 重复提交同一 transaction（同 wx）

- **WHEN** 同一 Apple `transactionId`（或等价渠道交易号）被**同一** `wx_id` 再次 verify
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误的时长

#### Scenario: ASN 已履约后再 verify（同 wx）

- **WHEN** 同一 `transactionId` 已由 ASN 履约为同一 `wx_id`，用户再次调用 verify
- **THEN** 系统 MUST 幂等成功且 MUST NOT 再次延长权益

#### Scenario: 异 wx 复用同一 transaction

- **WHEN** 某 `transactionId` 已为账号 A 履约成功，账号 B 登录后对该 `transactionId` 调用 verify
- **THEN** 系统 MUST 拒绝，错误信息 MUST 为「一个 Apple 账号仅支持开通一个胖宝账号」，MUST NOT 为 B 开通或续期 VIP

### Requirement: 查询当前 VIP 状态

cash-service MUST 提供 `GET /cash/app/api/vip/status`（需登录），返回当前 Header `wxId` 的 `isVip` 与 `expireAt`（无权益时 `isVip=false`）。

#### Scenario: 会员中查询状态

- **WHEN** 权益未过期的用户请求 status
- **THEN** 响应 MUST 含 `isVip=true` 与未来的 `expireAt`

---

### Requirement: VIP 商品承载现价与原价

`ai_voice_cash.vip_product` MUST 提供现价字段 `price_fen` 与原价字段 `original_price_fen`（单位：分）。`original_price_fen=0` 表示无划线原价。改价 MUST 通过更新该表完成；写路径 MUST 为开通功能管理所调用的 VIP 商品 Admin API（及等价手工 SQL）。系统 MUST NOT 使用环境变量或 yaml 作为定价或 Apple 商品 ID 的真相源。

#### Scenario: 原价为零

- **WHEN** 商品行 `original_price_fen` 为 0
- **THEN** 读价 API MUST 返回 `originalPriceFen=0`（客户端可不展示划线价）

#### Scenario: 原价与现价均可读

- **WHEN** 商品行 `price_fen=1900` 且 `original_price_fen=9900`
- **THEN** 读价 API MUST 分别返回对应分值

#### Scenario: Admin 改价后读价一致

- **WHEN** 管理员经 VIP 商品 Admin API 将 `price_fen` 更新为新值
- **THEN** `GET /cash/app/api/vip/product` 与后续建单 `amount_fen` MUST 使用该新值

### Requirement: 匿名读取 VIP 商品价格

cash-service MUST 提供 `GET /cash/app/api/vip/product`，响应至少包含 `productCode`、`priceFen`、`originalPriceFen`、`durationDays`。该接口 MUST 允许未携带 Bearer / 未注入有效 `X-Internal-Wx-Id` 的调用方成功读取。gateway-app MUST 将该路径列入 Bearer 鉴权白名单（精确 GET）。支付宝等建单路径 MUST 仍要求登录，且建单金额 MUST 使用库内 `price_fen`。

#### Scenario: 未登录拉价成功

- **WHEN** 客户端不带 Authorization 请求 `GET /cash/app/api/vip/product`
- **THEN** gateway MUST 放行且 cash-service MUST 返回上架商品的现价与原价

#### Scenario: 建单仍用库现价

- **WHEN** 已登录用户创建支付宝 VIP 订单
- **THEN** 订单 `amount_fen` MUST 等于当时 `vip_product.price_fen`，MUST NOT 使用请求体中的任意标价覆盖

---

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

---

### Requirement: wx 表承载账号 VIP 标志

device-service 所属库 `ai_voice_device` 的 `wx` 表 MUST 提供账号 VIP 列 `is_vip`（`TINYINT NOT NULL DEFAULT 0`：`0` 表示非 VIP，`1` 表示 VIP）。该列以 `wx.id` 为账号主键语义，MUST NOT 以 `device_no` 作为 VIP 归属键。

#### Scenario: 默认非 VIP

- **WHEN** 新建或历史迁移后的 `wx` 行未显式设置 VIP
- **THEN** `is_vip` MUST 为 `0`（非 VIP）

#### Scenario: VIP 归属账号而非设备

- **WHEN** 同一 `device_no` 关联的账号上下文需要判断 VIP
- **THEN** 系统 MUST 按该请求对应的 `wx.id` 行上的 `is_vip` 判定，MUST NOT 用设备号反推「设备级 VIP」

### Requirement: device 内部按 wxId 查询 VIP

device-service MUST 提供经内部密钥鉴权的 HTTP 接口，按正整数 `wxId` 返回该账号是否 VIP。voice-service 及其他跨域调用方 MUST 经此契约读取，MUST NOT 在非 device 进程内直查 `wx` 表。

#### Scenario: 存在账号且为 VIP

- **WHEN** 内部调用方以合法密钥请求 `wxId` 对应行且 `is_vip=1`
- **THEN** 响应 MUST 表示 `isVip=true`，并包含该 `wxId`

#### Scenario: 存在账号且非 VIP 或行不存在

- **WHEN** 内部调用方请求的 `wxId` 对应行 `is_vip=0`，或该主键不存在
- **THEN** 响应 MUST 表示 `isVip=false`（或等价：调用方视为非 VIP），MUST NOT 要求调用方因此中断无关主业务（由调用方定义降级）

#### Scenario: 缺少内部密钥

- **WHEN** 请求未携带正确的 device 内部密钥
- **THEN** device-service MUST 拒绝该请求

---
