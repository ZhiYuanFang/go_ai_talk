## MODIFIED Requirements

### Requirement: Web video upload response SHALL include contentHash for register

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 且上传成功时，响应 `data` MUST 含：

- `contentHash`：64 位小写 hex SHA-256，对 **OSS 上最终对象字节** 计算（v1 直传为原始字节；服务端转码为转码后字节）
- `transformVersion`：字符串 `v1` 或 `v2`，指示客户端 `RegisterMedia` 应使用的版本（v1 直传路径为 `v1`；服务端转码路径为 `v2`）

Web 客户端与 Flutter Web（经 gateway 同域 upload）MUST 使用响应中的 `transformVersion` 与 `contentHash` 配对 register；MUST NOT 在转码路径仍 register `v1`。

#### Scenario: Web direct upload returns v1 hint

- **WHEN** Web 成功上传 v1 合规视频（直传路径）
- **THEN** JSON `data` MUST 含 `contentHash`（长度 64）且 `transformVersion` MUST 为 `v1`

#### Scenario: Web transcode upload returns v2 hint

- **WHEN** Web 上传 v1 不合规但服务端转码成功
- **THEN** JSON `data` MUST 含 `contentHash` 且 `transformVersion` MUST 为 `v2`

#### Scenario: Gateway forwards transformVersion unchanged

- **WHEN** 请求经 gateway-app 反向代理至 ucg-service
- **THEN** 响应 JSON MUST 透传 `transformVersion` 字段（无裁剪）
