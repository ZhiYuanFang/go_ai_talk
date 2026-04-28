## Voice Session Multi-Instance Verification

用于 `Task 4.4`：验证 Redis 会话外置后，多实例会话连续性与故障切换行为。

### 前置条件

- 已启动 Redis 集群（参考 `docs/redis-cluster-local.md`）。
- 已启动两个 gateway 实例（示例）：
  - gateway-A: `http://127.0.0.1:9701`
  - gateway-B: `http://127.0.0.1:9702`
- 两个实例均开启：
  - `VOICE_SESSION_BACKEND=redis`
  - `VOICE_GUARD_REDIS_ENABLED=true`
- 两个实例连接同一 Redis 集群种子地址。

### 自动化验证（推荐）

执行：

`powershell -ExecutionPolicy Bypass -File "hack/voice-session-verify.ps1" -GatewayA "http://127.0.0.1:9701" -GatewayB "http://127.0.0.1:9702" -AdminPassword "a521521521" -DeviceNo "device-verify-001"`

脚本会：

1. 在 A 发起首轮文本对话
2. 在 B 发起同设备第二轮文本对话
3. 打印两轮响应与验证状态
4. 提示 Redis 故障切换人工演练命令

### 故障切换人工演练

1. 先记录当前接口成功率与延迟基线。
2. 暂停部分 Redis 节点（例如一个主节点）：
   - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml stop redis-node-1`
3. 重复执行脚本，确认请求仍可成功（允许降级内存路径）。
4. 恢复节点：
   - `docker compose -f manifest/docker/docker-compose.redis-cluster.yml start redis-node-1`
5. 再次执行脚本，确认恢复后行为稳定。

### 验收清单

- [ ] 同一 `deviceNo` 在 A/B 实例间连续请求均成功
- [ ] Redis 局部故障下对话链路不完全中断
- [ ] 恢复后无持续错误告警
- [ ] 已记录故障演练结果与回滚步骤
