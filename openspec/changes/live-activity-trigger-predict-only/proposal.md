## Why

统一 Live Activity 后，history **C/U/D** 与 **active-eta** 也会 Refresh，常在 pending 被清空、客户端尚未回写预测时推出「仅进行中」或重复扇出；而 C/U/D 后客户端都会重算并 **PUT 预测**。应将 LA 触发收敛为 **预测变更 + 预测到点**，并用短窗合并避免连打。

## What Changes

- **LA 触发收敛**：仅在（1）imminent **pending PUT** 成功、（2）预测 **MQ fire** 过闸并通知时，调用统一 `RefreshBabyLiveActivity`。
- **history C/U/D**：**MUST** 继续 Del imminent pending（作废旧日程/对账）；**MUST NOT** 再 Refresh / 扇出 LA。
- **active-eta PUT**：**MUST NOT** 触发 LA Refresh（不推送实时活动）。
- **合并**：同 `deviceNo` 对 Refresh 做 **trailing 300ms** debounce；窗结束只投影一次，再按绑定 wx 扇出（每 wx 本轮至多 1 次 APNs）。
- **非目标**：不改普通 `predict_imminent` 闸门/文案；不改 lead 进窗与 items≤3 组装语义；不改 token invalid→push_to_start。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `unified-live-activity-push`：收敛 LA 触发点为预测 PUT + MQ 到点；C/U/D 只清 pending 不 Refresh；去掉 active-eta 触发；增加 deviceNo trailing 300ms 合并。

## Impact

- `internal/services/history/live_activity_refresh.go`、`local.go`：CUD 旁路拆成「仅 invalidate」与「可合并的 Refresh」。
- `internal/services/voice/predict_imminent.go`：pending PUT / fire 仍调 Refresh（经合并入口）。
- `internal/services/voice/predict_active_eta.go`：去掉 `refreshUnifiedLiveActivity`。
- 依赖客户端约定：C/U/D 后会 PUT pending（含清空）；否则 LA 直至下次预测 PUT/到点才更新。
- 无新 App path；usage 不改。
