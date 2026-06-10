## Why

UCG 动态目前仅依赖阿里云 Green 异步机审；运营无法在后台查看全部动态并对违规内容做人工下架。需要补齐管理端「列表 + 批量驳回」能力，使已发布或待审帖子可快速改为 `rejected`，作者仅在「我的动态」自行查看原因，不额外推送通知。

## What Changes

- 在 `ucg-service` 新增 Admin API：`GET /ucg/admin/api/posts/list`（分页、按 status 筛选）与 `POST /ucg/admin/api/posts/reject`（批量 `postIds`）。
- 批量驳回复用现有 `rejectPost` 语义：`status=3`、`reject_reason` 非空、`updated_at` 更新；不通知作者、不记录操作人。
- 扩展 `resource/public/ucg-admin.html`：在现有 AI 配置页增加「动态审查」Tab，支持全选本页、批量勾选、确认后批量驳回；`status=rejected` 行不可选。
- `resource/public/admin.html` 入口文案由「UCG AI 配置」调整为「UCG 管理」（链接不变）。
- 本变更**不涉及**客户端发布时间展示、人工通过 pending、操作审计、作者通知。

## Capabilities

### New Capabilities

- `ucg-admin-post-moderation`：UCG 管理端动态列表与批量人工驳回（认证、API 契约、页面交互、驳回语义与边界）。

### Modified Capabilities

（无。Green 自动审核路径与 C 端 Feed 可见性规则不变，仅新增管理端补充能力。）

## Impact

- **服务**：`ucg-service`（`internal/services/ucg`、`internal/controller/ucg_admin_api.go`、`api/v1/ucg_admin_http.go`）。
- **静态页**：`resource/public/ucg-admin.html`、`resource/public/admin.html`。
- **网关**：已有 `/ucg/admin/api/*` 反代与 `X-Admin-Password` 豁免，无需改路由契约。
- **数据库**：复用 `ucg_post` 现有字段，无 schema 迁移。
- **依赖**：无新第三方依赖。
