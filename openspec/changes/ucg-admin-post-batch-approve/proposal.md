## Why

UCG 管理端「动态审查」目前仅有批量驳回，且驳回理由可选（UI 甚至不传 reason），作者常只能看到默认「违规已下架」，无法了解具体原因。同时待审（`pending_audit`）、机审异常（`moderation_failed` / `apply_failed`）帖子无法人工放行，运营只能等待 MQ 或无法处理卡单。

## What Changes

- 新增 **`POST /ucg/admin/api/posts/approve`**：批量人工通过，语义对称于现有 `posts/reject`（`postIds` 最多 100，返回 `approved` / `skipped` / `failed`）。
- 人工通过目标 status：**1 pending_audit、4 apply_failed、5 moderation_failed**；已是 **2 published** 计入 `skipped`；**0 draft / 3 rejected** 计入 `failed` 或 `skipped`（见 design）。
- 人工通过 MUST 调用与 MQ Phase2 等价的 publish 路径（`published` + `syncPublishedPostRedis`），**不再调 Green**。
- **BREAKING**：`POST /ucg/admin/api/posts/reject` 的 `reason` 改为**必填**（trim 后非空）；移除空 reason 时默认「违规已下架」的行为。
- `ucg-admin.html` 动态审查 Tab：增加「批量通过」；批量驳回前 MUST 弹出理由输入（对齐资料机审 Tab）；列表筛选增加 status **4 / 5**；表格增加「驳回原因」列（已驳回帖展示）。
- 动态审查 **工具栏行**（状态筛选 + 刷新 + 批量通过/驳回 + 已选提示）：子元素 MUST **纵向居中对齐**（label / select / button / hint 同一视觉基线）。
- App 端 **无 API 结构变更**；作者仍通过「我的动态」`rejectReason` 查看驳回原因（已实现）。

## Capabilities

### New Capabilities

（无独立新 capability；能力归入既有 admin 域。）

### Modified Capabilities

- `ucg-admin-post-moderation`：新增批量 approve API 与 UI；reject reason 必填；列表 status 筛选扩展 4/5；admin 页双按钮批量操作；动态审查工具栏纵向居中。

## Impact

- **代码**：`internal/services/ucg/post_admin.go`、`audit_post.go`（或新 `post_admin_approve.go`）、`internal/controller/ucg_admin_api.go`、`api/v1/ucg_admin_http.go`、`resource/public/ucg-admin.html`、`resource/public/admin-pages.css`（或页面内 scoped 样式）
- **规格**：`ucg-admin-post-moderation` delta
- **App / Flutter**：无需改动（只读 `rejectReason`）
- **通知**：approve/reject 均不发作者通知（延续现有 reject 语义）
