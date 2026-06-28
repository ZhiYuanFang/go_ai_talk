## 1. 契约与 DTO（go_ai_talk）

- [x] 1.1 `contracts.AIQuotaSnapshot` 增加 `Degraded bool`；App 读 DTO（`api/v1` polish/voice quota 响应）同步 `degraded` JSON 字段
- [x] 1.2 `UcgPostsPolishRes` 增加 `QuotaDegraded bool`（`json:"quotaDegraded"`）
- [x] 1.3 `GetPolishAIQuotaAppStatus` / `GetVoiceAIQuotaAppStatus` 填充 `Degraded = !Allowed`（或 `used >= limit`）

## 2. UCG 润笔 degraded 路径（go_ai_talk）

- [x] 2.1 `compose_ai.go`：支持传入 `aimodel.Profile`（或 `PolishPostTextWithProfile`）；degraded 使用 `DefaultSeedProfile(LanePolish)`
- [x] 2.2 `PostsPolish`（`ucg_app_api.go`）：`allowed=false` 时走 degraded 路径、不返回 40302、不 `ConsumePolishAIQuota`；成功设 `QuotaDegraded=true`
- [x] 2.3 `allowed=true` 路径保持 consume；成功设 `QuotaDegraded=false` 或省略
- [x] 2.4 `AIQuotaGet` 响应返回 `polish.degraded`

## 3. Voice 胖宝 clinic degraded 路径（go_ai_talk）

- [x] 3.1 `CheckClinicAIQuota` 改为返回 snapshot（或新增 `CheckClinicAIQuotaSnap`），供 `HandleQuestion` 分支
- [x] 3.2 `clinic_service.go`：`allowed=false` 时用 `DefaultSeedProfile(LaneClinic)` 调 LLM，跳过 40302；`answer_done` 路径不 `ConsumeClinicAIQuota`
- [x] 3.3 `allowed=true` 路径保持现有 consume 语义
- [x] 3.4 voice App quota API 返回 `clinicAi.degraded`（及 `voiceAi.degraded` 若 DTO 统一）

## 4. voice_ai 不变性确认（go_ai_talk）

- [x] 4.1 确认 `guardVoiceAIQuota` / `voice_chat_understanding.go` 额度用尽仍返回 40302，无 degraded 分支
- [x] 4.2 grep 确认 polish/clinic 路径不再对**仅额度**场景映射 40302

## 5. Flutter UI（flutter_ai_talk，`d:\work\flutter_ai_talk`）

- [x] 5.1 `ai_quota_models.dart`：`AiQuotaFeatureStatus` 增加 `degraded` 字段与 fromJson
- [x] 5.2 `ai_quota_remaining_hint.dart`：`remaining=0 && degraded` 时展示润笔/胖宝降速文案
- [x] 5.3 `ucg_compose_screen.dart`：解析 `quotaDegraded`；可选 snackbar；刷新 quota provider
- [x] 5.4 `pangbao_ai_screen.dart`：确保 clinic hint 使用 degraded 状态；额度用尽不期待 40302
- [x] 5.5 确认 `ai_quota_errors.dart`：voice 40302 弹框保留；polish/clinic 不因仅额度用尽弹框

## 6. 构建与验证

- [x] 6.1 `go build ./...`（go_ai_talk 根目录）通过
- [x] 6.2 Flutter analyze/build（flutter_ai_talk）通过（若本变更一并实施）

## 7. 手工测试说明

- [ ] 7.1 **润笔额度内**：used < limit 时润笔成功、consume +1、`quotaDegraded=false`、hint 显示剩余次数
- [ ] 7.2 **润笔额度用尽**：used=limit 时润笔仍成功、`quotaDegraded=true`、used 不变、hint「本月润笔额度已用完，已降速」、无 40302
- [ ] 7.3 **胖宝额度内**：clinic 成功、`answer_done` 后 consume +1
- [ ] 7.4 **胖宝额度用尽**：仍可问答、`answer_done` 后 clinic_ai used 不变、hint 降速文案、无 WS 40302
- [ ] 7.5 **喂养 voice_ai 额度用尽**：仍 WS 40302「本月额度已用完」、App 弹框不变
- [ ] 7.6 **非额度错误**：40301/50301/42901 行为与变更前一致
- [ ] 7.7 **额度读 API**：`/ucg/app/api/ai-quota` 与 `/voice/app/api/ai-quota` 在 used=limit 时 `degraded=true`
