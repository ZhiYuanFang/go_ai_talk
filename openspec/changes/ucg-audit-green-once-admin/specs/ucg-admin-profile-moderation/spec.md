## ADDED Requirements

### Requirement: UCG Admin SHALL list profile audit jobs in moderation_failed state

ucg-service MUST 提供 admin HTTP API（Header `X-Admin-Password` 与现有 UCG admin 一致）：

- `GET /ucg/admin/api/profile-audit-jobs/list` — 分页查询 `ucg_profile_audit_job`；默认筛选 `status=ProfileJobStatusModerationFailed(5)`；可选 query 覆盖 status/page/pageSize。

响应 MUST 含：jobId、wxId、auditVersion、nickname/avatarKey/bio（job patch 字段）、rejectReason（机审失败日志）、createdAt、updatedAt。

#### Scenario: 列表默认仅机审失败

- **WHEN** 管理员 GET list 且未传 status
- **THEN** 响应 SHALL 仅包含 `status=5` 的 job 行

#### Scenario: 未授权拒绝

- **WHEN** 请求缺少或错误 `X-Admin-Password`
- **THEN** API SHALL 返回未授权错误且 MUST NOT 返回 job 数据

### Requirement: UCG Admin SHALL manually resolve moderation_failed profile jobs

`POST /ucg/admin/api/profile-audit-jobs/resolve` MUST 接受 `jobId`（必填）、`action`（`approve` | `reject`）、`reason`（reject 时必填）。

- **approve**：CAS 将 job 从 `moderation_failed(5)` 转为 `approved(2)`；MUST 按 job 非空 patch 字段更新 `ucg_profile`（与 `approveProfileJobCAS` 语义一致）；SHOULD 补写 `moderation_verdict=pass` 与 `moderation_at`。
- **reject**：CAS 将 job 转为 `rejected(3)` 并写入 `reject_reason`；MUST NOT 更新已发布 profile。

操作 MUST 使用 CAS（`status=5` + 匹配 `audit_version`），并发重复请求 MUST 幂等（0 行 affected 返回成功或明确 skip）。

#### Scenario: 人工通过应用 patch

- **WHEN** 管理员对 status=5 的 job 执行 approve 且 job.bio 非空
- **THEN** job SHALL 变为 approved，且 `ucg_profile.bio` SHALL 更新为 job 中 bio

#### Scenario: 人工驳回

- **WHEN** 管理员对 status=5 的 job 执行 reject 并填写 reason
- **THEN** job SHALL 变为 rejected 且 reject_reason SHALL 等于所填 reason，且已发布 profile MUST 不变

#### Scenario: 非 moderation_failed 拒绝

- **WHEN** 管理员对 status≠5 的 job 调用 resolve
- **THEN** API SHALL 返回参数/状态错误且 MUST NOT 变更 job 或 profile

### Requirement: ucg-admin.html SHALL provide profile moderation_failed review UI

静态页 MUST 在「资料机审失败」Tab 内：

- 调用 list API 展示表格（wxId、patch 摘要、失败原因、时间）；
- 每行提供「通过」「驳回」按钮；驳回 MUST 弹窗收集 reason；
- 操作成功后 MUST 刷新当前页列表。

#### Scenario: 通过后列表减少

- **WHEN** 管理员点击某行「通过」且 API 成功
- **THEN** 该行 MUST 从当前列表消失（或刷新后不再出现）
