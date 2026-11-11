## 1. Start 前 push_device 闸

- [x] 1.1 在 `dispatchLiveActivity`：当 `sendEvent=="start"` 时调用 `ListPushDevicesForWx`；空则 skip 并打 `no_push_device` 日志（中文注释说明近似代理通知开关）
- [x] 1.2 确认 `update`/`end` 路径不进入该闸

## 2. 删除 invalid→restart

- [x] 2.1 去掉 update invalid 后对 `restartLiveActivityViaPushToStart` 的调用；保留 invalid 删 activity 行
- [x] 2.2 删除 `restartLiveActivityViaPushToStart` 函数及相关注释/日志引用
- [x] 2.3 更新 `FanoutLiveActivity` / APNs 文件中仍写「invalid→restart」的中文注释

## 3. 验收

- [x] 3.1 无 `push_device`、有 push_to_start：upsert→start 被 skip，日志含 `no_push_device`
- [x] 3.2 有任意 channel 的 `push_device`：start 仍可发出
- [x] 3.3 update invalid：只删 activity、无同拍 start
