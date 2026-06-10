## Context

- UCG 帖子状态：`0=draft, 1=pending_audit, 2=published, 3=rejected`（`ucg_post`）。
- Green 异步 worker（`audit_post.go`）负责 `pending_audit → published/rejected`；`rejectPost()` 为包内私有，仅更新 DB 三字段，不发通知。
- 已有 UCG 管理端：`/device/admin/ucg-admin.html` + `/ucg/admin/api/ai-config`，认证为 Header `X-Admin-Password`（`ucg.admin.password`），网关反代至 `ucg-service`。
- 运营需要人工复核全部动态并批量下架，与 Green 机审互补。

## Goals / Non-Goals

**Goals:**

- 管理端分页列出全部动态（可按 `status` 筛选），展示摘要、媒体缩略图、时间戳。
- 支持勾选多条（本页全选可驳回项）并一次 API 批量驳回。
- 驳回语义与 Green 失败一致：`status=3`、`reject_reason` 默认「违规已下架」。
- 扩展 `ucg-admin.html` Tab，不新建独立 HTML 路由。

**Non-Goals:**

- 人工通过 pending、自定义驳回原因输入、操作人审计、作者通知/站内信。
- 发布时间字段语义或客户端展示修复。
- `ucg_post_recommend` 行清理（Feed 已按 `status=published` 过滤，残留无用户可见影响）。
- 数据库 schema 变更。

## Decisions

### 1. 新增独立 capability 与 service 文件

- **决定**：新增 `internal/services/ucg/post_admin.go`，导出 `ListPostsForAdmin`、`AdminBatchRejectPosts`；`rejectPost` 保留在 `audit_post.go`，admin 层调用同一逻辑（可抽 `rejectPostByID` 避免重复）。
- **理由**：与 Green 审核共用驳回语义，避免两套 UPDATE 分叉。
- **备选**：在 `post.go` 内追加 — 文件已较长，admin 职责分离更清晰。

### 2. 批量驳回 API 形态

- **决定**：`POST /ucg/admin/api/posts/reject`，body `{ postIds: number[], reason?: string }`；响应 `{ rejected, skipped, failed }` 数组。
- **理由**：部分成功报告便于运营；`skipped` 覆盖已是 `rejected` 的幂等场景。
- **上限**：单次最多 100 个 `postId`（与 admin 分页 `pageSize` 上限一致）。
- **备选**：单条 REST `POST .../posts/{id}/reject` — 批量场景需 N 次请求，不符合需求。

### 3. 可驳回状态范围

- **决定**：`draft`、`pending_audit`、`published` 均可驳回；`rejected` 跳过计入 `skipped`。
- **理由**：用户要求「针对所有动态」；已驳回无需重复操作。

### 4. 列表 API

- **决定**：`GET /ucg/admin/api/posts/list?page&pageSize&status`；`status` 省略表示全部；排序 `updated_at DESC`。
- **DTO**：复用 `PostDTO` 核心字段 + 管理端不需 `likedByMe`；作者昵称经 `GetPublicProfile` 填充（失败则省略）。
- **备选**：按 `created_at` 排序 — 管理端更关心最近变动，用 `updated_at`。

### 5. 前端交互

- **决定**：`ucg-admin.html` 增加 Tab「动态审查」；表格 + checkbox；「全选本页可驳回项」；确认对话框后调用 reject API。
- **理由**：复用现有登录态与 `localStorage` 口令；与 AI 配置同域同认证。
- **admin.html**：链接文案改为「UCG 管理」。

### 6. 认证与路由

- **决定**：沿用 `UcgAdminCtrl.requireAdmin` + 网关 `/ucg/admin/api/*` 反代；不新增 JWT 或 device admin 口令混用。
- **理由**：与现有 UCG AI 配置一致，零网关改动。

## Risks / Trade-offs

- **[Risk] 批量请求部分失败** → 响应区分 `rejected/skipped/failed`，前端展示摘要并刷新列表。
- **[Risk] 运营误驳已发布热帖** → 确认对话框；无撤销 API（需重新发帖），接受为运营流程约束。
- **[Risk] Green worker 与人工竞态** → 双方均为 `UPDATE status=3` 或 `2`，最终态一致；人工驳回后 worker 扫到 pending 时 UPDATE 无实质影响。
- **[Trade-off] 不清理 recommend 表** → 实现简单；Feed 查询已过滤，可接受。

## Migration Plan

1. 部署 `ucg-service`（新 API）与网关静态资源（`ucg-admin.html`）。
2. 无需 DB 迁移；回滚仅需还原服务与 HTML。
3. 配置项无变更（复用 `ucg.admin.password`）。

## Open Questions

（无。探索阶段已确认：仅驳回、批量、不通知、不记操作人。）
