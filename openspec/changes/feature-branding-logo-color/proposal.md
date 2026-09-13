## Why

开通功能管理页表单仅纵向堆叠，宽屏空白浪费、页过长；功能尚无独立视觉品牌，App 开通中心 / AI 分析 / 弹窗 / 喂养分析与成长轨迹详情只能用主题色，无法像喂养事件 sheet 那样按业务色做玻璃拟态。功能开通未发布，可直接改 v1 契约，无需历史兼容。

## What Changes

- **`feature_def` 增加 `logo` / `color`**（与事件字段同名）：库内 `logo` 存 OSS objectKey（前缀 `feature/`），`color` 存 `#RGB`/`#RRGGBB`。
- **种子默认色**（logo 种子为空，App 无图则占位）：  
  - `prediction_unlock` → `#3B82F6`  
  - `care_alert_smart_remind` → `#0D9488`  
  - `growth_trajectory_predict` → `#EA580C`  
  运维已改色时种子 MUST NOT 覆盖。
- **BREAKING（v1）**：`GET /cash/app/api/feature/catalog` 与 Admin 功能定义读写结构增加 `logo`（HTTP 边界为 CDN URL）、`color`；功能开通未发布，不保留旧字段兼容层。
- **Admin「开通功能管理」**：功能定义表单按语义分组（文案 / 开通策略 / 视觉）并适当横排；视觉区支持 Logo 上传预览与主色选择（对齐事件色板交互）。套餐表单可同步压高度，SKU 不挂视觉字段。
- **Logo 上传**：经 ucg OSS 契约（`feature/` 前缀），cash 经 `clients/ucg` 调用，禁止 import device 业务包。
- **跨仓 Flutter**（`flutter_ai_talk`）：解析 catalog 的 `logo`/`color`；在功能开通界面、AI 分析、功能开通弹窗、喂养记录分析详情、成长轨迹详情中，标题左侧展示功能 logo，页面/对话框玻璃拟态主色随功能 `color`（对齐喂养事件 sheet）。

## Capabilities

### New Capabilities

- `feature-visual-branding`：功能定义视觉字段、种子色、catalog/Admin 下发、OSS 上传与 CDN 映射。
- `feature-admin-semantic-layout`：开通功能管理页语义分组 + 横排布局与视觉编辑控件。
- `feature-branding-client`：Flutter 消费 catalog 视觉字段并在指定界面应用 logo + 主色玻璃拟态。

### Modified Capabilities

- （无强制改写已归档 `openspec/specs/`）行为增量以本变更 capabilities 为准，并与 `commercial-feature-entitlement` / `prediction-default-invite-admin-ux` 的 catalog、Admin 约定对齐。

## Impact

- **进程**：`cash-service`（schema / catalog / Admin）；`ucg-service`（内部上传前缀或等价）；`gateway-app-server`（静态页、反代既有路径）。
- **库**：仅 `ai_voice_cash.feature_def` 增列；EnsureSchema 迁移。
- **API**：`api/v1` `CashFeatureCatalogItem`、`CashAdminFeatureDefItem`、Admin upsert；可选独立 logo 上传 Admin 路径。
- **Redis**：功能定义 catalog 热读投影须含 `logo`/`color`（沿用既有 `cachekit` 键，不新增读缓存产品决策）。
- **跨仓**：`D:\work\flutter_ai_talk` 模型与 UI；本仓库 tasks 列出跨仓清单。
- **非目标**：新建微服务；改开通履约语义；改 usage 统计策略；为 SKU 单独配色；新增 `*_test.go`；VIP 写功能表。
