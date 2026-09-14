## Context

当前 `ucg.oss`：`pang-bao` / `cn-beijing` / 公网 `oss-cn-beijing.aliyuncs.com`。Flutter Web 已走 `POST /media/upload`；原生 App 仍 `presign` + 直传。产品决定迁桶 `caidoukeji`（杭州）、服务端内网上传，且 App 未推广故统一经服务器；旧桶弃用、CDN 已绑新桶、历史对象人工同 key 拷贝。

## Goals / Non-Goals

**Goals:**

- 配置与运行时指向 `caidoukeji` + 杭州内网 endpoint。
- 全客户端（含 iOS/Android）仅经服务端 multipart 上传。
- 停用 App 预签名直传主路径，避免内网 URL 发给手机。

**Non-Goals:**

- 自动迁移 `pang-bao` 对象或改库 objectKey。
- 双桶读回退、双 endpoint 长期并存（开发可用 env 临时公网覆盖）。
- 改 CDN 域名、改 Green 地域、改 objectKey 前缀规则。

## Decisions

### D1：单 endpoint（内网）+ 全员服务端上传

- **选择**：yaml `endpoint: oss-cn-hangzhou-internal.aliyuncs.com`，`bucket: caidoukeji`，`region: cn-hangzhou`。所有 `oss.New` 用此配置。
- **理由**：App 不再直传，无需公网 SignURL；内网省流量。
- **备选**：双 endpoint — 已否决。

### D2：SDK endpoint 形态

- **选择**：`endpoint` 为 `oss-cn-hangzhou-internal.aliyuncs.com`（**不含** bucket 主机前缀）；`bucket` 字段单独为 `caidoukeji`。
- **理由**：避免 `caidoukeji.caidoukeji.oss-...`。

### D3：Flutter 统一 `uploadMediaViaGateway`

- **选择**：`uploadMediaBytes` 非 hit 路径一律 multipart `/media/upload`；辩论分享等改走同一路径；删除或不再调用 `presignMedia` 主流程。
- **理由**：与 Web 一致；视频仍受现有 20MB/网关 body 上限约束。

### D4：Presign API

- **选择**：`POST /ucg/app/api/media/presign` 返回明确业务错误（如「请使用 /media/upload」）或实现内标记废弃并拒绝签发；**禁止**再签发可被客户端使用的上传 URL。
- **理由**：防止旧客户端或误调拿到内网 URL。

### D5：开发覆盖

- **选择**：允许 `UCG_OSS_ENDPOINT`（或现有等价覆盖）在非 VPC 环境指向公网 `oss-cn-hangzhou.aliyuncs.com`；生产 compose 不设覆盖则用内网。
- **理由**：笔记本无法访问 internal。

### D6：旧桶与 CDN

- **选择**：代码不引用 `pang-bao`；CDN `cdnBaseUrl` 保持现配置；运维保证同 key 已在新桶。
- **理由**：产品已确认人工迁移 + CDN 已切。

## Risks / Trade-offs

- [大视频经网关] → 与现 Web 同上限；可后续再议分片（本期不做）。
- [生产非杭州 VPC] → 内网失败；部署前确认地域。
- [迁移未完成同 key] → CDN 404；运维 checklist。
- [旧 App 包仍调 presign] → API 拒绝 + 未推广可接受。

## Migration Plan

1. 运维：完成 `pang-bao` → `caidoukeji` 同 key 拷贝；确认 CDN。
2. 发版 ucg（配置 + presign 拒绝）+ gateway（body 上限已具备）。
3. 发版 Flutter（统一服务端上传）。
4. 回滚：yaml 临时改公网 endpoint / 旧桶仅紧急；App 回滚包。

## Open Questions

无。
