## ADDED Requirements

### Requirement: 新上传 APK SHALL 经 ucg 写入 OSS apk 前缀

gateway-app-server 在管理员通过 `POST /device/app/api/version/admin/upload` 上传合法 `.apk` 时，MUST 经 `internal/clients/ucg` 调用 ucg-service 内部上传接口，由 ucg 使用内网 OSS endpoint 将对象写入前缀为 `apk/` 的 objectKey；MUST NOT 将新包仅落盘到 `apkStorageDir` 作为权威存储；MUST NOT 从 gateway-app import `hello/internal/services/ucg`。

#### Scenario: 上传成功得到 objectKey

- **WHEN** 已鉴权管理员上传合法 APK 且 ucg OSS 可用
- **THEN** ucg MUST 返回非空 `objectKey`（以 `apk/` 开头）与对应 `cdnUrl`
- **AND** gateway-app MUST 将 `objectKey` 写入新版本行的 `download_url`

#### Scenario: OSS 上传失败不写版本行

- **WHEN** ucg 内部上传失败或超时
- **THEN** gateway-app MUST NOT 插入新的版本行
- **AND** MUST 向管理员返回明确失败语义

### Requirement: ucg 内部 APK 上传 SHALL 支持大文件流式写入

ucg-service MUST 提供独立于事件/功能 logo（2MB）与社交媒体（25MB）上限的 APK 上传能力；单文件上限 MUST 可配置且默认与 gateway-app `apkMaxBytes`（约 200MB）对齐；实现 MUST 流式 PutObject（或等价），MUST NOT 因整文件读入内存导致必然失败；鉴权 MUST 使用与现有 internal media 相同的内部密钥头校验。

#### Scenario: 超过 APK 上限拒绝

- **WHEN** 上传体超过 APK 配置上限
- **THEN** ucg MUST 拒绝写入 OSS 并返回 4xx 语义

#### Scenario: 非 apk 扩展名拒绝

- **WHEN** 内部上传声明的扩展名或文件名不是 `.apk`
- **THEN** ucg MUST 拒绝 PutObject

### Requirement: HTTP 边界 downloadUrl SHALL 映射为 CDN 绝对 URL

凡版本相关 HTTP 响应中面向客户端的 `downloadUrl`（含 `version/check`、官网聚合 Android 下载、管理列表/上传成功响应），当库内 `download_url` 为 `apk/` objectKey（或已是 `http(s)` CDN URL）时，MUST 返回 CDN 绝对 URL（`https://` + 配置的 CDN 基址 + objectKey）；MUST NOT 对这类值再经「剥 host 只留 path」的归一化后当作应用内路径返回。

#### Scenario: version/check 返回 CDN

- **WHEN** 最新版本行的 `download_url` 为 `apk/` 前缀 objectKey
- **THEN** `GET /device/app/api/version/check` 的 `downloadUrl` MUST 以 `https://` 开头且指向 CDN 域名

#### Scenario: 上传响应返回 CDN

- **WHEN** 管理员上传成功
- **THEN** 上传接口 JSON 中的 `downloadUrl` MUST 为 CDN 绝对 URL（可与写库 objectKey 不同形态）

### Requirement: 历史 path-only 行 SHALL 继续本地下载

当 `download_url` 为 `/device/app/apk/<filename>.apk` 时，系统 MUST 继续通过既有 `GET /device/app/apk/*` 从 `apkStorageDir` 提供文件；官网聚合 MUST 仍可用 `publicBaseUrl`（或请求 Host）拼出绝对下载地址。本变更 MUST NOT 要求批量迁移历史磁盘 APK。

#### Scenario: 未发新版时官网仍可下载旧包

- **WHEN** 版本表最新行仍为 path-only `download_url` 且对应文件存在于 `apkStorageDir`
- **THEN** 官网 Android 下载入口与二维码 MUST 仍指向可成功下载的地址（经 gateway-app 本地路径）

### Requirement: 删除 OSS 版本行 SHALL 尽力删除对象

删除 `download_url` 为 `apk/` objectKey 的版本行时，gateway-app MUST 经 ucg 内部能力尽力删除对应 OSS 对象；删除对象失败时 MUST 仍完成数据库删行并记录可观测警告。path-only 行 MUST 继续尽力删除本地文件。邀请群二维码文件 `er_code.png` MUST 不受版本删除逻辑误删。

#### Scenario: 删除 OSS 行尽力删对象

- **WHEN** 管理员删除 `download_url` 为 `apk/...` 的版本行
- **THEN** 系统 MUST 尝试删除该 OSS 对象
- **AND** 无论 OSS 删除是否成功，数据库行 MUST 已删除
