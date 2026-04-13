## 1. Implementation
- [x] 1.1 增加 DeepSeek/STT/TTS 的地址、API Key、采样率/位深/声道、HTTP 超时等配置项。
- [x] 1.2 实现语音管线服务：校验 PCM 元数据，必要时转码，调用配置的 STT 转文字，调用 DeepSeek 对话，再用配置的 TTS 合成回复音频。
- [x] 1.3 增加需鉴权的 POST 接口（如 /voice/chat），消费 `application/octet-stream`，读取 `X-Audio-*` 头，调用语音管线，返回 `Content-Type: application/octet-stream` 且带 `Content-Length` 的 PCM 音频。
- [x] 1.4 增加错误处理与日志：涵盖音频格式错误、上游 API 失败/超时；成功时保持二进制返回路径，失败时返回结构化错误。
- [x] 1.5 补充测试（单测+集成：正确流程、非法头、超时）：覆盖转写、DeepSeek 调用、合成，以及接口响应头与长度校验。
- [x] 1.6 调整智能对话音频返回格式为 WAV Base64，并同步配置默认值与文档说明。
- [x] 1.7 运行时打印智能对话 API 返回结果摘要（不输出 audio 内容，仅输出元信息与音频字节长度）。
