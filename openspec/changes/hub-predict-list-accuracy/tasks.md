## 1. 计分与 Admin API

- [x] 1.1 实现与 Flutter 对齐的纯函数/服务：1800s 命中、`expect_start_time>0`、实际 `start_time>0`、Asia/Shanghai 归日、日 `round(100*hits/n)`、无样本→null；禁止重放
- [x] 1.2 Admin `GET .../predict-accuracy`：近 7 日 + 按根（目录可得）/eventId 卡片；只读；一次时间窗查询内存聚合
- [x] 1.3 Admin `GET .../predict-accuracy/day`：返回预计/实际开始与 hit 列表；空日空数组
- [x] 1.4 注册 history Admin 路由与鉴权；确认 gateway `/device/admin/api/history/*` 透传

## 2. Hub history.html

- [x] 2.1 新增「预测准确率」区（与「预测上报」分离）；宽屏左右分栏；纵向事件卡 + 7 柱（null 日无柱无 0%）
- [x] 2.2 点有样本柱加载日小结（预计/实际/命中）；「近7日」返回；空日不进入；空卡空态文案
- [x] 2.3 展示时区说明（Asia/Shanghai）；刷新按钮

## 3. 验收

- [x] 3.1 有 expect* 样本：柱与 % 与手工按 1800s 规则一致；点柱见小结列
- [x] 3.2 无 expect* / 空窗：空态；预测上报卡行为不变
