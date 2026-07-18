# 小智 MCP 服务接入 Spec

## Why
小智 AI 平台（xiaozhi.me）支持通过 MCP 接入点把外部工具暴露给其语音终端的大模型。当前项目的 `/device/history/api/chat` 已具备文本对话能力（最终委派至 voice-service 的 `voice.Voice().TextChat`），但小智平台无法直接调用该能力。需要一个独立的 MCP 服务进程，主动连接小智的 `wss://api.xiaozhi.me/mcp/?token=...` 接入点，把「文本对话」作为 MCP 工具暴露给小智，使语音终端用户可以通过小智触发本系统的设备对话。

## What Changes
- 新增独立 Go 进程 `cmd/mcp-service`，作为 MCP Server 通过 WebSocket 主动拨号连接小智 MCP 接入点（小智为 MCP Client，本服务为 MCP Server，连接方向为 Server → Client）。
- 新增 MCP 工具 `chat`：接收 `transcript` 文本参数，进程内直接调用 `histsvc.DelegateTextChat(ctx, deviceNo, transcript, 0)` 完成对话，返回 `reply` 字符串。
- `deviceNo` 从环境变量 `XIAOZHI_MCP_DEVICE_NO` 读取（单设备绑定场景，硬编码）；`wxID` 传 0（不参与额度校验）。
- 新增 WebSocket 传输适配器 `internal/services/mcpbridge`，负责：
  - 拨号连接 `wss://api.xiaozhi.me/mcp/?token=...`
  - 维持长连接、心跳保活、断线重连（指数退避，上限 60s）
  - 在 WebSocket 与本地 MCP JSON-RPC handler 之间桥接帧（手动实现，非 mcp-go）
- 新增配置文件 `manifest/config/config.mcp-service.yaml`，仅承载 MCP 服务专属配置（接入点 URL、deviceNo、重连参数）。
- 新增 Dockerfile `manifest/docker/Dockerfile.mcp-service` 与 docker-compose / kustomize 部署片段。
- **不引入 `github.com/mark3labs/mcp-go`**：经核实所有 mcp-go 版本均要求 Go 1.23+，与本项目 `go 1.19` 不兼容；升级 Go 版本超出本变更范围。改为基于已存在的 `github.com/gorilla/websocket v1.5.3` 手动实现最小 MCP JSON-RPC 2.0 协议子集（`initialize` / `notifications/initialized` / `tools/list` / `tools/call`），不新增任何外部依赖。

## Impact
- Affected specs（v2.0.24 基线）:
  - `history-voice-delegation`：本变更复用 `DelegateTextChat` 委派路径，不修改其行为。
  - `history-service-db-ownership`：新进程不直连任何 MySQL 库，仅经 HTTP 委派 voice-service，符合服务边界。
  - `background-loop-task-governance`：MCP WebSocket 长连接 + 重连 goroutine 属于「网络保活」而非「扫表循环」，需在 design 显式声明宿主进程与失败语义。
  - `gateway-app-server`：本变更不经过 gateway-app，不涉及 App HTTP 接口登记。
- Affected code:
  - 新增 `cmd/mcp-service/main.go`
  - 新增 `internal/services/mcpbridge/`（WebSocket 传输 + 工具注册 + 桥接循环）
  - 新增 `manifest/config/config.mcp-service.yaml`
  - 新增 `manifest/docker/Dockerfile.mcp-service`
  - 修改 `manifest/docker/docker-compose.microservices.yml`、`docker-compose.microservices.local.yml`（追加 mcp-service 段）
  - 修改 `manifest/deploy/kustomize/base/`（追加 mcp-service deployment + service）
  - 修改 `go.mod` / `go.sum`（追加 mcp-go 依赖）
  - 修改 `manifest/docker/.env.example`（追加 `XIAOZHI_MCP_*` 变量）

## ADDED Requirements

### Requirement: 小智 MCP 接入服务
系统 SHALL 提供一个独立 Go 进程 `cmd/mcp-service`，作为 MCP Server 通过 WebSocket 主动连接小智 MCP 接入点 `wss://api.xiaozhi.me/mcp/?token=...`，向小智平台暴露本系统的设备文本对话能力。

#### Scenario: 启动并建立 MCP 连接
- **WHEN** `cmd/mcp-service` 启动且 `XIAOZHI_MCP_TOKEN` 与 `XIAOZHI_MCP_DEVICE_NO` 已配置
- **THEN** 服务拨号连接 `wss://api.xiaozhi.me/mcp/?token=<token>`
- **AND** 连接建立后向小智平台响应 `initialize` 握手，声明本服务为 MCP Server
- **AND** 当小智发起 `tools/list` 请求时，返回包含 `chat` 工具的列表

#### Scenario: WebSocket 断线自动重连
- **WHEN** 与小智的 WebSocket 连接因网络异常断开
- **THEN** 服务以指数退避策略重连（初始 2s，上限 60s），重连成功后重新完成 MCP 握手
- **AND** 重连期间日志输出 `mcp-bridge reconnect` 关键字与等待时长
- **AND** 重连不导致进程退出

#### Scenario: 缺失关键配置时启动失败
- **WHEN** `XIAOZHI_MCP_TOKEN` 或 `XIAOZHI_MCP_DEVICE_NO` 为空
- **THEN** 进程在启动阶段 fail-fast，日志输出缺失的环境变量名并退出码非 0

### Requirement: MCP chat 工具
系统 SHALL 在 MCP Server 上注册名为 `chat` 的工具，接收文本输入，进程内调用 `histsvc.DelegateTextChat` 完成设备对话并返回回复文本。

#### Scenario: 成功对话
- **WHEN** 小智平台调用 `tools/call`，工具名为 `chat`，参数 `transcript` 非空
- **THEN** 服务以 `XIAOZHI_MCP_DEVICE_NO` 作为 deviceNo、`transcript` 作为文本、`wxID=0`，调用 `histsvc.DelegateTextChat(ctx, deviceNo, transcript, 0)`
- **AND** 将返回的 `reply` 字符串作为工具结果返回给小智
- **AND** 工具结果以 MCP `content` 数组（type=text）封装

#### Scenario: transcript 为空
- **WHEN** `tools/call` 的 `transcript` 参数为空字符串或仅空白
- **THEN** 返回 MCP 错误响应，`message` 为「transcript 不能为空」，不发起下游调用

#### Scenario: 下游对话失败
- **WHEN** `DelegateTextChat` 返回 error（如 voice-service 不可达、AI 配额耗尽、未登录）
- **THEN** 返回 MCP 错误响应，`message` 携带错误摘要
- **AND** 日志记录 `mcp-bridge chat failed` 与原始错误

### Requirement: 配置与部署
系统 SHALL 通过环境变量提供运行配置，并通过独立的配置文件 / Dockerfile / kustomize 资源与现有微服务部署体系一致。

#### Scenario: 环境变量配置
- **WHEN** 部署 mcp-service
- **THEN** 以下环境变量 MUST 可配置：
  - `XIAOZHI_MCP_TOKEN`：小智接入点 token（拼接到 wss URL）
  - `XIAOZHI_MCP_DEVICE_NO`：绑定的设备号
  - `XIAOZHI_MCP_BASE_URL`：接入点基址，默认 `wss://api.xiaozhi.me/mcp/`
  - `XIAOZHI_MCP_RECONNECT_MIN_MS`：重连初始退避，默认 2000
  - `XIAOZHI_MCP_RECONNECT_MAX_MS`：重连退避上限，默认 60000
  - `DEVICE_GATEWAY_INTERNAL_SECRET`：`DelegateTextChat` 内部 HTTP 鉴权所需（已存在）
  - `VOICE_SERVICE_URL`：voice-service 基址（已存在，默认 http://127.0.0.1:9802）

#### Scenario: 配置文件边界
- **WHEN** 查看 `manifest/config/config.mcp-service.yaml`
- **THEN** 文件仅包含 mcp-service 专属字段（接入点、重连参数），不承载 `database` 段、不承载 voiceChat 段
- **AND** 顶部注释说明本进程不连 MySQL、依赖 `VOICE_SERVICE_URL` 与 `DEVICE_GATEWAY_INTERNAL_SECRET`

#### Scenario: 部署资源齐全
- **WHEN** 在 K8s 或 docker-compose 部署 mcp-service
- **THEN** 存在 `Dockerfile.mcp-service`、kustomize `mcp-deployment.yaml` + `mcp-service.yaml`、compose `mcp-service` 段
- **AND** `.env.example` 包含 `XIAOZHI_MCP_*` 变量占位

## MODIFIED Requirements

### Requirement: 背景循环任务治理（v2.0.24 `background-loop-task-governance`）
本变更在 `internal/services/mcpbridge` 引入常驻 goroutine 维持 WebSocket 长连接与断线重连。该循环不属于「扫表 reconciler / outbox 轮询 / HTTP Pull 队列」，而是「对外协议保活」，类比 voice-service 的 WS 会话维持。

- 任务名：`mcp-bridge-reconnect`
- 宿主进程：`cmd/mcp-service`
- 触发条件：WebSocket 读循环返回 err 或 EOF
- 失败语义：指数退避重连，不退出进程；重连失败上限 60s
- 开关：`XIAOZHI_MCP_TOKEN` 为空时进程直接退出，不进入循环
- 关闭方式：进程收到 SIGTERM 时关闭 WebSocket 并退出

## REMOVED Requirements
无。
