## Context

- **gateway-app-server** 已提供 App 登录、刷新、`GET /device/app/api/version/check` 等接口，版本元数据来自 `ai_voice_app.version`（`dao.AppVersion`）。
- 现网 APK 需由运维上传并维护 `download_url`；云服务器上期望「管理页上传 → 落盘 → 可公网下载的 URL → 写库」闭环。
- 仓库约定：`manifest` 中网关进程独立配置；敏感口令 **SHALL** 通过配置/环境变量注入，**禁止**在源码中写死生产口令（部署侧按运营约定设置，例如与本次需求一致的口令由运维写入私有配置）。

## Goals / Non-Goals

**Goals:**

- 提供受口令保护的「版本管理」页面（HTML 即可，可与现有 `resource/public` 风格一致或独立路由）。
- 接收 Android APK 上传，保存到 **`/apk/ai_talk/`**（不存在则 `MkdirAll`），文件名安全、可追溯（建议：`{sanitizedVersion}_{UTC时间戳}.apk` 或 UUID 前缀防冲突）。
- 根据 **`publicBaseUrl`（或等价配置项）** 生成 **绝对 HTTPS/HTTP 下载 URL**（例如 `{publicBaseUrl}/device/app/apk/{filename}`），并写入 **`version.download_url`**；若同次提交维护 `latest_version`、`release_notes`、`force_update`、`release_date`，则 **INSERT** 新行（与现有「按 `id` 降序取最新一条」的读逻辑一致）。
- 提供 **GET** 下载路由：仅允许映射到上述目录下已登记文件名或白名单规则，**禁止**路径穿越与非 `.apk` 扩展名。

**Non-Goals:**

- iOS IPA / App Store 链接自动化。
- 分片续传、多租户多应用线（本变更仅服务当前 App 线）。
- 将 APK 存入对象存储（OSS）；若后续需要，可再增适配层。

## Decisions

| 决策 | 说明 | 备选（未采纳） |
|------|------|----------------|
| 鉴权模型 | 首屏输入管理员口令；校验通过后签发 **短期 HttpOnly Cookie**（或服务端 session + cookie），后续上传与写库接口校验该凭证。 | 每次 multipart 带口令：易泄露到代理日志。 |
| 口令来源 | `gateway-app-server` 独立配置（如 `gatewayApp.versionAdmin.password` 或 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`），与主网关 `9701` 配置隔离。 | 写死在代码：违反仓库安全约定。 |
| 下载 URL 形态 | **绝对 URL** 写入 `download_url`，便于客户端在非同源场景直接使用。`publicBaseUrl` 必填校验（部署文档说明云主机域名/反代前缀）。 | 相对路径：客户端拼接易错。 |
| 文件落盘路径 | 固定 **`/apk/ai_talk/`**（Linux 绝对路径），与需求一致；进程用户需对该路径有写权限（部署 runbook 说明 `chown`/`mkdir`）。 | 可配置根目录：可作为后续增强，首版可按需求固定。 |
| 写库策略 | **INSERT** 新版本行，保证历史可追溯；`version/check` 已按 `id` DESC 取最新。 | 仅 UPDATE 最后一行：无审计、易丢历史。 |
| 静态页挂载 | 挂在 gateway-app-server 已有 HTTP 栈（如 `resource/public` 下新 html + 专用 API），或注册独立 `GET /device/app/admin/version` 返回页面。 | 独立前端工程：超出本变更范围。 |

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 根目录 `/apk` 权限不足导致创建失败 | 部署文档写明目录预创建与 systemd 用户；启动时若创建失败打明确日志并拒绝上传接口。 |
| 大文件占满磁盘 | 配置 **单文件大小上限**（如 200MB）；超限返回 413。 |
| 口令泄露 | 强制 HTTPS、HttpOnly Cookie、口令仅环境变量注入；可选 IP 白名单（后续）。 |
| `publicBaseUrl` 配错导致客户端无法下载 | 管理页展示「即将写入的完整 URL」预览；健康检查或文档示例。 |
| 同域名下 APK 被枚举 | 文件名含不可猜 UUID；不列出目录索引；可选短期 signed query（非首版必做）。 |

## Migration Plan

1. 在云主机创建 `/apk/ai_talk/` 并授权给 gateway-app-server 运行用户（若进程内 `MkdirAll` 成功可省略预创建，但 `/apk` 父级权限仍需满足）。
2. 在 **非提交** 的私有配置或环境变量中设置管理员口令与 `publicBaseUrl`。
3. 发布新 gateway-app-server 二进制；验证：登录管理页 → 上传测试 APK → DB `download_url` 可访问 → 客户端 `version/check` 返回一致。

## Open Questions

- 是否在首版支持 **仅更新 `download_url` 而不改版本号**（例如热修包）：当前建议每次上传绑定表单中的 `latest_version` 一并提交，减少半条记录状态。
