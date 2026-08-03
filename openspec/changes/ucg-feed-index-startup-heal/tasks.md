## 1. 配置与键

- [x] 1.1 `cachekit` 登记 startup-heal 锁键 builder（中文注释：TTL/跨进程语义）
- [x] 1.2 `feed` 配置/env：startup check/heal 开关、`indexHealGapThreshold`、`indexStartupHealMaxPosts`、锁 TTL；写入 `config.ucg-service.yaml` 默认值

## 2. 冷判据与共享缺口检测

- [x] 2.1 抽出 publishedCount / zcard / gap 判断（threshold）；修正 `isFeedIndexCold` / needsWarm：短缺亦触发请求路径有界 warm
- [x] 2.2 更新 lazy warm 日志，区分「空索引」与「短缺补齐」

## 3. 启动自检 + 异步 heal

- [x] 3.1 实现 `StartFeedIndexStartupHeal`（或等价）：check → ERROR 日志 → 可选异步 heal（独立锁、分页 `syncPublishedPostRedis`、有界 maxPosts、单帖失败继续）
- [x] 3.2 `cmd/ucg-service/main.go` 在依赖检查后、HTTP Run 前启动该任务（不阻塞 Listen）

## 4. 校验

- [x] 4.1 `openspec validate ucg-feed-index-startup-heal --strict` 通过
- [x] 4.2 确认无新增 ticker；heal 可关；多副本锁语义在注释/日志中可观测
