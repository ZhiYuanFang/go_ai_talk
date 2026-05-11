## Context

仓库已存在独立 `gateway`（根 `main`）、`device-service`、`history-service`，网关通过环境变量将 `/device/history/api/*`、`/device/admin/api/*`、`/voice/text/*` 等路径以 `ReverseProxy` 方式委派到下游；响应封装与领域逻辑在各自进程内。Redis 已有 `internal/platform/cachekit` 封装。`AGENTS.md` 要求跨域数据经 HTTP 契约、进程独立配置、网关不承载 worker 类后台任务。

本变更为 App 增加第二套网关进程 **gateway-app-server**：与现网关共享代理与横切思路，但增加 wx 主键鉴权、令牌、版本库、历史 WS；device/history 扩展 HTTP 契约；history 通过 **Redis Pub/Sub** 驱动 gateway-app 侧推送。

## Goals / Non-Goals

**Goals:**

- gateway-app-server 与现 gateway **能力对齐**（静态页、领域代理模式、跨切中间件），并增加 **Bearer → wx.id → wxCode → `X-Internal-Wx-Code`** 的代理链增强。
- **access_token 为纯 JWT**（RFC 7519 JWS 紧凑序列化），载荷经签名校验后可得到 **`wx` 表主键 id**（见 D9）；网关通过 **device 只读契约** 解析 `wxCode`，并对缓存做短 TTL 与写路径失效。
- **access / refresh 的生成、校验、刷新接口仅在 gateway-app**；device **`POST /device/app/api/user/login`** 只处理 wx 行与业务返回字段（与网关聚合 **`POST /device/app/api/login`** 路径区分）。
- **history CUD** 后 **PUBLISH**；gateway-app **SUBSCRIBE** 并向已认证 **device_no** 的 WS 客户端推送含 **create/update/delete** 的消息体。
- **Redis**：KV 场景优先 `cachekit`；Pub/Sub 使用与现网相同的 Redis 配置族，订阅逻辑在 gateway-app 进程生命周期内管理（启动订阅、退出取消）。

**Non-Goals:**

- 不改变 voice WebSocket 的既有边缘代理语义（仍由现 gateway 配置负责）。
- 不在 device 进程内查询或缓存 history 表数据。
- 不规定具体 JWT 实现库品牌；实现阶段 SHALL 选用成熟 JWT 库完成签发与校验，签名密钥仅由 gateway-app 配置加载。

## Decisions

### D1：进程与配置边界

- **决策**：新增 `cmd/gateway-app-server`，通过 `GF_GCFG_FILE` 指向 **`manifest/config/config.gateway-app-server.yaml`**；其中包含 **server 地址**、**redis**、**database 分组 `app`（`ai_voice_app`）**、以及 **DEVICE_SERVICE_URL / HISTORY_SERVICE_URL**（或沿用现有 `*_PROXY_URL` 命名族，在 tasks 与实现时与现有一致性对齐）。
- **理由**：符合仓库既有「每进程独立配置」约束；`version` 表仅 app 库，不应写入 `manifest/config/config.yaml` 主文件承载业务项。

### D2：登录与令牌拆分

- **决策**：**device** `POST /device/app/api/user/login` 返回 **wxId、wxCode、device_no（可空）、is_new_user** 等；**gateway-app** `POST /device/app/api/login` 调用 device 后签发 **access_token**（**纯 JWT**，见 D9）与 **refresh_token**（高熵随机串 + Redis 存储，**非 JWT**）；刷新接口例如 `POST /device/app/api/token/refresh`（路径在 tasks 中最终确定，前缀固定为 `/device/app/api/`）。
- **备选**：由 device 直接返回 JWT（已拒绝，与「签发仅在 gateway-app」冲突）。

### D3：下游注入 wxCode

- **决策**：鉴权中间件在解析出 **id > 0** 后获取 `wxCode`，对代理请求统一设置 **`X-Internal-Wx-Code`**；**不修改** POST body（与现 ReverseProxy 约束一致）。
- **理由**：避免双写响应体与 body 缓冲问题；device 已有 profile 接口从 body 取参的模式，本变更对 wx 相关接口约定为 **从 Header 读 wxCode**（网关注入）。

### D4：device 提供 id → wxCode

- **决策**：新增 **仅内网** 可调用的只读接口 **`GET /device/app/api/user/internal/by-id`**，入参 **id**，出参 **wxCode**；网关调用时携带可选 **`X-Internal-Gateway`** 或依赖网络策略；失败时中间件返回 401/403 与明确错误码。
- **缓存**：`cachekit.SetEX`，TTL 建议 60–300s；**bindwx / 更新 wx 行** 后对该 id（及 wxCode）做 **Del** 或版本键失效。

### D5：历史 WS 与授权

- **决策**：WS 路径挂在 gateway-app；首帧 `{"type":"auth","access_token":"...","device_no":"..."}`。服务端 **校验 access 中 id** 对应 wx 行 **是否允许访问该 device_no**（例如 `wx.device_no` 与首帧 `device_no` 一致，或业务规定的绑定关系）；未通过则关闭连接或返回错误文本帧。
- **理由**：仅 JWT 验签无法防止「用合法 token 订阅他人设备号」。

### D6：Pub/Sub 形状

- **决策**：使用 **固定 channel 前缀**（实现时定为常量，例如 `app:history:notify`）；消息为 **单条 JSON 字符串**，字段至少包含：`device_no`、`action`（`create`|`update`|`delete`）、`payload`（与列表 API 一致的历史记录结构子集或完整体，以便前端就地更新 UI）。
- **备选**：Redis Streams（更重，本变更按产品选择采用 Pub/Sub）。

### D7：history `piece` 与缓存

- **决策**：`GET /device/history/api/piece` 在 **history-service** 实现，查询 `dao.History` 时间区间与 `eventId`、`deviceNo` 过滤；结果使用 **Redis 缓存**，key 包含四元组哈希；在 **同一 history 进程内** 对 history 表 CUD 成功后 **删除或失效**与该 `device_no` 相关的 `piece` 缓存键（可采用前缀删除或维护版本号键）。

### D8：`auto_save` 无设备号时的创建设备

- **决策**：`POST /device/app/api/user/auto_save` 在处理请求时，若当前 wx（由 Header `X-Internal-Wx-Code` 识别）**尚未绑定** `device_no`，则 SHALL **生成一个新的设备号**：长度为 **6** 的字符串，字符集为 **大写英文字母 A–Z**，在 `user`（或项目既有的设备注册表）中 **插入新设备行**；将该 `device_no` 与 wx 行**绑定**；再写入画像（`birthday`、`sex`，语义与现有 `SaveUserProfile` 对齐）。若 wx **已绑定** `device_no`，则 SHALL **不重新生成**设备号，仅更新画像并返回已有 `device_no`。
- **全局唯一**：`device_no` 在设备注册表中 **MUST 全局唯一**；随机生成的候选值 **MUST NOT** 与库中**任意已有** `device_no` 冲突。实现 SHALL 依赖数据库 **UNIQUE 约束**作为最终权威，并在冲突时 **重新生成候选**（或插入前 `SELECT`/存在性检查后再插入，二者至少满足其一且与唯一约束一致）；重试直至成功或达到实现约定的最大次数后返回可观测错误。
- **理由**：与「登录时尚无 device_no」的 App 流程衔接，设备创建权威仍在 device 域，且不跨库访问 history。

### D9：access_token 的 JWT 形态

- **决策**：`access_token` **MUST** 为符合 **RFC 7519** 的 **JWT**（三段式 JWS compact serialization），由 gateway-app **签发与校验**；**refresh_token** 保持 **不透明串**，不得伪装为 JWT。
- **Claims**：`sub` **MUST** 承载 `wx` 表主键 id（与 device 返回的 wxId 一致，编码为字符串或整型之一并在实现中固定）；**MUST** 包含 `exp` 与 `iat`；算法（如 **HS256** 或 **RS256**）与签名密钥 **MUST** 来自 gateway-app 专用配置（不写死密钥）。
- **校验**：Bearer 中间件与 WS auth **MUST** 先完成 JWT 签名与 `exp` 校验，再解析 `sub` 为大于 0 的 id 并继续 wxCode 解析链。
- **TTL**：access 过期时间 **MUST** 为可配置项（建议在配置中默认 15–60 分钟量级，具体值在实现 PR 中落默认）；refresh 生命周期与是否单次旋转仍由实现与 code review 在单一策略下定稿。

## Risks / Trade-offs

- **[Risk] Pub/Sub 不持久**：断连期间消息丢失 → **缓解**：App 侧仍以轮询或进入前台时拉取列表为主；WS 为增量提示。
- **[Risk] Cluster 与 SUBSCRIBE**：跨 slot 与客户端重连 → **缓解**：单订阅连接、记录重连退避；实现阶段按 goframe/redis 文档验证。
- **[Risk] 内部头伪造**：若 App 可直连 device → **缓解**：网络策略仅允许 gateway 网段访问带 internal 契约的路径；或 internal 路径要求 mTLS/共享密钥（后续加固项）。
- **[Trade-off] 网关多一次 device RTT**：id→wxCode → **缓解**：Redis 短缓存。

## Migration Plan

1. 先部署 **device / history** 新接口与发布逻辑（向后兼容，旧客户端不调用则无影响）。
2. 部署 **gateway-app-server** 与 Redis 配置，联调登录、刷新、代理、WS。
3. 切 App 域名或路径到 gateway-app；现 gateway 保持不变供 Web/管理端使用。
4. **回滚**：App 回退到旧入口；停用 gateway-app Deployment；新接口若无调用可保留。
