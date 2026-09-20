## ADDED Requirements

### Requirement: 可能无行的单行查询 MUST 使用 One + IsEmpty

在 `internal/services/**` 与 `internal/controller/**` 中，凡对 GoFrame/gdb 模型执行**单行**查询且业务上允许零行（可选配置行、按条件取最新、按业务键探测是否存在等），实现 MUST 使用 `.One()`，并在 `err == nil` 后通过 `IsEmpty()` 区分「无行」与「有行」；有行后再 `Struct`（或等价字段读取）。MUST NOT 依赖单结构体 `.Scan(&T)` 在空集时返回 `nil`。若因历史原因保留 `Scan`，则 MUST 将 `errors.Is(err, sql.ErrNoRows)`（及驱动等价空结果错误）视为「无行」，MUST NOT 将其作为系统失败向上抛出或打成失败级日志。

#### Scenario: 作者无 pending 资料审核 job

- **WHEN** 调用方请求作者视角 profile 合并，且该 `wx_id` 在 `ucg_profile_audit_job` 中不存在 `status=pending` 行
- **THEN** 系统 MUST 将结果视为「无待审 job」（例如 `ok=false`），MUST NOT 因空结果集将 `sql.ErrNoRows` 记为失败 WARN，MUST 仍返回已发布 profile 视图

#### Scenario: 可选覆盖行不存在

- **WHEN** 业务读取用户级可选覆盖表（如 AI quota override）且该用户无覆盖行
- **THEN** 系统 MUST 回落默认配额/空覆盖视图并成功返回，MUST NOT 因空集 `Scan` 错误导致接口失败

### Requirement: 空集 MUST NOT 冒充系统失败或无限 requeue

当「无行」属于预期业务状态时，系统 MUST NOT：将裸 `sql.ErrNoRows`（或空集 Scan 错误）作为 HTTP/RPC 系统错误返回给调用方；或在审核等 MQ consumer 中因该错误对消息 Nack/requeue。按实体 ID 加载且行已不存在时，consumer MUST 视为过期/缺失并 Ack（或等价丢弃），MUST NOT 因空集读失败反复入队。

#### Scenario: 审核消息指向已删除实体

- **WHEN** 审核 MQ 消息携带的 jobId/postId/commentId/messageId 在库中已不存在
- **THEN** consumer MUST 跳过该消息并 Ack，MUST NOT 因单行空查询错误导致消息持续 requeue

#### Scenario: 业务必须存在的行缺失

- **WHEN** 查询语义要求行必须存在（例如按订单号履约）且库中无行
- **THEN** 系统 MUST 返回明确业务错误（如订单不存在），MUST NOT 向调用方暴露裸 `sql.ErrNoRows`

### Requirement: 列表 Scan 与必有行回读不受本约定禁止

对切片目标的 `.Scan(&[]T)`（列表/分页/批量）以及在同一流程中刚成功插入、或已 `Ensure*Row`/`Count>0` 保证存在后的单行回读，实现 MUST 被允许继续使用 `Scan`；本约定 MUST NOT 要求将其改为 `One`。本约定的禁止范围 MUST 仅针对「空集为正常可能」且「错误被当作系统失败」的单结构体路径。

#### Scenario: 管理端分页列表

- **WHEN** 管理接口按分页查询多行并 `Scan` 到切片
- **THEN** 零行结果 MUST 表现为空列表，且本约定 MUST NOT 要求改为 `One`

### Requirement: 全局文档与评审 MUST 记载空集查询约定

`openspec/project.md` MUST 包含「GoFrame 单行空结果查询」强制约定（含推荐写法、禁止写法、排除项与评审检查）。`AGENTS.md` MUST 包含可执行摘要并指向 `project.md` 对应章节。涉及新增/修改单行「可能无行」查询的变更，评审 MUST 检查是否违反该约定。

#### Scenario: 文档可被 AI 与评审发现

- **WHEN** 贡献者或 AI 阅读 `AGENTS.md` / `openspec/project.md`
- **THEN** 均可定位到单行空结果查询的强制约定，且 `AGENTS.md` 摘要与 `project.md` 权威描述一致指向
