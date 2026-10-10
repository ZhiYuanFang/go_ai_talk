## ADDED Requirements

### Requirement: App 强制重连小智绑定

系统 SHALL 提供经 gateway-app 鉴权的 App 接口 `POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect`，供当前登录用户对属于自己的 active 小智绑定触发立即重拨。该接口 MUST 使用新增的请求/响应类型登记 `g.Meta`；MUST NOT 修改既有 List/Add/Alias/Delete 的 v1 结构体字段。

#### Scenario: 归属用户成功触发重连

- **WHEN** 当前 wx 拥有 id 对应的 active 绑定，并携带合法 Bearer 调用 reconnect
- **THEN** device-service MUST 加载该行的 mcpToken、deviceNo、wxId
- **AND** MUST 经 `clients/xiaozhimcp` 调用 xiaozhi-mcp-service 内部 reconnect（禁止 device 直连 mcpbridge 包）
- **AND** 内部侧 MUST 取消既有会话（若有）并 `startLocked` 启动新会话，即使 token/deviceNo/wxId 与当前会话相同
- **AND** MUST 重置该会话退避并立即发起 redial
- **AND** App 接口 MUST 在触发完成后快速返回成功（MUST NOT 阻塞等待 WebSocket 已连通）

#### Scenario: 非归属或无效绑定

- **WHEN** 绑定不存在、非 active，或不属于当前 wx
- **THEN** reconnect MUST 失败并返回明确业务错误
- **AND** MUST NOT 调用内部 reconnect，MUST NOT 修改绑定行

#### Scenario: 同 key 的 Upsert 不足以重连

- **WHEN** 会话已存在且 token/deviceNo/wxId 未变
- **THEN** 普通 Upsert MAY 继续 noop
- **AND** reconnect/ForceRestart MUST 仍执行 cancel + startLocked（不得仅依赖无 force 的 Upsert）

### Requirement: Entitlement 不阻断既有绑定重连

系统 SHALL 允许对「绑定行 active 且归属当前 wx」的记录执行 reconnect，即使该 wx 的小智 MCP entitlement 已过期或未解锁。MUST NOT 在 reconnect 路径强制调用开通/加购校验（如 EnsureAccessForAdd）。

#### Scenario: 未解锁仍可重连自有绑定

- **WHEN** 用户 entitlement 非 active，但仍有归属自己的 active 绑定行
- **THEN** reconnect MUST 仍可成功触发内部 ForceRestart
- **AND** 若后续因 entitlement/usage 策略被 Manager 侧 lazy-remove，属既有链路行为，本接口不额外保证长期 connected

### Requirement: gateway-app 可达与鉴权

新增 reconnect App 路由 MUST 落在既有 device App 反代前缀内；MUST 要求 Bearer；MUST NOT 加入 `gateway_app_auth_exempt`。usage 是否计入 MUST 在实现前向负责人确认；未获明确答复前 MUST NOT 修改 `usagestats/maintenance_skip.go`。

#### Scenario: 经网关需登录

- **WHEN** 客户端无合法 App Bearer 调用 reconnect
- **THEN** gateway-app MUST 拒绝（与其它 device App 写接口一致）
- **AND** 该 path MUST NOT 出现在 auth exempt 列表

### Requirement: 内部重连接口与客户端

`xiaozhi-mcp-service` SHALL 提供仅内网、需内部密钥的 `POST /xiaozhi-mcp/internal/api/bindings/reconnect`（路径可与实现微调但语义等价）。`internal/clients/xiaozhimcp` SHALL 提供对应出站方法供 device-service 调用。

#### Scenario: 内部鉴权

- **WHEN** 请求缺少合法内部密钥
- **THEN** reconnect 内部接口 MUST 拒绝
- **AND** MUST NOT 经 gateway-app 对 App 直接暴露该内部路径
