## ADDED Requirements

### Requirement: Runbook SHALL 文档化 2G 双栈 survival 配置

`docs/runbooks/release-deploy-and-run.md` SHALL 包含 **2G ECS**（或 documented 低内存同机双栈）专节，至少包括：

- 生产 Redis **3 主 0 从** 与测试 Redis **单机** 拓扑及迁移步骤（含 `down -v` 与数据丢失说明）
- prod/test 启动命令叠加 **`docker-compose.resources.{prod,test}.yml`**
- 默认 **mem_limit / cpus** 对照表与 `docker stats` 验收
- MySQL 同机 **`innodb_buffer_pool_size`** 建议（如 256M 级）
- ASR 验收约定：生产微服务保持 Up，**仅 test 域名** 进行语音压测，避免 prod 并发 ASR
- OOM / swap 排错

#### Scenario: 运维按 2G 文档完成测试 Redis 迁移

- **WHEN** 运维阅读 runbook 2G 专节并按步骤从六节点 test cluster 迁到 standalone
- **THEN** 其 SHALL 能完成 standalone Redis 启动、`.env.test` 更新与微服务 recreate，且 **无需** `cluster create`

#### Scenario: 运维按文档叠加资源 limits 启动双栈

- **WHEN** 运维按 runbook 生产/测试启动命令启动双栈
- **THEN** 命令示例 SHALL 包含 `-f docker-compose.resources.prod.yml` 或 test 等价文件
