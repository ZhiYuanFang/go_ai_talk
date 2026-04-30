## Why

当前仓库虽然已建立 `internal/services/voice|device|history` 目录，但主要业务实现仍集中在 `internal/service`，导致“目录边界”与“代码边界”不一致。现在需要执行一次完整迁移，消除遗留文件，确保后续服务演进不再回到统一 service 包。

## What Changes

- 将 `internal/service` 中全部业务实现按领域归属迁移到 `internal/services/*`，不保留遗留实现文件。
- 建立 `internal/shared` 的准入规则，仅允许无领域语义的公共代码进入共享目录。
- 调整 import/调用入口，使 gateway、voice-service、device-service、history-service 均引用迁移后路径。
- 增加迁移分批策略与回滚方案，确保每批迁移后可编译、可启动、可验证。
- **BREAKING**: 统一 `internal/service` 包路径将被拆解，内部引用路径与包名语义将发生变化。

## Capabilities

### New Capabilities
- `service-code-full-cutover`: 定义 `internal/service` 全量迁移到 `internal/services/*` 的目标态与不留遗留文件约束。
- `domain-package-boundary-enforcement`: 定义领域包命名、依赖方向与 `internal/shared` 准入边界，防止迁移后回流。
- `service-migration-safety-and-rollback`: 定义分批迁移的编译校验、运行验证与按服务回滚要求。

### Modified Capabilities
- None.

## Impact

- 影响代码：`internal/service/*.go`、`internal/services/**`、`internal/shared/**`、各服务入口与控制器中的导入路径。
- 影响构建：Go 包引用关系会调整，需保证 `go test ./cmd/... ./internal/...` 持续通过。
- 影响运行：服务启动入口与跨服务契约路径保持不变，但内部实现归属将完全按领域拆分。
