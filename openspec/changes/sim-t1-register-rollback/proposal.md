## Why

T1 注册任务在 `simRegister` 成功后若昵称/头像 AI、媒体上传或 profile 写入任一步失败，当前实现仅 `RecordTaskRun(false)` 并 return，已创建的 `wx`（及 `sim_wx_credential`）仍保留，导致 sim 管理页出现无昵称/头像的「裸号」，并占用 `maxSimUsers` 名额。需在 profile 完整落地前将失败视为「未注册成功」，并回滚已创建的 sim wx。

## What Changes

- T1（`RunRegisterTask`）在 **profile PUT 成功之前** 任一步失败时，MUST 调用已有 device internal sim 注销回滚 wx，并清理 sim 侧 credential（若已写入）。
- 将 `InsertWxCredential` 调整为 **profile 写入成功之后** 再落库，避免失败路径留下凭据脏数据。
- 昵称 AI 返回空串、头像 AI/上传/profile 失败与 nickname/avatar AI 错误同等对待，均触发回滚。
- 回滚失败时 MUST 记录 warning 日志，任务仍记为失败；不新增 Admin/App HTTP 接口。
- 手动触发 T1（sim-admin「执行」）与定时调度共用同一 rollback 语义。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `sim-user-service`：T1 注册任务增加「profile 完成前失败须回滚 wx」的原子性要求；调整 credential 写入时机。

## Impact

- **进程**：仅 `sim-user-service`（`RunRegisterTask` 及 sim 内部 rollback helper）。
- **依赖**：复用已实现 `POST /device/internal/api/sim/wx/{wxId}/deactivate` 与 `DeleteWxCredentialByWxID`；无新跨服务契约。
- **数据库**：无新表；失败路径减少 `sim_wx_credential` 脏行。
- **OpenSpec 基线**：增量修改 `sim-user-service` 中 T1 Requirement（与 `sim-admin-user-list-deregister` 变更后语义对齐）。
- **Usage 统计 / Redis**：无变更。
