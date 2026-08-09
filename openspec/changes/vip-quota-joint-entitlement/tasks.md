## 1. 合约与数据模型

- [x] 1.1 新增 `aimodel.LaneCareAlert`；`contracts.AIQuotaFeature` 增加 `care_alert`；扩展额度 DTO/App/Admin 字段（含 careAlert）
- [x] 1.2 `llm_lane_config`（voice）与 polish 存储增加 `free_provider`/`free_model`（EnsureSchema/DDL）；careAlert 种子行；Sim 不增加 free
- [x] 1.3 扩展 `VoiceAdminLLMLaneItem` / ucg polish Admin item：`freeProvider`/`freeModel`；GET/PUT `/voice/admin/api/llm-lanes` 含 `careAlert`
- [x] 1.4 Python 请求体：`model`/`model_cfg` 支持 omit（指针或 omitempty）；详细中文注释

## 2. 公共权益与选模出口

- [x] 2.1 实现特权 ctx（Account/Hardware）与入口打标：`/voice/chat/ws`、MCP/internal text chat
- [x] 2.2 实现 `ResolveLaneEntitlement`：VIP（cash Remote）∪ 硬件 ∪ quota.Allowed；VIP 时 Snapshot Allowed=true；失败降级 Warning
- [x] 2.3 实现 `ResolveLaneModel`：premium→正式模；否则 free；free 空→nil；禁止业务私自拼模
- [x] 2.4 VIP/硬件路径 consume 改为 no-op；非 VIP 成功仍计次

## 3. 业务接线（voice）

- [x] 3.1 喂养意图/流式/成长建议/历史问答：经公共出口选模；去掉 `loadVoiceUnderstandingProfile` degraded 硬编码依赖
- [x] 3.2 clinic 流式：经公共出口 + `clinic_ai`；额尽走 clinic free/omit
- [x] 3.3 TipStream：挂 `clinic_ai` + `clinic` lane，同一套权益/选模/计次
- [x] 3.4 care-alert：强制 wxId；`care_alert` check；`careAlert` lane Acquire；成功非 VIP consume；删除 DeepSeek/Zhipu 硬切
- [x] 3.5 App/Admin `ai-quota` 暴露 careAlert；VIP 用户 allowed=true

## 4. 业务接线（ucg）

- [x] 4.1 polish：VIP∪额度；成功 VIP 不计次；额尽走 polish free；停用 `DegradedPolishProfile` 真相源
- [x] 4.2 ucg App/Admin ai-quota 与 VIP 语义对齐

## 5. Admin UI 与文档

- [x] 5.1 更新 `ai-model-admin.html`：Voice 三条 + careAlert；业务 lane/polish 的 free 编辑；Sim 无 free；保存联调
- [x] 5.2 更新 `llm-care-alert-daily/CONTRACT.md`、相关 runbook（VIP∪额度、careAlert lane、非 VIP 计次）；必要时 `vip-commercial-config.md` 简述
- [x] 5.3 评审：无跨库直查 cash；Redis 仅经 cachekit；无新 `*_test.go`；gateway-app 静态页 usage 排除仍成立
- [x] 5.4 更新 `voice-admin.html`：全局默认与用户列表暴露 `careAlertMonthlyLimit` / `careAlert`（已用+上限）；PUT 与 API 三字段对齐

## 6. 验收

- [ ] 6.1 VIP：各 AI 路径正式模、不计次、ai-quota allowed=true
- [ ] 6.2 非 VIP 有额：正式模并计次；额尽：free 或 omit（Python 自选）
- [ ] 6.3 硬件 chat WS / MCP：正式模、不计次；care-alert/clinic/tip/polish 无硬件特权泄漏
- [ ] 6.4 care-alert：有额非 VIP=正式模+计次；VIP=正式模不计次；Admin 可改 careAlert lane/free 与 care_alert 月度额度
- [ ] 6.5 tip 与 clinic 共享 clinic_ai；polish 行为与策略一致
