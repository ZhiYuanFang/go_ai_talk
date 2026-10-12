## Context

`hub-op-flywheel-admin` 已落地 OP / 数量 / R·source Hub 页与门户重排。兄弟仓此后又挂好：

| 层 | Collection | Admin | payload |
|----|------------|-------|---------|
| 时间 | `feeding_time_cues` | `/v1/admin/time-cache` list/bulk/delete | 相对配方（禁 Unix 主缓存） |
| 备注 | `feeding_remark_cues` | `/v1/admin/remark-cache` list/bulk/delete | `remark`（可空串） |
| R·predict | `feeding_predict_cues` | `/v1/admin/predict-cache` list/bulk/delete | `predict_channel` + `time_answer` |
| U/D 定位 | `feeding_locate_cues` | `/v1/admin/locate-cache` list/bulk/delete | `strategy`（禁 historyId） |

上游 Python change：`time-relative-flywheel`、`fix-read-route-qty-remark-flywheel`、`intent-dim-resolve-align`（Admin 任务已勾；Hub 标可选兄弟仓）。

Go Hub 现状：门户仍写「TIME 无飞轮」；缺上述四页；`PythonAIClient` 已有通用 `listCueCache` / `bulkUpsertCueCache` / `deleteCueCache` 可复用。

约束：Admin JWT + gateway 注入 `X-Admin-Password`；不直连 Chroma；Admin 不碰 usage；`/voice/admin/api/*` 反代已覆盖；不新增 Redis / 背景循环；不新增测试文件。

## Goals / Non-Goals

**Goals:**

- Hub 可整理四维新飞轮：时间、备注、R·predict、U/D 定位（list/bulk/delete）。
- 门户补四卡；修正 TIME 文案（有飞轮、缺口仍驱动扩规则）。
- 规则缺口页同步去掉「无时间飞轮」表述。

**Non-Goals:**

- 改 Python store / 配方校验 / 热路径顺序。
- 为新四层补 dedupe（上游无路由）。
- 在 Hub 做配方物化预览或 unix→配方反推。
- 合并多飞轮为单页 Tab；新建 MySQL 表。
- 改 App 接口 / usage 统计。

## Decisions

### D1：四组代理镜像 op-vectors，复用 cue 通用客户端

Hub `*-vectors` → Python `*-cache`：

| Hub | Python |
|-----|--------|
| `/voice/admin/api/time-vectors*` | `/v1/admin/time-cache*` |
| `…/remark-vectors*` | `/v1/admin/remark-cache*` |
| `…/predict-vectors*` | `/v1/admin/predict-cache*` |
| `…/locate-vectors*` | `/v1/admin/locate-cache*` |

每组：`List` / `BulkUpsert` / `Delete`；`VerifyVoiceAdminPassword`；`api/v1` DTO + 四个薄 Ctrl（或一个文件多 Ctrl，与现有 op/qty/read-source 一致）。

**备选**：gateway 直打 Python → 否决（破坏 Hub→voice 口令惯例）。

### D2：四张独立静态页，克隆 qty / op 骨架

页面提示（灌种契约，透传校验由上游）：

- **时间**：整段 document；payload 为相对配方（`kind` / `points` / `duration_minutes`）；**禁止**以 `start_time`/`end_time` Unix 作主载荷；无 dedupe。
- **备注**：payload `{remark}`，允许 `""`；read 轨热路径不写本库（页头可提示）。
- **R·predict**：`predict_channel`=`active|next`，`time_answer`=`clock|remain`。
- **定位**：`strategy`=`latest|open_timer|prefer_completed`；禁止 historyId。

登记：`admin-modules.js`（`showInNav: false`）、`admin_static_pages.go`、`gateway_app_auth_exempt.go`。

### D3：门户与规则缺口文案

`intent-agent-admin.html` 飞轮区追加：时间、备注、R·predict、U/D 定位；hint 改为「TIME 有相对配方飞轮；缺口仍见工单」。

`rule-gaps-admin.html`：TIME 主工单 + **可配合时间飞轮**；QTY 仍可配合数量飞轮；默认筛选语义不变。

### D4：部署边界

须发：`voice-service`、`gateway-app`。依赖 python-ai-talk 四套 admin 已上线；Python 未发时 Hub 页 list 失败 → adminFetch 可读错误即可。

## Risks / Trade-offs

- [Python 未发版] → 页内错误提示检查 `PYTHON_AI_TALK_URL` / 兄弟仓发版；门户其它卡不受影响
- [运维误灌绝对 Unix 进时间库] → 页头强提示相对配方；上游 bulk 拒绝非法条目
- [门户卡片变多] → 接受网格扩张；不做 Tab 合并（Non-Goal）
- [定位误带 historyId] → 页头禁止；上游 `as_locate_payload` 拒绝

## Migration Plan

1. 确认线上 python-ai-talk 已挂四套 `*-cache`。
2. 发 voice-service（代理）+ gateway-app（静态/豁免/门户）。
3. 目视：门户十入口可进；四新页 list/bulk/delete；规则缺口无「无时间飞轮」。
4. 回滚：撤 gateway 静态与 voice 代理即可；不碰 Chroma 数据。

## Open Questions

（无。四维全齐与文案对齐已在 explore 拍板；payload 枚举以上游 store 为准。）
