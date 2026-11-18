## ADDED Requirements

### Requirement: wx 表 SHALL 持久化可空设备型号字段

The system MUST persist an optional device model display string on the `wx` row as column `device_model` (JSON `deviceModel`). The value MUST be client-supplied free text; the server MUST NOT maintain an enum/allowlist of brands or models. Missing or empty values MUST be allowed for legacy clients and non-App platforms.

系统 MUST 在 `wx` 表以可空列 `device_model` 持久化客户端上报的设备型号展示文案；MUST NOT 做品牌/机型枚举校验；旧客户端与 Web 可保持为空。

#### Scenario: Ensure 幂等加列

- **WHEN** `device-service` 启动且 `wx.device_model` 尚不存在
- **THEN** 系统 MUST 幂等添加该列且不影响既有行（既有行型号为空）

#### Scenario: 列已存在

- **WHEN** `device-service` 启动且列已存在
- **THEN** Ensure MUST 成功跳过且 MUST NOT 报错中断启动

### Requirement: App API SHALL accept logged-in device model updates

`device-service`（经 gateway-app 反代）MUST 提供 `PUT /device/app/api/user/device-model`，须有效 Bearer；从网关注入的内部 wxId 定位行。请求体 MUST 含字符串字段 `deviceModel`（可为空串表示清空）。trim 后长度 MUST ≤ 128；超长或含非法控制字符时 MUST 拒绝且 MUST NOT 写入。值与库中当前值相同时 MUST 跳过 UPDATE。服务端 MUST NOT 将客户端机型映射为枚举码。

#### Scenario: 登录用户写入型号

- **WHEN** 已登录用户提交 `deviceModel` 为 `huawei`（或任意合法非空文案）且与库中不同
- **THEN** 系统 MUST 将该 wx 行的 `device_model` 更新为提交文案并返回成功

#### Scenario: 空串清空

- **WHEN** 已登录用户提交 `deviceModel` 为空串（或仅空白）
- **THEN** 系统 MUST 将该行型号置空（空串/NULL 实现约定一致）并返回成功

#### Scenario: 未登录拒绝

- **WHEN** 请求无法解析出正整数 wxId
- **THEN** 系统 MUST 拒绝写入

### Requirement: Device-model App API usage-stats policy MUST be owner-confirmed

Whether `PUT /device/app/api/user/device-model` counts toward App API usage stats MUST be confirmed by the product owner before changing `usagestats/maintenance_skip`. Until confirmed, implementers MUST NOT assume denylist membership. If owner confirms「不统计」, the exact METHOD+path MUST be added to `maintenance_skip`; if「统计」, it MUST NOT be listed there and MUST appear in apiregistry summaries normally.

型号上报接口是否计入 App API usage MUST 经负责人确认；未确认前 MUST NOT 擅自改 `maintenance_skip`。

#### Scenario: 确认不统计后登记

- **WHEN** 负责人确认该接口不计入 usage
- **THEN** `maintenance_skip` MUST 含精确项 `PUT /device/app/api/user/device-model`（或最终锁定路径）
