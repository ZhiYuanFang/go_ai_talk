## MODIFIED Requirements

### Requirement: 系统 MUST 提供与 history 解耦的预约下次读写 API

系统 MUST 提供 App 路径读写宝宝某事件的下次约定时间 **与备注**（与 history CRUD 解耦）。请求 MUST 含 `deviceNo` 与根 `eventId`；写路径 MUST 接受 `nextAt`（unix 秒，`0`=无约定）与 `remark`（字符串，可空）。读路径无行时 MUST 返回 `nextAt=0` 且 `remark` 为空串。`nextAt=0` 的写入 MUST 将约定视为清空，且 MUST 将 `remark` 置为空串。正 `nextAt` 写入 MUST 持久化所传 `remark`（含空串清备注但保留约定）。该 API MUST NOT 要求修改 history 行。

#### Scenario: GET 返回备注

- **WHEN** 某 `(deviceNo, eventId)` 已存 nextAt=T 与 remark=R
- **THEN** GET MUST 返回 `nextAt=T` 与 `remark=R`

#### Scenario: PUT 正 nextAt 写备注

- **WHEN** 客户端 PUT `nextAt=T>0` 与 `remark=R`
- **THEN** 系统 MUST 持久化 T 与 R，后续 GET MUST 读到相同值

#### Scenario: PUT nextAt=0 清空备注

- **WHEN** 客户端 PUT `nextAt=0`
- **THEN** 系统 MUST 将 nextAt 视为无约定，且 MUST 将 remark 置为空串

#### Scenario: Internal GET 返回备注

- **WHEN** 内部调用方 GET `/device/internal/api/appointment/next` 且该键已存 nextAt=T 与 remark=R
- **THEN** 响应 MUST 含 `nextAt=T` 与 `remark=R`

#### Scenario: Internal GET 无行

- **WHEN** 内部调用方 GET 某键且存储无行
- **THEN** 响应 MUST 为 `nextAt=0` 且 `remark` 为空串

## ADDED Requirements

### Requirement: appointment_next 表 MUST 持久化 remark 列

`appointment_next`（或等价表）MUST 包含 `remark` 列（字符串，默认空）。启动 ensure MUST 幂等补齐该列。新安装建表 MUST 含该列。

#### Scenario: ensure 补列

- **WHEN** 已有部署缺少 `remark` 列且服务启动 ensure
- **THEN** 系统 MUST 添加该列且 MUST NOT 破坏既有 next_at 数据
