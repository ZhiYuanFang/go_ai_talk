## Why

联调可见：Timing 闭合 / Imminent pending 清场后锁屏条仍挂着，且时间错乱。根因是 APNs `event=end` 未带过去的 `dismissal-date`（代码把 0 当「立即」却省略该字段，Apple 则视为可在锁屏再挂数小时），同时 end 的 ContentState 过瘦把 `startTime`/`nextAt` 解成 0。删除未闭合 history 也不发 Timing end。token 已上报，不必改 token 分层。

## What Changes

- **立即 dismiss**：所有 Live Activity **end**（Imminent 清场/换槽、Timing 闭合/删除）组 APNs 体时 MUST 写入 `dismissal-date` 为 **过去的 unix 秒**；`DismissalDate==0` MUST 视为立即并填过去时刻，禁止省略该键。
- **end ContentState 不抹计时字段**：Timing end MUST 保留可读的 `startTime`/`eventName`（外加 `phase`）；Imminent end MUST 保留 `nextAt`/`title`（外加 `phase`），避免 dismiss 延迟时锁屏解成 0 导致时间错乱。
- **删除 history 发 Timing end**：`DeleteDeviceHistory` 成功后 MUST 按 `timing:{deviceNo}:{historyId}` 向绑定用户扇出 end（未闭合必发；已闭合再删 MAY 再发一次幂等 end）。
- **锁屏标题走固定 Attributes**：Push-to-Start 常不套用 start 的 ContentState，导致锁屏落到「事件计时中 / 事件即将发生」。Timing start Attributes MUST 含 `eventName`（空则按 eventId 主档名回填）与 `startTime`；Imminent start Attributes MUST 含 `title`。ContentState 仍带同名字段。兄弟仓 Flutter 锁屏 MUST 在 state 为空时回退读 attributes。
- **非目标**：不改 `push_to_start`/`activity` 分工、1/min、pending 清场挂点、token 注册。不在 register 后再发一轮 update 作为标题主路径。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-timing-live-activity-push`：闭合与删除均 MUST 发可立即 dismiss 的 Timing end；end 载荷不得把 `startTime` 抹成缺省 0；start 固定 Attributes MUST 含展示名与 `startTime`。
- `predict-imminent-live-activity-push`：pending 清场与单槽换新的 Imminent end MUST 带过去的 `dismissal-date`；end 载荷不得把 `nextAt` 抹成缺省 0；start 固定 Attributes MUST 含 `title`。

> 上述 capability 来自未归档 change `predict-live-activity-push`（及 `live-activity-lock-ui-payload` 增量）；本变更对其做 end/dismiss 增量。v3.0.3 基线尚无这两章。

## Impact

- `internal/services/push/push_live_activity_apns.go`：end 必写过去 `dismissal-date`
- `internal/services/history/live_activity_timing.go`、`local.go`：闭合/删除 end；start Attributes 补 `eventName`/`startTime`；空名走主档
- `internal/services/voice/predict_live_activity.go`：Imminent end 补全；start Attributes 补 `title`
- 兄弟仓 `flutter_ai_talk`：Attributes 解码新字段；锁屏标题/计时 state 空则回退 attributes
- 无新 HTTP path / 无新表 / 无 Android 要求
