## ADDED Requirements

### Requirement: 新增微服务 MUST 同步 CI 与发版清单

凡变更引入或首次对外部署新的微服务进程（含 `cmd/*-service` 与对应 `Dockerfile.*`），OpenSpec proposal/design/tasks 与实现 MUST 同步更新下列发版清单，不得仅完成基线 Compose / 本地 Dockerfile：

1. **GitHub Actions**：至少 `.github/workflows/docker-acr.yml`（服务别名、`dockerfile_for`、`ALL_SERVICES`、全量服务数注释与错误提示）。
2. **环境 overlay**：`manifest/docker/docker-compose.microservices.test.yml` 与 `docker-compose.microservices.prod.yml` 中的 ACR 镜像、`container_name`（若适用）、ports（若服务暴露 HTTP）、networks。
3. **发布 runbook**：`docs/runbooks/release-deploy-and-run.md` 中全量服务数、构建范围别名表与相关说明。
4. **ACR 仓库**：提醒在目标命名空间创建与 matrix `id` 一致的仓库（运维操作，可记在 tasks）。

全局约定 MUST 写入 `openspec/project.md`（微服务 CI / ACR 发版约定），并作为评审硬性检查项。

#### Scenario: propose 新微服务时列出 CI 任务

- **WHEN** 某变更提案新增独立微服务进程
- **THEN** tasks MUST 含更新 `docker-acr.yml`、test/prod overlay 与 runbook 别名/服务数的可勾选项，且 `openspec/project.md` MUST 已存在对应强制约定可供引用

#### Scenario: 仅改基线 Compose 视为不完整

- **WHEN** 评审发现新服务仅出现在 `docker-compose.microservices.yml` 而未出现在 `docker-acr.yml` 的 `ALL_SERVICES`
- **THEN** 评审 MUST 要求补齐后再合并（或等价记录明确豁免理由）

### Requirement: cash-service 纳入 ACR 构建矩阵

`.github/workflows/docker-acr.yml` MUST 将 `cash-service` 列入全量构建列表，MUST 支持别名 `cash` 与 `cash-service` 解析为 matrix id `cash-service`，MUST 将 dockerfile 映射为 `manifest/docker/Dockerfile.cash-service`。全量构建服务数文案 MUST 与列表长度一致（本变更后为 10）。

#### Scenario: 全量 tag 构建含 cash-service

- **WHEN** 推送不含 `+` 后缀的预发布或正式 tag（或 workflow_dispatch 且 services 留空）
- **THEN** 构建 matrix MUST 包含 id `cash-service` 且使用 `Dockerfile.cash-service`

#### Scenario: 部分构建仅 cash

- **WHEN** git tag 带 `+cash` 后缀，或 workflow_dispatch 的 `services` 为 `cash`
- **THEN** CI MUST 仅 build/push `cash-service`（不强制 retag 其他服务）

### Requirement: cash-service 纳入 test/prod Compose overlay

测试与生产 overlay MUST 声明服务 `cash-service`，镜像为 `${REGISTRY}/cash-service:${IMAGE_TAG}`，并加入对应外部网络。test overlay MUST 映射宿主机端口 `19807`→容器 `9807`；prod overlay MUST 映射 `9807`→`9807`。业务环境变量仍由基线 compose + env-file 注入，overlay MUST NOT 重复书写业务 env（与现有 overlay 惯例一致）。

#### Scenario: 测试栈可 pull cash 镜像

- **WHEN** 使用 `.env.test` 与 microservices + test overlay 执行 `compose pull`
- **THEN** MUST 能解析并拉取 `${REGISTRY}/cash-service:${IMAGE_TAG}`（在 ACR 已存在该 tag 的前提下）

#### Scenario: 生产栈端口不与测试冲突

- **WHEN** 同机分别使用 test / prod overlay 启动
- **THEN** cash-service 宿主机端口 MUST 分别为 `19807` 与 `9807`，MUST NOT 互相占用
