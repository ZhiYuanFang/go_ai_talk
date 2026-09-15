## Why

预测事项的时间点目前只存在于客户端；App 离线时用户无法在「事件即将发生」时被提醒。需要服务端记住宝宝待发生节点，并在约定提前量（5 分钟）通过系统推送叫醒所有绑定该宝宝的用户，且客户端每次刷新预测后服务端闹钟必须立刻跟上，不能停在旧的未发送时间点。

## What Changes

- 新增 App API：客户端在预测更新后，将「有确定下次发生时间」的事件全量上报；服务端按宝宝（`deviceNo`）**整表替换** Redis 待办列表（不落库；无 TTL，靠替换维护）。
- 写入时按 `nextAt - 5min` 向 RabbitMQ **延时交换机**投递叫醒消息（`rabbitmq_delayed_message_exchange` + `x-delay` 毫秒）；消费宿主为 **voice-service**。
- 到点后：再读 Redis 校验是否仍匹配 → 5 分钟推送去重 → 查 history 近 30 分钟是否已有同类记录 → 向绑定该宝宝的全部 `wxId` 发系统推送。
- device 域新增 `ListWxIdsByDeviceNo`（同 `device_no` 全部 wx），替代现有 LIMIT 1 反查。
- 抽离/扩展**公共推送**能力并新增业务类型（非 UCG 社区角标语义）；voice 经契约调用，禁止跨包 import。
- Consumer **一期对业务失败一律 Ack**：推送失败可接受；禁止 Nack requeue、禁止消费路径二次 Publish 延时消息（避免审核 Green 类重试风暴）。
- Redis 丢失或刷库后闹钟清空，靠客户端下次打开 App 同步恢复（产品接受）。

## Capabilities

### New Capabilities

- `predict-imminent-schedule`：待办全量同步、Redis 权威、延时 MQ 叫醒、消费前校验与去重/历史闸、Ack 失败语义。
- `predict-imminent-push`：公共推送契约、新 bizType、按宝宝绑定用户扇出。
- `device-wx-by-device-no-list`：device 按 `device_no` 列出全部绑定 `wxId` 的内部/契约能力。

### Modified Capabilities

- （无强制改写已归档 `openspec/specs/` 基线 Requirement；本变更以增量 specs 约束新行为。RabbitMQ 拓扑与插件启用在 design/tasks 声明。）

## Impact

- **进程**：`voice-service`（同步 API、Redis、延时 Publish、AMQP consumer、调 history/device/push）；`device-service`（`ListWxIdsByDeviceNo`）；承载公共 push 的进程（从 UCG 推送栈抽契约，一期可仍由 ucg-service 实现 HTTP）；`gateway-app-server`（App 路由反代、Bearer/usage 策略）。
- **基础设施**：RabbitMQ 启用 `rabbitmq_delayed_message_exchange`；新增 delayed exchange + 业务队列；`hack/rabbitmq-init` 与 compose/镜像需同步。
- **存储**：仅 Redis（cachekit 键登记）；无新业务表。
- **客户端**：`flutter_ai_talk` 预测更新后调用同步 API；需注册/复用系统推送 token（与公共 push 注册约定对齐）。
- **非目标**：服务端重算预测模型；推送失败自动重试/outbox；MQ 延时消息主动 cancel；新建微服务；新增 `*_test.go`。
