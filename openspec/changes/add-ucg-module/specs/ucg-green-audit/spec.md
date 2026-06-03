## ADDED Requirements

### Requirement: Post and profile content SHALL use Green async audit Option A

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter `pending_audit` visibility: ONLY author MAY see content in feeds/profile until Green pass sets `published` or profile active state; on fail content SHALL be `rejected` with reason visible to author in 我的动态 or profile edit feedback as「违规已下架」.

#### Scenario: 提交后仅作者可见
- **WHEN** 用户发布帖子且 Green 未完成
- **THEN** 其他用户请求 Feed SHALL NOT 包含该帖；作者 我的动态 SHALL 显示审核中

#### Scenario: 审核通过公开
- **WHEN** Green 返回 pass
- **THEN** post status SHALL 变为 2 且 SHALL 出现在推荐/关注 Feed

#### Scenario: 审核失败
- **WHEN** Green 返回 fail
- **THEN** post status SHALL 变为 3，作者 SHALL 见 reject_reason；其他用户 SHALL NOT 见该帖

### Requirement: Chat messages SHALL use Green audit Option C before delivery

Chat messages SHALL be visible as pending to sender only until Green pass; on pass message MUST be delivered to recipient via WS; on fail sender MUST receive failure notification and recipient MUST NOT receive message.

#### Scenario: 发送后收件人不可见
- **WHEN** 用户发送聊天消息且 Green 未完成
- **THEN** 收件人 WS SHALL NOT 收到该消息

#### Scenario: 审核通过后投递
- **WHEN** Green pass
- **THEN** 收件人 SHALL 通过 WS 收到 `message_delivered` 事件

#### Scenario: 审核失败
- **WHEN** Green fail
- **THEN** 发送方 SHALL 收到 `audit_failed` 含 reason；消息 SHALL NOT 进入收件人会话
