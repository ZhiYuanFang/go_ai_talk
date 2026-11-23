# Proposal

## Why

Flutter 客户端将喂养历史实时推送从鉴权 WebSocket 硬切为 SSE，以避免 refresh 失败后的 WS 重连空转。网关现仅提供 `/device/app/ws/history`；必须在 `gateway-app-server` 新增对等 SSE 通道，复用既有 history 推送源（Redis 订阅 → 按 `device_no` 扇出），供兄弟仓 `flutter_ai_talk` 变更 `history-sse-retire-ucg-clinic-ws` 消费。**已上线仍走 history WS 的客户端 MUST 不受本变更影响。**

## What Changes

- 新增 **`GET /device/app/sse/history`**（名称可微调但须与 Flutter design 一致）：`Authorization: Bearer` 校验 JWT；`deviceNo` 与 JWT `device_no` 对齐；响应 `text/event-stream`。
- SSE 下行业务 JSON **与现网 history WS 推送兼容**（`action` create/update/delete + `payload`），便于客户端复用解析；SSE 包装（`event:` 行等）仅作用于 SSE 连接。
- 鉴权失败返回 HTTP 401（或等价），**不得**升级为 WebSocket。
- **MUST 完整保留** `/device/app/ws/history` 现网契约（路径、Bearer 豁免、首帧 auth / `auth_ok`、ping/pong、下行 JSON 形态）；本变更 **MUST NOT** 删除或改写该端点行为。Flutter 硬切稳定后可另变更下线 WS。
- **MUST NOT** 为本变更修改 history-service Redis 发布体或写库语义。
- **非目标**：UCG chat WS、voice WS、clinic WS 的 SSE 化。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `gateway-edge`: 注册历史 SSE；JWT/device_no 校验；扇出与 usage/CORS/豁免策略对齐；明确与 history WS **并存且行为不变**，直至客户端硬切完成后再另开 retire 变更。
- `history-feeding`: 历史实时推送 MUST 也可经 SSE 送达已认证订阅者；事件载荷与 list/WS 字段一致（含 expect* 等既有约定）；MUST NOT 改变落库与 Redis 发布规则。

## Impact

- `gateway-app-server` 路由注册、`gateway_app_history` WS **旁路**新增 SSE handler、Redis 订阅扇出共享（WS 写路径形态不变）。
- 反代/负载均衡需对 **SSE 路径** 禁用缓冲（不得全局误伤其它入口）。
- 联调对象：Flutter `history-sse-retire-ucg-clinic-ws`；回归对象：未升级 App 的 history WS。
