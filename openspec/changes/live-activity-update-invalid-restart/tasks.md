## 1. Token 查询辅助（按需）

- [x] 1.1 若现有 API 不足：增加按 `(wx_id, device_key, token_kind=push_to_start)` 取 token 的辅助（或在 List 结果中按 device_key 过滤）；中文注释说明仅用于 update invalid 自愈

## 2. dispatch 同拍自愈

- [x] 2.1 在 `dispatchLiveActivity` 发送循环：当 `sendEvent==update` 且 `invalidToken` 时，删除该 activity 行后，对同 `device_key` 的 push_to_start 发 `start`（复用 ContentState/Attributes/alert；Attributes 补 `activityId` 与现 start 分支一致）
- [x] 2.2 非 invalid 的 update 失败：不删、不 restart；仅日志
- [x] 2.3 `sendEvent==end`：invalid 可删行，MUST NOT restart；保留既有 end 后 cleanup
- [x] 2.4 start failover 失败仅 WARN，同 dispatch 不重试；无 push_to_start 时 WARN 跳过
- [x] 2.5 可观测日志（如 `la_update_invalid_restart`）含 wxId、deviceKey、tokenId

## 3. 校验

- [x] 3.1 `gofmt`；编译 push 相关包
- [x] 3.2 自检：update invalid→删+同机 start；429 类失败不动 token；end invalid 不 start
