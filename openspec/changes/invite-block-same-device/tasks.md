## 1. cash → device 取主人设备号

- [x] 1.1 在 cash 增加 `FetchDeviceNoByWxID`（或等价）：调用 `GET .../device/app/api/user/internal/device-no-by-wx-id`，密钥头与现有 device 客户端一致；失败返回 error

## 2. Redeem 同设备拒绝

- [x] 2.1 在 `RedeemInviteCode`：非自用校验通过后拉取主人 `device_no`；与兑换者 `deviceNo` 均非空且相等则拒绝；主人空则跳过同设备规则；拉取失败 fail-closed
- [x] 2.2 更新 `RedeemInviteCode` 中文注释：补充同宝宝（同 device_no）不可兑

## 3. 校验

- [ ] 3.1 手工：同设备两账号互兑应失败；不同设备应成功；主人未绑机可兑（若可构造）
