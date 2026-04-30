## 1. 基线盘点与迁移边界冻结

- [x] 1.1 盘点 `internal/service` 全部文件并建立归属矩阵（voice/device/history/shared），标记目标目录与目标包名。
- [x] 1.2 冻结迁移期间目录准入规则：新增实现文件不得进入 `internal/service`，共享目录需通过“无领域语义”审查。
- [x] 1.3 为每个批次定义完成标准（编译通过、关键服务可启动、关键链路可验证）与回滚点。

## 2. 分批迁移实现文件到 `internal/services/*`

- [x] 2.1 完成 history 批次迁移（`device_history*.go` 等）并修正引用路径与包语义。
- [x] 2.2 完成 device 批次迁移（`device_admin.go`、`device_profile_adapter.go` 等）并修正引用路径与包语义。
- [x] 2.3 完成 voice 第一批迁移（`voice_chat.go`、`voice_chat_deepseek.go`）并修正引用路径与包语义。
- [x] 2.4 完成 voice 第二批迁移（`voice_chat_understanding.go`、STT/TTS、WS 管理等）并修正引用路径与包语义。
- [x] 2.5 完成异步与共享能力迁移（producer/consumer/outbox/http deps/api types 等）并确保 shared 准入合规。

## 3. 迁移后引用收口与遗留清理

- [x] 3.1 全量替换调用路径，确保 `cmd`、`controller`、`service entry` 不再引用旧 `internal/service` 代码路径。
- [x] 3.2 清理 `internal/service` 中全部遗留实现文件，保留策略文件仅用于迁移说明（如需要）。
- [x] 3.3 增补治理文档与评审检查项，固化“目录边界=包边界=服务边界”与“禁止回流”规则。

## 4. 验证、回滚演练与验收

- [x] 4.1 每批迁移后执行编译校验（至少 `go test ./cmd/... ./internal/...`）并记录结果。
- [ ] 4.2 验证 gateway/voice/device/history 关键链路无回归，含 local/remote/canary 关键场景。
- [ ] 4.3 演练单服务维度回滚（如 voice 批次失败仅回滚 voice 相关迁移）并记录恢复步骤。
- [x] 4.4 收口验收：确认 `internal/service` 无遗留实现文件且迁移映射文档可追踪。
