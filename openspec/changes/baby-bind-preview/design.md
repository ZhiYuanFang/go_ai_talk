## Context

Flutter 扫码绑定需要在调用 `POST /device/app/api/user/bindwx` 之前展示宝宝昵称确认。`GET /device/app/api/user/get` 已能按 `deviceNo` 返回 `babyName`/`birthday`/`sex`，但：

- 语义是通用画像查询，不是「绑定前预览」；
- 会多返回生日/性别（产品确认框只要昵称）；
- 未注册设备时的错误语义未必适合扫码失败文案。

姊妹仓 `flutter_ai_talk` change `baby-share-qr-bind` 依赖本接口先部署。

## Goals / Non-Goals

**Goals:**

- 提供登录态只读预览：校验 `deviceNo` 已注册后返回 `deviceNo` + `babyName`。
- **不得** 修改 `wx.device_no` 或其它绑定副作用。
- 与 `bindwx` 相同的 App 鉴权注入（`X-Internal-Wx-Id`）。

**Non-Goals:**

- Flutter UI / 二维码载荷格式（姊妹仓）。
- 改 `bindwx`、`auto_save`、`user/get` 契约。
- 返回性别/生日/头像 URL；限流/邀请短期 token；已绑定用户换绑策略。

## Decisions

1. **路径与方法**  
   - `POST /device/app/api/user/bind_preview`  
   - Body：`{ "deviceNo": "<string>" }`（trim；空 → 参数错误）  
   - 成功 data：`{ "deviceNo": "<echo>", "babyName": "<string，可空>" }`  
   - 选用 POST（与 `bindwx` 同族、body 传 dn）而非复用 GET `user/get`。

2. **鉴权**  
   - 与 `bindwx` 一致：gateway 注入 `X-Internal-Wx-Id`；controller 解析失败则参数/未授权错误。  
   - **不得** 加入匿名 auth-exempt。  
   - 预览**不**要求调用方当前 `wx.device_no` 为空（客户端已保证未绑定才进绑定页；服务端不做换绑策略）。

3. **存在性与读档**  
   - 先 `EnsureRegistered`（或等价：无 user/device 行则业务错误「设备未注册」类），再 `DeviceProfile().GetProfile`。  
   - `babyName` 空字符串合法（客户端展示「宝宝」）。  
   - **不得** 调用 `WxBindDevice`。

4. **相对 `user/get`**  
   - 保留 `user/get` 不变；预览专用最小字段，避免客户端误把查档当确认契约。

5. **Gateway / usage**  
   - 确认 `/device/app/api/user/*` 代理已覆盖新 path；**不得** auth-exempt。  
   - **负责人确认：`POST …/bind_preview` 计入 usage**（不加入 `maintenance_skip`；与既有 `GET …/user/get` 不计入对照）。

6. **双仓顺序**  
   - 本仓 apply → 部署 `device-service`（及必要时 gateway）→ 再 Flutter apply。

## Risks / Trade-offs

- [有码即可读昵称] → 与今日手输 ID 后 bind 同级；接受；响应字段保持最小。  
- [与 `user/get` 并存] → 文档标明预览走 `bind_preview`，避免客户端混用。  
- [EnsureRegistered 文案偏「设备」] → 业务 message 可沿用现网「设备未注册」类；Flutter 映射为宝宝共享码无效（姊妹仓负责）。

## Migration Plan

- 无 DDL。  
- 回滚：下线路由或网关拒转。

## Open Questions

- （已定）路径、最小字段、无绑定副作用、Go 先于 Flutter。
