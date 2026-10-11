## ADDED Requirements

### Requirement: Voice Admin 只读预测上报

系统 SHALL 在 voice-service 提供管理端只读接口，按 `deviceNo` 返回当前 Redis 中的预测临近待办与进行中 ETA，且 MUST 使用与既有 voice Admin 相同的口令校验（`X-Admin-Password`）。接口 MUST 复用既有 Load 语义：缺键、空载荷或非法 JSON 时 MUST 返回空列表而非仅因此报错。接口 MUST NOT 修改 Redis、MUST NOT 投递延时 MQ、MUST NOT 发送推送。

响应字段 MUST 与 App 同步条目对齐：

- 临近待办：`events[]` 含 `eventId`、`nextAt`、`title` 及可选 `expectKind`、`expectDurationSec`、`expectAmount`、`expectUnit`、`expectLeafTitle`
- 进行中 ETA：`items[]` 含 `rootEventId`、`leafEventId`、可选 `historyId`、可选 `title`、`startAt`、`expectEndAt`

#### Scenario: 读取临近待办有数据

- **WHEN** 运维已通过 Hub Admin 鉴权，且该 `deviceNo` 的 imminent Redis 键含合法列表
- **THEN** `GET /voice/admin/api/predict/imminent/pending?deviceNo=` 返回 `code=0` 且 `events` 字段与 Redis 条目对齐（含可选 expect 字段）

#### Scenario: 读取进行中 ETA 缺键

- **WHEN** 运维已通过 Hub Admin 鉴权，且该 `deviceNo` 无 active-eta Redis 键
- **THEN** `GET /voice/admin/api/predict/active-eta?deviceNo=` 返回 `code=0` 且 `items` 为空数组

#### Scenario: 口令错误拒绝

- **WHEN** 请求缺少有效 `X-Admin-Password`（含未注入的直连）
- **THEN** 系统 MUST 拒绝该 Admin 预测只读请求（未授权），且 MUST NOT 返回 Redis 列表正文

### Requirement: 设备数据运维预测上报卡片

Hub 设备数据运维页（`resource/public/history.html`）SHALL 提供一张「预测上报」卡片，内含两段：临近待办与进行中 ETA。页面 MUST 使用既有 Admin JWT（`AdminCommon.adminFetch`）调用上述 voice Admin 只读接口，MUST NOT 在浏览器持有或发送 internal secret。卡片 MUST 完整展示各段全部契约字段；时间类字段 MUST 以可读本地时间展示（MAY 同时保留 unix 秒便于核对）。空列表 MUST 明示暂无上报。卡片 MUST 提供刷新；进入页面且 Admin 已登录时 MUST 自动拉取当前页 `deviceNo` 的两段数据。卡片 MUST NOT 提供清空或改写 Redis 的操作。

#### Scenario: 一张卡两段展示

- **WHEN** 运维打开 `/device/admin/history/{deviceNo}` 且 Admin 已登录
- **THEN** 页面展示一张预测上报卡片，内含临近待办与进行中 ETA 两段，并自动请求对应 Admin GET

#### Scenario: 空列表提示

- **WHEN** 任一 Admin GET 返回空 `events` 或空 `items`
- **THEN** 对应段 MUST 显示暂无上报类提示，且 MUST NOT 将空列表表现为系统故障

#### Scenario: 刷新两段

- **WHEN** 运维点击预测上报卡片的刷新
- **THEN** 页面 MUST 重新请求临近待办与进行中 ETA 两段数据并更新展示
