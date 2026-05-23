## Context

- 网关 `POST /device/app/api/device_login` 调用 device `POST /device/app/api/user/device_login`，从返回 JSON 的 `data` 读取 `wxId`、`deviceNo`、`isNewUser` 后签发 JWT。
- `api/v1` 中 `GatewayAppDeviceLoginRes`、`DeviceWxDeviceLoginRes` 已含 `deviceNo` 字段；`device.WxDeviceLoginByDeviceNo` 在多数分支已填充 `DeviceNo`。
- 仍可能出现 **下游 JSON 未带 `deviceNo`** 或 **解析结果为空** 时网关直接报错或返回空字段，客户端感知为「登录成功但无设备号」。

## Goals / Non-Goals

**Goals:**

- 成功登录响应中 **`deviceNo` 对调用方可见且非空**（与本次登录所用设备号一致，或与库内权威值一致）。
- 网关在解析 `data.deviceNo` 为空时，**用请求体 `deviceNo` 兜底** 再签发与回包，避免无谓失败与空字段。

**Non-Goals:**

- 不改变登录失败（未注册、参数非法）的 HTTP/业务码语义。
- 不扩展 refresh/login 以外接口的字段（除非规格另行变更）。

## Decisions

1. **兜底主战场放在网关 `DeviceLogin`**  
   - **理由**：聚合层掌握请求入参，且不增加 device 对外契约的二次往返；与「签发 access 所需 device_no」同源。

2. **device 服务保持显式返回 `deviceNo`**  
   - **理由**：直连 device 的客户端（内部或其它网关）也能依赖同一字段；若实现已满足则仅补规格与回归。

3. **兜底仅使用 trim 后的请求 `deviceNo`，且仅在业务成功分支**  
   - **理由**：失败路径不应伪造设备号；成功且下游缺字段时，请求值已通过 `EnsureRegistered` 等价校验链。

## Risks / Trade-offs

- **[Risk] 请求体与库内规范形式（大小写/空格）不一致**  
  - **缓解**：统一 `strings.TrimSpace`；JWT 签发与回包使用同一字符串。

## Migration Plan

- 先部署 **gateway-app-server** 即可见效；若 device 有独立发布节奏，可仅网关先行。

## Open Questions

- 无。
