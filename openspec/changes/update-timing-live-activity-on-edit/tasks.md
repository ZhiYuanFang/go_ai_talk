## 1. Timing update 扇出

- [x] 1.1 `updateTimingLiveActivity`：按 `timing:{deviceNo}:{historyId}` `action=update`；CS 含 `phase=timing`、主档优先的 `eventName`、已知 unix 秒 `startTime`（禁止写 0）
- [x] 1.2 `UpdateDeviceHistory` 成功且 `endTime==0` 时调用；`endTime!=0` 仍只 end

## 2. 验收

- [ ] 2.1 真机：改未闭合 history 的开始时间或名称，锁屏计时/标题更新且条不消失
- [ ] 2.2 真机：闭合仍立刻摘条（既有 end）

### 联调说明

`PUSH_LIVE_ACTIVITY_ENABLED=true` 且该条已有 `token_kind=activity`、`activity_id=timing:{deviceNo}:{historyId}`。改未闭合行后看 `[history-la]` / `[push-la] action=update`；锁屏应变且不 dismiss。闭合仍走 end + 过去 `dismissal-date`。
