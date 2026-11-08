## Context

现网 LA：history C/U/D（`afterHistoryMutationLiveActivity`：Del pending + Refresh）、pending PUT、MQ fire、active-eta PUT 都会触发 `RefreshBabyLiveActivity`。C/U/D 先清空 pending 再投影，易推出「仅进行中」；多入口连打导致同 wx 多次 APNs。产品约定：C/U/D 后客户端会重传预测，故 LA **只**在预测变更（pending PUT）与预测到点（MQ fire）投影；active-eta 不推 LA；同 deviceNo trailing 300ms 合并。

约束：history 不 import voice；Refresh 仍在 history；voice 经 `clients/history`；普通推送闸门不变；禁止新增扫表 ticker（本设计的 `time.AfterFunc` 仅为按 deviceNo 的短窗合并，不是 reconciler）。

## Goals / Non-Goals

**Goals:**

- C/U/D：仅 `invalidateImminentPending`，不 Refresh。
- pending PUT / MQ fire：经合并入口调度 Refresh。
- active-eta：去掉 `refreshUnifiedLiveActivity`。
- `ScheduleRefreshBabyLiveActivity(deviceNo)`（名可调整）：trailing **300ms**，最后一次触发后执行一次真实 `RefreshBabyLiveActivity`；扇出仍按绑定 wx，每 wx 本轮一次。

**Non-Goals:**

- 改 lead 过滤、items≤3、invalid→push_to_start、普通 push 文案/闸门。
- 客户端强制 PUT 的实现（契约依赖，本仓只声明）。
- 跨进程分布式 debounce（单 history 进程内存 map 即可；多实例最坏重复一次，可接受）。

## Decisions

1. **触发矩阵**  
   | 入口 | Del pending | LA Refresh |  
   |------|-------------|------------|  
   | history C/U/D | 是 | 否 |  
   | pending PUT | 否（整表替换） | 是（合并） |  
   | MQ fire | 否 | 是（合并） |  
   | active-eta | 否 | 否 |  
   *备选：U 仍 Refresh — 产品已确认 U 也会重传预测，否决。*

2. **拆分 CUD 旁路**  
   `afterHistoryMutationLiveActivity` 改为只 invalidate；或拆成 `invalidate…` + 调用方不再 Refresh。禁止 CUD 路径调用 schedule Refresh。

3. **trailing 300ms 挂在 history**  
   voice PUT/fire 调 `clients/history` 的 Refresh HTTP → history 侧 **schedule** 而非立即跑；本进程内 CUD 不再 schedule。  
   实现：`sync.Mutex` + `map[deviceNo]*time.Timer`，每次 schedule `Stop` 旧 timer，`AfterFunc(300ms, runRefresh)`。  
   *备选：push 侧去重 — 减 APNs 但不减错误时机投影；否决为主方案。*  
   *备选：singleflight — 盖不住串行连打；否决为主方案（可作辅）。*

4. **空 pending PUT**  
   仍 schedule Refresh（可能仅进行中或 end）——属于「预测变更」。

5. **普通推送**  
   MQ fire 仍 `PushByBizType`；iOS LA 开时仍 `la_replaces_apns_alert`。

## Risks / Trade-offs

- [C/U/D 后至 pending PUT 前 LA 短暂陈旧] → 接受；换「仅进行中」空窗包。  
- [客户端漏 PUT] → LA 不更新直至到点/下次 PUT；依赖 App 约定。  
- [多 history 实例各 debounce] → 可能双发；低频可接受。  
- [进程退出丢失未触发 timer] → 下次 PUT/fire 补偿。  
- [AfterFunc 被误认为扫表 ticker] → design/OpenSpec 声明：按键短窗合并，无周期扫库。

## Migration Plan

1. 改 CUD / active-eta / schedule Refresh；发 history + voice。  
2. 回滚：恢复 CUD+eta Refresh、去掉 debounce。

## Open Questions

- （无）多实例强一致 debounce 不做。
