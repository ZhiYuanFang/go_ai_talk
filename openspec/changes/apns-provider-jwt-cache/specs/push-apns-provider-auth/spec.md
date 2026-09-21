## ADDED Requirements

### Requirement: APNs Provider JWT 进程内缓存与复用

push-service 经 APNs HTTP/2 发送推送时，MUST 使用 Apple Provider Authentication Token（JWT，`Authorization: bearer`）。同一进程内，对同一配置密钥（`keyId`/`teamId`/密钥材料）MUST 缓存并复用已签发的 JWT，MUST NOT 在每一次设备发送时重新签发带新 `iat` 的 JWT。

缓存的 JWT 刷新间隔 MUST 不短于 20 分钟，且 MUST 在签发后 60 分钟内完成下一次刷新（实现推荐约 45–50 分钟复用窗口）。并发获取 bearer 时 MUST 互斥，避免短窗内并行签发多把不同 `iat` 的 token。

本要求仅约束 Provider JWT；设备 device token 的注册、删除与无效判定语义 MUST NOT 因此改变。MUST NOT 为此引入新的 Redis 键或跨进程共享 JWT（除非未来独立变更明确批准）。

#### Scenario: 连续两次 APNs 发送复用同一 Provider JWT

- **WHEN** push-service 在缓存有效期内向两个不同 device token 连续发送 APNs 推送
- **THEN** 两次请求 MUST 使用同一 Provider JWT 字符串（或等价：同一 `iat`），MUST NOT 为第二次发送重新签发新 JWT

#### Scenario: 缓存过期后才换发

- **WHEN** 缓存中的 Provider JWT 已超过实现设定的复用窗口（且该窗口 ≥20 分钟、<60 分钟）
- **THEN** 下一次 APNs 发送 MUST 签发新 JWT 并更新缓存，供后续请求复用

#### Scenario: 并发发送不并行换发

- **WHEN** 多个 goroutine 在缓存未命中时同时请求 APNs bearer
- **THEN** 系统 MUST 经互斥使该窗口内至多成功签发一把新 JWT 并写入缓存，其余调用方复用该结果

### Requirement: TooManyProviderTokenUpdates 不得删除设备 token

当 APNs 响应为 HTTP 429 且 reason 为 `TooManyProviderTokenUpdates`（或等价 Provider token 更新过频语义）时，系统 MUST 将该次发送记为失败（如 `send_failed`），MUST NOT 将该 reason 视为设备 token 无效，MUST NOT 因此删除对应 `push_device` 行。

#### Scenario: 429 Provider 限流不删设备

- **WHEN** APNs 返回 status=429 且 reason=`TooManyProviderTokenUpdates`
- **THEN** dispatcher MUST 记发送失败，且 MUST NOT 调用按设备 id 删除 token 的路径
