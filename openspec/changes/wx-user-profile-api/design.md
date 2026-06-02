## Context

device 用户域在 `wx-username-auth` 变更后支持两类账号形态：微信登录（`unionid` 非空）与用户名密码（`account` 非空，且 `unionid` 可为空）。客户端在个人中心、绑定引导等场景需要展示当前账号状态，但现有 `GET /device/app/api/user/detail` 仅返回 `deviceNo`，无法一次获取「是否绑定微信」与「用户名账号」。

约束：
- 会话主体统一为 `wx.id`，由 gateway-app 从 access JWT `sub` 注入 `X-Internal-Wx-Id`。
- 数据读取仅在 device 域内完成，复用 `wx` 表与现有分层。
- 不向客户端泄露 `unionid`、`password` 等敏感字段。

## Goals / Non-Goals

**Goals:**
- 提供 `GET /device/app/api/user/profile`，无额外入参。
- 一次返回 `isWxBound`、`account`（有值时）、`deviceNo`。
- 对齐现有错误语义（头非法、wx 行不存在）。

**Non-Goals:**
- 不扩展返回宝宝画像（`babyName`/`birthday`/`sex`），该能力仍由 `/user/get` 承担。
- 不废弃或修改 `GET /device/app/api/user/detail` 行为。
- 不在 gateway-app 新增聚合层或改 JWT 结构。
- 不新增 Redis 缓存键（读频率低，单次查库即可）。

## Decisions

1. **接口路径与方法**
   - 采用 `GET /device/app/api/user/profile`。
   - 原因：语义为「当前登录用户的账号 profile」，与按 `deviceNo` 查询的 `/user/get`（设备宝宝画像）区分。
   - 备选：`/user/account`；未采用，用户已选定 `/profile` 且与「个人中心」语义更贴近。

2. **身份来源**
   - 从请求头 `X-Internal-Wx-Id` 读取 wx 主键，复用 `wxIDFromAppUserHeader`。
   - 原因：与 `detail`、`bindwx`、`deactivate` 等接口一致。
   - 备选：body/query 传 `wxId`；未采用，存在伪造风险且违背 token 会话模型。

3. **数据读取**
   - 单次调用 `wxRowByWxID(ctx, wxID)` 读取整行，从同一行派生三个响应字段。
   - 原因：避免分别调用 `WxDeviceNoByWxID` 与额外查询造成两次 DB/缓存访问。
   - 备选：复用 `WxDeviceNoByWxID` + 新查 account；未采用，冗余且逻辑分散。

4. **字段语义**
   - `isWxBound`：`strings.TrimSpace(row.Unionid) != ""`，始终返回 bool。
   - `account`：`strings.TrimSpace(row.Account)`，非空时返回；JSON 使用 `omitempty` 省略空值。
   - `deviceNo`：`strings.TrimSpace(row.DeviceNo)`，始终返回；未绑定时为 `""`（与 `detail` 一致）。

5. **错误语义**
   - 头缺失/非法：参数错误（与现有 app 用户域一致）。
   - wx 行不存在：404（对齐 `deactivate` 的 `ErrWxDeactivateNotFound` 映射）。

## Risks / Trade-offs

- [风险] `/user/profile` 与 `/user/get` 均含「profile」字样，易混淆。  
  → 缓解：API 文档与 OpenSpec 明确 `/get` 为设备宝宝画像、`/profile` 为账号状态。
- [风险] `detail` 与 `profile` 均返回 `deviceNo`，存在冗余调用。  
  → 缓解：保留 `detail` 兼容旧客户端；新客户端统一使用 `profile`。
- [权衡] 不缓存 profile 读路径，实现简单但高并发下略增 DB 压力。  
  → 缓解：个人中心读频率低；若后续有热点可再引入只读缓存。

## Migration Plan

1. 在 `api/v1` 新增 `DeviceUserProfileReq/Res` 并绑定 `GET /device/app/api/user/profile`。
2. 在 `DeviceAppUserCtrl` 新增 `Profile` 方法，复用头解析。
3. 在 `internal/services/device/wx.go` 新增 `WxUserProfileByWxID`（或等价函数）封装字段派生。
4. 联调验证：纯微信、纯用户名、两者兼有、未绑设备、wx 行不存在、头非法。
5. 回滚：回滚代码即可；新增只读接口，无数据迁移。

## Open Questions

- （无）
