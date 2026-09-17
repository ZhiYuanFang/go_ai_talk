## ADDED Requirements

### Requirement: 预测临近可见推送正文 MUST 使用昵称与 title 模板

当 `voice-service` 在预测临近叫醒路径上准备向绑定用户发送 `bizType=predict_imminent` 的**可见**系统推送时，通知正文（传入推送契约的 `alert`）MUST 为：`{宝宝昵称}要{事件名}了`，其中：

- **事件名** MUST 为该叫醒载荷中的 `title`（trim 后的非空字符串），MUST NOT 由服务端展开事件树选取子事件名，MUST NOT 使用「事件将在约 N 分钟内发生」类时间回落文案作为事件名。
- **宝宝昵称** MUST 在发送前经 device 画像契约按 `deviceNo` 读取；当返回的 `babyName` 为空、或画像读取失败时，昵称 MUST 回落为字面量「宝宝」。
- 各厂商通知**标题**（如「胖宝」）本需求 MUST NOT 要求修改。

#### Scenario: 有昵称与 title

- **WHEN** 画像 `babyName` 为「小宝」且叫醒载荷 `title` 为「换尿布」，且其它推送闸均通过
- **THEN** 发往推送契约的 `alert` MUST 等于「小宝要换尿布了」

#### Scenario: 昵称为空回落

- **WHEN** 画像 `babyName` 为空串（或画像读取失败）且 `title` 为「吃奶」，且其它推送闸均通过
- **THEN** `alert` MUST 等于「宝宝要吃奶了」

#### Scenario: 不使用旧前缀

- **WHEN** 校验通过并发送可见预测临近推送
- **THEN** `alert` MUST NOT 以「宝宝提醒：」开头

### Requirement: 空 title 时 MUST NOT 发送可见推送

当叫醒载荷中的 `title` 经 trim 后为空时，系统 MUST NOT 调用推送发送（MUST NOT 向任何绑定 wx 下发该次可见 `predict_imminent` 通知），MUST 记录可观测日志，且消费路径 MUST 仍按现有约定 Ack（MUST NOT 因空 title 而 Nack requeue 或再次 Publish 延时重试）。

#### Scenario: 缺 title 跳过

- **WHEN** Redis 匹配与其它闸均可通过，但载荷 `title` 为空或仅空白
- **THEN** 系统 MUST NOT 发送该次可见推送，且 MUST Ack 该叫醒消息

#### Scenario: 有 title 才发送

- **WHEN** 载荷 `title` 为非空「睡觉」且其它闸通过
- **THEN** 系统 MUST 按模板拼装正文并尝试向绑定用户发送可见推送
