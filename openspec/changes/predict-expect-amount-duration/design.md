## Context

预测临近链路：`PUT /device/api/predict/imminent/pending` → Redis 整表 → 延时 MQ → consumer 对账后推送。当前项仅 `eventId`/`nextAt`/`title`；可见正文 `{昵称}要{title}了`（`predict-imminent-push-copy`）。孪生 Flutter change 将择一上报 duration 或 amount 结构化字段；本仓负责存储透传与后缀拼接。

## Goals / Non-Goals

**Goals:**

- API / Redis / MQ 载荷可选 expect 字段与 Flutter D4 对齐。
- 推送在基句后追加「约」后缀；非法或缺省时仅基句。
- 保持调度、闸门、Ack 语义不变。
- 秒→中文时长算法与 Flutter `formatTodayTotalAmount`（小时浮点 + round）等价，避免联调差一分钟。

**Non-Goals:**

- 服务端重算用量/用时；改基句为「即将…」；改 lead/去重/进行中/历史闸。
- 厂商标题「胖宝」变更。
- **不**把 expect 写入推送 `data` map（仅拼进 `alert`）；客户端深度链路由既有 `eventId`/`nextAt` 即可。

## Decisions

### D1：字段与合法 expect 矩阵（与 Flutter 对齐）

`PredictImminentPendingEventItem` / Redis item / fire payload 增加（均为可选，JSON omitempty）：

- `expectKind`：`duration` | `amount`
- `expectDurationSec`：int64，>0 时用于 duration
- `expectAmount`：float64，>0 时用于 amount（展示时整数无小数）
- `expectUnit`：string（可空；空则只拼数量）
- `expectLeafTitle`：string（可空；trim 后空则省略）

整表替换时原样写入（零值可用 omitempty 不落盘）；**推送拼接时**按下列矩阵判定「合法」，非法则忽略整段 expect、仍发基句（**不**因此拒绝整表同步）：

| 条件 | 行为 |
|------|------|
| `expectKind` 非 `duration`/`amount`（空、未知） | 忽略 expect |
| `kind=duration` 且 `expectDurationSec>0` | 用 duration 后缀；同条附带的 amount/unit/leaf **忽略** |
| `kind=duration` 且 sec≤0 或缺失 | 忽略 expect |
| `kind=amount` 且 `expectAmount>0` | 用 amount 后缀；unit/leaf trim 后非空才拼接；同条 durationSec **忽略** |
| `kind=amount` 且 amount≤0 或缺失 | 忽略 expect |

### D2：MQ 必须带 expect

延时消息体今日已带 `title`；同步扩展 expect，避免消费时再查客户端。Redis 对账仍以 `eventId`+`nextAt` 为准；expect 不参与匹配。

### D3：后缀模板与时长算法（统一「约」）

```
alert = nick + "要" + title + "了"
if valid expect:
  alert += suffix
```

**duration**（`kind=duration` 且 sec>0）：

`，预计这次约` + `formatExpectDurationCN(sec)`，算法等价 Flutter 今日汇总（先把秒换成小时浮点再 round）：

```
hours = sec / 3600.0
若 hours < 1/60          →「不足1分钟」
若 hours < 1             → m = round(hours*60)；m<1 →「不足1分钟」；否则「{m}分钟」
否则 h = floor(hours)，m = round((hours-h)*60)
  m==0 →「{h}小时」；否则「{h}小时{m}分钟」
```

参考边界：`59`→不足1分钟；`90`→2分钟；`3600`→1小时；`3900`→1小时5分钟；`3661`→1小时1分钟。

**amount**（`kind=amount` 且 amount>0）：

`，预计这次约` + 格式化数量（整数无小数；否则去尾零）+（非空）unit +（非空）leafTitle，**无空格**：

- 例：`，预计这次约120毫升配方奶`
- 无叶子：`，预计这次约120毫升`
- 无单位：`，预计这次约120` 或 `，预计这次约120配方奶`

不使用「需要」措辞。

**leafTitle**：trim 后截断至 **32** 个 Unicode 字符（契约，防推送体膨胀）。

### D4：兼容与同路径扩展

- 旧客户端不传字段 → 仅基句。
- 部署：Go 先于或同发 App；新客户端 + 旧二进制时字段被忽略。
- **API 版本豁免**：本接口仅追加可选 omitempty 字段，缺省行为与现网一致（非 BREAKING）。允许在既有 `PUT /device/api/predict/imminent/pending` 同路径扩展结构体，**不**新建 v2；与仓库「破坏性改结构须升版」区分——此处为可选加字段。

### D5：联调可观测

`sync_recv` 事件摘要宜附带 expect 摘要（如 `kind` + durationSec 或 amount），便于 grep 核对客户端是否上报；不改变对外 API。

## Risks / Trade-offs

- [时长格式与客户端细微差异] → D3 算法锁定 round；可接受极端浮点边界的分钟级差。
- [恶意超长 leafTitle] → D3 截断 32。
- [amount 浮点] → 展示去掉无意义小数。

## Migration Plan

1. 上线 voice（含 API 结构体）；无需迁 Redis 旧数据（下次客户端全量替换覆盖）。
2. 与 Flutter 联调后再放量 App。
3. 回滚：忽略 expect 拼接即可；旧字段路径不变。

## Open Questions

（已在探索中拍板并写入 D1–D5；无阻塞项。）
