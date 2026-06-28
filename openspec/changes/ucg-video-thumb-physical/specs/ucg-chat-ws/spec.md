## ADDED Requirements

### Requirement: Chat messages with video SHALL include mediaThumbnailUrl as physical thumb CDN URL

ucg-service 在 enrich 聊天消息媒体（含 HTTP 消息列表、WS 实时 `message_delivered` 及审核通过后推送）时，当消息含非空 `videoKey` 时 MUST 设置 `mediaThumbnailUrl` 为 `BuildVideoThumbnailURL(videoKey)`。该 URL MUST NOT 含 `x-oss-process`。`mediaCdnUrl` MUST 仍为视频原片 CDN URL。

本要求 MUST 在 Chat WebSocket v1 首版即满足，MUST NOT 推迟至后续 API 版本。MUST NOT 改变 v1 消息 JSON 字段结构，仅填充/修正 `mediaThumbnailUrl` 值。

#### Scenario: WS 实时视频消息含 mediaThumbnailUrl

- **WHEN** 收件方经 WS 收到含 `videoKey` 的 `message_delivered` 且 OSS 存在对应 `_thumb.jpg`
- **THEN** 消息 body SHALL 含非空 `mediaThumbnailUrl` 且 URL SHALL NOT 含 `x-oss-process`

#### Scenario: HTTP 消息历史与 WS 语义一致

- **WHEN** 客户端 `GET` 会话消息列表且行含 `videoKey`
- **THEN** 每项 SHALL 含与 WS 相同规则的 `mediaThumbnailUrl`

#### Scenario: 图片消息行为不变

- **WHEN** 消息仅含 `imageKey`
- **THEN** `mediaThumbnailUrl` SHALL 仍为 `BuildImageThumbnailURL(imageKey)` 物理缩略图 URL
