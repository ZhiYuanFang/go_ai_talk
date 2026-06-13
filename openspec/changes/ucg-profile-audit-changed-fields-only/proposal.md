# 提案：资料审核仅提交相对已发布 profile 的变更字段

## 问题

App PUT `/profile/me` 常携带全量 nickname/avatarKey/bio。服务端将请求中非空字段全部写入 `ucg_profile_audit_job` 并机审，导致用户仅改简介时仍对未改昵称调用 Green（`nickname_detection`），放大 MQ 重试时的 API 费用。

## 方案

- 入队前对比 **`ucg_profile` 已发布行**（非 pending job 预览），仅将**相对已发布值有变化**且请求非空的字段写入 audit job。
- 若三者均无实质变更：不 enqueue、不调 Green，返回当前作者视图。
- Consumer Phase1 兜底：机审前再次跳过与已发布 profile 相同的 job 字段（兼容 MQ 积压旧消息）。

## 范围

- `UpdateMyProfile` / `EnqueueProfileAuditJob` 入参语义不变（空字段仍表示「本次不提交该字段」）。
- 不修改 App 协议；不新增 Redis。
