# Proposal: app-status-banner-service

## 背景

gateway-app 维护或 MySQL/Redis 故障时，App 无法通过现有 gateway 接口获知「请弹窗告知用户」。需要独立、极轻、无持久化依赖的 status 微服务，在 gateway 不可用时仍能下发维护/公告通知。

## 目标

- 新增 **app-status-service**（`:9806`），进程内内存 banner，重启即清空（默认无通知）。
- 公开 `GET /app/api/status/banner`（无鉴权，`Cache-Control: public, max-age=30`）。
- 同进程 Admin：`POST /admin/api/login`、`GET/PUT /admin/api/banner`（Hub 同账号 Admin JWT）。
- 静态运维页 `/admin`（`app-status-admin.html`），gateway 维护期间可独立使用。
- Flutter 经 `STATUS_BASE_URL` 拉取 banner，主页 postFrame 展示；可取消模式支持「不再提示」（contentKey = title+\n+message）。

## 非目标

- 不经 gateway-app 反代；不计入 usage 统计。
- 无 MySQL、Redis、RabbitMQ、后台 ticker。
- 多副本一致性（设计为单副本）。

## 影响范围

- **go_ai_talk**：`cmd/app-status-service`、Docker/Compose、Admin Hub 外链。
- **flutter_ai_talk**：`AppEnv.statusBaseUrl`、主页弹窗与 dismiss 存储。
