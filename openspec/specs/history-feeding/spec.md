# history-feeding

## Purpose

喂养历史、事件层级、effective-day、DeepSeek 缓存偏好与 history 服务委托的行为规格。 History write/read models and feeding-day core without clinic product lines.

<!-- migrated-from: appointment-next-at, deepseek-history-redis-prefer, device-event-cache-rebuild-on-mutate, device-event-hierarchy, device-event-logo-color, device-event-type-field, device-event-update-parent-id, end-open-history-by-event, event-is-appointment, feeding-effective-day-core, feeding-shell-b-removal, history-delegate-downstream-urls, history-device-sync-cache-projection, history-event-unit-denorm, history-expect-fields, history-filter-api, history-list-v2-api, history-piece-and-realtime-notify, history-profile-nickname, history-service-db-ownership, history-voice-delegation, voice-history-http-contract -->

## Requirements

### Requirement: 系统 MUST 按宝宝与事件持久化下次约定时间

`device-service` MUST 在 device 域库提供存储，键为 `(deviceNo, eventId)`，值为下次约定 unix 秒 `nextAt`。同一宝宝同一事件 MUST 仅有一条有效约定（唯一约束）。`nextAt=0` MUST 表示「无下次约定」。该存储 MUST NOT 作为预测临近推送的权威，MUST NOT 由服务端据此写入 predict Redis 或发布叫醒 MQ。

#### Scenario: 写入非零下次约定

- **WHEN** 已登录且绑机一致的客户端对某 `(deviceNo, eventId)` 写入 `nextAt=T`（T>0）
- **THEN** 随后读取同一键 MUST 得到 `nextAt=T`

#### Scenario: 清空下次约定

- **WHEN** 客户端对已有约定的键写入 `nextAt=0`
- **THEN** 随后读取同一键 MUST 得到 `nextAt=0`（无下次约定）

#### Scenario: 与 history 行无关

- **WHEN** 同一宝宝同一 `eventId` 存在多条 history 记录
- **THEN** 读取该事件下次约定 MUST 返回同一 `nextAt`，MUST NOT 因所选 history 行不同而变化

### Requirement: 系统 MUST 提供与 history 解耦的预约下次读写 API

系统 MUST 提供 App 路径读写宝宝某事件的下次约定时间 **与备注**（与 history CRUD 解耦）。请求 MUST 含 `deviceNo` 与根 `eventId`；写路径 MUST 接受 `nextAt`（unix 秒，`0`=无约定）与 `remark`（字符串，可空）。读路径无行时 MUST 返回 `nextAt=0` 且 `remark` 为空串。`nextAt=0` 的写入 MUST 将约定视为清空，且 MUST 将 `remark` 置为空串。正 `nextAt` 写入 MUST 持久化所传 `remark`（含空串清备注但保留约定）。该 API MUST NOT 要求修改 history 行。

#### Scenario: GET 返回备注

- **WHEN** 某 `(deviceNo, eventId)` 已存 nextAt=T 与 remark=R
- **THEN** GET MUST 返回 `nextAt=T` 与 `remark=R`

#### Scenario: PUT 正 nextAt 写备注

- **WHEN** 客户端 PUT `nextAt=T>0` 与 `remark=R`
- **THEN** 系统 MUST 持久化 T 与 R，后续 GET MUST 读到相同值

#### Scenario: PUT nextAt=0 清空备注

- **WHEN** 客户端 PUT `nextAt=0`
- **THEN** 系统 MUST 将 nextAt 视为无约定，且 MUST 将 remark 置为空串

#### Scenario: Internal GET 返回备注

- **WHEN** 内部调用方 GET `/device/internal/api/appointment/next` 且该键已存 nextAt=T 与 remark=R
- **THEN** 响应 MUST 含 `nextAt=T` 与 `remark=R`

#### Scenario: Internal GET 无行

- **WHEN** 内部调用方 GET 某键且存储无行
- **THEN** 响应 MUST 为 `nextAt=0` 且 `remark` 为空串

### Requirement: 网关 MUST 放行预约 App 路径到 device-service

`gateway-app-server` MUST 将预约 App 路径反代至 device-service（在 `DEVICE_API_ROUTE_MODE` 启用代理时），MUST NOT 误送到 history/voice。该路径 MUST NOT 加入 Bearer 匿名白名单。是否计入 usage 统计 MUST 以负责人确认为准；未确认前 MUST NOT 擅自加入或移出 `maintenance_skip`。

#### Scenario: 反代可达

- **WHEN** 边缘已配置 device 代理且客户端请求预约读写路径
- **THEN** 请求 MUST 到达 device-service 对应 handler，而非本机 404 或其它域服务

### Requirement: 预约下次约定 MUST NOT 改变预测临近服务端行为

本能力 MUST NOT 要求修改 `PUT /device/api/predict/imminent/pending` 的校验、整表替换、延时投递或消费对账语义。客户端清空可推送约定时，产品约定为从 pending 列表移除该事件；服务端继续依赖既有 Redis 对账使旧 `nextAt` 延时消息作废。

#### Scenario: 预约写零不自动清 Redis

- **WHEN** 客户端仅调用预约写接口将 `nextAt` 置 0，未调用 pending 同步
- **THEN** 系统 MUST NOT 因此自动删除或改写该宝宝的 predict Redis 待办（推送侧行为保持既有客户端上报权威）

### Requirement: appointment_next 表 MUST 持久化 remark 列

`appointment_next`（或等价表）MUST 包含 `remark` 列（字符串，默认空）。启动 ensure MUST 幂等补齐该列。新安装建表 MUST 含该列。

#### Scenario: ensure 补列

- **WHEN** 已有部署缺少 `remark` 列且服务启动 ensure
- **THEN** 系统 MUST 添加该列且 MUST NOT 破坏既有 next_at 数据

---

### Requirement: DeepSeek 历史读取 Redis 优先

系统在执行 **voiceUnderstanding lane** 所触发的 LLM 意图分析与对话补全前，历史上下文读取 SHALL 优先命中 Redis 历史读模型，并在不可用时回源到历史服务或数据库。

#### Scenario: LLM 请求前 Redis 命中

- **WHEN** voiceUnderstanding 路径请求前需要最近历史上下文且 Redis 命中
- **THEN** 系统 MUST 使用缓存历史组装 prompt

#### Scenario: 缓存未命中回源

- **WHEN** voiceUnderstanding 路径请求前历史缓存未命中
- **THEN** 系统 MUST 回源获取历史并回填缓存后继续调用 LLM

### Requirement: 上下文读取一致性与降级可观测

系统 MUST 对 **voiceUnderstanding** LLM 上下文读取提供命中率、回源率与降级原因可观测性，并在异常时保持功能可用。

#### Scenario: 请求完成可观测

- **WHEN** 任一 voiceUnderstanding LLM 请求完成上下文装配
- **THEN** 系统 SHOULD 输出结构化日志含缓存命中/回源标记

#### Scenario: Redis 不可用降级

- **WHEN** Redis 读模型暂时不可用
- **THEN** 系统 MUST 降级回源并继续完成 LLM 调用，同时输出结构化告警日志

### Requirement: 历史窗口语义一致

系统读取最近 N 小时历史用于 **voiceUnderstanding** LLM 时 MUST 遵守既有 history 读模型契约，MUST NOT 新增跨库直查他域表。

#### Scenario: 历史问答使用读模型

- **WHEN** 历史问答等价路径加载 12 小时历史
- **THEN** MUST 经 Redis 优先读模型路径

---

### Requirement: 事件表变更后 Redis 缓存必须从数据库重建

当 `device-service` 成功执行对 `event` 表的插入、更新或删除后，系统 SHALL 使用数据库当前全量事件行（含 `logo`、`color` 与 **`parent_id`**）重建 Redis 中的事件选项缓存，且 SHALL NOT 通过先调用 `ListEvents`（可能仅返回变更前缓存）再写回的方式刷新缓存。

#### Scenario: 更新事件 color 后缓存含新色值

- **WHEN** 管理员通过 API 成功更新某事件的 `color`
- **THEN** 随后对 `ListEvents` 或等价读路径的调用在缓存命中时 SHALL 返回包含新 `color` 的该事件行

#### Scenario: 更新事件 logo 后缓存含新 path

- **WHEN** 管理员成功上传并更新某事件的 `logo` 路径
- **THEN** Redis 事件选项快照中该事件的 `logo` 字段 SHALL 与数据库一致

#### Scenario: 新增事件后缓存含新行

- **WHEN** 管理员成功新增一条事件记录
- **THEN** 随后缓存命中时 SHALL 包含该新事件

#### Scenario: 删除事件后缓存不含已删行

- **WHEN** 管理员成功删除一条事件记录
- **THEN** 随后缓存命中时 SHALL NOT 包含已删除的事件 id

#### Scenario: 新增子事件后缓存含 parentId

- **WHEN** 管理员成功新增 `parent_id=5` 的子事件
- **THEN** 随后缓存命中时该事件行 SHALL 包含 `parentId=5`

### Requirement: 写后刷新失败可观测

若重建 Redis 缓存失败，系统 SHALL 记录警告级别日志且 SHALL NOT 将写库事务回滚（写库已成功）。

#### Scenario: Redis 不可用时的行为

- **WHEN** 数据库写入成功但 `RebuildEventCache` 因 Redis 错误失败
- **THEN** 系统 SHALL 记录可观测警告日志
- **AND** API 仍可对客户端返回写库成功（与现网语义一致）

---

### Requirement: 事件表以 parent_id 表达通用树

`device-service` 持久化的 `event` 行 SHALL 使用 `parent_id` 表示层级：`0`（或等价空值约定）为根节点；非零值 MUST 指向已存在的父事件 id。系统 SHALL NOT 在业务逻辑中读写 `child_ids` 列。

#### Scenario: 新增子事件写入 parent_id

- **WHEN** 管理员提交 `POST /device/admin/api/event/add` 且表单 `parentId=5`
- **THEN** 新行 `parent_id` SHALL 为 `5`
- **AND** 父行 SHALL NOT 依赖 `child_ids` 维护子列表

#### Scenario: 根事件 parent_id 为零

- **WHEN** 管理员新增根事件且未提交 `parentId` 或 `parentId=0`
- **THEN** 新行 `parent_id` SHALL 为 `0`

### Requirement: 同父下事件名唯一

创建或更新事件时，系统 SHALL 在相同 `parent_id` 下保证 `name` 唯一；不同 `parent_id` 下 MAY 存在相同 `name`。

#### Scenario: 同父重复名称被拒绝

- **WHEN** 父 id=5 下已存在名为「大便」的事件
- **AND** 客户端在同一 `parentId=5` 下再次提交 `name=大便`
- **THEN** API SHALL 返回业务错误且 SHALL NOT 插入

#### Scenario: 不同父允许同名

- **WHEN** 父 id=5 下已存在「其他」
- **AND** 客户端在 `parentId=10` 下提交 `name=其他`
- **THEN** API SHALL 允许创建

### Requirement: 有子节点的事件不可删除

`DeleteEvent` SHALL 在存在任意 `parent_id` 等于待删 id 的行时拒绝删除。

#### Scenario: 删除有子的父事件失败

- **WHEN** 事件 id=5 存在 `parent_id=5` 的子行
- **AND** 客户端请求删除 id=5
- **THEN** API SHALL 返回可识别业务错误
- **AND** 数据库 SHALL 保留 id=5 行

#### Scenario: 删除叶子事件成功

- **WHEN** 事件 id=12 无子行
- **THEN** 删除 SHALL 成功且 SHALL 触发事件缓存重建

### Requirement: ListEvents 返回 parentId

`GET /device/admin/api/event/list` 及内部 `ListEvents` 契约 SHALL 在每条事件记录中包含 `parentId` 字段。

#### Scenario: 列表含 parentId

- **WHEN** 客户端请求事件列表
- **THEN** 每项 SHALL 包含与数据库 `parent_id` 一致的 `parentId`

### Requirement: 新增子事件不继承父 logo 与 color

带非零 `parentId` 创建事件时，系统 SHALL NOT 从父行复制 `logo` 或 `color`；新行视觉字段 SHALL 仅来自本次提交或系统默认值。

#### Scenario: 子事件使用表单色值而非父色

- **WHEN** 父事件 `color=#FF0000`
- **AND** 子事件创建表单提交 `color=#4A90D9` 与 `parentId=5`
- **THEN** 新行 `color` SHALL 为 `#4A90D9`
- **AND** SHALL NOT 自动设为 `#FF0000`

---

### Requirement: 事件 logo 与 color SHALL 可配置且列表可见

device-service 事件字典 MUST 支持 `logo`（OSS objectKey，前缀 `event/`）与 `color` 的持久化；所有返回事件字典列表的 HTTP 接口 MUST 在 JSON 中将 `logo` 序列化为 CDN 绝对 URL（无 logo 时为空串），并 MUST 包含 `color` 字段。Redis 事件选项缓存（`device:event:options:all`）MAY 仅持久化 objectKey；**任何** HTTP 响应边界 MUST 在返回前完成 CDN 映射，不得因缓存命中而跳过。

#### Scenario: 管理端事件列表返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/admin/api/event/list`
- **THEN** 响应 `list[]` 中每项 MUST 包含 `logo` 与 `color` 字段
- **AND** `logo` 若有值 MUST 为 CDN 绝对 URL（`https://` 开头）

#### Scenario: 历史与内部事件选项返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/history/api/event/options` 或 `GET /device/internal/api/event/options`
- **THEN** 响应 `list[]` MUST 同样包含 CDN 形式的 `logo` 与 `color`

#### Scenario: 管理端更新 logo 后 history 缓存命中仍返回 CDN

- **WHEN** 管理员成功更新某事件 logo 且 device-service 已重建 Redis 事件选项缓存（缓存内 logo 为 objectKey）
- **AND** App 或客户端请求 `GET /device/history/api/event/options` 且 history-service 命中该 Redis 键
- **THEN** 响应 `list[]` 中每项 `logo` 若有值 MUST 仍为 CDN 绝对 URL（`https://` 开头）
- **AND** MUST NOT 返回裸 objectKey（如 `event/2026/06/xxx.png`）

### Requirement: 事件新增与更新 SHALL 支持 multipart 上传 logo

`POST /device/admin/api/event/add` 与 `POST /device/admin/api/event/update` MUST accept `multipart/form-data`，至少包含表单字段 `name`、`eventType`、`extraNames`、`color`；`update` MUST 包含 `id`。可选文件字段名 MUST 为 `logo`。

#### Scenario: 新增事件并上传 logo

- **WHEN** 客户端 multipart 提交有效 `name` 与合法图片 `logo`
- **THEN** 服务端 MUST 经 ucg internal 上传至 OSS
- **AND** MUST 将 `event.logo` 设为 `event/...` objectKey
- **AND** MUST 将 `color` 写入 `event.color`（若提供）

#### Scenario: 更新事件未传 logo 保留原值

- **WHEN** 客户端对已有事件 multipart 更新且未包含 `logo` 文件
- **THEN** 服务端 MUST 保留原 `event.logo` objectKey
- **AND** MAY 更新 `color` 与其它文本字段

### Requirement: 事件 Redis 缓存投影 SHALL 含 logo 与 color

`ListEvents` 使用的 Redis 事件选项缓存 MUST 与数据库查询一致，包含 `logo` 与 `color`，以便缓存命中时列表仍返回完整字段。

#### Scenario: 缓存命中仍含 logo

- **WHEN** `ListEvents` 从 Redis 命中事件选项
- **THEN** 返回的 `[]Event` MUST 含非省略的 `logo`、`color` 字段

### Requirement: history 写 Redis 事件选项缓存 SHALL 仅持久化 objectKey

history-service 向 `device:event:options:all` 写入事件选项快照时，SHALL 将每项 `logo` 规范为 OSS objectKey（无前缀域名）；SHALL NOT 将 CDN 绝对 URL 写入共享 Redis 键，以避免与 device-service 重建缓存的 objectKey 格式混写。

#### Scenario: history 回源 internal 后写缓存

- **WHEN** history-service cache miss 后从 device internal 拉取事件列表（响应 logo 已为 CDN URL）
- **AND** history-service 将结果写入 Redis 事件选项键
- **THEN** 写入 Redis 的 JSON 中 `logo` MUST 为 objectKey 形式
- **AND** 后续 history HTTP 响应 MUST 仍通过 CDN 映射返回绝对 URL

---

### Requirement: 事件主档必须持久化有效的 event_type

`device-service` 在创建或更新 `event` 表记录时，SHALL 接受并持久化 `event_type`，其值 MUST 为 `number`、`time` 或 `one` 之一。系统 SHALL NOT 再读写 `need_quantity` 列或 API 字段 `needQuantity`。

#### Scenario: 管理端新增事件带 eventType

- **WHEN** 客户端 `POST /device/admin/api/event/add` 提交合法 `eventType=number`
- **THEN** 数据库新行 `event_type` SHALL 为 `number`
- **AND** 随后 `ListEvents` 或缓存命中 SHALL 返回该 `eventType`

#### Scenario: 非法 eventType 被拒绝

- **WHEN** 客户端提交 `eventType` 为空或不在 `number|time|one`
- **THEN** API SHALL 返回参数错误且 SHALL NOT 插入或更新行

### Requirement: 事件选项 Redis 快照含 event_type

写库成功后，系统 SHALL 通过从数据库全量扫描（含 `event_type` 列）重建 Redis 事件 options，且 SHALL NOT 依赖可能过期的 `ListEvents` 缓存读回后写回。

#### Scenario: 更新事件后缓存含新类型

- **WHEN** 管理员成功更新某事件的 `eventType` 为 `one`
- **THEN** 重建后的 Redis 快照中该事件 SHALL 带有 `eventType` 为 `one`

### Requirement: 匹配已有事件时不改 event_type

当仅合并别名（`extra_names`）或命中已有事件名时，系统 SHALL NOT 更新该事件行的 `event_type`。

#### Scenario: DeepSeek 仅追加别名

- **WHEN** 抽取结果匹配已存在事件名且仅合并 `extraNames`
- **THEN** 该事件 `event_type` 列 SHALL 保持不变

### Requirement: 对话新建事件时由模型提供 event_type

经 voice 调用的 `InsertOrGetEventByNeedle` 或 DeepSeek 落库插入新事件时，系统 SHALL 将模型给出的 `event_type` 写入新行；若模型未给出合法值，SHALL 使用规范化默认值（`time`）并仍可观测。

#### Scenario: 语音路径插入新事件

- **WHEN** 用户话术导致新建事件且 DeepSeek 返回 `event_type` 为 `number`
- **THEN** 新插入的 `event` 行 `event_type` SHALL 为 `number`

---

### Requirement: UpdateEvent SHALL 支持修改 parent_id

`device-service` 的 `UpdateEvent`（及 `POST /device/admin/api/event/update`）SHALL 接受 **`parentId`**（非负整数，`0` 表示根）。当请求携带有效 `parentId` 且与库内当前值不同时，系统 SHALL 更新该行的 `parent_id` 字段，并 SHALL 在成功写库后触发与现有事件变更一致的 Redis 事件选项缓存重建。

#### Scenario: 将事件挂到新的父节点下

- **WHEN** 管理员提交 `id=10`、`parentId=5`，且 id=5 存在、不构成环
- **THEN** id=10 行的 `parent_id` SHALL 变为 `5`
- **AND** 随后 `ListEvents` / 缓存中该项的 `parentId` SHALL 为 `5`

#### Scenario: 将事件提升为根节点

- **WHEN** 管理员提交 `id=10`、`parentId=0`
- **THEN** id=10 行的 `parent_id` SHALL 为 `0`

#### Scenario: 未变更父节点时仅更新其他字段

- **WHEN** 管理员仅修改 `name` 且提交的 `parentId` 与库内一致
- **THEN** 系统 SHALL 仅更新非层级字段，且 SHALL NOT 产生无效的父节点写操作错误

### Requirement: 修改 parent_id 须校验父存在且无环

当 `parentId > 0` 时，系统 SHALL 校验对应父事件行存在。系统 SHALL 拒绝 `parentId` 等于待更新事件自身 id，SHALL 拒绝将父设为其**任意后代**（防止环）。违反时 SHALL 返回业务错误且**不得**部分更新。

#### Scenario: 父节点不存在

- **WHEN** 提交 `parentId=99999` 且库中无 id=99999
- **THEN** 请求 SHALL 失败并返回可识别的错误信息

#### Scenario: 父节点为自身

- **WHEN** 提交 `id=10`、`parentId=10`
- **THEN** 请求 SHALL 失败

#### Scenario: 父节点为子孙造成环

- **WHEN** 事件 10 拥有后代 20（`parent_id=10`）
- **AND** 提交 `id=10`、`parentId=20`
- **THEN** 请求 SHALL 失败

### Requirement: 修改 parent_id 后同父 name 唯一

更新名称或父节点后，系统 SHALL 在**目标父** `parent_id` 下保证 `name` 唯一（排除自身 id），规则与 `device-event-hierarchy` 中 AddEvent 一致。

#### Scenario: 移动后与兄弟同名冲突

- **WHEN** 父 5 下已存在 `name=大便` 的事件
- **AND** 将另一事件移至 `parentId=5` 且 `name=大便`
- **THEN** 请求 SHALL 失败并返回与「事件已存在」一致的业务错误

### Requirement: 有子节点的事件 MAY 修改 parent_id

存在 `parent_id = 待更新 id` 的子行时，系统 **MAY** 允许修改该事件的 `parent_id`；子行的 `parent_id` SHALL 仍指向原 id，除非另有删除/移动子树需求。

#### Scenario: 中间节点更换父级而子节点仍挂在其下

- **WHEN** 事件 10 有子事件 20（`parent_id=10`）
- **AND** 成功将事件 10 的 `parent_id` 改为 5
- **THEN** 事件 20 的 `parent_id` SHALL 仍为 `10`

---

### Requirement: EndLatest 按 eventId 闭合最近未结束记录

`history-service` 的 `EndLatestHistoryIfMatch`（及内部 HTTP `POST /device/history/api/event/end-latest`）MUST 在指定 `deviceNo` 与 `eventId` 下，查找 `end_time=0`（未闭合）的历史行，按 `id` 降序取最近一条并更新其结束时间；MUST NOT 仅因「全局最新一条 history 的 eventId 不等于目标」而返回未匹配。当 remark 非空时 MUST 同时覆盖该行备注；remark 为空串时 MUST NOT 修改原备注。

#### Scenario: 中间夹其它事件仍能结束睡眠

- **WHEN** 设备已有未闭合睡眠记录（`eventId=E` 且 `end_time=0`），其后又写入其它事件记录，使全局最新一条不再是睡眠
- **AND** 调用方对同一 `deviceNo` 请求结束 `eventId=E`（语音 `feeding+end` 或 App `end-latest`）
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=true`
- **AND** SHALL 更新该未闭合睡眠行的 `end_time`（而非新建一条睡眠历史）

#### Scenario: 无未闭合同 event 时返回未匹配

- **WHEN** 该 `deviceNo` 下不存在 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方请求结束 `eventId=E`
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=false`
- **AND** SHALL NOT 修改任何已结束（`end_time!=0`）的同 `eventId` 行

#### Scenario: 多条未闭合同 event 只闭合最近一条

- **WHEN** 同一 `deviceNo` 下存在多条 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方请求结束 `eventId=E`
- **THEN** 系统 SHALL 仅更新其中 `id` 最大的一条的 `end_time`
- **AND** 其余未闭合行保持 `end_time=0`

### Requirement: 语音 end 优先闭合未结束记录再降级新建

喂养事件 CRUD（含 start/end/one/multi）MUST 由 `python-ai-talk` 在意图落地阶段经 `POST /device/history/api/event/batch` 写入 history；Go `voice-service` MUST NOT 在 `applyUnifiedIntentResult` 或等价主路径内调用 `DeviceHistory().AddHistory` / `UpdateHistory` 执行喂养写库。当 Python 处理 `action=end` 且目标 event 为 E 时，batch MUST 优先使用 `op=end` 或等价语义经 `EndLatestHistoryIfMatch` 闭合；仅当无未闭合同 E 记录时 MAY 降级 `create` 瞬时结束行（`start_time=end_time`）。MUST NOT 在仍存在未闭合 E 记录时新建一条 E 的历史行作为「结束」手段。

#### Scenario: 孩子醒了结束已开始的睡眠

- **WHEN** Python 返回结束睡眠意图且该设备存在未闭合睡眠历史（同 eventId）
- **THEN** Python batch MUST 闭合该未闭合睡眠行（经 end-latest 或 batch op=end）
- **AND** MUST NOT 再新增一条睡眠历史行作为结束

#### Scenario: 从未开始睡眠时 end 可降级新建

- **WHEN** Python 返回结束睡眠意图，且该设备不存在未闭合的同 eventId 历史
- **THEN** batch MAY 写入瞬时结束记录
- **AND** Go voice MAY 透传 Python 成功话术

#### Scenario: 重复 start 被 history 拒绝

- **WHEN** Python batch 提交同 eventId 的 start（`endTime=0`）且该 event 已在进行中
- **THEN** history-service MUST 拒绝该 create
- **AND** batch 对应 item MUST 返回失败 reason

### Requirement: 禁止同 eventId 重复 start 进行中记录

`history-service` 在 `AddDeviceHistory`（及经其的 `POST /device/history/api/event/add`、`POST /device/history/api/event/batch` 之 `op=create`）插入新行时，若请求满足 `eventId>0` 且 `endTime=0`（进行中），MUST 先查询该 `deviceNo` 下是否已存在同 `eventId` 且 `end_time=0` 的历史行； 若存在 MUST 拒绝插入并返回可读错误（语义等价：`该事件已在进行中，请先结束后再开始`）。 `UpdateDeviceHistory`（及 `event/update`、batch `op=update`）将某行 `endTime` 改为 `0` 时 MUST 应用相同规则（排除当前行 `id`）。 本守卫 MUST NOT 限制不同 `eventId` 的并行进行中； MUST NOT 拦截 `endTime!=0` 的瞬时记录（含 `startTime=endTime`.

#### Scenario: 同 event 已有进行中时拒绝再次 start

- **WHEN** 设备 `deviceNo=D` 已存在 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方经 add 或 batch create 请求插入 `eventId=E`、`endTime=0` 的新行
- **THEN** history-service MUST 拒绝该插入
- **AND** MUST NOT 新增第二条同 E 的未闭合行

#### Scenario: 不同 event 允许并行进行中

- **WHEN** 设备已有 `eventId=E1` 且 `end_time=0` 的睡眠记录
- **AND** 调用方插入 `eventId=E2`（E2≠E1）且 `endTime=0` 的计时记录
- **THEN** history-service MUST 允许插入

#### Scenario: 瞬时 one 记录不受限

- **WHEN** 设备已有 `eventId=E` 且 `end_time=0` 的进行中行
- **AND** 调用方插入同 E 但 `endTime>0` 且 `endTime>=startTime` 的瞬时记录
- **THEN** history-service MUST 允许插入

#### Scenario: batch 单条失败不整单回滚

- **WHEN** Python 经 `event/batch` 提交多条 items，其中一条因重复 start 失败
- **THEN** 该 item MUST 在 results 中 `ok=false` 且 `reason` 含可读拒绝语义
- **AND** 其它 items MUST 仍按既有 batch 语义独立执行

---

### Requirement: 事件字典 MUST 支持预约标志且默认非预约

系统 MUST 在全局事件字典（`event`）持久化预约标志，对外 JSON 字段名为 `isAppointment`（或与现有 camelCase 约定一致的布尔字段）。新建事件时若未显式指定，该标志 MUST 为 false（非预约）。该标志 MUST NOT 复用或改写既有 `eventType`（`number`|`time`|`one`）语义。

#### Scenario: 新建事件默认非预约

- **WHEN** 管理员新增事件且未勾选预约
- **THEN** 持久化记录的预约标志 MUST 为 false，且后续列表/options 返回中该事件 `isAppointment` 为 false

#### Scenario: 与 eventType 正交

- **WHEN** 某事件 `eventType` 为 `one` 且预约标志为 false
- **THEN** 系统 MUST 仍按非预约事件暴露，MUST NOT 仅因 `one` 而视为预约

### Requirement: 后台 MUST 可编辑预约标志

设备域后台新增与更新事件时，MUST 允许设置预约标志。更新成功后，系统 MUST 使事件字典读模型与既有「变更后重建 EventOptions 缓存」约定一致，以便后续 options/list 读到新值。

#### Scenario: 后台将事件标为预约

- **WHEN** 管理员将已有事件更新为预约（`isAppointment=true`）且保存成功
- **THEN** 之后经事件列表或 options 读取该事件时 MUST 返回 `isAppointment=true`

### Requirement: 事件 options 与列表 MUST 附带预约标志

凡返回 `entity.Event`（或等价事件字典项）的 App/内部 options 与后台列表路径，MUST 包含预约标志字段，供 Flutter 判断是否按预约逻辑处理。字段缺失 MUST NOT 作为「非预约」的隐式协议以外的行为；实现 MUST 保证新列映射到响应（旧缓存淘汰后）。

#### Scenario: Flutter 可从 options 识别预约事件

- **WHEN** 客户端请求历史事件 options（或等价字典接口）且字典中存在已标记预约的事件
- **THEN** 响应列表中对应项 MUST 包含 `isAppointment=true`

---

### Requirement: 有效喂养日判定 MUST 独立于场景门槛且仅由 cash-service 执行

系统 MUST 在 **cash-service** 提供与场景解耦的有效日判定：使用 `Asia/Shanghai` 日历日；某日有效当且仅当该 `device_no` 在该日 history 记录数 ≥ 该次计算使用的 `minRecordsPerDay`；连续有效日 MUST 以 **上海昨日** 为锚向前连续累计（`days[0]` MUST 为昨天），遇无效日中断；**请求当日（今日）MUST NOT** 计入有效日或连续 streak。资格计算 MUST 经 history-service HTTP 按日计数，MUST NOT 由 cash 直查 history 库。UCG 入场与值得留意喂养资格的天数判断 MUST NOT 实现在 device-service、voice-service、history-service 或客户端权威路径。

#### Scenario: 日条数达门槛计有效

- **WHEN** 某上海已闭合日 history 条数等于该次 `minRecordsPerDay`
- **THEN** 该日 MUST 计为有效喂养日

#### Scenario: 不足门槛中断连续

- **WHEN** 从昨天往回扫描时某日条数小于 `minRecordsPerDay`
- **THEN** 该日 MUST NOT 计入，且连续计数 MUST 在该日中断

#### Scenario: 今日不计入

- **WHEN** 请求时刻为上海某日且当日 history 条数已 ≥ `minRecordsPerDay`
- **THEN** 该「今日」MUST NOT 增加 `effectiveDays`，连续扫描 MUST 自昨天起算

#### Scenario: 跨日立即合格

- **WHEN** 上海日切换后，以昨天为锚的连续有效日 ≥ 该场景 `requiredDays`
- **THEN** 当日资格请求 MUST 可返回 `qualified=true`（按请求日缓存键 miss 后重算即可），MUST NOT 依赖额外后台 ticker

#### Scenario: 资格合成宿主为 cash

- **WHEN** 计算 UCG 或值得留意喂养资格
- **THEN** 连续天数与 `qualified` 合成 MUST 在 cash-service 完成，history MUST 仅返回按日条数

### Requirement: 取数窗口 MUST 等于所需连续天数

调用 history 按日统计时，请求天数 MUST 等于本次合成所需的 `requiredDays`（单场景）；窗口 MUST 覆盖自 **上海昨日** 起往前共 `requiredDays` 个已闭合日历日，MUST NOT 包含今日；MUST NOT 无必要请求大于该阈值的固定窗口（例如仅为缓冲而固定 14）。若一次请求需服务多个场景，窗口 MUST 为各场景 `requiredDays` 的最大值且 MUST NOT 更大。history 实现 MUST 在数据库侧按上海日聚合 `COUNT`（或等价），MUST NOT 为资格统计拉取窗口内全部 history 行的 `start_time` 列表再在应用层逐条累加。

#### Scenario: UCG 仅拉 requiredDays 已闭合日

- **WHEN** 计算 `ucg_entry` 且配置 `requiredDays=7`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `7`，且返回序列 `days[0]` MUST 为上海昨天

#### Scenario: 值得留意仅拉 requiredDays 已闭合日

- **WHEN** 计算 `care_alert_entry` 且配置 `requiredDays=2`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `2`，且窗口实质为上海昨天与前天（不含今日）

#### Scenario: 按日 COUNT 返回

- **WHEN** history 处理 `feeding-day-stats` 且窗口内某日有多条记录
- **THEN** 响应该日 MUST 为聚合后的单条 `{date, count}`，MUST NOT 要求调用方再按原始行重算

### Requirement: 各场景 MUST 使用各自配置的连续天数与日条数门槛

场景 `ucg_entry` 与 `care_alert_entry` MUST 分别持久化 `requiredDays` 与 `minRecordsPerDay`；合成某一场景资格时 MUST 只使用该场景配置，MUST NOT 强制两场景共用同一日条数门槛。VIP 与功能权益 MUST NOT 改变任一场景的 `qualified`。

#### Scenario: 两场景门槛可不同

- **WHEN** `ucg_entry` 的 `minRecordsPerDay=10` 且 `care_alert_entry` 的 `minRecordsPerDay=5`
- **THEN** 同一设备同一日的 history 条数=6 时，该日对值得留意 MUST 计有效，对 UCG MUST NOT 计有效

#### Scenario: UCG 合格判据读配置

- **WHEN** `ucg_entry.requiredDays=7` 且连续有效日 ≥ 7
- **THEN** UCG eligibility MUST `qualified=true` 且 `requiredDays=7`

### Requirement: 各场景门槛语义不变且 care_alert 为促活门槛

场景 `ucg_entry` 与 `care_alert_entry` MUST 分别持久化并使用各自的 `requiredDays` 与 `minRecordsPerDay`；合成某一场景资格时 MUST 只使用该场景配置，MUST NOT 强制两场景共用同一日条数门槛。VIP 与功能权益 MUST NOT 改变任一场景的 `qualified`。`care_alert_entry` 的连续有效日门槛 MUST 视为功能解锁 / 促活门槛，MUST NOT 被定义为「为值得留意数据准确性而要求昨日有发生」的继承语义；排除今日后 `requiredDays=2` 几何上对应「昨天+前天」连续有效，属阈值结果而非旧客户端闸门的等价替换说明义务。

#### Scenario: 两场景门槛可不同

- **WHEN** `ucg_entry` 的 `minRecordsPerDay=10` 且 `care_alert_entry` 的 `minRecordsPerDay=5`
- **THEN** 同一设备同一已闭合日的 history 条数=6 时，该日对值得留意 MUST 计有效，对 UCG MUST NOT 计有效

#### Scenario: UCG 合格判据读配置

- **WHEN** `ucg_entry.requiredDays=7` 且以昨天为锚的连续有效日 ≥ 7
- **THEN** UCG eligibility MUST `qualified=true` 且 `requiredDays=7`

#### Scenario: care_alert 满两闭合日合格

- **WHEN** `care_alert_entry.requiredDays=2` 且上海昨天与前天均为有效喂养日
- **THEN** care-alert eligibility MUST `qualified=true`

---

### Requirement: 删除对外 history 文本喂养 API

系统 MUST NOT 再提供 `POST /device/history/api/chat` 与 `POST /device/history/api/chat/stream` 作为自然语言喂养入口。本变更 MUST 直接删除（或取消注册）上述路由与 `HistoryCtrl` 对应处理方法，MUST NOT 保留过渡期 200 兼容实现。

#### Scenario: chat 路径不可用

- **WHEN** 客户端请求 `POST /device/history/api/chat` 或 `.../chat/stream`
- **THEN** 网关/history MUST 返回未找到或未注册路由的失败语义（非成功 SSE/JSON 对话）

### Requirement: 删除 voice internal 文本喂养 API

voice-service MUST NOT 再提供 `POST /voice/internal/api/text/chat` 与 `POST /voice/internal/api/text/chat/stream`。history-service MUST 删除 `DelegateTextChat` 与 `DelegateTextChatStream` 及其调用点。若 `POST /voice/text/chat` 仅服务于已删除文本喂养调试路径，MUST 一并删除。

#### Scenario: internal text chat 不可用

- **WHEN** 带内部密钥请求 `/voice/internal/api/text/chat` 或 `.../stream`
- **THEN** voice-service MUST 不再成功执行文本喂养对话

### Requirement: 自然语言喂养仅经 chat WS

自然语言喂养（意图理解与业务落地）对外 MUST 仅经 `/voice/chat/ws`。系统 MUST NOT 再维护第二条 HTTP 文本喂养外壳。

#### Scenario: 文档与 Admin 说明一致

- **WHEN** 运维查看 voice-admin 或相关 runbook 中的喂养 AI 入口列表
- **THEN** 列表 MUST 以 `/voice/chat/ws` 为喂养对话入口，MUST NOT 再将已删的 history/internal text chat 路径列为可用入口

---

### Requirement: history-service 在隔离网络运行时 MUST 使用可路由的下游 HTTP 基址

当 `history-service` 进程运行在与其他领域服务**不同的网络命名空间**（例如 Docker 独立容器、Kubernetes Pod）且需要通过 HTTP 委派访问 `voice-service` 或 `device-service` 时，其 MUST 通过环境变量 `VOICE_SERVICE_URL` 与 `DEVICE_SERVICE_URL` 配置为可在该命名空间内解析并路由到目标服务的基址（例如同一编排系统中的服务 DNS 名），MUST NOT 依赖指向本容器 loopback 的默认基址（如 `http://127.0.0.1:9802`）作为跨容器访问手段。

#### Scenario: Docker Compose 微服务栈中 history 访问 device

- **WHEN** `history-service` 与 `device-service` 作为不同容器加入同一用户定义 bridge/overlay 网络，且 history 需调用 device 的 HTTP API
- **THEN** 部署配置 SHALL 为 history 设置 `DEVICE_SERVICE_URL`（例如 `http://device-service:9803`），使得 TCP 连接目标为 device 容器而非 history 容器自身

#### Scenario: Docker Compose 微服务栈中 history 访问 voice

- **WHEN** `history-service` 与 `voice-service` 作为不同容器在同一编排网络内，且 history 需调用 voice 的 HTTP API
- **THEN** 部署配置 SHALL 为 history 设置 `VOICE_SERVICE_URL`（例如 `http://voice-service:9802`），使得请求到达 voice 服务监听端口

### Requirement: 仓库参考 Compose 中 history 段落 SHALL 与 voice 下游配置语义一致

`manifest/docker/docker-compose.microservices.yml`（或其后继等价的官方微服务 Compose 参考）中，若同时定义 `history-service` 与 `voice-service`、`device-service`，则 `history-service` 的环境变量段落 SHALL 包含与同文件内其他服务一致的、基于服务名的 `VOICE_SERVICE_URL` 与 `DEVICE_SERVICE_URL`，以避免开箱即用栈中出现 history 误连 `127.0.0.1` 的失败。

#### Scenario: 审查者对比 voice 与 history 环境块

- **WHEN** 审查者检查微服务 Compose 文件中 voice 已配置 `DEVICE_SERVICE_URL` 指向 `device-service`
- **THEN** 其 SHALL 能在 history 段落找到对 voice、device 的显式 URL 配置，且主机部分为 compose 服务名而非 `127.0.0.1`

---

### Requirement: history 写操作 MUST 同步更新 Redis 读模型

`history-service`（及 voice adapter 本地 patch 路径）在 `history` 表 insert/update/delete **成功提交后**，MUST 在同请求内调用读模型 patch（`patchHistoryOnAdd` / `patchHistoryOnUpdate` / `patchHistoryOnDelete`）或等价逻辑更新 `history:record:list:*` 与 `history:record:latest:*`；MUST NOT 依赖 worker-service 或 domain_outbox 异步投影作为唯一更新路径。

#### Scenario: AddHistory 成功后列表与 latest 可读

- **WHEN** `AddDeviceHistory` 事务提交成功
- **THEN** 系统 MUST 尝试同步 patch Redis；后续 `GetLatestHistory` 在 cache miss 或 patch 成功时 MUST 返回含新记录的正确数据（以 MySQL 为准）

#### Scenario: 冷缓存新增记录仍更新 latest

- **WHEN** 设备 history 列表 Redis key 不存在且新增一条记录
- **THEN** 系统 MUST 至少写入 `history:record:latest:{deviceNo}`；列表 MAY 在首次 `ListHistory` miss 时从 MySQL 全量回填

### Requirement: list patch 失败 MUST 避免长期 stale hit

当 `setHistoryList` 失败时，系统 SHOULD best-effort 删除对应 list key，使下次读取走 MySQL 回源；MUST NOT 假设 worker 会异步修复。

#### Scenario: Redis 写入失败后读路径自愈

- **WHEN** prepend 列表缓存时 Redis 返回错误
- **THEN** 系统 MUST 记录 warning；SHOULD 删除 list key；读路径在 miss 时 MUST 从 MySQL 重建列表

### Requirement: device 域字典与画像 MUST 在写路径同步重建缓存

device admin 变更 `event`/`action`/`user` 后，MUST 在写库成功后调用 `refreshEventOptionsCacheAfterMutate`、`RebuildActionCache` 或 `setUserProfile` 等同步重建；MUST NOT 依赖 worker HTTP 回调作为唯一刷新路径。

#### Scenario: 后台新增事件后主数据可读

- **WHEN** 管理端新增事件并提交成功
- **THEN** 下一次 `ListEvents` MUST 能读到新事件（Redis hit 或 miss 回源 MySQL 后回填）

### Requirement: 读路径 MUST 在 cache miss 时回源 MySQL 并回填

`ListHistory`、`GetLatestHistory`、device `ListEvents` 等在 Redis miss 或不可用时，MUST 查询权威 MySQL 并回填 Redis；数据 correctness MUST 以 MySQL 为准。

#### Scenario: 列表 cache miss

- **WHEN** `getHistoryList` 未命中
- **THEN** adapter MUST 查库并 `setHistoryList` 回填

---

### Requirement: History 写入 MUST 反规范化事件单位

当新增或更新 `history` 记录且关联的 `eventId` 在 device 事件主档中存在非空 `unit` 时，history-service MUST 在持久化前将单位写入 `history.event_unit`（写入时刻快照）。若请求体已携带非空 `eventUnit`，MUST 优先使用该值；否则 MUST 经 device-service HTTP 契约解析单位。history-service MUST NOT 通过本进程 `default` 数据库连接直查 `event` 表。

#### Scenario: 微服务分库下新增历史带单位

- **WHEN** device 库中 `event.id=10` 的 `unit` 为 `ml`，客户端经 history HTTP 新增一条 `eventId=10` 的历史且请求体未传 `eventUnit`
- **THEN** history-service MUST 经 device 契约解析到 `ml` 并成功 INSERT，`history.event_unit` SHALL 为 `ml`

#### Scenario: 请求体显式携带单位

- **WHEN** 客户端 POST 新增历史且 `eventUnit` 为 `次`
- **THEN** history-service MUST 持久化 `history.event_unit=次`，且 MUST NOT 被主档单位覆盖

#### Scenario: 事件主档无单位

- **WHEN** 关联 `event.unit` 为空且请求体未传 `eventUnit`
- **THEN** history-service MUST 持久化 `history.event_unit` 为 NULL 或等效空值，且 MUST NOT 报错

#### Scenario: 禁止跨库直查 event 表

- **WHEN** history-service 运行于仅配置 `HISTORY_DB_LINK` 的环境
- **THEN** 补全 `event_unit` 的实现 MUST NOT 调用 `dao.Event` 或等价 history 库内 event 表访问

### Requirement: History 读路径 MUST 暴露 eventUnit

`GET /device/history/api/list`、`GET /device/history/api/piece` 及单条查询响应中的 history 实体 MUST 包含 `eventUnit` 字段（JSON camelCase），值与数据库 `event_unit` 一致。

#### Scenario: 列表返回单位

- **WHEN** 数据库行 `event_unit=ml`
- **THEN** 列表 JSON 中对应记录的 `eventUnit` SHALL 为 `ml`

### Requirement: 历史管理页 MUST 展示计数单位

`resource/public/history.html` 对 `eventType=number` 的记录，展示数量时 MUST 在数字后附加单位：优先使用记录 `eventUnit`；若为空且所选事件 option 含 unit，MAY 回退展示 option 单位。

#### Scenario: 有 eventUnit 的计数记录

- **WHEN** 列表项 `eventNumber=120` 且 `eventUnit=ml`
- **THEN** 页面数量列 SHALL 展示含 `ml` 的可读文本（如 `120ml` 或 `120 ml`，实现 MUST 文档化一种格式）

---

### Requirement: History service SHALL persist optional expect prediction fields

The History domain MUST support optional fields `expectStartTime`, `expectEndTime`, and `expectAmount` (JSON lowerCamelCase; These fields MUST be non-required on add. When present on add, the service MUST persist them. List/get and History WebSocket push payloads MUST include stored values when non-null. History 域 MUST 支持可选字段 `expectStartTime` / `expectEndTime` / `expectAmount`（JSON lowerCamelCase； add MUST 非必填； 有值 MUST 落库。 list/get 与 History WS 推送 MUST 在非空时回传。

#### Scenario: Add without expect fields succeeds

- **WHEN** 客户端调用历史 event/add 且 body 不含任何 expect*
- **THEN** 服务端 MUST 在其它字段合法时成功落库
- **AND** 回传实体上 expect* MUST 可为空

#### Scenario: Add with expectStartTime persists and lists

- **WHEN** add body 含合法 `expectStartTime`
- **THEN** 服务端 MUST 持久化该值
- **AND** 随后 list/WS MUST 回传同一 `expectStartTime`

### Requirement: History update SHALL NOT overwrite stored expect fields

On history event/update (including setting `endTime` to stop an active timing) and on voice/intent `event/batch` update/end operations that use the general history update column set, the service MUST NOT replace existing non-null `expectStartTime` / `expectEndTime` / `expectAmount` with request body values. The service MUST ignore expect* on those update requests. This freeze MUST NOT apply to the dedicated expect-batch endpoint defined in this change, which MAY write.

#### Scenario: Stop timing keeps expectStartTime

- **WHEN** 某记录已有 `expectStartTime`，客户端 update 仅更新 `endTime`
- **THEN** 库中 `expectStartTime` MUST 保持不变
- **AND** 即使 body 误带 expect* MUST NOT 覆盖原值

#### Scenario: Batch CRUD update still freezes expect

- **WHEN** 客户端经 `event/batch` op=update 更新备注或时间且 body 含 expect*
- **THEN** 服务端 MUST NOT 用该 expect* 覆盖库中已有值

### Requirement: History service SHALL expose expect-batch patch endpoint

The History service MUST provide an authenticated device API at `POST /device/history/api/event/expect-batch` that accepts `deviceNo` and a non-empty list of items each with history `id` and optional `expectStartTime` / `expectEndTime` / `expectAmount`. An empty `items` array MUST result in HTTP 400. More than 200 items MUST result in HTTP 400. Present keys MUST update the corresponding stored expect columns for rows that belong to that `deviceNo`; omitted keys MUST leave.

#### Scenario: Batch patches expectStartTime for owned ids

- **WHEN** 合法 Bearer 调用 expect-batch，items 含本 device 下两行 id 及新的 `expectStartTime`
- **THEN** 服务端 MUST 持久化这两行的新 `expectStartTime`
- **AND** 未出现在请求中的 expectEndTime/expectAmount MUST 保持原值（若库中已有）
- **AND** 响应 MUST 为 `{ "ok": true }`

#### Scenario: Empty items rejected

- **WHEN** 合法 Bearer 调用 expect-batch 且 `items` 为空数组
- **THEN** 服务端 MUST 返回 HTTP 400
- **AND** MUST NOT 修改任何行

#### Scenario: Foreign id does not block other patches

- **WHEN** items 中混有本机合法 id 与不属于该 `deviceNo` 的 id
- **THEN** 服务端 MUST 持久化合法 id 的 expect 变更
- **AND** MUST NOT 因外机 id 回滚已成功行
- **AND** 因至少一条成功，响应 MUST 为 `{ "ok": true }`
- **AND** 响应 MUST NOT 包含逐条错误明细

#### Scenario: All items fail yields ok false

- **WHEN** 合法 batch 中每条均无法更新（例如 id 全不存在）
- **THEN** 服务端 MUST 返回 HTTP 200
- **AND** 响应 MUST 为 `{ "ok": false }`
- **AND** MUST NOT 修改任何行

#### Scenario: Omitting expectAmount leaves amount unchanged

- **WHEN** 某行已有 `expectAmount`，batch item 仅带 `expectStartTime`
- **THEN** 库中 `expectAmount` MUST 不变

### Requirement: Expect-batch success SHALL notify with full expect fields

After a successful expect-batch item write, the service MUST reload the history row (or otherwise assemble a complete entity including stored expect*) before updating any in-process cache and before publishing the History WebSocket `update` event for that item. Failed/skipped items MUST NOT emit a stripping update. The service MUST NOT publish an update payload that strips previously stored expect* solely because the patch writer omitted reconstructing them. expect-batch.

#### Scenario: WS update after expect-batch includes expectStartTime

- **WHEN** expect-batch 成功将某行 `expectStartTime` 从空改为有值
- **THEN** 随后 WS update（或等价推送）MUST 携带该 `expectStartTime`

### Requirement: Expect-batch App API SHALL NOT count toward usage stats

The gateway-app usage stats path MUST treat `POST /device/history/api/event/expect-batch` as a maintenance/excluded API (exact method+path in `usagestats` maintenance skip), per owner confirmation for this change. Successful or failed calls MUST NOT increment App API usage counters for that apiKey.

gateway-app 使用统计 MUST 将 `POST /device/history/api/event/expect-batch` 记为维护/排除接口（`usagestats` maintenance 精确 method+path），按本变更负责人确认。成功或失败调用 MUST NOT 计入该 apiKey 的 App API usage。

#### Scenario: Expect-batch excluded from usage

- **WHEN** 客户端经 gateway-app 调用 `POST /device/history/api/event/expect-batch`
- **THEN** 该请求 MUST NOT 计入 App API 使用统计

---

### Requirement: 历史记录筛选 API

系统 SHALL 提供 `GET /device/history/api/filter` 接口，支持按设备号、事件ID列表、时间范围和返回条数上限筛选历史记录。

#### Scenario: 正常筛选 - 所有参数都传
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1,2,5`、`startTime=1234567890`、`endTime=1234567990`、`limit=50` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在时间范围内、事件ID为1/2/5的最多50条历史记录
- **AND** 结果 SHALL 按 id 倒序排列（最新在前）

#### Scenario: eventIds 为空时跳过事件过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=`（空串）、`startTime=1234567890`、`endTime=1234567990` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在时间范围内的所有事件类型的历史记录（不按事件ID过滤）

#### Scenario: startTime 为空时跳过开始时间条件
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1`、`startTime=0`、`endTime=1234567990` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在 endTime 之前、事件ID为1的历史记录（不限制开始时间）

#### Scenario: endTime 为空时跳过结束时间条件
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1`、`startTime=1234567890`、`endTime=0` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在 startTime 之后、事件ID为1的历史记录（不限制结束时间）

#### Scenario: limit 默认值为 100
- **WHEN** 客户端以 `deviceNo=XXX`、`limit=0`（或不传 limit）调用 filter 接口
- **THEN** 系统 SHALL 最多返回 100 条历史记录

#### Scenario: limit 上限为 500
- **WHEN** 客户端以 `deviceNo=XXX`、`limit=1000` 调用 filter 接口
- **THEN** 系统 SHALL 将 limit 限制为 500，最多返回 500 条历史记录

#### Scenario: deviceNo 为空时返回空列表
- **WHEN** 客户端以 `deviceNo=`（空串）调用 filter 接口
- **THEN** 系统 SHALL 返回空列表（不报错）

#### Scenario: 响应格式正确
- **WHEN** filter 接口成功返回
- **THEN** 响应 JSON SHALL 为 `{ "list": [ { "id": ..., "deviceNo": ..., "eventId": ..., ... } ] }` 格式
- **AND** 每条记录 SHALL 包含与 v1 list 接口相同的字段

#### Scenario: 支持 local/remote/canary 服务模式
- **WHEN** 服务运行在 local 模式
- **THEN** filter 接口 SHALL 直连数据库执行查询
- **WHEN** 服务运行在 remote 模式
- **THEN** filter 接口 SHALL 通过 HTTP 调用远程 history-service
- **WHEN** 服务运行在 canary 模式
- **THEN** filter 接口 SHALL 按设备号一致性分流到本地或远程

### Requirement: filter 支持 ignoreTimeRange 强制忽略时间窗

系统 SHALL 在既有 `GET /device/history/api/filter` 上接受可选查询参数 `ignoreTimeRange`。未传或为假时，时间筛选语义 MUST 与现网一致（`startTime`/`endTime` > 0 才施加对应条件，=0 跳过）。为真时，系统 MUST **完全忽略**请求中的 `startTime` 与 `endTime`（即使均为正数），MUST NOT 对其施加任何时间 WHERE 条件；`deviceNo`、`eventIds`、`remark`、`limit` 与按 id 倒序等既有规则 MUST 保持不变。本参数为 additive 扩展，MUST NOT 要求客户端升级后才能调用原筛选行为。

#### Scenario: 默认不忽略时间（兼容旧客户端）

- **WHEN** 客户端调用 filter 且未传 `ignoreTimeRange`（或显式为假），并传入 `startTime`/`endTime` 正数
- **THEN** 系统 SHALL 按现网规则对 `start_time` 施加对应上下界过滤

#### Scenario: ignoreTimeRange 为真时忽略已填时间

- **WHEN** 客户端以 `ignoreTimeRange` 为真，且同时传入非零 `startTime` 与/或 `endTime` 调用 filter
- **THEN** 系统 SHALL 返回结果时 MUST NOT 使用上述时间值作为过滤条件
- **AND** 其他条件（如 `eventIds`、`remark`、`limit`）SHALL 仍按既有规则生效

#### Scenario: ignoreTimeRange 为真且无事件/备注时仍按 limit 收口

- **WHEN** 客户端以 `ignoreTimeRange` 为真、`eventIds` 为空、`remark` 为空调用 filter
- **THEN** 系统 SHALL 在无时间条件下按设备维度返回最多 `limit` 条（默认/上限规则与现网一致）的历史记录（id 倒序）

#### Scenario: local/remote/canary 透传一致

- **WHEN** history 以 remote 或 canary 模式调用 filter
- **THEN** `ignoreTimeRange` 为真时 MUST 经 HTTP 契约透传至 history-service，本地与远程语义 MUST 一致

---

### Requirement: v2 历史列表 API

系统 SHALL 提供 `GET /device/history/api/v2/list` 接口，在 v1 分页基础上扩展 startTime、endTime、limit 可选参数。

#### Scenario: 不传新参数时行为与 v1 完全一致
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20` 调用 v2 list 接口（不传 startTime、endTime、limit）
- **THEN** 系统 SHALL 返回与 v1 list 接口完全相同的结果（分页、排序、字段完全一致）
- **AND** v1 list 接口 SHALL 保持完全不变

#### Scenario: 传了 limit 时优先使用 limit 替代 pageSize
- **WHEN** 客户端以 `deviceNo=XXX`、`page=2`、`pageSize=20`、`limit=100` 调用 v2 list 接口
- **THEN** 系统 SHALL 使用 limit=100 作为每页条数，page 忽略（固定为1）
- **AND** total 字段 SHALL 仍然返回符合条件的总记录数

#### Scenario: 传 startTime 时按开始时间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20`、`startTime=1234567890` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time >= 1234567890 的记录
- **AND** total 字段 SHALL 仅统计符合时间条件的记录数

#### Scenario: 传 endTime 时按结束时间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20`、`endTime=1234567990` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time <= 1234567990 的记录

#### Scenario: 同时传 startTime 和 endTime 时按时间区间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`startTime=1234567890`、`endTime=1234567990` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time 在 [1234567890, 1234567990] 区间内的记录

#### Scenario: pageSize 上限仍为 100（与 v1 一致）
- **WHEN** 客户端以 `deviceNo=XXX`、`pageSize=200` 调用 v2 list 接口（未传 limit）
- **THEN** 系统 SHALL 将 pageSize 限制为 100

#### Scenario: 排序使用 id 倒序
- **WHEN** v2 list 接口返回结果
- **THEN** 结果 SHALL 按 id 倒序排列（与 v1 一致，最新在前）

#### Scenario: 时间单位为 Unix 秒
- **WHEN** 客户端传 `startTime` 或 `endTime`
- **THEN** 系统 SHALL 按 Unix 秒解释该值（与 piece 接口一致）
- **AND** 值为 0 时 SHALL 视为"不限制"，跳过对应时间条件

#### Scenario: deviceNo 为空时返回空分页
- **WHEN** 客户端以 `deviceNo=`（空串）调用 v2 list 接口
- **THEN** 系统 SHALL 返回空列表和 total=0（不报错）

#### Scenario: 响应格式与 v1 一致
- **WHEN** v2 list 接口成功返回
- **THEN** 响应 JSON SHALL 为 `{ "list": [...], "total": N, "page": N, "pageSize": N }` 格式（与 v1 完全一致）

#### Scenario: 支持 local/remote/canary 服务模式
- **WHEN** 服务运行在 local 模式
- **THEN** v2 list 接口 SHALL 直连数据库执行查询
- **WHEN** 服务运行在 remote 模式
- **THEN** v2 list 接口 SHALL 通过 HTTP 调用远程 history-service
- **WHEN** 服务运行在 canary 模式
- **THEN** v2 list 接口 SHALL 按设备号一致性分流到本地或远程

---

### Requirement: 事件区间查询 piece 接口

history-service SHALL 提供 `GET /device/history/api/piece`，接受查询参数 `eventId`、`startTime`、`endTime`、`deviceNo`，并 SHALL 返回该设备在指定时间区间内、指定事件类型下的历史记录集合（字段与现有 history 列表语义一致或可被子集化但 MUST 文档化）。

#### Scenario: 有数据区间

- **WHEN** 参数合法且数据库存在匹配记录
- **THEN** 响应 SHALL 包含记录列表且顺序与产品设计一致（例如按时间升序）

#### Scenario: 无数据

- **WHEN** 区间内无匹配记录
- **THEN** 响应 SHALL 返回空列表而非错误，除非参数非法

### Requirement: piece 结果 Redis 缓存

history-service SHALL 对 piece 查询结果使用 Redis 缓存以降低数据库压力；缓存键 MUST 能区分 eventId、startTime、endTime、deviceNo 的组合。

#### Scenario: 缓存命中

- **WHEN** 相同查询在 TTL 内重复到达
- **THEN** 服务 MAY 从 Redis 返回缓存结果且结果与数据库一致

### Requirement: 历史 CUD 后发布 Redis 通知

history-service SHALL 在任何导致 history 表新增、更新或删除成功的业务路径完成后，向约定 Redis channel 发布一条消息；消息体 MUST 包含 `device_no`、操作类型 `action`（create、update、delete 之一）以及供前端更新的历史记录载荷。当记录存在非空 `event_unit` 时，载荷 MUST 包含 `eventUnit` 字段且值与数据库一致。

#### Scenario: 新增历史

- **WHEN** 新增一条 history 成功提交且 `event_unit=ml`
- **THEN** 系统 SHALL PUBLISH 一条 action 为 create 的消息且包含新记录标识、展示所需字段及 `eventUnit=ml`

#### Scenario: 更新或删除历史

- **WHEN** 更新或删除 history 成功提交
- **THEN** 系统 SHALL PUBLISH 对应 update 或 delete 的消息且包含受影响记录的主键或 event 关联信息；update 时 MUST 包含最新 `eventUnit`

### Requirement: CUD 后失效 piece 缓存

history-service SHALL 在 history 表发生增删改并成功提交后，使与该 device_no（及必要时 eventId）相关的 piece 缓存失效，以保证后续 piece 查询不返回陈旧数据。

#### Scenario: 写入后查询一致

- **WHEN** 同一 device_no 在写入后立刻发起 piece 查询
- **THEN** 查询结果 SHALL 反映刚写入的数据（通过失效缓存或直接读库达成）

### Requirement: History 投影事件 MUST 携带 event_unit

`history.record.created` / `history.record.updated` 域 outbox 与 Redis 缓存投影消费载荷 MUST 包含 `event_unit`（或等价 JSON 字段），与权威库 `history.event_unit` 一致；投影写入 Redis 读模型时 MUST 保留 `EventUnit`。

#### Scenario: 创建投影含单位

- **WHEN** outbox 发布 `history.record.created` 且对应行 `event_unit=次`
- **THEN** 投影载荷 MUST 含 `event_unit=次`，且 Redis 列表缓存中该条记录的 `eventUnit` SHALL 为 `次`

### Requirement: 语音结束事件 MUST 与 API 结束等价推送 WS

当用户通过语音链路触发「结束」计时类事件（`ActionTargetTypeEnd` / Python `target_type=feeding` 且 `action=end`）且系统向用户播报结束成功时，voice-service MUST 经 `DeviceHistory` 契约完成 history 写库，且该写库 MUST 在 history-service 侧触发与 App 调用 `POST /device/history/api/event/end-latest`（`updated=true`）或等价 `event/update` 相同的 Redis 通知：向 `app:history:notify` PUBLISH，`action` 为 `update` 或 `create`（与实落库操作一致），载荷含 `deviceNo`、记录主键及展示字段。 `EndLatestHistoryIfMatch` / `end-latest` 的匹配权威 MUST 为：该 `deviceNo` 下指定 `eventId` 且.

#### Scenario: 结束最近一条同 eventId 未闭合计时记录

- **WHEN** 用户语音或 App 结束事件 E，且 history-service 存在 `eventId=E` 且 `end_time=0` 的记录（取 `id` 最大者），即使全局最新一条 history 的 `eventId` 不等于 E
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=true`
- **AND** 系统 SHALL PUBLISH `action=update` 的 WS 通知
- **AND** 语音路径在写库成功后 MUST 表示结束成功

#### Scenario: EndLatest 未匹配时降级写库仍推送

- **WHEN** 用户语音结束事件 E，且该设备不存在 `eventId=E` 且 `end_time=0` 的记录，因而 `EndLatestHistoryIfMatch` 对 E 返回 `updated=false`
- **THEN** voice-service MUST 执行降级写库（至少 `AddHistory` 写入 E 的瞬时结束记录）
- **AND** 降级写库成功后系统 SHALL PUBLISH `action=create` 或 `update` 的 WS 通知
- **AND** 语音回复 MUST 仅在至少一次写库成功后可表示结束成功

#### Scenario: 与 App end-latest 行为一致

- **WHEN** 同一 `deviceNo` 下 App 调用 `event/end-latest` 成功（`updated=true`）与语音结束同一未闭合 event 在 history 库产生相同终态
- **THEN** 两种路径触发的 WS 载荷在 `action`、记录 `id`、`eventId`、`endTime` 语义上 MUST 一致（允许 `remark` 来源不同）

---

### Requirement: 历史画像接口返回昵称

系统 SHALL 扩展历史画像读取接口（`GET /device/history/api/birthday`）返回 `nickname` 字段；该字段 MUST 通过 device 画像契约获取，history-service SHALL NOT 直连查询 device 域库表。

#### Scenario: 已有昵称
- **WHEN** 目标设备存在可用昵称
- **THEN** 响应 SHALL 包含非空 `nickname`

#### Scenario: 无昵称
- **WHEN** 目标设备当前无昵称
- **THEN** 响应 SHALL 返回 `nickname` 为空串，且接口 SHALL 保持成功响应

### Requirement: 历史页面展示昵称

系统 SHALL 在历史记录页面展示 `nickname`，并与既有性别展示共存；接口返回为空时页面 MUST 显示空态文案而非报错。

#### Scenario: 页面加载成功展示昵称
- **WHEN** 页面加载到包含 `nickname` 的画像数据
- **THEN** 页面 SHALL 显示昵称文本并维持原有性别主题逻辑

#### Scenario: 昵称为空时降级展示
- **WHEN** 接口返回 `nickname` 为空
- **THEN** 页面 SHALL 显示“未设置昵称”或等价占位，并 SHALL NOT 阻断其他历史数据渲染

---

### Requirement: history-service 进程 MUST 仅直连本域持久化表

`history-service`（及其独立配置所连接的默认数据库）MUST 仅对 `history` 与 `domain_outbox` 表执行 DAO/SQL 读写（不含只读副本或显式配置的跨库迁移工具）。MUST NOT 在 history 进程内对 `user`、`event`、`action`、`qa`、`suggest` 等他域业务表执行直连访问。

#### Scenario: 独立部署 history 库仅含 history 与 outbox

- **WHEN** 运行 `history-service` 且数据库中仅存在 `history` 与 `domain_outbox` 业务表
- **THEN** 服务 MUST 能完成历史记录与 outbox 相关功能，且 MUST NOT 因缺少他域表而依赖本地 DAO 回退直查

#### Scenario: 代码评审检查 history 包 import

- **WHEN** 评审 `internal/services/history` 或 history 进程入口的变更
- **THEN** MUST NOT 引入对 `dao.User`、`dao.Event`、`dao.Suggest`、`dao.Qa`、`dao.Action` 等他域 DAO 的直连依赖用于业务读写

### Requirement: 跨域数据 MUST 通过契约获取

当 history 域逻辑需要设备画像、事件字典或语音建议等非 history 表数据时，MUST 通过 **device-service / voice-service** 的 HTTP（或已批准的事件契约）获取，MUST NOT 在同一进程内直查他域表。

#### Scenario: 元数据或画像由其他服务提供

- **WHEN** 上层仍通过统一 `Contract` 需要「事件选项」或「生日」等能力
- **THEN** 实现 MUST 路由到对应服务客户端，而非 `history/local.go` 内对 `dao.Event` 或 `dao.User` 的查询

---

### Requirement: history 文本 chat MUST HTTP 委派 voice-service

history-service 处理 `POST /device/history/api/chat` 时 MUST 经 HTTP 调用 voice-service internal 文本 chat 契约，MUST NOT 在 history 进程内 import 或调用 `voice.TextChat` / `voice.Voice()` 执行业务。

history-service MUST NOT 配置 `VOICE_DB_LINK` 或访问 voice 库表（含 `ai_quota_default`、`qa`、`suggest`、`llm_lane_config`）。

对外路径、请求体字段（`deviceNo`、`transcript`）与成功响应结构 MUST 保持不变。

#### Scenario: App chat 经委派完成

- **WHEN** App 调用 `POST /device/history/api/chat` 且 voice-service 可达
- **THEN** history-service MUST 向 voice-service 发起 internal text chat HTTP 请求并返回等价 `reply`

#### Scenario: 额度错误透传

- **WHEN** voice-service internal chat 返回 40301 或 40302
- **THEN** history-service 对外响应 MUST 携带相同 business code 与 message 语义

#### Scenario: 无 voice 库配置

- **WHEN** history-service 进程未配置 voice 数据库连接
- **THEN** chat 路径 MUST 仍可工作（依赖 voice-service HTTP，不依赖本地 voice 库）

---

### Requirement: Voice MUST 通过 History HTTP 契约获取历史域数据

`voice-service` 在涉及 **历史记录**（由 `history` 表承载的会话/事件时间线数据）的查询或写入时 MUST 通过 `history-service` 暴露的内部 HTTP 接口完成，MUST NOT 直接访问 history 领域数据库表。用户画像（生日、性别等）、事件类型字典、动作记录等 **非 history 表** 数据 MUST NOT 通过本需求所述的 history 接口冒充权威来源，MUST 分别遵循 device 与 voice 域契约。

#### Scenario: 查询历史记录用于对话生成

- **WHEN** voice 处理“查询历史记录”或需要最近历史上下文的请求
- **THEN** voice MUST 调用 history 内部查询接口获取数据，并使用返回结果生成回复

#### Scenario: History 服务不可达

- **WHEN** voice 调用 history 内部接口超时或网络失败
- **THEN** voice MUST 返回可观测的错误语义，并按照配置决定是否执行本地兜底（仅迁移期允许）

### Requirement: History 内部接口 MUST 提供稳定错误语义

history 内部接口 MUST 对参数错误、资源不存在、服务异常返回可区分的错误结构，供 voice 做一致错误处理与日志分类。

#### Scenario: 参数不合法
- **WHEN** voice 传入缺失 `deviceNo` 或非法参数
- **THEN** history MUST 返回明确的参数错误码与错误信息，voice MUST 将其映射为调用方可理解的失败结果

#### Scenario: 服务端内部异常
- **WHEN** history 在处理请求时发生内部错误
- **THEN** history MUST 返回统一错误结构，voice MUST 记录失败原因并输出统一告警日志

### Requirement: Voice MUST 通过 Device HTTP 契约访问用户画像与事件字典

`voice-service` 需要读取或更新 **设备用户画像**（如生日、性别）或 **事件/动作** 相关持久化数据时 MUST 通过 `device-service` 暴露的 HTTP 契约完成，MUST NOT 使用 `dao.User`、`dao.Event`、`dao.Action` 直连 device 库表。

#### Scenario: 事件抽取结果落库

- **WHEN** voice 在理解流程中创建或解析事件实体并需持久化
- **THEN** voice MUST 调用 device 服务接口（或已批准的适配层），MUST NOT 在 voice 进程内对 `event` 表执行 DAO Insert
- **AND** 新建事件时 MUST 向 device 传递合法 `eventType`（`number` | `time` | `one`）

#### Scenario: 读取事件选项列表

- **WHEN** voice 需要事件字典列表或 `eventType` 等元数据
- **THEN** voice MUST 从 device 服务获取，MUST NOT 依赖 history 服务返回的 `event` 表投影作为权威来源
- **AND** 响应项 SHALL 含 `eventType`，SHALL NOT 含 `needQuantity`

### Requirement: Voice MUST 在本域处理 suggest 表

`voice-service` 对 **`suggest` 表** 的读写 MUST 仅在 voice 进程内通过本域 DAO 或本域服务接口完成；history-service MUST NOT 作为 suggest 数据的权威存储进程。

#### Scenario: 写入每日建议

- **WHEN** voice 生成并保存建议文案
- **THEN** 持久化 MUST 发生在 voice 库 `suggest` 表路径上，MUST NOT 由 history 进程执行 `dao.Suggest` 写入

### Requirement: Voice 写入 history SHOULD 传递已知 eventUnit

当 voice-service 经 device HTTP 契约已解析出 `entity.Event` 且 `Unit` 非空时，调用 history 新增/更新接口的请求体 SHOULD 携带 `eventUnit` 字段。history-service MUST 仍支持未携带时经 device 契约补全。

#### Scenario: 语音记录奶量事件

- **WHEN** voice 解析到事件 `unit=ml` 并调用 history 新增一条计数记录
- **THEN** HTTP 请求体 SHOULD 包含 `eventUnit=ml`，且 history 持久化后 `event_unit` SHALL 为 `ml`

#### Scenario: Voice 未传 eventUnit 时由 history 补全

- **WHEN** voice 调用 history 新增记录未传 `eventUnit` 但 `eventId` 对应主档 `unit=ml`
- **THEN** history-service MUST 经 device 契约补全并成功写入 `event_unit=ml`

---
