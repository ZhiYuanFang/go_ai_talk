## ADDED Requirements

### Requirement: 小智 MCP 绑定存储

系统 SHALL 在 device 域持久化用户的小智 MCP 绑定，每条记录包含归属 `wx_id`、喂养落点 `device_no`、小智接入点 `mcp_token`、用户备注 `alias` 以及状态与时间戳。`mcp_token` MUST 在全局唯一（跨 wx）。

#### Scenario: 唯一约束

- **WHEN** 已存在某 `mcp_token` 的有效绑定
- **THEN** 另一账号（或其他记录）再次以相同 token 添加 MUST 失败并返回明确业务错误
- **AND** MUST NOT 写入第二条相同 token 的绑定

### Requirement: App 添加小智绑定

系统 SHALL 提供经 gateway-app 鉴权的 App 接口，供当前登录用户添加小智绑定。客户端 MUST 提交 `mcpToken` 与 `alias`（备注）；服务端 MUST 使用当前账号已绑定的宝宝 `deviceNo` 作为落点，MUST NOT 信任客户端随意指定他人 `deviceNo`。

#### Scenario: 成功添加

- **WHEN** 当前 wx 已绑定非空 `deviceNo`，且 `mcpToken` 非空、全局未被占用，`alias` 符合校验
- **THEN** 系统写入绑定行，`device_no` 等于该 wx 当前 `device_no`
- **AND** 写成功后 MUST 调用 xiaozhi-mcp-service 内部 Upsert（失败不回滚已写入绑定，依赖 reconcile 兜底；实现可记录告警日志）

#### Scenario: 未绑定宝宝

- **WHEN** 当前 wx 的 `device_no` 为空
- **THEN** 添加 MUST 失败，MUST NOT 写入绑定

#### Scenario: Token 冲突

- **WHEN** `mcpToken` 已被其他绑定使用
- **THEN** 添加 MUST 失败并提示 token 已被占用

### Requirement: App 列表与改备注、删除

系统 SHALL 提供经 gateway-app 鉴权的 App 接口：列出当前用户的小智绑定、更新备注、删除绑定。列表与详情 MUST NOT 返回完整 `mcp_token`，MUST 返回脱敏展示（如前后若干字符）及 `alias`、`deviceNo`、状态等。

#### Scenario: 列表脱敏

- **WHEN** 用户请求绑定列表
- **THEN** 响应每条含 id、alias、脱敏 token、deviceNo 等
- **AND** 响应 MUST NOT 包含完整 mcp_token 明文

#### Scenario: 更新备注

- **WHEN** 用户更新属于自己的绑定的 alias
- **THEN** 系统更新该行备注并成功返回
- **AND** MUST NOT 允许修改他人绑定

#### Scenario: 删除绑定

- **WHEN** 用户删除属于自己的绑定
- **THEN** 系统删除（或逻辑删除）该行
- **AND** 写成功后 MUST 调用 xiaozhi-mcp-service 内部 Remove，使对应 Bridge 停止

### Requirement: 内部全量列表

device-service SHALL 提供仅供内网调用的接口，返回全部 active 小智绑定（含完整 token 与 deviceNo），供 xiaozhi-mcp-service 启动与 reconcile 使用。该接口 MUST 校验内部密钥，MUST NOT 经 gateway-app 对 App 暴露。

#### Scenario: 内部拉取

- **WHEN** xiaozhi-mcp-service 携带合法内部密钥请求全量 active 绑定
- **THEN** 返回所有应用于拨号的绑定字段（含完整 mcp_token）

### Requirement: 小智绑定 App 接口计入 usage 统计

经 gateway-app 对外暴露的小智绑定增删改查 App HTTP 接口 MUST 计入 App API 使用统计。实现 MUST 在 `api/v1`（或规定版本）登记完整 `g.Meta` path/method/summary；MUST NOT 将这些路径写入 `usagestats/maintenance_skip.go`。

#### Scenario: 统计纳入

- **WHEN** 运维查看功能使用统计且用户调用了小智绑定列表或增删改接口
- **THEN** 对应 path 出现在 usage 统计中（按现有归一化规则）
- **AND** maintenance_skip 中 MUST NOT 包含这些精确 path

### Requirement: gateway-app 登记

新增小智绑定 App 路由 MUST 落在已有 device App 反代前缀内（或按项目约定补 Bind）；若需 Bearer，MUST NOT 误加入 auth exempt。proposal/tasks 自检：反代、exempt、usage 结论（本能力为「统计」）。

#### Scenario: 经网关可达

- **WHEN** App 携带合法 Bearer 调用小智绑定 CRUD 路径
- **THEN** 请求经 gateway-app 到达 device-service 并按鉴权头识别当前 wx
