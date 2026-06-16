## go_ai_talk

### 1. session_sync 下行帧

- [x] 1.1 `clinic_session.go`：新增 `BuildSessionSync(ctx, wxID)`（或等价），读 Redis session、过滤 question/answer 均非空轮次、计算 `expiresAt`（`firstQuestionAt + TTL`）
- [x] 1.2 `voice_clinic_ws.go`：`auth_ok` 成功后立即 `writeJSON(session_sync)`；读 session 失败时下发空 `turns` 并打 warning 日志，不阻断后续 `question`
- [x] 1.3 补充中文注释：`session_sync` 时序、不含 thinking、与 12h 固定 TTL 关系

### 2. 用户向文档

- [x] 2.1 `resource/public/privacy-policy.html`：「胖宝 AI 诊室」→「胖宝诊疗」；更新生效日期
- [x] 2.2 代码内用户向注释/日志（如有「胖宝 AI 诊室」）对齐「胖宝诊疗」；**不**改 `clinic` 包名、路径、Redis 键

### 3. 验收（go_ai_talk）

- [x] 3.1 手工验证：`auth` → `auth_ok` → `session_sync`（有/无 session、expiresAt 合理）
- [x] 3.2 手工验证：重连后再次 `auth` 收到一致 `session_sync`；`session_sync` 帧无 thinking 字段
- [x] 3.3 确认无新增 `*_test.go`、无 background ticker、无新 Redis 键

---

## flutter_ai_talk

### 1. 用户可见命名

- [x] 1.1 `pangbao_ai_screen.dart`：AppBar title **「胖宝诊疗」**；consent 标题改为「使用胖宝诊疗前请知悉」（或等价）
- [x] 1.2 `ai_quota_remaining_hint.dart`（及诊疗页内联 hint）：**「本月胖宝诊疗剩余 N 次」**
- [x] 1.3 输入 hint（可选）：**「问问胖宝诊疗…」**；`home_immersive_header.dart` tooltip **保持「胖宝」**不变

### 2. session_sync 恢复

- [x] 2.1 `clinic_ws_client.dart`：解析 `session_sync` 帧并暴露给页面（stream/callback）
- [x] 2.2 `pangbao_ai_screen.dart`：收到 `session_sync` 以服务端 turns **全量填充** `_items`（user + assistant + 免责）；不渲染历史 thinking；不与进行中 `_activeAssistant` 冲突

### 3. Thinking 内层滚动

- [x] 3.1 `_ThinkingBlock` 改为 `StatefulWidget`：内层 `ScrollController`；折叠态 **移除** `NeverScrollableScrollPhysics`，改用 `ClampingScrollPhysics`
- [x] 3.2 `thinking_delta` / `didUpdateWidget`：post-frame `jumpTo(maxScrollExtent)`（对齐 `home_voice_message_strip`）
- [x] 3.3 内层 scroll pin 检测 + **「跟随最新」** 按钮恢复 auto-scroll
- [x] 3.4 （可选）折叠区顶部渐变 fade、流式 pulse 光标

### 4. 验收（flutter_ai_talk）

- [x] 4.1 端到端：提问完成 → 退出诊疗页 → 再进入 → 历史 Q&A 展示（12h 内）
- [x] 4.2 流式 thinking 折叠态展示**最新行**；pin 后停止跟随；「跟随最新」恢复
- [x] 4.3 文案：AppBar/consent/额度为「胖宝诊疗」；首页 tooltip 仍为「胖宝」
