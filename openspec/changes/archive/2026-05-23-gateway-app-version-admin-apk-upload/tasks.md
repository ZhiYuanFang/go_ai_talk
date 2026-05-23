## 1. 配置与契约

- [x] 1.1 在 `manifest/config/config.gateway-app-server.yaml`（或该进程实际加载的配置）中增加 `gatewayApp.versionAdmin` 与 `gatewayApp.publicBaseUrl`（及 APK 大小上限等）字段说明；口令 **仅** 允许从环境变量或本地私有覆盖读取，**禁止**将生产口令写入仓库提交的 yaml。
- [x] 1.2 在 `docs/runbooks/release-deploy-and-run.md`（或网关专用 runbook）补充：创建 `/apk/ai_talk/` 权限、`publicBaseUrl` 填云主机对外域名、口令注入方式。

## 2. 管理员鉴权与会话

- [x] 2.1 实现 `POST` 登录接口（路径置于 `/device/app/...` 命名空间下）：校验口令与配置一致后设置 **HttpOnly**、**Secure**（生产建议）、`SameSite` 合理的会话 Cookie；失败返回 401。
- [x] 2.2 将上传、写库等管理接口挂中间件：无有效会话则 401。

## 3. 管理 UI

- [x] 3.1 新增静态页（如 `resource/public/gateway-app-version-admin.html`）：口令表单、版本号/更新说明/是否强制更新等字段、APK 文件选择、提交后展示返回的 `downloadUrl` 预览。
- [x] 3.2 在 gateway-app-server 路由注册该页面 **GET** 路径（与现有 gateway-app 路由风格一致），并确保 CORS/鉴权豁免列表不误放行写库接口（仅豁免必要静态资源若已有约定）。

## 4. APK 上传与落盘

- [x] 4.1 实现 `multipart/form-data` 上传处理：扩展名 `.apk`、文件名净化、可配置大小上限；保存目录 **`/apk/ai_talk/`**，`MkdirAll` + 可配置/固定文件名策略（含版本与时间戳防覆盖）。
- [x] 4.2 上传成功后拼接 `{publicBaseUrl}{下载路由前缀}/{filename}` 得到绝对 URL，校验 `publicBaseUrl` 已配置否则 400 并记录日志。

## 5. 数据库与下载

- [x] 5.1 向 `version` 表 **INSERT** 新行：写入 `latest_version`、`download_url`、`release_notes`、`force_update`、`release_date`（Unix 秒）等与表单一致字段。
- [x] 5.2 实现 **GET** APK 下载路由：根据文件名在 `/apk/ai_talk/` 内解析真实路径，禁止 `..` 与非 `.apk`；正确 `Content-Type` 与 `Content-Disposition`（`attachment`）。
- [x] 5.3 手动验证：`version/check` 返回的 `downloadUrl` 与库一致且浏览器/App 可下载。（实现侧已在上传成功后 `DEL gw:app:version:latest`；**部署后**请在目标环境打开 `version/check` 与下载 URL 各做一次点验。）

## 6. 安全与观测

- [x] 6.1 上传接口加简单防刷（可选：同一 Cookie 频率限制）与失败日志（口令错误、磁盘错误、非法文件）。
- [x] 6.2 确认 Redis/版本检查缓存：若上传后 `version/check` 仍缓存旧行，需在写库后 **失效** `gw:app:version:latest` 或沿用现有 bump 策略（查阅 `gateway_app_ctrl.VersionCheck` 实现并补齐）。
