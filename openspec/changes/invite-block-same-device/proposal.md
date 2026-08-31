## Why

同一宝宝（同一 `device_no`）下多个微信号可互兑邀请码：兑码给本机功能加码、码主人获客原力 +100，形成家庭内刷。现网只禁「自用」（同 `wxId`），挡不住同机不同账号。

## What Changes

- **BREAKING（兑码规则）**：兑换邀请码时，若码主人当前绑定的 `device_no` 与兑换者请求中的 `device_no` 相同（均非空），MUST 拒绝；适用于所有经 `invite_code` 开通的功能。
- 仍拒绝自用（同 `wxId`）；不同设备好友互兑仍允许。
- 边界：主人尚未绑机（查得空 `device_no`）→ **放行**；查询 device 契约失败 → **fail-closed 拒绝兑码**（防刷优先）。
- 不恢复已废除的「一家锁定」（不限制兑换者只能兑某一个 owner）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `invite-code-identity`：兑码增加同设备拒绝；修正「互兑」场景为双方设备号不同。

## Impact

- **cash-service**：`RedeemInviteCode`；经 `DEVICE_SERVICE_URL` 调 device `device-no-by-wx-id`（复用现有内部契约，禁止直查 device 库）。
- **device-service**：无必改（已有内部按 wxId 取 device_no）。
- **App**：错误文案需可展示（服务端返回明确错误即可）。
- **非目标**：改获客原力数额；改人×码×功能去重；新建测试文件；背景补偿任务。
