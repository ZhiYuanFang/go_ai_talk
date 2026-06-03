## Context

`go_ai_talk` 已按微服务拆分：`history-service`（`:9801`，`cmd/history-service/main.go` + `config.history-service.yaml`）、device-service、gateway-app-server（`RegisterGatewayAppHTTP`，`internal/controller/gateway_app_register.go`）。网关通过 `installDomainProxyMiddlewares`（`domain_route_proxy.go`）反向代理 `/device/history/api/*` 等，Bearer JWT 校验后注入 `X-Internal-Wx-Union-Id` / `X-Internal-Wx-Id`。

UCG 为全新域，须：

- 独立进程 `ucg-service`（~`:9804`）与库 `ai_voice_ucg`；
- **禁止**跨库读 `wx`（`service-boundary-no-cross-db`）；
- **ucg-service 仅作内部微服务**；App 对外 **HTTP + WebSocket 全部经 gateway-app-server 同域暴露**（与 `/device/app/api/*`、`/device/app/ws/history` 一致，无独立 UCG 公网网关或直连域名）。

## Goals / Non-Goals

**Goals:**

- 脚手架对齐 `history-service`：专属配置、依赖检查、单 default DB。
- 完整 DDL（见下文）与 REST + WS API。
- gateway-app 代理 `/ucg/app/api/*`（HTTP）与 `/ucg/app/ws/chat`（WebSocket 升级透传 → ucg-service `/ws/chat`）。
- OSS presign（bucket `pang-bao`，prefix `social/`）+ CDN `https://resorce.cuplay.top/{objectKey}`。
- Green 异步审核：帖子/资料 Option A；聊天 Option C。
- Redis 聊天消息永久保留；推荐混合排序。
- device-service internal API：wx validate/batch、baby_name。

**Non-Goals:**

- 拉黑/举报、群聊、管理后台审核台（MVP 仅 Green 自动审核）。
- 为 UCG 单独部署对外网关或让 App 直连 ucg-service 公网地址。
- 跨服务 DB join。

## Decisions

### 1. 服务脚手架

**Decision**：`cmd/ucg-service/main.go` 镜像 history：`GF_GCFG_FILE=manifest/config/config.ucg-service.yaml`；`UCG_SERVICE_ADDR` 默认 `:9804`；`UCG_DB_LINK` 覆盖 `GF_DATABASE_DEFAULT_LINK`；`runtimecheck.CheckDependencies` 含 MySQL + Redis。

**Register**：`controller.RegisterUcgServiceHTTP(s)` 绑定 `/ucg/app/api/*` 与内部 `/ws/chat`（仅集群内可达；App 连接 gateway 对外路径 `/ucg/app/ws/chat`）。

### 2. 身份与 device 契约

**Decision**：App HTTP 经 gateway 注入 `X-Internal-Wx-Id`（与现有 JWT `sub` 一致）。ucg-service internal 调 device：

| 内部接口 | 用途 |
|----------|------|
| `POST /device/internal/api/ucg/wx/validate` | 校验 wxId 存在 |
| `POST /device/internal/api/ucg/wx/batch` | 批量 nickname/avatar 展示字段 |
| `GET /device/internal/api/ucg/wx/{wxId}/baby-name` | 默认昵称 `{babyName}的家长` |

Header：`X-Device-Gateway-Internal-Secret: ${DEVICE_GATEWAY_INTERNAL_SECRET}`。

**Why**：wx 表归属 device；ucg 仅存 `wx_id` 外键语义。

### 3. Gateway 代理（HTTP + WebSocket）

**HTTP Decision**：`ucg_route_proxy.go` 注册 middleware：`/ucg/app/api/*` → `UCG_SERVICE_BASE_URL`（env，如 `http://ucg-service:9804`）；复用 Bearer + 注入 `X-Internal-Wx-Id`；CORS 与 history 代理一致（`gateway-app-cors-reverse-proxy`）。

**WebSocket Decision**：gateway-app 注册 `/ucg/app/ws/chat` WebSocket 升级代理至 ucg-service `/ws/chat`，实现模式参考 `internal/controller/ws_route_proxy.go`（`installVoiceWSProxyMiddleware` + `httputil.ReverseProxy`）：

| 对外（App） | 对内（ucg-service） |
|-------------|---------------------|
| `GET /ucg/app/ws/chat`（Upgrade） | `GET /ws/chat`（Upgrade） |

- 环境变量：`UCG_WS_ROUTE_MODE=proxy`（默认 proxy）；`UCG_WS_PROXY_URL` 指向 ucg-service WS 入口（如 `http://ucg-service:9804/ws/chat` 或等价 ws/wss 基址）。
- Bearer 白名单：`/ucg/app/ws/chat` 加入 `gatewayAppAuthExemptExactGET`（与 `/voice/chat/ws`、`/device/app/ws/history` 一致）；JWT 认证在连接后首帧 JSON `auth` 由 ucg-service 处理（不经 gateway 解析首帧）。
- 配置非法或下游不可达时返回结构化 `ws_proxy` 错误，与 voice WS 透传一致。

**Why**：App 仅需 `apiBaseUrl` 同 host；运维 TLS 终止在 gateway 层；ucg-service 不暴露公网 WS 端口。

### 4. OSS

**Decision**：

| 项 | 值 |
|----|-----|
| Bucket | `pang-bao` |
| Region | `cn-beijing` |
| Endpoint | `oss-cn-beijing.aliyuncs.com` |
| ObjectKey 前缀 | `social/` |
| CDN 展示 | `https://resorce.cuplay.top/{objectKey}` |
| AccessKey ID | `LTAI5t6tomJZp4im2H32FSMT` |
| AccessKey Secret | `LVCECT4exrGkkhI85HmyD4P2e6wJZW` |

- ObjectKey 格式：`social/{yyyy}/{mm}/{uuid}.{ext}`
- Presign：`POST /ucg/app/api/media/presign` 返回 uploadUrl、objectKey、headers
- 凭证：**写入** `manifest/config/config.ucg-service.yaml`（用户明确要求入库；生产可用 env 覆盖）

DB **仅**存 objectKey；响应 DTO 可附带 `cdnUrl` 便于客户端校验。

### 5. 帖子状态机（Option A）

```
draft(0) --submit--> pending_audit(1) --Green pass--> published(2)
                              \--Green fail--> rejected(3)
```

- `pending_audit` / `rejected`：仅作者在 feed/我的动态可见。
- `published`：关注/推荐/public profile 可见。
- 视频：同流程；Green 支持 video URL 审核。

Worker：进程内 goroutine 或独立 worker 轮询 `pending_audit` 行，调 Green API，更新 status + reject_reason。

### 6. 聊天审核（Option C）

- 客户端经 gateway 连接 `/ucg/app/ws/chat`，WS 发送 → 服务端写 Redis（key 设计见下）+ DB 可选 archive，状态 `pending`。
- 仅回 ACK 给发送方（含 clientMsgId）。
- Green pass → 标记 `delivered`，WS push 给收件方；更新 conversation unread。
- Green fail → 仅发送方 WS `audit_failed` + reason。

### 7. Redis 聊天模型

**Decision**（示例键名）：

- `ucg:chat:conv:{convId}:msgs` — Redis List 或 Stream，元素 JSON 消息体，**无 TTL**（永久保留）。
- `ucg:chat:user:{wxId}:conversations` — ZSET，score=lastMsgTime。
- `ucg:chat:conv:{convId}:unread:{wxId}` — INT 未读数。

可选 MySQL `ucg_chat_message_archive` 异步落库（非 MVP 阻塞项）。

### 8. 推荐算法（MVP）

**Decision**：定时任务计算 score 写入 `ucg_post_recommend`（可选表）或 Redis ZSET：

```
score = w_new * exp(-age_hours / τ) + w_like * log(1+likes) + w_comment * log(1+comments)
```

默认 `w_new=1.0, τ=72, w_like=0.3, w_comment=0.5`；分页按 score DESC + post_id tie-break。

关注 Feed：`published` 且 author 在 follow 集合，按 `published_at DESC`。

### 9. 分页

**Decision**：统一 query `page`（从 1）、`pageSize`（默认 20，最大 50）；响应 `{ list, total, page, pageSize }`。

## 数据模型（DDL）

```sql
CREATE DATABASE IF NOT EXISTS ai_voice_ucg DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_voice_ucg;

CREATE TABLE ucg_profile (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  wx_id         BIGINT UNSIGNED NOT NULL COMMENT 'device wx.id',
  nickname      VARCHAR(64)     NOT NULL DEFAULT '',
  avatar_key    VARCHAR(256)    DEFAULT NULL COMMENT 'OSS objectKey only',
  bio           VARCHAR(512)    DEFAULT NULL,
  created_at    BIGINT          NOT NULL COMMENT 'unix seconds',
  updated_at    BIGINT          NOT NULL,
  UNIQUE KEY uk_wx_id (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_post (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  author_wx_id  BIGINT UNSIGNED NOT NULL,
  content       TEXT,
  status        TINYINT         NOT NULL DEFAULT 0 COMMENT '0 draft 1 pending_audit 2 published 3 rejected',
  reject_reason VARCHAR(512)    DEFAULT NULL,
  media_type    TINYINT         NOT NULL DEFAULT 0 COMMENT '0 none 1 images 2 video',
  like_count    INT UNSIGNED    NOT NULL DEFAULT 0,
  comment_count INT UNSIGNED    NOT NULL DEFAULT 0,
  created_at    BIGINT          NOT NULL,
  updated_at    BIGINT          NOT NULL,
  published_at  BIGINT          DEFAULT NULL,
  KEY idx_author_status (author_wx_id, status, created_at),
  KEY idx_feed_published (status, published_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_post_media (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  post_id       BIGINT UNSIGNED NOT NULL,
  sort_order    INT             NOT NULL DEFAULT 0,
  object_key    VARCHAR(256)    NOT NULL,
  media_kind    TINYINT         NOT NULL COMMENT '1 image 2 video',
  duration_ms   INT             DEFAULT NULL,
  size_bytes    BIGINT          DEFAULT NULL,
  KEY idx_post (post_id, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_follow (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  follower_wx_id  BIGINT UNSIGNED NOT NULL,
  followee_wx_id  BIGINT UNSIGNED NOT NULL,
  created_at      BIGINT          NOT NULL,
  UNIQUE KEY uk_follow (follower_wx_id, followee_wx_id),
  KEY idx_followee (followee_wx_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_post_like (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  post_id    BIGINT UNSIGNED NOT NULL,
  wx_id      BIGINT UNSIGNED NOT NULL,
  created_at BIGINT          NOT NULL,
  UNIQUE KEY uk_post_wx (post_id, wx_id),
  KEY idx_wx (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_post_comment (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  post_id       BIGINT UNSIGNED NOT NULL,
  author_wx_id  BIGINT UNSIGNED NOT NULL,
  content       VARCHAR(1024)   NOT NULL,
  created_at    BIGINT          NOT NULL,
  KEY idx_post (post_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_conversation (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  type       TINYINT         NOT NULL DEFAULT 1 COMMENT '1 direct',
  created_at BIGINT          NOT NULL,
  updated_at BIGINT          NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_conversation_member (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  conversation_id  BIGINT UNSIGNED NOT NULL,
  wx_id            BIGINT UNSIGNED NOT NULL,
  pinned           TINYINT         NOT NULL DEFAULT 0,
  deleted_at       BIGINT          DEFAULT NULL COMMENT 'user soft delete',
  last_read_msg_id BIGINT UNSIGNED DEFAULT NULL,
  unread_count     INT UNSIGNED    NOT NULL DEFAULT 0,
  updated_at       BIGINT          NOT NULL COMMENT 'last activity; drives idx_wx_list sort',
  UNIQUE KEY uk_conv_wx (conversation_id, wx_id),
  KEY idx_wx_list (wx_id, pinned, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE ucg_post_recommend (
  post_id      BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  score        DOUBLE          NOT NULL,
  computed_at  BIGINT          NOT NULL,
  KEY idx_score (score DESC, post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## API 概要

### App 对外（经 gateway-app，与现有 App API 同域）

| 类型 | 对外路径 | 代理目标 |
|------|----------|----------|
| HTTP | `/ucg/app/api/*` | ucg-service 同路径 |
| WebSocket | `/ucg/app/ws/chat` | ucg-service `/ws/chat` |

**REST 示例**（前缀 `/ucg/app/api`）：

| Method | Path | 说明 |
|--------|------|------|
| GET | `/profile/me` | 当前用户 profile（自动创建默认昵称） |
| PUT | `/profile/me` | 更新 nickname/avatar_key/bio → pending Green |
| GET | `/profile/{wxId}` | 公开 profile + 帖子摘要 |
| GET | `/feed/recommend` | 推荐分页；匿名可选 |
| GET | `/feed/following` | 关注 Feed；需登录 |
| POST | `/posts` | 创建/提交（draft 或 pending_audit） |
| PUT | `/posts/{id}` | 编辑自己的帖 |
| DELETE | `/posts/{id}` | 删除自己的帖 |
| GET | `/posts/mine` | 我的动态（含全 status） |
| POST | `/media/presign` | OSS 直传签名 |
| POST | `/follow/{wxId}` | 关注 |
| DELETE | `/follow/{wxId}` | 取消关注 |
| GET | `/follow/following` | 我关注的人 |
| POST | `/posts/{id}/like` | 点赞 |
| DELETE | `/posts/{id}/like` | 取消赞 |
| GET | `/posts/{id}/comments` | 评论分页 |
| POST | `/posts/{id}/comments` | 发表评论 |
| DELETE | `/comments/{id}` | 删自己的评论 |
| GET | `/conversations` | 会话列表 |
| POST | `/conversations` | 创建/获取 1:1 会话 |
| GET | `/conversations/{id}/messages` | 历史消息分页（Redis+DB） |
| POST | `/conversations/{id}/read` | 标记已读 |

**WS** `GET /ucg/app/ws/chat`（gateway 对外；透传至 ucg-service `/ws/chat`）：auth 帧 JWT；事件 `message`、`message_ack`、`message_delivered`、`audit_failed`、`ping`/`pong`。

## Risks / Trade-offs

- **[Risk] Green 延迟导致作者长时间只见 pending** → UI 文案「审核中」；超时告警监控。
- **[Risk] Redis 永久聊天占内存** → 可选 archive + 冷数据策略 Phase 6；MVP 接受。
- **[Risk] OSS AK 入库泄露** → 用户明确要求；仍建议生产用 env 覆盖。
- **[Risk] WS 升级代理链路增加一跳延迟** → 与 voice WS 透传一致；MVP 可接受。
- **[Risk] 推荐刷榜** → MVP 无风控；Phase 6 tuning。

## Migration Plan

1. 创建 `ai_voice_ucg` 库执行 DDL。
2. 部署 ucg-service；gateway 增加 HTTP + WS 代理 env。
3. device-service 发布 internal API。
4. Flutter 分 Phase 对接（见 tasks.md）。
5. 回滚：gateway 移除 proxy 路由；App 隐藏 UCG 入口。

## Open Questions

- Green 视频审核是否走 OSS 临时 URL 还是帧截图 — 实现时按 Green SDK 能力选型。
