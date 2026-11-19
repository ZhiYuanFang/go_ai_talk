## Context

- Explore 已定：对外不做 AIGC 内容生成叙事，也不输出健康建议类免责（产品不宣称健康内容输出）。
- 范围覆盖 `user-agreement.html`、`privacy-policy.html`、`pangbao-home.html`。
- 云端语音识别供应商对外**不点名**，统一「经具备相应资质的云端语音识别服务处理」；工程实现仍可使用 DashScope，本变更只改静态文案。
- 关联未归档 change `inline-home-promo-and-privacy-wording` 已改本地媒体/预测/实体音箱；本变更在其上继续去 AI/就医措辞。

## Goals / Non-Goals

**Goals:**

- 三页用户可见文案不再呈现「AI 问答 / 生成回答 / AIGC / 不能替代就医」等信号。
- 能力叙事改为：成长记录、基于历史的预测与提醒、语音记事与查询、实体音箱连接。
- 隐私仍披露麦克风与云端语音识别处理事实，但不点名厂商、不写「生成回复」。

**Non-Goals:**

- 不改 ASR/LLM 后端、模型配置或 App 功能开关。
- 不删除「喂养预测与提醒」产品能力（可保留「预测」，避免「AI 预测」「生成」）。
- 不改 Flutter 应用内文案（若有同类措辞，另案处理）。

## Decisions

1. **用户协议删除整节「AI 服务免责」**  
   - 不改写成「健康免责」（产品明确无健康内容输出）。  
   - §6/§7 顺延为 §5/§6。

2. **隐私：匿名化云端语音识别**  
   - 音频条与第三方列表：删除「阿里云 DashScope」「语音 AI」「生成回复」「不能替代医疗诊断」。  
   - 统一：「经具备相应资质的云端语音识别服务处理」。  
   - 「第三方 AI 服务」标题改为「第三方技术服务」或并入既有第三方列表，避免「AI」字样。

3. **官网 `#ai` 改叙事**  
   - 标题/导航：如「语音记事」或「语音查询」（实现时选一，全文一致）。  
   - 三步闭环：开口说 → 写入时间线 → 口语查记录；删除「参考作答 / 建议 / 就医」。  
   - 删除 `.ai-disclaimer` 就医免责句。  
   - Hero/meta 同步去掉「语音 AI」「作答」。

4. **规格**  
   - `user-agreement-company`：REMOVED「含 AI 免责」旧要求，ADDED 无 AIGC/健康免责的条款要求。  
   - `privacy-policy-company`：MODIFIED 披露范围（不点名、无生成回复/医疗免责）。  
   - `gateway-app-official-site`：REMOVED「必须展示 AI/就医免责」；MODIFIED 能力柱为语音记事/查询。

## Risks / Trade-offs

- [Risk] 隐私不点名供应商，透明度略降 → 用「具备相应资质的云端语音识别」保留处理事实。  
- [Risk] 工程仍接大模型，文案与实现不完全同词 → 对外按工具叙事，属产品披露策略，不在本变更改实现。  
- [Risk] 官网仍有「智能预测」含「智能」→ 可接受；MUST NOT 写成 AI 生成建议。

## Migration Plan

1. 改三份 HTML，更新修订日期。  
2. 部署 `gateway-app-server`。  
3. 回滚：还原三页静态文件。

## Open Questions

（无；DashScope 不点名已确认。）
