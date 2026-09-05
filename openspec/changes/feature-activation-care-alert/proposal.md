## Why

值得留意已有 cash 喂养资格（连续有效作息日），但业务接口与客户端未强制「功能开通」门槛；开通履约又散落在支付/邀请/广告分支里，邀请天数写死、广告与邀请效果不一致，难以复用到下一个付费功能。需要把「共用开通原子工具 + 值得留意双门禁 + 设备维共享」一次补齐。

## What Changes

- 在 **cash-service** 引入功能开通共用原子入口：按入参区分激活主体（`user` | `device`）、开通通道（`payment` | `invite_code` | `ad`），统一解析授予效果（永久 / 限时 / 条数增量），供支付回调、邀请兑换、广告完成汇入。
- 种子新功能「值得留意智能提醒」（稳定 `featureId`，与客户端约定）：付费永久；邀请码与广告授予天数 **Admin 可配**（默认 7）；广告效果与邀请码一致；所有开通事件统一保留 `ad` 能力（一期可不做 Flutter 广告入口）。
- 邀请码对值得留意：**设备维**权益（`device_no` 全家共享）；**一个 `deviceNo` 对本功能仅能成功兑码一次**；获客原力 / 兑换流水仍按使用邀请码的用户记录。
- VIP 与预测槽一致：时效覆盖值得留意使用权，**不写**功能权益表；读模型由客户端/`isVip` 与设备 entitlement OR。
- **cash internal** 提供「是否可看值得留意」合成契约（喂养资格 ∧（设备开通 ∨ VIP））；**voice** `CareAlertDaily`（及必要删除/反馈入口）经该契约 dual-gate，禁止直查 cash 库；cash 不可达 fail-closed。
- Admin「开通功能管理」：功能定义表单暴露「邀请/广告授予天数」；支付仍以 SKU `duration_days` 为准。
- **Flutter**（孪生仓 `flutter_ai_talk`）：喂养合格但未开通（且非 VIP）时展示可点开通引导，跳转开通中心；合格且开通/VIP 后再拉 daily。
- 喂养资格 `qualified` **不得**被 VIP/功能开通短路（延续既有规格）。

## Capabilities

### New Capabilities

- `feature-activation-toolkit`：共用开通原子工具（主体/通道枚举、效果解析、支付/邀请/广告汇入）；邀请/广告天数可配；ad 与 invite 同源效果。
- `care-alert-feature-unlock`：值得留意功能定义与设备维开通语义（付费永久、邀请/广告限时、device 兑码一次、VIP 覆盖）。
- `care-alert-access-gate`：cash internal 可看合成契约 + voice 服务端 dual-gate（先喂养资格，再开通/VIP）。
- `care-alert-unlock-client`：Flutter 三态 UX（未喂养合格 / 合格未开通→开通中心 / 已开通或 VIP→拉数）；目录与 VIP 覆盖对齐。

### Modified Capabilities

- （无强制改写已归档 `openspec/specs/` 基线 Requirement 文本；喂养资格相关行为以既有 `feeding-eligibility-*` / `care-alert-feeding-eligibility` 变更为准并在本变更 specs 中引用约束。）

## Impact

- **进程**：`cash-service`（开通原子、功能种子、Admin、internal access）；`voice-service`（CareAlertDaily 调 cash）；`gateway-app-server`（internal/App 反代核验、usage skip 若新增查询路径）。
- **库**：仅 `ai_voice_cash`（`feature_def` 等；可能扩展邀请设备去重表/列）；禁止 cash 直查 history/device；voice 禁止直查 cash。
- **客户端**：`flutter_ai_talk` 开通中心 + 值得留意卡片闸门（本仓规格约束孪生实现）。
- **非目标**：改预测条数累加语义；UCG 入场资格算法；新建微服务；新增 `*_test.go`；广告服务端验真；一期强制上线 Flutter 广告按钮。
