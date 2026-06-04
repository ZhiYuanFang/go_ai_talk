## ADDED Requirements

### Requirement: release-deploy-and-run SHALL 文档化 Compose 镜像版本控制

`docs/runbooks/release-deploy-and-run.md` SHALL 包含章节说明：`docker-compose.microservices.yml` 与 prod/test overlay 的关系；镜像仓库 `${REGISTRY}` 与 `${IMAGE_TAG}` 用法；测试默认 `develop`、生产 semver 钉扎；registry `pull` + `--no-build` 部署；禁止生产使用 `:local`/`develop`；按服务镜像 tag 回滚步骤。

#### Scenario: 运维按 runbook 回滚生产镜像

- **WHEN** 生产发版后需回滚至上一 semver
- **THEN** runbook SHALL 提供将 `.env.prod` 中 `IMAGE_TAG` 改回上一版本并 `pull` + `force-recreate` 的步骤

### Requirement: release-deploy-and-run SHALL 文档化生产测试双栈部署

`docs/runbooks/release-deploy-and-run.md` SHALL 包含生产/测试双栈对照表（网络、端口、库名、静态目录、中间件端口）、测试栈启动顺序（test 网络 → test Redis cluster → test Rabbit 初始化 → microservices test）、Nginx 反代 `test.pangbao.cuplay.top:9701/9702`、健康检查 URL（对外形态与生产一致仅域名不同）、脱敏种子刷新与发版前 checklist。

#### Scenario: 运维首次搭建测试栈

- **WHEN** 运维仅阅读 `release-deploy-and-run.md`
- **THEN** 其 SHALL 能按文档顺序完成测试中间件与微服务启动，并验证 `https://test.pangbao.cuplay.top:9702/api.json`（或 documented 等价 HTTPS 探活）

## MODIFIED Requirements

### Requirement: Runtime docs SHALL be centralized and governed

`dao-sync-by-domain.md` and `release-deploy-and-run.md` MUST be maintained in a dedicated runtime-docs directory, and change governance MUST require synchronized updates when runtime behavior changes. Changes that introduce or alter Compose prod/test dual-stack deployment, registry image tagging, or test seed desensitization MUST update `release-deploy-and-run.md` in the same change.

#### Scenario: Docs location is consolidated

- **WHEN** checking runtime operation documents
- **THEN** both target documents MUST exist under one dedicated new folder

#### Scenario: Governance requires synchronized update

- **WHEN** project runtime/deployment/database boundary rules change
- **THEN** project governance (`openspec/project.md`) MUST require updating both runtime docs

#### Scenario: Dual-stack change updates runbook

- **WHEN** a change adds prod/test Compose overlays or test seed procedures
- **THEN** `release-deploy-and-run.md` MUST be updated in that change before merge
