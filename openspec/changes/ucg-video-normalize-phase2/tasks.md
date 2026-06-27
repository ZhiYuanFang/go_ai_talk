## 1. 可解码探测

- [x] 1.1 `video_validate.go` 新增 `ProbeVideoDecodable(data []byte) error`：ffprobe 成功且含 video stream；失败语义与 `videoInvalidErr` 对齐
- [x] 1.2 补充中文注释：与 v1/v2 验真分工、Web B 分支门禁

## 2. Web 代理上传转码兜底

- [x] 2.1 `UploadMediaResult` 增加 `TransformVersion string`（`v1`|`v2`）
- [x] 2.2 改造 `UploadMediaObject`（`mediaKind=2`）：v1 通过 → 直传 + `TransformVersion=v1`；v1 失败 → `ProbeVideoDecodable` → 失败 4xx；成功 → 复用 `UploadVideoTranscodedObject` + `TransformVersion=v2`
- [x] 2.3 确认转码失败返回 5xx、不留下 OSS 对象；与 Phase 1 semaphore/超时一致

## 3. Controller 与 API 响应

- [x] 3.1 `ucg_media_upload`：`mediaKind=2` 成功时响应 `data.transformVersion`（来自 `UploadMediaResult.TransformVersion`）
- [x] 3.2 确认 gateway-app 对 `/ucg/app/api/*` 全量透传（无字段裁剪）

## 4. 文档与验收

- [x] 4.1 更新 `docs/runbooks/release-deploy-and-run.md`：Web Phase 2 三分支（直传 v1 / 转码 v2 / 4xx）、register 须读 `transformVersion`
- [x] 4.2 test 环境人工回归：v1 Baseline+AAC 直传；HEVC 或无音轨转码 v2；损坏文件 4xx；v1 无 faststart 仍直传
- [x] 4.3 `openspec validate ucg-video-normalize-phase2 --strict` 通过
- [x] 4.4 评审：App API usage — 无新路径，响应增字段，**不统计**；Flutter Web 客户端 follow-up 记 sibling repo

## 5. 显式不在本变更

- [x] 5.1 sim-user-service / internal upload-video 行为变更（无）
- [x] 5.2 Flutter Web 读取 `transformVersion`（sibling repo，本任务组仅 design 备注）
