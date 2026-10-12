## ADDED Requirements

### Requirement: voice Admin 代理 R·数据源飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/read-source-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/read-source-vectors*`。MUST NOT 直连 Chroma 或新建 read-source 表。MUST NOT 提供 dedupe 代理。

#### Scenario: 列出 R·source 飞轮

- **WHEN** 运维携带合法 Admin 凭据请求 R·数据源向量列表
- **THEN** voice-service 调用 Python read-source-cache 列表并返回分页结果

#### Scenario: bulk 写入 read_source

- **WHEN** 运维提交含 `document` 与 `payload.read_source`（`record` 或 `predict`）的 bulk 请求
- **THEN** 系统代理至 Python read-source-cache bulk 并返回上游汇总

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 接口 MUST 拒绝

### Requirement: Hub R·数据源飞轮整理页

Hub SHALL 提供 R·数据源飞轮静态管理页，调用 `/voice/admin/api/read-source-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_read_source_cues`、document 为整段原文、payload 为 `read_source`（record=查史 / predict=预测）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 门户进入并可列表

- **WHEN** 运维从门户进入 R·数据源页且 Admin 已登录
- **THEN** 可见列表（或空态）并由列表接口填充
