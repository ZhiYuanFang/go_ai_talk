# 提案：设备管理端设备记录分页与最近 HTTP 接口

## 背景

设备管理页 `admin.html` 当前一次拉取全量设备列表，无分页与搜索；`device_no` 不可点击跳转历史页。运营需要按设备号模糊查询、每页 5 条浏览，并查看各设备最近一次带 `deviceNo` 的 **HTTP** API（不含 WebSocket）。

## 目标

- 管理端新增分页 API：`GET /device/admin/api/user/list`（`page`、`pageSize` 默认 5、`q` 模糊匹配 `device_no`）。
- `user` 表增加 `last_api_path`、`last_api_at`；经网关边缘 middleware 异步写入（任意带 device 的 HTTP，排除 WS 与 `/device/internal/*`）。
- `admin.html`：注册区与「设备记录」分卡片；`device_no` 链接至 `/device/history/{deviceNo}`；展示最近接口与时间。

## 非目标

- 独立全屏设备列表页。
- WebSocket 路径统计。
- 修改 `history.html` 行为。

## 影响范围

- `device-service`：`user` 表、admin/internal API、领域服务。
- `gateway` 与 `gateway-app-server`：HTTP touch middleware。
- `resource/public/admin.html`、`api/v1/device_admin_http.go`、`api/v1/device_internal_full_http.go`。
