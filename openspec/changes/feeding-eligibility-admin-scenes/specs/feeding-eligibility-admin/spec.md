## ADDED Requirements

### Requirement: Admin MUST 可配置喂养资格场景阈值

cash-service MUST 提供管理端 API（及 Hub 静态页入口）读取与更新已登记场景的 `requiredDays`、`minRecordsPerDay`。场景键 `ucg_entry`、`care_alert_entry` MUST 由种子/发版约定，Admin MUST NOT 创建任意新 `scene_key`。写操作 MUST 校验 Admin 鉴权；成功后 MUST 使相关资格缓存失效（或等价使旧缓存不可命中）。

#### Scenario: 更新 UCG 连续天数

- **WHEN** 运维将 `ucg_entry.requiredDays` 从 7 改为 5 并保存
- **THEN** 后续 UCG eligibility 的 `requiredDays` MUST 为 5，且缓存 MUST NOT 长期返回改前结果

#### Scenario: 拒绝未知场景新建

- **WHEN** 运维尝试写入未登记的 `scene_key`
- **THEN** 系统 MUST 拒绝该写入

### Requirement: 场景配置 MUST 有安全默认种子

EnsureSchema（或等价迁移）MUST 种子 `ucg_entry`（默认 requiredDays=7、minRecordsPerDay=10）与 `care_alert_entry`（默认 requiredDays=2、minRecordsPerDay=10），幂等且 MUST NOT 在重复启动时用种子覆盖运维已改值（或仅 INSERT 忽略已存在）。

#### Scenario: 首次部署有两行默认

- **WHEN** 空库执行 EnsureSchema
- **THEN** 配置表 MUST 含上述两场景默认行
