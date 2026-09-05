## Why

运行时已拆成多进程与一服务一库，但源码仍按「共享域包」互相 `import`：出站 HTTP 客户端放在被调方包内（如 `cash.Remote*`）、`controller` 平铺且文件名与宿主进程错位、跨域密钥/请求头工具散落在 device/voice/gatewayapp。结果是 IDE 依赖图与部署边界不一致，后续加功能易继续写穿。需要一次按进程对齐包布局与 import 禁令，且 **不得改动 App 可见 URL/契约**。

## What Changes

- **B（clients）**：新增中立 `internal/clients/{target}/`，迁入所有域间出站 HTTP/WS 客户端（含 `cash.RemoteIsVip*` / `RemoteCareAlertAccess`、各 `*_client.go` 等）；被调方 `internal/services/{domain}` **仅保留本进程业务实现**，不再承载「给别人调自己」的 Remote SDK。
- **C（controller）**：`internal/controller/` 按进程分子包（cash/voice/device/history/ucg/gatewayapp/simuser/notify）；各 `register_*_service` 只 Bind 本子包；允许重命名误导文件（如 voice 宿主的 `device_care_alert_*`），**禁止**修改 `g.Meta` / 路由中的对外 path 字符串。
- **D（包隔离）**：业务服务包不得 import 他域业务实现包；跨域只许经 `clients/*` HTTP（或既有事件）与 `contracts` / `platform`；将请求头常量、内部密钥校验、`ParseHeaderWxID`、`ConstantTimeEqual` 等伪公共沉到 `platform`（或等价共享无业务包）。
- 增加自动化门禁脚本（如 `hack/check-service-import.sh`）与 `AGENTS.md` / `openspec/project.md` 约定更新。
- **非 BREAKING（对 App）**：Flutter/网关对外 path、method、请求响应字段名保持不变；本变更纯服务端源码结构。

## Capabilities

### New Capabilities

- `service-outbound-clients`：中立 `internal/clients/{target}` 出站客户端归属与迁入规则；业务包禁止再导出 Remote* 供他进程 import。
- `controller-process-layout`：controller 按进程分子包与 register 绑定规则；对外路由字符串冻结。
- `service-import-isolation`：跨服务业务 import 禁令、platform 共享工具下沉、门禁脚本与文档约定。

### Modified Capabilities

- （无）不改归档基线中面向 App 的业务 Requirement；本变更不新增/修改对外 HTTP 行为规格。

## Impact

- **进程**：全部 `cmd/*-service`（cash/voice/device/history/ucg/gateway-app/sim-user/notify/mcp 等）的 import 与 controller 注册路径；mcp 无 controller 则仅 clients/文档侧对齐。
- **包路径**：大量 Go import 改写；二进制行为与对外 URL 应不变。
- **文档**：`AGENTS.md`、`openspec/project.md`；可选 runbook 一句说明包布局。
- **非目标**：多 Go module / 多仓拆分；改 App URL；改业务语义；新建 `*_test.go`；与 `feature-activation-care-alert` 功能逻辑纠缠（宜先独立合并/归档功能变更）。
