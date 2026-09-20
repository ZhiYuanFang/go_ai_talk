## Context

胖宝官网由 `gateway-app-server` 托管，公网域名为 `https://www.pangbao.cuplay.top`（Nginx 反代至 App 网关）。根路径已有显式路由：`/` → `pangbao-home.html`、`/privacy-policy.html`、`/user-agreement.html`、Apple AASA 等；Bearer 白名单与 usage 跳过列表对上述匿名静态路径单独维护。

微信侧要求在站点根目录放置校验文件：

- 路径：`/90fafbe9bf8308ecbd2063d7b479a309.txt`
- 内容：`41a8fe99b666f6646724f29ea87962e34c6df2b6`（纯文本）

目的：完成域名归属校验后，可在微信后台配置该域名，使聊天内点击官网链接优先在微信内置浏览器打开。

## Goals / Non-Goals

**Goals:**

- 公网 `https://www.pangbao.cuplay.top/90fafbe9bf8308ecbd2063d7b479a309.txt` 匿名 GET/HEAD 返回 200，body 为微信下发校验串（无 HTML/JSON 壳、无登录跳转）。
- 实现与现有根路径静态页模式一致，可随官网静态挂载发布。

**Non-Goals:**

- 不负责微信后台「业务域名 / 安全域名」勾选与填写（属运维/运营操作）。
- 不保证微信策略下所有机型/版本都禁止「用浏览器打开」入口。
- 不改动 Universal Links、AASA、`/wx/ulink/`、微信登录 OAuth。
- 不新增业务 App API；不引入 Redis / 背景循环。

## Decisions

### D1：静态文件 + 显式 BindHandler（对齐隐私政策）

- **选择**：文件落在 `resource/public/90fafbe9bf8308ecbd2063d7b479a309.txt`；在 `RegisterGatewayAppHTTP` 中 `BindHandler` + `ServeFile`，响应 `Content-Type` 保持文本、`Cache-Control` 可用短缓存或 `no-store`（校验场景优先正确性，建议 `no-store` 与合规页一致或短 public cache）。
- **备选 A**：仅 Nginx `alias` — 快但不出仓库，换机易丢；否决（要随官网一并上线、可复现）。
- **备选 B**：Handler 内联写死字符串 — 可行，但与「校验文件」运维心智不一致；优先文件 + ServeFile。

### D2：鉴权与 usage 与 `/robots.txt` 对齐

- **选择**：将精确路径加入 `gatewayAppAuthExemptExactGETHEAD`；在 `usagestats` 静态/壳路径跳过列表中加入同路径。属维护型校验资源，不计入 App usage。
- **备选**：依赖文件服务器默认放行 — 根路径并非仅靠 `SetServerRoot` 暴露 `resource/public`，不可靠。

### D3：校验串写死在仓库

- **选择**：本次微信下发的文件名与内容写入仓库；若微信重新下发校验串，另开小变更替换。
- **备选**：配置项注入 — 过重，单次归属校验无此必要。

## Risks / Trade-offs

- **[仅有文件无 BindHandler → 公网 404]** → tasks 强制注册路由并 curl 验收。
- **[未进白名单 → 401/跳转]** → 同步改 auth exempt。
- **[TXT 验过但聊天仍外开]** → 文档/ tasks 注明须在微信后台完成域名配置与真机验收；本变更只解决归属文件。
- **[校验串过期/更换]** → 替换文件内容与规格中的字面量即可。

## Migration Plan

1. 合并后发布 gateway-app（`resource/public` 若 volume 挂载，文件可即时可见；BindHandler / 白名单变更须重启 gateway-app）。
2. 验收：`curl -i https://www.pangbao.cuplay.top/90fafbe9bf8308ecbd2063d7b479a309.txt` → 200，body 精确匹配校验串。
3. 运营在微信后台点击验证，并完成域名白名单配置；用聊天链接做端到端抽检。
4. 回滚：移除 BindHandler、白名单项与静态文件，回退发布。

## Open Questions

- （无阻塞）微信后台具体菜单名称以运营控制台为准；技术侧交付根路径校验文件即可。
