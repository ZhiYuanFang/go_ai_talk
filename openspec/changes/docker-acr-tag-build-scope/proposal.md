## Why

当前 `docker-acr` workflow 在 push 任意 `v*` tag 时 **固定构建并 push 全部 6 个微服务镜像**，小范围改动（如仅 `ucg-service`）也会消耗约 6 倍 GitHub Actions 分钟。需要在 **保持 tag 自动触发** 的前提下，支持通过 git tag 后缀 **显式缩小构建范围**，节省 CI 额度。

## What Changes

- 支持 git tag **`+服务`** 后缀：`v1.0.0-rc.3+ucg` 表示 **仅 build/push `ucg-service`**；无 `+` 后缀时行为与现网一致（全量 6 服务）。
- 解析规则：`+` 之前为 **镜像主 tag**（`primary_tag`，与 ACR / 服务器 `IMAGE_TAG` 一致）；`+` 之后为逗号分隔的服务别名列表（如 `ucg,gateway`）。
- **不实现 retag**：未纳入构建范围的服务 **不在本次 workflow 中 push**；ACR 上不存在对应 tag 的镜像是 **预期行为**。
- 部署侧：运维 **仅 pull/up 变更的服务**（如 `docker compose pull ucg-service && up -d --no-build ucg-service`）；若 pull 失败表示 tag 后缀或构建范围有误。
- 扩展 `workflow_dispatch`：可选 `services` 输入，与 tag 后缀语义一致。
- 更新 `docs/runbooks/release-deploy-and-run.md`：文档化 `+ucg` 用法、部分构建部署步骤与注意事项。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ci-acr-github-secrets`：tag 解析、`docker-acr` 选择性 matrix、无 retag 语义。
- `compose-prod-test-dual-stack`：补充「部分镜像 tag 发布 + 按服务 pull/up」部署约定。

## Impact

- **CI**：`.github/workflows/docker-acr.yml`（`resolve-env` 解析、`build-push` 动态 matrix）。
- **文档**：`docs/runbooks/release-deploy-and-run.md`。
- **不影响**：compose 文件结构、ACR 凭证、test/prod 环境路由规则（仍基于 `+` 前的 base tag）。
- **运维习惯**：部分发版时 **禁止** 对全栈执行 blind `compose pull`（会因缺镜像失败）；须按服务选择性拉取。
