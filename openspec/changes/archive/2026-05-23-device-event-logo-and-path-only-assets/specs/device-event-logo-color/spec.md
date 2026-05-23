## ADDED Requirements

### Requirement: 事件 logo 与 color SHALL 可配置且列表可见

device-service 事件字典 MUST 支持 `logo`（应用内路径）与 `color`（色值字符串）的持久化；所有返回事件字典列表的 HTTP 接口 MUST 在 JSON 中包含 `logo` 与 `color` 字段（无值时为空串）。

#### Scenario: 管理端事件列表返回 logo 与 color

- **WHEN** 客户端携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/event/list`
- **THEN** 响应 `list[]` 中每项 MUST 包含 `logo` 与 `color` 字段
- **AND** `logo` 若有值 MUST 为以 `/` 开头的路径（如 `/ai_talk_images/...`），MUST NOT 为含 scheme 的绝对 URL（新数据）

#### Scenario: 历史与内部事件选项返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/history/api/event/options` 或 `GET /device/internal/api/event/options`
- **THEN** 响应 `list[]` MUST 同样包含 `logo` 与 `color`

### Requirement: 事件新增与更新 SHALL 支持 multipart 上传 logo

`POST /device/admin/api/event/add` 与 `POST /device/admin/api/event/update` MUST 接受 `multipart/form-data`，至少包含表单字段 `name`、`needQuantity`、`extraNames`、`color`；`update` MUST 包含 `id`。可选文件字段名 MUST 为 `logo`。

#### Scenario: 新增事件并上传 logo

- **WHEN** 客户端 multipart 提交有效 `name` 与合法图片 `logo`
- **THEN** 服务端 MUST 在 `/ai_talk_images/` 目录（不存在则创建）保存文件
- **AND** MUST 将 `event.logo` 设为 `/ai_talk_images/<安全文件名>` 且不包含域名
- **AND** MUST 将 `color` 写入 `event.color`（若提供）

#### Scenario: 更新事件未传 logo 保留原值

- **WHEN** 客户端对已有事件 multipart 更新且未包含 `logo` 文件
- **THEN** 服务端 MUST 保留原 `event.logo` 路径
- **AND** MAY 更新 `color` 与其它文本字段

### Requirement: 事件 logo 静态读 SHALL 由 device-service 提供

device-service MUST 注册 `GET /ai_talk_images/*`，从配置的 `eventImageStorageDir`（默认 `/ai_talk_images/`）安全读取文件；路径 MUST 拒绝 `..` 与非法字符。

#### Scenario: 按路径读取已上传 logo

- **WHEN** 请求 `GET /ai_talk_images/event_1_abc.png` 且文件存在
- **THEN** 服务端 MUST 返回对应图片 Content-Type

### Requirement: 事件 Redis 缓存投影 SHALL 含 logo 与 color

`ListEvents` 使用的 Redis 事件选项缓存 MUST 与数据库查询一致，包含 `logo` 与 `color`，以便缓存命中时列表仍返回完整字段。

#### Scenario: 缓存命中仍含 logo

- **WHEN** `ListEvents` 从 Redis 命中事件选项
- **THEN** 返回的 `[]Event` MUST 含非省略的 `logo`、`color` 字段
