## ADDED Requirements

### Requirement: ucg-service SHALL expose internal sim unread conversation sample API

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/chat/sim-unread-sample`，供 `sim-user-service` 等受信内部调用方抽取 **一条** eligible 未读会话样本。鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。

请求体 MUST 含 `simWxIds`（`int64` 数组，非空）。MAY 含 `messageLimit`（默认 20，最大 50）。

**eligible 定义**（direct 1:1 会话）：

- sim 侧成员 `m`：`m.wx_id IN simWxIds` 且 `m.unread_count > 0` 且 `m.deleted_at = 0`
- 对端成员 `peer`：同 `conversation_id` 且 `peer.wx_id != m.wx_id` 且 `peer.wx_id NOT IN simWxIds` 且 `peer.deleted_at = 0`

响应 MUST NOT 调用 App 层 author 富化或 recommend 路径。命中时 MUST 返回 `conversationId`、`simWxId`、`peerWxId`、`unreadCount` 及 `messages[]`（最近消息，元素含 `senderWxId`、`content`）。无 eligible 时 MUST 返回 `found=false`（或等价空结果）与 HTTP 200（code=0），MUST NOT 500。

#### Scenario: Sample returns one eligible unread conversation

- **WHEN** 存在 sim 用户 S∈simWxIds 对真人 P 的未读会话，且请求含完整 simWxIds
- **THEN** 响应 MUST 含 `found=true`、`simWxId=S`、`peerWxId=P`、`conversationId>0` 及非空 `messages`

#### Scenario: No eligible unread

- **WHEN** simWxIds 非空但无任何 eligible 未读会话
- **THEN** 响应 MUST 含 `found=false` 且 MUST NOT 返回会话 id

#### Scenario: Invalid secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Sim-sim unread excluded

- **WHEN** 仅存在 sim↔sim 双方均在 simWxIds 内的未读会话
- **THEN** 响应 MUST 为 `found=false`

### Requirement: sim unread sample MUST use bounded ID probe without retry loop

抽样 MUST 在 eligible 集合上对 `ucg_conversation_member.id` 使用 **MIN/MAX + 均匀锚点 + `id >= anchor ORDER BY id LIMIT 1`**（空洞回退 minId），固定 **2 条**有界 SQL。MUST NOT 使用 `ORDER BY RAND()`。MUST NOT 对 `simWxIds` 内单个 wx 循环重试或分页扫描会话列表。

消息加载 MUST 为单条有界查询（`LIMIT messageLimit`），MUST NOT 调用 device HTTP。

#### Scenario: Bounded SQL on eligible subset

- **WHEN** 代码评审 sample 实现
- **THEN** MUST 为 eligible 子集上的 MIN/MAX + LIMIT 1 探测，MUST NOT 全表 `ORDER BY RAND()`

#### Scenario: No cross-domain DAO in ucg

- **WHEN** 代码评审 ucg internal sample 实现
- **THEN** MUST NOT import device 域 DAO；sim 身份过滤 MUST 依赖请求体 `simWxIds`

#### Scenario: Message limit clamp

- **WHEN** 请求 `messageLimit` 为 100
- **THEN** 实际返回 MUST 最多 50 条消息
