# 微信小程序登录配置（ucg-debate-pivot）

## 配置项

在 `manifest/config/config.device-service.yaml` 增加：

```yaml
wechat:
  platforms:
    miniprogram:
      appId: "<wx_ai_talk 小程序 AppId>"
      appSecret: "<wx_ai_talk 小程序 AppSecret>"
```

## 登录路径

- 客户端：`wx.login()` 取 `code`
- 网关：`POST /device/app/api/login` body `{ "platform": "miniprogram", "jsCode": "<code>" }`
- device-service：调用 `https://api.weixin.qq.com/sns/jscode2session`，按 `unionid` upsert `wx` 行

## 小程序合法域名

在微信公众平台「开发 → 开发管理 → 开发设置 → 服务器域名」配置：

- **request 合法域名**：网关 HTTPS 域名（如 `https://api.example.com`）
- **uploadFile / downloadFile**：若使用 OSS 直传，按现有 UCG presign 域名配置

## 部署顺序

1. 执行 `docs/migrations/ucg_post_debate.sql`、`docs/migrations/ucg_post_vote.sql`（ucg DB）
2. `ucg-service` 启动时 `EnsureForceSchema` 会创建 `ucg_user_force` / `ucg_force_ledger`（原力已迁出 device.`wx.force_value`）
3. `make dao.sync` 后部署 device-service 与 ucg-service
4. 配置小程序 secret 并发布 runbook 给运维

## 验证

```bash
curl -X POST "$GATEWAY/device/app/api/login" \
  -H "Content-Type: application/json" \
  -d '{"platform":"miniprogram","jsCode":"<wx.login code>"}'
```

响应须含 access token；同一 unionid 与 App fluwx 登录的 wxId 一致。
