## Why

`ai_voice_device.user` 已新增 `baby_name` 字段，但当前设备画像接口、历史页面接口与网页仅覆盖生日/性别，导致宝宝名字无法读取、展示与更新。该缺口会造成前端体验与数据库能力不一致，且影响用户档案完整性。

## What Changes

- 扩展设备画像相关接口（`/device/app/api/user/get`、`/device/app/api/user/save`、`/device/app/api/user/auto_save`）支持 `babyName` 字段读写。
- 扩展历史页面使用的画像接口（`/device/history/api/birthday`、`/device/history/api/birthday/save`）返回并保存 `babyName`，与前端页面实际调用链对齐。
- 扩展设备域与历史域适配层/缓存结构，将 `babyName` 纳入统一画像缓存与远程契约透传。
- 更新网页画像区：在“性别”之外展示“宝宝名字”，并支持修改后保存。
- 兼容旧客户端：`babyName` 作为可选字段，未传入时保持空串语义，不引入破坏性变更。

## Capabilities

### New Capabilities
- （无）

### Modified Capabilities
- `device-wx-profile-apis`: 设备画像接口能力从“生日+性别”扩展为“宝宝名字+生日+性别”，并要求历史页面画像接口链路同步支持 `babyName`。

## Impact

- 受影响代码：
  - API 定义：`api/v1/device_app_user_http.go`、`api/v1/device_history_http.go`
  - 控制器：`internal/controller/device_app_user.go`、`internal/controller/device_history.go`
  - 服务与契约：`internal/services/device/*`、`internal/services/history/*`、`internal/services/contracts/runtime_contracts.go`
  - 网页：`resource/public/history.html`
- 受影响缓存：设备画像缓存与历史画像缓存 JSON 结构将新增 `babyName` 字段。
- 数据库与依赖：不新增表，不新增外部依赖；基于既有 `user.baby_name` 字段完成能力接入。
