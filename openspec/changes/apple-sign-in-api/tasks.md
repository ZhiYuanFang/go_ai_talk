## 1. 数据库迁移（device-service / `database.default`）

- [x] 1.1 编写并评审 DDL：`ALTER TABLE wx ADD COLUMN apple_sub VARCHAR(255) NULL COMMENT 'Apple JWT sub'`；`CREATE UNIQUE INDEX uk_wx_apple_sub ON wx (apple_sub)`
- [x] 1.2 在目标环境（staging → prod）按 runbook 顺序执行 DDL，确认既有微信/用户名行 `apple_sub` 均为 NULL 且无冲突
- [x] 1.3 运行 `gf gen dao`（或等价流程）同步 `internal/model/entity/wx.go` 与 `internal/dao` 列定义

## 2. Apple JWT 校验（device-service）

- [x] 2.1 新增 `internal/services/device/apple_auth.go`：拉取 `https://appleid.apple.com/auth/keys`（JWKS 内存缓存 + TTL）、校验 `identityToken` 签名、`iss`、`aud`（读配置 `apple.ios.bundleId`）、`exp`；不向日志写入 token 原文
- [x] 2.2 在 `manifest/config/config.device-service.yaml` 增加 `apple.ios.bundleId: com.fzy.pangbao` 及中文注释说明用途
- [x] 2.3 实现 `WxAppleLogin`（`wx.go` 或 `apple_login.go`）：`verifyAppleIdentityToken` → 按 `apple_sub` 查/建 `wx` 行，返回 `WxLoginResult`；**不**存储 email/name

## 3. API 类型定义（`api/v1/`）

- [x] 3.1 在 `api/v1/device_app_user_http.go` 增加 `DeviceAppleLoginReq`（`POST /device/app/api/user/apple/login`，字段 `identityToken`、`platform`、可选 `authorizationCode`）与 `DeviceAppleLoginRes`（`wxId`、`deviceNo`、`isNewUser`）
- [x] 3.2 在 `api/v1/gateway_app_http.go` 增加 `GatewayAppAppleLoginReq`（`POST /device/app/api/apple_login`）与 `GatewayAppAppleLoginRes`（与 `GatewayAppLoginRes` 字段对齐）
- [x] 3.3 增加 `DeviceAppleBindReq`（`POST /device/app/api/user/apple/bind`）与 `DeviceWxBindWxReq`（`POST /device/app/api/user/wx/bindwx`）类型定义

## 4. 控制器与网关聚合

- [x] 4.1 在 `internal/controller/device_app_user.go` 增加 `AppleLogin` 处理器，调用 `device.WxAppleLogin`
- [x] 4.2 在 `internal/controller/gateway_app_ctrl.go` 增加 `AppleLogin`：转发 device `user/apple/login` 后 `SignAccess` + `IssueRefreshToken`（对齐 `UsernameLogin` 逻辑）
- [x] 4.3 确认 GoFrame 路由注册已绑定新 API 类型（`cmd/device-service`、`cmd/gateway-app-server` 启动无未注册路由）

## 5. 账号绑定（device-service）

- [x] 5.1 实现 `WxBindApple`（或等价）：Bearer 当前 `wxId`；校验 `identityToken` 得 `sub`；当前行 `apple_sub` 为空；`sub` 未占用其他 `wxId`；UPDATE 当前行；冲突返回 `ErrAppleSubTakenByOther`
- [x] 5.2 实现 `WxBindWxByCode`（或等价）：抽取/泛化 `WxUsernameBindWxByCode`，新增 `POST /device/app/api/user/wx/bindwx`；**去掉**必须已设用户名密码限制；`unionid` 冲突返回 `ErrUnionIDTakenByOther`
- [x] 5.3 定义并实现 `ErrAccountMergeConflict`（或文档化与 TakenByOther 的映射）：当绑定请求等价于合并两条已独立完整 `wx` 行时拒绝
- [x] 5.4 在 `device_app_user.go` 增加 `AppleBind`、`WxBindWx` 处理器；确认 bind 路由**不在** gateway 匿名白名单

## 6. Profile 绑定状态（可选最小）

- [x] 6.1 在 `WxUserProfile` 与 profile API 响应增加 `isAppleBound`（`apple_sub` 非空）与 `authProviders`（按行派生，如 `apple`/`wechat`/`username`）
- [x] 6.2 确认不改变 `unionid`/`password` 等敏感字段暴露；与 v1.0.3 `isWxBound` 风格一致

## 7. 鉴权白名单

- [x] 7.1 在 `internal/controller/gateway_app_auth_exempt.go` 的 `gatewayAppAuthExemptExactPOST` 增加 `/device/app/api/apple_login` 与 `/device/app/api/user/apple/login`（**不含** bind 路径）

## 8. 联调页与文档

- [x] 8.1 更新 `resource/public/gateway-app-integration-test.html`：新增 Apple 登录区块（`apple_login`）与绑定区块（`apple/bind`、`wx/bindwx`，需 Bearer）
- [x] 8.2 在 `docs/runbooks/release-deploy-and-run.md`（或等价 runbook）补充：`wx.apple_sub` 迁移步骤、部署顺序（先 DDL → device-service → gateway-app）、`apple.ios.bundleId` 配置核对项
- [x] 8.3 核对 `manifest/docker/.env.example` 与 `config.device-service.yaml` 顶部注释：本变更不新增 `*_DB_LINK`，但 runbook 须标明迁移归属 `DEVICE_DB_LINK` 对应库

## 9. 验证与跨仓库联调

- [x] 9.1 staging 联调页：有效 `identityToken` 登录成功并拿到 JWT；无效/过期 token 返回业务错误
- [x] 9.2 绑定：Apple-only 行绑定微信成功；WeChat-only 行绑定 Apple 成功；标识符已被其他 `wxId` 占用时返回 `ErrAppleSubTakenByOther` / `ErrUnionIDTakenByOther`
- [x] 9.3 确认**不**支持合并两条已独立完整账号（各 `wxId` 已含对应标识）；尝试绑定时返回明确冲突错误
- [x] 9.4 通知 `flutter_ai_talk` `add-apple-sign-in` Phase 2：staging `apple_login` 与 bind API 已就绪，可开始客户端联调
