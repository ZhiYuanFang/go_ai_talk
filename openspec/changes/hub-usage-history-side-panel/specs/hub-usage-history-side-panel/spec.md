## ADDED Requirements

### Requirement: Usage stats user detail SHALL default to history tab with usage detail available

On Hub pages `api-usage-stats` and `client-usage-stats`, when the「按用户」panel shows a selected wx in the master–detail layout, the detail pane MUST provide two tabs: **历史事件** (default selected) and **使用明细**. Selecting **使用明细** MUST show the page’s previous per-user detail content (API call list on api-usage-stats; feature summary plus timeline on client-usage-stats). Narrow-screen stacking rules for `.hub-md` MUST remain.

「按用户」选中某 wx 时，右侧 MUST 提供 Tab：**历史事件**（默认选中）与 **使用明细**。使用明细 MUST 保留该页原有人均详情。`.hub-md` 窄屏回退 MUST 保持。

#### Scenario: Default tab is history

- **WHEN** 运维在按用户 Tab 点击某一 wx 行
- **THEN** 右侧 MUST 默认展示「历史事件」Tab 为选中态

#### Scenario: Switch back to usage detail

- **WHEN** 运维切换到「使用明细」
- **THEN** 系统 MUST 展示该页原先的人均使用明细内容（或等价数据）

### Requirement: History tab SHALL load device history via existing admin list without time filter

When the history tab is active for a selected user row that has a non-empty `deviceNo`, the page MUST fetch via existing `GET /device/admin/api/history/list` using only `deviceNo`, `page`, and `pageSize` (MUST NOT send `startTime`/`endTime` or otherwise require server API changes). History pagination page size MUST be fixed at **50**. The table MUST present columns aligned with the Hub device history page list (at least: ID, 事件名, 数量/时长, 开始时间, 结束时间, 备注) in read-only form (MUST NOT require edit/delete actions in this panel). An empty `deviceNo` MUST show an empty-state and MUST NOT call list with a blank device. The UI MUST offer a link to the full device history admin page for that `deviceNo`.

历史 Tab 在有非空 `deviceNo` 时 MUST 仅用既有 `GET /device/admin/api/history/list` 的 `deviceNo`/`page`/`pageSize` 拉取，MUST NOT 传时间窗参数、MUST NOT 依赖服务端契约变更。每页 MUST 固定 **50**。表列 MUST 对齐运维历史列表（至少含 ID、事件名、数量/时长、开始、结束、备注），只读。无 `deviceNo` MUST 空态。MUST 提供完整运维页链接。

#### Scenario: Fetch without time bounds

- **WHEN** 选中一带 `deviceNo` 的用户且历史 Tab 激活
- **THEN** list 请求 MUST 含 `deviceNo` 与 `pageSize=50`，MUST NOT 依赖 startTime/endTime 查询参数

#### Scenario: Columns match history page

- **WHEN** 历史列表有数据
- **THEN** 表头 MUST 含事件名、数量/时长、开始时间、结束时间、备注（或与 history.html 历史表等价列集合，不含操作）

#### Scenario: No device number

- **WHEN** 选中行 `deviceNo` 为空
- **THEN** 历史 Tab MUST 显示无设备提示且 MUST NOT 发起有效 list 拉取

### Requirement: User list defaults SHALL be baby-named, page size 100, and days 3

On both usage stats pages, the「按用户」controls MUST default to: baby-name filter **有宝宝名** (`yes`), wx list page size **100**, and usage days select **3**. Operators MUST still be able to change these. Baby-name filtering MAY remain client-side on the currently loaded page. Usage `days` MUST NOT be required to filter the embedded history list.

两页按用户控件 MUST 默认：宝宝名=有宝宝名、每页=100、days=3。宝宝名 MAY 保持当前页客户端过滤。usage `days` MUST NOT 作为嵌入历史列表的时间过滤条件。

#### Scenario: Fresh load defaults

- **WHEN** 运维打开 api 或 client 使用统计页并进入按用户
- **THEN** 宝宝名筛选 MUST 为「有宝宝名」、每页 MUST 为 100、时间范围 MUST 为近 3 天（除非另有覆盖）
