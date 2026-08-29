## ADDED Requirements

### Requirement: 功能编号在 Admin 中 MUST 只读且不可随意新建

Admin「开通功能」界面与 API MUST NOT 允许运维修改已有功能的 `featureId`，MUST NOT 允许通过管理页创建任意新 `featureId`（新功能仅种子/发版写入）。Admin MUST 允许编辑标题、简介、开通方式、上架、排序及预测默认开通数等非编号字段。App catalog MUST 继续返回服务端 `title`/`description`，供客户端直接展示。

#### Scenario: 编辑文案不改编号

- **WHEN** 运维在管理页修改某已有功能的标题与简介并保存
- **THEN** 系统 MUST 持久化新文案，且 `featureId` MUST 保持不变；后续 catalog MUST 返回新文案

#### Scenario: 拒绝新建未知功能编号

- **WHEN** 运维尝试保存一个库中不存在的 `featureId`
- **THEN** 系统 MUST 拒绝该写入

### Requirement: 售卖套餐商品编码 MUST 可自动生成且所属功能为选择

创建功能 SKU 时，若未提供 `productCode`，服务端 MUST 自动生成全局唯一商品编码；Admin UI MUST 以只读方式展示已生成编码，MUST NOT 要求运维手填编码作为新建前提。所属功能 MUST 从已有功能定义中选择（下拉），MUST NOT 依赖自由文本输入功能编号。

#### Scenario: 新建套餐自动生成编码

- **WHEN** 运维选择所属功能、填写价格与授予参数并保存，且未填写商品编码
- **THEN** 系统 MUST 生成唯一 `productCode` 并落库，且列表可展示该编码

#### Scenario: 所属功能来自已有定义

- **WHEN** 运维打开售卖套餐表单
- **THEN** 所属功能控件 MUST 为已有 `feature_def` 的选择列表，而非任意字符串输入框

### Requirement: 开通方式 MUST 为枚举多选

Admin 配置功能允许的开通方式时，MUST 提供固定枚举（至少 `payment`、`invite_code`、`ad`）的多选控件，MUST NOT 以「逗号拼接自由字符串」作为唯一录入方式。持久化与 App 响应可仍使用逗号分隔字符串以兼容现网。

#### Scenario: 多选保存为兼容串

- **WHEN** 运维勾选「付费」与「邀请码」并保存
- **THEN** 持久化的开通方式 MUST 同时包含 `payment` 与 `invite_code`（顺序可实现自定），且 MUST NOT 依赖运维手输逗号串

### Requirement: Admin 操作按钮 MUST 与表单纵向对齐

开通功能管理页与邀请码管理页的主操作按钮（保存/创建/刷新等）MUST 与表单字段采用清晰的纵向对齐布局（同一表单流内垂直堆叠或独立操作行），避免与输入框在同一 flex 行末尾错位导致难用。

#### Scenario: 开通功能页按钮位置

- **WHEN** 运维打开开通功能管理页
- **THEN** 「保存功能定义」「保存套餐」等按钮 MUST 与对应表单区域纵向对齐，而非仅挤在字段行尾且与标签基线错乱
