# Design

## Context

见 `proposal.md`。现网：`gateway_app_history` 注册 `/device/app/ws/history`（Bearer 豁免），首帧 auth 后按 `device_no` 入推送集合，消费 Redis 历史变更并 **原文** `BroadcastText` 下行。Flutter 变更 `history-sse-retire-ucg-clinic-ws` 将硬切 SSE；未升级客户端仍依赖完整 WS 契约。

## Goals / Non-Goals

**Goals:**

- 新增历史 SSE，鉴权与 device 绑定对齐 HTTP/WS
- 与 WS **共用** Redis 订阅源；SSE 下行业务字段与 WS 兼容
- **Additive-only**：旧客户端 history WS 行为与部署前一致
- 文档化反代对 SSE 路径不缓冲

**Non-Goals:**

- 删除或改写 history WS 协议（另开 retire 变更）
- 改 UCG/voice/clinic WS
- 改 history-service 业务写入或 Redis 发布 JSON

## Decisions

### 1. 路径与帧格式（与 Flutter 锁定）

| 项 | 值 |
|----|-----|
| Path | `GET /device/app/sse/history` |
| Auth | 中间件或 handler 内 Bearer JWT（**非**豁免名单） |
| Query | `deviceNo` |
| Headers out | `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`, 视部署加 `X-Accel-Buffering: no` |
| SSE event | `event: history` |
| data | 业务 JSON 与现网 WS 帧兼容：至少含 `action` + `payload`；若直接转发 Redis 原文（含顶层 `device_no`）亦可，由 SSE writer 决定，**不得**为对齐 SSE 而改 WS `BroadcastText` |
| 心跳 | 每 N 秒 SSE comment (`: ping`) 或 `event: ping`，防中间设备断闲置连接 |

### 2. 实现落点

- 在 `gateway_app_register` 旁注册 SSE handler（与 `gatewayAppHistoryWS` 并列）。
- Redis 订阅回调可分叉：WS 继续 `HistoryHub().BroadcastText(ctx, deviceNo, payload)`；SSE 走独立 writer。
- 可并列维护 SSE 订阅集合；**不要**把 SSE 连接塞进 `HistoryWSConn` 类型强行共用写路径。
- **不要**把 SSE 放进 `gateway_app_auth_exempt`；**不要**从豁免表移除 `/device/app/ws/history`。

### 3. 与 WS 并存（旧客户端兼容硬约束）

本期 MUST：

1. 保留路径 `/device/app/ws/history`。
2. 保留 GET 精确豁免（首帧 auth，非 HTTP Bearer）。
3. 保留首帧 `type=auth` → `auth_ok`、后续 `ping`/`pong` 语义。
4. 保留 Redis → WS 下行：对匹配 `device_no` 的连接 `WriteText` 现网 JSON（含顶层 `device_no`、`action`、`payload`）；本变更 MUST NOT 为 SSE 规范化而改写该字节流。
5. MUST NOT 修改 history-service `publishHistoryChange` 消息体。

Flutter 硬切后观测无 WS 流量，再开 retire-WS 变更。

### 4. 上线顺序

1. 本变更部署 SSE（WS 不动）
2. 回归：未升级 App 仍能经 WS 收 create/update/delete
3. Flutter 联调硬切
4. （可选，另变更）下线 history WS

### 5. 反代与 usage

- 禁缓冲配置 MUST 限定在 SSE 路径（或等价 location），不得全局关闭缓冲。
- SSE 长连接 usage 策略可单独文档化；修改 skip 规则时 MUST NOT 改变现有 WS Upgrade 排除语义。

## Risks / Trade-offs

- **[Risk] 共享扇出重构改坏 WS 下行** → Mitigation：WS 写路径保持 `BroadcastText` 原文；SSE 包装只在 SSE writer。
- **[Risk] 误改 Bearer 豁免表** → Mitigation：任务显式核对 `/device/app/ws/history` 仍在 `gatewayAppAuthExemptExactGET`；SSE 不在表中。
- **[Risk] 反代缓冲导致 SSE 延迟** → Mitigation：路径级禁缓冲头 + 部署检查清单。
- **[Risk] 双集合漏推** → Mitigation：同一 Redis 回调内对 WS 与 SSE 各写一次。
- **[Trade-off] 短期维护两通道** → 换客户端安全硬切，保护未升级用户。

## Open Questions

（无）
