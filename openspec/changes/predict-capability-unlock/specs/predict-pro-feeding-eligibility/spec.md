## ADDED Requirements

### Requirement: 系统 MUST 提供预测 Pro 独立喂养资格场景

cash-service EnsureSchema MUST 幂等种子 `feeding_eligibility_scene.scene_key=predict_pro_entry`，默认 `requiredDays=7`、`minRecordsPerDay=10`，MUST NOT 在重复启动时覆盖运维已改值（INSERT IGNORE 或等价）。该场景 MUST 独立于 `ucg_entry` 与 `care_alert_entry`。资格合成算法 MUST 与现有喂养资格引擎同构（自昨天起向前连续有效日，今日不计入窗口）。

#### Scenario: 首次部署含 predict_pro_entry

- **WHEN** 空库执行 EnsureSchema
- **THEN** 配置表 MUST 含 `predict_pro_entry` 默认 7/10 行

#### Scenario: 改 Pro 场景不影响 UCG

- **WHEN** 运维将 `predict_pro_entry.requiredDays` 改为 5
- **THEN** `ucg_entry.requiredDays` MUST 保持不变

### Requirement: App MUST 可查询预测 Pro 喂养资格且计入 usage

cash-service MUST 提供已登录 App API `GET /cash/app/api/predict-pro/eligibility`（路径精确以实现登记为准）。`device_no` MUST 取自网关注入内部头。响应 MUST 同构 UCG/值得留意资格字段：`qualified`、`requiredDays`、`effectiveDays`、`remainingDays`、可选 `message`。该接口 **MUST 计入** App API 使用统计：MUST NOT 写入 `usagestats/maintenance_skip.go`。MUST 具备 `g.Meta` summary 供 apiregistry。MUST NOT 加入 Bearer 匿名白名单。

#### Scenario: 未满 7 日返回进度

- **WHEN** 绑机用户请求 Pro eligibility 且连续有效日不足配置值
- **THEN** `qualified` MUST 为 false，且 `effectiveDays` / `remainingDays` / `requiredDays` MUST 反映进度

#### Scenario: 请求计入 usage

- **WHEN** 客户端调用 Pro eligibility
- **THEN** gateway-app usage 统计 MUST 可计到该 METHOD+path（维护排除列表 MUST NOT 含该项）

### Requirement: Pro 试用与付费与邀请开通 MUST 校验喂养资格

对 `predict_capability_pro` 的点试用、功能建单（创建订单）、邀请码兑换成功履约前，系统 MUST 确认当前请求设备在 `predict_pro_entry` 下 `qualified=true`；否则 MUST 拒绝且 MUST NOT 授予权益。基础版开通路径 MUST NOT 要求该资格。Pro 开通 MUST NOT 要求基础版已开通。

#### Scenario: 未合格拒绝 Pro 试用

- **WHEN** 用户点 Pro 试用但喂养资格未合格
- **THEN** 系统 MUST 拒绝，MUST NOT 标记试用 used，MUST NOT 写 Pro 权益

#### Scenario: 合格后可付费开通 Pro 且无需基础版

- **WHEN** 设备喂养合格且用户无基础版权益，支付 Pro SKU 成功
- **THEN** 系统 MUST 授予 Pro 人权益

#### Scenario: 基础版开通不查 Pro scene

- **WHEN** 用户开通基础版（支付/邀请/试用）
- **THEN** 系统 MUST NOT 因 `predict_pro_entry` 未合格而拒绝
