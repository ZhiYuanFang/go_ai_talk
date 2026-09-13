## ADDED Requirements

### Requirement: 开通功能管理页 MUST 按语义分组展示功能定义表单

`cash-feature-admin.html` 的功能定义编辑区 MUST 分为至少三个语义分组：**文案**、**开通策略**、**视觉**。文案组 MUST 含功能编号（只读）、显示名称、简介。开通策略组 MUST 含开通方式、邀请/广告授予天数、默认开通条数、开放主体、是否上架、排序。视觉组 MUST 含 Logo 预览/上传与主色选择。

#### Scenario: 分组标题可见

- **WHEN** 管理员打开开通功能管理页并载入某功能
- **THEN** 表单 MUST 能区分文案、开通策略、视觉三组字段

### Requirement: 功能定义表单 MUST 适当横向排列以降低页长

短输入字段（如天数、条数、主体、上架、排序、主色与 Logo 控件）MUST 在宽屏下横向排列（两列或等价 grid），MUST NOT 将全部字段单列纵向堆叠。简介与开通方式多选可跨列全宽。保存/刷新按钮 MUST 横向排列。

#### Scenario: 宽屏下短字段并排

- **WHEN** 管理员在桌面宽度浏览功能定义表单
- **THEN** 开通策略中的短字段 MUST 至少两列并排，页面纵向长度相对原单列布局明显缩短

### Requirement: 视觉组 MUST 支持 Logo 与主色编辑

视觉组 MUST 提供：当前 Logo 预览（无则占位）、文件选择上传、`<input type="color">` 或等价色板与 hex 展示。保存功能定义时 MUST 将所选 color 与（若已上传）logo 经 Admin API 持久化。交互风格 SHOULD 对齐事件管理中的 Logo/色调编辑习惯。

#### Scenario: 选择主色并保存

- **WHEN** 管理员在视觉组选择主色并点击保存功能定义
- **THEN** 后续刷新列表/载入编辑 MUST 显示该主色

#### Scenario: 上传 Logo 后预览更新

- **WHEN** 管理员成功上传 Logo
- **THEN** 视觉组预览 MUST 更新为新图（本地 object URL 或 CDN URL）

### Requirement: 售卖套餐表单 MAY 横排短字段且 MUST NOT 配置功能视觉

售卖套餐编辑区可将价格、天数、授予数量等短字段横向排列。SKU 表单 MUST NOT 提供独立于功能定义的 logo/color 配置。

#### Scenario: SKU 无视觉字段

- **WHEN** 管理员编辑售卖套餐
- **THEN** 表单 MUST NOT 出现功能级 Logo/主色控件
