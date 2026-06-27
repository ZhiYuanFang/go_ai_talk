## MODIFIED Requirements

### Requirement: T5 and E1 SHALL handle chat with real users

每 `intervals.chat`（默认 1h，± jitter）T5（`chat_scan`）MUST 执行：

1. `GET /device/internal/api/sim/wx/ids` 取得**全量** sim wxId 列表
2. `POST /ucg/internal/api/chat/sim-unread-sample`，请求体含完整 `simWxIds`
3. 若无 eligible 未读（`found=false`）→ MUST `RecordTaskRun(chat_scan, false, "无未读会话")` 并结束
4. 若有命中 → 结合 sample 返回的 `messages` 经 `simText`（`chat_reply` prompt）生成 **一条**回复，并 `POST /ucg/internal/api/chat/send`

每 tick MUST 最多回复 **一条**会话。MUST NOT spawn detached goroutine 或临时聊天窗口循环。MUST NOT 使用 App `GET /conversations` 扫描全量未读。peer MUST 为真人（由 ucg sample 的 `peer NOT IN simWxIds` 保证）。

#### Scenario: Single reply per tick

- **WHEN** T5 tick 且 ucg sample 返回一条 eligible 未读
- **THEN** 系统 MUST 发送恰好一条 chat 消息并 MUST NOT 启动后台聊天 goroutine

#### Scenario: Skip when no unread

- **WHEN** 全量 simWxIds 非空但 ucg sample 返回 `found=false`
- **THEN** MUST 记录失败或 skip 语义（非 success 空消息）且 MUST NOT 调用 LLM

#### Scenario: Real peer only via sample

- **WHEN** 未读会话对端 wxId 属于 simWxIds
- **THEN** ucg sample MUST NOT 返回该会话，T5 MUST NOT 回复

#### Scenario: Full sim coverage

- **WHEN** sim 用户总数超过 200
- **THEN** T5 MUST 仍通过 ids 接口覆盖全部 sim 用户，MUST NOT 仅使用前 200 个 id

## REMOVED Requirements

### Requirement: E1 ephemeral chat window

**Reason**: 连续多轮回复改由 T5 周期轮询承担；删除 detached goroutine 降低扇出与运维复杂度。

**Migration**: 移除 `ephemeralChatLoop`/`ephemeralChatWindow` 配置；升级后已运行的 E1 goroutine 自然结束，不保证强杀。
