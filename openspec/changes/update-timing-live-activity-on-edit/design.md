## Context

Timing LA：Add 未闭合 → start；闭合/删除 → end。`UpdateDeviceHistory` 注释写明「未闭合 update 不重复 start」，因此改 `startTime`/`eventName` 锁屏不变。

ActivityKit `update` 只能改 ContentState；Attributes（含 `eventId` 与图）开局锁死。产品决定：**只 update 文案/时间，图标可以旧**。

## Goals / Non-Goals

**Goals:**

- 未闭合 history 更新后锁屏标题与已计时跟新行走。
- 换 `eventId` 不重建活动。

**Non-Goals:**

- end+start 换图标。
- Imminent。
- 无 activity token 时补 start（与 end 一样 skip）。

## Decisions

1. **挂 `UpdateDeviceHistory` 成功路径。**  
   `endTime!=0` → 既有 end。`endTime==0` → `updateTimingLiveActivity`。不在 controller 另钩。  
   *备选：仅当 startTime/eventName 变化才发 — 要读旧行对比；少一次读，接受幂等 update。*

2. **ContentState：`eventName`、`startTime`、`phase=timing`。**  
   `eventName` 空则 `notifyEventDisplayName` 主档回填；`startTime>0` 才写入，禁止显式 0。

3. **换 eventId 不 end+start。**  
   图可旧；标题随 CS `eventName` 变。

4. **无 token skip。**  
   与 end 相同，走 `FanoutLiveActivity(action=update)`。

## Risks / Trade-offs

- [改 eventId 后图与标题不一致] → 产品接受。  
- [start 后尚未 register 就 update] → `no_la_token`，锁屏暂旧；下次改或 end 再试。  
- [Flutter 若 state 空仍回退旧 attributes.startTime] → 本路径会写入非空 CS，state 应覆盖。

## Migration Plan

1. 合并后真机：改未合计时的开始时间/名称，锁屏应变，条不消失。  
2. 回滚：去掉未闭合 update 调用。

## Open Questions

- 无。
