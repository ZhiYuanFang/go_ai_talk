## SLO、告警与故障响应手册（Task 7.5）

### SLO 目标（首版）

- Gateway 可用性：`>= 99.9%`（按 5 分钟窗口聚合）
- Gateway 5xx 比例：`< 1%`（日均）
- History 查询成功率：`>= 99.5%`
- Worker 任务最终成功率（含重试）：`>= 99%`
- Outbox 补投延迟：`P95 < 5 分钟`

### 告警规则（PrometheusRule）

文件：`manifest/deploy/kustomize/overlays/develop/prometheus-rules.yaml`

- `GatewayDown` / `HistoryServiceDown` / `WorkerDown`（critical）
- `GatewayHighErrorRate`（warning）
- `WorkerHighCPU`（warning）

### 告警分级与响应时限

- `critical`：5 分钟内响应，30 分钟内给出止血方案
- `warning`：30 分钟内响应，2 小时内完成原因定位

### 标准处置流程（Incident）

1. **确认影响面**
   - 看 Prometheus/Grafana 告警面板
   - 关注网关可用性、5xx、worker backlog
2. **快速止血**
   - 网关异常：优先回滚最近版本或临时扩容
   - 历史服务异常：切换 `HISTORY_SERVICE_MODE=local` 回退到单体实现
   - worker 异常：临时关闭 `MQ_CONSUMER_ENABLED` 并保留 outbox 积压
3. **恢复一致性**
   - MQ 恢复后观察 outbox 自动补投
   - 必要时执行：
     - `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1" -ResetFailedToPending`
4. **复盘与预防**
   - 记录时间线、影响范围、根因、改进项
   - 将改进项进入下一迭代任务

### 回滚演练清单

- [ ] 演练网关版本回滚（deployment image 回滚）
- [ ] 演练 history 路由回退（proxy -> local）
- [ ] 演练 worker 停止消费与恢复补投
- [ ] 演练 outbox failed 重放路径
- [ ] 演练告警触发 -> 通知 -> 恢复 -> 复盘闭环
