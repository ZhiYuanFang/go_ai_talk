## Why

当前 `worker-service` 仍使用主配置 `manifest/config/config.yaml`，而该主配置历史上承载了数据库信息，已经与“服务独立配置、服务独占数据库”的微服务边界目标不一致。随着 `gateway-service` 已确认不再访问数据库，主配置中的数据库项应被彻底移除，避免错误依赖与跨服务耦合继续存在。

## What Changes

- 为 `worker-service` 新增专属配置文件，并将 worker 默认配置加载与部署入口统一切换到专属配置。
- 从主配置 `manifest/config/config.yaml` 中删除数据库配置，仅保留网关与全局公共配置项。
- 移除 `internal/dao/domain_db.go` 多数据库组回退机制，统一为“服务进程只使用本服务 default 数据库”。
- 梳理并收敛 DAO 扩展层（`*_ext.go`）职责，仅保留确有业务必要的扩展；对无实际增量价值的扩展进行清理或回并。
- 更新并重构运维文档：将 `dao-sync-by-domain.md` 与 `release-deploy-and-run.md` 迁移到新的统一目录，并按当前微服务原则改写。
- 在 `openspec/project.md` 增加常驻约定：  
  - 一个服务对应一个数据库；  
  - 跨服务数据通过 API 获取，不跨库直查；  
  - 各服务配置仅维护 `database.default`。

## Capabilities

### New Capabilities

- `worker-dedicated-config-loading`: 规范 worker-service 专属配置加载、部署入口与回滚方式。
- `main-config-without-database`: 规范主配置文件不再承载数据库连接信息。
- `single-default-db-per-service`: 规范每个服务仅使用本服务 `database.default`，移除多数据库组分流能力。
- `dao-extension-layer-simplification`: 规范 DAO 扩展层清理准则与最小保留面。
- `runtime-docs-centralization-and-governance`: 规范运行/发布文档集中管理与变更同步要求。

### Modified Capabilities

- 无

## Impact

- 受影响代码：`cmd/worker-service/**`、`manifest/config/**`、`manifest/docker/**`、`manifest/deploy/kustomize/**`、`internal/dao/**`、`internal/services/async/**`。
- 受影响文档：`docs/dao-sync-by-domain.md`、`docs/release-deploy-and-run.md`、`openspec/project.md`。
- 运行影响：worker 配置来源变化；主配置瘦身；DAO 访问路径简化。
- 风险点：若 worker 专属配置数据库未对齐 outbox 所在库，可能导致 outbox relay 无法正确处理事件。
