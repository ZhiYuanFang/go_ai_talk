## Why

运营希望用户在微信聊天中点击 `https://www.pangbao.cuplay.top` 链接时，直接在微信内置浏览器打开官网，而不被引导到外部浏览器。微信后台要求先完成域名归属文件校验：在站点根路径提供指定 TXT 文件与固定内容。现网官网由 gateway-app 根路径托管，尚未暴露该校验路径，导致无法完成验证与后续业务域名配置。

## What Changes

- 在 `resource/public/` 增加微信域名校验文件 `90fafbe9bf8308ecbd2063d7b479a309.txt`，内容为微信下发的校验串。
- 在 `gateway-app-server` 对 `GET/HEAD /90fafbe9bf8308ecbd2063d7b479a309.txt` 显式注册静态响应（与隐私政策等根路径静态页同一模式），保证公网 `https://www.pangbao.cuplay.top/...` 可匿名访问并返回纯文本。
- 将该路径加入 Bearer 白名单与 usage 静态路径跳过列表（对齐 `/robots.txt`）。
- **无 BREAKING**：不改动现有 App API、官网 HTML、Universal Links / AASA。

## Capabilities

### New Capabilities

- `wechat-site-domain-verify`：gateway-app 在官网域名根路径提供微信域名归属校验 TXT，供微信后台验证 `www.pangbao.cuplay.top`。

### Modified Capabilities

- （无）

## Impact

- **静态资源**：`resource/public/90fafbe9bf8308ecbd2063d7b479a309.txt`
- **gateway-app**：`internal/controller/gatewayapp/gateway_app_register.go`（BindHandler）、`gateway_app_auth_exempt.go`（白名单）、`usagestats` 静态路径跳过
- **发布**：随官网静态挂载 / gateway-app 发布上线；验证前运维用 curl 确认公网 200 与 body 一致
- **运维后续（本变更范围外）**：微信后台完成文件校验后，仍须在后台把 `www.pangbao.cuplay.top` 加入对应域名白名单，并用真实聊天链接验收「微信内打开」
- **不涉及**：新增 App HTTP 业务路由语义、Redis、背景循环；usage 策略为维护型静态资源跳过（与 robots 同类，无需另行询问计入统计）
