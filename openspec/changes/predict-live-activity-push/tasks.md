## 1. Token 与 liveactivity 发送

- [x] 1.1 新增 App LA token 注册/更新 HTTP path（并列于 `/app/api/push/register`，不改 device register 语义）；push 域分表存储 Push-to-Start 与 per-activity update token；MUST NOT 写入 `push_device.uk_token`；cachekit builder 跟踪活动态键
- [x] 1.2 新 path：核对 gateway-app 反代、Bearer 白名单；**usage 是否计入先问负责人**后再改 `maintenance_skip`（若需要）
- [x] 1.3 push/APNs 支持 liveactivity start/update/end：`apns-push-type=liveactivity`，topic=`{bundleId}.push-type.liveactivity`；复用现有 `.p8`（`PUSH_APNS_*`）
- [x] 1.4 iOS Imminent MUST NOT 再发普通 `predict_imminent` alert；无 LA token 跳过该接收方

## 2. Imminent 限流、清场、单槽

- [x] 2.1 每用户（wxId）60s SETNX 闸接入开火扇出前；Timing 不起用
- [x] 2.2 `SyncPredictImminentPending`：写 Redis/重布闹钟前向绑定用户 end 全部 Imminent；不得 end Timing（每次成功处理都清）
- [x] 2.3 开火 start 前 end 该 wx 旧 Imminent 槽（或同槽更新）；未过 1/min 不 start
- [x] 2.4 Redis 跟踪 Imminent activity id：**每 wx 单槽**（多监护人各一条）

## 3. Timing history 钩子

- [x] 3.1 未闭合计时 create → Timing LA start（activity id 含 historyId）
- [x] 3.2 闭合 endTime → 同 id end
- [x] 3.3 **单次落库成功触发**：审计所有写入口汇入 history 持久化成功路径一次扇出；MUST NOT voice adapter + history local 双推
- [x] 3.4 遵守包边界：经 clients/契约，禁止 services 互引；多 Timing 并存

## 4. 联调与开关

- [x] 4.1 特性开关：默认可关；与 `flutter_ai_talk` change `predict-ios-live-activity` 载荷/深链/Attributes 对齐联调
- [ ] 4.2 对照规格验收：1/min、pending 清场、per-wx 槽、iOS LA、Timing 开停无延迟闸、Android 无锁屏要求
