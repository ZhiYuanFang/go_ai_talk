# 客户端功能使用统计（client-usage）

与 [App API 使用统计](./vip-commercial-config.md) 并存：本能力统计 **App 主动上报的页面/功能事件**，不是 HTTP 接口调用量。

## App 上报

- `POST /device/app/api/client-usage/report`（须 Bearer 登录）
- Body：`{ "featureId": "vip_paywall_show", "description": "展示 VIP 购买页" }`
  - `featureId`：自由字符串，最长 128；**勿含 `|`**
  - `description`：运维可读说明，必填，最长 128；**勿含 `|`**
- 同账号 **3 秒**内最多成功 1 次；模拟用户跳过；**不计入** App API 使用统计
- 数据仅 Redis，约 **30 天**；时间线每用户最多 **1 万**条（成员 `unix|featureId|description`；旧数据无描述段时 Admin 显示 `-`）

### Flutter 示例

```dart
// 登录后调用；失败可丢；本地不必排队
await api.post('/device/app/api/client-usage/report', body: {
  'featureId': 'growth_trajectory_open',
  'description': '展示成长轨迹页',
});
```

## Hub

设备管理 → **客户端使用统计**（`/device/admin/client-usage-stats`）：按功能 / 按用户；列表与时间线均展示 **事件名 + 说明**。
