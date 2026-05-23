# gateway-app-path-only-assets Specification

## Purpose
TBD - created by archiving change device-event-logo-and-path-only-assets. Update Purpose after archive.
## Requirements
### Requirement: APK download_url SHALL 仅存应用内路径

gateway-app-server 在 APK 上传写库时，`app_version.download_url` MUST 仅存以 `/` 开头的路径，格式为 `/device/app/apk/<filename>.apk`；MUST NOT 将 `publicBaseUrl` 或任何域名写入该列（新写入）。

#### Scenario: 上传 APK 写库为路径

- **WHEN** 管理员成功上传 APK 并完成数据库 Insert
- **THEN** `download_url` MUST 等于 `/device/app/apk/` 加安全文件名
- **AND** 上传接口 JSON 响应中的 `downloadUrl` MUST 为同一路径（非绝对 URL）

#### Scenario: 上传不再因缺少 publicBaseUrl 拒绝写库

- **WHEN** `gatewayApp.publicBaseUrl` 未配置
- **AND** 上传文件与其它表单字段合法
- **THEN** 服务端 MUST 仍能完成落盘与数据库写入（路径存库）

### Requirement: 版本检查接口 SHALL 返回 path 型 downloadUrl

`GET /device/app/api/version/check` 响应中的 `downloadUrl` MUST 为应用内路径（新数据）；若库内仍为历史绝对 URL，服务端 MUST 在返回前归一化为路径（仅保留 path 部分）。

#### Scenario: 版本检查返回路径供客户端拼接

- **WHEN** 版本表存在最新行且 `download_url` 已按 path 存储
- **THEN** `downloadUrl` MUST 形如 `/device/app/apk/xxx.apk`
- **AND** MUST NOT 以 `http://` 或 `https://` 开头

### Requirement: gateway-app SHALL 代理事件 logo 静态路径

gateway-app-server MUST 注册 `GET /ai_talk_images/*`，将请求转发至 device-service 同路径（或等价静态源），使客户端可通过 gateway-app 端口（如 `:9702`）访问 `https://<host>:9702/ai_talk_images/...` 而无需直连 device 端口。

#### Scenario: 经 gateway-app 访问事件 logo

- **WHEN** 客户端请求 gateway-app 的 `GET /ai_talk_images/event_1.png`
- **AND** device-service 上对应文件存在
- **THEN** gateway-app MUST 返回成功图片响应（经反代或共享存储）

### Requirement: APK 下载路径契约保持不变

既有 `GET /device/app/apk/*filename` 下载处理器 MUST 继续从 `apkStorageDir` 提供文件；与 path-only `download_url` 组合后，客户端完整下载地址为 `<gateway-app-base>` + `downloadUrl`。

#### Scenario: path 与下载路由一致

- **WHEN** `download_url` 为 `/device/app/apk/foo.apk`
- **THEN** 对 gateway-app 发起 `GET /device/app/apk/foo.apk` MUST 可下载该文件

