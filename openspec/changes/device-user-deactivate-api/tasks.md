## 1. API 合约与路由入口

- [x] 1.1 在 `api/v1/device_app_user_http.go` 新增 `DeviceUserDeactivateReq/Res`，路径为 `POST /device/app/api/user/deactivate`
- [x] 1.2 在 `internal/controller/device_app_user.go` 新增 `Deactivate` 方法，复用 `X-Internal-Wx-Id` 解析并完成参数校验

## 2. 设备域注销业务实现

- [x] 2.1 在 `internal/services/device/wx.go` 新增按 `wxId` 删除单条 `wx` 记录的方法
- [x] 2.2 在删除成功后执行 `wxId` 相关缓存失效，避免陈旧映射读取
- [x] 2.3 统一错误语义：参数非法、记录不存在、删除失败

## 3. 联调与回归检查

- [x] 3.1 本地验证三类场景：成功注销、请求头非法、重复注销（记录不存在）
- [x] 3.2 检查现有 `/device/app/api/user/login`、`/detail`、`/bindwx` 路径行为未被破坏
- [x] 3.3 更新必要文档说明（接口用途、请求头要求、错误语义）
