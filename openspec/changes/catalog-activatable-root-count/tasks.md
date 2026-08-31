## 1. device 一级根计数

- [x] 1.1 将 `CountNonLeafEvents` 改为统计 `parent_id = 0`（或重命名为 `CountRootEvents` 并保留旧入口委托）；更新中文注释：语义为一级根，含无子根
- [x] 1.2 更新内部 API summary/dc：标明返回一级根数量；路径可仍为 `/device/internal/api/event/non-leaf-count`

## 2. cash 注释对齐

- [x] 2.1 更新 `FetchNonLeafEventCount` / `feature_catalog` 中文注释：`totalActivatableCount` = 一级根天花板，非 voice「非叶子」

## 3. 校验与客户端提醒

- [x] 3.1 对照 DB：实现等价于 `SELECT COUNT(*) FROM event WHERE parent_id=0`；发版后在目标环境用该 SQL 与 catalog `totalActivatableCount` 对一下（含无子根如洗澡）
- [x] 3.2 **Flutter（本仓不改）**：预测锁行顺序与「全部激活」判断 MUST 按**一级根**列表，与 `totalActivatableCount` 同源；若仍按「有子节点的父」对齐会与新天花板不一致
