## ADDED Requirements

### Requirement: tip 反馈字段与 URL 对齐 Go API
Flutter tip 客户端 SHALL 以 `POST /device/api/tip/feedback` 提交反馈，JSON Body 使用 `answerId`（string）与 `feedback`（int：1 或 -1），MUST NOT 继续使用 `tipId` / `feedbackResult`。

#### Scenario: tip submitFeedback 对齐
- **WHEN** 用户在首页小贴士面板点击 thumbs up/down 且本地已有非空 `answerId`
- **THEN** 客户端 SHALL 向 `/device/api/tip/feedback` 发送 `{ "answerId": "<id>", "feedback": 1|-1 }`
- **AND** 已反馈后 MUST NOT 允许再次更改

#### Scenario: 无 answerId 不提交
- **WHEN** 本地 `answerId` 为空或缺失
- **THEN** 客户端 MUST NOT 发起 tip feedback HTTP 请求

### Requirement: tip 流式 done 写入 answerId
Flutter tip 流式接收完成后 SHALL 将服务端返回的 `answerId` 写入 tip 状态，供反馈提交使用。该能力依赖包 B tip SSE；若 B 未完成，本变更 SHALL 补最小接线或在实现报告中标明 tip 反馈 blocker。

#### Scenario: done 携带 answerId
- **WHEN** tip SSE 流结束且 done 帧含 `answerId`
- **THEN** tip 状态 SHALL 保存该字符串，反馈按钮可用

#### Scenario: 包 B 阻塞时的降级
- **WHEN** 包 B 未完成导致 tip 无法获得 `answerId`
- **THEN** 实现报告 SHALL 明确 tip 反馈 blocker
- **AND** clinic 反馈与 Go Bind/反代/skip MUST 仍可完成

### Requirement: clinic 反馈 UI 与 HTTP
胖宝诊疗屏（`pangbao_ai_screen`）SHALL 在助手回答完成且存在 `answerId` 时展示 thumbs up/down，并经 HTTP `POST /device/api/clinic/feedback` 提交 `{answerId, feedback}`。

#### Scenario: clinic 回答完成后可反馈
- **WHEN** 某条助手消息已 `answer_done` 且 `answerId` 非空，且尚未反馈
- **THEN** UI SHALL 展示反馈按钮
- **AND** 用户点击后 SHALL POST `/device/api/clinic/feedback` 并标记该条已反馈

#### Scenario: 无 answerId 或已反馈
- **WHEN** `answerId` 为空，或该条 `feedbackGiven` 已为 true
- **THEN** UI MUST NOT 再次提交 clinic feedback
