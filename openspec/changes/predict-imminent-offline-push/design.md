## Context

预测事项时间点由 Flutter 本地计算；服务端仅有 `prediction_unlock` 条数门禁与「值得留意 / 成长轨迹」等相邻能力，**没有**「下次发生时间」存储，也没有育儿向离线推送管道。现有系统推送仅在 UCG（APNs/HMS/MiPush），且 `device_no → wxId` 反查为 LIMIT 1。RabbitMQ 仅有即时 `voice.events` topic，无延时能力。UCG 审核曾因 handler `return err` → `Nack(requeue=true)` 导致 Green 调用风暴，本设计必须显式规避同类语义。

约束：跨域走契约/clients；Redis 经 cachekit 且键在 `keys_*.go` 登记；新 App 接口须过 gateway；背景 **ticker 扫表默认禁止**，AMQP push consumer 为例外且须声明队列与宿主；不新增 `*_test.go`。

## Goals / Non-Goals

**Goals:**

- 客户端全量同步宝宝待发生事件节点到 Redis；打开 App 刷新后服务端闹钟立刻以新列表为准。
- 提前固定 5 分钟（unix 秒差值）经延时 MQ 叫醒 voice consumer，校验后系统推送给该宝宝全部绑定用户。
- 推送失败 best-effort；消费路径禁止无限 requeue / 循环 Publish。

**Non-Goals:**

- 服务端预测算法；MySQL 持久化待办；推送失败自动重试/outbox。
- 主动 cancel 已入队延时消息（靠 Redis 校验作废幽灵消息）。
- 新建微服务；改 UCG 社区角标算法本身（仅抽公共推送契约并加 bizType）。

## Decisions

### D1. Redis 为调度权威，延时 MQ 只负责叫醒

- **选择**：`deviceNo` 维度待办列表存 Redis（无 TTL，靠全量替换）；同步时按每条 `fireAt=nextAt-300` 发延时消息；消费时必须再读 Redis，确认同 `eventId` 且 `nextAt` 一致后才推。
- **理由**：满足「刷新后立刻改写」；旧延时消息无法 cancel 也可安全丢弃。
- **备选**：纯 ZSET + ticker（不做本地轮询被否）；TTL keyspace（不可靠且本仓无先例）。

### D2. 延时实现：`rabbitmq_delayed_message_exchange` + `x-delay`

- **选择**：独立 delayed exchange（建议名 `voice.delayed`，勿与即时 `voice.events` 混用）；routing key 如 `voice.predict.imminent.fire`；队列如 `voice.predict.imminent.q`；宿主 **voice-service** 启动 AMQP consumer（`autoAck=false`）。
- **理由**：产品要求不做业务侧轮询；插件语义清晰。
- **备选**：TTL+DLX（不装插件但拓扑复杂）— 未采用。
- **批准说明**：本变更批准 voice 进程内 **AMQP push consumer**（非 ticker 扫表）。不批准新增 Redis 到期扫描 ticker。

### D3. 写入语义：最后写入赢

- 任意已登录且绑定该 `deviceNo` 的会话可全量替换列表；多人家庭互相覆盖可接受。
- 同步事务顺序建议：device 级短锁或「先写列表再发 MQ / 先清理逻辑」保证同 device 并发不半更新；失败时以 Redis 最终内容为准，幽灵 MQ 靠校验消化。

### D4. Consumer Ack 语义（防重试风暴）

对齐 `eventkit`：`handler err` → `Nack(requeue=true)`。本能力 **handler 对业务路径一律返回 nil（Ack）**，包括：JSON 非法、Redis 不匹配、去重命中、history 已发生、推送厂商失败、列人失败。仅进程崩溃未 Ack 时由 broker 重投，靠 5 分钟去重键防双推。

**禁止**：消费内因推失败再 `Publish` 延时消息；禁止用 Nack 表达「跳过」。

### D5. 去重与历史闸

- 推送去重键：`predict:imminent:pushed:{deviceNo}:{eventId}`，TTL=5min（具体前缀以 cachekit 登记为准）。
- 推前调用 history `ListHistoryFilter`：`startTime=now-1800`，`endTime=now`，`limit=1`。若客户端上报的是**一级根** `eventId`，MUST 经 device 事件树展开为根+后代叶子再查，避免漏挡。

### D6. 收件人：`ListWxIdsByDeviceNo`

- device-service 新增契约：同 `device_no` 返回全部 `wx.id`（可分页或上限，一期设合理上限并打日志）。
- voice 经 `clients/device` 调用；再对每个 wx 调公共 push。

### D7. 公共推送 + 新 bizType

- 从 UCG 推送发送栈抽出可复用的「按 wxId + channel token 下发可见通知」契约（HTTP internal 或等价），新增 `bizType=predict_imminent`（名称以实现常量准）。
- 预测推送 MUST NOT 复用 UCG 未读角标计算；badge 策略按 bizType（一期可为固定/省略社区 unread）。
- Token 注册：复用现有设备 token 表或约定同一注册入口可服务多 biz（design 实现时优先复用 `ucg_push_device` 读模型经契约访问，避免 voice 直查 ucg 库）。

### D8. API 与版本

- 新 App 路由（建议 voice 域，经 gateway）：例如 `PUT/POST /device/api/predict/imminent/pending`（最终 path 在实现时与现有 `/device/api/*` 风格对齐，**新增**不改旧版结构）。
- 请求体：`deviceNo`（或信任登录绑机头）+ `events: [{ eventId, nextAt, title? }]`；`nextAt` 为 unix 秒。
- 空列表 = 清空该宝宝待办（并依赖校验使旧 MQ 失效）；是否仍发「取消」类 MQ：不发，仅覆盖 Redis。
- gateway：反代 + Bearer；**usage 是否计入须先问负责人**（任务中显式）。

### D9. 权益（一期默认）

- 一期：已登录且请求 `deviceNo` 与会话绑定一致即可同步；接收端为该宝宝全部绑定 wx（有 token 才推得出）。
- 不新增独立 featureId；不强制 `prediction_unlock` 服务端再门禁（客户端已按槽位展示）。若产品后续要付费墙，另开变更。

### D10. 过期仍推

- `delayMs = max(0, (nextAt - 300 - now) * 1000)`；已进入窗口或已过 `nextAt` 仍投递（delay=0），只要 Redis 仍匹配则推。

## Risks / Trade-offs

- [延时消息堆积 / 幽灵消息] → 消费前 Redis 精确匹配；Ack 丢弃；监控队列深度。
- [插件未启用导致发布失败] → 部署清单强制启用插件；同步 API 失败可观测，客户端下次重试。
- [根/叶 eventId 不一致导致 history 闸失效] → 规格强制展开；联调用例覆盖。
- [多 voice 实例重复消费] → 去重键；prefetch 不宜过大。
- [推送 token 仅 UCG 注册用户有] → 文档要求客户端注册；无 token 则静默跳过该 wx。
- [Redis 无 TTL 键增长] → 仅活跃宝宝有键；可用运维扫描，产品接受刷库丢失。

## Migration Plan

1. 构建/部署启用 delayed 插件的 RabbitMQ 镜像；跑 init 声明 exchange/队列。
2. 先发 device `ListWxIdsByDeviceNo` + 公共 push 契约，再发 voice 同步 API 与 consumer。
3. 客户端发版：预测更新后调同步；确保推送注册。
4. 回滚：关 consumer 开关 / 停发延时消息；Redis 键可保留无害；插件可留在 Broker。

## Open Questions

- 公共 push 契约最终挂在 ucg internal HTTP 还是独立 notify 路径（一期倾向 ucg internal，voice → `clients/ucg`）。
- App 通知文案/深链字段最终文案（规格要求可配置或固定模板，实现期与产品确认）。
- usage 统计是否计入（负责人确认前不改 `maintenance_skip`）。
