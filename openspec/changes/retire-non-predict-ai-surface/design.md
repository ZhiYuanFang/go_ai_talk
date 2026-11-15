## Context

产品收敛到「喂养预测 + Pro 开通壳」：推送已 ungate（`predict-pro-unlock-shell`），Pro 暂无业务挂载。仓库仍耦合 VIP OR、care-alert、growth-trajectory、clinic、润笔、voice/UCG AI 额度、`llm_lane_config` 选模、`aimodel` 多 lane 中台、`sim-user-service` 机器人农场。喂养理解实际由 Python 跑 LLM；Go 仍选模并可选 `Acquire`。小智 MCP 与 UCG 社区本体必须保留；`upload-video` 与 sim 曾共用，删 sim 时不得误删。

## Goals / Non-Goals

**Goals:**

- C 档硬删与预测无关的商业/AI/机器人表面，代码与 Hub/路由一并消失。
- 喂养理解：Go → Python intent，不传 `model`、不读额度/VIP/lane。
- 预测 Pro 支付/试用/邀请/`predict_pro_entry` 与广告清理、catalog `status` 行为保持可用。
- 删除 sim 整服务、device/ucg sim-only API、`wx.is_simulated`、gateway 模拟号 usage 跳过。
- 退役 `aimodel` 与 Hub「AI 模型与并发」。

**Non-Goals:**

- 不实现预测 Pro 的实际业务能力（另案）。
- 不删除 UCG 社区（发帖/Feed/社交/私信/资料/Green/Admin 审帖）。
- 不改动小智 MCP 开通/绑定/引导片；不删除 `POST /ucg/internal/api/media/upload-video`。
- 不强制物理 DROP VIP/feature/quota 等历史表（可停用路径；`wx.is_simulated` 例外 MUST DROP）。
- 不强制同 PR 删除 Python care-alert/clinic/growth 端点（Go 断调用为验收底线；Python 可跟进）。
- 不新增 Redis 读缓存；不新增背景 ticker。

## Decisions

### D1：单 change、按域切 tasks

**选择**：一个 OpenSpec change，tasks 按 cash / voice / ucg / sim+device / gateway+Hub / aimodel 分组。  
**备选**：拆多个 change → 并行安全但交叉依赖（理解去选模依赖额度删除）多，合并更易一次验收。  
**理由**：探索已定稿边界清晰，单 PR/单 apply 叙事完整。

### D2：理解路径去掉选模与额度

**选择**：删除 `resolveVoiceUnderstandingModel` / `ResolveLaneModel` / `lane_entitlement` / voice `ai_quota` 在理解路径上的使用；`AnalyzeIntent*` 请求体 **省略** `model`；不再 `aimodel.Acquire`（无 model 时本就可跳过）。  
**备选**：保留固定 YAML 选模仍经 Go 下发 → 与「Python 定死」产品陈述冲突。  
**理由**：用户明确 Python 定死、Go 不必传选模；额度与 VIP 一并砍干净。

### D3：成长建议与 clinic 一并删

**选择**：删除 `callDeepSeekGrowthSuggestion` 及相关死代码；删除 clinic WS/服务；不再调用 `/v1/clinic` 与 `/v1/clinic/stream`。  
**备选**：成长建议改调 intent → 产品不要该能力。  
**理由**：用户选 A；clinic 用不到。

### D4：UCG 只删润笔+额度，社区保留

**选择**：删除 polish API、UCG `ai_quota`（polish 列）、`LanePolish`、Hub AI 润笔配置；社区路由不动。  
**理由**：UCG 与 voice 额度同名表但分库（`ai_voice_ucg` / `ai_voice_voice`），可各自硬删。

### D5：sim 整服务 + 列 DROP + 先清假号

**选择**：停并移除 `sim-user-service`；删 device sim internal、`sim_user.go`；删 ucg sample/sim-unread-sample/chat/send；删 gateway `IsSimulatedWx`；Ensure 将 `wx.is_simulated` **DROP**（若存在）。部署前/启动迁移：**先删除或注销 `is_simulated=1` 行**（及 sim 库凭证），再 DROP 列，避免假号混入 Admin 列表。  
**备选**：只停服务留列 → 用户要求表字段直接删。  
**理由**：无模拟语义后列无读者。

### D6：aimodel 整包退役

**选择**：删除 `internal/services/aimodel` 及所有 ProfileStore/Invoke/Acquire 调用方（随 clinic/care/polish/sim/understanding 消失）。Hub `ai-model-admin.html` 与相关 Admin API 删除或改为空壳下线。  
**备选**：缩包只留 stub → 无调用方则整删更干净。  
**理由**：探索结论：选模中台无剩余消费者。

### D7：小智护栏

**选择**：tasks/自检显式列出 MUST KEEP：`upload-video`、`xiaozhi-mcp*`、device bindings、mcpbridge、cash 小智履约。  
**理由**：sim T4 曾用 upload-video，易误删。

### D8：对外 App 接口版本

**选择**：删除已废弃 App 路由（care-alert、growth-trajectory、polish、clinic WS、AI quota App API 等）属 **BREAKING**；不修改仍存活的 v1 结构字段语义去「悄悄改行为」。预测 Pro / catalog / eligibility 保持现有契约。  
**理由**：仓库规则「已有版本结构不可改」针对增量字段；下线整条废弃接口以删除绑定实现，而非改同 path 响应形。

## Risks / Trade-offs

- [Risk] Python intent 在无 `model` 时失败 → [Mitigation] apply 前/后用测试环境打 intent；失败则阻塞合并或临时 Python 小补丁。
- [Risk] 误删 `upload-video` 或小智路由 → [Mitigation] tasks 检查清单 + grep 小智路径。
- [Risk] 残留 `is_simulated=1` 行在 DROP 前列未清 → [Mitigation] 迁移脚本/运维步骤先 DELETE；design 强制顺序。
- [Risk] Hub/history「成长建议」`suggest` 读删残留 → [Mitigation] 生成路径必删；Hub suggest CRUD 可同 tasks 清理或标可选（历史行可留）。
- [Risk] 大 diff 难审 → [Mitigation] tasks 按域勾选；可分 commit 但仍同一 change。
- [Trade-off] 历史 VIP/quota 表不 DROP → 库有僵尸表，换代码路径干净与数据迁移成本。

## Migration Plan

1. 合并前：确认 Python intent 无 `model` 可用；备份 device `wx` 中 `is_simulated=1` 列表（可选）。
2. 部署顺序建议：先发领域服务断调用与删路由 → gateway Hub/反代 → 停 sim 容器并改 compose → device Ensure DROP 列（需已无代码读该列）。
3. 环境变量：移除 `SIM_*`、`SIM_SERVICE_BASE_URL`；gateway 重载。
4. 回滚：困难（硬删）；依赖 git revert + 重新部署旧镜像；`is_simulated` DROP 需再 ADD 才能完全回滚 sim（不预期回滚）。
5. 重部署清单：`cash-service`、`voice-service`、`ucg-service`、`device-service`、`gateway-app-server`；停止 `sim-user-service`。

## Open Questions

- Hub/history 对 `suggest` 表的运维读删是否同 PR 删除（生成已死）：默认 **同 PR 清理入口**，表可留。
- Python care-alert/clinic/growth 端点是否同仓删除：默认 **Go 先断**；Python 另提交可选。
