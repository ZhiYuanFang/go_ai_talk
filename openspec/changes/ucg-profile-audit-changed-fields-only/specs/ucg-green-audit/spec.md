## MODIFIED Requirements

### Requirement: Profile patch SHALL use MySQL audit job and MQ instead of Redis pending queue

资料变更（nickname/avatar/bio）MUST 写入 MySQL `ucg_profile_audit_job` 并 Publish `ucg.profile.patch.submitted`。入队前 MUST 将请求字段与 **`ucg_profile` 已发布行**对比，**仅**将相对已发布值发生变化的非空字段写入 job；若 nickname、avatarKey、bio 均无实质变更，MUST NOT enqueue 且 MUST NOT 触发 Green。Consumer 机审时 SHOULD 跳过 job 中与已发布 profile 相同的字段（兼容历史全量 job 消息）。

#### Scenario: 仅 bio 变更

- **WHEN** 作者 PUT profile 携带未改 nickname 与新 bio
- **THEN** job MUST 仅含 bio 非空；Green MUST NOT 调用 `nickname_detection` 审该 nickname

#### Scenario: 全量相同无变更

- **WHEN** PUT 三字段均与已发布 profile 相同
- **THEN** MUST NOT 创建新 audit MQ 消息；HTTP MAY 返回 200 与当前作者 profile
