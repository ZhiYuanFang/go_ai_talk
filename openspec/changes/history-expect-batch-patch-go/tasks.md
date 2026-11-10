## 1. API 与写路径

- [x] 1.1 定义 `expect-batch` 请求/响应 DTO（`deviceNo`、`items`；响应仅 `{ok}`；无 results）
- [x] 1.2 改为逐条 expect_* patch（WHERE id+deviceNo）；omit 不改列；失败条跳过不回滚其它；v1 不做 null 清列
- [x] 1.3 空 items → 400；`len(items)>200` → 400；≥1 条成功 → `{ok:true}`，全失败 → `{ok:false}`（合法 batch HTTP 200）

## 2. 通知、冻结与网关

- [x] 2.1 每条成功 reload 全行后更新 cache 并 WS `update`（含 expect*）（实现随 1.2 保持：仅成功条通知）
- [x] 2.2 确认 `historyUpdateData` / event/batch update 仍不写 expect_*（本 change 不修 EventBatch 通知剥字段）
- [x] 2.3 注册 `POST /device/history/api/event/expect-batch`；contracts 增加 `HistoryEventExpectBatchPath`；gateway 反代确认 fuzzy 已覆盖
- [x] 2.4 `usagestats/maintenance_skip.go` 精确排除 `POST /device/history/api/event/expect-batch`（负责人确认不计入 usage）

## 3. 联调

- [ ] 3.1 curl/手工：空 items 400；混入外机 id 时本机仍写入且 `{ok:true}`；全失败 `{ok:false}`；omit 语义；WS 仅成功条带 expect*
- [ ] 3.2 与 Flutter `history-expect-anchor-chain-scheduler` 联调一批连锁补丁
