## ADDED Requirements

### Requirement: 最新行为 OSS 时 downloadUrl SHALL 为 CDN 绝对 URL

当 `GET /device/app/api/version/check` 所选取的最新版本行的 `download_url` 为 OSS objectKey（前缀 `apk/`）或已是 `http(s)` CDN URL 时，响应中的 `downloadUrl` MUST 为客户端可直接使用的 CDN 绝对 URL；MUST NOT 要求客户端再拼接 gateway-app 基址。

#### Scenario: OSS 最新行返回可直用 CDN

- **WHEN** 客户端请求 `version/check` 且最新行 `download_url` 为 `apk/` objectKey
- **THEN** 响应 `downloadUrl` MUST 以 `https://` 开头
- **AND** 客户端无需拼接 gateway 基址即可发起下载
