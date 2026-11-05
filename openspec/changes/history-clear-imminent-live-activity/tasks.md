## 1. Push：end 后清理 activity token（Imminent + Timing 共用）

- [x] 1.1 `dispatchLiveActivity`（或等价）在 `action=end` 扇出尝试结束后，按 `activityId` 删除全部 `token_kind=activity` 行；**不分** `attributesKind`（imminent / timing 均适用）；MUST NOT 删 `push_to_start`
- [x] 1.2 中文注释说明：计时闭合/删除与 Imminent 清屏/换槽共用，避免结束后无效 update token 累计；与 APNs invalid 删行兼容

## 2. History：落库成功清 Imminent + Del pending

- [x] 2.1 在 history 包新增 `clearImminentOnHistoryMutation(ctx, deviceNo)`：开关门控、列绑定 wx、Imminent end 扇出、Del per-wx slot、Del `PredictImminentPendingKey`；禁止 import voice；经 clients/device + clients/push + cachekit
- [x] 2.2 `AddDeviceHistory` 成功后调用（含瞬时 `endTime!=0`）
- [x] 2.3 `UpdateDeviceHistory` 成功后调用（一律清，不区分字段）
- [x] 2.4 `DeleteDeviceHistory` 成功后调用
- [x] 2.5 审计其它落库成功旁路（如 EnsureEnd / 闭合未闭合）是否汇入上述三函数；漏则同钩补齐，禁止 voice adapter 双钩

## 3. Voice：pending PUT 去掉清屏

- [x] 3.1 `SyncPredictImminentPending` 移除 `clearImminentLiveActivitiesForDevice` 调用；保留写 Redis / 空列表 Del / 延时 MQ
- [x] 3.2 保留开火 `startImminent…` 前单槽 `endImminentSlotForWx`；注释标明全员清屏已改挂 history

## 4. 联调验收

- [ ] 4.1 `PUSH_LIVE_ACTIVITY_ENABLED=true`：Add / Update（含改备注）/ Delete 后 Imminent 立刻消失；若有进行中 Timing，Imminent 清屏 MUST NOT 误 end 该 Timing
- [ ] 4.2 清屏后 Redis imminent pending 键为空；旧延时消息消费不因旧列表再推
- [ ] 4.3 Imminent end 后该 activityId 的 activity token 行已删；push_to_start 仍在
- [ ] 4.4 Timing 闭合或删除未闭合行后：锁屏 Timing 摘条，且该 `timing:…` activityId 的 activity token 行已删；push_to_start 仍在；未闭合仅 update 时 token 行仍在
- [ ] 4.5 仅 `PUT …/pending`（无 history 写）不再单独触发全员清屏；重新 PUT 仍可写入新日程
