## 1. HMS category 映射

- [x] 1.1 在 `push_hms.go` 增加按规范化 `bizType` 返回 `WORK` / `IM` / 空（省略）的辅助逻辑，并附中文注释说明华为自分类与省略规则
- [x] 1.2 在构建 `androidNotif` 时：仅当映射非空才设置 `category`；`ucg_silent_badge`、空与未知 `bizType` 不写入该键

## 2. 校验

- [x] 2.1 确认未改动 APNs/MiPush 与 `push_by_biz` 常量；本变更仅触及 HMS 载荷
- [x] 2.2 对照 design：预测=`WORK`、UCG 可见=`IM`、静默/未知省略；必要时用本地构造 payload 核对 JSON 字段存在性（不新增 `*_test.go`）
