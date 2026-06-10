## MODIFIED Requirements

### Requirement: 事件 logo 与 color SHALL 可配置且列表可见

device-service 事件字典 MUST 支持 `logo`（OSS objectKey，前缀 `event/`）与 `color` 的持久化；所有返回事件字典列表的 HTTP 接口 MUST 在 JSON 中将 `logo` 序列化为 CDN 绝对 URL（无 logo 时为空串），并 MUST 包含 `color` 字段。

#### Scenario: 管理端事件列表返回 logo 与 color

- **WHEN** 客户端携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/event/list`
- **THEN** 响应 `list[]` 中每项 MUST 包含 `logo` 与 `color` 字段
- **AND** `logo` 若有值 MUST 为 CDN 绝对 URL（`https://` 开头）

#### Scenario: 历史与内部事件选项返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/history/api/event/options` 或 `GET /device/internal/api/event/options`
- **THEN** 响应 `list[]` MUST 同样包含 CDN 形式的 `logo` 与 `color`

### Requirement: 事件新增与更新 SHALL 支持 multipart 上传 logo

`POST /device/admin/api/event/add` 与 `POST /device/admin/api/event/update` MUST accept `multipart/form-data`，至少包含表单字段 `name`、`eventType`、`extraNames`、`color`；`update` MUST 包含 `id`。可选文件字段名 MUST 为 `logo`。

#### Scenario: 新增事件并上传 logo

- **WHEN** 客户端 multipart 提交有效 `name` 与合法图片 `logo`
- **THEN** 服务端 MUST 经 ucg internal 上传至 OSS
- **AND** MUST 将 `event.logo` 设为 `event/...` objectKey
- **AND** MUST 将 `color` 写入 `event.color`（若提供）

#### Scenario: 更新事件未传 logo 保留原值

- **WHEN** 客户端对已有事件 multipart 更新且未包含 `logo` 文件
- **THEN** 服务端 MUST 保留原 `event.logo` objectKey
- **AND** MAY 更新 `color` 与其它文本字段

## REMOVED Requirements

### Requirement: 事件 logo 静态读 SHALL 由 device-service 提供

**Reason**: logo 改由 OSS/CDN 提供，不再使用宿主机 `/ai_talk_images/`。

**Migration**: 执行 `docs/runbooks/event-logo-oss-migration.md` 一次性迁移后部署本变更。
