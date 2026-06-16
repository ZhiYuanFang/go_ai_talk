## go_ai_talk

### 1. WS 协议与连接状态

- [x] 1.1 `voice_clinic_ws.go`：扩展帧解析，支持 `question{text, turnId}`、`cancel{turnId}`；缺少/空 `turnId` 返回 400 error
- [x] 1.2 `voice_clinic_ws.go`：连接级 `activeTurnID` + `cancelTurn` + `turnMu`；读循环非阻塞，每条 `question` 启动 goroutine
- [x] 1.3 `voice_clinic_ws.go`：新 `question` 时 supersede 旧 turn（cancel ctx + 写 `turn_cancelled{reason:superseded}`）；`cancel` 匹配 active 时写 `turn_cancelled{reason:cancelled}`
- [x] 1.4 `voice_clinic_ws.go`：连接关闭 defer 取消 active LLM ctx；补充中文注释（非阻塞读循环、显式 cancel 语义）

### 2. HandleQuestion 可取消与下行帧

- [x] 2.1 `clinic_service.go`：`HandleQuestion` 签名增加 `turnCtx`、`turnId`；所有下行帧含 `turnId`
- [x] 2.2 `clinic_service.go` / `streamClinicLLM`：LLM HTTP 流使用 `turnCtx`；ctx 取消时停止读流、不写 `answer_done`
- [x] 2.3 `clinic_service.go`：turn 正常结束写 `answer_done`（含 `turnId`）；cancel 路径不写 `answer_done`
- [x] 2.4 新增或内联 `emitTurnCancelled(writeJSON, turnId, reason)` 供 WS handler 与 service 复用

### 3. 额度、限流与 session

- [x] 3.1 `clinic_rate.go`：将 `INCR` 从 question 入口移至 **`answer_done` 成功后**（`recordClinicRateLimitOnSuccess`）；question 前仅检查窗口计数
- [x] 3.2 `clinic_service.go`：cancel/supersede/disconnect 路径 **MUST NOT** `ConsumeClinicAIQuota`；仅 success `answer_done` 后 consume
- [x] 3.3 `clinic_service.go`：仅 `answer_done` 成功路径 `appendClinicTurn`；partial turn 不写入 session
- [x] 3.4 补充中文注释：cancel 不扣费、限流仅计成功轮次

### 4. 验收（go_ai_talk）

- [x] 4.1 手工验证：question 流式中发 `cancel` → `turn_cancelled{cancelled}`，无 `answer_done`，额度不变
- [x] 4.2 手工验证：流式中发新 question → 旧 turn `turn_cancelled{superseded}`，新 turn 正常 `answer_done` 且仅扣 1 次额度
- [x] 4.3 手工验证：WS  abrupt 关闭 → LLM 停止（日志可观测），无 session append
- [x] 4.4 确认无新增 `*_test.go`、无 background ticker、无新 Redis 键

---

## flutter_ai_talk

### 1. ClinicWsClient 协议

- [x] 1.1 `clinic_ws_client.dart`：`sendQuestion` 生成 UUID `turnId` 并上行 `{type:question, text, turnId}`
- [x] 1.2 `clinic_ws_client.dart`：新增 `sendCancel(turnId)` → `{type:cancel, turnId}`
- [x] 1.3 `clinic_ws_client.dart`：解析 `turn_cancelled`、下行 delta/`answer_done` 的 `turnId`；暴露 active turnId 状态

### 2. 页面生命周期与 cancel

- [x] 2.1 `pangbao_ai_screen.dart` / client：`dispose` 与 `AppLifecycleState.paused` **先** `sendCancel(activeTurnId)` 再 `disconnect`
- [x] 2.2 移除流式期间全局 `_busy` 输入锁；保留发送/停止：新问 supersede、停止发 cancel
- [x] 2.3 收到 `turn_cancelled` 清除进行中 thinking/answer UI；`cancelled` 可选 toast「已停止」

### 3. Stale 帧过滤与改问 UX

- [x] 3.1 `pangbao_ai_screen.dart`：仅应用与 `activeTurnId` 一致的 `thinking_delta` / `answer_delta` / `answer_done`
- [x] 3.2 流式中发送新问题：新 user 气泡 + 新 turnId，等待新流式回复
- [x] 3.3 （可选）tap 用户气泡编辑预填 text，发送新 question（新 turnId）supersede

### 4. 验收（flutter_ai_talk）

- [x] 4.1 端到端：thinking 流中点停止 → UI 停止更新，额度不减少（对比 Admin/读 API）
- [x] 4.2 端到端：thinking 流中改问 → 旧流停止，新回答正常完成
- [x] 4.3 离开诊疗页 / App 后台 → 服务端 LLM 停止（配合 go 日志或断流观察）
- [x] 4.4 确认 `session_sync` 历史仍正常；被取消 partial 不出现在重进页面历史
