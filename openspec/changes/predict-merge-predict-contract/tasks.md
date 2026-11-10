## 1. Pending / active-eta 契约

- [x] 1.1 imminent pending：DTO/Redis/GET/internal 增加可选 `parentEventName`；移除对 `expectLeafTitle` 的绑定与透传；amount alert 不拼 leaf
- [x] 1.2 active-eta：`expectEndAt` 省略/null/0 视为无结束且保留项；正值仍校验 `>= startAt`；LA 投影容忍无结束（不写 0）

## 2. 批量最近发生

- [x] 2.1 history-service：实现 `GET /device/history/api/event/latest-by-event`（纯 history；每 eventId 按 `start_time` 最大一条；含未闭合）
- [x] 2.2 gateway：确认 history 反代覆盖；Bearer **不**豁免；`maintenance_skip.go` 精确登记该 GET（不计入 usage）

## 3. 兼容与联调说明

- [x] 3.1 核对旧客户端：无 parentName、仍带 leafTitle、active 正结束均可成功；推送 amount 不再含 leaf
- [x] 3.2 与 Flutter / Python 锁定字段名 `parentEventName`（仅 pending；active-eta 不带）
