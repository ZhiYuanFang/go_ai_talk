## Context

- sim 任务经 `randomSimSession` / `sessionForWx` / `listSimWxIDs` 选取模拟用户；当前均依赖 `GET /device/internal/api/sim/wx/list?page=1&pageSize=200`。
- `randomSimSession` 先调 `listSimWxIDs` 再 **重复** 请求同一 list 取 `account`；`accountForWx` 再拉整页 list 线性查找。
- device `ListSimulatedWx` 分页上限 200，全库 sim 用户超过 200 时随机池不完整。
- ucg `posts/sample` `mode=random` 使用幂次偏置 ID 探测；**sim wx random 采用均匀锚点**，所有模拟用户平等对待。

## Goals / Non-Goals

**Goals:**

- device 提供 `GET /device/internal/api/sim/wx/random`，返回 0 或 1 条 `{wxId, account}`，覆盖全库 `is_simulated=1`。
- 查询有界：MIN/MAX + 单次 `id >= R LIMIT 1`（共 2 次 SQL）；MUST NOT `ORDER BY RAND()`。
- sim `randomSimSession` 单次 HTTP 取得 account 并登录；T6 follow 选取两个不同 wxId 且不再经 list 扫 account。
- 保留 `sim/wx/list` 分页与 `countSimUsers`（pageSize=1 取 total）行为。

**Non-Goals:**

- 不删除 list 接口；不改 sim 注册 T1 流程。
- 不引入 Redis SRANDMEMBER（sim 不直连 device Redis）。
- 不新增 `*_test.go`。
- 首期不暴露 env/Admin 配置（包内均匀随机即可）。

## Decisions

### 1. 新接口：`GET /device/internal/api/sim/wx/random`

**选择**：独立 path，响应 `{ wxId, account }` 或空（无 sim 用户时 200 + 空 data / found=false，与 list 空列表语义一致）。

**鉴权**：与 list 相同，有效 `X-Device-Gateway-Internal-Secret`。

**备选**：list 加 `mode=random` — 可行但 GET list 语义混杂；独立 random 更清晰。

### 2. 随机算法：均匀 ID 探测

```
1) SELECT MIN(id), MAX(id) FROM wx WHERE is_simulated=1

2) 无行或 min=max=0 → 空

3) U ~ Uniform(0,1)  (crypto/rand)

4) R = minId + floor((maxId - minId) * U)   // 均匀，无 α 偏置

5) SELECT id, account FROM wx
   WHERE is_simulated=1 AND id >= R
   ORDER BY id ASC LIMIT 1

6) 若 5 无行且 eligible 存在 → 回退 minId 一条
```

**说明**：与 ucg 帖抽样不同，sim 用户随机 **不做新用户权重**；id 空洞时回退 minId 与 ucg sample 一致。

实现于 `internal/services/device/sim_user.go`：`PickRandomSimulatedWx(ctx) (SimWxListItem, ok bool, err error)`。

### 3. sim 客户端简化

```
randomSimSession:
  pickRandomSimWx()  → GET .../sim/wx/random  → account
  usernameLogin(account)

RunFollowTask:
  a := pickRandomSimWx()
  b := pickRandomSimWx()  // 若 b.WxId==a.WxId 且总数≥2，重试至多 K 次
  login(a.Account) → POST follow/{b.WxId}
```

删除或内联仅 random 路径使用的 `listSimWxIDs`；删除 `accountForWx`（或改为按 wxId 的 internal 单条查询若别处仍需要 — 当前仅 follow 用，可删）。

`countSimUsers` 继续 `list?page=1&pageSize=1` 读 `total`。

### 4. 契约文件

- `api/v1/device_sim_internal_http.go`：`DeviceSimWxRandomReq` / `DeviceSimWxRandomRes`。
- `internal/controller/device_sim_internal.go`：`SimWxRandom` handler。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| follow 仅 1 个 sim 用户时无法互关 | 保持「sim 用户不足」失败语义 |
| ID 空洞导致 probe miss | minId 回退，与 ucg random 一致 |
| device 未部署 random 接口时 sim 升级 | 部署顺序：先 device 后 sim |

## Migration Plan

1. 部署 **device-service**（新 random 路由）。
2. 部署 **sim-user-service**（改 clients/tasks）。
3. 回滚：revert sim 仍可调 list（旧版 sim 兼容旧行为）；revert device 仅影响新 sim。

## Open Questions

- random 响应无用户时 sim 统一错误文案「无模拟用户」（与现网一致）。
