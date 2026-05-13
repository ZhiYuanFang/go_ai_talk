## Why

客户端在「设备号登录」成功后需要稳定拿到 **设备号**（用于后续请求头、埋点或与 JWT 内 `device_no` 声明对齐）。当前 `GatewayAppDeviceLoginRes` / `DeviceWxDeviceLoginRes` 虽已声明 `deviceNo` 字段，但在部分链路（例如下游 JSON 缺字段、或网关仅解析 `data` 未兜底）可能出现 **响应体中无可用设备号**，导致前端无法展示或无法缓存。本变更将 **成功响应中必须带回非空设备号** 固化为契约，并在网关侧增加与请求入参一致的兜底，避免再出现「登录成功却拿不到号」的断裂体验。

## What Changes

- 明确 **POST `/device/app/api/device_login`（网关聚合）** 与 **POST `/device/app/api/user/device_login`（device 纯业务）** 在 **业务成功（code=0）** 时，`data` 中 **MUST** 包含 **`deviceNo`**，且为 trim 后的非空字符串（与登录所用设备号一致，或与库内权威 `device_no` 一致）。
- **网关聚合实现**：若解析 device 返回的 `data.deviceNo` 为空，**SHALL** 回退为本次请求体中的 `deviceNo`（trim 后），再用于签发 access/refresh 与写回响应；若仍为空则保持现有「内部错误 / 参数错误」语义。
- **device 服务实现**：复核 `WxDeviceLoginByDeviceNo` → `DeviceWxDeviceLoginRes` 路径，确保序列化出的 JSON 始终带 `deviceNo`（与现有一致则可为 no-op，仅规格锁定）。

## Capabilities

### New Capabilities

- `gateway-app-device-login-device-no`：设备号登录成功响应中 `deviceNo` 字段的契约与网关兜底策略。

### Modified Capabilities

- （无）`openspec/specs/` 中无同名基线规格；本变更在 change 内新增能力规格即可。

## Impact

- 代码：`internal/controller/gateway_app_ctrl.go`（`DeviceLogin`）、必要时 `internal/controller/device_app_user.go`（仅复核）。
- 对外 JSON：成功响应体字段不变名（`deviceNo`），行为为 **更强保证**；**非 BREAKING**（客户端已按字段解析则更稳；未解析者可新用）。
