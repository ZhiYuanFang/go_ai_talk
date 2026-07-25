## ADDED Requirements

### Requirement: 生产与测试主机职责分离

切流完成后，系统部署 MUST 满足：生产 Docker 栈与生产 MySQL 仅运行在新 4C8G ECS；测试 Docker 栈与 `ai_voice_*_test` 库仅留在旧 ECS。旧 ECS MUST NOT 再运行生产微服务 compose（可短期保留生产库只读回滚）。

#### Scenario: 切流后拓扑验收

- **WHEN** 运维完成 DNS/反代切流并执行验收
- **THEN** 生产域名请求解析到的服务 MUST 连接新机 MySQL 生产库（库名无 `_test` 后缀），且测试域名请求 MUST 仍连接旧机 `_test` 库

#### Scenario: 旧机停止生产栈

- **WHEN** 生产流量已切至新机
- **THEN** 旧机生产微服务 compose MUST 处于 down，以避免容器继续写入旧机 `ai_voice_*`

### Requirement: 生产库迁移范围与改址方式

生产切流 MUST 仅迁移 MySQL 生产库 `ai_voice_*`（含 history/device/voice/app/ucg/sim 等域约定库）。测试库 `ai_voice_*_test` MUST NOT 迁往新机。应用进程改连新库 MUST 通过更新 `.env.prod` 的 `MYSQL_TCP_HOST` 完成；`*_DB_LINK` MUST 继续使用 `mysql-host` 占位符与生产库名。

#### Scenario: 仅改 MYSQL_TCP_HOST

- **WHEN** 新机生产服务启动且 `MYSQL_TCP_HOST` 已设为新机可达地址
- **THEN** 启动日志 MUST 显示各 `database.*` 主机为该地址，且库名为对应 `ai_voice_*`（非 `_test`）

#### Scenario: 测试 env 不跟迁

- **WHEN** 生产已迁至新机
- **THEN** 旧机 `.env.test` 的 `MYSQL_TCP_HOST` MUST 仍指向旧机 MySQL，且库名 MUST 保持 `_test` 后缀

### Requirement: 新机 MySQL buffer 首日规格

新机生产 MySQL 在首日上线时 MUST 将 `innodb_buffer_pool_size` 配置为 **1G**。旧机（测试）MySQL MUST NOT 跟随调整到 1G，MUST 维持适合约 2G 主机的较小 buffer（例如约 256M 量级）。

#### Scenario: 新机 buffer 验收

- **WHEN** 新机 MySQL 完成配置并重启后执行 `SHOW VARIABLES LIKE 'innodb_buffer_pool_size'`
- **THEN** 返回值 MUST 为 1G（1073741824 字节）或运维文档等效写法对应的 1G

#### Scenario: 旧机 buffer 不跟调

- **WHEN** 生产切流完成
- **THEN** 旧机 MySQL `innodb_buffer_pool_size` MUST 仍为切流前测试适用规格，不得因新机升配而被改为 1G

### Requirement: 生产 Redis 空集群冷启

新机生产 Redis Cluster MUST 以空 volume 冷启：首次 `up` 后执行一次三主 `cluster create`，达到 `cluster_state:ok` 后再启动依赖 Redis 的生产微服务。MUST NOT 将旧机生产 Redis AOF/volume 拷贝为新机初始数据。切流 runbook MUST 声明冷启的用户可感知后果：App refresh token 失效需重新登录；语音/诊所会话上下文丢失；AI 月额度 Redis 计数归零；Feed 索引/快照需经读路径或写路径重建；UCG 私信正文以 MySQL 为权威不因冷启丢失。

#### Scenario: 新集群首次创建

- **WHEN** 新机首次初始化生产 Redis Cluster
- **THEN** 运维 MUST 在空节点上执行一次 `cluster create`，且 `CLUSTER INFO` MUST 显示 `cluster_state:ok`

#### Scenario: 冷启后私信仍可读

- **WHEN** Redis 为空且用户拉取已持久化到 MySQL 的会话历史
- **THEN** 系统 MUST 仍能返回私信正文（MySQL fallback / warm）

#### Scenario: 冷启后需重新登录

- **WHEN** 用户持有切流前签发的 refresh token 访问新机生产环境
- **THEN** 该 refresh token MUST 不再有效，用户 MUST 通过登录流程获取新凭证

### Requirement: 换机 runbook 与档位文档

仓库 MUST 提供可执行的生产换机切流说明（含准备、dump/restore、Redis 冷启、停写、DNS 切流、验收、回滚、旧库保留后再清理），并 MUST 在内存档位文档中记录 **4C8G / 单 prod + 本机 MySQL** 的建议，其中 MySQL buffer 起点为 1G。

#### Scenario: runbook 可独立执行

- **WHEN** 运维按仓库 runbook 执行生产换机
- **THEN** 文档 MUST 覆盖新机 MySQL buffer=1G、仅迁生产库、`MYSQL_TCP_HOST` 改址、Redis 空集群冷启、旧机 test 留守与验收命令

#### Scenario: 档位文档含 4C8G

- **WHEN** 查阅 `memory-sizing-guide`（或等价档位文档）
- **THEN** 文档 MUST 包含 4C8G prod-only 建议，并写明新机 `innodb_buffer_pool_size` 首日 1G
