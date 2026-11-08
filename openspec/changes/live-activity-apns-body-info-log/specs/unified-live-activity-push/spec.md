## ADDED Requirements

### Requirement: Live Activity APNs request body MUST be logged at Info before send

Before each successful attempt to HTTP POST a Live Activity payload to APNs, after the final JSON request body has been built, the push-service MUST emit an **Info**-level log that includes at least the Live Activity `event` (`start`|`update`|`end`), the device-token length (or equivalent redaction), and the full request body JSON string. The log MUST NOT include the full device token value. This logging MUST apply to all Live Activity sends that go through the shared APNs send path (including update, start via push-to-start, end, and same-dispatch restart after invalid update). MUST NOT change send success/failure semantics solely due to logging.

每次经共享 APNs 发送路径发出 Live Activity 前，MUST 以 Info 记录 event、token 长度脱敏与完整请求体 JSON；MUST NOT 因日志改变发送语义。

#### Scenario: Update path logs body at Info

- **WHEN** the system is about to send a Live Activity APNs request with event=update
- **THEN** an Info log MUST be written containing event=update and the full JSON body
- **AND** the log MUST NOT contain the complete device token string

#### Scenario: Start via push-to-start also logs body

- **WHEN** the system sends event=start (including after invalid-update restart)
- **THEN** an Info log MUST likewise include the full start request body before the APNs POST
