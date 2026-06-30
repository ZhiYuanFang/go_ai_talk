## 1. sim-user-service：rollback 与 T1 重构

- [x] 1.1 新增 `rollbackSimRegistration(ctx, wxID)`：device sim deactivate + `DeleteWxCredentialByWxID`，失败打 warning
- [x] 1.2 重构 `RunRegisterTask`：`defer` 在未 commit 且 `wxID>0` 时 rollback；成功边界为 profile PUT 之后
- [x] 1.3 将 `InsertWxCredential` 移至 profile PUT 成功之后；移除注册后立即写入
- [x] 1.4 昵称 trim 为空、头像 URL 为空视为失败并走 rollback

## 2. 验收

- [x] 2.1 `go build ./...` 通过
- [x] 2.2 代码评审：`RunRegisterTask` 任一步失败路径均触发 rollback（含 login/Prompt/upload/profile）
- [x] 2.3 确认无新 HTTP 路由与 usage 统计变更
