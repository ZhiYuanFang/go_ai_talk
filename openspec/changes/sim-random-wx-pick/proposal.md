## Why

sim 各任务（T2 评论、T3 图文、T4 视频、E1 聊天、T6 关注）经 `randomSimSession` 随机选取模拟用户时，当前实现拉取 `GET /device/internal/api/sim/wx/list?page=1&pageSize=200` 分页列表并在内存随机，且 `randomSimSession` 对同一 list **重复请求两次**；超过 200 个 sim 用户时后段用户永不可选。随机选 1 个用户不应依赖分页语义，应改为 device 侧有界 ID 探测返回单条。

## What Changes

- device internal 新增 `GET /device/internal/api/sim/wx/random`：返回 0 或 1 条 `{wxId, account}`，覆盖全库 `is_simulated=1` 集合（有界 MIN/MAX + `id >= R LIMIT 1`，对齐 ucg sample random 思路）。
- 保留现有 `GET /device/internal/api/sim/wx/list` 分页接口（Admin / `countSimUsers` 取 total 等场景不变）。
- sim-user-service：`randomSimSession` 改为单次调用 random 接口；T6 `RunFollowTask` 改为两次 random（第二次排除第一个 wxId）或等价逻辑，删除 `listSimWxIDs` / `accountForWx` 在随机路径上的分页拉全表用法。
- 不新增 Redis 读缓存；不新增 `*_test.go`。

## Capabilities

### New Capabilities

（无 — 在既有 device sim 与 sim-user-service 能力上增量修改。）

### Modified Capabilities

- `device-sim-user`：新增 random 单条选取 internal API；list 分页 MUST 保留供列举/计数，MUST NOT 作为 sim 随机选取的唯一路径。
- `sim-user-service`：随机 sim 用户 MUST 经 device random 接口一次 HTTP 取得 `wxId+account`；MUST NOT 对 sim wx list 分页结果再做 `rand`；MUST NOT 为查 account 再拉整页 list 线性扫描。

## Impact

- **代码**：`internal/services/device/sim_user.go`、`internal/controller/device_sim_internal.go`、`api/v1/device_sim_internal_http.go`、`internal/services/simuser/clients.go`、`internal/services/simuser/tasks.go`（T6 follow）。
- **进程**：先 **device-service**，再 **sim-user-service**。
- **DB**：无迁移（复用 `wx.is_simulated` 与主键 `id`）。
- **OpenSpec**：delta 挂于本 change。
