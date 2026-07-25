## ADDED Requirements

### Requirement: 历史记录筛选 API

系统 SHALL 提供 `GET /device/history/api/filter` 接口，支持按设备号、事件ID列表、时间范围和返回条数上限筛选历史记录。

#### Scenario: 正常筛选 - 所有参数都传
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1,2,5`、`startTime=1234567890`、`endTime=1234567990`、`limit=50` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在时间范围内、事件ID为1/2/5的最多50条历史记录
- **AND** 结果 SHALL 按 id 倒序排列（最新在前）

#### Scenario: eventIds 为空时跳过事件过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=`（空串）、`startTime=1234567890`、`endTime=1234567990` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在时间范围内的所有事件类型的历史记录（不按事件ID过滤）

#### Scenario: startTime 为空时跳过开始时间条件
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1`、`startTime=0`、`endTime=1234567990` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在 endTime 之前、事件ID为1的历史记录（不限制开始时间）

#### Scenario: endTime 为空时跳过结束时间条件
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1`、`startTime=1234567890`、`endTime=0` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在 startTime 之后、事件ID为1的历史记录（不限制结束时间）

#### Scenario: limit 默认值为 100
- **WHEN** 客户端以 `deviceNo=XXX`、`limit=0`（或不传 limit）调用 filter 接口
- **THEN** 系统 SHALL 最多返回 100 条历史记录

#### Scenario: limit 上限为 500
- **WHEN** 客户端以 `deviceNo=XXX`、`limit=1000` 调用 filter 接口
- **THEN** 系统 SHALL 将 limit 限制为 500，最多返回 500 条历史记录

#### Scenario: deviceNo 为空时返回空列表
- **WHEN** 客户端以 `deviceNo=`（空串）调用 filter 接口
- **THEN** 系统 SHALL 返回空列表（不报错）

#### Scenario: 响应格式正确
- **WHEN** filter 接口成功返回
- **THEN** 响应 JSON SHALL 为 `{ "list": [ { "id": ..., "deviceNo": ..., "eventId": ..., ... } ] }` 格式
- **AND** 每条记录 SHALL 包含与 v1 list 接口相同的字段

#### Scenario: 支持 local/remote/canary 服务模式
- **WHEN** 服务运行在 local 模式
- **THEN** filter 接口 SHALL 直连数据库执行查询
- **WHEN** 服务运行在 remote 模式
- **THEN** filter 接口 SHALL 通过 HTTP 调用远程 history-service
- **WHEN** 服务运行在 canary 模式
- **THEN** filter 接口 SHALL 按设备号一致性分流到本地或远程
