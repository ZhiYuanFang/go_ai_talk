## 1. 全局约定文档

- [x] 1.1 在 `openspec/project.md` 新增「GoFrame 单行空结果查询约定（强制）」：推荐 `.One()` + `IsEmpty()`、禁止空集单结构体 `Scan` 当系统失败、列表/`Ensure` 后回读排除项、评审检查项
- [x] 1.2 在 `AGENTS.md` 增加短摘要并指向 `project.md` 对应章节（写法对齐既有 Redis 约定摘要）

## 2. UCG 高危修复

- [x] 2.1 将 `LoadLatestPendingProfileJob`（`audit_profile_job.go`）改为 `.One()` + `IsEmpty()`；确认 `mergeProfileForAuthor` 无 pending 时不再打「读取待审 job 失败」WARN
- [x] 2.2 将 `loadProfileAuditJob` / `loadPostForAudit` / `loadCommentForAudit` / `loadChatMessageForAudit`（`audit_moderation.go` 等）改为 One + IsEmpty；缺行返回零值并由 caller Ack/skip，禁止因 `ErrNoRows` requeue
- [x] 2.3 修复 `chat_persist.go` 中「会话最后一条 / MAX(id)」等可能空会话的单行 Scan
- [x] 2.4 修复 `oss_delete.go` 按 `object_key` 查 blob：无行返回业务约定的空/nil，而非内部 Scan 错误
- [x] 2.5 修复 `ai_quota.go` 用户 override 单行查询（admin/读路径）：无覆盖行成功回落默认

## 3. UCG 中危（吞错）修复

- [x] 3.1 `EnqueueProfileAuditJob` 内 pending `Scan`、`audit_moderation` 读 published profile、`profile.go` apply_failed/rejected 预览查询改为 One + IsEmpty，保留「无行则走分支」语义并暴露真 DB 错
- [x] 3.2 `force_store.go` / `ai_config.go` / `ai_quota.go` 等 `_ = Scan` 可选单行改为 One（与已有 `AddForceDelta` 风格一致）
- [x] 3.3 复核 `audit_moderation.go` CAS 后按 PK 再读：缺行按 missing 处理，不与「可选空集」混淆

## 4. Voice / Cash / Device / History / Simuser

- [x] 4.1 Voice：`care_alert_service` / `growth_trajectory_service` latest 加载、`ai_quota_store` override 等空集正常路径改为 One + IsEmpty
- [x] 4.2 Cash：`feeding_eligibility` 场景缺行回落默认、`apple_notify` 商品、`invite_group_qr`、`feature_invite` 锁行等改为 One 或显式业务 NotFound（禁止裸 ErrNoRows）
- [x] 4.3 Device：`feedback` 按 id、`profile_adapter` 按 device_no 等改为 One + IsEmpty / 明确业务错误
- [x] 4.4 History：`local.go` 最新/进行中 history 无行改为成功空语义
- [x] 4.5 Simuser：`sim_config` / `sim_prompt` / runtime 加载等改为 One + IsEmpty（缺行按既有默认或明确业务错误）

## 5. 验收与边界确认

- [x] 5.1 全仓复核：`rg '\.Scan\(&' internal/services`（及必要 controller）确认「可能无行」单结构体路径已处理；列表 `Scan(&[]T)` 与插入后回读可保留
- [x] 5.2 手工/日志验收：无 pending 资料 job 时 profile/发帖路径不再出现 `[ucg-profile] 读取待审 job 失败 … no rows`
- [x] 5.3 确认本变更 **不新增** App HTTP 路由（gateway-app / usage 统计 N/A）、**不新增** Redis 读缓存、**不新增** 背景循环任务
- [x] 5.4 （可选）若实现中样板过多，评估是否补充 `hack/check-gdb-empty-scan`；默认本 change 不做自动化门禁
