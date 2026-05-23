## Context

- 管理入口：`GET /device/admin` → `admin.html`，API 走 `GET/POST /device/admin/api/*`（主网关反代至 device-service）。
- 事件 `logo` 存 path-only（`/ai_talk_images/event_<id>_...`）；`color` 为 `#RRGGBB` 等字符串。
- `POST /device/admin/api/event/update` 为 **multipart**，必填 `id`、`name`、`needQuantity`、`extraNames`、`color`；`logo` 文件可选，不传则保留原 logo。
- 仓库内 `admin.html` 可能已有 Logo/色调列草稿，但现网未生效或 logo 指向 :9702 导致「有列无图」；用户确认需 **同源** 与 **行内点击修改**。

## Goals / Non-Goals

**Goals:**

- 登录后事件表**始终可见** Logo、色调两列及可辨别的预览。
- Logo 请求 URL = `location.origin` + 规范化 path。
- 主网关 :9701 可 GET `/ai_talk_images/*`（反代 device-service）。
- 点击色块 / 图片即可更新，无需打开完整编辑弹窗。
- 更新成功后刷新列表并保留现有事件缓存失效语义（device-service 内已有）。

**Non-Goals:**

- 不改 `event` 表结构；不新增 PATCH 单字段 API。
- 不在本变更改 App 端（:9702）事件展示逻辑。
- 不实现 logo 裁剪、多尺寸、批量导入。

## Decisions

### 1. 同源 logo 基址

```javascript
function eventLogoUrl(logoPath) {
  const p = String(logoPath || '').trim();
  if (!p) return '';
  if (/^https?:\/\//i.test(p)) return p; // 历史绝对 URL 兼容
  const origin = window.location.origin;
  return origin + (p.startsWith('/') ? p : '/' + p);
}
```

**弃用** `gatewayAppBase()` 拼 logo（仅版本管理链接仍可用 :9702）。

### 2. 主网关暴露 `/ai_talk_images`

与 `gateway_app_image_proxy.go` 相同模式：`ReverseProxy` 到 `DEVICE_API_PROXY_URL`（或 `readDomainProxyConfig` 目标）下的 device-service。

```
浏览器 :9701/admin.html
    │  <img src=":9701/ai_talk_images/...">
    ▼
gateway-service  GET /ai_talk_images/*
    │  reverse proxy
    ▼
device-service   GET /ai_talk_images/*  (deviceEventImageServe)
```

**备选**：仅改前端、不反代主网关 — 拒绝，同源 URL 在 :9701 会 404。

实现可抽取 `installEventImageProxy(s, targetURL)` 供 gateway 与 gateway-app 复用，避免双份逻辑（本变更至少保证 gateway `register.go` 安装）。

### 3. 行内编辑交互（方案 A）

每行挂载隐藏控件：

| 点击目标 | 行为 |
|----------|------|
| 色块 / 色值区域 | `input[type=color].click()` → `change` 时 `submitEventUpdate(row, { color })` |
| Logo 图或占位 | `input[type=file].click()` → `change` 时 `submitEventUpdate(row, { logoFile })` |

`submitEventUpdate` 构造 `FormData`：`id`、`name`、`needQuantity`、`extraNames`、`color`（来自行缓存），有文件则 `append('logo', file)`。

行级状态：`data-event-id`、提交中禁用、错误写入 `eventMsg`。

**不**在点击时打开 `eventModal`（与「点选即改」一致）；名称等仍用「编辑」按钮。

### 4. 列表列与占位

| 列 | 内容 |
|----|------|
| Logo | `max-height:40px` 缩略图；无图时虚线框「点击上传」，`cursor:pointer` |
| 色调 | 24×24 色块 + `#RRGGBB` 文本，可点击 |

色值非法时回退展示 `#cccccc` 并允许点击重选。

### 5. 静态页与部署

- `BindHandler("/device/admin")` 增加 `Cache-Control: no-store`（若与 `ServeFile` 同路由）。
- 验收需 **重建 gateway-service**（反代 + 静态）并 **device-service**（列表字段）；浏览器 Ctrl+F5。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 现网 device-service 过旧，list 无 logo/color | 部署说明写明依赖 `device-event-logo-and-path-only-assets` |
| 误触行内上传 | 仅 logo/色块可点；提交前无 confirm（色块可即时提交，考虑 300ms debounce 可选） |
| 主网关反代目标配错 | 复用现有 `DEVICE_API_PROXY_URL`，与 `/device/admin/api` 一致 |
| 历史绝对 URL logo | 前端保留 http(s) 直通 |

## Migration Plan

1. 部署 device-service（若尚未含 logo/color 列表）。
2. 部署 gateway-service（含 `/ai_talk_images` 反代 + 新 `admin.html`）。
3. 运维强刷管理页，验证列表列与点击改色/换图。

## Open Questions

- 无。用户已确认：当前无两列、logo 用同源。
