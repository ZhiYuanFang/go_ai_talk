## 1. ucg 契约与 service

- [x] 1.1 `api/v1/ucg_internal_posts_sample_http.go`：请求体增加 `Mode string`（`latest` / `random`），补充中文注释
- [x] 1.2 `post_sample_internal.go`：实现 `SampleRandomPublishedPost`（MIN/MAX + 幂次偏置 α=0.65 + `id>=R LIMIT 1`）；抽取与 latest 共用的字段/SQL 片段
- [x] 1.3 `post_sample_internal.go`：random 路径使用 `crypto/rand`；无 published 帖、min=max 单帖等边界处理
- [x] 1.4 `ucg_internal_posts_sample.go`：`mode=random` 调 random 函数；缺省/`latest` 保持 `SamplePublishedPosts`

## 2. sim T2

- [x] 2.1 `tasks.go` `RunCommentTask`：请求改为 `{ "mode": "random" }`；去掉 `rand.Intn`；直接使用 `sample.List[0]`（先判空）
- [x] 2.2 确认 T2 仍不调用 `feed/recommend`；空 list 失败语义不变

## 3. 文档与校验

- [x] 3.1 runbook（若已有 T2/sample 说明）：补充 random 模式语义一行
- [x] 3.2 `openspec validate sim-t2-random-post-probe` 通过
