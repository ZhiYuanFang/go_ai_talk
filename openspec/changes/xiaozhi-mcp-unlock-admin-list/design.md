## Context

`xiaozhi-mcp-permanent-unlock` 已落地独立 Hub 页（区 A SKU + 区 B 手工授/撤销）与表 `xiaozhi_mcp_entitlement`；试用门禁在 `xiaozhi-mcp-trial-usage-gate` 扩展了 `expires_at` / `unlock_method=trial`。运维缺少「谁已开通」视图。开通功能管理已有可对标模式：`AdminListFeatureActivationSnapshot` + 静态页行内撤销。

约束：Admin 口令鉴权；cash 经 `clients/ucg` 拉昵称，禁止跨库直查 ucg；不新增 Redis；不新增背景循环；非 App 接口不碰 usage。

## Goals / Non-Goals

**Goals:**

- Admin 分页列出 `xiaozhi_mcp_entitlement` 当前快照（含有效与已失效行）。
- 列表补昵称；Hub 区 C 展示并对齐功能页交互密度。
- 行内撤销仅对仍有效的 `unlock_method=admin` 行，复用既有 revoke API。

**Non-Goals:**

- 不改支付/App 开通/device Add 门禁语义。
- 不新增「撤销支付开通」能力。
- 不做完整历史流水（仅权益表当前行快照）。
- 不把列表并入开通功能管理页。
- 不为列表加 Redis 缓存。

## Decisions

### 1. API：`GET /cash/admin/api/xiaozhi-mcp/entitlements`

- **选择**：独立 list 接口，query：`limit`（默认 50，最大 200）、`offset`（默认 0）。响应：`total`、`note`、`list[]`。
- 行字段：`wxId`、`nickname`、`unlockMethod`、`channelRef`、`unlockedAt`、`expiresAt`、`active`、`remainingSeconds`、`status`、`revokedAt`、`updatedAt`。
- **active 计算**：`status=有效` 且（`expires_at=0` 永久 **或** `expires_at > now` 试用未过期）；否则 `active=false`。
- **理由**：与 `/cash/admin/api/feature/activations` 同形；路径落在既有 cash Admin 反代。
- **备选**：扩展 product GET 夹带 list → 破坏区 A 单一职责，否决。

### 2. 数据范围：表内全量分页，非「仅有效」

- **选择**：`COUNT(*)` + `ORDER BY updated_at DESC, wx_id DESC` 分页扫全表；前端用 `active`/badge 区分状态。
- **理由**：对齐功能开通快照（含过期行便于审计）；撤销后行仍在，运维可核对。
- **备选**：默认仅 active → 丢审计上下文；若日后量很大再加 `activeOnly` 过滤。

### 3. 昵称：复用 `ucgclient.FetchUcgNicknames`

- **选择**：对本页 `wxId` 批量调用；失败时昵称空串、列表仍成功返回（与 feature snapshot `nicks, _ :=` 一致）。
- **理由**：零新依赖；符合服务边界。

### 4. 行内撤销：复用 `POST .../grants/revoke`

- **选择**：前端仅对 `active && unlockMethod==='admin'` 渲染撤销按钮；调用既有 `AdminRevokeXiaozhiMcp`（支付/试用不可由此撤）。
- **理由**：业务规则已实现；本变更只补发现与操作入口。
- **备选**：新 revoke-by-id → 无必要。

### 5. Hub UI：区 C 固定展示

- **选择**：`cash-xiaozhi-mcp-admin.html` 在区 B 下增加「区 C · 已开通」表；进页加载；刷新按钮；简单 prev/next 或 offset 分页；授/撤成功后刷新区 C。
- **列建议**：wxId、昵称、方式、状态、到期/剩余、开通时间、更新时间、channel_ref、操作。
- **理由**：运维一屏可见，无需再点「查看快照」。

### 6. 实现落点

- 领域：`internal/services/cash` 新增 `AdminListXiaozhiMcpEntitlements`（可与 unlock 同文件或旁路小文件）。
- 控制器：`CashXiaozhiMcpController` 增方法；DTO 进 `api/v1/cash_xiaozhi_mcp_http.go`。
- 鉴权：与现有 Xiaozhi MCP Admin 一致（Admin 口令中间件已覆盖 `/cash/admin/api/*`）。

## Risks / Trade-offs

- [Risk] 权益表变大后全表分页变慢 → **Mitigation**：首版量小；已有 `idx_status`；必要时再加 `updated_at` 索引与 `activeOnly`。
- [Risk] ucg 昵称失败导致运维困惑 → **Mitigation**：昵称失败降级为空；wxId 仍可操作。
- [Risk] 试用过期行显示为无效，运维误以为「列表漏人」 → **Mitigation**：`note` 文案说明「当前权益快照，含已撤销/试用过期」；badge 标明无效原因语义（方式=trial + 过期）。

## Migration Plan

- 纯增量：发版 cash-service + 静态页即可；无 schema 变更、无数据迁移。
- 回滚：去掉 list 路由与区 C；权益数据不受影响。

## Open Questions

- （无阻塞项。若产品后续要「仅有效」过滤，可加 query 而不改默认。）
