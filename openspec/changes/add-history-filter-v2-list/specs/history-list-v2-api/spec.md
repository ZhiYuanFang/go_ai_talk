## ADDED Requirements

### Requirement: v2 历史列表 API

系统 SHALL 提供 `GET /device/history/api/v2/list` 接口，在 v1 分页基础上扩展 startTime、endTime、limit 可选参数。

#### Scenario: 不传新参数时行为与 v1 完全一致
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20` 调用 v2 list 接口（不传 startTime、endTime、limit）
- **THEN** 系统 SHALL 返回与 v1 list 接口完全相同的结果（分页、排序、字段完全一致）
- **AND** v1 list 接口 SHALL 保持完全不变

#### Scenario: 传了 limit 时优先使用 limit 替代 pageSize
- **WHEN** 客户端以 `deviceNo=XXX`、`page=2`、`pageSize=20`、`limit=100` 调用 v2 list 接口
- **THEN** 系统 SHALL 使用 limit=100 作为每页条数，page 忽略（固定为1）
- **AND** total 字段 SHALL 仍然返回符合条件的总记录数

#### Scenario: 传 startTime 时按开始时间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20`、`startTime=1234567890` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time >= 1234567890 的记录
- **AND** total 字段 SHALL 仅统计符合时间条件的记录数

#### Scenario: 传 endTime 时按结束时间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20`、`endTime=1234567990` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time <= 1234567990 的记录

#### Scenario: 同时传 startTime 和 endTime 时按时间区间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`startTime=1234567890`、`endTime=1234567990` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time 在 [1234567890, 1234567990] 区间内的记录

#### Scenario: pageSize 上限仍为 100（与 v1 一致）
- **WHEN** 客户端以 `deviceNo=XXX`、`pageSize=200` 调用 v2 list 接口（未传 limit）
- **THEN** 系统 SHALL 将 pageSize 限制为 100

#### Scenario: 排序使用 id 倒序
- **WHEN** v2 list 接口返回结果
- **THEN** 结果 SHALL 按 id 倒序排列（与 v1 一致，最新在前）

#### Scenario: 时间单位为 Unix 秒
- **WHEN** 客户端传 `startTime` 或 `endTime`
- **THEN** 系统 SHALL 按 Unix 秒解释该值（与 piece 接口一致）
- **AND** 值为 0 时 SHALL 视为"不限制"，跳过对应时间条件

#### Scenario: deviceNo 为空时返回空分页
- **WHEN** 客户端以 `deviceNo=`（空串）调用 v2 list 接口
- **THEN** 系统 SHALL 返回空列表和 total=0（不报错）

#### Scenario: 响应格式与 v1 一致
- **WHEN** v2 list 接口成功返回
- **THEN** 响应 JSON SHALL 为 `{ "list": [...], "total": N, "page": N, "pageSize": N }` 格式（与 v1 完全一致）

#### Scenario: 支持 local/remote/canary 服务模式
- **WHEN** 服务运行在 local 模式
- **THEN** v2 list 接口 SHALL 直连数据库执行查询
- **WHEN** 服务运行在 remote 模式
- **THEN** v2 list 接口 SHALL 通过 HTTP 调用远程 history-service
- **WHEN** 服务运行在 canary 模式
- **THEN** v2 list 接口 SHALL 按设备号一致性分流到本地或远程
