## 1. API 与 gateway 编排（同 path 补字段）

- [x] 1.1 `DeviceWxListPage` / `deviceWxListItem` 解析并透传 device `wx/list` 的 `babyName`
- [x] 1.2 `api/v1`：`DeviceAdminUsageWxListReq/Item` 增加可选 `days` 与响应 `babyName`/`lastAt`；`DeviceAdminClientUsageWxListReq/Item` 同等
- [x] 1.3 `usagestats`（及 clientusage 等价）实现本页 wxId 批量取窗口内 `lastAt`（读既有 last 键取 max + days 过滤）
- [x] 1.4 `UsageWxList` / `ClientUsageWxList` handler 组装 `babyName`、`lastAt` 并传递 `days`

## 2. 运维 Hub 与反馈样式

- [x] 2.1 `admin.html`：将 `#feedbackCard` 移到 `#deviceRecordCard` 之后
- [x] 2.2 `admin.html`：反馈头部常驻「查看全部」→ `/device/admin/feedback-records`（不依赖 total>10）
- [x] 2.3 `pangbao-theme.css` + Hub/全页反馈表：仅「问题」列使用顶对齐专用 class

## 3. 使用统计两页前端

- [x] 3.1 `api-usage-stats.html`：用户维度设备号 `.device-link` → `/device/admin/history/{deviceNo}`（空不可点；避免与行选中冲突）
- [x] 3.2 `api-usage-stats.html`：用户表增加序号、最近使用时间列；本页 lastAt 排序 / 默认 wxId；有无宝宝名筛选；请求带 `days`
- [x] 3.3 `client-usage-stats.html`：与 3.1–3.2 同等行为（走 client-usage wx-list）

## 4. 自检

- [x] 4.1 本地打开 Hub：反馈位置与「查看全部」、问题列顶对齐目视确认
- [x] 4.2 两页统计：设备号跳转、序号、本页筛选/排序、无 babyName/lastAt=0 展示正常
