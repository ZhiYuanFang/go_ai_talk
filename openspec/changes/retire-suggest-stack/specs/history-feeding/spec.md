# Spec Delta

## REMOVED Requirements

### Requirement: Voice MUST 在本域处理 suggest 表

**Reason**: 成长建议写入与读删业务已退役；不再要求 voice 维护 suggest 权威读写面。  
**Migration**: 拆除 List/Delete API 与 Hub 列表；历史行 MAY 留库；宝宝号迁移 MUST NOT 再 UPDATE `suggest`。

## ADDED Requirements

### Requirement: suggest 读删契约 MUST 退役

系统 MUST NOT 再提供设备历史、运维 Hub 或 voice 内部的成长建议列表/删除 HTTP 契约。history-service MUST NOT 再委托 voice 查询或删除 `suggest` 行作为正式产品行为。

#### Scenario: 设备历史 suggest 口下线

- **WHEN** 客户端请求原 `/device/history/api/suggest` 或删除等价路径
- **THEN** 系统 MUST NOT 返回正式业务列表契约（路由不存在或不可用）

#### Scenario: Hub 运维 suggest 口下线

- **WHEN** 运维请求原 `/device/admin/api/history/suggest`
- **THEN** 系统 MUST NOT 再提供可列出成长建议行的正式管理接口

### Requirement: 宝宝号迁移 MUST NOT 触碰 suggest

history-service 的宝宝号分区键迁移 MUST 仅更新仍有效的本域业务表（至少含 `history`），MUST NOT 对 `suggest` 表执行 UPDATE/DELETE/INSERT。无 `suggest` 表的环境 MUST 仍能成功完成 history 段迁移。

#### Scenario: 无 suggest 表时迁移仍成功

- **WHEN** 可信调用方请求将旧宝宝号迁移到新宝宝号，且目标库不存在 `suggest` 表
- **THEN** history 行 MUST 迁至新号，且迁移 MUST NOT 因缺失 `suggest` 表而失败

#### Scenario: 有遗留 suggest 行也不跟迁

- **WHEN** voice 库仍存在旧号下的遗留 `suggest` 行且执行宝宝号迁移
- **THEN** 迁移流程 MUST NOT 要求或保证这些行改挂到新号
