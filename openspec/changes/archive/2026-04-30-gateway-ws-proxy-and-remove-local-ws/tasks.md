## 1. Gateway WS 透传能力

- [x] 1.1 为 `/voice/chat/ws` 增加可配置的 WebSocket 透传组件，支持将连接转发到 `voice-service`。
- [x] 1.2 增加 WS 路由配置解析（`VOICE_WS_ROUTE_MODE`、`VOICE_WS_PROXY_URL`）与非法配置处理逻辑。
- [x] 1.3 在 proxy 模式下补充握手失败与目标不可达错误返回，确保不回落本地业务执行。

## 2. Gateway 职责收敛

- [x] 2.1 从 gateway 路由注册中移除本地 `registerVoiceChatWS` 绑定。
- [x] 2.2 确认 gateway 在 `/voice/chat/ws` 仅保留边缘层职责（路由/策略/元数据透传）。

## 3. 编排与配置同步

- [x] 3.1 在 `manifest/docker/docker-compose.microservices.yml` 增加/更新 WS 透传相关环境变量。
- [x] 3.2 在部署清单或运行配置中同步 WS 透传变量，确保各环境一致。

## 4. 验证与文档

- [ ] 4.1 执行冒烟验证：前端地址不变时，WS 能通过 gateway 成功握手并完成一轮消息收发。
- [ ] 4.2 验证 proxy 模式下目标不可达时返回明确错误，且不会触发本地业务处理。
- [x] 4.3 更新端点归属矩阵与网关契约文档，标记 `/voice/chat/ws` 已完成委派收敛。
