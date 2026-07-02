## Context

胖宝官网由 `gateway-app-server` 在 `GET /` 返回 `resource/public/pangbao-home.html`，动态数据经匿名接口 `GET /device/app/api/site/home` 聚合（事件列表、Android 下载、Hero 文案字段）。v2.0.14 基线 `gateway-app-official-site` 仅要求「母婴喂养定位 + 事件卡片 + 下载区」。

App 已具备三类能力，但官网未展示：

| 能力 | App 实现 | 官网现状 |
|------|---------|---------|
| 智能喂养 | 语音/文本记事件、趋势、家人协同 | 部分（事件卡片） |
| 同月龄轻社交 | UCG：发帖、关注、私信、Feed、AI 润笔 | 无 |
| AI 陪伴 | 语音 AI、智能问答（胖宝诊疗）、成长建议 | 无 |

目标用户：**0–18 月新手妈妈**（高频喂养、强焦虑、手忙、需同路人连接）。

## Goals / Non-Goals

**Goals:**

- 首屏 3 秒内传达「喂养 + 同月龄社交 + AI」三角定位，并锚定 0–18 月新手妈妈。
- 新增 `#capabilities` 三大能力柱（静态 HTML），沿用现有玻璃拟态视觉，移动端单列堆叠。
- 更新 Nav 锚点：`#capabilities`（产品能力）、`#events`（喂养事件）、`#download`（下载）。
- 同步更新 `SiteHome` 接口 Hero 文案字段与 HTML 静态 fallback 一致。
- Footer 增加隐私政策、用户协议链接（已有静态页 `/privacy-policy.html`、`/user-agreement.html`）。
- AI 区块使用「智能喂养问答」等合规措辞，附简短免责声明。

**Non-Goals:**

- 不新增 API 版本或扩展 `GatewayAppSiteHomeRes` 结构（v1 结构不可变）。
- 不在官网嵌入真实 UGC 数据、WebSocket 或 App 截图依赖（若无素材则用文案 + 图标/emoji 示意）。
- 不改动 voice/ucg/history 业务逻辑。
- 不新增 Redis 读缓存或后台循环任务。
- 不修改主网关或其他微服务根路径。

## Decisions

### 1. 实现方式：静态 HTML 能力柱 + API 文案字段

**选择**：三大能力柱文案写死在 `pangbao-home.html`；Hero 三字段由 `gateway_app_site.go` 返回值覆盖。

**理由**：能力柱为稳定营销文案，无需 DB 驱动；Hero 已通过 `SiteHome` 聚合，改字段值即可保持一致，无需 v2 接口。

**备选**：扩展 API 返回 `features[]` ——  rejected，违反 v1 不可改结构约定且过度设计。

### 2. 信息架构与 Nav 顺序

```
Nav: 产品能力 (#capabilities) | 喂养事件 (#events) | 应用下载 (#download)

Hero → #capabilities（三柱）→ #events → #download → Footer
```

**理由**：0–18 月用户先确认「是不是给我用的」，再看差异化（社交/AI），最后看具体事件与下载。

### 3. 定稿文案（Hero + 三柱）

**Kicker：** `0–18 月新手妈 · 记录 · 连接 · 问答`

**Hero Title：** `每一口奶、每一次觉，都有人帮你记着`

**Hero Subtitle：** `宝宝还在小月龄，手在忙、心在慌。胖宝用语音帮你记喂奶和睡眠，连接同阶段妈妈，AI 结合喂养记录按月龄回答你的疑问。`

**Service Summary：** `胖宝面向 0–18 月新手妈妈：智能喂养记录、同月龄轻社交、AI 随时在。`

**Metric 卡片（替换原三卡）：**

| 标题 | 副文案 |
|------|--------|
| 腾不出手也能记 | 喂奶时按住说话，奶量与次数自动记进时间线 |
| 同月龄妈妈懂你 | 不是育儿大群，是和宝宝差不多大的家长，轻轻聊 |
| 问 AI，它看过你的记录 | 结合近 7 天喂养和宝宝月龄，给更贴地的建议 |

**能力柱 #capabilities：**

| 柱 | 标题 | 要点 |
|----|------|------|
| 喂养 | 智能喂养记录 | 语音/文本记事件；吃奶、睡觉、换尿布；家人协同；趋势一目了然 |
| 社交 | 同月龄轻社交 | 发帖分享成长瞬间；关注同路人；私信交流；AI 帮写朋友圈文案 |
| AI | AI 智能陪伴 | 语音喂养助手；智能喂养问答（近 7 天摘要 + 月龄）；每日成长建议 |

**事件区副标题：** `小月龄最高频的照护事件 — 吃奶、睡觉、换尿布，语音一说就记好。`

**AI 免责声明（能力柱或 Footer 小字）：** `AI 回答仅供参考，不能替代医生诊断；如有健康疑虑请及时就医。`

### 4. SEO

- `<title>`：`胖宝 - 0-18月新手妈妈的喂养记录与AI陪伴`
- `<meta name="description">`：含「喂养记录」「同月龄妈妈」「AI 智能问答」「0-18 月」关键词。

### 5. 视觉

- 延续现有 `:root` 粉/蓝/紫渐变与 `.glass` 样式。
- 三柱使用 `.capability-grid` 三列布局，`@media (max-width:960px)` 单列。
- 不引入新 CSS 文件或构建链。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 「0–18 月」可能让更大月龄用户觉得不适用 | Kicker 明确阶段，Hero 仍强调「新手妈妈」；产品内无硬限制 |
| 「胖宝诊疗」对外名称与 App 内不一致 | 官网统一用「智能喂养问答」，App 内名称不变 |
| 静态能力柱与 App 功能迭代不同步 | 文案聚焦稳定核心能力；大改走新 OpenSpec |
| 无 App 截图，转化弱于有图 | 后续可独立变更加截图，本次不阻塞 |

## Migration Plan

1. 部署 `pangbao-home.html` 与 `gateway_app_site.go` 至 `gateway-app-server`。
2. 验证 `GET /`、`GET /device/app/api/site/home` 匿名可访问。
3. 回滚：还原上述两文件即可，无 DB/Redis 迁移。

## Open Questions

- （无阻塞项）若后续获得 Flutter App 截图，可另开变更替换 emoji/纯文案示意。
