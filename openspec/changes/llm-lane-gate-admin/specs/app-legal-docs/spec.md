## MODIFIED Requirements

### Requirement: 隐私政策 SHALL 披露胖宝 AI 与模型供应商

`resource/public/privacy-policy.html` MUST 说明喂养语音 AI、胖宝诊疗与润笔等功能可能将用户输入或摘要发送至 **可配置的第三方大模型服务**（包括但不限于智谱 GLM、DeepSeek、阿里云 DashScope，以运维后台实际配置为准）。文档 MUST 在 AI 相关章节说明 **胖宝诊疗**：为生成回答，系统会读取用户近 **7 天喂养记录聚合摘要**（非完整原始记录全文），并 MAY 在 App 内展示 AI **思考过程**（thinking）供用户参考。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读胖宝诊疗相关说明

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 7 天喂养摘要与 thinking 展示说明，且用户向名称 SHALL 为「胖宝诊疗」

#### Scenario: 供应商描述与可配置 provider 一致

- **WHEN** 系统默认使用智谱模型且 Admin 可切回 DeepSeek/DashScope
- **THEN** 隐私政策 SHALL 表述为可配置多供应商，且 MUST NOT 声称仅使用单一固定供应商
