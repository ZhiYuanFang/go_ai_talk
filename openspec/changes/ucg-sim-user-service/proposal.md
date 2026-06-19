## Why

UCG 广场在早期真实用户稀少，推荐 Feed、评论与互动密度不足，影响新用户留存与产品观感。需要一套**在生产环境运行**的模拟用户系统，按固定节奏自动注册宝妈 persona 账号、发帖、评论、关注与回复私聊，使广场呈现真实活跃状态；同时模拟行为不得污染 App API 使用统计，且 AI 调用须与现有微服务共用 Redis 并发闸门。

## What Changes

- 新增独立进程 **`sim-user-service`**，承载 6 个周期任务、1 个视频轮询任务，以及按需触发的「真人聊天 30 分钟临时窗口」。
- **device-service**：`wx` 表新增 `is_simulated`；新增 internal **`POST /device/internal/api/sim/username/register`**（方案 A：复用 `WxUsernameRegister` 并原子标记模拟用户）；新增 internal 列表/批量查询 sim 用户契约。
- **ucg-service**：新增 internal **`POST /ucg/internal/api/chat/send`**，仅允许 `is_simulated=1` 的发送方，复用 `ProcessOutboundChatMessage`（正常 Green 审核）。
- **aimodel**：扩展 CogView 生图、CogVideoX 视频提交/轮询；新增 sim 相关 Lane，共用按 model 的 Redis 闸门。
- **gateway-app**：usage 统计写入前跳过 `is_simulated=1` 的 wxId；静态页 **`/device/admin/sim-admin.html`** 与 sim Admin API（Prompt 模板、模拟用户上限、总开关）。
- **Docker/配置**：`cmd/sim-user-service`、独立 config、compose 服务项；测试与生产 overlay 可开关 `SIM_USER_SERVICE_ENABLED`。
- 模拟用户账号 `ptest{N}` 递增（用户不可见），默认上限 100（管理页可改）；模拟用户对真人可见；内容正常走 Green，无 fast-track。
- 每隔 7h 随机 sim 关注另一 sim；AI 队列满/超时则跳过本次任务，不额外处理 polish 月度额度。

## Capabilities

### New Capabilities

- `sim-user-service`：模拟用户微服务宿主、周期任务调度、视频 job 轮询、真人聊天临时窗口、gateway/ucg/device HTTP 客户端编排。
- `sim-user-admin`：模拟管理 Web 与 Admin API（Prompt 可配置、`maxSimUsers`、总开关、任务观测字段）。
- `device-sim-user`：wx 模拟用户标识、internal 注册/列表/批量 is_simulated 契约。
- `ucg-sim-chat-internal`：UCG internal 发消息契约（替代 sim 侧 WebSocket）。
- `aimodel-media-gen`：aimodel 包扩展生图/生视频及 sim Lane profile。

### Modified Capabilities

- `gateway-app-api-usage-stats`：成功 App HTTP 调用统计 MUST 排除 `is_simulated=1` 的 wxId（负责人已确认按用户维度全量跳过）。

## Impact

- **新进程**：`cmd/sim-user-service`；`manifest/config/config.sim-user-service.yaml`；`manifest/docker` compose 与 prod/test overlay。
- **数据库**：device 库 `wx.is_simulated` 迁移；sim 服务自有表（config、prompt、video_job、序号计数）。
- **Redis**：沿用 `ai:llm:gate:{model}:*`；可选 `usage:sim_wx_ids` 缓存集；sim 服务不新增业务读缓存（除闸门与 usage 辅助集）。
- **外部 API**：智谱 GLM-4.7-Flash、GLM-4.6V-Flash、CogView-3-Flash、CogVideoX-Flash（`/paas/v4/videos/generations`）。
- **OpenSpec 强制项**：6 周期 ticker + 1min 视频轮询须在 design 中声明宿主、周期、开关与失败语义；usage 统计策略已确认；不新增 `*_test.go`。
- **风险（已接受）**：生产环境真人可能与模拟用户互动并收到 push；Green 拒审可能导致任务成功但内容未上线；模拟 AI 与 voice/ucg 竞争 LLM 槽位。
