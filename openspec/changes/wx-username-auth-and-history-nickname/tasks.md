## 1. 数据模型与安全基线

- [x] 1.1 校验并固化 `ai_voice_device.wx` 列约束：`user_name` 唯一、`unionid` 可空且唯一语义不变，补充迁移脚本/说明
- [x] 1.2 引入并封装密码哈希能力（bcrypt），统一注册、登录、改密的哈希/比对入口
- [x] 1.3 定义用户名规范化规则（trim/大小写策略/长度）与统一错误语义常量

## 2. device-service 用户名能力

- [x] 2.1 在 `api/v1/device_app_user_http.go` 新增用户名相关请求/响应模型（注册、登录、绑定微信、绑定设备、改密、微信下创建用户名）
- [x] 2.2 在 `internal/controller/device_app_user.go` 实现上述接口的参数校验、错误码映射与头部主体解析
- [x] 2.3 在 `internal/services/device/wx.go` 实现账号状态流转与冲突判定（用户名唯一、微信唯一绑定、已绑定不可重复绑定）
- [x] 2.4 实现用户名账号绑定设备号逻辑并复用设备注册校验（不跨域直查）
- [x] 2.5 对写路径补齐缓存失效（`wxId->unionid`、`wxId->deviceNo`）与关键日志

## 3. gateway-app 聚合登录扩展

- [x] 3.1 在 `api/v1/gateway_app_http.go` 新增用户名聚合登录契约
- [x] 3.2 在 `internal/controller/gateway_app_ctrl.go` 增加用户名聚合登录转发与 token 签发流程
- [x] 3.3 更新白名单与鉴权说明，确保用户名聚合登录可匿名访问且不影响既有微信/设备号登录

## 4. 历史画像昵称扩展

- [x] 4.1 扩展 `device/history` 画像读取契约返回 `nickname`（经 device 契约获取，不新增跨域直查）
- [x] 4.2 更新 `api/v1/device_history_http.go` 与 `internal/controller/device_history.go` 响应结构，透传 `nickname`
- [x] 4.3 更新 `resource/public/history.html`，在性别旁展示昵称并处理空值占位

## 5. 规格与联调收口

- [x] 5.1 修订受影响 OpenSpec 文档一致性（本变更下 specs 与现有能力语义对齐）
- [x] 5.2 执行接口联调清单：6 类用户名接口 + 微信兼容登录 + 历史页昵称展示
- [x] 5.3 运行 `openspec validate wx-username-auth-and-history-nickname --strict` 并修复校验问题
