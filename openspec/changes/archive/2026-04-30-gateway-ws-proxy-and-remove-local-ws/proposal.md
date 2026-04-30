## Why

当前 `gateway-service` 与 `voice-service` 同时绑定了 `/voice/chat/ws`，导致入口职责不清、迁移路径不完整。需要将 gateway 的 WebSocket 能力收敛为“边缘透传”，并移除 gateway 本地 WS 业务处理，确保领域执行只在 `voice-service`。

## What Changes

- 为 gateway 增加 `/voice/chat/ws` 的可配置 WebSocket 透传能力（`local|proxy`），支持将连接代理到 `voice-service`。
- 在 gateway 删除本地 `registerVoiceChatWS` 业务绑定，避免同一路径“双实现”并存。
- 为 WS 透传增加运行时配置与失败语义：目标不可达时明确返回握手/代理错误，不回落本地业务。
- 更新容器编排与运行文档，补充 WS 透传相关环境变量与验证步骤。

## Capabilities

### New Capabilities
- `gateway-ws-edge-proxy`: 定义 gateway 对 `/voice/chat/ws` 的边缘透传能力、配置模型与错误语义。
- `gateway-ws-delegation-convergence`: 定义 gateway 删除本地 WS 领域执行并收敛为委派层的目标行为。

### Modified Capabilities
- None.

## Impact

- 影响代码：`internal/controller/register.go`、gateway 代理相关控制器文件、`manifest/docker/docker-compose.microservices.yml`。
- 影响运行配置：新增/调整 WS 透传环境变量（如 `VOICE_WS_ROUTE_MODE`、`VOICE_WS_PROXY_URL`）。
- 影响接口行为：前端 WS 对外入口路径可保持不变，但后端执行责任转移到 `voice-service`。
