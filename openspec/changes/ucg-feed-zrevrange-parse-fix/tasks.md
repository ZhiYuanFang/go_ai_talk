## 1. cachekit 解析修复

- [x] 1.1 新增 `parseZRevRangeWithScoresResult`（`zset_parse.go` 或 `cache_geo.go`）：嵌套 `[[m,s],...]` 与扁平 `[m,s,m,s,...]` 双路径；member/score 字符串化与空 member 跳过
- [x] 1.2 `SortedSetRevRangeWithScores` 改为调用 helper，删除原 `i+=2` 单一假设；补充中文注释说明 go-redis 9.x 嵌套形态
- [x] 1.3 `go build ./internal/platform/cachekit/... ./cmd/ucg-service/...` 通过

## 2. 文档

- [x] 2.1 在 `docs/runbooks/release-deploy-and-run.md` UCG Feed 章节补充：Feed 空且 ZCARD>0 时 MONITOR 若见 `GEOPOS ... "[\"id\",score]"` → ZREVRANGE 解析 bug（本 fix）

## 3. 验收

- [ ] 3.1 测试栈：ZCARD>0、snapshot 齐全；`GET /feed/recommend?pageSize=20` 返回非空 list（带/不带 lat/lng）
- [ ] 3.2 MONITOR：`GEOPOS ucg:feed:geo` 参数为纯 postId（如 `"1"`），非 JSON 数组串
- [ ] 3.3 ucg-service 日志含 `mget` / snapshot 读取（候选池非空）
