## MODIFIED Requirements
### Requirement: 语音管线编排
系统 SHALL 在语音转写后，基于请求中的 `deviceNo` 组装 DeepSeek 多轮消息上下文：按顺序包含系统提示（如配置）、该设备未过期历史消息以及当前用户消息；DeepSeek 成功返回后再进入 TTS 合成并返回音频。

#### Scenario: 新设备首次请求按单轮处理并建立会话
- **WHEN** `deviceNo` 首次发起语音对话且缓存中无可用历史
- **THEN** 系统仅发送系统提示（如有）与当前用户消息给 DeepSeek
- **AND** 在 DeepSeek 成功返回后，写入该设备本轮 user/assistant 消息

#### Scenario: 同设备连续请求携带历史上下文
- **WHEN** 同一 `deviceNo` 在 TTL 内再次发起语音对话
- **THEN** 系统在 DeepSeek 请求中包含该设备最近 N 轮历史消息与当前用户消息
- **AND** 返回成功后更新并裁剪历史，仅保留最近 N 轮

#### Scenario: DeepSeek 失败不污染会话历史
- **WHEN** DeepSeek 调用超时或返回错误
- **THEN** 系统返回结构化错误
- **AND** 不写入本轮不完整消息到该设备会话缓存

#### Scenario: 识别文本长度不足时直接拦截
- **WHEN** 语音转写得到的文本长度小于配置阈值（默认 2）
- **THEN** 系统直接返回结构化错误
- **AND** 不调用 DeepSeek

## ADDED Requirements
### Requirement: 设备级临时会话缓存
系统 SHALL 提供进程内设备级会话缓存，以 `deviceNo` 为键保存最近问答历史，并通过 TTL 与容量上限控制内存；该缓存不做持久化，服务重启后丢失属于预期行为。

#### Scenario: 会话历史过期后重置
- **WHEN** 设备会话超过配置 TTL（默认 30 分钟）未活跃
- **THEN** 系统在后续请求中将其视为新会话，不再使用过期历史

#### Scenario: 服务重启后历史丢失
- **WHEN** 服务实例重启
- **THEN** 所有设备会话缓存被清空
- **AND** 后续请求按新会话开始处理

#### Scenario: 达到容量上限触发淘汰
- **WHEN** 活跃设备会话数超过配置上限（如 maxDeviceSessions）
- **THEN** 系统按最久未活跃优先淘汰旧会话
- **AND** 保证新增设备请求可继续建立会话
