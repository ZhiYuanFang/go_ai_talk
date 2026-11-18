## 1. 数据模型与 Ensure

- [x] 1.1 `device-service` 增加幂等 Ensure：`wx.device_model` 可空字符串列（含中文注释）
- [x] 1.2 更新 `dao`/`entity`/`do` 的 Wx 列与结构体字段 `DeviceModel` / `deviceModel`

## 2. App 写入 API

- [x] 2.1 新增 `PUT /device/app/api/user/device-model`（api + controller + service）：校验长度/空白、值未变跳过写、须登录
- [x] 2.2 确认 gateway `/device/app/api/*` 反代覆盖该路径；**不要**加入 Bearer 白名单
- [ ] 2.3 **与负责人确认是否计入 App API usage**；若「不统计」则将精确 `PUT /device/app/api/user/device-model` 写入 `maintenance_skip.go` 并在 PR 注明；若「统计」则不写 denylist

## 3. Admin / usage 读路径

- [x] 3.1 `wx/list` 查询 Fields 与 `DeviceAdminWxListItem` 增加 `deviceModel`
- [x] 3.2 gateway `usagestats` enrich / usage wx-list、client-usage wx-list item 透传 `deviceModel`
- [x] 3.3 `api-usage-stats.html`、`client-usage-stats.html` 按用户表增加「型号」列（空显示 —）

## 4. 文档与验收

- [x] 4.1 更新 `docs/runbooks/client-feature-usage-stats.md`（或等价）说明型号来源与可空
- [x] 4.2 自检：跨服务无直查；无新增 Redis 键；接口版本未改既有 login/report body
- [x] 4.3 对照 Flutter 仓 `report-device-model-to-server` 字段名 `deviceModel` 一致
