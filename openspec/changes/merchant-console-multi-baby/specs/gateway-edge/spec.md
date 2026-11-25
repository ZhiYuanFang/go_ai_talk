# Spec Delta

## ADDED Requirements

### Requirement: gateway SHALL 校验商户 JWT 并与 wx 会话隔离
gateway-app-server MUST 识别商户后台路由的 Bearer 为 **商户 JWT**（issuer/claim 与 wx access 可区分）。校验成功 MUST 注入商户身份头（如 `X-Internal-Merchant-Id`）。MUST NOT 将商户 JWT 的 `sub` 当作 `wxId` 注入 `X-Internal-Wx-Id`。wx App 路由 MUST 继续只接受 wx access。

#### Scenario: 商户路由接受商户令牌
- **WHEN** 请求携带有效商户 JWT 访问商户 API 前缀
- **THEN** 网关 MUST 放行并注入商户身份头

#### Scenario: wx 令牌访问商户 API
- **WHEN** 请求携带 wx access 访问商户专属 API
- **THEN** 网关 MUST 拒绝

### Requirement: gateway SHALL 提供商户多宝宝历史 SSE
gateway-app-server MUST 注册商户 SSE 端点（路径以实现锁定），`Content-Type: text/event-stream`，鉴权为商户 JWT。连接后 MUST 按该商户当前宝宝 `device_no` 集合订阅与 App 历史通知相同的 Redis 通道，过滤后下行，每帧 MUST 含宝宝标识（`deviceNo` 或稳定 babyId）。宝宝集合变更（创建/删除归属）后 MUST 使订阅集更新（重载或热更新）。反代禁缓冲 MUST 路径限定。MUST NOT 将商户 SSE 计入 WebSocket usage 的错误桶（与历史 SSE 同类处理）。

#### Scenario: 合法商户打开 SSE
- **WHEN** 有效商户 JWT 请求商户历史 SSE
- **THEN** 网关 MUST 建立 event-stream 且按名下 device 集合订阅

#### Scenario: 无令牌拒绝
- **WHEN** 无 Bearer 请求商户 SSE
- **THEN** MUST NOT 建立流

### Requirement: gateway SHALL 托管或反代商户 Vue SPA
gateway（或约定边缘）MUST 向商户提供 SPA 入口（静态托管或反代），与 `/device/admin` Hub 入口分离。商户 SPA 静态资源路径 MUST 列入适当鉴权豁免（仅静态，API 仍需商户 JWT）。

#### Scenario: 商户 SPA 可加载
- **WHEN** 用户打开商户台站点根路径
- **THEN** MUST 返回 SPA 壳且 MUST NOT 要求 Hub Admin 登录
