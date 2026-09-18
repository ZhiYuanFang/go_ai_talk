## 1. 保存只更新

- [x] 1.1 `AdminUpsertFeatureProduct`：空编码或编码不存在则拒绝，不再生成或插入 `fp_` 编码
- [x] 1.2 已有编码只更新非主键字段

## 2. Hub

- [x] 2.1 打开功能时按「非 `fp_` 上架 → 上架 → 第一条」填入套餐；没有套餐则编码留空
- [x] 2.2 去掉「新建套餐」按钮与清空编码逻辑，文案改为只改当前编码

## 3. 自检

- [x] 3.1 确认未改 App 建单字段、未改 `maintenance_skip`、未批量改已有 `fp_` 行
- [x] 3.2 `openspec validate feature-sku-update-only --strict` 通过
