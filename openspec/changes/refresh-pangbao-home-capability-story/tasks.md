## 1. 公开引导视频读路径

- [x] 1.1 向负责人确认 `GET /cash/app/api/xiaozhi-mcp/guide-videos`（或公开等价路径）是否计入 usage；按答复更新/不更新 `maintenance_skip`
- [x] 1.2 将引导视频列表 GET 加入 gateway-app Bearer 豁免（或新增公开路径并豁免）；保证匿名与带 Bearer 返回同一 list 语义
- [x] 1.3 自检：cash 反代前缀覆盖、apiregistry/`g.Meta` 若有新 path 则登记

## 2. 官网改版

- [x] 2.1 重写 `pangbao-home.html` 视觉 token（浅色高档、非粉紫玻璃；顶栏不 sticky）
- [x] 2.2 按 design 信息架构重排区块与文案（预测 → 宝宝信息记录 → 语音 AI → 小智）；禁止付费/邀请/社区/诊疗主叙事
- [x] 2.3 宝宝信息记录区：消费 `site/home` 事件 logo 叠放 + 示意统计图
- [x] 2.4 下载改为弹层（Android 二维码/链接 + iOS）；压缩页脚但保留主体/ICP/邮箱/法律链
- [x] 2.5 小智区匿名拉取引导视频并播放；双空隐藏入口
- [x] 2.6 更新 SEO title/description；必要时清洗 `site/home` 默认 Hero 旧社交文案

## 3. 法律页

- [x] 3.1 `privacy-policy.html` 视觉对齐官网（不 sticky、`a.brand` 无下划线）；保留已对齐现网正文不回退
- [x] 3.2 重写 `user-agreement.html` 正文：去注册地址/UGC；微信 unionid；服务内容对齐现网；保留极简付费与 AI 免责；更新修订日期
- [x] 3.3 `user-agreement.html` 视觉对齐官网（同 3.1 规则）

## 4. 验收与部署说明

- [x] 4.1 对照本变更 specs 做页面与匿名 API 走查清单
- [x] 4.2 在实现回复中列出需重部署进程（预期 `gateway-app-server`；cash 仅在有 handler 变更时）
