## ADDED Requirements

### Requirement: voice Admin 代理类型向量列表

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `GET /v1/admin/type-cache`，支持分页查询参数（与意图向量缓存同类），并将结果返回给 Hub。MUST NOT 在 Go 进程内直连 Chroma 或新建类型向量表。

#### Scenario: 列出类型向量

- **WHEN** 运维携带合法 Admin 凭据请求类型向量列表
- **THEN** 系统返回分页结果（含 id、document、payload、quality 或 Python 已有等价字段、创建时间等）
- **AND** 上游失败时返回明确错误（不得假装空成功掩盖故障——上游空库 total=0 除外）

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 列表接口 MUST 拒绝

### Requirement: voice Admin 代理类型向量 bulk / dedupe / delete

系统 SHALL 提供经 voice Admin 口令鉴权的接口，分别代理：

- `POST /v1/admin/type-cache/bulk`（批量写入；同 document 跳过策略由上游保证）
- `POST /v1/admin/type-cache/dedupe`
- `DELETE /v1/admin/type-cache/{id}`

Hub 路径命名可使用 `/voice/admin/api/type-vectors*`，但 MUST 映射到上述 Python 契约。

#### Scenario: bulk 写入

- **WHEN** 运维提交含 `document` 与 `payload`（含 `target_type`/`op`/`history_mode`）的 bulk 请求
- **THEN** 系统代理至 Python 并返回上游汇总（如 added/skipped/failed）

#### Scenario: 删除不存在的向量

- **WHEN** 删除目标 id 在 Python 侧不存在
- **THEN** 系统 MUST 返回失败，MUST NOT 静默成功

### Requirement: Hub 类型向量管理页

系统 SHALL 在运维 Hub 提供独立「类型向量」管理页（导航可发现，与意图向量缓存并列），调用上述 Admin API 完成列表、批量写入、去重与删除。页面 MUST 提示 payload 语义为类型飞轮字段（非事件短说法），数据在 Python 类型向量库。

#### Scenario: 进页可见列表

- **WHEN** 运维已登录 Hub 并打开类型向量页
- **THEN** 可见向量表格（或空态）并由列表接口填充

#### Scenario: 页面可 bulk 与删除

- **WHEN** 运维在页内执行 bulk 或删除操作
- **THEN** 页面调用对应代理 API 并刷新或反馈结果
