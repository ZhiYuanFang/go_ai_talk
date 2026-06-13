# 提案：UCG 审核 Green 单次调用 + 资料机审失败人工复核

## Why

资料审核已引入两阶段机审与 `moderation_verdict`，但帖子/评论/私信仍可能在 MQ 重投时重复调用 Green；资料侧 Green 或写 verdict 失败时虽可落入 `ProfileJobStatusModerationFailed(5)`，作者侧无感知且运维无法在管理页处理，存在内容长期不落库与合规盲区。需要在全审核链路统一「Green 至多一次」语义，并为资料机审失败 job 提供 UCG 管理端人工通过/驳回能力。

## What Changes

- **全实体 Green 单次**：profile / post / comment / chat 对每个 `(entity_id, audit_version)`，一旦 Phase1 已发起 Green 请求（含 API 错误路径），后续 MQ 消费 MUST NOT 再次调用 Green；以 `moderation_verdict≠0` 或机审失败终态（如 `moderation_failed`）为跳过依据。
- **资料 Phase1 收口**：修复 Green 失败后仍尝试 persist、以及 Phase1 仍可能 requeue 的遗留逻辑；Green/`persist verdict` 失败 → `status=moderation_failed`，Ack MQ，不调 Phase2。
- **帖子 Phase1 对齐资料**：Green 或 persist verdict 失败不再 `return err` 无限 requeue；新增帖子侧机审失败终态（与 profile 对称）。
- **评论/私信两阶段化**：为 `ucg_post_comment`、`ucg_chat_message` 增加与 post 同族的 `moderation_verdict` / apply 字段；Consumer 拆 Phase1 Green + Phase2 CAS，Green 单次语义与 profile 一致。
- **作者侧**：`moderation_failed` job 不对 App 作者展示专用文案或审核中态（作者仍见已发布 profile）；与人工复核队列分离。
- **UCG 管理页**：新增「资料机审失败」Tab/模块，列表展示 `status=moderation_failed` 的 `ucg_profile_audit_job`（含 wxId、patch 字段、reject_reason、时间）；提供单条「通过」「驳回」人工操作（CAS 迁移至 approved/rejected 并更新 `ucg_profile` 或写入驳回原因）。
- **Admin API**：`GET /ucg/admin/api/profile-audit-jobs/list`、`POST /ucg/admin/api/profile-audit-jobs/resolve`（Header `X-Admin-Password`）；路径属 `/ucg/admin/api/*`，**不计入** App usage 统计（维护型 admin API）。

## Capabilities

### New Capabilities

- `ucg-admin-profile-moderation`：UCG 管理页资料机审失败列表与人工通过/驳回 API 及静态页 UI。

### Modified Capabilities

- `ucg-green-audit`：全 UGC 审核实体 Green 单次、两阶段机审、机审失败终态与作者不可见语义；收紧 comment/chat 由单阶段改为两阶段。
- `ucg-admin-post-moderation`：ucg-admin.html Tab 结构扩展说明（新增资料机审 Tab，与动态审查并列）。

## Impact

- **进程**：`ucg-service`（MQ consumer、`audit_moderation.go`、comment/chat 审核、admin service）。
- **网关**：`gateway-app-server` 注册新 admin HTTP 路由（经现有 UCG admin 控制器）。
- **数据库**：`ai_voice_ucg` — comment/chat 表 DDL（moderation/apply 字段）；post 可选 `status=5 moderation_failed` 或等价标记。
- **前端**：`resource/public/ucg-admin.html`。
- **API**：`api/v1/ucg_admin_http.go` 新增 profile audit job 管理接口。
- **规格基线**：变更 `ucg-green-audit`（v2.0.3 §ucg-green-audit）、扩展 `ucg-admin-post-moderation` 管理页范围。
- **Out of scope**：帖子/评论/私信 `moderation_failed` 的管理页（本变更仅 profile job）；Green API 参数与 pass/fail 判定逻辑不变；不新增 `*_test.go`。
