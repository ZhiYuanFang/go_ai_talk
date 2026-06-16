## Context

- **现状**（v2.0.3 + `add-pangbao-ai-clinic-room` 增量）：AI 月度额度三 feature（`polish`、`voice_ai`、`clinic_ai`）由 **device-service** 在 `ai_voice_device` 库统一持久化（`ai_quota_default` / `ai_quota_user_override`），Redis 用量键 `ai:usage:{feature}:{wxId}:{YYYYMM}` 亦由 device 进程读写。ucg/voice 经 HTTP internal 调用 device check/consume；ucg-admin 经 ucg 转发 device admin API 管理三字段；App 经 `GET /device/app/api/ai-quota` 聚合读取。
- **问题**：违反「一服务一数据库、各域自管数据」——voice 与 ucg 的配额配置不应落在 device 库；ucg-admin 不应管理 voice/clinic 字段。
- **约束**：voice 禁止跨库直查 device/ucg 表；ucg 禁止跨库直查 device/voice 表；40301/40302 语义不变；禁止新增测试文件与背景 ticker；Redis 用量键沿用既有模式（proposal 已记录负责人确认沿用）。
- **前置**：`add-pangbao-ai-clinic-room` clinic 运行时与 device 侧临时 clinic_ai 实现须已落地；本变更在其上重构配额层。

## Goals / Non-Goals

**Goals:**

- device-service **完全退出** AI 配额（表、DAO、internal/app API、gateway 反代路径）。
- voice-service 在 **`ai_voice_voice`**（`VOICE_DB_LINK` / `database.default`）自管 `voice_ai` + `clinic_ai` 配置与用量。
- ucg-service 在 **`ai_voice_ucg`**（`UCG_DB_LINK` / `database.default`）自管 `polish` 配置与用量。
- App 读 API **分域暴露**，gateway **不做** BFF 聚合。
- gateway-app 注册 `/voice/app/api/*`、`/voice/admin/api/*` HTTP 反代；voice-admin HTML + Hub 导航。
- ucg-admin **polish only**；维护窗口一次性迁移，接受 15–30min 配额写入停机。

**Non-Goals:**

- 双写/灰度切流（用户已确认不需要）。
- gateway 聚合 `GET /device/app/api/ai-quota` 兼容层（App/Flutter 直接改分域 API）。
- 变更额度计费模型或自然月桶规则（仍 `Asia/Shanghai` YYYYMM）。
- 新增 Redis 读缓存层（配置仍 MySQL 权威，用量仍 Redis INCR）。

## Decisions

### 1. 数据模型与库归属

| Feature | 权威进程 | MySQL 库 | 表（建议名） | Redis 用量键 |
|---------|----------|----------|--------------|--------------|
| `voice_ai` | voice-service | `ai_voice_voice` | `ai_quota_default`, `ai_quota_user_override` | `ai:usage:voice_ai:{wxId}:{YYYYMM}` |
| `clinic_ai` | voice-service | 同上 | 同上（列 `voice_ai_monthly_limit`, `clinic_ai_monthly_limit`） | `ai:usage:clinic_ai:{wxId}:{YYYYMM}` |
| `polish` | ucg-service | `ai_voice_ucg` | `ai_quota_default`, `ai_quota_user_override` | `ai:usage:polish:{wxId}:{YYYYMM}` |

- **默认上限**：voice_ai=5、clinic_ai=30、polish=5（与现网一致）。
- **voice/ucg 各自 internal API**：本进程内 check/consume，**不再** HTTP 调 device。

### 2. API 边界

**voice-service（新增/迁移）**

- Internal（`X-Voice-Internal-Secret` 或沿用既有 voice internal 密钥约定）：
  - `POST /voice/internal/api/ai-quota/check` — body `{ wxId, feature: voice_ai|clinic_ai }`
  - `POST /voice/internal/api/ai-quota/consume` — 同上
- App：`GET /voice/app/api/ai-quota` → `{ voiceAi: {used,limit}, clinicAi: {used,limit} }`
- Admin：`GET/PUT /voice/admin/api/ai-quota/default`、`GET/PUT /voice/admin/api/ai-quota/user`

**ucg-service（本地化）**

- Internal：进程内 check/consume（可保留 package 级函数，**不**再转发 device）
- App：`GET /ucg/app/api/ai-quota` → `{ polish: {used,limit} }`
- Admin：`GET/PUT /ucg/admin/api/ai-quota/default|user` — **仅** `polishMonthlyLimit`

**device-service（删除）**

- 移除 `/device/internal/api/ai-quota/*`、`GET /device/app/api/ai-quota`、admin ai-quota 接口及 DAO。

**wxId-by-deviceNo 迁移**

- voice 在需由 deviceNo 反查 wxId 时，改调 device **user 域**已有 internal API（如 `GET /device/app/api/user/internal/by-id` 或等价 wxId 反查契约），**禁止**再经 ai-quota internal API 附带反查。

### 3. Gateway 反代

参照 `device_route_proxy.go` / `ucg_route_proxy.go` 模式，新增 `voice_route_proxy.go`：

| 路径前缀 | 目标 | 环境变量 |
|----------|------|----------|
| `/voice/app/api/*` | voice-service | `VOICE_API_PROXY_URL` + `VOICE_API_ROUTE_MODE` |
| `/voice/admin/api/*` | voice-service | 同上 |

- gateway-app **HookBeforeServe** 对 App 路径注入 Bearer 头；Admin 路径经 Admin JWT + `InjectAdminDownstreamPassword` 注入 `X-Admin-Password`。
- **扩展** `InjectAdminDownstreamPassword`：`/voice/admin/api/` → `VOICE_ADMIN_PASSWORD` env（GoFrame 配置 `voice.admin.password`）。
- gateway-service **SHOULD** 同步 voice HTTP 反代路径以对齐。
- **移除** gateway `device_route_proxy.go` 中 `/device/app/api/ai-quota` 登记。

### 4. Admin UI

- 新建 `resource/public/voice-admin.html`：「AI 配置」Tab — 全局 `voiceAiMonthlyLimit` + `clinicAiMonthlyLimit`；wxId override 两字段。
- `admin-modules.js` 增加 Hub 入口（id: `voice-admin`，pagePath: `/device/admin/voice-admin.html`）。
- `ucg-admin.html`：移除 voiceAi/clinicAi 字段与相关 JS。

### 5. voice/ucg 业务调用链

- voice `/voice/chat/ws`、`/voice/clinic/ws`：LLM 前本地 check voice_ai/clinic_ai；成功后本地 consume。
- ucg `POST /ucg/app/api/posts/polish`：本地 check/consume polish。
- 错误码：**40301** 未登录、**40302** 额度用尽（不变）。

### 6. Flutter 分域读取

- 拆分 `aiQuotaStatusProvider` → `voiceAiQuotaProvider`（`/voice/app/api/ai-quota`）+ `polishAiQuotaProvider`（`/ucg/app/api/ai-quota`）。
- UI 组件按 feature 订阅对应 provider；移除 `/device/app/api/ai-quota` 调用。

## Risks / Trade-offs

- **[Risk] 维护窗口内配额写入不可用** → 提前公告；窗口内禁止 AI 润笔/喂养/胖宝调用；只读展示可降级为「暂时不可用」。
- **[Risk] Redis 用量键迁移遗漏导致计数重置** → runbook 脚本按 wxId 批量 COPY 旧键到新键（feature 前缀不变，仅写入进程变更）；迁移前后抽样对账。
- **[Risk] MySQL 配置行迁移错误** → 自 device 表导出 INSERT 至 voice/ucg 对应表；voice 仅迁 voice_ai/clinic_ai 列，ucg 仅迁 polish 列。
- **[Risk] App 旧版仍调 `/device/app/api/ai-quota`** → 该路径返回 404/410；与 App 发版同步。
- **[Trade-off] 无 gateway BFF** → Flutter 两次 HTTP 请求；换取清晰域边界与独立演进。

## Migration Plan

1. **准备**：voice/ucg DDL 建表；voice/ucg 服务代码就绪但未切流量。
2. **维护窗口开始**：停止 device/voice/ucg 相关实例或开启维护页；备份 device ai_quota 表与 Redis `ai:usage:*`。
3. **数据迁移**：MySQL 行拆分写入 voice/ucg 库；Redis 键 verify（键名不变，归属进程变更）。
4. **部署**：先发 voice/ucg 新版本 → gateway 注册 voice 反代、移除 device ai-quota 反代 → 停 device ai-quota 代码路径。
5. **验证**：internal check/consume、App 分域读 API、admin 配置、40301/40302 回归。
6. **App 发版**：Flutter 分域 provider。
7. **回滚**：恢复 device 旧版本 + gateway ai-quota 反代（需窗口内 DB/Redis 备份回滚）；voice/ucg 新表可保留待下次窗口。

## Open Questions

- **App API usage 统计**：`GET /voice/app/api/ai-quota`、`GET /ucg/app/api/ai-quota` 是否计入运维统计 — **apply 前须向负责人确认**。
- voice internal API 密钥头名称是否与现有 voice internal 契约统一（实现时对齐 `manifest/config` 与 device internal 模式）。
