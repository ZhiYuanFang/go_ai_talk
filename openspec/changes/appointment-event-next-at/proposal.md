## Why

疫苗等喂养事件的下次发生时间由预约决定，不遵循宝宝作息间隔。客户端若仍用间隔预测，卡片与提醒不准；若只存在本地，重装/换端会丢失「下次约定」。需要在事件字典上标识预约型事件，并按宝宝持久化下次触发时间，供 Flutter 编辑回填与预测输入，且不改动既有预测临近推送链路。

## What Changes

- 全局事件字典 `event` 增加预约标志（如 `isAppointment`），后台新增/编辑可选，**默认非预约**；经事件列表 / options 等既有返回路径附带该字段供 Flutter 分流。
- 在 **device-service** 新增按 `(deviceNo, eventId)` 存储「下次约定」unix 秒的表与**独立 App API**（与 history 解耦）；`nextAt=0` 表示无下次约定（新增选填未填、或主动清空）；允许编辑 history 时读旧值原样写回。
- **不改** `predict-imminent` 服务端行为：推送权威仍为客户端 `PUT` pending 整表替换；清空可推送约定 = **从 pending 列表移除该事件**（旧延时消息靠消费时 Redis 对账作废）。预约表 **不** 驱动 Redis/MQ/推送。
- 客户端契约（孪生 Flutter，本仓不实现 UI）：预约事件不做间隔预测；预测卡片展示「补充下次触发时间」而非「补充间隔」；有约定则以其为预测结果并按既有流程上报 pending。

## Capabilities

### New Capabilities

- `event-is-appointment`：事件字典预约标志的存储、后台编辑与对客户端/内部 options 的暴露。
- `appointment-next-at`：按宝宝+事件持久化下次约定时间的表、独立读写 API、网关放行与鉴权语义。

### Modified Capabilities

- （无）——明确 **不** 修改 `predict-imminent-schedule` / push 相关需求。

## Impact

- **device-service / device DB**：`event` 表新列；新预约下次表；admin 事件表单字段；App 预约读写 handler；事件 options 缓存重建后快照含新字段（沿用既有 `EventOptions` 缓存模式，**不为预约 nextAt 新建 Redis 读缓存**）。
- **history-service**：经 device 契约拿到的 `entity.Event` 自然含新字段；**无** history 表结构变更、无跨库直查。
- **gateway-app**：为预约 App 路径登记 device 反代前缀；Bearer **不**豁免（须登录且绑机校验，与同类 App 写接口一致）；**usage 统计是否计入须向负责人确认**（未确认前不得改 `maintenance_skip.go`）。
- **voice / predict**：无代码与规格变更。
- **Flutter 孪生仓**：消费 `isAppointment`、调用预约 API、卡片与预测逻辑（本 change 任务以本仓后端为准，设计中约定契约）。
- **API 兼容**：对既有 options/list 为**可选加字段**（非 BREAKING 删改）；预约为**新增**路由，不修改 history CRUD 结构。
