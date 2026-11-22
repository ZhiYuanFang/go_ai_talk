# Proposal

## Why

「成长建议」生成与 Python 落库路径早已退役（`clinic-bridge` 已要求删除生成），但 voice List/Delete、history 反代、Hub 历史页列表，以及宝宝号改名时对 `suggest` 表的 UPDATE 仍留在代码里。测试等环境甚至没有该表，导致 `hub-rename-baby-device-no` 的 history 迁移在 1146 上失败。写入侧已无数据可落库，整段读删壳属于死代码，应一次性拆除。

## What Changes

- **BREAKING**：下线 voice 内部 suggest List/Delete API，以及 `/device/history/api/suggest*`、`/device/admin/api/history/suggest` 等对外/运维读删口。
- 删除 Hub `history.html` 中成长建议区块与 `loadSuggest` 调用链。
- 删除 voice `suggest_internal`、相关 controller/契约方法/HTTP path 辅助函数与 history 侧委托实现。
- **宝宝号改名**：`history.MigrateDeviceNo` MUST NOT 再 UPDATE `suggest`（只迁 `history` 等仍有效分区表）；注释与内部 migrate API 文案同步去 suggest。
- 规格对齐：强化「无 suggest 读写面」；清理 `voice-chat` / `llm-lanes` 中仍假定成长建议 LLM 路径的条款；`history-feeding` 中「Voice MUST 在本域处理 suggest 表」改为退役语义。
- **不** DROP 库表、**不**强制清历史行（MAY 留库）；**不**改 Flutter（若仍有调用应 404/无路由）；**不**含 Hub 自由填写 OLD→NEW 补迁入口（属改名运维，另案）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `clinic-bridge`：将「生成已删、Hub 可读删历史行」收紧为 MUST 拆除剩余 suggest 读删运维面与 API。
- `history-feeding`：退役 suggest 列表/删除契约与 history→voice 委托；改名迁移 MUST NOT 触碰 `suggest`。
- `voice-chat`：去掉独立成长建议意图/落库相关要求；`target_type=suggest` 不得再驱动业务写路径。
- `llm-lanes`：从 voiceUnderstanding 必经路径说明中去掉「成长建议」专径与对应 Scenario。
- `platform-runtime`：分库 ownership 中 `suggest` 改为遗留表、进程 MUST NOT 再作为业务读写面（表可物理残留）。

## Impact

- **进程**：`voice-service`、`history-service`（及绑定 history 管理 API 的 device/gateway 反代面）、静态 Hub（`history.html`）。
- **API**：`/voice/internal/api/suggest/*`、`/device/history/api/suggest*`、`/device/admin/api/history/suggest`。
- **改名**：依赖 `MigrateDeviceNo` 的 Hub 改名在无 `suggest` 表环境可继续完成 history 段。
- **DAO**：应用层不再引用 `dao.Suggest`；entity/do/dao 生成物随无引用删除或标遗留（实现阶段择一保持可编译）。
- **兄弟仓**：Python 已不返回可落库 suggest；Flutter 若仍有历史 suggest 页属于另仓清理，本仓只保证 Go 侧无入口。
