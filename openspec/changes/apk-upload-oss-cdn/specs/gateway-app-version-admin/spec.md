## MODIFIED Requirements

### Requirement: Android APK 落盘路径

系统对 **新上传** 的 Android APK MUST 经 ucg 写入 OSS（objectKey 前缀 `apk/`），MUST NOT 再要求将新包保存至 `/apk/ai_talk/` 作为权威存储。`apkStorageDir`（默认 `/apk/ai_talk/`）MUST 仍可用于服务历史本地 APK 与同目录其它本地资源（如邀请群二维码）；若目录不存在且仍需写入本地资源，MAY 继续 `MkdirAll`。

#### Scenario: 新上传权威存储为 OSS

- **WHEN** 管理员已鉴权且上传合法 APK 且 OSS 可用
- **THEN** 系统 MUST 在 OSS 中存在对应 `apk/` 对象
- **AND** MUST NOT 将「仅本地落盘成功」视为上传成功的充分条件

### Requirement: 下载 URL 与数据库一致性

上传成功后，系统 MUST 将 OSS **objectKey** 写入 `ai_voice_app` 版本表（或 `app_version` 等价表）对应新记录的 **`download_url`** 字段；对外返回的下载链接 MUST 为指向该对象的 **CDN 绝对 URL**。客户端使用该 CDN URL 发起 GET 时 SHALL 能获取 APK 内容（由 CDN/OSS 提供），路径穿越与本地目录外访问防护仍适用于本地 `GET /device/app/apk/*` 处理器。

#### Scenario: 客户端可经 CDN 下载

- **WHEN** 任意客户端使用版本检查或上传响应中的 CDN `downloadUrl` 发起 GET 请求
- **THEN** 响应 SHALL 返回对应 APK 内容（`application/vnd.android.package-archive` 或等价二进制流）

### Requirement: 与现有版本检查行为兼容

新增写库记录后，`GET /device/app/api/version/check` 所依据的「最新版本行」语义 SHALL 与现有实现一致（按主键或约定排序取最新一条）。当该行为 OSS 行时，返回的 **`downloadUrl`** MUST 为 CDN 绝对 URL（由 objectKey 映射），MUST NOT 要求与库内 objectKey 字符串字面相等。

#### Scenario: 新插入 OSS 行成为最新发版

- **WHEN** 管理员通过本功能插入一条包含 `latest_version` 与 OSS `download_url`（objectKey）的新版本行且其排序上为最新
- **THEN** 客户端调用版本检查接口时 SHALL 收到该行的 `latestVersion` 与 CDN 形式的 `downloadUrl`
