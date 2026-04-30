## 1. 基线与运行时规范

- [x] 1.1 定义并落地统一的 Redis key 命名空间与 TTL 规则，覆盖 voice/device/history 的缓存、守卫与幂等键。
- [x] 1.2 定义并落地统一的 RabbitMQ exchange、queue、routing-key 命名规范，覆盖跨服务事件。
- [x] 1.3 定义启动依赖检查契约（Redis 与 RabbitMQ 连通性/拓扑）以及发布失败阻断错误契约。
- [x] 1.4 设计并创建统一缓存适配层（`cache-kit`），暴露标准接口（读写、TTL、幂等键、错误语义）。
- [x] 1.5 设计并创建统一事件适配层（`event-kit`），暴露标准接口（发布、路由键约定、发布失败语义）。
- [x] 1.6 在 `cache-kit`/`event-kit` 中统一接入超时、日志、指标与追踪元数据透传能力。
- [x] 1.7 明确约束：业务服务不得直接调用底层 Redis/MQ 客户端，必须通过统一适配层访问。

## 2. 在现有服务中执行强依赖收敛

- [x] 2.1 删除 voice/device/history 运行时代码中的 Redis/memory 切换分支与内存兜底逻辑，并改为通过 `cache-kit` 调用 Redis。
- [x] 2.2 删除生产/消费链路中的 MQ enabled/disabled 分支与降级路径，并改为通过 `event-kit` 发布/消费消息。
- [x] 2.3 增加启动期 fail-fast 检查：Redis 或 RabbitMQ 检查失败时服务进程启动失败。
- [x] 2.4 更新必需事件发布路径：RabbitMQ 发布失败时返回阻断错误并中止请求。

## 3. 拆分 Voice 与 Device 服务

- [x] 3.1 创建/完善 `voice-service` 与 `device-service` 的独立入口与运行配置。
- [x] 3.2 定义并实现 gateway 到 voice/device 的内部调用契约。
- [x] 3.3 将 gateway 本地执行的 voice/device 领域处理逻辑迁移到对应独立服务。
- [x] 3.4 确保 voice/device/history 跨域数据访问通过服务契约或 RabbitMQ 事件完成。
- [x] 3.5 创建服务目录结构：`internal/services/voice`、`internal/services/device`、`internal/services/history`，并建立 `internal/shared` 公共目录。
- [x] 3.6 按服务归属迁移仅在单一服务使用的文件到对应目录，保持编译与运行行为不变。
- [x] 3.7 将跨服务复用文件收敛到公共目录，并建立准入规则（仅允许无领域语义的通用代码进入公共目录）。

## 4. gateway 收敛为流量与策略层

- [x] 4.1 建立端点归属矩阵，明确哪些端点委派给 voice/device/history 服务。
- [x] 4.2 在保持外部 API 兼容的前提下，移除 gateway 中的领域业务执行逻辑。
- [x] 4.3 保留并强化 gateway 横切能力（鉴权、路由、策略、元数据透传、流量控制）。
- [x] 4.4 完成目标态路由配置，使 gateway 全量委派到下游领域服务。

## 5. 禁测文件策略与运行时核验

- [x] 5.1 在本次迁移范围内删除仓库中的现有 `*_test.go` 文件。
- [x] 5.2 清理本地脚本/工作流中对测试文件存在性的依赖引用。
- [x] 5.3 建立运行时冒烟核验清单/脚本，覆盖启动检查、缓存读写、必需事件发布路径。

## 6. 灰度发布、可观测与回滚

- [x] 6.1 为 gateway/voice/device/history 增加或调整健康探针与依赖就绪检查。
- [x] 6.2 增加 Redis 依赖失败、RabbitMQ 发布失败与委派链路错误率的指标看板与告警。
- [x] 6.3 执行分阶段发布（local、canary、full），并按运行时清单验证委派与服务行为。
- [x] 6.4 文档化并演练依赖故障或错误率超阈值场景的回滚流程。
