## 1. 官网 HTML 结构与导航

- [x] 1.1 更新 `resource/public/pangbao-home.html` 的 `<title>` 与 `<meta name="description">`，覆盖 0–18 月、喂养记录、AI、同月龄关键词
- [x] 1.2 更新 Nav：新增「产品能力」链至 `#capabilities`，保留「喂养事件」「应用下载」锚点
- [x] 1.3 升级 Hero 区：kicker、H1、副标题、三个 metric 卡片，按 design.md 定稿文案（含静态 fallback，供 API 失败时展示）
- [x] 1.4 新增 `#capabilities` 区块：三列 glass 卡片（智能喂养 / 同月龄轻社交 / AI 智能陪伴），含 AI 免责声明小字；960px 以下单列堆叠

## 2. 既有区块微调与 Footer

- [x] 2.1 微调 `#events` 区块副标题，对齐小月龄高频场景（吃奶、睡觉、换尿布、语音记录）
- [x] 2.2 更新 Footer：保留现有品牌句，新增「隐私政策」「用户协议」链接至 `/privacy-policy.html`、`/user-agreement.html`
- [x] 2.3 确认 AI 文案使用「智能喂养问答」，不在主宣传位单独使用「诊疗」

## 3. 后端 Hero 文案同步

- [x] 3.1 更新 `internal/controller/gateway_app_site.go` 中 `SiteHome` 的 `HeroTitle`、`HeroSubtitle`、`ServiceSummary` 字段值，与 design.md 定稿一致
- [x] 3.2 确认 `GatewayAppSiteHomeRes` 结构未变更（v1 契约不变）

## 4. 验收

- [x] 4.1 本地或测试环境访问 `GET /`，目视确认 Hero、三能力柱、事件区、下载区、Footer 链接与移动端布局
- [x] 4.2 请求 `GET /device/app/api/site/home`，确认 Hero 三字段返回新文案且 `code=0`
- [x] 4.3 对照 `specs/gateway-app-official-site/spec.md` 增量 Requirement/Scenario 逐项自检
- [x] 4.4 确认变更仅影响 `gateway-app-server`，主网关根路径未改动
