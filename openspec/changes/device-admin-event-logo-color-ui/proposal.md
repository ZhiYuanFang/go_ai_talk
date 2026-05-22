## Why

设备管理页（`admin.html`，主网关 **:9701**）当前线上仍表现为**无 Logo、无色调列**，且历史实现将 logo 预览拼到 **:9702 App 网关**，与管理页不同源，易导致图片不显示。后端 `device-event-logo-and-path-only-assets` 已支持 `logo`/`color` 存库与列表返回，需在管理 UI 上**可见、可同源访问、可点选即改**，与运维实际使用路径一致。

## What Changes

- **事件列表 UI**：在事件管理表格中固定展示 **Logo（缩略图）**、**色调（色块 + 色值）** 列；无 logo 时显示可点击占位。
- **Logo URL 同源**：管理页预览与点击上传使用的图片地址 MUST 为 `当前页面 origin + path`（如 `https://host:9701/ai_talk_images/...`），**不再**依赖 `gatewayAppBase()`（:9702）。
- **主网关静态读**：在 **gateway-service（:9701）** 增加 `GET /ai_talk_images/*` 反代至 device-service（与 gateway-app 行为对齐），使同源 URL 可访问。
- **行内编辑**：点击列表中的**色调色块**可调起取色并提交更新；点击 **Logo 图/占位** 可调起选图并提交更新；未改字段须随 multipart 一并提交以符合现有 `event/update` 契约。
- **静态页缓存**：管理页响应建议 `Cache-Control: no-store`，降低浏览器长期缓存旧版 HTML 的概率。
- 弹窗「编辑 / 新增」保留，用于名称、扩展名、计数等；行内仅负责 logo/color。

## Capabilities

### New Capabilities

- `device-admin-event-logo-color-ui`：管理端事件列表展示 logo/color、同源资源访问、行内点击更新 logo 与色调。

### Modified Capabilities

（无。`openspec/specs/` 下尚无已归档的 device 事件 logo 规格；后端列表字段由变更 `device-event-logo-and-path-only-assets` 覆盖，本变更聚焦网关与管理页。）

## Impact

- **前端**：`resource/public/admin.html`
- **主网关**：`internal/controller/register.go`（或抽取与 gateway-app 共用的 `/ai_talk_images` 反代安装）
- **依赖**：device-service 已提供 `GET /ai_talk_images/*` 与 `event/list` 含 `logo`/`color`；需确认现网 device-service 与主网关均已部署相关版本
- **无破坏性 API 变更**：仍使用现有 `POST /device/admin/api/event/update` multipart
