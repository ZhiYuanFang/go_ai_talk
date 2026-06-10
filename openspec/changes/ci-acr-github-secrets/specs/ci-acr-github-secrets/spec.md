## ADDED Requirements

### Requirement: CI ACR 凭证来自 GitHub Secrets 而非仓库 .env 文件

`docker-acr` workflow MUST NOT 依赖 git 仓库内存在的 `manifest/docker/.env.test` 或 `manifest/docker/.env.prod` 来获取 ACR 登录信息。构建 push 所需的 `REGISTRY`、`ACR_USERNAME`、`ACR_PASSWORD` MUST 从 GitHub Actions Secrets 加载。仓库 `.gitignore` SHALL 继续忽略 `**/.env.*`（`.env.example` 除外），且 MUST NOT 要求将含真实 ACR 密码的 `.env.test|prod` 提交至 git。

#### Scenario: tag push 触发测试环境构建

- **WHEN** 开发者 push git tag `v2.0.0-beta.1` 且 GitHub Environment `test` 已配置 `REGISTRY`、仓库 Secrets 已配置 `ACR_USERNAME` 与 `ACR_PASSWORD`
- **THEN** workflow SHALL 选择 `test` 环境、从 Secrets 读取 ACR 配置、成功 login 并向对应命名空间 push 七微服务镜像（含 tag 与 git sha）

#### Scenario: 缺少 Secrets 时明确失败

- **WHEN** workflow 运行且目标环境的 `REGISTRY` 或 `ACR_USERNAME` 或 `ACR_PASSWORD` 未配置
- **THEN** workflow SHALL 在 ACR login 之前失败，并输出可操作的错误信息（指明缺失项与 runbook 配置章节）

#### Scenario: checkout 无 .env 文件仍可构建

- **WHEN** 仓库 checkout 后不存在 `manifest/docker/.env.test` 与 `.env.prod`
- **THEN** `docker-acr` workflow SHALL 仍能完成镜像 build 与 push（不因「缺少环境文件」失败）

### Requirement: CI push 地址从 REGISTRY 推导公网域名

workflow SHALL 从 Secrets 中的 `REGISTRY`（允许含 `-vpc` 专线域名）推导 CI push 用的公网 registry 地址（去掉 host 中的 `-vpc` 段）。若推导结果仍含 `-vpc` 或缺少命名空间路径，workflow MUST 失败并给出格式说明。

#### Scenario: REGISTRY 含 vpc 域名

- **WHEN** Environment `test` 的 `REGISTRY` 为 `crpi-xxx-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test`
- **THEN** workflow SHALL 使用 `crpi-xxx.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test` 作为 push 地址

#### Scenario: REGISTRY 格式非法

- **WHEN** `REGISTRY` 不含 `/` 命名空间段
- **THEN** workflow MUST 失败并提示格式须为 `<域名>/<命名空间>`

### Requirement: tag 路由规则保持不变

workflow SHALL 保留现有环境选择规则：`workflow_dispatch` 使用输入 `target_env`；tag push 时 `vMAJOR.MINOR.PATCH`（无预发布后缀）→ `prod`；其余 `v*` 预发布 tag → `test`。

#### Scenario: 正式 semver tag 路由生产

- **WHEN** push tag `v2.0.3`
- **THEN** workflow SHALL 使用 GitHub Environment `prod` 的 Secrets

#### Scenario: 预发布 tag 路由测试

- **WHEN** push tag `v2.0.3-beta.2`
- **THEN** workflow SHALL 使用 GitHub Environment `test` 的 Secrets

### Requirement: Runbook 区分 CI Secrets 与 ECS .env

`docs/runbooks/release-deploy-and-run.md` SHALL 文档化：GitHub Actions 使用 Environments/Secrets 配置 ACR；ECS 部署使用本地 `manifest/docker/.env.test|prod`（从 `.env.example` 复制，不上传 git）。文档 MUST 列出所需 Secret 名称与各 Environment 的 `REGISTRY` 配置方式，并 MUST NOT 再声明「无需 GitHub Secrets」。

#### Scenario: 运维按 runbook 配置 CI

- **WHEN** 运维阅读 runbook「ACR 与 CI 凭证」章节
- **THEN** 文档 SHALL 提供 `ACR_USERNAME`、`ACR_PASSWORD`、`REGISTRY`（test/prod 分环境）的配置步骤与验证方式（如 workflow_dispatch）

#### Scenario: 运维按 runbook 配置 ECS

- **WHEN** 运维在 ECS 部署测试栈
- **THEN** runbook SHALL 说明 `.env.test` 仍须含 `REGISTRY`、`IMAGE_TAG`、`ACR_*` 等完整字段，且该文件仅存在于服务器、不进 git

### Requirement: ACR 密码不得出现在 workflow 日志

workflow MUST 对 `ACR_PASSWORD` 使用 GitHub `::add-mask::`（或等价机制），防止明文密码写入 Actions 日志。

#### Scenario: 成功 login 后日志无密码

- **WHEN** workflow 完成 ACR login
- **THEN** Actions 日志 SHALL NOT 包含 `ACR_PASSWORD` 明文
