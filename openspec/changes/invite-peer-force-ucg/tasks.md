## 1. device：非叶子计数与去原力

- [x] 1.1 实现事件字典非叶子计数（有子节点的事件数）及 internal HTTP 契约
- [x] 1.2 注册成功路径触发 cash EnsureInviteCode（或文档化仅懒创建并在 tasks 验收）
- [x] 1.3 移除 `wx.force_value` 读写、`UcgWxIncrementForceValue` 及 Batch/Validate 中的 force 字段与内部 increment 路由

## 2. ucg：原力本域

- [x] 2.1 schema：`ucg_user_force` + `ucg_force_ledger`（含中文注释）
- [x] 2.2 辩论自投改为本域 +1 + ledger；profile/feed enrich 改读本域
- [x] 2.3 internal：获客加分 +100 + ledger（供 cash）
- [x] 2.4 App：force ledger（及 summary/profile 字段对齐）；向负责人确认 usage 后再改 skip 列表

## 3. cash：邀请身份与兑码

- [x] 3.1 EnsureInviteCode + App `invite/mine`、`invite/invitees`（昵称经契约）
- [x] 3.2 重写 Redeem：人×码×功能；去一家锁/码级范围有效期；预测 +1；成功后调 ucg 获客加分；获客 `redeemed_count++`
- [x] 3.3 删除邀请 Admin 页、Hub 模块、静态路由、exempt 与造码 CRUD API

## 4. cash：预测 +1 与 catalog 天花板

- [x] 4.1 支付/广告预测路径强制 grantQuantity=1 的 allowed_count_delta；移除邀请 T1 全开调用
- [x] 4.2 catalog 调用 device 非叶子计数，写入 `totalActivatableCount`；VIP 不改写预测 allowedCount 为 -1
- [x] 4.3 种子/校验：预测商品 grantQuantity=1；清理无用 invite 子表/bind 路径（未发布可删）

## 5. 联调与文档

- [x] 5.1 gateway 放行新 App/internal 路径；确认 Bearer/usage
- [x] 5.2 更新 runbook（去掉 wx force DDL 依赖说明）
- [x] 5.3 `openspec validate invite-peer-force-ucg --strict`
