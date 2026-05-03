# voice-device-profile-http-contract Specification

## ADDED Requirements

### Requirement: 画像链路的「本地」实现 MUST 不依赖进程内 user DAO

即使存在 `localDeviceProfileAdapter` 类实现，`voice-service` 在生产单库模式下 MUST 使用 **HTTP 远程实现**（或指向 device 基址的 HTTP local）；MUST NOT 依赖在 voice 进程内对 `dao.User` 的查询作为获取画像的主路径。

#### Scenario: 生产 voice 配置

- **WHEN** 部署 `voice-service` 且 `database.default` 仅含 voice 域表
- **THEN** 设备画像 MUST 通过 `device-service` 的 HTTP 接口获取；若误配为进程内 local 适配器，MUST 视为配置错误并暴露启动或运行期检查（若实现）
