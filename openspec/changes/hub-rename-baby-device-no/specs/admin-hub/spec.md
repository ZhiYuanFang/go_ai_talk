# Spec Delta

## ADDED Requirements

### Requirement: Hub 设备管理 MUST 提供宝宝号改名操作

加载设备管理 Hub 页时，系统 MUST 为已注册宝宝提供改名入口：运维可指定或生成新 `deviceNo`，经二次确认后调用 device 管理改名 API。成功后页面 MUST 展示新号；失败 MUST 展示可读错误且 MUST NOT 静默假装成功。Hub MUST 另提供可不依赖列表当前行的自由填写入口，使运维可同时输入旧号与新号以发起或重试改名。

#### Scenario: 运维确认改名

- **WHEN** 运维在 Hub 对某已注册宝宝提交新号并确认
- **THEN** Hub MUST 调用改名 API，成功后 MUST 以新号更新展示

#### Scenario: 取消确认

- **WHEN** 运维在二次确认时取消
- **THEN** Hub MUST NOT 调用改名 API

#### Scenario: 自由填写旧号与新号补迁

- **WHEN** 运维在 Hub 自由填写入口输入非空 `oldDeviceNo` 与 `newDeviceNo` 并确认
- **THEN** Hub MUST 以该双号调用改名 API（MUST NOT 要求旧号仍出现在当前设备列表行上）
