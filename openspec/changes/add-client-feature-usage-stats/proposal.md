## Why

现有 App API 使用统计只能反映网关看到的 HTTP 调用，无法回答「用户在客户端页面上用了什么功能、按时间做了什么」。产品需要客户端主动上报的功能触达统计，且接受仅 Redis、可丢失。

## What Changes

- 在 **gateway-app** 新增登录态 App 上报接口：客户端提交自由字符串 `featureId`（事件名），服务端记聚合 + 用户时间线。
- 数据 **仅存 Redis**（可丢）；聚合与时间线窗口统一 **最多 30 天**；单用户时间线硬上限 **1 万条**（超则丢最旧）。
- 同 wxId **3 秒内最多成功上报 1 次**；模拟用户跳过；未登录拒绝。
- 上报接口 **不计入** App API 使用统计（`maintenance_skip`；负责人已确认）。
- Hub 新增独立入口 **「客户端使用统计」**，交互对齐「App API 使用统计」：按功能 / 按用户；用户下钻含时间线（事件名 + 服务端时间）。
- 无 MySQL、无 MQ、无扫表 ticker；不改计费/开通路径。

## Capabilities

### New Capabilities

- `client-feature-usage-stats`: 客户端功能上报、Redis 聚合/时间线、限流与跳过规则、Admin 读 API 与 Hub 页。

### Modified Capabilities

- （无既有 capability 需求条文变更；仅向 App API usage 的 denylist 增补一条精确 path，属实现约定。）

## Impact

- **进程**：`gateway-app-server`（本机处理 report + Admin 读；与现有 `usagestats` 同进程同 Redis）。
- **API**：新增 App `POST` 上报；新增 Admin `GET` 列表/详情/时间线；路径建议挂在 `/device/app/api/…` 与 `/device/admin/api/…`（本机，勿被 device 反代吞掉）。
- **Redis**：`cachekit` 新键族（如 `gw:featusage:*`）；访问经 `cachekit`；TTL/裁剪见 design。
- **Hub**：`admin-modules.js` + 静态 HTML；Admin JWT。
- **Flutter**：本变更可只交付后端+Hub；客户端埋点调用为联调依赖（跨仓，tasks 标明）。
- **约定**：新增 App 路由已确认 **不计入** usage；gateway 反代前缀无需扩展（本机路由）。
