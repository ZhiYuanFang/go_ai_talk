# 预测临近离线提醒 — 客户端（Flutter 孪生仓）任务

本仓已提供服务端能力；以下须在 `flutter_ai_talk`（或等价 App）完成，方可端到端生效。

## API

- **同步待办**：`PUT /device/api/predict/imminent/pending`（经 gateway-app，需登录 Bearer）
  - Body：`{ "deviceNo": "...", "events": [ { "eventId": 1, "nextAt": 1710000000, "title": "可选" } ] }`
  - `nextAt`：unix 秒；空 `events` 清空服务端闹钟
  - **时机**：每次本地预测更新后立即调用（打开 App / 预测刷新均可）
  - 多人同宝宝：最后一次写入覆盖（产品接受）

## 推送注册

- 离线系统通知复用 UCG 推送 token：`POST /ucg/app/api/push/register`（`channel`=`apns`|`hms`|`mipush`，`token`，`deviceKey`）
- 未注册则服务端扇出时该账号静默跳过
- 通知 `bizType`/`data` 可能含 `deviceNo`、`eventId`、`nextAt`；深链页由产品定

## 验收建议

1. 上报后约 `nextAt-5min` 收到可见推送（多绑定账号均应尝试）
2. 再次上报改 `nextAt` 后，旧时间点不应再推
3. 已写入同类 history（30 分钟内）不应再推
