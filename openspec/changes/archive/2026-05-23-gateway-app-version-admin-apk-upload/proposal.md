## Why

App 发版依赖 `ai_voice_app.version` 表中的 `download_url` 等字段，目前缺少在 **gateway-app-server** 侧一站式维护版本与上传 Android APK 的入口，运维需手工改库、手工托管安装包，易出错且下载地址与网关对外域名不一致。需要在 App 网关提供受控的管理页：保存 APK 到约定目录、生成可对外访问的下载 URL 并写回数据表。

## What Changes

- 在 **gateway-app-server** 增加「版本管理」静态页面（或同域下的管理 UI），访问前须通过 **管理员口令** 校验（口令由网关进程配置注入，**不得**在代码中写死生产口令；部署文档说明如何设置）。
- 管理页支持上传 **Android APK**，服务端将文件落盘至 **`/apk/ai_talk/`**（若目录不存在则自动创建），文件名策略在 `design.md` 中约定（避免覆盖不可追溯）。
- 上传成功后，根据网关对外可访问的 **Base URL（配置项）** 拼接 APK 的 **HTTP(S) 下载路径**，将完整 URL 写入 **`version` 表**对应记录的 `download_url`（并与现有版本检查接口返回的 `downloadUrl` 对齐）；可同时维护 `latest_version`、`release_notes`、`force_update` 等字段（最小闭环以 `download_url` + 版本号为主）。
- 提供 APK 的 **匿名 GET 下载**（或带短期 token，见设计权衡），以便客户端版本检查拿到的 URL 可直接下载；须限制路径 traversal、仅允许 `.apk`、可选大小上限。

## Capabilities

### New Capabilities

- `gateway-app-version-admin`：App 网关上的版本与 APK 管理（管理员鉴权、文件存储、下载 URL 生成、`version` 表更新）。

### Modified Capabilities

- （无）本变更为网关侧新能力；若未来将 `openspec/specs/` 下已有能力与发版流程绑定，再另起 delta。

## Impact

- **进程**：`gateway-app-server`（路由、静态页、上传处理、文件系统、可选静态文件对外映射）。
- **配置**：管理员口令、APK 根目录、对外下载 Base URL（云服务器公网域名或反代前缀）。
- **数据**：`ai_voice_app` 库 `version` 表（`dao.AppVersion` 已有实体）。
- **安全**：口令校验、上传 MIME/扩展名校验、磁盘配额与速率限制（设计阶段细化）。
