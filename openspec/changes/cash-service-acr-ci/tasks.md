## 1. 全局约定

- [x] 1.1 在 `openspec/project.md` 新增「微服务 CI / ACR 发版约定（强制）」：新增微服务 MUST 同步 `docker-acr.yml`、test/prod overlay、runbook 别名/服务数，并提醒 ACR 建仓
- [x] 1.2 在 `openspec/project.md`「重要约束」增加对应评审检查项

## 2. CI workflow

- [x] 2.1 更新 `.github/workflows/docker-acr.yml`：`resolve_service_alias` 增加 `cash`/`cash-service`；`dockerfile_for` 映射 `Dockerfile.cash-service`；`ALL_SERVICES` 追加 `cash-service`
- [x] 2.2 同步文件头注释、`workflow_dispatch.services` 描述与未知别名错误文案中的服务列表/「全量 10 服务」

## 3. Compose overlay

- [x] 3.1 `docker-compose.microservices.test.yml` 增加 `cash-service`：`${REGISTRY}/cash-service:${IMAGE_TAG}`、`go-ai-talk-cash-service-test`、`19807:9807`、test 网络
- [x] 3.2 `docker-compose.microservices.prod.yml` 增加 `cash-service`：同镜像变量、`9807:9807`、prod 网络

## 4. Runbook 与运维检查

- [x] 4.1 更新 `docs/runbooks/release-deploy-and-run.md`：docker-acr 相关全量服务数、别名表加入 `cash`/`cash-service`、必要时补充 `+cash` 示例
- [x] 4.2 自检：ACR `pangbao-test` / `pangbao-release` 是否已有仓库 `cash-service`（未建则运维创建；可记结论于 PR）— **结论：本环境无法访问 ACR 控制台；已在 runbook 写明须先建仓；合并前请运维确认两命名空间均有 `cash-service`**
- [x] 4.3 自检：`rg`/`grep` 确认 `docker-acr.yml` 的 `ALL_SERVICES` 含 cash，且 test/prod overlay 均含 `cash-service` 镜像行
