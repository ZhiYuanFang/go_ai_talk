## ADDED Requirements

### Requirement: voiceUnderstanding 不再配置 free 模型

喂养 lane `voiceUnderstanding` 的 Admin 配置与 API MUST 仅包含正式模字段：`provider`、`model`、`maxInFlight`、`maxWaiters`（及既有审计字段）。MUST NOT 再要求或展示 `freeProvider`/`freeModel`。PUT 若仍收到 free 字段，MUST 忽略或拒绝，且 MUST NOT 再将其作为额尽降级选模真相源。

#### Scenario: Admin 保存 VU 无 free 控件

- **WHEN** 运维打开 AI 模型与并发页的 voiceUnderstanding（喂养默认智能体）区块并保存
- **THEN** UI MUST NOT 提供 free 模型下拉；保存体 MUST NOT 依赖 free 字段才能成功

#### Scenario: 硬件喂养只用正式模

- **WHEN** `/voice/chat/ws` 触发喂养 LLM
- **THEN** 选模 MUST 使用 VU 正式 provider/model，MUST NOT 读取 VU free 配置

### Requirement: 额度 Admin 隐藏喂养月度项

Voice AI 额度管理界面 MUST 隐藏喂养（`voice_ai` / `voiceAiMonthlyLimit`）全局默认与 per-user 覆盖的展示与编辑入口。clinic / care-alert 等其他 feature 的额度配置 MUST 保持可用（若该页已有）。

#### Scenario: 额度页无喂养表单项

- **WHEN** 管理员打开 voice-admin 额度功能区
- **THEN** 页面 MUST NOT 展示可编辑的喂养月度额度字段

### Requirement: 喂养对话与 voice_ai 额度解耦

自然语言喂养对话路径 MUST NOT 再以 `voice_ai` 月度 used/limit 作为拒绝或 free 降级条件。`voice_ai` 存储与内部 API MAY 短期保留，但 MUST NOT 作为 chat WS 对话门禁。

#### Scenario: 无 40302 挡喂养 WS

- **WHEN** 某 wx 的 `voice_ai` 已用尽且客户端经 `/voice/chat/ws` 发起喂养对话
- **THEN** 服务端 MUST NOT 因该额度用尽返回阻断对话的 40302（或等价拒绝）；MUST 仍可走 VU 正式模（受并发闸门约束）
