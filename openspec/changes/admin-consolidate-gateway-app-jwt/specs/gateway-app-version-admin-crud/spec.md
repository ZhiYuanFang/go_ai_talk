## MODIFIED Requirements

### Requirement: 版本管理历史列表

gateway-app-server SHALL 向已通过 **Admin JWT** 鉴权的客户端提供历史发版列表接口，返回 `ai_voice_app.version` 表中的记录，默认按主键 `id` 降序排列。

#### Scenario: 已登录管理员获取列表

- **WHEN** 客户端携带有效 Admin JWT 并请求列表接口且 Hub admin 登录功能已启用
- **THEN** 系统 SHALL 返回 `code=0` 及包含版本记录的列表（含 `id`、`latestVersion`、`releaseDate`、`releaseNotes`、`downloadUrl`、`forceUpdate`、`minVersion`）

#### Scenario: 未登录拒绝列表

- **WHEN** 客户端未携带有效 Admin JWT 即请求列表接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 返回版本行数据

#### Scenario: 分页参数生效

- **WHEN** 客户端传入合法的 `limit` 与 `offset`
- **THEN** 系统 SHALL 按 `id` 降序返回不超过 `limit` 条记录（`limit` 不得超过约定上限）

### Requirement: 管理页展示与操作

`gateway-app-version-admin.html`（或等价路由页面）SHALL 在持有有效 Admin JWT 时展示历史版本列表，并 SHALL 提供触发列表刷新、编辑元数据、删除记录及上传新版本 APK 的交互；所有写操作请求 SHALL 携带 `Authorization: Bearer` Admin JWT，MUST NOT 依赖版本管理独立 Cookie。

#### Scenario: 登录后可见历史表格

- **WHEN** 管理员已在 Hub 取得 Admin JWT 并打开版本管理页
- **THEN** 页面 SHALL 请求列表接口并展示历史版本行

#### Scenario: 操作后刷新列表

- **WHEN** 上传、更新或删除任一操作成功
- **THEN** 页面 SHALL 刷新列表以反映当前表状态

### Requirement: 按 id 查询单条版本

系统 SHALL 支持已携带有效 Admin JWT 的管理员按主键 `id` 查询单条 `version` 记录。

#### Scenario: 存在记录时返回详情

- **WHEN** 客户端携带有效 Admin JWT 请求存在的 `id`
- **THEN** 系统 SHALL 返回该行的完整发版字段

#### Scenario: 不存在记录

- **WHEN** 客户端携带有效 Admin JWT 请求不存在的 `id`
- **THEN** 系统 SHALL 返回未找到响应且 SHALL NOT 返回伪造数据

### Requirement: 新增发版与现有上传兼容

「新增（Create）」SHALL 继续通过现有 `POST /device/app/api/version/admin/upload` multipart 接口完成（APK 校验、落盘、插入新行、写入 path-only `download_url`），行为与变更前一致；调用 MUST 携带有效 Admin JWT。

#### Scenario: 上传成功仍插入新行

- **WHEN** 客户端携带有效 Admin JWT 上传合法 APK 及 `latestVersion`
- **THEN** 系统 SHALL 插入新的 `version` 行并失效最新版本缓存，且 `download_url` SHALL 为可经 `GET /device/app/apk/` 下载的 path-only 值

### Requirement: 版本管理未启用时的错误语义

当网关进程未配置 `GATEWAY_APP_ADMIN_PASSWORD`（Hub 登录不可用）时，受保护的管理接口（含列表、查询、更新、删除、上传）SHALL 与 Hub login 一致返回服务不可用语义，且 SHALL NOT 暴露写库能力。

#### Scenario: 未配置 admin 密码拒绝受保护接口

- **WHEN** `GATEWAY_APP_ADMIN_PASSWORD` 为空且客户端调用受保护版本管理接口
- **THEN** 系统 SHALL 返回与「管理未启用」一致的不可用响应
