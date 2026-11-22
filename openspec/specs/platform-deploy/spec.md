# platform-deploy

## Purpose

Compose/Docker/CI/ACR 发版、双栈环境、资源限制、主机卷与 cutover 运维的行为规格。 Covers release pipelines and deploy topology for go_ai_talk services.

<!-- migrated-from: ci-acr-github-secrets, compose-container-resource-limits, compose-host-root-asset-volumes, compose-mysql-endpoint-via-env, compose-mysql-test-seed-desensitization, compose-prod-test-dual-stack, compose-redis-topology-2g, docker-deploy-logging, microservice-ci-release, prod-ecs-cutover -->

## Requirements

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

workflow SHALL 保留现有环境选择规则：`workflow_dispatch` 使用输入 `target_env`；**tag push 时以 base tag（`+` 之前）判定**：`vMAJOR.MINOR.PATCH`（无预发布后缀）→ `prod`；其余 `v*` 预发布 base tag → `test`。git tag 含 `+` 构建后缀 MUST NOT 改变上述路由逻辑。

#### Scenario: 正式 semver tag 路由生产

- **WHEN** push tag `v2.0.3`
- **THEN** workflow SHALL 使用 GitHub Environment `prod` 的 Secrets

#### Scenario: 预发布 tag 路由测试

- **WHEN** push tag `v2.0.3-beta.2`
- **THEN** workflow SHALL 使用 GitHub Environment `test` 的 Secrets

#### Scenario: 带构建后缀的预发布 tag 仍路由测试

- **WHEN** push tag `v2.0.3-beta.2+ucg`
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

### Requirement: docker-acr SHALL support tag build-scope suffix for selective service builds

`docker-acr` workflow SHALL 支持 git tag **`+` 构建范围后缀**：`+` **之前** 的片段为 **base tag**（用作 ACR 主 tag、`primary_tag`、与服务器 `IMAGE_TAG` 一致）； 无 `+` 后缀时 SHALL 构建全部 6 个微服务（与变更前行为一致）。 workflow MUST NOT 对未纳入构建范围的服务执行 retag 或 manifest 复制。 该状态 SHALL 视为预期，而非 workflow 失败条件。 服务别名 MUST 至少支持：`gateway`、`gateway-app`、`history`/`history-service`、`voice`/`voice-service`、`device`/`device-service`、`ucg`/`ucg-service`； `all` MUST 表示全量 6 服务。 非法别名 MUST 导致 workflow 在 build 前失败并输出可操作的错误信息。 test/prod 环境路由.

#### Scenario: 预发布 tag 仅构建 ucg

- **WHEN** 开发者 push git tag `v1.0.0-rc.4+ucg` 且 Secrets 配置正确
- **THEN** workflow SHALL 仅 build/push `ucg-service` 至 `:${v1.0.0-rc.4}` 与 `:${git_sha}`，且 SHALL NOT push 其他五服务镜像至 `v1.0.0-rc.4`

#### Scenario: 无后缀全量构建保持不变

- **WHEN** 开发者 push git tag `v1.0.0-rc.4`（无 `+`）
- **THEN** workflow SHALL build/push 全部 6 个微服务至 `:${v1.0.0-rc.4}` 与 `:${git_sha}`

#### Scenario: 非法服务别名失败

- **WHEN** push tag `v1.0.0-rc.4+unknown-svc`
- **THEN** workflow MUST 失败且 MUST NOT push 任意镜像

#### Scenario: base tag 环境路由不受后缀影响

- **WHEN** push tag `v2.0.0-rc.1+ucg`
- **THEN** workflow SHALL 使用 GitHub Environment `test`（因 base tag 为预发布），且 push 主 tag MUST 为 `v2.0.0-rc.1`（不含 `+ucg`）

### Requirement: workflow_dispatch SHALL accept optional services scope

手动触发 `docker-acr` 时，workflow SHALL 支持可选输入 `services`（逗号分隔别名，空=全量 6 服务），语义 MUST 与 tag `+` 后缀一致。`image_tag` 输入 MUST 为 base tag（不含 `+` 后缀）。

#### Scenario: 手动仅构建 ucg

- **WHEN** 运维 workflow_dispatch 选择 `target_env=test`、`image_tag=v1.0.0-rc.4`、`services=ucg`
- **THEN** workflow SHALL 仅 build/push `ucg-service` 至 `:v1.0.0-rc.4`

---

### Requirement: 生产与测试 Compose SHALL 定义容器 CPU 与内存上限

仓库 SHALL 提供 **`manifest/docker/docker-compose.resources.prod.yml`** 与 **`manifest/docker/docker-compose.resources.test.yml`**（或后继等价 overlay），为下列组件定义 **`mem_limit`** 与 **`cpus`**（或 compose 规范中等价、在非 Swarm 模式下对 `docker compose up` 生效的字段）：

- 生产/测试 **全部** 微服务（gateway、gateway-app、history、voice、device、worker、ucg）
- 生产 Redis、测试 Redis、生产/测试 RabbitMQ

runbook SHALL 文档化默认配额表及「2G ECS survival profile」说明。`voice-service` 测试实例 SHALL 拥有 **不低于** 其它微服务的 memory limit（ documented 起步值 **512M**）。

#### Scenario: 启动命令叠加 resources overlay

- **WHEN** 运维按 runbook 启动生产微服务并叠加 `-f docker-compose.resources.prod.yml`
- **THEN** `docker inspect` 或 `docker stats` SHALL 显示对应容器配置了 memory/cpu 上限

#### Scenario: 本地开发不受 prod/test limits 强制约束

- **WHEN** 开发者仅使用基线 `microservices.yml` + `microservices.local.yml` 且 **不** 叠加 `resources.*.yml`
- **THEN** 本地容器 **MAY** 无 cgroup 上限（便于调试）

### Requirement: limits SHALL 防止单容器耗尽宿主机

资源上限的配置意图 SHALL 在 runbook 中说明：当某容器内存超过 `mem_limit` 时，内核 **MAY** OOM kill 该容器，**SHALL NOT** 无限制占用同机其它栈（含 MySQL 宿主机进程）的全部物理内存。runbook SHALL 包含 OOM 排查步骤（`dmesg`、`docker stats`、调高 voice-test limit 等）。

#### Scenario: 文档化 OOM 语义

- **WHEN** 运维查阅 `release-deploy-and-run.md` 资源 limits 章节
- **THEN** 文档 SHALL 说明 limits 与宿主机 2G 物理内存的关系，以及 ASR 验收时优先保障 test voice 的建议

---

### Requirement: device-service 事件 logo 持久化到宿主机根目录

在 Docker Compose 部署下，device-service 容器 SHALL 通过 bind mount 将 **`/ai_talk_images`** 映射到 Linux 宿主机同路径 **`/ai_talk_images`**，使 `SaveEventLogo` 写入的文件出现在宿主机上。

#### Scenario: 上传 logo 后宿主机可见

- **WHEN** 管理员通过 API 上传事件 logo 且 device-service 使用默认 `eventImageStorageDir` `/ai_talk_images/`
- **THEN** 宿主机路径 `/ai_talk_images/` 下 SHALL 存在对应文件
- **AND** 容器内同路径 SHALL 可读取该文件

#### Scenario: 容器重建后文件保留

- **WHEN** 宿主机 `/ai_talk_images` 已存在 logo 文件且运维对 device-service 执行 `docker compose up --force-recreate`
- **THEN** 重建后容器 SHALL 仍能读取宿主机挂载目录中的同名文件

### Requirement: gateway-app APK 持久化到宿主机根目录

在 Docker Compose 部署下，gateway-app 容器 SHALL 通过 bind mount 将 **`/apk/ai_talk`** 映射到 Linux 宿主机同路径 **`/apk/ai_talk`**，使版本管理上传的 APK 出现在宿主机上。

#### Scenario: 上传 APK 后宿主机可见

- **WHEN** 管理员通过版本管理接口上传 APK 且 gateway-app 使用默认 `apkStorageDir` `/apk/ai_talk/`
- **THEN** 宿主机路径 `/apk/ai_talk/` 下 SHALL 存在对应 `.apk` 文件

#### Scenario: 容器重建后 APK 保留

- **WHEN** 宿主机 `/apk/ai_talk` 已存在 APK 且运维对 gateway-app 执行 `docker compose up --force-recreate`
- **THEN** 重建后 gateway-app SHALL 仍能通过 `GET /device/app/apk/` 提供该文件

### Requirement: 挂载路径与配置默认一致

Compose 卷挂载点 SHALL 与代码/配置默认存储目录一致：`/ai_talk_images`（device）、`/apk/ai_talk`（gateway-app）。未通过环境变量修改存储路径时，SHALL NOT 要求额外配置即可满足本需求。

#### Scenario: 默认配置下路径一致

- **WHEN** 未设置 `DEVICE_EVENT_IMAGE_STORAGE_DIR` 与 `GATEWAY_APP_APK_STORAGE_DIR`
- **THEN** 写盘路径与 bind mount 目标路径 SHALL 均为上述宿主机根下目录

### Requirement: 部署文档说明宿主机准备

项目 runbook 或等价部署文档 SHALL 说明：Linux Docker **生产**部署前建议执行 `mkdir -p /ai_talk_images /apk/ai_talk`；**测试**部署前建议执行 `mkdir -p /ai_talk_images_test /apk/ai_talk_test`；并给出验证宿主机与容器内文件一致的示例命令。

#### Scenario: 运维可按文档验收生产静态目录

- **WHEN** 运维按文档创建生产目录并启动 prod compose 后上传 logo 与 APK
- **THEN** 文档中的 `ls` 或 `docker exec` 验收步骤 SHALL 能确认宿主机 `/ai_talk_images` 与 `/apk/ai_talk` 非空

#### Scenario: 运维可按文档验收测试静态目录

- **WHEN** 运维按文档创建测试目录并启动 test compose 后上传 logo 与 APK
- **THEN** 文档中的验收步骤 SHALL 能确认宿主机 `/ai_talk_images_test` 与 `/apk/ai_talk_test` 非空

### Requirement: 测试栈事件 logo 持久化到独立宿主机目录

在 Docker Compose **测试** overlay 部署下，device-service 容器 SHALL 通过 bind mount 将容器内 **`/ai_talk_images`** 映射到 Linux 宿主机 **`/ai_talk_images_test`**，与生产目录 `/ai_talk_images` 隔离。

#### Scenario: 测试上传 logo 不写入生产目录

- **WHEN** 管理员在测试环境上传事件 logo 且 device-service 使用默认 `eventImageStorageDir` `/ai_talk_images/`
- **THEN** 文件 SHALL 出现在宿主机 `/ai_talk_images_test/` 下
- **AND** 宿主机 `/ai_talk_images/`（生产）SHALL NOT 因该上传而新增或修改同名文件

### Requirement: 测试栈 APK 持久化到独立宿主机目录

在 Docker Compose **测试** overlay 部署下，gateway-app 容器 SHALL 通过 bind mount 将容器内 **`/apk/ai_talk`** 映射到 Linux 宿主机 **`/apk/ai_talk_test`**，与生产目录 `/apk/ai_talk` 隔离。

#### Scenario: 测试上传 APK 不写入生产目录

- **WHEN** 管理员在测试环境版本管理页上传 APK
- **THEN** 文件 SHALL 出现在宿主机 `/apk/ai_talk_test/` 下
- **AND** 宿主机 `/apk/ai_talk/`（生产）SHALL NOT 因该上传而新增或修改同名文件

---

### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入各服务 MySQL 连接串

yml`（或其后继官方等价文件）SHALL 允许在**不修改已提交 YAML 内口令占位**的前提下，通过环境变量为 `history-service`、`device-service`、`voice-service`、`worker`、`ucg-service` 及 `gateway-app`（`APP_DB_LINK`）注入数据库连接：其中 history/device/voice/ucg SHALL 分别支持 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`、`UCG_DB_LINK` 覆盖默认库； worker SHALL 支持 `WORKER_OUTBOX_DB_LINK`（由 cmd 写入 `GF_DATABASE_OUTBOX_LINK`）； gateway-app SHALL 支持 `APP_DB_LINK` 覆盖 `database. env` 文件 SHALL 分别注入对应库名。

#### Scenario: device 使用注入的 link 启动

- **WHEN** 部署者在启动 Compose 前设置 `DEVICE_DB_LINK` 为合法 MySQL DSN
- **THEN** `device-service` 进程 SHALL 使用该 DSN 作为 `GF_DATABASE_DEFAULT_LINK`，而不依赖仅写在镜像内配置文件中的占位地址

#### Scenario: 测试 worker 使用 test outbox 库

- **WHEN** 测试栈设置 `WORKER_OUTBOX_DB_LINK` 指向 `ai_voice_worker_test`
- **THEN** worker-service SHALL 使用该 DSN 作为 outbox 库连接，SHALL NOT 写入生产 `ai_voice_worker`

### Requirement: 参考 Compose MUST 为访问宿主机 MySQL 提供 host.docker.internal 解析

当 MySQL 监听在 **运行 Docker 的宿主机** 上且业务容器使用 bridge 网络时，参考 Compose 中需访问该 MySQL 的服务 SHALL 配置 `extra_hosts`，使主机名 `host.docker.internal` 解析到宿主机（例如 `host-gateway` 语义），以便连接串中使用 `tcp(host.docker.internal:3306)` 等地址时行为可预期。

#### Scenario: Linux 上 compose up 后容器解析 host.docker.internal

- **WHEN** 在支持 `host-gateway` 的 Docker Engine 上执行 `docker compose up` 使用该参考文件
- **THEN** 业务容器内 SHALL 能将 `host.docker.internal` 解析到宿主机侧地址，从而可与宿主机上监听的 MySQL 建立 TCP 连接（在 DSN 已正确配置且 mysqld 对 Docker 网桥来源放行时）

### Requirement: 仓库 MUST 提供无密钥的 Compose 数据库环境样例

仓库 SHALL 提供一份可复制为本地 `.env` 的示例文件（例如 `manifest/docker/.env.example`），其中 SHALL 用中文或英文注释说明：**MySQL 与 Docker 同机**时推荐将主机设为 `host.docker.internal`；**MySQL 在其它机器**时将主机设为从容器网络可达的 DNS 或 IP（如 RDS、内网 IP），且 SHALL NOT 包含真实生产口令。

#### Scenario: 新成员首次接 Compose 栈

- **WHEN** 开发者复制示例为 `.env` 并按注释填写自己的 MySQL 拓扑
- **THEN** 其 SHALL 能区分同机与异机两种填法，且无需从 git 历史中寻找口令

### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入 Redis 地址

`manifest/docker/docker-compose.microservices.yml`（或其后继官方等价文件）SHALL 允许通过环境变量 **`GF_REDIS_DEFAULT_ADDRESS`** 覆盖 GoFrame `redis.default.address`，作用于 **所有** 依赖 Redis 的微服务（含 gateway、gateway-app、history、voice、device、worker、ucg）。当变量为空或未设置时，SHALL 回退镜像内 yaml 默认地址（cluster 三主种子）。`.env.test.example` SHALL 文档化测试单机地址 `redis-test:6379`；`.env.prod.example` **SHALL NOT** 要求填写该变量（生产使用 yaml 默认 cluster 种子）。

#### Scenario: 测试栈注入单机 Redis 地址

- **WHEN** `.env.test` 设置 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379` 且启动测试微服务栈
- **THEN** 各服务容器环境 SHALL 包含该变量，且 Redis 客户端 SHALL 连接 `redis-test:6379`

#### Scenario: 生产栈不注入时沿用 yaml cluster 种子

- **WHEN** 生产 `.env.prod` 未设置 `GF_REDIS_DEFAULT_ADDRESS`
- **THEN** 微服务 SHALL 使用 config yaml 中的 `redis-node-1:7001,redis-node-2:7002,redis-node-3:7003`

### Requirement: 仓库 MUST 提供分环境的 Compose 数据库环境样例

除现有 `manifest/docker/.env.example` 外，仓库 SHALL 提供 `manifest/docker/.env.prod.example` 与 `manifest/docker/.env.test.example`。prod 示例 SHALL 说明各 `*_DB_LINK` 指向生产库名（无 `_test` 后缀）及 `IMAGE_TAG` 为 semver。test 示例 SHALL 说明各 `*_DB_LINK` 指向 `ai_voice_*_test` 库、`IMAGE_TAG=develop`、`GATEWAY_APP_PUBLIC_BASE_URL=https://test.pangbao.cuplay.top:9702`，且 SHALL NOT 包含真实生产口令。

#### Scenario: 新运维区分 prod 与 test env 文件

- **WHEN** 运维复制 `.env.test.example` 为 `.env.test` 并按注释填写
- **THEN** 其 SHALL 能将 `HISTORY_DB_LINK` 指向 `ai_voice_history_test` 且将 `IMAGE_TAG` 设为 `develop`，且 SHALL NOT 误用生产库 DSN

### Requirement: 测试栈 Compose MUST 支持 Redis 地址环境注入

测试 overlay 或 `.env.test.example` SHALL 文档化并通过 compose `environment` 注入 `GF_REDIS_DEFAULT_ADDRESS`，指向测试 Redis cluster 三主种子（测试网络内服务名与端口，与 prod 物理隔离）。

#### Scenario: 测试 gateway-app 使用 test Redis

- **WHEN** 测试栈 gateway-app 启动且 `GF_REDIS_DEFAULT_ADDRESS` 已按 `.env.test.example` 配置为 test cluster 种子
- **THEN** 版本检查等 Redis 缓存 SHALL 读写 test cluster，SHALL NOT 依赖 prod cluster 的节点地址

---

### Requirement: 测试 MySQL 库 SHALL 与生产库名隔离

测试环境各业务库 SHALL 使用与生产对应、带 `_test` 后缀的库名（至少包含 `ai_voice_history_test`、`ai_voice_device_test`、`ai_voice_voice_test`、`ai_voice_worker_test`、`ai_voice_app_test`、`ai_voice_ucg_test`）。`.env.test` 中各 `*_DB_LINK` SHALL 指向上述测试库，MUST NOT 指向生产库名。

#### Scenario: 测试 device-service 连接测试库

- **WHEN** 测试栈 device-service 启动且 `DEVICE_DB_LINK` 已按 `.env.test.example` 配置
- **THEN** 进程 SHALL 连接 `ai_voice_device_test`（或 documented 等价名），SHALL NOT 连接 `ai_voice_device`

### Requirement: 仓库 SHALL 提供生产到测试的脱敏种子流程

仓库 SHALL 在 runbook 和/或 `hack/mask-seed-data.sh` 中描述可重复流程：从生产 `ai_voice_*` 导出 → 脱敏 → 导入 `ai_voice_*_test`。脱敏 SHALL 至少处理：用户手机号、微信 openid/unionid、refresh token/session、设备标识（替换或前缀化）。导入 SHALL 覆盖测试库既有数据（运维须在 runbook 中警告）。

#### Scenario: 脱敏后测试库无原始手机号

- **WHEN** 运维按文档完成脱敏 import
- **THEN** 测试库 user 相关表中 SHALL NOT 保留与生产 export 完全相同的手机号明文

#### Scenario: 发版前刷新测试种子

- **WHEN** 准备发布新的 release candidate
- **THEN** runbook SHALL 要求（或 recommend 作为 checklist 必项）在测试验收前执行一次脱敏种子刷新

### Requirement: 脱敏种子 SHALL 同步测试静态资源

当生产种子包含 `/ai_talk_images/` 路径引用时，运维 SHALL 将对应 logo 文件同步至宿主机 `/ai_talk_images_test/`（或 test overlay documented 路径），使测试管理页与 App 静态读链路可验收。

#### Scenario: 测试环境 logo 可读

- **WHEN** 测试库 event 行引用 `/ai_talk_images/<file>` 且文件已同步至测试静态目录
- **THEN** 经 test gateway 或 gateway-app 反代的静态请求 SHALL 返回 200

---

### Requirement: 仓库 SHALL 提供生产与测试双栈 Compose overlay

仓库 SHALL 在 `manifest/docker/` 提供 `docker-compose.microservices.prod.yml` 与 `docker-compose.microservices.test.yml`，与基线 `docker-compose.microservices.yml` 组合使用。prod/test overlay SHALL 使用 `${REGISTRY}/<service>:${IMAGE_TAG}` 引用镜像仓库（如 `${REGISTRY}/gateway:${IMAGE_TAG}`，无 `go-ai-talk/` 路径前缀，以适配阿里云 ACR 等单段仓库名），且 SHALL NOT 包含 `build` 段。基线文件 MAY 保留 `build` 与 `:local` 供本机开发。

#### Scenario: 测试栈从 registry pull 启动

- **WHEN** 运维设置 `REGISTRY`、 `IMAGE_TAG=develop` 并执行 `docker compose -f ...microservices.yml -f ...microservices.test.yml pull && up -d --no-build`
- **THEN** 各业务容器 SHALL 使用 registry 中 `:develop` 镜像启动，且 SHALL NOT 在宿主机执行源码 build

#### Scenario: 生产栈使用 semver tag

- **WHEN** 运维在 `.env.prod` 设置 `IMAGE_TAG=v1.0.0` 并 pull + up
- **THEN** 生产容器 SHALL 使用 `:v1.0.0` 镜像，且 SHALL NOT 使用 `:develop` 或 `:local`

### Requirement: 生产与测试 SHALL 使用独立 Docker 网络完全隔离

生产栈与测试栈 SHALL 分别仅加入独立的 external Docker 网络（约定名 `go-ai-talk-prod-net` 与 `go-ai-talk-test-net`）。同一宿主机上 prod 与 test 的中间件与微服务 SHALL NOT 共用同一 bridge 网络的 DNS 解析。

#### Scenario: test 网络内 rabbitmq 不可被 prod 容器解析

- **WHEN** prod 与 test 栈同时运行且各自 RabbitMQ 仅加入对应网络
- **THEN** prod 容器内 SHALL NOT 通过服务名 `rabbitmq` 解析到 test 的 RabbitMQ 实例

### Requirement: 测试栈 SHALL 独立 Redis Cluster 与 RabbitMQ

仓库 SHALL 提供 `docker-compose.redis-cluster.test.yml` 与 `docker-compose.rabbitmq.test.yml`。测试 Redis cluster 宿主机映射端口 SHALL 使用 17001–17006（或与 prod 7001–7006 不冲突的 documented 端口段）。测试 RabbitMQ 宿主机映射 SHALL 使用 5673/15673（或与 prod 5672/15672 不冲突的 documented 端口段）。测试微服务 SHALL 通过环境变量 `GF_REDIS_DEFAULT_ADDRESS` 与 `MQ_HTTP_API_BASE` 指向 test 网络内中间件。

#### Scenario: 测试 history 与 worker 使用 test RabbitMQ

- **WHEN** 测试栈 `history-service` 与 `worker` 已启动且 `OUTBOX_RELAY_ENABLED`/`MQ_CONSUMER_ENABLED` 为 true
- **THEN** 二者 SHALL 仅与 test 网络内 RabbitMQ 通信，且 prod worker SHALL NOT 消费 test 队列中的消息

### Requirement: 测试栈后端端口 SHALL 与生产错开

测试栈微服务宿主机端口映射 SHALL 为：gateway 19701、gateway-app 19702、history 19801、voice 19802、device 19803、ucg 19804、worker 19901（或与 runbook  documented 表一致且不与 prod 9701–9901 冲突）。测试栈 container_name SHALL 与 prod 不同（例如带 `-test` 后缀或使用 `COMPOSE_PROJECT_NAME` 前缀）。

#### Scenario: 同机 prod 与 test 同时监听

- **WHEN** prod 与 test 栈同时 up
- **THEN** 宿主机 SHALL 可同时访问 `127.0.0.1:9701`（prod gateway）与 `127.0.0.1:19701`（test gateway）且无端口绑定冲突

### Requirement: 测试对外访问形态 SHALL 与生产一致

测试环境对外入口 SHALL 为 `https://test.pangbao.cuplay.top:9701`（主网关）与 `https://test.pangbao.cuplay.top:9702`（App 网关），由 Nginx（或等价反代）转发至测试后端 19701/19702。测试栈 SHALL 设置 `GATEWAY_APP_PUBLIC_BASE_URL` 为 `https://test.pangbao.cuplay.top:9702`（或 runbook documented 等价 HTTPS 基址）。

#### Scenario: 客户端仅换域名访问测试 App 网关

- **WHEN** 客户端将生产基址 `www.pangbao.cuplay.top:9702` 换为 `test.pangbao.cuplay.top:9702` 且路径不变（如 `/device/app/api/version/check`）
- **THEN** 请求 SHALL 到达测试 gateway-app 且 API 路径语义与生产一致

### Requirement: 镜像 tag 策略 SHALL 区分测试浮动与生产钉死

测试默认 `IMAGE_TAG=develop`（CI 覆盖的浮动 tag）。生产 MUST 使用 semver release tag（如 `v1.0.0`），MUST NOT 在生产 `.env.prod` 中使用 `develop` 或 `latest`。CI SHOULD 同时 push 不可变 `:<git-sha>` tag 供排错。

#### Scenario: 生产 env 拒绝 develop

- **WHEN** 运维检查生产部署配置
- **THEN** `.env.prod` 中 `IMAGE_TAG` SHALL 为 semver 形式且 SHALL NOT 等于 `develop`

### Requirement: 部分 CI 构建发版 SHALL 支持按服务 pull 与 up

当 CI 使用带 `+` 构建后缀的 tag（如 `v1.0.0-rc.4+ucg`）仅 push 部分服务镜像时，运维 MUST NOT 假设 ACR 上存在全部六服务 `:${base_tag}` 镜像。部署 MUST 通过 **按服务** `docker compose pull <service>` 与 `up -d --no-build <service>` 更新变更服务；对未构建服务 **MAY** 继续运行已拉取的旧 tag 本地镜像，直至下次无后缀全量 tag 发版。

`docs/runbooks/release-deploy-and-run.md` MUST 文档化：git tag 全名与 `.env` 中 `IMAGE_TAG`（base tag）的关系、`+ucg` 等后缀含义、按服务 pull/up 示例命令，以及全栈 `compose pull` 在部分构建 tag 下 **预期失败** 的说明（表示构建范围与部署操作不匹配）。

#### Scenario: 部分构建后仅更新 ucg

- **WHEN** CI 已对 tag `v1.0.0-rc.4+ucg` 仅 push `ucg-service:v1.0.0-rc.4`，且运维将 `.env.test` 中 `IMAGE_TAG` 设为 `v1.0.0-rc.4`
- **THEN** 运维 SHALL 执行 `pull`/`up` 仅针对 `ucg-service`；其他服务容器 MAY 保持上一版本镜像运行

#### Scenario: 全栈 pull 在部分 tag 下失败为预期

- **WHEN** ACR 不存在 `gateway:v1.0.0-rc.4`（因本次 CI 为 `+ucg` 部分构建）且运维执行全栈 `docker compose pull`
- **THEN** pull MUST 对缺失镜像报错；运维 SHALL 改为按服务 pull 或先打无后缀全量 tag 构建

---

### Requirement: 生产 Redis Cluster SHALL 为 3 主 0 从

`manifest/docker/docker-compose.redis-cluster.yml` SHALL 仅定义 **3** 个 Redis 服务（`redis-node-1`..`redis-node-3`），端口 **7001–7003**。仓库 runbook SHALL 文档化初始化命令：`redis-cli --cluster create` 仅包含上述三节点，且 **`--cluster-replicas 0`**。应用 config 中三主种子地址 `redis-node-1:7001,redis-node-2:7002,redis-node-3:7003` SHALL 与拓扑一致，**无需**为缩容修改 Go 代码。

#### Scenario: 生产 cluster 初始化成功

- **WHEN** 运维在空 volume 上启动 3 节点 compose 并执行 documented `cluster create`
- **THEN** `CLUSTER INFO` SHALL 报告 `cluster_state:ok`，且 `CLUSTER NODES` SHALL 显示 3 个 master、0 个 replica

#### Scenario: 生产微服务连接 Redis

- **WHEN** 生产微服务在 `go-ai-talk-net` 上启动且未设置 `GF_REDIS_DEFAULT_ADDRESS`
- **THEN** 进程 SHALL 通过 yaml 默认三主种子连接生产 3 节点 cluster

### Requirement: 测试 Redis SHALL 为单机 standalone

仓库 SHALL 提供 `manifest/docker/docker-compose.redis-standalone.test.yml`（或后继等价文件），定义 **单** Redis 服务（约定服务名 **`redis-test`**，容器端口 **6379**），且 **仅** 加入 `go-ai-talk-test-net`。**SHALL NOT** 要求测试栈执行 `redis-cli --cluster create`。测试栈 **SHALL NOT** 依赖 `docker-compose.redis-cluster.test.yml` 六节点拓扑作为默认路径。

#### Scenario: 测试 Redis 启动无需 cluster create

- **WHEN** 运维 `up -d` 测试 standalone Redis compose
- **THEN** 容器 running 后 SHALL 可直接 `redis-cli PING` 返回 `PONG`，且 **无需** cluster 初始化步骤

#### Scenario: 测试与生产 Redis 网络隔离

- **WHEN** 生产与测试栈同时运行
- **THEN** 测试 Redis 容器 SHALL 不在 `go-ai-talk-net` 上，生产 Redis 容器 SHALL 不在 `go-ai-talk-test-net` 上

### Requirement: 测试 Redis 地址 SHALL 经环境变量注入单机地址

测试部署 MUST 通过 `manifest/docker/.env.test` 设置 **`GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`**（或 runbook documented 等价单地址）。基线 `docker-compose.microservices.yml` SHALL 为需 Redis 的服务提供 `${GF_REDIS_DEFAULT_ADDRESS:-}` 注入；未设置时 SHALL 回退 yaml 默认 cluster 种子（供生产/local cluster 使用）。

#### Scenario: 测试微服务读写 test 单机 Redis

- **WHEN** 测试栈微服务启动且 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`
- **THEN** 经 `internal/platform/cachekit`（如启动探活 Ping）或 `internal/platform/redismsgkit` 建立的 Redis 连接 SHALL 指向测试单机 `redis-test:6379`，**SHALL NOT** 连接生产 cluster 节点

---

### Requirement: 生产与测试 Compose 栈 MUST 限制容器 json-file 日志保留量

`manifest/docker` 下用于 **生产** 与 **测试** 长期运行的 Compose 服务（微服务基线六件套、RabbitMQ prod/test、Redis prod cluster / test standalone）MUST 配置 `logging.driver=json-file`，且 MUST 设置 `max-size` 与 `max-file` 轮转选项。微服务与 Redis 默认 MUST 为 `max-size=10m`、`max-file=3`；RabbitMQ MUST 为 `max-size=20m`、`max-file=3`。策略 MUST 在 prod 与 test 对齐（同一 compose 源或等价 anchor）。

#### Scenario: 微服务容器日志有上限

- **WHEN** 运维在 ECS 上对测试或生产微服务栈执行 `docker compose up -d --force-recreate`
- **THEN** 各微服务容器 inspect 的 LogConfig MUST 含 `max-size` 10m 与 `max-file` 3

#### Scenario: RabbitMQ 容器日志有上限

- **WHEN** 运维 recreate 生产或测试 RabbitMQ compose 栈
- **THEN** Rabbit 容器 LogConfig MUST 含 `max-size` 20m 与 `max-file` 3

### Requirement: RabbitMQ MUST 降低 stdout 日志级别至 warning

生产与测试 RabbitMQ compose MUST 挂载仓库内 `manifest/docker/rabbitmq/rabbitmq.conf`，且该文件 MUST 将 `log.console.level`、`log.connection.level`、`log.channel.level` 设为 `warning`。Management 插件与 AMQP/HTTP 业务行为 MUST NOT 因该配置而关闭。

#### Scenario: Rabbit 正常收发时 stdout 更安静

- **WHEN** 客户端 HTTP Publish 与 AMQP consume 正常运行
- **THEN** Rabbit 容器 `docker logs` MUST NOT 以 info 级别刷屏连接/channel 常规行（warning 及以上仍可见）

### Requirement: runbook MUST 说明日志策略生效与磁盘清理

`docs/runbooks/release-deploy-and-run.md` MUST 说明：logging 变更需 recreate 容器；已有巨型 `*-json.log` 的清理方式（删容器或 truncate）；可用 `docker system df -v` 验收。

#### Scenario: 运维按 runbook 部署后验收

- **WHEN** 运维完成 logging 变更的 recreate
- **THEN** runbook MUST 提供检查 LogConfig 与磁盘占用的命令示例

---

### Requirement: 新增微服务 MUST 同步 CI 与发版清单

*`），OpenSpec proposal/design/tasks 与实现 MUST 同步更新下列发版清单，不得仅完成基线 Compose / 本地 Dockerfile：

1. 全局约定 MUST 写入 `openspec/project.

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

---

### Requirement: 生产与测试主机职责分离

切流完成后，系统部署 MUST 满足：生产 Docker 栈与生产 MySQL 仅运行在新 4C8G ECS；测试 Docker 栈与 `ai_voice_*_test` 库仅留在旧 ECS。旧 ECS MUST NOT 再运行生产微服务 compose（可短期保留生产库只读回滚）。

#### Scenario: 切流后拓扑验收

- **WHEN** 运维完成 DNS/反代切流并执行验收
- **THEN** 生产域名请求解析到的服务 MUST 连接新机 MySQL 生产库（库名无 `_test` 后缀），且测试域名请求 MUST 仍连接旧机 `_test` 库

#### Scenario: 旧机停止生产栈

- **WHEN** 生产流量已切至新机
- **THEN** 旧机生产微服务 compose MUST 处于 down，以避免容器继续写入旧机 `ai_voice_*`

### Requirement: 生产库迁移范围与改址方式

生产切流 MUST 仅迁移 MySQL 生产库 `ai_voice_*`（含 history/device/voice/app/ucg/sim 等域约定库）。测试库 `ai_voice_*_test` MUST NOT 迁往新机。应用进程改连新库 MUST 通过更新 `.env.prod` 的 `MYSQL_TCP_HOST` 完成；`*_DB_LINK` MUST 继续使用 `mysql-host` 占位符与生产库名。

#### Scenario: 仅改 MYSQL_TCP_HOST

- **WHEN** 新机生产服务启动且 `MYSQL_TCP_HOST` 已设为新机可达地址
- **THEN** 启动日志 MUST 显示各 `database.*` 主机为该地址，且库名为对应 `ai_voice_*`（非 `_test`）

#### Scenario: 测试 env 不跟迁

- **WHEN** 生产已迁至新机
- **THEN** 旧机 `.env.test` 的 `MYSQL_TCP_HOST` MUST 仍指向旧机 MySQL，且库名 MUST 保持 `_test` 后缀

### Requirement: 新机 MySQL buffer 首日规格

新机生产 MySQL 在首日上线时 MUST 将 `innodb_buffer_pool_size` 配置为 **1G**。旧机（测试）MySQL MUST NOT 跟随调整到 1G，MUST 维持适合约 2G 主机的较小 buffer（例如约 256M 量级）。

#### Scenario: 新机 buffer 验收

- **WHEN** 新机 MySQL 完成配置并重启后执行 `SHOW VARIABLES LIKE 'innodb_buffer_pool_size'`
- **THEN** 返回值 MUST 为 1G（1073741824 字节）或运维文档等效写法对应的 1G

#### Scenario: 旧机 buffer 不跟调

- **WHEN** 生产切流完成
- **THEN** 旧机 MySQL `innodb_buffer_pool_size` MUST 仍为切流前测试适用规格，不得因新机升配而被改为 1G

### Requirement: 生产 Redis 空集群冷启

新机生产 Redis Cluster MUST 以空 volume 冷启：首次 `up` 后执行一次三主 `cluster create`，达到 `cluster_state:ok` 后再启动依赖 Redis 的生产微服务。MUST NOT 将旧机生产 Redis AOF/volume 拷贝为新机初始数据。切流 runbook MUST 声明冷启的用户可感知后果：App refresh token 失效需重新登录；语音/诊所会话上下文丢失；AI 月额度 Redis 计数归零；Feed 索引/快照需经读路径或写路径重建；UCG 私信正文以 MySQL 为权威不因冷启丢失。

#### Scenario: 新集群首次创建

- **WHEN** 新机首次初始化生产 Redis Cluster
- **THEN** 运维 MUST 在空节点上执行一次 `cluster create`，且 `CLUSTER INFO` MUST 显示 `cluster_state:ok`

#### Scenario: 冷启后私信仍可读

- **WHEN** Redis 为空且用户拉取已持久化到 MySQL 的会话历史
- **THEN** 系统 MUST 仍能返回私信正文（MySQL fallback / warm）

#### Scenario: 冷启后需重新登录

- **WHEN** 用户持有切流前签发的 refresh token 访问新机生产环境
- **THEN** 该 refresh token MUST 不再有效，用户 MUST 通过登录流程获取新凭证

### Requirement: 换机 runbook 与档位文档

仓库 MUST 提供可执行的生产换机切流说明（含准备、dump/restore、Redis 冷启、停写、DNS 切流、验收、回滚、旧库保留后再清理），并 MUST 在内存档位文档中记录 **4C8G / 单 prod + 本机 MySQL** 的建议，其中 MySQL buffer 起点为 1G。

#### Scenario: runbook 可独立执行

- **WHEN** 运维按仓库 runbook 执行生产换机
- **THEN** 文档 MUST 覆盖新机 MySQL buffer=1G、仅迁生产库、`MYSQL_TCP_HOST` 改址、Redis 空集群冷启、旧机 test 留守与验收命令

#### Scenario: 档位文档含 4C8G

- **WHEN** 查阅 `memory-sizing-guide`（或等价档位文档）
- **THEN** 文档 MUST 包含 4C8G prod-only 建议，并写明新机 `innodb_buffer_pool_size` 首日 1G

---

### Requirement: Apply 收尾 MUST 列出重部署进程

每次 OpenSpec 实现收尾（/opsx:apply 或等价交付）时，助手 MUST 在回复中列出需要重新部署的 Compose/cmd 进程（如 device-service、gateway-app-server），或显式写明不必重部署及原因；若依赖新 env/compose 注入 MUST 一并说明。 The system SHALL treat redeploy clarity as part of deploy hygiene.

#### Scenario: 领域服务二进制变更

- **WHEN** 变更仅修改 device-service 构建产物且 gateway 路径未变
- **THEN** 回复 MUST 列出 device-service，并可写不必重部署 gateway-app-server

#### Scenario: 仅文档变更

- **WHEN** 变更仅修改 openspec/changes/** 文档
- **THEN** 回复 MUST 写明无需部署业务进程

