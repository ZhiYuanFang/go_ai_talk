## 1. ucg-service：posts/sample 扩展

- [x] 1.1 `post_sample_internal.go`：响应增加 `authorWxId`；支持 `excludeAuthorWxIds` 过滤（latest + random）
- [x] 1.2 `ucg_internal_posts_sample.go` + `api/v1/ucg_internal_posts_sample_http.go`：请求/响应契约

## 2. sim-user-service：T6 重写

- [x] 2.1 `clients.go`：新增 `sampleRandomRealAuthor(ctx, excludeSimWxIds)`；删除 `pickTwoDistinctSimWx`
- [x] 2.2 `tasks.go`：重写 `RunFollowTask`（sim → 真人 author，自关注重试）

## 3. 验收

- [x] 3.1 `go build` ucg-service 与 sim-user-service 通过
- [x] 3.2 `openspec validate sim-t6-follow-real-author` 通过
