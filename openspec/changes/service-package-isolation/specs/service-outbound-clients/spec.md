## ADDED Requirements

### Requirement: 域间出站客户端 MUST 位于中立 clients 包

系统 MUST 将调用他域服务的 HTTP/WS 出站客户端放置于 `internal/clients/{target}/`（`target` 为被调服务域名，如 `cash`、`device`、`history`、`ucg`、`voice`）。被调方业务包 `internal/services/{target}` MUST NOT 再向其他服务进程导出 `Remote*` 或等价「调本域 internal API」的客户端符号供其 `import`。调用方业务包 MUST 通过 `clients/{target}` 发起跨进程调用，MUST NOT 为调用他域而 `import` 他域 `internal/services/{other}` 业务实现包。

#### Scenario: voice 调 cash VIP/access

- **WHEN** voice-service 需要查询 VIP 或值得留意 access
- **THEN** 源码 MUST `import` `internal/clients/cash`（或等价 clients 路径），MUST NOT `import hello/internal/services/cash`

#### Scenario: cash 调 device/history

- **WHEN** cash-service 需要设备号或喂养日统计
- **THEN** 出站客户端 MUST 位于 `internal/clients/device` 与 `internal/clients/history`（或从原 `cash/*_client.go` 迁入后由此引用），MUST NOT 要求 device/history 业务包反向依赖 cash

#### Scenario: 非本仓域依赖可例外

- **WHEN** 出站目标为 Python AI、阿里云内容安全或外部 LLM
- **THEN** 客户端 MAY 保留在原业务包或非域名 clients 子目录，MUST NOT 因此允许域业务包互相 import

### Requirement: 出站客户端迁入 MUST NOT 改变内部 API 契约语义

迁入 clients 时，请求的 path、方法、内部密钥头与响应解析语义 MUST 与迁入前一致（允许包路径与符号名变更）。MUST NOT 借机修改 App 对外路径。

#### Scenario: CareAlert access 路径不变

- **WHEN** voice 经 clients/cash 调用值得留意 access
- **THEN** 请求 URL path MUST 仍为既有 `/cash/internal/api/care-alert/access`（或迁入前实际使用的同一 path）
