## MODIFIED Requirements

### Requirement: 推送前 MUST 做五分钟去重与三十分钟历史闸

对同一 `(deviceNo, eventId)`，在**进行中闸已通过**（无同类 `end_time=0` 记录且查询成功）之后，成功进入推送发送路径前 MUST 检查推送去重状态：去重窗口 MUST 等于当前生效的 `leadSeconds`（与同步路径同一配置源 `VOICE_PREDICT_IMMINENT_LEAD_SECONDS`，非法或未配置时回退实现默认秒数；用于去重 TTL 时若 `leadSeconds < 1` 则 MUST 按 1 秒计）。若该窗口内已推送过（或已占位），MUST 跳过并 Ack。若未去重命中，MUST 查询 history：该宝宝在过去 **1800 秒**内是否存在同类事件记录（`start_time` 落在窗口内）；若存在 MUST 跳过推送并 Ack。当上报 `eventId` 为事件树一级根时，历史查询 MUST 包含该根及其后代叶子 `eventId` 集合（经 device 事件目录解析），MUST NOT 仅用根 id 查询导致漏判。通过闸后 MUST 写入去重标记，TTL MUST 为上述去重窗口（成功发送或「已尝试发送」的语义由实现固定为「进入发送前占位」或「发送成功后写入」，但 MUST 保证 leadSeconds 窗口内不会对用户重复可见推送风暴）。进行中闸命中或进行中查询失败时 MUST NOT 适用本要求的去重占位（见「推送前 MUST 抑制同类进行中事件」；进行中闸语义本变更 MUST NOT 修改）。历史闸 **1800 秒** MUST NOT 因本变更改为跟随 `leadSeconds`。

#### Scenario: 用户已记录喂养则不推

- **WHEN** Redis 仍匹配某喂养类预测节点，但 history 在近 30 分钟内已有对应叶子事件
- **THEN** 系统 MUST NOT 发送该节点的临近推送

#### Scenario: leadSeconds 窗口内不重复推同一事件

- **WHEN** `leadSeconds` 生效为 600，且同一 `(deviceNo, eventId)` 在 600 秒内第二次通过 Redis 匹配进入推送逻辑（且进行中闸已通过）
- **THEN** 系统 MUST 跳过第二次可见推送

#### Scenario: leadSeconds 为 0 时去重仍至少 1 秒

- **WHEN** `VOICE_PREDICT_IMMINENT_LEAD_SECONDS=0` 且同一 `(deviceNo, eventId)` 在约 1 秒内第二次进入推送逻辑（进行中闸已通过）
- **THEN** 系统 MUST 仍能通过去重跳过第二次可见推送（TTL MUST NOT 为 0）
