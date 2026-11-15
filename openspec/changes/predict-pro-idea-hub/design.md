## Context

Flutter 预测墙 Pro 档收集「想法」（非设置反馈）。device 域已有 JWT、Admin 口令、`user.baby_name`、`feedback-records` 页模式。基线：`openspec/specs/v3.1.9/spec.md`。探索补充：gateway 须显式绑 App 子路径；usage 计入；任意登录用户可用；日限额用表 COUNT。

## Goals / Non-Goals

**Goals:**

- 持久化想法；App 可 list/submit/like；列表含宝宝名、热度、是否已赞、是否本人。
- Admin 可分页查看（wxId + 宝宝名）、改展示文案、删除。
- 列表默认按 `created_at` 降序（新在前）。
- gateway 反代 / Admin 静态 / usage 计入齐备。

**Non-Goals:**

- Flutter UI；catalog description 编辑；采纳标记 / 自动发 Pro；按热度排序；新微服务。
- 服务端校验预测 Pro 开通态（不调 cash / 不查 `feature_user_entitlement`）。
- 为本能力新增 Redis 读缓存或日限额键。

## Decisions

1. **域：device-service**（wx + Admin + babyName 同构 feedback）。

2. **表**  
   - `predict_pro_idea`：`id`、`wx_id`、`device_no`、`baby_name`、`content`、`display_text`、`created_at`、`updated_at`、`deleted_at`；可选缓存 `like_count`（若缓存则同事务维护，否则 list 以 like 表 COUNT 为准）。  
   - `predict_pro_idea_like`：`idea_id`、`wx_id`、`created_at`，UNIQUE(`idea_id`,`wx_id`)。  
   - EnsureSchema 幂等；库 `ai_voice_device` / `default` / `DEVICE_DB_LINK`。

3. **路径**  
   - App：`/device/app/api/predict-pro-ideas/list|submit|like|unlike`  
   - Admin：`/device/admin/api/predict-pro-ideas/list|update|delete`

4. **gateway-app**  
   - `device_route_proxy.go`：`BindMiddleware("/device/app/api/predict-pro-ideas/*", …)`。  
   - `admin_static_pages.go`：`/device/admin/predict-pro-ideas` → `predict-pro-ideas.html`。  
   - `gateway_app_auth_exempt.go`：仅登记 Admin **HTML** 路径；App API **禁止** exempt。  
   - usage：**计入**；不改 `maintenance_skip.go`。

5. **门禁（已定）**：JWT 即可 list/like/unlike；submit 另须绑定 `deviceNo`。**MUST NOT** 查 cash Pro 权益。

6. **可见性**：提交即公开；Admin 改 `display_text`；软删下线。无审核队列。

7. **宝宝名**：submit 时从绑定 `deviceNo` 的 `user.baby_name` 快照；无绑定或无宝宝名拒绝提交。App list **下发** `babyName`；**不下发** wxId。

8. **list 项**：`{ id, text, babyName, likeCount, likedByMe, isMine }`；排序 `created_at DESC`。

9. **点赞**：like 插入、unlike 删除；幂等；返回 `{ likeCount, likedByMe }`。

10. **日限额（已定）**：默认 10；窗口为 Asia/Shanghai **自然日**；实现为对 `predict_pro_idea` 按 `wx_id` + 当日 `created_at` 范围 **COUNT**（含软删与否：建议 **含软删计入配额**，防删了再刷）。**不得**用 Redis。

11. **采纳权益**：本版服务端不建模；无自动履约。

12. **Admin UI**：`predict-pro-ideas.html` + `/device/admin/predict-pro-ideas`；`admin.html` 卡片；加载 `admin-pages.css`，近全宽；宽屏列表+编辑宜左右分栏。

## Risks / Trade-offs

- [刷赞/刷想法] → 日限额 COUNT + Admin 删除；like 唯一键防重复。  
- [宝宝名公开] → 产品需要「{名}的家长」；不下发 wxId。  
- [like_count 缓存不一致] → 优先 like 表 COUNT，或同事务维护缓存。  
- [未绑 Pro 也可发想法] → 产品已选；依赖 Flutter 入口与运营删除。

## Migration Plan

1. 部署 `device-service`（建表）+ `gateway-app-server`（反代/静态/exempt）。  
2. 验 App 401/未绑定、list 空、submit、like、usage 有计数。  
3. 再发 Flutter。  
回滚：停路由/入口；表可留。

## Open Questions

- （已定）日限额默认 10 + 表按日 COUNT；任意登录可提交/点赞；usage **计入**；无宝宝名拒绝提交；排序新在前；采纳仅客户端文案。  
- App list 分页：本版可不做；若量暴涨再另案加 `limit`/`page`。
