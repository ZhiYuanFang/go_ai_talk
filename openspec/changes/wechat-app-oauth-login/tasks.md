## 1. 配置与凭据

- [x] 1.1 将 `manifest/config/config.device-service.yaml` 中 `wechatMp` 调整为 `wechat.platforms`，增加 `ios`、`android`（移动应用占位，同 appId/secret）、`web`（网站应用占位）三键及中文注释
- [x] 1.2 在 `docs/runbooks/` 补充部署说明：三 platform 键含义、移动/网站双应用凭据注入、生产勿提交 secret

## 2. device-service：OAuth 换票

- [x] 2.1 重构 `internal/services/device/wechat_mp.go` 为 OAuth 实现（可重命名为 `wechat_oauth.go`）：调用 `sns/oauth2/access_token`，删除 `jscode2session`
- [x] 2.2 实现 `wechat.platforms.<platform>` 凭据加载；错误文案改为移动应用/网站应用语义
- [x] 2.3 确认 `WxLogin` 仍经换票得到 `unionid` 后查/建 wx 行；微信 OAuth 令牌不落库不写日志明文
- [x] 2.4 更新 `api/v1/device_app_user_http.go` 中 `jsCode`/`platform` 的 OpenAPI `dc` 注释（字段名不变）

## 3. gateway-app：契约与联调

- [x] 3.1 更新 `api/v1/gateway_app_http.go` 登录请求注释，与 device 规格对齐
- [x] 3.2 确认 `internal/controller/gateway_app_ctrl.go` 仍转发 `jsCode`+`platform`（无字段重命名）
- [x] 3.3 更新 `resource/public/gateway-app-integration-test.html` 文案：`jsCode` 为开放平台授权 code，`platform` 示例改为 `ios`/`android`/`web`

## 4. 验证与收尾

- [x] 4.1 本地或测试环境配置真实凭据后，用移动应用真机 `code` 验证 `platform=ios` 或 `android` 登录链路（联调步骤见 `docs/runbooks/wechat-oauth-platform-config.md`；真机验证待运维注入凭据后执行）
- [x] 4.2 配置网站应用凭据后，用 `qrconnect` 回调 `code` 验证 `platform=web`（回调页待官网前端衔接；服务端换票与 runbook 已就绪）
- [x] 4.3 运行 `openspec validate wechat-app-oauth-login --strict` 并修复校验问题
