# voice-device-profile-http-contract Specification

## Purpose
TBD - created by archiving change enforce-service-boundary-no-cross-db. Update Purpose after archive.
## Requirements
### Requirement: Voice MUST 通过 Device 服务契约获取设备画像数据
`voice-service` 在涉及设备信息、用户生日、性别、注册状态等 device/profile 领域数据时 MUST 通过 `device-service` 暴露的内部 HTTP 接口获取，MUST NOT 直接访问 `user/device` 领域数据库表。

#### Scenario: 通用问答需要设备画像
- **WHEN** voice 在生成通用问答提示词时需要生日或性别等画像信息
- **THEN** voice MUST 调用 device 内部接口获取画像数据，并将结果用于提示词构建

#### Scenario: 设备信息接口不可达
- **WHEN** voice 调用 device 内部画像接口超时或网络失败
- **THEN** voice MUST 返回可观测的错误语义，并按配置决定是否执行迁移期兜底

### Requirement: Device 内部画像接口 MUST 提供一致错误结构
device 内部画像接口 MUST 对参数错误、设备不存在、服务异常返回统一错误结构，供 voice 侧做稳定错误映射。

#### Scenario: 设备不存在
- **WHEN** voice 传入的 `deviceNo` 在 device 服务中不存在
- **THEN** device MUST 返回可区分的业务错误码，voice MUST 返回可理解的失败信息

#### Scenario: 参数缺失
- **WHEN** voice 调用画像接口时缺失关键参数
- **THEN** device MUST 返回参数错误结构，voice MUST 记录请求参数异常日志

### Requirement: 画像链路的「本地」实现 MUST 不依赖进程内 user DAO

即使存在 `localDeviceProfileAdapter` 类实现，`voice-service` 在生产单库模式下 MUST 使用 **HTTP 远程实现**（或指向 device 基址的 HTTP local）；MUST NOT 依赖在 voice 进程内对 `dao.User` 的查询作为获取画像的主路径。

#### Scenario: 生产 voice 配置

- **WHEN** 部署 `voice-service` 且 `database.default` 仅含 voice 域表
- **THEN** 设备画像 MUST 通过 `device-service` 的 HTTP 接口获取；若误配为进程内 local 适配器，MUST 视为配置错误并暴露启动或运行期检查（若实现）

