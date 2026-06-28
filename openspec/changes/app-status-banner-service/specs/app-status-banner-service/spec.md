# app-status-banner-service Specification Delta

## ADDED Requirements

### Requirement: app-status-service SHALL expose public banner API without auth

`app-status-service` MUST 提供 `GET /app/api/status/banner`，**无鉴权**。响应 MUST 使用 `{ code, message, data }` 信封。`data.active=false` 时 MUST 仅返回 `{ "active": false }`。`data.active=true` 时 MUST 含 `title`、`message`、`dismissible`、`updatedAt`，MAY 含 `expectedEndAt`（unix 秒）与 `contentKey`。响应 MUST 设置 `Cache-Control: public, max-age=30`。

#### Scenario: inactive banner

- **WHEN** 进程内 `active=false`
- **THEN** `GET /app/api/status/banner` 返回 `data.active=false` 且无 title/message

#### Scenario: active maintenance banner

- **WHEN** 运维 PUT `active=true` 且 `dismissible=false`
- **THEN** App 可读全字段且 `dismissible` 为 false

### Requirement: app-status-service SHALL provide admin banner CRUD with Hub credentials

同进程 MUST 提供 `POST /admin/api/login`（`GATEWAY_APP_ADMIN_USERNAME/PASSWORD`）、`GET /admin/api/banner`、`PUT /admin/api/banner`（Admin JWT）。状态 MUST 存于进程内存；重启 MUST 恢复默认 inactive。MUST NOT 使用 MySQL/Redis。

#### Scenario: admin saves banner

- **WHEN** 管理员 Bearer 有效并 PUT 新 title/message
- **THEN** 随后 GET public banner 立即反映新文案

### Requirement: app-status-service SHALL serve standalone admin page

MUST 托管 `resource/public/app-status-admin.html` 于 `GET /admin`，含登录、启用、文案、expectedEndAt、dismissible、保存、立即关闭、预览与 contentKey 快照提示。

#### Scenario: gateway down

- **WHEN** gateway-app 不可用但 status 服务存活
- **THEN** 运维仍可通过 status 域名打开 `/admin` 并修改 banner
