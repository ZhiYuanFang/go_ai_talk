## Context

`cash-service`（`:9807`）已在基线 `docker-compose.microservices.yml` 与 `Dockerfile.cash-service` 落地；gateway-app / voice 已依赖 `CASH_SERVICE_URL`。发版路径仍为「九微服务」矩阵：`docker-acr.yml` 的 `ALL_SERVICES` 与 test/prod overlay 均未登记 cash。`openspec/project.md` 已有「新进程须同步 `.env.example` / 基线 compose / runbook」约定，但未显式要求 GitHub Actions workflow 与 ACR overlay。

## Goals / Non-Goals

**Goals:**

- CI 全量 tag / `workflow_dispatch` 留空 services 时 build/push 含 `cash-service`；支持别名 `cash`、`cash-service` 做部分构建。
- test/prod overlay 可按 `${REGISTRY}/cash-service:${IMAGE_TAG}` pull/up。
- runbook 与服务数/别名文案一致（十服务）。
- `openspec/project.md` 固化「新增微服务 → 同步 workflows + overlay + runbook + ACR 建仓」强制约定与评审检查项。

**Non-Goals:**

- 不改 cash 业务逻辑、API、DDL、支付配置。
- 不改 Flutter；不新建测试文件。
- 不在本变更内代建 ACR 控制台仓库（仅文档/tasks 提醒）。
- 不调整其他服务端口或镜像命名惯例。

## Decisions

1. **镜像仓库名单段 `cash-service`**  
   与现有 `voice-service`、`notify-service` 一致；CI matrix `id` = ACR 仓库名 = overlay `image` 路径末段。  
   备选：`cash`（短名）→ 与进程/compose 服务名不一致，放弃。

2. **端口映射**  
   - 容器内保持 `:9807`（基线已定，避开 notify `:9806`）。  
   - test overlay：`19807:9807`（沿用 198xx 域服务端口段）。  
   - prod overlay：`9807:9807`。  
   mcp 无 HTTP ports；cash 有 `/api.json` 健康检查，需映射以便宿主机探测（与 notify 同理）。

3. **全局约束落点：`openspec/project.md`**  
   新增小节「微服务 CI / ACR 发版约定（强制）」，并在「重要约束」增加评审检查项。不强制改仓库根 `AGENTS.md` 托管大段（可由 project.md 引用）；若根 `AGENTS.md` 已有部署检查列表则可加一行交叉引用，避免双源冗长。  
   备选：仅改 runbook → AI propose/apply 不一定读 runbook，约束力不足。

4. **文档数字**  
   workflow 头注释与 runbook 中「全量 N 服务」统一为 **10**（原 9 + cash）；历史段落中「七服务」若仍指旧基线，本变更仅修正与 docker-acr 直接相关的表述，避免无关大段改写。

5. **部分构建示例**  
   在 workflow 头注释增加 `+cash` 示例，与现有 `+ucg` / `+sim` / `+notify` 并列。

## Risks / Trade-offs

- **[Risk] ACR 未建仓导致 push denied** → Mitigation：tasks/runbook 明确先建 `cash-service`；首次可用 `+cash` 验证单仓。
- **[Risk] 全量 pull 时缺 cash 镜像失败** → Mitigation：合并后首次全量 tag 或 dispatch 全量；或先 `services=cash` 再全量。
- **[Risk] 服务数文案残留「9」** → Mitigation：tasks 含 grep 自检 `docker-acr.yml` / runbook 别名段。
- **Trade-off**：全局约束偏文档与评审，无法在 CI 自动断言「新服务已登记」；后续若需要可再加脚本检查（本期不做）。

## Migration Plan

1. ACR test/prod 命名空间创建仓库 `cash-service`。
2. 合并本变更后，以预发布 tag 全量或 `+cash` 推送镜像。
3. 更新服务器 `.env.test` / `.env.prod`（`CASH_DB_LINK` 等已在 example；确认已填）。
4. `compose pull` + `up -d --no-build`（含 cash）；按 runbook 顺序：cash → gateway-app → voice。
5. 回滚：`.env` 的 `IMAGE_TAG` 回退；若仅 cash 新仓无旧 tag，可临时从栈中 scale/stop cash（业务 VIP 不可用，与现状一致）。

## Open Questions

- ACR 两命名空间是否已有人建好 `cash-service` 仓（实现前运维确认即可，不阻塞代码合并）。
