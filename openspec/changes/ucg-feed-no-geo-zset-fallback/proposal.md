## Why

`ucg-feed-geo-composite-score` 规格要求 Feed **合并 GEO 候选与 ZSET 中无 GEO 的帖**（含历史无坐标帖）。现网实现中 `collectFeedCandidates` 在 `hasViewer=true` 时 **所有半径阶梯（含 `radiusKm=0`）均走 GEO 分支**，导致仅存在于 ZSET、不在 `ucg:feed:geo` 的帖（MySQL 无 lat/lng）**永远进不了候选集**。lazy warm 已成功灌库（ZCARD>0、snapshot 齐全）后 App 仍返回空列表，与运维/用户预期不符。

## What Changes

- 修正 `collectFeedCandidates` 半径阶梯语义：`radiusKm=0`（unlimited）**MUST** 走 ZSET 扫描，**MUST NOT** 在 `hasViewer` 时仍用 GEO 20000km 替代。
- 当 viewer 含有效坐标时，在 unlimited 步从 ZSET 补全候选（含无坐标帖；已在 GEO 阶梯出现的帖由 pool/session 去重）。
- 无 viewer 坐标时行为不变（各阶梯 ZSET，等价 baseScore 排序）。
- runbook 补充：Feed 空但 ZCARD>0 时排查 GEOPOS / 无坐标帖；与 lazy warm 分工说明。
- **无**新 HTTP 接口；**无** usage 统计变更；**无** Flutter 必改项（修后端后 App 现有 lat/lng 请求应恢复有数据）。

## Capabilities

### New Capabilities

- `ucg-feed-no-geo-zset-fallback`：推荐 Feed 在 viewer 有坐标时 unlimited 半径步 ZSET 回退，保证无 GEO 帖与远距 GEO 帖可进入候选集。

### Modified Capabilities

- `ucg-recommend-feed`：明确 `radiusKm=0` 时 ZSET 全量扫语义；有 viewer 坐标时无 GEO 帖 MUST 仍可出现在 Feed（与 geo change 原 Scenario「历史帖无坐标仍参与排序」对齐）。

## Impact

- **go_ai_talk**：`internal/services/ucg/feed.go`（`collectFeedCandidates`）；可选 runbook 段落。
- **Redis / MySQL**：无 schema 变更；依赖现有 ZSET/GEO/snapshot。
- **flutter_ai_talk**：无强制变更；修后端后推荐 Tab 应恢复列表。
- **与 lazy warm 关系**：warm 解决索引冷启动；本 change 解决「有索引但读路径漏帖」。
