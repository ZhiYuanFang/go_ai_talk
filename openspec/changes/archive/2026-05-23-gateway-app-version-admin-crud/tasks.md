## 1. 后端：会话与路由

- [x] 1.1 在 `gateway_app_version_admin.go` 抽取 `requireVersionAdminSession`（含口令未配置 503、未登录 401）
- [x] 1.2 在 `gateway_app_register.go` 注册 `list` / `get` / `update` / `delete` 路由
- [x] 1.3 在 `gateway_app_auth_exempt.go` 登记新路径（与 upload 相同：跳过 App JWT，handler 内校验管理会话）

## 2. 后端：CRUD 实现

- [x] 2.1 实现 `GET .../admin/list`（`id` DESC，`limit`/`offset`，返回 `items` + `total` 或仅 items）
- [x] 2.2 实现 `GET .../admin/get?id=`（404 语义）
- [x] 2.3 实现 `POST .../admin/update`（JSON：可改字段校验，`download_url` 不可改）
- [x] 2.4 实现 `POST .../admin/delete`（删行 + `tryRemoveApkForDownloadPath` + `InvalidateAppVersionLatestCache`）
- [x] 2.5 确认 `upload` 在 insert 后已调用缓存失效（已有则仅回归）

## 3. 前端：版本管理页

- [x] 3.1 扩展 `gateway-app-version-admin.html`：历史版本表格（列：id、版本号、上线时间、downloadUrl、强制更新、备注）
- [x] 3.2 标出当前 `version/check` 使用的行（最大 id，列表首行或徽章）
- [x] 3.3 行内/弹窗编辑：调用 update API（含 `minVersion`、`releaseDate`）
- [x] 3.4 删除确认后调用 delete API 并刷新列表
- [x] 3.5 上传成功后自动 `refreshList()`；统一错误展示

## 4. 文档与验收

- [x] 4.1 在 `README.MD` 或 `docs/runbooks/release-deploy-and-run.md` 补充管理 API 路径与口令环境变量（若尚未写明 list/更新/删除）
- [x] 4.2 手工验收：登录 → 列表 → 上传 → 检查列表 → 编辑 → 删除 → `GET /device/app/api/version/check` 与最大 id 行一致
