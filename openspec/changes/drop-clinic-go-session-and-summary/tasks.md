## 1. Clinic WS：去掉 session_sync

- [x] 1.1 修改 `internal/controller/voice_clinic_ws.go`：`auth_ok` 成功后不再调用 `BuildSessionSync` / 不下发 `session_sync`；中文注释说明 UI 历史由前端、多轮由 Python 负责
- [x] 1.2 确认读循环在 `auth_ok` 后即可接收 `question`/`cancel`

## 2. 删除 Go 对话会话与摘要/画像死路径

- [x] 2.1 删除 `internal/services/voice/clinic_session.go`（含 `appendClinicTurn` / `BuildSessionSync` / session Redis）
- [x] 2.2 删除 `internal/services/voice/clinic_summary.go`（含 `ensureClinicSummary` 与 summary Redis）
- [x] 2.3 删除 `internal/services/voice/clinic_profile.go`（含 `loadClinicBabyProfile`）
- [x] 2.4 改写 `clinic_service.go` `HandleQuestion`：去掉 summary/session/profile 调用与 `appendClinicTurn`；保留限流、额度、闸门、Python 流转发、`answer_done`
- [x] 2.5 瘦身 `clinic_llm.go`：去掉未使用的 `baby` / `summaryJSON` / `prior` 参数，仅转发 ClinicStream 所需字段；补中文注释

## 3. 配置与 cachekit 清理

- [x] 3.1 更新 `clinic_config.go`：移除 `SessionTTLSeconds` / `SummaryTTLSeconds` 及 session/summary 键前缀常量注释；收紧 `aiClinic` 加载默认值
- [x] 3.2 更新 `manifest/config/config.voice-service.yaml`：删除 `sessionTtlSeconds` / `summaryTtlSeconds`；删除或改写依赖「近 7 天摘要」的 `systemPrompt`（无引用则删除该字段）
- [x] 3.3 更新 `internal/platform/cachekit/keys_voice.go`：删除 `VoiceClinicSessionKey` / `VoiceClinicSummaryKey`（及前缀常量）；保留 rate 键；全文 grep 确认无残留引用

## 4. 隐私与自检

- [x] 4.1 修订 `resource/public/privacy-policy.html` 中与「本系统读取近 7 天喂养聚合摘要」冲突且归因于已删除 Go 管线的表述
- [x] 4.2 grep 确认：无 `session_sync` 下发、无 `voice:clinic:session` / `voice:clinic:summary` 写入、无 `loadClinicBabyProfile` / `ensureClinicSummary` / `appendClinicTurn`；`go build`（voice-service 相关包）通过
- [x] 4.3 对照本 change `specs/pangbao-ai-clinic/spec.md` 自检：六种下行 type、auth_ok 后无 sync、限流/额度仍生效
