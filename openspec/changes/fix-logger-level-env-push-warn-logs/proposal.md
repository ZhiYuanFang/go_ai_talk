## Why

生产 `.env` 已配置 `GF_LOGGER_LEVEL=prod`，compose 也注入该变量，但 GoFrame 的 `gins.Log()` 只从 yaml 读取 `logger.level`（各服务均为 `all`），**不会**被 `GF_LOGGER_LEVEL` 覆盖。运维误以为 prod 已抑制 Info/Debug，实际全量输出；同时 push-service 内部发送路径在 prod 意图下几乎不可观测（入口无日志、凭证未配仅 Debug、成功无日志），导致 voice `push_attempt` 成功后无法在 push 侧对齐排查。

## What Changes

- 新增 `internal/platform/loggercfg`：进程启动在 yaml logger 初始化之后，若 `GF_LOGGER_LEVEL` 非空则 `SetLevelStr` 覆盖默认 logger（与 compose/`x-logger-env` 注释语义对齐）。
- 各长期运行服务 `main`/`prepare*Runtime` 末尾调用该 Apply（voice/push/device/history/ucg/cash/gateway-app/notify/sim-user/mcp 等）。
- push-service：internal `by-biz-type` 受理、异步 dispatch、凭证未配置、发送成功等关键路径补 **Warning** 级可观测日志（prod 可见；不打印 token/alert 全文）。
- 更新 `manifest/docker/.env.example` 与 `manifest/docker/env/.env.prod`（及同结构的 test/release/local 若已有该键）注释，说明 `GF_LOGGER_LEVEL` 取值效果与「留空=跟 yaml」语义，便于随时改。

## Capabilities

### New Capabilities

- `logger-level-env`：环境变量 `GF_LOGGER_LEVEL` 在进程启动后覆盖默认 logger 级别的契约与生效时机。
- `push-dispatch-observability`：push-service internal 发送与厂商 dispatch 的 Warning 级排障日志契约。

### Modified Capabilities

- （无）既有 `openspec/specs/` 为版本快照，本变更以 change 内增量能力规格为准。

## Impact

- 代码：`internal/platform/loggercfg`（新）、各 `cmd/*/main.go` 启动路径、`internal/controller/push`、`internal/services/push`（dispatcher/async/厂商 sender）。
- 配置文档：`manifest/docker/.env.example`、`manifest/docker/env/.env.*` 注释；compose `x-logger-env` 注释可微调为「经 loggercfg 生效」。
- 行为：当生产设 `GF_LOGGER_LEVEL=prod` 后，现有 `Infof`（含 voice `push_attempt`）将不再输出；排障依赖本变更的 Warning 点或临时改 env 为 `all`/`info`。
- 无 API/库表 BREAKING；无新背景 ticker。
