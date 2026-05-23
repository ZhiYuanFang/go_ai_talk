## Context

- **现状**：`gateway-app-server` 在 `ai_voice_app.version` 表维护发版行；管理员通过 `gateway-app-version-admin.html` 登录后调用 `POST .../admin/upload` 插入新行，`download_url` 存 **path-only**（如 `/device/app/apk/<file>.apk`）。`GET /device/app/api/version/check` 取 **`id` 最大** 的一行并缓存于 Redis `gw:app:version:latest`。
- **约束**：版本域数据仅 gateway-app 的 `app` 连接组访问；管理 API 与上传共用 Redis 会话 Cookie（`VersionAdminSessionCookieName`，路径 `/device/app`）；鉴权白名单在 `gateway_app_auth_exempt.go` 登记。

## Goals / Non-Goals

**Goals:**

- 管理员登录后可 **列出** 全部历史版本（默认按 `id` 降序），可 **查看单条**、**更新** 可编辑字段、**删除** 记录；**新增** 继续走现有 multipart 上传（视为 Create）。
- 写库/删行后 **失效** 最新版本缓存，与 `version/check` 行为一致。
- 管理页 UI 与 API 错误语义与现有登录/上传风格一致（`code`/`message` JSON）。

**Non-Goals:**

- 不改变 App 客户端 `version/check` 请求/响应字段或比较算法。
- 不支持批量导入、版本号自动递增策略、灰度分渠道发版。
- 不在本变更中替换 APK 文件（更新仅改元数据；换包仍通过「上传新版本」插入新行）。
- 不将 `gateway-app-version-admin` 能力归档合并进 `openspec/specs/`（留待 apply 完成后 archive）。

## Decisions

### 1. API 形态（REST 风格，JSON）

| 操作 | 方法 | 路径 | 说明 |
|------|------|------|------|
| List | GET | `/device/app/api/version/admin/list` | 查询参数 `limit`（默认 50，上限 200）、`offset`（默认 0） |
| Get | GET | `/device/app/api/version/admin/get?id=` | 单条，不存在返回 404 |
| Update | PUT 或 POST | `/device/app/api/version/admin/update` | JSON body：`id` + 可改字段 |
| Delete | POST 或 DELETE | `/device/app/api/version/admin/delete` | JSON/query：`id` |
| Create | POST | `/device/app/api/version/admin/upload` | **已有**，不变 |

鉴权：与 `upload` 相同，校验管理会话 Cookie + Redis；未登录 401。口令未配置时 list/get 亦返回 503（与 login 一致），避免泄露表是否为空。

**备选**：GraphQL / 单一路由 `action=` — 拒绝，与现有 handler 风格不一致。

### 2. 可更新字段

允许更新：`latestVersion`、`releaseNotes`、`forceUpdate`（0/1）、`minVersion`、`releaseDate`（Unix 秒）。**不允许**通过 update 修改 `download_url`（防指向任意路径）；换 APK 必须新上传一行。

上传（Create）继续写入：`latestVersion`、`releaseNotes`、`forceUpdate`、`download_url`（由服务端生成 path）、`release_date`（默认 `time.Now().Unix()`）。

### 3. 「最新发版」语义

保持 `version/check`：**`ORDER BY id DESC LIMIT 1`**。删除当前最大 `id` 行后，检查接口自动落到次新行；每次 insert/update/delete 调用 `InvalidateAppVersionLatestCache`。

**备选**：按 `release_date` 取最新 — 拒绝，避免与现网逻辑不一致。

### 4. 删除与 APK 文件

删行前读取 `download_url`，若匹配前缀 `/device/app/apk/` 且文件名为 `apkFilenameSafe`，则在 `ApkStorageDir` 下 `os.Remove`；失败只打 Warning，仍删 DB 行。

**风险**：误删仍被其它行引用的文件（若人工改库重复 path）— 接受；正常每行独立文件名。

### 5. 列表响应与对外展示

列表项返回实体字段（camelCase JSON），`downloadUrl` 为库中 path；管理页展示完整 path，可选拼接 `location.origin` 作为可点击下载提示（与设备管理页 gateway 基址一致时用户自行访问 9702）。

### 6. 实现分层

- 控制器：`gateway_app_version_admin.go` 增加 list/get/update/delete handler；抽取 `requireVersionAdminSession(r)` 减少重复。
- 路由：`gateway_app_register.go` 绑定；`gateway_app_auth_exempt.go` 将新路径加入豁免列表（仅 login 免鉴权，list 等需会话 — **注意**：现网 upload 在 exempt 列表是因网关中间件跳过 JWT，仍靠 handler 内会话校验）。
- 前端：单页表格 + 行内编辑/删除确认；上传成功后 `refreshList()`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 删除最大 id 行导致客户端突然看到旧版本为「最新」 | 产品预期；删前 UI 提示「将影响 version/check」 |
| 大量历史行 list 无分页误用 | 默认 limit 50，上限 200 |
| update 改 `latestVersion` 与 semver 检查不一致 | 管理页说明版本号格式；不强制 DB 唯一约束 |
| 未配置口令时 503 | 与现 login 行为一致；文档/runbook 已强调环境变量 |

## Migration Plan

1. 部署新 `gateway-app` 镜像/二进制（无需 DDL，`version` 表已存在）。
2. 配置 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`（若尚未配置）。
3. 回滚：回退 gateway-app；DB 已删行无法自动恢复，需备份或重新上传。

## Open Questions

- 是否在列表中标示「当前 version/check 使用的行」（`id === max(id)`）— **建议实现**，仅 UI 高亮，无额外 API。
