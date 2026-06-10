## ADDED Requirements

### Requirement: 事件 logo SHALL 存储于 OSS event 前缀

device-service 在管理端上传或迁移脚本写入时，MUST 将 `event.logo` 设为 OSS objectKey，前缀 MUST 为 `event/`，MUST NOT 含 scheme 或域名。

#### Scenario: 管理端上传新 logo

- **WHEN** 客户端 multipart 提交合法图片至 event add/update
- **THEN** device-service MUST 经 HTTP 调用 ucg-service 内部上传接口
- **AND** MUST 将返回的 objectKey 写入 `event.logo`

#### Scenario: 迁移脚本回填

- **WHEN** 迁移脚本成功上传历史 logo
- **THEN** `event.logo` MUST 为 `event/{id}/logo.{ext}` 形式

### Requirement: 对外 API logo 字段 SHALL 返回 CDN 绝对 URL

所有返回事件字典列表的 HTTP 接口（含 admin list、history/internal event options、gateway site home 的 logoUrl）MUST 将 logo 序列化为 `https://{cdnHost}/{objectKey}`，MUST NOT 返回 `/ai_talk_images/` path-only。

#### Scenario: history event options 返回 CDN

- **WHEN** 客户端请求 `GET /device/history/api/event/options`
- **THEN** 每项 `logo` 若有值 MUST 以 `https://` 开头且指向 CDN 域名

### Requirement: ucg-service SHALL 提供内部 OSS 上传供 device 调用

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload`，接受 multipart 文件，鉴权 MUST 使用网关内部密钥；响应 MUST 含 `objectKey` 与 `cdnUrl`。

#### Scenario: device 内部上传成功

- **WHEN** device-service 携带有效内部密钥上传 png
- **THEN** ucg-service MUST 写入 OSS 并返回 objectKey 与 cdnUrl

### Requirement: 迁移完成后 SHALL 下线 ai_talk_images 静态链路

系统 MUST NOT 再注册 `GET /ai_talk_images/*` 静态读或网关反代；Docker compose MUST NOT 再挂载宿主机 `/ai_talk_images` 目录。

#### Scenario: 旧静态 URL 不可用

- **WHEN** 迁移与部署完成后请求 `GET /ai_talk_images/event_1.png`
- **THEN** 网关或 device-service MUST NOT 再提供该静态文件
