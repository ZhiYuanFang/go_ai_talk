# Change: 添加语音对话接口

## Why
- 前端需要一个 HTTP 接口，接收原始 PCM 音频，转文字后通过 DeepSeek 对话，再把回复合成为 PCM 音频返回。
- 现有 API 仅覆盖文字/游戏流程，没有语音管线或 DeepSeek 集成。

## What Changes
- 增加一个需要鉴权的 POST 接口，接受 16-bit PCM（`application/octet-stream`）并读取 `X-Audio-*` 元数据头，成功时返回带 `Content-Length` 的 PCM 音频流。
- 新增语音管线：可配置 STT（默认 Whisper API）把 PCM 转文字，把文本发给 DeepSeek，对回复文本用可配置 TTS（默认 OpenAI TTS）合成 PCM。
- 增加配置项：外部 API Key/地址、音频参数（采样率/声道/位深）、超时，以及管线的日志与可选指标。

## Impact
- 影响规格：audio-chat
- 影响代码：新增接口的路由/控制器，STT+DeepSeek+TTS 编排服务模块，外部 API 配置与密钥管理，可选的错误/指标工具。
