## 1. Go：voice Bind 与 gateway

- [x] 1.1 在 `register_voice_service.go` Bind `DeviceClinicFeedbackController`（中文注释说明 clinic/tip 反馈宿主）
- [x] 1.2 在 `voice_route_proxy.go` 增加 `/device/api/clinic/*`、`/device/api/tip/*` → VOICE 反代（与 device `/device/app/api/feedback/*` 区分）
- [x] 1.3 在 `maintenance_skip.go` 精确排除 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`（usage 不统计；不碰 tip generate）

## 2. Flutter：tip 反馈对齐

- [x] 2.1 检查包 B `wire-tip-generate-on-voice`：若 tip SSE/answerId 未就绪，补最小 answerId 接线（不重复 Bind TipCtrl / tip generate 反代）
- [x] 2.2 `tip_models` / `tip_provider` / `tip_repository`：`submitFeedback` 改为 `answerId` + `feedback`，URL `/device/api/tip/feedback`

## 3. Flutter：clinic 反馈

- [x] 3.1 `pangbao_ai_screen`（及必要 client）补反馈 UI：有 `answerId` 且未反馈时展示 thumbs；POST `/device/api/clinic/feedback`

## 4. 验证

- [x] 4.1 `go build` 相关包（voice/controller/gateway usagestats）通过
- [x] 4.2 确认 Flutter 改动语法合理（字段/URL 一致；无残留 tipId/feedbackResult）
