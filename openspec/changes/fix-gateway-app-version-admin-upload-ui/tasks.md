## 1. 上传体积极限

- [x] 1.1 `config.gateway-app-server.yaml` 配置 `server.clientMaxBodySize: "220MB"`（注释说明须 ≥ apkMaxBytes）
- [x] 1.2 `docs/runbooks/release-deploy-and-run.md` 版本管理小节补充 clientMaxBodySize 与失败现象

## 2. 管理页 UI 与体验

- [x] 2.1 `gateway-app-version-admin.html` 玻璃拟态布局（对齐 pangbao-home 设计 token）
- [x] 2.2 `apiJson` 对 `Failed to fetch` 等网络错误给出可操作的提示文案
