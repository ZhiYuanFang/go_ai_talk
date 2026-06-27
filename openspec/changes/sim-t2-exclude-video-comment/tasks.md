## 1. ucg sample 过滤

- [x] 1.1 API/controller：`excludeMediaTypes []int` 解析并传入 service
- [x] 1.2 `post_sample_internal.go`：`SampleRandomPublishedPost` / `SamplePublishedPosts` / baseModel 支持 media_type 排除；bounds 与 probe 同一 filter

## 2. sim T2

- [x] 2.1 `RunCommentTask` sample 请求增加 `"excludeMediaTypes": []int{2}`
- [x] 2.2 防御：`mediaType==2` 时 warning + 失败退出，不评论

## 3. 校验

- [x] 3.1 `go build` 相关包通过
- [x] 3.2 `openspec validate sim-t2-exclude-video-comment` 通过
