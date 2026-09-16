## ADDED Requirements

### Requirement: GF_LOGGER_LEVEL 覆盖默认 logger

系统 MUST 在长期运行服务启动时，于 yaml logger 配置已应用到默认 logger 之后，读取环境变量 `GF_LOGGER_LEVEL`：若 trim 后非空，则对该默认 logger 调用与 GoFrame 一致的级别字符串设置（如 `SetLevelStr`），使其覆盖 yaml 中的 `logger.level`；若为空，则 MUST 保持 yaml 级别不变。

#### Scenario: 生产设为 prod

- **WHEN** 进程环境 `GF_LOGGER_LEVEL=prod` 且服务已调用 `loggercfg.ApplyFromEnv`
- **THEN** 默认 logger 仅输出 WARN / ERRO / CRIT（及框架强制保留级别），`Infof`/`Debugf` 不再出现在该进程日志中

#### Scenario: 留空沿用 yaml

- **WHEN** `GF_LOGGER_LEVEL` 未设置或为空字符串
- **THEN** 默认 logger 级别保持配置文件中的 `logger.level`（当前服务配置多为 `all`）

#### Scenario: 生效可观测

- **WHEN** 非空 `GF_LOGGER_LEVEL` 成功应用
- **THEN** 进程 MUST 打出一条 Warning 级日志，标明服务名与所应用级别字符串

#### Scenario: 非法级别

- **WHEN** `GF_LOGGER_LEVEL` 为 GoFrame 不识别的字符串
- **THEN** 系统 MUST 记录 Error 且 MUST NOT 静默成功；logger 级别 MUST 保持应用前状态（yaml 已加载的级别）
