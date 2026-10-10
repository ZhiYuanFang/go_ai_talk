## ADDED Requirements

### Requirement: voice Admin 代理规则缺口列表

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `GET /v1/admin/rule-gaps`，支持 `offset`、`limit`、`dimension`、`status` 查询参数，并将分页结果返回给 Hub。MUST NOT 在 Go 进程内直读 Python JSONL 文件或新建缺口表。

#### Scenario: 列出 open 缺口

- **WHEN** 运维携带合法 Admin 凭据请求规则缺口列表且 `status=open`（或默认）
- **THEN** 系统返回 `total` 与 `items`（含 dimension、segment/整句、reason、创建时间等 Python 已有字段）
- **AND** 请求失败时返回明确错误（不得假装空成功掩盖上游故障——空文件导致 total=0 除外）

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 列表接口 MUST 拒绝

### Requirement: voice Admin 代理标已处理

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `PATCH /v1/admin/rule-gaps/{id}`，将缺口 `status` 更新为 `dismissed`（或请求体指定的合法状态）。

#### Scenario: 标 dismissed

- **WHEN** 运维对存在的缺口 id 提交标已处理
- **THEN** 系统代理成功后该条在默认 open 列表中不再出现（除非筛选包含 dismissed）

#### Scenario: 缺口不存在

- **WHEN** 目标 id 在 Python 侧不存在
- **THEN** 系统 MUST 返回失败（透传或映射为业务错误），MUST NOT 静默成功

### Requirement: Hub 规则缺口管理页

系统 SHALL 在运维 Hub 提供独立「规则缺口」管理页（导航可发现，与意图向量缓存并列），调用上述 Admin API 展示缺口列表，支持按维度/状态筛选与分页，并支持将条目标为已处理。页面 MUST 提示数据来自 Python 收件箱、整理规则须改代码发版而非热更新。

#### Scenario: 进页可见列表

- **WHEN** 运维已登录 Hub 并打开规则缺口页
- **THEN** 可见缺口表格（或空态）并由列表接口填充

#### Scenario: 行内标已处理

- **WHEN** 运维对某条 open 缺口确认标已处理
- **THEN** 页面调用代理 PATCH 并刷新列表
