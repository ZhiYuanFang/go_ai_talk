## Context

App 网关已具备 **`POST /device/app/api/login`**（微信 `jsCode` → device 换票 → 签发 JWT）与 **`POST /device/app/api/user/login`**（device 纯业务）。设备域存在 **`user`/`wx` 绑定** 模型：`wx.device_no` 关联已注册设备。联调需要 **不调用微信** 即可拿到与微信登录 **同形态** 的 access/refresh，以便验证 Bearer 与下游。

## Goals / Non-Goals

**Goals:**

- device 提供 **`POST /device/app/api/user/device_login`**：入参仅 **`deviceNo`**；校验设备已注册且 **wx 行已绑定该设备**；返回 **`wxId`、`deviceNo`、`isNewUser`**（语义与现网 device 微信登录响应子集一致，**无 token**）。
- gateway 提供 **`POST /device/app/api/device_login`**：转发 device 设备登录后 **签发 access/refresh**（与微信聚合登录复用 **`SignAccess`/`IssueRefreshToken`** 路径）；加入 **Bearer 白名单**。
- 联调页增加调用 **`/device/app/api/device_login`** 的按钮与日志区。

**Non-Goals:**

- 不引入「无 wx 行、纯设备 JWT」的新身份模型（`sub` 仍为 **wx 主键**）。
- 不将 `device_no` 作为长期密码等价物做复杂风控（生产是否开放由部署与产品策略决定，本变更以**可观测错误语义**与**仅已绑 wx 的设备**为硬条件）。

## Decisions

### D1：device 侧成功条件

- **决策**：`device_no` **MUST** 在设备注册表（如 `user`）中存在；**MUST** 存在 **`wx` 行**满足 `wx.device_no = device_no`（trim 后一致），否则返回业务错误（设备不存在 / 未绑定账号等文案在实现中区分或合并，**评审时单一化**）。
- **理由**：网关签发 JWT 的 `sub` 依赖 **wx.id**；无 wx 则无法与 refresh 会话模型对齐。

### D2：响应字段

- **决策**：device 返回 **`wxId`、`deviceNo`、`isNewUser: false`**（固定 false，与「设备登录非首登」语义一致）。网关聚合响应与 **`GatewayAppLoginRes`** 字段对齐（**`accessToken`、`refreshToken`** 等）。
- **备选**：省略 `isNewUser` → 为减少客户端分支，**建议保留且恒 false**。

### D3：网关路径与白名单

- **决策**：聚合路径定为 **`POST /device/app/api/device_login`**（与 **`/device/app/api/login`** 并列）；device 路径 **`POST /device/app/api/user/device_login`** 加入网关 **Bearer 豁免**（与 `user/login` 同类）。
- **理由**：与现有「网关聚合 vs device 业务」命名对称。

### D4：安全与滥用面

- **决策**：接口 **不** 新增独立共享密钥；依赖 **device_no 已注册且已绑 wx** 与现网一致的数据前置条件。生产若需限流/审计，在实现阶段打结构化日志（**禁止**在响应中回显敏感配置）。
- **缓解**：错误响应不泄露「设备存在但未绑 wx」与「设备不存在」的细微差别，若产品要求防枚举，可在实现中统一错误文案（OpenSpec 场景可写「统一错误」）。

## Risks / Trade-offs

- **[Risk] 设备号可被枚举尝试**：与任意「已知 device_no」接口同类 → **缓解**：产品侧限流、WAF、或仅内网联调开启。
- **[Risk] 与微信登录并存两套入口**：文档需标明 **`device_login` 仅适用于已绑机场景** → **缓解**：联调页文案 + README/OpenAPI summary。

## Migration Plan

1. 先部署 **device-service**（新路由）。
2. 再部署 **gateway-app-server**（聚合 + 白名单）。
3. 静态资源联调页随网关发布。

## Open Questions

- 是否在 **config** 增加开关（如 `gatewayApp.deviceLoginEnabled`）以便生产默认关闭；若需要，在 **tasks** 中列为可选任务。
