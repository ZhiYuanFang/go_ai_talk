## 1. Platform — cachekit 扩展

- [x] 1.1 扩展 `cachekit.Cache` 接口与 `RedisCache`/`observedCache`：Decr、TTL、Set（无 EX）、Persist、Hash*、List*、Set*、SortedSetAdd
- [x] 1.2 实现 `HashGetAll` 并迁入 `usagestats/redisHashToMap` 解析逻辑（含 flat `[]string`）
- [x] 1.3 新增 `cachekit.Default()`（或等价）返回 `WithObserver(NewRedisCache(), LoggingObserver{})`
- [x] 1.4 新增 `keys_device.go`、`keys_history.go`、`keys_gateway.go`、`keys_voice.go`、`keys_ucg.go`、`keys_ai.go`；扩展 `keys.go` 域常量（DomainUCG、DomainAI 等）
- [x] 1.5 登记全部现有键 builder（策略 A：字符串与线上一致），附中文注释（TTL/共享域/失效）

## 2. Platform — redismsgkit

- [x] 2.1 新建 `internal/platform/redismsgkit`：Publisher、Subscriber、Observer、errors
- [x] 2.2 `channels.go`：`ChannelAppHistoryNotify = "app:history:notify"`
- [x] 2.3 收编 `gatewayapp/history_subscriber.go` 的 go-redis standalone/cluster 订阅逻辑
- [x] 2.4 新增 `DefaultPublisher()`（WithObserver + LoggingObserver）

## 3. 业务迁移 — gateway / device

- [x] 3.1 `gatewayapp/usagestats/store.go` → cachekit + gateway 键 builder；删除本地 `redisHashToMap`
- [x] 3.2 `gatewayapp/usagestats/simulated_wx.go`、`device/sim_user.go` → cachekit Set* + `GatewayUsageSimWxSetKey`
- [x] 3.3 `gatewayapp/version_admin.go`、`controller/gateway_app_ctrl.go` → cachekit + version 键 builder
- [x] 3.4 `gatewayapp/refresh_store.go` → 改用 platform refresh 键 builder（已合规访问路径，统一键源）
- [x] 3.5 `gatewayapp/history_subscriber.go` → 薄包装调用 redismsgkit

## 4. 业务迁移 — history / voice

- [x] 4.1 `history/realtime_notify.go` → cachekit Incr（piece ver）+ redismsgkit Publish；键 builder
- [x] 4.2 `history/piece.go` → cachekit Get/SetEX + history piece 键 builder
- [x] 4.3 `history/cache_repo.go` → 确认已合规（必要时统一 Default()）
- [x] 4.4 `voice/voice_chat.go` → guard rate/idem 键 builder（session 已用 cache）
- [x] 4.5 `voice/clinic_session.go`、`clinic_rate.go`、`clinic_summary.go` → cachekit + voice 键 builder
- [x] 4.6 `voice/ai_quota_store.go` → cachekit + `AIQuotaUsageKey`
- [x] 4.7 `device/wx.go` → 改用 platform wx 键 builder（保留 `dev:` 前缀）

## 5. 业务迁移 — ucg / aimodel

- [x] 5.1 `ucg/chat_store.go`、`chat_persist.go`、`audit_chat.go` → cachekit + ucg chat 键 builder；**删除** `ucg/chat_keys.go`
- [x] 5.2 `ucg/profile_audit.go`、`ip_location.go`、`ai_quota.go` → cachekit + 对应 builder
- [x] 5.3 `ucg/recommend_mq_consumer.go` → `WithObserver` + recommend throttle 键 builder
- [x] 5.4 `aimodel/gate.go` → cachekit + `keys_ai` gate builder

## 6. 治理与文档

- [x] 6.1 更新 `AGENTS.md`：Redis 访问（强制）+ Redis 键命名（强制）+ 评审 grep
- [x] 6.2 更新 `openspec/project.md` 第 32/95 行与 cachekit/redismsgkit 交叉引用
- [x] 6.3 （可选）新增 `hack/check-redis-bypass.sh` 并在 design/runbook 引用
- [x] 6.4 验收 grep：`rg 'g\.Redis\(\)' internal/services internal/controller` 为 0；业务层无 Redis 键字面量

## 7. 验收

- [x] 7.1 编译通过：`go build ./...`
- [ ] 7.2 手工：usage Admin 列表在 Redis 有数据时非空（HGETALL flat 路径）
- [ ] 7.3 手工：App 历史变更 WS 通知仍可达（publish + subscribe）
- [ ] 7.4 手工：ucg 聊天读写、clinic session 固定 TTL 语义不变
- [x] 7.5 `openspec validate enforce-redis-platform-access --strict` 通过
