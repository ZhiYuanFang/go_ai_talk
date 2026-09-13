## ADDED Requirements

### Requirement: 客户端 MUST 从功能目录解析 logo 与 color

Flutter App MUST 在功能 catalog 模型（如 `FeatureCatalogItem`）中解析并保存每项的 `logo`（CDN URL 字符串）与 `color`（hex 字符串）。解析失败或缺字段时，`logo`/`color` MUST 视为空串，MUST NOT 导致目录整表解析失败。

#### Scenario: 解析含视觉字段的目录项

- **WHEN** catalog 返回某项含非空 `logo` 与 `color`
- **THEN** 客户端模型 MUST 暴露对应值供 UI 使用

### Requirement: 指定界面 MUST 在标题左侧展示功能 logo

下列界面在展示对应功能标题时，MUST 在标题左侧展示该功能 logo（网络图）；当 `logo` 为空时 MUST 使用本地占位，MUST NOT 留白导致标题错位：

1. 功能开通界面（各功能卡）
2. AI 分析界面（喂养分析入口与成长轨迹入口各自对应功能）
3. 功能开通弹窗
4. 喂养记录分析详情页
5. 成长轨迹详情页

#### Scenario: 开通中心卡片标题带 logo

- **WHEN** 用户打开功能开通界面且某功能已配置 logo
- **THEN** 该功能卡标题左侧 MUST 显示其 logo

#### Scenario: 无 logo 时占位

- **WHEN** 某功能 `logo` 为空
- **THEN** 上述界面 MUST 显示占位图/图标且标题仍可读

### Requirement: 指定界面 MUST 使用功能 color 作为玻璃拟态主色

上述界面中，对应功能所在的页面或对话框背景/玻璃面板 MUST 使用该功能 `color` 作为 accent（混入方式对齐喂养事件 sheet：`HistoryEditGlassPanel` / `AppModalGlassPanel` / `panelGlassGradient` 等既有 accent 混合）。AI 分析界面若同时展示喂养分析与成长轨迹入口，MUST 分别使用各自功能的 color，MUST NOT 整页强制单一功能色。

当 `color` 为空时，MUST 回退应用主题 primary（或等价），MUST NOT 崩溃。

#### Scenario: 开通弹窗随功能着色

- **WHEN** 用户打开某功能的开通弹窗且该功能 color 为非空 hex
- **THEN** 对话框玻璃拟态 MUST 呈现该主色倾向

#### Scenario: 成长轨迹详情随功能着色

- **WHEN** 用户进入成长轨迹详情页且 `growth_trajectory_predict` 的 color 非空
- **THEN** 页面/玻璃背景 MUST 使用该 color 作为 accent

#### Scenario: AI 分析双卡分色

- **WHEN** 用户打开 AI 分析界面
- **THEN** 喂养分析入口与成长轨迹入口 MUST 分别使用 care 与 growth 功能的 color（若已配置）
