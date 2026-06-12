## ADDED Requirements

### Requirement: ucg_post SHALL include audit_version for CAS

表 `ucg_post` MUST 新增列 `audit_version INT NOT NULL DEFAULT 1`。新建 pending 帖 MUST 从 1 起始。用户再提审进入 pending 时 MUST `audit_version++`。消费者 CAS 审态变更成功 MUST NOT 递增该列。

#### Scenario: 新帖默认版本

- **WHEN** 用户 submit 创建帖子
- **THEN** 插入行 MUST 含 `audit_version=1` 且 `status=1`

#### Scenario: 再提审递增版本

- **WHEN** 作者对已发布或驳回帖 submit 再提审
- **THEN** 行 MUST `audit_version` 递增且 `status=1`；MUST NOT 回滚帖文正文等业务字段

### Requirement: ucg_post_comment SHALL include audit status columns

表 `ucg_post_comment` MUST 新增：

- `status TINYINT NOT NULL` — 0 draft（保留）、1 pending_audit、2 published、3 rejected
- `audit_version INT NOT NULL DEFAULT 1`
- `reject_reason VARCHAR`（可空）

首评 `audit_version=1`；再提审 MUST 递增。消费者 CAS MUST 使用 `status` + `audit_version`。

#### Scenario: 新评论 pending

- **WHEN** 用户发表评论
- **THEN** 插入 MUST `status=1`、`audit_version=1`

### Requirement: ucg_profile_audit_job SHALL store pending profile patches with audit_version

数据库 MUST 提供 `ucg_profile_audit_job`（或 design 等价命名）存储待审资料 patch，至少含：`wx_id`、`nickname`、`avatar_key`、`bio`、`status`（1/2/3）、**`audit_version`**、`reject_reason`、时间戳。

`audit_version` 为该 job 的 **唯一**审核轮次源；资料 CAS 与 MQ 载荷 MUST 读此列。MUST NOT 用 Redis 或 `ucg_profile` 表列作为审核版本。

#### Scenario: 资料提交创建 job

- **WHEN** 用户 PUT profile 变更
- **THEN** MUST 插入 pending job 行（`audit_version=1` 或再提审递增）而非仅写 Redis

### Requirement: ucg_chat_message SHALL include audit_status and audit_version

表 `ucg_chat_message` MUST 新增（或规范既有 `status` 字符串）：

- `audit_status`：`pending` | `approved` | `rejected`
- `audit_version INT NOT NULL DEFAULT 1`
- `reject_reason`（可空）

Redis 消息 JSON MUST 镜像 `audit_status` 与 `audit_version`，且与 MySQL 列语义一致。私信 CAS 与 MQ 载荷 MUST 以 `ucg_chat_message.audit_version` 为权威。

#### Scenario: 聊天消息初始 pending

- **WHEN** WS 发送消息并投递 Redis
- **THEN** `ucg_chat_message`（或等价 outbox 镜像）MUST `audit_status=pending` 且 `audit_version=1`；Redis JSON MUST 含相同 `audit_version`
