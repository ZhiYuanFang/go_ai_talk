## Context

- 事件主档在 device 库 `event` 表；层级用 `parent_id`（0=根）；已有正交标志 `is_appointment`（任意层级可勾，默认 0），经 Hub + options 供 Flutter 分流预约逻辑。
- Flutter 预测需要另一配置位：某**有子树的根**是否把子事件历史合并进同一条预测。算法在客户端；Go 只持久化与下发。
- 复用约定（v3.0.3）：`device-event-hierarchy`、`device-event-cache-rebuild-on-mutate`；模式对齐 `appointment-event-next-at` 中的 `event-is-appointment`（列 ensure + Hub + EventOptions）。

## Goals / Non-Goals

**Goals:**

- 持久化 `merge_predict`，默认 0；JSON `mergePredict`。
- 仅「根且至少有一个直接子」可置 1；Hub 禁用非法勾选；服务端写路径强制清零。
- options/list/内部字典透出；写后重建 EventOptions。

**Non-Goals:**

- 不改 voice/history 预测、临近推送、LA、Hub 准确率同根聚合。
- 不实现 Flutter 预测算法。
- 不为该字段新建 Redis 键或 App 专用 API；不改 usage 统计。
- 不处理微服务启动依赖改造。

## Decisions

1. **列与字段命名**  
   - DB：`merge_predict TINYINT NOT NULL DEFAULT 0`（跟进 `is_appointment`）。  
   - JSON/API/表单：`mergePredict`（0/1 或 bool，Hub 与预约列一致用 0/1）。  
   - 备选 `is_merge_predict` 更啰嗦；用户指定 `merge_predict`。

2. **可勾条件（产品硬约束）**  
   - MUST：`parent_id == 0` **且** `EXISTS child WHERE parent_id = 本 id`。  
   - 叶子、无子根：Hub checkbox disabled；`AddEvent`/`UpdateEvent` 若请求 1 则强制写 0（可 WARN，不必 4xx，避免破坏批量/兼容路径）。  
   - **删光最后子节点**：若父仍为根且 `merge_predict=1`，在**删除子**成功路径或**父的下一次 Update** 清零；优先在 `DeleteEvent` 成功后若被删节点之父变为「无子根」则对该父 `UPDATE merge_predict=0` 并重建缓存（最少额外写、避免脏配置长期挂着）。  
   - **根改挂为非根**（`UpdateEvent` 改 `parentId`）：目标 `parent_id≠0` 时强制 `merge_predict=0`。  
   - **新增子到无子根**：不自动改标志；运维再勾。

3. **落点与契约**  
   - 实现在 `device-service`：`EnsureEventMergePredictColumn`（幂等 ALTER，放在既有 schema ensure 旁）；entity/dao/do；`AddEvent`/`UpdateEvent`/list fields；admin multipart；`clients/device` 与 `contracts` 签名跟预约一样扩参。  
   - Hub：`admin.html` 增加列与 checkbox；前端根据当前列表的 `parentId` + 是否有子计算 `canMergePredict`。  
   - 读路径：`entity.Event` 带字段即可覆盖 history EventOptions / device internal options。

4. **缓存**  
   - 沿用写后全量重建 EventOptions；不为 merge 单独加 Redis 读缓存。

5. **API 版本**  
   - 既有 admin/options 加可选字段，不新开 v2。

## Risks / Trade-offs

- [脏配置：删光子后仍为 1] → DeleteEvent 后清父标志 + Hub 展示仍以库为准；Flutter 对无子根 merge=1 应当 no-op。  
- [Hub 与服务端规则漂移] → 服务端强制 0 为权威；Hub 仅 UX。  
- [旧 EventOptions 缺字段] → 缺省当 0；下次事件写触发重建。  
- [contracts/clients 签名扩参] → 与 isAppointment 同模式；漏改编译期可见。

## Migration Plan

1. 部署 **device-service**（ensure 加列；admin 读写）。  
2. 部署 **gateway-app-server**（Hub 静态页）。  
3. Flutter 另行发版消费 `mergePredict`。  
4. 回滚：忽略新列/隐藏 Hub 列即可；列可保留。

## Open Questions

- 无（「无子的根不许勾」已确认；删光子清零采用 DeleteEvent 后清父）。
