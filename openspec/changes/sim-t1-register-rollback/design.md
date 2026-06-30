## Context

- T1 当前顺序：`simRegister` → `InsertWxCredential` → `usernameLogin` → AI 昵称/头像 → upload → `PUT profile/me`；中间任一步失败仅 `return`，wx 仍计入 sim 用户数。
- 已有 `DeactivateSimUserForAdmin`（device internal sim deactivate + `DeleteWxCredentialByWxID`），可复用于失败补偿。
- wx 必须先存在才能 login 与上传头像，无法延后 wx 创建；只能 **失败时回滚**。

## Goals / Non-Goals

**Goals:**

- profile（nickname + avatarKey）写入成功前，任一步失败视为 T1 未成功。
- 失败时删除 `is_simulated=1` 的 wx，不占用 `maxSimUsers`；sim 库不保留 credential 脏行。
- 昵称 AI 空串、头像 AI 失败、upload/profile 失败与 AI 错误同等触发回滚。
- 手动 T1 与 scheduler 行为一致。

**Non-Goals:**

- 不清理 UCG 侧已创建的 profile 行（正常失败路径在 PUT 前不应有 profile；若 PUT 部分成功需单独评估，首期以 PUT 失败即 rollback wx 为准）。
- 不回收 OSS 孤儿媒体（sim 场景可接受）。
- 不批量清理历史裸号；不新增 HTTP API。

## Decisions

### 1. 成功边界

**注册成功** = `PUT /ucg/app/api/profile/me` 返回成功（含非空 nickname 与 avatarKey 已通过校验写入）。

此前步骤失败（含 `usernameLogin`、Prompt 加载、AI、upload）均触发 rollback。

### 2. credential 写入时机

**从「wx 注册后立即写」改为「profile PUT 成功之后写」。**

理由：失败路径无需删 credential；admin 列表与 T2–T6 仅对完整注册用户暴露凭据。

同步 **MODIFIED** `sim-admin-user-list-deregister` 中「注册后立即 INSERT credential」的规格表述（本变更 delta 覆盖）。

### 3. rollback 实现

新增 `rollbackSimRegistration(ctx, wxID int64)`（sim-user 内部）：

1. 若 `wxID <= 0` 则 no-op
2. `deviceInternalPost` → `/device/internal/api/sim/wx/{wxId}/deactivate`
3. `DeleteWxCredentialByWxID`（幂等）
4. 注销失败打 `[simuser] rollback failed wxId=... err=...` warning，不 panic

`RunRegisterTask` 结构：

```text
wxID := 0
committed := false
defer: if !committed && wxID>0 { rollbackSimRegistration }

simRegister → wxID
login → AI → upload → profile PUT
InsertWxCredential
committed = true
RecordTaskRun(success)
```

任一步 error：`RecordTaskRun(false)` + return（defer 执行 rollback）。

### 4. 空昵称 / 空头像

- `strings.TrimSpace(nickname) == ""`  after AI → 视为失败，不 PUT profile。
- `imgRes.URL` 为空 → 视为失败。

### 5. 与 admin deactivate 复用

rollback 与 admin 注销共用 device deactivate 路径；rollback **不** skip video job（新注册用户无 pending job）。可抽 `deactivateSimWxInternal(ctx, wxID, skipVideo bool)` 或 rollback 直接调 device POST + DeleteCredential。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| rollback 调用 device 失败，裸 wx 残留 | warning 日志 + 任务 fail；运维可 sim-admin 手动注销 |
| PUT 成功但 InsertCredential 失败 | defer 仍 rollback wx（profile 行可能残留，极少见）；任务 fail |
| device 短暂不可用 | T1 fail，不占名额；下次 tick 重试 |

## Migration Plan

1. 部署 sim-user-service 新版本即可。
2. 历史裸号不自动清理；可选运维手动注销。
3. 无配置/DB 迁移。

## Open Questions

- 无。
