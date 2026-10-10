## ADDED Requirements

### Requirement: Hub 小智区断连显示重连

Hub 设备详情页（`resource/public/history.html`）小智 MCP 绑定列表 SHALL 在条目 `connected` 为 false 时展示「重连」操作；为 true 时 MUST NOT 展示该操作（或展示为不可用）。

#### Scenario: 未连接显示重连

- **WHEN** 绑定列表某行 `connected === false` 且操作者已 App 登录
- **THEN** 该行 MUST 显示可用的「重连」控件

#### Scenario: 已连接不显示重连

- **WHEN** 绑定列表某行 `connected === true`
- **THEN** 该行 MUST NOT 提供可点击的「重连」（可不渲染或禁用）

### Requirement: Hub 重连调用 App API

「重连」MUST 以 App Bearer 调用 `POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect`；MUST NOT 新增依赖 Admin JWT 的重连业务 API。成功后 MUST 刷新绑定列表以更新 `connected` 展示。

#### Scenario: 点击重连

- **WHEN** 操作者点击某行「重连」且 App 已登录
- **THEN** 前端 MUST 对该 id 发起上述 POST
- **AND** 成功后 MUST 重新拉取列表
- **AND** 失败时 MUST 向操作者展示错误信息

#### Scenario: 未登录禁用

- **WHEN** 操作者未完成 App 登录
- **THEN** 「重连」MUST 不可用或提交被前端拒绝（与添加/改备注/删除一致）
