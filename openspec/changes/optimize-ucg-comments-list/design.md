## Context

- 评论存储于 `ai_voice_ucg.ucg_post_comment`，帖子 `comment_count` 在发表/删除时递增/递减（`AddComment` / `DeleteComment`）。
- 当前 `ListComments`：`COUNT(*)` + `LIMIT/OFFSET` 分页 + 循环 `GetPublicProfile`（每条 1 次 `ucg_profile` 查询 + 可选 device IP 属地）。
- 客户端（Flutter）为拉全评论对同一接口循环翻页，进入详情时串行多次 HTTP。
- 帖子作者 profile 与评论作者 profile 均走「实时」`ucg_profile` + device 补全（默认昵称等），与帖子列表一致；**不在评论行写入快照**。
- 仓库 Redis 读缓存政策：负责人已决策评论列表 **不使用 Redis**（见 `openspec/project.md`「Redis 读缓存约定」）。

## Goals / Non-Goals

**Goals:**

- 单次 HTTP 返回帖子下全部评论（升序 `created_at`，UI 底部为最新）
- 列表路径 profile 查询降为 **1 次批量 IN**（+ 已有 `IpLocationForWxIDs` 批量）
- 去掉评论列表专用 `COUNT(*)`，减少一次 DB 往返
- 客户端发帖评论后乐观追加 POST 响应，避免再次 GET 全列表
- 超长帖评论数保护：可配置硬上限（默认 500）+ `truncated`

**Non-Goals:**

- Redis 缓存评论列表或 profile
- `ucg_post_comment` 表新增 nickname/avatar 快照列
- 评论分页能力（历史 `page`/`pageSize` 不再作为列表契约）
- 改造点赞列表 `ListPostLikes` 的 N+1（可后续独立变更）
- WebSocket 推送新评论（仍依赖刷新或乐观 UI）

## Decisions

### 1. 改造现有 GET 评论接口（非新路径）

- **选择**：保持 `GET /ucg/app/api/posts/{id}/comments`，变更响应形状与查询语义。
- **理由**：gateway 反代与 App 路由已固定；Flutter 同步改造即可。
- **替代**：新增 `/comments/all` — 多一条注册表与双端维护成本，收益有限。

### 2. 响应 `{ list, total, truncated }`，移除 `page`/`pageSize`

- `total`：默认取帖子 `comment_count`（进入列表前通常已从帖子详情获得）；若未查帖或计数不可用，回退 `len(list)`。
- `truncated`：当 `comment_count > cap` 且只返回前 `cap` 条（按 ASC 取最早 cap 条）时为 `true`；此时 `total` 仍为 `comment_count`，提示 UI「仅展示前 N 条」。
- **BREAKING**：旧客户端依赖分页字段将失效，需 Flutter 同步发版。

### 3. 批量 profile：`GetPublicProfilesByWxIDs`

在 `profile.go` 新增批量函数（或同等内部方法）：

1. `dao.UcgProfile` `WHERE wx_id IN (...)` 一次查询
2. 对缺省昵称行按需 `refreshDefaultNicknameIfNeeded`（仅命中行，避免全表）
3. `IpLocationForWxIDs(ctx, wxIDs)` 一次批量（已有 device batch 契约）
4. `ListComments` 内组装 `map[uint64]*ProfileDTO`，循环评论行 O(1) 挂载 `author`

- **理由**：与单条 `GetPublicProfile` 语义一致，仅合并 IO。
- **替代**：评论表快照 — 已被负责人否决。

### 4. 去掉 ListComments 内 `COUNT(*)`

- 列表查询前 `ensurePublishedPost` 或轻量读取帖子行获取 `comment_count`（可与校验合并为一次 `ucg_post` 读取）。
- **理由**：`comment_count` 为列表展示权威计数；避免与 `COUNT(*)` 双写不一致时的额外查询（极端并发下计数与实数可能短暂偏差，与现网帖子计数语义一致）。

### 5. 硬上限默认 500，配置项可选

- 配置键建议：`ucg.comments.listMax`（`config.ucg-service.yaml`），默认 500，0 表示不限制（仅内测慎用）。
- 查询：`ORDER BY created_at ASC LIMIT cap`（cap>0 时）。

### 6. AddComment 不变更契约

- POST 仍返回完整 `UcgCommentItem`（含 `author`），Flutter 追加到本地 `list` 并递增本地 `commentCount`。
- 服务端仍单条 `GetPublicProfile`（发表频率低，可接受）。

### 7. Flutter 交互

- 进入详情：一次 `fetchComments(postId)`。
- 发表评论成功：将 POST 响应 append 到列表末尾（与 ASC 序一致），`commentCount++`，不调用 `fetchComments`。
- 删除评论（若已有）：本地 remove + `commentCount--`（本变更 tasks 可仅列评论列表优化，删除逻辑沿用）。

### 8. App API 使用统计（负责人已确认：不统计）

- **决策**：`GET /ucg/app/api/posts/{id}/comments` **不计入** gateway-app App API 使用统计（进入帖子详情时高频读路径，与 token 刷新/版本探测同类维护型排除）。
- **gateway 原始路径**：经 gateway-app 反代 UCG，`req.URL.Path` 形如 `/ucg/app/api/posts/<postId>/comments`（**非** `/device/app/api/...`）。
- **归一化 apiKey**（`apiregistry.Normalize`）：`GET /ucg/app/api/posts/{id}/comments`（登记于 `api/v1/ucg_app_http.go` `UcgPostCommentsGetReq`）。
- **实现位置**：`internal/services/gatewayapp/usagestats/maintenance_skip.go`。`isMaintenanceAPI` 在 `RecordHTTPRequest` 中于 apiregistry 归一化**之前**执行，依据原始 `method` + `path`。
- **匹配方式**：路径含动态 `{id}`，不可仅用 `maintenanceExactAPI` 字面量；MUST 增加 **GET-only** 匹配规则，使原始 path 满足 `/ucg/app/api/posts/<numericPostId>/comments` 时跳过统计（可参考 `gateway_app_auth_exempt.go` 中 posts 路径分段校验）。**POST** `POST /ucg/app/api/posts/{id}/comments`（发表评论）**仍计入**统计。
- **排除项摘要**：`GET /ucg/app/api/posts/{id}/comments` → 不统计。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 单帖评论数极大时单次响应体过大 | 默认 cap=500 + `truncated`；监控 P99 响应大小 |
| `comment_count` 与实数短暂不一致 | 与现网点赞/评论计数语义一致；UI 以 `total` 显示计数、以 `list` 展示内容 |
| API BREAKING 旧 App | 与 Flutter 同版本发布；文档标注废弃分页参数 |
| 批量 profile 缺 profile 行 | 与单条逻辑一致：无行则 `author` 省略或按现有错误语义 |
| usage 统计语义 | 负责人已确认 GET 评论列表不统计；实现 `maintenance_skip.go` GET-only 排除，POST 发表评论仍统计 |

## Migration Plan

1. 部署 ucg-service 与 gateway-app（含 `maintenance_skip.go` 排除项）新版本
2. 同步发布 Flutter（单次 GET + 乐观 append）
3. 无 DB 迁移；回滚仅需还原服务与客户端

## Open Questions

（无）
