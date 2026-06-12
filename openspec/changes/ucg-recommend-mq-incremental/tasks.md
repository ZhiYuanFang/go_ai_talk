## 1. DDL 与 MQ 拓扑

- [x] 1.1 新增 `hack/sql/ucg_recommend_mq.sql`：`ucg_recommend_hot_scan_state`（`last_post_id`、`round_hot_cutoff`、`updated_at`）
- [x] 1.2 `eventkit/routing_keys.go` 注册六个 recommend routing key
- [x] 1.3 `hack/rabbitmq-init.ps1` / `.sh` 增加 `ucg.recommend.score.q` binding
- [x] 1.4 `config.ucg-service.yaml`：`hotWindowHours`、`hotScanPageSize`、`hotScanIntervalSeconds`、`likeThrottleMs`；废弃/移除全表 `refreshIntervalSeconds` 语义

## 2. eventkit 共用 AMQP connection

- [x] 2.1 Refactor `eventkit/amqp_consumer.go`：`SharedAMQPRunner`（单 connection、多 queue channel 注册、统一重连）
- [x] 2.2 迁移 audit consumer 注册到 SharedAMQPRunner（4 channel）
- [x] 2.3 ucg-service 启动时创建单 runner，audit + recommend 一并注册

## 3. 推荐分核心逻辑

- [x] 3.1 抽取 `RecomputeRecommendScore(postId)`、`RemoveRecommendScore(postId)`（复用 `computeRecommendScore`）
- [x] 3.2 实现 `recommend_publisher.go` HTTP Publish 六类事件
- [x] 3.3 实现 `recommend_mq_consumer.go`：单 key throttle（500ms/postId）；`unpublished` DELETE 0 行不报错且永远 Ack
- [x] 3.4 实现 `recommend_hot_reconciler.go`：轮首固定 `round_hot_cutoff`、分页扫热区、无互动也 Recompute

## 4. 业务 Publish 挂点

- [x] 4.1 `audit_post.go`：`publishPostCAS` → `published`；published 驳回 → `unpublished`
- [x] 4.2 `post.go` `DeletePost` → `unpublished`
- [x] 4.3 `post_admin.go` 管理端驳回（原 published）→ `unpublished`
- [x] 4.4 `social.go` like/unlike → `liked`/`unliked`
- [x] 4.5 `audit_comment.go` 过审 / 删评论 → `comment.published`/`removed`

## 5. 清理与验证

- [x] 5.1 删除 `RefreshRecommendScores` 全表 `All()` 与原 `StartRecommendWorker` 全表 ticker
- [x] 5.2 `go build ./...` 通过
- [x] 5.3 runbook 文档化：禁止全表、热区 reconciler、like throttle、双端口 Publisher/Consumer
- [x] 5.4 确认无新 gateway-app App API / usage 统计变更
- [x] 5.5 手动验收：发帖入推荐、like 风暴 throttle（允许短期误差）、热区 reconciler 收敛、unpublished 0 行 Ack、轮内 hotCutoff 不变、无冷区扫表
