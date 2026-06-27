## 1. ucg sample 封面 CDN

- [x] 1.1 `PostSampleItem` / `api/v1/ucg_internal_posts_sample_http.go` 增加 `coverCdnUrl` 字段与中文注释
- [x] 1.2 `post_sample_internal.go`：行映射时按 `mediaType` 填 `coverCdnUrl`（图 `BuildCdnURL`、视频 `BuildVideoSnapshotURL`）

## 2. sim T2 多模态评论

- [x] 2.1 `simuser` 包内 helper：由 `coverCdnUrl` + 渲染 prompt 构建 `aimodel.Message` Content（多模态或纯文本）
- [x] 2.2 `RunCommentTask`：解析 sample 的 `coverCdnUrl`；单次 `LaneSimVision` Invoke；CDN 缺失时 warning + 降级
- [x] 2.3 `schema.go` 默认 `comment` prompt 微调（只输出评论正文、结合配图与正文）

## 3. 校验

- [x] 3.1 `go build` 相关包通过
- [x] 3.2 `openspec validate sim-t2-comment-multimodal` 通过
