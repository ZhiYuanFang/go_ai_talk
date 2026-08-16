## Why

智能对话调用 OpenClaw 时出现 `Post "$%7BOPENCLAW_GATEWAY_URL%7D/v1/chat/completions": unsupported protocol scheme ""`。根因是 `config.voice-service.yaml` 把 `openclaw.gatewayUrl` / `gatewayToken` / `apiToken` 写成字面量 `${ENV}`，GoFrame **不会**展开 YAML 中的 `${VAR}`；而 `OpenClawFromCfg` 只在配置值为空时才回落 `os.Getenv`，导致占位符字符串盖住容器内已正确注入的环境变量。

## What Changes

- 将 `manifest/config/config.voice-service.yaml` 中 `openclaw.gatewayUrl` / `gatewayToken` / `apiToken` 改为空字符串（与同文件 `voice.admin.password` 等密钥占位一致），由进程环境变量提供真实值。
- **不改** `OpenClawFromCfg` 的优先级语义：配置非空优先，空则 `OPENCLAW_GATEWAY_URL` / `OPENCLAW_GATEWAY_TOKEN` / `PANGBAO_API_TOKEN`，URL 仍可回退本地默认。
- **不改** docker-compose 已注入的 `OPENCLAW_*` / `PANGBAO_API_TOKEN` 环境块；`.env.prod` / `.env.example` 继续作为运维真源。
- 在 YAML 注释中明确：禁止 `${VAR}` 字面量，须由环境变量或显式非空配置覆盖。
- **非 BREAKING**：对外 HTTP/API 不变；仅修正配置加载行为，使已部署 env 生效。

## Capabilities

### New Capabilities

- `openclaw-client-config`：voice-service 解析 OpenClaw 门禁 base URL 与 G/A Token 的配置/环境约定（空 YAML + env 回落；禁止未展开的 `${VAR}` 占位）。

### Modified Capabilities

- （无）基线 `openspec/specs/` 无独立 OpenClaw capability 需 delta。

## Impact

- **进程**：`voice-service`（配置文件；运行时仍用既有 `OpenClawFromCfg`）。
- **配置**：`manifest/config/config.voice-service.yaml`；compose / env 文件仅文档对齐（若 `.env.example` 缺 OpenClaw 键则补注释示例）。
- **不改**：OpenClaw HTTP 客户端协议、Intent/Clinic/Care 编排路径、gateway 反代、App 对外接口、Redis/DB。
- **运维**：修复后须保证容器 env 已设置 `OPENCLAW_GATEWAY_URL`（compose 默认 `http://python-ai-talk:8000/agent-gate`）及 Token；未设 Token 时鉴权仍会失败（与修前字面量 Token 同等需配置）。
