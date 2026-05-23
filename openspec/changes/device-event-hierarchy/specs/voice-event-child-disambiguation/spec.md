## ADDED Requirements

### Requirement: 事件匹配时子节点优先于父节点

voice-service 在本地文本匹配事件时，SHALL 对候选事件按**深度降序、名称长度降序**排序后再匹配，使叶子或深层子节点优先于浅层父节点命中。

#### Scenario: 同时提及父名与子名时命中叶子

- **WHEN** 事件树含「换尿布(父)」与「大便(子)」
- **AND** 用户说「换尿布，拉了大便」
- **THEN** 匹配结果 SHALL 为「大便」事件 id
- **AND** SHALL NOT 进入父节点追问流程

### Requirement: 命中非叶子事件时必须追问且不写 history

当匹配到的事件存在子节点（`EXISTS parent_id = 该 id`）时，voice SHALL NOT 调用 `AddHistory` / 等价写库；SHALL 设置该设备的 pending 子事件上下文；SHALL 回复列出**直接子节点名称**的选择问句；且 SHALL 令 `finishTalk=false` 等待用户下一轮输入。

#### Scenario: 仅说换尿布时追问

- **WHEN** 「换尿布」有子节点「大便」「小便」
- **AND** 用户说「换尿布了」且命中「换尿布」
- **THEN** 系统 SHALL NOT 写入 history
- **AND** 回复 SHALL 引导用户在「大便」「小便」中选择（语义等价即可）
- **AND** `finishTalk` SHALL 为 false

#### Scenario: 三级树下追问直接子节点

- **WHEN** 「换尿布」子节点含「排泄类」，「排泄类」下含「大便」「小便」
- **AND** 用户仅命中「换尿布」
- **THEN** 第一轮追问 SHALL 仅针对「换尿布」的直接子节点（如「排泄类」与其它同级子名）
- **AND** SHALL NOT 在第一轮直接询问「大便还是小便」

### Requirement: pending 期间仅在当前父的直接子节点中匹配

存在 pending 子事件上下文时，voice SHALL 仅在 `pending.ParentEventId` 的**直接**子节点集合中执行文本匹配；命中仍为非叶子则 SHALL 更新 pending 并继续追问；命中叶子则 SHALL 清除 pending 并按原动作类型写 history。

#### Scenario: 第二轮回答大便后落库

- **WHEN** pending 父为「排泄类」且子含「大便」「小便」
- **AND** 用户第二轮说「大便」
- **THEN** 系统 SHALL 清除 pending
- **AND** SHALL 以「大便」叶子 event id 写入 history（在动作流程允许写库时）

### Requirement: pending 为内存态且不跨会话恢复

pending 子事件上下文 SHALL 存储于 voice 进程内存（按 deviceNo 键）；SHALL NOT 写入 Redis 或与 session 同步持久化；会话 TTL 或进程重启后 pending 丢失时，后续输入 SHALL 按无 pending 的新轮次处理。

#### Scenario: 超时后大便按新对话处理

- **WHEN** 用户第一轮触发「换尿布」pending 后长时间无后续
- **AND** pending 已因会话过期或重启而丢失
- **AND** 用户再说「大便」
- **THEN** 系统 SHALL NOT 假定仍在「换尿布」追问上下文中
- **AND** MAY 将「大便」作为独立话术在全树中匹配

### Requirement: 仅叶子事件 id 可写入 history

voice 写入 `history.event_id` 时，所选事件 MUST 为无子节点的叶子；非叶子 id SHALL NOT 作为新 history 行的 event_id。

#### Scenario: 追问完成前无 history 行

- **WHEN** 用户仅命中非叶子「换尿布」且处于追问态
- **THEN** 在该轮及 pending 未清除前 SHALL NOT 产生以「换尿布」为 event_id 的新 history 行
