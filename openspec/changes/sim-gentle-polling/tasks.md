## 1. ucg internal 轻量帖抽样 API

- [x] 1.1 新增 `api/v1/ucg_internal_posts_sample_http.go` 契约：`POST /ucg/internal/api/posts/sample`（limit 默认 20、上限 50）
- [x] 1.2 实现 `internal/services/ucg/post_sample_internal.go`：单 SQL 查 `ucg_post`（published）+ 可选首图 `ucg_post_media`；禁止 `postsFromResult`/device 调用
- [x] 1.3 新增 `internal/controller/ucg_internal_posts_sample.go` handler + internal 密钥校验
- [x] 1.4 在 `register_ucg_service.go` 注册路由；补充中文 summary

## 2. sim 周期 env 化与启动错峰

- [x] 2.1 扩展 `LoadRuntimeFlags`：接入 `SIM_INTERVAL_*`、`SIM_EPHEMERAL_CHAT_*`、`SIM_STARTUP_STAGGER_MAX`（复用 `envDuration`）
- [x] 2.2 `scheduler.go`：各 `runPeriodic` 首次 tick 前增加 `randomStartupDelay()`（0–staggerMax 均匀随机）
- [x] 2.3 文档化默认值与现网写死周期对齐（T3=3h30m 等）

## 3. P1 自适应轮询

- [x] 3.1 实现 `runAdaptivePeriodic`：tick 结束后按 `ListPendingVideoJobs` 是否为空选择 idle/active 下一间隔
- [x] 3.2 `scheduler.go` 将 P1 从固定 `runPeriodic` 改为 `runAdaptivePeriodic`；env `SIM_INTERVAL_VIDEO_POLL_IDLE`/`ACTIVE`
- [x] 3.3 确认无 pending 时不调智谱、不发 UCG 帖

## 4. E1 降频

- [x] 4.1 `spawnEphemeralChat` 使用 `RuntimeFlags` 的 loop/window 配置（默认 5m/15m）
- [x] 4.2 硬停与 `(sim,peer)` 去重语义保持不变

## 5. HTTP 全局限速

- [x] 5.1 在 `clients.go` 为 gateway App API（`appGet`/`appPost`/`appPut`）接入 `rate.Limiter`；env `SIM_UCG_RATE_LIMIT_RPS`（默认 2）、burst 4
- [x] 5.2 `ucgInternalPost`（sample、chat/send）可选同限或独立配置；首期与 App API 共用 limiter 即可
- [x] 5.3 队列满时阻塞等待，记录 Debug 日志（不丢请求）

## 6. T2 改走 sample API

- [x] 6.1 `simuser/clients.go` 新增 `ucgInternalPost` 调用 sample 端点
- [x] 6.2 `RunCommentTask` 移除 `feed/recommend`；改为 sample → 随机选帖 → simVision → POST comments
- [x] 6.3 无样本帖时 `RecordTaskRun` 记失败语义「无已发布帖」，与现网「无推荐帖」对齐

## 7. 配置与文档

- [x] 7.1 更新 `manifest/docker/.env.example`：全部新 `SIM_*` 变量与推荐生产配方注释
- [x] 7.2 更新 `docs/runbooks/release-deploy-and-run.md`：「长期开 sim + 共享 MySQL」观测指标与 env 配方
- [x] 7.3 运行 `openspec validate sim-gentle-polling --strict` 并通过
- [x] 7.4 `docker-compose.microservices.yml` 透传全部 `SIM_INTERVAL_*` / `SIM_TASK_*` / `SIM_STARTUP_STAGGER_MAX` / 限速 env 至 sim-user-service 容器

## 8. 验收（手工）

- [ ] 8.1 生产/测试栈：`SIM_USER_SERVICE_ENABLED=true` 启动后各任务首次执行时间错峰；`Threads_running` 低于变更前峰值
- [ ] 8.2 T2 日志与 access 路径无 `feed/recommend`；sample internal 返回帖可成功评论
- [ ] 8.3 无 pending video job 时 P1 约 10m 才唤醒；注入 pending 后约 2m 轮询
