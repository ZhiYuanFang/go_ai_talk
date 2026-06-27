## Why

当前 T5（`chat_scan`）随机选一个 sim 用户后扫描全部未读会话，并为每条真人未读 **spawn E1** 子 goroutine（5min 循环 × 15min 窗口），导致：单次 tick HTTP/LLM 扇出不可控、进程内存在 detached 循环任务、且 `sim/wx/list` 分页语义不适于「全库 sim 覆盖」。运营期望改为：**device 提供全量 sim wxId → ucg 一次 O(1) 抽样 eligible 未读会话 → 单次 LLM 回复**，仅依赖 T5 周期轮询，不连续对话。

## What Changes

- **device internal 新增** `GET /device/internal/api/sim/wx/ids`：返回**全部** `is_simulated=1` 的 wxId 列表（单条 SQL、无分页截断；有安全上限）；T5 **MUST NOT** 用 `sim/wx/list` 第一页代替全集。
- **ucg internal 新增** `POST /ucg/internal/api/chat/sim-unread-sample`：请求体含 `simWxIds`；在 eligible 集合（sim 侧 `unread_count>0` 且 peer **不在** `simWxIds`）上有界 ID 锚点随机返回 **0 或 1** 条会话，并附带最近消息供 LLM；MUST NOT `ORDER BY RAND()`，MUST NOT 循环重试 sim 用户。
- **T5 重写**：`RunChatScanTask` 编排 device ids → ucg sample → `aimodel` `chat_reply` → `sendInternalChat`；每 tick **最多 1 条**回复；无 eligible 时记 skip（非假 success）。
- **删除 E1**：移除 `spawnEphemeralChat`、`ephemeralMu`/`ephemeralActive` 及全部 `ephemeralChatLoop`/`ephemeralChatWindow` 配置（runtime_json、env、Admin 表单、effects、scheduler 日志）。
- **不新增** Redis 读缓存；**不新增** `*_test.go`；**不新增** App HTTP 接口。

## Capabilities

### New Capabilities

- `ucg-sim-chat-unread-sample`：ucg internal 模拟用户未读会话抽样 API（含最近消息、真人 peer 过滤、有界 SQL）。

### Modified Capabilities

- `device-sim-user`：新增全量 sim wxId internal 列表 API；明确 list 分页 MUST NOT 作为 T5 全集来源。
- `sim-user-service`：T5 改为 poll-reply 流水线；删除 E1 与子 goroutine；`task_llm` usage 标签更新。
- `sim-runtime-config`：移除 `ephemeralChatLoop`/`ephemeralChatWindow` 及相关 env 默认值。
- `sim-user-admin`：Admin API 与 `sim-admin.html` 移除 E1 字段与 `ephemeral_may_continue` effect。

## Impact

- **代码**：`internal/services/device/sim_user.go`、`internal/controller/device_sim_internal.go`、`api/v1/device_sim_internal_http.go`；`internal/services/ucg/`（新 sample 服务 + internal handler）；`internal/services/simuser/`（tasks、clients、runtime*、config_admin、scheduler_manager、task_llm、manual_run）；`resource/public/sim-admin.html`、`api/v1/sim_admin_http.go`。
- **进程**：**device-service** → **ucg-service** → **sim-user-service**（部署顺序）；静态页经 gateway-app。
- **DB**：无表结构迁移；只读 `wx` 与 `ucg_conversation_member`/`ucg_chat_message`。
- **OpenSpec 基线**：对照 v2.0.5；本变更 delta 挂于 `sim-t5-unread-sample`。
- **App usage 统计**：无新增 App HTTP 接口，无需确认。
