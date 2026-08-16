## ADDED Requirements

### Requirement: OpenClaw 客户端配置经空 YAML 回落环境变量

voice-service 的默认配置 MUST 将 `openclaw.gatewayUrl`、`openclaw.gatewayToken`、`openclaw.apiToken` 设为空字符串（或省略等效空值），不得写入未展开的 `${ENV_NAME}` 字面量。

`OpenClawFromCfg`（或等价初始化逻辑）MUST 按以下优先级解析：

1. 配置中非空字符串优先；
2. 否则分别读取环境变量 `OPENCLAW_GATEWAY_URL`、`OPENCLAW_GATEWAY_TOKEN`、`PANGBAO_API_TOKEN`；
3. 若 URL 仍为空，MUST 回退到本地开发默认 `http://127.0.0.1:8000/agent-gate`（或变更文档明确的等价默认）。

解析得到的 base URL MUST 可作为带 scheme 的绝对 URL 用于 `POST {base}/v1/chat/completions`；MUST NOT 将字面量 `${OPENCLAW_GATEWAY_URL}` 用作请求目标。

#### Scenario: 容器仅注入环境变量时采用真实 Gateway URL

- **WHEN** `config.voice-service.yaml` 中 `openclaw.gatewayUrl` 为空，且进程环境 `OPENCLAW_GATEWAY_URL=http://python-ai-talk:8000/agent-gate`
- **THEN** OpenClaw 客户端 base URL MUST 为 `http://python-ai-talk:8000/agent-gate`（允许去掉末尾 `/`）
- **AND** 流式/非流式请求 MUST NOT 使用包含字面量 `${OPENCLAW_GATEWAY_URL}` 的 URL

#### Scenario: Token 空配置回落环境变量

- **WHEN** 配置中 `openclaw.gatewayToken` 与 `openclaw.apiToken` 均为空，且环境已设置 `OPENCLAW_GATEWAY_TOKEN` 与 `PANGBAO_API_TOKEN`
- **THEN** 客户端 MUST 使用对应环境变量值设置 `Authorization: Bearer …` 与 `x-pangbao-api-token`

#### Scenario: 禁止 YAML 假占位符

- **WHEN** 审查默认 `manifest/config/config.voice-service.yaml` 的 `openclaw` 段
- **THEN** `gatewayUrl` / `gatewayToken` / `apiToken` MUST 为空或显式真实值
- **AND** MUST NOT 出现未展开的 `${OPENCLAW_GATEWAY_URL}`、`${OPENCLAW_GATEWAY_TOKEN}`、`${PANGBAO_API_TOKEN}` 字符串
