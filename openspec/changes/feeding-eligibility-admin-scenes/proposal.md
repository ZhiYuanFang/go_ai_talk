## Why

UCG 入场资格的「连续有效喂养日」阈值（7 天 / 日 ≥10 条）硬编码在 cash，且多拉 14 天统计窗口；「值得留意」仅靠客户端「上海昨日有发生」闸门，与喂养有效日语义不一致、无法运维调参。需要把有效日判定抽成独立能力，按场景配置阈值，并为值得留意补齐与 UCG 同构的服务端资格与客户端进度展示。

## What Changes

- **宿主强制**：UCG 入场与值得留意的**喂养有效日 / 连续天数资格判断** MUST 全部实现在 **`cash-service`**（库 `ai_voice_cash`）；MUST NOT 放入 device / voice / history / Flutter 本地重算权威资格。history 仅提供按日条数契约；客户端只消费 cash 资格 API。
- **独立有效日逻辑**（cash 内）：上海日历日、按 history 日条数判定是否有效日、从今日起向前连续 streak；与具体场景解耦。
- **场景配置（Admin 可改）**：至少 `ucg_entry`、`care_alert_entry` 两场景，各自配置 `requiredDays`、`minRecordsPerDay`（日条数门槛互不共用）；种子默认 UCG=7/10、值得留意=2/10（条数可按种子调整，以 Admin 为准）。
- **取数窗口**：history 按日统计窗口 MUST 等于该次计算所需天数（单场景=`requiredDays`；若一次算多场景则 `max(requiredDays)`），MUST NOT 无必要多拉（废除固定 14）。
- **UCG**：继续 `GET /cash/app/api/ucg/eligibility`（cash），阈值改读场景配置；行为与现网一致，仅去硬编码。
- **值得留意资格**：新增 App eligibility API（如 `GET /cash/app/api/care-alert/eligibility`，**cash 暴露**）；**替代**客户端「昨日有发生」闸门。
- **客户端（Flutter）**：先请求 cash 值得留意资格；未合格仍展示「值得留意」卡片并展示还需累计有效天数；合格后再请求原值得留意生成/日列表（device）并按原逻辑展示。UCG 继续跟服务端 `requiredDays` 文案。
- **Admin**：Hub 可编辑两场景阈值（cash Admin API）；写后失效资格 Redis 缓存。
- **usage**：资格类 GET 拟与 UCG 相同不计入；实现前 **向负责人确认** 后再改 `maintenance_skip`（未确认不得假定）。
- VIP / 功能开通 MUST NOT 短路喂养资格。

## Capabilities

### New Capabilities

- `feeding-effective-day-core`：独立有效喂养日与连续 streak（**仅 cash-service**）；按场景阈值合成资格；取数窗口等于所需天数。
- `feeding-eligibility-admin`：场景阈值 Admin 读写与缓存失效（cash）。
- `care-alert-feeding-eligibility`：值得留意喂养资格 App API（cash）+ 客户端先资格后生成的展示流（替代昨日发生闸门）。

### Modified Capabilities

- （无）`ucg-entry-eligibility` 尚未归档入 `openspec/specs/`；UCG 读配置与窗口收敛以本变更 capabilities 为准，并与既有 commercial 变更对齐。

## Impact

- **进程**：资格判定与配置 **仅** `cash-service`；`history-service` 仍仅按日计数契约；`gateway-app`（反代/静态页/usage skip 待确认）；Flutter 只调 cash 资格 API + 合格后调 device care-alert。
- **库**：仅 `ai_voice_cash` 新增场景配置表；EnsureSchema 种子。
- **非目标**：在 device/voice/history 内复制资格算法；改 history 表结构；VIP 短路；值得留意 LLM/额度语义；新建测试文件。
