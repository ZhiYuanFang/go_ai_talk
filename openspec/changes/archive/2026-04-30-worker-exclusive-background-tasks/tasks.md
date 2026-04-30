## 1. 启动职责收敛

- [x] 1.1 从 `gateway-service` 启动路径移除 `StartBackgroundWorkers` 调用。
- [x] 1.2 保持 `worker-service` 启动路径继续调用 `StartBackgroundWorkers`，并确认依赖检查顺序不变。
- [x] 1.3 检查并补充必要注释，明确“后台任务仅由 worker 启动”的角色约束。

## 2. 配置与文档对齐

- [x] 2.1 校验并更新 compose/kustomize 中 gateway 与 worker 的 MQ/outbox 开关说明与默认值。
- [x] 2.2 更新运行文档，明确 gateway 无业务后台任务职责、worker 独占异步任务执行。

## 3. 验证与回归

- [ ] 3.1 验证 gateway 运行时不再出现后台任务消费/中继日志。
- [ ] 3.2 验证 worker 可正常消费 `voice.task.requested.q` 并执行 outbox relay。
- [x] 3.3 验证重复启动入口不会导致同进程内重复拉起后台 goroutine。
