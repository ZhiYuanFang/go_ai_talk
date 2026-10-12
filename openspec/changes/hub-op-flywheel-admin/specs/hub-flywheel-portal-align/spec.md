## ADDED Requirements

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
