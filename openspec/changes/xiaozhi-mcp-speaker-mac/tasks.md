## 1. 表结构与规范化

- [x] 1.1 Ensure 增加 `speaker_mac` 列；清理空 MAC 历史行后加全局唯一索引
- [x] 1.2 实现 MAC 规范化/校验（统一为 `xx:xx:xx:xx:xx:xx` 小写）
- [x] 1.3 更新 entity/do/dao 列定义

## 2. 添加语义（upsert by MAC）

- [x] 2.1 App 添加请求增加必填 `speakerMac`；列表响应增加 `speakerMac` 与 `connected`
- [x] 2.2 `AddXiaozhiMcpBinding`：normalize MAC+token；MAC 属他宝宝拒绝；同宝宝更新 token/alias；新 MAC 插入；token 去重排除自身
- [x] 2.3 token 变更时 mcp Remove(旧)+Upsert(新)；仅 alias 变更可跳过停桥

## 3. mcp 连接状态（内存 + 内部 HTTP）

- [x] 3.1 Bridge/Manager：拨号进入读循环标 `connected=true`；断线/Remove 标 false 或删条目（键=规范化 token）
- [x] 3.2 内部 API：`/xiaozhi-mcp/internal/api/bindings/connection-status`（密钥鉴权，批量 tokens → 各 bool）
- [x] 3.3 `clients/xiaozhimcp` 增加批量查询；device `ListXiaozhiMcpBindingsForWx` 拼装 `connected`；mcp 失败则该项 false 且列表仍成功

## 4. Hub UI

- [x] 4.1 `history.html` 表单增加 MAC 输入；列表展示 MAC；提交带 `speakerMac`
- [x] 4.2 列表按 `connected` 显示绿/红灯（仅两态）

## 5. 喂养确认/取消工具描述

- [x] 5.1 扩展 `mcpbridge/tools.go` 中 `chatToolDescription`：确认词 + 取消词须调用 `baby_feeding_advisor`，原话进 transcript；不新增工具、不改 Handler

## 6. 自检

- [x] 6.1 一宝宝两 MAC 成功；MAC 跨宝宝拒绝；同 MAC 换 token 行数不变且桥切换
- [x] 6.2 桥连通列表 `connected=true`/绿灯；断线或停桥为 false/红灯；mcp 不可达时列表成功且红灯
- [x] 6.3 `tools/list` 描述含确认/取消选路；工具名仍仅 `baby_feeding_advisor`
- [x] 6.4 无新增 App path；usage/maintenance_skip 无需改；未引入 Redis 存连接态
