## 1. device 绑定模型与内部 API

- [x] 1.1 新增小智 MCP 绑定表（wx_id、device_no、mcp_token、alias、状态、时间戳）及 entity/DAO；`mcp_token` 全局唯一索引
- [x] 1.2 实现绑定领域服务：添加（取当前 wx.device_no、未绑宝宝拒绝、token 去重）、列表（脱敏）、改 alias、删除；归属校验
- [x] 1.3 新增 device 内部 HTTP：全量 active 绑定 list（含完整 token），内部密钥鉴权
- [x] 1.4 新增 device → xiaozhi-mcp 出站 client（`internal/clients`）：Upsert/Remove；配置基址环境变量

## 2. App CRUD 与 gateway / usage

- [x] 2.1 新增 `api/v1`（或约定版本）App 路由：列表/添加/改备注/删除，完整 `g.Meta` path/method/中文 summary
- [x] 2.2 实现 device controller：从 `X-Internal-Wx-Id` 取当前用户，接领域服务
- [x] 2.3 gateway-app 自检：device App 反代前缀已覆盖；**不得**加入 auth exempt；**已确认计入 usage**——确认未写入 `maintenance_skip.go`
- [x] 2.4 写路径：绑定添加/删除成功后调用 mcp Upsert/Remove；失败打告警日志不回滚 DB

## 3. xiaozhi-mcp-service Manager 与内部 HTTP

- [x] 3.1 将 `cmd/mcp-service` 及配置/Docker/日志标识改名为 `xiaozhi-mcp-service`
- [x] 3.2 实现 Bridge Manager（map、Upsert、Remove、Reconcile）；每绑定独立 Bridge+ChatHandler(deviceNo)
- [x] 3.3 启动时从 device 内部 list 全量 Reconcile；主路径不再依赖必填单 env token（可选 env fallback）
- [x] 3.4 新增内部 HTTP 服务：Upsert/Remove/health；内部密钥鉴权；进程同时维持出站 MCP
- [x] 3.5 实现 `xiaozhi-mcp-binding-reconcile`（可配置间隔，0=关）；失败仅日志
- [x] 3.6 确认 tools/call 使用 Bridge 绑定 deviceNo，不被工具入参 xzDeviceNo 覆盖

## 4. 部署与文档

- [x] 4.1 更新 compose / kustomize：改名、`replicas: 1`、新增 ClusterIP Service 与内部端口、环境变量（DEVICE↔MCP URL、reconcile 间隔）
- [x] 4.2 更新 ACR workflow 服务名/Dockerfile 映射、`.env.example`、runbook 中 mcp 章节
- [x] 4.3 移除或迁移旧 `xiaozhi-mcp-credentials` 单 token 密钥依赖说明（稳态以 DB 绑定为准）

## 5. 自检

- [x] 5.1 服务 import / Redis bypass 门禁（若触及）通过；voice 域无他域 DAO
- [x] 5.2 手工或日志验证：添加两台不同 token → 两条桥；删一台 → 对应桥停；token 重复添加失败
- [x] 5.3 确认 App 四类接口出现在 usage 统计路径登记且不在 maintenance_skip
