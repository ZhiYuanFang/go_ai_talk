## 1. 配置修正

- [x] 1.1 将 `manifest/config/config.voice-service.yaml` 的 `openclaw.gatewayUrl` / `gatewayToken` / `apiToken` 改为空字符串
- [x] 1.2 更新同段中文注释：说明由 `OPENCLAW_GATEWAY_URL` / `OPENCLAW_GATEWAY_TOKEN` / `PANGBAO_API_TOKEN` 注入，禁止 `${VAR}` 字面量

## 2. 运维文档对齐

- [x] 2.1 检查 `manifest/docker/.env.example`：若缺 OpenClaw 相关键，补充注释示例（URL/Token/PANGBAO_API_TOKEN）
- [x] 2.2 确认 `docker-compose.microservices.yml` 中 voice-service 已注入上述三键（已有则无需改代码，仅核对）

## 3. 验收

- [x] 3.1 确认 `OpenClawFromCfg` 在空配置下会读 env（逻辑已存在则只做代码走读核对，不改行为）
- [x] 3.2 自检：YAML 中无 `${OPENCLAW` / `${PANGBAO` 字面量
