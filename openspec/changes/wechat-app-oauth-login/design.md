## Context

胖宝 App 网关与 device-service 已实现微信登录骨架：`POST /device/app/api/login`（网关聚合）→ `POST /device/app/api/user/login`（device 换票 + wx 落库）→ 网关签发 JWT。当前换票实现调用微信小程序专用 API `sns/jscode2session`，配置键为 `wechatMp.platforms`。

产品实际为：
- **iOS / Android**：微信开放平台「移动应用」，SDK `SendAuth` 取得临时 `code`；
- **网页**：微信开放平台「网站应用」，`qrconnect` 扫码回调取得 `code`；
- iOS 与 Android 共用**同一移动应用** `appId`/`appSecret`，但客户端分别传入 `platform=ios` 与 `platform=android`；
- 网页使用**独立**网站应用凭据，客户端传入 `platform=web`。

客户端已约定请求体字段 **`jsCode`** 与 **`platform`**，本变更**不得**修改 JSON 字段名或登录路径。Universal Links（`https://www.pangbao.cuplay.top/wx/ulink/`）由既有变更覆盖，与本换票改造正交。

## Goals / Non-Goals

**Goals:**

- 将 device-service 换票统一为微信开放平台 `sns/oauth2/access_token`，`jsCode` 承载授权临时 `code`。
- 配置 `wechat.platforms` 支持 `ios`、`android`（同凭据）、`web`（网站应用凭据）。
- 保持 `wx.union_id` 身份模型、登录响应字段、网关 JWT/refresh 行为不变。
- 更新 OpenAPI 注释、联调页文案与部署说明。

**Non-Goals:**

- 不实现微信小程序 `wx.login` / `jscode2session`（产品确认不需要）。
- 不新增平行登录 HTTP 路径；不修改 App 已发布的 `jsCode`/`platform` 字段名。
- 不在本变更内实现官网 `qrconnect` 前端页面或 OAuth 回调路由（仅文档约定；可与 `pangbao-official-site-homepage` 后续衔接）。
- 不实现微信公众号内 H5 网页授权（`snsapi_base`/`snsapi_userinfo`）。
- 不持久化微信 OAuth 返回的 `access_token`/`refresh_token`（仅用于换取 `unionid` 后丢弃）。

## Decisions

### D1：换票 API 选型

**Decision**：删除 `jscode2session`，统一使用 `GET https://api.weixin.qq.com/sns/oauth2/access_token`，参数 `appid`、`secret`、`code`（来自请求体 `jsCode`）、`grant_type=authorization_code`。

**Why**：移动应用与网站应用在开放平台文档中均通过该接口用 `code` 换 `openid`/`unionid`；与客户端已传的 `code` 语义一致。

**Alternatives**：
- 保留双路径（mp + oauth）：增加分支与测试面，产品已明确不要小程序。
- 客户端直传 unionid：违反现有安全规格，否决。

### D2：配置结构与 platform 映射

**Decision**：`manifest/config/config.device-service.yaml` 使用 `wechat.platforms.<platform>.appId|appSecret`；必须配置 `ios`、`android`、`web` 三键。`ios` 与 `android` 的 `appId`/`appSecret` 在部署文档中约定为**相同移动应用凭据**；`web` 为网站应用凭据。

**Why**：`platform` 以客户端传入为准，便于 `wx.platform` 区分首登来源；凭据按微信开放平台应用类型分离。

**Alternatives**：
- 继续 `wechatMp` 键名：名称误导，实现时重命名为 `wechat`。
- 单键 `pangbao` 忽略客户端 platform：与用户「ios/android 分别传入」冲突，否决。

### D3：HTTP 契约冻结

**Decision**：`POST /device/app/api/login` 与 `POST /device/app/api/user/login` 的请求/响应 JSON 字段名不变；仅服务端换票与 OpenAPI `dc` 注释更新。

**Why**：避免 iOS/Android 客户端发版；网页端复用同一契约。

### D4：代码组织

**Decision**：将 `internal/services/device/wechat_mp.go` 重构/重命名为 `wechat_oauth.go`（或等价），导出 `exchangeAuthCodeForUnionID(ctx, platform, code)`；`WxLogin` 签名可保留 `(jsCode, platform)` 参数名以减少调用链改动。

**Why**：文件名与小程序脱钩，逻辑单一。

### D5：微信 OAuth 令牌处理

**Decision**：从 `oauth2/access_token` 响应解析 `openid`、`unionid`；微信侧 `access_token`/`refresh_token` 不落库、不写日志明文。

**Why**：登录仅需 `unionid`；与现网丢弃 `session_key` 一致。

### D6：错误语义

**Decision**：`unionid` 为空、凭据未配置、`errcode` 非 0（如 40029 code 无效）时返回业务错误；日志可记录 `platform` 与 `errcode`，不向客户端透传微信 `errmsg` 全文。

## Risks / Trade-offs

- **[Risk] 仅配置移动应用、未配置网站应用** → `platform=web` 登录失败。  
  **Mitigation**：配置模板与 runbook 明确要求三键；启动或集成测试可对 `ios`/`android` 凭据一致性做可选校验。

- **[Risk] 运维将 ios/android 配成不同 secret** → 一端正常一端失败。  
  **Mitigation**：文档强调双键同凭据；可选启动日志 warning（非阻断）。

- **[Risk] 网页回调页未实现即期望扫码登录可用** → 用户无法拿到 `code`。  
  **Mitigation**：本变更交付服务端；网页前端列为独立任务或官网变更后续项。

- **[Risk] 行为变更导致旧联调依赖小程序 code** → 联调失败。  
  **Mitigation**：联调页改文案；保留 `device_login` 无微信路径；产品已确认不要小程序。

- **[Trade-off] `wx.platform` 仅在首登写入** → 用户换端后库内 platform 不更新。  
  **Acceptable**：身份以 `union_id` 为准；与现网一致。

## Migration Plan

1. 在 device-service 配置注入移动应用与网站应用凭据（`ios`/`android`/`web`）。
2. 部署含 oauth 换票的 device-service。
3. 部署 gateway-app（若仅有注释变更可同版本）。
4. 真机验证 iOS/Android `SendAuth` → `/login`；配置就绪后验证 `platform=web`。
5. **回滚**：回退 device-service 至 `jscode2session` 版本（仅当仍需小程序时；产品侧不计划）。

## Open Questions

- 官网 `qrconnect` 回调路径最终 URL（如 `/wx/callback`）由官网前端变更确定，本变更不强制网关新增路由。
- 是否在启动时对 `ios`/`android` appId 不一致打 warning：实现阶段可选。
