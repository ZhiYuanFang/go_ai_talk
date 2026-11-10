## 1. 数据模型与 schema

- [x] 1.1 device DB：`event` 表增加 `merge_predict TINYINT NOT NULL DEFAULT 0`；更新 entity / dao columns / do；JSON 字段 `mergePredict`
- [x] 1.2 新增 `EnsureEventMergePredictColumn`（幂等 ALTER），并在 device-service 启动 ensure 链中调用

## 2. 写路径与约束

- [x] 2.1 实现「有子的根」判定辅助；`AddEvent` / `UpdateEvent` 规范化 0/1，不满足条件强制写 0；根改挂为非根时清零
- [x] 2.2 `DeleteEvent` 成功后：若原父变为无子根则清零该父 `merge_predict` 并走既有 EventOptions 重建
- [x] 2.3 admin multipart / API 请求体支持 `mergePredict`；list 查询 fields 含新列；变更后重建 EventOptions
- [x] 2.4 同步扩 `contracts.DeviceAdmin`、`clients/device` 与 controller 入参（对齐 `isAppointment` 扩参模式）

## 3. Hub UI

- [x] 3.1 `admin.html` 事件列表增加「合并预测」列；行数据保留 `mergePredict`
- [x] 3.2 事件表单增加 checkbox；仅当前行为有子根时可勾，叶子与无子根 disabled；提交带 `mergePredict=0|1`

## 4. 验收与部署说明

- [x] 4.1 人工核对：options/list 返回含 `mergePredict`；非法勾选落库为 0；删最后子后父标志清零
- [x] 4.2 确认重部署清单：`device-service` + `gateway-app-server`；不必改 voice/history 业务逻辑；无新 usage / 无新 App 路由
