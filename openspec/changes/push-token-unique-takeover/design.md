## Context

`push_device`（库 `ai_voice_push`）当前唯一键为 `(wx_id, device_key, channel)`，注册路径 `RegisterPushDevice` 仅在该键上 upsert `token`。同一物理设备换登录账号时，厂商 token 常不变，于是出现多行同 token、不同 `wx_id`。预测临近等对「同宝宝全部绑定 wx」扇出时，同一 token 被发送多次 → 用户看到重复通知。

约束：跨域走 clients/契约；Admin 静态页挂 gateway `resource/public` + `admin-modules.js`；Admin API 宿主 push-service；Redis/背景 ticker 与本变更无关；中文注释；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 注册时 **token 字符串全局唯一**；冲突则删除旧行，当前登录用户后来顶上。
- 表级 `UNIQUE(token)` + EnsureSchema 存量去重，防止并发双插。
- 运维 Hub 可分页查看 `push_device` 并按行删除。

**Non-Goals:**

- 不按 `device_key` / `channel` 做跨 wx 接管（产品明确只要 token）。
- 不限制「同一用户多手机」（token 不同则各自保留）。
- 不改 predict 扇出、不改 App register/unregister 路径字段、不做试推。

## Decisions

### D1. 唯一维：仅 `token`（非 token+channel）

- **选择**：`UNIQUE KEY uk_token (token)`；比较与删除均按 trim 后的 token 全文。
- **理由**：产品「一部手机一个 token」；跨 channel 撞串极罕见，全局唯一更简单。
- **备选**：`UNIQUE(token, channel)` — 已否决。

### D2. 注册顺序：先删冲突再 upsert

```
Register(wxB, channel, token T, deviceKey D):
  1. DELETE FROM push_device WHERE token = T
     （或 WHERE token = T AND NOT (即将被 upsert 命中的同键行) —— 实现可「先删全部同 token 再 Save」，同账号重注册亦安全）
  2. Save/upsert (wx_id=B, device_key=D, channel) SET token=T, updated_at=now
```

- 保留现有 `(wx_id, device_key, channel)` upsert，兼容同账号同机刷新 token。
- 同账号两台手机：token 不同 → 两行并存（符合预期）。

### D3. 存量去重再加索引

EnsureSchema：

1. 若尚无 `uk_token`：删除重复 token 中非「最新」行（`updated_at` 最大，并列取 `id` 最大保留）。
2. `ALTER TABLE ... ADD UNIQUE KEY uk_token (token)`。
3. 保留既有 `uk_wx_device_channel`。

### D4. Admin API 与鉴权

- 路径：`GET /push/admin/api/devices`（分页 + 可选 `wxId` / `channel` / token 前缀检索）、`DELETE /push/admin/api/devices/{id}`。
- 宿主 push-service；校验 `X-Admin-Password`（与 cash/ucg 同模式）。
- gateway：`BindMiddleware("/push/admin/api/*")` 反代至 `PUSH_API_PROXY_URL`；`InjectAdminDownstreamPassword` 增加 `/push/admin/api/` 分支。
- 口令：`PUSH_ADMIN_PASSWORD`，若空则回退 `DeviceAdminPassword()`（降低新 env 摩擦；compose/.env.example 注明可单独配置）。

### D5. Admin UI

- 静态页：`resource/public/push-device-admin.html`，路由 `/device/admin/push-device-admin.html`。
- `AdminCommon.requireAdmin` + `adminFetch`；列表展示 id、wx_id、channel、device_key、token **截断**、updated_at；删除二次确认。
- `admin-modules.js` 增加入口「推送设备」；`admin_static_pages.go` / Bearer 白名单静态页列表同步登记。

### D6. usage 统计

- App `register`/`unregister` 已在 `maintenance_skip`，本变更不改。
- Admin API / 静态页属运维路径，不计入 App usage；**不**改 `maintenance_skip`（无需再问负责人：无新增对外 App 业务接口）。

### D7. GoFrame `OnDuplicate` 误用说明（增写失败根因）

先删再增两条独立 SQL 的产品顺序保持不变；关键是**增语句必须可执行**。

- **误用**：`OnDuplicate(g.Map{"token": token, "updated_at": now})`。GoFrame 将 Map 的 **value 当作 `VALUES()` 内的列名**（官方例：`{"nickname": "passport"}` → `nickname=VALUES(passport)`），不是要写入的业务值。
- **后果**：生成非法子句如 `` `token`=VALUES(`<整段厂商token>`) ``、`` `updated_at`=VALUES(`1736…`) ``。`Save` 无论是否撞唯一键都会附带该子句；MySQL 报 Unknown column 等错误 → **整句 INSERT 失败**。于是出现「① DELETE 已提交、② Save 失败 → 只删不插」。
- **正确**：传列名到列名，例如 `OnDuplicate("token", "updated_at")`，生成 `` `token`=VALUES(`token`), `updated_at`=VALUES(`updated_at`) ``。删光同 token 后：无 `(wx,device,channel)` 行则纯插入；该键上仍有其它 token 旧行则靠此子句更新。
- **非目标**：不因此改为「删插必须同事务」；并发撞 `uk_token` 仍可返回可观测错误并由客户端重试。

## Risks / Trade-offs

- [并发双注册同 token] → 唯一索引 + 注册先删后写；撞唯一则返回可观测错误，客户端可重试。
- [误删其它用户 token] → 仅当 token 字符串完全相同；符合「一机一 token」产品假设。
- [Admin 口令回退 device] → 部署文档写明；需要隔离时设 `PUSH_ADMIN_PASSWORD`。
- [列表暴露 token] → UI 截断；完整 token 仅服务端持有，删除按 id。
- [OnDuplicate 误传业务值] → 见 D7；已改为列名形式，避免删后增失败空窗。

## Migration Plan

1. 发版 push-service：EnsureSchema 去重 + uk_token + 注册接管逻辑。
2. 发版 gateway：反代 `/push/admin/api/*`、口令注入、静态页与 Hub 入口。
3. 验证：同机换号 register 后库中仅一行该 token；预测扇出对旧 wx `no_device`。
4. 回滚：回退镜像；唯一索引可保留（有益无害）；Admin 入口可留空页或回退静态资源。

## Open Questions

- （无阻塞）若后续要在 Admin 展示「该 wx 绑了哪个宝宝」，需经 device 契约联查，本期不做。
