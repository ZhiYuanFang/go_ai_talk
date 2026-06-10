## Why

`ucg-service` 的阿里云 OSS AccessKey 与 DashScope API Key 当前以明文写入 `manifest/config/config.ucg-service.yaml`，并出现在 OpenSpec 聚合规格中。凭证进 git 增加泄露与误公开风险；与仓库内既有模式（`GATEWAY_APP_VERSION_ADMIN_PASSWORD`、`*_DB_LINK` 经 `.env.*` 注入）不一致。Go 侧已支持 `UCG_OSS_ACCESS_KEY_*` / `UCG_DASHSCOPE_API_KEY` 覆盖，但部署链路与 yaml 尚未收口。本变更先在主仓库完成 env 化，不拆分独立 secrets 仓库。

## What Changes

- 清空 `config.ucg-service.yaml` 中 `ucg.oss.accessKeyId`、`ucg.oss.accessKeySecret`、`ucg.ai.dashscope_api_key` 明文，保留 bucket/endpoint 等非敏感项；顶部注释说明 env 名。
- 在 `manifest/docker/.env.local`、`.env.test`、`.env.prod` 写入真实 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY`（与现有 DB/JWT 密钥同一模式）。
- 在 `docker-compose.microservices.yml` 的 `ucg-service.environment` 增加上述三个变量的 pass-through。
- 新增 `manifest/docker/.env.example`（占位符 + 注释），供 project.md / runbook 引用；**不**含真实密钥。
- 更新 `docs/runbooks/release-deploy-and-run.md` 中 ucg 部署说明，列出必需 env。
- 从 OpenSpec 变更增量规格中移除明文凭证要求，改为 env 注入约定（Green 继续复用 OSS 凭证，无需独立 env）。
- **不**改动 Go 加载逻辑（`LoadOSSConfig` / `LoadGreenConfig` / `LoadAIConfig` 已满足 env 优先）；**不**在本变更内轮换已泄露密钥或拆分私有 secrets 仓库。

## Capabilities

### New Capabilities

- `ucg-aliyun-secrets-env`：ucg-service 阿里云 OSS/Green（复用 OSS AK）与 DashScope 凭证经 Compose env 注入、yaml 不含明文、`.env.example` 与 runbook 约定。

### Modified Capabilities

- `ucg-oss-presign`：凭证配置要求从「yaml 可含明文 AK/SK」改为「yaml MUST 留空，运行时 MUST 经环境变量注入」。

## Impact

- **配置**：`manifest/config/config.ucg-service.yaml`；`manifest/docker/.env.local|test|prod`；新增 `.env.example`；`docker-compose.microservices.yml`。
- **文档**：`docs/runbooks/release-deploy-and-run.md`；变更增量规格 `ucg-oss-presign`。
- **服务**：`ucg-service` 启动依赖 env 中 OSS/DashScope 凭证；本地/测试/生产 compose 须同步更新 `.env.*`。
- **CI**：`docker-acr.yml` 仍读 `.env.test|prod` 的 ACR 字段，本变更不改动 CI 密钥来源。
- **边界**：不涉及 voice/device 进程；不新增测试文件（符合 AGENTS.md）。
