## Context

`llm-care-alert-daily` 已在 voice-service 落地日缓存与选模骨架：`resolveCareAlertModelProfile` → `isAccountVIP`，但后者以 `deviceNo` 为入参且恒返回 false。Flutter 主设计要求 Go 读「账号 VIP」；仓库内 `wx` / `user` entity 尚无 VIP 列。账号主键在全站为 `wx.id`（JWT `sub` → gateway `X-Internal-Wx-Id`）；clinic / voice 额度亦按 `wxId`，care-alert VIP 应对齐。

约束：voice 不得直查 device 域 `wx` 表；跨域 MUST 经 HTTP 契约。本变更不引入新的 Redis 读缓存键（VIP 仅 miss 生成时读一次；失败降级）。

## Goals / Non-Goals

**Goals:**

- `wx` 表增加 VIP 列，device-service 可按 `wxId` 读取。
- care-alert 全路径强制 `wxId > 0`；选模按触发者 VIP；查 VIP 失败降级 Zhipu。
- 日缓存维度保持 `deviceNo + 上海日`（触发者权益，接受多看护 race）。

**Non-Goals:**

- 订阅购买、续费、后台改 VIP 的完整写路径（可后续 admin；本变更以列为准，写路径仅需保证 DDL 默认可读）。
- 「宝宝侧最高 VIP」聚合。
- 按 VIP 分拆日缓存键或缓存模型名。
- 扣减 clinic / voice AI 配额。

## Decisions

### D1：VIP 落在 `wx.is_vip`

- **选择**：`ai_voice_device.wx` 新增 `is_vip TINYINT NOT NULL DEFAULT 0`（0=非 VIP，1=VIP）。
- **理由**：账号主键即 `wx.id`；与「权益跟人」一致；DDL 最小。
- **备选**：独立 `subscription` 表 → 本期过重；`vip_expire_at` → 可后续加列，本期布尔足够。
- **交付**：`hack/ddl_wx_is_vip.sql` + 同步 `entity.Wx` / `dao` columns（GoFrame 生成或手改对齐现网惯例）。

### D2：device internal 读契约，voice 经 HTTP

- **路径建议**：`GET /device/app/api/user/internal/vip-by-wx-id?wxId=`（与现有 `wx-id-by-device-no` / `by-id` 同族；Header `X-Device-Gateway-Internal-Secret`）。
- **响应**：`{ "wxId": <int64>, "isVip": <bool> }`；无行时 `isVip=false`（或 404 由调用方视为非 VIP——实现取「空行 = false」更简）。
- **voice**：扩展 `device.Remote*` / `userInternalHTTP` 客户端；`isAccountVIP(ctx, wxID int64) bool` 仅调该契约。
- **禁止**：`internal/services/voice` import `dao.Wx`；禁止 `ResolveVoiceWxID` 的 deviceNo fallback 用于 care-alert 鉴权。

### D3：care-alert 强制 Header wxId

- controller（或 service 入口）解析 `X-Internal-Wx-Id`，`wxId <= 0` → 明确错误（建议对齐账号类「缺少/无效 X-Internal-Wx-Id」或 AI 未登录语义；实现时与现有 device App 错误码风格一致即可）。
- GET / DELETE / POST 三条均强制；`deviceNo` 仍必填（宝宝维度）。
- 网关已 Bearer 注入头；本变更不改反代前缀，不改 auth 白名单（care-alert 保持需登录）。

### D4：触发者权益 + 查失败降级

```
miss 生成：
  wxId(必填) ──HTTP──▶ device isVip?
                         │
              ok+true ───┼──▶ DeepSeek
              ok+false ──┤
              err/timeout┘──▶ Zhipu + Warning（不失败请求）
hit 缓存：不查 VIP
```

- 日志：生成成功日志带 `wxId` 与 `vip=`（降级时注明 `vipDegraded=true` 可选）。

### D5：不单独缓存 VIP

- VIP 变更低频；仅 miss 路径一次跨服务读；失败已降级。避免新 Redis 键与负责人确认成本。

### D6：文档

- 更新 `openspec/changes/llm-care-alert-daily/CONTRACT.md`：VIP 主键、强制 wxId、降级语义、去掉「恒 false」。
- DDL 执行说明写入 tasks；若触及 runbook 惯例可在 `docs/runbooks/release-deploy-and-run.md` 补一句（与历史 `ddl_wx_*` 一致）。

## Risks / Trade-offs

- [多看护 race] VIP 先生成 → 非 VIP 同日吃到 DeepSeek 列表；非 VIP 先生成 → VIP 同日吃到 Zhipu → **接受**（触发者权益）。
- [降级掩盖故障] device VIP API 长期失败会全员 Zhipu → Warning 可观测；不阻断主路径优先可用性。
- [无写路径] 列默认 0，上线后需运维/后续 admin 置位才有 VIP 流量 → 可接受；选模链路可先用手工 UPDATE 验收。
- [entity 与库漂移] DDL 未执行则读列失败 → 按降级处理；发布 checklist 须先 DDL。

## Migration Plan

1. 在 device 库执行 `hack/ddl_wx_is_vip.sql`（prod/test 各一份）。
2. 发布 device-service（internal VIP API + entity）。
3. 发布 voice-service（强制 wxId + Remote VIP + 选模）。
4. 回滚：voice 可临时将 `isAccountVIP` 恒 false（行为回退）；列可保留（兼容）。

## Open Questions

- （无阻塞）VIP 写路径 / admin UI 是否另开变更——默认另开。
- 错误码文案：统一用 `缺少 X-Internal-Wx-Id` vs `40301`——实现时对齐邻近 device App / voice 惯例即可，不阻塞设计。
