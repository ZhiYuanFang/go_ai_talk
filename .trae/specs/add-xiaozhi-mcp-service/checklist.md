# Checklist

## 代码与边界
- [x] `cmd/mcp-service/main.go` 已创建，缺失 `XIAOZHI_MCP_TOKEN` / `XIAOZHI_MCP_DEVICE_NO` 时 fail-fast 退出码非 0
- [x] `internal/services/mcpbridge/bridge.go` 实现 WebSocket 拨号、读循环、写响应、指数退避重连（2s→60s）
- [x] `internal/services/mcpbridge/tools.go` 注册 `chat` 工具，schema 含必填 `transcript` 参数
- [x] `chat` 工具 handler 空字符串校验通过后调用 `histsvc.DelegateTextChat(ctx, deviceNo, transcript, 0)`，未自行直连 voice 库表
- [x] 新进程不 import `internal/dao`、不连 MySQL、不出现 `g.DB()` 调用（grep 验证 0 匹配）
- [x] 新进程未直接 `g.Redis()`，未引入新 Redis 键空间（MCP 桥接不持久化状态）（grep 验证 0 匹配）
- [x] `manifest/config/config.mcp-service.yaml` 顶部注释说明「不连 MySQL、依赖 VOICE_SERVICE_URL 与 DEVICE_GATEWAY_INTERNAL_SECRET」
- [x] `manifest/config/config.mcp-service.yaml` 不含 `database` 段、不含 `voiceChat` 段、不含他域业务字段（仅 logger 段）
- [x] 所有新增导出函数/结构体/关键字段附中文注释，覆盖失败语义与边界条件

## MCP 协议与连接
- [x] 服务拨号连接 `wss://api.xiaozhi.me/mcp/?token=<token>`，URL 拼接正确（baseURL 末尾斜杠与 token query 不重复/丢失）
- [x] 连接建立后能响应小智的 `initialize` 握手与 `tools/list`，返回包含 `chat` 的工具列表
- [x] `tools/call` 调用 `chat` 时成功返回 `content` 数组（type=text）封装的 reply
- [x] transcript 为空时返回 MCP error，不发起下游调用
- [x] 下游 `DelegateTextChat` 失败时返回 MCP error 并记录 `mcp-bridge chat failed` 日志
- [x] WebSocket 断开后自动重连，重连期间日志含 `mcp-bridge reconnect`，进程不退出
- [x] SIGTERM/SIGINT 触发优雅关闭，WebSocket 正常 close（gctx.New() 自动监听信号并 cancel ctx）

## 配置与部署
- [x] `go.mod` / `go.sum` 未引入新依赖（mcp-go 因 Go 版本不兼容改为手动实现），`go build ./...` 通过（git diff 验证无变更）
- [x] `manifest/docker/Dockerfile.mcp-service` 已创建，多阶段构建，入口 `cmd/mcp-service`
- [x] `manifest/docker/docker-compose.microservices.yml` 已追加 `mcp-service` 段（prod/test overlay 同步追加）
- [x] `manifest/deploy/kustomize/base/mcp-deployment.yaml` 已创建并加入 `kustomization.yaml`（无 Service，因不监听 HTTP）
- [x] `manifest/docker/.env.example` 已追加 `XIAOZHI_MCP_TOKEN` / `XIAOZHI_MCP_DEVICE_NO` 占位与注释
- [x] `docs/runbooks/release-deploy-and-run.md` 已追加 mcp-service 部署说明（E 章节）

## 服务边界合规
- [x] 评审 grep `internal/dao` 在 `internal/services/mcpbridge` 与 `cmd/mcp-service` 期望 0 匹配（已验证）
- [x] 评审 grep `g\.Redis\(\)` 在新代码期望 0 匹配（已验证）
- [x] 评审 grep `g\.DB\(\)` 在新代码期望 0 匹配（已验证）
- [x] 未修改任何 v1/v2 已有接口结构（本变更纯新增，git diff device_history_http.go / delegate_http.go 无变更）
- [x] `DelegateTextChat` 行为未被修改，仅被新进程调用（git diff 验证无变更）
