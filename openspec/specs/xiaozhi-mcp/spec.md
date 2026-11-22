# xiaozhi-mcp

## Purpose

小智 MCP 绑定、桥接、试用解锁、指南视频与 Hub 绑定 UI 的行为规格。 MCP bridge manager and speaker/device unlock gates.

<!-- migrated-from: mcp-feeding-ws-text, xiaozhi-mcp-baby-tool-desc, xiaozhi-mcp-binding, xiaozhi-mcp-binding-reconnect, xiaozhi-mcp-bridge-manager, xiaozhi-mcp-guide-videos, xiaozhi-mcp-hub-binding-ui, xiaozhi-mcp-hub-reconnect-ui, xiaozhi-mcp-permanent-unlock, xiaozhi-mcp-speaker-mac, xiaozhi-mcp-trial, xiaozhi-mcp-unlock-admin, xiaozhi-mcp-unlock-admin-list, xiaozhi-mcp-usage-gate -->

## Requirements

### Requirement: MCP 喂养工具改经 chat WS 文模式

mcp-service 的喂养对话工具（现名 `baby_feeding_advisor` 或等价）MUST 通过 `/voice/chat/ws` 完成一轮对话：使用 text 输入与 text 输出模态，发送工具入参文本，读取服务端 `answer`（及可选 `thinking_delta`）作为工具返回。MUST NOT 再调用 `DelegateTextChat` 或 `/voice/internal/api/text/chat`。

#### Scenario: 工具调用成功返回回复

- **WHEN** 小智侧调用 MCP 喂养工具并传入非空 `transcript`
- **THEN** mcp-service MUST 对配置的设备建立（或复用）chat WS，以文模式提交该文本，并在收到业务 `answer` 后将其作为工具结果返回

#### Scenario: 空 transcript 仍本地校验

- **WHEN** 工具收到空或仅空白的 transcript
- **THEN** mcp-service MUST 在本地返回错误，MUST NOT 无意义地占用 WS 意图轮次

### Requirement: MCP 不再依赖 history 文本委派

mcp-service MUST NOT import 或调用 `histsvc.DelegateTextChat` / `DelegateTextChatStream` 作为喂养实现。进程依赖与 runbook MUST 改为描述 WS 连接所需基址/密钥（若有），而非 internal text HTTP。

#### Scenario: 无 DelegateTextChat 调用点

- **WHEN** 检索 mcpbridge / mcp-service 喂养工具实现
- **THEN** MUST NOT 存在对 `DelegateTextChat` 的调用

---

### Requirement: 宝宝话题经同一 MCP 工具进入现有智能体

`xiaozhi-mcp-service` 暴露的工具 `baby_feeding_advisor` 的 description MUST 表明：用户谈到宝宝时调用本工具，范围不限于喂养记录。描述 MUST 覆盖喂养记录或查询、宝宝信息（至少包括名字、生日、月龄、性别、身高体重、睡眠、发育与日常照护），以及用户对上一条本工具操作的确认或取消短句（原话作为 `transcript`）。描述 MUST NOT 以多句「必须」堆叠选路。工具名、入参与执行路径 MUST 保持为既有 `ChatHandler` → `/voice/chat/ws` 文模式，MUST NOT 新增 MCP 工具。

#### Scenario: tools/list 描述含宝宝信息且不限喂养

- **WHEN** 客户端请求 `tools/list`
- **THEN** `baby_feeding_advisor` 的 description 包含宝宝信息示例，并说明调用范围不限于喂养
- **AND** 仍为唯一聊天工具名 `baby_feeding_advisor`

#### Scenario: 确认短句仍走同一工具

- **WHEN** 阅读该 description 中关于确认或取消的条目
- **THEN** 短句示例仍要求调用本工具并把用户原话放入 `transcript`

---

### Requirement: 小智 MCP 绑定存储

系统 SHALL 在 device 域持久化用户的小智 MCP 绑定，每条记录包含归属 `wx_id`、喂养落点 `device_no`、小智接入点 `mcp_token`、用户备注 `alias` 以及状态与时间戳。`mcp_token` MUST 在全局唯一（跨 wx）。

#### Scenario: 唯一约束

- **WHEN** 已存在某 `mcp_token` 的有效绑定
- **THEN** 另一账号（或其他记录）再次以相同 token 添加 MUST 失败并返回明确业务错误
- **AND** MUST NOT 写入第二条相同 token 的绑定

### Requirement: App 添加小智绑定

系统 SHALL 提供经 gateway-app 鉴权的 App 接口，供当前登录用户添加小智绑定。客户端 MUST 提交 `mcpToken` 与 `alias`（备注）；服务端 MUST 使用当前账号已绑定的宝宝 `deviceNo` 作为落点，MUST NOT 信任客户端随意指定他人 `deviceNo`。

#### Scenario: 成功添加

- **WHEN** 当前 wx 已绑定非空 `deviceNo`，且 `mcpToken` 非空、全局未被占用，`alias` 符合校验
- **THEN** 系统写入绑定行，`device_no` 等于该 wx 当前 `device_no`
- **AND** 写成功后 MUST 调用 xiaozhi-mcp-service 内部 Upsert（失败不回滚已写入绑定，依赖 reconcile 兜底；实现可记录告警日志）

#### Scenario: 未绑定宝宝

- **WHEN** 当前 wx 的 `device_no` 为空
- **THEN** 添加 MUST 失败，MUST NOT 写入绑定

#### Scenario: Token 冲突

- **WHEN** `mcpToken` 已被其他绑定使用
- **THEN** 添加 MUST 失败并提示 token 已被占用

### Requirement: App 列表与改备注、删除

系统 SHALL 提供经 gateway-app 鉴权的 App 接口：列出当前用户的小智绑定、更新备注、删除绑定。列表与详情 MUST NOT 返回完整 `mcp_token`，MUST 返回脱敏展示（如前后若干字符）及 `alias`、`deviceNo`、状态等。

#### Scenario: 列表脱敏

- **WHEN** 用户请求绑定列表
- **THEN** 响应每条含 id、alias、脱敏 token、deviceNo 等
- **AND** 响应 MUST NOT 包含完整 mcp_token 明文

#### Scenario: 更新备注

- **WHEN** 用户更新属于自己的绑定的 alias
- **THEN** 系统更新该行备注并成功返回
- **AND** MUST NOT 允许修改他人绑定

#### Scenario: 删除绑定

- **WHEN** 用户删除属于自己的绑定
- **THEN** 系统删除（或逻辑删除）该行
- **AND** 写成功后 MUST 调用 xiaozhi-mcp-service 内部 Remove，使对应 Bridge 停止

### Requirement: 内部全量列表

device-service SHALL 提供仅供内网调用的接口，返回全部 active 小智绑定（含完整 token 与 deviceNo），供 xiaozhi-mcp-service 启动与 reconcile 使用。该接口 MUST 校验内部密钥，MUST NOT 经 gateway-app 对 App 暴露。

#### Scenario: 内部拉取

- **WHEN** xiaozhi-mcp-service 携带合法内部密钥请求全量 active 绑定
- **THEN** 返回所有应用于拨号的绑定字段（含完整 mcp_token）

### Requirement: 小智绑定 App 接口计入 usage 统计

经 gateway-app 对外暴露的小智绑定增删改查 App HTTP 接口 MUST 计入 App API 使用统计。实现 MUST 在 `api/v1`（或规定版本）登记完整 `g.Meta` path/method/summary；MUST NOT 将这些路径写入 `usagestats/maintenance_skip.go`。

#### Scenario: 统计纳入

- **WHEN** 运维查看功能使用统计且用户调用了小智绑定列表或增删改接口
- **THEN** 对应 path 出现在 usage 统计中（按现有归一化规则）
- **AND** maintenance_skip 中 MUST NOT 包含这些精确 path

### Requirement: gateway-app 登记

新增小智绑定 App 路由 MUST 落在已有 device App 反代前缀内（或按项目约定补 Bind）；若需 Bearer，MUST NOT 误加入 auth exempt。proposal/tasks 自检：反代、exempt、usage 结论（本能力为「统计」）。

#### Scenario: 经网关可达

- **WHEN** App 携带合法 Bearer 调用小智绑定 CRUD 路径
- **THEN** 请求经 gateway-app 到达 device-service 并按鉴权头识别当前 wx

---

### Requirement: App 强制重连小智绑定

系统 SHALL 提供经 gateway-app 鉴权的 App 接口 `POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect`，供当前登录用户对属于自己的 active 小智绑定触发立即重拨。该接口 MUST 使用新增的请求/响应类型登记 `g.Meta`；MUST NOT 修改既有 List/Add/Alias/Delete 的 v1 结构体字段。

#### Scenario: 归属用户成功触发重连

- **WHEN** 当前 wx 拥有 id 对应的 active 绑定，并携带合法 Bearer 调用 reconnect
- **THEN** device-service MUST 加载该行的 mcpToken、deviceNo、wxId
- **AND** MUST 经 `clients/xiaozhimcp` 调用 xiaozhi-mcp-service 内部 reconnect（禁止 device 直连 mcpbridge 包）
- **AND** 内部侧 MUST 取消既有会话（若有）并 `startLocked` 启动新会话，即使 token/deviceNo/wxId 与当前会话相同
- **AND** MUST 重置该会话退避并立即发起 redial
- **AND** App 接口 MUST 在触发完成后快速返回成功（MUST NOT 阻塞等待 WebSocket 已连通）

#### Scenario: 非归属或无效绑定

- **WHEN** 绑定不存在、非 active，或不属于当前 wx
- **THEN** reconnect MUST 失败并返回明确业务错误
- **AND** MUST NOT 调用内部 reconnect，MUST NOT 修改绑定行

#### Scenario: 同 key 的 Upsert 不足以重连

- **WHEN** 会话已存在且 token/deviceNo/wxId 未变
- **THEN** 普通 Upsert MAY 继续 noop
- **AND** reconnect/ForceRestart MUST 仍执行 cancel + startLocked（不得仅依赖无 force 的 Upsert）

### Requirement: Entitlement 不阻断既有绑定重连

系统 SHALL 允许对「绑定行 active 且归属当前 wx」的记录执行 reconnect，即使该 wx 的小智 MCP entitlement 已过期或未解锁。MUST NOT 在 reconnect 路径强制调用开通/加购校验（如 EnsureAccessForAdd）。

#### Scenario: 未解锁仍可重连自有绑定

- **WHEN** 用户 entitlement 非 active，但仍有归属自己的 active 绑定行
- **THEN** reconnect MUST 仍可成功触发内部 ForceRestart
- **AND** 若后续因 entitlement/usage 策略被 Manager 侧 lazy-remove，属既有链路行为，本接口不额外保证长期 connected

### Requirement: gateway-app 可达与鉴权

新增 reconnect App 路由 MUST 落在既有 device App 反代前缀内；MUST 要求 Bearer；MUST NOT 加入 `gateway_app_auth_exempt`。usage 是否计入 MUST 在实现前向负责人确认；未获明确答复前 MUST NOT 修改 `usagestats/maintenance_skip.go`。

#### Scenario: 经网关需登录

- **WHEN** 客户端无合法 App Bearer 调用 reconnect
- **THEN** gateway-app MUST 拒绝（与其它 device App 写接口一致）
- **AND** 该 path MUST NOT 出现在 auth exempt 列表

### Requirement: 内部重连接口与客户端

`xiaozhi-mcp-service` SHALL 提供仅内网、需内部密钥的 `POST /xiaozhi-mcp/internal/api/bindings/reconnect`（路径可与实现微调但语义等价）。`internal/clients/xiaozhimcp` SHALL 提供对应出站方法供 device-service 调用。

#### Scenario: 内部鉴权

- **WHEN** 请求缺少合法内部密钥
- **THEN** reconnect 内部接口 MUST 拒绝
- **AND** MUST NOT 经 gateway-app 对 App 直接暴露该内部路径

---

### Requirement: 进程命名为 xiaozhi-mcp-service

系统 SHALL 将原 `mcp-service` 进程、镜像与部署资源命名为 `xiaozhi-mcp-service`（含 cmd 入口、Dockerfile、compose、kustomize、CI 镜像映射、配置文件名与服务日志/OTEL 标识）。历史上无生产用户，允许直接改名而无并行双进程兼容期。

#### Scenario: 部署标识

- **WHEN** 查看生产/测试编排中的该服务
- **THEN** 服务名与镜像名体现 `xiaozhi-mcp-service`（或项目统一的等价命名）
- **AND** MUST NOT 再依赖名为 `mcp-service` 的主部署资源作为稳态入口

### Requirement: 单副本部署

`xiaozhi-mcp-service` 的 Deployment MUST 约定 `replicas: 1`，以保证同一 `mcp_token` 全球至多一条由本系统发起的出站 MCP 连接。

#### Scenario: 副本数

- **WHEN** 应用本变更后的 kustomize/compose 配置
- **THEN** xiaozhi-mcp-service 副本数为 1

### Requirement: Bridge Manager 多连接

`xiaozhi-mcp-service` SHALL 在进程内维护 Manager：为每条 active 绑定建立独立 Bridge（token → 小智接入点长连接；deviceNo → `ChatViaVoiceWS` 落点）。每条 Bridge MUST 沿用断线指数退避重连，且互相隔离（单条失败 MUST NOT 拖垮整个进程）。

#### Scenario: 多 token 并存

- **WHEN** 内部状态中存在两条不同 token 的绑定
- **THEN** Manager MUST 分别维持两条出站连接
- **AND** 小智经某 token 对应连接发起 `tools/call` 时，喂养对话 MUST 使用该绑定的 `deviceNo`

#### Scenario: 移除绑定停连

- **WHEN** Manager 收到对该 token 的 Remove（或 reconcile 发现绑定已不存在）
- **THEN** MUST 取消该 Bridge 的 ctx 并停止重连
- **AND** MUST 从活跃 map 中移除

### Requirement: 内部 HTTP 增量同步

`xiaozhi-mcp-service` SHALL 监听仅内网可达的 HTTP 端口，提供 Upsert/Remove（及健康检查）接口，供 device-service 在绑定写成功后推送。接口 MUST 校验内部密钥。

#### Scenario: Upsert 启桥

- **WHEN** 合法内部请求 Upsert 携带 token 与 deviceNo
- **THEN** Manager 若不存在该 token 则启动 Bridge；若已存在则可更新 deviceNo 或按实现重建会话
- **AND** 非法密钥 MUST 拒绝

#### Scenario: Remove 停桥

- **WHEN** 合法内部请求 Remove 指定 token（或绑定 id）
- **THEN** Manager 停止对应 Bridge

### Requirement: 启动全量拉取与低频 reconcile

`xiaozhi-mcp-service` 启动时 MUST 从 device-service 内部全量 list 拉取 active 绑定并 Reconcile。进程 MUST 运行已批准的后台任务 `xiaozhi-mcp-binding-reconcile`：按可配置间隔再次全量拉取并对齐 Manager（默认分钟级；间隔为 0 时关闭周期 reconcile，仍保留启动全量与写路径推送）。失败 MUST 打日志且 MUST NOT 退出进程。

#### Scenario: 启动对齐

- **WHEN** 进程启动且 device 内部 list 返回若干绑定
- **THEN** Manager 为返回的每条绑定建立 Bridge，且不保留 list 中不存在的陈旧 Bridge

#### Scenario: reconcile 补齐丢推送

- **WHEN** 某绑定已在 DB 但因写路径 HTTP 失败未 Upsert
- **THEN** 下一次 reconcile MUST 发现并启动对应 Bridge

### Requirement: 内部 Service 暴露

编排 MUST 为 `xiaozhi-mcp-service` 提供 ClusterIP（或 compose 网络别名）Service，使 device-service 能解析并调用其内部 HTTP。对外 App 流量 MUST NOT 经 gateway-app 暴露这些内部路径。

#### Scenario: device 可调用

- **WHEN** device-service 配置的 xiaozhi-mcp 基址指向集群内 Service
- **THEN** 写路径 Upsert/Remove 可到达唯一副本

### Requirement: 喂养工具行为保持

小智经 MCP 调用喂养顾问工具时，行为 MUST 继续经 `/voice/chat/ws` 文模式完成对话并返回 answer；deviceNo MUST 来自该 Bridge 绑定，而非进程级单一 env（稳态）。工具入参中的 `xzDeviceNo` 若存在，MUST NOT 覆盖绑定 deviceNo（除非后续变更另有规格）。

#### Scenario: 对话落点

- **WHEN** 绑定 token=T1、deviceNo=D1 的 Bridge 收到非空 transcript 的 tools/call
- **THEN** 系统以 D1 调用 voice chat WS 并返回 answer

---

### Requirement: cash SHALL 持久化小智引导视频双固定槽

cash-service MUST 维护固定两槽引导视频元数据，槽标识 MUST 为 `promo`（能力宣传）与 `tutorial`（绑定教学）。每槽 MUST 含：可编辑 `title`、可选视频 `object_key`（库内仅存 OSS objectKey，MUST NOT 存完整 CDN URL 作为权威源）、`updated_at`。空 `object_key` MUST 表示该槽未配置。系统 MUST 在 schema 初始化时种子两行（可带默认中文标题），Hub 可随后改标题。

#### Scenario: 种子两槽存在

- **WHEN** cash schema 初始化完成
- **THEN** `promo` 与 `tutorial` 两行 MUST 存在且均可被 Admin 读取

#### Scenario: 清空 object_key 表示未配置

- **WHEN** 某槽 `object_key` 为空
- **THEN** App 列表 MUST NOT 包含该槽

### Requirement: Hub Admin SHALL 管理引导视频标题与上传

系统 MUST 提供经 cash Admin 鉴权的接口，支持：列出两槽（含 CDN 预览 URL）、更新某槽标题、multipart 上传/覆盖某槽视频、清空某槽视频（清除 `object_key`）。上传 MUST 经 `clients/ucg` 写入 OSS（建议前缀 `xiaozhi-guide/`），MUST 使用与 UCG App 视频上传相同的 **v1 验真与体积上限**，MUST NOT 走服务端转码，MUST NOT 落 gateway 本机盘。Hub UI MUST 挂在现有「小智 MCP 开通」页（区 D），MUST NOT 要求新建独立导航模块（可复用既有开通页入口）。

#### Scenario: 运维上传合规视频

- **WHEN** 运维对 `tutorial` 槽上传通过 v1 验真的 mp4
- **THEN** 该槽 `object_key` MUST 更新，Admin GET MUST 返回可访问的 `videoCdnUrl`

#### Scenario: 不合规视频拒绝

- **WHEN** 运维上传未通过 v1 验真的文件
- **THEN** API MUST 返回错误且 MUST NOT 更新该槽已有 `object_key`（若曾配置）

#### Scenario: 改标题不强制已有视频

- **WHEN** 运维仅更新 `promo` 标题且该槽尚无视频
- **THEN** 标题 MUST 保存成功

### Requirement: App SHALL 返回已配置引导视频的 CDN 列表

系统 MUST 提供 `GET /cash/app/api/xiaozhi-mcp/guide-videos`（Bearer）。响应 `list` MUST 仅含 `object_key` 非空的槽，每项至少含 `slot`、`title`、`videoCdnUrl`；MAY 含 `coverCdnUrl`（视频首帧缩略图 CDN）。该接口 MUST 计入 gateway-app usage 统计（MUST NOT 加入维护型 skip 列表）。

#### Scenario: 仅一槽有视频

- **WHEN** 仅 `promo` 已配置视频
- **THEN** `list` 长度 MUST 为 1 且含 `slot=promo`

#### Scenario: 未登录拒绝

- **WHEN** 请求无有效 App Bearer
- **THEN** MUST 拒绝（MUST NOT 匿名放行）

### Requirement: Flutter 小智页 SHALL 按清单展示或隐藏入口

Flutter 客户端（`flutter_ai_talk`）在小智音箱相关页面 MUST 请求引导视频清单。当 `list` 为空时 MUST **完全不展示**引导视频入口。当存在一项或两项时 MUST 按返回的 `title` 展示可点入口，点击后 MUST 使用返回的 `videoCdnUrl` 播放。MUST NOT 在双槽皆空时展示占位「暂无视频」入口块（与「完全不露」一致）。

#### Scenario: 双空隐藏

- **WHEN** App 返回 `list` 为空数组
- **THEN** 小智页 MUST 不渲染引导视频入口 UI

#### Scenario: 有视频可点播

- **WHEN** `list` 含 `tutorial` 项
- **THEN** 用户点击对应入口后客户端 MUST 能打开该 `videoCdnUrl` 进行播放

---

### Requirement: 详情页小智绑定依赖 App 登录

Hub 设备详情页（`/device/admin/history/{deviceNo}`）SHALL 提供「小智 MCP 绑定」区块。未完成 App 用户登录时，MUST NOT 允许发起添加/改备注/删除绑定请求；MUST 提示需先 App 登录。

#### Scenario: 未登录禁用写操作

- **WHEN** 操作者仅持有 Hub Admin 会话、未 App 登录
- **THEN** 小智绑定添加/改备注/删除控件 MUST 不可用或提交被前端拒绝
- **AND** MUST 展示需 App 登录的说明

#### Scenario: App 登录后可操作

- **WHEN** 操作者在详情页完成 App 用户名密码登录并取得 App JWT
- **THEN** 小智绑定区块 MUST 允许调用绑定列表与写操作（在设备号一致前提下）

### Requirement: 复用 App 小智绑定 API

小智绑定 UI MUST 使用既有 App 接口完成 CRUD，且请求携带 App Bearer（非 Admin JWT）：

- 列表：`GET /device/app/api/xiaozhi-mcp/bindings`
- 添加：`POST /device/app/api/xiaozhi-mcp/bindings`（`mcpToken`、`alias`）
- 改备注：`PUT /device/app/api/xiaozhi-mcp/bindings/{id}/alias`
- 删除：`DELETE /device/app/api/xiaozhi-mcp/bindings/{id}`

MUST NOT 新增依赖 Admin JWT 的小智绑定业务 API 作为本能力的实现路径。

#### Scenario: 添加绑定

- **WHEN** App 已登录、设备号一致，操作者提交非空 mcpToken 与 alias
- **THEN** 前端 MUST 以 App Bearer 调用 POST bindings
- **AND** 成功后刷新列表

#### Scenario: 删除绑定

- **WHEN** App 已登录，操作者删除列表中某条绑定
- **THEN** 前端 MUST 以 App Bearer 调用 DELETE bindings/{id}
- **AND** 成功后自列表移除或刷新

### Requirement: Token 脱敏展示

列表与详情展示 MUST 使用接口返回的脱敏字段（如 `tokenMask`）；MUST NOT 在页面上持久展示或回显完整 `mcpToken`（添加表单输入框在提交前的明文输入除外）。

#### Scenario: 列表无全文 token

- **WHEN** 绑定列表加载成功
- **THEN** 每行展示 alias 与 tokenMask（或等价脱敏）
- **AND** 页面 DOM/文案 MUST NOT 包含完整 mcpToken

### Requirement: 登录设备与页内设备一致

App 登录后，系统 MUST 比较登录账号绑定的 `deviceNo` 与当前详情页 `deviceNo`。不一致时 MUST 向操作者告警，且 MUST 禁止添加小智绑定（避免串号落库）。

#### Scenario: 串号禁止添加

- **WHEN** App 登录成功但账号 deviceNo 与 URL 中 deviceNo 不同
- **THEN** UI MUST 提示不一致
- **AND** MUST NOT 成功发起添加绑定请求

#### Scenario: 一致允许添加

- **WHEN** App 登录账号 deviceNo 与当前页 deviceNo 相同（忽略首尾空白）
- **THEN** 在已登录前提下允许添加绑定

---

### Requirement: Hub 小智区断连显示重连

Hub 设备详情页（`resource/public/history.html`）小智 MCP 绑定列表 SHALL 在条目 `connected` 为 false 时展示「重连」操作；为 true 时 MUST NOT 展示该操作（或展示为不可用）。

#### Scenario: 未连接显示重连

- **WHEN** 绑定列表某行 `connected === false` 且操作者已 App 登录
- **THEN** 该行 MUST 显示可用的「重连」控件

#### Scenario: 已连接不显示重连

- **WHEN** 绑定列表某行 `connected === true`
- **THEN** 该行 MUST NOT 提供可点击的「重连」（可不渲染或禁用）

### Requirement: Hub 重连调用 App API

「重连」MUST 以 App Bearer 调用 `POST /device/app/api/xiaozhi-mcp/bindings/{id}/reconnect`；MUST NOT 新增依赖 Admin JWT 的重连业务 API。成功后 MUST 刷新绑定列表以更新 `connected` 展示。

#### Scenario: 点击重连

- **WHEN** 操作者点击某行「重连」且 App 已登录
- **THEN** 前端 MUST 对该 id 发起上述 POST
- **AND** 成功后 MUST 重新拉取列表
- **AND** 失败时 MUST 向操作者展示错误信息

#### Scenario: 未登录禁用

- **WHEN** 操作者未完成 App 登录
- **THEN** 「重连」MUST 不可用或提交被前端拒绝（与添加/改备注/删除一致）

---

### Requirement: wx 维永久小智 MCP 能力权益

系统 SHALL 以微信账号（`wx_id`）为开通主体持久化小智 MCP 绑定能力权益。权益一经有效开通 MUST 视为永久（无到期续期模型）。开通来源 MUST 支持支付履约与 Admin 手工授。

#### Scenario: 支付开通后永久有效

- **WHEN** 某 wx 的小智 MCP 订单支付履约成功
- **THEN** 系统为该 `wx_id` 写入（或保持）有效永久权益
- **AND** MUST NOT 写入限时 `expires_at` 依赖 `ActivateFeature` 天数模型

#### Scenario: 未开通无权益

- **WHEN** 某 `wx_id` 从未支付成功且未被手工授
- **THEN** 该账号 MUST 视为未开通小智 MCP 能力

### Requirement: 支付通道复用与 MCP 类型履约

系统 SHALL 复用现有功能开通支付通道（建单调起、支付宝/Apple 回调与验单）处理小智 MCP 买断。建单金额与 Apple 商品 ID MUST 取自小智 MCP 专用 SKU 表。履约时 MUST 识别 MCP 商品类型并授予永久权益，MUST NOT 调用要求「授予天数≥1」的通用功能开通路径。

#### Scenario: MCP 订单履约

- **WHEN** 支付回调或验单命中小智 MCP 的 `product_code` 且订单可履约
- **THEN** 系统标记订单已支付并为订单 `wx_id` 授予永久小智 MCP 能力
- **AND** MUST NOT 调用 `ActivateFeature` 对 care/growth 等功能写权益

#### Scenario: 非 MCP 功能订单不受影响

- **WHEN** 订单商品属于既有 `feature_product` 功能 SKU
- **THEN** 履约行为 MUST 保持原功能开通语义（限时等）

### Requirement: App 开通态与可购 SKU

系统 SHALL 提供经 gateway-app 鉴权的 App 接口，返回当前登录用户是否已开通小智 MCP 能力，以及可售永久 SKU（含 `productCode`、价格、`appleProductId` 等）。该接口 MUST NOT 依赖 `GET /cash/app/api/feature/catalog`。

#### Scenario: 已开通

- **WHEN** 当前 wx 已有有效永久权益并请求开通态
- **THEN** 响应 `unlocked=true`（或等价字段）

#### Scenario: 未开通且 SKU 上架

- **WHEN** 当前 wx 未开通且专用 SKU `status` 为上架
- **THEN** 响应 `unlocked=false` 且包含可建单所需 SKU 信息

### Requirement: App 建单

系统 SHALL 提供 App 建单接口，使用与功能开通一致的支付渠道参数创建小智 MCP 订单，并返回调起支付所需信息。订单 MUST 关联当前登录 `wx_id` 与 MCP `product_code`。

#### Scenario: 成功建单

- **WHEN** 用户已登录、SKU 上架且渠道合法
- **THEN** 系统创建待支付订单并返回调起参数

#### Scenario: 已开通再建单

- **WHEN** 用户已拥有有效永久权益仍请求建单
- **THEN** 系统 MUST 拒绝或幂等提示已开通（实现选一并在错误文案中明确），MUST NOT 重复收费履约出第二份冲突权益语义

### Requirement: 添加绑定须已开通

device 域添加小智 MCP 绑定（App `POST` 添加）MUST 在写入绑定前确认当前 `wx_id` 已具备有效小智 MCP 永久能力。校验 MUST 经 cash 服务接口契约完成，MUST NOT 在 device 进程直查 cash 库表。List、改备注、删除 MUST NOT 因未开通而拒绝。

#### Scenario: 未开通拒绝添加

- **WHEN** 当前 wx 未开通且请求添加绑定
- **THEN** 添加 MUST 失败并返回明确业务错误
- **AND** MUST NOT 写入绑定行

#### Scenario: 已开通允许添加

- **WHEN** 当前 wx 已开通且其它绑定校验通过
- **THEN** 添加 MUST 按既有绑定规则成功（含 speaker_mac / token 唯一等）

#### Scenario: 未开通仍可列表与删除

- **WHEN** 当前 wx 未开通但已有历史绑定行
- **THEN** 列表、改备注、删除 MUST 仍可按既有归属规则执行

### Requirement: 支付退款撤销能力

当小智 MCP 支付订单发生渠道退款时，系统 MUST 撤销对应 wx 的小智 MCP 永久能力。已存在的绑定行 MAY 保留，但后续 Add MUST 因未开通而失败。

#### Scenario: 退款后不可再添加

- **WHEN** MCP 订单退款处理完成
- **THEN** 该 wx 开通态为未开通
- **AND** 再次 Add MUST 失败

---

### Requirement: 添加绑定必填音箱 MAC

系统 SHALL 在创建/更新小智 MCP 绑定时要求客户端提供音箱 MAC。服务端 MUST 将 MAC 规范为小写冒号分隔的六段形式（示例：`3c:dc:75:fc:7f:c4`）。格式非法时 MUST 拒绝并返回明确错误。

#### Scenario: 合法 MAC 规范化

- **WHEN** 客户端提交 `speakerMac` 为 `3C-DC-75-FC-7F-C4` 或等价可解析形式，且其他字段合法
- **THEN** 持久化的 `speaker_mac` MUST 为 `3c:dc:75:fc:7f:c4`

#### Scenario: 非法 MAC

- **WHEN** `speakerMac` 无法解析为 6 字节 MAC
- **THEN** 添加 MUST 失败，MUST NOT 写入绑定行

#### Scenario: 缺少 MAC

- **WHEN** 添加请求未提供非空 `speakerMac`
- **THEN** 添加 MUST 失败

### Requirement: MAC 全局唯一且一宝宝多音箱

`speaker_mac` MUST 在全局唯一，以保证一台音箱只绑定一个宝宝（`device_no`）。同一 `device_no` MUST 允许存在多条不同 MAC 的绑定。

#### Scenario: 多音箱同宝宝

- **WHEN** 同一账号宝宝下先后添加两个不同合法 MAC
- **THEN** 两条绑定均 MUST 成功存在

#### Scenario: MAC 已被其他宝宝占用

- **WHEN** 某 MAC 已绑定 deviceNo=A，另一账号宝宝 deviceNo=B 再次以该 MAC 添加
- **THEN** 添加 MUST 失败并提示该音箱已绑定其他宝宝
- **AND** MUST NOT 修改 A 上的绑定

### Requirement: 同 MAC 再添加自动更新 token

当规范化后的 MAC 已存在且归属当前宝宝时，再次调用添加接口 MUST 更新该行的 `mcp_token`（经既有 token 规范化）及本次提交的 `alias`（若提供），MUST NOT 再插入新行。若 token 发生变化，系统 MUST 通知 xiaozhi-mcp-service 停止旧 token 对应 Bridge 并为新 token 建立 Bridge。

#### Scenario: 换 token

- **WHEN** 宝宝 D 已绑定 MAC=M、token=T1，再次添加 MAC=M、token=T2（T2≠T1，可含完整接入点 URL）
- **THEN** 该行 mcp_token MUST 变为规范化后的 T2
- **AND** 行数 MUST 不增加
- **AND** mcp 侧 MUST 不再使用 T1 拨号，并使用 T2 拨号

#### Scenario: 同 token 再提交

- **WHEN** 同 MAC、同宝宝、token 规范化后与库中一致再次添加
- **THEN** MUST 成功（可更新 alias），MUST NOT 报「token 已被绑定」误伤自身

### Requirement: 列表展示 MAC

App 绑定列表响应 MUST 包含每条绑定的 `speakerMac`（规范化形式）。Hub 设备详情小智绑定 UI MUST 提供 MAC 输入，并在列表中展示 MAC。

#### Scenario: 列表含 MAC

- **WHEN** 用户请求绑定列表且存在绑定行
- **THEN** 每条 MUST 含 `speakerMac` 字段

### Requirement: mcp 进程内维护 token 连接状态

xiaozhi-mcp-service MUST 在进程内存中按规范化后的 `mcp_token` 维护是否已连接小智 MCP WebSocket 的布尔状态。`connected=true` MUST 仅表示当前处于拨号成功后的读循环中；拨号失败、退避重连、会话已 Remove 或无会话 MUST 为未连接。系统 MUST NOT 使用 Redis 持久化该状态。

#### Scenario: 拨号成功标绿

- **WHEN** Bridge 对某 token 拨号成功并进入读循环
- **THEN** 该 token 的连接状态 MUST 为已连接

#### Scenario: 断线或停桥标红

- **WHEN** 读循环结束、Remove、或会话不再存在
- **THEN** 该 token 的连接状态 MUST 为未连接

### Requirement: 内部 HTTP 批量查询连接状态

xiaozhi-mcp-service MUST 提供经内部密钥鉴权的 HTTP 接口，接受一组 token 并返回各自是否已连接。未知 token MUST 视为未连接。

#### Scenario: 批量查询

- **WHEN** 调用方携带合法内部密钥提交若干 token
- **THEN** 响应 MUST 为每个 token 给出布尔连接结果
- **AND** 未在内存中的 token MUST 为未连接

### Requirement: 绑定列表返回连接状态

device-service 在返回小智 MCP 绑定列表时 MUST 对每条绑定按规范化 token 查询 mcp 连接状态，并在列表项中返回 `connected`（布尔）。查询 mcp 失败或超时时，列表请求 MUST 仍成功，相关项 `connected` MUST 为 `false`（不得因状态查询失败使整个列表 5xx）。

#### Scenario: 已连接

- **WHEN** 绑定 token 在 mcp 侧为已连接
- **THEN** 列表项 `connected` MUST 为 `true`

#### Scenario: 未连接或查询失败

- **WHEN** token 未连接，或 device 调用 mcp 状态接口失败
- **THEN** 对应列表项 `connected` MUST 为 `false`
- **AND** 绑定列表 HTTP 仍 MUST 成功返回

### Requirement: Hub 绿红灯展示连接状态

Hub 设备详情小智绑定列表 MUST 根据每条 `connected` 展示连接指示：已连接为绿灯，未连接为红灯。系统 MUST NOT 要求第三态（黄灯/重连中）。

#### Scenario: 绿灯

- **WHEN** 列表项 `connected` 为 `true`
- **THEN** Hub UI MUST 显示绿色连接指示

#### Scenario: 红灯

- **WHEN** 列表项 `connected` 为 `false` 或缺失
- **THEN** Hub UI MUST 显示红色连接指示

### Requirement: 喂养确认与取消短句须选路本工具

xiaozhi-mcp-service 暴露的 `baby_feeding_advisor` 工具描述 MUST 要求：当用户对上一条喂养相关操作进行确认或取消（含短确认词如「是的」「对的」「确定」「好的」「可以」「没问题」，及取消词如「取消」「不要」「算了」「不对」）时，智能体 MUST 调用本工具并将用户原话作为 `transcript`，MUST NOT 自行回答。系统 MUST NOT 为此新增独立 MCP 工具；执行路径 MUST 仍为既有 `ChatHandler` → voice chat WS。

#### Scenario: 描述包含确认与取消选路

- **WHEN** 客户端请求 `tools/list`
- **THEN** `baby_feeding_advisor` 的 description MUST 包含对喂养操作确认与取消短句须调用本工具的说明
- **AND** MUST 仍仅有该喂养顾问工具（不因本需求新增第二个工具名）

---

### Requirement: 每账号一次试用 24 小时

系统 SHALL 为小智 MCP 能力提供按 `wx_id` 计的免费试用：每账号仅可成功领取一次，领取后权益有效期 MUST 为 24 小时（与既有商业功能试用时长常量对齐）。试用 MUST NOT 写入永久权益（`expires_at` MUST 为未来截止时间，MUST NOT 为 0）。

#### Scenario: 首次领取得到限时权益

- **WHEN** 某 wx 试用状态为未使用且尚无有效小智 MCP 权益，并成功完成试用领取
- **THEN** 系统标记该账号试用已使用，并写入有效权益且 `expires_at` 约为领取时刻起 24 小时
- **AND** MUST NOT 将 `expires_at` 记为 0（永久）

#### Scenario: 不可重复领取

- **WHEN** 某 wx 已成功领取过试用（无论当前权益是否已过期）
- **THEN** 再次试用领取 MUST 失败或被跳过且不再授予新的试用时段

### Requirement: 首次成功 Add 时领取试用

系统 SHALL 在 App 添加小智绑定的成功路径上，对「无有效权益且试用未用」的账号执行试用领取。领取与绑定写入 MUST 保证：领取失败则 MUST NOT 留下本次新增绑定；已有有效权益（试用未过期或永久）时 MUST NOT 再次消耗试用。

#### Scenario: 未开通未试用时 Add 触发 claim

- **WHEN** 当前 wx 无有效权益、试用未用，且绑定其它校验通过
- **THEN** 系统先（或原子地）完成试用领取，再写入绑定
- **AND** 写成功后该 wx 在试用窗内视为已开通

#### Scenario: 试用已用且已过期时拒绝 Add

- **WHEN** 当前 wx 试用已用尽且当前无有效权益
- **THEN** 添加绑定 MUST 失败
- **AND** MUST NOT 写入新绑定行

#### Scenario: 永久或试用有效时 Add 不消耗试用

- **WHEN** 当前 wx 已有有效永久或未过期试用权益并添加绑定
- **THEN** 添加按既有规则成功
- **AND** MUST NOT 再次将试用从 unused 转为 used（若已 used 则保持）

### Requirement: 开通态暴露试用信息

App 小智 MCP 开通态接口 SHALL 返回是否已开通、试用是否仍可领取，以及当前权益到期时间（永久时以约定哨兵如 `expiresAt=0` 表示）。

#### Scenario: 可试用

- **WHEN** 用户未开通且试用未用
- **THEN** 响应含 `unlocked=false` 且 `trialAvailable=true`（或等价字段）

#### Scenario: 试用中

- **WHEN** 用户处于未过期试用
- **THEN** 响应 `unlocked=true` 且带有未来的 `expiresAt`

---

### Requirement: 小智 MCP SKU 专用表

系统 SHALL 使用独立于 `feature_product` 的数据表存储小智 MCP 永久买断 SKU（至少含商品编码、标题、现价分、原价分、Apple 商品 ID、上下架状态与更新时间）。种子 MUST 提供唯一（或约定的）永久商品编码；`product_code` MUST NOT 经 Admin 改写为任意新码集合（首版以更新种子行为主）。

#### Scenario: 与开通功能管理隔离

- **WHEN** 运维打开开通功能管理的功能 SKU 列表
- **THEN** 列表 MUST NOT 展示小智 MCP 专用 SKU 行

#### Scenario: 配置可被支付读取

- **WHEN** Admin 更新了 MCP SKU 的价格或 `appleProductId` 并保存
- **THEN** 后续 App 建单与 Apple 反查 MUST 使用更新后的值

### Requirement: 独立小 Admin 一页两区

系统 SHALL 在运维 Hub 提供独立于「开通功能管理」的小智 MCP 开通管理页，并在 Hub 导航中可发现。页面 MUST 包含两区：**SKU 配置**与**手工授**。

#### Scenario: 导航可达

- **WHEN** 运维已登录 Hub 并打开模块导航
- **THEN** 可见「小智 MCP 开通」（或同等标题）并进入该独立页

#### Scenario: SKU 区可保存

- **WHEN** 运维在 SKU 区修改价格、原价、Apple 商品 ID 或上下架并提交
- **THEN** 系统持久化专用表并成功返回

#### Scenario: 手工授区授予永久能力

- **WHEN** 运维填写有效 `wxId` 与非空授权理由并提交手工授
- **THEN** 系统为该 wx 授予永久小智 MCP 能力
- **AND** MUST 留下可审查的手工授痕迹（如 admin 订单或等价 channel_ref + grant_reason）

### Requirement: 手工授撤销

系统 SHALL 允许撤销某 wx **最近一笔**且仍有效的小智 MCP **手工授**（`unlock_method=admin`）。撤销后该 wx MUST 视为未开通（除非另有仍有效的支付开通——首版若一人仅一行权益，撤销即未开通）。支付开通 MUST NOT 被本撤销接口当作手工授撤销。

#### Scenario: 撤销最近手工授

- **WHEN** 目标 wx 最近有效开通来自手工授且运维确认撤销
- **THEN** 该 wx 小智 MCP 能力变为未开通

#### Scenario: 无手工授可撤

- **WHEN** 目标 wx 无有效手工授可撤（未开通或仅支付开通）
- **THEN** 撤销 MUST 失败并返回明确业务错误

### Requirement: Admin API 鉴权

小智 MCP 的 Admin 读写接口 MUST 使用与其它 cash Admin 一致的运维鉴权（如 Admin 口令头），MUST NOT 经 App Bearer 匿名可写。

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 Admin 凭据
- **THEN** Admin API MUST 拒绝

---

### Requirement: Admin 已开通人员快照列表

系统 SHALL 提供经 Admin 鉴权的接口，分页返回 `xiaozhi_mcp_entitlement` 当前快照（含有效与已失效行），并为本页 `wxId` 批量附带公开昵称（昵称拉取失败时 MUST 仍返回列表，昵称可为空）。列表 MUST NOT 经 App Bearer 匿名可访问。

#### Scenario: 分页列出含昵称

- **WHEN** 运维携带合法 Admin 凭据请求开通人员列表（含 limit/offset）
- **THEN** 系统返回 `total` 与当前页 `list`
- **AND** 每行至少包含 `wxId`、`unlockMethod`、`active`、`expiresAt`、`unlockedAt`、`updatedAt`、`channelRef`
- **AND** 在能拉到公开资料时填充 `nickname`

#### Scenario: active 语义

- **WHEN** 某行 `status` 为有效且（`expiresAt=0` 永久或 `expiresAt` 未过期）
- **THEN** 该行 `active` MUST 为 true
- **WHEN** 某行已撤销或试用已过期
- **THEN** 该行 `active` MUST 为 false 且仍可出现在列表中（非默认过滤掉）

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 Admin 凭据
- **THEN** 列表接口 MUST 拒绝

### Requirement: Hub 区 C 展示与行内撤销

系统 SHALL 在「小智 MCP 开通」静态页提供 **区 C · 已开通**，调用上述列表接口渲染表格（含昵称、方式、状态、到期、时间、痕迹），并支持刷新与分页。对仍有效且 `unlockMethod=admin` 的行 MUST 提供行内撤销入口，调用既有手工授撤销接口；支付或试用开通 MUST NOT 显示可成功撤销的行内操作（或不提供按钮）。区 B 手工授或撤销成功后，页面 SHOULD 刷新区 C。

#### Scenario: 进页可见已开通表

- **WHEN** 运维打开小智 MCP 开通页并通过 Admin 鉴权
- **THEN** 可见区 C 表格并由列表接口填充当前页数据

#### Scenario: 行内撤销手工授

- **WHEN** 某行 `active=true` 且 `unlockMethod=admin`，运维确认行内撤销
- **THEN** 系统按既有手工授撤销语义使该账号能力变为未开通（除非另有仍有效支付开通——首版一人一行权益则即未开通）
- **AND** 区 C 刷新后该行不再显示为有效 admin 开通

#### Scenario: 支付开通无行内撤销

- **WHEN** 某行仍有效且 `unlockMethod=payment`（或 `trial`）
- **THEN** 区 C MUST NOT 提供可对该行成功撤销的行内按钮（或按钮不可用）

---

### Requirement: tools/call 校验 wx 有效权益

xiaozhi-mcp-service 在执行喂养类 `tools/call` 前 MUST 校验该 Bridge 绑定所属 `wx_id` 仍具备有效小智 MCP 权益（永久或未过期试用）。校验 MUST 经 cash 服务接口契约完成，MUST NOT 在 mcp 进程直查 cash 库表。未开通或校验系统失败时 MUST fail-closed：返回工具层错误，MUST NOT 调用 voice 喂养对话。

#### Scenario: 有效权益允许调用

- **WHEN** 绑定 wx 具有有效权益且工具入参合法
- **THEN** 工具按既有逻辑执行喂养对话并返回结果

#### Scenario: 过期或未开通拒绝调用

- **WHEN** 绑定 wx 无有效权益（含试用已过期、退款撤销后）
- **THEN** `tools/call` MUST 返回错误结果
- **AND** MUST NOT 建立或使用 voice chat 完成喂养

#### Scenario: cash 不可达拒绝调用

- **WHEN** 开通态查询失败（超时/5xx/未授权等）
- **THEN** `tools/call` MUST 失败并提示暂时无法校验
- **AND** MUST NOT 因瞬时错误必然拆桥（与明确未开通区分）

### Requirement: 未开通时懒停桥

当 `tools/call` 因账号**明确未开通**而拒绝时，系统 SHOULD 尽力停止该 token 对应 Bridge（懒停）。绑定行 MUST 保留在 device 库中（不因懒停删除 token/MAC）。

#### Scenario: 拒答后拆桥

- **WHEN** tools 因未开通拒答
- **THEN** 系统尽力 Remove 该 Bridge，使出站连接停止
- **AND** `xiaozhi_mcp_binding` 对应行仍存在

#### Scenario: 再开通后可复用绑行

- **WHEN** 用户随后获得有效权益且绑行仍在
- **THEN** 经写路径 Upsert（或等价同步）后 MUST 能重新建立 Bridge
- **AND** MUST NOT 强制用户因过期而重新录入 token（除非产品另有操作）

### Requirement: Bridge 与同步载荷携带 wxId

device 通知 xiaozhi-mcp 的 Upsert 以及内部全量绑定列表 MUST 包含每条绑定的 `wxId`。mcp Manager/Bridge MUST 在会话中保存该 `wxId` 供 tools 校验。缺少有效 `wxId` 时 tools/call MUST fail-closed。

#### Scenario: Upsert 带 wxId

- **WHEN** device 添加或更新绑定并通知 mcp Upsert
- **THEN** 请求体含该绑定归属 `wxId`
- **AND** 后续该桥上的 tools/call 使用该 `wxId` 查权益

#### Scenario: 内部 list 带 wxId

- **WHEN** xiaozhi-mcp 拉取内部全量绑定做启动或 reconcile
- **THEN** 每条记录含 `wxId`
- **AND** 新建 Bridge 会话保存该值

---
