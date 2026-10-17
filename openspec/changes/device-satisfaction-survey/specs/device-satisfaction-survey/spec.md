## ADDED Requirements

### Requirement: 当前活动可配置且仅最新活动对 App 生效

device-service MUST 在 device 库保存满意度问卷活动与选项行。App 读取 MUST 只针对 id 最大的一条活动。Hub MUST 能编辑该活动的开关、说明、Asia/Shanghai 开始与结束时间，以及每一行选项的文案、排序与是否启用。Hub MUST 能开启新一轮（新活动 id，默认关闭）。就地修改选项文案 MUST NOT 清除已有提交。时间 MUST 以 unix 秒存储；`end_at` 早于 `start_at` 或缺少起止时 MUST 拒绝保存。

#### Scenario: 开关关闭则不展示

- **WHEN** 最新活动 `enabled=0`，且当前 unix 秒落在其起止之内
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 超出上海时区窗口不展示

- **WHEN** 最新活动开关为开，且当前 unix 秒早于 `start_at` 或晚于 `end_at`
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 新一轮不继承已提交状态

- **WHEN** 运维开启新一轮且新活动开关为开、处于有效期内
- **THEN** 曾提交上一轮的 wxId MUST 仍可能被判定为需要展示
- **AND** 上一轮提交记录 MUST 仍可按旧活动 id 查询

### Requirement: 已登录用户读取是否展示

`GET /device/app/api/survey/current` MUST 要求登录，MUST NOT 列入 Bearer 匿名白名单。身份 MUST 使用网关注入的 wxId。需要展示时响应 MUST 包含活动 id、说明、以及启用中的选项 id 与文案。无启用选项时，只要活动开关与时间窗允许且该用户未提交，MUST 仍返回需要展示。该用户已有该活动提交时 MUST 返回不展示。

#### Scenario: 未提交且活动有效

- **WHEN** 已登录用户请求 current，最新活动为开、处于有效期、该 wxId 无提交
- **THEN** 响应 MUST 表示需要展示，并带上启用选项

#### Scenario: 已提交不再展示

- **WHEN** 同一 wxId 已存在该活动的提交行
- **THEN** current MUST 表示不展示

### Requirement: 提交可只写意见且只记一次

`POST /device/app/api/survey/submit` MUST 要求登录。选项 id 列表与去除空白后的意见都为空时 MUST 拒绝，且 MUST NOT 写入提交行。至少有一项启用选项或非空意见时 MUST 写入一行，并快照所选选项当时的文案、会话 deviceNo 对应的宝宝昵称（没有则为空）、设备号与 wxId。同一活动同一 wxId MUST 至多一行。重复提交 MUST 保持单行并返回成功。选项不属于该活动或未启用时 MUST 拒绝。活动当前不可展示时 MUST 拒绝新插入。

#### Scenario: 只写意见

- **WHEN** 用户不选任何选项，意见为非空文本，且活动当前可展示
- **THEN** 系统 MUST 写入提交，选项快照为空
- **AND** 随后 current MUST 对该用户不展示

#### Scenario: 空卷不关闭

- **WHEN** 用户不选选项且意见为空白
- **THEN** 系统 MUST 拒绝
- **AND** MUST NOT 留下提交行

#### Scenario: 快照不被事后改文案影响

- **WHEN** 用户已提交某选项，运维随后修改该选项文案
- **THEN** 该提交行保存的文案 MUST 仍为提交当时的文本

### Requirement: 网关放行且不计入使用统计

gateway-app MUST 将 `/device/app/api/survey/*` 反代到 device-service。`GET /device/app/api/survey/current` 与 `POST /device/app/api/survey/submit` MUST 列入 usage 维护排除（精确方法与路径），不得计入功能使用统计。Hub 管理接口 MUST 使用 `/device/admin/api/survey/` 前缀。

#### Scenario: App 路径到达 device-service

- **WHEN** 已登录客户端请求上述 GET 或 POST
- **THEN** gateway-app MUST 按 device 反代规则转发，而不是当本机路由处理

#### Scenario: 成功调用不进使用统计

- **WHEN** 这两条 App 接口返回 2xx
- **THEN** usage 记录 MUST 跳过它们

### Requirement: Hub 可查看提交明细

Hub 问卷调查页 MUST 近全宽，宽屏下编辑区与结果区左右分栏。结果 MUST 展示用户 id、宝宝昵称、设备号、选项快照、意见与提交时间，并默认对应最新活动。

#### Scenario: 查看某次提交

- **WHEN** 运维打开问卷调查页且该活动已有提交
- **THEN** 页面 MUST 列出该条的 wxId、宝宝昵称、设备号、所选文案与意见
