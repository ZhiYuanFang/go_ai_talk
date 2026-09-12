## Context

成长轨迹（`growth_trajectory_predict`）已接入商业开通与 voice SSE，但权益落在 `feature_entitlement(device_no)`，与值得留意同为「设备全家共享」。业务要求改为：**开通跟人（wx.id）**；**最新报告按 device_no 读且免开通、仍须登录**；VIP 仍可免开通；付费 19 元/30 天不变。

当前 `ActivateFeature` 拒绝 `ActivationSubjectUser`；catalog 只按 device 合成开通态；Admin「开通功能管理」无主体字段；日限 Redis 按 device。本变更在 cash 开通原子层通用化 user 主体，并将成长轨迹切到该路径；值得留意等保持 device。

约束：跨域仍走 HTTP 契约；Redis 键走 `cachekit`；接口版本不破坏既有 v1 路径语义时优先兼容扩展字段；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- 落地 `activation_subject`（device|user）驱动 Activate / 三通道履约 / catalog / access。
- 成长轨迹：user 权益 ∨ VIP；latest 免开通仍登录；日限跟人；邀请「人一次 + 设备一次」。
- Admin 列表与表单标明「对人 / 对机」，可查看并以服务端字段为准维护（至少可读；一期允许 Admin 更新 subject，预测类强制 device）。

**Non-Goals:**

- 不把值得留意改成账号维。
- 不自动迁移已有成长轨迹 device 权益行。
- 不新增成长轨迹多版本历史表（仍 `growth_trajectory_latest` 一条）。
- 不改 VIP 商品与 VIP 履约表。
- 不引入新的 App 对外 path 版本（在现有 v1 上扩展字段/行为）。

## Decisions

### D1：并行 `feature_user_entitlement`，不改写旧表唯一键

- **选择**：新建 `feature_user_entitlement`（字段镜像设备表，`UNIQUE(wx_id, feature_id)`）；`feature_entitlement` 仅 device。
- **理由**：旧表 `uk_device_feature` 全是设备语义；硬加 subject 列迁数据风险高。
- **备选**：单表 `subject_type+subject_key` → 否决（迁移与索引复杂）。

### D2：`feature_def.activation_subject` 驱动，而非仅硬编码 featureId

- **选择**：列值 `device`（默认）|`user`；种子成长轨迹=`user`；Activate 与支付/邀请/广告读定义决定 SubjectType/SubjectKey。
- **理由**：方案 B 可扩展；Admin 可追踪。
- **规则**：`prediction_unlock` / `allowed_count` 类 MUST 保持 device；若 Admin 误设 user MUST 拒绝或忽略写入。

### D3：Activate 双路径 Grant

```
ActivateFeature
  subject=device → upsert feature_entitlement(device_no,…)
  subject=user   → upsert feature_user_entitlement(wx_id,…)
                 SubjectKey 解析为正整数 wxId
```

支付：`SubjectKey=order.WxId`（成长轨迹）；邀请/广告：`SubjectKey=redeemer/actor wxId`。`deviceNo` 仍用于订单审计、邀请设备闸、广告幂等（可保留 device 维幂等键，但权益写 user）。

### D4：成长轨迹 access / latest / turn

| API | 登录 | 开通∨VIP |
|-----|------|----------|
| `GET .../latest` | MUST | MUST NOT 校验 |
| `POST .../turn` | MUST | MUST（user 权益 ∨ VIP） |
| cash `.../growth-trajectory/access` | wxId MUST | 仅查 user 表 ∨ VIP；deviceNo 可选（兼容调用方可忽略） |

### D5：邀请双闸

- 保留 `InviteOncePerDevice`（含成长轨迹）→ `feature_invite_device_grant`。
- 新增 `InviteOncePerUser(featureID)`（至少成长轨迹）→ 聚合 `feature_invite_feature_grant` 按 `(redeemer_wx_id, feature_id)` 任意码已存在则拒；或等价专用表。人×码×功能去重仍保留。

### D6：日限跟人

- `cachekit`：`GrowthTrajectoryDailyUsageKey` identifier 改为 `wxId:yyyyMMdd`（或新 builder，废弃设备维语义注释）。
- turn 预检/INCR 使用 wxId；single-flight 建议按 wxId（与「一人一份」一致）。

### D7：Catalog 合成

- `GetFeatureCatalog` 增加 `wxID`：device 功能读 `feature_entitlement`；user 功能读 `feature_user_entitlement`。
- App 路由已有登录头时注入；未登录则 user 功能 `unlocked=false`。

### D8：Admin「对人/对机」

- Admin defs GET/POST 增加 `activationSubject`（`device`|`user`）。
- `cash-feature-admin.html`：列表列「开放主体」展示「对机」「对人」；编辑区下拉/只读说明（编号旁固定文案）；保存写回。
- 列表 MUST 一眼可区分，避免运维误以为成长轨迹仍全家共享。

### D9：存量

- 不迁移 device 行；access 不再认成长轨迹的 device entitlement。
- 若需补偿，运维按 `feature_order.wx_id` 手工/脚本回填（本变更不强制脚本）。

## Risks / Trade-offs

- [Risk] 已付费写在 device 上的用户升级后「突然未开通」→ 未上量可接受；已上量则需运维回填或临时双读一版。
- [Risk] catalog 未传 wx → 成长轨迹永远显示未开通 → 网关/controller MUST 注入 `X-Internal-Wx-Id`。
- [Risk] 广告 API 缺 wx → user 履约失败 → CompleteFeatureAd 签名补 wx，缺则拒。
- [Trade-off] 邀请保留设备一次：同设备第二账号无法再邀请开成长轨迹（即使权益跟人）→ 产品接受的防刷。
- [Trade-off] Admin 可改 subject：误改可能导致新履约写错表 → 文档+预测强制 device；可选后续加「已有权益时禁止改 subject」。

## Migration Plan

1. 部署 cash：EnsureSchema 新表/新列 + 种子成长轨迹 `activation_subject=user`；Activate/三通道/access/catalog/Admin。
2. 部署 voice：latest 去门禁；日限 key 跟人；turn 仍调 access。
3. 部署 gateway 静态 Admin 页（Ctrl+F5）。
4. 验收：A 付费开通 → 仅 A 可 turn；B 同设备不可 turn（非 VIP）；latest 双方登录均可读该 device 报告；Admin 列表显示成长轨迹「对人」。
5. 回滚：回退代码；新表可留；若已写 user 行，回退旧代码将再次只认 device（用户需重新以旧逻辑开通）——回滚窗口宜短。

## Open Questions

- （已决）日限跟人；邀请人一次+设备一次；方案 B；latest 免开通仍登录；VIP 免开通；19 元/30 天；不自动迁存量。
- Admin 一期是否允许改 `activationSubject`，或仅展示种子值？**默认：允许更新，预测类拒绝改为 user。**
