## Context

- 官网 `pangbao-home.html` 经匿名 `GET /cash/app/api/xiaozhi-mcp/guide-videos` 拉双槽后，用按钮打开弹层播放；`.video-modal video` 选择器未命中实际 DOM，视频按原尺寸撑破窄弹层，用户只能看到局部。
- Explore 已定：`promo` 留在 `#xiaozhi` 现位置并内嵌；官网不展示 `tutorial`；用户面改称「实体音箱连接能力」。
- 隐私政策上一轮对齐仍写「上传头像/配图媒体」「客户端同步预测时间点」「多供应商大模型」；产品事实为头像与成长记录配图仅本地、无成长记录视频、预测披露不暴露实现、AI 对外仅 DashScope。

## Goals / Non-Goals

**Goals:**

- 官网仅内嵌播放已配置的 `promo` 视频，随视口缩放且画面整段可见。
- 官网前端忽略 `tutorial`；不依赖改 cash API。
- 用户面「小智音箱」改为「实体音箱连接能力」等通俗表述（导航/Hero/区块/SEO 对齐）。
- 隐私政策修正本地媒体、预测措辞、仅 DashScope；更新修订日期。

**Non-Goals:**

- 不上移 `#xiaozhi` 到 Hero、不改聚合 `site/home` 契约。
- 不删除 Hub/`promo`/`tutorial` 双槽或改 App 引导视频语义。
- 不改 `user-agreement.html`（除非后续单独要求）。
- 不新增 Redis/后台任务/跨服务契约。

## Decisions

1. **仅前端过滤 tutorial**  
   - 渲染前 `list.filter(x => x.slot === 'promo' && x.videoCdnUrl)`。  
   - 备选：API 对匿名隐藏 tutorial → 拒绝（影响面更大且 App/官网同路径）。

2. **内嵌 video，去掉官网视频弹层**  
   - `#xiaozhi` 右侧（或单列时下方）直接 `<video controls playsinline>`，`width:100%; max-height` + `object-fit:contain`。  
   - 删除或闲置 `#videoModal` 及相关 JS，避免死代码误导。  
   - 未配置 promo 时保留「准备中」提示。

3. **文案替换范围**  
   - 用户可见文案中的「小智音箱」→「实体音箱」/「实体音箱连接能力」；场景说明强调「连接家里的实体音箱、少掏手机问喂养」。  
   - 运维/内部槽名 `promo`/`tutorial`、接口路径不变。

4. **隐私披露策略（推翻上一版部分要求）**  
   - 宝宝头像、成长记录配图：收集节删除上传表述；「不收集」节点名；权限节可保留本机用途并注明不上云。  
   - 成长记录：删除「视频」字样。  
   - 预测：收集节去掉「同步服务端」专条；推送/提醒用「可能发送预测临近提醒」一句带过。  
   - AI：第三节与音频条仅 DashScope；删除智谱/DeepSeek/「以运营配置为准」。

5. **规格增量**  
   - `gateway-app-official-site`：ADDED 内嵌 promo / 不展示 tutorial；MODIFIED 能力叙事中的「小智音箱」用户面叫法。  
   - `privacy-policy-company`：整段 MODIFIED「按现网披露」要求以匹配新策略。

## Risks / Trade-offs

- [Risk] promo 仍在页面中后段，首屏不一定「打开就见」→ 已由产品接受「放现位置」。  
- [Risk] 匿名 API 仍返回 tutorial URL → 可接受；官网不渲染即可。  
- [Risk] 自动播放策略：默认 `controls` 用户点击播放，避免浏览器拦静音自动播与流量争议。  
- [Risk] 竖屏源视频在横宽容器中上下留黑边 → `contain` 保证整段可见，属预期。

## Migration Plan

1. 部署 `gateway-app-server`（或含静态资源的等价发布）。  
2. 验收：有 promo 时内嵌可完整缩放播放；无 tutorial 入口；隐私四类措辞正确。  
3. 回滚：回退上述两个 HTML（及规格变更文档，若已归档则另议）。

## Open Questions

（无；Explore 已关闭。）
