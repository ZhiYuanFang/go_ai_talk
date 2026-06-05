## ADDED Requirements

### Requirement: 生产与测试 Compose SHALL 定义容器 CPU 与内存上限

仓库 SHALL 提供 **`manifest/docker/docker-compose.resources.prod.yml`** 与 **`manifest/docker/docker-compose.resources.test.yml`**（或后继等价 overlay），为下列组件定义 **`mem_limit`** 与 **`cpus`**（或 compose 规范中等价、在非 Swarm 模式下对 `docker compose up` 生效的字段）：

- 生产/测试 **全部** 微服务（gateway、gateway-app、history、voice、device、worker、ucg）
- 生产 Redis、测试 Redis、生产/测试 RabbitMQ

runbook SHALL 文档化默认配额表及「2G ECS survival profile」说明。`voice-service` 测试实例 SHALL 拥有 **不低于** 其它微服务的 memory limit（ documented 起步值 **512M**）。

#### Scenario: 启动命令叠加 resources overlay

- **WHEN** 运维按 runbook 启动生产微服务并叠加 `-f docker-compose.resources.prod.yml`
- **THEN** `docker inspect` 或 `docker stats` SHALL 显示对应容器配置了 memory/cpu 上限

#### Scenario: 本地开发不受 prod/test limits 强制约束

- **WHEN** 开发者仅使用基线 `microservices.yml` + `microservices.local.yml` 且 **不** 叠加 `resources.*.yml`
- **THEN** 本地容器 **MAY** 无 cgroup 上限（便于调试）

### Requirement: limits SHALL 防止单容器耗尽宿主机

资源上限的配置意图 SHALL 在 runbook 中说明：当某容器内存超过 `mem_limit` 时，内核 **MAY** OOM kill 该容器，**SHALL NOT** 无限制占用同机其它栈（含 MySQL 宿主机进程）的全部物理内存。runbook SHALL 包含 OOM 排查步骤（`dmesg`、`docker stats`、调高 voice-test limit 等）。

#### Scenario: 文档化 OOM 语义

- **WHEN** 运维查阅 `release-deploy-and-run.md` 资源 limits 章节
- **THEN** 文档 SHALL 说明 limits 与宿主机 2G 物理内存的关系，以及 ASR 验收时优先保障 test voice 的建议
