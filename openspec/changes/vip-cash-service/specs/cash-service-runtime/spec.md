## ADDED Requirements

### Requirement: cash-service 独立进程与配置

系统 MUST 提供独立进程 `cash-service`，默认配置文件为 `manifest/config/config.cash-service.yaml`（或等价），MUST NOT 将现金域专属配置回流到主配置 `manifest/config/config.yaml`。进程启动 MUST 能通过环境变量 **`CASH_DB_LINK`** 覆盖其 GoFrame `database.default` 连接。

#### Scenario: 默认配置指向本域

- **WHEN** 未设置 `GF_GCFG_FILE` 启动 `cmd/cash-service`
- **THEN** 运行时 MUST 使用 cash-service 专用配置文件，且 MUST NOT 依赖共享主配置承载支付/VIP 业务项

#### Scenario: 数据库连接可被 CASH_DB_LINK 覆盖

- **WHEN** 部署环境设置 `CASH_DB_LINK` 为 `ai_voice_cash` 的 DSN
- **THEN** cash-service MUST 使用该连接访问本域库，MUST NOT 直连 `ai_voice_device` / `ai_voice_voice` 等他域库承载 VIP 表

### Requirement: 部署清单包含 cash-service

微服务 Compose、`.env.example` 与发布 runbook MUST 登记 `cash-service`（镜像/服务名、`CASH_DB_LINK`、`CASH_SERVICE_ADDR`、他服所需的 `CASH_SERVICE_URL`）。

#### Scenario: Compose 可拉起 cash-service

- **WHEN** 按微服务 compose 启动含 cash-service 的栈且已配置 `CASH_DB_LINK`
- **THEN** 容器 MUST 监听配置地址（建议 `:9806`）并提供可探测的 HTTP 健康/API 入口（与现网其它 `*-service` 惯例一致）
