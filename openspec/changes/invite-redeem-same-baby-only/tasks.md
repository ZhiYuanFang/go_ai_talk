## 1. RedeemInviteCode 闸替换

- [x] 1.1 删除 `RedeemInviteCode` 中两处 `owner_wx_id == redeemer_wx_id` 自用拒绝及「不可使用自己的邀请码」文案
- [x] 1.2 在码有效后、人×码×功能去重前：经 `deviceclient.FetchDeviceNoByWxID(owner)` 取主人当前 `device_no`；失败 fail-closed；主人与兑换者 `device_no` 均非空且相等则拒绝，文案「不可使用同一宝宝下其他账号的邀请码」
- [x] 1.3 peek 与 TX 内锁码后均执行同宝宝闸（对齐原自用双检）；更新文件头注释（去掉自用、标明同宝宝）

## 2. Admin 文案

- [x] 2.1 更新 `feature_admin_rules.go` 的 `inviteCommonRulePrefix`：去掉「不可使用自己的邀请码；」，保留同宝宝句

## 3. 验收与边界

- [x] 3.1 确认 cash **不**直查 device 库；经 `clients/device`；无新 App 路由 / Redis 读缓存 / 背景 ticker
- [x] 3.2 场景自检：同机互兑拒；异机好友过；同人异机兑自己的码过；主人未绑机不过同宝宝闸；device 失败拒兑
