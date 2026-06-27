## 1. device-service：全量 sim wxId

- [x] 1.1 `sim_user.go` 实现 `ListAllSimulatedWxIds`（单 SQL、total>10000 报错）
- [x] 1.2 `api/v1/device_sim_internal_http.go` 增加 `DeviceSimWxIdsReq/Res` 与路由 `GET /device/internal/api/sim/wx/ids`
- [x] 1.3 `device_sim_internal.go` 注册 handler

## 2. ucg-service：未读会话抽样

- [x] 2.1 新增 `chat_sample_internal.go`：`SampleRandomSimUnreadChat(ctx, simWxIds, messageLimit)`（eligible JOIN + ID 锚点 + 消息 LIMIT）
- [x] 2.2 新增 `ucg_internal_chat_sample.go` handler：`POST /ucg/internal/api/chat/sim-unread-sample`
- [x] 2.3 在 ucg 路由注册 internal 路径（与 posts/sample 鉴权一致）

## 3. sim-user-service：T5 重写与 E1 删除

- [x] 3.1 `clients.go`：`listAllSimWxIds()`、`sampleSimUnreadChat(simWxIds)` 封装 device/ucg internal 调用
- [x] 3.2 `tasks.go`：重写 `RunChatScanTask`（ids → sample → LLM → sendInternalChat）；删除 `spawnEphemeralChat` 与 ephemeral 包级变量
- [x] 3.3 `task_llm.go`：`chat_scan` usage 改为「未读回复」
- [x] 3.4 `manual_run.go` / `scheduler_manager.go`：确认 T5 签名与调用无 ephemeral 依赖

## 4. runtime 与 Admin：移除 E1 配置

- [x] 4.1 `runtime_config.go` / `runtime.go` / `runtime_api.go` / `runtime_snapshot.go`：移除 ephemeral 字段与 env 映射
- [x] 4.2 `config_admin.go`：移除 ephemeral 比较、effects、`buildConfigEffects` 中 E1 文案
- [x] 4.3 `api/v1/sim_admin_http.go`：intervals 类型去掉 ephemeral 键（若 struct 显式声明）
- [x] 4.4 `resource/public/sim-admin.html`：移除 E1 表单、interval 标签与 runtime 只读展示

## 5. 验收

- [x] 5.1 `openspec validate sim-t5-unread-sample` 通过
- [x] 5.2 部署顺序文档化：device → ucg → sim-user；T5 手动执行：有未读回复 1 条、无未读 skip、无 E1 goroutine
