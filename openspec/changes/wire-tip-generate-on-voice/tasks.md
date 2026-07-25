## 1. Go：voice Bind TipCtrl

- [x] 1.1 在 `internal/controller/register_voice_service.go` 的公开路由 Group 中 `Bind(NewTipCtrl())`（或等价），并补充中文业务注释说明 tip 宿主为 voice
- [x] 1.2 确认 `TipCtrl.Generate` / `writeSSEEvent` 可用且 SSE 方言为 `event:` + 纯文本；**不要** Bind feedback 控制器

## 2. Go：gateway 反代与约定核对

- [x] 2.1 在 `voice_route_proxy.go`（或同类）增加 `/device/tip/*` → VOICE_API_PROXY 反代 pattern，并补充中文注释
- [x] 2.2 核对 `gateway_app_auth_exempt.go`：**不**将 `/device/tip/generate` 加入豁免（需登录）
- [x] 2.3 核对 `usagestats/maintenance_skip.go`：**不**排除 tip generate（保持统计）

## 3. Flutter：路径与 SSE 方言（仓 `d:\work\flutter_ai_talk`）

- [x] 3.1 修改 `app/lib/data/tip_repository.dart`：URL 改为 `/device/tip/generate`；SSE 解析对齐 chat（`event:` + data 纯文本；done 解析 `answerId`；识别 `[DONE]`）
- [x] 3.2 修改 `app/lib/data/tip_models.dart`：用 `String? answerId` 替代流依赖的 `tipId int`
- [x] 3.3 修改 `app/lib/providers/tip_provider.dart`：流式累积写入 `answerId`；`submitFeedback` 字段对齐 `answerId`（完整反馈飞轮留给包 C）
- [x] 3.4 确认 home create → tip 触发链路无需改动即可调用更新后的 provider

## 4. 编译 / 静态核对

- [x] 4.1 `go build` 相关包（至少 controller / voice-service 注册路径可编译）
- [x] 4.2 Flutter 改动文件语法合理（`dart analyze` 针对改动文件，或等价静态检查）
