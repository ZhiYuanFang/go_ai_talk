## Why

当前 `gateway-app-server` 的根路径 `/` 仅返回“智能语音 App 网关”文本，无法承载对外品牌展示、事件能力介绍与 App 下载说明。业务希望将官网直接放在 `gateway-app` 的 `/`，以品牌“胖宝”统一对外表达“专注母婴喂养，更便捷、更轻松地照顾孩子”的定位，同时复用现有事件 logo 与 Android 下载数据能力。

该变更需要现在推进，因为事件 logo、Android `downloadUrl`、同源图片代理与静态页承载已经具备基础能力，只缺少一个面向公众的官网入口与公开聚合数据接口。同时必须明确边界：官网替换的仅是 `gateway-app-server` 的 `/`，不得影响主网关或其他服务进程的既有入口。

## What Changes

- 在 `gateway-app-server` 的根路径 `/` 提供品牌官网页面，页面名称为“胖宝”，替换当前纯文本响应。
- 官网页面采用玻璃拟态视觉风格，突出“专注母婴喂养服务商”“更便捷、更轻松地照顾孩子”等品牌表达。
- 官网 SHALL 从数据库权威链路读取事件信息并展示事件 logo 与事件名；对外页面只读展示，不新增运营写入能力。
- 官网 SHALL 包含应用下载说明：
  - Android：从数据库读取最新下载链接，经网关聚合为官网可直接使用的下载地址，并在页面生成二维码展示。
  - iOS：展示“前往 App Store 搜索‘胖宝’下载”的固定说明。
- 新增官网公开聚合接口，由 `gateway-app-server` 经服务契约聚合事件列表与 Android 下载信息，避免前端直接访问受保护业务接口。
- 官网相关匿名访问白名单、静态资源与数据接口仅在 `gateway-app-server` 暴露，**不得** 扩散到主网关或其他服务。
- **BREAKING**：`gateway-app-server` 的 `/` 从纯文本探活页改为官网 HTML；依赖旧文本内容的人工访问习惯将变更，但服务进程、业务 API、主网关入口与下载路由语义保持不变。

## Capabilities

### New Capabilities
- `gateway-app-official-site`: `gateway-app-server` 对外提供官网首页、官网公开聚合数据接口，以及 Android 下载二维码与事件展示能力。

### Modified Capabilities
- 无

## Impact

- Affected code：`internal/controller/gateway_app_register.go`、`internal/controller/gateway_app_auth_exempt.go`、`internal/controller/gateway_app_ctrl.go` 或新增官网聚合控制器、`resource/public/` 下官网静态资源。
- Affected systems：`gateway-app-server`、device-service 事件字典读取链路、`ai_voice_app.version` 最新版本读取链路。
- APIs：新增官网公开聚合接口；`GET /` 在 `gateway-app-server` 上改为返回官网页面；现有 `/device/app/api/version/check`、`/ai_talk_images/*`、`/device/app/apk/*` 继续兼容。
- Dependencies：前端需引入浏览器端二维码生成方案；后端继续通过 HTTP 契约访问事件权威服务，禁止跨服务直连他域库。
