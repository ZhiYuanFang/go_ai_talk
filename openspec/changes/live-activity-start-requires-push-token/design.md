## Context

关消息通知 ≈ Flutter 调 `/app/api/push/unregister` → 无 `push_device`。LA token 分表独立，关通知后仍可能 `upsert→start`。产品要求暂不改 Flutter：start 前看有无任意 push token；并删除用不到的 invalid→restart。

## Goals / Non-Goals

**Goals:**

- `sendEvent=start` 前：`ListPushDevicesForWx(wxId)` 为空则 skip（日志 `no_push_device`）。
- update/end 不受影响。
- 删除 `restartLiveActivityViaPushToStart`；update invalid 仅删 activity 行，不同拍 start。

**Non-Goals:**

- 不按 deviceKey / 仅 apns 过滤（暂时任意 channel）。
- 不在关通知时主动 end 已挂 LA。
- 不改 history Refresh、不改 Flutter。

## Decisions

1. **闸位置：push-service `dispatchLiveActivity`**  
   在已确定 `sendEvent == "start"` 且已有 LA push_to_start token 之后、发 APNs 之前查 `push_device`。同包调用，无跨服务。

2. **任意 channel**  
   `ListPushDevicesForWx` 非空即放行；不区分 apns/hms/mipush。

3. **只挡 start**  
   已有 activity 的 update 继续；空列表 end 继续。

4. **删除 restart**  
   去掉函数与 `sendEvent==update && invalid` 分支中的 restart 调用；保留 invalid 删行。下次 upsert 无 activity 时再走 start（再过 push_device 闸）。客户端 clear / 再 register 仍可自然恢复。

## Risks / Trade-offs

- [多机：一机关通知另一机仍有 token → 仍可能 start] → 暂时接受；日后可改 deviceKey 对齐。  
- [关通知后已挂 LA 仍 update] → 产品接受只挡新弹出。  
- [删 restart 后脏 activity 需等下次 upsert/clear] → 产品确认用不到同拍补刀。

## Migration Plan

1. 部署 **push-service**。  
2. 回滚：恢复 restart 函数 + 去掉 push_device 闸。  
3. 无 DB / env / Flutter 变更。

## Open Questions

- 无（任意 token + 只挡 start + 删 restart 已定）。
