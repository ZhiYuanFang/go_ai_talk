## Context

- 官网由 `gateway-app-server` 根路径托管 `resource/public/pangbao-home.html`；匿名聚合 `GET /device/app/api/site/home` 已提供事件列表与下载信息。
- 现网客户端已关闭 UCG 入口，诊疗/值得留意等已退役；隐私政策正文已在 `align-privacy-policy-to-client-surface` 方向对齐，但视觉仍为粉紫玻璃，且 `user-agreement.html` 未改。
- 小智引导视频：`GET /cash/app/api/xiaozhi-mcp/guide-videos` 现要求 Bearer，规格明确禁止匿名；官网无法直接调用。

## Goals / Non-Goals

**Goals**

- 按现网能力重写官网叙事与区块顺序，视觉升级为浅色克制高档风。
- 下载入口改为弹层；页脚压缩但保留主体/ICP/邮箱/法律链。
- 官网匿名可读小智已配置引导视频。
- 隐私政策与用户协议视觉对齐官网；用户协议内容对齐隐私/现网；协议保留极简付费条款，官网不写付费。

**Non-Goals**

- 不改 Flutter 法律页 URL；不重开 UCG/诊疗产品面。
- 不把付费/邀请写进官网营销文案。
- 不新增测试文件；不默认引入新 Redis。
- 不在本变更内完成 `align-privacy-policy-to-client-surface` 归档（可并行，正文避免互相覆盖）。

## Decisions

### D1. 官网信息架构（区块顺序）

```
Hero（品牌 + 一句价值 + CTA「下载」）
  → 智能预测与提醒
  → 宝宝信息记录（事件 logo 叠放 + 示意趋势图；事件数据仍来自 site/home）
  → 语音 AI 问答闭环
  → 小智音箱（场景文案 + 引导视频播放）
  → Footer（压缩）
```

- 导航 CTA 打开下载弹层（Android 二维码/链接 + iOS App Store），页内不再占大段双列下载区。
- **禁止**：付费、会员、邀请开通、同月龄社区主叙事、诊疗主宣传、「值得留意」等退役能力。

### D2. 视觉系统（三页共用 token）

- 浅色底 + 克制渐变/纹理；避免粉紫玻璃英雄感、默认紫系、大圆角 pill 堆叠。
- 顶栏：**不 sticky**；品牌链选择器为 `a.brand { text-decoration: none }`。
- 法律页：同色板与顶栏规则；正文以可读为主，少装饰。
- 实现：三页内联 CSS 对齐同一套 CSS 变量即可（可不抽独立 css 文件，避免额外静态路径约定）；若抽公共片段，MUST 仍由 gateway-app 可静态服务。

### D3. 小智引导视频公开读（方案 A）

优先：**将既有** `GET /cash/app/api/xiaozhi-mcp/guide-videos` **加入 gateway-app Bearer GET 豁免**，响应体与 App 一致（仅已配置槽 + CDN URL）。Flutter 带 Bearer 仍可用；官网无 Token 也可调。

备选（仅当豁免同路径有合规顾虑）：新增精确公开路径（如 `/cash/app/api/xiaozhi-mcp/guide-videos/public`）并豁免，handler 复用同一 list 逻辑。

- 反代：已落在 `/cash/app/api/*`，无需新 Bind。
- **Usage**：按仓库约定，实现前 MUST 向负责人确认是否计入 usage；未答复前不得改 `maintenance_skip.go` 假定策略。
- cash Admin 上传/改标题路径不变。

### D4. site/home 与 Hero 文案

- 事件 logo 列表继续用 `site/home`；宝宝信息记录区用返回的 events 做叠放展示。
- 若聚合接口仍返回含「轻社交」等旧 Hero 字段：前端硬编码新文案优先，或同步清洗 gateway 默认 Hero 字符串（最小改动二选一，实现时选「前端硬编码为主、后端默认串顺手清洗」）。

### D5. 用户协议内容边界

| 项 | 处理 |
|----|------|
| 注册地址 | 删除 |
| 微信头像/昵称/OpenID | 改为 unionid；不收集头像/昵称 |
| UGC 节与社区 IP 授权 | 删除 |
| 服务内容 | 记录/统计可视化/预测提醒/语音 AI/小智等 |
| 付费 | **保留极简**（App 内开通以展示与渠道规则为准） |
| AI 免责 | 保留 |
| 修订日期 | 更新 |

隐私政策：以视觉对齐为主；正文若与 `align-privacy-policy-to-client-surface` 已落地版本一致则不无谓回退。

### D6. 页脚

- 压缩行高与块数；仍 MUST 含：公司中英文全称、`pangbao@cuplay.top`、浙ICP备2024101366号（链工信部）、隐私/协议链接、© 年份。
- MUST NOT 恢复运维旁白类文案。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 公开 guide-videos 被刷流量/CDN | 仅元数据+CDN URL；视频本身走 CDN；可后续加边缘限流 |
| Bearer 豁免改变「未登录拒绝」验收 | 规格 MODIFIED；Flutter 行为不变 |
| 官网改版与未归档隐私变更冲突 | apply 时 diff 三页，保留已对齐隐私正文 |
| 视觉「高档」主观 | design 固定禁区（粉紫玻璃/社区叙事），实现对照截图自检 |

## Migration Plan

1. 实现公开读 + 豁免（及 usage 确认结果）。
2. 重写官网 HTML/CSS/JS；接 guide-videos。
3. 改用户协议正文 + 两法律页换皮。
4. 重部署 `gateway-app-server`；若 cash 无代码变更则不必重部署 cash；若新增独立公开 handler 则部署 `cash-service`。

## Open Questions

（无）负责人已确认：公开 `guide-videos` **不计入** usage，已写入 `maintenance_skip`。
