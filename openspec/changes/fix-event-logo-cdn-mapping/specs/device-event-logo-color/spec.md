## MODIFIED Requirements

### Requirement: 事件 logo 与 color SHALL 可配置且列表可见

device-service 事件字典 MUST 支持 `logo`（OSS objectKey，前缀 `event/`）与 `color` 的持久化；所有返回事件字典列表的 HTTP 接口 MUST 在 JSON 中将 `logo` 序列化为 CDN 绝对 URL（无 logo 时为空串），并 MUST 包含 `color` 字段。Redis 事件选项缓存（`device:event:options:all`）MAY 仅持久化 objectKey；**任何** HTTP 响应边界 MUST 在返回前完成 CDN 映射，不得因缓存命中而跳过。

#### Scenario: 管理端事件列表返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/admin/api/event/list`
- **THEN** 响应 `list[]` 中每项 MUST 包含 `logo` 与 `color` 字段
- **AND** `logo` 若有值 MUST 为 CDN 绝对 URL（`https://` 开头）

#### Scenario: 历史与内部事件选项返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/history/api/event/options` 或 `GET /device/internal/api/event/options`
- **THEN** 响应 `list[]` MUST 同样包含 CDN 形式的 `logo` 与 `color`

#### Scenario: 管理端更新 logo 后 history 缓存命中仍返回 CDN

- **WHEN** 管理员成功更新某事件 logo 且 device-service 已重建 Redis 事件选项缓存（缓存内 logo 为 objectKey）
- **AND** App 或客户端请求 `GET /device/history/api/event/options` 且 history-service 命中该 Redis 键
- **THEN** 响应 `list[]` 中每项 `logo` 若有值 MUST 仍为 CDN 绝对 URL（`https://` 开头）
- **AND** MUST NOT 返回裸 objectKey（如 `event/2026/06/xxx.png`）

## ADDED Requirements

### Requirement: history 写 Redis 事件选项缓存 SHALL 仅持久化 objectKey

history-service 向 `device:event:options:all` 写入事件选项快照时，SHALL 将每项 `logo` 规范为 OSS objectKey（无前缀域名）；SHALL NOT 将 CDN 绝对 URL 写入共享 Redis 键，以避免与 device-service 重建缓存的 objectKey 格式混写。

#### Scenario: history 回源 internal 后写缓存

- **WHEN** history-service cache miss 后从 device internal 拉取事件列表（响应 logo 已为 CDN URL）
- **AND** history-service 将结果写入 Redis 事件选项键
- **THEN** 写入 Redis 的 JSON 中 `logo` MUST 为 objectKey 形式
- **AND** 后续 history HTTP 响应 MUST 仍通过 CDN 映射返回绝对 URL
