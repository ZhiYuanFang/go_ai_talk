## Why

App 版本管理页上传 25MB APK 时浏览器报 `failed to fetch`；根因是 GoFrame `ghttp.Server` 默认 `clientMaxBodySize` 为 8MB，大于该值的 multipart 在进 handler 前被断开。管理页 UI 与胖宝官网玻璃拟态风格不一致。

## What Changes

- `config.gateway-app-server.yaml` 增加 `server.clientMaxBodySize`（≥ `gatewayApp.apkMaxBytes`）。
- `release-deploy-and-run.md` 补充 APK 上传体积极限说明。
- `gateway-app-version-admin.html` 玻璃拟态 UI，并改进网络层失败提示。

## Capabilities

### Modified Capabilities

- （增量落在既有 gateway-app-version-admin 运维约定，无新 spec 目录）

## Impact

- **进程**：gateway-app-server（配置 + 静态页）
- **部署**：需重启 gateway-app 使 `clientMaxBodySize` 生效
