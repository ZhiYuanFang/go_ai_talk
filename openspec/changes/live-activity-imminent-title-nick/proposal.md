## Why

统一 Live Activity 锁屏预测行当前把 Redis pending 的裸事件名原样写入 ContentState `title`，锁屏只显示「吃奶」一类短名；产品期望与普通可见推送同形的成品句 `{nick}要{事件}了`。量/时预计已由结构化 `expect*` 字段交给 Flutter 右侧短槽展示，不得塞进标题。LA 触发面（pending PUT + MQ 到点）宽于普通推送（仅 MQ），标题拼装 MUST 落在 history Refresh，禁止挂在 voice 普通推送路径上透传。

## What Changes

- **Imminent `title` 成品句**：history `RefreshBabyLiveActivity` / `buildLiveActivityItems` 对 `kind=imminent` 项，用设备画像昵称（空则「宝宝」）与 pending 事件名拼成 `{nick}要{事件}了` 写入 ContentState `title`。
- **独立于普通推送**：不改 voice MQ `predict_imminent` alert 拼装路径；不把普通推送 alert 回填进 LA。
- **expect 不进标题**：继续经 `applyPendingExpect` 写 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit`；标题 MUST NOT 追加「，预计这次约…」或 leaf。
- **Redis pending**：仍存客户端上报的事件名（及既有 expect）；投影时再拼成品句。
- **alertBody**：若取自首条 imminent `title`，随成品句更新（有意）。
- **非目标**：不改 timing 行 `eventName`；不改 Flutter 解码（锁屏已原样展示 `title` + 右槽 expect）；不改 Island（Island 用本地目录短名，不吃推送 title）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：统一 Refresh 投影时，imminent 项 `title` MUST 为 `{nick}要{事件}了` 成品句；MUST NOT 把 expect 量/时文案拼进 title；拼装 MUST 仅在 history LA 投影路径完成。

## Impact

- **history-service**：`internal/services/history/live_activity_refresh.go`（投影拼 title；经既有 `GetBirthday`/profile 契约取 `babyName`）。
- **voice-service / 普通推送**：无行为变更（alert 仍为既有 `要…了` + 可选 expect 后缀）。
- **push-service / gateway-app**：无契约/鉴权/usage 变更。
- **兄弟仓 Flutter**：锁屏已消费成品 `title` + 结构化 expect 右槽；本仓可不改客户端代码；建议 twin 文档注明 title 为成品句（可选，非本仓交付）。
- **兼容**：旧 App 仍上报裸事件名进 pending；服务端投影即生效，无需客户端同发。
