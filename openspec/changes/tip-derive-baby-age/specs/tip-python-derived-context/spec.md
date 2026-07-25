## ADDED Requirements

### Requirement: Tip 月龄由 Python 根据生日自算

在 tip 生成图中，系统 SHALL 在 `fetch_baby_profile`（或等价画像拉取）完成之后，根据宝宝 `birthday` 与当前日期（时区 `Asia/Shanghai`）计算月龄，供提示词及依赖月龄的检索使用。调用方（Go/Flutter）MUST NOT 被要求提供 `baby_age_months` / `babyAgeMonths`。

#### Scenario: 有生日时写入计算出的月龄

- **WHEN** tip 图成功获取到可解析的 `birthday`
- **THEN** 系统 SHALL 按日历月差计算非负整数月龄（未来生日钳为 0）
- **AND** 提示词 SHALL 包含该月龄（例如「宝宝月龄：{n} 个月」）

#### Scenario: 无生日时提示「未知」而非 0 个月

- **WHEN** 画像缺失、无 `birthday` 或生日无法解析
- **THEN** 系统 SHALL NOT 将月龄表示为 `0` 个月以表示「未知」
- **AND** 提示词 SHALL 使用「宝宝月龄：未知」（或等价明确未知文案）
- **AND** 若知识检索依赖月龄，SHALL 跳过月龄过滤或使用不限月龄策略，MUST NOT 用 0 冒充新生儿月龄过滤

### Requirement: Tip 当前时间由 Python 在写提示词时生成

系统 SHALL 在构建 tip 提示词时生成当前时间上下文，时区 MUST 为 `Asia/Shanghai`。Tip 入站请求 MUST NOT 将 `current_time` / `currentTime` 作为必填字段。

#### Scenario: 提示词含上海时区当前时间

- **WHEN** tip 流式生成进入写提示词阶段
- **THEN** 提示词 SHALL 包含基于 `Asia/Shanghai` 的当前时间（可读本地时间；MAY 同时附带 Unix 秒）
- **AND** 该时间 SHALL 来自 Python 运行时时钟，而非请求体字段

#### Scenario: 请求体不再要求 current_time

- **WHEN** 调用方 POST tip stream 且 body **不含** `current_time` / `currentTime`
- **THEN** 系统 SHALL 仍能完成 tip 生成（在其它必填字段合法时）
- **AND** SHALL NOT 因缺少该字段返回 422
