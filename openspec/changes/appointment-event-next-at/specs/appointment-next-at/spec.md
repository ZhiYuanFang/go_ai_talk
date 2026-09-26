## ADDED Requirements

### Requirement: 系统 MUST 按宝宝与事件持久化下次约定时间

`device-service` MUST 在 device 域库提供存储，键为 `(deviceNo, eventId)`，值为下次约定 unix 秒 `nextAt`。同一宝宝同一事件 MUST 仅有一条有效约定（唯一约束）。`nextAt=0` MUST 表示「无下次约定」。该存储 MUST NOT 作为预测临近推送的权威，MUST NOT 由服务端据此写入 predict Redis 或发布叫醒 MQ。

#### Scenario: 写入非零下次约定

- **WHEN** 已登录且绑机一致的客户端对某 `(deviceNo, eventId)` 写入 `nextAt=T`（T>0）
- **THEN** 随后读取同一键 MUST 得到 `nextAt=T`

#### Scenario: 清空下次约定

- **WHEN** 客户端对已有约定的键写入 `nextAt=0`
- **THEN** 随后读取同一键 MUST 得到 `nextAt=0`（无下次约定）

#### Scenario: 与 history 行无关

- **WHEN** 同一宝宝同一 `eventId` 存在多条 history 记录
- **THEN** 读取该事件下次约定 MUST 返回同一 `nextAt`，MUST NOT 因所选 history 行不同而变化

### Requirement: 系统 MUST 提供与 history 解耦的预约下次读写 API

系统 MUST 经 gateway-app 对 App 暴露独立于 history CRUD 的读写接口（路径前缀 MUST 落在已登记的 device 反代模式内，推荐 `/device/app/api/appointment/*`）。读接口 MUST 接受 `deviceNo` 与 `eventId`；当存储无行时 MUST 返回 `nextAt=0`，MUST NOT 将无行视为系统错误。写接口 MUST 接受 `deviceNo`、`eventId`、`nextAt`（含 0）并 upsert。接口 MUST 要求有效登录；请求 `deviceNo` MUST 与当前会话绑定设备一致，否则 MUST 拒绝。上述接口 MUST NOT 修改 history 表。

#### Scenario: 无行时读为 0

- **WHEN** 某 `(deviceNo, eventId)` 尚无存储行且客户端调用读接口
- **THEN** 响应 MUST 成功且 `nextAt=0`

#### Scenario: 绑机不一致拒绝写

- **WHEN** 请求中的 `deviceNo` 与会话当前绑定设备不一致
- **THEN** 系统 MUST 拒绝写入且 MUST NOT 更新存储

#### Scenario: 原样写回

- **WHEN** 客户端先读到 `nextAt=T`，在仅修改 history 内容后将同一 `T` 再写入预约接口
- **THEN** 存储中的下次约定 MUST 仍为 `T`

### Requirement: 网关 MUST 放行预约 App 路径到 device-service

`gateway-app-server` MUST 将预约 App 路径反代至 device-service（在 `DEVICE_API_ROUTE_MODE` 启用代理时），MUST NOT 误送到 history/voice。该路径 MUST NOT 加入 Bearer 匿名白名单。是否计入 usage 统计 MUST 以负责人确认为准；未确认前 MUST NOT 擅自加入或移出 `maintenance_skip`。

#### Scenario: 反代可达

- **WHEN** 边缘已配置 device 代理且客户端请求预约读写路径
- **THEN** 请求 MUST 到达 device-service 对应 handler，而非本机 404 或其它域服务

### Requirement: 预约下次约定 MUST NOT 改变预测临近服务端行为

本能力 MUST NOT 要求修改 `PUT /device/api/predict/imminent/pending` 的校验、整表替换、延时投递或消费对账语义。客户端清空可推送约定时，产品约定为从 pending 列表移除该事件；服务端继续依赖既有 Redis 对账使旧 `nextAt` 延时消息作废。

#### Scenario: 预约写零不自动清 Redis

- **WHEN** 客户端仅调用预约写接口将 `nextAt` 置 0，未调用 pending 同步
- **THEN** 系统 MUST NOT 因此自动删除或改写该宝宝的 predict Redis 待办（推送侧行为保持既有客户端上报权威）
