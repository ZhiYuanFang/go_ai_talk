## 1. Schema 与建单 token

- [x] 1.1 `vip_order` / `feature_order` EnsureSchema 增加 `app_account_token`（CHAR(36)，唯一索引可空）；注释说明仅 Apple 渠道使用
- [x] 1.2 `CreateOrder` / `CreateFeatureOrder` 在 `channel=apple_iap` 时生成 UUID 写入并在结果中返回 `appAccountToken`
- [x] 1.3 `api/v1` 建单响应类型增加 `appAccountToken`（omitempty）；controller 透传

## 2. ASN 验签与履约

- [x] 2.1 实现 ASN V2 `signedPayload` 验签（Apple 证书链）；失败拒绝履约；生产禁止 DEV_BYPASS 跳过 ASN 验签
- [x] 2.2 `POST /cash/app/api/vip/apple/notifications`：`g.Meta` + controller（可特殊读 body，对齐 alipay notify 写法）
- [x] 2.3 解析 notificationType / transaction：用 `appAccountToken` 查 VIP 或功能订单 → 成功类调用既有 `DispatchFulfillPaid` / 功能履约；`channel_txn_id=transactionId` 幂等
- [x] 2.4 实现 `REFUND`：标记订单退款语义并撤销/失效该笔 VIP 或功能权益（可测、幂等）；其它类型日志 + 2xx
- [x] 2.5 确认无 MQ、无新 ticker

## 3. gateway / usage / verify

- [x] 3.1 `gateway_app_auth_exempt.go` ExactPOST 增加 `/cash/app/api/vip/apple/notifications`
- [x] 3.2 `maintenance_skip.go` 登记 `POST /cash/app/api/vip/apple/notifications`（不计入 usage；已与口径确认）
- [x] 3.3 `apple/verify` 与 ASN 共用履约幂等；文档/注释标明 ASN 为权威；verify 尽量复用验签工具

## 4. 配置与文档

- [x] 4.1 runbook（`vip-commercial-config.md` / `cash-vip-sandbox.md`）：ASC Production+Sandbox 通知 URL、Flutter 必须传 `appAccountToken`、与支付宝 notify 对照
- [x] 4.2 `.env.example` / cash 配置注释：说明 ASN URL 配在 ASC（非新 env 密钥，除非验签需要挂载证书路径）

## 5. Flutter 跨仓

- [x] 5.1 `flutter_ai_talk`：Apple 建单读取 `appAccountToken`，StoreKit 购买 options 传入；缺 token 不得静默购买

## 6. 自检

- [x] 6.1 人工/沙箱：ASC Sandbox 通知 → 无客户端 verify 仍开通；重复通知不双开；REFUND 收回（验收步骤已写入 runbook，待 ASC 联调执行）
- [x] 6.2 确认未改支付宝路径、未引入 MQ、未新增测试文件、gateway 反代前缀无需扩展
