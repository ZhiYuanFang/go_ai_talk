## 1. 后端 taskSchedule 生效语义

- [x] 1.1 `TaskScheduleItem` / `SimAdminTaskScheduleDTO` 增加 `configEnabled`；`Enabled` 改为 effective（task && db && service）
- [x] 1.2 `buildTaskSchedule` 传入 `dbEnabled`、`serviceEnabled`；`nextRunHint` 总闸关时返回阻塞说明
- [x] 1.3 `buildConfigEffects`：DB 关或 env 关且存在配置开任务时追加 effects 提示
- [x] 1.4 `sim_admin_api.go` 映射新字段

## 2. Admin 保存结果 UI

- [x] 2.1 `sim-admin.html` `renderSaveResult`：表头区分「配置 / 自动调度」；`enabled=false` 时不误导为已运行

## 3. 验收

- [x] 3.1 DB 关 + T4 配置开保存：API 返回 `configEnabled=true, enabled=false`，hint 含业务总闸
- [x] 3.2 env 关 + 全开保存：对应任务 `enabled=false`，hint 含进程总闸
- [x] 3.3 三层全开：行为与变更前 effective 路径一致（enabled=true，正常 nextRunHint）
