# Change: 增加按设备号的临时连续对话能力

## Why
- 当前语音对话在 DeepSeek 调用阶段是单轮请求，无法利用同一设备的历史上下文。
- 业务需要在智能对话过程中按设备号维持最近对话内容，以提升连续对话体验。
- 该上下文仅需临时缓存；服务重启后丢失可接受。

## What Changes
- 在语音对话请求中增加 `deviceNo`（设备号）入参，作为会话缓存键。
- 增加进程内会话缓存：按 `deviceNo` 存储最近 N 轮问答，并设置 TTL 过期时间。
- 调用 DeepSeek 时，按顺序拼接「系统提示 + 设备历史消息 + 当前用户消息」，实现连续对话。
- 约束缓存边界（最大轮数、TTL、可选最大设备会话数）并提供清理策略，避免内存无界增长。
- 明确失败语义：仅在 DeepSeek 成功返回后写入完整问答轮次，失败请求不污染会话历史。

## Impact
- Affected specs: audio-chat
- Affected code:
  - `api/v1/voice_chat.go`（请求结构新增 `deviceNo`）
  - `internal/controller/voice_chat.go`（透传设备号）
  - `internal/service/voice_chat.go`（会话缓存与 DeepSeek 消息构造）
  - `internal/service/voice_chat_test.go`（新增多轮、TTL、边界与重启语义相关测试）
  - 配置文件（新增会话缓存参数）
