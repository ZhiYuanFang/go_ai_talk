## Why

push-service 经 APNs 发送时，每次请求都会重新签发 Provider JWT（`iat=now`）。Apple 对同一 Key 的 Provider token 更新频率有限制（约 20 分钟内不可频繁更换），现网因此出现 `HTTP 429 TooManyProviderTokenUpdates`，导致 `ucg_alert` 等 iOS 推送 `send_failed`。HMS 通道已有进程内 access token 缓存可对齐。

## What Changes

- APNs Provider JWT 改为**进程内缓存并复用**：同一 key 在有效期内复用同一 bearer，禁止每条推送重新签发。
- 刷新节奏 MUST 满足 Apple 约束：缓存生命周期显著短于 JWT 1 小时上限，且两次签发间隔 MUST NOT 短于约 20 分钟（实现取约 45–50 分钟刷新窗口）。
- 并发签发 MUST 互斥（对齐 HMS `hmsTokenMu`），避免短窗内多把不同 `iat` 的 JWT。
- `TooManyProviderTokenUpdates` / 429 MUST NOT 被当作无效 **device** token（不得删除 `push_device` 行）。
- **无 BREAKING** 对外 App/internal 推送 API 契约、载荷或 channel 枚举。

## Capabilities

### New Capabilities

- `push-apns-provider-auth`：push-service APNs 通道的 Provider JWT 签发、缓存、刷新与 429 语义（与 device token 失效区分）。

### Modified Capabilities

- （无）

## Impact

- **代码**：`internal/services/push/push_apns.go`（`apnsBearerToken` 及进程内缓存）；可参考同包 `push_hms.go` 的 `hmsAccessToken` 模式
- **进程**：仅 `push-service`；不引入 Redis 新键、不改配置文件结构、不改跨服务契约
- **不涉及**：HMS/MiPush、gateway usage、新 App 路由、背景 ticker、多实例共享 JWT（本变更仅进程内缓存；多副本冷启动偶发仍可接受）
