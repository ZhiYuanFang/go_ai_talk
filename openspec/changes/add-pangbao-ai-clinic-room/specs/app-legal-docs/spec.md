## ADDED Requirements

### Requirement: 隐私政策 SHALL 披露胖宝 AI 与模型供应商

`resource/public/privacy-policy.html` MUST 修正既有 DashScope/DeepSeek 供应商描述不准确之处。文档 MUST 在 AI 相关章节说明 **胖宝 AI 诊室** 功能：为生成回答，系统会读取用户近 **7 天喂养记录聚合摘要**（非完整原始记录全文），并 MAY 在 App 内展示 AI **思考过程**（thinking）供用户参考。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读胖宝相关说明

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 7 天喂养摘要与 thinking 展示说明
- **AND** DeepSeek/DashScope 等模型供应商描述 SHALL 与后端实际调用一致

### Requirement: 合规文档 URL SHALL 保持不变

gateway-app 暴露的合规文档路径 MUST 仍为 `/privacy-policy.html` 与 `/user-agreement.html`；胖宝 consent 为 App 本地键 `pangbao_ai_consent_v1`，不改变合规 URL。

#### Scenario: 客户端加载隐私政策

- **WHEN** 客户端请求 `/privacy-policy.html`
- **THEN** gateway SHALL 返回含胖宝修订内容的 HTML
- **AND** 路径 SHALL 与修订前相同
