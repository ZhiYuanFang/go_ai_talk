## Context

`history-expect-fields-go` 已使 add 可落库 expect*，且 `historyUpdateData` **永不**写 `expect_*`。Flutter twin 改为 CRUD 后静默锚点链，需一次批量只改 expect。

探索迭代拍板：空 items→400、上限 200、响应仅 `{ok}`（无 results）、不计入 usage、contracts 登记 path；**不**修 EventBatch 通知层。  
**回摆（相对首版整批 TX）**：单条失败不得拖死整批——≥1 条成功即 `ok:true`；失败条静默跳过。

## Goals / Non-Goals

**Goals:**

- 窄口 `expect-batch`：按 id 批量写 expect 列；鉴权与 device 归属校验；成功行全量 WS/cache。
- 部分成功：能写的写上；有一条成功就算 ok。
- 普通 update/batch CRUD 冻结保持。
- usage 排除；contracts path 登记。

**Non-Goals:**

- 服务端不跑 interval 预测。
- v1 不做 expect 清空（JSON `null` → 清列）API。
- 不强制改 admin/internal。
- 不修 `EventBatch` op=update 裸实体导致的 cache/WS 短暂缺 expect*。
- 不返回逐条错误明细。

## Decisions

1. **新路由** `POST /device/history/api/event/expect-batch`  
   - 与现网 device history 同鉴权（Bearer + bound deviceNo）。  
   - 不塞进 `event/batch` 的 op 枚举。  
   - gateway fuzzy `/device/history/api/*` 已覆盖。  
   - contracts：`HistoryEventExpectBatchPath()`。

2. **请求**  
   ```json
   { "deviceNo": "...", "items": [
     { "id": 1, "expectStartTime": 1730000000, "expectEndTime": 1730001800, "expectAmount": 120.5 }
   ]}
   ```  
   - 缺键不更新该列；禁止用 `0` 表示缺省。  
   - **空 `items` → 400**；**`len(items) > 200` → 400**。  
   - `id<=0`：该条跳过（计失败），不因此 400 整请求。

3. **写路径（部分成功）**  
   - `historyExpectPatchData(item)` 只含出现的 expect_*。  
   - **逐条** `WHERE id=? AND device_no=?` Update（每条可用短事务或自动提交）；**禁止**外层「任一条失败则 ROLLBACK 全部」。  
   - 0 行 / 写错 / `id<=0` / 无列可写且归属失败 → **跳过**，继续下一条。  
   - 成功条：立即或批末 `Get` 全行 → cache + WS（仅成功条）。

4. **响应**  
   - 合法请求（非空、≤200、deviceNo 非空）→ **HTTP 200**。  
   - `{ "ok": true }` 当且仅当本批 **成功更新条数 ≥ 1**。  
   - `{ "ok": false }` 当全部未写上（仍 200，无 results）。  
   - 入参非法（空 items、超上限、空 deviceNo）→ **400**。

5. **通知**  
   - 仅成功条：reload 全行再 `patchHistoryOnUpdate` / `publishHistoryChange("update", full)`。  
   - **禁止** 裸结构体推送剥 expect*。

6. **与冻结关系**  
   - `historyUpdateData` **不改**。  
   - expect-batch 是唯一允许 update 语义写 expect_* 的业务入口（本 change）。

7. **usage**  
   - 负责人确认不计入；`maintenance_skip` 精确排除  
     `POST /device/history/api/event/expect-batch`。

8. **v1 不清空**  
   - omit = 不改列；不承诺 `null` 清列。

## Risks / Trade-offs

- [静默跳过失败] → 客户端看不到哪条失败；依赖 `ok` + 幂等重跑 / tip 链 epoch。产品明确接受。  
- [上限 200] → 控制锁与往返。  
- [EventBatch 通知剥 expect] → 本 change 不修。  
- [实现已按整批 TX 落地] → apply 时改回逐条部分成功。

## Migration Plan

1. 更新写路径语义 + 文档；手工 curl（部分成功 / 全失败 `ok:false`）。  
2. 与 Flutter twin 联调。  
3. 回滚：下线路由；冻结未动。

## Open Questions

- （无）全失败 → `{ok:false}` + HTTP 200；≥1 成功 → `{ok:true}`；无 results。
