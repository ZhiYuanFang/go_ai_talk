## 1. 依赖与权益模型

- [x] 1.1 确认 `xiaozhi-mcp-permanent-unlock` 权益表/Add 门禁已在分支可用（同 PR 或先合并）
- [x] 1.2 cash：`xiaozhi_mcp_entitlement` 增加 `expires_at`；`HasActive` 识别永久与未过期试用；支付/Admin 授写 `expires_at=0`
- [x] 1.3 cash：试用记账表（或等价）+ claim 24h（一次；失败回滚 unused）
- [x] 1.4 cash：`EnsureAccessForAdd(wx)`（有效则放行；可试用则 claim；否则拒绝）供 device 调用（App 或 internal）

## 2. App 开通态

- [x] 2.1 `GET .../xiaozhi-mcp/unlock` 增加 `trialAvailable`、`expiresAt`（语义见 spec）

## 3. device Add 与 wxId 下沉

- [x] 3.1 Add：改为调用 `EnsureAccessForAdd`（替代仅查 unlocked）；再写绑定
- [x] 3.2 Upsert 通知与内部全量 list 增加 `wxId`；clients/xiaozhimcp 与 device 内部 DTO 对齐

## 4. mcpbridge 使用闸门与懒停

- [x] 4.1 Manager/Bridge/ChatHandler 保存 `wxId`；缺省无效则 tools fail-closed
- [x] 4.2 tools/call 前 `clients/cash` 查开通；未开通拒答且不调 voice；cash 瞬时失败拒答但不拆桥
- [x] 4.3 明确未开通拒答后尽力 Remove 本桥（懒停）；绑行不删

## 5. 自检

- [x] 5.1 跑服务边界/import 检查（mcpbridge/device 不直连 cash 库）
- [ ] 5.2 手工：未试用首次 Add → 24h 开通；试用中 tools 成功；改 `expires_at` 过期 → tools 拒答且桥停、绑行仍在；再付费/手工授后可复建桥；试用用尽后 Add 失败
