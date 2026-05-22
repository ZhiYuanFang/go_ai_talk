## ADDED Requirements

### Requirement: 版本管理历史列表

gateway-app-server SHALL 向已通过管理员会话鉴权的客户端提供历史发版列表接口，返回 `ai_voice_app.version` 表中的记录，默认按主键 `id` 降序排列。

#### Scenario: 已登录管理员获取列表

- **WHEN** 客户端持有有效版本管理会话并请求列表接口且口令功能已启用
- **THEN** 系统 SHALL 返回 `code=0` 及包含版本记录的列表（含 `id`、`latestVersion`、`releaseDate`、`releaseNotes`、`downloadUrl`、`forceUpdate`、`minVersion`）

#### Scenario: 未登录拒绝列表

- **WHEN** 客户端未持有有效版本管理会话即请求列表接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 返回版本行数据

#### Scenario: 分页参数生效

- **WHEN** 客户端传入合法的 `limit` 与 `offset`
- **THEN** 系统 SHALL 按 `id` 降序返回不超过 `limit` 条记录（`limit` 不得超过约定上限）

### Requirement: 按 id 查询单条版本

系统 SHALL 支持管理员按主键 `id` 查询单条 `version` 记录。

#### Scenario: 存在记录时返回详情

- **WHEN** 已鉴权管理员请求存在的 `id`
- **THEN** 系统 SHALL 返回该行的完整发版字段

#### Scenario: 不存在记录

- **WHEN** 已鉴权管理员请求不存在的 `id`
- **THEN** 系统 SHALL 返回未找到响应且 SHALL NOT 返回伪造数据

### Requirement: 更新版本元数据

系统 SHALL 允许已鉴权管理员更新已有 `version` 行的元数据字段，且 SHALL NOT 通过本接口修改 `download_url`。

#### Scenario: 成功更新可编辑字段

- **WHEN** 已鉴权管理员提交有效 `id` 及至少一个允许字段（如 `latestVersion`、`releaseNotes`、`forceUpdate`、`minVersion`、`releaseDate`）
- **THEN** 系统 SHALL 持久化更新并返回成功响应

#### Scenario: 更新后版本检查缓存失效

- **WHEN** 更新操作成功提交
- **THEN** 系统 SHALL 删除或失效用于 `GET /device/app/api/version/check` 的最新行 Redis 缓存键 `gw:app:version:latest`

#### Scenario: 未鉴权拒绝更新

- **WHEN** 未持有有效管理会话的客户端调用更新接口
- **THEN** 系统 SHALL 拒绝写库

### Requirement: 删除版本记录

系统 SHALL 允许已鉴权管理员按 `id` 删除 `version` 表记录，并在安全条件下清理关联 APK 文件。

#### Scenario: 成功删除数据库行

- **WHEN** 已鉴权管理员删除存在的 `id`
- **THEN** 系统 SHALL 从 `version` 表移除该行并返回成功响应

#### Scenario: 删除后失效最新版本缓存

- **WHEN** 删除操作成功
- **THEN** 系统 SHALL 失效 `gw:app:version:latest` 缓存，使 `version/check` 按剩余行中最大 `id` 重新加载

#### Scenario: 尽力删除磁盘 APK

- **WHEN** 被删行的 `download_url` 为约定 path-only 形式（前缀 `/device/app/apk/`）且文件名为安全 APK 名且在存储目录内存在对应文件
- **THEN** 系统 SHALL 尝试删除该文件；若删除失败 SHALL 仍完成删行并记录可观测警告日志

#### Scenario: 非法路径不删盘外文件

- **WHEN** `download_url` 不含约定前缀或文件名未通过安全校验
- **THEN** 系统 SHALL 仅删除数据库行且 SHALL NOT 删除存储目录外或路径穿越目标

### Requirement: 新增发版与现有上传兼容

「新增（Create）」SHALL 继续通过现有 `POST /device/app/api/version/admin/upload` multipart 接口完成（APK 校验、落盘、插入新行、写入 path-only `download_url`），行为与变更前一致。

#### Scenario: 上传成功仍插入新行

- **WHEN** 已鉴权管理员上传合法 APK 及 `latestVersion`
- **THEN** 系统 SHALL 插入新的 `version` 行并失效最新版本缓存，且 `download_url` SHALL 为可经 `GET /device/app/apk/` 下载的 path-only 值

### Requirement: 与版本检查接口语义一致

增删改及上传写库后，`GET /device/app/api/version/check` 所依据的「最新版本行」SHALL 仍为 `version` 表中 **`id` 最大** 且 `latest_version` 非空的一行；返回的 `downloadUrl` SHALL 经现有 path 归一化后与库中 `download_url` 一致。

#### Scenario: 删除当前最大 id 后检查回落

- **WHEN** 表中仍存在其它发版行且管理员删除了当前 `id` 最大的行
- **THEN** 客户端调用 `version/check` 时 SHALL 使用剩余行中 `id` 最大者作为最新发版

### Requirement: 管理页展示与操作

`gateway-app-version-admin.html`（或等价路由页面）SHALL 在登录成功后展示历史版本列表，并 SHALL 提供触发列表刷新、编辑元数据、删除记录及上传新版本 APK 的交互；所有写操作请求 SHALL 携带同源管理 Cookie。

#### Scenario: 登录后可见历史表格

- **WHEN** 管理员口令校验通过
- **THEN** 页面 SHALL 请求列表接口并展示历史版本行

#### Scenario: 操作后刷新列表

- **WHEN** 上传、更新或删除任一操作成功
- **THEN** 页面 SHALL 刷新列表以反映当前表状态

### Requirement: 版本管理未启用时的错误语义

当网关进程未配置版本管理口令时，受保护的管理接口（含列表、查询、更新、删除、上传）SHALL 与登录接口一致返回服务不可用语义，且 SHALL NOT 暴露写库能力。

#### Scenario: 未配置口令拒绝受保护接口

- **WHEN** `GATEWAY_APP_VERSION_ADMIN_PASSWORD`（及等价配置）为空且客户端调用受保护管理接口
- **THEN** 系统 SHALL 返回与「版本管理未启用」一致的不可用响应
