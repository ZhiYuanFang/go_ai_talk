## Context

当前 App 登录链路：

- **微信**：客户端 `POST /device/app/api/login` → gateway-app 转发 `POST /device/app/api/user/login` → `WxLogin` 经微信 OAuth 换 `unionid` 查/建 `wx` 行 → gateway 签发 `accessToken`/`refreshToken`（`gateway_app_ctrl.go` `Login`）。
- **用户名**：`POST /device/app/api/username_login` → `POST /device/app/api/user/username/login` → `WxUsernameLogin` → 同上 JWT 签发（`UsernameLogin`）。
- **设备号**：`POST /device/app/api/device_login` → 业务校验后签发（`wxId` 可为 0）。

`wx` 表（`internal/model/entity/wx.go`）现有字段：`id`、`device_no`、`unionid`、`platform`、`account`、`password`；无 Apple 标识列。Flutter 客户端变更 `add-apple-sign-in` Phase 2 将向 gateway 提交 Apple `identityToken`（JWT），服务端须仅信任 JWT 内 `sub` 作为稳定身份键。

既有 `POST /device/app/api/user/username/bindwx` 仅允许**已设用户名密码**的 `wx` 行绑定微信（`WxUsernameBindWxByCode`）。产品要求泛化绑定：Apple-only 用户也可绑定微信；WeChat-only 用户也可绑定 Apple。

产品决策（已确认）：

- 用户**可以**在已登录状态下将第二登录方式绑定到**当前** `wx` 行（可选，非强制）。
- 用户**可以**仅使用单一登录方式，不绑定第二方式。
- **限制**：若 Apple 与微信登录曾各创建**独立** `wx` 行（两条 `wxId` 各含 `apple_sub` / `unionid`），**不得**事后合并；绑定仅当向当前行添加尚未被其他行占用的第二标识符。
- Apple 邮箱**不**存储，仅 `sub`。

## Goals / Non-Goals

**Goals:**

- 新增 `wx.apple_sub` 列（唯一索引），仅持久化 Apple JWT `sub`。
- 实现 `apple_auth.go`：Apple JWKS 校验 `identityToken`（`iss`、`aud`=`com.fzy.pangbao`、`exp`）。
- 实现 `WxAppleLogin`：按 `apple_sub` 查/建 `wx` 行，返回 `wxId`、`deviceNo`、`isNewUser`（不含 JWT）。
- 暴露 `POST /device/app/api/user/apple/login`（device 业务）与 `POST /device/app/api/apple_login`（gateway 聚合 + JWT），响应形态与 `Login`/`UsernameLogin` 一致。
- 实现 Bearer 绑定：`POST /device/app/api/user/apple/bind`、`POST /device/app/api/user/wx/bindwx`。
- 绑定冲突错误码：`ErrAppleSubTakenByOther`、`ErrUnionIDTakenByOther`、`ErrAccountMergeConflict`。
- Profile 最小扩展：`isAppleBound`、`authProviders`。
- 白名单、API 类型定义、`apple.ios.bundleId` 配置、联调页更新、DDL/runbook 说明。

**Non-Goals:**

- **合并两条已独立存在的完整 `wx` 行**（各 `wxId` 已分别通过 Apple / 微信登录创建且含对应标识）。
- 存储 Apple 邮箱、姓名或 `authorizationCode` 换票作为主路径（`authorizationCode` 仅作请求体可选字段文档化，本变更不强制服务端用其换 token）。
- 新建 `*_test.go` 或自动化测试文件（遵循项目约定）。

## Decisions

1. **身份键：仅 `apple_sub`**
   - 校验通过后提取 JWT `sub`，作为 `wx.apple_sub` 唯一查找键；与微信 `unionid` 可共存于同一 `wx` 行（绑定后）。
   - **不**解析或持久化 `email`、`name` claim（即使 Apple 首次授权附带）。
   - 备选：以 email 作辅助标识 — 拒绝（隐私与产品决策）。

2. **JWT 校验：`apple_auth.go`**
   - 从 `https://appleid.apple.com/auth/keys` 拉取 JWKS（带内存缓存与 TTL，避免每次登录请求 Apple）。
   - 校验项：`iss` 为 `https://appleid.apple.com`；`aud` 等于配置 `apple.ios.bundleId`（`com.fzy.pangbao`）；`exp` 未过期；签名算法与 kid 匹配。
   - 使用 Go 标准库 + 成熟 JWT 解析（如 `github.com/golang-jwt/jwt/v5` 或项目已有依赖），不将 identityToken 原文写入日志。
   - 备选：仅 base64 解码 payload 不验签 — 拒绝（安全要求）。

3. **业务函数：`WxAppleLogin`（`wx.go` 或独立 `apple_login.go`）**
   - 入参：`identityToken`（必填）、`platform`（客户端传 `ios`，写入 `wx.platform`）、可选 `authorizationCode`（本变更忽略，仅 API 层接受以便将来收紧校验）。
   - 流程：`verifyAppleIdentityToken` → 得 `sub` → `WHERE apple_sub = ?` → 无则 `INSERT`（`apple_sub`、`platform`）→ 返回 `WxLoginResult`。
   - 新用户 `isNewUser=true`、`deviceNo=""`；老用户返回已绑定 `deviceNo`（若有）。
   - 与 `WxLogin` 相同：响应不含 JWT、不含 `apple_sub` 明文回传客户端（客户端已持有 token）。

4. **API 分层（镜像微信/用户名模式）**
   - **Device** `DeviceAppleLoginReq/Res`：`POST /device/app/api/user/apple/login`，body `{ identityToken, platform, authorizationCode? }`。
   - **Gateway** `GatewayAppAppleLoginReq/Res`：`POST /device/app/api/apple_login`，转发 device 后 `SignAccess` + `IssueRefreshToken`（复制 `UsernameLogin` 控制器逻辑）。
   - JSON 字段 lowerCamelCase，与现有 `api/v1` 一致。

5. **绑定 API（Bearer，wx 主键来自 `X-Internal-Wx-Id`）**
   - **`POST /device/app/api/user/apple/bind`**：body `{ identityToken, platform }`；校验 token 得 `sub`；当前行 `apple_sub` 须为空；`sub` 不得已属于其他 `wxId`；成功则 UPDATE 当前行 `apple_sub`。
   - **`POST /device/app/api/user/wx/bindwx`**：body `{ jsCode, platform }`；`jsCode` 换 `unionid`；当前行 `unionid` 须为空；`unionid` 不得已属于其他 `wxId`；成功则 UPDATE 当前行 `unionid`（可复用/抽取 `WxUsernameBindWxByCode` 核心逻辑，**去掉**「必须已设用户名密码」限制）。
   - 既有 `POST /device/app/api/user/username/bindwx` **保留**（用户名账号专用路径）；新 `wx/bindwx` 为泛化入口，供 Apple-only 等账号使用。
   - bind 端点**不**加入 gateway 匿名白名单；须有效 Bearer access token。

6. **绑定算法与冲突语义**
   - 仅 UPDATE **当前**已登录 `wx` 行，添加第二标识符；**永不** DELETE/合并另一 `wx` 行或迁移其数据。
   - 若目标标识符已存在于 `wxId' != currentWxId` 的行 → `ErrAppleSubTakenByOther` 或 `ErrUnionIDTakenByOther`。
   - 若当前行已含目标标识符（重复绑定）→ 幂等成功或返回已绑定业务码（实现时选与 `username/bindwx` 一致语义）。
   - 若用户曾分别以两种方式登录产生两条完整行，后登录行尝试绑定另一行已有标识 → 命中 TakenByOther；客户端可映射为「无法合并两个已独立创建的账号」；可选用 `ErrAccountMergeConflict` 作为对外统一码（当两标识分属两行且均非空时）。

7. **Profile 扩展（最小）**
   - `isAppleBound`：`apple_sub` 非空为 `true`。
   - `authProviders`：按行非空字段派生，如 `unionid` → `wechat`，`apple_sub` → `apple`，`account`+`password` → `username`（顺序与命名在实现中与 v1.0.3 `isWxBound` 风格对齐）。
   - 不改变 `unionid`/`account` 等敏感字段暴露策略。

8. **鉴权白名单**
   - 在 `gateway_app_auth_exempt.go` 的 `gatewayAppAuthExemptExactPOST` 增加：
     - `/device/app/api/apple_login`
     - `/device/app/api/user/apple/login`
   - 与 `login`、`username_login` 同级，匿名可调用。
   - `apple/bind`、`wx/bindwx` **不在**白名单。

9. **配置**
   - `manifest/config/config.device-service.yaml`：
     ```yaml
     apple:
       ios:
         bundleId: "com.fzy.pangbao"
     ```
   - 生产可通过环境变量或 overlay 覆盖；`aud` 校验读取此配置。

10. **DDL**
    - `ALTER TABLE wx ADD COLUMN apple_sub VARCHAR(255) NULL COMMENT 'Apple JWT sub' AFTER unionid;`
    - `CREATE UNIQUE INDEX uk_wx_apple_sub ON wx (apple_sub);`（MySQL 允许多个 NULL，不冲突微信/用户名行）。
    - 通过 `gf gen dao` 或手工同步 `entity`/`dao` 列定义；tasks 记录执行顺序（先 DDL 再部署代码）。

11. **联调页**
    - `gateway-app-integration-test.html` 增加 Apple 登录区块与绑定区块（需先登录拿 Bearer）：`apple/bind`、`wx/bindwx`。

## Risks / Trade-offs

- **[Risk] JWKS 拉取失败或 Apple 服务不可用** → 登录/绑定返回明确业务错误；JWKS 内存缓存 + 合理 TTL；日志记录 kid/HTTP 状态，不记录 token 原文。
- **[Risk] `aud` 与客户端 Bundle ID 不一致** → 配置集中 `apple.ios.bundleId`；staging 联调前核对 `com.fzy.pangbao`。
- **[Risk] 用户曾独立创建双账号后尝试绑定** → `ErrAppleSubTakenByOther` / `ErrUnionIDTakenByOther` / `ErrAccountMergeConflict`；客户端展示明确不可合并文案。
- **[Risk] DDL 与代码部署顺序** → runbook/tasks 要求先执行迁移再滚动 device-service；回滚时保留列无害。
- **[Trade-off] 不复用 `unionid` 列存 Apple sub** → 语义清晰、避免与微信冲突；多一列但查询简单。
- **[Trade-off] 保留 `username/bindwx` 与新增 `wx/bindwx` 双路径** → 避免破坏既有用户名绑定契约；新路径服务 Apple-only 等场景。

## Migration Plan

1. **DDL**（维护窗口或低峰）：在 device 默认库对 `wx` 表执行 `apple_sub` 增列与唯一索引。
2. **配置**：各环境 `config.device-service.yaml`（或 Secret）写入 `apple.ios.bundleId`。
3. **部署**：先 `device-service`，再 `gateway-app-server`（新路由与白名单）。
4. **验证**：联调页 + Flutter `add-apple-sign-in` staging 真机 `identityToken` 联调；绑定成功/冲突路径。
5. **回滚**：下线新路由（旧客户端无 Apple 入口不受影响）；`apple_sub` 列可保留；已创建 Apple `wx` 行保留直至用户注销。

## Open Questions

- **`authorizationCode` 是否在未来用于 server-to-server 二次校验** — 本变更 API 接受可选字段但不处理；若 App Store 审核或安全评审要求，可在后续 change 增加换票校验而不破坏客户端契约。
- **JWKS 缓存 TTL** — 建议 24h 与 Apple 密钥轮换频率对齐；实现时在 `apple_auth.go` 中文注释说明刷新策略。
- **`ErrAccountMergeConflict` 与 TakenByOther 的对外映射** — 实现时二选一或并用：spec 允许 TakenByOther 覆盖合并场景，MergeConflict 为可选语义化包装。
