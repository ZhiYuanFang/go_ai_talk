## 1. 宝宝画像加载（voice-service）

- [x] 1.1 新增 `internal/services/voice/clinic_profile.go`：定义 `clinicBabyProfile` struct 与 `loadClinicBabyProfile(ctx, deviceNo)`，经 `DeviceProfile().GetProfile` 读取画像
- [x] 1.2 实现 `birthday`/`gender` 格式化（与 `VoiceService.loadDeviceProfile` 口径一致：`未设置` / `女` / `男`）
- [x] 1.3 实现 `age_months` 整月差计算；`birthday=0` 或拉取失败时 `age_months=0`
- [x] 1.4 `GetProfile` 失败时记录 warning 日志并降级，不返回 error

## 2. LLM prompt 注入

- [x] 2.1 扩展 `streamClinicLLM` 签名，接受 `clinicBabyProfile`，在 system prompt 中于喂养摘要 JSON 前追加 `宝宝信息（JSON）：{"birthday":"…","gender":"…","age_months":N}`
- [x] 2.2 在 `HandleQuestion` 中 `ensureClinicSummary` 之后调用 `loadClinicBabyProfile` 并传入 `streamClinicLLM`

## 3. 配置与文案

- [x] 3.1 更新 `manifest/config/config.voice-service.yaml` 的 `aiClinic.systemPrompt`：说明结合宝宝年龄、性别与喂养摘要；出生日期未设置时勿臆测月龄

## 4. 隐私政策

- [x] 4.1 更新 `resource/public/privacy-policy.html` 胖宝诊疗章节：补充读取宝宝出生日期与性别用于适龄建议；更新生效日期

## 5. 校验

- [x] 5.1 运行 `openspec validate clinic-inject-baby-profile --strict` 并通过
- [x] 5.2 本地编译 voice-service 确认无 import 边界违规（clinic 包不得新增 `hello/internal/dao` user 依赖）
