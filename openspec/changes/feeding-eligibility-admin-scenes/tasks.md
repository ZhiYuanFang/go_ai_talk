## 1. Schema 与核心算法（cash-service）

- [x] 1.1 EnsureSchema：`feeding_eligibility_scene` 表；种子 `ucg_entry`(7,10)、`care_alert_entry`(2,10)；重复启动不覆盖已改值
- [x] 1.2 在 **cash** 抽取独立有效日/连续 streak 合成函数；按场景读配置；history 窗口=`requiredDays`；确认 device/voice/history **无**资格合成副本
- [x] 1.3 Redis：资格缓存键含 scene（或版本）；配置变更失效；platform `cachekit` builder + 中文注释
- [x] 1.4 重构既有 UCG eligibility 走场景配置与新算法（去掉硬编码 7/10 与固定 14）

## 2. App / Admin API（cash-service）

- [x] 2.1 新增 `GET /cash/app/api/care-alert/eligibility`（字段同构 UCG，cash 控制器）；只信内部设备头；登录绑机
- [x] 2.2 `api/v1` g.Meta；确认 gateway-app 反代覆盖；不加入 Bearer 匿名豁免
- [x] 2.3 **向负责人确认** care-alert eligibility 是否 usage skip；确认后再改 `maintenance_skip.go`（未确认不改）
- [x] 2.4 Admin GET/POST 场景列表与更新（仅已有 scene_key）；静态页 + Hub 模块登记；写后失效缓存

## 3. Flutter 客户端

- [x] 3.1 新增 care-alert eligibility repository/provider（调 **cash**）；替代 `careAlertDailyFetchGate` 的「昨日有发生」前提
- [x] 3.2 未合格：仍展示值得留意卡片 + 剩余/所需有效日文案；不调生成/日列表接口
- [x] 3.3 合格：按原逻辑请求值得留意并展示；资格失败 fail-closed
- [x] 3.4 UCG 文案继续跟服务端 `requiredDays`（无需硬编码 7）

## 4. 验收

- [x] 4.1 自检：UCG 与值得留意资格均由 cash 返回；窗口=配置天数；Admin 改后生效
- [x] 4.2 自检：客户端未合格不打生成、合格才打；无新 `*_test.go`
- [x] 4.3 若无新 DB 连接则备注 runbook/资格约定；有则核 `.env.example`/compose
