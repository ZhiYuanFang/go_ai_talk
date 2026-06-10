## Why

仓库已从 git 移除 `manifest/docker/.env.test` / `.env.prod`（含 ACR 密码、DB 连接串等敏感信息），但 `.github/workflows/docker-acr.yml` 仍在 checkout 后读取这些文件获取 `REGISTRY`、`ACR_USERNAME`、`ACR_PASSWORD`，导致 push tag 触发 CI 时第一步即报「缺少环境文件」。需在**不把完整 `.env` 提交回仓库**的前提下恢复镜像构建与 push 能力。

## What Changes

- **BREAKING**：CI 不再从 git 中的 `manifest/docker/.env.test|prod` 读取 ACR 凭证；改为 GitHub Actions **Secrets**（按 test/prod 环境区分 `REGISTRY`）。
- 改造 `.github/workflows/docker-acr.yml`：按 tag 规则或 `workflow_dispatch` 选择目标环境，从 Secrets 加载 ACR 配置；保留现有 `-vpc` → 公网 push 地址推导与格式校验逻辑。
- 更新 `docs/runbooks/release-deploy-and-run.md`「ACR 与 .env 统一配置」：明确 **CI 用 GitHub Secrets**、**ECS 用本地 `.env.test|prod`**（从 `.env.example` 复制，不上传 git）。
- 更新 workflow 文件头注释，移除「无需 GitHub Secrets」表述。
- **不**恢复 `.env.test|prod|local` 进 git；**不**改动 Compose 部署链路与服务器侧 `.env` 字段约定；**不**新增 Go 代码或测试文件。

## Capabilities

### New Capabilities

- `ci-acr-github-secrets`：GitHub Actions 构建 push 七微服务镜像至 ACR 时，凭证与 registry 来自 GitHub Secrets/Environments，而非仓库内 `.env` 文件；test/prod 命名空间路由规则不变。

### Modified Capabilities

- （无）服务器侧 `.env` 与 Compose 行为不变，无既有 capability 的 REQUIREMENTS 级变更。

## Impact

- **CI**：`.github/workflows/docker-acr.yml`；GitHub 仓库 Settings → Secrets and variables → Actions（及可选 Environments `test` / `prod`）。
- **文档**：`docs/runbooks/release-deploy-and-run.md`；`.env.example` 注释（说明 ACR 段仅 ECS 使用，CI 见 runbook）。
- **运维**：首次合并后须在 GitHub 配置 Secrets 方可触发成功构建；ECS 上 `.env.test|prod` 仍手工维护，与 CI 配置独立。
- **安全**：ACR 固定密码不再出现在 git 工作树；git 历史中旧 `.env` 内容需运维自行评估是否轮换凭证。
