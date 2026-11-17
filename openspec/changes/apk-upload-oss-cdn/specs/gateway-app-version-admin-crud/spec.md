## MODIFIED Requirements

### Requirement: 删除版本记录

系统 SHALL 允许已鉴权管理员按 `id` 删除版本表记录，并在安全条件下清理关联 APK：**path-only** 行尽力删除 `apkStorageDir` 内文件；**`apk/` objectKey** 行尽力经 ucg 删除 OSS 对象。清理失败 MUST NOT 阻止删行成功，MUST 记录可观测警告。

#### Scenario: 成功删除数据库行

- **WHEN** 已鉴权管理员删除存在的 `id`
- **THEN** 系统 SHALL 从版本表移除该行并返回成功响应

#### Scenario: 删除后失效最新版本缓存

- **WHEN** 删除操作成功
- **THEN** 系统 SHALL 失效用于 `version/check` 的最新行缓存，使检查按剩余行中最大 `id` 重新加载

#### Scenario: 尽力删除磁盘 APK

- **WHEN** 被删行的 `download_url` 为约定 path-only 形式（前缀 `/device/app/apk/`）且文件名为安全 APK 名且在存储目录内存在对应文件
- **THEN** 系统 SHALL 尝试删除该文件；若删除失败 SHALL 仍完成删行并记录可观测警告日志

#### Scenario: 尽力删除 OSS APK

- **WHEN** 被删行的 `download_url` 为 `apk/` 前缀 objectKey
- **THEN** 系统 SHALL 尝试经 ucg 删除该 OSS 对象；若删除失败 SHALL 仍完成删行并记录可观测警告日志

#### Scenario: 非法路径不删盘外文件

- **WHEN** `download_url` 不含约定 path 前缀且不是安全的 `apk/` objectKey，或文件名未通过安全校验
- **THEN** 系统 SHALL 仅删除数据库行且 SHALL NOT 删除存储目录外或路径穿越目标

### Requirement: 新增发版与现有上传兼容

「新增（Create）」SHALL 继续通过现有 `POST /device/app/api/version/admin/upload` multipart 接口完成（APK 校验、经 ucg 上传 OSS、插入新行、写入 objectKey `download_url`）；调用 MUST 携带有效 Admin JWT。

#### Scenario: 上传成功仍插入新行

- **WHEN** 客户端携带有效 Admin JWT 上传合法 APK 及 `latestVersion`
- **THEN** 系统 SHALL 插入新的版本行并失效最新版本缓存
- **AND** `download_url` SHALL 为 `apk/` objectKey
- **AND** 响应 `downloadUrl` SHALL 为 CDN 绝对 URL

### Requirement: 与版本检查接口语义一致

增删改及上传写库后，`GET /device/app/api/version/check` 所依据的「最新版本行」SHALL 仍为版本表中 **`id` 最大** 且 `latest_version` 非空的一行；返回的 `downloadUrl` SHALL 按该行存储形态映射（OSS→CDN 绝对 URL；历史 path→path-only）。

#### Scenario: 删除当前最大 id 后检查回落

- **WHEN** 表中仍存在其它发版行且管理员删除了当前 `id` 最大的行
- **THEN** 客户端调用 `version/check` 时 SHALL 使用剩余行中 `id` 最大者作为最新发版
