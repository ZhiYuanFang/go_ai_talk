## Why

现网 Live Activity 拆成 Imminent / Timing 两套 Attributes，且 history CUD 对 Imminent 走「end 清屏」、预测到点走「先 end 再 start」。锁屏无法同时呈现「进行中 + 预测」，Flutter 也无法用列表态依次展示。需要收成**每用户一条**实时活动：ContentState 为列表，CUD / pending 同步 / MQ 到点都经同一 Refresh 投影；普通 `predict_imminent` 可见推送与延时 MQ **保持不变**。

## What Changes

- **统一单条 LA**：**BREAKING（LA 模型）** 取消 `attributesKind=imminent|timing` 双轨与多槽 Timing；每 wx 至多一条 Live Activity，ContentState 为 `items[]`（预测 + 进行中）。
- **公共 Refresh**：实现于 **history-service**；每次推送前重读 Redis imminent pending + 未闭合 history，排序后扇出。空列表 → `end` 摘屏；非空 → 有 `activity` token 则 `update`，否则 `start`（push_to_start）。
- **触发点**：
  - history 任意 CUD 成功：先 **Del** 该宝宝 imminent pending（保留现网作废语义，保证旧 MQ 对账失败、普通推送不受影响），再 Refresh；
  - `PUT …/imminent/pending` 写 Redis 成功后：Refresh；
  - MQ 预测到点（闸门通过后）：普通 `PushByBizType` **保留**，并额外 Refresh（不再 Imminent end+start）。
- **Token 表简化**：**BREAKING（register 契约）** `push_live_activity_token` 去掉业务分槽所需的 `activity_id`、`attributes_kind`；register 不再要求这两字段；APNs `attributes-type` 改为单一固定类型（或单一 env）。
- **移除**：CUD 路径 Imminent 清屏 end、开火前单槽换新 end、按 historyId 的 Timing start/update/end 多槽扇出。
- **非目标**：不改 delayed MQ 调度模型 / 对账闸门 / 可见推送文案；不改 pending PUT 整表替换；不强制 Flutter UI 实现细节（仅契约对齐兄弟仓）；不引入 Redis TTL 或扫表替代 MQ。

本变更 **supersede** 未归档双轨/清屏向变更意图（如 `predict-live-activity-push`、`history-clear-imminent-live-activity`、`update-timing-live-activity-on-edit`、`fix-live-activity-end-dismiss` 中与双 Attributes / CUD 清屏 LA 冲突的部分）；实现以本 change 规格为准。

## Capabilities

### New Capabilities

- `unified-live-activity-push`：统一列表态 Live Activity 的组装、排序、Refresh 触发、空列表 end、有无 activity token 的 update/start 选择，以及 LA token/register 契约收敛。

### Modified Capabilities

- `predict-imminent-schedule`：明确 history CUD **MUST 继续 Del** imminent pending（服务普通推送作废）；pending PUT 与 MQ 到点在保留既有调度/对账的前提下 **MUST** 触发统一 LA Refresh；CUD **MUST NOT** 再以 Imminent end 清屏替代 Refresh。

## Impact

- `internal/services/history/`：用 Refresh 替换 `clearImminentOnHistoryMutation` 的 LA end 侧与 Timing 多槽钩子；保留 Del pending；新增 internal 可被 voice 调用的 Refresh 入口（或等价 `clients/history`）。
- `internal/services/voice/`：pending PUT 成功后调 Refresh；`handlePredictImminentFire` 去掉 Imminent LA start/换槽，改调 Refresh；普通推送路径保留。
- `internal/services/push/`：Fanout 按 wx 有无 activity token 选 update/start；token 表/register API 去掉 `attributesKind`/`activityId`；单一 attributes-type。
- `api/v1/app_push_http.go`、gateway-app Bearer 豁免 path 若受 register 字段变更影响须同步；**兄弟仓 Flutter** Attributes/ContentState 列表契约须同发。
- 无新 App 业务 path（除非 Refresh 暴露为 history internal）；usage 统计：无新对外 App 接口则不改 `maintenance_skip`（若新增须先问负责人）。
