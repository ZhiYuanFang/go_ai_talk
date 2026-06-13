# 设计

## Diff 基准

以 `ucg_profile` 当前已发布列值为准（公众所见），**不**以 pending job 或 merge 后的作者 DTO 为准。

## 入队规则

对每个请求字段（trim 后）：

| 条件 | 写入 job |
|------|----------|
| 请求为空 | 不写（空串） |
| 请求非空且等于已发布值 | 不写 |
| 请求非空且不同于已发布值 | 写入 patch 值 |

全无 patch → 直接返回，不调用 `EnqueueProfileAuditJob`。

## Consumer 兜底

`runProfileGreenChecks` 读取 `ucg_profile`，若 `job.Nickname/Bio/AvatarKey` 与已发布相同则跳过对应 Green 调用（积压全量 job 消息时仍省 API）。

## 与 apply 一致

`approveProfileJobCAS` 本就「job 字段非空才 UPDATE profile」，patch-only job 行为不变。
