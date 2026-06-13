## 1. 资料审核 Phase1 收口

- [x] 1.1 `runProfileModerationPhase`：Green err / persist err 后 `return`，不再继续 `persistModerationVerdictProfile`
- [x] 1.2 `markProfileModerationFailed`：增加 `WHERE status=pending` CAS；更新 `audit_moderation.go` 顶部路径 A 注释
- [x] 1.3 确认 `auditProfileJobFromEvent` Phase1 永不向上抛 err；stale skip 覆盖 status=5

## 2. 帖子 Green 单次对齐

- [x] 2.1 `constants.go` 增加 `PostStatusModerationFailed = 5`
- [x] 2.2 实现 `markPostModerationFailed`；`runPostModerationPhase` 改 void 语义（Green/persist err → mark，不 return err）
- [x] 2.3 `auditPostFromEvent` 与 profile 对称：Phase1 后不 requeue Green

## 3. 评论 / 私信两阶段化（DDL + Consumer）

- [x] 3.1 编写并执行 migration：`ucg_post_comment`、`ucg_chat_message` 增加 `moderation_verdict/reason/at`、`apply_attempts`、`apply_failed_at`；同步 dao/entity
- [x] 3.2 实现 `persistModerationVerdictComment/Chat`、`runCommentModerationPhase/Apply`、`runChatModerationPhase/Apply`（复用 `audit_moderation.go` 模式）
- [x] 3.3 重构 `audit_comment.go`、`audit_chat.go` 为两阶段入口；Green API err → 机审失败终态 + Ack
- [x] 3.4 Chat Redis 竞态：明确 MySQL 未就绪时不调 Green 的策略并在代码注释

## 4. Admin API（资料机审失败）

- [x] 4.1 `api/v1/ucg_admin_http.go` 增加 list/resolve 请求响应结构
- [x] 4.2 `internal/services/ucg` 实现 `ListProfileAuditJobsForAdmin`、`ResolveProfileAuditJobAdmin`（approve/reject CAS 含 status=5）
- [x] 4.3 `internal/controller/ucg_admin_api.go` 注册 handler；确认路由经 gateway-app 可达
- [x] 4.4 确认新接口属 `/ucg/admin/api/*`，**不**写入 `usagestats` maintenance_skip（维护型 admin，无需统计）

## 5. ucg-admin.html UI

- [x] 5.1 新增「资料机审失败」Tab：列表表格 + 通过/驳回交互
- [x] 5.2 对接 list/resolve API；操作后刷新；错误 toast

## 6. 文档与校验

- [x] 6.1 `docs/runbooks/release-deploy-and-run.md` 增补：moderation_failed 识别、管理页人工处理、Green 单次语义
- [x] 6.2 `openspec validate ucg-audit-green-once-admin --strict`
