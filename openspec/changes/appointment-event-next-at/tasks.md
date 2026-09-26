## 1. 数据模型与字典标志

- [x] 1.1 device DB：`event` 表增加预约列（默认非预约）；生成/更新 DAO、entity、必要时 Do，JSON 字段 `isAppointment`
- [x] 1.2 device DB：新建 `(device_no, event_id)` 唯一的下次约定表；DAO/entity；`next_at=0` 语义在注释中固定（保留行或删行二选一）
- [x] 1.3 `AddEvent` / `UpdateEvent` / admin multipart 表单支持读写 `isAppointment`；变更后走既有 EventOptions 缓存重建

## 2. 预约下次 App API（device-service）

- [x] 2.1 新增 `api/v1`（或现行 App 约定）`GET`/`PUT` `/device/app/api/appointment/next` 请求响应结构与 `g.Meta`
- [x] 2.2 实现读：无行返回 `nextAt=0`（`.One()`+`IsEmpty()`）；写：upsert，允许 `nextAt=0`；校验登录与 `deviceNo` 绑机一致
- [x] 2.3 controller 注册于 device 进程；中文业务注释齐全；**禁止** import history/voice 业务包或直查他域表

## 3. gateway-app 放行与统计

- [x] 3.1 `installDeviceProxyMiddleware` 增加 `/device/app/api/appointment/*` 反代 pattern
- [x] 3.2 确认路径**不**加入 Bearer 匿名白名单
- [x] 3.3 **向负责人确认**预约 GET/PUT 是否计入 usage；记录结论后再决定是否改 `maintenance_skip.go`（未确认不得改）——结论：**不计入**，已加入 skip
- [x] 3.4 自检：领域路由 + 反代前缀 + auth + usage 结论 + `g.Meta` summary（apiregistry）

## 4. 回归与边界

- [x] 4.1 确认未改动 `predict_imminent` / pending HTTP 校验与消费逻辑
- [x] 4.2 冒烟：代码路径自检（绑机校验 / One+IsEmpty / upsert 0 / admin 勾选 / options 字段）；运行时联调待部署后验证
- [x] 4.3 （联调备忘，本仓可不改代码）Flutter：预约卡片补下次时间、清空时 pending 移除该事件、写预约表与 pending 两步一致——本工作区无 Flutter 孪生仓；契约见 design，客户端另仓实现
