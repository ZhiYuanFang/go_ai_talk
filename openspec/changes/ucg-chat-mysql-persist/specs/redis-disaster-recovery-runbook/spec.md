## ADDED Requirements

### Requirement: 项目 SHALL 提供 Redis 容灾 runbook 文档

仓库 SHALL 包含 `docs/runbooks/redis-disaster-recovery.md`，说明本项目 Redis 在 Docker 环境下的重启、卷保留、AOF 持久化与数据丢失场景下的恢复步骤。文档 SHALL 使用中文说明性文本；命令、路径、环境变量名可保留英文。

#### Scenario: 运维查阅容器重启恢复

- **WHEN** 运维人员 Redis 容器 stop/start 或 `docker compose up --force-recreate`（未使用 `down -v`）
- **THEN** runbook SHALL 提供验证步骤（PING、CLUSTER INFO、DBSIZE）
- **AND** SHALL 说明无需重复 `cluster create`（生产 cluster 卷已有元数据时）

#### Scenario: 运维查阅 volume 丢失影响

- **WHEN** 运维人员执行 `down -v` 或 volume 损坏
- **THEN** runbook SHALL 列出按数据类型的可恢复性（MySQL 权威 vs 仅 Redis vs 可 lazy warm）
- **AND** SHALL 说明 UCG 私信在方案 A 下可从 MySQL 读时回填 Redis

### Requirement: runbook SHALL 说明本项目 Redis 持久化配置

文档 SHALL 描述测试 standalone 与生产 cluster 的 compose 文件路径、`--appendonly yes`、volume 名称（如 `redis-test-data`、`redis-node-*-data`），以及 AOF 文件位于容器 `/data` 目录。

#### Scenario: 查阅 AOF 与 volume 位置

- **WHEN** 运维需要确认持久化是否启用
- **THEN** runbook SHALL 指向 `manifest/docker/docker-compose.redis-standalone.test.yml` 与 `docker-compose.redis-cluster.yml` 中的相关配置

### Requirement: runbook SHALL 提供 volume 备份与还原指引

文档 SHALL 包含 Docker volume 备份示例（如 `tar` 打包 `/data`）与还原步骤，并注明备份前应尽量降低写入或接受 point-in-time 一致性限制。

#### Scenario: 计划内维护前备份

- **WHEN** 运维在维护窗口前备份 Redis 数据
- **THEN** runbook SHALL 提供可复制的备份命令示例（测试 standalone 与生产 cluster 节点）

### Requirement: runbook SHALL 与 release-deploy-and-run 交叉引用

`docs/runbooks/redis-disaster-recovery.md` SHALL 链接至 `docs/runbooks/release-deploy-and-run.md` 中 Redis 日常重启、cluster create 误报等章节，避免重复维护冲突步骤。

#### Scenario: 生产 cluster 报 Node is not empty

- **WHEN** 运维误执行 cluster create
- **THEN** runbook SHALL 说明先查 `cluster_state:ok` 并引用 release runbook 对应排障节
