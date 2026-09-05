## ADDED Requirements

### Requirement: 业务服务包 MUST NOT 互相 import 实现

`internal/services` 下各业务域包（至少：cash、voice、device、history、ucg、gatewayapp、simuser、mcpbridge、appstatus）MUST NOT 直接 import 另一业务域包中的实现代码。跨域协作 MUST 通过 `internal/clients/*` 出站调用、消息事件或 `contracts` 接口（实现由本域或 clients 装配）。共享无业务工具 MUST 放在 `internal/platform/**`（或同等 platform 包）。`aimodel` 与 `contracts` MAY 被各业务包 import，视为共享基础设施而非业务线包。

#### Scenario: voice 不再 import services/cash

- **WHEN** 检查 `internal/services/voice` 的 Go import
- **THEN** MUST NOT 出现 `hello/internal/services/cash`

#### Scenario: ucg 不再 import services/device 业务实现

- **WHEN** 检查 `internal/services/ucg` 的 Go import
- **THEN** MUST NOT 出现对 `hello/internal/services/device` 的依赖（设备 HTTP 经 clients；类型/密钥头经 platform 或 clients DTO）

#### Scenario: 允许 platform 与 contracts

- **WHEN** voice 需要解析内部 wxId 头或调用 DeviceAdmin 契约
- **THEN** MAY import `platform` 与 `contracts`，MUST NOT 因此 import device 业务实现包

### Requirement: 伪公共工具 MUST 下沉 platform

下列能力 MUST 从业务域包迁出至 `internal/platform`（或等价 platform 子包），供多进程共用：内部网关密钥头与校验、App/内部注入头常量（如 wxId/deviceNo/clientIP）、从请求头解析 wxId、常量时间比较等。迁出后，原业务包 MUST NOT 再作为他域获取这些工具的唯一来源。

#### Scenario: cash controller 不 import voice 解析 wxId

- **WHEN** cash 控制器需要解析 `X-Internal-Wx-Id`
- **THEN** MUST 使用 platform（或等价）解析函数，MUST NOT import `internal/services/voice`

#### Scenario: history/ucg 不借 device 包做密钥校验

- **WHEN** history 或 ucg 内部接口校验网关内部密钥
- **THEN** MUST 使用 platform 校验，MUST NOT 仅为校验而 import device 业务包

### Requirement: 仓库 MUST 提供跨域 import 门禁

仓库 MUST 提供可运行的检查脚本（路径以实现为准，如 `hack/check-service-import.sh`），用于检测业务服务包之间的禁止 import；检查失败时 MUST 以非零退出码表示（可先 warn 后在同一变更内收紧为 fail）。`AGENTS.md` 与 `openspec/project.md` MUST 记载源码包边界约定，与「跨库须走 HTTP」并列。

#### Scenario: 门禁检出 voice→cash

- **WHEN** 有人重新引入 `services/voice` → `services/cash` 的 import 并运行门禁
- **THEN** 检查 MUST 失败（收紧后）或明确告警（过渡配置下）

#### Scenario: clients 与 platform 不误杀

- **WHEN** `services/voice` 仅 import `clients/cash` 与 `platform`
- **THEN** 门禁 MUST 通过
