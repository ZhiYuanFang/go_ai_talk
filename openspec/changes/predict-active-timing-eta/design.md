## Context

Flutter 上报进行中 ETA；Python intent `speak_history` 读取。须与 `PredictImminentPending`（下次发生 + MQ 推送）分离。

## Goals / Non-Goals

**Goals:**

- App PUT 整表替换 Redis；internal GET 按 deviceNo 返回。
- 字段透传；无业务重算（含不过滤「已过期」ETA）。

**Non-Goals:**

- MQ/推送；MySQL 权威；改 imminent 闸。

## Decisions

### D1：键、锁与 API（沿用 Imminent 模式）

- Redis 权威列表：`cachekit.PredictActiveEtaKey(deviceNo)`，**无 TTL**，靠客户端全量替换。
- 同步短锁：`cachekit.PredictActiveEtaSyncLockKey(deviceNo)`，`SetNX` + 短 TTL（秒级，对齐 imminent）；未抢到锁 → 业务 4xx「同步繁忙，请稍后重试」。
- App：`PUT /device/api/predict/active-eta/pending`（voice 宿主 + gateway `/device/api/predict/*`，登录 + 绑机）。
- Internal：`GET /voice/internal/api/predict/active-eta?deviceNo=`（供 Python；鉴权按现有 voice internal 约定）。
- Redis 策略：沿用 `PredictImminentPending` 同族模式（非新建读缓存收益率评估）；访问经 `cachekit.WithObserver` / `Default()`。

### D2：请求/响应 JSON

**App PUT**

```json
{ "deviceNo": "...", "items": [
  { "rootEventId": 1, "leafEventId": 2, "historyId": "h1",
    "title": "喂奶", "startAt": 1710000000, "expectEndAt": 1710003900 }
]}
```

响应：`{ "count": N }`（写入保留条数；清空成功为 0）。

**Internal GET** 查询参数 `deviceNo`；响应：`{ "items": [ /* 同上字段 */ ] }`；缺键或空 → `items: []`。

### D3：条目保留与去重

单条字段：`rootEventId`、`leafEventId`（int64）、`historyId`（string，omitempty；空串当无）、`title`（可空 trim）、`startAt`、`expectEndAt`（unix 秒）。

**保留条件（全部满足）：**

- `rootEventId > 0` **且** `leafEventId > 0`
- `startAt > 0` 且 `expectEndAt > 0`
- `expectEndAt >= startAt`（否则丢弃）

**不作为丢弃条件：**

- `title` 为空 → 仍保留
- `expectEndAt < now`（已过期）→ 仍保留（透传，由 Python/history 判断）

**同表去重：** 按 `leafEventId`；同 leaf 保留列表中靠后的（后写覆盖）。`historyId` 仅透传，不参与去重。

非法项静默丢弃，不导致整次同步失败。

### D4：无副作用

同步成功只写 Redis；**不**发延时消息、**不**调推送。

### D5：gateway-app 与 usage

- 反代：已有 `/device/api/predict/*`，**无需**扩展绑定。
- Auth exempt：**不**加入（须 Bearer 登录）。
- Usage：负责人确认 **不统计** `PUT /device/api/predict/active-eta/pending` → `maintenance_skip.go` 精确登记（对标 imminent）。

## Risks / Trade-offs

- [客户端停同步] → Redis 可能陈旧；Python 降级无 ETA；后续靠空列表或覆盖纠正。
- [与 history 不一致] → GET 原样返回；Python 以 history 进行中为准，ETA 仅增强文案。
- [已过期仍存] → 列表可能含历史 ETA；换取「服务端不重算」简单性。

## Migration Plan

上线即可；旧 Redis 无键视为空列表。回滚停路由后键可残留至下次清空。

## Open Questions

（无；2026-10-11 探索已敲定：usage 不统计、短锁、保留/去重/JSON。）
