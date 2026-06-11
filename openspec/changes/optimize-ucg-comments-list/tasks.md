## 1. 负责人确认（已完成）

- [x] 1.1 **App API 使用统计** — 负责人已确认：`GET /ucg/app/api/posts/{id}/comments` **不计入** usage 统计（归一化 apiKey：`GET /ucg/app/api/posts/{id}/comments`；gateway 原始 path：`/ucg/app/api/posts/<postId>/comments`）。POST 发表评论仍计入。

## 2. gateway-app 使用统计排除

- [x] 2.1 在 `internal/services/gatewayapp/usagestats/maintenance_skip.go` 增加 **GET-only** 排除：原始 path 匹配 `/ucg/app/api/posts/<numericPostId>/comments` 时 `isMaintenanceAPI` 返回 true；**不得**误排除 POST 同路径或其他 posts 子路径（如 `mine`、`/posts/{id}` 单帖 GET）
- [x] 2.2 补充中文注释：排除原因（负责人确认评论列表为高频读、不统计）及与 apiregistry 模板 apiKey 的对应关系

## 3. ucg-service 批量 profile

- [x] 3.1 在 `internal/services/ucg/profile.go` 实现 `GetPublicProfilesByWxIDs`（或等价导出函数）：`ucg_profile WHERE wx_id IN (...)` 单次查询 + 批量 `IpLocationForWxIDs`；缺省昵称刷新语义与 `GetPublicProfile` 对齐
- [x] 3.2 补充中文注释：批量路径与单条路径的语义一致性与失败时 `author` 省略行为

## 4. ListComments 重构

- [x] 4.1 重构 `internal/services/ucg/social.go` 的 `ListComments`：合并 `ensurePublishedPost` 与读取 `comment_count`；`ORDER BY created_at ASC` 单次列表查询；应用 `ucg.comments.listMax`（默认 500）
- [x] 4.2 移除 `ListComments` 内 `COUNT(*)`；`total` 使用帖子 `comment_count`，截断时设 `truncated=true`
- [x] 4.3 使用批量 profile 填充 `CommentDTO.Author`，消除 N+1
- [x] 4.4 定义列表结果类型（替代或扩展 `PageResult`）：含 `List`、`Total`、`Truncated`，不含 `Page`/`PageSize`

## 5. HTTP API 与控制器

- [x] 5.1 更新 `api/v1/ucg_app_http.go`：`UcgPostCommentsGetReq` 移除或标记废弃 `page`/`pageSize`；`UcgCommentsPageRes` 改为 `{ list, total, truncated }`（可保留类型名或更名为 `UcgCommentsListRes`）
- [x] 5.2 更新 `internal/controller/ucg_app_api.go` 的 `PostCommentsGet` 与 `commentsPageToRes` 映射
- [x] 5.3 在 `manifest/config/config.ucg-service.yaml` 增加 `ucg.comments.listMax`（默认 500）及顶部中文说明

## 6. Flutter（兄弟仓库 `d:\work\flutter_ai_talk`）

- [x] 6.1 `ucg_repository.dart`：`fetchComments` 改为单次 GET，解析 `list`/`total`/`truncated`；删除客户端 page 循环
- [x] 6.2 `ucg_post_detail_screen.dart`：加载详情时一次拉取评论；`POST` 评论成功后乐观 append 响应体到列表末尾并递增本地 `commentCount`，不再全量 `fetchComments`
- [x] 6.3 UI：当 `truncated=true` 时展示提示（如「仅显示前 500 条评论」）

## 7. 验收

- [x] 7.1 手工验证：10+ 评论帖子单次 GET 返回升序全量、`author` 均有昵称/头像，DB/日志无逐条 profile 查询风暴
- [x] 7.2 手工验证：发表评论后 Flutter 列表即时追加、无二次 GET
- [x] 7.3 手工验证：`comment_count > listMax` 时 `truncated=true` 且 `list` 长度为 `listMax`
- [x] 7.4 手工验证：成功 GET 评论列表不写入 usage 统计；成功 POST 发表评论仍写入
- [x] 7.5 确认未引入 Redis 读缓存、未新增 DB migration
