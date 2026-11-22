# push-notify

## Purpose

推送服务、APNs/HMS、token、dispatch 可观测与 notify 运行时的行为规格。 Push and notification delivery independent of prediction copy.

<!-- migrated-from: hms-notification-payload, notify-service-runtime, push-apns-provider-auth, push-app-and-internal-api, push-caller-migration, push-dispatch-observability, push-hms-message-category, push-token-uniqueness -->

## Requirements

### Requirement: HMS 可见通知消息载荷

当 push-service 经 HMS channel 发送**非静默**且 `alert` 非空的推送时，请求体 MUST 包含顶层 `message.notification`（至少含 `title` 与 `body`），并且 MUST 在 `message.android.notification` 中提供合法的 `click_action`（默认打开应用：`type` 为 `3`；若配置了自定义 intent，则 `type` 为 `1` 且带 `intent`）。

系统 MUST NOT 仅依赖「只有 `android.data`、无顶层 notification」的形态作为可见通知的主路径。

#### Scenario: 预测临近可见推送

- **WHEN** `bizType=predict_imminent`（或其它非静默+非空 alert）经 HMS 下发
- **THEN** 发往华为的 JSON MUST 含 `message.notification.title` 与 `message.notification.body`，且 `message.android.notification.click_action` 合法

#### Scenario: 静默不伪造通知正文

- **WHEN** 推送为 silent 或 alert 为空
- **THEN** 系统 MUST NOT 以空/占位正文冒充完整用户可见通知消息（可仅 badge/data）

### Requirement: HMS 业务成功码判定

HMS `messages:send` 在 HTTP 2xx 时，系统 MUST 解析响应体业务字段 `code`：仅当成功码为 `80000000`（字符串或与之等价的表示）时，方可视为发送成功并记 `send_ok`；否则 MUST 视为失败并记 `send_failed`（日志含 `code` 与截断错误信息，MUST NOT 含完整 device token）。

#### Scenario: HTTP 200 但业务失败

- **WHEN** 华为返回 HTTP 200 且 `code` 非 `80000000`
- **THEN** dispatcher 路径 MUST 走失败日志（非 `send_ok`）

#### Scenario: 业务成功

- **WHEN** 华为返回 HTTP 2xx 且 `code` 为 `80000000`
- **THEN** 系统 MAY 记 `send_ok`

---

### Requirement: notify-service SHALL run as dedicated microservice

平台 SHALL 提供 **`notify-service`** 进程（原 app-status 维护通知能力），监听 `NOTIFY_SERVICE_ADDR` 默认 `:9806`，加载 `manifest/config/config.notify-service.yaml`。进程 **MUST NOT** 依赖 MySQL、Redis 或 RabbitMQ 启动探活。Docker 镜像名与 ACR 仓库名 **MUST** 为 `notify-service`。

#### Scenario: 启动与配置隔离

- **WHEN** 启动 notify-service 且未设置 `GF_GCFG_FILE`
- **THEN** 进程 SHALL 加载 `config.notify-service.yaml`，且配置 **MUST NOT** 含 `database` 段

#### Scenario: HTTP 契约保持不变

- **WHEN** 客户端 `GET /app/api/status/banner`
- **THEN** 路径与响应语义 SHALL 与 rename 前一致

### Requirement: docker-acr workflow SHALL build notify-service image

`.github/workflows/docker-acr.yml` MUST 将 `notify-service` 纳入全量构建矩阵（与 gateway、ucg 等并列，共 8 服务）。构建别名 **MUST** 接受 `notify` 与 `notify-service`，映射 Dockerfile `manifest/docker/Dockerfile.notify-service`。Tag `vX.Y.Z+notify` **MUST** 仅构建并 push `notify-service` 镜像。

#### Scenario: 全量 tag 含 notify-service

- **WHEN** push tag `v1.0.0-rc.1`（无 `+` 后缀）
- **THEN** CI matrix **MUST** 包含 `notify-service` 且 push `${REGISTRY}/notify-service:${tag}`

#### Scenario: 单服务 notify 构建

- **WHEN** push tag `v1.0.0-rc.2+notify` 或 workflow_dispatch `services=notify`
- **THEN** CI **MUST** 仅 build/push `notify-service` 镜像

---

### Requirement: APNs Provider JWT 进程内缓存与复用

push-service 经 APNs HTTP/2 发送推送时，MUST 使用 Apple Provider Authentication Token（JWT，`Authorization: bearer`）。同一进程内，对同一配置密钥（`keyId`/`teamId`/密钥材料）MUST 缓存并复用已签发的 JWT，MUST NOT 在每一次设备发送时重新签发带新 `iat` 的 JWT。

缓存的 JWT 刷新间隔 MUST 不短于 20 分钟，且 MUST 在签发后 60 分钟内完成下一次刷新（实现推荐约 45–50 分钟复用窗口）。并发获取 bearer 时 MUST 互斥，避免短窗内并行签发多把不同 `iat` 的 token。

本要求仅约束 Provider JWT；设备 device token 的注册、删除与无效判定语义 MUST NOT 因此改变。MUST NOT 为此引入新的 Redis 键或跨进程共享 JWT（除非未来独立变更明确批准）。

#### Scenario: 连续两次 APNs 发送复用同一 Provider JWT

- **WHEN** push-service 在缓存有效期内向两个不同 device token 连续发送 APNs 推送
- **THEN** 两次请求 MUST 使用同一 Provider JWT 字符串（或等价：同一 `iat`），MUST NOT 为第二次发送重新签发新 JWT

#### Scenario: 缓存过期后才换发

- **WHEN** 缓存中的 Provider JWT 已超过实现设定的复用窗口（且该窗口 ≥20 分钟、<60 分钟）
- **THEN** 下一次 APNs 发送 MUST 签发新 JWT 并更新缓存，供后续请求复用

#### Scenario: 并发发送不并行换发

- **WHEN** 多个 goroutine 在缓存未命中时同时请求 APNs bearer
- **THEN** 系统 MUST 经互斥使该窗口内至多成功签发一把新 JWT 并写入缓存，其余调用方复用该结果

### Requirement: TooManyProviderTokenUpdates 不得删除设备 token

当 APNs 响应为 HTTP 429 且 reason 为 `TooManyProviderTokenUpdates`（或等价 Provider token 更新过频语义）时，系统 MUST 将该次发送记为失败（如 `send_failed`），MUST NOT 将该 reason 视为设备 token 无效，MUST NOT 因此删除对应 `push_device` 行。

#### Scenario: 429 Provider 限流不删设备

- **WHEN** APNs 返回 status=429 且 reason=`TooManyProviderTokenUpdates`
- **THEN** dispatcher MUST 记发送失败，且 MUST NOT 调用按设备 id 删除 token 的路径

---

### Requirement: App MUST 能注册与注销推送 token

系统 MUST 经 gateway-app 对外提供中性路径注册/注销接口（如 `POST /app/api/push/register` 与 `POST /app/api/push/unregister`），请求体含 `channel`、`token`、`deviceKey`（注销可按既有字段）。接口 MUST 使用登录用户 `wxId`（网关注入内部头），MUST NOT 要求微信 unionid 绑定作为前提。系统 MUST NOT 再提供 `/ucg/app/api/push/register` 与 `/ucg/app/api/push/unregister` 作为有效推送注册入口。

#### Scenario: 登录用户注册成功

- **WHEN** 已登录用户以合法 channel/token/deviceKey 调用注册接口
- **THEN** push-service MUST 持久化该 token，且后续按该 wxId 发送可见推送时可投递到对应通道

#### Scenario: 旧 UCG 注册路径移除

- **WHEN** 客户端请求已删除的 `/ucg/app/api/push/register`
- **THEN** 系统 MUST NOT 再将其作为推送注册成功路径（404 或网关未反代至 push 实现）

### Requirement: Internal MUST 支持按 bizType 发送且 badge 由调用方提供

push-service MUST 提供经内部密钥鉴权的发送接口，入参至少含 `wxId`、`bizType`、可见文案（或静默标志）与可选 `badge`、`data`。当请求携带 `badge` 时，下发载荷 MUST 使用该值；未携带时 MUST 默认 `0`。push-service MUST NOT 为填充 badge 去查询 UCG 未读。无有效 token 时 MUST 跳过发送且 MUST NOT 视为调用方硬失败（可返回成功空投递或等价 best-effort 语义，并在实现中固定）。

#### Scenario: UCG 传入未读角标

- **WHEN** 调用方传入 `badge=5` 与非空 alert
- **THEN** 厂商载荷中的角标 MUST 为 5

#### Scenario: 预测类默认零角标

- **WHEN** 调用方不传 badge 或显式传 0
- **THEN** 厂商载荷角标 MUST 为 0

### Requirement: push 注册注销 MUST 不计入 App usage 统计

经 gateway-app 暴露的 push 注册与注销 App 接口 MUST 列入 usage 维护型排除（`maintenance_skip` 精确 METHOD+path）。负责人结论：不统计。旧 ucg push 排除项 MUST 在路由删除后移除或替换为新 path，避免残留无效键。

#### Scenario: 注册不写入使用统计

- **WHEN** 客户端成功或失败调用 `POST /app/api/push/register`
- **THEN** 该次调用 MUST NOT 计入 App API 使用统计报表

---

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

---

### Requirement: Internal by-biz-type 受理可观测

`POST /push/internal/api/by-biz-type` 处理路径 MUST 在以下情况输出 **Warning** 级日志（前缀建议 `[push]`），且 MUST NOT 打印完整 device token 或完整 alert 正文：

- 内部密钥校验失败
- 请求体/必填字段校验失败
- 校验通过并受理发送（含 `wxId`、`bizType`、`badge`、`silent`、`alertLen`）

#### Scenario: 受理成功有日志

- **WHEN** 合法内部请求到达且校验通过
- **THEN** push-service 日志中 MUST 出现一条 Warning，表明已 `accepted` 该 `wxId` 与 `bizType`

#### Scenario: 鉴权失败有日志

- **WHEN** 内部密钥无效或缺失
- **THEN** 在返回 403 的同时 MUST 打 Warning（`auth_denied` 或等价 reason）

### Requirement: Dispatch 与厂商发送可观测

异步 dispatch 路径 MUST：

- 在查询到待发设备列表后打 Warning，含 `wxId`、`bizType`、`deviceCount`（`deviceCount=0` 时沿用或等价于现有 `no_device` skip）
- 某 channel 因凭证未配置而跳过时打 **Warning**（不得仅 Debug）
- 某 token 发送成功时打 Warning（`send_ok` 或等价）
- 异步 goroutine panic 时打 Error/Warning，不得空 recover 吞掉

#### Scenario: 无注册设备

- **WHEN** 受理后该 `wxId` 在 `push_device` 无可用 token
- **THEN** 日志 MUST 以 Warning 标明 skip（如 `no_device`），且在 `GF_LOGGER_LEVEL=prod` 下仍可见

#### Scenario: 凭证未配置

- **WHEN** 存在设备记录但对应 channel 厂商凭证未配置
- **THEN** 日志 MUST 以 Warning 标明该 channel 因凭证未配置而跳过（`GF_LOGGER_LEVEL=prod` 下仍可见）

#### Scenario: 发送成功

- **WHEN** 某 channel sender 返回成功（无 error）
- **THEN** 日志 MUST 以 Warning 标明 `send_ok`（或等价），含 `channel`、`wxId`、`bizType`

---

### Requirement: HMS 推送 MUST 按 bizType 设置 android.notification.category

当 push-service 经 HMS 通道发送推送时，若请求载荷中的业务类型 `bizType`（`payload.Data["bizType"]`，大小写不敏感比较前 MUST 规范化为小写 trim）为已知可见类型，则 MUST 在 `message.android.notification` 中设置 `category` 为华为官方大写枚举；MUST NOT 将 `category` 写入顶层 `message.notification` 以外的错误位置（本需求仅约束 android.notification）。

映射 MUST 为：

| bizType（规范化后） | category |
|---------------------|----------|
| `predict_imminent`  | `WORK`   |
| `ucg_alert`         | `IM`     |

#### Scenario: 预测临近推送带 WORK

- **WHEN** HMS 发送且规范化后的 `bizType` 为 `predict_imminent`
- **THEN** 请求体中 `message.android.notification.category` MUST 为字符串 `WORK`

#### Scenario: UCG 可见内容推送带 IM

- **WHEN** HMS 发送且规范化后的 `bizType` 为 `ucg_alert`
- **THEN** 请求体中 `message.android.notification.category` MUST 为字符串 `IM`

### Requirement: HMS 静默角标与未知 bizType MUST NOT 写入 category

当规范化后的 `bizType` 为 `ucg_silent_badge`，或为空，或为上述映射表以外的任意值时，HMS 请求体中的 `message.android.notification` MUST NOT 包含 `category` 字段（键不存在，而非空字符串）。

#### Scenario: UCG 静默角标不传 category

- **WHEN** HMS 发送且规范化后的 `bizType` 为 `ucg_silent_badge`
- **THEN** `message.android.notification` MUST NOT 含有 `category` 键

#### Scenario: 未知或空 bizType 不传 category

- **WHEN** HMS 发送且 `bizType` 为空或未落入 `predict_imminent` / `ucg_alert` / `ucg_silent_badge` 的已知映射
- **THEN** `message.android.notification` MUST NOT 含有 `category` 键

---

### Requirement: 推送 token MUST 全局唯一且注册后来顶上

系统在持久化推送设备（`push_device`）时，厂商 `token` 字符串 MUST 在全表唯一（MUST NOT 仅按 `wx_id`+`device_key`+`channel` 唯一而允许多行同 token）。当已登录用户调用注册接口提交某 `token`，若库中已存在相同 `token` 的任意行（无论原 `wx_id`、`channel`、`device_key` 是否相同），系统 MUST 删除这些既有行（或等价保证最终仅保留当前注册结果），再为当前 `wxId` 写入/更新对应 `(device_key, channel)` 记录。同一用户使用不同手机产生不同 `token` 时，系统 MUST 允许并存多行。系统 MUST 在库层提供 `token` 唯一约束；启动迁移 MUST 清理存量重复 token（保留最近更新的一行）后再添加约束。

#### Scenario: 同机换号接管 token

- **WHEN** 账号 A 已注册 token `T`，随后账号 B 在同一设备以相同 token `T` 调用注册成功
- **THEN** 库中 MUST 不再存在账号 A 对该 `T` 的行，且 MUST 存在账号 B 持有 `T` 的行

#### Scenario: 同用户两部手机

- **WHEN** 同一 `wxId` 先后注册两个不同的 token `T1`、`T2`（不同设备）
- **THEN** 库中 MUST 同时保留 `T1` 与 `T2` 两行（在各自 deviceKey/channel 语义下）

#### Scenario: 同账号刷新注册

- **WHEN** 同一 `wxId`、同一 `deviceKey`、同一 `channel` 再次注册（token 可变或不变）
- **THEN** 系统 MUST 成功更新该键上的记录，且 MUST NOT 因本需求导致该用户失去该设备推送能力

---
