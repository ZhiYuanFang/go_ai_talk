## Why

当前架构是单进程单体，将 HTTP、WebSocket、设备管理、历史查询和 AI 交互编排耦合在同一个部署单元中。这会显著增加扩缩容与故障隔离难度，也不利于基于真实场景学习分布式与微服务实践。

## What Changes

- 将单体拆分为可独立部署的服务：gateway、device service、history service、voice-session service、async AI worker。
- 引入同步 API 与异步事件的服务契约，降低服务演进耦合度。
- 从单节点数据组件迁移到分布式模式：按域拥有数据库、用于会话/缓存共享状态的 Redis 集群、用于异步流程的消息队列。
- 通过 worker 池引入分布式计算能力，承接高耗时/长链路 AI 处理阶段。
- 对全部服务进行容器化，并基于 Kubernetes 部署，具备健康检查、滚动更新、自动扩缩容和可观测能力。

## Capabilities

### New Capabilities
- `service-decomposition-and-gateway`：定义服务边界、服务间契约以及 gateway 路由/鉴权聚合能力。
- `distributed-data-and-cache`：定义按域拥有的数据拓扑，以及 Redis 集群在共享临时状态和缓存中的使用方式。
- `async-processing-and-orchestration`：定义事件驱动处理、worker 执行模型及容器编排运行能力。

### Modified Capabilities
- 无。

## Impact

- 受影响代码：
  - `internal/cmd/`
  - `internal/controller/`
  - `internal/service/`
  - 新增 gateway/services/workers 多服务目录与部署清单。
- 受影响 API：
  - 现有 HTTP 与 WebSocket 入口将迁移到 gateway 与服务契约之后。
  - 服务间新增内部 API/事件通信。
- 受影响依赖/系统：
  - MySQL 拓扑演进为按域拥有的分布式数据形态。
  - Redis 演进为集群模式，承接会话/缓存状态。
  - 新增消息队列与 worker 运行时。
  - 需要新增 Kubernetes 部署与可观测体系。
