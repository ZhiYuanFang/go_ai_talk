## ADDED Requirements

### Requirement: voice Admin 代理覆盖缺口列表与删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/miss-inbox` 的分页列表与按 id 删除。Hub 路径 SHALL 使用 `/voice/admin/api/miss-inbox` 与 `/voice/admin/api/miss-inbox/{id}`。列表 MUST 支持查询参数 `offset`、`limit`、`status`、`fail_kind`（空值语义与上游一致）。MUST NOT 在 Go 进程内直读 Python JSONL 或新建 miss 表。MUST NOT 提供 bulk、PATCH dismissed 或从 Hub 新建记录的代理。

#### Scenario: 列出覆盖缺口

- **WHEN** 运维已通过 Hub Admin 鉴权并请求覆盖缺口列表（含可选 status/fail_kind）
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/miss-inbox` 并返回分页结果（透传字段含 utterance、reply、fail_kind、device_no 等）

#### Scenario: 删除覆盖缺口

- **WHEN** 运维调用按 id 删除
- **THEN** 系统 MUST 代理至 Python `DELETE /v1/admin/miss-inbox/{id}`，且 MUST NOT 在 Go 侧改写其它收件箱

### Requirement: Hub 覆盖缺口查删页

Hub SHALL 提供覆盖缺口静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/miss-inbox*`，支持分页浏览、按 status/fail_kind 筛选与单条删除（删除前 MUST 确认）。页面 MUST 说明数据独立于规则缺口与时间飞轮、用途为终局答不出盲区巡检。默认列表 MUST 以 `status=open` 查询（除非运维改选全部）。MUST 登记静态路由与 auth exempt，并 MUST 可从意图智能体门户进入。

#### Scenario: 从门户进入并查删

- **WHEN** 运维从门户打开覆盖缺口页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、筛选与删除，且 MUST NOT 展示飞轮 bulk/dedupe 操作

### Requirement: device_no 链到 history 运维页

覆盖缺口列表中，当某条记录的 `device_no` 非空时，页面 SHALL 将其渲染为指向既有设备 history 运维页的链接，路径为 `/device/admin/history/{deviceNo}`（deviceNo 须做 URL 编码）。`device_no` 为空时 MUST NOT 生成无效链接。

#### Scenario: 有设备号可跳转

- **WHEN** 列表项含非空 `device_no`
- **THEN** 运维点击该设备号 MUST 打开对应 history 运维页

#### Scenario: 无设备号不链

- **WHEN** 列表项 `device_no` 为空
- **THEN** 页面 MUST 以非链接方式展示占位，且 MUST NOT 链到 history 根路径

### Requirement: 意图智能体门户暴露覆盖缺口入口

Hub 意图智能体门户页 SHALL 在规则工单区提供前往覆盖缺口页的入口（可与规则缺口并列）。门户 MUST NOT 将该入口放在飞轮整理主网格作为向量灌种卡。

#### Scenario: 门户工单区可见

- **WHEN** 运维已登录并打开意图智能体门户
- **THEN** 页面 MUST 展示覆盖缺口入口且可进入查删页
