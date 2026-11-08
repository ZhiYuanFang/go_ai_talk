## 1. Token 契约与 push 扇出

- [x] 1.1 调整 `push_live_activity_token` schema/ensure：去掉（或停用）`attributes_kind`、`activity_id` 分槽；唯一键收敛为 `(wx_id, device_key, token_kind)`（或设计等价）
- [x] 1.2 更新 `RegisterLiveActivityToken` 与 `api/v1` register body：不再要求 `attributesKind`/`activityId`；校验仅 `push_to_start|activity`
- [x] 1.3 实现单一 APNs `attributes-type`（常量或单一 env）；删除 imminent/timing 双类型分支
- [x] 1.4 `FanoutLiveActivity`：按 wx 是否存在 `token_kind=activity` 选择 update vs start；空业务列表由调用方发 end；end 后仍清 activity 行、保留 push_to_start
- [x] 1.5 确认 gateway-app register 免鉴权 path 与 usage 策略不受破坏（字段变更不新增 path 则不改 `maintenance_skip`；若负责人另有要求再问）

## 2. history Refresh 公共方法

- [x] 2.1 在 history-service 实现 `RefreshBabyLiveActivity(deviceNo)`：读 pending（cachekit）+ 未闭合 history；组装 `items`（预测 `nextAt` 升序，进行中 `startTime` 倒序）；禁止 logo/color
- [x] 2.2 空 items → 对各绑定 wx end；非空 → 有 activity token 则 update，否则 start
- [x] 2.3 暴露 history internal HTTP（或等价）供 voice 经 `clients/history` 调用；补客户端封装与中文注释
- [x] 2.4 （可选）Refresh 合并 active-eta 的 expectEndAt 到对应进行中项，替代独立 Timing update 扇出

## 3. 接线触发点

- [x] 3.1 history CUD（Add/Update/Delete/闭合旁路）：保留 Del imminent pending，**之后**调用 Refresh；删除 `clearImminent` 的 Imminent end 清屏与 Timing 多槽 start/update/end 钩子
- [x] 3.2 `SyncPredictImminentPending` 写 Redis 成功后调用 history Refresh（空列表清空后同样 Refresh）
- [x] 3.3 `handlePredictImminentFire`：保留普通 `PushByBizType` 与既有闸门；去掉 Imminent end+start；闸门通过后调用 Refresh
- [x] 3.4 审计 voice/history 无双发：remote 写 history 不得在 voice 再钩一套 LA

## 4. 清理与联调

- [x] 4.1 删除或掏空仅服务双轨的辅助（如独立 Timing 槽 Redis 键写入、Imminent 槽换新 end）并更新注释
- [x] 4.2 与 Flutter 统一 Attributes / `items[]` 契约对齐（类型名写入代码常量）；开关 `PUSH_LIVE_ACTIVITY_ENABLED` 下联调 CUD / PUT / 到点三条路径
- [x] 4.3 跑 `hack/check-service-import` 与 Redis 旁路检查；确认无 services 互引、无业务层直连 `g.Redis()`
