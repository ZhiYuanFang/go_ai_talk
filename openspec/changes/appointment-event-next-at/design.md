## Context

喂养事件字典在 device DB 全局表 `event`（无 `deviceNo`），经 Redis `EventOptions` 缓存与 history 代理的 `/device/history/api/event/options` 等路径下发。宝宝发生记录在 history；预测临近待办由 Flutter `PUT /device/api/predict/imminent/pending` 整表写入 voice Redis，延时 MQ 到点后对账 Redis，**不**以 MySQL 为推送权威。

产品需要把「疫苗」类事件标成预约型，并按宝宝记住下次约定时间，供编辑 sheet 回填与客户端预测输入；推送仍完全由客户端 pending 驱动。

## Goals / Non-Goals

**Goals:**

- 字典级 `isAppointment`（默认 false），后台可编辑，options/list 带回。
- device-service 持久化 `(deviceNo, eventId) → nextAt`；独立 App 读写 API；`0` = 无约定。
- 网关对预约 App 路径反代到 device-service；登录 + 绑机校验。
- 明确与 predict 解耦：清空推送靠 pending 移除事件。

**Non-Goals:**

- 修改 predict-imminent 校验、Redis、MQ、推送文案或闸逻辑。
- 预约表投影到 Redis / 服务端代算下次 / 服务端代报 pending。
- history 表挂 nextAt、或把预约 API 塞进 history CRUD。
- 为本能力新建 Redis 读缓存（未做收益率确认；直读 MySQL）。
- 本仓实现 Flutter UI（仅约定契约）。
- 复用或改写 `event_type=one` 语义为预约。

## Decisions

### D1：预约标志挂在全局 `event`，不按宝宝

- **选择**：`event.is_appointment`（JSON `isAppointment`），后台勾选，全宝宝共用语义。
- **理由**：疫苗等「天生预约」由运营配置即可；避免每宝宝一份目录偏好。
- **备选**：宝宝级覆盖表 → 第一期过重，不做。

### D2：下次约定独立表 + 独立 API，键为 `(deviceNo, eventId)`

- **选择**：device DB 新表（建议名 `appointment_next`），唯一键 `(device_no, event_id)`；列至少 `next_at`（unix 秒，`0` 允许）。
- **API（草案，实现可微调 path 但语义不变）**：
  - `GET /device/app/api/appointment/next`：`deviceNo` + `eventId` → `{ nextAt }`；无行视为 `nextAt=0`。
  - `PUT /device/app/api/appointment/next`：body `{ deviceNo, eventId, nextAt }` upsert；`nextAt=0` 表示清空约定（可保留行值为 0，或删行，**对外读语义均为 0**；实现选定一种并在代码注释固定）。
- **理由**：同一预约事件点任意 history 行，下次时间相同；与 history 解耦。
- **备选**：挂在 history 行 → 违背「与记录无关」；拒绝。

### D3：网关挂在 device App 子路径并显式反代

- **选择**：路径落在 `/device/app/api/appointment/*`，在 `installDeviceProxyMiddleware` 增加该 pattern（与现有 `user`/`feedback` 并列）。
- **理由**：预约状态属 device 域；现有 device 反代**没有** `/device/api/*` 总前缀（predict 在 voice）。
- **鉴权**：须 Bearer；**不**进 auth exempt；请求 `deviceNo` MUST 与会话绑机一致（对齐 predict/history 写接口习惯）。
- **usage**：新增 App 路由 MUST 先向负责人确认是否计入；tasks 含确认项，未确认不得改 `maintenance_skip.go`。

### D4：事件 options 加字段，沿用缓存重建

- **选择**：entity/DAO/`ListEvents`/admin add·update 支持 `isAppointment`；变更后走既有 `refreshEventOptionsCacheAfterMutate`。
- **理由**：history/voice 已消费 `[]entity.Event`，加可选字段即可；沿用 cachekit EventOptions 模式（design 声明，免新建 Redis 键）。
- **兼容**：加字段非删改旧字段；客户端忽略未知字段的旧版仍可用（仅无预约分流）。

### D5：与 predict 的客户端协作（服务端零改）

```
清空下次约定：
  PUT appointment nextAt=0
  PUT pending（列表中移除该 eventId）  ← 旧 nextAt=A 的 MQ 消费时对账失败 → 不推

有下次约定：
  PUT appointment nextAt=T
  Flutter 将 T 纳入预测结果并 PUT pending（含该事件）
```

- **选择**：产品接受「清空 = pending 移除」；HTTP 继续拒绝/不依赖 `nextAt=0` 进 pending。
- **理由**：现有整表替换 + `pendingMatches` 已足够。

### D6：空结果查询

- 预约 GET：无行 → 业务上 `nextAt=0`，MUST 用 `.One()` + `IsEmpty()`（或等价），禁止把无行当 5xx。

## Risks / Trade-offs

- **[Risk] Flutter 只写预约表、忘记更新 pending → 旧时间仍可能推送** → Mitigation：契约与联调清单写明「两步」；本仓无法强制，依赖客户端。
- **[Risk] 全局 isAppointment 误标常规事件 → 全用户预测 UI 分流错误** → Mitigation：默认 false；后台谨慎配置。
- **[Risk] options Redis 旧快照缺新字段** → Mitigation：发版后依赖 mutate 重建；必要时运维触发一次事件保存或清 EventOptions 键。
- **[Trade-off] 无 Redis 缓存预约 nextAt** → 编辑 sheet 多一次 MySQL；量小可接受。

## Migration Plan

1. device DB：`event` 加列默认 0；建 `appointment_next`（或最终表名）及唯一索引。
2. 部署 device-service（含 admin + App API）与 gateway-app（反代 pattern）。
3. 运营将疫苗等事件勾选预约；Flutter 发版后启用分流。
4. **Rollback**：网关去掉反代 pattern；忽略新列/新表；旧客户端不受影响。

## Open Questions

- usage 统计：预约 GET/PUT **不计入**（负责人确认；已写入 `maintenance_skip.go`）。
- PUT `nextAt=0`：实现已定为 **行保留 0**（不清行）。
