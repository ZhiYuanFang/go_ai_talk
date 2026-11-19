## Why

用户协议、隐私政策与官网仍大量使用「语音 AI 问答 / AI 生成或分析 / 不能替代就医」等表述，易被理解为 AIGC 内容生成与健康建议输出，触发不必要的 AIGC 合规审查；产品对外叙事应定位为成长记录、预测提醒与语音记事/查询工具，而非生成式回答。

## What Changes

- **用户协议**（`user-agreement.html`）：服务内容去掉「语音 AI 问答 / 智能服务 / 小智」等生成感措辞，改为记录、预测提醒、语音录入与查询、实体音箱连接；**删除整节「AI 服务免责」**（含就医/诊断类免责），后续章节顺延编号；更新修订日期。
- **隐私政策**（`privacy-policy.html`）：音频与第三方处理去掉「生成回复 / 语音 AI / 医疗免责」；云端处理**不点名** DashScope，仅写「经具备相应资质的云端语音识别服务处理」；「第三方 AI 服务」改写或并入非 AI 表述的技术服务披露；与 `inline-home-promo-and-privacy-wording` 已落地的本地媒体/预测措辞保持一致。
- **官网**（`pangbao-home.html`）：「语音 AI 问答」区块改为语音记事与查询叙事；去掉「参考作答 / 按月龄建议 / AI 回答不能替代医生诊断」等文案；导航、Hero、meta 对齐；保留预测提醒与实体音箱连接能力（不写成 AIGC）。
- **规格**：修订 `user-agreement-company`、`privacy-policy-company`、`gateway-app-official-site`（含取消「必须展示 AI/就医免责」类要求）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `user-agreement-company`：去 AIGC/AI 免责节；服务内容工具化表述。
- `privacy-policy-company`：语音处理不点名供应商；去掉 AI 生成回复与医疗免责表述。
- `gateway-app-official-site`：官网去「语音 AI 问答」与就医免责；改为语音记事/查询叙事。

## Impact

- **静态页**：`resource/public/user-agreement.html`、`privacy-policy.html`、`pangbao-home.html`（经 `gateway-app-server`）。
- **不必改**：后端 ASR/对话实现、cash API、Hub；工程仍可用 DashScope，仅对外文案不点名。
- **部署**：`gateway-app-server`（静态资源）；无需改 `.env`。
