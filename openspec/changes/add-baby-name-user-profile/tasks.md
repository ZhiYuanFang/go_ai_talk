## 1. 接口与契约扩展

- [x] 1.1 扩展 `api/v1/device_app_user_http.go`：在画像读写与 auto_save 请求/响应中新增可选 `babyName` 字段。
- [x] 1.2 扩展 `api/v1/device_history_http.go`：在 birthday 查询与保存请求/响应中新增可选 `babyName` 字段。
- [x] 1.3 更新 `internal/services/contracts/runtime_contracts.go` 与相关调用签名，确保画像契约统一包含 `babyName`。

## 2. 后端业务链路改造

- [x] 2.1 修改 `internal/controller/device_app_user.go`：`Get/Save/AutoSave` 读写并透传 `babyName`。
- [x] 2.2 修改 `internal/services/device/admin.go`、`internal/services/device/wx.go`、`internal/services/device/profile_adapter.go`、`internal/services/device/admin_http_client.go`，完成 `user.baby_name` 持久化与远程调用透传。
- [x] 2.3 修改 `internal/controller/device_history.go` 与 `internal/services/history/{local.go,delegate_http.go,adapter.go}`，使 history 画像接口完整透传 `babyName`。

## 3. 缓存与一致性

- [x] 3.1 修改 `internal/services/device/cache_repo.go`：画像缓存结构新增 `babyName`，写后缓存刷新包含该字段。
- [x] 3.2 修改 `internal/services/history/cache_repo.go`：birthday/profile 缓存结构新增 `babyName`，兼容旧缓存反序列化与空值语义。
- [x] 3.3 校验画像写入后的缓存一致性（同请求链路读回应返回最新 `babyName`）。

## 4. 前端页面与联调验证

- [x] 4.1 修改 `resource/public/history.html`：新增“宝宝名字”输入与展示，保持与生日/性别同一保存入口。
- [x] 4.2 更新前端加载与保存逻辑：读取 `babyName` 并在保存时提交 `babyName`，处理空串与 trim 规则。
- [x] 4.3 完成端到端自测：仅改名字、同时改名字+性别+生日、旧请求不传 `babyName` 三类场景均通过。
