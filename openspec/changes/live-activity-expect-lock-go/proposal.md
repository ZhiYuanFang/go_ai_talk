## Why

兄弟仓 Flutter change `live-activity-expect-lock` 将在锁屏 Live Activity 上展示结构化预计：Timing 副行「预计 HH:mm 结束」、Imminent 右侧短版 duration/amount。客户端只解码与拼文案；Go 须在 start/update ContentState 写入可选字段，否则真机无字。Imminent pending / active-eta Redis **已有**这些数据，缺口在 LA 扇出未透传，以及 active-eta 同步目前 **禁止** 推送。另：若 LA 标题继续塞中文「预计这次约…」后缀会与右侧结构化短文案叠字；history 改计时的 Timing update 整包替换 ContentState 时若不带 `expectEndAt` 会冲掉副行。

## What Changes

- **Imminent LA start**：从开火载荷把可选 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit` 写入 ContentState（有值才写；不写成品中文 label；不拼 leaf）。
- **Imminent LA 标题**：Attributes / ContentState / LA Alert 的 title 使用基句 **`{昵称}要{事件名}了`**，MUST NOT 含「，预计这次约…」后缀；普通 `PushByBizType` 通知正文可继续带 expect 中文后缀（Android / 非 LA 通道）。
- **Timing LA**：ContentState 可选 `expectEndAt`（unix 秒）。
- **active-eta 同步后更新 Timing**：`PUT .../active-eta/pending` 写 Redis 成功后，对含有效 `historyId` + 正 `expectEndAt` 的条目扇出 Timing **update**（activityId=`timing:{deviceNo}:{historyId}`）；空列表清空 Redis 时不强制清 LA 字段（计时 stop/end 仍走既有 Timing end）。
- **history 改计时 update**：`updateTimingLiveActivity` MUST 在可解析到同 `historyId` 的 active-eta 时把 `expectEndAt` 一并写回 ContentState，避免 APNs 整包替换冲掉副行。
- **兼容**：缺字段旧客户端忽略；新客户端缺字段不画。MUST NOT 推 logo/color。MUST NOT 在 Go 拼锁屏短文案（副行/右侧由客户端拼）。
- **非目标**：不改 Flutter UI；不改 Imminent 1/min / 清场；不改 Android 通知长文案拼装逻辑（仅与 LA 标题拆开）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-imminent-live-activity-push`：Imminent start ContentState MUST 透传可选结构化 expect；LA title MUST 为基句不含预计后缀。
- `predict-timing-live-activity-push`：Timing ContentState MUST 支持可选 `expectEndAt`；history 侧 update MUST 尽量保留已有 ETA。
- `predict-active-timing-eta`：同步成功后 MAY/MUST 触发 Timing LA update（修正原「MUST NOT 推送」边界，仅限 LA ContentState update，仍 MUST NOT 发延时 MQ / 普通可见推送）。

## Impact

- `internal/services/voice/predict_live_activity.go`：`startImminentLiveActivityForWx` 扩参并写 ContentState expect*
- `internal/services/voice/predict_imminent.go`：开火拆 `laTitle`（基句）与 `alert`（可含后缀）；LA 用基句 + expect*，普通推送用长 alert
- `internal/services/voice/predict_active_eta.go`：Sync 成功后扇出 Timing update（经 `clients/push`）
- `internal/services/history/live_activity_timing.go`：`updateTimingLiveActivity` 读 active-eta Redis（`cachekit.PredictActiveEtaKey`）合并 `expectEndAt`；禁止 import voice 包
- Twin：`flutter_ai_talk` / `live-activity-expect-lock`
- 无新 HTTP path / 无新表
