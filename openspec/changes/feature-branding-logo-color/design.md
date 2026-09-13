## Context

商业功能开通（catalog / Admin / 履约）已在 `cash-service` 落地，但 `feature_def` 仅有文案与开通策略字段；App 各开通/分析界面统一用主题色。事件域已有成熟模式：`event.logo`（OSS objectKey）+ `event.color`（hex），Admin 色板/上传，HTTP 边界经 `eventlogo` 映射 CDN，Flutter `HistoryEditGlassPanel(eventAccent:)` 玻璃拟态。

本变更在功能域复刻该视觉模型，并优化 `cash-feature-admin.html` 表单布局。功能开通未发布，允许直接扩展 v1 结构。Flutter 在独立仓库 `flutter_ai_talk`，本 change 跨仓交付。

约束：cash 禁止 import device 业务包；跨域上传经 `clients/ucg`；Redis 经 `cachekit`；不新增测试文件；不新增读缓存产品决策（沿用既有 defs 热读键时扩展投影即可）。

## Goals / Non-Goals

**Goals:**

- `feature_def` 持久化 `logo`/`color`；种子写入三功能默认色；logo 默认为空。
- App catalog 与 Admin defs 读写下发视觉字段（logo 为 CDN URL）。
- Admin 功能定义表单语义分组（文案 / 开通策略 / 视觉）并横排压高度；可上传 logo、选主色。
- Flutter 在约定界面应用功能 logo（标题左侧）与主色玻璃拟态。

**Non-Goals:**

- 改变开通/履约/资格判定语义。
- 为 `feature_product` 配置独立视觉。
- 历史 API 双读兼容层。
- 强制种子预置 logo 文件。
- 新建微服务或新增 usage 策略讨论。

## Decisions

### D1：字段名与事件对齐 `logo` / `color`

- **选择**：JSON 与 DB 列均用 `logo`、`color`。
- **理由**：Flutter/运维心智与事件一致；proposal 已拍板。
- **备选**：`primaryColor` / `logoUrl` — 语义更长，放弃。

### D2：直接改 v1，无兼容分支

- **选择**：扩展现有 `CashFeatureCatalogItem` / Admin def 结构体；不建 v2 catalog。
- **理由**：功能开通未发布；用户明确不要求兼容。
- **备选**：v2 并行 — 不必要成本。

### D3：种子默认色 + 空 logo

| featureId | color |
|-----------|-------|
| `prediction_unlock` | `#3B82F6` |
| `care_alert_smart_remind` | `#0D9488` |
| `growth_trajectory_predict` | `#EA580C` |

- INSERT 写入默认色；`ON DUPLICATE KEY UPDATE` 仅当现有 `color` 为空时填充，避免覆盖运维配置。
- `logo` 默认 `''`；App 无 URL 时使用本地占位（运营保证会配齐）。

### D4：上传与 CDN

- **选择**：
  1. 库内存 `feature/...` objectKey（可复用/扩展 `eventlogo` 式 helper，建议新建 `internal/shared/featurelogo` 或把 CDN 工具泛化到 `platform`，避免 cash→device）。
  2. cash Admin 提供独立 multipart 上传（如 `POST /cash/admin/api/feature/defs/logo`），经 `clients/ucg` 调 ucg internal media upload；ucg 侧支持 `feature/` 前缀（扩展现有 internal upload 的 kind/prefix，或新增等价入口）。
  3. 定义保存仍走现有 JSON upsert：`logo` 传 objectKey（或空表示保留/清空需约定：空+未传文件保留原值；显式 clear 可选后期）。
- **理由**：对齐事件「存 key、读 CDN」；与现有 JSON 表单解耦；边界清晰。
- **备选**：整表 multipart upsert — 与当前 Admin JSON 流不一致。群二维码本地落盘 — 不适合多功能 CDN 资源。

### D5：色值校验

- **选择**：与事件相同正则 `#RGB` / `#RRGGBB`（大小写不敏感）；Admin/服务端双端校验。
- **空串**：种子后正常行不应为空；若为空，App 回退主题 primary（防御），不作为产品主路径。

### D6：Admin 布局

```
文案：编号 | 名称 | 简介(全宽)
开通策略：开通方式(全宽) | 天数+条数+主体 | 上架+排序
视觉：Logo 预览+上传 | color picker
按钮：横排 保存/刷新
```

- CSS：分组标题 + `form-grid` 两列；短字段并排；textarea/checks 跨列。
- 套餐区：价格/天数/授予等短字段横排（无视觉块）。

### D7：Flutter 应用面

| 界面 | 行为 |
|------|------|
| 功能开通 hub | 卡标题左侧 logo；卡玻璃用该功能 `color` |
| 功能开通弹窗 | 标题行 logo；`AppModalGlassPanel` / 等价传入 accent |
| AI 分析 hub | 喂养分析卡 ↔ care 功能；成长卡 ↔ growth 功能；各自 logo+色 |
| 喂养记录分析详情 | AppBar 标题左侧 logo；页/卡玻璃用 care 色 |
| 成长轨迹详情 | AppBar logo；页渐变/玻璃用 growth 色 |

- 解析：`FeatureCatalogItem` 增加 `logo`/`color`；`resolveFeatureColor` 对齐 `resolveEventColor`。
- 无 logo：占位图标；运营配置后走网络图（可复用事件 Logo 组件模式）。

### D8：跨仓协作

- OpenSpec 与 Go/Admin 在 `go_ai_talk`；Flutter 任务在同 change `tasks.md` 标明仓库路径，实现时可并行 PR。

## Risks / Trade-offs

- **[Risk] ucg internal upload 仅写死 `event/` 前缀** → 扩展 kind/prefix 或专用 feature 上传；设计评审时确认 ucg `UploadEventLogoObject` 可参数化。
- **[Risk] catalog Redis 缓存未含新字段导致旧投影** → 写路径失效 defs 缓存；读路径结构体同步扩展。
- **[Risk] 默认色与最终品牌不符** → Admin 可随时改；种子不覆盖非空 color。
- **[Risk] Flutter/Go 发布窗口不一致** → 未发布功能，约定同窗口上线；App 对空 logo 已有占位。
- **[Trade-off] 校验逻辑可能与 device 重复** → 接受 cash/platform 轻量复制或抽 shared，禁止业务包互引。

## Migration Plan

1. EnsureSchema：`ALTER feature_def ADD logo VARCHAR(...)`、`ADD color VARCHAR(16)`；跑种子默认色（空才填）。
2. 部署 cash + ucg（上传前缀）+ gateway 静态页。
3. 运维在 Admin 为三功能上传 logo、按需微调色。
4. 发布 Flutter 消费新字段。
5. 回滚：代码回退即可；列可保留（兼容向前）；无需双写。

## Open Questions

- （无阻塞）ucg internal upload 是扩参还是新 path：实现时按最小改动选择，行为以「`feature/` objectKey + CDN」为准。
