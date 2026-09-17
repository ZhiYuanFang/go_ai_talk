# 预测临近离线提醒 — 客户端（Flutter 孪生仓）任务

本仓已提供服务端能力；以下须在 `flutter_ai_talk`（或等价 App）完成，方可端到端生效。

## API

- **同步待办**：`PUT /device/api/predict/imminent/pending`（经 gateway-app，需登录 Bearer）
  - Body：`{ "deviceNo": "...", "events": [ { "eventId": 1, "nextAt": 1710000000, "title": "换尿布" } ] }`
  - `nextAt`：unix 秒；空 `events` 清空服务端闹钟
  - **`title`**：父事件展示名；**须非空**，否则服务端到点校验通过后仍**不发**可见推送（见 `predict-imminent-push-copy`）
  - **时机**：每次本地预测更新后立即调用（打开 App / 预测刷新均可）
  - 多人同宝宝：最后一次写入覆盖（产品接受）

## 推送注册

- 离线系统通知复用全局推送 token：`POST /app/api/push/register`（`channel`=`apns`|`hms`|`mipush`，`token`，`deviceKey`）
- 未注册则服务端扇出时该账号静默跳过
- 通知 `bizType`/`data` 可能含 `deviceNo`、`eventId`、`nextAt`；深链页由产品定
- 可见正文模板：`{宝宝昵称}要{title}了`（昵称空则「宝宝」）；通知标题仍为「胖宝」

## 验收建议

1. 上报后约 `nextAt-5min` 收到可见推送（多绑定账号均应尝试），正文形如「小宝要换尿布了」
2. 再次上报改 `nextAt` 后，旧时间点不应再推
3. 已写入同类 history（30 分钟内）不应再推
4. `title` 为空的条目不应出现可见推送
