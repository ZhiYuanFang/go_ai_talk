## Context

现网：`CountNonLeafEvents` = `COUNT(DISTINCT parent_id) WHERE parent_id > 0`，经 cash 写入 `prediction_unlock.totalActivatableCount`。一级无子根（如洗澡）不计入，导致天花板小于产品期望的一级类目数。

本变更将可激活计数改为**一级根**（`parent_id = 0`），与预测槽位「一级事项」对齐；不改 voice「非叶子追问」语义。

## Goals / Non-Goals

**Goals:**

- device 计数 = 事件表 `parent_id = 0` 的行数（含无子根）。
- cash catalog `totalActivatableCount` 使用该数。
- 规格 / 注释与 voice「非叶子」术语脱钩。

**Non-Goals:**

- 改 voice 追问或 history 写叶子规则。
- 改 `allowedCount`、VIP、邀请原力。
- 强制同步改 Flutter（须在客户端仓对齐；本仓仅文档提示）。
- 新建测试文件。

## Decisions

### D1 — 口径 A：一级根

```sql
SELECT COUNT(*) FROM event WHERE parent_id = 0
```

（或 GoFrame 等价 `Where(parent_id, 0).Count()`；注意库内根是否用 `0` 而非 `NULL`——与现网 `normalizeEventParentID` 一致用 0。）

**否决**：非叶子∪叶子根（方案 B）——中间层父节点会抬高总数，易与一级锁 UI 不对齐。

### D2 — 路径名

保留 `GET /device/internal/api/event/non-leaf-count` 与 cash `FetchNonLeafEventCount` 符号，**只改 count 语义**与中文注释 / OpenAPI summary，避免无谓改调用方。长期可另开 rename。

函数可重命名为 `CountRootEvents` 并让旧名委托，或就地改实现与注释。

### D3 — 客户端

预测「已全部激活」与锁行顺序 MUST 以一级根列表为准；若 App 仍用「有子节点的父」列表，会出现 total 与行数不一致。实现任务含文档提醒；代码在 `flutter_ai_talk`。

## Risks / Trade-offs

- [BREAKING 数字变大] → 已开通条数未变，仅天花板升高；用户更易显示「未全部激活」。
- [术语残留 non-leaf] → 注释标明「历史路径名，语义=一级根」。
- [根用 NULL] → 实现前确认列默认；按现网 0 处理。

## Migration Plan

1. 发 device（计数实现）→ cash 无需改调用路径即可吃到新数（若仅改 device）。
2. 回滚：恢复 DISTINCT parent_id 查询。
3. App：确认一级根列表与 total 同源后发版（可并行）。

## Open Questions

- （无）口径已确认为方案 A。
