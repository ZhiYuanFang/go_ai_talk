## Context

`ResolveHTTPTargets()`（`internal/services/contracts/http_targets.go`）在未设置环境变量时使用 `http://127.0.0.1:9801/9802/9803` 作为 history / voice / device 的 HTTP 基址。本机多进程联调时合理；Docker 每服务一容器时，history 容器内访问 `127.0.0.1:9803` 不会路由到 `device-service` 容器。仓库内 `docker-compose.microservices.yml` 已为 `voice-service` 配置 `HISTORY_SERVICE_URL`、`DEVICE_SERVICE_URL`，但 `history-service` 段落缺少对 voice、device 的反向配置。

## Goals / Non-Goals

**Goals:**

- 参考部署（至少 `docker-compose.microservices.yml`）中，history 与 voice、device 处于同一用户定义网络时，history 的委派 HTTP 客户端 MUST 使用可解析的服务名（如 `voice-service`、`device-service`）而非 `127.0.0.1`。
- 文档与 history 专属配置注释中明确上述约束，降低「同网段仍用 localhost」的误配。

**Non-Goals:**

- 不修改 `http_targets.go` 的默认值（仍保留本机 go run 体验）。
- 不在本变更中强制增加运行时探测（可作为后续加固任务）。

## Decisions

1. **Compose 显式注入环境变量**  
   - **决策**：在 `history-service.environment` 增加 `VOICE_SERVICE_URL=http://voice-service:9802`、`DEVICE_SERVICE_URL=http://device-service:9803`。  
   - **理由**：与 voice 段落的写法一致，不依赖「碰巧」与宿主机端口映射一致；服务名随 compose project 解析。  
   - **备选**：改代码在检测到 `/.dockerenv` 时自动改写默认 URL——侵入性强、与 K8s 等非 Docker 环境重叠差，不采纳。

2. **Kustomize**  
   - **决策**：若仓库已有 history Deployment 清单，则补充同等环境变量（值可用 ConfigMap/占位，或文档化由 overlay 覆盖）；若无独立 history deployment，则仅在 runbook 说明 K8s 必配项。  
   - **理由**：保持部署面与 compose 语义对齐，避免仅 compose 正确、K8s 仍踩坑。

3. **文档**  
   - **决策**：在 `release-deploy-and-run.md` 与 `config.history-service.yaml` 顶部注释中增加「容器内 MUST 设置 VOICE_SERVICE_URL / DEVICE_SERVICE_URL」一句，并指向 compose 示例。

## Risks / Trade-offs

- **[Risk] 本地开发者复制 compose 片段但自定义了 service 名** → 需同步改 URL；缓解：使用 compose 默认 `service:` 键名作为文档基准。  
- **[Risk] 多网络或 mesh 下服务名不同** → 由环境变量覆盖；规范不绑定具体主机名，只要求「非本容器 loopback」。

## Migration Plan

1. 更新 compose 后 `docker compose up` 重建 history 容器。  
2. 无数据迁移；回滚为移除环境变量（恢复旧行为，仅不推荐）。

## Open Questions

- Kustomize base 是否已存在 `history-deployment.yaml`：实施阶段glob确认后再改或仅文档。
