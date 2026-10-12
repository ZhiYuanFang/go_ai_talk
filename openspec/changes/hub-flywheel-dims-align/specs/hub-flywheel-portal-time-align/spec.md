## ADDED Requirements

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
