## Why

当前 `history-service` 已使用专属配置，但 `voice-service` 与 `device-service` 仍部分复用主配置，导致服务边界与配置边界不一致。为达到目标态微服务架构，需要完成“服务专属配置 + 主配置瘦身 + 跨服务契约化访问”的一致收敛，避免后续演进继续耦合。

## What Changes

- 为 `voice-service`、`device-service` 建立独立配置文件与独立加载入口，对齐 `history-service` 的专属配置模式。
- 清理主配置中不应由 gateway 或跨服务共用的领域配置项，保留仅网关与全局公共配置。
- 对齐各 service 启动参数、环境变量覆盖规则与默认值语义，确保部署行为可预测。
- 继续收敛微服务边界：跨服务数据访问必须走契约接口，不允许通过主配置回流到跨库直查。
- 更新 compose/kustomize 与运行文档，明确每个服务的配置来源、职责边界与迁移回滚策略。

## Capabilities

### New Capabilities
- `service-dedicated-config-loading`: 定义 voice/device/history/gateway 的专属配置加载规则与环境变量覆盖策略。
- `main-config-boundary-pruning`: 定义主配置可包含与禁止包含的配置域，约束主配置瘦身。
- `microservice-boundary-final-alignment`: 定义目标态微服务边界对齐要求（配置边界、运行边界、数据访问边界一致）。

### Modified Capabilities
- None.

## Impact

- 影响代码：`cmd/voice-service/main.go`、`cmd/device-service/main.go`、`cmd/history-service/main.go`、`internal/cmd/cmd.go` 及相关配置读取逻辑。
- 影响配置：`manifest/config/config.yaml`、新增 `config.voice-service.yaml`、`config.device-service.yaml`，以及 compose/kustomize 的 `GF_GCFG_FILE` 与覆盖环境变量。
- 影响文档：服务配置规范、部署手册、边界治理检查项。
- 影响运行：服务配置与职责隔离更清晰，减少跨服务隐式耦合与误配置风险。
