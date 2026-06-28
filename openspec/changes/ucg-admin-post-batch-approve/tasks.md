## 1. 后端 API — 批量通过

- [x] 1.1 在 `post_admin.go` 新增 `AdminBatchApproveResult` 与 `AdminBatchApprovePosts`（上限 100，approved/skipped/failed 语义见 design）
- [x] 1.2 在 `audit_post.go`（或 `post_admin.go`）实现 `approvePostByIDAdmin`：status 1/4/5 分支 CAS + `syncPublishedPostRedis`；不调 Green
- [x] 1.3 `api/v1/ucg_admin_http.go` 新增 `UcgAdminPostsApproveReq/Res`；`internal/controller/ucg_admin_api.go` 注册 `POST /ucg/admin/api/posts/approve`

## 2. 后端 API — 驳回理由必填

- [x] 2.1 `AdminBatchRejectPosts`：`reason` trim 后为空返回 400；移除默认「违规已下架」回退
- [x] 2.2 `rejectPostByIDAdmin` 调用方保证 reason 非空（admin 路径不再 fallback default）

## 3. Admin 静态页

- [x] 3.1 `ucg-admin.html`：状态筛选增加 4（发布失败）、5（机审失败）；表格增加「驳回原因」列
- [x] 3.2 增加「批量通过」按钮：confirm → `POST /posts/approve` → 展示 approved/skipped/failed 计数并刷新
- [x] 3.3 批量驳回改为 prompt 必填 reason，body 携带 `reason`；空理由不提交
- [x] 3.4 工具栏行纵向居中：`#panelPosts` 工具栏 scoped CSS（`align-items: center`），可选为 `.row` 增加 `posts-toolbar` class；不修改全局 `.row`

## 4. 验证

- [x] 4.1 `go build ./...` 通过
- [ ] 4.2 test 环境：待审帖批量通过 → Feed 可见；驳回须填 reason → App「我的动态」可见 reason；status=3 批过计入 failed
- [ ] 4.3 浏览器目视：动态审查工具栏 label/select/按钮/已选提示纵向居中对齐
