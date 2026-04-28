## History Service Rollout Runbook

本 Runbook 用于执行 `3.4`：将 history 流量从本地处理逐步迁移到 `history-service`，并验证可快速回滚。

### 前置条件

- `history-service` 已启动并可访问（默认 `http://127.0.0.1:9801`）。
- gateway 与 history-service 的配置已完成：
  - `HISTORY_API_PROXY_URL`
  - `HISTORY_API_ROUTE_MODE`
  - `HISTORY_API_PROXY_CANARY_PERCENT`
- 已准备至少 1 个可重复验证的 `deviceNo` 样本。

### 分阶段迁移步骤

1. **阶段 0（基线）**
   - 目标：确认本地路径稳定。
   - 命令：`.\hack\history-rollout.ps1 -Stage local`
   - 验证：历史查询成功率、延迟与基线一致。

2. **阶段 1（10% 灰度）**
   - 命令：`.\hack\history-rollout.ps1 -Stage canary10`
   - 观察窗口：建议 15-30 分钟。
   - 通过条件：错误率无明显升高；关键接口响应结构不变。

3. **阶段 2（50% 灰度）**
   - 命令：`.\hack\history-rollout.ps1 -Stage canary50`
   - 观察窗口：建议 30-60 分钟。
   - 通过条件：延迟与错误率仍在阈值内；日志无持续代理失败。

4. **阶段 3（100% 灰度）**
   - 命令：`.\hack\history-rollout.ps1 -Stage canary100`
   - 通过条件：可稳定运行一个完整业务周期。

5. **阶段 4（全量代理）**
   - 命令：`.\hack\history-rollout.ps1 -Stage proxy`
   - 通过条件：与 canary100 一致，并确认配置简化可用。

### 回滚验证路径

- 任一阶段若异常（错误率升高、核心接口失败）立即执行：
  - `.\hack\history-rollout.ps1 -Stage rollback`
- 执行后重启 gateway，使环境变量生效。
- 验证以下接口恢复本地处理：
  - `GET /device/history/api/list?deviceNo=<sample>`
  - `GET /device/history/api/suggest?deviceNo=<sample>`

### 验收清单

- [ ] 已完成 local -> canary10 -> canary50 -> canary100 -> proxy 逐阶段切换
- [ ] 每阶段均记录成功率/延迟观测结果
- [ ] 已执行 rollback 并验证本地路径恢复
- [ ] 回滚操作可在 5 分钟内完成并恢复服务
