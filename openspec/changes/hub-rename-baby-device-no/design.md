# Design

## Context

- 宝宝身份键为 `user.device_no`；业务数据按该字符串分区（history/suggest、cash 对机权益、appointment、MCP 落点等）。见 proposal.md — Why。
- 跨服务禁止直查他域库；改名须由 device 编排 + `clients/{history,cash,...}` 契约迁移。
- 今日 refresh 键为 `gw:app:rt:{token}`，无按 wxId/deviceNo 反查；`TokenRefresh` 在 refresh 已携带 `wxId:device_no` 时**不会**回源 device（与 gateway-edge「refresh 同步 device_no」意图存在实现缺口）。
- Flutter 历史写入使用请求体 `deviceNo`（非仅 JWT）；`token/refresh` 响应不含 `deviceNo`；登录响应含 `deviceNo`。不改客户端时须逼重登。
- 本变更**不**为单宝宝行数做分批占库专项（产品确认可一次 `UPDATE`）。

## Goals / Non-Goals

**Goals:**

- Hub 可对已注册宝宝执行 OLD→NEW 原地改名并保留业务数据。
- 改名后 OLD 不可注册校验通过（阻断泄露号新登录/新绑定）。
- 吊销相关 refresh，使 App 走现有登录拿到 NEW（不改 Flutter）。
- history 写前校验已注册，避免旧号静默写空壳。

**Non-Goals:**

- 一账号多宝宝关系模型。
- 强断 Voice/History WS（依赖令牌失效与重连即可）。
- 问卷/订单审计快照跟迁。
- history 超大表分批迁移与占库专项优化。
- 已绑陌生人的智能识别（运维可另用现有 wx 列表人工处理；改名会把 `device_no=OLD` 的绑定一并迁到 NEW）。

## Decisions

### D1 — 编排宿主：device-service Admin API

- Hub 调 `device-service` 管理接口（如 `POST /device/admin/api/device/rename`，具体 path 实现时定），入参 `oldDeviceNo` + `newDeviceNo`（或「生成新号」标志）。
- device 本域事务/顺序更新：`user` → `wx` → `appointment_next` → `xiaozhi_mcp_binding`；再经 HTTP 调 history/cash migrate；最后调 gateway 吊销会话。
- **备选**：Hub 直接调多服务 — 拒绝（权限与顺序难统一）。

### D2 — 数据迁移范围

| 跟迁 | 不跟迁（默认） |
|------|----------------|
| `user.device_no`、`wx.device_no`、`appointment_next`、`xiaozhi_mcp_binding` | `survey_response.device_no`、`predict_pro_idea.device_no` |
| `history`、`suggest` | `feature_order` 等支付审计（可选后续） |
| `feature_entitlement`、`feature_allowed_count`、仍在用的 invite device 维表 | |

- NEW 必须尚未被 `user` 占用；冲突则整单失败。
- Redis：best-effort 删除 OLD 相关 cachekit 键；不作为正确性前提（miss 回源）。

### D3 — 会话失效（不改 Flutter 的杠杆）

- 改名成功后：列出该宝宝绑定的全部 `wxId`（已有 `ListWxIDsByDeviceNo` 语义），吊销其 refresh。
- **实现选型（实现阶段二选一写清）**：
  - **A（推荐）**：签发 refresh 时维护 `wxId → refresh token 集合`（或 sessionEpoch）；吊销时按 wx 删除；或
  - **B**：`wx`/独立表递增 `session_epoch`，refresh 载荷带 epoch，Consume 时校验，改名 bump epoch → 全部旧 refresh 失效（无需扫 token 键）。
- access JWT：无黑名单则依赖较短 TTL；改名后短窗内旧 access 仍可能带着 OLD body 打 history —— 由 **D4 写前注册校验** 挡住落库。
- 纯 `device_login`（wxId=0）会话：旧 refresh `0:OLD` 吊销或校验失败后，无法再用 OLD 登录；须运维把 NEW 告知持有者（本变更不解决无账号场景的自动恢复）。

### D4 — history 写前 EnsureRegistered

- `AddDeviceHistory` / batch create 等写路径 MUST 经 `clients/device` 校验 `EnsureRegistered(deviceNo)`；未注册返回可读错误。
- 与 voice 已有校验对齐，堵住 App 直写历史用旧号。

### D5 — refresh 权威 device_no（兑现既有规格）

- `wxId>0` 的 `token/refresh`：**MUST** 以 device-service `FetchDeviceNoByWxID` 结果写入新 access/refresh，**MUST NOT** 仅信任 refresh 载荷内旧 `device_no` 字符串。
- 仍不足以更新 Flutter 本地 body 号（refresh 响应无 `deviceNo`），故 **不能替代 D3 重登**；但可避免「未吊销时 refresh 继续钉死旧号」。

### D6 — Hub UI

- 设备列表行或详情：改名表单、二次确认文案（旧号失效、家人需重新登录、数据跟迁）。
- 成功后展示 NEW，并提示刷新列表；history 深链改用新号。
- 遵循 admin-hub 近全宽/危险操作确认惯例。

### D7 — 失败与重入

- 跨库无分布式事务：建议顺序 **device 本域先改注册与绑定 → history → cash → 吊销会话**；若后步失败，运维可对同一 OLD/NEW 重试（幂等：以「user 已是 NEW 且 OLD 无注册」为已完成 device 段，继续补迁残留行）。
- 不做自动回滚到 OLD（避免再次暴露泄露号）。

## Risks / Trade-offs

- [已绑攻击者随 `wx` 一起迁到 NEW] → 缓解：Hub 改名前/后展示该宝宝绑定 wx 列表，人工 deactivate；本变更不自动踢绑定。
- [access TTL 窗口仍持旧 JWT] → D4 拒绝未注册旧号写入；读旧号则空数据。
- [半迁残留] → 可重入补迁；Hub 返回分步错误信息。
- [纯设备会话无法自助恢复] → 运维线下告知 NEW；或要求用户使用微信/用户名账号。
- [吊销索引/epoch 为新基础设施] → 改名低频，但正确性依赖 D3；优先 sessionEpoch 以减少 Redis 集合维护。

## Migration Plan

1. 部署 gateway（refresh 权威 + 吊销能力）→ history（EnsureRegistered + migrate API）→ cash（migrate API）→ device（编排 + Hub API）→ 静态 Hub 页。
2. 回滚：停用 Hub 改名入口；已成功改名的数据不自动改回。
3. 重部署清单（apply 收尾再核）：`gateway-app-server`、`history-service`、`cash-service`、`device-service`；Hub 静态随 gateway-app 或主网关静态托管路径。

## Open Questions

- sessionEpoch（D3-B）落在 `wx` 列还是独立 Redis/表：实现时选维护成本更低者即可，不影响规格语义。
- `feature_order.device_no` 是否跟迁：默认不跟；若财务对账强依赖「当前号」再开增量。
