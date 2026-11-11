## Context

真机验证：注释掉全部 `aps.alert` 后，`event=start`（push-to-start）无法弹出 Live Activity；`update` 无 alert 仍可静默刷新 content-state。本变更从「全事件省略 alert」修正为「仅 start 写 alert」。

## Goals / Non-Goals

**Goals:**

- `event=start`（含 invalid→restart）APNs 体 MUST 含 `aps.alert`，保证呈现。
- `event=update` / `end` MUST NOT 含 `aps.alert`，避免反复叮/横幅。
- Refresh upsert 恢复伴随文案透传；组装层按 event 决定是否落盘。
- start 时若 title/body 皆空，组装层用默认文案兜底，避免某路径忘填再次「有 start 无弹」。

**Non-Goals:**

- 不改 ContentState / lead 窗 / end dismissal。
- 不改普通非 LA 推送。
- 不在本变更强制 `sound: ""`（start 可能仍有一声；可后续另开）。

## Decisions

1. **组装层按 event 分流（主闸）**  
   - `start`：写 `aps.alert`（优先 payload 文案，否则默认「宝宝动态」/「有新的动态」）。  
   - `update`/`end`：不写 alert，即使 payload 带了 title/body。

2. **Refresh 恢复填 AlertTitle/Body**  
   upsert 始终带「宝宝动态」+ `liveActivityAlertBody`；真正进 APNs 的只有变成 start 的发送。

3. **不改 internal JSON 字段名**  
   `alertTitle`/`alertBody` 继续透传。

4. **验收**  
   - start 的 `apns_body` 含 `"alert"`；update/end 不含。  
   - 真机：无 activity token 时能弹出；已有 activity 时 update 静默；空列表 end 仍可摘。

## Risks / Trade-offs

- [start 带 alert 可能叮一声/出横幅] → 为呈现所必需；update 仍静默。  
- [默认兜底文案与业务文案不一致] → 仅空字段时触发；Refresh 正常路径有成品句。

## Migration Plan

1. 部署 **push-service** + **history-service**。  
2. 回滚：恢复「全省略」或「全带 alert」旧注释策略即可。  
3. 无 DB / env 变更。

## Open Questions

- 无（真机已确认 start 必须有 alert）。
