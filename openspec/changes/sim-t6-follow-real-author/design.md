## Context

- T6 `RunFollowTask` 经 `pickTwoDistinctSimWx` 选两个 sim，调用 `POST /ucg/app/api/follow/{wxId}`。
- UCG `Follow()` 不区分 sim/真人；限制仅在 sim-user 选目标侧。
- T5 已建立模式：sim-user `listAllSimWxIds()` → ucg internal 传 `simWxIds` → SQL `NOT IN` 排除 sim。

## Goals / Non-Goals

**Goals**

- T6：sim follower + 随机真人 author（有 published 帖）→ follow。
- 复用 `posts/sample` `mode=random` ID 探测，与 T2 同帖源、不同过滤。
- 固定行为，无配置开关。

**Non-Goals**

- sim→sim 关注（删除）。
- 排除已关注、按 T5 peer 优先、admin 比例配置。
- device 新 random-real API。

## Decisions

### 1. 扩展 posts/sample 而非新接口

**选择**：`excludeAuthorWxIds` + 响应 `authorWxId`。

**理由**：random 探测已实现；T6 只需 author，不必拉帖内容做 LLM。

### 2. 真人池 = published 帖作者且不在 simWxIds

```
listAllSimWxIds()
pickRandomSimWx() → follower
POST posts/sample { mode:"random", excludeAuthorWxIds: simWxIds }
→ authorWxId
若 authorWxId == follower.WxId → 重试（最多 simFollowRandomMaxTry）
POST follow/{authorWxId}
```

与 T5 对称：`excludeAuthorWxIds` 等价于 T5 对 peer 的 `WhereNotIn(simWxIds)`。

### 3. 失败语义

| 条件 | 行为 |
|------|------|
| 无 sim 用户 | pickRandomSimWx 失败 |
| 无真人帖（sample 空） | `RecordTaskRun(..., false, "无真人作者")` |
| 重试后仍自关注 | `RecordTaskRun(..., false, "无可用关注目标")` |
| 已关注 | Follow 幂等成功 |

## Risks / Trade-offs

- sim 帖占比高时真人池变小 → 可能频繁「无真人作者」；可接受，与运营目标一致。
- `excludeAuthorWxIds` 列表可达 10000 → 与 T5 相同上限，单次 T6 tick 可接受。

## Migration

1. 部署 ucg-service（posts/sample 扩展向后兼容：未传 exclude 行为不变；T2 忽略新字段 authorWxId）。
2. 部署 sim-user-service（T6 新逻辑）。
