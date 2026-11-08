## Context

`SendLiveActivityAPNs` 已组装最终 APNs JSON，但发送前后仅有 Warning 级元数据（`send_ok` / `send_failed`）。联调需要看见完整请求体；产品确认 **只加 Info 日志**，不改业务。

## Goals / Non-Goals

**Goals:**

- `buildLiveActivityAPNsBody` 成功后、`client.Do` 前：`g.Log().Infof` 打印 `event`、`tokenLen`、body 字符串。
- 覆盖所有经该函数的路径（update / start / end / invalid→restart 的 start）。

**Non-Goals:**

- env 开关、截断 body、改 logger 配置、改截断/自愈语义、打普通 alert APNs body。

## Decisions

1. **单点打在 `SendLiveActivityAPNs`**  
   避免 dispatch / restart 各打一份漏路径。  
   *备选：仅 dispatch — 漏 restart，否决。*

2. **级别 Info**  
   与现有大量 Warning 元数据区分；push-service `level: all` 下可见。  
   *备选 Warning — 产品已定 Info。*

3. **脱敏**  
   日志字段用 `tokenLen=len(deviceToken)`，MUST NOT 打印完整 token。

4. **body 全量**  
   items≤3 后体积可控；不截断。

## Risks / Trade-offs

- [prod 日志量增加] → 接受；仅 LA 且频次随 Refresh。  
- [body 含宝宝事件名等业务文案] → 与现有 alert 日志同类；不记 token。

## Migration Plan

1. 发布 push-service 即可。  
2. 回滚：去掉 Info 行。

## Open Questions

- （无）
