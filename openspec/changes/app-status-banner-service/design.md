# Design: app-status-banner-service

## 架构

```
Flutter App  --STATUS_BASE_URL-->  app-status-service (:9806)
                                      ├─ GET  /app/api/status/banner
                                      ├─ POST /admin/api/login
                                      ├─ GET/PUT /admin/api/banner
                                      └─ GET  /admin  (静态页)
```

与 gateway-app **解耦**；JWT 密钥与 `GATEWAY_APP_ADMIN_*` 与 Hub 一致，便于运维记忆，token 在本服务校验。

## 状态模型

- `sync.RWMutex` + 单 struct：`active`, `title`, `message`, `expectedEndAt?`, `dismissible`, `updatedAt`。
- 进程启动默认 `active=false`；**重启清空**。
- `contentKey = trim(title) + "\n" + trim(message)`，供 Admin 快照与客户端 dismiss。

## dismissible 语义

| dismissible | App UI | 本地持久化 |
|-------------|--------|------------|
| false | 不可点遮罩关闭；仅「刷新」；inactive 后关窗 | 无 |
| true | 可关闭；可选「不再提示」 | SharedPreferences 存 contentKey |

**active false→true 不清客户端 dismiss**；同文案再次全员通知需改字或清 App 数据（Admin 页提示）。

## 部署

- 单副本；Nginx 独立域名（如 `status.pangbao.cuplay.top`）。
- Docker：`Dockerfile.app-status-service`，Compose 注入 `GATEWAY_APP_JWT_SECRET` / `GATEWAY_APP_ADMIN_*`。
- Hub：`admin-modules.js` 外链至 `APP_STATUS_ADMIN_PUBLIC_URL`（默认示例域名）。

## Flutter

- `STATUS_BASE_URL` dart-define，默认 `https://status.pangbao.cuplay.top`。
- 主页 `addPostFrameCallback` 拉 banner，**优先于**版本弹窗；未登录也执行。
