# wx 用户名密码体系迁移说明

## 目标

- 账号主键统一使用 `wx.id`
- 在 `wx` 表启用 `user_name` + `password`（bcrypt 哈希）
- 保持 `unionid` 可为空，且微信绑定保持 1:1

## 前置检查

1. 确认数据库为 `ai_voice_device`
2. 确认 `wx` 表已存在：`id`、`device_no`、`unionid`、`platform`
3. 检查是否已有重复用户名（若已人工写入）
4. 检查 `unionid` 唯一索引策略允许 `NULL`

## DDL 执行

执行 [hack/ddl_wx_username_auth.sql](../../hack/ddl_wx_username_auth.sql) 中脚本。

## 约束语义

- `user_name`：唯一，服务端统一 `trim + lowercase` 后入库
- `password`：仅存 bcrypt 哈希，禁止明文
- `unionid`：允许为空；非空时全局唯一
- 一个微信（同一 `unionid`）只能绑定一个 `wx.id`

## 发布顺序建议

1. 先上线 device-service（用户名业务接口、哈希校验）
2. 再上线 gateway-app-server（用户名聚合登录）
3. 最后上线历史页昵称展示

## 回滚

- 可下线用户名入口接口并回滚服务代码
- 不建议删除 `user_name/password` 列，避免丢失已注册账号数据
