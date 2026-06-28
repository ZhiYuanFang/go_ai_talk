## Context

- 管理端已有 `GET /posts/list`、`POST /posts/reject`（`post_admin.go` / `rejectPostByIDAdmin`），静态页 `ucg-admin.html` 仅批量驳回且无 reason 输入。
- MQ 自动审：`publishPostCAS`（1→2）、`rejectPostCAS`；异常态 **4 apply_failed**、**5 moderation_failed** 无 Admin 出口。
- 资料机审 Tab 已有 `resolve { approve | reject }` + 驳回必填 reason，可作为 UX 参考；本变更对帖子采用 **批量 approve/reject**（用户指定），与现有 checkbox 交互一致。
- App「我的动态」已展示 `rejectReason`，无需客户端改动。

## Goals / Non-Goals

**Goals:**

- 新增 `AdminBatchApprovePosts` + `POST /ucg/admin/api/posts/approve`，批量人工发布。
- 可批准 status **1 / 4 / 5**；发布后 `syncPublishedPostRedis`。
- **BREAKING**：reject API/UI 驳回理由必填。
- Admin 列表筛选支持 status 4、5；UI 双按钮 + 驳回 prompt。

**Non-Goals:**

- 不从 **3 rejected** 恢复发布（误驳回滚另开变更）。
- 不批准 **0 draft**（无提审语义）。
- 不新增作者通知/WS 推送。
- 不做单行 resolve API（与资料不同，保持批量对称）。
- 不修改 App v1/v2 HTTP 契约。

## Decisions

### 1. 批量 approve API 形态

- **决定**：`POST /ucg/admin/api/posts/approve`，body `{ postIds: uint64[] }`，响应 `{ approved, skipped, failed }`，上限 100，与 reject 对称。
- **理由**：用户明确要求批量 approve；与现有运营流（勾选多行）一致。

### 2. 可批准 status 与计数语义

| 当前 status | 行为 |
|-------------|------|
| 1 pending_audit | 人工 publish（见 §3）→ `approved` |
| 4 apply_failed | 人工 publish（verdict 已为 pass）→ `approved` |
| 5 moderation_failed | 人工 publish + 写 `moderation_verdict=pass` → `approved` |
| 2 published | 幂等 → `skipped` |
| 3 rejected | 不可批准 → `failed` |
| 0 draft | 不可批准 → `failed` |

### 3. 人工 publish 实现（不调 Green）

新增 `approvePostByIDAdmin(ctx, post)`，按 status 分支 CAS：

```
status=1 (pending_audit):
  Extra: moderation_verdict=pass（若原为 none），reject_reason=""，
         published_at，updated_at
  CAS: FromStatus=pending_audit, ToStatus=published, FromVersion=audit_version

status=4 (apply_failed):
  Extra: reject_reason=""，apply_failed_at=0，published_at，updated_at
  CAS: FromStatus=apply_failed → published

status=5 (moderation_failed):
  Extra: moderation_verdict=pass，moderation_at=now，reject_reason=""，
         published_at，updated_at
  CAS: FromStatus=moderation_failed → published
```

成功后统一 `syncPublishedPostRedis(postID)`（与 `publishPostCAS` 一致）。

- **备选**：复用 `publishPostCAS` 仅处理 status=1 → 无法覆盖 4/5，故独立 admin CAS。

### 4. 驳回 reason 必填（BREAKING）

- **决定**：`AdminBatchRejectPosts` 在 `reason` trim 后为空时返回 400；移除 `rejectReasonDefault` 回退。
- **UI**：批量驳回前 `prompt('请输入驳回原因：')`，空则阻断（对齐 `resolveProfileAuditJob`）。
- **理由**：作者需可读原因；默认「违规已下架」无运营价值。

### 5. Admin UI 选择逻辑

- Checkbox：仍对 **status≠3** 可选（可批驳、可批过）。
- **批量通过**：仅对选中项中 status∈{1,4,5} 生效；其余选中 id 计入 API `failed` 或前端预过滤并提示。
- **批量驳回**：选中非 3 均可；须填 reason。
- 表格增加 **驳回原因** 列（`rejectReason` truncate）；筛选下拉增加「机审失败(5)」「发布失败(4)」。

### 6. 动态审查工具栏纵向居中

- **问题**：全局 `.row` 使用 `align-items: flex-start`（`pangbao-theme.css`），动态审查工具栏内 label、select、button 高度不一致时视觉上顶对齐、显得参差。
- **决定**：为 `#panelPosts` 首行工具栏增加 scoped 样式（优先 `admin-pages.css` 中 `#panelPosts .posts-toolbar` 或 `#panelPosts > .row`）：`display: flex; align-items: center; flex-wrap: wrap; gap` 与现有一致；**不**修改全局 `.row`，避免影响 AI 配置等其它 Tab。
- **范围**：状态 label、`postStatusFilter` select、刷新/批量通过/批量驳回按钮、`postsSelectedHint` 提示文案均在同一 flex 行内垂直居中。

### 7. 认证与通知

- 沿用 `X-Admin-Password`；approve/reject 均 **不** 发通知（延续 spec）。

## Risks / Trade-offs

- [pending 无 verdict 即人工放行] → 仅 Admin 口令可调用；操作可观测日志 `[ucg-admin] post approve id=...`。
- [apply_failed 实为内容违规后误批] → 运营责任；status=4 列表应展示 `rejectReason`（系统文案）供判断。
- [reject reason 必填 BREAKING] → 旧脚本/自动化驳回须带 reason；OpenSpec delta 明确 BREAKING。
- [批量 approve 部分失败] → 响应 `failed` 数组；UI 展示计数并刷新列表。

## Migration Plan

1. 部署 ucg-service + 更新 `ucg-admin.html`（gateway 静态资源）。
2. 无需 DB migration（字段已存在）。
3. 运营：驳回时必须填写面向作者的 reason。

## Open Questions

- 无（批量 approve + 必填 reason 已在 explore 确认）。
