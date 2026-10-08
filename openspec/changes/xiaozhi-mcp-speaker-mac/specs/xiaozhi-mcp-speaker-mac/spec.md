## ADDED Requirements

### Requirement: 添加绑定必填音箱 MAC

系统 SHALL 在创建/更新小智 MCP 绑定时要求客户端提供音箱 MAC。服务端 MUST 将 MAC 规范为小写冒号分隔的六段形式（示例：`3c:dc:75:fc:7f:c4`）。格式非法时 MUST 拒绝并返回明确错误。

#### Scenario: 合法 MAC 规范化

- **WHEN** 客户端提交 `speakerMac` 为 `3C-DC-75-FC-7F-C4` 或等价可解析形式，且其他字段合法
- **THEN** 持久化的 `speaker_mac` MUST 为 `3c:dc:75:fc:7f:c4`

#### Scenario: 非法 MAC

- **WHEN** `speakerMac` 无法解析为 6 字节 MAC
- **THEN** 添加 MUST 失败，MUST NOT 写入绑定行

#### Scenario: 缺少 MAC

- **WHEN** 添加请求未提供非空 `speakerMac`
- **THEN** 添加 MUST 失败

### Requirement: MAC 全局唯一且一宝宝多音箱

`speaker_mac` MUST 在全局唯一，以保证一台音箱只绑定一个宝宝（`device_no`）。同一 `device_no` MUST 允许存在多条不同 MAC 的绑定。

#### Scenario: 多音箱同宝宝

- **WHEN** 同一账号宝宝下先后添加两个不同合法 MAC
- **THEN** 两条绑定均 MUST 成功存在

#### Scenario: MAC 已被其他宝宝占用

- **WHEN** 某 MAC 已绑定 deviceNo=A，另一账号宝宝 deviceNo=B 再次以该 MAC 添加
- **THEN** 添加 MUST 失败并提示该音箱已绑定其他宝宝
- **AND** MUST NOT 修改 A 上的绑定

### Requirement: 同 MAC 再添加自动更新 token

当规范化后的 MAC 已存在且归属当前宝宝时，再次调用添加接口 MUST 更新该行的 `mcp_token`（经既有 token 规范化）及本次提交的 `alias`（若提供），MUST NOT 再插入新行。若 token 发生变化，系统 MUST 通知 xiaozhi-mcp-service 停止旧 token 对应 Bridge 并为新 token 建立 Bridge。

#### Scenario: 换 token

- **WHEN** 宝宝 D 已绑定 MAC=M、token=T1，再次添加 MAC=M、token=T2（T2≠T1，可含完整接入点 URL）
- **THEN** 该行 mcp_token MUST 变为规范化后的 T2
- **AND** 行数 MUST 不增加
- **AND** mcp 侧 MUST 不再使用 T1 拨号，并使用 T2 拨号

#### Scenario: 同 token 再提交

- **WHEN** 同 MAC、同宝宝、token 规范化后与库中一致再次添加
- **THEN** MUST 成功（可更新 alias），MUST NOT 报「token 已被绑定」误伤自身

### Requirement: 列表展示 MAC

App 绑定列表响应 MUST 包含每条绑定的 `speakerMac`（规范化形式）。Hub 设备详情小智绑定 UI MUST 提供 MAC 输入，并在列表中展示 MAC。

#### Scenario: 列表含 MAC

- **WHEN** 用户请求绑定列表且存在绑定行
- **THEN** 每条 MUST 含 `speakerMac` 字段

### Requirement: mcp 进程内维护 token 连接状态

xiaozhi-mcp-service MUST 在进程内存中按规范化后的 `mcp_token` 维护是否已连接小智 MCP WebSocket 的布尔状态。`connected=true` MUST 仅表示当前处于拨号成功后的读循环中；拨号失败、退避重连、会话已 Remove 或无会话 MUST 为未连接。系统 MUST NOT 使用 Redis 持久化该状态。

#### Scenario: 拨号成功标绿

- **WHEN** Bridge 对某 token 拨号成功并进入读循环
- **THEN** 该 token 的连接状态 MUST 为已连接

#### Scenario: 断线或停桥标红

- **WHEN** 读循环结束、Remove、或会话不再存在
- **THEN** 该 token 的连接状态 MUST 为未连接

### Requirement: 内部 HTTP 批量查询连接状态

xiaozhi-mcp-service MUST 提供经内部密钥鉴权的 HTTP 接口，接受一组 token 并返回各自是否已连接。未知 token MUST 视为未连接。

#### Scenario: 批量查询

- **WHEN** 调用方携带合法内部密钥提交若干 token
- **THEN** 响应 MUST 为每个 token 给出布尔连接结果
- **AND** 未在内存中的 token MUST 为未连接

### Requirement: 绑定列表返回连接状态

device-service 在返回小智 MCP 绑定列表时 MUST 对每条绑定按规范化 token 查询 mcp 连接状态，并在列表项中返回 `connected`（布尔）。查询 mcp 失败或超时时，列表请求 MUST 仍成功，相关项 `connected` MUST 为 `false`（不得因状态查询失败使整个列表 5xx）。

#### Scenario: 已连接

- **WHEN** 绑定 token 在 mcp 侧为已连接
- **THEN** 列表项 `connected` MUST 为 `true`

#### Scenario: 未连接或查询失败

- **WHEN** token 未连接，或 device 调用 mcp 状态接口失败
- **THEN** 对应列表项 `connected` MUST 为 `false`
- **AND** 绑定列表 HTTP 仍 MUST 成功返回

### Requirement: Hub 绿红灯展示连接状态

Hub 设备详情小智绑定列表 MUST 根据每条 `connected` 展示连接指示：已连接为绿灯，未连接为红灯。系统 MUST NOT 要求第三态（黄灯/重连中）。

#### Scenario: 绿灯

- **WHEN** 列表项 `connected` 为 `true`
- **THEN** Hub UI MUST 显示绿色连接指示

#### Scenario: 红灯

- **WHEN** 列表项 `connected` 为 `false` 或缺失
- **THEN** Hub UI MUST 显示红色连接指示

### Requirement: 喂养确认与取消短句须选路本工具

xiaozhi-mcp-service 暴露的 `baby_feeding_advisor` 工具描述 MUST 要求：当用户对上一条喂养相关操作进行确认或取消（含短确认词如「是的」「对的」「确定」「好的」「可以」「没问题」，及取消词如「取消」「不要」「算了」「不对」）时，智能体 MUST 调用本工具并将用户原话作为 `transcript`，MUST NOT 自行回答。系统 MUST NOT 为此新增独立 MCP 工具；执行路径 MUST 仍为既有 `ChatHandler` → voice chat WS。

#### Scenario: 描述包含确认与取消选路

- **WHEN** 客户端请求 `tools/list`
- **THEN** `baby_feeding_advisor` 的 description MUST 包含对喂养操作确认与取消短句须调用本工具的说明
- **AND** MUST 仍仅有该喂养顾问工具（不因本需求新增第二个工具名）
