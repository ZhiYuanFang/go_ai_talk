## 1. Redis 键与存储

- [x] 1.1 在 `cachekit` 登记 `gw:featusage:*` 键 builder（日聚合 g/w/x、lastAt、时间线、限流），中文注释标明 TTL≈30 天与用途
- [x] 1.2 实现 client-usage store：聚合 HINCRBY + Expire；时间线 append + 裁剪至 ≤10000；条目仅 `featureId` + 服务端时间
- [x] 1.3 实现 per-wx 3 秒限流键；模拟用户跳过写入（复用现有 sim 判定）

## 2. App 上报 API

- [x] 2.1 `api/v1` + gateway-app controller：`POST /device/app/api/client-usage/report`（须登录；校验 featureId）
- [x] 2.2 本机注册路由，确认不被 device 反代吞掉；gateway Bearer 豁免 **不** 需要（须登录）
- [x] 2.3 `maintenance_skip.go` 精确登记该 POST（不计入 App API usage；已确认）

## 3. Admin 读 API 与 Hub

- [x] 3.1 Admin API：按功能列表、按用户、用户详情（含时间线）、功能下钻；Admin JWT；窗口 days≤30
- [x] 3.2 Hub：`admin-modules.js` 入口「客户端使用统计」+ 静态页（对齐 api-usage-stats 交互与 hint）
- [x] 3.3 静态路由 / `admin_static_pages` 登记（若项目需要）

## 4. 联调与自检

- [x] 4.1 自检：未登录拒绝；sim 不记；3s 第二次失败；时间线字段与 1 万裁剪；report 不进 API usage
- [x] 4.2 （可选跨仓）Flutter 调用 report 的埋点约定文档或示例；不阻塞后端合并
- [x] 4.3 确认无 MySQL、无 MQ、无新 ticker；未改支付宝/ASN/开通路径
