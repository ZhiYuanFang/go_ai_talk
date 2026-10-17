## MODIFIED Requirements

### Requirement: 当前活动可配置且仅最新活动对 App 生效

device-service MUST 只维护一份满意度问卷活动及其选项行。App 读取 MUST 只针对这一份。Hub MUST 能编辑该份的开关、说明、Asia/Shanghai 开始与结束时间。Hub MUST NOT 提供「开启新一轮」或按活动切换。就地修改选项文案 MUST NOT 清除已有提交，也 MUST NOT 使已提交的 wxId 再次变为可展示。时间 MUST 以 unix 秒存储。保存开关与起止时，若 `end_at` 早于 `start_at` 或缺少起止，MUST 拒绝该次设置保存。选项行的新增与修改 MUST NOT 因起止无效而失败。

#### Scenario: 开关关闭则不展示

- **WHEN** 这份问卷 `enabled=0`，且当前 unix 秒落在其起止之内
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 超出上海时区窗口不展示

- **WHEN** 这份问卷开关为开，且当前 unix 秒早于 `start_at` 或晚于 `end_at`
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 不能靠新一轮让已提交用户再弹

- **WHEN** 运维改选项文案或重新保存开关与有效期
- **THEN** 已有提交的 wxId MUST 仍然不展示
- **AND** 系统 MUST NOT 再创建第二份活动

### Requirement: Hub 可查看提交明细

Hub 问卷调查页 MUST 近全宽，宽屏下编辑区与结果区左右分栏。结果 MUST 展示用户 id、宝宝昵称、设备号、选项快照、意见与提交时间，并且 MUST 列出这一份问卷的全部提交，不得按活动分表或筛选。

#### Scenario: 查看某次提交

- **WHEN** 运维打开问卷调查页且已有提交
- **THEN** 页面 MUST 列出该条的 wxId、宝宝昵称、设备号、所选文案与意见

## ADDED Requirements

### Requirement: 选项行可单独编辑并立即持久化

Hub MUST 能新增一行选项并立即写入 device 库，随后 MUST 能修改该行文案、是否启用，以及删除该行，且每次操作立即持久化。刷新页面后这些行 MUST 仍在。空文案行 MAY 存在，但 App current MUST NOT 把它当作可选项返回。删除选项 MUST NOT 改写已有提交的文案快照。

#### Scenario: 新增后刷新仍在

- **WHEN** 运维新增一行并写入文案，然后重新打开问卷调查页
- **THEN** 该行 MUST 仍显示刚写入的文案

#### Scenario: 空文案不对 App 展示

- **WHEN** 存在文案为空白的选项行，且问卷对某用户需要展示
- **THEN** current 的选项列表 MUST NOT 包含该行
