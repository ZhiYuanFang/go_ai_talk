# UCG 图片物理缩略图 Backfill

一次性为历史图片 objectKey 在 OSS 上生成 `{stem}_thumb.{ext}` 物理对象，供 `BuildImageThumbnailURL` 返回无 `x-oss-process` 的 CDN URL。

## 前置条件

- 可访问目标环境 UCG MySQL（`UCG_DB_LINK`）与阿里云 OSS（`UCG_OSS_ACCESS_KEY_ID` / `UCG_OSS_ACCESS_KEY_SECRET`）。
- 本机或运维机已 checkout 仓库，Go 1.19+。
- **推荐发布顺序**：先部署含 `EnsureImageThumb` 写路径的版本 → 本脚本 backfill → 再部署/确认读路径已切换为物理 thumb URL。

## 环境变量

| 变量 | 说明 |
|------|------|
| `GF_GCFG_FILE` | 默认 `manifest/config/config.ucg-service.yaml` |
| `UCG_DB_LINK` | UCG 库 DSN（与 ucg-service 一致） |
| `UCG_OSS_ACCESS_KEY_ID` / `UCG_OSS_ACCESS_KEY_SECRET` | OSS 凭证 |
| `MYSQL_TCP_HOST` | 可选，覆盖 DSN 内 MySQL 主机 |

## 执行（test 示例）

在仓库根目录（默认加载 `manifest/docker/env/.env.test`，含 `UCG_DB_LINK`、`UCG_OSS_ACCESS_KEY_*`、`MYSQL_TCP_HOST` 等）：

```bash
# 1. 试跑：仅打印将处理的 key（无需手动 export）
go run ./cmd/ucg-image-thumb-backfill --dry-run

# 指定其他 env 文件或不加载
go run ./cmd/ucg-image-thumb-backfill --env-file manifest/docker/env/.env.test
go run ./cmd/ucg-image-thumb-backfill --env-file=""   # 仅用当前 shell 环境变量

# 2. 小批量验证
go run ./cmd/ucg-image-thumb-backfill --limit 10

# 3. 全量
go run ./cmd/ucg-image-thumb-backfill --concurrency 4
```

也可显式 export 覆盖 dotenv 中的单项（已存在的环境变量优先于文件）：

```bash
export UCG_DB_LINK='mysql:user:pass@tcp(127.0.0.1:3306)/ai_voice_ucg_test'
go run ./cmd/ucg-image-thumb-backfill --env-file=""
```

脚本输出汇总：`total / ok / missing_original / failed`。`failed > 0` 时退出码为 1，可修复后重跑（对已存在 thumb 幂等跳过）。

## 验证

1. 抽样原图 key `social/.../a.jpg`，确认 OSS 存在 `social/.../a_thumb.jpg`。
2. 打开 UCG Admin 动态审查列表，图片 `thumbnailUrl` 可加载且 URL **不含** `x-oss-process`。
3. App 侧头像、动态、聊天图片缩略图正常。

## prod

与 test 相同流程，使用 prod DSN 与 OSS 凭证；建议在低峰执行。backfill 完成并验证后再对外宣告读路径切换完成。

## 回滚

读路径若回退到 `x-oss-process` 可快速恢复展示；已生成的 `*_thumb.*` 对象可保留，无副作用。
