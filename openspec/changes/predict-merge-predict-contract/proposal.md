## Why

Flutter 按事件字典 `mergePredict` 调整预测维度与 pending/active-eta 上报形：非合并子事件需可选父事件名供 Python 意图反推；合并根进行中 ETA 可无预计结束；预测卡「上一次」需批量最近发生接口。现契约强制正 `expectEndAt`、无父名字段，且推送文案仍拼接 `expectLeafTitle`，阻碍三仓同发。

## What Changes

- **imminent pending**：items 增加可选 `parentEventName`（字典根事件 `name`；缺省兼容旧客户端）；`expectLeafTitle` **全链路丢弃**（不入库 Redis、不进 MQ、alert 不拼 leaf；旧客户端仍传则静默忽略）。
- **active-eta**：`expectEndAt` 可空；**省略 / null / 0** 均表示无预计结束；正值兼容旧客户端。本变更 **不** 给 active-eta 增加 `parentEventName`。
- **批量最近发生**：新增 App `GET /device/history/api/event/latest-by-event`（纯 history）：按 `deviceNo` 每 `eventId` 一条最新，`occurredAt=start_time`，含未闭合；**不含** `appointment_next`。须登录+绑机（Bearer **不**豁免）；**不计入** usage（负责人已确认，写入 `maintenance_skip.go`）。
- **非目标**：不实现 Flutter 预测算法 / Python 意图匹配细节；不改 `mergePredict` 字典列（见 `event-merge-predict`）。

## Capabilities

### New Capabilities

- `predict-latest-occurrence-batch`：纯 history 批量最近发生查询契约。
- `predict-pending-parent-optional`：imminent 可选 `parentEventName`；丢弃 `expectLeafTitle`；active-eta 结束可空（含 0）。

### Modified Capabilities

- 推送文案：amount 后缀不再拼接 `expectLeafTitle`（覆盖既有 `predict-expect-amount-duration` / 推送 copy 中 leaf 段；见本 change 内 pending 规格）。
- 既有 imminent pending / active-eta 写入与 voice internal GET 读出形（字段可选与忽略规则）。

## Impact

- **voice-service**：pending DTO/Redis/MQ/alert；active-eta 校验放宽；internal GET 投影 `parentEventName`。
- **history-service**：新 latest-by-event 路由与实现。
- **gateway-app**：history 反代前缀确认；`maintenance_skip` 精确登记新 GET；**不**进 auth exempt。
- **对端**：Flutter `predict-merge-predict-dimension`、Python `predict-merge-predict-intent` 须同步字段名 `parentEventName`（本仓仅 Go）。
- **兼容**：旧 App 无父名、仍带 leafTitle（被忽略）、active 必有正结束 → MUST 仍成功；推送 amount 文案对旧客户端也会少拼 leaf（有意行为变更）。
