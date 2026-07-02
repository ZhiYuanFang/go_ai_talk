## Why

当前胖宝官网（`pangbao-home.html`）仅展示「母婴喂养 + 事件卡片 + 应用下载」，目标用户 0–18 月新手妈妈无法从首屏感知 App 已具备的**同月龄轻社交**与 **AI 喂养陪伴**能力。产品与隐私政策/用户协议中已有完整能力描述，但官网叙事仍停留在「喂养工具」，导致转化与品牌定位不完整。

## What Changes

- 升级官网 Hero 与导航，明确面向 **0–18 月新手妈妈** 的「记录 · 连接 · 问答」三角定位。
- 新增**三大能力柱**静态区块：智能喂养、同月龄轻社交、AI 智能陪伴（含语音记事件、社区、智能问答、成长建议、AI 润笔等场景化文案）。
- 保留并微调现有「喂养事件卡片」与「应用下载」区块，事件区副文案对齐小月龄高频场景。
- 更新 `GET /device/app/api/site/home` 聚合接口中的 Hero 文案字段（`heroTitle`、`heroSubtitle`、`serviceSummary`），使 API 与页面首屏一致；**不修改 v1 响应结构**，仅更新字段值。
- 补充 Footer 链接至隐私政策与用户协议；更新 `<title>` 与 `<meta description>` 以覆盖 AI、社区、0–18 月关键词。
- AI 对外表述采用「智能喂养问答」等合规措辞，避免「诊疗」医疗暗示；附简短免责声明。

## Capabilities

### New Capabilities

（无新增 capability；本次为既有官网能力的增量展示与文案扩展。）

### Modified Capabilities

- `gateway-app-official-site`：扩展官网信息架构与文案要求，新增三大能力柱、0–18 月目标用户定位、AI/轻社交展示、Footer 法律页链接及 SEO meta 要求；更新 Hero 品牌定位 Scenario 的验收语义。

## Impact

- **前端静态页**：`resource/public/pangbao-home.html`（新增 section、Nav 锚点、样式、Footer 链接）。
- **后端聚合**：`internal/controller/gateway_app_site.go`（Hero 文案字段值更新）。
- **API 契约**：`api/v1/gateway_app_http.go` 中 `GatewayAppSiteHomeRes` 结构**不变**；无新版本接口。
- **进程边界**：变更仅作用于 `gateway-app-server` 根路径 `/` 与 `/device/app/api/site/home`；不涉及 voice/ucg/history 业务逻辑。
- **App API 使用统计**：`/device/app/api/site/home` 为既有匿名只读接口，文案变更不计入新接口；**无需** usage 统计确认。
- **Redis / 数据库**：无新增读缓存、无新增 DB 操作。
- **Spec 依据**：基于 v2.0.14 基线 `gateway-app-official-site`，增量见本变更 `specs/gateway-app-official-site/spec.md`。
