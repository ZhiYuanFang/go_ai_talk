## ADDED Requirements

### Requirement: tools/call 校验 wx 有效权益

xiaozhi-mcp-service 在执行喂养类 `tools/call` 前 MUST 校验该 Bridge 绑定所属 `wx_id` 仍具备有效小智 MCP 权益（永久或未过期试用）。校验 MUST 经 cash 服务接口契约完成，MUST NOT 在 mcp 进程直查 cash 库表。未开通或校验系统失败时 MUST fail-closed：返回工具层错误，MUST NOT 调用 voice 喂养对话。

#### Scenario: 有效权益允许调用

- **WHEN** 绑定 wx 具有有效权益且工具入参合法
- **THEN** 工具按既有逻辑执行喂养对话并返回结果

#### Scenario: 过期或未开通拒绝调用

- **WHEN** 绑定 wx 无有效权益（含试用已过期、退款撤销后）
- **THEN** `tools/call` MUST 返回错误结果
- **AND** MUST NOT 建立或使用 voice chat 完成喂养

#### Scenario: cash 不可达拒绝调用

- **WHEN** 开通态查询失败（超时/5xx/未授权等）
- **THEN** `tools/call` MUST 失败并提示暂时无法校验
- **AND** MUST NOT 因瞬时错误必然拆桥（与明确未开通区分）

### Requirement: 未开通时懒停桥

当 `tools/call` 因账号**明确未开通**而拒绝时，系统 SHOULD 尽力停止该 token 对应 Bridge（懒停）。绑定行 MUST 保留在 device 库中（不因懒停删除 token/MAC）。

#### Scenario: 拒答后拆桥

- **WHEN** tools 因未开通拒答
- **THEN** 系统尽力 Remove 该 Bridge，使出站连接停止
- **AND** `xiaozhi_mcp_binding` 对应行仍存在

#### Scenario: 再开通后可复用绑行

- **WHEN** 用户随后获得有效权益且绑行仍在
- **THEN** 经写路径 Upsert（或等价同步）后 MUST 能重新建立 Bridge
- **AND** MUST NOT 强制用户因过期而重新录入 token（除非产品另有操作）

### Requirement: Bridge 与同步载荷携带 wxId

device 通知 xiaozhi-mcp 的 Upsert 以及内部全量绑定列表 MUST 包含每条绑定的 `wxId`。mcp Manager/Bridge MUST 在会话中保存该 `wxId` 供 tools 校验。缺少有效 `wxId` 时 tools/call MUST fail-closed。

#### Scenario: Upsert 带 wxId

- **WHEN** device 添加或更新绑定并通知 mcp Upsert
- **THEN** 请求体含该绑定归属 `wxId`
- **AND** 后续该桥上的 tools/call 使用该 `wxId` 查权益

#### Scenario: 内部 list 带 wxId

- **WHEN** xiaozhi-mcp 拉取内部全量绑定做启动或 reconcile
- **THEN** 每条记录含 `wxId`
- **AND** 新建 Bridge 会话保存该值
