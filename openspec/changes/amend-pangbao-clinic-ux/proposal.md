## Why

`add-pangbao-ai-clinic-room` 已落地胖宝 WS 与 Flutter 初版，但上线前产品反馈三类体验缺口：**用户可见名称**仍混用「胖宝 AI / AI 诊室」，与「胖宝诊疗」品牌不一致；**离开诊疗页再进入**时聊天列表为空，尽管服务端 Redis 已存 12h 会话轮次；**思考区折叠**在流式输出时展示顶部旧行而非最新内容，违背原 spec 的 auto-scroll 意图。本变更在既有 `clinic`/`clinic_ai` 路径与额度域上做小步 UX 修正，不引入新后端域或 Redis 键。

## What Changes

- **修改** 用户可见文案统一为 **「胖宝诊疗」**（AppBar、consent 标题、额度 hint、输入 hint 可选「问问胖宝诊疗…」）；首页沉浸式头部 tooltip **保持「胖宝」**；首页品牌标题 **保持「胖宝」**；内部路径/API/Redis 键 `clinic`/`clinic_ai`/`voice:clinic:*` **不变**。
- **新增** WS 下行帧 **`session_sync`**：`auth_ok` 成功后（含每次重连后重新 `auth`）下发 Redis `voice:clinic:session:{wxId}` 中已完成的 Q&A 轮次；payload `{ turns: [{question, answer}], expiresAt }`；**不含** thinking（Redis 未存）。
- **修改** Flutter 收到 `session_sync` 后填充 `_items` 聊天气泡（仅已完成轮次）；重连不再呈现空白历史。
- **修改** thinking 折叠区：内层 `ScrollController` + `thinking_delta` 时 `jumpTo(maxScrollExtent)`，折叠态 **底对齐最新行**（对齐 `home_voice_message_strip`）；可选顶部渐变、流式 pulse 光标、「跟随最新」恢复按钮。
- **修改** 相关 OpenSpec / 隐私政策等 **用户向** 表述：「胖宝 AI 诊室」→「胖宝诊疗」。
- **不新增** 测试文件、背景 ticker、新 Redis 键、网关路径变更。

## Capabilities

### New Capabilities

（无——均为既有能力的增量修正。）

### Modified Capabilities

- `pangbao-ai-clinic`：`auth_ok` 后下发 `session_sync`；帧协议与 handler 用户向命名对齐「胖宝诊疗」。
- `pangbao-ai-clinic-flutter`：用户可见命名、session 恢复、thinking 内层滚动与跟随最新交互。
- `app-legal-docs`：隐私政策中胖宝功能用户向名称改为「胖宝诊疗」。

## Impact

- **go_ai_talk**
  - `internal/controller/voice_clinic_ws.go`：`auth_ok` 后读 session 并写 `session_sync`
  - `internal/services/voice/clinic_session.go`：暴露 session 读取与 `expiresAt` 计算（复用现有 Redis 结构）
  - `resource/public/privacy-policy.html`：用户向名称
  - OpenSpec 变更增量（本 change）
- **flutter_ai_talk**（`d:\work\flutter_ai_talk`）
  - `pangbao_ai_screen.dart`：AppBar「胖宝诊疗」、session_sync 填充、`_ThinkingBlock` 内层滚动
  - `clinic_ws_client.dart`：解析 `session_sync`
  - `ai_quota_remaining_hint.dart`、`pangbao_ai_consent_store` 相关文案
  - `home_immersive_header.dart`：tooltip 保持「胖宝」（确认不改为「胖宝诊疗」）
- **依赖**：建立在 `add-pangbao-ai-clinic-room`、`refactor-ai-quota-domain-ownership` 已实现能力之上；无新跨服务契约
- **App API usage 统计**：无新 HTTP 接口，无需 maintenance_skip 变更
