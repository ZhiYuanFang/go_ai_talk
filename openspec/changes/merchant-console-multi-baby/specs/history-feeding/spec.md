# Spec Delta

## ADDED Requirements

### Requirement: 商户欠费时 SHALL 拒绝新增 history 并放行 update 与 end
当 history 写入目标 `device_no` 归属某商户且该商户权益无效（`expire_at` 已过或无有效权益）时：`AddDeviceHistory` / batch `op=create`（及等价新增行路径）MUST 拒绝并返回可读业务错误；`UpdateDeviceHistory`、闭合进行中事件的 end/update、以及只读 list/latest MUST 仍允许。闸门 MUST 落在 history 权威写入口，使 App、voice/Python batch 与商户台同源生效。无商户归属的普通 C 端宝宝 MUST NOT 受本闸门影响。

#### Scenario: 欠费禁止 create
- **WHEN** 商户已欠费且其子账户对名下宝宝请求新增一条 history
- **THEN** history-service MUST 拒绝插入新行

#### Scenario: 欠费允许 end
- **WHEN** 商户已欠费且存在进行中 history，客户端请求 end/update 闭合
- **THEN** 系统 MUST 允许更新且 MUST NOT 因欠费拒绝

#### Scenario: 欠费允许只读
- **WHEN** 商户已欠费且客户端请求该宝宝 history 列表
- **THEN** 系统 MUST 返回数据

#### Scenario: 非商户宝宝不受影响
- **WHEN** 无商户归属的 parent 宝宝在任意权益状态下新增 history
- **THEN** MUST NOT 仅因本商户闸门拒绝
