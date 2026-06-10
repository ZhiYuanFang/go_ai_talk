## Context

- **现状**：`.github/workflows/docker-acr.yml` 在 tag push 或 `workflow_dispatch` 时，根据 tag 格式选择 `test`/`prod`，然后从 `manifest/docker/.env.${deploy_env}` 读取 `REGISTRY`、`ACR_USERNAME`、`ACR_PASSWORD`，登录 ACR 并 push 七微服务镜像。
- **触发变更**：commit `删除敏感信息` 后 `.gitignore` 忽略 `**/.env.*`，并从 git 删除 `.env.test|prod|local`；checkout 后 CI 找不到 env 文件即失败。
- **约束**：test/prod 使用同一 ACR 实例、同一套访问凭证，仅 **命名空间**（`REGISTRY` 路径后缀）不同；公网 runner push 时须从 `REGISTRY` 去掉 `-vpc`；服务器 pull 仍可用 `-vpc` 专线域名。
- **利益相关**：运维（配置 GitHub Secrets + ECS `.env`）、发版流程（tag 规则不变）。

## Goals / Non-Goals

**Goals:**

- CI 构建 push 不依赖 git 中的 `.env` 文件。
- 保留现有 tag → 环境路由（`vX.Y.Z` → prod；含 `-` 后缀 → test）与 push 地址校验逻辑。
- Runbook 清晰区分 **CI Secrets** 与 **ECS 本地 `.env`** 职责。
- ACR 凭证通过 GitHub Secrets 管理，workflow 日志中对密码 `add-mask`。

**Non-Goals:**

- 阿里云 OIDC / 短期 token 联邦（后续可评估）。
- 改动 ECS 部署 compose 或 `.env` 字段结构。
- 轮换 git 历史中已泄露的 ACR 密码（运维另行处理）。
- 新增 Go 代码或 `*_test.go`。

## Decisions

### 1. 使用 GitHub Environments + 分层 Secrets

- **决定**：
  - 仓库级 Secrets（两环境共用）：`ACR_USERNAME`、`ACR_PASSWORD`
  - Environment `test`：`REGISTRY`（完整 pull 地址，可含 `-vpc`）
  - Environment `prod`：`REGISTRY`
  - Job 设置 `environment: ${{ steps.acr.outputs.env }}`，使 Secrets 按环境隔离
- **理由**：test/prod 命名空间不同但凭证相同；Environment 边界清晰，prod 可后续加 protection rules。
- **备选**：全部放 Repository Secrets（`REGISTRY_TEST` / `REGISTRY_PROD`）— 更简单但无 prod 审批扩展点；本设计优先 Environments，若仓库未启用 Environments 可在 tasks 中注明 fallback 为 repo secrets 命名约定。

### 2. Workflow 内联读取 Secrets，不生成临时 `.env` 文件

- **决定**：「选择环境并加载」步骤改为从 `${{ secrets.REGISTRY }}` 等读取（经 `environment` 上下文），保留 `read_env_var` 以外的校验函数（`-vpc` 推导、`validate_push_registry`）。
- **理由**：避免在 runner 磁盘写入含密码的文件；减少与「`.env` 不上传」叙事冲突。
- **备选**：Secrets 写出临时 `manifest/docker/.env.*` — 改动面小但多一步 IO 与清理。

### 3. 删除「缺少环境文件」分支，改为「缺少 Secret」

- **决定**：若 `secrets.REGISTRY` / `ACR_USERNAME` / `ACR_PASSWORD` 为空，输出 `::error` 并指向 runbook 配置章节。
- **理由**：失败语义与根因一致，便于排错。

### 4. 服务器 `.env` 与 CI 完全解耦

- **决定**：`.env.example` 保留 `REGISTRY`/`ACR_*` 字段注释，标明「仅 ECS docker login / pull 使用；CI 见 runbook GitHub Secrets」；`.gitignore` 继续忽略真实 `.env.*`。
- **理由**：ECS 仍需要 ACR 凭证做 `docker pull`；字段名不变，运维复制 example 即可，无需与 CI Secrets 同步（值相同但存储位置不同）。

### 5. 规格：新增独立 capability `ci-acr-github-secrets`

- **决定**：不修改 compose / 部署相关既有 spec；CI 凭证来源为独立关注点。
- **理由**：服务器行为未变，避免污染 `v2.0.2` 等聚合规格。

## Risks / Trade-offs

- **[Risk] 合并后未配置 Secrets 导致首次 tag CI 失败** → Runbook 增加「合并前/后 checklist」；workflow 错误信息链到文档锚点。
- **[Risk] Environments 未在仓库启用** → Tasks 验收：若 `environment:` 不可用，document fallback 为 `REGISTRY_TEST`/`REGISTRY_PROD` repo secrets + 条件表达式。
- **[Risk] git 历史仍含旧 ACR 密码** → 提案已建议运维轮换；本变更不自动轮换。
- **[Trade-off] Secrets 与 ECS `.env` 双处维护 ACR 凭证** → 换密码时需同时更新 GitHub 与 ECS；runbook 明确步骤；长期可考虑 OIDC 消除固定密码。

## Migration Plan

1. **合并前**：在 GitHub 创建 Environments `test`、`prod`；配置 `ACR_USERNAME`、`ACR_PASSWORD`（repo 级）与各环境 `REGISTRY`（从本地 `.env.test|prod` 复制，勿提交 git）。
2. **合并**：更新 workflow + runbook + `.env.example` 注释。
3. **验证**：`workflow_dispatch` 选 test + 测试 tag，确认 login + push 成功；再测 prod tag 规则（或 dry-run 仅到 login 步骤）。
4. **回滚**：恢复 workflow 读 `.env` 文件（需临时 re-commit env 或 revert PR）；Secrets 可保留无害。

## Open Questions

- （无阻塞）是否在 prod Environment 启用 required reviewers — 运维策略，不在本变更代码范围。
