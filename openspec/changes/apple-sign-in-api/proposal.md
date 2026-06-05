## Why

iOS App Store 审核指南 4.8 要求：当应用提供第三方登录（本应用已在 iOS 提供微信登录）时，必须同时提供「通过 Apple 登录」。Flutter 客户端变更 `add-apple-sign-in`（`flutter_ai_talk`）的 Phase 2 依赖本仓库先交付 Phase 1 后端能力：校验 Apple `identityToken`、按 `sub` 查找或创建 `wx` 行、经 gateway-app 签发与微信/用户名登录相同形态的 JWT 会话；并支持已登录用户**可选**将第二登录方式绑定到**当前** `wx` 行。

## What Changes

- 在 `wx` 表新增 `apple_sub` 列（VARCHAR，唯一索引），**仅**持久化 Apple JWT 的 `sub`；**不得**存储 Apple 邮箱或姓名。
- 在 device-service 新增 `internal/services/device/apple_auth.go`：拉取 Apple JWKS，校验 `identityToken`（`iss`、`aud`=`com.fzy.pangbao`、`exp` 等）；可选接受 `authorizationCode` 并在 design 中说明用途（本变更以 identityToken 校验为主路径）。
- 在 device-service 新增 `WxAppleLogin` 业务：按 `apple_sub` 查找或创建 `wx` 行，返回 `wxId`、`deviceNo`、`isNewUser`（不含 JWT）。
- 新增 device 内部接口 `POST /device/app/api/user/apple/login` 与 gateway 聚合接口 `POST /device/app/api/apple_login`；JWT 签发逻辑与现有 `Login` / `UsernameLogin`（`gateway_app_ctrl.go`）对齐。
- 新增 **Bearer** 绑定接口：
  - `POST /device/app/api/user/apple/bind` — 将 Apple 身份绑定到**当前**已登录 `wx` 行（校验 `identityToken`，UPDATE 当前行 `apple_sub`）。
  - `POST /device/app/api/user/wx/bindwx` — 将微信身份绑定到**当前**已登录 `wx` 行（`jsCode` 换 `unionid`，不限于用户名账号）；可泛化自既有 `username/bindwx` 业务逻辑。
- 绑定错误码：`ErrAppleSubTakenByOther`、`ErrUnionIDTakenByOther`、`ErrAccountMergeConflict`（尝试绑定但标识符已属于另一完整账号 / 不可合并场景）。
- 绑定算法：向**当前** `wx` 行 UPDATE 第二标识符；若 `apple_sub` 或 `unionid` 已存在于**不同** `wxId`，拒绝绑定；**不得**合并两条已独立存在的完整 `wx` 行。
- Profile（可选最小扩展）：`isAppleBound`（`apple_sub` 非空）、`authProviders`（如 `["apple"]`、`["wechat"]`、`["apple","wechat"]`）。
- 在 `api/v1/` 补充 Gateway / Device 请求响应类型定义。
- 将 `apple_login` 与 `user/apple/login` 加入 gateway-app Bearer 鉴权白名单（`gateway_app_auth_exempt.go`）；bind 端点**须** Bearer 鉴权（非白名单）。
- 在 `manifest/config/config.device-service.yaml` 增加 `apple.ios.bundleId: com.fzy.pangbao` 配置项。
- 更新 `resource/public/gateway-app-integration-test.html` 联调页，增加 Apple 登录与绑定测试区块。
- 在 tasks / runbook 中记录 DDL 迁移与部署顺序说明。

**明确不在本变更范围**：

- **合并两条已独立存在的完整账号**：用户曾分别以 Apple、微信各登录并各产生独立 `wx` 行（两条 `wxId` 各含对应标识）后，**不得**再通过绑定将两行合并为一行。

## Capabilities

### New Capabilities

- `apple-sign-in-api`：定义 Apple JWT 校验、`wx.apple_sub` 持久化、device 业务登录与 gateway 聚合 `apple_login` 契约、可选账号绑定（`apple/bind`、`wx/bindwx`）、绑定冲突错误码、profile 绑定状态字段、鉴权白名单、配置项及联调页要求。

### Modified Capabilities

- `device-wx-profile-apis`（v1.0.3 基线）：**最小扩展** profile 读模型以暴露 `isAppleBound` 与 `authProviders`（若实现判定必要）；`username/bindwx` 既有行为保留，新增泛化 `wx/bindwx` 供 Apple-only 等账号绑定微信。

## Impact

- **device-service**：`wx` 表 DDL、`entity`/`dao` 生成、`apple_auth.go`、`wx.go`（或等价业务文件）、绑定业务（`WxBindApple`、`WxBindWxByCode` 或等价）、`device_app_user.go` 控制器、`api/v1/device_app_user_http.go`、profile 读模型。
- **gateway-app-server**：`gateway_app_ctrl.go`、`api/v1/gateway_app_http.go`、`gateway_app_auth_exempt.go`。
- **配置**：`manifest/config/config.device-service.yaml`（`apple.ios.bundleId`）。
- **联调与文档**：`resource/public/gateway-app-integration-test.html`；`docs/runbooks/release-deploy-and-run.md` 或等价 runbook 补充迁移说明（若 tasks 判定需同步）。
- **跨仓库依赖**：`flutter_ai_talk` `add-apple-sign-in` Phase 2 联调前须本变更已部署至目标环境。
- **数据库**：`device-service` 默认库 `database.default`（`DEVICE_DB_LINK`），仅 `wx` 表增列；不涉及跨库直查。
