## Why

邀请码改为用户互推后，尚无好友码的用户需要运营入口获取邀请码。开通功能管理页上传微信群二维码（固定文件名、可设有效期），开通中心在有效期内展示「加群拿邀请码」，降低获客门槛。

## What Changes

- Admin（开通功能管理）：上传群二维码图片，强制落盘为 `er_code.png`（仅保留一张，可覆盖）；设置/更新有效期；过期后 Admin 仍可预览。
- gateway-app：在既有 APK 存储目录写入/覆盖 `er_code.png`；经既有 `/device/app/apk/er_code.png` 对外可读（已 auth exempt）。
- cash：单行元数据表记录 `expires_at` / `updated_at`；Admin 读写有效期；合成 `GET feature/catalog` 时，仅当 `now < expires_at` 时在响应**顶层**返回 `inviteGroupQrUrl`（含 cache-bust query）；过期或不存在则省略/空。
- 孪生客户端：`flutter_ai_talk` 同名 change——开通中心页级展示二维码与文案「加入微信群获取邀请码」。

## Capabilities

### New Capabilities

- `invite-group-qr-admin`：Admin 上传（gateway 落盘）+ cash 有效期配置与预览语义。
- `invite-group-qr-catalog`：catalog 顶层条件返回 `inviteGroupQrUrl`。

### Modified Capabilities

- （无已归档基线 capability；catalog 字段增量以本 change specs 为准。）

## Impact

- gateway-app：上传 handler（或扩展 version-admin 旁路）、静态下载复用。
- cash-service：schema、Admin API、catalog 合成；依赖 `GATEWAY_APP_PUBLIC_BASE_URL`（或等价）拼绝对/相对 URL。
- `cash-feature-admin.html`：上传 + 有效期 UI。
- Flutter 开通中心；兄弟仓契约对齐。
- 不新增测试；不新增背景 ticker（过期仅读时判断）。
