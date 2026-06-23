## 1. voice internal chat API

- [x] 1.1 新增 `api/v1/voice_internal_text_chat_http.go`：`POST /voice/internal/api/text/chat`
- [x] 1.2 新增 `VoiceInternalTextChatCtrl`：secret 中间件组内注册，解析 `X-Internal-Wx-Id`，调用 `voice.TextChat`
- [x] 1.3 `http_targets.go` 增加 `VoiceInternalTextChatPath()`

## 2. history HTTP 委派

- [x] 2.1 扩展 `delegate_http.go`：支持请求 headers、envelope business code 透传（40301/40302）
- [x] 2.2 实现 `delegateTextChat(ctx, deviceNo, transcript, wxID)`
- [x] 2.3 `HistoryCtrl.Chat` 改调 delegate；移除 `Voice` 字段；更新 `NewHistoryCtrl` 与 `register_history_service.go`

## 3. 撤销跨库直连

- [x] 3.1 移除 history-service 的 `VOICE_DB_LINK` / voice DB 分组（`main.go`、compose、config 注释）
- [x] 3.2 回滚 `voice/db.go` 及 voice 包内 `voiceDB()`/`qaModel`/`suggestModel` 改造，恢复 voice-service 内 `g.DB()`/`dao.*` 直连

## 4. 部署与验证

- [x] 4.1 确认 history-service 环境含 `VOICE_SERVICE_URL` 与 `DEVICE_GATEWAY_INTERNAL_SECRET`（与 gateway-app 一致）
- [ ] 4.2 联调：App `POST /device/history/api/chat` 成功；额度用尽返回 40302；history 日志无 voice 库 SQL
