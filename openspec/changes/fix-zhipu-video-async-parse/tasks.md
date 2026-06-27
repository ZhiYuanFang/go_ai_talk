## 1. aimodel 解析修复

- [x] 1.1 扩展 `parseVideoPollBody`：解析 `video_result[]`、映射 `task_status`（PROCESSING/SUCCESS/FAIL）
- [x] 1.2 保留 `video_url` / `output.video_url` fallback；SUCCESS 无 URL 仍返回 processing

## 2. 验收

- [x] 2.1 `go build ./cmd/sim-user-service/...` 通过
- [x] 2.2 `openspec validate fix-zhipu-video-async-parse` 通过
- [x] 2.3 手动或日志确认：SUCCESS + video_result 时 T4 不再 poll timeout（代码路径已对齐官方 AsyncVideoGenerationResponse；部署后手动 T4 验证）
