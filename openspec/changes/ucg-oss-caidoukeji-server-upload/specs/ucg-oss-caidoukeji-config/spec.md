## ADDED Requirements

### Requirement: UCG OSS MUST use caidoukeji bucket on Hangzhou internal endpoint

`ucg-service` 的 OSS 配置 MUST 使用 bucket `caidoukeji`、region `cn-hangzhou`。服务端 SDK endpoint MUST 为 `oss-cn-hangzhou-internal.aliyuncs.com`（不含 bucket 虚拟主机前缀）。AccessKey MUST 继续经 `UCG_OSS_ACCESS_KEY_*` 注入，yaml 明文 MUST 留空。展示 URL MUST 继续基于配置的 `cdnBaseUrl` 与 objectKey 拼接。系统 MUST NOT 再将 `pang-bao` 或 `oss-cn-beijing.aliyuncs.com` 作为默认生产配置。

#### Scenario: 默认配置指向新桶内网

- **WHEN** 使用仓库默认 `config.ucg-service.yaml` 且未设置 endpoint 覆盖 env
- **THEN** `LoadOSSConfig`（或等价）SHALL 得到 bucket=`caidoukeji` 且 endpoint 为杭州内网主机

#### Scenario: 非 VPC 可用 env 覆盖 endpoint

- **WHEN** 开发环境设置了约定的 OSS endpoint 覆盖变量为公网杭州 endpoint
- **THEN** 服务端 OSS 客户端 SHALL 使用该覆盖值连接（便于本地联调）

### Requirement: Server-side OSS operations MUST use configured endpoint only

服务端 `PutObject`、删除、缩略图、转码读写、校验下载等 MUST 使用上述配置的 endpoint/bucket，MUST NOT 硬编码旧桶名。

#### Scenario: 媒体服务端上传写入新桶

- **WHEN** 调用 `POST /ucg/app/api/media/upload`（经网关）成功上传文件
- **THEN** 对象 SHALL 写入 `caidoukeji`，响应 SHALL 含 objectKey 与基于 cdnBaseUrl 的 cdnUrl
