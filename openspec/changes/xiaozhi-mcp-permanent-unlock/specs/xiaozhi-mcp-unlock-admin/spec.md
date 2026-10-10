## ADDED Requirements

### Requirement: 小智 MCP SKU 专用表

系统 SHALL 使用独立于 `feature_product` 的数据表存储小智 MCP 永久买断 SKU（至少含商品编码、标题、现价分、原价分、Apple 商品 ID、上下架状态与更新时间）。种子 MUST 提供唯一（或约定的）永久商品编码；`product_code` MUST NOT 经 Admin 改写为任意新码集合（首版以更新种子行为主）。

#### Scenario: 与开通功能管理隔离

- **WHEN** 运维打开开通功能管理的功能 SKU 列表
- **THEN** 列表 MUST NOT 展示小智 MCP 专用 SKU 行

#### Scenario: 配置可被支付读取

- **WHEN** Admin 更新了 MCP SKU 的价格或 `appleProductId` 并保存
- **THEN** 后续 App 建单与 Apple 反查 MUST 使用更新后的值

### Requirement: 独立小 Admin 一页两区

系统 SHALL 在运维 Hub 提供独立于「开通功能管理」的小智 MCP 开通管理页，并在 Hub 导航中可发现。页面 MUST 包含两区：**SKU 配置**与**手工授**。

#### Scenario: 导航可达

- **WHEN** 运维已登录 Hub 并打开模块导航
- **THEN** 可见「小智 MCP 开通」（或同等标题）并进入该独立页

#### Scenario: SKU 区可保存

- **WHEN** 运维在 SKU 区修改价格、原价、Apple 商品 ID 或上下架并提交
- **THEN** 系统持久化专用表并成功返回

#### Scenario: 手工授区授予永久能力

- **WHEN** 运维填写有效 `wxId` 与非空授权理由并提交手工授
- **THEN** 系统为该 wx 授予永久小智 MCP 能力
- **AND** MUST 留下可审查的手工授痕迹（如 admin 订单或等价 channel_ref + grant_reason）

### Requirement: 手工授撤销

系统 SHALL 允许撤销某 wx **最近一笔**且仍有效的小智 MCP **手工授**（`unlock_method=admin`）。撤销后该 wx MUST 视为未开通（除非另有仍有效的支付开通——首版若一人仅一行权益，撤销即未开通）。支付开通 MUST NOT 被本撤销接口当作手工授撤销。

#### Scenario: 撤销最近手工授

- **WHEN** 目标 wx 最近有效开通来自手工授且运维确认撤销
- **THEN** 该 wx 小智 MCP 能力变为未开通

#### Scenario: 无手工授可撤

- **WHEN** 目标 wx 无有效手工授可撤（未开通或仅支付开通）
- **THEN** 撤销 MUST 失败并返回明确业务错误

### Requirement: Admin API 鉴权

小智 MCP 的 Admin 读写接口 MUST 使用与其它 cash Admin 一致的运维鉴权（如 Admin 口令头），MUST NOT 经 App Bearer 匿名可写。

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 Admin 凭据
- **THEN** Admin API MUST 拒绝
