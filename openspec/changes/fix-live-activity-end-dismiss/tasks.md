## 1. APNs end 立即 dismiss

- [x] 1.1 `buildLiveActivityAPNsBody`：`event=end` 时 MUST 写入 `dismissal-date`；`DismissalDate<=0` 填过去 unix 秒（如 `now-1`），禁止省略
- [x] 1.2 更正 `DismissalDate` 注释：0 表示「组装层按立即 dismiss 填过去时刻」，不是「省略即立即」

## 2. Timing 闭合 / 删除

- [x] 2.1 Timing end ContentState：闭合路径带 `phase=ended` 以及已知的 `startTime`/`eventName`，MUST NOT 把已知 `startTime` 写成 0
- [x] 2.2 `DeleteDeviceHistory` 成功后按 `timing:{deviceNo}:{id}` 扇出 Timing end（不因 `endTime==0` 跳过）；无 token 跳过即可

## 3. Imminent 清场

- [x] 3.1 Imminent end/清场 ContentState 保留已知 `nextAt`/`title` 与 `phase`；`dismissal-date` 依赖 1.1 兜底

## 4. 锁屏标题（Attributes 回退）

- [x] 4.1 Timing start Attributes 写入 `eventName`、`startTime`；`eventName` 空则主档回填；ContentState 同步
- [x] 4.2 Imminent start Attributes 写入 `title`（与 CS 同一 alert）
- [x] 4.3 兄弟仓 Flutter：Attributes 解码新字段；锁屏标题 / Timing 计时在 state 为空时回退 attributes

## 5. 验收

- [ ] 5.1 真机：Timing 闭合后锁屏对应条立刻消失，不得残留错乱计时
- [ ] 5.2 真机：PUT imminent pending 后 Imminent 条立刻消失，不得误 end Timing
- [ ] 5.3 真机：删除未闭合 history 后对应 Timing 条立刻消失
- [ ] 5.4 真机：Timing/Imminent start 锁屏中间为事件名/临近句，不得停在「事件计时中 / 事件即将发生」

### 联调说明（部署后真机）

前提：`PUSH_LIVE_ACTIVITY_ENABLED=true`，该 wx 已有 `token_kind=activity` 且 `activity_id` 与扇出一致。

| 操作 | 期望 | push 日志 |
|------|------|-----------|
| 闭合未合计时（end-latest / 写 endTime） | Timing 锁屏立刻消失，不得从 00:00/超大数继续走 | `action=end` `kind=timing` `send_ok` |
| `PUT …/predict/imminent/pending` | Imminent 条立刻消失；已亮的 Timing 仍在 | `action=end` `kind=imminent` |
| 删除未闭合 history | 对应 `timing:{deviceNo}:{id}` 立刻消失 | 同上 timing end |
| Timing / Imminent start | 中间为「吃奶计时中」/「宝宝要睡觉了」类文案，不是兜底句 | start 载荷 attrs 含 eventName/title |

APNs 体 `event=end` 必含 `dismissal-date`（过去秒）。`no_la_token` 仍表示 activity_id 与表对不上，与本变更无关。
