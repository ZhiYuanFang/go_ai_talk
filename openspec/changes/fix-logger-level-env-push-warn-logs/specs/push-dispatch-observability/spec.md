## ADDED Requirements

### Requirement: Internal by-biz-type 受理可观测

`POST /push/internal/api/by-biz-type` 处理路径 MUST 在以下情况输出 **Warning** 级日志（前缀建议 `[push]`），且 MUST NOT 打印完整 device token 或完整 alert 正文：

- 内部密钥校验失败
- 请求体/必填字段校验失败
- 校验通过并受理发送（含 `wxId`、`bizType`、`badge`、`silent`、`alertLen`）

#### Scenario: 受理成功有日志

- **WHEN** 合法内部请求到达且校验通过
- **THEN** push-service 日志中 MUST 出现一条 Warning，表明已 `accepted` 该 `wxId` 与 `bizType`

#### Scenario: 鉴权失败有日志

- **WHEN** 内部密钥无效或缺失
- **THEN** 在返回 403 的同时 MUST 打 Warning（`auth_denied` 或等价 reason）

### Requirement: Dispatch 与厂商发送可观测

异步 dispatch 路径 MUST：

- 在查询到待发设备列表后打 Warning，含 `wxId`、`bizType`、`deviceCount`（`deviceCount=0` 时沿用或等价于现有 `no_device` skip）
- 某 channel 因凭证未配置而跳过时打 **Warning**（不得仅 Debug）
- 某 token 发送成功时打 Warning（`send_ok` 或等价）
- 异步 goroutine panic 时打 Error/Warning，不得空 recover 吞掉

#### Scenario: 无注册设备

- **WHEN** 受理后该 `wxId` 在 `push_device` 无可用 token
- **THEN** 日志 MUST 以 Warning 标明 skip（如 `no_device`），且在 `GF_LOGGER_LEVEL=prod` 下仍可见

#### Scenario: 凭证未配置

- **WHEN** 存在设备记录但对应 channel 厂商凭证未配置
- **THEN** 日志 MUST 以 Warning 标明该 channel 因凭证未配置而跳过（`GF_LOGGER_LEVEL=prod` 下仍可见）

#### Scenario: 发送成功

- **WHEN** 某 channel sender 返回成功（无 error）
- **THEN** 日志 MUST 以 Warning 标明 `send_ok`（或等价），含 `channel`、`wxId`、`bizType`
