# device-route-canary-management Specification

## Purpose
TBD - created by archiving change voice-device-canary-route-split. Update Purpose after archive.
## Requirements
### Requirement: Gateway MUST 为 device 路由提供独立可配置代理能力
gateway MUST 以独立中间件管理 `/device/admin/api/*` 路由，并支持 `local|proxy|canary` 三态。

#### Scenario: device 路由进入 local 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=local`
- **THEN** gateway MUST 执行本地处理链路，且 MUST NOT 将请求转发到 device-service

#### Scenario: device 路由进入 proxy 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=proxy` 且 `DEVICE_API_PROXY_URL` 可用
- **THEN** gateway MUST 将 `/device/admin/api/*` 请求全量转发到 device-service

#### Scenario: device 路由进入 canary 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=canary` 且配置了 `DEVICE_API_PROXY_CANARY_PERCENT`
- **THEN** gateway MUST 按稳定分流键执行百分比转发，其余请求保持本地处理

### Requirement: device canary 分流 MUST 保持同键稳定
gateway MUST 采用稳定分流键（如 deviceNo 或请求头标识）对 canary 流量做无状态一致性计算。

#### Scenario: 同一分流键连续请求
- **WHEN** 同一设备在 canary 模式下发起多次 `/device/admin/api/*` 请求
- **THEN** 请求 MUST 稳定命中同一流量路径（proxy 或 local）

