## Context

胖宝诊疗（`/voice/clinic/ws`）在 `HandleQuestion` 中已通过 `ensureClinicSummary` 注入近 7 天喂养事件聚合 JSON，经 `streamClinicLLM` 拼入 system prompt 后调用 `deepseek-v4-pro`。语音球路径已有 `loadDeviceProfile`（`DeviceProfile().GetProfile` HTTP 契约），clinic 尚未使用画像字段。

约束（仓库 AGENTS.md / 服务边界）：

- voice-service MUST 经 `DeviceProfile()` HTTP 访问 device 域，禁止进程内直连 `dao.User`。
- 不改动 WS 帧协议、Flutter、summary Redis 懒刷新逻辑。
- 不引入新的 Redis 读缓存（profile 每轮 HTTP 拉取）。
- 不新增测试文件（当前阶段约定）。

## Goals / Non-Goals

**Goals:**

- 每轮 `question` 前以 auth 锁定的 `deviceNo` 拉取宝宝 `birthday`、`sex`，计算 `age_months`，以 **A2 单行 JSON** 注入 system prompt（位于 7 天喂养摘要 JSON 之前）。
- `DeviceProfile` 失败时降级为 `birthday="未设置"`、`gender="女"`、`age_months=0`，记录 warning，继续 LLM。
- 与语音球 `loadDeviceProfile` 降级语义一致。
- 更新 `aiClinic.systemPrompt` 与隐私政策披露。

**Non-Goals:**

- 不改为原始 history 行 dump（仍用聚合摘要）。
- 不将 profile 写入 `voice:clinic:summary:*` 缓存。
- 不包含 `babyName`（本期仅 birthday/sex/age_months）。
- 不改 gateway、额度、限流、session 语义。

## Decisions

### 1. Prompt 形态：A2 单行 JSON（用户选定）

在 system prompt 中于喂养摘要前追加：

```
宝宝信息（JSON）：{"birthday":"2024-03-15","gender":"男","age_months":15}
```

**理由**：与语音球 `用户宝宝信息={"birthday":"…","gender":"…"}` 风格一致；改动面小于统一 snapshot struct（方案 B）。

**备选**：A1 自然语言列表 — 未选，用户指定 A2。

### 2. 集成位置：`HandleQuestion` + `streamClinicLLM`

```
HandleQuestion
  ensureClinicSummary(...)
  profile := loadClinicBabyProfile(ctx, deviceNo)   // 新增
  streamClinicLLM(ctx, profile, summary, question, prior, ...)
```

`streamClinicLLM` 签名增加 `profile clinicBabyProfile` 参数（或等价 struct），在现有 `system += 摘要` 之前插入宝宝 JSON 行。

**理由**：profile 与 summary 生命周期不同（summary 有 Redis watermark；profile 每轮 fresh），不宜合并进 summary 缓存。

### 3. Profile 加载：包级函数，复用 `DeviceProfile()` 契约

新增 `clinic_profile.go`：

- `clinicBabyProfile` struct：`Birthday string`, `Gender string`, `AgeMonths int`
- `loadClinicBabyProfile(ctx, deviceNo)` → 调 `DeviceProfile().GetProfile`
- 格式化逻辑独立于 `VoiceService.loadDeviceProfile`，避免 clinic 依赖 VoiceService 单例；字段映射与语音球一致：
  - `birthday`：`Birthday>0` → `2006-01-02`（本地时区），否则 `"未设置"`
  - `gender`：`Sex>0` → `"男"`，否则 `"女"`
  - `age_months`：`Birthday>0` 时整月差（见下），否则 `0`

### 4. age_months 计算

- 基准：`time.Now()` 与 `time.Unix(birthday, 0)` 的**整月**差（非四舍五入天数）。
- 算法：年差×12 + 月差，若当前日 < 生日日则减 1（与常见「月龄」口径一致）。
- `birthday=0` 或拉取失败：`age_months=0`，JSON 仍输出该字段，配合 systemPrompt 引导模型勿臆测。

### 5. 失败语义：降级继续（用户选定）

| 条件 | 行为 |
|------|------|
| `GetProfile` HTTP/解析错误 | Warning 日志；profile 降级；**继续 =** 500 error 帧 |
| `birthday=0`（未设置） | `birthday="未设置"`, `age_months=0` |
| `deviceNo` 空 | 不应发生（auth 已校验）；若出现则同降级 |

与 history 摘要失败（返回 500「喂养摘要加载失败」）**独立**：仅 profile 失败不阻断问诊。

### 6. systemPrompt 配置微调

`manifest/config/config.voice-service.yaml` `aiClinic.systemPrompt` 首句改为提及宝宝年龄、性别；增加一行：出生日期未设置时勿臆测月龄，可温和建议用户完善宝宝信息。

### 7. 日志可观测

profile 降级时：`glog.Warningf(ctx, "clinic baby profile degraded: deviceNo=%s err=%v", …)`，便于排查 device-service 不可用。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 每轮多一次 DeviceProfile HTTP | 与语音球一致；GetProfile 轻量；不缓存以避免画像滞后 |
| 降级默认值 `gender="女"` 可能误导 | 与语音球一致；systemPrompt 要求未设置时勿臆测；长期可改 `gender="未知"`（本期不引入新口径） |
| 隐私披露滞后 | 同步更新 `privacy-policy.html` 与 `app-legal-docs` spec |
| MODIFIED spec 归档遗漏 | delta spec 使用 ADDED（新 requirement）+ MODIFIED（隐私政策完整块） |

## Migration Plan

1. 部署 voice-service（向后兼容，无 WS 协议变更）。
2. 同步更新 `resource/public/privacy-policy.html`（gateway 静态资源）。
3. 无需数据迁移、Redis 键变更或 Flutter 发版。
4. 回滚： revert voice-service + 隐私 HTML 即可。

## Open Questions

（无 — A2 格式与降级语义已在提案阶段确认。）
