## Why

Hub「App API / 客户端使用统计」按用户时，右侧只有调用/功能明细；要看该设备真实历史事件仍须点设备号跳进独立运维页，对照成本高。运维希望点选用户后在右侧直接看历史列表，并默认收窄到「有宝宝名 + 较大页 + 近三天」活跃用户。

## What Changes

- **右侧 Tab**：两页（`api-usage-stats.html`、`client-usage-stats.html`）按用户 detail 改为 Tab——默认 **历史事件**，可切回 **使用明细**（api 页=原 API 调用表；client 页=原功能汇总+时间线）。
- **历史只读嵌入**：选中 wx 且有 `deviceNo` 时，调用**既有** `GET /device/admin/api/history/list`（仅 `deviceNo`/`page`/`pageSize`），表头与单元格展示对齐 `history.html` 历史表（只读、无操作列）；右侧历史分页 **固定每页 50**；链到完整运维页。无 `deviceNo` 时空态。
- **左栏默认**：宝宝名筛选默认「有宝宝名」；每页默认 **100**；usage `days` 默认 **3**（仅影响统计活跃窗与左栏，**不**用于过滤历史）。
- **非目标**：不改 history-service / admin API 契约；不做历史时间窗筛选；不把增删改搬进统计页；不改 App history；不改宝宝名服务端筛选；不改 usage Redis 口径。

## Capabilities

### New Capabilities

- `hub-usage-history-side-panel`：两页使用统计按用户右侧 Tab（历史默认 + 使用明细）、左栏默认（有宝宝名 / 100 / 近 3 天）、历史表样式对齐运维历史页且只读分页 50。

### Modified Capabilities

- （无）

## Impact

- **静态页**：`resource/public/api-usage-stats.html`、`client-usage-stats.html`
- **gateway-app**：静态资源发版即可
- **history-service / api**：无改动
- Twin Flutter：无
