## Why

UCG 媒体需迁到新 OSS 桶 `caidoukeji`（杭州），并用内网 endpoint 降低费用与延迟。App 尚未推广，可将原生上传改为与 Web 一致的服务端直传，从而无需为预签名保留公网 endpoint，也无需双桶读旧 `pang-bao`。

## What Changes

- **BREAKING（客户端）**：Flutter 原生不再走 `POST /media/presign` + App 直传 OSS；一律经 `POST /media/upload`（网关同域代理 → ucg `PutObject`）。
- ucg OSS 配置：`bucket=caidoukeji`，`region=cn-hangzhou`，`endpoint=oss-cn-hangzhou-internal.aliyuncs.com`（SDK 两段式，不含 bucket 前缀）。
- CDN 继续使用现有 `cdnBaseUrl`（运维已指向 `caidoukeji`）。
- 旧桶 `pang-bao`：**代码弃用**；历史对象由运维**手动**拷贝至新桶且 **保持相同 objectKey**（库表不改）。
- 预签名主路径停用或明确拒绝（避免误配内网 URL 给客户端）；服务端 Put/删/缩略图/转码全部走内网 endpoint。
- 本地/非 VPC 开发：允许 env 覆盖公网 endpoint（design 约定）。

## Capabilities

### New Capabilities

- `ucg-oss-caidoukeji-config`：新桶/地域/内网 endpoint 与配置加载约定。
- `ucg-media-server-upload-only`：全客户端经服务端上传；停用 App 预签名直传主路径。

### Modified Capabilities

- （无独立 `openspec/specs/<capability>` 活文档；既有归档条文中 `pang-bao` / 公网 endpoint / App presign 行为由本变更 ADDED 规格取代并在 design 说明迁移。）

## Impact

- **go_ai_talk**：`config.ucg-service.yaml`、`LoadOSSConfig`、presign/upload 相关；runbook / `.env.example` 注释。
- **flutter_ai_talk**：`ucg_repository` 上传分支、辩论分享等仍调 presign 的路径。
- **运维**：ucg 须在杭州同地域可访问内网 OSS；手动迁移 `pang-bao` → `caidoukeji`（同 key）；CDN 已切新桶。
- **不改**：DB schema、objectKey 生成规则、Green 审核 endpoint（可仍北京，AK 须有新桶权限）。
