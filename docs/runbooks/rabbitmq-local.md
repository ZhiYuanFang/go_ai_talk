## Local RabbitMQ Baseline

用于 `Task 5.1`：搭建本地消息队列基础设施并完成 topic/exchange 基线配置。

### 文件位置

- compose: `manifest/docker/docker-compose.rabbitmq.yml`
- init script: `hack/rabbitmq-init.ps1`

### 启动与初始化

1. 启动 RabbitMQ 并初始化交换机/队列绑定：
   - `powershell -ExecutionPolicy Bypass -File "hack/rabbitmq-init.ps1"`
2. 打开管理台：
   - [http://127.0.0.1:15672](http://127.0.0.1:15672)
   - 用户名/密码：`guest` / `guest`

### 基线拓扑

- Exchange:
  - `voice.events`（type=`topic`, durable）
- Queues:
  - `voice.task.requested.q`（`voice.task.requested`）
  - `voice.task.completed.q`（`voice.task.completed`）
  - `voice.task.failed.q`（`voice.task.failed`）
  - `notify.events.q`（`notify.*`）
  - `history.events.q`（`history.#`）

### 停止与清理

- 停止：
  - `docker compose -f manifest/docker/docker-compose.rabbitmq.yml down`
- 删除数据卷（重置）：
  - `docker compose -f manifest/docker/docker-compose.rabbitmq.yml down -v`

### 验收清单

- [ ] RabbitMQ 管理台可访问
- [ ] `voice.events` exchange 创建成功
- [ ] 5 个基线队列创建成功并完成绑定
- [ ] 能在管理台观察到基础路由键流转
