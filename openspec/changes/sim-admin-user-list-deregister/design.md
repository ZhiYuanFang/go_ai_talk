## Context

- 现有 `sim-admin.html` 仅展示 `simUserCount/maxSimUsers` 聚合值；device 已有 `GET /device/internal/api/sim/wx/list`（wxId + account），ucg 已有 `GetPublicProfilesByWxIDs` 服务函数但未暴露 internal HTTP。
- `wx.password` 存 bcrypt，无法反查明文；`wx` 表无 `created_at`。T1 使用 `NextAccountName()` → `ptest{N}` + 全局 `defaultPassword`（默认 `123456`），T2–T6 用同一密码登录随机 sim 用户。
- 服务边界：sim-user-service 不得 import device/ucg DAO；编排经 internal HTTP + 自有 `SIM_DB_LINK` 表。

## Goals / Non-Goals

**Goals:**

- sim-admin 单页展示分页 sim 用户列表（头像、昵称、账号、wxId、注册时间、密码明文、注销）。
- 注销语义与 App 一致：仅删除 device `wx` 行（须 `is_simulated=1`）。
- T1 起使用随机账号/密码，并在 sim 库持久化明文供 admin 与 T2–T6 登录。
- 最少 DB 操作：列表一页一次 device list + 一次 ucg batch + 一次 sim credential 批量查询。

**Non-Goals:**

- 不级联删除 UCG profile、帖子、评论、私信、关注关系。
- 不给 `wx` 表加 `created_at` 或明文密码列。
- 不对历史 `ptest*` 用户 backfill credential（仅 fallback 展示与登录）。
- 不新增 App 面向客户端 API；不新增测试文件。

## Decisions

### 1. 凭据存储：`sim_wx_credential`（sim-user DB）

| 字段 | 说明 |
|------|------|
| `wx_id` | PK，对应 device wx.id |
| `account` | 冗余，便于日志与列表 |
| `password_plain` | 注册时写入，仅 admin API 与 sim 任务读 |
| `created_at` | Unix 秒，作为 admin「注册时间」 |

**理由**：明文不应进入 device `wx` 表（也服务真人账号）；bcrypt 不可逆，必须在注册当下落库。

**替代方案**：加密存 sim 库 —— 运维页仍需解密展示，复杂度高于收益，首期明文 + admin 鉴权即可。

### 2. 随机账号/密码生成（T1）

- **账号**：`crypto/rand` 生成满足 `^[a-z0-9_]{4,32}$` 的字符串（建议前缀 `s` + 10–15 位 `[a-z0-9_]`）；碰撞 `ErrWxUsernameTaken` 时重试（上限如 8 次）。
- **密码**：12–16 位随机可打印 ASCII，满足 6–64 长度。
- **停用**：`NextAccountName` / `sim_account_seq` 不再参与 T1（表保留，不删）。

**理由**：消除可预测 `ptest{N}`；每用户独立密码后 T2–T6 必须 per-wxId 查密。

### 3. T2–T6 登录

- `randomSimSession` 流程：`sim/wx/random` → 按 `wxId` 查 `sim_wx_credential.password_plain`；无行则 fallback `LoadRuntimeFlags().DefaultPassword` 或 `"123456"`。
- 手动任务与 scheduler 共用同一 helper（如 `simLoginByWxPick`）。

### 4. Admin 列表 API 编排（sim-user-service）

`GET /sim/admin/api/users?page=&pageSize=`：

1. `device/internal/sim/wx/list`
2. `ucg/internal/profiles/batch`（wxIds）
3. `sim_wx_credential` WHERE wx_id IN (...)
4. 合并 DTO；历史用户：`passwordPlain` 为空时响应 `"123456"` + `passwordPlainLegacy: true`（或等价字段供 UI 标注）；`createdAt` 缺失为 0 / null，UI 显示「—」。

鉴权：与现有 sim-admin 一致（gateway 注入 `X-Admin-Password`）。

### 5. 注销 API 编排

`POST /sim/admin/api/users/{wxId}/deactivate`：

1. `device/internal/sim/wx/{wxId}/deactivate` — 服务端校验 `is_simulated=1`，调用 `WxDeactivateByID`，并从 `usage:sim_wx_ids` SREM（经 `cachekit`）。
2. sim-user：DELETE `sim_wx_credential` WHERE wx_id；UPDATE `sim_video_job` pending/processing → `skipped` WHERE wx_id。
3. 失败语义：非 sim / 不存在 → 4xx 业务错误；与 App deactivate 对齐。

**不**调用 ucg 删 profile。

### 6. ucg internal batch profiles

`POST /ucg/internal/api/profiles/batch`，body `{ "wxIds": [int64] }`，响应 `{ "list": [{ wxId, nickname, avatarUrl, avatarThumbnailUrl }] }`。

实现：复用 `GetPublicProfilesByWxIDs`；无 profile 行的 wxId 不出现在 list（与现有公开 profile 语义一致）。

### 7. UI（sim-admin.html）

- 在「当前生效快照」与「Prompt 模板」之间插入 `<h4>模拟用户</h4>` + 表格 + 分页控件。
- 列：头像（thumbnail）、昵称、账号、wxId、注册时间（`toLocaleString`）、密码（`<code>` + 历史标注）、注销按钮（`confirm`）。
- 注销成功后刷新列表 + `loadRuntime()` 更新计数。

### 8. API 版本

- 新增 sim-admin 路由为**新端点**，不修改既有 `/sim/admin/api/config|status|runtime` 响应结构（符合「旧版本不可改结构」约定）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 明文密码存 sim 库 | 仅 admin 鉴权 API 返回；禁止日志打印 body；Hub 内网运维入口 |
| 注销后 UCG 内容残留 | proposal 已明确 non-goal；列表注销后该行消失，广场内容仍可能存在 |
| 历史用户无 credential | fallback 默认密码登录 + UI 标注「默认密码（历史）」 |
| T6 等任务因错误密码失败 | credential 写入与 register 同事务顺序：先 device 注册成功再 insert credential，再 login |
| internal batch N+1 | ucg batch 单次 SQL WHERE IN；device list 已分页 |

## Migration Plan

1. 部署 ucg internal batch + device sim deactivate internal。
2. 部署 sim-user-service：`EnsureSchema` 建 `sim_wx_credential`；更新 T1/T2–T6 与 admin API。
3. 部署 gateway-app 更新 `sim-admin.html`（sim admin API 反代无需新路由）。
4. 无需数据迁移；旧 `ptest*` 继续可用至手动注销。
5. **回滚**：回退 sim-user 二进制与 HTML；credential 表可留空；旧代码仍用 defaultPassword 登录历史账号。

## Open Questions

- 无（历史用户密码展示已定为 fallback `123456` + 「默认密码（历史）」标注）。
