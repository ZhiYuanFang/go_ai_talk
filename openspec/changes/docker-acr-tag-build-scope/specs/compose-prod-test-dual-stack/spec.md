## ADDED Requirements

### Requirement: 部分 CI 构建发版 SHALL 支持按服务 pull 与 up

当 CI 使用带 `+` 构建后缀的 tag（如 `v1.0.0-rc.4+ucg`）仅 push 部分服务镜像时，运维 MUST NOT 假设 ACR 上存在全部六服务 `:${base_tag}` 镜像。部署 MUST 通过 **按服务** `docker compose pull <service>` 与 `up -d --no-build <service>` 更新变更服务；对未构建服务 **MAY** 继续运行已拉取的旧 tag 本地镜像，直至下次无后缀全量 tag 发版。

`docs/runbooks/release-deploy-and-run.md` MUST 文档化：git tag 全名与 `.env` 中 `IMAGE_TAG`（base tag）的关系、`+ucg` 等后缀含义、按服务 pull/up 示例命令，以及全栈 `compose pull` 在部分构建 tag 下 **预期失败** 的说明（表示构建范围与部署操作不匹配）。

#### Scenario: 部分构建后仅更新 ucg

- **WHEN** CI 已对 tag `v1.0.0-rc.4+ucg` 仅 push `ucg-service:v1.0.0-rc.4`，且运维将 `.env.test` 中 `IMAGE_TAG` 设为 `v1.0.0-rc.4`
- **THEN** 运维 SHALL 执行 `pull`/`up` 仅针对 `ucg-service`；其他服务容器 MAY 保持上一版本镜像运行

#### Scenario: 全栈 pull 在部分 tag 下失败为预期

- **WHEN** ACR 不存在 `gateway:v1.0.0-rc.4`（因本次 CI 为 `+ucg` 部分构建）且运维执行全栈 `docker compose pull`
- **THEN** pull MUST 对缺失镜像报错；运维 SHALL 改为按服务 pull 或先打无后缀全量 tag 构建
