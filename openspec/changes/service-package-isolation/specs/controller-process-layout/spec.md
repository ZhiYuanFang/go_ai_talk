## ADDED Requirements

### Requirement: HTTP controller MUST 按宿主进程分子包

`internal/controller` 下 MUST 按进程划分目录（至少包含：cash、voice、device、history、ucg、gatewayapp、simuser、notify）。各进程的 `Register*ServiceHTTP`（或等价注册入口）MUST 仅绑定本进程子包中的控制器类型。误导性文件名（例如实际由 voice 宿主注册却以 `device_` 为前缀的 care-alert/tip/clinic 控制器）MUST 迁入正确进程子包，并可重命名类型；gateway 反代与跨切面 MUST 归入 gatewayapp（或明确的 proxy 子包并由 gateway-app 注册）。

#### Scenario: voice 注册 care-alert

- **WHEN** 构建并注册 voice-service HTTP
- **THEN** care-alert 控制器 MUST 位于 voice 子包且由 voice 注册入口 Bind，MUST NOT 再与 device-service 注册入口混绑

#### Scenario: cash 仅绑 cash 控制器

- **WHEN** 注册 cash-service
- **THEN** Bind 集合 MUST 仅含 cash 子包控制器（VIP/功能等）

### Requirement: 对外 App 路由字符串 MUST 冻结

本变更中，面向 App/网关已暴露的路由 `path` 字符串（含 `g.Meta` 中的 path 与既有反代匹配模式所依赖的对外前缀）MUST NOT 被修改。允许调整 Go 包路径、类型名与文件名。internal 服务间 path 宜保持不变；若必须调整 MUST 同步所有调用方 clients，且仍 MUST NOT 影响 App 路径。

#### Scenario: care-alert App path 不变

- **WHEN** 完成本变更后的代码审查或自动化检查
- **THEN** App 使用的 `/device/api/care-alert/*`（及同类既有对外 path）MUST 与变更前字符串一致

#### Scenario: 重命名控制器不影响 path

- **WHEN** 将 `DeviceCareAlertController` 重命名并移入 `controller/voice`
- **THEN** 其 `g.Meta` path 字段 MUST 仍为原对外路径
