## MODIFIED Requirements

### Requirement: Comments API SHALL reflect audit pending and rejection

`POST .../posts/{id}/comments` 在 Green 完成前 MUST 返回 `status=1`（或等价 pending 字段）。`GET .../posts/{id}/comments` MUST 仅返回 `status=2`（published）评论给 **非作者** 视角；作者 MAY 看到自身 pending/rejected 评论及 reject_reason。

#### Scenario: 他人看不到待审评论

- **WHEN** 用户 A 发表评论且未审过，用户 B 拉取评论列表
- **THEN** 列表 MUST NOT 包含 A 的该条评论

#### Scenario: 作者见违规评论

- **WHEN** 用户 A 评论 Green fail
- **THEN** A 拉取评论或评论详情 MUST 可见 reject_reason

## ADDED Requirements

### Requirement: Profile me API SHALL read pending patch from audit job

`GET /ucg/app/api/profile/me` 对作者 MUST 合并 MySQL 最新 pending `ucg_profile_audit_job`（`auditPending=true`、预览 nickname/avatar/bio），MUST NOT 依赖 Redis profile pending 键作为长期权威。待审预览与 MQ/CAS 的版本语义 MUST 以 job 表 `audit_version` 为准，MUST NOT 从 `ucg_profile` 或 Redis 读取审核轮次。

#### Scenario: 待审头像预览

- **WHEN** 用户提交新 avatar 且 job pending
- **THEN** profile/me MUST 返回 `auditPending=true` 且 avatar 预览为新 key
