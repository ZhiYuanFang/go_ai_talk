## Context

`ucg-feed-geo-composite-score` 引入 Feed 读路径依赖 `SortedSetRevRangeWithScores` 扫描 `ucg:recommend:score`。实现使用 `g.Redis().Do(ctx, "ZREVRANGE", key, start, stop, "WITHSCORES")` 后 `g.NewVar(ret).Array()` 按 **i, i+1 步进扁平数组** 取 member/score。

现网测试栈（go-redis 9.2.1 + GoFrame 2.6.3 + Redis standalone）实际返回 **嵌套** `[]interface{}{ []interface{}{member, score}, ... }`。错误解析使：

- `z.Member` ≈ `["1",1.7182912825960477]`（整对序列化）
- `strconv.ParseUint(z.Member)` → `id=0` → 全部 skip
- `GEOPOS` 收到错误 member 参数（MONITOR 已证实）
- `collectFeedCandidates` pool 恒空 → Feed `list=[]`

`parseHashGetAllResult`（`hash_parse.go`）已有「多种 Redis 响应形态」先例；本 change 在 cachekit 层对齐同族模式。

约束：Redis 访问 MUST 经 `cachekit`；不新增 HTTP 路由；当前阶段不新增 `*_test.go`。

## Goals / Non-Goals

**Goals:**

- `SortedSetRevRangeWithScores` 正确解析嵌套与扁平两种 `ZREVRANGE WITHSCORES` 响应。
- 解析结果 `Member` 为十进制 postId 字符串（与 ZADD member 一致），`Score` 为 float64。
- 修复后：ZCARD>0、snapshot 齐全时 Feed 推荐 list 非空（测试栈复现条件）。
- runbook 补充 MONITOR 特征排查（GEOPOS 参数形态）。

**Non-Goals:**

- 修改 Feed 半径语义、lazy warm、composite 分公式（`ucg-feed-no-geo-zset-fallback` 已覆盖）。
- 替换 GoFrame Redis 客户端或升级 go-redis 大版本。
- 审计并修复 cachekit 内所有 Redis 命令解析（仅本函数及共用 helper 范围）。

## Decisions

### D1：共用 helper `parseZRevRangeWithScoresResult`（采用）

新增 `zset_parse.go`（或 `cache_geo.go` 内 unexported helper），逻辑：

1. `arr := g.NewVar(v).Array()`；空则返回 nil slice。
2. 若 `arr[0]` 可再 `.Array()` 且长度 ≥ 2 → **嵌套形态**：逐项取 `pair[0]`/`pair[1]`。
3. 否则 → **扁平形态**：`i += 2` 取 member/score（保留现逻辑）。
4. member/score 字符串化：优先 string/[]byte，fallback `fmt.Sprint`；trim 后 skip 空 member。

**理由**：与 `hash_parse.go` 一致；单测点集中；`SortedSetRevRangeWithScores` 仅调用 helper。

**备选（未采用）**：改用 GoFrame `ZRevRange(..., WithScores:true)` 高层 API — 仍依赖 adapter 返回形态，且与现有 `Do` 风格不一致。

### D2：member 合法性（采用）

解析后 **不** 在 cachekit 层校验 postId 数字格式；由调用方 `ParseUint` 处理。helper 只保证 member 不是 `[` 开头的 JSON 数组串。

**理由**：cachekit 为通用 ZSET 工具，不限 UCG postId。

### D3：可观测（采用）

修复后 MONITOR 中 GEOPOS 参数 MUST 为 `"1" "19"` 形态；Feed 日志应出现 `mget ucg:post:snapshot:*`。

### D4：验收（采用）

测试栈（无需 DEL 索引）：

```bash
curl -s "http://127.0.0.1:19804/ucg/app/api/feed/recommend?pageSize=20"
# list 非空
docker exec go-ai-talk-redis-test redis-cli MONITOR  # GEOPOS 参数为纯 postId
```

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 误判嵌套/扁平（混合数组） | 以首元素是否可 `.Array()` 且 len≥2 探测；扁平路径保留 |
| 其他 Redis 命令同类 bug | 本 change 仅 ZREVRANGE WITHSCORES；后续可开独立 audit |
| 生产 Cluster 响应形态不同 | helper 双路径覆盖；部署后抽样 MONITOR |

## Migration Plan

1. 部署含 fix 的 `ucg-service`（platform 包随二进制发布）。
2. **无需** Redis 迁移或 backfill。
3. 回滚：还原二进制。

## Open Questions

- 无（MONITOR 已确认根因；实现为 helper + 单点替换）。
