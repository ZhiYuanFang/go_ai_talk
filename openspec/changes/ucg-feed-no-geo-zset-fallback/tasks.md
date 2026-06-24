## 1. 读路径修复（go_ai_talk）

- [x] 1.1 在 `collectFeedCandidates` 将 GEO 分支条件由 `hasViewer && radiusKm >= 0` 改为 `hasViewer && radiusKm > 0`
- [x] 1.2 在 ZSET 分支：`radiusKm == 0` 时 **不** 对 `hasViewer` 执行 `inGeo` skip；`radiusKm > 0` 且 `!hasViewer` 的 ZSET 步保持现逻辑
- [x] 1.3 补充中文注释：unlimited 步与 geo change design D3 对齐；无 GEO 帖须在 viewer 有坐标时可见

## 2. 文档

- [x] 2.1 在 `docs/runbooks/release-deploy-and-run.md` UCG Feed 章节补充：Feed 空但 ZCARD>0 时查 GEOPOS / 无坐标帖；本 fix 与 lazy warm 分工

## 3. 验收

- [ ] 3.1 测试栈：ZSET 有帖、GEOPOS 空、snapshot 有；`GET /feed/recommend?lat=...&lng=...` 返回非空 list
- [ ] 3.2 无 lat/lng 请求仍返回非空 list（与变更前 ZSET 行为一致）
- [ ] 3.3 有坐标帖在 GEO 50–500km 内仍按 composite 分排序，近距帖优先
