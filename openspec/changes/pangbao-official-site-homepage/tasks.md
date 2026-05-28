## 1. gateway-app 后端：官网入口与公开聚合接口

- [x] 1.1 在 `gateway_app_register.go` 将 `GET /` 从纯文本响应切换为返回官网静态页（仅限 `gateway-app-server`）
- [x] 1.2 新增官网匿名只读聚合接口（如 `GET /device/app/api/site/home`），返回品牌文案、事件展示列表、Android 下载信息与 iOS 提示
- [x] 1.3 在 `gateway_app_auth_exempt.go` 为官网聚合接口登记匿名白名单；确认仅新增官网所需公开路径
- [x] 1.4 官网聚合逻辑通过服务契约读取事件权威数据，并复用 gateway-app 现有版本读取能力组装 Android 下载地址
- [x] 1.5 处理降级语义：事件列表为空时返回空数组；Android 无下载链接时返回显式不可用状态与提示文案

## 2. 官网前端：胖宝品牌首页

- [x] 2.1 在 `resource/public/` 新增胖宝官网页面与所需静态资源，UI 使用玻璃拟态风格
- [x] 2.2 首页首屏展示品牌名“胖宝”、母婴喂养定位文案，以及“更便捷、更轻松地照顾孩子”的核心表达
- [x] 2.3 通过官网聚合接口渲染事件卡片，展示事件 logo 与事件名，并保证 path-only logo 在当前站点同源可访问
- [x] 2.4 实现应用下载区：Android 展示二维码与下载入口；iOS 展示“前往 App Store 搜索‘胖宝’下载”说明
- [x] 2.5 对官网数据加载失败、Android 暂无下载、logo 缺失等情况提供可读的降级展示

## 3. 集成与边界验证

- [ ] 3.1 手工验证：访问 `gateway-app-server` 的 `/` 返回胖宝官网，而不是“智能语音 App 网关”纯文本
- [ ] 3.2 手工验证：官网匿名聚合接口可读，返回事件列表、Android 下载信息和 iOS 提示，且不暴露用户态或管理态数据
- [ ] 3.3 手工验证：官网中的事件 logo 可正常加载；Android 二维码扫码后可命中有效下载地址或正确降级提示
- [ ] 3.4 手工验证：`/device/app/api/version/check`、`/device/app/apk/*`、`/ai_talk_images/*` 等既有能力保持可用
- [ ] 3.5 手工验证：主网关根路径、代理能力与其他服务入口不受本次官网变更影响
- [x] 3.6 执行 `openspec validate pangbao-official-site-homepage --strict` 并确保通过
