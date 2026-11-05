## Why

Imminent Live Activity「清屏」挂在 `PUT …/predict/imminent/pending` 上：历史已增删改、预测尚未重报时，锁屏「即将发生」仍挂着；预测重报又会在无关时刻清掉刚亮的条。真正让临近预测失效的是 **history 权威写路径**，清屏应跟历史走，并同时作废 Redis 待办。另：Imminent / Timing 在 `event=end` 后 `push_live_activity_token` 的 activity 行常残留，需一并清理，避免累计无效 update token。

## What Changes

- **清屏挂点搬走**：**BREAKING（清屏触发语义）**：从 `SyncPredictImminentPending` **移除** Imminent 全员清场；改为 history 落库成功路径（`Add` / `Update` / `Delete`，含瞬时事件 Add、任意成功 Update）对绑定用户 **一律** end 全部 Imminent Live Activity；MUST NOT 误 end Timing。
- **一并作废日程**：清屏同时 MUST `Del` 该 `deviceNo` 的 Redis imminent pending 权威键；旧延时 MQ 继续靠消费对账作废（无 pending 匹配则不推）。
- **Imminent / Timing 共用：end 后清 DB token**：任意 Live Activity `action=end`（Imminent 清屏/换槽，**以及** Timing 闭合/删除 end）扇出尝试结束后，MUST 删除该 `activityId` 的 `token_kind=activity` 行；MUST NOT 删除 `push_to_start`。Timing 起停触发条件不变，仅补结束侧落库清理。
- **非目标**：不改 Timing start/update 触发与载荷；不因 Imminent 清屏去 end Timing；不改 Imminent 开火 1/min 与单槽换新（start 前 end 旧槽保留）；不改 pending PUT 的整表替换/延时投递语义（仅去掉清屏副作用）；不强制清空 active-eta Redis；不改 Flutter 必改契约。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-imminent-live-activity-push`：清屏触发从 pending PUT 改为 history 增/删/改成功；清屏 MUST 清空 Redis pending；Imminent end 后 MUST 删除该 activity 的 activity token 行。
- `predict-imminent-schedule`：history 写成功 MUST 作废该宝宝 imminent pending Redis（与 LA 清屏同钩）；pending PUT 不再承担清屏。
- `predict-timing-live-activity-push`：Timing Live Activity end（闭合/删除）后 MUST 同样删除该 `activityId` 的 `token_kind=activity` 行，避免计时结束后无效 token 累计。

> capability 来自未归档 `predict-live-activity-push` / `fix-live-activity-end-dismiss` 等增量；v3.0.3 基线尚无完整 LA 章。本变更对其清场挂点、pending 作废与 end 后 token 清理做增量。

## Impact

- `internal/services/voice/predict_imminent.go`：去掉 SyncPending 内 `clearImminent…`
- `internal/services/history/`：落库成功钩子统一清 Imminent + Del pending Redis（经 cachekit + clients/push / device，禁止 import voice）；既有 Timing end 路径受益于 push 层 token 清理
- `internal/services/push/`：`action=end` 后按 activityId 删除 activity token 行（Imminent 与 Timing 共用）
- 无新 App HTTP path；history 直清 Redis pending（与既有 Timing 槽键同模式）
- Twin Flutter：无强制改；history 变更后重报 pending 仍可，但清屏不再依赖该 PUT
