## 1. 存储与 App API

- [x] 1.1 `cachekit` 登记 `PredictActiveEtaKey`（无 TTL）与 `PredictActiveEtaSyncLockKey`（短锁 TTL 由业务 SetNXEX 传入）
- [x] 1.2 App PUT：`api/v1` 结构体（`deviceNo`+`items` → `count`）+ voice controller/service：绑机校验、SetNX 短锁、按 D3 清洗/去重、整表写 Redis、空列表 Del；**不** PublishDelayed、**不**推送
- [x] 1.3 gateway-app 自检：反代已覆盖 `/device/api/predict/*`；不进 auth exempt；`PUT /device/api/predict/active-eta/pending` 写入 `maintenance_skip.go`（负责人确认不统计）
- [x] 1.4 Redis 策略：沿用 `PredictImminentPending` 模式（权威/无 TTL/全量替换），经 `cachekit` Observer

## 2. 内部读

- [x] 2.1 `GET /voice/internal/api/predict/active-eta?deviceNo=` 返回 `{ items }`
- [x] 2.2 缺键/空 → 空列表；字段与 App items 对齐

## 3. 验收

- [x] 3.1 PUT 两条再 PUT 覆盖 → Redis 仅第二次；同 leaf 去重后写赢
- [x] 3.2 PUT 空 → GET `items: []`；非法项丢弃不影响合法项
- [ ] 3.3 与 Flutter/Python 联调字段名（`items` / id / 时间戳）一次
