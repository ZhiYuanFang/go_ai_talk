## 1. Tip SSE 下线

- [x] 1.1 从 `register_voice_service` 卸掉 `NewTipCtrl()` Bind；删除 `internal/controller/voice/device_tip.go`
- [x] 1.2 删除 `api/v1/device_tip_http.go`（或等价 tip generate 契约）
- [x] 1.3 删除 `VoiceService.TipStream` 及 `python_ai_client` 中 tip stream 方法/请求类型；清理 `contracts` 中仅 tip 使用的 `TipStream*`
- [x] 1.4 gateway：从 `voice_route_proxy` 移除 `/device/tip/*` 反代 pattern 与注释

## 2. Clinic / Tip HTTP 飞轮下线

- [x] 2.1 卸掉 `DeviceClinicFeedbackController` Bind；删除控制器文件
- [x] 2.2 删除 `api/v1/device_clinic_feedback_http.go`（clinic + tip feedback 契约）
- [x] 2.3 删除 Python 客户端 `ClinicFeedback` / `TipFeedback`（及仅被其使用的类型）；确认不影响 clinic WS / care-alert
- [x] 2.4 gateway：移除 `/device/api/clinic/*`、`/device/api/tip/*` 反代 pattern
- [x] 2.5 从 `maintenance_skip.go` 删除 clinic/tip feedback 排除项与过时注释

## 3. 运维与文档

- [x] 3.1 删除 `resource/public/history.html` 中 tip generate 调用与相关 UI
- [x] 3.2 若 `docs/runbooks/*` 仍描述 tip 运维调试，删除或改写相关段落

## 4. care-alert 保留验收（不得误伤）

- [x] 4.1 确认 care-alert 控制器仍 Bind；`POST /device/api/care-alert/feedback` 与 daily/delete 仍在
- [x] 4.2 确认 gateway 仍反代 `/device/api/care-alert/*`；`CareAlertFeedback` 服务与 Python CareAlertFeedback 仍在
- [x] 4.3 确认 clinic WS（`/voice/clinic/ws`）注册与 WS 反代未改坏

## 5. 编译与门禁

- [x] 5.1 `go build` voice-service 与 gateway-app-server 通过
- [x] 5.2 仓库内 grep：无残留 `/device/tip/generate`、`TipStream` 业务入口、clinic/tip feedback 路由注册（允许 openspec 历史文档）
- [x] 5.3 运行 `hack/check-service-import.ps1`（或 `.sh`）通过
