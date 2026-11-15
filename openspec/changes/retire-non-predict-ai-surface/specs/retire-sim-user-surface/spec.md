## ADDED Requirements

### Requirement: sim-user-service MUST 整进程下线

系统 MUST 移除 `sim-user-service` 进程（cmd、Dockerfile、compose 服务、ACR 镜像别名、`config.sim-user-service.yaml`、`SIM_*` / `SIM_SERVICE_BASE_URL` 环境变量）、`internal/services/simuser` 与 `controller/simuser`、gateway `/sim/admin/api/*` 反代，以及 Hub `sim-admin.html` 与导航入口。

#### Scenario: compose 无 sim 服务

- **WHEN** 运维查看 microservices compose
- **THEN** MUST NOT 再包含可启动的 `sim-user-service` 服务定义（或等价已删除）

### Requirement: device/ucg sim-only API MUST 删除

device-service MUST 删除 `/device/internal/api/sim/*` 与 `sim_user.go`。ucg-service MUST 删除 `POST /ucg/internal/api/posts/sample`、`POST /ucg/internal/api/chat/sim-unread-sample`、`POST /ucg/internal/api/chat/send`（仅允许模拟发送方的内部发送）。**MUST 保留** `POST /ucg/internal/api/media/upload-video`（小智引导片等）。

#### Scenario: sim register 不可用且 upload-video 仍在

- **WHEN** 调用方请求原 sim username register，同时另一调用方请求 `upload-video`（如 kind=xiaozhi-guide）
- **THEN** sim register MUST 不可用；upload-video MUST 仍可处理合法上传

### Requirement: wx.is_simulated 列 MUST 删除

device-service MUST 停止读写 `wx.is_simulated`，Ensure/迁移 MUST 在列存在时执行 `DROP COLUMN is_simulated`（或等价）。部署 MUST 先清理 `is_simulated=1` 行（删除或注销），再 DROP 列。Admin `wx/list` MUST NOT 再依赖该列过滤。`/device/internal/api/ucg/wx/batch` 响应 MUST NOT 再包含 `isSimulated` 字段。

#### Scenario: 启动后列不存在

- **WHEN** device-service 完成 schema Ensure 且库中曾有 `is_simulated` 列
- **THEN** `wx` 表 MUST 不再包含 `is_simulated` 列

### Requirement: gateway 模拟号 usage 跳过 MUST 删除

gateway-app MUST 删除 `IsSimulatedWx` 及 usagestats/clientusage 中对模拟号的跳过逻辑，以及相关 Redis sim wx set 键写入。删除后 usage 统计 MUST 仅按既有其它维护型 skip 规则过滤，MUST NOT 再查询模拟号集合。

#### Scenario: 统计路径无模拟号分支

- **WHEN** 成功 App 请求进入 usage 记录逻辑
- **THEN** 实现 MUST NOT 调用已删除的 `IsSimulatedWx` 判定

### Requirement: 小智 MCP 链路 MUST 不受影响

删除 sim 表面时，系统 MUST 保留小智 MCP 的 cash 开通/试用/引导视频、device bindings、mcpbridge/voice chat 相关路径，以及 ucg `upload-video`。

#### Scenario: 小智引导片上传仍可用

- **WHEN** Hub/cash 经既有客户端上传小智引导视频
- **THEN** `upload-video` MUST 成功写入对象键（或既有错误语义），MUST NOT 因删除 sim 而路由消失
