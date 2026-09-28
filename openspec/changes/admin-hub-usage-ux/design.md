## Context

运维 UI 为 gateway-app 托管的静态 HTML（`resource/public/`）。设备记录已用 `.device-link` → `/device/admin/history/{deviceNo}`；两页使用统计的用户维度走 `usage/wx-list` / `client-usage/wx-list`，经 `DeviceWxListPage` 拉 device `wx/list` 后再补 UCG 昵称，但**丢弃了** device 已返回的 `babyName`，且**无** per-wx `lastAt`。Hub 反馈卡片在问答库之后，入口仅「展开更多」（total>10）。反馈「问题」列使用 `wrapcell`（`vertical-align: middle`）。

约束：跨服务仍走契约；运维读接口同 path **加法字段**（非 App 对外 v+1）；不新增 Redis 读缓存键族以外的访问方式；不新增背景循环；不改 `maintenance_skip`。

## Goals / Non-Goals

**Goals:**

- Hub：反馈模块紧挨设备记录；常驻「查看全部」。
- 反馈：仅问题列顶对齐。
- 统计两页：设备号可点进运维；用户维度序号、本页 lastAt 排序、本页有无宝宝名筛选。
- 同 path 为 wx-list 补 `babyName` + `lastAt`（受 `days`）。

**Non-Goals:**

- 不改 App 对外 HTTP；不为 wx-list 新建 v2 路由。
- 不做全局跨页 lastAt 排序或服务端 `hasBabyName` 真分页。
- 不扩展 UCG nickname 语义筛选；序号不进 API。
- 不改设备记录链接目标以外的 history 业务。

## Decisions

### 1. 同 path 加法字段，不新开路由

- **选择**：在既有 `GET /device/admin/api/usage/wx-list` 与 `client-usage/wx-list` 上增加 `babyName`、`lastAt`，可选 `days`（与页顶时间窗一致；缺省与现有统计默认窗一致，如 API usage 7 天、client-usage 受 30 天上限约束）。
- **理由**：运维只读编排补全；device 已有 `babyName`；Redis 已有 per-wx last hash。
- **备选**：严格 v2 新 path（用户明确不要）；前端 N+1 打 `usage/user`（噪声大且拿不到 babyName）。

### 2. lastAt 计算

- **选择**：对本页每个 `wxId`，读 `gw:usage:last:w:{wxId}`（client：`gw:featusage:last:w:{wxId}`）HashGetAll，取字段值 max；若传入 `days>0`，仅保留 `lastAt` 落在窗口内的时间戳，否则视为 0。`days=0` 时与现有「全部」语义对齐（API usage 约 90 天 TTL 内；client-usage 不超过实现约定的 30 天窗）。
- **理由**：一页约 50 行、每行一次 HGETALL，可接受；无需扫交叉日桶。
- **备选**：扫 cross 日键聚合（更重）；全局 ZSET（新键，违反「不默认加 Redis」）。

### 3. 筛选与排序仅当前页（前端）

- **选择**：下拉「全部 / 有宝宝名 / 无宝宝名」滤 `babyName` trim 后非空；点「最近使用时间」表头对内存列表按 `lastAt` 降序，`lastAt` 相同则 wxId 升序；未启用 lastAt 排序时按 wxId 升序。序号为渲染后的 1…N。
- **理由**：产品确认不需要跨页全局排；实现简单。
- **代价**：筛选只影响本页可见行，total 仍为服务端分页 total。

### 4. 设备号跳转

- **选择**：与 Hub 一致：`/device/admin/history/` + `encodeURIComponent(deviceNo)`，空设备号不可点。行点击选中用户与链接点击需避免冲突（链接 `stopPropagation` 或设备号列单独 `<a>`）。

### 5. 反馈问题顶对齐

- **选择**：新增专用 class（如 `feedback-question-cell`），`vertical-align: top`；不改全局 `.wrapcell`。

### 6. Hub 反馈入口

- **选择**：常驻 `link-btn`「查看全部」→ `/device/admin/feedback-records`；可移除或保留「展开更多」逻辑，但入口不得再依赖 total>10。

## Risks / Trade-offs

- [本页筛选误解为全库过滤] → 文案/hint 标明「筛选当前页」；后续若要真分页再开 change。
- [lastAt 与明细略有偏差] → last hash 为该 API/功能维度最近成功时间的 max；与「用户明细里各行 lastAt」一致来源。
- [AGENTS「接口加字段须 v+1」] → 本变更为运维编排加法；design 锁定同 path；若评审坚持 v2，再拆 path。
- [行点击与设备号链接冲突] → 设备号列独立链接并 `stopPropagation`。

## Migration Plan

1. 先部署 gateway-app（含加法字段），旧页忽略新字段仍可用。
2. 再部署/刷新静态页（或同包发布）。
3. 回滚：回退静态页即可恢复旧交互；字段保留无害。

## Open Questions

- （无）explore 已确认：方案 ②、筛选 A、本页排序、两页都做、仅反馈问题列顶对齐。
