## REMOVED Requirements

### Requirement: TextChat SHALL 在 voice.task 发布失败时仍完成对话

**Reason**: `voice.task.requested` 发布与 worker consumer 全链路删除；TextChat 不再发布该事件。

**Migration**: 删除 `publishTaskRequested` 及 voice_chat 中相关调用；删除 rabbitmq-init 中 `voice.task.*` binding。

### Requirement: voice.task.requested 成功发布后继续 chat

**Reason**: 同上。

**Migration**: 无替代；审计需求若存在须另开变更。
