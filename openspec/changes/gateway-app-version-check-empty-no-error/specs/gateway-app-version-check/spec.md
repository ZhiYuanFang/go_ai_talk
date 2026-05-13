## ADDED Requirements

### Requirement: 版本表无数据时版本检查须成功且无需更新

gateway-app-server 对 **`GET /device/app/api/version/check`** SHALL 在版本配置表（如 `app_version`）中**不存在任何可用版本行**时，仍返回 **`code=0`**。响应 **`data.needUpdate`** SHALL 为 **`false`**。响应 SHALL NOT 因「结果集无行」或等价空表语义返回非 0 业务码。

#### Scenario: 表无任何记录

- **WHEN** 版本表为空或查询不到任何版本行
- **THEN** HTTP 业务包装 SHALL 为成功（`code=0`）且 **`needUpdate` 为 false**，且 SHALL NOT 将空结果集作为错误返回给客户端

#### Scenario: 有版本记录时行为不变

- **WHEN** 存在至少一条版本记录且 `latestVersion` 可解析
- **THEN** 系统 SHALL 继续按现有规则比较 `currentVersion` 与 `latestVersion` 并设置 **`needUpdate`**

### Requirement: 区分空表与数据库基础设施故障

当版本表查询因**连接、权限、语法等**失败时，系统 MAY 返回非 0 业务码或错误信息以便运维定位。系统 SHALL NOT 将**仅无匹配行**与上述基础设施错误等同为「统一失败」而掩盖空表成功语义。

#### Scenario: 真实读库错误

- **WHEN** 数据库返回非「无行」类错误（如连接失败）
- **THEN** 系统 MAY 返回错误响应且 SHOULD NOT 冒充「无需更新」的成功语义
