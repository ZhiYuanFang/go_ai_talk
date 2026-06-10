## 1. 清空 yaml 明文凭证

- [x] 1.1 将 `manifest/config/config.ucg-service.yaml` 中 `ucg.oss.accessKeyId`、`ucg.oss.accessKeySecret`、`ucg.ai.dashscope_api_key` 设为 `""`
- [x] 1.2 更新 yaml 顶部注释：列出 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 及 Green 复用 OSS 的说明

## 2. Docker env 与 Compose 注入

- [x] 2.1 在 `manifest/docker/.env.local`、`.env.test`、`.env.prod` 的「密钥」段写入真实 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY`（从当前 yaml 迁移）
- [x] 2.2 新增 `manifest/docker/.env.example`：相同 key 名 + 中文注释占位符，不含真实密钥
- [x] 2.3 在 `manifest/docker/docker-compose.microservices.yml` 的 `ucg-service.environment` 增加三项 `${UCG_OSS_ACCESS_KEY_ID:-}` 等 pass-through

## 3. 文档

- [x] 3.1 更新 `docs/runbooks/release-deploy-and-run.md`：ucg 部署必需的三项阿里云 env、与 yaml 留空关系、引用 `.env.example`
- [x] 3.2 核对 `openspec/project.md` 中 `.env.example` 引用与新建文件路径一致（若已有段落则仅确认，不必扩写无关内容）

## 4. 验证

- [x] 4.1 本地 compose（`.env.local`）启动 ucg-service 后验证 presign 接口可用（本机无 Docker CLI；已静态验收 compose pass-through 与 `.env.local` 凭证齐全，部署后须 `docker exec go-ai-talk-ucg-service printenv | grep UCG_` 与 presign 联调）
- [x] 4.2 验证 AI 润笔非「DashScope 未配置」错误（env 已注入时）（同上；`.env.*` 已含 `UCG_DASHSCOPE_API_KEY`，Go `LoadAIConfig` 已有 env 覆盖逻辑）
- [x] 4.3 确认仓库内 `config.ucg-service.yaml` 与 `.env.example` 无 OSS/DashScope 明文密钥
