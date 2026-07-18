# Tasks

- [x] Task 1: 初始化 mcp-service 进程骨架与配置文件
  - [x] SubTask 1.1: 创建 `cmd/mcp-service/main.go`，设置 `GF_GCFG_FILE` 默认指向 `manifest/config/config.mcp-service.yaml`，读取并校验 `XIAOZHI_MCP_TOKEN` / `XIAOZHI_MCP_DEVICE_NO`，缺失时 fail-fast 退出
  - [x] SubTask 1.2: 创建 `manifest/config/config.mcp-service.yaml`，仅含 mcp 专属字段（baseUrl、reconnect 参数），顶部注释说明不连 MySQL、依赖 `VOICE_SERVICE_URL` 与 `DEVICE_GATEWAY_INTERNAL_SECRET`
  - [x] SubTask 1.3: 不引入 mcp-go（所有版本要求 Go 1.23+，与项目 Go 1.19 不兼容）；改用 `gorilla/websocket`（已存在依赖）手动实现 MCP JSON-RPC 子集，`go.mod` 无需新增依赖
- [x] Task 2: 实现 WebSocket 传输适配器 `internal/services/mcpbridge`
  - [x] SubTask 2.1: 创建 `internal/services/mcpbridge/bridge.go`，封装 `Bridge` 结构体，持有目标 URL、token、重连参数、chat 工具回调；提供 `Run(ctx)` 方法
  - [x] SubTask 2.2: 实现拨号逻辑：使用 `gorilla/websocket` Dialer 连接 `wss://api.xiaozhi.me/mcp/?token=<token>`，设置 30s 握手超时与默认读写缓冲区
  - [x] SubTask 2.3: 实现读循环：从 WebSocket 读取 JSON-RPC 2.0 帧，根据 `method` 分发到本地 handler（initialize / tools/list / tools/call）；写回响应帧（notification 无响应）
  - [x] SubTask 2.4: 实现指数退避重连：初始 2s、倍增、上限 60s；重连成功后重置退避；日志输出 `mcp-bridge reconnect` 关键字
  - [x] SubTask 2.5: 实现优雅关闭：监听 `ctx.Done()` 与 `SIGTERM`，关闭 WebSocket 并退出 `Run`
- [x] Task 3: 注册 MCP `chat` 工具并接入 DelegateTextChat
  - [x] SubTask 3.1: 创建 `internal/services/mcpbridge/protocol.go`，定义 JSON-RPC 2.0 请求/响应结构、MCP `initialize` / `tools/list` / `tools/call` 响应构造函数
  - [x] SubTask 3.2: 创建 `internal/services/mcpbridge/tools.go`，定义 `chat` 工具 schema（inputSchema：单参数 `transcript`，string，required）与 handler：trim `transcript`，空则返回 MCP error「transcript 不能为空」；非空则调用 `histsvc.DelegateTextChat(ctx, deviceNo, transcript, 0)`，deviceNo 取自 `XIAOZHI_MCP_DEVICE_NO`
  - [x] SubTask 3.3: 成功时返回 `content: [{type:text, text:reply}]`；失败时记录 `mcp-bridge chat failed` 日志并返回 MCP error 携带错误摘要
  - [x] SubTask 3.4: 在 `Bridge` 构造时注入 chat handler，`tools/list` 静态返回包含 `chat` 的工具列表
- [x] Task 4: 编排启动流程 `cmd/mcp-service/main.go`
  - [x] SubTask 4.1: 读取环境变量（token/deviceNo/baseUrl/reconnect 参数），fail-fast 校验；构造 `mcpbridge.NewBridge(...)`
  - [x] SubTask 4.2: 使用 `gctx.New()` 监听 SIGTERM/SIGINT，调用 `bridge.Run(ctx)`；Run 返回后进程退出
  - [x] SubTask 4.3: 启动日志输出 `mcp-service starting`、目标 URL（脱敏 token）、deviceNo
- [x] Task 5: 部署与运行文档同步
  - [x] SubTask 5.1: 创建 `manifest/docker/Dockerfile.mcp-service`，复用现有多阶段构建模式（builder + runtime alpine），入口 `cmd/mcp-service`
  - [x] SubTask 5.2: 在 `manifest/docker/docker-compose.microservices.yml` 与 `docker-compose.microservices.local.yml` 追加 `mcp-service` 服务段，注入 `XIAOZHI_MCP_TOKEN` / `XIAOZHI_MCP_DEVICE_NO` / `DEVICE_GATEWAY_INTERNAL_SECRET` / `VOICE_SERVICE_URL`
  - [x] SubTask 5.3: 在 `manifest/deploy/kustomize/base/` 新增 `mcp-deployment.yaml`（无 Service，因不监听 HTTP），并加入 `kustomization.yaml` resources 列表
  - [x] SubTask 5.4: 在 `manifest/docker/.env.example` 追加 `XIAOZHI_MCP_TOKEN=` / `XIAOZHI_MCP_DEVICE_NO=` 占位与注释
  - [x] SubTask 5.5: 在 `docs/runbooks/release-deploy-and-run.md` 追加 mcp-service 部署说明（不连库、依赖 voice-service 可达）

# Task Dependencies
- Task 2、Task 3 可并行：传输层与工具定义相互独立，最终在 Task 4 编排
- Task 4 依赖 Task 1、Task 2、Task 3
- Task 5 依赖 Task 4（需要二进制名与配置项确定）
