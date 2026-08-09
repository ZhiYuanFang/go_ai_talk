## Context

当前选模与计次分散在三处：

1. **月度额度**（`voice_ai` / `clinic_ai` / `polish`）：用尽后 `Degraded=true`，业务用硬编码 `Degraded*Profile` 智谱种子模继续跑。
2. **care-alert**：仅 `cash.RemoteIsVipByWxID` → DeepSeek / Zhipu，无独立额度、借 clinic lane 闸门字段。
3. **硬件通道**：`/voice/chat/ws`、MCP→`DelegateTextChat(wxId=0)` 未纳入 VIP/额度统一语义。

约束：voice 不直查 cash 库（经 `RemoteIsVipByWxID`）；不新增 Redis 读缓存键缓存 VIP（沿用现 care-alert 决策：低频读、失败降级）；Admin 已有 `ai-model-admin.html` + `llm-lanes`；tip 今日复用 `LaneClinic`。

## Goals / Non-Goals

**Goals:**

- 单一权益公式：`isPremium = VIP ∨ HardwarePrivilege ∨ quota.Allowed`。
- VIP / 硬件：**不计次**；额度 API 对 VIP 呈现有额度。
- 非 premium：走 lane **free** 配置；free 为空则 Python 请求 **omit model**。
- care-alert：独立 `care_alert` feature + `careAlert` lane；非 VIP 成功计次；选模走统一权益。
- polish / tip：同一套策略；tip 挂靠 `clinic_ai` + `clinic` lane。
- Sim 四条不出现 free 配置。
- 原子公共 API，禁止各调用点各自拼 VIP/额度/模型。

**Non-Goals:**

- 不删除额度体系，不改现金支付/VIP 开通写路径。
- 不新增 tip 独立 feature/lane（默认挂 clinic）。
- 不强制本变更完成 Flutter UI 全面改版（API 语义先对齐；跨仓可跟）。
- 不引入 VIP Redis 缓存键（除非后续负责人确认）。
- 不改变 care-alert 日缓存键（仍 `deviceNo + 日`）与「触发者权益」race 接受语义。

## Decisions

### D1：公共权益与选模出口（voice / ucg 可复用契约形状）

- **选择**：在 voice（及 ucg 对称薄封装）提供：
  - `ResolveLaneEntitlement(ctx, wxID, feature, privilege) → {Premium, Snapshot}`
  - `ResolveLaneModel(ctx, wxID, lane, feature, privilege) → *ModelCfg`（nil=omit）
- **privilege**：`Account` | `Hardware`；仅硬件入口（chat WS、MCP/internal text chat）置 `Hardware`。
- **VIP**：`cash.RemoteIsVipByWxID`；失败 → 非 VIP + Warning（与现 care-alert 一致）。
- **VIP 时 Snapshot**：`Allowed=true`，`Degraded=false`（展示用 used/limit 可仍返回真实计数或占位，但不得挡主路径）；**consume 直接 no-op**。
- **备选**：各业务复制 if VIP… → 拒绝（无法保证原子性）。

### D2：硬件特权打标

- **选择**：入口显式写入 ctx（如 `WithModelPrivilegeHardware`），公共 API 只读标记。
- 覆盖：`voice_ws`（`/voice/chat/ws`）、`mcpbridge`→`DelegateTextChat` / voice internal text chat（无 Bearer wxId 的设备通道）。
- **不覆盖**：clinic WS、care-alert、tip、polish（账号通道）；即使 wxId=0 也按 Account（非 premium，除非另有额度——wxId=0 无额度则走 free/omit）。
- **备选**：仅用 wxId=0 推断硬件 → 与「未登录放行非 VIP」冲突，否决。

### D3：care-alert 独立额度

- **feature**：`care_alert`（contracts 新增）。
- **存储**：与 voice AI 额度同族（全局默认 + per-wxId override + Redis 月用量键）；Admin/App API 扩展字段（`careAlert`）。
- **计次**：生成成功且 **非 VIP** 时 `Consume(care_alert)`；VIP / 硬件不适用（care-alert 无硬件特权）。
- **选模**：`ResolveLaneModel(..., LaneCareAlert, care_alert, Account)`；废除 DeepSeek/Zhipu 硬切。
- **登录**：仍要求 `wxId>0`（不引入纯设备 care-alert）。

### D4：lane 与 free 配置

- 新增 `LaneCareAlert = "careAlert"`；种子行 + `EnsureLLMLaneDefaultRows` + Admin GET/PUT。
- `llm_lane_config`（或等价）增加 `free_provider` / `free_model`（可空字符串）；API JSON：`freeProvider` / `freeModel`。
- 适用 free 的 lane：`voiceUnderstanding`、`clinic`、`careAlert`、`polish`。
- **Sim**：API/UI 不接受、不返回 free 字段（或忽略）。
- premium → 正式 provider/model + `Acquire(lane)`；非 premium → 若 free 非空则传 free（仍建议走同一 lane 闸门或文档约定是否 Acquire——**默认仍 Acquire 该 lane**，避免免费流量打爆上游）；free 空 → omit，**可不 Acquire 上游 LLM 槽**（Python 自管）——实现取：**omit 时仍可跳过 Go 侧按正式模的 key 池，或对 free 配置单独池；首期：有 free 则按 free 的 provider+model 进现有共享池；omit 则不 Acquire aimodel 池**。

### D5：tip / polish

- tip：`clinic_ai` + `clinic` lane（含 free）；与诊疗共用月次数。
- polish：`polish` feature + `polish` lane + free；VIP 不计次；非 VIP 成功计次；额尽走 free/omit（替换 `DegradedPolishProfile`）。

### D6：Python 请求体

- `PythonModelCfg` / `model` / `model_cfg` 改为指针或 `omitempty`，nil/空不出现在 JSON。
- Feedback 类无 model，不动。

### D7：App 额度 API

- VIP 用户：各 feature `allowed=true`（且不宜再逼 40302）；可增加 `isVip` 顶栏字段（可选，便于 Flutter）；本变更至少保证 allowed/degraded 与选模一致。
- 非 VIP：保持 used/limit/degraded 语义，但 degraded 选模来源改为 Admin free。

## Risks / Trade-offs

- [VIP 查询延迟] 每请求多一次 cash HTTP → 沿用短超时；失败当非 VIP；不新增 Redis VIP 缓存（除非后续确认）。
- [care-alert 日缓存 race] VIP/有额与无额抢同一日缓存 → 仍接受「触发者权益」。
- [tip 挂 clinic] tip 与诊疗抢同一月次数 → 接受；若运营要拆开另开变更。
- [omit 与闸门] Python 自管免费模时 Go 不 Acquire → 上游并发可能升高 → 运维应用 free 非空守住闸门，或后续加全局免费池。
- [Flutter 滞后] API 已 VIP=有额度但 UI 仍只展示次数 → 跨仓跟进；Go 不阻塞。
- [Admin API 扩字段] 旧客户端 PUT 无 free → 视为空 free（Python 自选），需文档说明。

## Migration Plan

1. 部署 voice：DDL/EnsureSchema 加 free 列与 careAlert 行、`care_alert` 额度默认种子；发布选模/权益代码。
2. 部署 ucg：polish 走统一权益 + free。
3. 部署 gateway-app 静态 `ai-model-admin.html`。
4. 运维在 Admin 为各 lane 填正式模与（可选）free；care-alert 旧 DeepSeek 硬切消失后，正式模由 Admin 配置。
5. 回滚：回退服务版本；free 列可保留；care_alert 用量键可保留无害。

## Open Questions

- App `ai-quota` 是否新增显式 `isVip` 字段（非阻塞；缺省可不加，仅靠 allowed）。
- omit 路径是否要在日志中统一打 `premium=false freeEmpty=true` 便于对账（建议做，非产品决策）。
