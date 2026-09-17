## Why

同一部手机切换多个登录账号、且都绑定同一宝宝时，厂商推送 `token` 往往不变，但 `push_device` 现仅按 `(wx_id, device_key, channel)` 唯一，导致**同一 token 挂在多个 wx 下**。预测临近等按宝宝绑定账号扇出时，会对同一物理设备连发多条通知。需要「一部手机一个 token」：注册时 token 全局唯一，冲突则后来顶上。运维同时需要可查看/手动清理 `push_device` 的管理入口。

## What Changes

- **注册接管**：`POST /app/api/push/register` 写入前，若库中已存在相同 `token`（任意 `wx_id` / `channel` / `device_key`），MUST 删除既有行，再为当前登录用户 upsert；语义为后来顶上。
- **库约束**：`push_device` 增加 **`token` 全局唯一索引**；存量重复 token 在加索引前按「保留最新 `updated_at`（并列取更大 `id`）」清理。
- **不改**：同一用户多台手机（token 不同）各自保留；预测扇出「同宝宝全部绑定 wx」逻辑不变；App 注册/注销 path 与 usage 排除策略不变。
- **Admin**：运维 Hub 新增入口与静态页，分页展示 `push_device`，支持按条件筛选与按行手动删除；API 宿主 **push-service**，经 gateway 反代与 Admin 鉴权。

## Capabilities

### New Capabilities

- `push-token-uniqueness`：注册时 token 全局唯一与跨 wx 后来顶上；表级 UNIQUE(token) 与存量去重。
- `push-device-admin`：Admin 列表/删除 `push_device`、静态页与 Hub 入口、gateway 反代与鉴权。

### Modified Capabilities

- （无主线 `openspec/specs/` 下独立 push capability 需 delta；行为增量以本变更 New Capabilities 为准。既有 change `extract-push-service` 中注册语义被本变更收紧。）

## Impact

- **进程**：`push-service`（注册逻辑、schema、Admin API）；`gateway-app-server`（`/push/admin/api/*` 反代、Admin 口令注入、静态页与 `admin-modules` 入口）。
- **库**：`ai_voice_push.push_device` 增 `UNIQUE(token)`；启动 EnsureSchema 迁移清理重复。
- **客户端**：注册契约字段不变；换号登录后再次 register 即可接管 token（预期既有行为）。
- **非目标**：不按 `device_key` 跨 wx 顶掉；不改 predict 扇出算法；不做试推/批量导入。
