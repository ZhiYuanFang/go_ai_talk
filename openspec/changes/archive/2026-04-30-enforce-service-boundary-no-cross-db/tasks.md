## 1. History 内部 HTTP 契约落地

- [x] 1.1 梳理并确认 voice 所需的 history 查询/写入接口清单（历史记录、生日、事件选项、建议相关）。
- [x] 1.2 在 history-service 中补齐对应内部接口与统一错误结构，保证参数校验与错误码一致。
- [x] 1.3 更新 `internal/services/contracts/http_targets.go` 与相关文档，固化 history 内部契约路径与环境变量约定。

## 2. Device 画像契约补齐

- [x] 2.1 梳理 voice 所需的设备画像/注册状态字段，并定义 device 内部查询接口清单。
- [x] 2.2 在 device-service 中补齐画像查询接口与统一错误结构，保证参数校验与错误码一致。
- [x] 2.3 为 device 内部契约补充目标地址与路径约定，更新运行文档。

## 3. Voice 侧远程适配与边界收敛

- [x] 3.1 实现 `historyRemoteClient` 各方法的真实 HTTP 调用，移除 `not implemented yet` 占位行为。
- [x] 3.2 将 `voice-service` 中直接访问 `dao.History` 的路径替换为 `DeviceHistoryContract` 调用。
- [x] 3.3 将 `voice-service` 中直接访问 `dao.User` / 设备画像数据的路径替换为 device 契约调用。
- [x] 3.4 为 history/device remote 调用补充超时、错误映射与可观测日志，确保故障可定位。

## 4. 迁移与回退策略

- [ ] 4.1 保持 history/device 的 `local|remote|canary` 切换可用，并校验同分流键稳定命中。
- [ ] 4.2 实现并验证 history/device 远程失败时的 failover 回退语义（仅迁移期启用）。
- [x] 4.3 更新 compose 与 kustomize 示例配置，明确各环境默认模式与切流建议。

## 5. 验证与治理收口

- [ ] 5.1 进行端到端验证：voice 查询历史与设备画像在 remote/canary 模式下结果与 local 一致。
- [ ] 5.2 验证 history/device 服务不可达时的错误语义与 failover 行为符合预期。
- [x] 5.3 增补架构治理文档与评审检查项，明确“跨服务走接口，不跨库直查”为强制约束。
