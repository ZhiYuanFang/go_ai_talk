## ADDED Requirements

### Requirement: docker-acr SHALL support tag build-scope suffix for selective service builds

`docker-acr` workflow SHALL 支持 git tag **`+` 构建范围后缀**：`+` **之前** 的片段为 **base tag**（用作 ACR 主 tag、`primary_tag`、与服务器 `IMAGE_TAG` 一致）；`+` **之后** 为逗号分隔的服务别名列表，workflow **仅** build 并 push 列表对应的服务镜像。无 `+` 后缀时 SHALL 构建全部 6 个微服务（与变更前行为一致）。

workflow MUST NOT 对未纳入构建范围的服务执行 retag 或 manifest 复制。未构建的服务在 ACR 上 **MAY** 不存在 `:${base_tag}` 镜像；该状态 SHALL 视为预期，而非 workflow 失败条件。

服务别名 MUST 至少支持：`gateway`、`gateway-app`、`history`/`history-service`、`voice`/`voice-service`、`device`/`device-service`、`ucg`/`ucg-service`；`all` MUST 表示全量 6 服务。非法别名 MUST 导致 workflow 在 build 前失败并输出可操作的错误信息。

test/prod 环境路由 MUST 基于 **base tag**（去掉 `+` 后缀后）应用现有规则，MUST NOT 因 `+ucg` 等后缀改变命名空间选择。

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

## MODIFIED Requirements

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
