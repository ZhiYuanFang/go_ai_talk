## ADDED Requirements

### Requirement: 客户端 MUST 能全量替换宝宝待发生预测节点

系统 MUST 提供经 gateway-app 暴露的 App HTTP 接口，供已登录且绑定设备的客户端在预测更新后提交该宝宝（`deviceNo`）的待发生事件全量列表。每条事件 MUST 包含 `eventId` 与 `nextAt`（unix 秒）。服务端 MUST 以该列表**整表替换**该宝宝在 Redis 中的待办（MUST NOT 写入业务 MySQL 作为权威）。待办键 MUST 经 `cachekit` 登记且 MUST NOT 依赖 TTL 过期清理（靠后续替换维护）。后一次成功写入 MUST 覆盖前一次（最后写入赢）。空列表 MUST 清空该宝宝待办。

#### Scenario: 预测刷新后覆盖旧列表

- **WHEN** 客户端先后两次成功提交同一 `deviceNo` 的不同 `events` 列表
- **THEN** Redis 中该宝宝待办 MUST 仅等于第二次提交内容，且 MUST NOT 残留仅存在于第一次列表中的事件节点

#### Scenario: 空列表清空

- **WHEN** 客户端提交空 `events` 且请求成功
- **THEN** 该宝宝 Redis 待办 MUST 为空，后续基于旧节点的叫醒校验 MUST 失败并跳过推送

### Requirement: 同步成功后 MUST 按提前量投递延时叫醒消息

对提交列表中每一条有效事件，系统 MUST 计算 `fireAt = nextAt - 300`（秒），并以 `x-delay = max(0, fireAt - now) * 1000` 毫秒发布到已启用 `rabbitmq_delayed_message_exchange` 的延时交换机，routing 到 voice 预测临近队列。消息体 MUST 足以在消费时定位 `deviceNo`、`eventId`、`nextAt`。当 `fireAt <= now` 时 MUST 仍投递（delay 为 0 或等价即时投递），不得因「已过理想叫醒点」而拒绝入队。

#### Scenario: 五分钟前提醒入队

- **WHEN** 当前时间为 T0，某事件 `nextAt = T0 + 600`
- **THEN** 系统 MUST 发布延时消息且延迟约 300 秒量级（允许实现取整误差），消息携带该 `deviceNo`、`eventId`、`nextAt`

#### Scenario: 已进入窗口仍入队

- **WHEN** 某事件 `nextAt - 300 <= now`
- **THEN** 系统 MUST 仍发布叫醒消息（零延迟或即时），MUST NOT 静默丢弃该事件的叫醒

### Requirement: 消费叫醒时 MUST 以 Redis 校验为准且业务失败 MUST Ack

`voice-service` MUST 以 AMQP push consumer（`autoAck=false`）消费预测临近队列。处理单条 delivery 时 MUST 先读取该宝宝当前 Redis 待办：若不存在匹配的 `eventId` 与相同 `nextAt`，MUST 跳过推送并对该 delivery **Ack**。下列情况 MUST **Ack** 且 MUST NOT `Nack(requeue=true)`，MUST NOT 在消费路径再次 Publish 延时重试消息：载荷非法、Redis 不匹配、推送去重命中、历史闸命中、列绑定用户失败、系统推送发送失败。进程崩溃导致未 Ack 时 broker 可重投；应用层 MUST 依赖去重键限制重复可见推送。

#### Scenario: 列表已替换则旧消息作废

- **WHEN** 延时消息携带的 `nextAt` 与 Redis 当前该 `eventId` 的 `nextAt` 不一致或不存在该事件
- **THEN** consumer MUST NOT 发送系统推送，且 MUST Ack 该消息

#### Scenario: 推送失败不回队

- **WHEN** 校验通过但向厂商推送返回错误
- **THEN** consumer MUST 记录可观测日志、MUST Ack，且 MUST NOT Nack requeue，MUST NOT 再发布一条延时消息作为重试

### Requirement: 推送前 MUST 做五分钟去重与三十分钟历史闸

对同一 `(deviceNo, eventId)`，成功进入推送发送路径前 MUST 检查推送去重状态：若五分钟内已推送过，MUST 跳过并 Ack。若未去重命中，MUST 查询 history：该宝宝在过去 1800 秒内是否存在同类事件记录（`start_time` 落在窗口内）；若存在 MUST 跳过推送并 Ack。当上报 `eventId` 为事件树一级根时，历史查询 MUST 包含该根及其后代叶子 `eventId` 集合（经 device 事件目录解析），MUST NOT 仅用根 id 查询导致漏判。通过闸后 MUST 写入五分钟去重标记（成功发送或「已尝试发送」的语义由实现固定为「进入发送前占位」或「发送成功后写入」，但 MUST 保证五分钟内不会对用户重复可见推送风暴）。

#### Scenario: 用户已记录喂养则不推

- **WHEN** Redis 仍匹配某喂养类预测节点，但 history 在近 30 分钟内已有对应叶子事件
- **THEN** 系统 MUST NOT 发送该节点的临近推送

#### Scenario: 五分钟内不重复推同一事件

- **WHEN** 同一 `(deviceNo, eventId)` 在五分钟内第二次通过 Redis 匹配进入推送逻辑
- **THEN** 系统 MUST 跳过第二次可见推送

### Requirement: 叫醒消费宿主与拓扑 MUST 声明

本能力的 AMQP consumer MUST 运行于 `voice-service`。系统 MUST 使用独立于即时 `voice.events` 业务的延时交换机（启用 delayed message 插件），并声明专用队列与绑定。变更与部署文档 MUST 说明插件启用方式；插件不可用时同步路径 MUST 失败可观测，MUST NOT 静默假装已调度。

#### Scenario: voice 进程订阅专用队列

- **WHEN** `voice-service` 启动且 MQ 配置可用
- **THEN** 系统 MUST 订阅预测临近专用队列以消费延时到期消息
