## Why

history-service 的 `POST /device/history/api/chat` 当前**进程内**调用 `voice.TextChat`，导致 quota/qa/llm_lane 等 voice 库表访问落在 history 进程的 `default` DB（history 库）上，触发 `ai_quota_default` 不存在等错误，且违反「跨服务 MUST 走 HTTP 契约、禁止跨库直查他域表」的架构约定。suggest 已委派 voice-service HTTP，chat 应对齐同一模式。

## What Changes

- 新增 voice-service **internal** 文本对话 API（`POST /voice/internal/api/text/chat`），经内网 secret 鉴权，支持注入 `wxId` 供额度预检。
- history-service `HistoryCtrl.Chat` 改为 **HTTP 委派** voice-service，**移除**对 `voice.Voice()` 的进程内依赖。
- history-service **撤销** `VOICE_DB_LINK` / voice DB 分组配置；**回滚**为跨库直连 voice 表而引入的 `voiceDB()` 等改动（voice 库访问仅在 voice-service 进程内）。
- 委派层 MUST 透传 voice 域 AI 额度/登录业务码（40301/40302），App 侧行为与改前一致。
- 对外 App 路径 `/device/history/api/chat` **不变**（无 **BREAKING**）。

## Capabilities

### New Capabilities

- `voice-internal-text-chat`：voice-service 对内文本 chat 契约（鉴权、wxId、错误语义）。

### Modified Capabilities

- `history-voice-delegation`：history-service 文本 chat MUST 经 HTTP 访问 voice-service，禁止进程内执行 voice 业务。

## Impact

- `api/v1/voice_internal_text_chat_http.go`（新增）
- `internal/controller/voice_internal_text_chat.go`、`register_voice_service.go`
- `internal/services/history/delegate_http.go`（新增 chat 委派 + 业务码透传）
- `internal/controller/device_history.go`、`register_history_service.go`
- `internal/services/contracts/http_targets.go`
- `cmd/history-service/main.go`（移除 VOICE_DB 分组）
- `manifest/docker/docker-compose.microservices.yml`（history-service 移除 VOICE_DB_LINK）
- 回滚 `internal/services/voice/db.go` 及关联 `voiceDB()` 改造（若仍存在）
