## Context

依赖进行中的 `xiaozhi-mcp-permanent-unlock`：`xiaozhi_mcp_entitlement`（wx）、Add 前 cash 校验、支付/手工授永久、独立 Admin。当前权益偏「永久有效/无效」；Bridge 会话仅持 `deviceNo`，`tools/call` 无权益校验。试用过期或退款后若只禁 Add，已连音箱仍可喂养。

约束：跨服务经 clients；不新增未批准 ticker（懒停借 tools 路径触发）；不默认加 Redis；试用对齐 care/growth 的 24h、每账号一次。

## Goals / Non-Goals

**Goals:**

- 每 wx 一次试用 24h；首次成功 Add 时 claim。
- tools/call 校验 wx 有效权益；失败拒答并懒停当前桥。
- 过期/未开通后绑行保留；付费后可复建桥。
- Upsert/内部 list/Bridge 携带 wxId。

**Non-Goals:**

- 不主动扫表到期拆桥（不做新 reconcile 义务；可选后续增强）。
- 不删绑行、不改开通功能管理、不引入 VIP 旁路。
- 不在本仓改 Flutter UI（契约字段即可）。
- Hub 不单独豁免试用消耗（走同一 App Add 则会 claim；运维可用手工授永久后再绑）。

## Decisions

### 1. 权益模型扩展 `expires_at`

- **选择**：`xiaozhi_mcp_entitlement` 增加 `expires_at`（0=永久；>0=试用截止 Unix 秒）。`HasActive` = status 有效且（expires_at=0 ∨ now < expires_at）。
- 支付/Admin 授：`expires_at=0`。试用 claim：`expires_at=now+24h`，`unlock_method=trial`。
- **备选**：独立试用表 → 双读复杂，否决。

### 2. 试用记账（一次）

- **选择**：专用 `xiaozhi_mcp_trial`（或等价）`wx_id` + `status unused|used`，与 feature_trial 同构但独立（不进 feature_id）。
- claim：unused→used 与写入限时权益同成功语义；失败回滚试用标记（对齐 care `revertTrialToUnused` 思路）。

### 3. 首次 Add claim 顺序

- **选择**：device Add 调用 cash「EnsureAccessForAdd」：
  - 已有效 → 放行；
  - 无效且试用可用 → cash 内 claim 24h，成功则放行；
  - 否则拒绝。
- 再写 `xiaozhi_mcp_binding`。避免「绑上了但试用没落」。
- **备选**：先写绑定再 claim → 可能留下无权益绑行，否决。

### 4. tools/call 硬闸 + 懒停

- **选择**：Bridge/ChatHandler 持有 `wxId`；`Handle` 开头 `clients/cash` 查开通；未开通 → `NewErrorCallResult`，并异步/尽力 `Manager.Remove` 本 token（懒停）。
- cash 不可达 → fail-closed（拒答；可不拆桥以免误伤，或拆桥——选 **拒答且不因瞬时错误拆桥**，仅明确未开通时拆桥）。
- **备选**：仅拆桥不拒答 → 重连窗口仍可能打到旧进程逻辑；硬闸必须在 call 上。

### 5. wxId 下沉到 Bridge

- **选择**：device 写路径 Upsert 与内部全量 list 增加 `wxId`；mcp Manager 会话存 wxId；旧客户端缺省 wxId=0 时 tools 一律拒答（fail-closed）。
- reconcile desired 带 wxId。

### 6. App unlock 字段

- **选择**：`unlocked`、`trialAvailable`、`expiresAt`（0 表示永久或未开通时省略）、可售 `product` 仍按上架返回。

### 7. 过期后复建桥

- 绑行仍在；用户重新获得有效权益后：
  - 下次写路径（改 token/再 Add 同 MAC）或
  - 运维/用户触发的刷新、或 mcp 启动 reconcile（若 list 仍含该绑且实现不按权益过滤）
- 首版：**reconcile 不按权益过滤**（避免与懒停双源）；复建依赖写路径 Upsert 或进程重启后仍会建桥——若重启后无权益仍建桥，则靠 tools 闸门拦使用，再次 call 再懒停。可接受。

## Risks / Trade-offs

- [Risk] 过期后到下次 tools/call 前桥仍显示已连接 → **Mitigation**：产品接受懒停；文案可提示「过期后首次对话将断开」。
- [Risk] Hub App 登录 Add 消耗用户试用 → **Mitigation**：文档说明；运维先手工授再绑。
- [Risk] mcp→cash 延迟增加 tools 尾延迟 → **Mitigation**：5s 级超时；失败 fail-closed。
- [Risk] 与 permanent-unlock 未合入顺序 → **Mitigation**：本变更 apply 须在权益表/Add 门禁已存在之后；tasks 标明依赖。

## Migration Plan

1. 确认 `xiaozhi-mcp-permanent-unlock` 已部署或同 PR 合并。
2. cash：加列/试用表 + claim + unlock 字段。
3. device：Add 走 EnsureAccessForAdd；Upsert/list 带 wxId。
4. xiaozhi-mcp：会话 wxId + tools 闸门 + 懒停。
5. 回滚：关掉 tools 校验会恢复白嫖；应保留校验、仅可临时放宽 cash 全开放（不推荐）。

## Open Questions

- 无（懒停、Add claim、保留绑行已拍板）。usage 统计沿用 permanent-unlock 待确认项，本变更若无新 App path 则不新增确认。
