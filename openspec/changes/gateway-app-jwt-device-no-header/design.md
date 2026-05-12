## Context

当前 gateway-app 在 Bearer 鉴权通过后，依赖 **JWT `sub` = wx.id**，再 **HTTP 调用 device `GET /device/app/api/user/internal/by-id`** 得到 **unionid**，注入 **`X-Internal-Wx-Union-Id`**，下游 device 用户域接口据此查 wx 行；历史 WS 还会再走 **unionid → detail** 校验与首帧 `device_no` 的一致性。`union_id` 仍在 **登录换票写库** 路径使用，但 **per-request 拉 unionid** 增加延迟与故障面。

约束：**App 前端不改动**——不新增/改名对外 JSON 字段；JWT 对客户端为不透明字符串；仍使用 `Authorization: Bearer`。

## Goals / Non-Goals

**Goals:**

- access JWT 在签发时携带 **`device_no` 声明**（已绑设备时非空），网关 **仅本地解析** JWT 后注入 **`X-Internal-Device-No`**，供 history/voice 等反代或设备域读路径直接使用。
- 用 **`X-Internal-Wx-Id`**（来自 JWT **`sub`**，即 wx 主键）替代 **`X-Internal-Wx-Union-Id`** 作为 device **用户域写接口**识别 wx 行的依据，**去掉**网关 **id→unionid** 的 HTTP 热路径（及主要相关 Redis 缓存）。
- 历史 WS：用 **JWT 内 `device_no` 声明与首帧 `device_no` 一致** 完成设备绑定校验，**去掉** unionid 解析链。
- **refresh** 重新签发 access 时，**MUST** 写入与当前会话一致的 **`device_no` 声明**（与登录时 device 返回或刷新时权威读一致）。

**Non-Goals:**

- 不改变微信小程序 **jsCode + platform** 登录与 **unionid 落库** 语义（仍由 device 换票写 `wx.union_id`）。
- 不强制改变 App 在业务请求里 **继续传 `deviceNo` query/body** 的现状（可与头并存；冲突策略在实现中单一约定）。
- 不在本变更中引入第三方 IdP 或替换 JWT 为 Session Cookie。

## Decisions

### D1：JWT 载荷形态（兼容 refresh、对前端透明）

- **决策**：保留 **RFC7519 `sub`** = **`wx.id` 的十进制字符串**（与现 refresh / Redis 会话一致）；新增 **私有声明 `device_no`**（字符串，与库内 `device_no` 一致；未绑定时为空串或省略，由实现固定一种并在评审中单一化）。
- **理由**：refresh 消费仍以 **wxId** 为主键成本低；客户端 **不解析 JWT**，增加 claim **不改变** HTTP 响应字段集合。
- **备选**：`sub` 直接改为 `device_no` → refresh 与 wx 会话锚点需整体重做（已拒绝）。

### D2：网关注入头

- **决策**：Bearer 中间件在验签成功后：
  - 设置 **`X-Internal-Wx-Id`** = `sub`（trim，整数校验）；
  - 若 **`device_no` claim** 非空，设置 **`X-Internal-Device-No`** = 该值；否则 **不设置** `X-Internal-Device-No`（或仅对需要设备头的路由返回 403，由实现与清单统一）。
- **决策**：**不再**为注入目的调用 **`GET .../internal/by-id`**；**不再**设置 **`X-Internal-Wx-Union-Id`**（或保留极短迁移期 feature flag，默认关闭，非目标默认路径）。

### D3：device-service 控制器与 wx 服务

- **决策**：`bindwx` / `auto_save` / `detail` 等从 **`X-Internal-Wx-Union-Id`** 改为 **`X-Internal-Wx-Id`**，用 **主键 id** 定位 `wx` 行（`Where(id, ...)`）；权限语义与「当前登录用户」一致。
- **理由**：`sub` 已由网关验签，等价于已认证 wx 行；避免 unionid 字符串索引与头泄露面。

### D4：内部只读 `internal/by-id`

- **决策**：网关 Bearer 热路径 **不再依赖**；接口可 **保留** 供运维/其它服务，或标记弃用；**不得**再作为 gateway-app 每个请求的必经路径。

### D5：换绑与 token 一致性

- **决策**：当 **wx 绑定 device_no 变更** 后，**已签发 access** 在过期前可能仍携带旧 **`device_no` claim** → **缓解**：bindwx 成功后 **旋转 refresh 并建议客户端立即刷新 access**（产品约定）；或在 bindwx 后 **使该 wx 下 refresh 版本递增** 强制失效旧 refresh（若已有版本键则复用）。实现阶段选 **一种** 并在 tasks 中写清。

## Risks / Trade-offs

- **[Risk] 头伪造**：App 直连 device 并伪造 `X-Internal-Wx-Id` → **缓解**：生产 **仅允许网关网段** 访问带内部头的 user API；保留 **`DEVICE_GATEWAY_INTERNAL_SECRET`** 仅用于真正「内部」路径，不替代 JWT。
- **[Risk] claim 与库不一致**：换绑后 access 短期陈旧 → **缓解**：D5 策略 + 短 **access TTL**。
- **[Trade-off] JWT 略大**：增加 `device_no` claim 长度有限，可接受。

## Migration Plan

1. 先部署 **device**：同时识别 **`X-Internal-Wx-Id`**（新）与 **`X-Internal-Wx-Union-Id`**（旧）**或** 直接切换（无线上用户时可硬切，与产品确认）。
2. 再部署 **gateway-app**：签发新 JWT、改中间件与 WS。
3. **回滚**：回退网关版本；device 若保留双头读取则可平滑回滚。

## Open Questions

- **未绑定设备**时 access 是否允许 **`device_no` 缺省** 仍访问哪些路由（仅登录/绑定/画像初始化）——需在白名单与 403 语义上列清单。
- **history 反代**是否强制要求 **`X-Internal-Device-No`** 与 body/query 的 `deviceNo` 双检——建议 **MUST 一致** 否则 403，防纵向越权。
