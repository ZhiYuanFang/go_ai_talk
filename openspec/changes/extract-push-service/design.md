## Context

推送发送栈（APNs/HMS/MiPush、`ucg_push_device`、注册 API、`PushVisibleAlert`/`PushByBizType`）位于 `ucg-service`。预测临近等全 App 场景已经 internal 调用推送，但路径与配置仍姓 UCG。`notify-service` 仅 App Banner，与厂商推送无关。gateway-app 负责鉴权/反代/usage。产品未正式推广，允许 **BREAKING** 删除旧 API、一步迁库。

## Goals / Non-Goals

**Goals:**

- 独立 `push-service` 作为唯一厂商推送宿主；中性 App/Internal API；`clients/push` 供各域调用。
- 库 `ai_voice_push` 一步到位；部署纳入 ACR 与 microservices compose。
- badge 由调用方传入；push 不计算 UCG 未读。
- push 注册/注销 **不计入** App API usage（负责人已确认）。

**Non-Goals:**

- 兼容旧 `/ucg/.../push`；双写/过渡读旧表。
- 把推送实现放入 gateway-app 或合并进 notify-service。
- 推送失败 MQ 重试；改预测闹钟模型（仅改最后一跳 client）。

## Decisions

### D1. 独立进程 `push-service`，端口 `:9808`

- 不占用 notify `:9806`。包路径：`cmd/push-service`、`internal/services/push`、`internal/controller/push`、`config.push-service.yaml`。
- 备选（否）：挂 voice / 挂 gateway / 扩 notify。

### D2. Schema `ai_voice_push`，表建议 `push_device`

- 同 MySQL 实例新建库；列语义对齐现 `ucg_push_device`（wx_id、channel、token、device_key、时间戳）；唯一键 `(wx_id, device_key, channel)`。
- **无**从旧表双写迁移要求；部署时建新库即可（未推广可丢弃旧 token，客户端重新 register）。
- 备选（否）：继续写 ucg 库表。

### D3. API 与路径（无兼容）

| 用途 | Path（建议） |
|------|----------------|
| App 注册 | `POST /app/api/push/register` |
| App 注销 | `POST /app/api/push/unregister` |
| Internal 发送 | `POST /push/internal/api/by-biz-type`（或 `/app/internal/api/push/by-biz-type`，实现与 gateway 反代一致即可） |

- 鉴权：App 经 gateway Bearer → 注入 `X-Internal-Wx-Id`；Internal 用 `DEVICE_GATEWAY_INTERNAL_SECRET`。
- **删除**全部 UCG 推送 App/Internal 路由与 `services/ucg` 发送实现。

### D4. badge 语义

- Internal 请求含可选 `badge`（int）与 `alert`、`bizType`、`data`、`silent`。
- UCG 社区推送：调用前 `ComputeTotalUnread`，传入 `badge`。
- 预测临近：`badge=0`（或不传则默认 0）。
- push-service **MUST NOT** 查询 ucg 库算未读。

### D5. usage（负责人确认）

- `POST /app/api/push/register`、`POST /app/api/push/unregister`：**不统计**。
- 从 `maintenance_skip.go` 移除旧 ucg push 项（路由已删），加入上述新精确 apiKey。
- Internal 发送不在 App usage 范围内。

### D6. 调用方迁移

- 新增 `internal/clients/push`；voice 预测、`clients/ucg` 内推送封装删除，改调 push。
- UCG 进程内原 `PushVisibleAlert` / `PushSilentBadge` / `PushByBizType` 改为经 client 或同进程禁止再直连厂商（实现可留 thin wrapper 调 client，避免循环 import：ucg → clients/push）。

### D7. 配置与部署

- Env：`PUSH_APNS_*`、`PUSH_HMS_*`、`PUSH_MIPUSH_*`、`PUSH_SERVICE_URL`、`PUSH_SERVICE_ADDR`、`PUSH_DB_LINK`（或组名约定）。
- 必改：`Dockerfile.push-service`、`docker-compose.microservices.yml`（及 prod/local overlay、`.env.example`）、`.github/workflows/docker-acr.yml`（ALL_SERVICES + Dockerfile 映射 + 别名）。
- gateway：反代 `/app/api/push/*`；voice/ucg 容器注入 `PUSH_SERVICE_URL`。
- APNs `.p8` 挂载目标从 ucg 容器改为 push 容器。

### D8. 与 predict-imminent 关系

- 闹钟/Redis/延时 MQ 仍在 voice；仅将 `clients/ucg.PushByBizType` 替换为 `clients/push`；更新该 change 的 CLIENT/design 表述（可同 PR）。

## Risks / Trade-offs

- [旧客户端仍打 ucg push path] → 未推广可接受；发版说明改 path。
- [token 清空需重注册] → 预期；App 启动注册即可。
- [漏改 ACR/compose 导致无镜像] → tasks 强制勾选 workflow + compose。
- [ucg→clients/push 超时] → 合理 timeout；失败语义与现 best-effort 推送一致。

## Migration Plan

1. 建库 `ai_voice_push`；上线 push-service；gateway 反代新 path；usage skip 更新。
2. 切换 voice/ucg 调用；删 ucg 推送代码与 `UCG_APNS_*` 注入。
3. 客户端改注册 path；无需旧 API 灰度。
4. 回滚：仅能回退整包镜像（无双路径）；因未推广可接受。

## Open Questions

- （无阻塞）Internal path 最终字符串在实现时与现有 `/ucg/internal/api/*` 风格对齐即可。
- predict 同步 API 的 usage 是否计入：属另一 change；本变更不处理。
