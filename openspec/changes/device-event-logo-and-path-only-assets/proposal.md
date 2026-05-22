## Why

事件管理需要可配置的 **logo** 与 **color**，且对外资源 URL 应与环境域名/端口解耦：库内只存路径，由 App/管理端按当前网关基址拼接。现有 APK 上传将完整 `download_url`（含 `publicBaseUrl`）写入数据库，换域名或端口后需改库；与事件 logo 应统一为 **path-only** 存储。另，`event` 表已有 `logo`/`color` 列，但列表查询与 Redis 缓存未返回，管理端也无法上传。

## What Changes

### 事件（device-service）

- 事件 **add/update** 改为 `multipart/form-data`：字段 `name`、`needQuantity`、`extraNames`、`color`；可选文件 `logo`。
- logo 落盘目录 **`/ai_talk_images/`**（不存在则创建）；DB `event.logo` 仅存路径，如 `/ai_talk_images/event_<id>_<nonce>.png`，**不存域名**。
- `ListEvents` 与事件 Redis 投影 **包含 `logo`、`color`**；管理端列表、内部 `event/options`、历史 `event/options` 均返回这两字段。
- device-service 提供 **GET `/ai_talk_images/*`** 静态读（源站）。

### 网关静态访问（gateway-app-server :9702）

- 增加 **GET `/ai_talk_images/*`**：反代至 device-service 或等价静态源，供 `https://www.cuplay.top:9702/ai_talk_images/...` 访问（与 App 约定一致）。
- APK：`version.download_url` 改为仅存路径，如 **`/device/app/apk/<filename>.apk`**；上传写库不再依赖 `publicBaseUrl`。
- **BREAKING（契约）**：`GET /device/app/api/version/check` 的 `downloadUrl` 改为返回 **路径**（非绝对 URL）；App 须用配置的网关基址拼接。上传接口 JSON 响应中的 `downloadUrl` 同步为路径。
- 读库时若历史行为绝对 URL，实现可做兼容归一化为路径（迁移期）。

### 文档与 UI

- 更新 `README.MD`：事件 logo/color、path 拼接规则；APK 与事件资源基址均为 gateway-app（如 `:9702`）。
- 管理页 `admin.html`：事件表单支持 logo 上传与 color；列表展示 logo 预览（基址 + path）。

## Capabilities

### New Capabilities

- `device-event-logo-color`：事件 logo 上传、color 配置、path-only logo 存库、列表/缓存返回字段、device 静态读。
- `gateway-app-path-only-assets`：APK download_url path-only；gateway 暴露 `/ai_talk_images`；版本检查与上传响应契约。

### Modified Capabilities

（无 `openspec/specs/` 既有能力名直接覆盖；上列为新增增量规格。）

## Impact

- **device-service**：`admin.go`、`device_admin.go`、事件缓存、`register_device_service.go`、配置项 `eventImageStorageDir`。
- **gateway-app-server**：`gateway_app_version_admin.go`、`gateway_app_ctrl.go`（版本检查）、`gateway_app_register.go`（反代/静态）。
- **gateway-service (:9701)**：可选仅为管理页说明；事件 admin API 仍经现有反代。
- **客户端**：App 版本更新与事件展示须 **baseURL + path**；已存绝对 URL 需兼容或迁移。
- **部署**：`www.cuplay.top:9702` 须能访问 `/ai_talk_images/*` 与 `/device/app/apk/*`；device 与 gateway 需可达落盘目录或反代链。
