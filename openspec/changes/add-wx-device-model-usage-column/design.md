## Context

Hub「按用户」列表里的「平台」来自 `wx.platform`（登录写入的 ios/android/web），经 `device` `GET /device/admin/api/wx/list` → gateway `usage/wx-list` / `client-usage/wx-list` → 静态页渲染。它**不是** Redis usage 计数维度。

产品要在同表再加「型号」列，展示客户端原文（huawei / xiaomi / iphone11 等），可空、无服务端枚举。

基线：`openspec/specs/v3.2.0/spec.md`（`device-admin`、`gateway-app-api-usage-stats`、`client-feature-usage-stats`）。对口 Flutter：`report-device-model-to-server`。

## Goals / Non-Goals

**Goals:**

- `wx.device_model` 可空持久化；App 可覆盖写；Admin/usage 列表与两 Hub 页展示。
- 契约字段 JSON：`deviceModel`；DB：`device_model`。
- 最少 DB 操作：仅在值变化时 `UPDATE`；Ensure 加列一次。

**Non-Goals:**

- 不做机型枚举/品牌映射/归一化白名单。
- 不把型号接入 Redis usage 聚合或交叉维度。
- 不为该字段新建 Redis 读缓存。
- 不修改既有 `client-usage/report`、登录 body（避免改既有 v1 结构）。
- 不强制 Web/旧客户端上报。

## Decisions

1. **落库位置：`wx.device_model`（可空 VARCHAR）**  
   - 与 `platform` 同为账号侧展示属性，usage 页已有透传链路。  
   - 替代：仅挂在 client-usage Redis → 否决（api-usage 页无时间线、TTL 丢数据）。  
   - 替代：仅请求头临时解析 → 否决（无「加一列」持久展示）。

2. **写入 API：新路径 `PUT /device/app/api/user/device-model`**  
   - Body：`{ "deviceModel": "<string>" }`；须 Bearer；`X-Internal-Wx-Id` 定位行。  
   - `deviceModel` trim 后长度 0～128；空串表示清空为 SQL NULL/空串（实现统一存空串即可）。  
   - 含非法控制字符则 400；不做枚举校验。  
   - 与当前值相同则跳过 UPDATE（少写）。  
   - gateway 既有 `/device/app/api/*` → device 反代即可；**不**进 Bearer 白名单。

3. **读路径：扩展既有列表，不新建 Admin API**  
   - `device` `wx/list` Fields 增加 `device_model` → JSON `deviceModel`。  
   - gateway enrich（`usagestats.DeviceWxListPage` 等）透传至 usage / client-usage wx-list item。  
   - Hub：`api-usage-stats.html`、`client-usage-stats.html` 在平台列旁加「型号」。

4. **Schema Ensure**  
   - `device-service` 启动幂等 `ALTER TABLE wx ADD COLUMN device_model ...`（列已存在则跳过）。  
   - 同步 entity/dao 列定义（手写或 regenerate 按仓库惯例）。

5. **App API usage 统计**  
   - **默认提案倾向：不统计**（遥测/档案补丁，对标 `POST /device/app/api/client-usage/report`）。  
   - **Open Question**：须负责人确认后再改 `maintenance_skip.go`；未确认不得擅自加入/移出。

6. **限流**  
   - 不做独立 Redis 限流；依赖「值未变跳过写」+ 客户端节流。若遭刷写，后续可加 SetNX（非本变更必须）。

## Risks / Trade-offs

- [Risk] 换机后旧型号被覆盖 → 接受（展示「最近一台」；与 platform 覆盖语义类似）。  
- [Risk] 客户端字符串脏乱（厂商差异） → 接受「直接展示」；运维可读即可。  
- [Risk] 未确认 usage 就 apply denylist → Mitigation：tasks 单独确认项；未答复则实现接口但 denylist 暂缓并在 PR 标明。  
- [Risk] 只上 Go 未上 Flutter → 列长期为空 → 可接受。

## Migration Plan

1. 部署 `device-service`（Ensure 加列 + API + wx/list）。  
2. 部署 `gateway-app-server`（enrich + 两 HTML；若确认不统计则带 maintenance_skip）。  
3. 发 Flutter 含上报逻辑的版本。  
4. 回滚：停上报即可；列可保留（可空无害）。

## Open Questions

1. **新接口是否计入 App API usage？**（建议：不统计）— apply 前须书面确认。  
2. 空串语义：清空 vs 忽略 — 本 design 锁定为**清空为空白**（与「可空」一致）。
