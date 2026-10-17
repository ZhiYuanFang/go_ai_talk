## MODIFIED Requirements

### Requirement: 已登录用户读取是否展示

`GET /device/app/api/survey/current` MUST 要求登录，MUST NOT 列入 Bearer 匿名白名单。身份 MUST 使用网关注入的 wxId。需要展示时响应 MUST 包含活动 id、说明、以及启用中的选项 id 与文案。无启用选项时，只要活动开关与时间窗允许且该用户没有有效提交，MUST 仍返回需要展示。该用户存在 `valid=1` 的提交行时 MUST 返回不展示。仅有 `valid=0` 的提交行时 MUST 仍可返回需要展示。

#### Scenario: 未提交且活动有效

- **WHEN** 已登录用户请求 current，问卷为开、处于有效期、该 wxId 无提交
- **THEN** 响应 MUST 表示需要展示，并带上启用选项

#### Scenario: 有效提交不再展示

- **WHEN** 同一 wxId 已存在 `valid=1` 的提交行
- **THEN** current MUST 表示不展示

#### Scenario: 无效提交可以再展示

- **WHEN** 同一 wxId 只有 `valid=0` 的提交行，且问卷为开、处于有效期
- **THEN** current MUST 表示需要展示

### Requirement: 提交可只写意见且只记一次

`POST /device/app/api/survey/submit` MUST 要求登录。请求体 MUST 使用 `optionResults` 字符串数组与意见，MUST NOT 根据 `optionIds` 写入选项快照。任一结果 trim 后为空、结果之间重复、条数超过 40，或任一条超过 203 个 Unicode 码位时，系统 MUST 拒绝，且 MUST NOT 写入或改写提交行。每一条结果 MUST 等于该活动某条启用且文案非空的选项原文，再直接连接「未使用」「不好用」或「好用」之一；对不上时 MUST 拒绝。多个启用选项都能对上同一条结果时，系统 MUST 仍然接受，并保存客户端提交的那条字符串。`optionResults` 与去除空白后的意见都为空时 MUST 拒绝，且 MUST NOT 写入或改写提交行。至少有一条合法结果或非空意见时，若该 wxId 尚无行，MUST 写入一行且 `valid=1`，并把 `optionResults` 按提交顺序写入选项快照，同时保存会话 deviceNo 对应的宝宝昵称（没有则为空）、设备号与 wxId。系统 MUST NOT 回表把快照换成不含评价后缀的原文。同一活动同一 wxId MUST 至多一行。已有 `valid=1` 的行时重复提交 MUST 保持该行不变并返回成功。已有 `valid=0` 的行时 MUST 用本次结果、意见、设备号、宝宝昵称和提交时间覆盖该行，并把 `valid` 设为 `1`。活动当前不可展示时 MUST 拒绝新插入与覆盖。系统 MUST NOT 要求每个启用选项都出现在 `optionResults` 中。

#### Scenario: 只写意见

- **WHEN** 用户提交空的 `optionResults` 和非空意见，活动当前可展示，且该用户无有效提交
- **THEN** 系统 MUST 写入或覆盖提交，选项快照为空，`valid=1`
- **AND** 随后 current MUST 对该用户不展示

#### Scenario: 空卷不关闭

- **WHEN** 用户提交空的 `optionResults` 且意见为空白
- **THEN** 系统 MUST 拒绝
- **AND** MUST NOT 留下新的提交行，也 MUST NOT 把无效行改成有效

#### Scenario: 快照保存评价后缀

- **WHEN** 启用选项原文为「夜间唤醒」，用户提交的一条结果为「夜间唤醒好用」，且活动当前可展示
- **THEN** 该提交行的选项快照 MUST 包含「夜间唤醒好用」
- **AND** MUST NOT 只保存「夜间唤醒」

#### Scenario: 快照不被事后改文案影响

- **WHEN** 用户已提交某条结果，运维随后修改该选项文案
- **THEN** 该提交行保存的文案 MUST 仍为提交当时的结果字符串

#### Scenario: 非法结果不落库

- **WHEN** 用户提交的某条结果不是任一启用选项原文加上「未使用」「不好用」或「好用」
- **THEN** 系统 MUST 拒绝
- **AND** MUST NOT 写入或改写提交行

#### Scenario: 无效后再交卷覆盖同一行

- **WHEN** 该 wxId 已有 `valid=0` 的提交，用户再次提交合法结果或非空意见，且问卷可展示
- **THEN** 系统 MUST 更新这一行的快照、意见与时间
- **AND** 该行 `valid` MUST 变为 `1`
- **AND** MUST NOT 再插入第二行

#### Scenario: 不必交齐全部选项

- **WHEN** 活动有多条启用选项，用户只提交其中一条合法结果，意见为空，且问卷可展示
- **THEN** 系统 MUST 接受并只把这一条写入快照

### Requirement: Hub 可查看提交明细

Hub 问卷调查页 MUST 近全宽，宽屏下编辑区与结果区左右分栏。结果 MUST 展示用户 id、宝宝昵称、设备号、选项快照、意见、提交时间与是否有效，并且 MUST 列出这一份问卷的全部提交，包括无效行，不得按活动分表或筛选。

#### Scenario: 查看某次提交

- **WHEN** 运维打开问卷调查页且已有提交
- **THEN** 页面 MUST 列出该条的 wxId、宝宝昵称、设备号、所选文案、意见以及有效状态

### Requirement: 当前活动可配置且仅最新活动对 App 生效

device-service MUST 只维护一份满意度问卷活动及其选项行。App 读取 MUST 只针对这一份。Hub MUST 能编辑该份的开关、说明、Asia/Shanghai 开始与结束时间。Hub MUST NOT 提供「开启新一轮」或按活动切换。就地修改选项文案 MUST NOT 清除已有提交，也 MUST NOT 把 `valid=1` 的提交改成可展示。时间 MUST 以 unix 秒存储。保存开关与起止时，若 `end_at` 早于 `start_at` 或缺少起止，MUST 拒绝该次设置保存。选项行的新增与修改 MUST NOT 因起止无效而失败。

#### Scenario: 开关关闭则不展示

- **WHEN** 这份问卷 `enabled=0`，且当前 unix 秒落在其起止之内
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 超出上海时区窗口不展示

- **WHEN** 这份问卷开关为开，且当前 unix 秒早于 `start_at` 或晚于 `end_at`
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 不能靠新一轮让已提交用户再弹

- **WHEN** 运维改选项文案或重新保存开关与有效期
- **THEN** `valid=1` 的 wxId MUST 仍然不展示
- **AND** 系统 MUST NOT 再创建第二份活动

## ADDED Requirements

### Requirement: Hub 可修改提交是否有效

Hub MUST 能把某一条提交标为有效或无效，并且该操作 MUST 立即写入 device 库。标为无效 MUST NOT 删除该行。刷新页面后 MUST 仍显示这次设置的有效状态。

#### Scenario: 标成无效后仍留在列表

- **WHEN** 运维把一条有效提交改为无效并刷新页面
- **THEN** 该行 MUST 仍在，且显示为无效
