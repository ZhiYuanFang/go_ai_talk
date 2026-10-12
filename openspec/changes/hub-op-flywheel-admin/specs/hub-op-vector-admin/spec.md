## ADDED Requirements

### Requirement: voice Admin 代理 OP 轨飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/op-cache` 的 list、bulk 与按 id delete，支持分页查询参数（与既有向量缓存同类）。Hub 路径 SHALL 使用 `/voice/admin/api/op-vectors*`。MUST NOT 在 Go 进程内直连 Chroma 或新建 OP 飞轮表。MUST NOT 提供 dedupe 代理（上游无对应路由）。

#### Scenario: 列出 OP 飞轮

- **WHEN** 运维携带合法 Admin 凭据请求 OP 向量列表
- **THEN** voice-service 调用 Python op-cache 列表并返回分页结果
- **AND** 上游失败时返回明确错误（上游空库 total=0 除外）

#### Scenario: bulk 写入 op_rail

- **WHEN** 运维提交含 `document` 与 `payload.op_rail`（create|read|update|delete|exit|other）的 bulk 请求
- **THEN** 系统代理至 Python op-cache bulk 并返回上游汇总

#### Scenario: 删除不存在的向量

- **WHEN** 删除目标 id 在 Python 侧不存在
- **THEN** 系统 MUST 返回失败，MUST NOT 静默成功

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 接口 MUST 拒绝

### Requirement: Hub OP 轨飞轮整理页

Hub SHALL 提供 OP 轨飞轮静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/op-vectors*`，支持列表、bulk 与删除。页面 MUST 说明数据位于 Python `feeding_op_cues`、document 须为整段原文、payload 为 `op_rail`；MUST NOT 展示「整理重复」类 dedupe 操作。该页 MUST 可从意图智能体门户进入，并 MUST 登记静态路由与 auth exempt。

#### Scenario: 门户进入并可列表

- **WHEN** 运维从意图智能体门户进入 OP 轨页且 Admin 已登录
- **THEN** 可见列表（或空态）并由列表接口填充

#### Scenario: 页内 bulk

- **WHEN** 运维在页内提交合法 bulk JSON
- **THEN** 请求经 voice 代理到达 Python op-cache bulk（成功语义以上游为准）
