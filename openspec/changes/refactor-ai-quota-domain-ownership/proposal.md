## Why

当前 AI 月度额度（`polish`、`voice_ai`、`clinic_ai`）全部集中在 **device-service**（`ai_voice_device` 库表 + internal/app/admin API），且 **ucg-admin** 统一管理三个 feature 的全局默认与 per-wxId override。这违反「一服务一数据库、各域自管数据」的模块边界：**voice** 应拥有喂养 AI 与胖宝 AI 额度，**ucg** 应仅拥有润笔额度，**device** 不应承载 AI 配额业务。本变更在 `add-pangbao-ai-clinic-room` 落地 clinic 运行时之后，将配额层按域拆分，消除跨模块管理与跨库依赖。

## What Changes

- **BREAKING**：**device-service 完全退出** AI 配额——删除 `ai_quota_default` / `ai_quota_user_override` 表、DAO、internal API、App API 及 ucg 转发依赖。
- **voice-service** 在 **`ai_voice_voice`** 库自管 `voice_ai` + `clinic_ai` 配置（全局默认 + per-wxId override）与 Redis 用量（`ai:usage:voice_ai:*`、`ai:usage:clinic_ai:*`）；提供 internal check/consume、App 读 API、voice-admin API。
- **ucg-service** 在 **`ai_voice_ucg`** 库自管 `polish` 配置与 Redis 用量（`ai:usage:polish:*`）；internal check/consume 本地化，不再转发 device；ucg-admin 仅保留润笔字段。
- **App 读 API 拆分**（**无** gateway BFF 聚合）：
  - `GET /voice/app/api/ai-quota` → `{ voiceAi, clinicAi }`
  - `GET /ucg/app/api/ai-quota` → `{ polish }`
- **gateway-app-server** 注册 `/voice/app/api/*`、`/voice/admin/api/*` HTTP 反代至 voice-service（ucg 路径不变）。
- **voice-admin**：独立 HTML 页 `/device/admin/voice-admin.html`，Hub 导航入口（对齐 ucg-admin）；`VOICE_ADMIN_PASSWORD` env + `InjectAdminDownstreamPassword` 扩展。
- **ucg-admin**：移除 `voiceAi` / `clinicAi` 字段与相关 UI，**polish only**。
- **wxId-by-deviceNo**：从 ai-quota internal API 迁移至 device **user 域 API**（voice 反查登录态）。
- **迁移**：接受 15–30 分钟维护窗口配额写入停机，**不要求**双写；40301/40302 错误码语义不变；默认上限 voice_ai=5、clinic_ai=30、polish=5。
- **Flutter**：拆分 quota provider，分别从 voice/ucg App API 拉取，移除对 `/device/app/api/ai-quota` 的依赖。

## Capabilities

### New Capabilities

- `voice-ai-quota`：voice-service 权威维护 `voice_ai` + `clinic_ai` 配置、Redis 用量、internal check/consume、App 读 API 与 voice-admin 配置 API。
- `ucg-ai-quota`：ucg-service 权威维护 `polish` 配置、Redis 用量、本地化 internal check/consume 与 App 读 API。
- `voice-admin-ui`：voice 域 Admin HTML（全局默认 + per-wxId override）、Hub 导航与 gateway 口令注入。
- `gateway-voice-http-proxy`：gateway-app（及 gateway-service 对齐）对 `/voice/app/api/*`、`/voice/admin/api/*` 的 HTTP 反代注册。

### Modified Capabilities

- `ai-monthly-quota`：**移除** device-service 集中式权威与 `/device/app/api/ai-quota`；**移除** ucg 转发 device 的 admin 三字段模式；保留 40301/40302 跨 feature 错误语义与 Flutter 额度用尽弹框要求（数据源改为分域 API）。
- `gateway-app-server`：补充 voice App/Admin HTTP 反代路径注册（与既有 WS 透传并列）。

## Impact

- **go_ai_talk**
  - **删除**：`internal/services/device/ai_quota*`、`internal/controller/device_*ai_quota*`、`api/v1/device_ai_quota*`、device 域 ai_quota DAO/entity
  - **新增/迁移**：`internal/services/voice/ai_quota*`（voice + clinic）、voice DB 表与 DAO；`internal/services/ucg/ai_quota*`（polish 本地化）；voice/ucg controller 与 api/v1 契约
  - **gateway**：`gateway_app_register.go` voice HTTP 反代；`admin_auth_http.go` 扩展 `VOICE_ADMIN_PASSWORD`；`resource/public/voice-admin.html`、`admin-modules.js`
  - **ucg-admin.html**：移除 voice/clinic 字段
  - **配置/部署**：`VOICE_DB_LINK` 承载 quota 表；`UCG_DB_LINK` 承载 polish quota 表；`.env.example` / runbook 更新
- **flutter_ai_talk**（`d:\work\flutter_ai_talk`）：拆分 provider/model、更新 home/pangbao/ucg polish UI 与 `AiQuotaRemainingHint`
- **依赖**：须在 `add-pangbao-ai-clinic-room` clinic 运行时与 device 侧 clinic_ai 临时实现之后执行；维护窗口内一次性迁移 MySQL 配置行与 Redis 用量键
- **App API usage 统计**：新增 `GET /voice/app/api/ai-quota`、`GET /ucg/app/api/ai-quota` — **propose 阶段须向负责人确认是否计入 usage 统计**（tasks 留检查项）
