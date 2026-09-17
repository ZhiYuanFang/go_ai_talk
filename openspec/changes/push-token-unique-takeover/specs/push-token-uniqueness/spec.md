## ADDED Requirements

### Requirement: 推送 token MUST 全局唯一且注册后来顶上

系统在持久化推送设备（`push_device`）时，厂商 `token` 字符串 MUST 在全表唯一（MUST NOT 仅按 `wx_id`+`device_key`+`channel` 唯一而允许多行同 token）。当已登录用户调用注册接口提交某 `token`，若库中已存在相同 `token` 的任意行（无论原 `wx_id`、`channel`、`device_key` 是否相同），系统 MUST 删除这些既有行（或等价保证最终仅保留当前注册结果），再为当前 `wxId` 写入/更新对应 `(device_key, channel)` 记录。同一用户使用不同手机产生不同 `token` 时，系统 MUST 允许并存多行。系统 MUST 在库层提供 `token` 唯一约束；启动迁移 MUST 清理存量重复 token（保留最近更新的一行）后再添加约束。

#### Scenario: 同机换号接管 token

- **WHEN** 账号 A 已注册 token `T`，随后账号 B 在同一设备以相同 token `T` 调用注册成功
- **THEN** 库中 MUST 不再存在账号 A 对该 `T` 的行，且 MUST 存在账号 B 持有 `T` 的行

#### Scenario: 同用户两部手机

- **WHEN** 同一 `wxId` 先后注册两个不同的 token `T1`、`T2`（不同设备）
- **THEN** 库中 MUST 同时保留 `T1` 与 `T2` 两行（在各自 deviceKey/channel 语义下）

#### Scenario: 同账号刷新注册

- **WHEN** 同一 `wxId`、同一 `deviceKey`、同一 `channel` 再次注册（token 可变或不变）
- **THEN** 系统 MUST 成功更新该键上的记录，且 MUST NOT 因本需求导致该用户失去该设备推送能力
