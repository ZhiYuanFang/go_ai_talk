## 1. UCG：删除服务端转码

- [x] 1.1 修改 `UploadMediaObject`：`mediaKind=2` 仅 v1 验真通过则 PUT，失败直接返回错误；删除 Probe→Normalize 转码分支
- [x] 1.2 修改 `InternalMediaUploadVideo` / `UploadVideoTranscodedObject` 调用链为「v1 验真 + 直传」，响应含 `transformVersion=v1` 与对应 contentHash
- [x] 1.3 删除或停用 `NormalizeVideo` 及上传路径对 `ProbeVideoDecodable` 的调用；更新相关中文注释与 runbook（三分支→两分支）
- [x] 1.4 扩展 ucg internal 上传以支持引导片前缀（如 `kind=xiaozhi-guide`）：同 v1 验真与体积上限、无转码、写入 `xiaozhi-guide/` 前缀并返回 objectKey/cdnUrl

## 2. sim T4 对齐

- [x] 2.1 调整 T4：按 upload-video 响应的 `transformVersion`（v1）register；失败 skip，不回退 presign
- [x] 2.2 确认 T4 素材为 v1 合规 mp4（不合规则更换或文档说明）

## 3. cash：双槽与 Admin/App

- [x] 3.1 schema：新增 `xiaozhi_mcp_guide_video` 并种子 `promo`/`tutorial` 默认标题
- [x] 3.2 领域：读两槽、改标题、上传（经 `clients/ucg`）、清空 object_key；HTTP 边界映 CDN（及可选 cover）
- [x] 3.3 Admin API + 控制器：`GET/PUT` guide-videos、`POST …/{slot}/upload`、`DELETE …/{slot}`
- [x] 3.4 App API：`GET /cash/app/api/xiaozhi-mcp/guide-videos`（Bearer）；**不**写入 `maintenance_skip`（计入 usage）
- [x] 3.5 `clients/ucg`：封装引导视频上传（internal secret + kind/前缀）

## 4. Hub 区 D

- [x] 4.1 `cash-xiaozhi-mcp-admin.html` 增加区 D：两槽标题、预览、上传、清空与错误提示（合规 mp4 / 同 Flutter 上限）
- [x] 4.2 确认静态页与 Admin 鉴权路径可用（无需新导航项除非体验需要）

## 5. Flutter（兄弟仓，本迭代）

- [ ] 5.1 对接 `GET …/guide-videos`：空列表完全隐藏入口；有项按 title 展示并点播 `videoCdnUrl`
- [ ] 5.2 与本仓契约联调（登录 Bearer、usage 可观测）

## 6. 校验与文档

- [x] 6.1 更新 `docs/runbooks` 中 UCG 视频上传/转码相关说明
- [x] 6.2 `openspec validate xiaozhi-guide-videos-drop-transcode --strict` 通过
