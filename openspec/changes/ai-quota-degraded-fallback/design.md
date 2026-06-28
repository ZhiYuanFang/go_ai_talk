## Context

当前实现（v2.0.5 基线）：

- **润笔**：`PostsPolish` → `CheckPolishAIQuota` → `allowed=false` 时返回 **40302**（`internal/controller/ucg_app_api.go`）；成功路径 `PolishPostText` 经 `aimodel.LanePolish` + `ConsumePolishAIQuota`。
- **胖宝**：`HandleQuestion` → `CheckClinicAIQuota` → 40302 阻断（`internal/services/voice/clinic_service.go`）；成功 `answer_done` 后 `ConsumeClinicAIQuota`。
- **喂养**：`guardVoiceAIQuota` 额度用尽仍 40302，**本变更不改**。
- **模型**：`aimodel.DefaultSeedProfile` 已定义 LanePolish=`glm-4.6v-flash`、LaneClinic=`glm-4.1v-thinking-flash`（`internal/services/aimodel/config_default.go`）；Admin 可 override lane profile，degraded 路径 MUST 忽略 override，固定使用上述免费种子。

约束：跨服务边界不变；不新增 Redis 键；不新增后台循环任务；40302 业务码保留给 voice_ai 及非 degraded 场景。

## Goals / Non-Goals

**Goals:**

- 润笔/clinic 在 `used >= limit` 时自动走 degraded LLM 路径，用户可继续使用。
- Degraded 路径 **不** INCR 月度用量 Redis 键。
- 额度读 API 返回 `degraded: true` 供 Flutter 展示降速文案。
- 润笔 HTTP 响应可选 `quotaDegraded: true`；Flutter 可选 toast。
- voice_ai 行为与 40302 弹框 **完全不变**。

**Non-Goals:**

- 不改变 Admin 配额配置、默认 limit、override 语义。
- 不引入 degraded 专用 lane 枚举或新 Redis 闸门键（复用现有 `LanePolish` / `LaneClinic` 闸门，profile 来源不同）。
- 不修改隐私政策（模型仍为智谱 flash，与 llm-lane-gate 种子一致）。
- 不在 degraded 路径 consume 或「恢复额度」。

## Decisions

### 1. Degraded 判定与分支位置

**决策**：在 service 层以 `snap.Allowed == false`（即 `used >= limit`）判定 degraded；controller/WS handler 不再对 polish/clinic **仅额度**场景返回 40302。

**润笔流程**（`PostsPolish`）：

```
wxId 校验 → CheckPolishAIQuota
  ├─ allowed=true  → PolishPostText(profile=LoadProfile(LanePolish)) → ConsumePolishAIQuota → { polishedText, quotaDegraded:false }
  └─ allowed=false → PolishPostTextDegraded(profile=DefaultSeedProfile(LanePolish)) → 不 consume → { polishedText, quotaDegraded:true }
```

**胖宝流程**（`HandleQuestion`）：

```
限流 check → CheckClinicAIQuota 得 snap
  ├─ allowed=true  → 现有 clinic lane + answer_done 后 ConsumeClinicAIQuota
  └─ allowed=false → degraded clinic lane（DefaultSeedProfile）+ answer_done 后 **不** ConsumeClinicAIQuota
```

**备选**：单独 `LanePolishDegraded` 枚举 —— 拒绝，增加 Admin/闸门配置面，无必要。

### 2. Degraded Profile 来源

**决策**：degraded 路径 MUST 使用 `aimodel.DefaultSeedProfile(lane)`，确保 provider=智谱、model 分别为 `glm-4.6v-flash` / `glm-4.1v-thinking-flash`，**不**读取 DB Admin override。

正常路径继续使用 `aimodel.LoadProfile(lane)`（可被 Admin 配置为其他 provider/model）。

**实现**：`PolishPostText` / `streamClinicLLM` 增加可选 `profile aimodel.Profile` 参数，或新增 `PolishPostTextWithProfile` /  clinic 侧传入 profile；degraded 调用方传入 `DefaultSeedProfile`。

### 3. 额度快照 `degraded` 字段

**决策**：扩展 `contracts.AIQuotaSnapshot`（或 App DTO）增加 `Degraded bool`：

- `Degraded = !Allowed`（即 `used >= limit`）
- `Used` / `Limit` 仍反映真实月度计数，**不**因 degraded 伪造 remaining

App 读 API：

- `GET /ucg/app/api/ai-quota` → `{ polish: { used, limit, degraded } }`
- `GET /voice/app/api/ai-quota` → `{ voiceAi: {...}, clinicAi: {..., degraded} }`（voiceAi.degraded 仅 informational；用尽仍 40302）

**备选**：单独 `remaining` 字段 —— 拒绝，Flutter 已有 `limit - used` 计算。

### 4. 40302 保留范围

| 场景 | 40302 |
|------|-------|
| voice_ai 额度用尽 | **是** |
| polish 仅额度用尽 | **否**（degraded） |
| clinic 仅额度用尽 | **否**（degraded） |
| polish/clinic 未登录 40301 | 不变 |
| 50301 队列满 | 不变 |
| clinic 42901 限流 | 不变 |

Race：degraded 请求并发时 `ConsumePolishAIQuota` 仅在 allowed 路径调用，degraded 路径无 consume，无竞态 INCR。

### 5. Flutter 展示规则

**决策**（`app-ai-quota-degraded-ui` spec）：

| feature | remaining=0 & degraded | 40302 弹框 |
|---------|------------------------|------------|
| polish | 「本月润笔额度已用完，已降速」 | 不再因仅额度用尽触发 |
| clinicAi | 「本月胖宝诊疗额度已用完，已降速」 | 同上 |
| voiceAi | 保持「剩余 0 次」或现有 hint | **仍** 40302「本月额度已用完」 |

润笔成功且 `quotaDegraded=true`：compose 页 MAY 展示一次性 snackbar/toast「已切换至降速模式」。

## Risks / Trade-offs

- **[Risk] Degraded 与正常路径同 lane 闸门，高负载时 degraded 用户仍可能 50301** → 与正常路径一致，可接受；不在本变更引入独立闸门。
- **[Risk] Admin 配置 premium 模型后，额度内用户体验与 degraded 免费模型质量差** → 产品预期：降速=免费 flash；文档在 proposal 已说明。
- **[Risk] 旧 App 在额度用尽仍期待 40302** → BREAKING 标注；需 Flutter 同步发版。
- **[Risk] `degraded=true` 时用户误以为额度未用尽** → UI 强制「已降速」文案，不用仅「剩余 0 次」。

## Migration Plan

1. 先后端发版（degraded 路径 + API 字段）；旧客户端在 polish/clinic 额度用尽时将获得成功响应而非 40302（体验改善，无数据迁移）。
2. Flutter 发版展示 degraded 文案；未升级客户端不再弹 40302（polish/clinic），功能仍可用。
3. 回滚：恢复 controller/clinic 对 `!allowed` 返回 40302；移除 degraded 分支即可，无 schema 变更。

## Open Questions

无——需求已由负责人全部拍板。
