## Why

兄弟仓 `flutter_ai_talk` change `live-activity-native-register-no-auth` 将在系统 Push-to-Start 唤醒窗口内由 **iOS 原生**直报 `POST /app/api/push/live-activity/register`，不经 Flutter 登录壳。当前 gateway-app **强制 Bearer**，push 控制器只认 `X-Internal-Wx-Id`，原生无 JWT 时请求必然 401，activity token 仍落不了库，对端 update/end 与 pending 清场继续堆积。须在本仓对该 path **免 token 鉴权**，并按 body.`wxId` 落库，与客户端 twin 契约对齐。

## What Changes

- **BREAKING（鉴权）**：`POST /app/api/push/live-activity/register` 在 gateway-app **必须**加入 Bearer 豁免（与登录类匿名 path 同族维护）；无 Authorization 时 **必须**仍可到达 push-service 业务 handler。
- **BREAKING（身份来源）**：成功 upsert **必须**以请求体非零 `wxId` 为准落库；**不得**再把「仅有内部头 `X-Internal-Wx-Id` / 仅有有效 Bearer」当作该 path 的唯一身份来源。仍带 Bearer 时 MAY 注入内部头，但落库身份 MUST 以 body.`wxId` 为准（防头与 body 不一致时的歧义）。
- 请求体 **必须**新增/要求非零 `wxId`；既有 `deviceKey`、`tokenKind`、`attributesKind`、`token`、`activity` 时 `activityId` 规则不变；表仍为 `push_live_activity_token`，upsert 语义不变。
- **不得**对 `/app/api/push/register`、`unregister` 做同样免鉴权；**不得**改 Live Activity 扇出/1min/清场主逻辑（仅 token 入口契约）。
- 产品接受公开写面风险（最坏污染 token → 无法自动关 LA）；本变更 **不**强制限流（MAY 后续另开）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `predict-imminent-live-activity-push`：LA register path 免 Bearer；按 body.`wxId` upsert `push_live_activity_token`；与 Flutter twin `live-activity-native-register-no-auth` 对齐。

## Impact

- **gateway-app**：`gateway_app_auth_exempt.go`（精确 POST 豁免本 path）；确认 `StripSpoofedInternalHeaders` 仍先剥伪造内部头。
- **push-service**：`api/v1/app_push_http.go` 请求体加 `wxId`；`LiveActivityRegisterPost` 改读 body.`wxId`；`RegisterLiveActivityToken` 入参来源不变（已按 wxID 写表）。
- **usage**：本 path 已在 `maintenance_skip`，一般无需再改。
- **兄弟仓**：`flutter_ai_talk` / `live-activity-native-register-no-auth`（原生免鉴权 POST + 缓存 wxId）。
- 对照既有：`predict-live-activity-push`。
- **契约定型（已定）**：`api/v1` **同 path 原地加 `wxId`**，不走 `api/v2`。activity token register 尚未对用户发版，无历史客户端依赖旧「Bearer + 头身份」契约；相对 AGENTS「v1 结构冻结」按未发布 path 首次定型处理（见 design 决策 3）。
