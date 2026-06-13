## 1. 入队 diff

- [x] 1.1 `profile.go`：新增 `profileAuditPatchFromPublished`，对比 `ucg_profile` 生成 patch-only 三字段
- [x] 1.2 `UpdateMyProfile`：无实质变更时不 enqueue；有变更时仅传 patch 字段

## 2. Consumer 兜底

- [x] 2.1 `audit_moderation.go`：`runProfileGreenChecks` 跳过与已发布 profile 相同的非空 job 字段

## 3. 校验

- [x] 3.1 `go build ./...` 通过
