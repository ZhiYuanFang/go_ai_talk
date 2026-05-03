## Why

各服务已按「单进程单库」部署，但 `voice-service` 仍可能通过 `device.DeviceAdmin()` 等同进程符号调用，在运行时会错误地使用 **voice 进程配置的 `database.default`** 去执行他域 DAO，造成隐式跨库或连错库。架构上跨服务数据交互必须 **仅经 HTTP（或已批准的 RPC/消息）**，禁止在调用方进程内直接执行他域表的 DAO。

## What Changes

- 为 **voice → device** 的设备域访问（事件字典、动作词典、画像保存、语音链路写 event/action 等）建立 **HTTP 客户端契约与适配层**，与现有 `DeviceProfile` 的 local/remote 模式对齐；**BREAKING**：voice 独立部署且仅连接 voice 库时，不再允许依赖进程内 `DeviceAdmin` 直连。
- 在 `contracts` 中补齐 device 内部 HTTP 路径与 URL 解析；`domain_refs` 或等价入口改为返回 **契约接口**，由配置选择本地直连（仅迁移/开发显式允许）或远程 HTTP。
- 代码评审与 `AGENTS.md` 明确：**禁止**在 `voice` 包 import 或调用会触发他域 `dao.*` 的路径（除本域 `qa`/`suggest`）。
- 文档：运行手册与环境变量（`DEVICE_SERVICE_URL`、`DEVICE_*_MODE` 等）说明单库模式下的必配项。

## Capabilities

### New Capabilities

- `voice-device-domain-http-access`：定义 voice 访问 device 领域（画像、事件、动作、语音所需写路径）时 MUST 经 device-service HTTP；禁止在 voice 进程内执行 `dao.User`/`dao.Event`/`dao.Action`。

### Modified Capabilities

- `service-boundary-no-cross-db`：补充「单服务单库」部署下禁止进程内跨包 DAO 伪装成领域服务」的场景与约束。
- `voice-device-profile-http-contract`：强调画像获取的实现路径 MUST 为 HTTP 适配器（remote 或显式 local 仅用于同库迁移期），与全量 device 域 HTTP 策略一致。

## Impact

- **代码**：`internal/services/voice`（`domain_refs`、`voice_chat_understanding` 等）、`internal/services/device`（必要时暴露/稳定内部 API）、`internal/services/contracts`。
- **配置**：`manifest/config/config.voice-service.yaml`、`voice-deployment` 环境变量；各环境 `DEVICE_SERVICE_URL`。
- **运维**：voice 与 device 网络互通、超时与熔断；回滚时恢复旧适配开关（若保留）。
