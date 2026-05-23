## ADDED Requirements

### Requirement: 版本管理页访问控制

gateway-app-server SHALL 提供「版本管理」相关 UI 与 API；在未通过管理员鉴权前，SHALL NOT 暴露 APK 上传与写库能力。

#### Scenario: 口令错误拒绝管理操作

- **WHEN** 客户端在未持有有效管理会话的情况下调用上传或写库接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 写入磁盘或数据库

#### Scenario: 口令校验通过获得会话

- **WHEN** 客户端提交的管理员口令与网关进程配置的口令一致
- **THEN** 系统 SHALL 建立可校验的管理员会话（如 HttpOnly Cookie）并允许后续受保护操作

### Requirement: Android APK 落盘路径

系统 SHALL 将成功接收的 Android APK 保存至 **`/apk/ai_talk/`** 目录下；若该目录（或父路径中本进程可创建的部分）不存在，SHALL 使用等价于 `MkdirAll` 的方式创建后再写入。

#### Scenario: 目录不存在时自动创建

- **WHEN** 管理员已鉴权且上传合法 APK，且 `/apk/ai_talk/` 尚不存在但进程具备在路径上创建目录的权限
- **THEN** 系统 SHALL 创建目录并完成文件保存

### Requirement: 下载 URL 与数据库一致性

上传成功后，系统 SHALL 根据配置的对外 **`publicBaseUrl`**（或等价项）与约定的 **HTTP GET 下载路径规则** 生成 **完整绝对 URL**，并 SHALL 将该 URL 写入 `ai_voice_app.version` 表中对应新记录的 **`download_url`** 字段，且该 URL SHALL 指向已保存的 APK 文件。

#### Scenario: 客户端可下载

- **WHEN** 任意客户端使用版本表中记录的 `download_url` 发起 GET 请求
- **THEN** 系统 SHALL 返回对应 APK 内容（`application/vnd.android.package-archive` 或等价二进制流）且 SHALL NOT 允许访问约定目录之外的文件

### Requirement: 上传与文件名校验

系统 SHALL 仅接受扩展名为 **`.apk`**（大小写不敏感可统一规范）的上传；SHALL 拒绝路径分隔符、空文件名等非法文件名；SHALL 可配置单文件大小上限并在超限时拒绝。

#### Scenario: 非 APK 扩展名拒绝

- **WHEN** 管理员上传文件扩展名不是 `.apk`
- **THEN** 系统 SHALL 拒绝保存且不更新 `download_url`

### Requirement: 与现有版本检查行为兼容

新增写库记录后，`GET /device/app/api/version/check` 所依据的「最新版本行」语义 SHALL 与现有实现一致（按主键或约定排序取最新一条），且返回的 **`downloadUrl`** SHALL 与库中 `download_url` 一致。

#### Scenario: 新插入行成为最新发版

- **WHEN** 管理员通过本功能插入一条包含 `latest_version` 与 `download_url` 的新 `version` 行且其排序上为最新
- **THEN** 客户端调用版本检查接口时 SHALL 收到该行的 `latestVersion` 与 `downloadUrl`（在 semver/比较规则允许的前提下与现有版本检查逻辑一致）
