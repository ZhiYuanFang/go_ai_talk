## 1. API 合约与路由入口

- [x] 1.1 在 `api/v1/device_app_user_http.go` 新增 `DeviceUserProfileReq/Res`，路径为 `GET /device/app/api/user/profile`；响应字段 `isWxBound`、`account`（`omitempty`）、`deviceNo`
- [x] 1.2 在 `internal/controller/device_app_user.go` 新增 `Profile` 方法，复用 `wxIDFromAppUserHeader` 解析 `X-Internal-Wx-Id`

## 2. 设备域 profile 读实现

- [x] 2.1 在 `internal/services/device/wx.go` 新增 `WxUserProfileByWxID`（或等价函数）：单次 `wxRowByWxID` 派生 `isWxBound`、`account`、`deviceNo`
- [x] 2.2 统一错误语义：头非法返回参数错误；wx 行不存在返回 404（对齐 deactivate）

## 3. 联调与回归检查

- [x] 3.1 验证场景：纯微信、纯用户名、两者兼有、未绑设备、wx 行不存在、头非法
- [x] 3.2 确认 `account` 为空时 JSON 省略该字段；`deviceNo` 未绑定时为 `""`
- [x] 3.3 确认现有 `/device/app/api/user/detail` 与其他 user 路径行为未被破坏
