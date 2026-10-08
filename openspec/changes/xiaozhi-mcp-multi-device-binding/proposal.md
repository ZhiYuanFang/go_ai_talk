## Why

当前 `mcp-service` 以环境变量绑定**单一**小智 MCP token 与 CuPlay `deviceNo`，无法服务 C 端用户「一人多台小智音箱」的场景：用户从 xiaozhi.me 取得各音箱的 MCP token 后，需在 Flutter（已绑宝宝账号）中登记多条绑定，日后对话音箱即可经 MCP 完成喂养语音录入。历史上该进程尚无生产用户，可直接演进为多绑定连接池并改名为 `xiaozhi-mcp-service`。

## What Changes

- **device-service**：新增小智 MCP 绑定表与 App CRUD（列表/添加/改备注/删除）；添加时 `deviceNo` 取当前账号已绑宝宝；`mcp_token` **全局唯一**去重；备注 `alias` 便于用户管理。
- **gateway-app**：放行 device 前缀下新 App 路由；小智绑定 CRUD **计入 App API usage 统计**（负责人已确认；**不得**写入 `maintenance_skip.go`）。
- **xiaozhi-mcp-service**（由 `mcp-service` **直接改名**）：进程内 Manager 维护 N 条 Bridge（一 token 一条出站连接）；部署约定 **replicas=1**。
- **绑定同步**：启动全量拉取 device 内部 list + 写路径 **内部 HTTP** Upsert/Remove + **低频 reconcile** 兜底。
- 废弃「启动必填单个 env token」为主路径（可选保留 env 单设备作迁移期 fallback）。
- **不做**：多副本分片/选主、Redis Pub/Sub 写路径、Flutter 直连 mcp、`intent-mcp-service`（后续另案）。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-binding`：用户侧小智 MCP token 与宝宝 `deviceNo` 的绑定存储、App CRUD、全局 token 去重、内部全量 list。
- `xiaozhi-mcp-bridge-manager`：单副本 Manager、多 Bridge 拨号/重连、内部 HTTP 增量同步、启动全量与低频 reconcile、进程改名与部署。

### Modified Capabilities

- （无基线能力 REQUIREMENTS 变更；既有单 env 绑定行为被本变更新能力取代，以新 specs 为准。）

## Impact

- **代码**：`internal/services/device`、`api/v1` device App 路由、`cmd/mcp-service`→`cmd/xiaozhi-mcp-service`、`internal/services/mcpbridge`（Manager + 内部 HTTP）、`internal/clients`（device↔mcp 内部调用）、compose/kustomize/ACR/Dockerfile/config/runbook。
- **API**：新增经 gateway-app 的小智绑定 CRUD（计入 usage）；新增 mcp 内部 HTTP（不计入 App usage）。
- **依赖**：xiaozhi-mcp-service 新增内部监听端口与对 device-service 的出站 HTTP；仍不连业务 MySQL（绑定权威在 device DB）。
- **运维**：Deployment `replicas: 1`；补 K8s Service（内部口）；环境变量与镜像名随改名更新。
- **背景循环**：批准 `xiaozhi-mcp-binding-reconcile`（低频全量对齐），见 design。
