## Context

App 已通过 `PUT /device/api/predict/imminent/pending` 与 `PUT /device/api/predict/active-eta/pending` 将预测数据写入 Redis（voice 权威）；Python 经 internal GET 读取。Hub「设备数据运维」`history.html` 尚无只读展示。运维需在单设备页核对上报值（含进行中 `expectEndAt`）。

约束：浏览器不得持有 internal secret；Hub 已用 Admin JWT + `AdminCommon.adminFetch`，gateway-app 对 `/voice/admin/api/*` 注入 `X-Admin-Password`。

## Goals / Non-Goals

**Goals:**

- voice Admin 只读 GET，按 `deviceNo` 返回与 App/internal 对齐的完整字段。
- `history.html` 一张卡两段：临近待办 + 进行中 ETA；刷新；空列表明示。
- 复用 `LoadPredictImminentPending` / `LoadPredictActiveEta`；缺键/空/坏 JSON → 空列表。

**Non-Goals:**

- 清空或编辑 Redis；改 App PUT / MQ / 推送；新建 Redis 键；history/device BFF；新环境变量；usage 统计变更。

## Decisions

### D1：Admin 双 GET（不对齐成聚合 BFF）

新增：

- `GET /voice/admin/api/predict/imminent/pending?deviceNo=` → `{ "events": [...] }`
- `GET /voice/admin/api/predict/active-eta?deviceNo=` → `{ "items": [...] }`

条目字段分别复用 `PredictImminentPendingEventItem` / `PredictActiveEtaItem`（与 App PUT、internal GET 同形）。

**理由：** 与已有 internal 契约对称；无聚合 DTO；卡片可用 `Promise.all` 一次刷新。  
**备选：** 单接口 `{events,items}` —— 少一跳，但多一层聚合、与 internal 分叉，运维收益不大，否决。

### D2：鉴权与反代

- Controller 内 `VerifyVoiceAdminPassword`（同 rule-gaps / intent-vectors）。
- 浏览器只带 Admin JWT；gateway 注入口令。
- `/voice/admin/api/*` 已反代，**无需**新 proxy 绑定；路径不进 auth exempt（须已登录 Hub）。

### D3：UI 一张卡两段

位置：`history.html` 中「宝宝档案」与「历史事件」之间（先看上报状态再看历史表）。

| 段 | 数据源 | 列（完整字段） |
|----|--------|----------------|
| 临近待办 | imminent Admin GET | `eventId`、`nextAt`、`title`、`expectKind`、`expectDurationSec`、`expectAmount`、`expectUnit`、`expectLeafTitle` |
| 进行中 ETA | active-eta Admin GET | `rootEventId`、`leafEventId`、`historyId`、`title`、`startAt`、`expectEndAt` |

- 时间列：本地可读时间为主；可用 `title`/副文保留 unix 秒便于核对。
- 空数组：段内提示「暂无上报（Redis 无键或空列表）」，不当错误。
- 卡级「刷新」同时拉两段；进入页（Admin 已登录）自动拉一次。
- **不**提供写操作按钮。

### D4：服务边界与部署

- Load 仅在 voice-service；禁止他域直读 Redis 键。
- 部署：**voice-service**（新路由）+ **gateway-app**（`history.html`）。无新 env。

## Risks / Trade-offs

- [Redis 陈旧] → 卡片原样展示；提示数据来自客户端最后一次全量同步。
- [已过期 ETA 仍显示] → 与 active-eta 透传一致；UI 可不标过期（首版不做过滤/着色，可选后续）。
- [仅更 voice 未更网关静态] → 页面无卡；部署清单写明两者。

## Migration Plan

1. 上线 voice Admin GET。
2. 上线 gateway-app 静态 `history.html`。
3. 回滚：撤路由/撤卡片即可；Redis 数据不受影响。

## Open Questions

- （无；探索结论已定：一张卡两段、只读、字段全展示、新 change。）
