## Context

- T2 使用 `posts/sample` `mode=random` 全库 ID 探测选帖，当前含 `mediaType=2`（视频）。
- T2 已支持多模态评论（`coverCdnUrl`）；视频帖使用 `BuildVideoSnapshotURL` 首帧，效果弱于图文全图。
- T4 负责 sim 发视频；T2 排除视频与任务分工一致。
- 探索结论：在 **ucg SQL 源头** 过滤优于 sim 侧重试。

## Goals / Non-Goals

**Goals:**

- sample random/latest 支持 `excludeMediaTypes`，过滤 `ucg_post.media_type`。
- T2 固定排除 `[2]`；候选为纯文字（0）+ 图文（1）。
- MIN/MAX 与 `id>=R` 探测使用相同 WHERE 条件，保证随机分布仅在 eligible 集合内。

**Non-Goals:**

- 不排除纯文字帖（仍允许 `mediaType=0`）。
- 不为 T2 单独新增 HTTP 接口；不修改 recommend Feed。
- 不从 sample 响应删除 video 相关 `coverCdnUrl` 逻辑（其他调用方 future-proof）。

## Decisions

### 1. 契约：`excludeMediaTypes []int`

请求 body 示例（T2）：

```json
{ "mode": "random", "excludeMediaTypes": [2] }
```

- 缺省或空数组：不过滤（向后兼容）。
- 非法值（负数等）忽略或 400；实现采用 **静默忽略非已知 mediaType** 或只接受 0/1/2。
- 上限：数组长度 ≤ 8，防止滥用。

### 2. SQL 过滤

`postSampleBaseModel` / random bounds 共用 helper：

```go
func postSampleApplyMediaExcludes(m *gdb.Model, exclude []int) *gdb.Model
```

对 `excludeMediaTypes` 去重后 `WhereNotIn("p.media_type", exclude)`。

`SampleRandomPublishedPost(ctx, excludeMediaTypes []int)`：

- bounds 查询同样带 exclude
- 若 eligible 集合为空 → 返回 `[]PostSampleItem{}`

`SamplePublishedPosts(ctx, limit, excludeMediaTypes []int)`：latest 路径同步支持，便于一致性与未来调用。

### 3. sim T2

```go
ucgInternalPost(..., g.Map{
  "mode": "random",
  "excludeMediaTypes": []int{2},
}, ...)
```

防御性：

```go
if post.MediaType == 2 {
  glog.Warningf(...); RecordTaskRun(..., false, "跳过视频帖"); return
}
```

### 4. 常量

sim 侧使用字面 `2` 或包内常量注释对齐 ucg `MediaTypeVideo`；**禁止** sim import ucg 包——用注释说明与 ucg `media_type=2` 一致。

## Risks / Trade-offs

- **[Risk] 广场几乎全为视频时 T2 长期失败** → 可接受；运营可依赖图文/文字帖或调 sim 其他任务。
- **[Risk] `(status, media_type, id)` 无复合索引时 MIN/MAX 略慢** → T2 低频；与现 random 同级。
- **[Risk] ucg 未部署新字段、sim 已部署** → sim 防御 skip video；旧 ucg 仍可能返回 video 直至 ucg 升级。

## Migration Plan

1. 部署 ucg-service（`excludeMediaTypes`）→ 部署 sim-user-service（T2 传 `[2]`）。
2. 回滚：sim 去掉 exclude 字段；ucg 忽略未知字段无害。
3. 验收：T2 日志/抽样不再出现 `mediaType=2` 成功评论。

## Open Questions

- 是否将 exclude 写死进 random 模式默认值 — **否**，显式由 sim 传参，sample 保持通用。
