## Why

运维在「功能使用统计 / 客户端使用统计」的按用户列表里只能看到 `platform`（ios/android/web），无法区分具体机型（华为、小米、iPhone 11 等）。需要在用户维度再展示一列**客户端上报的设备型号原文**，服务端不做枚举映射，字段可空以兼容旧客户端与 Web。

## What Changes

- `wx` 表新增可空列 `device_model`（展示文案，非枚举）；`device-service` 启动 Ensure 幂等加列。
- 新增 App 写入接口（**新路径**，不改既有 v1 登录/client-usage body）：登录用户上报/覆盖本账号最近设备型号。
- `GET /device/admin/api/wx/list` 及 gateway 编排的 `usage/wx-list`、`client-usage/wx-list` 响应透传 `deviceModel`。
- Hub 页 `api-usage-stats.html`、`client-usage-stats.html` 的「按用户」表在「平台」旁增加「型号」列；空值显示「—」。
- **不**把型号做成 usage Redis 聚合维度；仅用户档案展示列。
- **不**新增 Redis 读缓存（沿用 MySQL `wx` 行字段）。
- Flutter 仓对口变更：`report-device-model-to-server`（客户端取机型并调用本仓新接口）。

## Capabilities

### New Capabilities

- `wx-device-model`：账号级设备型号存储、App 上报契约、Admin/usage 列表透传与 Hub 展示。

### Modified Capabilities

- `device-admin`：`wx/list` 每项增加可空 `deviceModel`。
- `gateway-app-api-usage-stats`：usage wx-list 与 `api-usage-stats.html` 增加型号列。
- `client-feature-usage-stats`：client-usage wx-list 与 `client-usage-stats.html` 增加型号列。

## Impact

- **进程**：`device-service`（schema + App API + admin wx/list）；`gateway-app-server`（反代已有 `/device/app/api/*`；usage/client-usage enrich 与两静态页）。
- **API**：新 App 路径（建议 `PUT /device/app/api/user/device-model`）；Admin 列表响应**加法字段**（Hub 自用，不破坏旧 App）。
- **DB**：device 库 `wx.device_model`。
- **usage 统计**：**待负责人确认**新 App 上报接口是否计入 App API usage（建议对标 `client-usage/report`：**不统计**并写入 `maintenance_skip`；未确认前不得改 denylist）。
- **基线**：对照 `openspec/specs/v3.2.0/spec.md`。
- **跨仓**：依赖 Flutter 发版后字段才会逐渐非空；旧数据保持空。
