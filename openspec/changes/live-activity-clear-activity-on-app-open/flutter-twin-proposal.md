# Flutter twin proposal: live-activity-clear-activity-on-app-open

> **已写入兄弟仓**：`flutter_ai_talk/openspec/changes/live-activity-clear-activity-on-app-open/`（proposal/design/specs/tasks）。本文档为镜像摘要，以 Flutter 仓制品为准。

## Why

Go 将移除 Live Activity update 的 ContentState.`ack`、原生免鉴权 `POST /app/api/push/live-activity/ack`，以及 `LIVE_ACTIVITY_ACK_TIMEOUT_MS` 超时 failover。用户划掉锁屏后，服务端可能仍持有 `token_kind=activity`，继续发 update 进空气。产品接受空窗；改为 App 进入前台且本机已无对应 Live Activity 时，由 Flutter 带鉴权告知服务端删除该机 activity token，使下次推送走 `start`。

## What Changes

### 删除（与 Go 同发）

- 移除 iOS 原生 / MethodChannel 在收到 update 后回传 ack 的逻辑。
- ContentState 模型去掉 `ack` 字段（可继续忽略未知字段，但不再解析/上报）。
- 删除对 `POST /app/api/push/live-activity/ack` 的调用与相关 native 代码。

### 新增

- 调用 Go：`POST /app/api/push/live-activity/activity/clear`
  - **必须带 Bearer**（登录用户）
  - body：`{ "deviceKey": "<本机 deviceKey>" }`
  - 成功 / 无行：均视为成功（幂等）

### 触发时机

App 冷启动完成登录态可用后，以及从后台回前台时：

1. 用 ActivityKit 查询本 App 是否仍有目标 Attributes 的 Live Activity。
2. **若不存在** → 调 clear。
3. **若仍存在** → 不调 clear（避免误删导致下次多余 start）。

### 边界

- 未登录：不调；登录后再跑一次检查。
- 仅清本机 `deviceKey`，不影响其他设备。
- clear **不会**立刻起新 LA；等服务端下次预测 upsert/到点推送走 start。
- `push_to_start` / `activity` 的 register 路径保持不变（含原生免鉴权 register）。

### 非目标

- 不在划掉瞬间（App 未开）做后台 HTTP。
- 不重新实现 update ack。
- 不改锁屏 UI / items 契约（除去掉 `ack`）。

## Twin

- Go：`go_ai_talk` / `live-activity-clear-activity-on-app-open`
- 须与 Go 同发：先或同发 Go（ack 下线 + clear 上线），Flutter 去掉 ack 并接 clear。

## Test plan

- [ ] 有 LA 时打开 App → 不调 clear（本地仍有 activity）
- [ ] 划掉 LA 后打开 App → 调 clear → 服务端无该机 activity 行 → 下次推送为 start
- [ ] 未登录打开 → 不 401 刷屏；登录后再 clear
- [ ] 多机：A 机 clear 不影响 B 机 activity token
- [ ] ContentState 无 `ack` 字段时锁屏仍正常
