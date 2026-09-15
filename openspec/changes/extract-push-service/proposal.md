## Why

系统推送已从「仅 UCG 社区」扩展为全 App 能力（含预测临近提醒），但仍实现在 `ucg-service`、路径挂 `/ucg/app/api/push/*`、配置前缀 `UCG_*`，语义与进程边界都绑死社区域。项目尚未正式推广，适合一次性拆出独立 `push-service`，去掉兼容层与双写。

## What Changes

- **新增**独立进程 `push-service`（**不**复用 `notify-service`）：负责 token 注册/注销、按 `bizType` 下发 APNs/HMS/MiPush；库为同 MySQL 实例独立 schema **`ai_voice_push`**（一步到位，无过渡双写）。
- **BREAKING**：删除 UCG 侧推送实现与路由（含 `/ucg/app/api/push/register|unregister`、internal `push/by-biz-type`）；客户端与调用方改用中性 App/Internal 路径。
- gateway-app 反代全局 App 路径（如 `/app/api/push/*`）；**负责人确认：push 注册/注销不计入 usage**，从旧 ucg skip 项迁到新 path 的 `maintenance_skip`。
- UCG 私信/评论等：改为 `clients/push`，请求中**自带 `badge`**（未读仍由 UCG 计算）；预测临近等：`badge=0` 或不传社区未读。
- 部署：新增 Dockerfile、`docker-compose.microservices.yml`（及关联 overlay/env）、`.github/workflows/docker-acr.yml` 纳入 `push-service`；推送证书 env 从 `UCG_*` 改为 `PUSH_*`。
- 同步修正 `predict-imminent-offline-push` 实现/文档：最后一跳改调 `clients/push`（闹钟模型不变）。

## Capabilities

### New Capabilities

- `push-service-runtime`：进程、库 schema、token 表、厂商发送、配置与部署挂载。
- `push-app-and-internal-api`：中性注册/注销与 internal by-biz-type；gateway 反代；usage 排除。
- `push-caller-migration`：ucg/voice（及既有 clients）迁到 `clients/push`；删除 ucg 推送代码。

### Modified Capabilities

- （无强制改写已归档基线 Requirement 文本；预测临近以同仓 change `predict-imminent-offline-push` 的宿主修正为准，本变更 specs 约束 push 平台行为。）

## Impact

- **进程**：新增 `push-service`；`ucg-service` 去掉推送；`voice-service` / `gateway-app-server` 改调用与反代；`notify-service` 不变。
- **库**：新建 `ai_voice_push`；废弃 `ucg_push_device` 写入路径（可删表或弃用，实现期清空 UCG 推送依赖）。
- **客户端**：推送注册 path 变更（孪生仓须同步）；无旧 path 兼容。
- **非目标**：新建通知 Banner 能力；推送失败重试队列；把实现塞进 gateway-app 进程。
