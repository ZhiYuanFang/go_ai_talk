## 1. UCG OSS 配置

- [x] 1.1 更新 `config.ucg-service.yaml`：`bucket=caidoukeji`、`region=cn-hangzhou`、`endpoint=oss-cn-hangzhou-internal.aliyuncs.com`；注释说明 SDK 两段式与内网前提
- [x] 1.2 `LoadOSSConfig` 支持 env 覆盖 endpoint（本地公网联调）；确认 AK 仍仅 env 注入
- [x] 1.3 `.env.example` / runbook：新桶、内网、手动同 key 迁移、CDN 已切说明

## 2. 停用预签名直传

- [x] 2.1 `PresignUpload` / `POST …/media/presign`：拒绝签发 uploadUrl，返回明确改走 `/media/upload` 的错误
- [x] 2.2 确认服务端 Put/删/缩略图/转码均只用配置 endpoint（无硬编码 pang-bao）

## 3. Flutter 统一服务端上传

- [x] 3.1 `uploadMediaBytes`：非 resolve-hit 路径一律 `uploadMediaViaGateway`（去掉原生 presign 分支）
- [x] 3.2 辩论分享等其它 `presignMedia` 调用改为服务端上传
- [x] 3.3 自检：Web/原生均不请求可用的 OSS 预签名 PUT

## 4. 收尾

- [x] 4.1 确认未引入双桶读、无自动迁移脚本、无改 DB objectKey
- [x] 4.2 部署检查清单：杭州 VPC、人工拷贝完成、CDN 指向 caidoukeji
