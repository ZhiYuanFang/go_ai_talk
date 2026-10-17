## Why

Hub 问卷调查把选项、开关和「第几轮活动」捆在一次整表保存里，新增文案不点保存就会丢，结果还要按活动筛选。运维只需要一份问卷、一份提交统计，每一行文案能改并且马上留下。

## What Changes

- 全库只保留一份满意度问卷。去掉「开启新一轮」和活动下拉。App 与 Hub 都只看这一份。
- 选项改为逐行新增、修改文案、启用或删除，每次操作立即落库，不依赖整页保存。
- 开关、上海时区起止和说明仍单独保存。关掉或不在有效期内不展示。起止无效时只拒绝这份设置，不挡住选项行保存。
- 提交列表不再按活动拆开。同一 wxId 在这一份问卷上仍只记一次；改文案不让已提交用户重新弹出。选项快照语义不变。
- App 的 `GET /device/app/api/survey/current` 与 `POST /device/app/api/survey/submit` 路径与「不计入使用统计」不变。Flutter 契约不改。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `device-satisfaction-survey`：取消多轮活动；Hub 改为单问卷逐行编辑；提交统计合并为一份。

## Impact

- **device-service**：问卷读写改为确保单行活动；新增选项行的增改删；提交查询与「是否已答」不再按多活动过滤。已存在的多行活动在启动或首次读写时并入保留的那一份。
- **Hub**：`resource/public/survey-admin.html` 去掉活动选择与「开启新一轮」，选项行即时保存。布局仍近全宽、左右分栏。
- **gateway-app / usage**：App 路径、反代、匿名白名单、`maintenance_skip.go` 排除项不变。
- **Flutter**：不改。
