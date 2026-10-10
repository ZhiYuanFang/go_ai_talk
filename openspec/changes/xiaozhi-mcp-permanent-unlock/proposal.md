## Why

小智 MCP 绑定能力需对 C 端做一次性买断（永久），与现有商业功能（限时、进「开通功能管理」）语义冲突，不能复用 `ActivateFeature` 的天数模型。绑定 App/Hub 写路径目前无开通门禁；需独立权益、独立 SKU 配置与手工授，同时复用既有支付通道仅增加类型分支。

## What Changes

- 新增 **wx 维永久**「小智 MCP 能力」权益（付费或 Hub 手工授）；**仅** `POST` 添加绑定校验已开通，List/改备注/Delete 不校验。
- Flutter 与 Hub（复用 App 登录绑接口）**同一闸门**：未开通不可 Add。
- 支付：复用现有功能订单/支付宝/Apple 回调骨架，在建单与履约处分出 **MCP 类型**，履约写入永久权益，**不**走 `ActivateFeature` 的「天数≥1」路径。
- SKU / 价格 / Apple 商品 ID：落 **专用表**，由 **独立小 Admin**（一页两区：SKU 配置 + 手工授/可选撤销）维护；**不**进入开通功能管理、`feature/catalog`、`feature_def` 瓷砖。
- App：提供开通态与可购 SKU 查询，以及经同一支付通道的建单入口（类型为 MCP）。
- 存量已绑定用户：本变更默认 **不**自动补授；未开通者保留已有绑定行，但无法再 Add（产品可后续补 grandfather，首版不做）。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-permanent-unlock`：wx 永久能力权益、支付类型分支履约、App 开通态/SKU/建单、绑定 Add 门禁。
- `xiaozhi-mcp-unlock-admin`：独立小 Admin（SKU 专用表配置 + 手工授/撤销）、Hub 导航与 Admin API。

### Modified Capabilities

- （基线 `openspec/specs/` 尚无已归档的 `xiaozhi-mcp-binding`；Add 门禁作为新能力要求约束既有 App 添加接口行为，实现时改 device 绑定写路径。）

## Impact

- **cash-service**：专用 SKU/权益表（或等价 schema）、建单/履约分支、Admin API、Apple/支付宝反查需识别 MCP product。
- **device-service**：`AddXiaozhiMcpBinding`（及 App Add 控制器）调用 cash 校验开通；跨域经 `clients/cash`，禁止 device 直查 cash 库。
- **gateway-app**：登记独立 Admin 静态页与导航；新 App/Admin 路由落在既有 cash/device 反代前缀内；新 App 接口 **须向负责人确认是否计入 usage**（未确认前不改 `maintenance_skip`）。
- **Flutter**（兄弟仓）：绑前查开通态、未开通走支付；本仓只提供契约。
- **Hub**：`history.html` 小智绑定区依赖同一 Add 门禁；运维改价/授在独立小页，不进开通功能管理。
