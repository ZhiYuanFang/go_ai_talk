# UCG OSS：caidoukeji（杭州内网）与服务端上传

## 配置

| 项 | 值 |
|----|-----|
| bucket | `caidoukeji` |
| region | `cn-hangzhou` |
| endpoint（生产） | `oss-cn-hangzhou-internal.aliyuncs.com`（SDK 两段式，**不含** bucket 主机前缀） |
| CDN | `https://resorce.cuplay.top`（运维已指向 `caidoukeji`） |
| AK | `UCG_OSS_ACCESS_KEY_ID` / `UCG_OSS_ACCESS_KEY_SECRET`（yaml 留空） |
| 本地覆盖 | 可选 `UCG_OSS_ENDPOINT=oss-cn-hangzhou.aliyuncs.com`（非 VPC） |

## 上传路径

- App / Web：**仅** `POST /ucg/app/api/media/upload`（经 gateway multipart）。
- `POST …/media/presign`：**拒绝签发** uploadUrl（提示改走 upload）。

## 旧桶 `pang-bao`

- 代码弃用；**无**双桶读、**无**自动迁移、**不改**库表 objectKey。
- 运维手动将对象以**相同 objectKey** 拷贝至 `caidoukeji`；未拷完则 CDN 可能 404。

## 部署检查清单

1. ucg-service 运行在**杭州同地域 VPC**（可访问内网 OSS）。
2. 人工拷贝 `pang-bao` → `caidoukeji`（同 key）已完成或可接受窗口。
3. CDN 已绑定 `caidoukeji`。
4. 发版 ucg（配置 + presign 拒绝）与 Flutter（统一服务端上传）。
5. 抽测：图片/视频经网关上传写入新桶；旧 objectKey CDN 可访问。
