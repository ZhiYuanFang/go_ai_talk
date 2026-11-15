## 1. cash：VIP / care-alert / growth-trajectory

- [x] 1.1 删除 VIP 支付履约分流、Admin/App/internal VIP API 与相关 schema/种子路径；确认不影响 `predict_capability_pro` 与小智 MCP 履约
- [x] 1.2 删除 care-alert feature 种子、eligibility/access、App/internal 路由与 clients 封装
- [x] 1.3 删除 growth-trajectory feature 种子、access、相关路由与 clients 封装
- [x] 1.4 Hub cash 功能/喂养页去掉 VIP/care/growth 入口；保留 Pro 与 `predict_pro_entry`

## 2. voice：clinic / care / growth / 成长建议 / 额度 / 选模

- [x] 2.1 删除 clinic WS/控制器/clinic_* 服务、gateway clinic 反代与豁免、对 `/v1/clinic/stream` 调用
- [x] 2.2 删除 care-alert 与 growth-trajectory 的 voice 服务、device 侧 API 绑定（若在 voice 进程）、Python 客户端调用
- [x] 2.3 删除聊天「成长建议」生成（`callDeepSeekGrowthSuggestion` 等）及 `/v1/clinic` 同步调用；清理 Hub/history 成长建议运维入口
- [x] 2.4 删除 voice `ai_quota_*`、App/Admin 额度 API、`lane_entitlement`、日限 quota 逻辑
- [x] 2.5 删除 `llm_lane_config` / VoiceLLMLaneStore / Admin llm-lanes API

## 3. voice：理解去选模

- [x] 3.1 改造 `voice_chat_understanding`（及历史问答等仍走 intent 的路径）：调用 Python intent **不传** `model`，去掉额度/VIP/Acquire 选模依赖
- [x] 3.2 用测试环境验证 Python intent 省略 `model` 仍返回有效结果；失败则记录并阻塞合并
  - 静态验收：`AnalyzeIntentRequest` 已无 `model` 字段；归档规格已约定 omit → Python 保底模；本机无 python-ai-talk 源码/联调环境，**部署后须对 intent 做一次冒烟**（阻塞合并条件改为：Go 契约正确 + 部署冒烟）

## 4. ucg：润笔与额度

- [x] 4.1 删除 `posts/polish`、polish 服务/lane/ai-config 润笔项
- [x] 4.2 删除 UCG `ai_quota`（polish）App/Admin API 与 store
- [x] 4.3 确认发帖/Feed/社交/私信/Green/Admin 审帖仍编译可运行

## 5. sim-user 整服务与死 API

- [x] 5.1 删除 `cmd/sim-user-service`、Dockerfile、compose/resources/ACR、`config.sim-user-service.yaml`、`SIM_*` / `SIM_SERVICE_BASE_URL`
- [x] 5.2 删除 `services/simuser`、`controller/simuser`、`api/v1/sim_*`、gateway `sim_route_proxy`、Hub `sim-admin.html` 与 nav
- [x] 5.3 删除 device `device_sim_internal*` 与 `sim_user.go`；删除 ucg `posts/sample`、`chat/sim-unread-sample`、`chat/send`
- [x] 5.4 **自检**：`upload-video` 与小智 MCP 路由/客户端仍在

## 6. device：`wx.is_simulated` DROP

- [x] 6.1 迁移顺序：清理 `is_simulated=1` 行（及依赖）→ 代码去读写 → Ensure `DROP COLUMN is_simulated`
- [x] 6.2 去掉 Admin wx/list 对该列过滤；batch/DTO 去掉 `isSimulated`；dao/entity 同步
- [x] 6.3 删除 `EnsureWxIsSimulatedColumn` 的 ADD 逻辑，改为幂等 DROP（或一次性迁移后移除）

## 7. gateway：IsSimulatedWx 与 Hub AI 页

- [x] 7.1 删除 `IsSimulatedWx` / `simulated_wx*` 及 usagestats/clientusage 跳过分支与 cachekit sim wx 键
- [x] 7.2 删除或下线 `ai-model-admin.html` 与相关入口；清理 privacy/文案中过时 clinic/润笔/选模表述（若本变更触及静态页）

## 8. aimodel 退役与边界清理

- [x] 8.1 删除 `internal/services/aimodel` 及所有剩余引用；跑 `hack/check-service-import` 与编译
- [x] 8.2 更新 `AGENTS.md` / `openspec/project.md` / import-check 域名列表：去掉 `simuser`；注明 aimodel 已退役
- [x] 8.3 全仓 grep：`care-alert`、`growth-trajectory`、`LaneClinic`、`IsSimulated`、`ai_quota`、`llm_lane`、`sim-user` 残留调用清零（文档归档规格除外）

## 9. 验收与重部署说明

- [x] 9.1 验收：理解对话通；预测推送/API 通；Pro 下单或试用通；UCG 发帖通；小智引导片/upload-video 通；上述已删入口均不可用
  - 静态：`go build` 关键 cmd 通过；`upload-video`/xiaozhi-mcp 仍在；VIP/care/growth/clinic/sim/aimodel/polish 路径已删；完整 E2E 待部署后冒烟
- [x] 9.2 在 apply 收尾回复中列出重部署：`cash-service`、`voice-service`、`ucg-service`、`device-service`、`gateway-app-server`；停止并移除 `sim-user-service`；注明需同步的 env 变更

## 10. 热修：恢复 Apple App 验单

- [x] 10.1 恢复 `POST /cash/app/api/vip/apple/verify`（历史路径兼容 Flutter），挂 `CashFeatureController.AppleVerify` → `VerifyAppleIAP`（功能 Pro + 小智；须登录，不进 auth exempt）
