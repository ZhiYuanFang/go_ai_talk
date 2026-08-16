## Context

`voice-service` 通过 `OpenClawHTTPClient`（`OpenClawFromCfg`）调用 OpenClaw 门禁（`/v1/chat/completions`）。配置键为 `openclaw.gatewayUrl` / `gatewayToken` / `apiToken`，并约定环境变量 `OPENCLAW_GATEWAY_URL` / `OPENCLAW_GATEWAY_TOKEN` / `PANGBAO_API_TOKEN`。

当前 `config.voice-service.yaml` 使用 `"${OPENCLAW_GATEWAY_URL}"` 等形式。GoFrame 配置加载**不会**做 shell 式变量展开；`OpenClawFromCfg` 仅在配置值为空串时读取 `os.Getenv`。结果客户端 baseURL 变成字面量 `${OPENCLAW_GATEWAY_URL}`，HTTP 报 `unsupported protocol scheme ""`。compose 已把真实 URL/Token 注入容器 env，但被非空假配置挡住。

方案 A（本变更）：YAML 留空，沿用既有「空配置 → env → URL 本地默认」逻辑，与仓库其它密钥占位方式一致。

## Goals / Non-Goals

**Goals:**

- 默认配置下，容器内已设置的 `OPENCLAW_GATEWAY_URL` 等 MUST 被 `OpenClawFromCfg` 采用。
- Token 与 URL 同一套空 YAML + env 约定，避免修好 URL 后仍带着字面量 `${TOKEN}` 鉴权失败。
- 注释明确禁止在 voice-service YAML 中写未展开的 `${VAR}`。

**Non-Goals:**

- 不引入 `os.ExpandEnv` 或占位符检测逻辑（方案 B/C）。
- 不改 OpenClaw HTTP 协议、请求头语义、Intent/Clinic/Care 业务路径。
- 不改 docker-compose 环境注入语法（已正确）。
- 不新增测试文件。

## Decisions

### 1. YAML 空字符串，不改 Go 回落逻辑

- **选择**：`gatewayUrl` / `gatewayToken` / `apiToken` 写为 `""`；保留 `OpenClawFromCfg` 现有优先级。
- **理由**：最小改动；与 `voice.admin.password: ""` 模式一致；compose 已负责注入。
- **备选**：`os.ExpandEnv`（方案 B）或识别 `\$\{.+\}`（方案 C）——能保留 `${VAR}` 书写习惯，但增加解析面与误用风险；本次明确选 A。

### 2. URL 本地默认保留

- 配置与 env 皆空时仍默认 `http://127.0.0.1:8000/agent-gate`（本地开发友好）。
- compose 生产默认已是 `http://python-ai-talk:8000/agent-gate`，经 env 注入，不依赖 YAML。

### 3. 运维真源仍是 env 文件

- `.env.prod` / compose `--env-file` 继续提供真实值；`.env.example` 若缺 OpenClaw 键则补注释行，避免后人再在 YAML 写 `${VAR}`。

## Risks / Trade-offs

- **[Risk] 未 recreate 的旧容器仍带旧镜像内嵌 YAML** → 发版后需重建 `voice-service`；配置挂载场景则重启即可。
- **[Risk] Token 未设时鉴权失败** → 与修前「字面量 Token」同样依赖 env；文档与 `.env.example` 标明必填。
- **[Trade-off] YAML 无法「看见」默认 URL** → 靠注释 + compose 默认值说明；比假占位符更安全。

## Migration Plan

1. 合并配置修改并发布含新 `config.voice-service.yaml` 的 `voice-service` 镜像。
2. 确认 `--env-file` 含 `OPENCLAW_GATEWAY_URL` / `OPENCLAW_GATEWAY_TOKEN` / `PANGBAO_API_TOKEN`（prod 已有）。
3. `up -d --force-recreate voice-service`（或等价）。
4. 冒烟：智能对话不再出现 `${OPENCLAW...}` URL；若 401 则查 Token env。
5. **Rollback**：回滚镜像即可；旧镜像仍有假占位符 bug，不宜长期回退。

## Open Questions

- （无）方案 A 已在 explore 中选定；实现仅触及配置与注释。
