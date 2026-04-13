## MODIFIED Requirements
### Requirement: 文字对话接口
系统 SHALL 提供需鉴权的 POST `/text/chat` 接口，接受 JSON 请求体（包含 `deviceNo` 与 `text`），并返回 JSON 回复（包含 `reply`）。

#### Scenario: 合法文本请求返回回复
- **WHEN** 客户端携带有效 `Token`，提交已注册 `deviceNo` 与非空 `text`
- **THEN** 系统返回 200 且响应体包含非空 `reply`

#### Scenario: 文本为空被拒绝
- **WHEN** 客户端提交空 `text`
- **THEN** 系统返回 400，说明参数校验失败

#### Scenario: 未注册设备被拒绝
- **WHEN** 客户端使用未注册的 `deviceNo` 调用 `/text/chat`
- **THEN** 系统返回业务错误，提示先注册设备

### Requirement: 文字对话成功后更新设备最后对话信息
系统 MUST 在 `/text/chat` 每次成功返回回复后，按当前 `deviceNo` 更新 `user.last_talk_time` 为当前时间，并写入 `user.last_talk_ask`（本次提问）与 `user.last_talk_answer`（本次回答）。

#### Scenario: 文字对话成功更新时间
- **WHEN** 已注册设备调用 `/text/chat` 并成功获得回复
- **THEN** 系统更新该设备在 `user` 表的 `last_talk_time`、`last_talk_ask`、`last_talk_answer`

#### Scenario: 文字对话失败不更新时间
- **WHEN** `/text/chat` 调用过程中出现参数校验失败或上游失败
- **THEN** 系统不更新该设备的 `last_talk_time`、`last_talk_ask`、`last_talk_answer`
