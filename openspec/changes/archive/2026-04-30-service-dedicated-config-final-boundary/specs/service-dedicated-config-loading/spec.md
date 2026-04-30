## ADDED Requirements

### Requirement: 服务配置 MUST 按服务进程独立加载
`voice-service`、`device-service`、`history-service`、`gateway-service` MUST 具备独立默认配置文件，服务启动时 MUST 优先使用本服务默认配置，并允许通过 `GF_GCFG_FILE` 显式覆盖。

#### Scenario: voice-service 启动未指定 GF_GCFG_FILE
- **WHEN** 启动 `voice-service` 且环境变量未设置 `GF_GCFG_FILE`
- **THEN** 系统 MUST 加载 `voice-service` 专属默认配置文件

#### Scenario: device-service 启动指定 GF_GCFG_FILE
- **WHEN** 启动 `device-service` 且设置 `GF_GCFG_FILE`
- **THEN** 系统 MUST 使用指定配置文件并覆盖默认路径

### Requirement: 服务级覆盖变量 MUST 仅影响本服务
服务级环境变量覆盖（如数据库连接、监听地址）MUST 仅影响当前服务实例，不得通过同名变量隐式影响其他服务配置行为。

#### Scenario: 设置 VOICE_DB_LINK
- **WHEN** 部署仅设置 `VOICE_DB_LINK`
- **THEN** 系统 MUST 只影响 `voice-service` 数据库连接，不得改变 `history-service` 与 `device-service` 连接
