## 1. 校验文件与根路径路由

- [x] 1.1 新增 `resource/public/90fafbe9bf8308ecbd2063d7b479a309.txt`，内容精确为 `41a8fe99b666f6646724f29ea87962e34c6df2b6`（纯文本，无多余空白/HTML）
- [x] 1.2 在 `internal/controller/gatewayapp/gateway_app_register.go` 为 `/90fafbe9bf8308ecbd2063d7b479a309.txt` 增加 `BindHandler`，`ServeFile` 上述静态文件（对齐隐私政策根路径模式）

## 2. 鉴权与 usage 豁免

- [x] 2.1 将 `/90fafbe9bf8308ecbd2063d7b479a309.txt` 加入 `gateway_app_auth_exempt.go` 的 `gatewayAppAuthExemptExactGETHEAD`
- [x] 2.2 将同路径加入 `internal/services/gatewayapp/usagestats/skip.go` 的 `isStaticOrShellPath`（与 `/robots.txt` 同类）

## 3. 验收与边界

- [x] 3.1 本地或测试环境 `curl -i` 校验路径：期望 200，body 为校验串；无 Bearer 亦可访问
- [x] 3.2 确认**不新增**业务 App API、**不新增** Redis、**不新增** 背景循环；不改动 AASA / `/wx/ulink/`
- [x] 3.3 发版后公网抽检 `https://www.pangbao.cuplay.top/90fafbe9bf8308ecbd2063d7b479a309.txt`；提醒运营在微信后台完成验证与域名白名单配置，并用聊天链接做端到端验收
