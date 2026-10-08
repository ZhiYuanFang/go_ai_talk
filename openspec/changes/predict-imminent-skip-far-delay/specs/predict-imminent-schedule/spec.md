## MODIFIED Requirements

### Requirement: 同步成功后 MUST 按提前量投递延时叫醒消息

对提交列表中每一条有效事件，系统 MUST 计算 `fireAt = nextAt - leadSeconds`（秒；`leadSeconds` 默认 300，可由环境变量覆盖），并计算 `delayMs = max(0, fireAt - now) * 1000`。当且仅当 `delayMs` **不超过** RabbitMQ delayed message 插件可表达上限（`2^32−1` 毫秒，约 49.7 天；实现 MUST 使用显式常量并注释对齐该上限）时，系统 MUST 将消息发布到已启用 `rabbitmq_delayed_message_exchange` 的延时交换机（headers `x-delay=delayMs`），routing 到 voice 预测临近队列；消息体 MUST 足以在消费时定位 `deviceNo`、`eventId`、`nextAt`。当 `fireAt <= now` 时 MUST 仍投递（delay 为 0 或等价即时投递），不得因「已过理想叫醒点」而拒绝入队。

当 `delayMs` **大于**上述上限时，系统 MUST NOT 为该事件调用延时发布（MUST NOT 投递会导致插件 nodelay 立刻路由的超大 `x-delay`），MUST 仍将该事件保留在 Redis 待办中，MUST NOT 因此使整表同步失败，MUST 记录可观测跳过日志（至少含 `eventId` 与跳过原因）。远点的后续叫醒依赖客户端在 `delayMs` 可表达时再次全量同步以补发近端延时消息；本 Requirement MUST NOT 要求消费路径二次 Publish 或后台 ticker 补投。

#### Scenario: 五分钟前提醒入队

- **WHEN** 当前时间为 T0，某事件 `nextAt = T0 + 600`
- **THEN** 系统 MUST 发布延时消息且延迟约 300 秒量级（允许实现取整误差），消息携带该 `deviceNo`、`eventId`、`nextAt`

#### Scenario: 已进入窗口仍入队

- **WHEN** 某事件 `nextAt - leadSeconds <= now`
- **THEN** 系统 MUST 仍发布叫醒消息（零延迟或即时），MUST NOT 静默丢弃该事件的叫醒

#### Scenario: 超过插件延时上限则只写 Redis 不发 MQ

- **WHEN** 某有效事件计算得 `delayMs > 4294967295`（例如 `nextAt` 约在两年后）且整表同步成功
- **THEN** 该事件 MUST 出现在 Redis 待办中，系统 MUST NOT 为其发布延时叫醒消息，同步 API MUST 仍成功，且 MUST 有可观测跳过日志

#### Scenario: 同批近端远点互不影响

- **WHEN** 同一次提交同时包含 `delayMs` 在上限内与超过上限的有效事件
- **THEN** 近端事件 MUST 照常入队延时消息，远端事件 MUST 仅保留于 Redis 且不入队，整表同步 MUST 成功
