## ADDED Requirements

### Requirement: 跨域调用 MUST 经 clients/push

voice、ucg 及其他域服务发送系统推送时 MUST 通过 `internal/clients/push`（或同等 clients 包）出站 HTTP 调用 push-service，MUST NOT import `internal/services/push` 业务实现包，MUST NOT 再 import `internal/services/ucg` 仅为发送推送。UCG 社区可见/静默角标推送 MUST 在调用前计算未读（若需要）并将 `badge` 放入请求。

#### Scenario: voice 预测不依赖 ucg 推送包

- **WHEN** voice 预测临近校验通过并发送提醒
- **THEN** 出站目标 MUST 为 push-service，MUST NOT 调用已删除的 ucg internal push 路径作为唯一实现

### Requirement: UCG 进程 MUST 移除厂商推送实现

`ucg-service` MUST 删除（或停止编译）原 `push_*.go` 厂商发送与 `ucg_push_device` 写路径权威，MUST 删除 App/Internal 推送注册与 by-biz-type 路由。社区私信/评论等原推送触发点 MUST 改为经 `clients/push` 发送。

#### Scenario: ucg 不再直连 APNs 配置作为推送宿主

- **WHEN** 审查 ucg-service 运行时配置
- **THEN** 推送厂商证书/密钥 MUST NOT 再作为 ucg-service 发送推送的必需依赖（可整段移除 `UCG_APNS_*` 推送注入）
