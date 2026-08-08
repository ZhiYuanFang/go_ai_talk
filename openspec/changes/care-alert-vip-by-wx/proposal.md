## Why

护理留意（care-alert）日生成的 VIP→DeepSeek / 非 VIP→Zhipu 选模当前按 `deviceNo` 占位，且 `isAccountVIP` 恒为 false；VIP 是账号权益，主键应为 `wx.id`。同时本需求不允许纯设备会话调用 care-alert，必须携带有效 `wxId`，并在 `wx` 表落地 VIP 列供真实读取。

## What Changes

- 在 `ai_voice_device.wx` 增加账号 VIP 列（布尔或等价语义），由 **device-service** 维护与读取。
- 新增 device 域内部查询契约：按 `wxId` 返回是否 VIP；**voice-service 禁止直查 wx 表**。
- care-alert 三条 HTTP（GET daily / DELETE item / POST feedback）**必须**携带有效 `X-Internal-Wx-Id`（`wxId > 0`）；禁止用 `deviceNo` 反查 wx 作为登录旁路。
- `isAccountVIP` / 选模入参改为 `wxId`；日缓存仍按 `deviceNo + Asia/Shanghai day`（触发者权益：仅 miss 生成时读触发者 VIP）。
- VIP 查询失败或超时：**降级为非 VIP（Zhipu）**，不阻断日列表主路径；打 Warning 日志。
- 同步修正 `llm-care-alert-daily/CONTRACT.md` 中「恒 false / deviceNo VIP」表述。

## Capabilities

### New Capabilities

- `wx-account-vip`：`wx` 表 VIP 列语义与 device-service 对外（internal）按 `wxId` 读取契约。
- `care-alert-vip-selection`：care-alert 强制 `wxId`、按触发者 VIP 选模、查失败降级、与宝宝日缓存维度分离。

### Modified Capabilities

- （无）基线 `openspec/specs/v3.0.0` 尚无独立 care-alert / wx-VIP capability；本变更以增量规格引入，归档合并时再并入版本基线。

## Impact

- **进程 / 库**：`device-service` → `ai_voice_device.wx`（`DEVICE_DB_LINK` / `database.default`）；DDL 与 entity/dao 同步。
- **进程**：`voice-service` 编排 care-alert 选模，经 HTTP 调 device internal VIP 查询；不新增 Redis 读缓存（沿用既有日缓存键，VIP 不单独缓存）。
- **API**：care-alert 鉴权语义收紧（缺/无效 wxId 拒绝）；gateway 已注入 `X-Internal-Wx-Id`，反代前缀 `/device/api/care-alert/*` 不变。
- **相关文件**：`care_alert_service.go`、`device_care_alert_controller.go`、device wx entity/dao、device internal API、`openspec/changes/llm-care-alert-daily/CONTRACT.md`。
- **非目标**：不实现订阅购买/续费写路径；不按「宝宝侧最高 VIP」聚合；不扣 clinic 配额。
