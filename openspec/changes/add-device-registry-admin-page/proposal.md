# Change: 新增设备注册管理网页与设备活跃时间落库

## Why
- 目前系统没有可视化入口来登记设备号，也无法统一查看设备激活与最近对话时间。
- 现有语音/文字智能对话虽使用 `deviceNo`，但未将设备活跃信息持久化到 `user` 表，无法支撑运营排查与设备状态追踪。

## What Changes
- 新增一个设备管理网页，进入页面前需输入固定口令 `a521521521` 才能访问管理功能。
- 网页支持注册 `deviceNo`：仅校验唯一性，注册成功时写入 `user.device_no` 与 `user.active_time`。
- 网页支持列表查看全部设备，展示 `device_no`、`active_time`、`last_talk_time`。
- 在每次智能对话成功后（文字、语音 HTTP、语音 WebSocket 每次 `start...end` 成功轮次），按当前 `deviceNo` 更新 `user.last_talk_time`。
- 在每次智能对话成功后，额外记录最后一次提问与回答：`last_talk_ask`、`last_talk_answer`。
- 明确设备落库规则：
  - 注册接口创建设备记录。
  - 智能对话仅更新已存在设备记录的 `last_talk_time`/`last_talk_ask`/`last_talk_answer`；若设备未注册则直接拒绝并返回业务错误，提示先注册设备。

## Impact
- Affected specs:
  - device-admin（新增）
  - text-chat（修改：增加对话成功后的 `last_talk_time`、`last_talk_ask`、`last_talk_answer` 持久化要求）
  - audio-chat（修改：增加对话成功后的 `last_talk_time`、`last_talk_ask`、`last_talk_answer` 持久化要求）
- Affected code (planned):
  - 路由与控制器：新增设备管理页面相关路由/API（页面鉴权、注册、列表）
  - `internal/service/voice_chat.go`：在对话成功路径更新设备最近对话时间
  - `internal/dao/user.go` 与相关 `logic`：补充设备注册与时间更新的数据访问封装
  - 数据库迁移：为 `user` 表补充 `last_talk_ask`、`last_talk_answer` 字段
  - `resource/public` 或 `template`：新增管理页面静态资源
  - 测试：新增设备注册/查询/唯一性与对话更新时间回归测试

## Confirmed Scope
- “智能对话”包含 `/voice/chat/ws` 的每次成功轮次（`start...end` 成功结果）。
- 未注册 `deviceNo` 调用智能对话时直接拒绝。
