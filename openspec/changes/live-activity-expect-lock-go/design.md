## Context

Flutter twin `live-activity-expect-lock` 约定方案 B：Go 推结构化字段，Swift 拼短文案。Imminent pending / fire payload 已含 `expectKind` 等；active-eta Redis 已含 `expectEndAt` + `historyId`。当前 `startImminentLiveActivityForWx` 只写 title/phase/nextAt，且开火把带「预计这次约…」的长 alert 整段塞进 LA title；Timing start 只写 eventName/startTime/phase；`SyncPredictActiveEta` 注释写死 MUST NOT 推送；history `updateTimingLiveActivity` 只推 phase/eventName/startTime，APNs ContentState 整包替换会冲掉已灌入的 `expectEndAt`。

跨服务：Timing start 在 **history-service**；ETA 在 **voice-service**。voice 已可经 `clients/push` 扇出 LA，不必回调 history。history 可通过 `cachekit.PredictActiveEtaKey` 只读 Redis（平台键），**禁止** import `services/voice`。

## Goals / Non-Goals

**Goals:**

- Imminent start ContentState 透传可选 expect*。
- Imminent LA 标题仅为 `{昵称}要{事件名}了`；预计中文后缀只留在普通推送 alert。
- Timing ContentState 支持 `expectEndAt`；以 active-eta 同步后的 **update** 为灌入主路径。
- history 改计时的 Timing update 在 ETA 可解析时回写 `expectEndAt`，不冲掉副行。
- 有值才写键；0 / 空 omit；unix 秒；amount 浮点。
- 修正 active-eta「禁止一切推送」→ 允许 LA update，禁止 MQ/普通 alert。

**Non-Goals:**

- Go 拼「预计 HH:mm 结束」/ 锁屏右侧短时长（客户端职责）。
- 改 pending HTTP 契约字段名。
- Flutter Extension / Island。
- Timing start 时强依赖已有 ETA（竞态常见：先 start 后 ETA）。
- 改 Android 通知拼装算法（仅与 LA 标题输入拆开）。

## Decisions

1. **Imminent 结构化 expect**  
   扩展 `startImminentLiveActivityForWx`（或等价）接收 expect 结构化参数；`expectKind` 为 `duration`/`amount` 时按维写入 ContentState。duration 仅当 `expectDurationSec>0`；amount 仅当 `expectAmount>0`；`expectUnit` 非空才写。不写 `expectLeafTitle` 进 LA。

2. **Imminent 标题拆分**  
   开火路径：`laTitle = "{昵称}要{事件名}了"`；`alert = laTitle + buildPredictImminentExpectSuffix(...)`（后缀可空）。  
   - LA：Attributes.`title`、ContentState.`title`、以及 LA 的 AlertTitle/AlertBody（若沿用 title 参数）使用 **`laTitle`**。  
   - `PushByBizType`：继续用 **`alert`**（可含「，预计这次约…」）。  
   避免锁屏主标题与右侧结构化预计叠两句「预计」。

3. **Timing + active-eta**  
   `SyncPredictActiveEta` 写 Redis 成功后：遍历 cleaned items，解析正 `historyId`，若 `expectEndAt>0`，对绑定用户列表 `FanoutLiveActivity` action=`update`，`attributesKind=timing`，`activityId=timing:{deviceNo}:{historyId}`，ContentState 至少含 `phase=timing`、`expectEndAt`；有 `title`/`startAt` 时可一并带上以免 update 覆盖丢字段。空列表仅清 Redis，不批量 end Timing。

4. **history Timing update 保留 expectEndAt**  
   修正原「update 可不写 expectEndAt、靠 Flutter 再 sync」：因 APNs `content-state` **整包替换**，history `updateTimingLiveActivity` MUST 在扇出前只读 `PredictActiveEtaKey`，按字符串 `historyId` 匹配当前行；若解析到正 `expectEndAt` 则写入 ContentState。读失败 / 无匹配 / ≤0 → 省略键（不阻断改名/改 startTime 的 update）。不 import voice；JSON 结构与 `PredictActiveEtaItem` 字段名对齐即可（本地小 struct 或匿名解析）。Timing **start** 仍不强制读 ETA。

5. **仍禁止** logo/color；时间字段秒级。

## Risks / Trade-offs

- [active-eta 改推送边界] → 规格明确仅 LA update，无 alert/MQ。  
- [update 整包替换丢字段] → active-eta 与 history update 两侧都尽量带齐 eventName/startTime/expectEndAt。  
- [history 读 voice Redis] → 仅经 platform `cachekit` 键；评审禁止 `import .../services/voice`。  
- [改 startTime 后 ETA 暂时不准] → 副行保留旧 expectEndAt 直至下次 sync；优于闪没。  
- [historyId 解析失败] → 跳过该条 LA，打 debug/warn，不影响 Redis 成功。  
- [先 LA 后 ETA] → 副行延迟出现，可接受。

## Migration Plan

1. 与 Flutter twin 同发或 Go 稍后；缺字段客户端静默。  
2. 回滚：停写 ContentState 新键、恢复长 alert 进 LA title 即可（不推荐）。

## Open Questions

- 无。
