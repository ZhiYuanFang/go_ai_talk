## MODIFIED Requirements

### Requirement: APK download_url SHALL 仅存应用内路径

gateway-app-server 在 **历史兼容** 场景下仍可识别以 `/` 开头的 path-only `download_url`（格式 `/device/app/apk/<filename>.apk`）。**新上传写库**时，`app_version.download_url` MUST 存 OSS objectKey（前缀 `apk/`，无 scheme/域名）；MUST NOT 将 `publicBaseUrl` 或 CDN 域名写入该列（新写入）。上传接口与对外读接口的 JSON `downloadUrl` 对 OSS 行 MUST 为 CDN 绝对 URL，对 path-only 历史行 MUST 保持可下载语义（check 可为 path；site 拼绝对 URL）。

#### Scenario: 上传 APK 写库为 objectKey

- **WHEN** 管理员成功上传 APK 并完成数据库 Insert
- **THEN** `download_url` MUST 为以 `apk/` 开头的 objectKey（无 `http://`/`https://`）
- **AND** 上传接口 JSON 响应中的 `downloadUrl` MUST 为 CDN 绝对 URL

#### Scenario: 上传不再因缺少 publicBaseUrl 拒绝写库

- **WHEN** `gatewayApp.publicBaseUrl` 未配置
- **AND** 上传文件与其它表单字段合法且 OSS 上传成功
- **THEN** 服务端 MUST 仍能完成写库（objectKey 存库）

### Requirement: 版本检查接口 SHALL 返回 path 型 downloadUrl

`GET /device/app/api/version/check` 响应中的 `downloadUrl`：当最新行为 **OSS objectKey**（或库内已是 CDN 绝对 URL）时，MUST 返回 **CDN 绝对 URL**（以 `https://` 开头），客户端 MUST 可直接使用而不再拼接 gateway 基址；当最新行为 **历史 path-only**（`/device/app/apk/...`）时，MUST 仍返回应用内路径（供客户端拼接网关基址），与变更前兼容。MUST NOT 对 CDN/OSS 值使用「剥 host 只留 path」后再当作应用内 APK 路径返回。

#### Scenario: 版本检查对 OSS 行返回 CDN

- **WHEN** 版本表存在最新行且 `download_url` 为 `apk/` objectKey
- **THEN** `downloadUrl` MUST 以 `https://` 开头且指向 CDN
- **AND** MUST NOT 形如仅 `/device/app/apk/xxx.apk` 的 path-only

#### Scenario: 版本检查对历史 path 行仍返回路径

- **WHEN** 版本表最新行 `download_url` 为 `/device/app/apk/xxx.apk`
- **THEN** `downloadUrl` MUST 形如 `/device/app/apk/xxx.apk`
- **AND** MUST NOT 以 `http://` 或 `https://` 开头

### Requirement: APK 下载路径契约保持不变

既有 `GET /device/app/apk/*filename` 下载处理器 MUST 继续从 `apkStorageDir` 提供文件，供 **历史 path-only** `download_url` 与邀请群二维码等本地资源使用。新上传 APK 的权威下载 MUST 走 CDN，不依赖本路由作为唯一分发通道。

#### Scenario: path 与下载路由一致

- **WHEN** `download_url` 为 `/device/app/apk/foo.apk`
- **THEN** 对 gateway-app 发起 `GET /device/app/apk/foo.apk` MUST 可下载该文件
