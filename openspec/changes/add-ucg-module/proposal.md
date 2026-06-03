## Why

胖宝平台需要独立的用户生成内容（UCG）域：动态发布、关注关系、推荐 Feed、一对一私信，且内容须经阿里云 Green 违规检测后方可对外可见。现有 `history-service` / `device-service` 边界不应承载社交数据；须新建 `ucg-service` 微服务、独立库 `ai_voice_ucg`，**所有 App 对外流量（HTTP REST + 聊天 WebSocket）统一经 gateway-app-server（app 网关）同域暴露**，ucg-service 仅作为内部微服务被网关反向代理。

该变更遵循既有微服务治理（单服务单库、禁止跨库读 `wx`）与 `history-service` 脚手架模式，补齐 OSS 直传、Green 异步审核、Redis 聊天持久化与混合推荐算法。

## What Changes

- **新微服务 `ucg-service`**：端口约 `:9804`；专属配置 `config.ucg-service.yaml`；数据库 `ai_voice_ucg`（与同实例其他 `ai_voice_*` 库并列）；**不**单独对 App 暴露公网域名。
- **数据表（DDL）**：`ucg_profile`、`ucg_post`、`ucg_post_media`、`ucg_follow`、`ucg_post_like`、`ucg_post_comment`、`ucg_conversation`、`ucg_conversation_member`；可选 `ucg_post_recommend`。帖子状态：`0=draft, 1=pending_audit, 2=published, 3=rejected`。
- **服务边界**：`wx` 表仍在 device-service；ucg-service **不得**跨库读 device DB；通过 HTTP 内部 API + `DEVICE_GATEWAY_INTERNAL_SECRET` 校验 wx、批量 wx、默认昵称 `{babyName}的家长`。
- **gateway-app**：对外统一入口，与现有 App API **同域**：
  - HTTP 反向代理 `/ucg/app/api/*` → ucg-service（参考 `history_route_proxy` / `device_route_proxy`）
  - WebSocket 升级代理 `/ucg/app/ws/chat` → ucg-service `/ws/chat`（参考 `ws_route_proxy.go` / voice WS 透传模式；客户端 URL 推导参考 history WS `/device/app/ws/history`）
- **OSS**：bucket `pang-bao`，region `cn-beijing`，endpoint `oss-cn-beijing.aliyuncs.com`，objectKey 前缀 `social/`；presign 上传；CDN 展示 `https://resorce.cuplay.top/{objectKey}`；DB 仅存 objectKey；凭证写入仓库配置（见 `design.md`）。
- **Green 审核（Option A — 帖子/资料）**：提交 → `pending_audit`，**仅作者可见**；异步 Green → pass → `published`；fail → `rejected`，作者在「我的动态」见「违规已下架」+ reason。视频同流程（审核通过前仅作者可见）。
- **Green 审核（Option C — 聊天）**：发送 → 仅发送方见 pending；pass → WS 投递收件人；fail → 仅发送方见失败，不投递。
- **Redis**：聊天消息永久保留（+ 可选 MySQL 归档）；未读计数与会话列表缓存。
- **推荐 Feed**：混合算法（新帖权重 + 互动衰减）；关注 Feed 按时间倒序。
- **MVP 范围**：不含拉黑/举报。

## Capabilities

### New Capabilities

- `ucg-service-runtime`：ucg-service 进程、配置、依赖检查、端口与 DB 归属。
- `ucg-data-model`：MySQL DDL、状态枚举、索引与分页约定。
- `ucg-app-http-api`：App 侧 REST（profile、feed、posts CRUD、presign、follow、conversations/messages 列表）。
- `ucg-gateway-proxy`：gateway-app 对 `/ucg/app/api/*` 的 Bearer 注入与 HTTP 反向代理，以及对 `/ucg/app/ws/chat` 的 WebSocket 升级代理。
- `ucg-oss-presign`：social/ 前缀 OSS 直传与 objectKey 契约。
- `ucg-green-audit`：帖子/资料/聊天 Green 异步审核与状态机（Option A + Option C）。
- `ucg-chat-ws`：WebSocket 聊天（ucg-service 内部 `/ws/chat`）、Redis 消息存储、投递与未读。
- `ucg-recommend-feed`：推荐流混合排序与可选 `ucg_post_recommend` 表。
- `ucg-device-internal-api`：device-service 暴露给 ucg 的 wx 校验/批量与 baby_name 内部 HTTP。

### Modified Capabilities

- `gateway-app-server`：注册 UCG HTTP 代理、UCG 聊天 WS 升级代理路由与鉴权白名单扩展（delta）。
- `service-boundary-no-cross-db`：明确 ucg 域不得直连 device 库，仅 HTTP 契约（delta，若需重申 ucg 边界）。

## Impact

- **Affected code**：`cmd/ucg-service/main.go`（新建）、`internal/controller/ucg_*`、`internal/controller/gateway_app_register.go`、`internal/controller/ucg_route_proxy.go`（或 `domain_route_proxy` 扩展）、`internal/controller/ucg_ws_proxy.go`（或扩展 `ws_route_proxy.go`）、`manifest/config/config.ucg-service.yaml`、`docker-compose` / 部署清单。
- **Affected systems**：gateway-app-server、ucg-service、device-service（新增 internal routes）、MySQL `ai_voice_ucg`、Redis、阿里云 OSS、阿里云 Green。
- **APIs**：App 对外 `/ucg/app/api/*` 与 `/ucg/app/ws/chat`（均经 gateway-app，与现有 App API 同域）；ucg-service 内部 `:9804` 提供 HTTP + `/ws/chat`；device internal `/device/internal/api/ucg/*`（路径以实现为准）。
- **Dependencies**：遵循 `history-service` 模板、`RegisterGatewayAppHTTP` 代理链、`ws_route_proxy.go` WS 透传模式；Green SDK、OSS SDK。
