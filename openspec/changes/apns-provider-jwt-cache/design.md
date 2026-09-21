## Context

push-service 的 `ApnsSender.Send` 每次调用 `apnsBearerToken`：读 `.p8`、解析 ECDSA、用当前秒作 `iat` 签 ES256 JWT。Apple 文档要求同一 Provider Key 下更换认证 token 不宜快于约 20 分钟；现网密集推送触发 `429 TooManyProviderTokenUpdates`。同进程 HMS 已用 `hmsTokenMu` + 过期缓冲缓存 OAuth access token，可作为对齐模板。

约束：仅改 `internal/services/push`；禁止为 JWT 引入新 Redis 读缓存（未获负责人批准）；不改 App/internal API；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 进程内缓存并复用 APNs Provider JWT，消除「每推送一次签发」导致的 429。
- 并发安全：多 goroutine 同时 `Send` 时同一时刻至多签发一把新 JWT。
- 保持 device token 失效判定不变：`TooManyProviderTokenUpdates` 不得触发 `DeletePushDeviceByID`。

**Non-Goals:**

- 跨 push-service 实例共享 JWT（Redis / 中心化签发）。
- 429 自动重试退避队列（可后续增量；本变更以缓存消除主因）。
- 更换 HTTP/2 客户端库、改 sandbox/production 切换、改 `.p8` 配置来源。
- 修改 HMS/MiPush 或 dispatcher 扇出语义。

## Decisions

### 1. 进程内缓存对齐 HMS，不用 Redis

- **选择**：包级 `sync.Mutex` + 缓存字符串 JWT + 过期时间（或签发时间 + TTL）。
- **理由**：与 `hmsAccessToken` 一致；单变更面最小；AGENTS.md 禁止默认引入新 Redis 读缓存。
- **备选**：Redis 共享 JWT → 多副本更整齐，但本问题主因是「每请求签发」，进程内缓存即可；延后。

### 2. 刷新窗口约 45–50 分钟

- **选择**：JWT `iat` 一次签发后复用至约 `45*time.Minute`（或 `50`）再换新；MUST ≥ 20 分钟、MUST < 60 分钟（Apple JWT 有效期上限）。
- **理由**：留足时钟与 Apple 校验余量，又远高于 20 分钟下限。
- **备选**：严格 20 分钟换一次 → 无必要更勤；接近 60 分钟 → 边界风险更高。

### 3. 私钥与 PEM 可读缓存期内再读，或随 JWT 一起缓存

- **选择**：签发路径仍可读 `.p8`；缓存命中时不读盘。可选：把解析后的 `*ecdsa.PrivateKey` 与 JWT 同生命周期缓存，减少重复 I/O（非功能必须）。
- **理由**：命中路径零文件 I/O；未命中时行为与现网一致，便于排查密钥路径错误。

### 4. 429 语义

- **选择**：继续走 `send_failed`；`isApnsInvalidToken` 不把 `TooManyProviderTokenUpdates` 当无效 device token（现状已满足，实现时保持）。
- **理由**：该错误指向 Provider auth，删设备 token 会误伤用户。
- **备选**：收到 429 后强制清空 JWT 缓存并等 ≥20 分钟再签 → 可作为后续加固，本变更不强制实现重试。

### 5. 多实例

- **选择**：每进程独立缓存；接受滚动发布时短窗内多把 JWT（副本数通常很小）。
- **理由**：Non-Goal 排除跨实例共享；副本数有限时冷启动更新次数远低于「每条推送一次」。

## Risks / Trade-offs

- [多副本冷启动仍偶发 429] → 监控；若复发再评估共享缓存或启动抖动。
- [进程长期运行、系统时钟回拨导致过早/过晚刷新] → 用单调墙钟 `time.Now()` + 固定 TTL；异常时可重启进程清缓存。
- [密钥轮换后旧 JWT 仍缓存] → 运维轮换 `.p8`/keyId 后需重启 push-service；文档写入 tasks/运维备注即可。
- [缓存实现 bug 导致永不刷新，JWT 过期] → TTL 必须严格小于 60 分钟；过期后签发失败会体现在 `send_failed`，便于发现。

## Migration Plan

1. 合并后仅滚动 `push-service`。
2. 观察 APNs `send_failed` 中 `TooManyProviderTokenUpdates` 是否消失；`send_ok` 恢复。
3. 回滚：回退该 commit 并重启 push-service（无 DB/配置迁移）。

## Open Questions

- （无阻塞）多副本生产副本数若显著增大且仍见 429，再开增量变更评估跨实例 JWT。
