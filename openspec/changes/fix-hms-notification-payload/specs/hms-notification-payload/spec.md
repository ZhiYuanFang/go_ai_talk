## ADDED Requirements

### Requirement: HMS 可见通知消息载荷

当 push-service 经 HMS channel 发送**非静默**且 `alert` 非空的推送时，请求体 MUST 包含顶层 `message.notification`（至少含 `title` 与 `body`），并且 MUST 在 `message.android.notification` 中提供合法的 `click_action`（默认打开应用：`type` 为 `3`；若配置了自定义 intent，则 `type` 为 `1` 且带 `intent`）。

系统 MUST NOT 仅依赖「只有 `android.data`、无顶层 notification」的形态作为可见通知的主路径。

#### Scenario: 预测临近可见推送

- **WHEN** `bizType=predict_imminent`（或其它非静默+非空 alert）经 HMS 下发
- **THEN** 发往华为的 JSON MUST 含 `message.notification.title` 与 `message.notification.body`，且 `message.android.notification.click_action` 合法

#### Scenario: 静默不伪造通知正文

- **WHEN** 推送为 silent 或 alert 为空
- **THEN** 系统 MUST NOT 以空/占位正文冒充完整用户可见通知消息（可仅 badge/data）

### Requirement: HMS 业务成功码判定

HMS `messages:send` 在 HTTP 2xx 时，系统 MUST 解析响应体业务字段 `code`：仅当成功码为 `80000000`（字符串或与之等价的表示）时，方可视为发送成功并记 `send_ok`；否则 MUST 视为失败并记 `send_failed`（日志含 `code` 与截断错误信息，MUST NOT 含完整 device token）。

#### Scenario: HTTP 200 但业务失败

- **WHEN** 华为返回 HTTP 200 且 `code` 非 `80000000`
- **THEN** dispatcher 路径 MUST 走失败日志（非 `send_ok`）

#### Scenario: 业务成功

- **WHEN** 华为返回 HTTP 2xx 且 `code` 为 `80000000`
- **THEN** 系统 MAY 记 `send_ok`
