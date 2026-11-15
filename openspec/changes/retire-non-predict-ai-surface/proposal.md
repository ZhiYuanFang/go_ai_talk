## Why

产品要聚焦喂养预测：预测推送已全面开放，商业侧只需 **预测 Pro 开通壳**（支付/试用/邀请），Pro 实际能力另案补充。仓库里仍残留 VIP、值得留意、成长轨迹、胖宝诊疗、润笔、AI 额度/选模、sim 机器人农场等杂项，增加维护成本并干扰产品叙事。本变更以 **C 档硬删代码** 收敛表面，只保留预测、Pro 壳、UCG 社区本体、喂养语音理解（Python 定死 LLM）与小智链路。

## What Changes

- **BREAKING**：删除 VIP 全链路（表访问/支付履约分流/Admin/App/internal OR 判定）。
- **BREAKING**：删除值得留意（care-alert）全栈（device App API、cash 开通/eligibility/access、voice 服务、Python 调用、Hub/额度/lane）。
- **BREAKING**：删除 7 日成长轨迹（growth-trajectory）全栈（device API、cash access/feature、voice 服务、Python turn）。
- **BREAKING**：删除胖宝诊疗 clinic（WS、clinic_*、clinic_ai 额度、LaneClinic、Python `/v1/clinic/stream` 调用）。
- **BREAKING**：删除聊天侧「成长建议」生成路径（含对 Python `/v1/clinic` 同步调用）。
- **BREAKING**：删除 voice 域 AI 额度（`ai_quota_*`）、`lane_entitlement`、care/growth 日限读 quota；删除 `llm_lane_config`、Hub「AI 模型与并发」、请求体携带 provider/model。
- **BREAKING**：喂养语音理解改为 Go **直调** Python `POST /v1/analyze/intent[/stream]`，**不传** `model`、不查 VIP/额度；Python 侧定死 LLM（验收：省略 `model` 仍可用）。
- **BREAKING**：删除 UCG 润笔（`POST .../posts/polish`）及 UCG polish 额度/选模；**保留** UCG 社区本体（发帖/Feed/社交/私信/资料/Green/Admin 审帖）。
- **BREAKING**：删除 `sim-user-service` 整进程（compose/cmd/config/Hub/gateway 反代）。
- **BREAKING**：删除 device `sim` internal API 与 `sim_user.go`；删除 ucg `posts/sample`、`chat/sim-unread-sample`、`chat/send`（仅 sim）；删除 gateway `IsSimulatedWx` usage/clientusage 过滤。
- **BREAKING**：`wx.is_simulated` 列 DROP（建议先清理 `is_simulated=1` 行）；batch/DTO 去掉 `isSimulated`；Admin wx/list 去掉按该列过滤。
- **BREAKING**：退役 `aimodel` 多 lane 中台与 Hub AI 模型页（随 clinic/care/polish/sim/understanding 选模一并消失）。
- **保留（Non-Goals）**：预测 Pro 支付/试用/邀请/`predict_pro_entry` 壳；广告清理与 catalog `status`；喂养预测 API/推送/LA；小智 MCP 全链路；**尤其保留** `POST /ucg/internal/api/media/upload-video`（小智引导片）。

## Capabilities

### New Capabilities

- `retire-vip-care-growth`：硬删 VIP、值得留意、7 日成长轨迹的商业与调用面，且不得误伤预测 Pro 壳。
- `retire-clinic-growth-suggest`：硬删胖宝 clinic 与聊天成长建议。
- `voice-understanding-python-fixed`：喂养理解去 Go 选模/额度，直调 Python intent。
- `retire-ai-quota-llm-aimodel`：硬删 voice/UCG AI 额度、llm lane 配置与 aimodel/Hub 选模中台。
- `retire-ucg-polish`：硬删润笔 API 与 polish 专用配置，保留社区本体。
- `retire-sim-user-surface`：硬删 sim-user 整服务、device/ucg sim-only API、`wx.is_simulated`、gateway 模拟号 usage 跳过；不得误伤小智。

### Modified Capabilities

- （无独立活跃基线 capability 目录；版本归档规格中 VIP/care/clinic/sim/额度等条款由本变更增量规格在下一版归档时整体取代。实施以本 change 的 specs 为准。）

## Impact

- **进程**：`cash-service`、`voice-service`、`ucg-service`、`device-service`、`gateway-app-server`；**停止并移除** `sim-user-service`。
- **Python**：Go 侧断 care-alert/clinic/growth-trajectory/成长建议调用；intent 须支持无 `model`。Python 端点物理删除可同发或后续跟进（本变更以 Go 断调用为门禁）。
- **Compose/env/ACR**：去掉 sim 服务与 `SIM_*` / `SIM_SERVICE_BASE_URL`；gateway Hub 静态页清理。
- **库表**：可不迁历史 VIP/feature/quota 表数据；`wx.is_simulated` MUST DROP；Ensure 不再 ADD 该列。
- **后续**：另开提案为预测 Pro 挂实际能力；本变更 Pro 仅开通壳。
