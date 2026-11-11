## Context

统一 Live Activity 由 history `RefreshBabyLiveActivity` 组装 ContentState，经 `pushclient.FanoutLiveActivity` 到 push-service，最终在 `buildLiveActivityAPNsBody` 写 APNs JSON。

当前 upsert（非空 items）固定传：

- `AlertTitle = "宝宝动态"`
- `AlertBody = liveActivityAlertBody(items)`

组装层在 title/body 非空时写入 `aps.alert`。`alert` 只负责额外提醒（横幅/声音），锁屏条内容由 `content-state` 决定。产品要求注释/停用该 alert，使 LA 推送静默更新。

## Goals / Non-Goals

**Goals:**

- 所有经共享 LA APNs 路径发出的 start/update/end **请求体均不含 `aps.alert`**。
- Refresh 侧不再填充伴随文案，避免误导后续维护者「还有 alert」。
- 既有 Info 级 `apns_body` 日志可直接验收「无 alert 键」。

**Non-Goals:**

- 不改 ContentState / Attributes / lead 窗 / end+dismissal。
- 不改普通 APNs `predict_imminent`（非 LA）与其它 biz 推送。
- 不引入 `sound: ""` 或静音音频资源（本次是整段省略 alert，不是只关声）。
- 不删除 `LiveActivityOp.AlertTitle/AlertBody` 字段（可留空兼容 internal JSON；避免无关契约抖动）。

## Decisions

1. **在 APNs 组装层硬省略 alert（主闸）**  
   在 `buildLiveActivityAPNsBody` 注释或删除写入 `aps.alert` 的分支，使任何调用方传入的 title/body 都不进载荷。  
   - 备选：只在 Refresh 清空 title/body → 拒绝，因 invalid→restart 等路径仍可能复用字段，组装层硬闸更稳。

2. **Refresh 同步注释赋值（辅）**  
   注释 `AlertTitle`/`AlertBody` 赋值与（若因此无引用）`liveActivityAlertBody`；或保留 helper 但不再调用。优先少留死代码：无引用则注释函数并加一行说明「alert 已省略」。

3. **不改 internal API schema**  
   `alertTitle`/`alertBody` JSON 字段可继续存在，值为空即可；push controller → `LiveActivityOp` 透传不变。

4. **验收靠现有 body Info 日志**  
   联调看 `[push-la] apns_body ...` 中无 `"alert"`；真机确认锁屏条仍随 content-state 更新且无提示音/横幅。

## Risks / Trade-offs

- [部分 iOS 对 push-to-start 无 alert 呈现变差] → 真机验 start；若回归则另开变更恢复最小 alert 或静音 sound，不在本变更内回滚一半逻辑。
- [去掉 alert 后用户少一层「叮一声」提醒] → 产品有意；iOS 临近仍可有锁屏条；Android 普通推送不受影响。
- [struct 仍留 Alert 字段] → 可接受的短债；注释标明已忽略，避免再填业务文案。

## Migration Plan

1. 部署 **push-service**（组装闸）与 **history-service**（Refresh 停填）。
2. 回滚：还原两处注释/删除即可恢复旧伴随提醒。
3. 无需 DB / Redis / env 变更。

## Open Questions

- 无（产品已定为「注释 alert」整段省略，而非只关 sound）。
