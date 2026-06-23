## Context

- App 经 gateway-app 调 `POST /device/history/api/chat`；gateway 代理至 history-service。
- history-service 已用 `delegate_http.go` 委派 suggest → voice、event options/profile → device。
- `HistoryCtrl.Chat` 是唯一在 history 进程内调用 `voice.TextChat` 的路径；会触发 `ai_quota_default`、`qa`、`llm_lane_config`、`suggest` 等 voice 库读写。
- voice-service 已有 internal 额度 API（`/voice/internal/api/ai-quota/check|consume`），但 chat 整段未 HTTP 化。
- 临时方案（history 配置 `VOICE_DB_LINK` + `voiceDB()`）违反服务边界，本变更撤销。

## Goals / Non-Goals

**Goals:**

- history-service **零 voice 库访问**；chat 全量在 voice-service 执行。
- 对外 `/device/history/api/chat` 请求/响应与错误码语义不变。
- internal chat 契约与 suggest/ai-quota internal 路由风格一致（voice-service 注册 + history HTTP client）。

**Non-Goals:**

- 变更 App 调用路径或 WS 语音链路。
- 改造 `/voice/text/chat`（Admin 口令调试接口）；新建 internal 专用路径。
- history-service 内嵌 voice 包的其他潜在用途（当前无）。

## Decisions

### 1. 新增 internal 文本 chat API

```
POST /voice/internal/api/text/chat
Headers:
  X-Device-Gateway-Internal-Secret  (必须)
  X-Internal-Wx-Id                  (App 经 gateway 注入，可选但 chat 额度需要)
Body:
  { "deviceNo": "...", "transcript": "..." }
Response:
  { "reply": "..." }
```

- 注册在 `deviceUcgInternalSecretMiddleware` 保护的路由组（与 ai-quota internal 同组）。
- Handler 内 `voice.WithVoiceWxID(ctx, wxId)` + `voice.Voice().TextChat(...)`。
- 额度/登录错误经现有 `mapAIQuotaErr` 映射为 40301/40302。

**不复用** `/voice/text/chat`（需 `X-Admin-Password`，不适合 App 链路）。

### 2. history 委派实现

- `delegateTextChat(ctx, deviceNo, transcript, wxID int64) (reply string, err error)` 于 `history/delegate_http.go`。
- 扩展 HTTP helper：支持附加 headers（internal secret、wxId）；解析响应 envelope 时 **保留 `code`**，映射为 `gerror.NewCode`（40301/40302 与 voice 一致）。
- `HistoryCtrl.Chat` 仅调用 delegate；**移除** `HistoryCtrl.Voice` 字段与 `NewHistoryCtrl` 的 voice 参数。
- `register_history_service.go`：`NewHistoryCtrl(history.DeviceHistory())`。

### 3. 配置与部署

- history-service compose/env：**删除** `VOICE_DB_LINK`；保留已有 `VOICE_SERVICE_URL`（history 容器内已配置）。
- history-service `main.go`：**删除** `dbcfg.ApplyGroupFromEnv(..., "voice", "VOICE_DB_LINK")`。
- voice-service 无需新增 DB 配置；继续 `VOICE_DB_LINK` → default + voice 分组（可选，仅 voice 进程）。

### 4. 回滚临时跨库改动

- 删除 `internal/services/voice/db.go`（或恢复 voice 包内 `g.DB()`/`dao.Qa` 直连，仅 voice-service 使用）。
- `ai_quota_store.go`、`llm_lane_store.go`、`qa`/`suggest` 等恢复 `g.DB()` / `dao.*`（voice-service 进程内 default=voice 库）。

### 5. 超时与失败

- chat 委派 HTTP timeout 建议 **30s**（LLM 可能较慢；原进程内调用无额外 hop 限制）。
- voice-service 不可达：history 返回 5xx/明确 delegate 错误，不 fallback 进程内 voice。

## Risks / Trade-offs

- [多一跳 HTTP + LLM 延迟] → 30s client timeout；可观测日志 `[history-local-delegate] text_chat`。
- [internal secret 未配置] → history 启动或首次 chat 失败；compose 已有 `DEVICE_GATEWAY_INTERNAL_SECRET` 于 gateway-app，history 需同值 env（若尚未注入则本变更一并补充 compose 注释/变量）。

## Migration Plan

1. 部署 voice-service（含新 internal 路由）。
2. 部署 history-service（委派 + 移除 VOICE_DB_LINK）。
3. 验证 App `POST /device/history/api/chat` 与额度错误码。
4. 回滚：还原 history 内嵌 `voice.TextChat`（不推荐，仅应急）。

## Open Questions

- history-service 容器是否已注入 `DEVICE_GATEWAY_INTERNAL_SECRET`：实现阶段检查 compose，若缺失则与 gateway-app 对齐注入。
