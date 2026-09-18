## ADDED Requirements

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
