## 1. GitHub 仓库配置（合并前/后人工）

- [ ] 1.1 在 GitHub 仓库 Settings → Environments 创建 `test` 与 `prod` 两个 Environment
- [ ] 1.2 在 Repository secrets 配置 `ACR_USERNAME`、`ACR_PASSWORD`（test/prod 共用同一 ACR 实例凭证）
- [ ] 1.3 在 Environment `test` 配置 secret `REGISTRY`（从本地 `.env.test` 复制，可含 `-vpc` 域名）
- [ ] 1.4 在 Environment `prod` 配置 secret `REGISTRY`（从本地 `.env.prod` 复制）

## 2. Workflow 改造

- [x] 2.1 在 `build-push` job 增加 `environment: ${{ needs.resolve-env.outputs.deploy_env }}`（`resolve-env` job 先解析环境）
- [x] 2.2 重写「选择环境并加载 .env」步骤：移除对 `manifest/docker/.env.*` 文件存在性检查与 `read_env_var` 读文件逻辑
- [x] 2.3 从 `${{ secrets.REGISTRY }}`、`${{ secrets.ACR_USERNAME }}`、`${{ secrets.ACR_PASSWORD }}` 加载 ACR 配置；缺失时输出 `::error` 并指向 runbook
- [x] 2.4 保留 `-vpc` → 公网 push 推导与 `validate_push_registry` 校验；对 `ACR_PASSWORD` 保留 `::add-mask::`
- [x] 2.5 更新 workflow 文件头注释：说明凭证来自 GitHub Secrets/Environments，删除「无需 GitHub Secrets」表述

## 3. 文档与模板

- [x] 3.1 更新 `docs/runbooks/release-deploy-and-run.md`「ACR 与 CI 凭证」：拆分 **CI Secrets** 与 **ECS 本地 .env** 两节，列出 secret 名称与配置步骤
- [x] 3.2 更新 `manifest/docker/.env.example` 中 `REGISTRY`/`ACR_*` 注释：标明仅 ECS 使用，CI 见 runbook
- [x] 3.3 确认 `.gitignore` 仍忽略 `**/.env.*` 且保留 `!**/.env.example`（无需恢复 `.env.test|prod` 进 git）

## 4. 验收

- [x] 4.1 本地/空 checkout 场景：确认 workflow 逻辑不依赖磁盘上的 `.env.test|prod` 文件
- [ ] 4.2 使用 `workflow_dispatch`（target_env=test + 测试 image_tag）触发一次构建，确认 ACR login 与 push 成功
- [ ] 4.3 核对 Actions 日志：无 `ACR_PASSWORD` 明文；push 地址无 `-vpc` 段
- [x] 4.4 确认 runbook 中 ECS 部署步骤（`docker compose --env-file .env.test` 等）未因本文档改动而失效
