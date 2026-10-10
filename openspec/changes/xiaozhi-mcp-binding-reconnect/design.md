## Context

小智 MCP 绑定由 device-service 持久化；`xiaozhi-mcp-service`（`mcpbridge.Manager`）为每条 active 绑定维护出站 WebSocket 会话。列表接口经 `clients/xiaozhimcp.ConnectionStatus` 返回 `connected`。断线后 Manager 靠内置退避重拨；用户侧无「立即重连」入口。

今日 `Manager.Upsert`：同 token 且 deviceNo/wxId 不变时仅刷新 binding id 并 **noop**，不会 cancel/restart。因此「再调一次 Upsert」无法满足重连。Python 不在本链路，本设计仅覆盖 Go 与 Hub 静态页。

约束：服务边界 device → `clients/xiaozhimcp` HTTP；禁止跨包 import `services/mcpbridge`；不改既有 v1 List/Add 结构；无 Redis；无新 ticker；gateway-app 新 App 路径须 Bearer（不 exempt）；usage 是否计入须先问负责人。

## Goals / Non-Goals

**Goals:**

- 提供专用 App 重连 API，归属校验后强制 Manager cancel + `startLocked`、重置退避、立即 redial。
- 接口快速返回；客户端刷新列表观测 `connected`。
- Hub `history.html` 小智区在 `!connected` 时提供「重连」。
- 文档化 entitlement 与重连的交互（允许对仍属本 wx 的 active 行重连）。

**Non-Goals:**

- Flutter UI（兄弟仓后续）。
- 任何 Python 改动。
- 修改 List/Add/Alias/Delete 的既有请求/响应字段。
- Redis 缓存、新的后台 reconcile/ticker。
- 保证重连后一定 `connected=true`（网络/鉴权失败仍可能 false；客户端刷新即可）。

## Decisions

### 1. 专用 reconnect，不用 Upsert force 标志（首选）

- **决定**：新增内部 `POST /xiaozhi-mcp/internal/api/bindings/reconnect`，body 含 `id`/`mcpToken`/`deviceNo`/`wxId`（与 Upsert 对齐）；Manager 暴露 `ForceRestart`（或同名）：若存在会话则 cancel + 从 map 移除，再 `startLocked`；无会话则直接 `startLocked`。
- **理由**：今日同 key Upsert 为 noop；在 Upsert 上加 `force` 会改变语义并易被误用。专用路径意图清晰。
- **备选**：Upsert + `force=true` — 可行但耦合写路径；本变更不采用。

### 2. 调用链

```
Flutter / Hub
  POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect  (Bearer)
  → gateway-app 反代 device-service
  → device: 校验 wx 拥有该 binding 且 active；读 token/deviceNo/wxId
  → clients/xiaozhimcp.BindingReconnect → 内部 reconnect
  → Manager.ForceRestart → 快速 200
客户端再 GET bindings 看 connected
```

- device 失败语义：不存在/非本人 → 业务 4xx；mcpbridge 不可达 → 5xx/包装错误，**不**改 DB。
- **不**在 reconnect 路径再次调用 cash EnsureAccessForAdd；List 本就可在未解锁时查看已有绑定。

### 3. Entitlement：允许对归属行重连

- **决定**：只要绑定行 active 且归属当前 wx，即允许 reconnect；**不**因 entitlement 过期而拒绝。
- **理由**：与「可 List/Delete 已有绑定」一致；用户意图是恢复已绑定音箱的链路。
- **副作用**：若 entitlement 已失效，Manager/tools 侧可能 redial 后再次因鉴权/usage 被 lazy-remove；属既有行为，文档说明即可，本变更不新增「过期即删」逻辑。

### 4. 速率限制（轻量，SHOULD）

- **决定**：设计层 SHOULD 在 device 或 mcpbridge 侧对同一 binding id（或同一 wx）做轻量节流（如数秒内重复 reconnect 直接成功返回或 429）；实现阶段可选，不阻塞主路径。
- **理由**：防止连点打爆 dial；非硬门禁。

### 5. API 版本与契约

- **决定**：在 `api/v1` **新增** reconnect Req/Res 类型与 `g.Meta` 路径；**禁止**改动既有 List/Add 等结构体字段。
- gateway：路径落在既有 `/device/app/api/*` 反代；**不**加入 `gateway_app_auth_exempt`。
- usage：tasks 中单独项「向负责人确认是否计入 usage」；未确认前 **不得** 改 `maintenance_skip.go`。既有绑定 CRUD 规格倾向计入；默认实现倾向不写入 skip，但仍以负责人为准。

### 6. Hub UI

- `history.html` 列表行：`connected === false` 时显示「重连」按钮；点击 `POST .../bindings/{id}/reconnect`（App Bearer），成功后刷新列表。
- 未 App 登录时与其它写操作一致禁用。

### 7. 无 Redis / 无新 ticker

- 连接态仍由现有 ConnectionStatus 查询；重连不落缓存键。
- 不新增后台循环；仅请求驱动 ForceRestart。

## Risks / Trade-offs

- **[Risk] 重连成功但短暂仍 `connected=false`** → 客户端提示「已触发重连，请稍后刷新」；接口不阻塞等待 WS 握手完成。
- **[Risk] entitlement 过期后重连又被摘掉** → 文档说明；产品可引导去开通页；不在本 API 强拦。
- **[Risk] 连点重连** → SHOULD 轻量节流；Hub/Flutter 按钮 loading 防抖。
- **[Risk] mcpbridge 调用失败** → device 返回错误，绑定行不变；用户可重试。

## Migration Plan

1. 先部署 `xiaozhi-mcp-service`（含 reconnect 内部路由与 ForceRestart），再部署 `device-service` + gateway-app 静态 `history.html`。
2. 回滚：去掉 App/内部路由即可；Manager 无状态迁移问题。
3. Flutter：Go API 稳定后再在兄弟仓接线。

## Open Questions

- usage 统计：负责人确认 `POST .../bindings/{id}/reconnect` **计入** App API usage → 不写入 `maintenance_skip.go`。
