## 1. gateway 落盘

- [x] 1.1 Admin 上传 API：强制保存为 `ApkStorageDir/er_code.png`（覆盖）；鉴权对齐现有 Admin
- [x] 1.2 确认 `/device/app/apk/er_code.png` 可匿名下载（复用现路由）

## 2. cash 元数据与 catalog

- [x] 2.1 schema：`invite_group_qr` singleton（expires_at、updated_at、file_name）
- [x] 2.2 Admin GET/POST 有效期（及 touch updated_at）；过期语义对 App 失效、Admin 仍可读
- [x] 2.3 catalog 顶层条件返回 `inviteGroupQrUrl`（PublicBaseURL + path + `?v=`）

## 3. Admin UI

- [x] 3.1 `cash-feature-admin.html`：上传、有效期、预览（含已过期提示）

## 4. 验收

- [x] 4.1 `openspec validate invite-group-qr --strict`
