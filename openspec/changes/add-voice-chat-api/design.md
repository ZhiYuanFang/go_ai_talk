## Context
- 需要一个语音对话接口：客户端上传原始 PCM（头部给出采样率/位深/声道），先转写成文本，发给 DeepSeek 对话，再把回复文本合成音频，返回 PCM。
- 服务器使用 GoFrame 与统一响应中间件，新流程引入外部 API 与二进制载荷处理。

## Goals / Non-Goals
- Goals：单一 POST 实现 PCM 输入/输出；可配置 STT/TTS/DeepSeek 地址与凭证；保持或必要时规范化音频格式；对格式错误或上游失败有清晰错误路径；安全管理 API Key。
- Non-Goals：实时/流式多轮；长期会话存储；超出提供商默认的说话人分离或语言检测。

## Decisions
- 语音转写：使用 OpenAI Whisper API（可配置地址/Key）；按客户端头接受 PCM，只有在提供商要求时才重采样。
- DeepSeek 对话：调用配置的 DeepSeek Chat Completion，将转写文本作为用户消息，系统提示可由配置提供。
- 语音合成：使用 OpenAI TTS（可配置声音/模型）；尽量输出匹配请求采样率/位深，若提供商受限则转码为 16-bit PCM。
- 接口形态：POST /voice/chat（沿用统一响应中间件），`Content-Type: application/octet-stream`，头含 `X-Audio-Sample-Rate|X-Audio-Bits|X-Audio-Channels`；响应同样为 `application/octet-stream`，带 `Content-Length`，并回传音频格式头。
- 校验与超时：缺失/不支持的头直接拒绝；限制最大音频时长/大小；对转写/对话/合成设置超时与重试，失败返回结构化错误。
- 配置与密钥：增加 DeepSeek/STT/TTS（地址、Key、模型、超时）配置，密钥不入仓库。

## Risks / Trade-offs
- 时延：串行 STT→对话→TTS 可能需数秒；通过合理超时和日志减轻，后续可考虑批处理/并行。
- 音质：提供商采样率限制可能导致重采样；需向客户端明确说明。
- 成本/外部依赖：依赖 Whisper/TTS 与 DeepSeek 可用性；需优雅降级并给出清晰错误。
- 二进制处理：需避免响应包装中间件破坏 PCM；对成功路径保持原始二进制返回。

## Migration Plan
- 先加入配置键，默认安全，不接入管线前无行为变化。
- 新增服务/控制器，必要时用特性开关控制上线。
- 本地与集成测试用样例 PCM 与模拟上游验证后再启用。

## Open Questions
- 是否有指定的 STT/TTS 提供商或音色？（当前假设 Whisper + OpenAI TTS，可配置切换。）
- 接受的最大音频时长与大小？（需设限保护服务。）
- 是否需要多轮上下文（session id），还是每次请求独立？
