## Context

划掉锁屏后 APNs `update` 仍可能 200，现有 `invalidToken → 删 activity → 同机 push_to_start start` 进不去，导致 Refresh 反复 update、锁屏空白。客户端确认：iOS **原生层**可在 App 未启动时于数秒内 HTTP 回执。触发面已收敛为预测 pending PUT / MQ 到点（`live-activity-trigger-predict-only`），update 频率可控，适合「发后短等 ack」。

约束：跨域经 gateway `/app/api/push/*`；ack 身份对齐 register（body `wxId`、免 Bearer）；Redis/业务键命名走 platform；禁止新增扫表 ticker（本设计 `AfterFunc` 仅为单次超时，非周期 reconciler）；接口版本不改旧 register，新增 ack 路径。

## Goals / Non-Goals

**Goals:**

- update（含 upsert→update）注入临时 `ack`，send_ok 后等待客户端回执。
- 超时未 ack → 删该 activity 行 + 同 `device_key` push_to_start start（复用 `restartLiveActivityViaPushToStart`）。
- 超时毫秒可配：`LIVE_ACTIVITY_ACK_TIMEOUT_MS`，默认 3000。
- App `POST /app/api/push/live-activity/ack`；gateway Bearer 豁免对齐 register；**不计入 usage**。
- activity 行持久化 `last_sent_ack` / `last_acked_ack`（及补刀所需发送时刻）。

**Non-Goals:**

- start / end 不做 ack 超时 failover。
- 不改 invalidToken 判定集合；不改 LA 触发矩阵、lead/截断、普通 alert。
- 不在 widget extension 内实现（客户端原生层，仓库外）。
- 不做跨实例分布式定时器协调（丢 timer 用 DB 状态补刀即可）。

## Decisions

1. **探测模型：发后超时，而非「下次比对」**  
   下一轮 Refresh 才发现对「本发空气 update」太晚。update `send_ok` 后启动超时；到期 `last_acked != last_sent` 则 failover。  
   *备选：仅下次 update 前比对 sent/acked — 否决（用户已明确）。*  
   *备选：双边 ++ 计数 — 否决（非幂等、易漂）。*

2. **临时不透明 ack，注入点在 push dispatch**  
   ContentState 增加 `ack`（与 `items` 并列）；由 **push** 在临发 update 前生成并写入 payload（history 仍只组业务字段）。成功发出后写该 activity 行 `last_sent_ack` + `ack_sent_at`。  
   *备选：history 生成 ack — 否决（多一次跨服务状态，且扇出在 push）。*

3. **仅 update 等待；start 不动**  
   start 后靠客户端 register 新 activity token；不对 start 做同等超时。end 不 ack、不因超时 start。

4. **超时实现：`AfterFunc` + DB 补刀**  
   - 进程内：`send_ok` → `AfterFunc(timeout, checkAndFailover)`。  
   - 补刀：若进程重启丢 timer，下次对该行准备 update 时若 `last_sent != last_acked` 且 `now - ack_sent_at >= timeout`，先走失效重启再决定是否新发。  
   *备选：dispatch 内阻塞 sleep — 否决（多机串行差）。*

5. **ack HTTP 语义**  
   `POST /app/api/push/live-activity/ack`：`wxId`、`deviceKey`、`ack`。命中 `(wx, device_key, activity)`：仅当 `ack == last_sent_ack` 时更新 `last_acked_ack`；否则 no-op 成功或业务可区分日志（防乱序旧包把 acked 打回旧值）。晚到 ack（行已删）→ 忽略。  
   鉴权：**与 register 相同**——gateway 精确豁免 `/app/api/push/live-activity/ack`；落库身份以 body `wxId` 为准。

6. **usage**  
   负责人确认 **不计入**：`maintenance_skip` 增加 `POST /app/api/push/live-activity/ack`（对标 register）。

7. **gateway 路由**  
   既有 `BindMiddleware("/app/api/push/*")` 已覆盖；MUST 补 Bearer 豁免精确 path。push-service 注册 `g.Meta` 路由。

8. **超时配置**  
   `LIVE_ACTIVITY_ACK_TIMEOUT_MS`：解析为正整数毫秒，非法/空 → 默认 3000；建议上限钳制（如 ≤60000）避免误配挂起过久。

9. **与 invalidToken 关系**  
   两条进门、一个出口：APNs invalid 立即删+restart；ack 超时同样删+restart。互不替代。

## Risks / Trade-offs

- [原生 ack 偶发 > 超时] → 误起 start、锁屏闪一下；用可配超时调参；日志打 `la_ack_timeout_restart`。  
- [AfterFunc 进程退出丢失] → DB `ack_sent_at` 下次 update 前补刀。  
- [免鉴权伪造 ack] → 仅接受 `ack==last_sent`，伪造成功只是「假装收到」，划掉场景仍可能被下次超时纠正；接受与 register 同级风险。  
- [ContentState 多 `ack` 字段] → UI 忽略；需客户端 ContentState 模型容忍未知/新字段。  
- [AfterFunc 被误认为扫表] → 注释与 OpenSpec 声明：单次超时，无周期扫库。

## Migration Plan

1. schema ensure 加列；实现 dispatch/ack/豁免/skip；登记 env 默认。  
2. 与客户端约定 ContentState.`ack` 与 ack API 后联调。  
3. 回滚：关超时（或将 timeout 配极大并停 failover 代码路径）+ 去掉豁免不影响 register。

## Open Questions

- （无）网关豁免对齐 register、不计入 usage 已确认；超时默认 3s 可配已确认。
