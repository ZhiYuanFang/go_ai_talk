## ADDED Requirements

### Requirement: 推送前 MUST 抑制同类进行中事件

当 `voice-service` 处理预测临近叫醒且 Redis 待办已匹配该 `eventId`/`nextAt` 后，在五分钟去重占位与三十分钟 `start_time` 历史闸之前，系统 MUST 判断该宝宝是否存在同类**进行中**历史：进行中定义为 history 行满足 `end_time=0`。当叫醒载荷 `eventId` 为事件树一级根时，查询 MUST 包含该根及其后代叶子 `eventId` 集合（与三十分钟历史闸同一展开规则）；若展开集合中**任一** id 存在 `end_time=0` 行，系统 MUST 跳过可见推送并对该 delivery **Ack**，MUST NOT 写入五分钟去重键，MUST 记录可观测跳过原因（实现日志 reason MAY 为 `in_progress`）。查询 MUST 经 history 服务契约完成，MUST NOT 由 voice 直查 history 库表。契约查询 MUST 以存在性为目标（最少行扫描，例如 `end_time=0` 且 eventId IN 集合 LIMIT 1）。当进行中查询失败时，系统 MUST fail-closed：MUST NOT 发送该次可见推送，MUST Ack，且 MUST NOT 因此写入五分钟去重键。

#### Scenario: 长时段睡眠进行中不推

- **WHEN** Redis 仍匹配某睡眠类预测节点，且 history 存在对应叶子（或根）`end_time=0` 的未闭合行，即使该行 `start_time` 早于近 1800 秒窗口
- **THEN** 系统 MUST NOT 发送该节点的临近推送，MUST Ack，且 MUST NOT 写入该 `(deviceNo, eventId)` 的五分钟去重键

#### Scenario: 根预测因叶子进行中而抑制

- **WHEN** 叫醒 `eventId` 为一级根，其某后代叶子存在 `end_time=0` 记录
- **THEN** 系统 MUST 跳过该根节点的可见推送

#### Scenario: 不同根并行进行中不互相抑制

- **WHEN** 设备存在事件根 A 的进行中记录，叫醒载荷为不同根 B 且 B 子树无进行中、其它闸可通过
- **THEN** 系统 MUST NOT 仅因根 A 进行中而跳过根 B 的推送

#### Scenario: 进行中查询失败不推且不占去重

- **WHEN** Redis 已匹配，但进行中存在性查询返回错误
- **THEN** 系统 MUST NOT 发送可见推送，MUST Ack，且 MUST NOT 写入五分钟去重键

## MODIFIED Requirements

### Requirement: 推送前 MUST 做五分钟去重与三十分钟历史闸

对同一 `(deviceNo, eventId)`，在**进行中闸已通过**（无同类 `end_time=0` 记录且查询成功）之后，成功进入推送发送路径前 MUST 检查推送去重状态：若五分钟内已推送过，MUST 跳过并 Ack。若未去重命中，MUST 查询 history：该宝宝在过去 1800 秒内是否存在同类事件记录（`start_time` 落在窗口内）；若存在 MUST 跳过推送并 Ack。当上报 `eventId` 为事件树一级根时，历史查询 MUST 包含该根及其后代叶子 `eventId` 集合（经 device 事件目录解析），MUST NOT 仅用根 id 查询导致漏判。通过闸后 MUST 写入五分钟去重标记（成功发送或「已尝试发送」的语义由实现固定为「进入发送前占位」或「发送成功后写入」，但 MUST 保证五分钟内不会对用户重复可见推送风暴）。进行中闸命中或进行中查询失败时 MUST NOT 适用本要求的去重占位（见「推送前 MUST 抑制同类进行中事件」）。

#### Scenario: 用户已记录喂养则不推

- **WHEN** Redis 仍匹配某喂养类预测节点，但 history 在近 30 分钟内已有对应叶子事件
- **THEN** 系统 MUST NOT 发送该节点的临近推送

#### Scenario: 五分钟内不重复推同一事件

- **WHEN** 同一 `(deviceNo, eventId)` 在五分钟内第二次通过 Redis 匹配进入推送逻辑（且进行中闸已通过）
- **THEN** 系统 MUST 跳过第二次可见推送

### Requirement: 消费叫醒时 MUST 以 Redis 校验为准且业务失败 MUST Ack

`voice-service` MUST 以 AMQP push consumer（`autoAck=false`）消费预测临近队列。处理单条 delivery 时 MUST 先读取该宝宝当前 Redis 待办：若不存在匹配的 `eventId` 与相同 `nextAt`，MUST 跳过推送并对该 delivery **Ack**。下列情况 MUST **Ack** 且 MUST NOT `Nack(requeue=true)`，MUST NOT 在消费路径再次 Publish 延时重试消息：载荷非法、Redis 不匹配、**进行中闸命中或进行中查询失败**、推送去重命中、历史闸命中、列绑定用户失败、系统推送发送失败。进程崩溃导致未 Ack 时 broker 可重投；应用层 MUST 依赖去重键限制重复可见推送（进行中闸跳过 MUST NOT 依赖去重键）。

#### Scenario: 列表已替换则旧消息作废

- **WHEN** 延时消息携带的 `nextAt` 与 Redis 当前该 `eventId` 的 `nextAt` 不一致或不存在该事件
- **THEN** consumer MUST NOT 发送系统推送，且 MUST Ack 该消息

#### Scenario: 推送失败不回队

- **WHEN** 校验通过但向厂商推送返回错误
- **THEN** consumer MUST 记录可观测日志、MUST Ack，且 MUST NOT Nack requeue，MUST NOT 再发布一条延时消息作为重试
