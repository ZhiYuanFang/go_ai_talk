## Context

- `entity.Event` 已有 `Logo`、`Color` 字段；`ListEvents` / `RebuildEventCache` 仅查四列，列表 API 实际不返回 logo/color。
- 事件 CRUD 归属 **device-service**；App 读事件字典经 history 委派 `device/internal/event/options`。
- APK 由 **gateway-app-server** 上传至 `gatewayApp.apkStorageDir`（默认 `/apk/ai_talk/`），DB `app_version.download_url` 当前为 `publicBaseUrl + /device/app/apk/<file>`。
- 产品约定：对外资源经 **`www.cuplay.top:9702`**（gateway-app）；DB 不绑域名。

## Goals / Non-Goals

**Goals:**

- `event.logo`、`version.download_url`（及同类字段）库内为 **以 `/` 开头的应用内路径**。
- 事件 multipart add/update；logo 文件 + `color` 持久化。
- 所有事件列表类 API JSON 含 `logo`、`color`（path 或空串）。
- gateway-app **9702** 可 GET `/ai_talk_images/*`（反代 device）与既有 `/device/app/apk/*`。
- 上传 APK 不再要求配置 `publicBaseUrl` 才能写库。

**Non-Goals:**

- CDN、图片处理/缩略图、删除事件时清理磁盘旧 logo（可后续迭代）。
- 主网关 **9701** 反代 `/ai_talk_images`（管理页预览用 9702 基址或完整 URL 拼接说明即可）。
- 修改 voice/history 业务逻辑 beyond 事件字典字段透传。

## Decisions

### 1. Path-only 存库规范

| 资源 | DB 示例 | 对外 GET（gateway-app） |
|------|---------|-------------------------|
| 事件 logo | `/ai_talk_images/event_12_abc.png` | `https://<host>:9702/ai_talk_images/event_12_abc.png` |
| APK | `/device/app/apk/1.2.0_123.apk` | `https://<host>:9702/device/app/apk/1.2.0_123.apk` |

- 路径 MUST 以 `/` 开头，禁止存 `http://`/`https://`（新写入）。
- 读取时若检测到历史绝对 URL，MUST 提取 path 部分再返回（兼容旧行）。

### 2. 事件 logo 上传（device-service）

- **add**：multipart → 插入事件行得 `id` → 若有 logo 则保存为 `/ai_talk_images/event_<id>_<nonce>.<ext>` → 更新 `logo` 列。
- **update**：有 `id`；新 logo 则覆盖写新文件并更新 path；未传文件则保留原 `logo`。
- **color**：表单字段，校验建议 `#RGB` 或 `#RRGGBB`（大小写不敏感），可空。
- 文件类型：`.png`、`.jpg`、`.jpeg`、`.webp`；单文件大小上限可配置（默认如 2MB）。
- 存储目录：`device.eventImageStorageDir`，默认 `/ai_talk_images/`，`MkdirAll(0755)`。

### 3. 事件列表与缓存

- `ListEvents`、`RebuildEventCache`、`cache_repo` 投影同步字段增加 `Logo`、`Color`（或去掉 Fields 限制扫描全表相关列）。
- 变更后沿用现有 `RoutingDeviceEventChanged` 刷新缓存。

### 4. 静态与反代

```
App / 浏览器  GET :9702/ai_talk_images/xxx.png
       │
       ▼
gateway-app  反代 ──▶  device-service:9803/ai_talk_images/xxx.png
                              │
                              ▼
                         读 eventImageStorageDir 文件
```

- device-service：`BindHandler("/ai_talk_images/*", ...)` 安全读盘（防 `..`）。
- gateway-app：新增 `/ai_talk_images/*` 反代至 `DEVICE_SERVICE_URL` 同路径（与 admin API 反代目标一致）。
- APK 下载保持 gateway 本机 `apkStorageDir` + `/device/app/apk/*`（不变）。

### 5. APK path-only

- 上传成功写库：`download_url = "/device/app/apk/" + serverName`（与现 `dlPath` 一致，去掉 `base` 前缀）。
- 上传响应 `downloadUrl` 字段改为 path；移除「未配置 publicBaseUrl 则 400」的写库硬依赖（`publicBaseUrl` 可保留供管理页展示可选绝对链，非必须）。
- `VersionCheck` / Redis 缓存中的 `downloadUrl` 返回 path；文档要求 App 拼接 `gatewayApp` 基址。

### 6. API 形态变更

- `DeviceAdminEventAddReq` / `UpdateReq`：由 JSON 改为 **自定义 multipart handler**（或 GoFrame 不支持 file 的 struct 时拆 handler），保留路径与口令校验。
- `DeviceAdminContract.AddEvent` / `UpdateEvent` 扩展参数：`color string`，logo 由 handler 落盘后传 path。

### 7. 管理端 admin.html

- 新增/编辑：`<input type="file" accept="image/*">`、color `<input type="color">` 或 text。
- 提交 `FormData`；列表列可增加缩略图（`imageBase = 配置的 9702 基址` + `it.logo`）。

## Risks / Trade-offs

- **[BREAKING]** 旧 App 若假定 `downloadUrl` 为绝对 URL，需发版拼接逻辑。
- **[Risk]** 仅暴露 9701 的环境，管理预览须用 9702 基址。
- **[Risk]** device 与 gateway 分机部署时反代链或共享卷配置错误 → 图片 404。
- **[Risk]** 旧 `download_url` 全 URL 需归一化或手工迁移。

## Migration Plan

1. 部署 device-service + gateway-app。
2. 可选 SQL：将 `download_url` 中已知 host 前缀 strip 为 path（或依赖运行时归一化）。
3. 发布 App：版本检查与事件 UI 使用 `baseURL + path`。
4. 回滚：恢复写绝对 URL 的旧 gateway 上传逻辑（需同时回滚 App 若已适配 path）。

## Open Questions

- 无（path-only + 9702 访问、`multipart` add/update 已由产品确认）。
