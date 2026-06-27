## Why

T6 当前仅 sim 互关（`pickTwoDistinctSimWx`），无法在广场侧让模拟用户关注真人博主，与 T2（评论任意帖）、T5（仅回复真人）的开放度不一致。运营期望 sim 通过关注有内容的真人用户提升互动图谱真实感。

## What Changes

- **ucg internal `posts/sample`**：响应增加 `authorWxId`；请求增加可选 `excludeAuthorWxIds`（sim-user 传入全量 sim wxId），random/latest 抽样 SQL MUST 排除这些作者。
- **T6 重写**：随机 sim 登录 → `posts/sample` random 取真人 author → `POST follow/{authorWxId}`；**完全删除** sim→sim 关注路径与 `pickTwoDistinctSimWx`。
- **自关注防护**：若 `authorWxId == followerWxId`，重试抽样（有界次数），仍失败则记 task 失败。
- **不新增** admin/runtime 配置项、device 新接口、Redis、App HTTP、`*_test.go`。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-sim-feed-sample`：`posts/sample` 增加 `authorWxId` 与 `excludeAuthorWxIds` 过滤语义。
- `sim-user-service`：T6 从 sim→sim 改为 sim→真人 author（发过 published 帖）。

## Impact

- **代码**：`internal/services/ucg/post_sample_internal.go`、`internal/controller/ucg_internal_posts_sample.go`、`api/v1/ucg_internal_posts_sample_http.go`；`internal/services/simuser/tasks.go`、`clients.go`。
- **部署**：ucg-service → sim-user-service。
- **DB**：只读 `ucg_post`；无迁移。
