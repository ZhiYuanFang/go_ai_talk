## ADDED Requirements

### Requirement: predict_pro_idea tables SHALL persist ideas and likes with baby name snapshot

The device-service MUST create and maintain via idempotent EnsureSchema in `ai_voice_device`（test: `ai_voice_device_test`）:

1. `predict_pro_idea` with at least: `id` (PK), `wx_id` (NOT NULL, indexed), `device_no` (NOT NULL), `baby_name` (snapshot at submit), `content` (NOT NULL), `display_text` (NOT NULL), `created_at`, `updated_at`, `deleted_at` (nullable soft-delete); optional denormalized `like_count`.
2. `predict_pro_idea_like` with at least: `idea_id`, `wx_id`, `created_at`, and a UNIQUE constraint on (`idea_id`, `wx_id`).

These tables MUST NOT reuse `feedback`. 设备库 **必须** 幂等维护想法表与点赞表；提交时 **必须** 快照宝宝名；**不得** 复用 `feedback`。

#### Scenario: EnsureSchema creates tables

- **WHEN** device-service 执行 EnsureSchema 且表不存在
- **THEN** 系统 SHALL 创建上述两表且可重复执行无错误

#### Scenario: Soft-delete keeps row

- **WHEN** Admin 删除一条想法
- **THEN** 系统 SHALL 设置 `deleted_at`（或等价软删）且 App list MUST 随后不可见该行
