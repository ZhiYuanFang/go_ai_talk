# Tasks

## 1. Gateway 会话：refresh 权威与吊销

- [x] 1.1 实现 `wxId>0` 的 `token/refresh` 始终以 device `FetchDeviceNoByWxID` 为 `device_no` 权威（忽略载荷旧号），并用换绑/改名后 refresh 响应 JWT claim 核对
- [x] 1.2 落地 D3 会话吊销（优先 sessionEpoch 或 wx→refresh 索引），提供按 wxId（批量）失效 refresh 的内部/管理入口，并验证吊销后 refresh 返回未授权
- [x] 1.3 在 `cachekit`/`keys_gateway.go`（及必要注释）登记吊销相关键或 epoch 语义，确认业务层无字面量拼键

## 2. history / cash 迁移与写前校验

- [x] 2.1 history-service 增加可信 migrate API（`history` OLD→NEW，不迁 suggest），用样例行验证迁后按新号可查、旧号无残留
- [x] 2.2 history 新增/batch create 写前经 `clients/device` `EnsureRegistered`，未注册号验证拒绝且无插入
- [x] 2.3 cash-service 增加可信 migrate API（至少 `feature_entitlement`、`feature_allowed_count`），验证迁后新号持有权益、旧号无对机行；冲突时失败可观测

## 3. device 编排与本域迁移

- [x] 3.1 实现 Admin 改名 API：校验 OLD 存在、NEW 未占用；更新 `user`/`wx`/`appointment_next`/`xiaozhi_mcp_binding`；best-effort 删本域旧 Redis；验证 `EnsureRegistered(OLD)` 失败且 NEW 成功
- [x] 3.2 编排调用 history/cash migrate 与 gateway 按绑定 wx 列表吊销；部分失败返回可读分步错误且支持幂等重试（用故意失败下游验证）
- [x] 3.3 `clients/device`（及 history/cash/gateway 所需 client）补齐调用封装，确认无跨服务 import 业务包（跑 `hack/check-service-import.ps1`）

## 4. Hub UI

- [x] 4.1 `admin.html`（或设备管理相关页）增加改名入口、二次确认、成功展示新号/失败提示，并实机点选确认走通 Admin API
- [x] 4.2 改名成功后列表/深链使用新号（旧 history URL 不再作为唯一入口），人工核对跳转

## 5. 集成核对

- [x] 5.1 端到端：注册样例宝宝→写入 history/权益→Hub 改名→旧号 device_login/bind/写 history 失败→绑定账号 refresh 失败→重登后登录 JSON 为新号且历史/权益仍在
- [x] 5.2 对照 design 重部署清单核对改动进程；`openspec validate hub-rename-baby-device-no`（或项目等价校验）通过

## 6. 补迁入口与先探再改

- [x] 6.1 history/cash migrate 与 gateway revoke 支持 `dryRun` 探测（可达+鉴权+表可读，不改写/不 bump）；device `RenameDeviceNo` 在本域写与下游调用前先探测，失败则中止且旧号仍注册时可验证本域未改
- [x] 6.2 Hub `admin.html` 增加自由填写 OLD+NEW（及可选生成新号）改名入口，二次确认后调同一 rename API；半失败场景可不依赖列表行上的旧号
- [x] 6.3 `openspec validate hub-rename-baby-device-no` 通过；列出需重部署进程（device / history / cash / gateway-app + Hub 静态）

## Workflow follow-up

- 实现合并后按仓库惯例归档本 change（合并 delta 进主规格并删除 change 目录）。
- 一账号多宝宝另开独立 change，不在本任务范围。
