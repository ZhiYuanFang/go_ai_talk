## Context

`api/v1/device_clinic_feedback_http.go` 与 `DeviceClinicFeedbackController` 已实现 clinic/tip 反馈，内部经 `PythonAIClient.ClinicFeedback` / `TipFeedback`（JSON Body `answer_id` + `feedback`）调用 Python。但 `register_voice_service.go` 未 Bind，gateway `voice_route_proxy.go` 未反代 `/device/api/clinic|tip/*`，Flutter tip 仍提交 `tipId`/`feedbackResult`，clinic 屏有 `answerId` 但无反馈 UI。

包 B `wire-tip-generate-on-voice` 负责 tip generate 宿主与 SSE/answerId；本包只闭环反馈飞轮。tip 宿主 = voice 已锁定（generate 属 B）。

## Goals / Non-Goals

**Goals:**
1. voice-service Bind 反馈控制器，对外可调两 feedback POST
2. gateway 反代 `/device/api/clinic/*`、`/device/api/tip/*` → VOICE
3. usage：两 feedback POST 写入 `maintenance_skip.go` 精确排除（不统计）
4. Flutter tip 字段/URL 对齐；clinic 补反馈 UI + HTTP
5. 若 B 未完成 tip SSE answerId，补最小接线或明确 blocker，仍完成 clinic + Go Bind

**Non-Goals:**
- tip generate Bind、`/device/tip/*` 反代、SSE 方言整包改造（包 B）
- Python Body 改造（已完成）
- 改 `/device/app/api/feedback/*`（用户建议反馈，device 域）
- 新增测试文件

## Decisions

### 决策1：tip 宿主 = voice（generate 属包 B；本包只做 feedback）
- **方案**：feedback 与 tip generate 同宿主 voice-service；本包只 Bind `DeviceClinicFeedbackController`，不 Bind `TipCtrl`
- **理由**：Python AI 与 clinic/tip 流式同属 voice 进程；路径 `/device/api/*` 与 generate `/device/tip/*` 分前缀，可独立接线
- **备选**：feedback 挂 device-service — 排除，会跨域调 Python 且违背 tip 宿主 = voice

### 决策2：本轮做 clinic + tip 反馈飞轮
- **方案**：Go Bind + gateway 反代 + Flutter tip/clinic 两侧一并闭环
- **理由**：API/控制器已齐，缺的是接线与客户端；拆包只会延长不可用窗口
- **备选**：只做 Go — 排除，客户端仍无法提交

### 决策3：usage — tip generate 统计（B）；feedback 不统计
- **方案**：在 `maintenance_skip.go` 的 `maintenanceExactAPI` 增加：
  - `POST /device/api/clinic/feedback`
  - `POST /device/api/tip/feedback`
- **理由**：负责人已拍板 feedback 为维护型交互、不计入 App API 使用统计；generate 属包 B 且统计
- **备选**：feedback 也统计 — 排除，与拍板不符

### 决策4：feedback 路径保持 `/device/api/clinic/feedback` 与 `/device/api/tip/feedback`
- **方案**：不改 `g.Meta` path；gateway 用前缀 `/device/api/clinic/*`、`/device/api/tip/*` 反代到 VOICE
- **理由**：API 已登记；与 `/device/app/api/feedback/*`（device）及 `/device/tip/generate`（B）前缀分离，避免误路由
- **备选**：统一到 `/device/tip/feedback` — 排除，破坏已有契约

### 决策5：SSE 方言不在本包改
- **方案**：本包不改 tip generate SSE 帧格式；Flutter tip 反馈仅依赖 done 侧 `answerId` 字符串
- **理由**：方言对齐属包 B；本包最多补「从 done 写入 answerId」的最小接线
- **备选**：本包重做 SSE 解析 — 排除，与包边界冲突

### 决策6：gateway 反代挂在 voice_route_proxy，非 device_route_proxy
- **方案**：在 `installVoiceProxyMiddleware` 增加 `/device/api/clinic/*`、`/device/api/tip/*`
- **理由**：目标上游是 VOICE；挂 device 会误进 DEVICE_API_PROXY
- **备选**：device 中间件特例转发 voice — 排除，配置与观测更乱

### 决策7：包 B 依赖策略
- **方案**：apply 时检查 `wire-tip-generate-on-voice` tasks/代码；若 tip SSE/answerId 未就绪，则：
  1. 尽量在 tip 客户端补最小 `answerId` 解析（不 Bind TipCtrl、不改 `/device/tip/*` 反代）；或
  2. 报告 tip 反馈 blocker，仍完成 clinic 反馈 + Go Bind + maintenance_skip
- **理由**：避免重复整包 B，同时不让 clinic/Go 被 tip 阻塞

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 包 B 未完成 → tip 无 answerId，反馈无法提交 | 最小接线或报告 blocker；clinic/Go 仍交付 |
| `/device/api/*` 与 `/device/app/api/*` 混淆 | design/tasks 写明前缀；反代只挂 api/clinic|tip |
| feedback 误计入 usage | 精确 METHOD+path 写入 maintenance_skip，proposal/tasks 记录拍板 |
| Flutter tip 仍用旧 tipId 字段 | models/provider/repository 一并改为 answerId + feedback |

## Migration Plan

1. 先上 Go Bind + gateway 反代 + maintenance_skip（无 Flutter 亦可 curl 验证）
2. 再发 Flutter tip/clinic 反馈改动
3. 回滚：去掉 Bind/反代/skip 项即可；API 定义可保留

## Open Questions

- （无）锁定决策 1–5 已由产品/负责人确认
