## Why

Voice 运维额度页仅支持「手输 wxId → 单人 override」，无法浏览全量用户有效额度与身份信息，运维排查成本高且与「按设备找人改额度」的日常流程不匹配。需要以 device 域全部 wx 为全集分页展示已用/上限，并去掉功能重叠的单人 override 模块。

## What Changes

- **新增** voice admin 用户额度分页列表 API：按 device 全部真实 wx 分页，聚合当月已用、有效上限（override ∨ 全局默认）与用户身份字段；支持按 `deviceNo` 查询。
- **改造** `voice-admin.html`：保留全局默认；**移除**「单人 override（wxId）」表单模块；新增分页表格（`deviceNo` 为第一列），一行同时展示喂养 + 胖宝的「已用 / 上限」四列额度，并支持行内修改上限（写回既有 per-wxId override）。
- **说明文档**：在变更说明中列明喂养 AI（`voice_ai`）与胖宝 AI（`clinic_ai`）额度变更影响的业务接口清单。
- **非范围**：不改 App 读额度契约、不改 check/consume 语义、不改月度 Redis 键、不改 UCG 润笔额度页、不新增后台 ticker、不为列表默认引入新 Redis 读缓存键族。

## Capabilities

### New Capabilities

- （无）行为落在既有 voice 额度与运维 UI 能力上扩展。

### Modified Capabilities

- `voice-admin-ui`：额度区从「全局默认 + 单人 override」改为「全局默认 + 全量 wx 分页额度表」；去掉单人 override 模块；表格列与 deviceNo 查询行为。
- `voice-ai-quota`：新增 admin 用户额度分页列表 API 契约（聚合身份 + used/limit）；明确改上限仍经既有 PUT user / override 语义。

## Impact

- **前端**：`resource/public/voice-admin.html`（及如需的 admin 样式/公共脚本复用）。
- **后端**：`api/v1/voice_ai_quota_http.go`、`internal/controller/voice_app_ai_quota.go`、`internal/services/voice/ai_quota*.go`；经 `DeviceAdmin().ListWxPage`（及必要时扩展 wx 列表项含 `babyName`）聚合身份，禁止 voice 直查 device/history 表。
- **网关**：沿用既有 `/voice/admin/api/*` 反代与 Admin JWT → `X-Admin-Password` 注入；新增路径须落在已绑定 admin 前缀内。
- **受影响业务入口（改上限后 limit/degraded 阈值变化，路由本身不变）**：
  - **喂养 AI (`voice_ai`)**：WS `/voice/chat/ws`；`POST /voice/internal/api/text/chat`；`POST /voice/internal/api/text/chat/stream`；`POST /device/history/api/chat`；`POST /device/history/api/chat/stream`；`POST /voice/text/chat`。
  - **胖宝 AI (`clinic_ai`)**：WS `/voice/clinic/ws`。
  - **共用读写**：`GET /voice/app/api/ai-quota`；`POST /voice/internal/api/ai-quota/check|consume`；`GET/PUT /voice/admin/api/ai-quota/default`；`GET/PUT /voice/admin/api/ai-quota/user`；**新增**列表 API。
- **usage 统计**：本变更为 voice **admin** API / 静态运维页，不新增 gateway-app 对外 App HTTP；无需改 `usagestats/maintenance_skip.go`（若评审认定否则再确认负责人）。
