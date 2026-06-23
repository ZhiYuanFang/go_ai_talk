## 1. 共享命名与 URL

- [x] 1.1 新增 `internal/shared/mediacdn/thumb.go`：`ImageThumbSuffix`、`ThumbObjectKey`、`IsThumbObjectKey`（扩展名前插入 `_thumb`，扩展名与原图一致）
- [x] 1.2 修改 `BuildImageThumbnailURL`：改为 `BuildCdnURL(ThumbObjectKey(objectKey))`，移除图片侧 `imageThumbProcess` / `appendOssProcess` 用于图片的分支

## 2. 缩略图生成与 OSS 集成

- [x] 2.1 新增 `internal/services/ucg/oss_thumb.go`：按扩展名组装 OSS process 参数、`EnsureImageThumb`（HEAD 幂等、GetObject+Process、PutObject）
- [x] 2.2 `RegisterMedia` miss 成功路径调用 `EnsureImageThumb`；失败时 register 返回错误
- [x] 2.3 `putOSSObject` 在 `mediaKind==1` 上传成功后调用 `EnsureImageThumb`
- [x] 2.4 `DeleteOwnedMedia` 删除原图时一并 `DeleteObject(ThumbObjectKey(key))`（忽略 thumb 404）

## 3. Backfill CLI 与 Runbook

- [x] 3.1 新增 `cmd/ucg-image-thumb-backfill/main.go`：从 UCG DB 收集去重图片 objectKey，支持 `--dry-run`/`--limit`/`--concurrency`，调用 `EnsureImageThumb`，输出汇总
- [x] 3.2 新增 `docs/runbooks/ucg-image-thumb-backfill.md`：环境变量、test/prod 执行顺序、验证抽样、与读路径切换的发布顺序

## 4. 部署与验证

- [ ] 4.1 test 环境执行 backfill（先 dry-run 再正式），抽样确认原图与 `*_thumb.*` 成对存在
- [ ] 4.2 prod 环境 backfill 完成后发布读路径切换；验证 Admin 帖子列表、App 头像/动态/聊天图片 `thumbnailUrl` 可加载且无 `x-oss-process`
