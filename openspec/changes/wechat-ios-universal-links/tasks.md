## 1. 服务端托管与配置

- [x] 1.1 为 `gateway-app-server` 设计并接入 AASA 输出方案，支持 `/apple-app-site-association` 与 `/.well-known/apple-app-site-association` 两个公开入口
- [x] 1.2 增加 Team ID 独立配置项或环境变量读取，并在缺失时返回显式不可用语义与可观测日志
- [x] 1.3 固化 AASA 内容模板：Bundle ID 使用 `com.fzy.pangbao`，放行路径使用 `/wx/ulink/*`
- [x] 1.4 当公网站点切到根路径后，同步调整 Universal Links 前缀与边缘代理规则，并记录部署前提

## 2. 文档与接入说明

- [x] 2.1 在 `docs/runbooks/` 新增或更新 Universal Links runbook，说明 AASA 必须从 `www.pangbao.cuplay.top` 主机根路径获取
- [x] 2.2 在 runbook 中补充微信开放平台填写值，明确 Universal Links 使用 `https://www.pangbao.cuplay.top/wx/ulink/`
- [x] 2.3 在 runbook 中补充 GitHub 打包链路说明：Associated Domains、`applinks:www.pangbao.cuplay.top`、entitlements、Provisioning Profile、Secrets 与签名校验
- [x] 2.4 在 runbook 中说明 Team ID 待补位策略，列出部署前需补齐的环境变量/私有配置项

## 3. 验证与发布准备

- [x] 3.1 使用 `curl` 或等价命令验证两个 AASA 入口返回 `200`、无重定向且内容一致
- [ ] 3.2 手工检查 `Content-Type`、HTTPS 证书链和边缘代理行为，确认不会把 AASA 请求改写到其它业务路径
- [ ] 3.3 对照 runbook 检查 iOS GitHub 打包产物是否包含正确的 Associated Domains / entitlements
- [ ] 3.4 在 Team ID 补齐后，完成微信开放平台与真机 Universal Links 联调验证，并记录最终回填值
