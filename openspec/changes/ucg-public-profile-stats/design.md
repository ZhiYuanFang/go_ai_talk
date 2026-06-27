## Context

- `enrichProfileStats` 已对 `profile/me` 路径生效：3 条有界 COUNT（`ucg_follow` ×2、`ucg_post` ×1）。
- `GetPublicProfilesByWxIDs` 用于 Feed/列表 author 填充，**刻意不含** count（避免关注后整页快照失效）；本 change 不扩展批量路径。

## Decisions

- 仅在 `GetPublicProfile` 单条公开读路径补 `enrichProfileStats`，与 `mergeProfileForAuthor` 统计来源一致。
- 不引入 Redis 读缓存（个人主页低频，与项目 Redis 读缓存约定一致）。

## Risks

- 每次他人主页打开多 3 条 COUNT；可接受，与 `profile/me` 相同。
