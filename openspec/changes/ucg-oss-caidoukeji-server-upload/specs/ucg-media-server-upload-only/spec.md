## ADDED Requirements

### Requirement: All app clients MUST upload media via server multipart API

UCG 媒体上传主路径 MUST 为服务端 multipart 上传（`POST /ucg/app/api/media/upload` 或经 gateway 同域代理的等价路径）。iOS/Android 客户端 MUST 与 Web 一样走该路径，MUST NOT 再依赖「预签名 URL + 客户端直传 OSS」作为主上传流程。

#### Scenario: 原生上传经服务端

- **WHEN** 原生 App 上传图片或视频且 resolve 未命中去重
- **THEN** 客户端 SHALL 将文件以 multipart 提交至服务端上传接口，SHALL NOT PUT 至 OSS 预签名 URL

### Requirement: Media presign MUST NOT issue client-usable upload URLs

`POST /ucg/app/api/media/presign`（若路由仍保留）MUST 拒绝签发可供客户端直传的预签名上传 URL（返回明确错误）。服务端 MUST NOT 向客户端返回指向 OSS 内网域名的上传地址。

#### Scenario: 调用 presign 被拒绝

- **WHEN** 客户端请求 media presign
- **THEN** 接口 SHALL 失败并提示改用服务端 upload，且 SHALL NOT 返回可用的 uploadUrl

### Requirement: Historical pang-bao objects are out of scope for dual-read

系统 MUST NOT 实现运行时回退读取旧桶 `pang-bao`。历史可见性依赖运维将对象以相同 objectKey 拷贝到 `caidoukeji` 且 CDN 已指向新桶。

#### Scenario: 同 objectKey 在新桶可读

- **WHEN** 数据库中已有 objectKey 且该对象已按相同 key 存在于 `caidoukeji`，CDN 指向新桶
- **THEN** 既有展示 URL（cdnBaseUrl + objectKey）SHALL 可访问，无需改库
