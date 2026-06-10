## ADDED Requirements

### Requirement: ucg-service 阿里云凭证 MUST 经环境变量注入且 yaml 不得含明文

`ucg-service` 运行时使用的阿里云 OSS AccessKey ID/Secret 与 DashScope API Key MUST 来自容器环境变量，不得依赖 `manifest/config/config.ucg-service.yaml` 中的明文值。仓库内该 yaml 的 `ucg.oss.accessKeyId`、`ucg.oss.accessKeySecret`、`ucg.ai.dashscope_api_key` MUST 留空字符串。Green 内容审核 MUST 复用 OSS 凭证（`LoadGreenConfig` fallback），MUST NOT 要求独立的 Green AccessKey 环境变量。

环境变量名 MUST 为：

| 用途 | 环境变量 |
|------|----------|
| OSS AccessKey ID | `UCG_OSS_ACCESS_KEY_ID` |
| OSS AccessKey Secret | `UCG_OSS_ACCESS_KEY_SECRET` |
| DashScope API Key | `UCG_DASHSCOPE_API_KEY` |

#### Scenario: Compose 启动 ucg-service 时注入 OSS 凭证

- **WHEN** 运维使用 `docker compose --env-file manifest/docker/.env.prod` 启动含 `ucg-service` 的栈，且 `.env.prod` 含 `UCG_OSS_ACCESS_KEY_ID` 与 `UCG_OSS_ACCESS_KEY_SECRET`
- **THEN** ucg-service 容器环境 MUST 可见上述变量且 presign 接口 SHALL 可成功返回 uploadUrl

#### Scenario: yaml 无明文时仍可通过 env 润笔

- **WHEN** `config.ucg-service.yaml` 中 `dashscope_api_key` 为空且容器 env 设置 `UCG_DASHSCOPE_API_KEY`
- **THEN** AI 润笔运行时配置 MUST 使用该 key 且 SHALL NOT 因 yaml 为空而单独失败

#### Scenario: Green 复用 OSS env 凭证

- **WHEN** `ucg.green.enabled` 为 true 且 Green yaml 中 accessKey 为空，容器 env 已设置 `UCG_OSS_ACCESS_KEY_*`
- **THEN** Green 客户端 MUST 使用与 OSS 相同的 AccessKey 发起审核

### Requirement: Docker Compose 基线 MUST pass-through ucg 阿里云 env

`manifest/docker/docker-compose.microservices.yml` 的 `ucg-service.environment` MUST 包含 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 的 `${VAR:-}` 引用，使 local/test/prod overlay 共享同一注入点。

#### Scenario: 基线 compose 引用 env 文件变量

- **WHEN** 开发者查看 `docker-compose.microservices.yml` 中 `ucg-service` 段
- **THEN** MUST 可见上述三个 environment 条目且格式与其它 secrets pass-through 一致

### Requirement: 部署 env 文件 MUST 含真实凭证且 example 仅含占位符

`manifest/docker/.env.local`、`.env.test`、`.env.prod` MUST 包含真实值的 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY`（与本变更实施时 yaml 中既有凭证一致，由实施者填入）。仓库 MUST 提供 `manifest/docker/.env.example`，列出相同 key 名与注释说明，MUST NOT 含真实密钥。

#### Scenario: example 文件可供新环境复制

- **WHEN** 新成员复制 `.env.example` 为本地 `.env` 并填入凭证
- **THEN** key 名 MUST 与 compose pass-through 及 Go 代码读取的 env 名完全一致

### Requirement: runbook MUST 文档化 ucg 阿里云 env 约定

`docs/runbooks/release-deploy-and-run.md` MUST 说明 ucg-service 部署时必需的三项阿里云相关 env、与 yaml 留空的关系，并指向 `.env.example`。

#### Scenario: runbook 检索 env 名

- **WHEN** 运维查阅 release runbook 中 ucg-service 或 secrets 相关章节
- **THEN** MUST 能找到 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 的说明
