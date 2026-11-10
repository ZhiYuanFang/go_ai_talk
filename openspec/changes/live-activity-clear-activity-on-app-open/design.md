## Context

`live-activity-update-ack-timeout` 已落地：update 注入 ContentState.`ack`，`send_ok` 后 `AfterFunc(LIVE_ACTIVITY_ACK_TIMEOUT_MS)`，超时未回执则删 activity 并用同机 `push_to_start` start；另有 `POST /app/api/push/live-activity/ack`（gateway 免鉴权，对齐 register）。产品现接受划掉到打开 App 的空窗，希望去掉整条 ack 链路，改由 Flutter 登录后检测本机无 LA 时调鉴权接口清 activity token。

约束：跨域经 gateway `/app/api/push/*`；clear **必须鉴权**（与 device register 同族，身份 `X-Internal-Wx-Id`）；禁止新增扫表 ticker；接口版本上 **删除** 未稳定依赖的 ack path、**新增** clear path（ack 为近期能力，无历史前端依赖，按移除处理）；表级 `updated_at` 必须保留。

## Goals / Non-Goals

**Goals:**

- 彻底移除 update ack 注入、超时、补刀、ack HTTP、env、豁免与 skip 条目。
- DROP `last_sent_ack` / `last_acked_ack` / `ack_sent_at`；保留 `updated_at`。
- 新增鉴权 `POST /app/api/push/live-activity/activity/clear`：按 `(wx, deviceKey)` 删 `token_kind=activity`；不删 `push_to_start`；不同拍 start。
- 与 Flutter twin 契约对齐（见 `flutter-twin-proposal.md`）。

**Non-Goals:**

- 不在 clear 时主动 APNs start。
- 不改 invalidToken 同拍 restart。
- 不改 LA 触发矩阵、lead、items≤3、register 免鉴权。
- 不恢复「App 未启动时的送达探测」。
- 不删表级 `updated_at`。

## Decisions

1. **自愈模型：App 打开清 token，而非发后短等 ack**  
   产品确认空窗可接受。划掉后至下次打开 App，服务端可能继续空气 update；打开后 clear → 下次 upsert 走 start。  
   *备选：保留 ack 超时 — 否决（用户明确删除）。*  
   *备选：clear 后立刻 push_to_start start — 否决（用户要求仅影响下次推送）。*

2. **Clear 接口契约**  
   - Path：`POST /app/api/push/live-activity/activity/clear`  
   - Auth：**必须 Bearer**；gateway **不得**豁免；wx 身份来自 `X-Internal-Wx-Id`（对齐 `/app/api/push/register`）。  
   - Body：`deviceKey`（required）。  
   - 语义：`DELETE` 匹配 `(wx_id, device_key, token_kind=activity)`；0 行仍 200；MUST NOT 删 `push_to_start`。  
   *备选：按 wx 清全部设备 activity — 否决（多机误伤）。*

3. **Schema 清理**  
   ensure：已有库幂等 `ALTER DROP` 三列；新库 CREATE 不含三列；实体与 upsert `OnDuplicate` 去掉 ack 字段；activity 重注册不再清 ack 列。  
   **禁止** DROP `updated_at`。

4. **Dispatch 回退**  
   `push_live_activity.go` 去掉 `failoverStaleActivityIfAckTimedOut`、ack 注入、`MarkLiveActivityAckSent`、`scheduleLiveActivityAckTimeout`；update/start/end 载荷仅业务 ContentState。删除整个 `push_live_activity_ack.go`（或清空后移除文件）。

5. **Gateway / usage**  
   - 从 `gatewayAppAuthExemptExactPOST` 移除 `/app/api/push/live-activity/ack`。  
   - 从 `maintenance_skip` 移除 ack 条目。  
   - clear **不计入 usage**：负责人已确认；`maintenance_skip` 增加 `POST /app/api/push/live-activity/activity/clear`（对标 register）。

6. **Env**  
   删除 `LIVE_ACTIVITY_ACK_TIMEOUT_MS`（compose 注入、`.env.example`、`.env.prod` 等注释/赋值）。

7. **与 invalidToken 关系**  
   invalidToken 路径保留；clear 是第三条进门（用户主动），出口同为「无 activity 行 → 下次 start」。

## Risks / Trade-offs

- [划掉后空窗空气 update] → 产品已接受；invalidToken 仍可能提前自愈。  
- [Flutter 未升级仍发 ack] → Go 无路由 → 客户端失败可忽略；ContentState 无 ack 后旧原生回执无意义。  
- [误调 clear 时本机仍有 LA] → Flutter 须先查 ActivityKit；服务端无法校验本机是否仍有活动，接受「客户端契约」。  
- [未问清 usage 就改 skip] → tasks 强制先确认。  
- [误 DROP updated_at] → design/tasks 显式禁止；评审核对 CREATE/DROP 清单。

## Migration Plan

1. 发版 Go：schema DROP 三列 + 下线 ack + 上线 clear；同步去掉 env。  
2. 同发或紧随 Flutter：去 ack、接 clear。  
3. 回滚：恢复 ack 代码成本高；短期可仅保留 clear、不恢复 ack（空窗策略已定）。若需紧急回退 clear，客户端停调即可（服务端 clear 幂等无害）。

## Open Questions

- （无）clear 不计入 usage 已确认。
