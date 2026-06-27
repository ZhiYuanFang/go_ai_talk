## 1. 运行时配置与数据结构

- [x] 1.1 `RuntimeConfigDB` / `RuntimeFlags`：新增 `PostVideoPollInterval`、`PostVideoPollMaxWait`；移除 `VideoPoll`、idle/active；旧 JSON 迁移（active→interval）
- [x] 1.2 `runtime_api.go`、`sim_admin_http.go`、`sim_admin_api.go`：DTO 与 PUT/GET 映射新字段；移除 P1 字段
- [x] 1.3 `config_admin.go`：reload diff 排除仅 poll 参数变更；`taskSchedule` 删除 P1 行

## 2. T4 内联轮询与 P1 删除

- [x] 2.1 `tasks.go`：实现 `spawnVideoPoll`（自 P1 迁入逻辑 + deadline）；全局 `videoPostInFlight`；T4 改流水线语义与 `RecordTaskRun` 时机
- [x] 2.2 手动执行路径：`isManualRun` 时同步等待 poll 结束；调度路径异步 goroutine
- [x] 2.3 `store.go`：启动清理 `DiscardPendingVideoJobsOnStartup`；`scheduler_manager.go` 调用并移除 P1 调度
- [x] 2.4 删除 `RunVideoPollTask`；`manual_run.go`、`task_llm.go` 移除 `video_poll`

## 3. Admin 管理页

- [x] 3.1 `sim-admin.html`：移除 P1 UI；新增 T4 轮询间隔/最大等待输入；runtime 只读区同步
- [x] 3.2 T4 手动执行：延长 status 轮询上限以覆盖 `postVideoPollMaxWait`；按钮「执行中…」至 lastRunAt 更新

## 4. 验收

- [x] 4.1 手动 T4：提交后轮询直至发帖或超时，任务表 success/fail 与 job 状态一致
- [x] 4.2 轮询进行中：调度 tick skip；重复手动执行返回忙
- [x] 4.3 重启后 pending job 为 skipped；Admin 保存 poll 参数在线生效且不误触发 scheduler reload
