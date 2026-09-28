## Why

运维 Hub 与使用统计页存在若干交互缺口：统计列表里设备号不可跳转运维页、用户维度缺少「有无宝宝名」筛选与最近使用时间、Hub 上用户反馈模块位置偏后且入口依赖「展开更多」、反馈问题列纵向居中不便阅读。需要在不大改业务契约的前提下补齐运维效率。

## What Changes

- **运维 Hub（`admin.html`）**：用户反馈卡片上移至设备记录下方；头部常驻「查看全部」链至 `/device/admin/feedback-records`（不再依赖 total>10）。
- **反馈展示**：仅「问题」列取消纵向居中，顶对齐正常展示；其它列保持现状。
- **功能使用统计 / 客户端使用统计（两页）**：
  - 设备号可点击进入 `/device/admin/history/{deviceNo}`（复用既有 `.device-link` 样式）。
  - 用户维度：列表最前列展示序号（当前渲染序 1…N，非排序键）；增加「最近使用时间」列；表头点击仅对**当前页已加载列表**按 lastAt 降序重排；未启用 lastAt 排序时默认按 wxId 升序；增加「有无宝宝名」筛选（基于 `babyName` 非空，仅筛当前页）。
- **同 path 补字段（运维只读，非 App 契约）**：`GET .../usage/wx-list` 与 `GET .../client-usage/wx-list` 响应每项透传 `babyName`，并按顶栏时间窗口补 `lastAt`（该 wx 在窗口内最近成功使用 Unix 秒；无记录为 0）。可选查询参数 `days` 与页面时间范围对齐。不新开路由、不改 App 对外 API。

## Capabilities

### New Capabilities

- `admin-hub-ops-ux`：运维 Hub 反馈卡片顺序与常驻入口、反馈问题列顶对齐、统计页设备号跳转与用户维度纯前端交互（序号/筛选/本页排序）。

### Modified Capabilities

- `gateway-app-api-usage-stats`：usage `wx-list` 同 path 增加 `babyName`/`lastAt`（及 `days`），用户维度页消费上述字段。
- `client-feature-usage-stats`：client-usage `wx-list` 同等补字段与页面行为对齐。

## Impact

- **静态页**：`resource/public/admin.html`、`api-usage-stats.html`、`client-usage-stats.html`、`feedback-records.html`（若问题列 class）、`pangbao-theme.css`（反馈问题专用顶对齐）。
- **gateway-app**：`usagestats.DeviceWxListPage` 解析透传 `babyName`；usage / client-usage 的 WxList handler 补 `lastAt`（读既有 Redis last 键取 max，受 `days` 约束）；`api/v1` 对应 Req 结构体加法字段。
- **device-service**：无变更（`wx/list` 已返回 `babyName`）。
- **App 网关 / Bearer / usage 统计策略**：不新增对外 App 接口；运维读接口本就不计入 App usage；**不**改 `maintenance_skip.go`。
- **兼容**：运维 JSON 加法字段；旧前端忽略新字段仍可用。
