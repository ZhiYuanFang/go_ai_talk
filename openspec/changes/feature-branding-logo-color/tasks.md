## 1. Schema 与种子

- [x] 1.1 `feature_def` EnsureSchema 增加 `logo`、`color` 列（及既有库 ALTER）
- [x] 1.2 种子为三功能写入默认 color（`#3B82F6` / `#0D9488` / `#EA580C`），logo 空；`ON DUPLICATE` 仅空 color 才填充
- [x] 1.3 扩展内存/DAO 行结构与 Admin/Catalog 读写映射含 `logo`/`color`

## 2. CDN、上传与 ucg

- [x] 2.1 新增 feature logo CDN/objectKey helper（`feature/` 前缀；勿 import device）
- [x] 2.2 ucg internal 上传支持 feature 前缀（扩参或等价 path）
- [x] 2.3 `clients/ucg`（或 cash 薄封装）供 cash 上传 logo
- [x] 2.4 Admin `POST` 功能 logo 上传 API：校验类型/大小，返回 objectKey（及可选 CDN URL）
- [x] 2.5 color 校验（`#RGB`/`#RRGGBB`）；非法拒绝保存

## 3. Catalog 与 Admin API

- [x] 3.1 `CashFeatureCatalogItem` / 合成 catalog 下发 CDN `logo` + `color`；defs 缓存投影同步
- [x] 3.2 Admin 功能定义列表/更新契约与实现读写 `logo`/`color`（更新不误清 logo）
- [x] 3.3 controller 装配上传与 upsert；gateway 反代路径确认可达

## 4. Admin 语义布局与视觉控件

- [x] 4.1 `cash-feature-admin.html` 功能定义表单改为文案 / 开通策略 / 视觉分组 + 短字段横排 + 按钮横排
- [x] 4.2 视觉组：Logo 预览、上传、color picker；保存走 Admin API
- [x] 4.3 套餐表单短字段横排；不出现 logo/color
- [x] 4.4 列表载入编辑时回填 logo/color；上传成功刷新预览

## 5. Flutter 跨仓（`D:\work\flutter_ai_talk`）

- [x] 5.1 `FeatureCatalogItem` 解析 `logo`/`color`；增加 `resolveFeatureColor`（空则主题 primary）
- [x] 5.2 功能 logo 小组件（网络图 + 空则占位）
- [x] 5.3 功能开通 hub：标题左侧 logo；卡玻璃用功能 color
- [x] 5.4 功能开通弹窗：标题行 logo；玻璃 accent 用功能 color
- [x] 5.5 AI 分析：喂养/成长入口各自 logo + color
- [x] 5.6 喂养记录分析详情：AppBar logo + 页/卡 accent
- [x] 5.7 成长轨迹详情：AppBar logo + 页渐变/玻璃 accent

## 6. 自检

- [x] 6.1 跑 `hack/check-service-import.ps1`（或 sh）：cash 无 device 业务 import
- [x] 6.2 人工：Admin 改色/上传 → catalog 可见 → Flutter 五处界面着色与 logo
- [x] 6.3 确认未改 usage skip / 履约语义 / 未新增测试文件
