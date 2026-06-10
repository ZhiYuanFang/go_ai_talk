## Context

- **AI 润笔**：`POST /ucg/app/api/posts/polish`（ucg-service）→ DashScope；已要求 `X-Internal-Wx-Id > 0`（`wxIDFromUcgHeader`）。
- **语音喂养 AI**：`GET /voice/chat/ws`（voice-service）→ DeepSeek/TTS；HTTP 层 Bearer 豁免但可注入 `X-Internal-Wx-Id`；当前按 deviceNo 分钟级 Redis 限流，无月度 wxId 额度。
- **用户域**：wxId 权威在 device-service（`wx` 表）；ucg/voice MUST NOT 跨库直查额度表。
- **管理端**：`ucg-admin.html` 已有「AI 配置」Tab（模型、图片数），口令 `ucg.admin.password`；device 管理页使用 `X-Admin-Password`（device admin 口令）。
- **探索结论**：两功能**独立**月度额度；全局默认各 5、可独立调整（如润笔 10、喂养 5）；per-wxId override；wxId=0 禁止 AI；**成功返回后扣减**；月界 `Asia/Shanghai` 自然月。

## Goals / Non-Goals

**Goals:**

- device-service 作为额度权威：全局默认、per-wxId override、Redis 月度用量。
- ucg 润笔、voice 喂养 LLM 调用前预检（不扣减），AI 成功后扣减。
- App 可读当前 used/limit；超额返回稳定错误供 Flutter 弹框「本月额度已用完」。
- ucg-admin「AI 配置」Tab 扩展：全局默认 + wxId override（单口令登录）。

**Non-Goals:**

- `/voice/asr/ws` 纯听写、模式切换、规则回复、LLM 失败兜底计入额度。
- 与 gateway-app-api-usage-stats 合并或复用 Redis key。
- 操作审计、导出、Prometheus 指标。
- flutter_ai_talk 仓库内具体 UI 实现（本变更仅定义 API/错误契约；Flutter 任务单列说明）。

## Decisions

### 1. 额度权威与存储

- **决定**：device-service 独占 `ai_quota_default`（singleton id=1）、`ai_quota_user_override`（`wx_id` UNIQUE）；Redis 键 `ai:usage:{feature}:{wxId}:{YYYYMM}`，feature 为 `polish` | `voice_ai`；TTL 90 天；月桶按 `time.Now().In(Asia/Shanghai).Format("200601")`。
- **有效上限**：`override` 非 NULL 字段优先，否则取 global default；两 feature 独立。
- **理由**：wxId 在用户域；与 AGENTS.md 服务边界一致。
- **备选**：ucg 库存用量 — voice 无法合规读取。

### 2. Internal 契约（ucg / voice 调用）

- **决定**：
  - `POST /device/internal/api/ai-quota/check` body `{ wxId, feature }` → `{ allowed, used, limit }`
  - `POST /device/internal/api/ai-quota/consume` body `{ wxId, feature }` → `{ used, limit }` 或超额错误
- **鉴权**：Header `X-Device-Internal-Secret`（与现有 device internal 一致）。
- **consume 语义**：服务端 `INCR` 后若 `used > limit` 则 `DECR` 回滚并返回超额（防止成功路径并发穿透）；check 仅读。
- **理由**：「成功才扣」= 业务侧先 check → 调 AI → 成功则 consume；失败不调 consume。
- **备选**：check+consume 原子一步 — 无法在 AI 失败时免扣。

### 3. 预检与扣减落点

| 功能 | 预检 (check) | 扣减 (consume) |
|------|--------------|----------------|
| 润笔 | `PostsPolish` 调 DashScope 前 | `PolishPostText` 成功返回正文后 |
| 喂养 AI | `chatWithResult` / 流式 casual **即将调 LLM 前** | 母婴 DeepSeek 成功 / casual 流式完整成功返回后 |

- **wxId=0**：润笔已有 header 拒绝；voice 在 LLM 前若 wxId≤0（Header 或 device internal `wx-id-by-device-no`）返回 WS 错误帧 `code=40301`（或等价）、message 引导登录，**非**额度文案。
- **不计次**：模式切换、待补量词、纯规则回复、LLM 错误兜底、ASR-only。

### 4. App 读 API

- **决定**：`GET /device/app/api/ai-quota`（JWT + `X-Internal-Wx-Id`）→ `{ polish: { used, limit }, voiceAi: { used, limit } }`；wxId≤0 返回 401。
- **理由**：Flutter 可选展示剩余次数；非强制。

### 5. Admin 写读与 ucg-admin 单口令

- **决定**：
  - device-service 实现 internal admin 逻辑（读/写 default、读/写 user override）。
  - **ucg-service** 暴露 `GET/PUT /ucg/admin/api/ai-quota/default`、`GET/PUT /ucg/admin/api/ai-quota/user?wxId=`，校验 `ucg.admin.password` 后 **HTTP 转发** device internal admin 路径（复用 `deviceServiceUrl` + `deviceInternalSecret`）。
  - `ucg-admin.html` 仅调用 `/ucg/admin/api/ai-quota/*`。
- **理由**：运营在 UCG 管理页一次登录；device 仍为权威，ucg 不缓存额度。
- **备选**：ucg-admin 直接调 `/device/admin/api/*` — 需第二套口令，体验差。

### 6. 超额与 Flutter 错误契约

- **决定**：HTTP API 超额使用业务码 **40302**（或 gcode 等价），message **「本月额度已用完」**（固定中文）。wxId 无效 **40301**，message **「请先登录账号」**。
- **Voice WS**：超额 JSON 帧 `{ type: "error", code: 40302, message: "本月额度已用完" }`。
- **理由**：Flutter 按 code 分支，避免解析英文。

### 7. voice 解析 wxId

- **决定**：优先 WS 握手请求头 `X-Internal-Wx-Id`（gateway 注入）；缺失或 ≤0 时调用 device internal `device-no → wxId`（已有或新增轻量接口）；仍 ≤0 则 40301。
- **理由**：设备会话可能未带 wx header；与 wxId=0 禁止 AI 一致。

### 8. 全局默认初始值

- **决定**：DB migration/seed：`polish_monthly_limit=5`，`voice_ai_monthly_limit=5`；yaml 可同名 fallback。
- **运营可调**：两字段独立 PUT，允许润笔 10、喂养 5。

## Risks / Trade-offs

- **[Risk] check 与 consume 之间并发导致略超额度** → consume 用 INCR+比较+回滚；可接受软额度。
- **[Risk] voice WS 未带 Bearer 导致 wxId 缺失** → deviceNo 反查；反查失败 40301。
- **[Risk] ucg→device internal 不可用** → 润笔 fail-closed（返回 503，不调 DashScope）；voice 同理不调 LLM。
- **[Risk] 上海时区与服务器 UTC 配置** → 代码显式 `LoadLocation("Asia/Shanghai")`，不依赖 OS 默认。
- **[Trade-off] 成功才扣但 check 后 AI 前仍占「逻辑名额」** → 极端并发可能多调一次 AI；成本可控。

## Migration Plan

1. device 库建表 + seed default；部署 device-service（internal + app + internal-admin API）。
2. 部署 ucg-service（admin 代理 + 润笔 check/consume）；更新 `ucg-admin.html`。
3. 部署 voice-service（LLM 前 check、成功后 consume）；配置 `DEVICE_SERVICE_URL`。
4. Flutter 识别 40302 弹框（独立仓库发版）。
5. 回滚：还原服务版本；Redis 键自然过期；DB 表可保留。

## Open Questions

（无。探索阶段决策已闭合。）
