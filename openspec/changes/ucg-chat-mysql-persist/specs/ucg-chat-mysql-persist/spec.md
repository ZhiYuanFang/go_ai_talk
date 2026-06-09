## ADDED Requirements

### Requirement: 私信发送 SHALL 同步写入 Redis 与 MySQL outbox

`DeliverChatMessage` 在 Green 审核通过后，SHALL 先执行 Redis `INCR seq` 与 `RPUSH` 消息 JSON 至 `ucg:chat:conv:{conversationId}:msgs`；随后 SHALL 同步 INSERT 一行至 `ucg_chat_message_outbox`（payload 含完整消息字段，status=pending）。Redis 写入失败 SHALL 中止发送；outbox 写入失败 SHALL 记录 Error 日志且 SHALL NOT 阻止 WS `message_delivered` 推送。

#### Scenario: 正常发送双写

- **WHEN** 用户发送私信且 Green 审核通过
- **THEN** Redis LIST 中 SHALL 包含该消息 JSON
- **AND** MySQL `ucg_chat_message_outbox` SHALL 存在对应 pending 行

#### Scenario: Redis 写入失败

- **WHEN** Redis RPUSH 失败
- **THEN** 系统 SHALL NOT 写入 outbox
- **AND** 系统 SHALL NOT 向接收方推送 message_delivered

### Requirement: persist worker SHALL 异步将 outbox flush 至 ucg_chat_message

ucg-service SHALL 运行 `StartChatPersistWorker`，轮询 `ucg_chat_message_outbox` 中 status=pending 或 failed（未超最大 attempts）的记录；SHALL INSERT 至 `ucg_chat_message`（utf8mb4），幂等处理重复 `(conversation_id, id)`；成功后 SHALL 将 outbox 行标记为 done。

#### Scenario: outbox 成功落库

- **WHEN** worker 处理 pending outbox 行且 MySQL 可用
- **THEN** `ucg_chat_message` SHALL 存在与 Redis msg id 一致的记录
- **AND** outbox 行 status SHALL 变为 done

#### Scenario: 重复 flush 幂等

- **WHEN** worker 对同一 outbox 行重试 INSERT
- **THEN** `ucg_chat_message` SHALL NOT 出现重复 `(conversation_id, id)` 行

### Requirement: 读消息 SHALL Redis 优先并在 miss 时回源 MySQL

`listChatMessages` SHALL 优先从 Redis LRANGE 分页返回。当 Redis `LLEN` 为 0 且 MySQL `ucg_chat_message` 对该会话 `COUNT(*) > 0` 时，SHALL 从 MySQL 按 id 升序分页查询并返回等效 `ChatMessage` 结构。

#### Scenario: Redis 有缓存

- **WHEN** Redis LIST 非空
- **THEN** 系统 SHALL 从 Redis 返回消息，SHALL NOT 查询 MySQL

#### Scenario: Redis 空且 MySQL 有历史

- **WHEN** Redis LIST 为空且 MySQL 存在该会话消息
- **THEN** 系统 SHALL 从 MySQL 返回消息列表

### Requirement: MySQL 回源读 SHALL 按需 lazy warm Redis 并对齐 seq

从 MySQL fallback 读取时，系统 SHALL 将读到的消息 JSON RPUSH 回 Redis（至少当前页；可配置整会话阈值）；SHALL 将 `ucg:chat:conv:{id}:seq` 设为 MySQL 该会话 `MAX(id)`（当 Redis seq 缺失或小于该值时）。lazy warm SHOULD 使用短 TTL 分布式锁避免并发重复重建。

#### Scenario: Redis 丢数据后首次读会话

- **WHEN** Redis 消息 LIST 为空但 MySQL 有消息且用户拉取历史
- **THEN** 用户 SHALL 看到 MySQL 中的消息
- **AND** Redis seq SHALL 对齐至 MySQL MAX(id)

### Requirement: ucg_chat_message 表 SHALL 使用 utf8mb4 存储正文

`ucg_chat_message.content` 及表默认字符集 SHALL 为 utf8mb4，以支持 emoji 等四字节 Unicode 入库。

#### Scenario: 含 emoji 的私信持久化

- **WHEN** 用户发送含 emoji 的私信且 worker flush 完成
- **THEN** MySQL 中该消息 content SHALL 原样保留 emoji
