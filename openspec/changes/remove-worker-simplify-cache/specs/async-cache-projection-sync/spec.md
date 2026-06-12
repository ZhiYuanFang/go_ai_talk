## REMOVED Requirements

### Requirement: 写入后异步更新缓存投影

**Reason**: 由 `history-device-sync-cache-projection` 取代：写后同步 patch，无 domain_outbox worker relay。

**Migration**: 删除 worker outbox 与 enqueue；见 `remove-worker-simplify-cache` design D1/D3。

### Requirement: 乱序与重复事件保护（domain_outbox version + worker ApplyProjection）

**Reason**: 异步 outbox 投影链路删除；同步 patch 在单请求内顺序执行；读路径 miss 全量重建无版本乱序问题。

**Migration**: 删除 `history.ApplyProjection` 的 outbox 调用链；可选保留 version key 供未来扩展但无 worker 驱动。

### Requirement: 失败补偿与可重建

**Reason**: worker 重试与 projection repair ticker 删除；补偿改为 list key DEL + read miss rebuild + 运维手动 rebuild。

**Migration**: 见 `history-device-sync-cache-projection` REMOVED 条目。
