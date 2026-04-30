## 1. Worker 专属配置切换

- [x] 1.1 新增 `manifest/config/config.worker-service.yaml`，仅保留 worker 所需的 server/logger/database.default 基础项。
- [x] 1.2 更新 `cmd/worker-service/main.go` 默认配置加载逻辑，未设置 `GF_GCFG_FILE` 时回退 worker 专属配置。
- [x] 1.3 更新 `manifest/docker/docker-compose.microservices.yml` 中 worker 的 `GF_GCFG_FILE` 指向 worker 专属配置。
- [x] 1.4 更新 `manifest/deploy/kustomize/base/worker-deployment.yaml` 的 `GF_GCFG_FILE` 指向 worker 专属配置。
- [x] 1.5 更新 `manifest/docker/Dockerfile.worker-service` 的默认配置环境变量与 COPY 路径为 worker 专属配置。

## 2. 主配置瘦身与 DAO 访问模型收敛

- [x] 2.1 从 `manifest/config/config.yaml` 删除数据库配置，仅保留网关与公共项。
- [x] 2.2 删除 `internal/dao/domain_db.go`，移除多数据库组选择与 fallback 逻辑。
- [x] 2.3 全仓扫描 DAO 调用路径，确保不再依赖 `database.<group>.link` 多组配置。
- [x] 2.4 评估并整理 `internal/dao/*_ext.go`：无增量价值的删除/合并，有业务增量的保留并补充注释说明。

## 3. 运行文档迁移与治理规则更新

- [x] 3.1 新建运行文档目录（如 `docs/runbooks/`），迁移 `dao-sync-by-domain.md` 与 `release-deploy-and-run.md` 到该目录。
- [x] 3.2 基于当前微服务边界更新 `dao-sync-by-domain.md`（单服务单库、default-only、跨服务走 API）。
- [x] 3.3 基于当前发布路径更新 `release-deploy-and-run.md`（worker 专属配置、主配置无数据库项）。
- [x] 3.4 在 `openspec/project.md` 增加治理约束：后续需求变更涉及运行/发布/DAO 边界时，必须同步更新上述两份运行文档。
- [x] 3.5 在 `openspec/project.md` 固化架构原则：每服务一个数据库，跨服务数据通过 API 获取，各服务配置仅维护 `database.default`。

## 4. 验证与回滚

- [x] 4.1 执行编译校验：`go test ./cmd/... ./internal/...` 确认配置与 DAO 收敛后可通过。
- [ ] 4.2 验证 worker 在专属配置下可启动，且 outbox relay 可正确读取并更新 `domain_outbox` 状态。
- [ ] 4.3 验证 gateway 在主配置去数据库后可正常启动并保持代理链路可用。
- [x] 4.4 制定回滚步骤：worker 配置路径回切、文档路径回切、DAO 访问模型回退（仅在应急时启用）。
