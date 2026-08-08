## Why

`vip-cash-service` 已落地本地 Compose / Dockerfile / runbook，但发版链路未纳入 `cash-service`：`.github/workflows/docker-acr.yml` 全量仍为九服务，test/prod overlay 无 ACR 镜像声明。打 tag 后测试/生产拉不到 cash，VIP 支付与 care-alert 选模在目标环境不可用。同时缺少全局约束，后续新增微服务易再次漏改 CI。

## What Changes

- 将 `cash-service` 纳入 `docker-acr.yml`：别名、`dockerfile_for`、`ALL_SERVICES`、注释与错误文案中的服务数（九→十）。
- 在 `docker-compose.microservices.test.yml` / `.prod.yml` 增加 `cash-service` 的 `${REGISTRY}/cash-service:${IMAGE_TAG}`、container_name、端口（test `19807:9807` / prod `9807:9807`）与网络。
- 同步 `docs/runbooks/release-deploy-and-run.md` 中全量服务数、构建别名（含 `cash`/`cash-service`）与 ACR 仓库说明。
- 在 `openspec/project.md`（及必要时仓库级 `AGENTS.md` 对齐提示）写明强制约定：新增微服务 MUST 同步更新 GitHub workflows（至少 `docker-acr.yml`）、test/prod Compose overlay、runbook 别名/服务数，并提醒 ACR 控制台建仓。
- 运维侧（非代码）：在 ACR `pangbao-test` / `pangbao-release` 创建 `cash-service` 仓库（tasks 中列为检查项）。

## Capabilities

### New Capabilities

- `microservice-ci-release`：新增或扩展微服务时的 CI/ACR/overlay/runbook 登记要求；以及本变更将 `cash-service` 补入发版清单的具体要求。

### Modified Capabilities

- （无）基线 `openspec/specs/` 下尚无已归档的 cash 发版 capability；`vip-cash-service` 的 `cash-service-runtime` 已要求 Compose/env/runbook，本变更补强 CI/workflow 缺口并以新 capability 固化。

## Impact

- **CI**：`.github/workflows/docker-acr.yml`（全量与 `+cash` 部分构建）。
- **部署**：`manifest/docker/docker-compose.microservices.test.yml`、`docker-compose.microservices.prod.yml`。
- **文档**：`docs/runbooks/release-deploy-and-run.md`；全局约定 `openspec/project.md`（评审硬性检查项）。
- **无业务 API / DB 变更**；无 Flutter 变更。依赖既有 `Dockerfile.cash-service` 与基线 compose 中的 `cash-service` 服务定义。
- **运维**：ACR 未建仓时 push 会失败，须先建 `cash-service` 仓库。
