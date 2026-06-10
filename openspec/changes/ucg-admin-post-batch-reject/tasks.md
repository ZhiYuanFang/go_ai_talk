## 1. 服务层与驳回复用

- [x] 1.1 在 `audit_post.go` 抽取或暴露可复用的 `rejectPostByID`（供 Green 与 Admin 共用），保持默认原因「违规已下架」
- [x] 1.2 新增 `internal/services/ucg/post_admin.go`：实现 `ListPostsForAdmin`（分页、`status` 筛选、`updated_at DESC`、填充媒体与作者摘要）
- [x] 1.3 实现 `AdminBatchRejectPosts`（最多 100 条、返回 `rejected/skipped/failed`、不发送通知）

## 2. Admin API 契约与控制器

- [x] 2.1 在 `api/v1/ucg_admin_http.go` 增加 `posts/list` 与 `posts/reject` 请求/响应类型
- [x] 2.2 在 `internal/controller/ucg_admin_api.go` 增加 `PostsList` 与 `PostsReject` 处理器，复用 `requireAdmin`

## 3. 管理页 UI

- [x] 3.1 扩展 `resource/public/ucg-admin.html`：Tab 切换（AI 配置 / 动态审查）、状态筛选、分页表格、缩略图展示
- [x] 3.2 实现本页 checkbox、全选可驳回项、确认对话框与批量驳回 API 调用及结果提示
- [x] 3.3 更新 `resource/public/admin.html` 入口链接文案为「UCG 管理」

## 4. 验证

- [x] 4.1 本地验证：列表分页与 status 筛选、批量驳回 published/pending、已 rejected 幂等 skipped、错误口令 401
- [x] 4.2 确认驳回后 Feed 查询不再包含该帖，作者「我的动态」可见 `rejectReason`
