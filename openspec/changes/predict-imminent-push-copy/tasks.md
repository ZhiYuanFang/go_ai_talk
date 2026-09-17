## 1. 文案拼装

- [x] 1.1 在 `handlePredictImminentFire`：`title` trim 为空则打 skip 日志并 return nil（不调用 `PushByBizType`）
- [x] 1.2 推送前经 `DeviceProfile().GetProfile(deviceNo)` 取 `babyName`；空或失败回落「宝宝」
- [x] 1.3 将 `alert` 改为 `{昵称}要{title}了`；移除「宝宝提醒：」前缀与「事件将在约 N 分钟内发生」作为事件名的回落

## 2. 文档对齐

- [x] 2.1 在本变更或 `predict-imminent-offline-push/CLIENT.md` 注明：空 `title` 服务端不发可见推送，客户端同步时应带父事件展示名

## 3. 自检

- [x] 3.1 确认未改 pending API 结构、未改 push-service 标题「胖宝」、未引入事件 logo / 叶子展开逻辑
- [x] 3.2 确认空 title 路径仍 Ack、无 Nack requeue / 无二次 Publish
