## ADDED Requirements

### Requirement: 查史播报向量 Admin 代理

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python 查史播报缓存 Admin API（契约路径 `/v1/admin/history-speak-cache` 及其 bulk/dedupe/delete），支持列表、批量写入、去重与按 id 删除。Hub 侧路径 SHALL 使用 `/voice/admin/api/history-speak-vectors*`（与 intent/type vectors 命名对称）。MUST NOT 在 Go 进程内直连 Chroma 或新建查史播报向量表。批量条目 payload MUST 透传上游字段（含 `point_pick`、`open_speak`）；同 document 跳过策略由上游保证。

#### Scenario: 列表代理

- **WHEN** 运维已通过 Hub Admin 鉴权并请求查史播报向量列表
- **THEN** voice-service 调用 Python history-speak-cache 列表接口并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id 删除
- **THEN** 系统将请求代理至对应 Python admin 接口，且 MUST NOT 因代理路径改写 Redis/MySQL 权威数据

#### Scenario: 口令拒绝

- **WHEN** 请求缺少有效 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝该 Admin 请求

### Requirement: 查史播报向量 Hub 整理页

Hub SHALL 提供查史播报向量静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/history-speak-vectors*` 接口，支持列表浏览、bulk 写入、dedupe 与删除。页面 MUST 说明数据位于 Python `feeding_history_speak_cues`，payload 维为 `point_pick`（`latest`|`prefer_completed`）与 `open_speak`（`elapsed`|`in_progress`）。该页 MUST 可从意图智能体门户进入，并 MUST 登记静态路由与 auth exempt（与既有向量管理页同级）。

#### Scenario: 从门户进入并列表

- **WHEN** 运维从意图智能体门户进入查史播报向量页且 Admin 已登录
- **THEN** 页面可拉取并展示向量列表（或在上游未就绪时给出可理解错误提示）

#### Scenario: bulk 示例载荷

- **WHEN** 运维在页内按提示 bulk 写入含合法 `point_pick` 与 `open_speak` 的条目
- **THEN** 请求经 voice 代理到达 Python history-speak-cache bulk（成功语义以上游为准）
