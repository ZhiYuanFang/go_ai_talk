## Context

`RedeemInviteCode` 已禁同 `wxId` 自用，未比对码主人与兑换者 `device_no`。同宝宝多账号可互兑刷预测条数与获客原力。主人设备权威在 device `wx`；cash 不得直查 device 库，可复用 `GET /device/app/api/user/internal/device-no-by-wx-id`（网关已用）。

## Goals / Non-Goals

**Goals:**

- 同 `device_no`（双方均非空且相等）拒绝兑码。
- device 查询失败 fail-closed；主人未绑机（空 device_no）放行。
- 所有 `invite_code` 功能开通均适用。

**Non-Goals:**

- 恢复「一家锁定」；限制互兑（不同设备仍可 A↔B）。
- 改原力 / 去重键；直查 device DB；新建测试文件。

## Decisions

### D1 — 比对时机

在确认码有效且非自用之后、写 grant 之前：用 `owner_wx_id` 拉 `ownerDeviceNo`，与入参 `deviceNo`（兑换者，网关内部头）比较。

### D2 — device 契约

cash 新增薄封装（如 `FetchDeviceNoByWxID`），`DEVICE_SERVICE_URL` + `X-Device-Gateway-Internal-Secret`，路径与 gatewayapp 一致。禁止 import device DAO。

### D3 — 边界

| 条件 | 行为 |
|------|------|
| `ownerDeviceNo == redeemerDeviceNo` 且均非空 | 拒绝 |
| `ownerDeviceNo == ""` | 放行 |
| HTTP/业务失败 | 拒绝兑码 + Warning 日志 |

**否决**：查失败放行（可被刷）；主人未绑也拒绝（误伤未完成绑机的正常邀请）。

### D4 — 文案

错误码 `CodeInvalidParameter`（或等价），文案示例：「不可使用同一宝宝下其他账号的邀请码」。

## Risks / Trade-offs

- [device 抖动导致兑码失败] → fail-closed 取舍；依赖现网 device 可用性。
- [换机后仍同机判定] → 以**当前**绑定为准，符合「同一宝宝」。
- [历史已同机互兑记录] → 不追溯撤销。

## Migration Plan

1. 发 cash（含 device 客户端调用）；device 无需发版（契约已存在）。
2. 回滚：去掉同设备比对即可。

## Open Questions

- （无）边界已按探索结论锁定。
