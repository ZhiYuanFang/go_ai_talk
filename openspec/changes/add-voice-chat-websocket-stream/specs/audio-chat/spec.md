## MODIFIED Requirements
### Requirement: 语音对话接口
系统 SHALL 支持 WebSocket 流式语音对话，并在连接阶段要求设备头与音频元数据头；服务端按设备号隔离管理连接与音频接收。

#### Scenario: 设备头合法时建立流式会话
- **WHEN** 客户端连接 `/voice/chat/ws` 并携带合法 `X-Device-No` 与 `X-Audio-*` 头
- **THEN** 服务端建立连接并接收二进制音频帧，返回处理后的音频数据

#### Scenario: 升级成功后按 start/bin/end 协议通信
- **WHEN** 客户端连接 `/voice/chat/ws` 并完成 WebSocket Upgrade
- **THEN** 服务端返回 101 Switching Protocols
- **AND** 客户端发送首条文本 `{"type":"start","deviceNo":"...","sampleRate":16000,"bits":16,"channels":1,"length":...}` 后，服务端开始接收二进制音频分片
- **AND** 客户端发送 `{"type":"end"}` 后，服务端执行收尾处理并返回结果

#### Scenario: 同设备重复连接时替换旧连接
- **WHEN** 相同 `X-Device-No` 再次建立连接
- **THEN** 服务端替换旧连接并仅保留最新连接用于后续音频接收与发送

#### Scenario: 客户端发送显式结束消息触发收尾
- **WHEN** 客户端先发送二进制 PCM 音频帧，再发送文本控制帧 `{"type":"end"}`
- **THEN** 服务端立即结束当前音频片段接收
- **AND** 立即执行 STT、对话与 TTS 收尾流程并返回结果音频

#### Scenario: 服务端返回可读结果与错误
- **WHEN** 收尾处理成功
- **THEN** 服务端连续返回文本消息 `{"type":"audio_chunk","audio":"...","sampleRate":16000}`，并在结束时返回 `{"type":"audio_end","code":0}`（或兼容空 `result.audio`）
- **AND** 当请求状态非法或解析失败时，服务端返回可读 `type=error` 消息而不是直接无提示断连

### Requirement: 语音管线编排
系统 SHALL 在 DeepSeek 调用中支持流式响应解析，并聚合增量文本进入后续 TTS 合成流程。

#### Scenario: DeepSeek 以 SSE 流式返回
- **WHEN** DeepSeek 返回 `data:` 增量消息直到 `[DONE]`
- **THEN** 服务端按顺序聚合内容并继续执行 TTS 合成
