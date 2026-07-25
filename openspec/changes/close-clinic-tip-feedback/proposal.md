## Why

诊疗（clinic）与小贴士（tip）的反馈控制器与 API 已落地，但未 Bind 到 voice-service，gateway 也未反代 `/device/api/clinic|tip/*`，Flutter tip 反馈字段仍为错误的 `tipId`/`feedbackResult`，clinic 端有 `answerId` 存储却无反馈 UI/HTTP。端到端反馈飞轮无法闭环，Python 侧已对齐的 Body 反馈接口无人调用。

## What Changes

- voice-service 注册 `DeviceClinicFeedbackController`，对外暴露已有 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`
- gateway-app 新增 `/device/api/clinic/*`、`/device/api/tip/*` → `VOICE_API_PROXY` 反代（与 device 域 `/device/app/api/feedback/*` 区分；需登录，不进 Bearer exempt）
- `maintenance_skip.go` 精确排除两反馈 POST（usage **不统计**）；tip generate 统计属包 B，本包不改
- Flutter tip：`submitFeedback` 字段对齐 `answerId` + `feedback`，URL `/device/api/tip/feedback`；流式 done 写入 `answerId`（依赖包 B SSE；若 B 未完成则补最小接线或报告 blocker）
- Flutter clinic：`pangbao_ai_screen` 补反馈 UI + HTTP POST `/device/api/clinic/feedback`（复用已有 `answerId`）
- **本包不做** tip generate Bind / `/device/tip/*` 反代 / SSE 方言整包改造（属包 B `wire-tip-generate-on-voice`）

## Capabilities

### New Capabilities

- `clinic-tip-feedback-voice-host`：将已有 clinic/tip feedback 控制器接到 voice-service，并由 gateway 反代 `/device/api/clinic/*`、`/device/api/tip/*` 到 voice；usage 精确排除两 feedback POST
- `clinic-tip-feedback-flutter`：Flutter tip/clinic 反馈字段与路径对齐 Go API，补齐 clinic 反馈 UI 与 tip `answerId` 提交

### Modified Capabilities

- （无）主规格库无对应反馈飞轮能力需增量修改；承接 `fix-python-api-alignment` 的 Body 对齐与 `tip-chat-streaming` 的 tip 宿主 = voice 决策

## Impact

- **Go**：`register_voice_service.go`、`voice_route_proxy.go`、`maintenance_skip.go`；确认已有 `api/v1/device_clinic_feedback_http.go` 与 `device_clinic_feedback_controller.go`
- **Flutter 仓**（`d:\work\flutter_ai_talk`）：`tip_repository.dart`、`tip_models.dart`、`tip_provider.dart`、`pangbao_ai_screen.dart`（及可选 clinic client）
- **契约**：App 路径保持 `/device/api/clinic/feedback`、`/device/api/tip/feedback`；Body `{answerId, feedback}`
- **usage**：负责人已确认 feedback **不统计** → 写入 `maintenance_skip.go`；`tip generate` 统计属包 B
- **依赖**：包 B `wire-tip-generate-on-voice`（tip SSE done → answerId）；本包不重复 Bind TipCtrl
- **非目标**：Python Body 改造（已完成）、新增测试文件、改 SSE 方言整包、device 域反馈建议接口
