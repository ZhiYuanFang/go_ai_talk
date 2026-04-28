## DAO 分库同步说明

目标：避免 `gf gen dao` 覆盖我们在 DAO 层的分库逻辑，同时支持按不同数据库同步表结构。

### 1. 防覆盖策略

- 生成文件：`internal/dao/*.go` 与 `internal/dao/internal/*.go`
- 自定义分库逻辑：统一放在 `internal/dao/*_ext.go`
  - `device_ext.go`
  - `voice_ext.go`
  - `history_ext.go`
  - `domain_db.go`

这样即使 `gf gen dao` 更新了生成文件，自定义分库方法仍然保留。

### 2. 数据库与表映射

- `ai_voice_device`：`user,event,action`
- `ai_voice_voice`：`qa,suggest`
- `ai_voice_history`：`history`

### 3. 同步命令

在项目根目录执行：

- 全量同步（按三库配置）：
  - `make dao.sync`
- 只同步 device 域：
  - `make dao.sync.device`
- 只同步 voice 域：
  - `make dao.sync.voice`
- 只同步 history 域：
  - `make dao.sync.history`

### 4. 对应配置文件

- 全量：`hack/config.yaml`
- device：`hack/config.dao.device.yaml`
- voice：`hack/config.dao.voice.yaml`
- history：`hack/config.dao.history.yaml`
