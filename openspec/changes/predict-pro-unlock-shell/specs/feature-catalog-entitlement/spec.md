## ADDED Requirements

### Requirement: 本变更 MUST NOT 回退 catalog 下架 status 语义

`GET /cash/app/api/feature/catalog` MUST 继续返回含 `status` 的项（含下架 `status=0`），由前端决定展示；本变更 MUST NOT 恢复「停用功能从目录消失且无 status 字段」的旧行为。

#### Scenario: 下架项仍带 status=0

- **WHEN** 某功能 status=0 且客户端请求 catalog
- **THEN** 响应 MUST 仍含该项且 `status` 为 0
