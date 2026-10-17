## Why

同一用户对这份问卷只能成功提交一次，交过就不再弹出。运维需要在 Hub 上把某条提交标成无效，让这个人可以再答一次，同时仍能看见他上次写了什么。

## What Changes

- `survey_response` 增加有效标记，已有行默认有效。
- 只有「存在且有效」的提交才挡住弹出。标成无效后，开关与有效期仍允许时，current 对该用户返回需要展示。
- 用户再次提交时更新这一行的选项快照、意见、宝宝昵称、设备号和时间，并把有效改回开。不新增第二行，不新开活动。
- Hub 结果表增加「有效」开关，点一下立即落库。无效行仍留在列表里。
- 改选项文案、开关、有效期仍不会让有效提交的用户再弹。
- App 路径、登录要求和「不计入使用统计」不变。Flutter 不改。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `device-satisfaction-survey`：提交是否挡住弹出改为看有效标记；Hub 可改这一标记；无效后再交卷覆盖同一行。

## Impact

- **device-service**：`survey_response` 新列；current / submit 语义；Hub 结果多一个立即保存的有效开关。
- **gateway-app**：不新增 App 路由，不改 `maintenance_skip.go`。新的 Hub 接口仍在 `/device/admin/api/survey/` 下。
- **Flutter**：无改动。
