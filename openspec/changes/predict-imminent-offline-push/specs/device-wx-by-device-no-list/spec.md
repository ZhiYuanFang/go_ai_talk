## ADDED Requirements

### Requirement: device MUST 能按 device_no 列出全部绑定 wxId

device-service MUST 提供内部（或服务间）契约：输入非空 `deviceNo`，返回 `wx` 表中 `device_no` 等于该值的全部用户主键 `wxId` 列表（顺序不限）。MUST NOT 仅返回 `LIMIT 1` 单条作为「全部绑定用户」的语义。实现 MUST 对单次返回数量设上限（配置或常量）；超过上限时 MUST 截断并记录告警日志，或提供分页（一期可选截断+告警）。voice 等调用方 MUST 经 `clients/device` 访问，MUST NOT 直查 device 库。

#### Scenario: 同设备多账号全部返回

- **WHEN** 两个不同 `wx.id` 的行均绑定同一 `device_no`
- **THEN** 列表接口 MUST 返回包含这两个 `wxId` 的结果集

#### Scenario: 无绑定时返回空列表

- **WHEN** 没有任何 `wx` 行绑定该 `deviceNo`
- **THEN** 接口 MUST 成功返回空列表且 MUST NOT 报错为系统故障
