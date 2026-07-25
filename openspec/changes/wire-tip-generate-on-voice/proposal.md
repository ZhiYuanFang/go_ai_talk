## Why

`tip-chat-streaming` 已落地 `TipCtrl`、`api/v1` 的 `POST /device/tip/generate` 与 voice 侧 `TipStream`，但控制器未 Bind 到 voice-service，gateway 也未反代 `/device/tip/*`，Flutter 仍走错误路径且按 JSON 帧解析 SSE，导致端到端小贴士流式生成不可用。本变更承接该「接线遗漏」，不硬扩已勾完的 tip-chat-streaming tasks。

## What Changes

- voice-service 注册 `NewTipCtrl()`，对外提供已有 `POST /device/tip/generate` SSE
- gateway-app 新增 `/device/tip/*` → `VOICE_API_PROXY` 反代；需登录（不进 Bearer exempt）；usage **统计**（不写入 maintenance_skip）
- Flutter `tip_repository` 改为 `/device/tip/generate`，SSE 解析对齐 Go 方言（`event:` + data 纯文本；done 解析 `answerId`）
- tip models / provider 改用 `answerId`（字符串），弃用不正确的流内 `tipId int`
- **本包不做** clinic tip feedback 控制器 Bind 与完整反馈飞轮（属包 C `close-clinic-tip-feedback`）；反馈 submit 字段可先对齐 `answerId`

## Capabilities

### New Capabilities

- `tip-generate-voice-host`：将孤儿 TipCtrl 接到 voice-service，并由 gateway 反代 `/device/tip/*` 到 voice
- `tip-flutter-sse-go-dialect`：Flutter tip 客户端路径与 SSE 方言对齐 Go（event + 纯文本 + answerId）

### Modified Capabilities

- （无）主规格库无对应 tip 宿主/接线能力需增量修改；承接既有 change `tip-chat-streaming` 的 Decision 3（tip 宿主 = voice）

## Impact

- **Go**：`register_voice_service.go`、`voice_route_proxy.go`（或同类 domain 反代）；确认 `device_tip.go` 可用；不 Bind feedback
- **Flutter 仓**（`d:\work\flutter_ai_talk`）：`tip_repository.dart`、`tip_models.dart`、`tip_provider.dart`；home create → tip 触发链路保持
- **契约**：App 路径仍为 `/device/tip/generate`（api/v1 已有）；SSE 方言锁定跟 Go
- **usage**：`POST /device/tip/generate` 计入统计
- **非目标**：feedback 完整闭环、新增测试文件、改 history 宿主
