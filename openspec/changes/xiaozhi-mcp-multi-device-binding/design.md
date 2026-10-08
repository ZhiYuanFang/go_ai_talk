## Context

现有 `cmd/mcp-service` 作为 MCP Server **出站**拨号 `wss://api.xiaozhi.me/mcp/?token=...`，环境变量固定一对 `(token, deviceNo)`，工具 `baby_feeding_advisor` 经 `/voice/chat/ws` 文模式落喂养对话。C 端需要：用户在 Flutter 为已绑宝宝账号登记多台小智音箱 token，一人多音箱、数据落同一宝宝 `deviceNo`。

约束：服务边界（绑定在 device、mcp 不直连业务库）；gateway-app App 接口与 usage 约定；背景循环须 OpenSpec 批准；进程历史上无生产用户，可直接改名。

## Goals / Non-Goals

**Goals:**

- App 可对小智绑定做增删改查；添加取当前 `wx.device_no`；token 全局唯一；列表脱敏 + alias。
- 单副本 `xiaozhi-mcp-service` 按绑定集维护 N 条 Bridge；写路径内部 HTTP 增量同步；启动全量 + 低频 reconcile。
- 进程/镜像/部署资源改名为 `xiaozhi-mcp-service`；Deployment `replicas: 1`。
- 小智绑定 App CRUD **计入 usage 统计**。

**Non-Goals:**

- 多副本分片/选主/抢锁。
- Redis Pub/Sub 写路径；为绑定引入新 Redis 读缓存。
- Flutter 直连 mcp；`intent-mcp-service`。
- 改造 `wx.device_no` 1:1 模型；多宝宝选设备 UI（首版固定当前绑定宝宝）。

## Decisions

### D1. 绑定权威在 device-service

- **选择**：新表存 `wx_id` / `device_no` / `mcp_token` / `alias` / 状态与时间戳；App CRUD 走 `/device/app/api/...`（Bearer → `X-Internal-Wx-Id`）。
- **理由**：与账号/宝宝身份同域；可镜像 `push_device` 多登记，而非复用 `bindwx` 覆盖语义。
- **备选**：mcp 自建库 → 违反边界且无用户体系。

### D2. 表单字段

- **选择**：客户端提交 `mcpToken` + `alias`；服务端用当前 wx 已绑 `deviceNo`；未绑宝宝拒绝。
- **理由**：业务「账号下挂多音箱 → 同一宝宝」。
- **备选**：显式传 deviceNo → 首版不做。

### D3. Token 全局唯一

- **选择**：添加（及若允许改 token）时对 `mcp_token` 做全局唯一约束（DB unique + 业务校验）；冲突返回明确业务错误。
- **理由**：防两用户抢同一音箱接入点。
- **说明**：去重不解决多副本双拨号；出站连接靠 D4。

### D4. 单副本出站

- **选择**：`replicas: 1`；Manager 对全量 active 绑定 dial。
- **理由**：实现简单；避免同 token 多进程互踢。
- **代价**：进程宕机期间全桥短暂不可用，直到拉起。

### D5. 同步：全量拉取 + 内部 HTTP + 低频 reconcile

```
device 写成功 → HTTP 调 xiaozhi-mcp internal Upsert/Remove
xiaozhi-mcp 启动 → GET device internal list → Reconcile
xiaozhi-mcp ticker → 低频再全量 Reconcile（丢推送兜底）
```

- **选择**：写路径 **内部 HTTP**（单副本打唯一实例）；鉴权复用 `DEVICE_GATEWAY_INTERNAL_SECRET`（或同等内部密钥头）。
- **理由**：单副本下比 Pub/Sub 少 Redis 依赖；实时性好于纯轮询。
- **备选**：Pub/Sub → 多副本 fan-out 更优，本变更不需要。

### D6. 批准的背景循环

| 字段 | 值 |
|------|-----|
| 任务名 | `xiaozhi-mcp-binding-reconcile` |
| 宿主 | `cmd/xiaozhi-mcp-service` |
| 周期 | 可配置，默认 5–15 分钟量级 |
| 开关 | 环境变量（如 `XIAOZHI_MCP_RECONCILE_INTERVAL_MS`；`0` 关闭 reconcile，仍保留启动全量与写路径推送） |
| 失败语义 | 单次失败打日志，不退出进程；下次周期重试 |
| 关闭 | SIGTERM 取消 ctx |
| 为何需要 | 写路径 HTTP 可能丢（mcp 短暂不可达）；不能仅靠请求内同步保证长期一致 |

既有 Bridge 断线重连循环（`mcp-bridge-reconnect` 语义）保留，按每条 Bridge 独立 `Run(ctx)`。

### D7. 进程改名

- **选择**：`mcp-service` → `xiaozhi-mcp-service`（cmd、Dockerfile、compose、kustomize、ACR 映射、config、OTEL_SERVICE_NAME、日志前缀）。
- **包名**：`mcpbridge` 可暂留或随实现改为更明确包名；MCP `serverName` 已为 `xiaozhi-mcp-service`。
- **K8s**：新增 ClusterIP Service 暴露内部 HTTP；liveness 可改为 HTTP 探针（可选，至少暴露端口）。

### D8. Usage 统计

- **负责人确认**：小智绑定列表的**增删改查** App 接口 **计入** usage。
- **实现**：`g.Meta` 完整 path/method/summary；**不得**写入 `usagestats/maintenance_skip.go`。
- **内部** mcp/device internal 路径不计入 App usage。

### D9. 安全与脱敏

- 列表/详情不返回完整 token（mask）；日志沿用 maskToken。
- 库内明文或可逆加密：首版允许受控明文 + 权限与传输保护；若后续加固加密，另案。
- 删除绑定必须停对应 Bridge。

### D10. 迁移期 env fallback（可选）

- 若启动时内部 list 为空且仍配置了 `XIAOZHI_MCP_TOKEN` + `DEVICE_NO`，可注入一条临时 Bridge，便于灰度；稳态以 DB 绑定为准。design 实现时可保留开关，默认以 DB 为准。

## Risks / Trade-offs

- [单副本 SPOF] → 接受；监控进程存活；后续再分片。
- [写路径 HTTP 失败导致绑定已入库但未连] → reconcile 兜底；App 可提示「稍后自动连接」；可选添加接口同步等待 Upsert 结果并表面 status。
- [token 泄露] → HTTPS + Bearer；不回传全文；运维日志脱敏。
- [deviceNo 变更（用户换绑宝宝）] → 首版不自动迁移旧绑定 rows；可在改绑时文档要求用户重绑小智，或后续加同步策略。
- [改名镜像/编排遗漏] → tasks 列全量清单；CI ACR 服务名同步。

## Migration Plan

1. 合并代码：device 表迁移 + App API + mcp 改名与 Manager；部署 `replicas: 1` + Service。
2. 配置 `XIAOZHI_MCP_SERVICE_URL`（device → mcp）与 mcp 侧 `DEVICE_SERVICE_URL`（或既有命名）。
3. 下线对「唯一 env token」的运维依赖；用户经 App 添加绑定。
4. 回滚：回退镜像；绑定表可保留（无害）；旧单 env 进程可临时回切（若保留 Dockerfile 别名则另说——本变更直接改名，回滚用上一镜像 tag）。

## Open Questions

- 内部 HTTP 路径前缀最终命名（建议 `/xiaozhi-mcp/internal/...` 与 `/device/internal/.../xiaozhi-mcp-bindings`）在实现时与现有 internal 风格对齐即可。
- reconcile 默认间隔精确值（建议 10min）实现时写入 env.example。
- 换绑宝宝后是否级联更新绑定 `device_no`：首版 **不自动级联**；若产品要改，另开增量。
