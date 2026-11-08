## Context

统一 LA 后，`dispatchLiveActivity` 对 `upsert`：有 `activity` token 则 APNs `update`，否则 `push_to_start`→`start`。发送层已对 `isApnsInvalidToken` 删对应 token 行，但 **不会**在同一次 dispatch 内改走 start。用户划掉锁屏后 activity token 失效，脏行导致后续 Refresh 反复 update 失败、锁屏空白。

方案 B（已定）：仅 invalidToken 触发删行 + 同拍 start；临时错误不动库。

## Goals / Non-Goals

**Goals:**

- `update` / `upsert→update`：APNs 返回 invalidToken 时删除该 activity 行，并对**同一 `device_key`** 的 `push_to_start` 发 `start`（复用本次 payload 的 ContentState / Attributes / alert）。
- 临时错误（非 invalid）仅日志，保留 token。
- `end`：invalid 仍可删脏行，禁止 failover start。
- start failover 失败仅 WARN，禁止同循环再试。

**Non-Goals:**

- 把「任意 APNs 非 200」当成手动删除（方案 A）。
- 改 `isApnsInvalidToken` 判定集合（沿用现网 device/LA 共用逻辑，除非联调证明缺 reason 再开变更）。
- 改 history Refresh 组装、register API、lead 窗截断。
- 跨 device_key 整 wx 清 activity 后全量 start（避免误伤其它机）。

## Decisions

1. **失效判定 = 现有 `invalidToken`（方案 B）**  
   复用 `SendLiveActivityAPNs` → `isApnsInvalidToken`。  
   *备选 A 任意错误 — 429/5xx 误杀，否决。*

2. **自愈挂在 push `dispatchLiveActivity` 内**  
   history 仍只发 `upsert`；失效感知与 APNs 同进程，避免跨服务回调。  
   *备选：history 根据 push HTTP 错误重试 — Fanout 异步且不回传 per-token 结果，否决。*

3. **按 `device_key` 对齐 push_to_start**  
   失效 activity 行带 `device_key`；只向同 key 的 `push_to_start` 发 start。无匹配行则 WARN 跳过。  
   *备选：对该 wx 全部 push_to_start 广播 — 多机误起，否决。*

4. **实现形状**  
   在 update 发送循环中：`invalid && sendEvent=="update"` → `DeleteLiveActivityTokenByID`（已有）→ 查同 `device_key` 的 push_to_start → 构造 `event=start` 的 payload（Attributes 补齐 `activityId` 逻辑与现 start 分支一致）→ `SendLiveActivityAPNs` 一次。  
   可抽小函数 `restartLiveActivityViaPushToStart(ctx, wxID, deviceKey, payload)` 保持 dispatch 可读。

5. **多 activity token 同 wx**  
   逐 token 处理：仅对失败且 invalid 的那条做上述 failover；其它机 update 成功则不动。

6. **end 路径**  
   保持现有：invalid 删行 + 末尾 `cleanupActivityTokensAfterEnd`；不 start。

## Risks / Trade-offs

- [APNs 误报 invalid 导致多余 start] → 沿用既有 invalid 集合；可观测日志 `la_update_invalid_restart`。  
- [无 push_to_start] → 无法同拍恢复；依赖客户端再次 register；接受。  
- [start 成功但客户端未回传新 activity token] → 下次 Refresh 仍走 start（无 activity），行为正确。  
- [同拍 start 与客户端本地残留活动冲突] → 用户已划掉或 token 已废；Apple 侧以新 start 为准。  
- [异步 Fanout 内连续两次 APNs] → 可接受延迟；禁止同步阻塞 history。

## Migration Plan

1. 仅 push-service 发布；开关仍 `PUSH_LIVE_ACTIVITY_ENABLED`。  
2. 回滚：回退二进制即去掉 failover（删 invalid token 的旧行为可保留或一并回退）。

## Open Questions

- （无阻塞）若实网出现「活动已 dismiss 但 reason 不在 invalid 列表」，再增量扩 reason；本变更不预扩。
