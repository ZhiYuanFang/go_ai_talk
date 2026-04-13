## 1. Implementation
- [x] 1.1 新增设备管理能力的 API/页面路由：页面口令校验、设备注册、设备列表查询。
- [x] 1.2 实现设备注册逻辑：仅校验 `deviceNo` 唯一；注册成功写入 `user.device_no` 与 `user.active_time`。
- [x] 1.3 实现设备列表逻辑：返回全部设备的 `device_no`、`active_time`、`last_talk_time`。
- [x] 1.4 数据库变更：为 `user` 表新增 `last_talk_ask`、`last_talk_answer` 字段。
- [x] 1.5 在智能对话成功路径补充设备活跃信息更新：按 `deviceNo` 更新 `user.last_talk_time`、`user.last_talk_ask`、`user.last_talk_answer`（覆盖文字、语音 HTTP、语音 WS 每次成功轮次）。
- [x] 1.6 对未注册设备发起智能对话增加业务校验：直接拒绝并提示先注册设备。
- [x] 1.7 新增或更新前端页面资源：输入固定口令后可注册设备并查看设备列表。

## 2. Validation
- [x] 2.1 单元测试：设备注册成功、设备号重复失败、设备列表返回字段完整。
- [x] 2.2 集成/接口测试：已注册设备完成文字对话后 `last_talk_time`、`last_talk_ask`、`last_talk_answer` 被更新。
- [x] 2.3 集成/接口测试：已注册设备完成语音 HTTP 与语音 WS 每次成功轮次后 `last_talk_time`、`last_talk_ask`、`last_talk_answer` 被更新。
- [x] 2.4 负向测试：未注册设备调用智能对话被直接拒绝且不写 `last_talk_time`/`last_talk_ask`/`last_talk_answer`。
- [x] 2.5 回归测试：既有文字/语音对话主流程响应结构与成功语义保持兼容。
