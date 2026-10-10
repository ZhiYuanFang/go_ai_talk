## 1. Admin 列表 API

- [x] 1.1 `api/v1/cash_xiaozhi_mcp_http.go`：新增 `GET /cash/admin/api/xiaozhi-mcp/entitlements` 请求/响应 DTO（limit/offset、total、note、list 行字段含 wxId/nickname/unlockMethod/channelRef/unlockedAt/expiresAt/active/remainingSeconds/status/revokedAt/updatedAt）
- [x] 1.2 cash 领域：实现 `AdminListXiaozhiMcpEntitlements`——分页读 `xiaozhi_mcp_entitlement`（全表、按 updated_at/wx_id 倒序），计算 active/remainingSeconds，经 `ucgclient.FetchUcgNicknames` 补昵称（失败降级空昵称）
- [x] 1.3 `CashXiaozhiMcpController` 绑定 list 方法；鉴权与其它 cash Admin 一致

## 2. Hub 静态页区 C

- [x] 2.1 `cash-xiaozhi-mcp-admin.html`：区 B 下增加「区 C · 已开通」表格（列对齐 design）、刷新与简单分页（limit/offset）
- [x] 2.2 行内撤销：仅 `active && unlockMethod===admin` 显示按钮；调用既有 `POST /cash/admin/api/xiaozhi-mcp/grants/revoke`；成功后刷新列表
- [x] 2.3 进页加载区 C；区 B 授/撤成功后刷新区 C

## 3. 自检

- [x] 3.1 确认路径落在既有 cash Admin 反代与口令鉴权内；无 App 新路径、不改 `maintenance_skip`
- [x] 3.2 跑 `hack/check-service-import`（cash 不 import 他域服务实现包）
- [x] 3.3 手工走通：有权益数据时区 C 有行与昵称（或空昵称）；admin 行可撤；payment/trial 无行内撤；撤销后列表刷新为无效
