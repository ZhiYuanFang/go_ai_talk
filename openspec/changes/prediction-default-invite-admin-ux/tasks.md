## 1. Schema 与数据模型

- [x] 1.1 `EnsureSchema`：`feature_def` 增加 `default_allowed_count`（INT，默认 0）；种子 `prediction_unlock` 可设初值（缺省 0，待产品确认）
- [x] 1.2 `EnsureSchema`：`feature_allowed_count` 增加 `full_access`（TINYINT）与 `full_access_expires_at`（BIGINT）；兼容已有行
- [x] 1.3 Admin/内部读写结构体与 DAO 字段对齐上述列（含中文注释）

## 2. 预测有效条数与 T1 履约

- [x] 2.1 实现有效条数合成：无临时全开时 `allowedCount = defaultFree + permanentDelta`；有有效临时全开时 `allowedCount = -1` 并设置 `expiresAt`（有限期）
- [x] 2.2 邀请码兑换 `prediction_unlock`：按 `grant_duration_days` 写/续期临时全开；**不**递增永久 `allowed_count`；失效设备缓存
- [x] 2.3 付费/广告预测履约仍只增加永久 `allowed_count`（不碰临时全开，除非产品另有要求）
- [x] 2.4 `api/v1` catalog 类型注释标明 `-1` 哨兵语义；必要时 Admin 定义响应带上 `defaultAllowedCount`

## 3. Admin API 契约

- [x] 3.1 功能定义更新：仅允许已存在 `featureId`；拒绝新建任意 ID；支持更新 title/description/unlockMethods/status/sort/defaultAllowedCount
- [x] 3.2 功能 SKU 创建：`productCode` 空则服务端生成唯一码；所属 `featureId` 必须已存在；更新不得改码
- [x] 3.3 列表/详情返回默认开通数与只读 `featureId`/`productCode` 足够支撑 UI

## 4. Admin UI

- [x] 4.1 开通功能页：功能编号只读（编辑已有项）；展示并编辑简介、默认开通数；开通方式改为枚举多选；按钮与表单纵向对齐
- [x] 4.2 售卖套餐：所属功能下拉；商品编码新建留空/只读展示自动生成结果；按钮纵向对齐
- [x] 4.3 邀请码页：授予有效天数文案标明「预测为临时全开期限」；功能选择避免纯逗号手输（勾选/下拉）；按钮纵向对齐
- [x] 4.4 硬刷新/缓存提示：Hub 静态资源变更后需强刷（可在页眉 hint 一句）

## 5. 验收与文档

- [x] 5.1 自检：无设备行 + 默认 N → catalog N；邀请兑换后 -1 与 expiresAt；到期后回落；永久累加仍在
- [x] 5.2 自检：Admin 无法改 featureId / 无法新建未知 ID；SKU 自动编码；开通方式多选落库兼容串
- [x] 5.3 若影响库连接约定则核 `.env.example`/runbook（本变更预期仅 EnsureSchema，无新 DB 连接）；在 PR/备注写明客户端 `-1` 约定
- [x] 5.4 不新增 `*_test.go` 等测试文件
