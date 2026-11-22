# observability

## Purpose

日志级别、背景循环治理、中文文档/语言合规与可观测约定的行为规格。 Keeps ops signals, docs language, and background-task approvals consistent.

<!-- migrated-from: background-loop-task-governance, chinese-documentation-governance, documentation-language-compliance, logger-level-env -->

## Requirements

### Requirement: 循环后台任务 MUST 经 OpenSpec 明确批准后方可引入

系统 MUST **默认禁止**在 `internal/services/**` 业务实现中新增循环后台任务，包括但不限于：`time.NewTicker` / `time.Tick` 轮询、`for { select {} }` 常驻 goroutine 扫描 MySQL/Redis/outbox 表、定时 reconciler、pending 业务表全表/分页扫描兜底。

新增上述任务 **MUST** 在 OpenSpec **proposal 与 design** 中写明：任务名称、宿主进程、周期/触发条件、环境开关、失败语义、关闭方式，以及 **为何不能** 在请求内同步完成或使用 AMQP push consumer 替代。

#### Scenario: 未批准的后台 ticker 被拒绝合入

- **WHEN** PR 在业务包内新增 `Start*Reconciler` 或 ticker 扫表且无对应 OpenSpec 变更引用
- **THEN** 评审 MUST 要求补充已批准的变更或删除该后台任务

#### Scenario: 已批准 UCG outbox relay 仍允许

- **WHEN** 变更引用已归档或进行中的 UCG OpenSpec（如 audit publish outbox、chat persist）且在 design 中声明 ticker 语义
- **THEN** 该后台任务 MAY 存在于 `ucg-service` 进程

### Requirement: AMQP push consumer 与 ticker 扫表 MUST 区分治理

经 RabbitMQ **broker push**（`autoAck=false`）的消息 consumer **不视为** ticker 扫表任务，但 **MUST** 在 OpenSpec 变更中声明队列名、routing key 与宿主进程。HTTP Management API Pull 轮询队列 **视为** 循环后台任务，适用批准流程。

#### Scenario: UCG AMQP consumer 合规

- **WHEN** `ucg-service` 启动 AMQP push consumer 消费 `ucg.audit.*` 或 `ucg.recommend.score.q`
- **THEN** 该 consumer MUST 有 OpenSpec 依据且 MUST NOT 部署在已删除的 worker-service 内

### Requirement: 角色边界变更 MUST 伴随文档更新

当后台任务执行角色或宿主进程发生调整时，运行文档与部署说明 MUST 同步更新，以确保运维与开发对角色边界认知一致。

#### Scenario: 移除 worker-service 后文档更新

- **WHEN** 完成 worker-service 删除与缓存同步简化
- **THEN** runbook 与部署清单 MUST 说明 history/device 不再依赖 worker；gateway MUST NOT 启动业务后台任务；UCG 后台任务宿主为 ucg-service

---

### Requirement: OpenSpec 工件默认使用中文

系统在创建或更新 OpenSpec 变更工件时，说明性文本 SHALL 使用中文撰写，包括 proposal、design、specs、tasks。

#### Scenario: 创建新变更工件
- **WHEN** 用户通过 OpenSpec 工作流生成新的 proposal/design/specs/tasks
- **THEN** 生成内容中的说明性文本 SHALL 为中文

### Requirement: 必要技术标识允许保留英文

系统在文档中文化过程中 SHALL 允许保留必要英文技术标识，包括环境变量、路径、接口、协议和代码符号。

#### Scenario: 文档包含技术标识
- **WHEN** 文档中出现环境变量名、API 路径或代码符号
- **THEN** 这些标识 SHALL 保持英文原文，不做强制翻译

---

### Requirement: 变更文档需要通过语言合规检查

系统在变更进入实施阶段前 SHALL 完成文档语言合规检查，若说明性文本以英文为主则不得作为可实施输入。

#### Scenario: 变更进入 apply 前校验
- **WHEN** 变更已生成 proposal、design、specs、tasks 并准备进入实施阶段
- **THEN** 文档语言合规检查 SHALL 确认说明性文本为中文，否则应阻止进入实施

### Requirement: 语言规则适用于增量更新

系统在对已有变更进行增量更新时 SHALL 继续遵循中文文档规则，不因历史内容存在英文而豁免。

#### Scenario: 更新已有变更工件
- **WHEN** 用户对现有变更工件进行追加或修订
- **THEN** 新增与修改内容 SHALL 使用中文说明并保持术语一致

---

### Requirement: GF_LOGGER_LEVEL 覆盖默认 logger

系统 MUST 在长期运行服务启动时，于 yaml logger 配置已应用到默认 logger 之后，读取环境变量 `GF_LOGGER_LEVEL`：若 trim 后非空，则对该默认 logger 调用与 GoFrame 一致的级别字符串设置（如 `SetLevelStr`），使其覆盖 yaml 中的 `logger.level`；若为空，则 MUST 保持 yaml 级别不变。

#### Scenario: 生产设为 prod

- **WHEN** 进程环境 `GF_LOGGER_LEVEL=prod` 且服务已调用 `loggercfg.ApplyFromEnv`
- **THEN** 默认 logger 仅输出 WARN / ERRO / CRIT（及框架强制保留级别），`Infof`/`Debugf` 不再出现在该进程日志中

#### Scenario: 留空沿用 yaml

- **WHEN** `GF_LOGGER_LEVEL` 未设置或为空字符串
- **THEN** 默认 logger 级别保持配置文件中的 `logger.level`（当前服务配置多为 `all`）

#### Scenario: 生效可观测

- **WHEN** 非空 `GF_LOGGER_LEVEL` 成功应用
- **THEN** 进程 MUST 打出一条 Warning 级日志，标明服务名与所应用级别字符串

#### Scenario: 非法级别

- **WHEN** `GF_LOGGER_LEVEL` 为 GoFrame 不识别的字符串
- **THEN** 系统 MUST 记录 Error 且 MUST NOT 静默成功；logger 级别 MUST 保持应用前状态（yaml 已加载的级别）

---
