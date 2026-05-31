## ADDED Requirements

### Requirement: 历史画像接口返回昵称
系统 SHALL 扩展历史画像读取接口（`GET /device/history/api/birthday`）返回 `nickname` 字段；该字段 MUST 通过 device 画像契约获取，history-service SHALL NOT 直连查询 device 域库表。

#### Scenario: 已有昵称
- **WHEN** 目标设备存在可用昵称
- **THEN** 响应 SHALL 包含非空 `nickname`

#### Scenario: 无昵称
- **WHEN** 目标设备当前无昵称
- **THEN** 响应 SHALL 返回 `nickname` 为空串，且接口 SHALL 保持成功响应

### Requirement: 历史页面展示昵称
系统 SHALL 在历史记录页面展示 `nickname`，并与既有性别展示共存；接口返回为空时页面 MUST 显示空态文案而非报错。

#### Scenario: 页面加载成功展示昵称
- **WHEN** 页面加载到包含 `nickname` 的画像数据
- **THEN** 页面 SHALL 显示昵称文本并维持原有性别主题逻辑

#### Scenario: 昵称为空时降级展示
- **WHEN** 接口返回 `nickname` 为空
- **THEN** 页面 SHALL 显示“未设置昵称”或等价占位，并 SHALL NOT 阻断其他历史数据渲染
