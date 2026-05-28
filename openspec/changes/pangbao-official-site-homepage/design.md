## Context

当前 `gateway-app-server` 在 `GET /` 上仅返回纯文本“智能语音 App 网关”，而 `resource/public/` 已经承载多个静态页，说明该进程适合托管官网类页面。现有基础能力已经覆盖本次需求的大部分数据来源：

- 事件字典可通过 device-service 权威链路读取，事件实体已包含 `name`、`logo`、`color`、`parentId`。
- 事件 logo 已统一为 path-only，并可通过 `gateway-app-server` 同源代理 `/ai_talk_images/*` 访问。
- Android 最新版本下载信息已由 `gateway-app-server` 从 `ai_voice_app.version` 表读取，并通过 path-only `downloadUrl` 对外暴露。

本次变更的核心约束有三条：

1. 官网入口放在 `gateway-app-server` 的 `/`，替换现有纯文本，但不得影响主网关进程。
2. 跨服务读取必须遵守服务边界：事件数据走服务契约，不能在 gateway-app 内直接读 device 域库表。
3. 官网面向匿名公众访问，不能复用依赖 Bearer 或管理口令的业务接口给前端直连。

## Goals / Non-Goals

**Goals:**
- 在 `gateway-app-server` 的 `/` 提供“胖宝”官网首页。
- 首页使用玻璃拟态风格，展示品牌定位、事件 logo/名称和下载说明。
- 提供一个匿名只读聚合接口，统一给官网返回事件列表、Android 下载地址与 iOS 提示。
- 保持主网关、既有业务 API、事件图片代理与 APK 下载路由行为稳定。

**Non-Goals:**
- 不改动主网关根路径，不将官网扩展到 `register.go` 绑定的主网关进程。
- 不新增事件编辑、版本上传或其他运营后台能力。
- 不新增 iOS 下载地址管理或 IPA 分发能力；iOS 仅展示 App Store 搜索提示。
- 不引入跨服务数据库直连，也不重构现有版本检查逻辑。

## Decisions

### 1. 官网仅挂载在 `gateway-app-server` 的 `/`
**Decision**：将 `internal/controller/gateway_app_register.go` 中的 `/` 处理器从纯文本改为返回官网静态页，例如 `resource/public/pangbao-home.html`，且不修改主网关注册逻辑。

**Why**：用户明确要求官网放在 `/`，同时要求“不能影响主网关”。当前只有 `gateway-app-server` 已承载 App 侧静态资源和下载相关能力，因此把官网限制在该进程最小化影响面。

**Alternatives considered:**
- **挂到 `/official` 等子路径**：风险更低，但不满足“官网放在 /”的产品要求。
- **同步替换主网关 `/`**：会扩大影响面，与“不能影响主网关”冲突。

### 2. 新增官网匿名聚合接口，而不是让前端直连现有业务接口
**Decision**：新增一个官网专用匿名只读接口，例如 `GET /device/app/api/site/home`，由 `gateway-app-server` 组装官网需要的全部数据。

**Why**：现有事件接口位于业务域，匿名访问策略并不适合直接暴露给官网前端。新增聚合接口可以将匿名公开数据与业务接口隔离，同时保持前端实现简单、缓存策略清晰。

**Alternatives considered:**
- **前端直接调 `event/options` 与 `version/check`**：需要重新审视匿名白名单，容易把业务接口暴露给公众页面。
- **服务端 SSR 将所有数据写死进 HTML**：无法优雅处理版本变更与事件更新，也不利于后续缓存与前端刷新。

### 3. 事件数据经服务契约读取，Android 下载信息复用 gateway-app 既有能力
**Decision**：`gateway-app-server` 通过已有 HTTP 契约读取 device 域事件字典，并复用本进程读取最新版本行的逻辑，生成官网响应所需的 Android 下载信息。

**Why**：事件权威不在 gateway-app，本仓库明确禁止跨服务直查他域 DAO/库表。Android 版本数据本就由 gateway-app 持有并对外提供版本检查，因此复用本进程能力比重复造轮子更安全。

**Alternatives considered:**
- **gateway-app 直接查询 `event` 表**：违反服务边界。
- **官网前端自行拼接 path-only `downloadUrl`**：可行，但会把站点基址、缺省空值处理和二维码条件逻辑散落在前端；由聚合接口返回最终可用地址更稳。

### 4. 二维码采用浏览器端生成，避免新增图片生成后端
**Decision**：官网页面在浏览器端根据聚合接口返回的 Android 下载地址生成二维码，可采用本地静态资源方式引入轻量二维码库或等价实现。

**Why**：二维码属于展示层逻辑，不必为此增加新的图片渲染 API。浏览器端生成能减少后端复杂度，也避免缓存与二进制输出处理。

**Alternatives considered:**
- **后端提供二维码 PNG 接口**：额外增加输出格式、缓存和签名处理，收益有限。
- **使用第三方在线二维码服务**：引入外部依赖，不利于内网或离线部署。

### 5. 官网数据返回“可直接展示”的字段，而不是暴露底层原始字段
**Decision**：官网聚合接口直接返回页面所需的结构，例如品牌文案、事件展示列表、Android `downloadUrl`/`qrValue`、iOS 提示文案和可用状态。

**Why**：官网页面是品牌展示页，不需要感知底层 `version` 行或完整事件原始结构。输出前收敛字段可以减少前端对内部模型的耦合。

**Alternatives considered:**
- **完全透传 `entity.Event` 和版本检查响应**：短期快，但前端需要自行补齐绝对地址、空值处理和展示语义。

## Risks / Trade-offs

- **[Risk] 根路径替换后，依赖旧文本“智能语音 App 网关”的人工巡检习惯会失效** → 保留健康检查与业务 API 不变；如有需要，可在 runbook 中补充新的首页验收方式。
- **[Risk] 匿名聚合接口扩大公开面** → 接口仅返回品牌展示所需只读字段，不返回管理信息、用户态数据或内部调试信息。
- **[Risk] 事件 logo 或 Android 下载路径为空时页面体验不稳定** → 聚合接口输出显式可用状态，前端对空值提供占位与降级文案。
- **[Risk] 事件数量较多或层级较深导致首页过载** → 首页仅展示可读性友好的事件卡片视图，具体层级结构不作为首版重点。
- **[Risk] 前端二维码库引入额外静态资源** → 采用本地静态资源方式托管，避免 CDN 依赖和外部网络风险。

## Migration Plan

1. 在 `gateway-app-server` 中新增官网静态页与官网聚合接口。
2. 将 `GET /` 绑定切换为返回官网页，仅修改 `gateway_app_register.go`，不触碰主网关注册代码。
3. 部署后验证：
   - `GET /` 返回胖宝官网；
   - `GET /device/app/api/site/home` 匿名可读；
   - `/device/app/api/version/check`、`/device/app/apk/*`、`/ai_talk_images/*` 继续可用；
   - 主网关根路径与业务代理不受影响。
4. 若需回滚，只需恢复 `gateway-app-server` 根路径处理器到原文本输出，并移除官网静态页与接口绑定；其他链路无需数据迁移。

## Open Questions

- 首页事件卡片是否展示全部事件，还是仅展示根事件/精选事件，需要产品进一步确认。
- Android 下载区是否需要同时展示“最新版号/更新说明”，当前需求仅强制要求二维码与下载说明。
- 官网是否需要额外的备案、客服联系方式、隐私说明等品牌合规内容，当前需求未覆盖。
