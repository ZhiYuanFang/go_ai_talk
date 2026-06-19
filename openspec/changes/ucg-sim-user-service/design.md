## Context

- UCG 广场早期真实用户少，需在生产环境注入模拟宝妈 persona 的注册、发帖、评论、关注与私聊回复，使 Feed 与互动可见。
- 现有架构：`gateway-app` 签发 JWT；`device-service` 管 `wx` 与 username 注册；`ucg-service` 管动态/评论/关注/聊天（出站仅 WebSocket）；`aimodel` 包提供按 **model** 的 Redis 闸门（`ai:llm:gate:{model}:*`），已有 `LanePolish`（`glm-4.6v-flash`）、voice lanes（`glm-4.7-flash` 等）。
- 仓库约定：跨服务 MUST 走 HTTP 契约；禁止 sim 进程直查 device/ucg DAO；新增 ticker 背景任务 MUST OpenSpec 批准；usage 统计变更须负责人确认（本变更：**凡 `is_simulated=1` 的 wxId 全量跳过**）。

## Goals / Non-Goals

**Goals:**

- 新增 `sim-user-service` 独立进程，6 个周期任务 + 1min 视频轮询 + 按需 30min 真人聊天窗口。
- device internal 注册模拟用户（方案 A），`wx.is_simulated=1`；账号 `ptest{N}` 递增，默认上限 100（管理页可改）。
- UCG internal 发消息契约；模拟内容正常 Green 审核；模拟用户对真人可见。
- 扩展 `aimodel`：CogView 生图、CogVideoX 提交/轮询；sim Lane 共用 Redis 闸门。
- 模拟管理 Web：Prompt 模板、`maxSimUsers`、总开关可配置。
- gateway usage 统计排除模拟 wxId。

**Non-Goals:**

- 不为模拟用户 bypass Green 或推荐算法。
- 不在 App 客户端标注「机器人」。
- sim 服务不维护 WebSocket 长连接。
- 视频生成失败不重试；不实现 sim↔sim 私聊主动回复（除 T5 触发真人窗口外）。
- 不新增 `*_test.go`；任务周期间隔首期写死在代码/env（管理页仅改 prompt 与上限）。

## Decisions

### 1. 进程边界：`cmd/sim-user-service`

独立微服务，配置 `manifest/config/config.sim-user-service.yaml`，compose 新增服务项。依赖：`GATEWAY_APP_URL`、`DEVICE_SERVICE_URL`、`UCG_SERVICE_URL`、Redis、`GLM_API_KEY`、内部密钥 `DEVICE_GATEWAY_INTERNAL_SECRET`（与现有 internal 一致）。

**备选**：嵌入 ucg-service → 否决（职责混杂 + ucg 已有大量 MQ consumer）。

### 2. 注册：device internal 方案 A

```
POST /device/internal/api/sim/username/register
Header: X-Device-Gateway-Internal-Secret
Body: { account, password }
→ WxUsernameRegister + UPDATE wx SET is_simulated=1（同事务或注册后立即更新）
→ { wxId, account }
```

- 账号由 sim 服务分配 `ptest{N}`（查 sim 本地序号表，N 递增）。
- 达 `maxSimUsers`（默认 100）时 T1 跳过。
- 公开 `username/register` 不变，无法自行标记 sim。

**备选**：公开 register + 后置 mark → 否决（竞态 + 安全风险）。

### 3. 模拟用户身份与 usage 剔除

- `wx` 表新增 `is_simulated TINYINT NOT NULL DEFAULT 0`。
- device internal 提供 `POST /device/internal/api/sim/wx/list`（分页 wxId 列表）与 batch 字段 `isSimulated`（扩展现有 `wx/batch` 或 sim 专用 batch）。
- gateway `usagestats.RecordAsync`：若 `wxId > 0` 且 `IsSimulatedWx(wxId)` 则 return。
- 缓存：`Redis SET usage:sim_wx_ids` + 注册/变更时 SADD；miss 时 device internal 回源。进程内短 TTL map 兜底。

### 4. App 路径编排（sim 服务作为 HTTP 客户端）

典型 flow 使用 **gateway-app** 聚合登录与 UCG App API（带 Bearer），device internal 仅用于注册/列表/is_simulated：

| 任务 | 主要 HTTP 路径 |
|------|----------------|
| T1 注册 | device internal register → gateway username_login → aimodel 昵称/头像 → UCG profile/media → PUT profile |
| T2 评论 | login → GET feed/recommend → aimodel vision 评论 → POST comments |
| T3 图文 | login → aimodel text+image → media presign/upload/register → POST posts submit=true |
| T4 视频 | login → aimodel text → SubmitVideoGeneration → INSERT sim_video_job |
| T5 聊天 | login → GET conversations → 真人未读则 spawn E1 或 LLM 回复 → ucg internal chat/send |
| T6 关注 | login → pick 2 sim wxId → POST follow |
| P1 轮询 | GET zhipu video status → 成功则 OSS + POST posts |

### 5. UCG internal 发消息

```
POST /ucg/internal/api/chat/send
Header: X-Device-Gateway-Internal-Secret
Body: { senderWxId, conversationId, clientMsgId, content, imageKey?, videoKey? }
```

- MUST 校验 sender `is_simulated=1`（经 device internal）。
- 调用 `ProcessOutboundChatMessage`（Green + push 真人 peer 不变）。
- 403 若 sender 非 sim；404 非会话成员。

### 6. aimodel 扩展（方案 A）

新增 Lane 常量（sim 进程注册 `ProfileStore`）：

| Lane | 默认 model | 用途 |
|------|------------|------|
| `simText` | glm-4.7-flash | 昵称/文案/聊天（与 voiceUnderstanding **同 Redis 池**） |
| `simVision` | glm-4.6v-flash | 读帖评论（与 polish 同池） |
| `simImageGen` | cogview-3-flash | 头像/配图 |
| `simVideoGen` | cogvideox-flash | 视频提交（Acquire 仅 POST 时） |

新增 API：

- `GenerateImage(ctx, lane, prompt) (imageURL or []byte, err)`
- `SubmitVideoGeneration(ctx, lane, prompt) (taskID, err)`
- `PollVideoGeneration(ctx, taskID) (status, videoURL, err)` — **轮询不占 inflight 槽**

Provider：智谱 OpenAI 兼容 + `/paas/v4/videos/generations`；API Key 仍 env `GLM_API_KEY`。

### 7. 背景任务（OpenSpec 批准清单）

| 任务 ID | 宿主 | 周期 | 开关 | 失败语义 |
|---------|------|------|------|----------|
| T1 register | sim-user-service | 24h ± jitter | `SIM_TASK_REGISTER_ENABLED` | 达上限跳过；AI/注册失败打日志，下次周期重试 |
| T2 comment | sim-user-service | 6h ± jitter | `SIM_TASK_COMMENT_ENABLED` | 无帖/队列满/审核前失败则跳过 |
| T3 post_image | sim-user-service | 3.5h ± jitter | `SIM_TASK_POST_IMAGE_ENABLED` | 同上 |
| T4 post_video_submit | sim-user-service | 6.5h ± jitter | `SIM_TASK_POST_VIDEO_ENABLED` | 提交失败跳过；成功写 job 表 |
| T5 chat_scan | sim-user-service | 1h ± jitter | `SIM_TASK_CHAT_ENABLED` | 无未读跳过；真人未读可 spawn E1 |
| T6 follow | sim-user-service | 7h ± jitter | `SIM_TASK_FOLLOW_ENABLED` | 已关注幂等跳过 |
| P1 video_poll | sim-user-service | 1min | `SIM_VIDEO_POLL_ENABLED` | failed → job=skipped；success → 发帖 |
| E1 ephemeral_chat | sim-user-service | 按需 30min×1min | 随 T5/检测触发 | 到期硬停；同 (sim,peer) 最多 1 窗口 |

总开关：`SIM_USER_SERVICE_ENABLED=false` 时进程不启动 ticker。

**AI 额度**：sim 直连 `aimodel`，不消耗 UCG polish 月度额度；`ErrQueueFull`/超时 → 当前 tick 提前结束，不阻塞其他任务 goroutine。

### 8. 视频 job 状态机

表 `sim_video_job`：`id, wx_id, content, task_id, status(pending|processing|done|skipped|failed), created_at, updated_at`。

- T4：同一 sim **最多 1 个 pending/processing job**。
- P1：每分钟 `PollVideoGeneration`；`success` → 下载 → UCG media 链路 → `CreatePost(submit=true)` → done；`failed` → skipped（不重试）。

### 9. 真人聊天临时窗口 E1

触发：T5 或会话扫描发现 sim 用户 S 对真人 P 有 `unreadCount>0`。

```
spawn EphemeralChatWindow(S, P):
  deadline = now + 30min
  loop every 1min until deadline:
    if S 在会话(S,P) unread>0:
      拉 messages → aimodel simText → ucg internal chat/send
  // 不续期；不影响 T2–T6
```

同 `(S,P)` 用 sync.Map 去重，防止重复 goroutine。

### 10. 模拟管理 Admin

- 静态页 `/device/admin/sim-admin.html`（gateway-app `admin_static_pages.go`）。
- sim-service Admin API（`X-Admin-Password` 或复用 Hub JWT + gateway 反代，design 实现期与 ucg-admin 对齐）：  
  - `GET/PUT /sim/admin/api/config` — `enabled`, `maxSimUsers`  
  - `GET/PUT /sim/admin/api/prompts/{taskType}` — 各任务 Prompt 模板  
  - `GET /sim/admin/api/status` — 各任务上次运行、计数、pending video jobs  
- Prompt 存 sim DB `sim_prompt(task_type, system_prompt, user_prompt_template, updated_at)`；运行时渲染 `{{post_content}}` 等变量。

### 11. sim 服务数据存储

sim 进程自有 MySQL 库（或独立 schema）：`sim_config`、`sim_prompt`、`sim_account_seq`、`sim_video_job`、`sim_task_run_log`（可选）。**禁止**存 device/ucg 业务表副本。

## Risks / Trade-offs

- **[Risk] 生产真人误认 sim 为真实宝妈** → 已接受；不在 UI 标注。
- **[Risk] Green 拒审导致广场仍空** → 任务日志 + admin status 可观测；不 fast-track。
- **[Risk] push 骚扰真人** → 已接受；与正常 UGC 互动一致。
- **[Risk] LLM 槽位竞争** → jitter + 队列满跳过；与 voice/ucg 共用池。
- **[Risk] 100 sim × 多任务 LLM 成本** → maxSimUsers 管理页可调低；总开关可关。
- **[Risk] internal 密钥泄露可批量造 sim** → 网络隔离 + 密钥轮换；仅 internal 注册。

## Migration Plan

1. 部署 device 迁移 `wx.is_simulated` + internal API（向后兼容，默认 0）。
2. 部署 ucg internal chat/send + aimodel 扩展。
3. 部署 gateway usage 剔除 + sim-admin 静态页路由。
4. 部署 sim-user-service，`SIM_USER_SERVICE_ENABLED=false` 启动验证健康检查。
5. 测试环境开启开关，观察 24h；生产逐步开启，先 `maxSimUsers=10` 再调高。
6. 回滚：关 `SIM_USER_SERVICE_ENABLED`；已注册 sim 用户保留（`is_simulated=1` 仍不计 usage）。

## Open Questions

- sim Admin API 鉴权最终对齐 ucg-admin（`X-Admin-Password`）还是 Hub JWT — 实现期与现有 admin 页一致即可。
- CogView 返回 URL 的下载超时与最大图片字节 — 实现期按 UCG media 上限配置。
