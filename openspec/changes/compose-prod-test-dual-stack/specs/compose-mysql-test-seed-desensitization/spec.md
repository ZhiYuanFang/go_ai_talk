## ADDED Requirements

### Requirement: 测试 MySQL 库 SHALL 与生产库名隔离

测试环境各业务库 SHALL 使用与生产对应、带 `_test` 后缀的库名（至少包含 `ai_voice_history_test`、`ai_voice_device_test`、`ai_voice_voice_test`、`ai_voice_worker_test`、`ai_voice_app_test`、`ai_voice_ucg_test`）。`.env.test` 中各 `*_DB_LINK` SHALL 指向上述测试库，MUST NOT 指向生产库名。

#### Scenario: 测试 device-service 连接测试库

- **WHEN** 测试栈 device-service 启动且 `DEVICE_DB_LINK` 已按 `.env.test.example` 配置
- **THEN** 进程 SHALL 连接 `ai_voice_device_test`（或 documented 等价名），SHALL NOT 连接 `ai_voice_device`

### Requirement: 仓库 SHALL 提供生产到测试的脱敏种子流程

仓库 SHALL 在 runbook 和/或 `hack/mask-seed-data.sh` 中描述可重复流程：从生产 `ai_voice_*` 导出 → 脱敏 → 导入 `ai_voice_*_test`。脱敏 SHALL 至少处理：用户手机号、微信 openid/unionid、refresh token/session、设备标识（替换或前缀化）。导入 SHALL 覆盖测试库既有数据（运维须在 runbook 中警告）。

#### Scenario: 脱敏后测试库无原始手机号

- **WHEN** 运维按文档完成脱敏 import
- **THEN** 测试库 user 相关表中 SHALL NOT 保留与生产 export 完全相同的手机号明文

#### Scenario: 发版前刷新测试种子

- **WHEN** 准备发布新的 release candidate
- **THEN** runbook SHALL 要求（或 recommend 作为 checklist 必项）在测试验收前执行一次脱敏种子刷新

### Requirement: 脱敏种子 SHALL 同步测试静态资源

当生产种子包含 `/ai_talk_images/` 路径引用时，运维 SHALL 将对应 logo 文件同步至宿主机 `/ai_talk_images_test/`（或 test overlay documented 路径），使测试管理页与 App 静态读链路可验收。

#### Scenario: 测试环境 logo 可读

- **WHEN** 测试库 event 行引用 `/ai_talk_images/<file>` 且文件已同步至测试静态目录
- **THEN** 经 test gateway 或 gateway-app 反代的静态请求 SHALL 返回 200
