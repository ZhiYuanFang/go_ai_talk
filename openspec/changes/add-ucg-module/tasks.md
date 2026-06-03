## 1. Phase 1 — ucg-service 脚手架 + DB + gateway HTTP/WS 代理 + OSS

- [x] 1.1 创建 `cmd/ucg-service/main.go`（镜像 `history-service`：配置、端口 `:9804`、`runtimecheck`）
- [x] 1.2 新增 `manifest/config/config.ucg-service.yaml`（DB `ai_voice_ucg`、Redis、OSS、Green、device internal URL；OSS AK/SK 见 `design.md`）
- [x] 1.3 执行 `design.md` DDL 创建 `ai_voice_ucg` 及全部 `ucg_*` 表（含 `hack/config.yaml` ucg 域配置与 `gf gen dao` 生成 9 表 DAO/entity/do）
- [x] 1.4 实现 `RegisterUcgServiceHTTP` 与健康检查
- [x] 1.5 实现 `ucg_route_proxy.go` + `gateway_app_register.go` 注册 `/ucg/app/api/*` HTTP 代理与 `UCG_SERVICE_BASE_URL`
- [x] 1.6 实现 UCG WebSocket 升级代理：`/ucg/app/ws/chat` → ucg-service `/ws/chat`（参考 `ws_route_proxy.go` / `installVoiceWSProxyMiddleware`；`UCG_WS_ROUTE_MODE` + `UCG_WS_PROXY_URL`）
- [x] 1.7 扩展 gateway Bearer 白名单：`/ucg/app/ws/chat` 与 UCG 匿名只读 HTTP 路径（见 `gateway-app-server` delta）
- [x] 1.8 实现 `POST /ucg/app/api/media/presign`（bucket pang-bao、prefix social/、AK 配置）
- [x] 1.9 docker-compose / 部署清单增加 ucg-service 服务项；同步更新 `manifest/docker/.env.example`（见下）

## 2. Phase 1 — device internal API（ucg-device-internal-api）

- [x] 2.1 device-service 新增 internal routes：`wx/validate`、`wx/batch`、`wx/{id}/baby-name`
- [x] 2.2 校验 `X-Device-Gateway-Internal-Secret` 中间件
- [x] 2.3 ucg-service HTTP client 封装调用上述 internal API

## 3. Phase 2 — Profile + Posts CRUD + Feed 基础（ucg-app-http-api）

- [x] 3.1 `GET/PUT /ucg/app/api/profile/me`（自动创建 `{babyName}的家长`）
- [x] 3.2 `GET /ucg/app/api/profile/{wxId}` 公开资料
- [x] 3.3 Posts CRUD + `GET /posts/mine`；status draft/pending 流转
- [x] 3.4 `GET /feed/recommend`（匿名可选）与 `GET /feed/following`（需 wxId）
- [x] 3.5 统一分页 `{ list, total, page, pageSize }`

## 4. Phase 3 — Green 审核 worker（ucg-green-audit Option A）

- [x] 4.1 集成阿里云 Green SDK（文本/图片/视频/资料字段）
- [x] 4.2 异步 worker：扫描 `pending_audit` posts 与 profile 变更
- [x] 4.3 pass → `published` + `published_at`；fail → `rejected` + `reject_reason`
- [x] 4.4 Feed 查询仅返回他人 `published`；作者 我的动态返回全 status

## 5. Phase 4 — 互动与关注

- [x] 5.1 Follow/unfollow + following 列表 API
- [x] 5.2 Like/unlike + 计数维护
- [x] 5.3 Comment CRUD + 计数维护
- [x] 5.4 关注 Feed SQL/缓存实现

## 6. Phase 5 — 聊天 Redis + WS（ucg-chat-ws + Green Option C）

- [x] 6.1 Redis 键设计：会话列表、消息永久存储、未读计数
- [x] 6.2 ucg-service 内部 `GET /ws/chat` WebSocket（JWT 首帧 auth）
- [x] 6.3 发消息 → pending → Green → deliver/fail（Option C WS 事件）
- [x] 6.4 `GET/POST /conversations*`、`GET messages`、`POST read` HTTP API
- [x] 6.5 会话 pin/delete（`ucg_conversation_member`）
- [x] 6.6 联调 gateway `/ucg/app/ws/chat` 升级代理与 Flutter 同域 WS URL

## 7. Phase 6 — 推荐算法与 polish（ucg-recommend-feed）

- [x] 7.1 后台任务计算 mixed score 写入 `ucg_post_recommend` 或 Redis ZSET
- [x] 7.2 推荐 Feed 按 score 分页；参数 tuning 文档化
- [x] 7.3 联调 Flutter：presign、发帖审核、经 gateway 聊天投递、服务边界评审（无跨库 wx DAO）
