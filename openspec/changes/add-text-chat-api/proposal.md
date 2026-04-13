# Change: 增加文字输入连续对话接口

## Why
- 当前仅提供 `/voice/chat`（语音输入）对话能力；前端/设备侧需要在无语音场景下直接提交文本并获取回复。
- 文字输入对话需要复用现有“智能对话”能力（DeepSeek + `deviceNo` 会话缓存），保持跨语音/文字的连续对话体验。

## What Changes
- 新增需鉴权的 POST `/text/chat` 接口，接受 JSON 请求体（含 `deviceNo` 与 `text`）。
- 接口调用与 `/voice/chat` 相同的 DeepSeek 对话逻辑与设备会话缓存（按 `deviceNo`、N 轮 + TTL、进程内临时缓存），实现连续对话。
- 成功时返回 JSON：仅包含 `reply` 文本。
- 失败时沿用现有错误语义：返回结构化错误（并保证失败不写入不完整会话历史）。

## Impact
- Affected specs:
  - text-chat（新增）
  - audio-chat（行为关联：同一 deviceNo 跨端共享历史上下文）
- Affected code (planned):
  - `api/v1/text_chat.go`（新增请求/响应结构）
  - `internal/controller/text_chat.go`（新增控制器）
  - `internal/service/voice_chat.go`（复用/抽取 DeepSeek 对话与会话缓存能力，供文字接口调用）
  - 路由注册（将 `/text/chat` 挂载到现有鉴权中间件 `Token`）
  - 测试（新增文字接口与跨端共享历史场景测试）
