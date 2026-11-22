# admin-hub

## Purpose

Hub/设备运维后台 UI 与管理 API：布局、向量门户、统计侧栏、事件树与运维交互的行为规格。 Operator surfaces under resource/public admin and hub pages.

<!-- migrated-from: admin-event-inline-color-confirm, admin-hub-ops-ux, admin-qa-library, admin-usage-wx-list-pager, cash-admin-hub-tiles, cash-vip-admin-ui, device-admin, device-admin-event-logo-color-ui, device-admin-event-parent-picker-ui, device-admin-event-tree-ui, device-admin-user-list, device-ops-ai-debug, feature-admin-activation-snapshot, feature-admin-activation-subject, feature-admin-api, feature-admin-contract-ux, feature-admin-rule-readonly, feature-admin-semantic-layout, feature-admin-ui, feeding-eligibility-admin, feeding-llm-admin-simplify, hub-flywheel-portal-align, hub-flywheel-portal-time-align, hub-history-show-expect-start, hub-history-speak-vector-admin, hub-intent-agent-portal, hub-intent-document-lookup, hub-intent-seed-prompts, hub-layout-horizontal-first, hub-locate-vector-admin, hub-long-pages-master-detail, hub-miss-inbox-admin, hub-near-fullwidth, hub-op-vector-admin, hub-predict-list-accuracy, hub-predict-report-card, hub-predict-vector-admin, hub-qty-vector-admin, hub-read-source-vector-admin, hub-remark-vector-admin, hub-rule-gaps-admin, hub-rule-gaps-time-qty-ux, hub-stats-master-detail, hub-time-vector-admin, hub-type-vector-admin, hub-usage-history-side-panel, hub-usage-stats-days-presets, invite-group-qr-admin, predict-pro-idea-admin, push-device-admin, sim-user-admin, vip-product-admin, voice-admin-ui -->

## Requirements

### Requirement: 行内色调编辑须确认后保存

事件管理页在行内修改事件色调时，SHALL 提供明确的用户确认步骤；系统 SHALL NOT 在取色器 `change` 事件发生时自动调用更新接口。

#### Scenario: 打开色调编辑浮层

- **WHEN** 已登录管理员点击某行「色调」展示区域
- **THEN** 页面 SHALL 显示包含取色控件及 **确定**、**取消** 控件的编辑浮层
- **AND** 取色器 SHALL 初始化为该行当前 `color`（合法时）或默认色

#### Scenario: 点击确定后提交

- **WHEN** 用户在浮层中调整颜色并点击 **确定**
- **THEN** 客户端 SHALL 调用 `POST /device/admin/api/event/update` 并携带该行完整字段与新 `color`
- **AND** 成功后 SHALL 关闭浮层并刷新事件列表

#### Scenario: 点击取消不提交

- **WHEN** 用户在浮层中点击 **取消** 或等价取消操作
- **THEN** 系统 SHALL NOT 调用更新接口
- **AND** 浮层 SHALL 关闭且列表数据保持不变

#### Scenario: 提交中防重复

- **WHEN** 色调更新请求正在进行
- **THEN** 浮层内 **确定** 按钮 SHALL 处于不可用或加载状态直至请求结束

### Requirement: 弹窗编辑与其它行内能力不受影响

弹窗「编辑事件」中的色调与提交行为 SHALL 保持可用；行内 Logo 点击上传流程 SHALL 不因本变更而不可用。

#### Scenario: 弹窗编辑仍可用

- **WHEN** 用户点击行内「编辑」按钮
- **THEN** 页面 SHALL 仍打开含色调字段的编辑弹窗并按原逻辑提交

---

### Requirement: Hub 用户反馈卡片 MUST 位于设备记录下方且常驻全页入口

`resource/public/admin.html` 在管理员登录后展示的卡片顺序中，用户反馈卡片（`#feedbackCard`）MUST 紧挨设备记录卡片（`#deviceRecordCard`）之后。反馈卡片头部 MUST 常驻可见链接，文案为「查看全部」（或等价），目标为 `/device/admin/feedback-records`。该入口 MUST NOT 依赖反馈条数是否超过预览页大小才显示。

#### Scenario: 登录后可见常驻入口

- **WHEN** 管理员在运维 Hub 登录成功且反馈卡片已显示
- **THEN** 页面 MUST 展示可点击的「查看全部」链至反馈全页
- **AND** 即使用户反馈预览不足 10 条，该链接 MUST 仍可见

#### Scenario: 卡片相对位置

- **WHEN** 管理员查看登录后的 Hub 主列卡片
- **THEN** 用户反馈卡片 MUST 出现在设备记录卡片下方，且 MUST 在事件管理等后续卡片之前（相对本变更前的「问答库之后」位置已上移）

### Requirement: 反馈问题列 MUST 顶对齐展示

运维 Hub 反馈预览表与 `/device/admin/feedback-records` 全页反馈表中，「问题」列单元格 MUST 使用顶对齐（`vertical-align: top` 或等价），MUST NOT 使用导致多行问题相对单元格纵向居中的样式。其它列 MUST NOT 因本要求改变既有对齐行为。

#### Scenario: 多行问题顶对齐

- **WHEN** 某条反馈问题文本换行超过一行
- **THEN** 问题列内容 MUST 从单元格顶部开始排布，MUST NOT 垂直居中于单元格

### Requirement: 使用统计页设备号 MUST 可跳转设备数据运维

`api-usage-stats.html` 与 `client-usage-stats.html` 用户维度列表中的非空 `deviceNo` MUST 渲染为可点击链接，导航至 `/device/admin/history/{deviceNo}`（deviceNo MUST URL 编码）。空设备号 MUST NOT 渲染为有效运维链接。链接样式 SHOULD 与 Hub 设备记录 `.device-link` 一致。

#### Scenario: 点击设备号进入运维页

- **WHEN** 管理员在功能使用统计或客户端使用统计的用户列表中点击非空设备号
- **THEN** 浏览器 MUST 导航至同源 `/device/admin/history/{deviceNo}`

#### Scenario: 空设备号不可点

- **WHEN** 某行 `deviceNo` 为空
- **THEN** 该行设备号单元格 MUST NOT 提供指向 history 运维页的有效链接

### Requirement: 用户维度列表 MUST 展示序号且支持本页宝宝名筛选与最近使用排序

两页使用统计的「按用户」列表 MUST：

1. 在表格最前列展示序号，值为当前**渲染结果**从上到下的 1…N；序号 MUST NOT 作为排序键写入服务端。
2. 提供「有无宝宝名」筛选（全部 / 有 / 无），依据响应项 `babyName` trim 后是否非空，且 MUST 仅作用于**当前已加载页**的行集。
3. 展示「最近使用时间」列（格式化展示 `lastAt`；为 0 时显示空或「—」）。
4. 支持点击「最近使用时间」表头对当前已加载页按 `lastAt` 降序重排；未启用该排序时 MUST 按 `wxId`（`id`）升序。
5. 重排或筛选后序号 MUST 按新的可视顺序重新编号。

#### Scenario: 默认按 wxId

- **WHEN** 管理员打开用户维度且未点击最近使用时间排序
- **THEN** 列表行 MUST 按 wxId 升序展示（在服务端返回顺序基础上由前端保证），序号从 1 递增

#### Scenario: 本页按最近使用时间排序

- **WHEN** 管理员点击「最近使用时间」表头
- **THEN** 当前页已加载行 MUST 按 `lastAt` 降序排列，且序号 MUST 按新顺序从 1 重新编号

#### Scenario: 本页筛有宝宝名

- **WHEN** 管理员选择「有宝宝名」且当前页存在 `babyName` 为空与非空的行
- **THEN** 表格 MUST 仅展示 `babyName` 非空的行，且 MUST NOT 因此自动请求其它页

---

### Requirement: 使用统计按用户列表 MUST 支持服务端分页与每页条数

`api-usage-stats.html` 与 `client-usage-stats.html` 的「按用户」wx 账号表 MUST 通过对应 `wx-list` 接口的 `page`、`pageSize` 做服务端分页，MUST NOT 再写死仅请求第一页固定 50 条。默认 `pageSize` MUST 为 **20**；页面 MUST 允许运维在至少包含 20、50、100 的选项中切换每页条数（不得超过后端允许的最大值 100）。页面 MUST 展示当前页码、总页数与总条数，并提供上一页、下一页，以及输入页码后跳转；越界页码 MUST 夹到合法范围。切换每页条数后 MUST 从第 1 页重新加载。

#### Scenario: 默认每页 20

- **WHEN** 管理员打开功能使用统计或客户端使用统计并切换到按用户且未改每页条数
- **THEN** 请求 wx-list 的 `pageSize` MUST 为 20 且 `page` MUST 为 1

#### Scenario: 下一页

- **WHEN** 总条数大于当前 pageSize 且管理员点击下一页
- **THEN** 页面 MUST 请求 `page` 为当前页 +1 的 wx-list 并刷新表格

#### Scenario: 修改每页条数回到首页

- **WHEN** 管理员将每页条数从 20 改为 50
- **THEN** 页面 MUST 以 `page=1&pageSize=50` 重新请求并展示第一页

#### Scenario: 页码跳转越界

- **WHEN** 总页数为 5 且管理员输入页码 99 并跳转
- **THEN** 页面 MUST 加载第 5 页（或等价合法最后一页），MUST NOT 发起非法超大 page 导致空态且无提示的静默失败

### Requirement: 分页下列表序号 MUST 为当前页内序号

按用户表最前列序号 MUST 为当前渲染结果（含本页宝宝名筛选后）从上到下的 1…N，MUST NOT 使用跨页全局偏移（如 `(page-1)*pageSize+i`）作为本变更要求。

#### Scenario: 第二页序号从 1 起

- **WHEN** 管理员在第 2 页且未筛选，表格有多行
- **THEN** 第一行序号 MUST 为 1，而非从 pageSize+1 起算

---

### Requirement: 开通管理页顶部 MUST 平铺 VIP、上架功能与群二维码

静态页 `cash-feature-admin.html` MUST 在顶部平铺可点选的项：固定的 VIP、`feature_def.status=1` 的各功能、以及群二维码。`prediction_unlock` MUST NOT 出现在平铺中。同一时刻 MUST 只展示当前选中项的相关内容，MUST NOT 在未选中时同时展开全部功能定义表、全部 SKU 表与 VIP 套餐表。

#### Scenario: 点选一项后只看该项

- **WHEN** 已登录运维点击「值得留意」
- **THEN** 页面 MUST 展示该功能的定义、该功能的 SKU、开通快照与手工授，且 MUST NOT 同时展示 VIP 套餐表或其它功能的编辑表

#### Scenario: 预测槽位不在平铺中

- **WHEN** 页面加载功能平铺
- **THEN** 平铺 MUST NOT 包含 `prediction_unlock`

### Requirement: VIP 面板 MUST 收纳套餐、权益列表、授予与撤销

选中 VIP 时，页面 MUST 提供现有 VIP 套餐编辑，以及权益列表（含 `channel`、`grantReason`）、手工授表单。列表中 `channel=admin` 且仍有效的行 MUST 提供撤销，提交 MUST 调用 VIP 撤销 API。成功后 MUST 刷新列表。浏览器 MUST NOT 发送 `X-Admin-Password`。`cash-vip-admin.html` MUST NOT 再作为与本面板并行的第二套可写入口（可改为跳转到本页并选中 VIP）。

#### Scenario: 从 VIP 行撤销手工授

- **WHEN** 运维在 VIP 列表对 `channel=admin` 且有效的行确认撤销
- **THEN** 页面 MUST 调用撤销 API，刷新后该行 MUST 显示为已过期或不再提供撤销

#### Scenario: 付费行没有撤销

- **WHEN** 列表行渠道为支付宝或 Apple
- **THEN** 该行 MUST NOT 提供撤销按钮

### Requirement: 功能面板 MUST 支持对该功能手工授与撤销

选中某一上架功能时，手工授表单的功能编号 MUST 固定为该项，MUST NOT 再让运维从全量下拉里改选。快照中 `unlock_method=admin` 且仍有效的行 MUST 提供撤销，并调用功能撤销 API。群二维码 MUST 仅在其自身面板中编辑，MUST NOT 作为某个功能的子表单。

#### Scenario: 在功能面板授功能

- **WHEN** 运维在已选功能面板填写主体标识、期限与非空理由并提交
- **THEN** 系统 MUST 以该平铺项的 `featureId` 调用授功能 API，并刷新该功能快照

#### Scenario: 群二维码独立面板

- **WHEN** 运维点选群二维码
- **THEN** 页面 MUST 展示上传与有效期，且 MUST NOT 要求选择 `featureId`

---

### Requirement: Hub SHALL 提供 VIP 权益只读管理页

系统 MUST 在 `gateway-app-server` 注册静态页 `/device/admin/cash-vip-admin.html`（静态文件 `resource/public/cash-vip-admin.html`）。页面 MUST 使用 `resource/public/admin-common.js` 的 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch`（或等价封装）初始化。页面 MUST 为只读：MUST NOT 提供改价、退款、手工开通或编辑权益的表单/按钮。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 已登录 Hub 打开页面

- **WHEN** 管理员已在运维 Hub 登录并访问 `/device/admin/cash-vip-admin.html`
- **THEN** 主内容区 MUST 可见且 MUST 能发起对 `/cash/admin/api/vip/entitlements` 的加载请求

#### Scenario: 未登录访问

- **WHEN** 管理员未登录即打开该静态页
- **THEN** 页面 MUST 按 admin-common 惯例引导登录或隐藏需鉴权的主内容，MUST NOT 在浏览器侧携带运维口令头

### Requirement: Admin Hub SHALL 登记 cash-vip-admin 模块

`resource/public/admin-modules.js` MUST 保留模块登记：`id: cash-vip-admin`（或等价稳定 id）、`pagePath` 为 `/device/admin/cash-vip-admin.html`、`showInNav: false`。Hub 主导航 MUST NOT 展示「VIP 权益」独立入口。`RegisterAdminStaticPages`（或等价单点注册表）MUST 仍包含该 `pagePath`。PR MUST NOT 仅在主网关注册该页而 App 网关不可见。权益只读页的运维入口 MUST 为开通功能管理页文案为「VIP列表」的按钮（或等价控件）并导航至上述 `pagePath`。

#### Scenario: Hub 导航不展示 VIP 模块

- **WHEN** 运维已登录 `/device/admin` Hub
- **THEN** 模块列表 MUST NOT 以独立导航项展示「VIP 权益」

#### Scenario: 静态页仍可直达

- **WHEN** 已登录管理员打开 `/device/admin/cash-vip-admin.html`
- **THEN** 主内容区 MUST 可见且 MUST 能发起对 `/cash/admin/api/vip/entitlements` 的加载请求

### Requirement: VIP 管理页 SHALL 展示权益与激活金额列

`cash-vip-admin.html` MUST 将 `GET /cash/admin/api/vip/entitlements` 结果以表格或等价列表展示，列至少含：`wxId`、是否仍有效、到期时间、剩余时间（过期时 MUST 展示「已过期」或等价文案，而非伪装为正剩余）、激活金额、支付渠道、最近支付时间。页面 MUST 支持分页；宜提供按 `wxId` 查询。激活金额展示 MUST 与 API 的最近一次 paid `amount_fen` 语义一致（页内须能识别单位为分或换算为元）。

#### Scenario: 列表含已过期行

- **WHEN** API 返回 `isVip=false` 且剩余秒数 ≤0 的行
- **THEN** UI MUST 仍展示该行，且剩余时间文案 MUST 表明已过期

#### Scenario: 展示最近实付金额

- **WHEN** API 返回某行 `lastPaidAmountFen` 为非零
- **THEN** UI MUST 在激活金额列展示对应该值的金额（不得展示商品现价替代）

---

### Requirement: device admin SHALL provide paginated wx account list

`device-service` MUST 暴露 `GET /device/admin/api/wx/list`，鉴权与现有 device admin 一致（Header `X-Admin-Password`）。查询参数：`page`（默认 1）、`pageSize`（默认 20，最大 100）、可选 `q`（模糊匹配 wx.id、deviceNo、unionid、account）。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`（用户名账号若有）、`createdAt`（wx 账号创建 Unix 秒，0 表示未知）。列表 MUST **仅**包含 `is_simulated=0` 的真实用户；`is_simulated=1` 的行 MUST NOT 出现在 `list` 或 `total` 计数中。

#### Scenario: 默认分页列表

- **WHEN** 管理员携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/wx/list` 且未传分页参数
- **THEN** 响应 SHALL 返回第一页 wx 记录且 `page=1`
- **AND** `list` MUST NOT 含 `is_simulated=1` 的 wxId

#### Scenario: 关键字搜索

- **WHEN** 管理员请求 `q=138` 且存在 deviceNo 或 id 匹配的非模拟 wx 行
- **THEN** 响应 `list` SHALL 仅包含匹配的非模拟项

#### Scenario: 经 gateway-app 反代可达

- **WHEN** 管理员经 gateway-app 携带 Admin JWT 请求 `/device/admin/api/wx/list`
- **THEN** gateway-app SHALL 反代至 device-service 并成功返回列表

#### Scenario: createdAt on new account

- **WHEN** 列表项对应迁移后新注册的 wx 账号
- **THEN** 该项 `createdAt` MUST 大于 0

#### Scenario: createdAt zero displays as unknown

- **WHEN** 列表项 `createdAt` 为 0
- **THEN** 消费方展示层 MAY 显示「—」

---

### Requirement: 事件管理列表展示 Logo 与色调

设备管理页（`admin.html` 或等价路由）在登录并加载事件列表后，SHALL 在**树形**表格中展示 **Logo** 与 **色调** 列；每行（含根、中间与叶子节点）SHALL 根据 `GET /device/admin/api/event/list` 返回的 `logo`、`color`、`parentId` 渲染预览与层级缩进。

#### Scenario: 列表含 logo 与 color 字段时展示预览

- **WHEN** 事件列表项包含 `logo` 路径与有效 `color` 色值
- **THEN** 页面 SHALL 在 Logo 列显示可识别的缩略图
- **AND** 色调列 SHALL 显示与 `color` 一致的色块及可读色值文本

#### Scenario: 无 logo 时展示可点击占位

- **WHEN** 事件项 `logo` 为空
- **THEN** Logo 列 SHALL 显示明确占位（如「点击上传」）
- **AND** 占位区域 SHALL 可触发 logo 更换流程

#### Scenario: 子节点独立展示父级同级列

- **WHEN** 子事件在树中缩进展示
- **THEN** 该行 SHALL 仍含 Logo 与色调列且使用**该子事件自身**的 `logo`/`color` 字段

### Requirement: 管理页 Logo 预览使用同源 URL

管理页用于 `<img src>` 的 logo 地址 SHALL 直接使用 API 返回的 CDN 绝对 URL；SHALL NOT 再拼接当前页面 origin 与 `/ai_talk_images/` path。

#### Scenario: API 返回 CDN logo 时展示

- **WHEN** `GET /device/admin/api/event/list` 返回 `logo` 为 `https://resorce.cuplay.top/event/...`
- **THEN** 页面 SHALL 将该 URL 直接用于 `<img src>`

#### Scenario: 无 logo 时展示可点击占位

- **WHEN** 事件项 `logo` 为空
- **THEN** Logo 列 SHALL 显示明确占位（如「点击上传」）
- **AND** 占位区域 SHALL 可触发 logo 更换流程

### Requirement: 主网关提供同源事件图片访问

gateway-service（管理页常用入口，如 :9701）SHALL 注册 `GET /ai_talk_images/*`（及 HEAD），将请求反代或等价转发至 device-service 的同名静态读能力，使管理页同源 URL 可成功返回图片。

#### Scenario: 经主网关读取已上传 logo

- **WHEN** 客户端请求 `GET https://<gateway-host>/ai_talk_images/<安全文件名>` 且 device-service 上文件存在
- **THEN** gateway-service SHALL 返回对应图片内容且 SHALL NOT 要求 App 网关 Bearer

### Requirement: 点击色调即可更新 color

管理页 SHALL 允许用户通过点击列表行中的色调展示区域修改该事件的 `color`，并在成功后刷新列表。

#### Scenario: 点击色块修改 color

- **WHEN** 用户点击某行色调区域并选择新色值后确认提交
- **THEN** 客户端 SHALL 调用 `POST /device/admin/api/event/update`（multipart）并携带该行 `id`、`name`、`needQuantity`、`extraNames` 及新 `color`
- **AND** 未选择新 logo 文件时 SHALL NOT 清除原有 `logo`

#### Scenario: 更新成功后列表反映新色值

- **WHEN** 更新接口返回成功
- **THEN** 页面 SHALL 刷新事件列表且该行色调展示与新 `color` 一致

### Requirement: 点击 Logo 即可更新 logo 文件

管理页 SHALL 允许用户通过点击列表行中的 Logo 缩略图或占位触发文件选择，上传新图并更新该事件。

#### Scenario: 点击 Logo 上传新图

- **WHEN** 用户点击 Logo 区域并选择合法图片文件（如 png/jpeg/webp）
- **THEN** 客户端 SHALL 以 multipart 调用 `POST /device/admin/api/event/update`，包含 `logo` 文件及该行完整文本字段
- **AND** 成功后服务端 `event.logo` SHALL 更新为 path-only 新路径

#### Scenario: 更新成功后列表展示新缩略图

- **WHEN** logo 更新成功且列表重新加载
- **THEN** 该行 Logo 列 SHALL 使用同源 URL 展示新图

### Requirement: 行内编辑与弹窗编辑并存

名称、事件扩展、是否需要计数等字段 SHALL 仍可通过既有「编辑」弹窗修改；行内交互仅负责 **logo** 与 **color**，SHALL NOT 要求用户仅为改色/改图打开完整弹窗。

#### Scenario: 编辑按钮仍打开完整表单

- **WHEN** 用户点击行内「编辑」按钮
- **THEN** 页面 SHALL 打开包含名称与其它字段的编辑弹窗（行为与变更前一致）

### Requirement: 树形列表每行可新增子事件

除顶部「新增事件」外，事件管理页每一行 SHALL 提供「新增子事件」操作，打开创建表单并携带该行 id 作为 `parentId`。

#### Scenario: 父行展示新增子事件按钮

- **WHEN** 用户查看事件树中任意节点行
- **THEN** 操作列 SHALL 含「新增子事件」入口

#### Scenario: 新增子事件成功后树刷新

- **WHEN** 用户通过「新增子事件」成功创建记录
- **THEN** 页面 SHALL 重新加载列表且新节点出现在对应父节点下

---

### Requirement: 编辑事件时 SHALL 可选择父事件

设备管理页在**编辑**已有事件时，SHALL 提供父事件选择控件（含「无 / 根」选项，对应 `parentId=0`）。提交 `POST /device/admin/api/event/update` 时 SHALL 在 multipart 表单中包含 **`parentId`** 字段。

#### Scenario: 打开编辑弹窗默认选中当前父

- **WHEN** 管理员编辑 `parentId=5` 的事件
- **THEN** 父事件选择器 SHALL 默认选中 id=5 的项（或等价展示父名称）

#### Scenario: 提交修改父节点

- **WHEN** 管理员将父改为根并保存
- **THEN** 请求 SHALL 包含 `parentId=0`
- **AND** 成功后列表树形结构 SHALL 反映该节点位于根层

### Requirement: 父事件选择器 SHALL 排除非法选项

选择器 SHALL NOT 提供**当前事件自身**及其**全部后代**作为父选项，以避免必然触发后端成环校验失败。

#### Scenario: 编辑叶子事件时不出现自身为父

- **WHEN** 编辑 id=20 的叶子事件
- **THEN** 父事件下拉 SHALL NOT 包含 id=20

#### Scenario: 编辑有子节点时不出现其子孙为父

- **WHEN** 编辑 id=10 且存在 `parent_id=10` 的子事件 20
- **THEN** 父事件下拉 SHALL NOT 包含 id=20

---

### Requirement: 事件管理页树形展示层级

设备管理页事件模块 SHALL 根据 `ListEvents` 返回的扁平数组（含 `parentId`）渲染**树形**列表：根节点按 id 或现有排序规则排列，子节点缩进展示在其父节点之下；深度 SHALL 支持通用树（不限两级）。

#### Scenario: 换尿布与子事件分级可见

- **WHEN** 列表含 `换尿布(parentId=0)` 与 `大便(parentId=换尿布.id)`
- **THEN** 页面 SHALL 将「大便」行展示在「换尿布」之下并带可视缩进

#### Scenario: 多级中间节点可展开式展示

- **WHEN** 存在根 → 中间 → 叶子三级关系
- **THEN** 页面 SHALL 按 parentId 递归嵌套展示全部层级

### Requirement: 支持新增根事件与新增子事件

页面 SHALL 提供「新增事件」创建根节点；每一行（含中间节点）SHALL 提供「新增子事件」入口，提交时携带 `parentId` 为该行 id。

#### Scenario: 从换尿布行新增子事件

- **WHEN** 用户点击「换尿布」行的「新增子事件」
- **AND** 填写名称「小便」并提交
- **THEN** 客户端 SHALL `POST /device/admin/api/event/add` 且表单含 `parentId=<换尿布.id>`
- **AND** 成功后列表 SHALL 在「换尿布」下展示「小便」

#### Scenario: 新增根事件不带 parentId

- **WHEN** 用户点击顶部「新增事件」并提交
- **THEN** 请求 SHALL NOT 携带非零 `parentId`（或显式 `parentId=0`）

### Requirement: 树形列表保留 Logo 与色调行内编辑

树形结构中每一节点 SHALL 独立展示并支持行内 **Logo**、**色调** 编辑（行为与扁平列表时期一致）；子节点 SHALL NOT 因存在父节点而隐藏 logo/color 列。

#### Scenario: 中间节点可上传独立 Logo

- **WHEN** 用户为中间节点「排泄类」点击 Logo 上传新图
- **THEN** 仅该节点 `logo` SHALL 更新
- **AND** 父节点「换尿布」的 `logo` SHALL 保持不变

### Requirement: 子事件创建表单不预填父 logo 与 color

打开「新增子事件」弹窗时，色调与 Logo SHALL 使用与「新增根事件」相同的默认空状态，SHALL NOT 预填父节点当前 `color` 或 `logo` 预览为默认值。

#### Scenario: 子事件弹窗色值非父色

- **WHEN** 父节点 color 为 `#FF0000`
- **AND** 用户打开该父下的「新增子事件」弹窗
- **THEN** 颜色选择器 SHALL NOT 因父色而默认选中 `#FF0000`（除非用户手动选择）

---

### Requirement: 管理端设备记录分页列表

device-service SHALL 提供 `GET /device/admin/api/user/list`，要求 Header `X-Admin-Password` 有效。查询参数 `page`（默认 1）、`pageSize`（默认 5，最大 100）、可选 `q`（`device_no` 模糊包含，大小写不敏感以库排序规则为准）。响应 MUST 为 `{ list, total, page, pageSize }`，`list` 每项含 `deviceNo`、`activeTime`、`lastTalkTime`、`lastTalkAsk`、`lastTalkAnswer`、`lastApiPath`、`lastApiAt`。

#### Scenario: 默认分页

- **WHEN** 管理员请求 `GET /device/admin/api/user/list` 且未传 `pageSize`
- **THEN** 返回最多 5 条记录且 `pageSize` 字段为 5

#### Scenario: 模糊搜索

- **WHEN** 管理员请求带 `q=abc`
- **THEN** 仅返回 `device_no` 包含子串 `abc` 的设备

### Requirement: 最近 HTTP 接口落库

对任意经网关处理的 HTTP 请求，若可解析出非空 `deviceNo` 且路径不是 WebSocket、不以 `/device/internal/` 开头，系统 SHALL 异步更新该设备 `last_api_path`（`METHOD /path`）与 `last_api_at`（Unix 秒）。WebSocket 升级请求 MUST NOT 触发更新。

#### Scenario: 带 query 的 history 列表

- **WHEN** 客户端请求 `GET /device/history/api/list?deviceNo=d1&page=1`
- **THEN** 设备 `d1` 的 `last_api_path` 更新为 `GET /device/history/api/list`

### Requirement: 管理端设备号跳转历史页

`admin.html` 设备记录表中 `device_no` MUST 为指向 `/device/history/{deviceNo}` 的链接（URL 编码 deviceNo）。

#### Scenario: 点击设备号

- **WHEN** 管理员点击列表中某行的设备号链接
- **THEN** 浏览器导航至同源的 `/device/history/{deviceNo}` 历史管理页

---

### Requirement: 运维页支持 App 用户名登录

设备数据运维单设备页（`/device/admin/history/{deviceNo}` 所载 `history.html`）MUST 提供 App 用户名密码登录入口，调用既有 `POST /device/app/api/username_login`，并将返回的 `accessToken` 存于与 Admin JWT **不同**的存储键。Admin JWT 继续用于既有 admin API；tip 与 care-alert 调试请求 MUST 使用 App `accessToken` 作为 Bearer。

#### Scenario: 未 App 登录不可调 tip/care-alert 调试

- **WHEN** 运维未完成 App 用户名登录即点击小贴士生成或值得留意拉取
- **THEN** 页面 MUST 提示需先登录，MUST NOT 使用 Admin JWT 调用上述 App 接口

#### Scenario: Admin 与 App token 分钥

- **WHEN** 运维同时持有 Admin JWT 与 App accessToken
- **THEN** 历史 CRUD MUST 仍带 Admin Bearer；tip/care-alert MUST 带 App Bearer

### Requirement: 运维文字对话调试走 chat WS 文模式

运维页 MUST 提供文字对话调试面板：连接 `/voice/chat/ws`，使用 text 输入与 text 输出模态，展示服务端 `thinking_delta` 与 `answer`。MUST NOT 为此面板新增 gateway 鉴权豁免或 Admin 伪造身份接口。

#### Scenario: 发送文本可见思考与回答

- **WHEN** 运维在已 start（文模式）的连接上发送非空文本
- **THEN** 面板 MUST 展示思考增量与最终 answer（允许忽略音频帧）

### Requirement: 运维小贴士调试走正式 tip SSE

运维页 MUST 通过 `POST /device/tip/generate`（body 含当前 `deviceNo` 及可选 event 字段）发起 SSE，展示 `thinking`/`answer`/`done`（或等价事件）。MUST NOT 新增 tip 的 Admin 旁路或鉴权豁免。重复点击生成即视为强刷（服务端无日缓存门禁时仍走正式额度语义）。

#### Scenario: tip SSE 展示思考过程

- **WHEN** 运维已 App 登录并对当前设备触发小贴士生成
- **THEN** 页面 MUST 以流式方式展示思考与最终建议内容

### Requirement: 运维值得留意调试走正式 care-alert API

运维页 MUST 通过 `GET /device/api/care-alert/daily` 拉取当日列表并展示；MUST 提供强刷操作，请求带 force 语义（见 `care-alert-force-refresh`）。MUST NOT 使用 Admin JWT 或伪造 `X-Internal-Wx-Id` 调用该接口。

#### Scenario: 拉取当日值得留意

- **WHEN** 运维已 App 登录并点击拉取值得留意
- **THEN** 页面 MUST 展示返回的 `day` 与 `items`（可空列表）

#### Scenario: 强刷按钮带 force

- **WHEN** 运维点击值得留意强刷
- **THEN** 页面 MUST 使用带 force 的 daily 请求（仍携带 App Bearer）

---

### Requirement: Admin 按功能查看当前开通快照

cash Admin MUST 提供按 `featureId` 查询的开通快照接口（及管理页入口），返回该功能当前持久化开通主体列表（可分页），每项至少包含：主体标识（`deviceNo` 或 `wxId`）、最近开通方式（若适用）、是否未过期有效、到期时间或剩余时效（`expires_at=0` 表示永久；预测条数类可用条数代替时效）。快照 MUST 来自当前权益/条数权威表（方案 A），MUST NOT 要求新建事件账本。页面 MUST 注明不含 VIP 旁路覆盖、非完整历史事件流。

#### Scenario: 成长轨迹列出账号权益

- **WHEN** 运维打开 `growth_trajectory_predict` 详情快照
- **THEN** 列表 MUST 来自 `feature_user_entitlement`（或等价），含 wx、方式、是否有效与到期/永久

#### Scenario: 值得留意列出设备权益

- **WHEN** 运维打开 `care_alert_smart_remind` 详情快照
- **THEN** 列表 MUST 来自 `feature_entitlement`，含 deviceNo、方式、是否有效与到期/永久

#### Scenario: 预测列出条数

- **WHEN** 运维打开 `prediction_unlock` 详情快照
- **THEN** 列表 MUST 展示设备与永久条数（或等价字段），MUST NOT 伪造权益型到期字段为必需

#### Scenario: 已过期仍可见

- **WHEN** 某主体权益 `expires_at` 已早于当前时间且行仍存在
- **THEN** 快照 MUST 仍可列出并标记为未在有效期内

---

### Requirement: Admin 功能定义 API MUST 暴露开通主体

`GET /cash/admin/api/feature/defs` 返回的每一项 MUST 包含 `activationSubject`（`device`|`user`）。`POST /cash/admin/api/feature/defs` MUST 接受可选 `activationSubject`；若提供则 MUST 校验取值并持久化。将 `prediction_unlock`（或其它条数类）更新为 `user` MUST 被拒绝。未传该字段时 MUST 保持库中原值不变。

#### Scenario: 列表返回人对机

- **WHEN** 管理员拉取功能定义列表
- **THEN** 每条 MUST 含 `activationSubject`，成长轨迹为 `user`，值得留意为 `device`

#### Scenario: 禁止预测改为人

- **WHEN** Admin 将预测功能 `activationSubject` 设为 `user` 并保存
- **THEN** 系统 MUST 拒绝且 MUST NOT 修改该行主体

### Requirement: 开通功能管理页 MUST 标明对人或对机

`resource/public/cash-feature-admin.html`（经 gateway 运维 Hub 入口）MUST 在功能定义列表中展示每条功能的开放主体，文案 MUST 区分「对人」与「对机」（或等价清晰中文），MUST NOT 仅依赖英文枚举让运维猜测。编辑区 MUST 展示当前主体；若允许修改，MUST 提供明确控件并在保存后与列表一致。

#### Scenario: 列表一眼可辨成长轨迹对人

- **WHEN** 运维打开开通功能管理页并加载定义列表
- **THEN** `growth_trajectory_predict` 行 MUST 显示「对人」（或含「人/账号」语义的同等文案）

#### Scenario: 值得留意显示对机

- **WHEN** 列表展示 `care_alert_smart_remind`
- **THEN** 该行 MUST 显示「对机」（或含「设备」语义的同等文案）

---

### Requirement: cash-service MUST 提供付费功能管理 Admin CRUD API

cash-service MUST 在 `/cash/admin/api/feature/**`（或等价前缀）提供管理端 API，支持功能定义（`feature_def`）、解锁方式、功能 SKU（`feature_product`）、预测数量授予相关配置的创建/读取/更新/停用。写操作 MUST 校验 Admin 鉴权（与现网 cash Admin 一致）。Admin API MUST NOT 改变 App usage 统计 denylist 的默认假设（管理端前缀）。

#### Scenario: 创建功能定义

- **WHEN** 管理员提交合法的新功能定义
- **THEN** 系统 MUST 持久化到 `ai_voice_cash` 并在后续目录/管理列表中可见

#### Scenario: 配置功能 SKU

- **WHEN** 管理员创建或更新功能 `product_code` 与价格/授予语义
- **THEN** 系统 MUST 持久化，且支付建单 MUST 能引用该 SKU

#### Scenario: 未鉴权拒绝

- **WHEN** 请求未携带有效 Admin 凭证
- **THEN** 系统 MUST 拒绝写操作

### Requirement: cash-service MUST 提供邀请码管理 Admin API 含兑换明细

系统 MUST 提供邀请码 Admin API（建议 `/cash/admin/api/invite-code/**`）：支持创建/读取/更新/停用邀请码；创建时 MUST 绑定 `owner_wx_id`；一期 MUST 强制一 `owner_wx_id` 仅一条有效码；MUST 可配置有效期与可开通功能集合。MUST 提供按码查询兑换明细：至少含 `redeemer_wx_id`、`device_no`、`featureId`、`redeemed_at`，供追踪销售推广。

#### Scenario: 创建邀请码绑定 owner

- **WHEN** 管理员为某 `owner_wx_id` 创建邀请码并指定有效期与可开功能
- **THEN** 系统 MUST 保存码记录；若该 owner 已有码则 MUST 拒绝或按产品约定停用旧码（一期一人一码）

#### Scenario: 查看某码被哪些人使用

- **WHEN** 管理员查询某邀请码的兑换明细
- **THEN** 系统 MUST 返回使用该码成功兑换的记录列表，含兑换人 `wx_id` 与时间

### Requirement: Admin API MUST 经 gateway-app 反代且禁止他域直查

`/cash/admin/api/feature/*` 与邀请码 Admin 路径 MUST 由 gateway-app cash 反代到达 cash-service。其他非 cash 进程 MUST NOT 直查 `ai_voice_cash` 功能表。

#### Scenario: 反代可达

- **WHEN** 已登录 Hub 的管理员调用功能或邀请码 Admin API
- **THEN** 请求 MUST 经 gateway 到达 cash-service 并成功鉴权（与 VIP Admin 同模式）

---

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

---

### Requirement: 功能定义展示只读开通规则

开通功能管理（Admin）在展示某一功能定义时 MUST 提供只读开通规则说明（服务端按 `feature_id` 派生或常量映射），涵盖：开放主体（对人/对机）、邀请开通约束与效果口径、付费主体口径，以及共用「同宝宝禁兑」说明（MUST NOT 再将「不可使用自己的邀请码」列为独立共用前置）。运维 MUST NOT 通过该字段编辑并改写真实闸门逻辑。

#### Scenario: 成长轨迹规则文案

- **WHEN** 运维查看 `growth_trajectory_predict` 的功能定义/详情
- **THEN** 只读规则 MUST 表明对人开通、同机可兑不同邀请码、每人邀请仅一次，且提及同宝宝账号互兑禁止，MUST NOT 再单独要求「不可使用自己的邀请码」作为共用前置句

#### Scenario: 值得留意规则文案

- **WHEN** 运维查看 `care_alert_smart_remind`
- **THEN** 只读规则 MUST 表明对机开通且同宝宝邀请开通仅一次，且共用说明含同宝宝禁兑、不含独立自用禁兑句

### Requirement: 同宝宝禁兑继续生效

系统兑换邀请码时 MUST 拒绝码主人当前绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换；MUST NOT 因此写入开通或获客原力。拒绝提示 MUST 明确不可使用同一宝宝下其他账号的邀请码（或同等语义）。系统 MUST NOT 再以 `owner_wx_id == redeemer_wx_id` 作为独立拒绝条件。

#### Scenario: 同宝宝拒兑

- **WHEN** 兑换者 `deviceNo` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 不再因同 wx 单独拒兑

- **WHEN** 兑换者与码主人 `wx_id` 相同，但双方当前 `device_no` 不同或主人未绑机，且其它校验通过
- **THEN** 系统 MUST NOT 仅因同 wx 拒绝兑换

---

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

---

### Requirement: Hub MUST 提供「开通功能管理」静态页

系统 MUST 在 `gateway-app-server` 注册付费功能管理静态页（建议 `/device/admin/cash-feature-admin.html`）。页面 MUST 使用 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch` 访问功能 Admin API，支持功能定义、解锁方式、功能 SKU 等 CRUD。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 登录后可打开功能管理页

- **WHEN** 管理员已登录 Hub 并访问开通功能管理页
- **THEN** 页面 MUST 完成 Admin 校验并可通过 adminFetch 加载功能/SKU 列表

### Requirement: Hub MUST 提供独立「邀请码管理」静态页

系统 MUST 注册独立邀请码管理静态页（建议 `/device/admin/cash-invite-code-admin.html`）。页面 MUST 支持邀请码 CRUD（有效期、owner_wx_id、可开功能）及按码查看兑换明细（`wx_id` / 设备 / 功能 / 时间）。MUST 使用 AdminCommon，MUST NOT 在浏览器附带 `X-Admin-Password`。

#### Scenario: 可查看兑换明细

- **WHEN** 管理员在邀请码管理页打开某码的使用记录
- **THEN** 页面 MUST 展示该码成功兑换的用户与时间等信息

### Requirement: Admin Hub SHALL 登记两个模块入口

`admin-modules.js` MUST 增加至少两个模块：开通功能管理、邀请码管理（稳定 id、中文标题、`pagePath`、`showInNav: true`）。`RegisterAdminStaticPages` 与静态页 Bearer 白名单 MUST 包含上述 `pagePath`。

#### Scenario: 导航可见两个入口

- **WHEN** 管理员打开运维 Hub 导航
- **THEN** 模块列表 MUST 同时包含开通功能管理与邀请码管理入口

### Requirement: 管理页交互 MUST 对齐现有 cash-vip-admin 模式且可写

两页的登录、同源 API、错误提示与退出行为 MUST 对齐 Hub 惯例；MUST 为可写 CRUD，MUST NOT 误用 VIP 只读页的禁止写约束。

#### Scenario: 使用 AdminCommon 而非裸口令头

- **WHEN** 页面发起 Admin API 请求
- **THEN** 请求 MUST 经 Admin JWT / adminFetch，MUST NOT 在浏览器侧附带 `X-Admin-Password`

### Requirement: 开通功能管理页 MUST 提供 VIP列表入口

`cash-feature-admin.html` MUST 提供按钮或等价控件，可见文案 MUST 为「VIP列表」，激活后 MUST 导航至 `/device/admin/cash-vip-admin.html`。该入口 MUST 替代 Hub 主导航上的「VIP 权益」模块作为常规运维路径。

#### Scenario: 按钮文案与跳转

- **WHEN** 已登录管理员打开开通功能管理页
- **THEN** 页面 MUST 展示文案为「VIP列表」的入口，激活后 MUST 进入 VIP 权益只读页

---

### Requirement: Admin MUST 可配置喂养资格场景阈值

cash-service MUST 提供管理端 API（及 Hub 静态页入口）读取与更新已登记场景的 `requiredDays`、`minRecordsPerDay`。场景键 `ucg_entry`、`care_alert_entry`、**`predict_pro_entry`** MUST 由种子/发版约定，Admin MUST NOT 创建任意新 `scene_key`。写操作 MUST 校验 Admin 鉴权；成功后 MUST 使相关资格缓存失效（或等价使旧缓存不可命中）。

#### Scenario: 更新 UCG 连续天数

- **WHEN** 运维将 `ucg_entry.requiredDays` 从 7 改为 5 并保存
- **THEN** 后续 UCG eligibility 的 `requiredDays` MUST 为 5，且缓存 MUST NOT 长期返回改前结果

#### Scenario: 更新预测 Pro 场景天数

- **WHEN** 运维将 `predict_pro_entry.requiredDays` 从 7 改为 10 并保存
- **THEN** 后续 Pro eligibility 的 `requiredDays` MUST 为 10

#### Scenario: 拒绝未知场景新建

- **WHEN** 运维尝试写入未登记的 `scene_key`
- **THEN** 系统 MUST 拒绝该写入

### Requirement: 场景配置 MUST 有安全默认种子

EnsureSchema（或等价迁移）MUST 种子 `ucg_entry`（默认 requiredDays=7、minRecordsPerDay=10）、`care_alert_entry`（默认 requiredDays=2、minRecordsPerDay=10）与 **`predict_pro_entry`（默认 requiredDays=7、minRecordsPerDay=10）**，幂等且 MUST NOT 在重复启动时用种子覆盖运维已改值（或仅 INSERT 忽略已存在）。

#### Scenario: 首次部署有三行默认

- **WHEN** 空库执行 EnsureSchema
- **THEN** 配置表 MUST 含 `ucg_entry`、`care_alert_entry`、`predict_pro_entry` 默认行

---

### Requirement: voiceUnderstanding 不再配置 free 模型

喂养 lane `voiceUnderstanding` 的 Admin 配置与 API MUST 仅包含正式模字段：`provider`、`model`、`maxInFlight`、`maxWaiters`（及既有审计字段）。MUST NOT 再要求或展示 `freeProvider`/`freeModel`。PUT 若仍收到 free 字段，MUST 忽略或拒绝，且 MUST NOT 再将其作为额尽降级选模真相源。

#### Scenario: Admin 保存 VU 无 free 控件

- **WHEN** 运维打开 AI 模型与并发页的 voiceUnderstanding（喂养默认智能体）区块并保存
- **THEN** UI MUST NOT 提供 free 模型下拉；保存体 MUST NOT 依赖 free 字段才能成功

#### Scenario: 硬件喂养只用正式模

- **WHEN** `/voice/chat/ws` 触发喂养 LLM
- **THEN** 选模 MUST 使用 VU 正式 provider/model，MUST NOT 读取 VU free 配置

### Requirement: 额度 Admin 隐藏喂养月度项

Voice AI 额度管理界面 MUST 隐藏喂养（`voice_ai` / `voiceAiMonthlyLimit`）全局默认与 per-user 覆盖的展示与编辑入口。clinic / care-alert 等其他 feature 的额度配置 MUST 保持可用（若该页已有）。

#### Scenario: 额度页无喂养表单项

- **WHEN** 管理员打开 voice-admin 额度功能区
- **THEN** 页面 MUST NOT 展示可编辑的喂养月度额度字段

### Requirement: 喂养对话与 voice_ai 额度解耦

自然语言喂养对话路径 MUST NOT 再以 `voice_ai` 月度 used/limit 作为拒绝或 free 降级条件。`voice_ai` 存储与内部 API MAY 短期保留，但 MUST NOT 作为 chat WS 对话门禁。

#### Scenario: 无 40302 挡喂养 WS

- **WHEN** 某 wx 的 `voice_ai` 已用尽且客户端经 `/voice/chat/ws` 发起喂养对话
- **THEN** 服务端 MUST NOT 因该额度用尽返回阻断对话的 40302（或等价拒绝）；MUST 仍可走 VU 正式模（受并发闸门约束）

---

### Requirement: 意图智能体门户按热路径飞轮重排

Hub 意图智能体门户页 SHALL 以运维可发现的入口聚合热路径飞轮与规则缺口。门户 MUST 提供前往以下能力的入口：OP 轨、事件、数量、R·数据源、查史播报、规则缺口。门户 MUST NOT 再将「类型向量」作为主卡片入口（类型页深链 URL MAY 仍可直接访问）。

#### Scenario: 门户可见新飞轮卡

- **WHEN** 运维已登录 Hub 并打开意图智能体门户
- **THEN** 可见 OP 轨、事件、数量、R·数据源、查史播报与规则缺口入口
- **AND** MUST NOT 在门户主网格展示「类型向量」卡片

#### Scenario: 类型深链仍可用

- **WHEN** 运维直接访问既有 `type-vector-admin.html` 路径
- **THEN** 页面仍可加载（须 Admin 登录门禁与既有页一致）

### Requirement: 类型向量页遗留提示

类型向量管理页 MUST 明示该集合热路径已停用、新种子应使用 OP 轨飞轮；MUST NOT 继续引导运维以旧 `target_type`（feeding/history/predict/conversation）作为主灌种策略而不加遗留说明。

#### Scenario: 打开类型页见遗留说明

- **WHEN** 运维打开类型向量页
- **THEN** 页内可见热路径停用 / 改用 OP 轨的提示文案

### Requirement: 规则缺口文案对齐数量飞轮

规则缺口页 MUST 保留默认 TIME+QTY 展示语义（空维度不传 `dimension`，由 Python 合并）。文案 MUST 标明 TIME 为扩本地规则主工单；MUST 说明数量维可配合数量飞轮整段灌种，MUST NOT 暗示数量问题只能通过改规则代码解决。

#### Scenario: 缺口页提示可读

- **WHEN** 运维打开规则缺口页
- **THEN** 可见 TIME 主工单与 QTY/数量飞轮关系的说明，且默认筛选语义仍为 TIME+QTY

---

### Requirement: 意图智能体门户暴露四维新飞轮入口

Hub 意图智能体门户页 SHALL 在飞轮整理区提供前往以下能力的入口：时间、备注、R·predict、U/D 定位（与既有 OP / 事件 / 数量 / R·数据源 / 查史播报并列或同区）。门户 MUST NOT 再宣称「TIME 无飞轮」或等价「时间维禁止向量飞轮」表述。

#### Scenario: 门户可见四新卡且 TIME 文案已对齐

- **WHEN** 运维已登录 Hub 并打开意图智能体门户
- **THEN** 页面 MUST 展示时间、备注、R·predict、U/D 定位入口，且 MUST NOT 显示「TIME 无飞轮」类文案

### Requirement: 规则缺口页对齐时间飞轮语义

规则缺口管理页 SHALL 说明 TIME 仍为主工单（扩本地规则须改 Python 发版），并 MUST 说明口语时间可配合「时间飞轮」整段相对配方灌种以减少重复 TIME LLM；MUST NOT 再暗示时间维完全无飞轮。默认 dimension 筛选语义 MUST 保持不变（空则吃上游 TIME+QTY 默认）。

#### Scenario: 缺口页文案含时间飞轮配合提示

- **WHEN** 运维打开规则缺口页
- **THEN** 页面 MUST 同时保留 TIME 主工单语义与「可配合时间飞轮」提示，且 MUST NOT 写「无时间飞轮」

---

### Requirement: Hub history event table SHALL show expected start time

html`) history events table MUST include a read-only column **预计开始** that displays each row’s `expectStartTime` when present (Unix seconds formatted consistently with the page’s start-time display helpers). When `expectStartTime` is missing, null, or non-positive, the cell MUST show a clear empty placeholder (e. This MUST NOT require new backend API fields beyond what list already returns. Hub 设备历史页历史事件表 MUST 含只读列 **预计开始**，有 `expectStartTime` 时按与开始时间一致的可读格式展示； 缺失/空/非正 MUST.

#### Scenario: Row with expectStartTime

- **WHEN** list 某行含正 `expectStartTime`
- **THEN** 「预计开始」单元格 MUST 显示对应可读时间

#### Scenario: Row without expect

- **WHEN** list 某行无 `expectStartTime` 或值无效
- **THEN** 「预计开始」单元格 MUST 显示占位符（如 `-`）

### Requirement: Usage stats embedded history tab SHALL show the same expected start column

On `api-usage-stats` and `client-usage-stats` user-detail **历史事件** tables, the column set MUST include **预计开始** with the same semantics and formatting intent as the Hub `history.html` history events table, so operators see consistent fields when drilling from usage stats.

使用统计两页右侧历史 Tab 表 MUST 含 **预计开始**，语义与格式意图与 `history.html` 历史事件表一致。

#### Scenario: Side panel shows expect start

- **WHEN** 运维在使用统计按用户选中设备并打开历史 Tab 且某行有 `expectStartTime`
- **THEN** 该行「预计开始」MUST 可见且非空占位

---

### Requirement: 查史播报向量 Admin 代理

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python 查史播报缓存 Admin API（契约路径 `/v1/admin/history-speak-cache` 及其 bulk/dedupe/delete），支持列表、批量写入、去重与按 id 删除。Hub 侧路径 SHALL 使用 `/voice/admin/api/history-speak-vectors*`（与 intent/type vectors 命名对称）。MUST NOT 在 Go 进程内直连 Chroma 或新建查史播报向量表。批量条目 payload MUST 透传上游字段（含 `point_pick`、`open_speak`）；同 document 跳过策略由上游保证。

#### Scenario: 列表代理

- **WHEN** 运维已通过 Hub Admin 鉴权并请求查史播报向量列表
- **THEN** voice-service 调用 Python history-speak-cache 列表接口并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id 删除
- **THEN** 系统将请求代理至对应 Python admin 接口，且 MUST NOT 因代理路径改写 Redis/MySQL 权威数据

#### Scenario: 口令拒绝

- **WHEN** 请求缺少有效 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝该 Admin 请求

### Requirement: 查史播报向量 Hub 整理页

Hub SHALL 提供查史播报向量静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/history-speak-vectors*` 接口，支持列表浏览、bulk 写入、dedupe 与删除。页面 MUST 说明数据位于 Python `feeding_history_speak_cues`，payload 维为 `point_pick`（`latest`|`prefer_completed`）与 `open_speak`（`elapsed`|`in_progress`）。该页 MUST 可从意图智能体门户进入，并 MUST 登记静态路由与 auth exempt（与既有向量管理页同级）。

#### Scenario: 从门户进入并列表

- **WHEN** 运维从意图智能体门户进入查史播报向量页且 Admin 已登录
- **THEN** 页面可拉取并展示向量列表（或在上游未就绪时给出可理解错误提示）

#### Scenario: bulk 示例载荷

- **WHEN** 运维在页内按提示 bulk 写入含合法 `point_pick` 与 `open_speak` 的条目
- **THEN** 请求经 voice 代理到达 Python history-speak-cache bulk（成功语义以上游为准）

---

### Requirement: 意图智能体门户与导航收口

Hub 运维导航 SHALL 以单一入口「意图智能体」暴露意图智能体相关工具，并 MUST 提供门户页列出子能力入口。下列模块 MUST NOT 再作为顶栏平级 `showInNav` 入口：意图向量缓存页、类型向量页、规则缺口页。上述子页的既有 URL MUST 仍可直接访问（深链）。门户 MUST 至少包含指向以下能力的入口：事件向量（原意图向量缓存页）、类型向量、查史播报向量、规则缺口。

#### Scenario: 顶栏仅见意图智能体

- **WHEN** 运维在 Hub 主页查看导航链接（`showInNav`）
- **THEN** 可见「意图智能体」入口，且 MUST NOT 再平级展示「意图向量缓存」「类型向量」「规则缺口」三个独立导航项

#### Scenario: 门户聚合四子入口

- **WHEN** 运维打开意图智能体门户页
- **THEN** 页面提供前往事件向量、类型向量、查史播报向量、规则缺口的链接或等价入口

#### Scenario: 旧深链仍可用

- **WHEN** 运维直接访问既有 `intent-vector-admin.html` / `type-vector-admin.html` / `rule-gaps-admin.html` 路径
- **THEN** 页面仍可加载（不因导航收口而删除路由）

---

### Requirement: 门户按 document 查询九个飞轮

voice-service MUST 提供需管理员身份的 `GET /voice/admin/api/intent-cues/by-document`。它 MUST 把查询参数 `document` 转发给 Python `GET /v1/admin/cues/by-document`，并返回该接口的命中列表。去除空白后的 `document` 为空时，服务端 MUST 拒绝且 MUST NOT 调用 Python。该接口 MUST NOT 通过逐页拉取各维度列表再在本地过滤来实现。

#### Scenario: 转发精确查询

- **WHEN** 管理员传入一句非空原文
- **THEN** voice-service MUST 向 Python 聚合接口查询该原文，并把命中的维度、id 与 payload 返回给 Hub

### Requirement: 命中行可按维度删除

意图智能体门户 MUST 提供原文输入与查询。每条命中 MUST 单独占一行，并展示维度、向量 id 与 payload。对已知维度，删除 MUST 调用该维度已有的 `DELETE /voice/admin/api/*-vectors/{id}`，且 MUST NOT 用同一个删除路径删除另一个维度的 id。

#### Scenario: 删除备注命中

- **WHEN** 管理员删除一条 `dimension=remark` 的命中
- **THEN** 门户 MUST 调用 `DELETE /voice/admin/api/remark-vectors/{id}`，其中 id 为该命中的向量 id

#### Scenario: 空输入不查询

- **WHEN** 管理员在原文为空时点击查询
- **THEN** 门户 MUST NOT 发起查询请求

---

### Requirement: 各维向量页提供可复制的种子提示词

意图门户所链的 OP、事件、数量、时间、备注、R·数据源、R·predict、查史播报、U/D 定位管理页，MUST 在批量上传区展示只读提示词，并 MUST 提供复制到剪贴板的操作。提示词 MUST 要求其它模型只输出 JSON 数组，元素含 `document` 与 `payload`，且 payload 字段与枚举与当前 Python 该集合约定一致。除事件维（短说法，payload 为 `event_id`/`event_name`/`kind`）外，document MUST 说明为整段原话。

#### Scenario: 复制预测维提示词含三字段

- **WHEN** 运维打开 R·predict 向量页并复制种子提示词
- **THEN** 文本包含 `predict_channel`、`predict_aspect`、`time_answer` 的合法取值说明
- **AND** 同页批量示例 JSON 含 `predict_aspect`

#### Scenario: 时间与定位禁止项写在提示词内

- **WHEN** 运维复制时间维或定位维提示词
- **THEN** 时间维文本禁止以绝对 Unix 作为主缓存
- **AND** 定位维文本禁止 `historyId`

### Requirement: 类型向量不回到门户

`feeding_type_cues` 管理页 MUST NOT 重新出现在意图智能体门户的飞轮入口中。若该页展示提示词，文案 MUST 标明热路径已停。

#### Scenario: 门户无类型向量卡片

- **WHEN** 运维打开意图智能体门户
- **THEN** 飞轮入口不包含类型向量管理页链接

---

### Requirement: Hub 运维页优先左右分栏

`openspec/project.md` 与 `AGENTS.md` MUST 增加 Hub 运维页布局约定：新增或改造 Hub 静态运维页时，主列表与详情/侧栏内容在宽屏 MUST 优先左右利用空间；禁止将主从内容默认整页纵向堆叠导致长时间下滚（窄屏允许回退为上下布局）。

#### Scenario: 规范文档可检索

- **WHEN** 评审或 AI 查阅全局约定
- **THEN** 可在 `openspec/project.md` 与 `AGENTS.md` 中找到「Hub 运维页布局」相关强制条款，并包含宽屏左右、窄屏可堆叠、后续 proposal/design 须说明空间利用方式等要点

#### Scenario: 后续变更受约束

- **WHEN** 后续 OpenSpec 变更新增或改造 Hub 运维页且存在「列表 + 详情」交互
- **THEN** proposal/design MUST 说明宽屏左右分栏方案（或书面说明为何不适用），评审 MUST 检查是否违反该约定

---

### Requirement: voice Admin 代理定位飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/locate-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/locate-vectors*`。MUST NOT 直连 Chroma 或新建定位飞轮表。MUST NOT 提供 dedupe 代理。批量 payload MUST 透传上游（含 `strategy`）；MUST NOT 在 Go 侧要求或持久化 `historyId`。

#### Scenario: 列定位飞轮

- **WHEN** 运维已通过 Hub Admin 鉴权并请求定位飞轮列表
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/locate-cache` 并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id delete
- **THEN** 系统 MUST 代理至 Python 对应路由，且 MUST NOT 在 Go 侧落库或写入 historyId

### Requirement: Hub U/D 定位飞轮整理页

Hub SHALL 提供 U/D 定位飞轮静态管理页，调用 `/voice/admin/api/locate-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_locate_cues`、document 为整段原文、payload 为 `strategy`（`latest`|`open_timer`|`prefer_completed`），并 MUST 明示禁止 `historyId`；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 从门户进入并整理

- **WHEN** 运维从门户进入定位飞轮页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、bulk 写入与按 id 删除，并展示 strategy 枚举与禁 historyId 提示

---

### Requirement: 长 Hub 页宽屏左右分栏

下列页在宽屏 MUST 将主从或上下长段改为左右（或两列卡片），以减少整页下滚：开通功能管理（磁贴 | 表单区）、各飞轮向量管理页（列表 | 批量上传）、小智 MCP 开通（配置区 | 列表与视频）、运维 Hub 首页与设备数据运维页的多卡片区（宽屏两列）。

#### Scenario: 开通功能管理对照

- **WHEN** 宽屏打开开通功能管理并点选磁贴
- **THEN** 磁贴留在左侧可视区，表单/区段在右侧，无需先滚过整页磁贴带再找表单

#### Scenario: 向量页对照

- **WHEN** 宽屏打开任一 `*-vector-admin` 页
- **THEN** 向量列表与批量上传区左右并列

#### Scenario: 窄屏回退

- **WHEN** 视口低于约 900–1100px 断点
- **THEN** 上述布局可回退为上下堆叠且功能可用

---

### Requirement: voice Admin 代理覆盖缺口列表与删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/miss-inbox` 的分页列表与按 id 删除。Hub 路径 SHALL 使用 `/voice/admin/api/miss-inbox` 与 `/voice/admin/api/miss-inbox/{id}`。列表 MUST 支持查询参数 `offset`、`limit`、`status`、`fail_kind`（空值语义与上游一致）。MUST NOT 在 Go 进程内直读 Python JSONL 或新建 miss 表。MUST NOT 提供 bulk、PATCH dismissed 或从 Hub 新建记录的代理。

#### Scenario: 列出覆盖缺口

- **WHEN** 运维已通过 Hub Admin 鉴权并请求覆盖缺口列表（含可选 status/fail_kind）
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/miss-inbox` 并返回分页结果（透传字段含 utterance、reply、fail_kind、device_no 等）

#### Scenario: 删除覆盖缺口

- **WHEN** 运维调用按 id 删除
- **THEN** 系统 MUST 代理至 Python `DELETE /v1/admin/miss-inbox/{id}`，且 MUST NOT 在 Go 侧改写其它收件箱

### Requirement: Hub 覆盖缺口查删页

Hub SHALL 提供覆盖缺口静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/miss-inbox*`，支持分页浏览、按 status/fail_kind 筛选与单条删除（删除前 MUST 确认）。页面 MUST 说明数据独立于规则缺口与时间飞轮、用途为终局答不出盲区巡检。默认列表 MUST 以 `status=open` 查询（除非运维改选全部）。MUST 登记静态路由与 auth exempt，并 MUST 可从意图智能体门户进入。

#### Scenario: 从门户进入并查删

- **WHEN** 运维从门户打开覆盖缺口页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、筛选与删除，且 MUST NOT 展示飞轮 bulk/dedupe 操作

### Requirement: device_no 链到 history 运维页

覆盖缺口列表中，当某条记录的 `device_no` 非空时，页面 SHALL 将其渲染为指向既有设备 history 运维页的链接，路径为 `/device/admin/history/{deviceNo}`（deviceNo 须做 URL 编码）。`device_no` 为空时 MUST NOT 生成无效链接。

#### Scenario: 有设备号可跳转

- **WHEN** 列表项含非空 `device_no`
- **THEN** 运维点击该设备号 MUST 打开对应 history 运维页

#### Scenario: 无设备号不链

- **WHEN** 列表项 `device_no` 为空
- **THEN** 页面 MUST 以非链接方式展示占位，且 MUST NOT 链到 history 根路径

### Requirement: 意图智能体门户暴露覆盖缺口入口

Hub 意图智能体门户页 SHALL 在规则工单区提供前往覆盖缺口页的入口（可与规则缺口并列）。门户 MUST NOT 将该入口放在飞轮整理主网格作为向量灌种卡。

#### Scenario: 门户工单区可见

- **WHEN** 运维已登录并打开意图智能体门户
- **THEN** 页面 MUST 展示覆盖缺口入口且可进入查删页

---

### Requirement: Hub 运维页统一近全宽

凡同时使用 `pangbao-theme.css` 与 `admin-pages.css` 的 Hub 运维静态页，内容壳 MUST 为近全宽（视口减去约 32px 边距），MUST NOT 再以约 1180px 或 1400px 硬封顶；原 `page-narrow` / `hub-page-wide` MUST 与默认 Hub 宽度一致。

#### Scenario: 宽屏无大块两侧留白

- **WHEN** 运维在宽屏打开任意 Hub 运维页（含统计、功能管理、向量、设备历史等）
- **THEN** 主内容区接近视口宽度，仅保留约 32px 级外边距

#### Scenario: 非 Hub 协议页不受影响

- **WHEN** 用户打开仅使用 theme、不加载 `admin-pages.css` 的协议/官网页
- **THEN** 仍使用 theme 原有 `.page` 宽度约定

---

### Requirement: voice Admin 代理 OP 轨飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/op-cache` 的 list、bulk 与按 id delete，支持分页查询参数（与既有向量缓存同类）。Hub 路径 SHALL 使用 `/voice/admin/api/op-vectors*`。MUST NOT 在 Go 进程内直连 Chroma 或新建 OP 飞轮表。MUST NOT 提供 dedupe 代理（上游无对应路由）。

#### Scenario: 列出 OP 飞轮

- **WHEN** 运维携带合法 Admin 凭据请求 OP 向量列表
- **THEN** voice-service 调用 Python op-cache 列表并返回分页结果
- **AND** 上游失败时返回明确错误（上游空库 total=0 除外）

#### Scenario: bulk 写入 op_rail

- **WHEN** 运维提交含 `document` 与 `payload.op_rail`（create|read|update|delete|exit|other）的 bulk 请求
- **THEN** 系统代理至 Python op-cache bulk 并返回上游汇总

#### Scenario: 删除不存在的向量

- **WHEN** 删除目标 id 在 Python 侧不存在
- **THEN** 系统 MUST 返回失败，MUST NOT 静默成功

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 接口 MUST 拒绝

### Requirement: Hub OP 轨飞轮整理页

Hub SHALL 提供 OP 轨飞轮静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/op-vectors*`，支持列表、bulk 与删除。页面 MUST 说明数据位于 Python `feeding_op_cues`、document 须为整段原文、payload 为 `op_rail`；MUST NOT 展示「整理重复」类 dedupe 操作。该页 MUST 可从意图智能体门户进入，并 MUST 登记静态路由与 auth exempt。

#### Scenario: 门户进入并可列表

- **WHEN** 运维从意图智能体门户进入 OP 轨页且 Admin 已登录
- **THEN** 可见列表（或空态）并由列表接口填充

#### Scenario: 页内 bulk

- **WHEN** 运维在页内提交合法 bulk JSON
- **THEN** 请求经 voice 代理到达 Python op-cache bulk（成功语义以上游为准）

---

### Requirement: Hub SHALL expose admin read-only 7-day predict accuracy by event root

The history-service MUST provide an Admin-authenticated read-only HTTP API that, given a `deviceNo`, returns near-7-calendar-day prediction accuracy aggregated per event root for that device. Scoring MUST match Flutter `trend_predict_accuracy` / `predict-list-accuracy-bars`: a row is an eligible sample only when stored `expectStartTime` (unix seconds) is &gt; day assignment MUST use the actual start instant in timezone `Asia/Shanghai`; day percent MUST be `round(100 * hits.

#### Scenario: Day with samples returns percent

- **WHEN** Admin 查询某 `deviceNo` 且某根在某上海自然日有 2 条合格样本、1 条命中
- **THEN** 该日对该根的 `dailyPercents` MUST 为 `50`（或等价整数 50）

#### Scenario: Day without samples omits percent

- **WHEN** 某根在窗口某日零合格样本
- **THEN** 该日 percent MUST 为 null/省略
- **AND** MUST NOT 返回 `0` 冒充有样本

#### Scenario: Row without expectStartTime is not scored

- **WHEN** history 行缺少正 `expectStartTime`
- **THEN** 该行 MUST NOT 计入任何日的 samples/hits

### Requirement: Hub SHALL expose admin day summary of expect vs actual samples

The history-service MUST provide an Admin-authenticated read-only API to list eligible accuracy samples for one `deviceNo`, one event root (or fallback event id), and one calendar day in `Asia/Shanghai`. Each item MUST include `expectStartTime`, `actualStartTime` (unix seconds), and `hit` (boolean). The API MUST use the same eligibility and hit rules as the 7-day aggregate. Empty day MUST return an empty list (not an error solely for emptiness). history-service MUST 提供.

#### Scenario: Day summary lists hit and miss

- **WHEN** 某日某根有一条命中与一条未命中合格样本
- **THEN** 响应列表 MUST 含两条，且 `hit` 分别为 true 与 false
- **AND** 各条 MUST 带对应的 expect/actual 开始时刻

### Requirement: Device history Hub page SHALL show a vertical accuracy card list like the App prediction list

html` for a device MUST include a dedicated「预测准确率」section (separate from the Redis「预测上报」card) that presents a vertical list of event-root cards loaded from the Admin accuracy API. Each card MUST show near-7-day accuracy bars (height 0–100) with integer `%` labels on days that have percents, and MUST NOT draw a bar or `0%` on null days. On wide viewports the page MUST use a side-by-side layout (card list + day summary). Tapping a day bar with samples MUST show that day’s.

#### Scenario: Bars mirror API null days

- **WHEN** API 某日 percent 为 null
- **THEN** Hub 该日槽位 MUST 不显示柱与 `0%`

#### Scenario: Tap bar shows summary columns

- **WHEN** 运维点击某卡有样本的日柱
- **THEN** 小结区 MUST 列出预计开始、实际开始、是否命中
- **AND** MUST NOT 仅依赖跳转其它路由作为主呈现

---

### Requirement: Voice Admin 只读预测上报

系统 SHALL 在 voice-service 提供管理端只读接口，按 `deviceNo` 返回当前 Redis 中的预测临近待办与进行中 ETA，且 MUST 使用与既有 voice Admin 相同的口令校验（`X-Admin-Password`）。 接口 MUST 复用既有 Load 语义：缺键、空载荷或非法 JSON 时 MUST 返回空列表而非仅因此报错。 接口 MUST NOT 修改 Redis、MUST NOT 投递延时 MQ、MUST NOT 发送推送。 响应字段 MUST 与 App 同步条目对齐：

- 临近待办：`events[]` 含 `eventId`、`nextAt`、`title` 及可选 `expectKind`、`expectDurationSec`、`expectAmount`、`expectUnit`、`expectLeafTitle`
- 进行中 ETA：`items[]` 含 `rootEventId`、`leafEventId`、可选 `historyId`、可选.

#### Scenario: 读取临近待办有数据

- **WHEN** 运维已通过 Hub Admin 鉴权，且该 `deviceNo` 的 imminent Redis 键含合法列表
- **THEN** `GET /voice/admin/api/predict/imminent/pending?deviceNo=` 返回 `code=0` 且 `events` 字段与 Redis 条目对齐（含可选 expect 字段）

#### Scenario: 读取进行中 ETA 缺键

- **WHEN** 运维已通过 Hub Admin 鉴权，且该 `deviceNo` 无 active-eta Redis 键
- **THEN** `GET /voice/admin/api/predict/active-eta?deviceNo=` 返回 `code=0` 且 `items` 为空数组

#### Scenario: 口令错误拒绝

- **WHEN** 请求缺少有效 `X-Admin-Password`（含未注入的直连）
- **THEN** 系统 MUST 拒绝该 Admin 预测只读请求（未授权），且 MUST NOT 返回 Redis 列表正文

### Requirement: 设备数据运维预测上报卡片

Hub 设备数据运维页（`resource/public/history.html`）SHALL 提供一张「预测上报」卡片，内含两段：临近待办与进行中 ETA。页面 MUST 使用既有 Admin JWT（`AdminCommon.adminFetch`）调用上述 voice Admin 只读接口，MUST NOT 在浏览器持有或发送 internal secret。卡片 MUST 完整展示各段全部契约字段；时间类字段 MUST 以可读本地时间展示（MAY 同时保留 unix 秒便于核对）。空列表 MUST 明示暂无上报。卡片 MUST 提供刷新；进入页面且 Admin 已登录时 MUST 自动拉取当前页 `deviceNo` 的两段数据。卡片 MUST NOT 提供清空或改写 Redis 的操作。

#### Scenario: 一张卡两段展示

- **WHEN** 运维打开 `/device/admin/history/{deviceNo}` 且 Admin 已登录
- **THEN** 页面展示一张预测上报卡片，内含临近待办与进行中 ETA 两段，并自动请求对应 Admin GET

#### Scenario: 空列表提示

- **WHEN** 任一 Admin GET 返回空 `events` 或空 `items`
- **THEN** 对应段 MUST 显示暂无上报类提示，且 MUST NOT 将空列表表现为系统故障

#### Scenario: 刷新两段

- **WHEN** 运维点击预测上报卡片的刷新
- **THEN** 页面 MUST 重新请求临近待办与进行中 ETA 两段数据并更新展示

---

### Requirement: voice Admin 代理 R·predict 飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/predict-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/predict-vectors*`。MUST NOT 直连 Chroma 或新建预测飞轮表。MUST NOT 提供 dedupe 代理。批量 payload MUST 透传上游（含 `predict_channel` 与 `time_answer`）。

#### Scenario: 列出预测飞轮

- **WHEN** 运维已通过 Hub Admin 鉴权并请求预测飞轮列表
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/predict-cache` 并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id delete
- **THEN** 系统 MUST 代理至 Python 对应路由，且 MUST NOT 在 Go 侧落库

### Requirement: Hub R·predict 飞轮整理页

Hub SHALL 提供 R·predict 飞轮静态管理页，调用 `/voice/admin/api/predict-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_predict_cues`、document 为整段原文、payload 为 `predict_channel`（`active`|`next`）与 `time_answer`（`clock`|`remain`）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 从门户进入并整理

- **WHEN** 运维从门户进入 R·predict 飞轮页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、bulk 写入与按 id 删除，并展示两维枚举提示

---

### Requirement: voice Admin 代理数量飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/qty-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/qty-vectors*`。MUST NOT 直连 Chroma 或新建数量飞轮表。MUST NOT 提供 dedupe 代理。

#### Scenario: 列出数量飞轮

- **WHEN** 运维携带合法 Admin 凭据请求数量向量列表
- **THEN** voice-service 调用 Python qty-cache 列表并返回分页结果

#### Scenario: bulk 写入 quantity

- **WHEN** 运维提交含 `document` 与正整数 `payload.quantity` 的 bulk 请求
- **THEN** 系统代理至 Python qty-cache bulk 并返回上游汇总

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 接口 MUST 拒绝

### Requirement: Hub 数量飞轮整理页

Hub SHALL 提供数量飞轮静态管理页，调用 `/voice/admin/api/qty-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_qty_cues`、document 为含数量线索的整段原文、无量线索勿灌种；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 门户进入并可列表

- **WHEN** 运维从门户进入数量飞轮页且 Admin 已登录
- **THEN** 可见列表（或空态）并由列表接口填充

---

### Requirement: voice Admin 代理 R·数据源飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/read-source-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/read-source-vectors*`。MUST NOT 直连 Chroma 或新建 read-source 表。MUST NOT 提供 dedupe 代理。

#### Scenario: 列出 R·source 飞轮

- **WHEN** 运维携带合法 Admin 凭据请求 R·数据源向量列表
- **THEN** voice-service 调用 Python read-source-cache 列表并返回分页结果

#### Scenario: bulk 写入 read_source

- **WHEN** 运维提交含 `document` 与 `payload.read_source`（`record` 或 `predict`）的 bulk 请求
- **THEN** 系统代理至 Python read-source-cache bulk 并返回上游汇总

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 接口 MUST 拒绝

### Requirement: Hub R·数据源飞轮整理页

Hub SHALL 提供 R·数据源飞轮静态管理页，调用 `/voice/admin/api/read-source-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_read_source_cues`、document 为整段原文、payload 为 `read_source`（record=查史 / predict=预测）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 门户进入并可列表

- **WHEN** 运维从门户进入 R·数据源页且 Admin 已登录
- **THEN** 可见列表（或空态）并由列表接口填充

---

### Requirement: voice Admin 代理备注飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/remark-cache` 的 list、bulk 与按 id delete。Hub 路径 SHALL 使用 `/voice/admin/api/remark-vectors*`。MUST NOT 直连 Chroma 或新建备注飞轮表。MUST NOT 提供 dedupe 代理。批量条目 payload MUST 透传上游（含允许空字符串的 `remark`）。

#### Scenario: 列出备注飞轮

- **WHEN** 运维已通过 Hub Admin 鉴权并请求备注飞轮列表
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/remark-cache` 并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id delete
- **THEN** 系统 MUST 代理至 Python 对应路由，且 MUST NOT 在 Go 侧落库

### Requirement: Hub 备注飞轮整理页

Hub SHALL 提供备注飞轮静态管理页，调用 `/voice/admin/api/remark-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_remark_cues`、document 为整段原文、payload 为 `remark`（允许空串表示已确认无明细）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 从门户进入并整理

- **WHEN** 运维从门户进入备注飞轮页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、bulk 写入与按 id 删除

---

### Requirement: voice Admin 代理规则缺口列表

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `GET /v1/admin/rule-gaps`，支持 `offset`、`limit`、`dimension`、`status` 查询参数，并将分页结果返回给 Hub。MUST NOT 在 Go 进程内直读 Python JSONL 文件或新建缺口表。

#### Scenario: 列出 open 缺口

- **WHEN** 运维携带合法 Admin 凭据请求规则缺口列表且 `status=open`（或默认）
- **THEN** 系统返回 `total` 与 `items`（含 dimension、segment/整句、reason、创建时间等 Python 已有字段）
- **AND** 请求失败时返回明确错误（不得假装空成功掩盖上游故障——空文件导致 total=0 除外）

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 列表接口 MUST 拒绝

### Requirement: voice Admin 代理标已处理

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `PATCH /v1/admin/rule-gaps/{id}`，将缺口 `status` 更新为 `dismissed`（或请求体指定的合法状态）。

#### Scenario: 标 dismissed

- **WHEN** 运维对存在的缺口 id 提交标已处理
- **THEN** 系统代理成功后该条在默认 open 列表中不再出现（除非筛选包含 dismissed）

#### Scenario: 缺口不存在

- **WHEN** 目标 id 在 Python 侧不存在
- **THEN** 系统 MUST 返回失败（透传或映射为业务错误），MUST NOT 静默成功

### Requirement: Hub 规则缺口管理页

系统 SHALL 在运维 Hub 提供独立「规则缺口」管理页（导航可发现，与意图向量缓存并列），调用上述 Admin API 展示缺口列表，支持按维度/状态筛选与分页，并支持将条目标为已处理。页面 MUST 提示数据来自 Python 收件箱、整理规则须改代码发版而非热更新。

#### Scenario: 进页可见列表

- **WHEN** 运维已登录 Hub 并打开规则缺口页
- **THEN** 可见缺口表格（或空态）并由列表接口填充

#### Scenario: 行内标已处理

- **WHEN** 运维对某条 open 缺口确认标已处理
- **THEN** 页面调用代理 PATCH 并刷新列表

---

### Requirement: 规则缺口页默认 TIME/QTY 展示语义

Hub「规则缺口」页在运维未指定维度筛选时，MUST NOT 向列表接口传递 `dimension` 查询参数（从而依赖 Python 默认仅返回 `time` 与 `quantity`）。页面文案 MUST 标明默认展示 TIME/QTY，MUST NOT 再将「空维度」描述为「全部维度」或引导以 type/event/shape 为默认工单。

#### Scenario: 默认进页不传 dimension

- **WHEN** 运维打开规则缺口页且未填写/未选择具体维度
- **THEN** 列表请求省略 `dimension` 参数
- **AND** 页面提示默认仅为 TIME/QTY

#### Scenario: 显式筛选单一维度

- **WHEN** 运维选择或填写 `time` 或 `quantity`（或排障填写 legacy 维度）
- **THEN** 列表请求携带对应 `dimension`
- **AND** 结果按上游过滤返回

### Requirement: 存量 JSONL 清理不经 Go Hub 热路径

本能力 MUST NOT 要求 Go 进程读写 Python `gaps.jsonl`，MUST NOT 新增 Hub「清理存量 JSONL」热按钮或 Go purge API。存量 legacy 维度行的清理若执行，SHALL 由运维在 Python 数据卷按 runbook 可选完成（清空或只保留 time/quantity）。

#### Scenario: Hub 无 JSONL 清理动作

- **WHEN** 运维使用规则缺口管理页
- **THEN** 页面提供列表/筛选/标已处理等既有能力
- **AND** MUST NOT 出现触发 Go 删除/改写 Python JSONL 文件的操作入口

---

### Requirement: 使用统计按用户为左右 master–detail

客户端使用统计与 App API 使用统计页在「按用户」视图下，宽屏 MUST 将 wx 列表（含分页）置于左侧，将选中用户的汇总与明细/时间线置于右侧；点击列表行后 MUST 仅刷新右侧内容，MUST NOT 依赖用户整页下滚才能看到详情。

#### Scenario: 宽屏选中用户看详情

- **WHEN** 运维在宽屏打开上述任一统计页的「按用户」Tab 并点击某一 wx 行
- **THEN** 右侧展示该用户对应明细（client：功能汇总与操作时间线；api：API 调用明细），且列表仍留在可视区左侧

#### Scenario: 窄屏回退堆叠

- **WHEN** 视口宽度低于约定断点（约 900px）
- **THEN** 列表与详情可上下堆叠展示，功能仍可用

### Requirement: 使用统计按功能明细亦为右侧分栏

两页「按功能」视图中，功能表选中后的明细面板（功能下用户列表或 API 详情）宽屏 MUST 展示在列表右侧，而非整页底部独占纵向空间。

#### Scenario: 宽屏选中功能看明细

- **WHEN** 运维在宽屏「按功能」Tab 点击某一功能/API 行
- **THEN** 右侧展示该功能对应明细面板，左侧功能列表保持可见

### Requirement: 统计页内容区可容纳双栏

上述两统计页 MUST 使用足以容纳左右分栏的内容宽度（相对 Hub 默认约 1080px 放宽），且 MUST NOT 改变统计 API 请求参数与响应语义。

#### Scenario: 布局变更不改 API

- **WHEN** 运维筛选 days、分页或打开用户/功能详情
- **THEN** 请求 path/query 与变更前一致，仅 DOM 布局与样式变化

---

### Requirement: voice Admin 代理时间飞轮列表与写入删除

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `/v1/admin/time-cache` 的 list、bulk 与按 id delete，支持分页查询参数（与既有向量缓存同类）。Hub 路径 SHALL 使用 `/voice/admin/api/time-vectors*`。MUST NOT 在 Go 进程内直连 Chroma 或新建时间飞轮表。MUST NOT 提供 dedupe 代理（上游无对应路由）。批量 payload MUST 透传上游；配方合法性校验由 Python 保证。

#### Scenario: 列出时间飞轮

- **WHEN** 运维已通过 Hub Admin 鉴权并请求时间飞轮列表
- **THEN** voice-service MUST 将请求转发至 Python `/v1/admin/time-cache` 并返回分页结果

#### Scenario: 批量写入与删除

- **WHEN** 运维调用 bulk 或按 id delete
- **THEN** 系统 MUST 代理至 Python 对应路由，且 MUST NOT 在 Go 侧落库或改写配方语义

### Requirement: Hub 时间飞轮整理页

Hub SHALL 提供时间飞轮静态管理页，经 `AdminCommon.adminFetch` 调用上述 `/voice/admin/api/time-vectors*`，支持列表、bulk 与删除。页面 MUST 说明集合为 `feeding_time_cues`、document 为整段原文、payload 为相对时间配方（禁止以绝对 Unix 作主缓存）；MUST NOT 展示 dedupe。MUST 可从意图智能体门户进入并完成静态路由与 auth exempt 登记。

#### Scenario: 从门户进入并整理

- **WHEN** 运维从门户进入时间飞轮页且 Admin 已登录
- **THEN** 页面 MUST 可完成列表浏览、bulk 写入与按 id 删除，并展示相对配方灌种提示

---

### Requirement: voice Admin 代理类型向量列表

系统 SHALL 提供经 voice Admin 口令鉴权的接口，代理 Python `GET /v1/admin/type-cache`，支持分页查询参数（与意图向量缓存同类），并将结果返回给 Hub。MUST NOT 在 Go 进程内直连 Chroma 或新建类型向量表。

#### Scenario: 列出类型向量

- **WHEN** 运维携带合法 Admin 凭据请求类型向量列表
- **THEN** 系统返回分页结果（含 id、document、payload、quality 或 Python 已有等价字段、创建时间等）
- **AND** 上游失败时返回明确错误（不得假装空成功掩盖故障——上游空库 total=0 除外）

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 voice Admin 凭据
- **THEN** 列表接口 MUST 拒绝

### Requirement: voice Admin 代理类型向量 bulk / dedupe / delete

系统 SHALL 提供经 voice Admin 口令鉴权的接口，分别代理：

- `POST /v1/admin/type-cache/bulk`（批量写入；同 document 跳过策略由上游保证）
- `POST /v1/admin/type-cache/dedupe`
- `DELETE /v1/admin/type-cache/{id}`

Hub 路径命名可使用 `/voice/admin/api/type-vectors*`，但 MUST 映射到上述 Python 契约。

#### Scenario: bulk 写入

- **WHEN** 运维提交含 `document` 与 `payload`（含 `target_type`/`op`/`history_mode`）的 bulk 请求
- **THEN** 系统代理至 Python 并返回上游汇总（如 added/skipped/failed）

#### Scenario: 删除不存在的向量

- **WHEN** 删除目标 id 在 Python 侧不存在
- **THEN** 系统 MUST 返回失败，MUST NOT 静默成功

### Requirement: Hub 类型向量管理页

系统 SHALL 在运维 Hub 提供独立「类型向量」管理页（导航可发现，与意图向量缓存并列），调用上述 Admin API 完成列表、批量写入、去重与删除。页面 MUST 提示 payload 语义为类型飞轮字段（非事件短说法），数据在 Python 类型向量库。

#### Scenario: 进页可见列表

- **WHEN** 运维已登录 Hub 并打开类型向量页
- **THEN** 可见向量表格（或空态）并由列表接口填充

#### Scenario: 页面可 bulk 与删除

- **WHEN** 运维在页内执行 bulk 或删除操作
- **THEN** 页面调用对应代理 API 并刷新或反馈结果

---

### Requirement: Usage stats user detail SHALL default to history tab with usage detail available

On Hub pages `api-usage-stats` and `client-usage-stats`, when the「按用户」panel shows a selected wx in the master–detail layout, the detail pane MUST provide two tabs: **历史事件** (default selected) and **使用明细**. Selecting **使用明细** MUST show the page’s previous per-user detail content (API call list on api-usage-stats; hub-md` MUST remain. 「按用户」选中某 wx 时，右侧 MUST 提供 Tab：**历史事件**（默认选中）与 **使用明细**。 使用明细 MUST 保留该页原有人均详情。 hub-md` 窄屏回退 MUST 保持。

#### Scenario: Default tab is history

- **WHEN** 运维在按用户 Tab 点击某一 wx 行
- **THEN** 右侧 MUST 默认展示「历史事件」Tab 为选中态

#### Scenario: Switch back to usage detail

- **WHEN** 运维切换到「使用明细」
- **THEN** 系统 MUST 展示该页原先的人均使用明细内容（或等价数据）

### Requirement: History tab SHALL load device history via existing admin list without time filter

When the history tab is active for a selected user row that has a non-empty `deviceNo`, the page MUST fetch via existing `GET /device/admin/api/history/list` using only `deviceNo`, `page`, and `pageSize` (MUST NOT send `startTime`/`endTime` or otherwise require server API changes). History pagination page size MUST be fixed at **50**. The table MUST present columns aligned with the Hub device history page list (at least: ID, 事件名, 数量/时长, 开始时间, 结束时间, 备注) in read-only form.

#### Scenario: Fetch without time bounds

- **WHEN** 选中一带 `deviceNo` 的用户且历史 Tab 激活
- **THEN** list 请求 MUST 含 `deviceNo` 与 `pageSize=50`，MUST NOT 依赖 startTime/endTime 查询参数

#### Scenario: Columns match history page

- **WHEN** 历史列表有数据
- **THEN** 表头 MUST 含事件名、数量/时长、开始时间、结束时间、备注（或与 history.html 历史表等价列集合，不含操作）

#### Scenario: No device number

- **WHEN** 选中行 `deviceNo` 为空
- **THEN** 历史 Tab MUST 显示无设备提示且 MUST NOT 发起有效 list 拉取

### Requirement: User list defaults SHALL be baby-named, page size 100, and days 3

On both usage stats pages, the「按用户」controls MUST default to: baby-name filter **有宝宝名** (`yes`), wx list page size **100**, and usage days select **3**. Operators MUST still be able to change these. Baby-name filtering MAY remain client-side on the currently loaded page. Usage `days` MUST NOT be required to filter the embedded history list.

两页按用户控件 MUST 默认：宝宝名=有宝宝名、每页=100、days=3。宝宝名 MAY 保持当前页客户端过滤。usage `days` MUST NOT 作为嵌入历史列表的时间过滤条件。

#### Scenario: Fresh load defaults

- **WHEN** 运维打开 api 或 client 使用统计页并进入按用户
- **THEN** 宝宝名筛选 MUST 为「有宝宝名」、每页 MUST 为 100、时间范围 MUST 为近 3 天（除非另有覆盖）

---

### Requirement: 客户端与 API 使用统计页时间范围含 1/3 天且默认 1 天

`client-usage-stats.html` 与 `api-usage-stats.html` 的时间范围选择器 SHALL 提供以下选项（值分别为请求参数 `days`）：近 1 天（`1`）、近 3 天（`3`）、近 7 天（`7`）、近 30 天（`30`）。页面首次加载时 MUST 默认选中近 1 天。切换选项后发起的列表/明细请求 MUST 携带对应该选项的 `days` 查询参数。MUST NOT 因此变更 wx 列表排序语义或新增自填天数控件。

#### Scenario: 默认近 1 天

- **WHEN** 运维打开客户端使用统计页或 App API 使用统计页
- **THEN** 时间范围 MUST 默认选中近 1 天，且首屏请求的 `days` MUST 为 `1`

#### Scenario: 可选近 3 天

- **WHEN** 运维将时间范围改为近 3 天并触发刷新/重载
- **THEN** 后续统计请求的 `days` MUST 为 `3`

#### Scenario: 保留原 7/30

- **WHEN** 运维查看时间范围选项
- **THEN** 列表 MUST 仍包含近 7 天与近 30 天

---

### Requirement: Admin can upload er_code.png to apk storage

The system MUST accept an Admin multipart upload that saves the image as exactly `er_code. The system MUST reject uploads that are not a PNG when the implementation chooses reject-non-png. The download MUST remain available at `/device/app/apk/er_code. The upload path MUST be classified as a Gateway Admin API path (Admin JWT accepted;

#### Scenario: Overwrite

- **WHEN** Admin uploads a new PNG as the group QR
- **THEN** the storage path contains only `er_code.png` as the group QR asset (prior content replaced)

#### Scenario: Admin JWT accepted on upload

- **WHEN** Admin calls the group QR upload endpoint with a valid Admin JWT
- **THEN** the request MUST NOT be rejected as “admin token 不能访问 App 用户接口”

### Requirement: Admin configures expiry and can preview after expiry

The system MUST allow Admin to set `expires_at` for the group QR in cash. When `expires_at` is 0 or in the past, App catalog MUST NOT expose the URL, but Admin preview MUST still be able to load the current `er_code.png`.

#### Scenario: Expired still previewable

- **WHEN** `expires_at` is in the past
- **THEN** Admin UI can still preview the image and see the expiry, and App catalog omits `inviteGroupQrUrl`

---

### Requirement: Admin predict-pro-ideas list API SHALL paginate and expose wxId and baby name

`GET /device/admin/api/predict-pro-ideas/list` MUST require admin authentication (`X-Admin-Password` or documented equivalent). It SHALL accept `page` (default 1) and `pageSize` (default 20, max 100). Results MUST be ordered by `created_at` descending and include `list`, `total`, `page`, `pageSize`. Each item MUST expose at least `id`, `wxId`, `babyName`, `content`, `displayText`, `likeCount`, `createdAt`.

#### Scenario: Authenticated page

- **WHEN** 管理员请求 list 且口令正确
- **THEN** 系统 SHALL 返回分页 list，项内含 `wxId` 与 `babyName`

#### Scenario: Bad password

- **WHEN** 请求缺少有效 Admin 口令
- **THEN** 系统 MUST 拒绝

### Requirement: Admin SHALL update display text of an idea

`POST /device/admin/api/predict-pro-ideas/update` MUST accept `{ "id": <int64>, "displayText": "<text>" }`. After trim, `displayText` MUST be non-empty and MUST NOT exceed 200 characters. The server SHALL update `display_text` for a non-deleted row and MUST NOT change `content` by default. 管理员 **必须** 能改展示文案。

#### Scenario: Update display text

- **WHEN** Admin 对未删除想法提交合法 displayText
- **THEN** 系统 SHALL 更新 `display_text` 且 App list 随后返回新文案

### Requirement: Admin SHALL delete an idea

`POST /device/admin/api/predict-pro-ideas/delete` MUST accept `{ "id": <int64> }` and soft-delete the row. Subsequent App list MUST exclude it. 管理员 **必须** 能删除想法。

#### Scenario: Delete hides from App

- **WHEN** Admin 删除想法 id=N
- **THEN** App list MUST NOT 再包含 id=N

### Requirement: Admin predict-pro-ideas page SHALL mirror feedback-records pattern

The system MUST serve `resource/public/predict-pro-ideas. The page MUST load `admin-pages. 宽屏若同时存在列表与编辑区 MUST 左右分栏或等价避免整页长滚。 The page MUST use Admin JWT（`sessionStorage. Hub 入口 MUST 登记于 `resource/public/admin-modules. html` MUST NOT 再保留独立的「预测 Pro 想法」中间卡片作为入口。 gateway-app MUST 继续注册 HTML 于 `admin_static_pages` 且 Bearer 豁免 **HTML shell path only**。

#### Scenario: Open management page from top nav

- **WHEN** 管理员已登录主 Hub 并点击顶栏「预测 Pro 想法」
- **THEN** 浏览器 SHALL 进入 `/device/admin/predict-pro-ideas`，且在有效 Admin token 下 SHALL 请求 list 并展示想法表

#### Scenario: No mid-page card entry

- **WHEN** 管理员查看主 Hub 登录后首页
- **THEN** 页面 MUST NOT 再展示「预测 Pro 想法」中间卡片入口；MUST 能通过顶栏按钮进入管理页

### Requirement: This release SHALL NOT model adopted-idea Pro grants

The server MUST NOT add an「已采纳」persistence field or automatic one-month Pro entitlement grant for idea authors in this change. 本版 **不得** 实现采纳态或自动发 Pro。

#### Scenario: No adopt API

- **WHEN** 本变更交付
- **THEN** Admin/App API MUST NOT 提供采纳标记或自动履约接口作为本能力一部分

---

### Requirement: 运维 MUST 能分页查看并删除 push_device

系统 MUST 在 push-service 提供经 Admin 鉴权的管理 API，用于分页列出 `push_device` 行（至少含 `id`、`wxId`、`channel`、`deviceKey`、`token`、`updatedAt`），并支持按行 `id` 删除。列表 MUST 支持合理分页；MAY 支持按 `wxId`、`channel` 或 token 前缀筛选。删除成功后该行 MUST 不再参与后续按 wx 扇出推送。API MUST 经 gateway-app 反代暴露（建议前缀 `/push/admin/api/`），MUST 使用与现网一致的 Admin JWT + 下游 `X-Admin-Password` 注入模式；浏览器 MUST NOT 直接携带运维口令头。

#### Scenario: 列表可见

- **WHEN** 已登录运维人员请求设备列表接口且鉴权通过
- **THEN** 响应 MUST 返回当前 `push_device` 分页数据（字段至少含 id 与 wxId、channel、token、updatedAt）

#### Scenario: 手动删除

- **WHEN** 运维人员对某 `id` 调用删除且鉴权通过
- **THEN** 该行 MUST 从 `push_device` 移除，后续对该原 `wxId` 的推送若无其它 token 则 MUST 跳过该设备投递

### Requirement: 运维 Hub MUST 提供推送设备管理入口与静态页

系统 MUST 在 gateway-app 注册推送设备管理静态页（建议 `/device/admin/push-device-admin.html`），并在运维 Hub 模块列表（`admin-modules`）中提供可点击入口。页面 MUST 使用 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch` 访问上述 Admin API，支持查看列表与确认后删除。页面展示 token 时 MUST 截断或脱敏，MUST NOT 以明文完整 token 为默认展示（MAY 提供展开，但默认截断）。

#### Scenario: Hub 可进入

- **WHEN** 运维人员打开运维 Hub 并已登录
- **THEN** MUST 能通过登记入口打开推送设备管理页并加载列表（在 API 可用前提下）

---

### Requirement: sim admin API SHALL expose per-task AI model catalog

`GET /sim/admin/api/status` 与 `GET /sim/admin/api/runtime` 响应 MUST 含 `taskAiModels`（object）：键为调度任务名（至少含 `register`、`comment`、`post_image`、`post_video_submit`、`post_debate`、`debate_comment`、`chat_scan`、`follow`），值为数组。**MUST NOT** 含 `video_poll` 键。

数组元素 MUST 含 `laneKey`、`usage`（可选）、`provider`、`model`。`post_video_submit` MUST 含 simText 与 simVideoGen 相关条目。`post_debate` 与 `debate_comment` MUST 含 simText 条目。

#### Scenario: Status includes T7 T8 models

- **WHEN** GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_debate` 与 `taskAiModels.debate_comment` MUST 存在且含 simText

#### Scenario: Status returns AI for post video task

- **WHEN** 已鉴权 GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_video_submit` MUST 含 simText 与 simVideoGen 信息且 MUST NOT 存在 `video_poll` 键

### Requirement: sim admin UI SHALL display AI models per scheduled task

`sim-admin.html` 任务状态表 MUST 增加「AI 模型」列。每行 MUST 根据 `taskAiModels[taskName]` 展示：

- 用途说明（若有 `usage`）
- lane 键（如 `simText`）
- `provider/model`；未配置时 MUST 可读提示（如「未配置」）

无 LLM 的任务 MUST 显示「—」。页面 MUST 说明修改模型须至 ai-model-admin，本页 MUST NOT 提供 lane 编辑。

#### Scenario: Task table shows AI column

- **WHEN** 管理员打开 sim-admin 并加载状态
- **THEN** 表格 MUST 含「AI 模型」列且 T1 行展示 simText 与 simImageGen 信息

#### Scenario: Refresh reflects lane change

- **WHEN** 管理员在 ai-model-admin 修改 simText model 后点击 sim-admin「刷新状态」
- **THEN** 任务表中依赖 simText 的行 MUST 展示新 model

#### Scenario: No lane editor on status table

- **WHEN** 管理员查看任务状态表
- **THEN** UI MUST NOT 提供修改 provider/model 的输入控件

### Requirement: sim admin web SHALL be served from gateway-app

系统 MUST 在 `gateway-app-server` 注册静态页 `/device/admin/sim-admin.html`（`resource/public/sim-admin.html`）。页面 MUST 要求管理员已在运维 Hub 登录（与 ucg-admin 一致的鉴权模式）。

#### Scenario: Admin page reachable

- **WHEN** 管理员已登录 Hub 并访问 `/device/admin/sim-admin.html`
- **THEN** 浏览器 MUST 加载模拟管理界面

### Requirement: sim admin API SHALL expose config and prompts

`sim-user-service` MUST 提供 Admin HTTP API（经 gateway 反代或直连，鉴权与 ucg-admin 对齐）：

- `GET/PUT /sim/admin/api/config` — 字段至少含 `enabled`（bool）、`maxSimUsers`（int，默认 100）
- `GET/PUT /sim/admin/api/prompts/{taskType}` — `taskType` 至少含：`register_nickname`、`register_avatar`、`comment`、`post_image_text`、`post_video_text`、`post_debate_text`、`debate_comment`、`chat_reply`； 每项含 `systemPrompt`、`userPromptTemplate`
- `GET /sim/admin/api/status` — 各任务上次运行时间、成功/失败计数、pending video job 数
- `GET.

#### Scenario: Debate prompts editable

- **WHEN** GET `/sim/admin/api/prompts/debate_comment`
- **THEN** MUST 返回 userPromptTemplate 且保存后下一次 T8 MUST 使用新模板

#### Scenario: Update max sim users

- **WHEN** 管理员 PUT `maxSimUsers=50`
- **THEN** 下一次 T1 MUST 在 sim 用户数 ≥50 时停止注册

#### Scenario: Update comment prompt

- **WHEN** 管理员修改 `comment` 的 `userPromptTemplate`
- **THEN** 下一次 T2 MUST 使用新模板渲染变量

### Requirement: sim admin API SHALL expose read-only runtime snapshot

`sim-user-service` MUST 提供 `GET /sim/admin/api/runtime`（鉴权与现有 sim-admin API 一致）。 响应 MUST 包含字段，反映**当前 DB 生效值**（非 env 只读镜像）：

- `serviceEnabled`（bool）— 对应 `SIM_USER_SERVICE_ENABLED`（env 硬闸）
- `dbEnabled`（bool）— 对应 `sim_config. enabled`
- `databaseName`（string）— 自 `SIM_DB_LINK` 解析的库名，MUST NOT 含账号密码或 host
- `simUserCount`（int）— 当前 `is_simulated=1` 用户数； 值为字符串时长
- `rateLimitRps`（number）、`rateLimitBurst`（int）

响应 MUST NOT 含 DSN 凭据、`SIM_ADMIN_PASSWORD`、`GLM_API_KEY` 或默认登录密码明文。

#### Scenario: Admin reads runtime from DB

- **WHEN** 已鉴权管理员 GET `/sim/admin/api/runtime` 且 DB 中 comment 周期为 3h
- **THEN** 响应 `intervals.comment` MUST 为 3h 且 MUST NOT 依赖 env 覆盖

#### Scenario: Runtime excludes secrets

- **WHEN** 管理员 GET `/sim/admin/api/runtime`
- **THEN** 响应 body MUST NOT 包含 `@tcp(` 连接串或 `password` 字段

### Requirement: sim admin UI SHALL show structured task status

`sim-admin.html` MUST 将 `GET /sim/admin/api/status` 结果以结构化形式展示（表格或等价列表），字段至少含：任务名、上次运行时间、成功次数、失败次数、最近错误；并单独展示 `pendingVideoJobs`。

原始 JSON dump  alone MUST NOT 作为唯一展示方式（可保留「查看原始 JSON」折叠为可选）。

#### Scenario: Task status table

- **WHEN** 管理员加载页面或点击刷新状态
- **THEN** 各 `status.tasks` 条目 MUST 以可读表格行展示而非仅 `JSON.stringify` 整块输出

### Requirement: sim admin UI SHALL edit runtime configuration with save effect feedback

`sim-admin.html` MUST 提供可编辑运行时配置表单，字段覆盖 `taskSwitches`、`intervals`、`rateLimitRps`、`rateLimitBurst`、`ephemeralChatLoop`、`ephemeralChatWindow`（及既有 `enabled`、`maxSimUsers`）。保存 MUST 调用扩展后的 **`PUT /sim/admin/api/config`**。保存成功后 MUST 展示 API 返回的 **`effects`** 与 **`taskSchedule`**（含各任务「立即生效 / 预计下一跑」提示）。页面 MUST 区分 `serviceEnabled`（env，说明须改 env 并 recreate）与 `dbEnabled`（可在线保存）。页面 MUST NOT 提供 sim LLM lane 编辑（链至 ai-model-admin）。`GET /sim/admin/api/status` 结构化任务状态展示 MUST 保留。

#### Scenario: 可编辑任务开关

- **WHEN** 管理员在 sim-admin 取消勾选 chat 并保存
- **THEN** 页面 MUST 调用 PUT config 且 MUST 展示保存结果面板含 schedule 提示

#### Scenario: 保存后展示下一跑提示

- **WHEN** PUT config 返回 `taskSchedule` 含 postImage 的 `nextRunHint`
- **THEN** sim-admin MUST 在保存结果区展示该提示

#### Scenario: Interval inputs are editable

- **WHEN** 管理员查看运行配置区块
- **THEN** UI MUST 提供周期间隔输入控件（非只读文本）

### Requirement: sim config PUT response taskSchedule SHALL reflect effective scheduling state

`PUT /sim/admin/api/config` 响应中的 `taskSchedule[]` 每一项 MUST 含：

- `configEnabled`（bool）— `runtime_json` 中对应任务开关（配置层）
- `enabled`（bool）— **自动调度实际是否启用**：MUST 等于 `configEnabled && dbEnabled && serviceEnabled`，其中 `dbEnabled` 为保存后的 `sim_config.enabled`，`serviceEnabled` 为当前进程 `SIM_USER_SERVICE_ENABLED`
- `nextRunHint`（string）— 当 `enabled=false` 时 MUST 说明阻塞原因（任务配置关 / 业务总闸关 / 进程总闸关），MUST NOT 在总闸关闭时给出「约 X 后」类下一跑时间

`name`、`label`、`intervalSec`、`lastRunAt` 语义不变。

#### Scenario: DB off and T4 config on shows not effectively enabled

- **WHEN** PUT 保存 `sim_config.enabled=false` 且 `taskPostVideo=true`
- **THEN** `taskSchedule` 中 `post_video_submit` MUST 含 `configEnabled=true` 且 `enabled=false`，且 `nextRunHint` MUST 表明业务总闸关闭

#### Scenario: Env off and task config on

- **WHEN** 进程 `SIM_USER_SERVICE_ENABLED=false` 且 PUT 保存某任务 `configEnabled=true` 与 `sim_config.enabled=true`
- **THEN** 该任务 `enabled=false` 且 `nextRunHint` MUST 表明进程总闸关闭

#### Scenario: All gates on and task on

- **WHEN** `serviceEnabled=true`、`dbEnabled=true`、任务 `configEnabled=true`
- **THEN** `enabled=true` 且 `nextRunHint` MAY 基于 `lastRunAt` 与周期估算下一跑

### Requirement: sim config PUT effects SHALL note scheduler blocked when gates closed

当保存后存在 `configEnabled=true` 的任务且（`dbEnabled=false` 或 `serviceEnabled=false`）时，`effects[]` MUST 含可读提示：任务开关已保存但自动调度未启动，并 MAY 说明可手动执行任务。

#### Scenario: Effects on DB disabled save

- **WHEN** PUT 将 `enabled=false` 且至少一任务配置开关为 true
- **THEN** `effects` MUST 含业务总闸相关提示

### Requirement: sim admin save result UI SHALL distinguish config vs effective task state

`sim-admin.html` 保存结果中的 `taskSchedule` 表格 MUST 区分 **配置开关**（`configEnabled`）与 **自动调度是否生效**（`enabled`）。当 `enabled=false` 时 UI MUST NOT 仅展示「开」而暗示任务已在运行。

#### Scenario: Save result shows effective off for T4

- **WHEN** 管理员保存 DB 关、T4 配置开
- **THEN** 保存结果表 MUST 显示 T4 配置为开、自动调度为关（或等价文案）

### Requirement: sim-admin SHALL expose editable runtime configuration aligned with DB schema

`sim-admin.html` MUST 提供可编辑运行时配置表单，字段覆盖 `taskSwitches`、`intervals`、`rateLimitRps`、`rateLimitBurst`（及既有 `enabled`、`maxSimUsers`）。**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow` 或 E1 相关文案。保存 MUST 调用 **`PUT /sim/admin/api/config`**。保存成功后 MUST 展示 API 返回的 **`effects`** 与 **`taskSchedule`**。页面 MUST 区分 `serviceEnabled`（env）与 `dbEnabled`（可在线保存）。`GET /sim/admin/api/status` 结构化任务状态 MUST 保留；`taskAiModels.chat_scan` usage MUST 为「未读回复」（非「E1 回复」）。

#### Scenario: No E1 fields in admin form

- **WHEN** 管理员打开 sim-admin 运行配置区
- **THEN** MUST NOT 展示 E1 循环/窗口输入框

#### Scenario: Save without ephemeral effects

- **WHEN** PUT 保存 task/interval 变更
- **THEN** 响应 `effects` MUST NOT 含 `ephemeral_may_continue`

#### Scenario: Chat scan task AI label

- **WHEN** GET status 或 runtime 含 `taskAiModels.chat_scan`
- **THEN** lane usage MUST 为「未读回复」

### Requirement: sim-admin runtime panel SHALL display effective task schedule

`GET /sim/admin/api/runtime` 与 status 中 `intervals` 键 MUST 含 `register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`；**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow`、`videoPollIdle`、`videoPollActive`。

`taskSchedule[]`（config PUT 响应与 admin UI）MUST 含 `post_debate`（T7 辩论发帖）与 `debate_comment`（T8 辩论论点）项，含 `configEnabled`、`enabled`、`intervalSec`、`nextRunHint`。

#### Scenario: Admin shows T7 T8 schedule

- **WHEN** GET config 或 PUT 保存后返回 `taskSchedule`
- **THEN** MUST 含 name=`post_debate` 与 name=`debate_comment`

#### Scenario: Runtime intervals without ephemeral

- **WHEN** GET runtime
- **THEN** `intervals` MUST NOT 含 E1 键

### Requirement: sim admin runtime intervals SHALL include T4 inline poll parameters

`GET /sim/admin/api/config` 与 `GET /sim/admin/api/runtime` 的 `intervals` MUST 含：

- `postVideoPollInterval` — T4 提交后 async-result 轮询间隔（字符串 duration）
- `postVideoPollMaxWait` — T4 视频发布最大等待（字符串 duration）

MUST NOT 再返回 `videoPollIdle`、`videoPollActive`。`taskSwitches` MUST NOT 含 `videoPoll`。

`PUT /sim/admin/api/config` MUST 接受并持久化上述字段。

#### Scenario: Config get includes poll intervals

- **WHEN** 管理员 GET `/sim/admin/api/config`
- **THEN** `intervals.postVideoPollInterval` 与 `intervals.postVideoPollMaxWait` MUST 为可读 duration 字符串

### Requirement: sim admin UI SHALL reflect T4 inline video flow

`sim-admin.html` MUST：

- 移除 P1 任务开关、P1 周期输入、P1 手动执行行及 runtime 只读区中的 P1 项
- 在 T4 相关区域提供可编辑 `postVideoPollInterval`、`postVideoPollMaxWait`（保存走 `PUT /sim/admin/api/config`）
- T4 手动「执行」按钮：自点击至 `sim_task_run.post_video_submit.lastRunAt` 更新前 MUST 显示「执行中…」且 disabled；**MUST NOT** 依赖新增 status 字段
- status 轮询对 `post_video_submit` MUST 足够长以覆盖 `postVideoPollMaxWait`（避免过早恢复按钮）

#### Scenario: T4 button stays busy until task run updates

- **WHEN** 管理员手动执行 T4 且视频轮询需 5 分钟完成
- **THEN** 5 分钟内按钮 MUST 保持「执行中…」直至 status 中 lastRunAt 变化

#### Scenario: P1 removed from admin form

- **WHEN** 管理员打开 sim-admin 配置区
- **THEN** MUST NOT 显示 P1 开关或 P1 idle/active 输入

### Requirement: sim-user-admin SHALL expose paginated simulated user list API

`sim-user-service` MUST 提供 `GET /sim/admin/api/users`（鉴权与现有 sim-admin 一致）。 Query 参数 MUST 支持 `page`（默认 1）、`pageSize`（默认 20，最大 200）。 响应 MUST 含 `list`、`total`、`page`、`pageSize`。 `list` 每项 MUST 含：

- `wxId`（int64）
- `account`（string）
- `nickname`（string，无 UCG profile 时为空字符串）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）
- `createdAt`（int64 Unix 秒； 无 credential 时 MUST 为 `123456` 且 `passwordPlainLegacy` MUST 为 true）
- `passwordPlainLegacy`（bool，可选； true.

#### Scenario: Admin lists sim users with profile

- **WHEN** 已鉴权 GET `/sim/admin/api/users?page=1&pageSize=20` 且存在带 UCG profile 的 sim 用户
- **THEN** 响应 `list` MUST 含对应 `nickname` 与非空 `avatarUrl` 或 `avatarThumbnailUrl`（若 profile 有头像）

#### Scenario: Legacy user password fallback

- **WHEN** 列表项 wxId 在 `sim_wx_credential` 无记录（历史 `ptest*` 用户）
- **THEN** 该项 `passwordPlain` MUST 为 `123456` 且 `passwordPlainLegacy` MUST 为 true

#### Scenario: CreatedAt from credential

- **WHEN** T1 注册后写入 `sim_wx_credential`
- **THEN** 列表该项 `createdAt` MUST 等于 credential 的 `created_at`

### Requirement: sim-user-admin SHALL expose simulated user deactivate API

`sim-user-service` MUST 提供 `POST /sim/admin/api/users/{wxId}/deactivate`（鉴权与现有 sim-admin 一致）。路径参数 `wxId` MUST 为正整数。

成功时 MUST：调用 device internal sim deactivate 删除 `wx` 行；删除 sim 库 `sim_wx_credential` 对应行；将该 wxId 的 `sim_video_job` 中 `pending`/`processing` 行标为 `skipped`。

MUST NOT 调用 ucg 或 gateway App API 删除帖子/profile。注销语义 MUST 与 App `POST /device/app/api/user/deactivate` 一致（仅删 wx）。

#### Scenario: Deactivate simulated user success

- **WHEN** 已鉴权 POST `/sim/admin/api/users/1001/deactivate` 且 wxId=1001 为 `is_simulated=1`
- **THEN** HTTP MUST 成功且该 wx 行 MUST 自 device 库删除且 credential 行 MUST 删除

#### Scenario: Reject non-sim wxId

- **WHEN** wxId 存在但 `is_simulated=0`
- **THEN** MUST 返回 4xx 业务错误且 MUST NOT 删除 wx

#### Scenario: Reject invalid wxId

- **WHEN** wxId 不存在或已注销
- **THEN** MUST 返回明确业务错误（已注销或不存在）

### Requirement: sim-admin UI SHALL display simulated user table with deactivate

`sim-admin.html` MUST 在现有页面内嵌入「模拟用户」区块（MUST NOT 要求单独 Hub 模块页）。表格 MUST 展示：头像（优先 `avatarThumbnailUrl`）、UCG 昵称、账号、wxId、注册时间、密码、注销操作。

- 注册时间：`createdAt=0` 时 MUST 显示「—」
- 密码：`passwordPlainLegacy=true` 时 MUST 标注「默认密码（历史）」
- 注销：点击前 MUST `confirm`；成功后 MUST 刷新列表并更新 runtime 区 `simUserCount`

分页 MUST 调用 `GET /sim/admin/api/users`。页面 MUST NOT 提供 LLM lane 编辑（既有约束不变）。

#### Scenario: Admin sees user row

- **WHEN** 管理员已登录 Hub 并打开 sim-admin 且存在 sim 用户
- **THEN** 表格 MUST 展示至少一行含昵称/账号/wxId/密码列

#### Scenario: Deactivate refreshes list

- **WHEN** 管理员确认注销某一 sim 用户且 API 成功
- **THEN** 该行 MUST 从表格消失且 runtime 模拟用户数 MUST 减少

---

### Requirement: cash-service MUST 提供 VIP 商品 Admin 读与写

cash-service MUST 提供 `GET /cash/admin/api/vip/product` 与 `POST /cash/admin/api/vip/product`。 鉴权 MUST 与既有 cash Admin 一致（网关注入 `X-Admin-Password`）。 GET MUST 返回一期 VIP 商品（`productCode` 为 `vip_monthly_19`）的 `title`、`priceFen`、`originalPriceFen`、`durationDays`、`appleProductId`、`status`。 POST MUST 更新该行对应字段并持久化到 `ai_voice_cash. MUST NOT 创建第二行 VIP SKU，MUST NOT 修改 `product_code`。 `priceFen` MUST 为非负整数（分）； gateway-app MUST 经既有 `/cash/admin/api/*` 反代到达，MUST NOT 新增顶层前缀。 本路径为运维通道，MUST NOT.

#### Scenario: 管理员读取 VIP 商品

- **WHEN** 已鉴权管理员请求 `GET /cash/admin/api/vip/product`
- **THEN** 系统 MUST 返回 `vip_monthly_19` 的现价、原价、时长与 Apple 商品 ID

#### Scenario: 管理员更新现价与 Apple 商品 ID

- **WHEN** 已鉴权管理员 POST 合法的 `priceFen` 与 `appleProductId`
- **THEN** 系统 MUST 写入 `vip_product`，且后续 App 读价与建单 MUST 使用新现价，Apple 建单/验单 MUST 使用新 `appleProductId`

#### Scenario: 未鉴权拒绝写

- **WHEN** 请求未携带有效 Admin 凭证
- **THEN** 系统 MUST 拒绝 POST 且 MUST NOT 修改 `vip_product`

### Requirement: VIP 与功能建单 MUST 只认库内 Apple 商品 ID

`GetActiveProduct`、VIP Apple 验单以及 EnsureSchema 种子 MUST NOT 读取 `CASH_APPLE_PRODUCT_ID` 或 yaml `cash.appleProductId`。VIP 的 Apple 商品 ID 真相源 MUST 仅为 `vip_product.apple_product_id`。当该字段为空时，VIP Apple 建单 MUST NOT 伪造 productId；响应可带说明需在开通功能管理配置的 `payTip`。部署模板（`.env.example`、cash-service compose 环境注入、cash-service 配置注释）MUST 删除 `CASH_APPLE_PRODUCT_ID`。`CASH_APPLE_BUNDLE_ID` MUST 保留（JWS bundle 校验）。功能 SKU 的 Apple ID 仍在 `feature_product`，MUST NOT 改回 env。

#### Scenario: 库中已配置 Apple 商品 ID

- **WHEN** `vip_product.apple_product_id` 为非空且进程环境未设置 `CASH_APPLE_PRODUCT_ID`
- **THEN** `GET /cash/app/api/vip/product` 与 VIP Apple 建单 MUST 返回该库值

#### Scenario: 库中未配置 Apple 商品 ID

- **WHEN** `vip_product.apple_product_id` 为空
- **THEN** 系统 MUST NOT 从环境变量填入 Apple 商品 ID

#### Scenario: 部署模板无 CASH_APPLE_PRODUCT_ID

- **WHEN** 运维按 `.env.example` 与 compose 部署 cash-service
- **THEN** 模板与 compose 注入 MUST NOT 再出现 `CASH_APPLE_PRODUCT_ID`

### Requirement: 开通功能管理页 MUST 可编辑 VIP 套餐

`cash-feature-admin.html` MUST 提供独立于功能定义列表的 VIP 套餐表单，字段至少含现价（分）、原价（分）、时长、Apple 商品 ID；保存 MUST 调用 VIP 商品 Admin 写接口。页面 MUST NOT 把 VIP 渲染为一条可点选的 `feature_def`。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 保存 VIP 价格

- **WHEN** 管理员在开通功能管理填写 VIP 现价并保存
- **THEN** 页面 MUST 经 adminFetch POST Admin VIP 商品接口且成功后主内容反映新值

#### Scenario: VIP 不出现在功能定义列表

- **WHEN** 管理员查看开通功能管理的功能定义列表
- **THEN** 列表 MUST NOT 将 `vip` 或 `vip_monthly_19` 作为功能编号行展示

---

### Requirement: voice-admin HTML SHALL configure voice_ai and clinic_ai quota

系统 MUST 提供独立 Admin 页面 **`/device/admin/voice-admin. 行内修改上限 MUST 调用既有 **`PUT /voice/admin/api/ai-quota/user`**。 页面 MUST **移除**手输 wxId 加载/保存的单人 override 模块（含清除 checkbox UI）。 表格列顺序 MUST 为（左→右）：`deviceNo`、喂养已用、喂养上限、胖宝已用、胖宝上限、`wxId`、`account`、`babyName`。 一行 MUST 同时展示喂养（`voice_ai`）与胖宝（`clinic_ai`）两额度。 页面 MUST 提供按 **`deviceNo`** 查询并触发列表刷新。 页面 MUST NOT 包含润笔（polish）字段。 页面 MUST 保留指向 **`/device/admin/ai-model-admin. html`** 的可见链接（文案含「模型与并发」或等价），MUST NOT 在本页恢复 LLM lane 编辑 Tab。 页面 MUST.

#### Scenario: 管理员修改喂养 AI 全局默认

- **WHEN** 运维在 voice-admin 页提交 voiceAi=5、clinicAi=30
- **THEN** 页面 SHALL 调用 PUT `/voice/admin/api/ai-quota/default` 且 voice-service 配置 SHALL 更新

#### Scenario: 页面不含润笔字段

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示 `polishMonthlyLimit` 输入控件

#### Scenario: 无单人 override 模块

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示手输 wxId 的「单人 override」加载/保存表单

#### Scenario: 额度表列顺序与双额度

- **WHEN** 列表返回至少一行
- **THEN** 表头第一列 SHALL 为 deviceNo，且同一行 SHALL 含喂养已用/上限与胖宝已用/上限

#### Scenario: 按 deviceNo 查询

- **WHEN** 运维输入 deviceNo 并查询
- **THEN** 页面 MUST 请求 `/voice/admin/api/ai-quota/users` 且携带该 deviceNo 过滤参数并刷新表格

#### Scenario: 行内修改上限

- **WHEN** 运维修改某行喂养或胖宝上限并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/ai-quota/user`（含对应 wxId 与限额字段）

#### Scenario: Hub 登录后主面板可见

- **WHEN** 运维已在 `/device/admin` Hub 登录并打开 voice-admin.html
- **THEN** 页面 MUST 展示全局默认与用户额度表（非仅页头标题）

#### Scenario: 链至统一模型页

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 MUST 展示指向 ai-model-admin 的链接且 MUST NOT 含 LLM lane 编辑表单

### Requirement: Admin Hub SHALL link voice-admin alongside ucg-admin

`resource/public/admin-modules.js` MUST 增加 voice-admin 模块入口，模式对齐 ucg-admin：`id: voice-admin`、导航至 `/device/admin/voice-admin.html`。Hub 登录后 MUST 可点击进入 voice-admin。

#### Scenario: Hub 导航可见 voice-admin

- **WHEN** 管理员登录 Admin Hub
- **THEN** 模块列表 SHALL 包含 voice-admin 入口且链接至 `/device/admin/voice-admin.html`

### Requirement: ucg-admin SHALL remove voice and clinic quota fields

`resource/public/ucg-admin.html`「AI 配置」Tab MUST **移除** `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit` 相关表单与 API 字段，**仅保留** `polishMonthlyLimit` 全局默认与 per-wxId override。同一 Tab MUST 扩展 **润笔 lane** 的 `provider`、`maxInFlight`、`maxWaiters` 配置（`visionModel` 作为 polish 的 model 选择），并调用扩展后的 **`/ucg/admin/api/ai-config`**。

#### Scenario: ucg-admin 仅润笔配置

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 SHALL 展示 polish 相关字段（含模型、并发、缓冲池）且 SHALL NOT 调用 voice/clinic 配额 API

---
