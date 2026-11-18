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

「按用户」表除 **平台**（`wx.platform`：ios/android/web）外另有 **型号**列（`deviceModel`）：来自客户端上报的 `wx.device_model` 原文，可空；旧客户端/Web 未上报时显示「—」。型号**不是** Redis 聚合维度。

App 上报型号（与功能事件分开）：

- `PUT /device/app/api/user/device-model`（须 Bearer）
- Body：`{ "deviceModel": "huawei" }`（可空串清空；最长 128；服务端无枚举）

Admin API（`/device/admin/api/client-usage/*`）由 **gateway-app 本机**处理，device 反代须旁路（与 `/device/admin/api/usage/*` 同）；漏旁路时会 404。
