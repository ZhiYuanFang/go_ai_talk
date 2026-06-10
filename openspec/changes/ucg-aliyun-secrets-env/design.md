## Context

- **ucg-service** 使用阿里云 OSS（presign/upload/delete）、Green 内容审核（AK 复用 OSS）、DashScope（AI 润笔 vision）。
- Go 加载链已存在：
  - `LoadOSSConfig`：`UCG_OSS_ACCESS_KEY_ID` / `UCG_OSS_ACCESS_KEY_SECRET` 覆盖 yaml
  - `LoadGreenConfig`：green yaml 为空时 fallback 到 OSS 凭证
  - `loadAIConfigFresh`：`UCG_DASHSCOPE_API_KEY` 覆盖 yaml
- **现状问题**：yaml 与 v2.0.2 聚合规格仍含明文 AK/SK 与 DashScope key；`docker-compose.microservices.yml` 未向 ucg 容器传递上述 env；`.env.example` 在 project.md 中被引用但文件不存在。
- **约束**：本变更仅在主仓库完成 env 化；`.env.local|test|prod` 继续 tracked 写真实值（与 DB/JWT 一致）；不新增 `*_test.go`；不拆私有 secrets 仓库。

## Goals / Non-Goals

**Goals:**

- yaml 中 OSS AK/SK 与 DashScope key 留空，非敏感 OSS/Green/AI 参数保留。
- `.env.local`、`.env.test`、`.env.prod` 写入真实 `UCG_OSS_*` 与 `UCG_DASHSCOPE_API_KEY`。
- Compose 基线 `ucg-service` 注入三个 env 变量（`${VAR:-}` pass-through）。
- 提供 `.env.example` 占位模板与 runbook 说明。
- 增量规格明确：凭证 MUST NOT 以明文写入提交到 git 的 yaml；Green 无需独立 AK env。

**Non-Goals:**

- 修改 Go 加载逻辑或新增 fail-fast 启动校验。
- 轮换已在 git/spec 中出现的密钥（运维另行处理）。
- 拆分 `manifest/docker` 到独立私有仓库或迁移 CI 至 GitHub Secrets。
- 改动 `docker-acr.yml` 的 ACR 凭证读取方式。
- voice/device 等其他进程的第三方密钥治理。

## Decisions

### 1. 凭证来源：env 优先 + yaml 留空（对齐 gateway-app 口令模式）

- **决定**：仓库内 `config.ucg-service.yaml` 的 `accessKeyId`、`accessKeySecret`、`dashscope_api_key` 设为 `""`；运行时以 `.env.*` → Compose → 容器 env 为准。
- **理由**：与 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`、`*_DB_LINK` 一致；Go 代码已支持，改动面最小。
- **备选**：env-only 且启动失败 — 过度严格，本地忘配 env 时难排查；yaml fallback 留空即可使 OSS presign 返回明确「未配置」错误。

### 2. Green 不新增独立 env

- **决定**：仅配置 `UCG_OSS_ACCESS_KEY_*`；`LoadGreenConfig` 继续从 OSS 配置继承。
- **理由**：同一阿里云 RAM 账号；减少 env 面与文档重复。
- **备选**：`UCG_GREEN_ACCESS_KEY_*` — 仅当未来 OSS/Green 分账号时需要。

### 3. `.env.*` 继续进主仓库（tracked）

- **决定**：真实值写入 `.env.local|test|prod`（用户已确认）；另建 `.env.example` 仅含占位符与注释。
- **理由**：与现有 DB/ACR 密钥管理方式一致；CI 仍读同目录 `.env.test|prod`。
- **备选**：gitignore `.env.*` — 需同步改 CI，留待后续 secrets 仓库变更。

### 4. Compose 仅在基线 `docker-compose.microservices.yml` 增加 pass-through

- **决定**：在 `ucg-service.environment` 增加：
  - `UCG_OSS_ACCESS_KEY_ID: "${UCG_OSS_ACCESS_KEY_ID:-}"`
  - `UCG_OSS_ACCESS_KEY_SECRET: "${UCG_OSS_ACCESS_KEY_SECRET:-}"`
  - `UCG_DASHSCOPE_API_KEY: "${UCG_DASHSCOPE_API_KEY:-}"`
- **理由**：local/test/prod overlay 均继承基线；一处修改全覆盖。
- **备选**：仅 prod overlay — local 开发会漏配。

### 5. 规格增量：新 capability + 修改 ucg-oss-presign

- **决定**：新增 `ucg-aliyun-secrets-env` 覆盖 Compose/env/example/runbook；MODIFIED `ucg-oss-presign` 凭证段落，移除明文表格。
- **理由**：部署约定与 OSS 行为分属不同关注点，但 OSS 凭证要求必须在原 capability 上 delta。

## Risks / Trade-offs

- **[Risk] git 历史中仍有旧明文** → 运维侧建议轮换 AK/SK 与 DashScope key；本变更不自动轮换。
- **[Risk] 漏配 env 导致 OSS/润笔不可用** → `oss_presign` 已有空凭证校验；润笔返回未配置；runbook 列出必需项。
- **[Risk] `.env.example` 与 `.env.*` 字段漂移** → tasks 验收项要求三者 key 名一致。
- **[Trade-off] 真实密钥仍在主仓库 `.env.*`** → 接受（用户选择）；后续可迁私有 secrets 库而不改 env 名。

## Migration Plan

1. 合并前：在 `.env.local|test|prod` 填入当前 yaml 中的真实凭证（可从 yaml 复制）。
2. 合并：清空 yaml 明文 → 更新 compose → 更新 runbook / `.env.example`。
3. 部署 test/prod：pull 后确认 `.env.*` 含新三项 → `docker compose ... up -d` 重建 ucg-service。
4. 验证：`POST /ucg/app/api/media/presign` 成功；Green enabled 时审核可调；润笔接口非「未配置」。
5. 回滚：恢复 yaml 明文或临时在服务器 export env；Compose pass-through 可保留。

## Open Questions

- （无阻塞项）密钥轮换与独立 secrets 仓库列为后续运维/架构议题，不在本变更范围。
