## Why

当前 App 版本管理页仅支持口令登录后**单次上传 APK 并插入一行**，无法查看 `ai_voice_app.version` 表中的历史发版记录，也无法修正元数据或清理误发版本。运维在发版纠错、审计与回滚场景下需要完整的**列表展示与增删改查**，且须与现有 `GET /device/app/api/version/check`（按最大 `id` 取最新行）语义兼容。

## What Changes

- 在 **gateway-app-server** 增加受管理员会话保护的版本管理 API：**列表、按 id 查询、更新元数据、删除记录**；保留并复用现有 **登录 + multipart 上传** 作为「新增发版（Create）」路径。
- 扩展 `gateway-app-version-admin.html`：登录后展示历史版本表格（版本号、上线时间、下载路径、强制更新等），支持编辑、删除与继续上传新版本。
- 任意写库或删行后 **失效** `gw:app:version:latest` Redis 缓存，保证 `version/check` 立即反映当前最大 `id` 行。
- 删除版本行时：若 `download_url` 指向本进程可管理的 APK 路径且磁盘文件存在，**尽力删除**对应 APK 文件（失败仅记日志，不阻塞删行）；路径校验与现有下载接口一致，禁止目录穿越。
- **不修改**客户端 `version/check` 的对外 JSON 契约与 semver 比较规则；**不引入**跨服务或跨库访问。

## Capabilities

### New Capabilities

- `gateway-app-version-admin-crud`：版本管理页与管理员 API 的历史列表、查询、更新、删除，及与上传/版本检查缓存的一致性要求。

### Modified Capabilities

（无。`openspec/specs/` 下尚无已归档的 gateway-app 版本管理规格；本变更以新能力规格为准。）

## Impact

- **进程**：`gateway-app-server`（`internal/controller/gateway_app_version_admin.go`、`gateway_app_register.go`、`gateway_app_auth_exempt.go`）。
- **数据**：`ai_voice_app.version`（`dao.AppVersion`，连接组 `app`）。
- **前端**：`resource/public/gateway-app-version-admin.html`；设备管理入口 `admin.html` 链接不变。
- **缓存**：`gatewayapp.InvalidateAppVersionLatestCache` 在 update/delete/upload 后调用。
- **配置**：继续依赖 `GATEWAY_APP_VERSION_ADMIN_PASSWORD` / `gatewayApp.versionAdmin.*`；无需新环境变量（可选后续在 compose 示例中补充口令说明，非本变更必需）。
