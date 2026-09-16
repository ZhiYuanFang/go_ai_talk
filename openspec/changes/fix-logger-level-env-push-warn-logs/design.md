## Context

Compose 已通过 `x-logger-env` 向各微服务注入 `GF_LOGGER_LEVEL`，`.env.prod` 设为 `prod`，意图抑制 Info/Debug。但 GoFrame v2.6.3 的 `gins.Log()` 仅用 `Config().Get("logger")` 读 yaml（`GetWithEnv` 也不会覆盖已存在的 yaml 键），各服务 `manifest/config/config.*.yaml` 均写死 `logger.level: "all"`。结果是 env 开关无效，运维看到 Info（如 voice `[predict-imminent] push_attempt`）会误判「prod 未注入」。

同期：voice 经 `clients/push` 调用 `POST /push/internal/api/by-biz-type` 成功（无 `push_failed`）后，push-service 入口零日志；凭证未配走 Debug；发送成功无 Info。在真实 prod 级别下几乎无法对齐排查。

约束：跨服务共享逻辑放 `internal/platform/**`；不新增测试文件；不引入背景 ticker；注释中文。

## Goals / Non-Goals

**Goals:**

- `GF_LOGGER_LEVEL` 非空时，在 yaml logger 初始化之后覆盖默认 logger 级别，使 compose/.env 语义成真。
- push internal 发送链路在 **Warning** 级别可观测（与 prod 开关兼容）。
- `.env.example` 与 `env/.env.prod`（及同类文件）注释清楚各取值效果，便于运维改。

**Non-Goals:**

- 不改各 yaml 默认 `level: all`（本地/留空 env 仍全量）。
- 不改推送业务语义、不重试 MQ、不改 API 契约字段。
- 不为一次性 backfill 工具强制接 loggercfg（可选）。
- 不引入按服务独立的 logger level 变量（仍统一 `GF_LOGGER_LEVEL`）。

## Decisions

### D1. `internal/platform/loggercfg.ApplyFromEnv(service string)`

- 读 `os.Getenv("GF_LOGGER_LEVEL")`，trim 后空则 no-op。
- 先触达 `g.Log()`（或等价）确保 yaml `SetConfigWithMap` 已执行，再 `SetLevelStr`；失败打 Error（无效字符串）。
- 生效确认用 **Warning**：`[loggercfg] service=... GF_LOGGER_LEVEL=... applied`（prod 下仍可见）。
- 备选（否）：删 yaml `level` 指望 `GetWithEnv`——`gins.Log` 不用 `GetWithEnv`，无效。
- 备选（否）：仅改 yaml 为 prod——本地与单一镜像配置冲突。

### D2. 调用时机：各服务 `prepare*Runtime` / `main` 末尾、任意业务逻辑前

- MUST 在可能触发 `gins.Log()` 初始化的代码之后调用（如 `dbcfg.ApplyGroupFromEnv` 内已有 `glog.Printf`）。
- 覆盖进程：`push-service`、`voice-service`、`device-service`、`history-service`、`ucg-service`、`cash-service`、`gateway-app-server`、`notify-service`、`sim-user-service`、`mcp-service`；gateway 若独立 main 同样接入。
- `ucg-*-backfill`：建议接入同一 Apply（启动短、成本低），非硬性。

### D3. push 可观测日志一律 Warning，前缀 `[push]`

| 点 | 内容（脱敏） |
|----|----------------|
| Internal 鉴权失败 | `reason=auth_denied` |
| 参数校验失败 | `reason=bad_request` + 简短原因 |
| 受理成功 | `accepted wxId bizType badge silent alertLen` |
| async recover | `reason=async_panic` + panic 值 |
| dispatch 开始 | `dispatch wxId bizType deviceCount`（沿用/加强现有 skip Warning） |
| 凭证未配置 | 原 Debug → Warning：`channel skipped: credentials not configured` |
| 发送成功 | `send_ok channel wxId bizType` |

禁止日志完整 token、完整 alert 正文。

### D4. env 注释（运维面）

在 `manifest/docker/.env.example` 与 `manifest/docker/env/.env.prod`（及 `.env.test` / `.env.release` / `.env.local` 若含该键）于 `GF_LOGGER_LEVEL` 处写明：

| 取值 | 效果（GoFrame） |
|------|-----------------|
| 留空 | 不覆盖，沿用 yaml（当前多为 `all`） |
| `prod` / `product` | 仅 WARN / ERRO / CRIT |
| `info` | INFO 及以上（无 DEBU） |
| `all` / `dev` / `debug` | 含 DEBU |
| 其它非法值 | Apply 失败并打 Error，级别保持 yaml |

注明：须 **重新创建/重启容器** 后生效；改文件不自动热更新进程内 logger。

### D5. compose 注释修正

`docker-compose.microservices.yml` 的 `x-logger-env` 注释改为：非空时由各服务 `loggercfg.ApplyFromEnv` 覆盖默认 logger（非 GoFrame 自动覆盖 yaml）。

## Risks / Trade-offs

- [生产突然变成真正 prod] → voice 等现有 Infof 消失；依赖本变更 Warning 点 + 临时改 `GF_LOGGER_LEVEL=all` 排障。发版说明写清。
- [Apply 早于 logger 初始化被 yaml 再次覆盖] → MUST 放在 prepare 末尾且先触达 `g.Log()`；tasks 验收「启动日志含 loggercfg applied」。
- [Warning 量：每推一条 accepted + send_ok] → 预测临近频率低可接受；若后续吵可再降为可配置，本变更不引入新开关。
- [.env.prod 含密钥且可能不进 git] → 仍改本机/服务器文件注释；仓库以 `.env.example` 为权威说明副本。

## Migration Plan

1. 合并代码并发布含 `loggercfg` 的镜像。
2. 确认 `.env.prod` 仍为 `GF_LOGGER_LEVEL=prod`（或按需临时 `all`）。
3. `compose up` 强制 recreate 相关服务；查启动 `[loggercfg] ... applied`。
4. 复现 predict-imminent：voice 无 Info 属预期；push 应见 `[push] accepted` / `dispatch` / `no_device` 或 `send_ok` / credentials Warning。
5. 回滚：回退镜像，或临时设 `GF_LOGGER_LEVEL=` / `all` 并 recreate（旧镜像无 Apply 时 env 本就无效）。

## Open Questions

- （无阻塞）gateway 二进制若与 gateway-app 分离，实现时按实际 `cmd` 列表勾选即可。
