## MODIFIED Requirements

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
