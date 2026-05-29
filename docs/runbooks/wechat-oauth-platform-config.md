# 微信开放平台 OAuth 登录（device-service）部署指引

适用场景：胖宝 iOS/Android 移动应用与官网网页通过微信开放平台登录，网关 `POST /device/app/api/login` 入参为 **`jsCode`** + **`platform`**（字段名与 App 约定一致，勿改）。

## 配置位置

`device-service` 配置文件 `manifest/config/config.device-service.yaml`（或部署挂载覆盖）：

```yaml
wechat:
  platforms:
    ios:
      appId: "wx移动应用AppID"
      appSecret: "移动应用AppSecret"
    android:
      appId: "wx移动应用AppID"      # 与 ios 相同
      appSecret: "移动应用AppSecret"  # 与 ios 相同
    web:
      appId: "wx网站应用AppID"      # 独立申请
      appSecret: "网站应用AppSecret"
```

### platform 键含义

| platform | 客户端 | 微信应用类型 | 凭据 |
|----------|--------|--------------|------|
| `ios` | iOS App，`SendAuth` 取得 code | 移动应用 | 与 android 相同 |
| `android` | Android App，`SendAuth` 取得 code | 移动应用 | 与 ios 相同 |
| `web` | 浏览器 `qrconnect` 回调 code | 网站应用 | 独立 appId/secret |

**生产环境** MUST 通过环境变量或私有挂载覆盖 `appSecret`，禁止将真实密钥提交到 git。

## 换票语义

device-service 调用微信 `sns/oauth2/access_token`，将请求体 `jsCode` 作为 `code` 参数换 `openid`/`unionid`。身份键为 `wx.union_id`；微信返回的 access_token 不落库。

不再使用微信小程序 `jscode2session`。

## 网页端前置（web platform）

1. 在微信开放平台创建**网站应用**，配置授权回调域与 `redirect_uri`。
2. 前端跳转 `qrconnect` 取得 `code` 后，调用：

```http
POST /device/app/api/login
Content-Type: application/json

{
  "jsCode": "<回调 URL 中的 code>",
  "platform": "web"
}
```

回调页实现可与 `pangbao-official-site-homepage` 变更衔接；本服务端仅要求 `platform=web` 时 `wechat.platforms.web` 已配置。

## 联调验证

### 无微信路径

使用 `POST /device/app/api/device_login`（仅 `deviceNo`）验证 JWT 与下游，见 `gateway-app-integration-test.html`。

### 移动应用（ios / android）

1. 在测试环境为 `wechat.platforms.ios` 与 `android` 注入真实移动应用凭据（两键相同）。
2. 真机微信 SDK `SendAuth` 取得 `code`（一次性、短时效）。
3. 请求网关：

```http
POST /device/app/api/login
Content-Type: application/json

{
  "jsCode": "<SendAuth 返回的 code>",
  "platform": "ios"
}
```

4. 期望：`accessToken`、`refreshToken`、`wxId`；`unionid` 不返回客户端。

### 网站应用（web）

1. 配置 `wechat.platforms.web` 网站应用凭据。
2. 完成 `qrconnect` 扫码回调，从 URL 取 `code`。
3. `POST /device/app/api/login`，`platform` 为 `web`。
4. 期望与移动应用相同形态的 token 响应。

### 常见错误

| 现象 | 可能原因 |
|------|----------|
| 未配置微信凭据 | `platform` 键在 yaml 中缺失或 secret 为空 |
| errcode 40029 | code 无效、过期或已使用 |
| unionid 为空 | 应用未正确绑定微信开放平台账号 |

## 相关文档

- iOS Universal Links：`docs/runbooks/wechat-ios-universal-links.md`
- 网关部署：`docs/runbooks/release-deploy-and-run.md`
