## Context

- Flutter twin：`predict-list-accuracy-bars` + 纯函数 `trend_predict_accuracy.dart`（`kPredictAccuracyHitWindowSec=1800`）。预测竖列表大卡图区：近 7 日准确率柱；点柱同卡换「预测小结」。
- Go：`history` 表已有可空 `expect_start_time` 等（`history-expect-fields-go`）；Hub `history.html` 已有「预测上报」Redis 卡，**不是**准确率。
- 权威数据在 history MySQL；运维低频读，默认不加准确率 Redis 缓存。

## Goals / Non-Goals

**Goals:**

- Hub 在 `history.html` 展示仿竖列表：按事件根纵向卡 + 近 7 日准确率柱 + 日小结。
- 服务端计分与 Flutter 一致，避免 JS 双份规则漂移。
- 宽屏左右分栏（卡列表 | 小结）；符合 Hub 近全宽约定。

**Non-Goals:**

- 改 Flutter / App 契约；计 duration/amount；重放伪造 expect；新后台 ticker；准确率 Redis 缓存。
- 替换或合并「预测上报」Redis 卡。

## Decisions

1. **计分（对齐 Flutter，服务端单点）**  
   - 合格：`expect_start_time > 0`；实际开始取 `start_time`（>0）；否则该行跳过（Hub 无 Flutter `createdAt` 回退亦可：无 start 则跳过，与「须有实际开始」一致）。  
   - 命中：`abs(actualStart − expectStartTime) ≤ 1800`。  
   - 归日：按实际开始在 **`Asia/Shanghai`** 的自然日（Hub 固定时区；design 注明与手机本地可能偏差，运维以服务端为准）。  
   - 日 %：`round(100 * hits / n)`；`n=0` → 该日 `percent` 省略/null。  
   - 窗口：上海日历今天往前共 7 日（含今天）。  
   - 排除：无 expect*；不引入回忆种子（Hub 无此类行）。

2. **聚合键：事件根**  
   - 优先经 `clients/device` 事件目录把 history.`event_id` 映射到 root；同根合并计分与展示。  
   - 目录不可用时回退：按 `event_id` 自身为卡键，`event_name` 取窗内最近一条。  
   - 仅返回窗内**至少有一条合格样本**的根卡（全空则空列表 + 空态文案）；避免刷出全目录空卡。

3. **API**  
   - `GET /device/admin/api/history/predict-accuracy?deviceNo=`  
     返回：`timezone`、`days[]`（7 个日键 unix 日始或 `YYYY-MM-DD`）、`cards[]`：`{ rootEventId, title, dailyPercents: (int|null)[7], }`。  
   - `GET .../predict-accuracy/day?deviceNo=&rootEventId=&day=`  
     返回该根该日样本：`expectStartTime`、`actualStartTime`、`hit`、可选 `eventName`/`historyId`。  
   - 鉴权同既有 history Admin；只读；最少 DB：一次时间窗 list（或 filter 等价）内存聚合。

4. **Hub UI**  
   - 新卡片「预测准确率」（与「预测上报」分开）。  
   - 左：纵向卡列表，每卡简易 7 柱 + 顶 `%`；右：选中日小结表（预计/实际/命中）+「近7日」清除选中。  
   - 点空日无操作；柱区点击不误触其它控件。  
   - 纯 CSS/JS，无新前端框架；可用简单 div 柱，不必上 Chart.js。

5. **进程边界**  
   - 实现在 **history-service**；禁止 voice 直查 history 表。gateway 静态页随 gateway-app 部署。

## Risks / Trade-offs

- [时区与 App 不完全一致] → 固定 Asia/Shanghai 并在卡上展示；若需宝宝本地再开 follow-up。  
- [根映射失败] → 按 leaf eventId 展示，数字仍按同 id 聚合，可能与 App 根卡切分不同。  
- [打戳前冷数据] → 空态「近 7 日暂无预测对照」。  
- [大 history] → 仅拉 7 日窗；limit 对齐既有 filter 上限并文档说明截断风险。

## Migration Plan

1. 部署 history-service（API）+ gateway-app（`history.html`）。  
2. 回滚：隐藏准确率卡 / 下线 Admin 路由即可。

## Open Questions

- 无（时区已定 Asia/Shanghai）。
