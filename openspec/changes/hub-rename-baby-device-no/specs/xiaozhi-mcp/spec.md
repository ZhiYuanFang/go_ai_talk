# Spec Delta

## ADDED Requirements

### Requirement: 宝宝号改名时 MCP 喂养落点 MUST 跟随迁移

当 device 域执行宝宝号原地改名时，系统 MUST 将 `xiaozhi_mcp_binding` 中 `device_no=oldDeviceNo` 的喂养落点更新为 `newDeviceNo`。改名后新的喂养写入 MUST 落在新号对应宝宝分区。

#### Scenario: 绑定落点随改名更新

- **WHEN** 某音箱绑定的喂养落点为旧宝宝号且该宝宝完成改名
- **THEN** 该绑定行的 `device_no` MUST 为新宝宝号
