## Why

Flutter twin `history-expect-anchor-chain-scheduler` 将在历史 CRUD 成功后于后台按「前一条 tip」重算本条及后续行的 `expect*`，并希望**一次 HTTP** 批量落库。现网 `event/update` / `event/batch` 对 `expect_*` **硬冻结**且不适合作 expect-only 补丁；需要窄口批量写接口，同时保留停表/改备注路径的冻结语义。tip 链中偶发脏 id / 已删行时，**不得**因单条失败回滚整批已可成功的补丁。

## What Changes

- 新增 `POST /device/history/api/event/expect-batch`：body 含 `deviceNo` 与 `items[{id, expectStartTime?, expectEndTime?, expectAmount?}]`。
- 校验 `id` 属于该 `deviceNo` 后，**仅**更新请求中出现的 expect 列（omit = 不改）；v1 **MUST NOT** 支持用 `null` 清空列（除非后续单开）。
- 空 `items` **MUST** 返回 **400**；`items` 上限 **200**；超出 **MUST** 400。
- 批量 **MUST** 逐条尝试写入（**非**整批一棍子打死）：某条外机/不存在/`id<=0`/写失败 → **跳过该条**，继续处理其余；**不得**回滚已成功行。
- 响应仅 **`{ "ok": bool }`**，**无** per-item `results` / 逐条错误信息。只要本批 **≥1 条**成功更新 → `{ "ok": true }`；全部未写上 → `{ "ok": false }`（合法 batch 仍 HTTP 200）。
- 每条写成功后 MUST reload 完整行再更新 cache / `publishHistoryChange("update", …)`，**MUST NOT** 用缺 expect 的裸结构体覆盖。
- **保持**：普通 `event/update`、`event/batch` 的 update/end、停表路径 **MUST NOT** 写入 `expect_*`（`historyUpdateData` 冻结不变）。
- gateway-app：**不计入** usage 统计（负责人确认）；`maintenance_skip` 登记精确 `POST /device/history/api/event/expect-batch`。
- contracts：登记 `HistoryEventExpectBatchPath()`（风格统一；本变更无 Go 强制调用方）。

## 非目标

- 服务端不重算预报。
- 不改 add 可选 expect 落库能力（兼容旧客户端）。
- **不**顺手修 `event/batch` update 的 cache/WS 缺 expect 问题：语音完成的喂养记录当下可不带预测信息，产品可接受；该脚枪另议。
- 不强制 admin/internal 镜像（默认仅 device 路径）。
- 不向客户端返回逐条失败 reason（失败静默跳过）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `history-expect-fields`：在保留 update 冻结的前提下，**增加** expect-batch 专用写路径；明确部分成功（≥1 成功即 `ok:true`）、无 results、空 items→400、usage 排除。

## Impact

- `history-service`：HTTP handler、req/res DTO、逐条 expect-only patch、WS/cache reload（已落地实现需按本契约从「整批事务」改为「部分成功」）。
- `internal/services/contracts`：`HistoryEventExpectBatchPath`。
- `gateway-app`：`usagestats/maintenance_skip.go` 排除项；反代 `/device/history/api/*` 通常已覆盖，无需新 Bind。
- Twin：`flutter_ai_talk` / `history-expect-anchor-chain-scheduler`（可按 `ok` 清 dirty；全失败再试）。
