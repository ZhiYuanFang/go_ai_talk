## 1. mcpbridge ForceRestart 与内部 API

- [x] 1.1 `Manager`：实现 `ForceRestart`（或等价）——有会话则 cancel + 删 map，再 `startLocked`；无会话则直接 `startLocked`；重置退避并立即 redial（即使 token/device/wx 未变）
- [x] 1.2 `InternalHTTP`：新增 `POST /xiaozhi-mcp/internal/api/bindings/reconnect`，内部密钥鉴权，解析 id/mcpToken/deviceNo/wxId 后调用 ForceRestart；快速返回
- [x] 1.3 （可选 SHOULD）同 binding id 轻量节流，防止连点

## 2. clients/xiaozhimcp

- [x] 2.1 新增 `BindingReconnect`（或同名）出站方法，POST 上述内部路径；错误语义与 Upsert/Remove 一致

## 3. device-service App API

- [x] 3.1 `api/v1`：**新增** reconnect Req/Res（`POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect`）；**禁止**改动既有 List/Add 等结构
- [x] 3.2 device 服务：校验 wx 拥有 active 绑定 → 读 token/deviceNo/wxId → 调 `clients/xiaozhimcp`；不调用 EnsureAccessForAdd；不改 DB
- [x] 3.3 controller 绑定方法；注册进 device-service 路由

## 4. gateway-app 与 usage

- [x] 4.1 确认 path 落在 device App 反代前缀；**不**加入 `gateway_app_auth_exempt`
- [x] 4.2 **向负责人确认** `POST .../bindings/{id}/reconnect` 是否计入 App API usage；未获明确答复前 **不得** 修改 `usagestats/maintenance_skip.go`
  - 结论：负责人确认 **计入统计**；`maintenance_skip.go` 不新增该 path
- [x] 4.3 按负责人结论处理 usage（计入则保持不进 skip；若明确排除再写入 skip）
  - 计入统计 → 不写入 `maintenance_skip.go`（已核实 skip 列表无 reconnect path）

## 5. Hub UI

- [x] 5.1 `resource/public/history.html` 小智列表：`!connected` 且已 App 登录时显示「重连」；点击 POST reconnect，成功后刷新列表；未登录禁用

## 6. 自检与跟进说明

- [x] 6.1 自检：领域路由、反代、exempt、usage 结论；无 Redis、无新 ticker；无 Python 改动
- [x] 6.2 文档/备注：Flutter 兄弟仓在 Go API 落地后另开变更（本仓 tasks 不实现 Flutter）
