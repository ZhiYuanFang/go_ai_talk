## MODIFIED Requirements

### Requirement: 隐私政策 SHALL 披露胖宝诊疗与模型供应商

`resource/public/privacy-policy.html` MUST 修正既有 DashScope/DeepSeek 供应商描述不准确之处。文档 MUST 在 AI 相关章节说明 **胖宝诊疗** 功能（用户可见名称；内部实现仍为 clinic）：为生成回答，系统会读取用户近 **7 天喂养记录聚合摘要**（非完整原始记录全文），并 MAY 在 App 内展示 AI **思考过程**（thinking）供用户参考。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读胖宝诊疗相关说明

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 7 天喂养摘要与 thinking 展示说明，且用户向名称 SHALL 为「胖宝诊疗」
- **AND** DeepSeek/DashScope 等模型供应商描述 SHALL 与后端实际调用一致
