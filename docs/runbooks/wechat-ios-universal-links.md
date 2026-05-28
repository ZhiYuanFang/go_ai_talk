# 微信开放平台 iOS Universal Links 接入指引

适用场景：胖宝 iPhone 应用（Bundle ID：`com.fzy.pangbao`）在**微信开放平台**创建/更新 iOS 应用配置，并通过 **GitHub Actions** 而非本地 Xcode 手工打包发布。

## 1. 目标值与关键约束

- 对外站点基址：`https://www.pangbao.cuplay.top/`
- 微信开放平台 `Universal Links`：`https://www.pangbao.cuplay.top/wx/ulink/`
- Apple `Associated Domains`：`applinks:www.pangbao.cuplay.top`
- AASA 文件访问地址（两条都要可用）：
  - `https://www.pangbao.cuplay.top/apple-app-site-association`
  - `https://www.pangbao.cuplay.top/.well-known/apple-app-site-association`

注意：AASA 必须从 **host 根路径** 获取，不能只依赖某个业务子路径暴露。即使后续站点再挂载到其它前缀，Apple 仍只会读取根路径上的 `apple-app-site-association`。

## 2. 服务端配置

`gateway-app-server` 已约定输出 AASA 内容，部署时补齐以下配置：

- 环境变量：`GATEWAY_APP_IOS_TEAM_ID=395D9NUCNF`
- 可选配置回退：`manifest/config/config.gateway-app-server.yaml` 中 `gatewayApp.ios.teamId`
- 官网基址：`GATEWAY_APP_PUBLIC_BASE_URL=https://www.pangbao.cuplay.top`

注意：你最新提供的是 `http://www.pangbao.cuplay.top/`，但 Apple Universal Links 与 AASA 校验 **必须使用 HTTPS**；HTTP 只能用于普通网页访问，不能作为微信开放平台或 iOS `universalLink` 的最终配置值。

当前 Apple Team ID 已确认为 `395D9NUCNF`。当 `GATEWAY_APP_IOS_TEAM_ID` 未配置时，AASA 端点会返回 `503` 与显式错误 JSON，用于提醒当前环境尚未完成 Universal Links 接入。

## 3. AASA 内容约束

服务端生成的 AASA 关键结构如下：

```json
{
  "applinks": {
    "details": [
      {
        "appIDs": ["395D9NUCNF.com.fzy.pangbao"],
        "components": [
          {
            "/": "/wx/ulink/*"
          }
        ]
      }
    ]
  }
}
```

其中：

- `com.fzy.pangbao` 为固定 Bundle ID
- Apple Developer 后台中的真实 Team ID 为 `395D9NUCNF`
- `/wx/ulink/*` 必须与微信开放平台填写值、iOS 工程 `universalLink` 配置保持一致

## 4. 边缘代理 / Nginx 透出规则

如果公网入口已经直接挂在 `https://www.pangbao.cuplay.top/` 根路径下，可参考如下 Nginx 规则：

```nginx
location = /apple-app-site-association {
    proxy_pass http://127.0.0.1:9702/apple-app-site-association;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto https;
}

location = /.well-known/apple-app-site-association {
    proxy_pass http://127.0.0.1:9702/.well-known/apple-app-site-association;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto https;
}

location / {
    proxy_pass http://127.0.0.1:9702/;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto https;
}
```

部署要求：

- AASA 请求必须走 HTTPS
- AASA 请求不能 301/302 到其它路径
- 返回头必须包含 `Content-Type: application/json`
- 不要把 AASA 当作附件下载

## 5. 微信开放平台填写

在微信开放平台创建/编辑 iPhone 应用时：

- `Bundle ID`：`com.fzy.pangbao`
- `Universal Links`：`https://www.pangbao.cuplay.top/wx/ulink/`

注意事项：

- 必须以 `/` 结尾
- 不能写成 `http://www.pangbao.cuplay.top/wx/ulink/`（HTTP 不满足 Apple 校验要求）
- 不能写成 `https://www.pangbao.cuplay.top`（缺少业务前缀与尾部 `/`）

## 6. GitHub Actions 打包要求

虽然不是在本地 Xcode 界面打包，但 **GitHub Actions 最终仍调用 Xcode 工具链签名与归档**，所以 Universal Links 是否生效，仍取决于 iOS 工程的 entitlements 和描述文件能力。

至少需要满足以下条件：

1. iOS 工程启用 `Associated Domains`
2. entitlements 中包含：

```xml
<key>com.apple.developer.associated-domains</key>
<array>
    <string>applinks:www.pangbao.cuplay.top</string>
</array>
```

3. Provisioning Profile 对应 `com.fzy.pangbao`，并已启用 `Associated Domains`
4. 微信 SDK 注册时传入的 `universalLink` 必须是：`https://www.pangbao.cuplay.top/wx/ulink/`

GitHub Actions 常见 Secrets / 变量建议：

- `APPLE_TEAM_ID=395D9NUCNF`
- `IOS_BUNDLE_ID=com.fzy.pangbao`
- `WECHAT_UNIVERSAL_LINK=https://www.pangbao.cuplay.top/wx/ulink/`
- `BUILD_CERTIFICATE_BASE64`
- `P12_PASSWORD`
- `BUILD_PROVISION_PROFILE_BASE64`
- `KEYCHAIN_PASSWORD`
- 如需上传 App Store Connect，再补：`APPSTORE_ISSUER_ID`、`APPSTORE_KEY_ID`、`APPSTORE_PRIVATE_KEY`

无论你使用 `xcodebuild`、`fastlane gym` 还是其它 GitHub 封装脚本，都要确保最终归档产物里带着正确的 entitlements。

## 7. 发布前验证

### 7.1 校验 AASA 端点

```bash
curl -i https://www.pangbao.cuplay.top/apple-app-site-association
curl -i https://www.pangbao.cuplay.top/.well-known/apple-app-site-association
```

期望结果：

- 返回 `200 OK`
- 无 `Location:` 跳转到其它路径
- `Content-Type` 为 `application/json`
- 两个地址返回内容一致

### 7.2 校验 GitHub 打包结果

在 GitHub Actions 构建完成后，至少检查：

- 使用的 Bundle ID 是 `com.fzy.pangbao`
- 归档使用的 Team ID 与 AASA 中一致
- entitlements 中存在 `applinks:www.pangbao.cuplay.top`
- 使用的微信 SDK `universalLink` 为 `https://www.pangbao.cuplay.top/wx/ulink/`

### 7.3 真机联调

当 Team ID 已补齐、AASA 已对外可读、构建包已重新签名后：

1. 重新安装测试包到真机
2. 在微信开放平台保存最新 Universal Links 配置
3. 触发微信登录/分享等需要回跳 iOS App 的链路
4. 确认系统优先拉起 App，而不是停留在浏览器

## 8. 故障排查

### 症状：微信后台提示 Universal Links 校验失败

优先检查：

- Team ID 是否已经补齐
- AASA 是否错误地放到了某个业务子路径下
- 反向代理是否把根路径请求改写到了其它路径
- 证书链是否完整、HTTPS 是否可信

### 症状：GitHub Actions 打包成功，但真机无法拉起 App

优先检查：

- Provisioning Profile 是否包含 `Associated Domains`
- 归档产物中的 entitlements 是否真的含有 `applinks:www.pangbao.cuplay.top`
- 微信 SDK `universalLink` 是否与 AASA / 微信后台填写值完全一致（`https://www.pangbao.cuplay.top/wx/ulink/`）
- 安装包是否为更新后的重新签名版本（iOS 对 AASA 与签名缓存较敏感）