## DAO 同步与边界说明

当前项目以“单服务单数据库”为原则，DAO 生成与使用也按该原则收敛：

- 每个服务进程只使用自身配置中的 `database.default`
- 不再维护 `database.<group>.link` 多库分组配置
- 跨服务数据访问统一走服务 API，不跨库直查

### 1. DAO 目录与生成范围

- 生成目录：`internal/dao/*.go`、`internal/dao/internal/*.go`
- 当前不再保留基于多数据库分组回退的 `*_ext.go` 与 `domain_db.go`

### 2. 服务与数据库对应关系

- `voice-service` -> `ai_voice_voice`
- `device-service` -> `ai_voice_device`
- `history-service` -> `ai_voice_history`
- `worker-service` -> 根据其专属配置中的 `database.default`（当前用于 outbox relay）
- `gateway-service` -> 不访问数据库

### 3. 同步命令

在项目根目录执行：

- 全量：`make dao.sync`
- device：`make dao.sync.device`
- voice：`make dao.sync.voice`
- history：`make dao.sync.history`

### 4. 同步配置文件

- 全量：`hack/config.yaml`
- device：`hack/config.dao.device.yaml`
- voice：`hack/config.dao.voice.yaml`
- history：`hack/config.dao.history.yaml`

### 5. 变更检查清单

- 新需求若涉及数据访问，先确认是否属于本服务本库
- 若需要他域数据，先补跨服务 API 契约，再接入调用
- 需求完成后同步更新本文档与 `release-deploy-and-run.md`
