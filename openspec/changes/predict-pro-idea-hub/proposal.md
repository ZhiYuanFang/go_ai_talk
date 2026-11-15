## Why

预测 Pro 本阶段从「付费开通壳」转为「收集用户对 Pro 能力的想法」。需要服务端独立承载想法的提交、公开列表、点赞热度与 HUB 运营管理（改展示文案/删除，并展示 wxId 与宝宝名），且不得混入既有设置「用户反馈」问答链路。

## What Changes

- 在 **device-service** / `ai_voice_device` 新增：
  - `predict_pro_idea`：想法原文、运营展示文案、wxId、deviceNo、提交时宝宝名快照、时间戳、软删。
  - `predict_pro_idea_like`：`(idea_id, wx_id)` 唯一，支撑点赞/取消与热度。
- App API（JWT；提交须绑定设备）：
  - `GET …/list`：未删除想法，**新在前**；项含 `id`、`text`（display_text）、`babyName`、`likeCount`、`likedByMe`、`isMine`；**不得**下发 wxId。
  - `POST …/submit`：提交想法；快照宝宝名；`display_text` 初始=原文。
  - `POST …/like` 与 `POST …/unlike`：更新点赞关系并返回最新热度/是否已赞。
- **门禁**：任意**已登录**用户可 list/submit/like（submit 另须绑定设备）；服务端 **MUST NOT** 校验预测 Pro 权益（`feature_user_entitlement` / cash）；Pro 入口若需限制仅由 Flutter 侧处理。
- **提交日限额**：同 `wx_id` 自然日（Asia/Shanghai）最多 10 条；以 `predict_pro_idea` 表按日 **COUNT** 判定；**不得**为此新加 Redis。
- Admin API（`X-Admin-Password`）：分页列表（wxId、宝宝名、原文、展示文案、热度等）；改 `display_text`；软删。
- Admin 静态页：独立管理页（镜像 `feedback-records`）+ `admin.html` 入口；近全宽 Hub 布局约定。
- **gateway-app 自检**：
  - 反代：`installDeviceProxyMiddleware` **MUST** 增加 `/device/app/api/predict-pro-ideas/*`（现网非 `/device/app/api/*` 总前缀）。
  - Admin HTML：`admin_static_pages` + `gateway_app_auth_exempt` 登记 `/device/admin/predict-pro-ideas`（对齐 feedback-records）。
  - App 路径 **MUST NOT** 进 Bearer 白名单（须 JWT）。
- **App usage 统计（已确认）**：**计入** usage；`api/v1` 须含 `g.Meta` summary；**不得**写入 `usagestats/maintenance_skip.go`。
- **本版不做**「已采纳」状态或自动发放 1 个月 Pro（客户端文案承诺，运营线下处理）。
- **不得**复用 `feedback` 表与 `/device/*/api/feedback/*`。
- 兄弟仓 Flutter：`predict-pro-idea-wall`；本变更不改 Flutter。

## Capabilities

### New Capabilities

- `predict-pro-idea-data-model`：想法表、点赞表、宝宝名快照与软删。
- `predict-pro-idea-api`：App list/submit/like、鉴权（登录即可）、校验、日限额 COUNT、排序、`isMine`。
- `predict-pro-idea-admin`：Admin 列表/改文案/删除与 HUB 页。

### Modified Capabilities

- （无）既有 `feedback-*` 不变；仅参照其 Admin 鉴权与页面模式。

## Impact

- 进程：**重部署** `device-service`、`gateway-app-server`（反代 + 静态页 + auth exempt）。
- 路径：`/device/app/api/predict-pro-ideas/*`、`/device/admin/api/predict-pro-ideas/*`、`/device/admin/predict-pro-ideas`。
- **不必重部署**：`cash-service` / `voice-service` / `ucg-service`（本变更不读 Pro 权益、不改支付）。
- 无新微服务进程；CI matrix N/A。
