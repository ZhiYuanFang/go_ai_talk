## MODIFIED Requirements
### Requirement: 语音对话接口
系统 SHALL 提供语音智能对话能力（HTTP `/voice/chat` 与 WebSocket `/voice/chat/ws`），并在请求中接收 `deviceNo` 以标识设备会话。

#### Scenario: 合法语音请求返回音频回复
- **WHEN** 客户端提交合法语音请求且 `deviceNo` 已注册
- **THEN** 系统返回对话结果音频

#### Scenario: 未注册设备被拒绝
- **WHEN** 客户端使用未注册 `deviceNo` 调用语音智能对话接口
- **THEN** 系统返回业务错误，提示先注册设备

### Requirement: 语音对话成功后更新设备最后对话信息
系统 MUST 在语音智能对话每次成功完成后，按当前 `deviceNo` 更新 `user.last_talk_time` 为当前时间，并写入 `user.last_talk_ask`（本次提问）与 `user.last_talk_answer`（本次回答）。

#### Scenario: HTTP 语音对话成功更新时间
- **WHEN** 已注册设备调用 `/voice/chat` 并成功获得音频回复
- **THEN** 系统更新该设备在 `user` 表的 `last_talk_time`、`last_talk_ask`、`last_talk_answer`

#### Scenario: WebSocket 语音对话成功更新时间
- **WHEN** 已注册设备通过 `/voice/chat/ws` 完成一次 `start...end` 对话并成功收到结果
- **THEN** 系统更新该设备在 `user` 表的 `last_talk_time`、`last_talk_ask`、`last_talk_answer`

#### Scenario: 语音对话失败不更新时间
- **WHEN** 语音对话因校验失败、上游失败或会话错误未成功完成
- **THEN** 系统不更新该设备的 `last_talk_time`、`last_talk_ask`、`last_talk_answer`
