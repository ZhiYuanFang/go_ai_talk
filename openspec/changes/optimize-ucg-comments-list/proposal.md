## Why

UCG 帖子评论列表当前采用分页 + 逐条 `GetPublicProfile`，客户端需多轮 GET 才能拉全评论，且每条评论额外触发 profile 查询（典型 N+1）。进入帖子详情时延迟与 DB 压力随评论数线性增长。负责人已确认：**评论读路径不引入 Redis**，且 **不在评论表落 profile 快照**，需在保持与帖子一致的「实时 profile」前提下优化列表与客户端交互。

## What Changes

- **ListComments**（`internal/services/ucg/social.go`）：单次查询返回该帖全部评论（`ORDER BY created_at ASC`，最新在列表底部）；批量 `IN` 查询 `ucg_profile` 填充作者信息，消除 N+1；移除独立 `COUNT(*)`，总数优先使用帖子 `commentCount`，截断场景配合 `truncated` 标志
- 可选硬上限（默认 **500** 条）：超过上限时仅返回最早 500 条并设 `truncated=true`
- **GET `/ucg/app/api/posts/{id}/comments`** 响应契约调整：**BREAKING** — 由分页 `{ list, total, page, pageSize }` 改为 `{ list, total, truncated }`；`page`/`pageSize` 查询参数废弃（可忽略或后续从 API 定义移除）
- **AddComment** 保持返回完整 `CommentDTO`（含 `author`），供客户端乐观追加
- **Flutter**（兄弟仓库 `flutter_ai_talk`）：`fetchComments` 单次 GET；`ucg_post_detail_screen` 发帖评论后乐观 append，不再全量 refetch
- **不**新增 Redis 读缓存；**不**新增评论表 profile 快照列或 schema 迁移

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-app-http-api`：评论列表由分页改为单次全量（可截断）列表；批量 profile 填充；响应字段与排序语义

## Impact

- `internal/services/ucg/social.go` — `ListComments`、可能新增/复用批量 profile 辅助
- `internal/services/ucg/profile.go` — 批量公开 profile 查询（`ucg_profile IN (...)` + 批量 IP 属地）
- `api/v1/ucg_app_http.go` — `UcgPostCommentsGetReq` / `UcgCommentsPageRes`（或更名）响应结构
- `internal/controller/ucg_app_api.go` — 评论列表映射
- **gateway-app** 使用统计：负责人已确认 **`GET /ucg/app/api/posts/{id}/comments` 不计入** App API 使用统计；实现阶段在 `maintenance_skip.go` 增加 GET-only 排除（见 `design.md`）
- **Flutter** `lib/.../ucg_repository.dart`、`ucg_post_detail_screen.dart`（文档化于 tasks，不在本仓库实现）
- 部署：仅需 **ucg-service**（及 gateway-app 若改 usage 排除列表）镜像更新；无 DB 迁移
