# Proposal

## Why

宝宝逻辑身份键 `device_no` 一旦泄露，陌生人可通过 `device_login` / `bindwx` 查看或绑定该宝宝数据。Hub 今日仅能注册与列表，无法在保留历史/权益数据的前提下更换号面。需要运维可在 Hub 将已泄露号原地改为新号，并迫使 App 重新登录以拿到新号（不改 Flutter 代码）。

## What Changes

- Hub 设备管理增加「修改宝宝号 / 改名」入口：列表行改名，以及可自输入 OLD+NEW 的补迁入口；输入或生成 `NEW`，二次确认后执行原地改名。
- 改名编排在写本域 / 调下游前先探测 history、cash、gateway 可达（dryRun），避免下游不可达时中途半迁。
- device 域编排跨服务迁移：将各业务库中该宝宝分区键从 `OLD` 替换为 `NEW`（保数据连续）；`OLD` 离开注册表后不可再 `EnsureRegistered` / 登录 / 绑定。
- 改名成功后吊销该宝宝相关账号的 refresh（并尽快使旧 access 不可用），迫使 Flutter 走现有登录流；登录响应返回新 `deviceNo`，客户端无需改代码即可覆盖本地号。
- history 写路径对 `deviceNo` 增加「须已注册」校验，避免改名后仍用旧号静默写入空壳分区。
- **非目标**：一账号多宝宝（后续独立变更）；不强断 WS 以外的「物理踢人」作为独立产品能力（会话失效即足够）；不为单宝宝 `UPDATE` 做分批占库专项（运维接受短事务一次替换）。
- 问卷/订单等审计快照中的旧号默认保留不跟迁（展示历史提交时的号面）。

## Capabilities

### New Capabilities

- （无）本变更落在既有域行为上，不新增独立 capability 目录。

### Modified Capabilities

- `identity-device`：新增 Hub/管理侧宝宝号原地改名编排；改名后旧号对注册校验与绑定/设备登录失效；绑定引用（`wx.device_no` 等）随号迁移。
- `history-feeding`：history 分区键随改名迁移（不跟迁已退役 suggest）；喂养历史写入前 MUST 校验 device 域已注册该 `deviceNo`。
- `care-commerce`：对机权益与次数等以 `device_no` 为键的行随改名迁移。
- `xiaozhi-mcp`：喂养落点 `xiaozhi_mcp_binding.device_no` 随改名迁移。
- `gateway-edge`：支持按宝宝相关会话吊销 refresh（及必要的 access 失效策略），使客户端在不改 Flutter 的前提下重登获取新 `deviceNo`；可选修正 refresh 时对 wx 会话以 device 绑定为权威（避免 refresh 载荷钉死旧号）。
- `admin-hub`：设备管理页提供改名 UI、确认与结果反馈（含新号展示）。

## Impact

- **进程**：`device-service`（编排与本域表）、`history-service`、`cash-service`、`gateway-app-server`（令牌吊销 / refresh 语义）、Hub 静态页；MCP 绑定在 device 库则随 device 迁移，桥接进程靠重连/对账。
- **API**：新增 device 管理改名接口（及 history/cash 内部 migrate 契约）；gateway 内部或管理侧吊销能力。
- **客户端**：不改 Flutter；依赖 refresh 失败 → 现有重登 → 登录 JSON `deviceNo`。
- **数据**：device / history / cash 多库字符串替换；Redis 旧键可删可等 TTL；跨库无分布式事务，接受短窗不一致。
- **后续**：一账号多宝宝单独提案；本设计以「宝宝身份键」为主语，绑定层仅「更新所有指向该宝宝的引用」。
