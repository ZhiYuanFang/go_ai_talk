## 1. 契约确认

- [x] 1.1 与 Flutter twin `live-activity-native-register-no-auth` 核对字段表：`wxId`（非零）+ `deviceKey` / `tokenKind` / `attributesKind` / `token` / `activityId?`；同 path；同发或 Go 先合均可（无线上历史客户端）
- [x] 1.2 **已定**：同 path 原地加 `wxId`，不走 `api/v2`（activity token register 未发版，无历史客户端；见 design 决策 3）

## 2. gateway-app 鉴权豁免

- [x] 2.1 将 `POST /app/api/push/live-activity/register` 加入 `gatewayAppAuthExemptExactPOST`（`gateway_app_auth_exempt.go`）
- [x] 2.2 确认 `StripSpoofedInternalHeaders` 仍在豁免分支之前执行；**不得**豁免整个 `/app/api/push/` 前缀
- [x] 2.3 确认 device `register`/`unregister` 仍强制 Bearer

## 3. push-service 请求与落库

- [x] 3.1 `AppPushLiveActivityRegisterPostReq` 增加非零 `wxId`（json:`wxId`）校验
- [x] 3.2 `LiveActivityRegisterPost`：以 `req.WxId` 调用 `RegisterLiveActivityToken`；**不得**再以 `wxIDFromHeader` 作为唯一/回退身份（本变更不兼容旧客户端）
- [x] 3.3 回归：合法 body upsert `push_live_activity_token`；缺 `wxId` / 非法字段返回参数错误；仍 MUST NOT 写入 `push_device` uk_token

## 4. 验证

- [x] 4.1 无 Authorization + 合法 body（含 `wxId`）→ 200 且表中有对应行（代码路径：精确豁免 + body.wxId → RegisterLiveActivityToken；本地 `go build` 通过）
- [x] 4.2 无 `wxId`（即使带 Bearer）→ 拒绝（`v:"required|min:1"` + controller `WxId<=0` 挡；不读 header）
- [x] 4.3 无 Authorization 调 `/app/api/push/register` → 仍 401（仅 live-activity/register 入豁免表）
- [ ] 4.4 与 Flutter 真机联调：killed → Push-to-Start → 不打开 App → 服务端出现 `tokenKind=activity` 行
