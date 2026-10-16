## 1. 工具描述

- [x] 1.1 重写 `internal/services/mcpbridge/tools.go` 的 `chatToolDescription`：宝宝顾问；谈到宝宝时调用、不限于喂养；三条覆盖喂养、宝宝信息、确认/取消；用「调用」而非堆叠「必须」
- [x] 1.2 确认工具名、`chatToolInputSchema`、`ChatHandler.Handle` 未改

## 2. 验收

- [x] 2.1 对照 spec：描述含宝宝信息且写明不限于喂养；仍只有 `baby_feeding_advisor`
