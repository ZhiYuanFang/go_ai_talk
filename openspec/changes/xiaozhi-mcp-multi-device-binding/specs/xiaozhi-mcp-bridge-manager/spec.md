## ADDED Requirements

### Requirement: 进程命名为 xiaozhi-mcp-service

系统 SHALL 将原 `mcp-service` 进程、镜像与部署资源命名为 `xiaozhi-mcp-service`（含 cmd 入口、Dockerfile、compose、kustomize、CI 镜像映射、配置文件名与服务日志/OTEL 标识）。历史上无生产用户，允许直接改名而无并行双进程兼容期。

#### Scenario: 部署标识

- **WHEN** 查看生产/测试编排中的该服务
- **THEN** 服务名与镜像名体现 `xiaozhi-mcp-service`（或项目统一的等价命名）
- **AND** MUST NOT 再依赖名为 `mcp-service` 的主部署资源作为稳态入口

### Requirement: 单副本部署

`xiaozhi-mcp-service` 的 Deployment MUST 约定 `replicas: 1`，以保证同一 `mcp_token` 全球至多一条由本系统发起的出站 MCP 连接。

#### Scenario: 副本数

- **WHEN** 应用本变更后的 kustomize/compose 配置
- **THEN** xiaozhi-mcp-service 副本数为 1

### Requirement: Bridge Manager 多连接

`xiaozhi-mcp-service` SHALL 在进程内维护 Manager：为每条 active 绑定建立独立 Bridge（token → 小智接入点长连接；deviceNo → `ChatViaVoiceWS` 落点）。每条 Bridge MUST 沿用断线指数退避重连，且互相隔离（单条失败 MUST NOT 拖垮整个进程）。

#### Scenario: 多 token 并存

- **WHEN** 内部状态中存在两条不同 token 的绑定
- **THEN** Manager MUST 分别维持两条出站连接
- **AND** 小智经某 token 对应连接发起 `tools/call` 时，喂养对话 MUST 使用该绑定的 `deviceNo`

#### Scenario: 移除绑定停连

- **WHEN** Manager 收到对该 token 的 Remove（或 reconcile 发现绑定已不存在）
- **THEN** MUST 取消该 Bridge 的 ctx 并停止重连
- **AND** MUST 从活跃 map 中移除

### Requirement: 内部 HTTP 增量同步

`xiaozhi-mcp-service` SHALL 监听仅内网可达的 HTTP 端口，提供 Upsert/Remove（及健康检查）接口，供 device-service 在绑定写成功后推送。接口 MUST 校验内部密钥。

#### Scenario: Upsert 启桥

- **WHEN** 合法内部请求 Upsert 携带 token 与 deviceNo
- **THEN** Manager 若不存在该 token 则启动 Bridge；若已存在则可更新 deviceNo 或按实现重建会话
- **AND** 非法密钥 MUST 拒绝

#### Scenario: Remove 停桥

- **WHEN** 合法内部请求 Remove 指定 token（或绑定 id）
- **THEN** Manager 停止对应 Bridge

### Requirement: 启动全量拉取与低频 reconcile

`xiaozhi-mcp-service` 启动时 MUST 从 device-service 内部全量 list 拉取 active 绑定并 Reconcile。进程 MUST 运行已批准的后台任务 `xiaozhi-mcp-binding-reconcile`：按可配置间隔再次全量拉取并对齐 Manager（默认分钟级；间隔为 0 时关闭周期 reconcile，仍保留启动全量与写路径推送）。失败 MUST 打日志且 MUST NOT 退出进程。

#### Scenario: 启动对齐

- **WHEN** 进程启动且 device 内部 list 返回若干绑定
- **THEN** Manager 为返回的每条绑定建立 Bridge，且不保留 list 中不存在的陈旧 Bridge

#### Scenario: reconcile 补齐丢推送

- **WHEN** 某绑定已在 DB 但因写路径 HTTP 失败未 Upsert
- **THEN** 下一次 reconcile MUST 发现并启动对应 Bridge

### Requirement: 内部 Service 暴露

编排 MUST 为 `xiaozhi-mcp-service` 提供 ClusterIP（或 compose 网络别名）Service，使 device-service 能解析并调用其内部 HTTP。对外 App 流量 MUST NOT 经 gateway-app 暴露这些内部路径。

#### Scenario: device 可调用

- **WHEN** device-service 配置的 xiaozhi-mcp 基址指向集群内 Service
- **THEN** 写路径 Upsert/Remove 可到达唯一副本

### Requirement: 喂养工具行为保持

小智经 MCP 调用喂养顾问工具时，行为 MUST 继续经 `/voice/chat/ws` 文模式完成对话并返回 answer；deviceNo MUST 来自该 Bridge 绑定，而非进程级单一 env（稳态）。工具入参中的 `xzDeviceNo` 若存在，MUST NOT 覆盖绑定 deviceNo（除非后续变更另有规格）。

#### Scenario: 对话落点

- **WHEN** 绑定 token=T1、deviceNo=D1 的 Bridge 收到非空 transcript 的 tools/call
- **THEN** 系统以 D1 调用 voice chat WS 并返回 answer
