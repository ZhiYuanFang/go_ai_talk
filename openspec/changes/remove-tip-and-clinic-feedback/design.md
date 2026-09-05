## Context

小贴士 SSE（`POST /device/tip/generate`）与 clinic/tip HTTP 飞轮曾挂在 voice-service，路径前缀为 `/device/...`，经 gateway `VOICE_API_PROXY` 反代。Flutter 已移除 tip SSE 与 clinic/tip feedback 调用；运维 `history.html` 仍可能触发 tip。care-alert 日列表与飞轮仍活跃，宿主与额度留在 voice（探索阶段曾考虑迁 device，已撤销）。

基线行为见 `openspec/specs/v3.0.0/spec.md`（tip 暴露、clinic/tip feedback Bind、usage skip）。本变更以删除死面为主，不迁域、不改库。

## Goals / Non-Goals

**Goals:**

- 从 voice 注册、service、Python 客户端、contracts、gateway 反代与运维页彻底移除 tip 生成与 clinic/tip feedback。
- 保证 care-alert 全路径（含 `POST /device/api/care-alert/feedback`）与 clinic WS 不受影响。
- 清理 `maintenance_skip` 中仅服务于已删路径的条目，避免遗留噪声。

**Non-Goals:**

- 不迁 care-alert 宿主或配额到 device。
- 不删 care-alert feedback / Python CareAlertFeedback。
- 不改 `/voice/clinic/ws`、`clinic_ai` 额度或 clinic lane。
- 不改 Flutter 仓；不新增 `*_test.go`；不新增微服务或 DB 连接。

## Decisions

### D1：物理删除而非 410 桩

- **选择**：删除控制器、Bind、API 类型、`TipStream`、Python tip/clinic feedback 方法（若仅 tip/clinic 飞轮使用）、gateway 模式。
- **理由**：无 App 调用方；运维同步删除；保留桩会继续污染 voice 杂货面。
- **备选**：返回 410 Gone — 拒绝，增加长期噪音。

### D2：gateway `/device/api/clinic/*` 整段解绑

- **选择**：当前 voice 反代中 `/device/api/clinic/*` 仅服务 clinic feedback；删除反馈后移除该 pattern。
- **约束**：MUST NOT 误删 WS 路径（clinic WS 在 `ws_route_proxy` 的 `/voice/clinic/ws` 等）。
- **备选**：保留空反代 — 拒绝。

### D3：care-alert 飞轮显式保留

- **选择**：评审与实现自检清单必须包含「care-alert feedback 仍 Bind、反代仍含 `/device/api/care-alert/*`、`CareAlertFeedback` 服务与 Python 客户端仍在」。
- **理由**：产品明确要求保留；与 clinic/tip 飞轮删除刻意区分。

### D4：TipStream 从 contracts 接口移除

- **选择**：若 `contracts` 上 `TipStream` 仅为 tip 使用，同步删除接口方法与相关类型，避免假实现。
- **约束**：不得牵连 clinic WS / care-alert 契约。

### D5：不改 usage 统计策略文件的「新增排除」

- **选择**：从 `maintenance_skip.go` **删除** clinic/tip feedback 排除项（路径已不存在）；tip generate 本就不在 skip 列表，随路由删除即可。
- **注意**：本变更不向负责人新问 usage（无新 App 接口）；仅清理死配置。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 外部脚本/旧 App 仍打 tip 或 clinic feedback | **BREAKING** 在 proposal 标明；网关将 404；无兼容期 |
| 误删 care-alert 相关代码 | tasks 含显式保留验收；grep 反代与 Bind |
| `history.html` 删 tip 后页面残留报错按钮 | 同步删 UI/JS 调用 |
| Python 侧 tip/clinic feedback 端点仍存在 | 本变更仅 Go/gateway/运维；Python 可另开清理，非阻塞 |

## Migration Plan

1. 合并后发版 voice + gateway-app；无 DB 迁移、无 Redis 键变更。
2. 回滚：从 git 恢复删文件与 Bind/反代即可；无数据回填。
3. 归档前：对照 v3.0.0 tip/feedback Requirement，合并进新版基线或标注 REMOVED。

## Open Questions

- （无）care-alert 飞轮保留、运维 tip 删除、clinic feedback 删除均已由产品确认。
