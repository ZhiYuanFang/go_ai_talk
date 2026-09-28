## ADDED Requirements

### Requirement: 使用统计按用户列表 MUST 支持服务端分页与每页条数

`api-usage-stats.html` 与 `client-usage-stats.html` 的「按用户」wx 账号表 MUST 通过对应 `wx-list` 接口的 `page`、`pageSize` 做服务端分页，MUST NOT 再写死仅请求第一页固定 50 条。默认 `pageSize` MUST 为 **20**；页面 MUST 允许运维在至少包含 20、50、100 的选项中切换每页条数（不得超过后端允许的最大值 100）。页面 MUST 展示当前页码、总页数与总条数，并提供上一页、下一页，以及输入页码后跳转；越界页码 MUST 夹到合法范围。切换每页条数后 MUST 从第 1 页重新加载。

#### Scenario: 默认每页 20

- **WHEN** 管理员打开功能使用统计或客户端使用统计并切换到按用户且未改每页条数
- **THEN** 请求 wx-list 的 `pageSize` MUST 为 20 且 `page` MUST 为 1

#### Scenario: 下一页

- **WHEN** 总条数大于当前 pageSize 且管理员点击下一页
- **THEN** 页面 MUST 请求 `page` 为当前页 +1 的 wx-list 并刷新表格

#### Scenario: 修改每页条数回到首页

- **WHEN** 管理员将每页条数从 20 改为 50
- **THEN** 页面 MUST 以 `page=1&pageSize=50` 重新请求并展示第一页

#### Scenario: 页码跳转越界

- **WHEN** 总页数为 5 且管理员输入页码 99 并跳转
- **THEN** 页面 MUST 加载第 5 页（或等价合法最后一页），MUST NOT 发起非法超大 page 导致空态且无提示的静默失败

### Requirement: 分页下列表序号 MUST 为当前页内序号

按用户表最前列序号 MUST 为当前渲染结果（含本页宝宝名筛选后）从上到下的 1…N，MUST NOT 使用跨页全局偏移（如 `(page-1)*pageSize+i`）作为本变更要求。

#### Scenario: 第二页序号从 1 起

- **WHEN** 管理员在第 2 页且未筛选，表格有多行
- **THEN** 第一行序号 MUST 为 1，而非从 pageSize+1 起算
