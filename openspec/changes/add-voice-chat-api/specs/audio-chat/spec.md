## ADDED Requirements
### Requirement: 语音对话接口
系统 SHALL 提供需鉴权的 POST /voice/chat 接口，接受 `application/octet-stream` 的原始 PCM 音频，并要求 `X-Audio-Sample-Rate`、`X-Audio-Bits`、`X-Audio-Channels` 头；成功时返回带 `Content-Length` 且含同类元数据头的 PCM 回复音频。

#### Scenario: 合法 PCM 请求返回音频
- **WHEN** 客户端提交 16-bit PCM 且采样率/声道与头信息一致，并提供有效 `Token`
- **THEN** 服务器返回 200，`Content-Type: application/octet-stream`，`Content-Length` 与 PCM 长度一致，`X-Audio-*` 反映回复格式

### Requirement: 语音管线编排
系统 SHALL 将输入 PCM 通过配置的 STT 提供商转为文本，把文本作为用户消息提交给 DeepSeek，对返回文本用配置的 TTS 提供商合成 WAV 音频，并以 Base64 字符串回传客户端。

#### Scenario: 对话回复被合成
- **WHEN** 转写成功且 DeepSeek 返回文本回复
- **THEN** 系统将回复文本合成为 WAV，并将 WAV 字节进行 Base64 编码后返回

#### Scenario: 上游失败被处理
- **WHEN** 转写、DeepSeek 或合成失败/超时
- **THEN** 系统返回非 200，附带结构化 JSON 错误，说明阶段与原因，且不输出部分 PCM 数据

### Requirement: 音频校验与限制
系统 SHALL 校验音频元数据（位深需支持 16-bit），限制最大载荷大小/时长，对缺失或不支持的音频头在调用外部服务前直接拒绝。

#### Scenario: 非法音频元数据被拒绝
- **WHEN** 请求缺少必需的 `X-Audio-*` 头或位深不受支持
- **THEN** 系统返回 400，说明校验失败，不调用外部服务
