## ADDED Requirements

### Requirement: wx 表承载账号 VIP 标志

device-service 所属库 `ai_voice_device` 的 `wx` 表 MUST 提供账号 VIP 列 `is_vip`（`TINYINT NOT NULL DEFAULT 0`：`0` 表示非 VIP，`1` 表示 VIP）。该列以 `wx.id` 为账号主键语义，MUST NOT 以 `device_no` 作为 VIP 归属键。

#### Scenario: 默认非 VIP

- **WHEN** 新建或历史迁移后的 `wx` 行未显式设置 VIP
- **THEN** `is_vip` MUST 为 `0`（非 VIP）

#### Scenario: VIP 归属账号而非设备

- **WHEN** 同一 `device_no` 关联的账号上下文需要判断 VIP
- **THEN** 系统 MUST 按该请求对应的 `wx.id` 行上的 `is_vip` 判定，MUST NOT 用设备号反推「设备级 VIP」

### Requirement: device 内部按 wxId 查询 VIP

device-service MUST 提供经内部密钥鉴权的 HTTP 接口，按正整数 `wxId` 返回该账号是否 VIP。voice-service 及其他跨域调用方 MUST 经此契约读取，MUST NOT 在非 device 进程内直查 `wx` 表。

#### Scenario: 存在账号且为 VIP

- **WHEN** 内部调用方以合法密钥请求 `wxId` 对应行且 `is_vip=1`
- **THEN** 响应 MUST 表示 `isVip=true`，并包含该 `wxId`

#### Scenario: 存在账号且非 VIP 或行不存在

- **WHEN** 内部调用方请求的 `wxId` 对应行 `is_vip=0`，或该主键不存在
- **THEN** 响应 MUST 表示 `isVip=false`（或等价：调用方视为非 VIP），MUST NOT 要求调用方因此中断无关主业务（由调用方定义降级）

#### Scenario: 缺少内部密钥

- **WHEN** 请求未携带正确的 device 内部密钥
- **THEN** device-service MUST 拒绝该请求
