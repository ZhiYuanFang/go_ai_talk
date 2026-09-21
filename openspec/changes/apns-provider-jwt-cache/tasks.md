## 1. APNs Provider JWT 缓存

- [x] 1.1 在 `push_apns.go` 增加进程内 mutex + JWT 字符串 + 过期时间（对齐 `hmsAccessToken` 模式）
- [x] 1.2 改造 `apnsBearerToken`：缓存命中直接返回；未命中时签发并写入缓存；复用窗口约 45–50 分钟（≥20min、<60min）
- [x] 1.3 确认 `isApnsInvalidToken` 不将 `TooManyProviderTokenUpdates` / 429 视为无效 device token（必要时显式排除并补中文注释）
- [x] 1.4 为改动补充文件/方法级中文业务注释（签发、缓存、刷新、与 device token 区分）

## 2. 校验与发布注意

- [x] 2.1 本地或 staging 连续触发两次以上 APNs 发送，确认日志无新的 `TooManyProviderTokenUpdates`（或对比改造前后行为）
- [x] 2.2 确认运维轮换 APNs `.p8`/keyId 后需重启 push-service 以清空 JWT 缓存（可写在变更说明或代码注释）
