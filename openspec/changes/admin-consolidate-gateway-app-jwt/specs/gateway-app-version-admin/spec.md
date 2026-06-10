## MODIFIED Requirements

### Requirement: 版本管理页访问控制

gateway-app-server SHALL 提供「版本管理」相关 UI 与 API；在未通过 **Admin JWT**（`aud=gateway-admin`，由 Hub `POST /device/admin/api/login` 签发）鉴权前，SHALL NOT 暴露 APK 上传与写库能力。系统 MUST NOT 再使用独立版本管理口令、`POST /device/app/api/version/admin/login` 或 `gw_ver_admin` Cookie 会话作为鉴权手段（B1）。

#### Scenario: 无 Admin JWT 拒绝管理操作

- **WHEN** 客户端在未携带有效 Admin JWT 的情况下调用上传或写库接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 写入磁盘或数据库

#### Scenario: Admin JWT 校验通过允许操作

- **WHEN** 客户端携带由 Hub 登录签发的有效 Admin JWT
- **THEN** 系统 SHALL 允许后续受保护操作（上传、写库、列表等）

## REMOVED Requirements

### Requirement: 口令校验通过获得会话

**Reason**: B1 并入统一 Admin JWT；版本管理不再使用独立口令 + HttpOnly Cookie 会话。

**Migration**: 运维在 `/device/admin` Hub 使用 `GATEWAY_APP_ADMIN_USERNAME` / `GATEWAY_APP_ADMIN_PASSWORD` 登录一次，再访问版本管理页；废弃 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`。
