## Context

ActivityKit 远程 **end** 与普通通知 dismiss 不同：必须打到 **per-activity update token**（`token_kind=activity`），且 APNs 体含 `event=end`。Apple：省略 `dismissal-date` 时 ended 活动可在锁屏再挂最多约 4 小时；要立刻摘掉须把 `dismissal-date` 设为 **过去时刻**。

现网：

- 注释写 `DismissalDate==0` 表示立即，实现却是 `>0` 才写入 → 业务从未传该字段 → **省略**。
- Timing end ContentState 仅 `{phase:ended}`；Flutter 缺省 `startTime=0` 且 UI 不认 phase → 条还在时从 epoch 计时。
- Imminent 清场 `{phase:cleared}` 同理可能把 `nextAt` 解成 0。
- `DeleteDeviceHistory` 不扇出 Timing end。
- start 后客户端会上报 activity token；清场无效与「没钥匙」不是同一类问题。
- 锁屏落到「事件计时中 / 事件即将发生」：Flutter 只读 ContentState；Push-to-Start 常只稳套 Attributes。Imminent 倒计时可回退 `attributes.nextAt`，标题没有回退。

## Goals / Non-Goals

**Goals:**

- 所有 LA `event=end` 的 APNs 体 MUST 含过去 unix 秒的 `dismissal-date`。
- Timing 闭合与删除均按 `timing:{deviceNo}:{historyId}` 发可立即摘条的 end。
- end ContentState 不得让客户端把计时/倒计时字段解成 0。
- start 固定 Attributes 携带锁屏展示名（及 Timing `startTime`），使 Push-to-Start 未套用 ContentState 时仍能出事件名。

**Non-Goals:**

- 改 token 种类、register path、push_to_start 起活动。
- Flutter 按 phase 停表（仍可选）。
- 用 push_to_start 控显隐。
- 改 Imminent pending 挂点或 1/min。
- register activity token 后再发一轮 update 作为标题主路径。

## Decisions

1. **在 APNs 组装层兜底立即 dismiss。**  
   `buildLiveActivityAPNsBody`：当 `event=end` 时，若 `DismissalDate<=0` 则写入 `time.Now().Unix()-1`（或等价过去秒）。业务可不改每个调用点。  
   *备选：每个 end 调用手写 DismissalDate — 易漏删除/清场；否决。*

2. **end ContentState 带齐解码所需字段。**  
   Timing end：`phase=ended`，并带本次已知的 `startTime`/`eventName`（删除路径若库行已无，`startTime` MAY 省略但 MUST NOT 显式写 0；有缓存/调用方数据则带上）。Imminent end：`phase=cleared`（或换槽等价），保留旧 `nextAt`/`title` 若调用方有。  
   *备选：只靠 Flutter 认 phase — 本仓仍须保证 dismiss；ContentState 是防花屏。*

3. **删除走独立 end，不复用 `endTime==0` 早退。**  
   `maybeEndTimingLiveActivity` 今日 `endTime==0` 直接 return。删除应调用「强制 end」（例如 `endTime` 传 `time.Now().Unix()` 或拆 `endTimingLiveActivity` 不检查 endTime）。activity id 仍 `timing:{deviceNo}:{id}`。已闭合再删：再发 end 幂等（无 token 则 skip）。

4. **展示名写入固定 Attributes，Flutter state 空则回退。**  
   Timing start attrs 增 `eventName`、`startTime`；`eventName` 空则 `notifyEventDisplayName`（主档优先）。Imminent start attrs 增 `title`（与 CS 同一 alert 句）。CS 仍写同名字段。兄弟仓 Attributes 解码缺省空串/0；锁屏标题与 Timing 计时：state 空则读 attributes。  
   *备选：仅靠 start 后 update CS — activity token 有竞态；否决为主路径。*

5. **phase 停表仍不作为本仓必做。**  
   dismiss 正确后条应消失；残留花屏已由 end CS 不抹零减轻。

## Risks / Trade-offs

- [activity_id 与上报不一致] → 日志 `no_la_token`；本变更不改 id 格式。联调对照表字段。  
- [dismiss 后 token 行仍在] → 可接受；下次 end 无活动时 APNs 可能 invalid 再删行。  
- [过去 dismissal-date 与时钟偏差] → 用 `now-1` 秒足够；禁止用 0 或省略。  
- [仅 Go 上 Attributes、旧 App 不读] → 旧客户端忽略未知键，仍走 CS；须与 Flutter 同步发版才稳定出标题。

## Migration Plan

1. 合并后与 Flutter 真机：闭合/删除 Timing、PUT pending 立刻摘条；start 锁屏中间为事件名而非「事件计时中 / 事件即将发生」。  
2. 回滚：去掉 attrs 展示名与过去 `dismissal-date`。

## Open Questions

- 无（已定：过去 `dismissal-date`、删除 Timing end、展示名进 Attributes + Flutter 回退）。
