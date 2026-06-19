## MODIFIED Requirements

### Requirement: voice-service SHALL 提供胖宝 AI 诊室 WebSocket handler

`voice-service` MUST 在路径 `/voice/clinic/ws` 注册 WebSocket handler（`BindHandler`），处理经 gateway-app 透传而来的客户端文本提问并流式返回 **clinic lane 配置的上游 LLM** 回答（默认种子为智谱 `glm-4.1v-thinking-flash`，Admin 可切回 DeepSeek 等）。用户可见功能名称为 **胖宝诊疗**；实现与配置仍使用 `clinic` / `clinic_ai` / `voice:clinic:*` 命名。该路径为**集群内业务端点**；App 对外入口 MUST 为 gateway-app-server 同路径透传（见 `gateway-ws-edge-proxy`）。实现 MUST NOT 将连接注册到 `VoiceWSManager`。实现 MUST NOT 提供 TTS 或音频上行能力（MVP 纯文本）。

#### Scenario: 经 gateway-app 透传后握手成功

- **WHEN** 客户端经 gateway-app 透传对 voice-service `/voice/clinic/ws` 完成 WebSocket Upgrade
- **THEN** voice-service SHALL 接受连接并等待首帧 `auth` JSON

### Requirement: Clinic SHALL 实施 Redis 限流（per wxId）

系统 MUST 对 clinic 提问路径实施 Redis 限流，键 **`voice:clinic:rate:{wxId}`**。限流计数 **MUST** 在 **`answer_done` 成功** 后递增；**cancelled**、**superseded**、**disconnected** 或失败 turn **MUST NOT** 递增限流计数。处理新 `question` 前 MUST 检查当前窗口计数；超限时 MUST 返回 WS `error` code **42901** 且 MUST NOT 调用 LLM。42901 与 **50301**（LLM lane 队列满）为不同语义：50301 MUST 在额度与 42901 检查通过后、调用上游前返回。

#### Scenario: 短时间频繁提问

- **WHEN** 同一 wxId 在限流窗口内已成功完成（`answer_done`）次数超过阈值
- **THEN** 下一次 `question` SHALL 返回 42901 且 MUST NOT 调用 LLM

#### Scenario: supersede 未完成 turn 不计入限流

- **WHEN** 用户在窗口内多次改问但均未产生 `answer_done`
- **THEN** 限流计数 MUST NOT 因 supersede 而额外递增

#### Scenario: 队列满返回 50301

- **WHEN** 42901 与 clinic_ai 检查通过但 clinic lane 闸门队列满
- **THEN** WS MUST 返回 50301 且 MUST NOT 调用 LLM

## ADDED Requirements

### Requirement: Clinic LLM SHALL 经 clinic lane 与可配置 provider 调用

胖宝 `question` 处理中的 LLM MUST 经 `aimodel.InvokeStream(LaneClinic)` 调用；provider 与 model MUST 来自 Admin/DB profile，MUST NOT 硬编码 DeepSeek。thinking 流式下行语义（`thinking_delta` / `answer_delta`）MUST 保持与现有 WS 协议一致。

#### Scenario: 默认种子为智谱 thinking 模型

- **WHEN** 新部署且 DB 为种子 A 默认值
- **THEN** clinic lane MUST 使用 `provider=zhipu` 且 `model=glm-4.1v-thinking-flash`

#### Scenario: Admin 切回 DeepSeek

- **WHEN** Admin 将 clinic profile 改为 `provider=deepseek`、`model=deepseek-v4-pro`
- **THEN** 下一笔 `question` MUST 调用 DeepSeek 适配器且 MUST 使用 `deepseek-v4-pro` 闸门池
