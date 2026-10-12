## ADDED Requirements

### Requirement: voice Admin 代理时间飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/time-cache` 的 list、bulk 与按 id delete，支持分页查询参数（与既有向量缓存同类）。Hub 路径 SHALL 使用 `/voice/admin/api/time-vectors*`。MUST NOT 在 Go 进程内直连 Chroma 或新建时间飞轮表。MUST NOT 提供 dedupe 代理（上游无对应路由）。批量 payload MUST 透传上游；配方合法性校验由 Python 保证。

#### Scenario: 列出时间飞轮

- **WHEN** 运维已通过 Hub Admin 鉴权并请求时间飞轮列表
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/time-cache` 并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id delete
- **THEN** 系统 MUST 代理至 Python 对应路由，且 MUST NOT 在 Go 侧落库或改写配方语义

### Requirement: Hub 时间飞轮整理页

Hub SHALL 提供时间飞轮静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/time-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_time_cues`、document 为整段原文、payload 为相对时间配方（禁止以绝对 Unix 作主缓存）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 从门户进入并整理

- **WHEN** 运维从门户进入时间飞轮页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、bulk 写入与按 id 删除，并展示相对配方灌种提示
