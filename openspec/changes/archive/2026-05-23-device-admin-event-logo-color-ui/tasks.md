## 1. 主网关：同源图片访问

- [x] 1.1 在 gateway-service 安装 `GET /ai_talk_images/*` 反代至 device-service（复用或抽取与 gateway-app 相同的 proxy 逻辑，目标 URL 与 `DEVICE_API_PROXY_URL` 一致）
- [x] 1.2 确认反代路径不与管理 API、其它路由冲突；本地/Compose 下用 wget/curl 验证 `:9701/ai_talk_images/...` 可 200

## 2. 管理页：列表展示与同源 URL

- [x] 2.1 更新 `admin.html`：事件表含 Logo、色调列；`eventLogoUrl` 改为 `location.origin` + path（保留 https 绝对 URL 兼容）
- [x] 2.2 Logo 无图占位可点击；色调列色块 + 文本；列表加载失败时在 `eventMsg` 明示
- [x] 2.3 `register.go` 为 `/device/admin` 静态响应加 `Cache-Control: no-store`（若由网关 ServeFile）

## 3. 管理页：行内点击编辑

- [x] 3.1 实现 `submitEventRowUpdate(row, { color?, logoFile? })`：multipart 提交完整字段至 `event/update`
- [x] 3.2 色调列：隐藏 `input[type=color]`，点击色块触发；变更后提交并 `loadEventList()`
- [x] 3.3 Logo 列：隐藏 `input[type=file]`，点击图/占位触发；选文件后提交并刷新
- [x] 3.4 提交中行级 loading/禁用，错误汇总到 `eventMsg`

## 4. 文档与验收

- [x] 4.1 更新 `README.MD` 或 runbook：事件 logo 在管理页为**主网关同源**访问；部署需 gateway + device-service
- [x] 4.2 手工验收：:9701 打开管理页 → 见 Logo/色调列 → 图可显示 → 点击改色/换图 → 刷新后持久化
