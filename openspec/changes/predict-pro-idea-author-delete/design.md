## Context

`predict-pro-idea-hub` 已提供 App list/submit/like/unlike 与 Admin 软删（`SoftDeletePredictProIdea`，写 `deleted_at` unix 秒）。未删谓词：`deleted_at IS NULL OR deleted_at = 0`（见 `fix-predict-pro-idea-deleted-at`）。Flutter 需要作者自删 App API；姊妹仓同名 change 后对接。

## Goals / Non-Goals

**Goals:**

- App 作者软删端点；鉴权 + 所有权校验。
- 与 Admin 软删同一数据语义，list 立即不可见。
- 先于 Flutter apply 部署。

**Non-Goals:**

- Flutter UI（姊妹仓）。
- 改 Admin 路径或 Hub 页。
- 硬删物理行、改日限额规则（仍按既有 COUNT 语义）。

## Decisions

1. **路径与 body**  
   - `POST /device/app/api/predict-pro-ideas/delete`  
   - Body：`{ "id": <int64> }`（required）  
   - 成功：空 data 或 `{}` 即可（envelope code=0）。

2. **授权**  
   - 从 App 会话取 `wxId`。  
   - 加载活跃想法；若 `idea.wx_id != caller` → 业务错误（禁止删他人）。  
   - 已删/不存在 → 业务错误（与 like 对缺失想法一致；**不**对非作者伪装成 not-found 以外的成功）。  
   - 幂等：对已软删行再次 delete → 业务错误即可（Flutter 以成功路径本地移除为主）。

3. **实现复用**  
   - 优先：新增 `SoftDeletePredictProIdeaByAuthor(ctx, ideaID, wxID)`（校验后写 `deleted_at`），或给现有 SoftDelete 增加 optional owner 校验；**禁止** App 直接调无校验的 Admin 软删。  
   - Controller：`DeviceAppCtrl` 注册新 handler；`api/v1` 增加 Req 结构体。

4. **Gateway**  
   - 确认 `/device/app/api/predict-pro-ideas/*` 代理已覆盖 `delete`；若精确白名单/usage skip 需登记，按 `openspec/project.md` gateway-app 约定处理。  
   - **不得** 加入匿名 auth exempt。

5. **双仓顺序**  
   - 本仓 apply → 部署 `device-service`（及必要时 gateway）→ 再 Flutter apply。

## Risks / Trade-offs

- [误开放无 owner 校验] → 单测/联调必须覆盖非作者拒绝。  
- [与 Admin 双入口] → 同一软删字段，可接受。

## Migration Plan

- 无 DDL；沿用 `deleted_at`。  
- 回滚：下线路由或拒绝请求。

## Open Questions

- （已定）路径对齐 like、软删、作者校验、Go 先于 Flutter。
