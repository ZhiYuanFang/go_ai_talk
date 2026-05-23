## 1. 实现缓存重建

- [x] 1.1 在 `admin.go`（或 `cache_rebuild.go` 旁）新增 `refreshEventOptionsCacheAfterMutate`，内部调用 `RebuildEventCache` 并打失败日志
- [x] 1.2 将 `AddEvent`、`UpdateEvent`、`DeleteEvent` 写库成功后的 `ListEvents`+`setEventOptions` 替换为 `refreshEventOptionsCacheAfterMutate`
- [x] 1.3 同样替换 `InsertOrGetEventByNeedle`、`ApplyDeepSeekEventExtractPersistence` 及其它写 `event` 表后错误刷新的调用点（`grep ListEvents` 核对）

## 2. 验收

- [x] 2.1 修改事件 `color` 后：DB 与 Redis 事件 options JSON 一致（或 `ListEvents` 返回新 color）
- [x] 2.2 修改 `logo` 后：内部/管理 `event/list` 返回新 `logo` path
- [x] 2.3 删除事件后：缓存列表不再含该 id
