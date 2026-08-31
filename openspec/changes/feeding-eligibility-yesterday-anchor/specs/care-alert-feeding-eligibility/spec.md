## MODIFIED Requirements

### Requirement: 客户端 MUST 以 cash 服务端资格替代「昨日有发生」闸门

Flutter 值得留意展示流 MUST 先请求 **cash** care-alert eligibility。当 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 仍展示「值得留意」卡片，并 MUST 使用 eligibility 返回的 `effectiveDays` / `requiredDays` / `remainingDays` **由客户端拼接**进度文案，告知用户已累计的有效日（及/或剩余所需日），且 MUST NOT 调用值得留意生成或日列表数据接口。当 `qualified=true` 时，客户端 MUST 再按原逻辑请求值得留意接口并展示内容。客户端 MUST NOT 再以「上海昨日有发生」作为是否可拉取值得留意数据的前提，MUST NOT 用本地 history 自行判定权威 `qualified`。客户端 MUST NOT 被要求向用户声明「今日不计入」；服务端可选 `message` MUST NOT 作为进度数字的权威来源。本资格门槛的产品目的为促活 / 解锁更多体验，MUST NOT 被实现或文案绑定为「为保证值得留意数据准确而要求昨日有发生」。

#### Scenario: 未合格展示已累计有效日

- **WHEN** eligibility 返回 `qualified=false`、`effectiveDays=1`、`requiredDays=2`、`remainingDays=1`
- **THEN** UI MUST 展示值得留意卡片，并由客户端拼出反映已累计有效日与所需/剩余日的进度提示，且 MUST NOT 请求 care-alert daily/生成接口

#### Scenario: 合格后按原逻辑拉数

- **WHEN** eligibility 返回 `qualified=true`
- **THEN** 客户端 MUST 允许按既有逻辑请求值得留意数据接口并展示结果

#### Scenario: 资格失败 fail-closed

- **WHEN** eligibility HTTP 失败
- **THEN** 客户端 MUST NOT 当作已合格去请求生成接口

#### Scenario: 不强制解释今日排除

- **WHEN** 展示未合格进度文案
- **THEN** 文案 MUST NOT 被规格要求包含「今日不计入」或等价规则说明书表述

## ADDED Requirements

### Requirement: UCG 进度文案同样由客户端拼字段

UCG 入场资格展示（若展示未合格进度）MUST 使用 cash UCG eligibility 的 `effectiveDays` / `requiredDays` / `remainingDays` 由客户端拼接；MUST NOT 依赖服务端 `message` 作为已累计有效日数字的唯一来源；MUST NOT 强制向用户声明「今日不计入」。

#### Scenario: UCG 未合格展示有效日进度

- **WHEN** UCG eligibility 返回 `qualified=false` 且 `effectiveDays=3`、`requiredDays=7`
- **THEN** 客户端进度展示 MUST 能体现已累计有效日为 3（或等价「3/7」「还差 4」等由字段可推导的表述）
