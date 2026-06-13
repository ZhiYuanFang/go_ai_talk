## 1. Green dataId 规范化

- [x] 1.1 在 `internal/services/ucg/green_client.go` 新增 `greenDataIDFromMediaURL(url string) string`：`url.Parse` 取 path、`/`→`_`、过滤非法字符、空则返回空串、超 64 截断
- [x] 1.2 修改 `ModerateImageURL`：构建 `ServiceParameters` 时仅当 dataId 非空才写入；禁止 `dataId=imageURL`
- [x] 1.3 修改 `ModerateVideoURL`：与图片路径共用 dataId  helper
- [x] 1.4 补充中文注释：dataId 约束、为何从 URL path 推导、空则省略

## 2. 错误信息可观测

- [x] 2.1 改进 `parseImageModeration`：`body.Code!=200` 时输出 `code` + `Msg`；HTTP 非 200 时带 status
- [x] 2.2 改进 `parseVideoModeration`：与图片路径对齐错误格式

## 3. 校验与文档

- [x] 3.1 `go build ./...` 通过
- [x] 3.2 `openspec validate fix-ucg-green-media-dataid --strict` 通过
- [x] 3.3 在 `docs/runbooks/release-deploy-and-run.md` 补充：dataId 修复说明、新发图帖验证步骤、status=5 存量不自动重审
