## Context

`POST /device/app/api/survey/submit` 现在收 `optionIds`，`surveyOptionTexts` 按 id 读取启用选项原文写入 `option_snapshot`。评价发生在客户端，id 列表带不走「未使用 / 不好用 / 好用」。

`survey-response-validity` 已经规定：有效行挡住再次提交，无效行再交卷时覆盖同一行并恢复 `valid=1`。这次只换快照的来源，不改有效标记和 Hub 列表。

## Goals / Non-Goals

**Goals:**

- 提交体改为 `optionResults []string`。
- 每条结果必须等于某条启用选项原文直接加上三个后缀之一。
- 通过校验后按提交顺序把这些字符串写入 `option_snapshot`。

**Non-Goals:**

- 不改 current 响应，不新增评价枚举字段。
- 不改 `option_snapshot` 列，不要求 Hub 拆出评价。
- 不要求一次提交覆盖每一个启用选项。
- 不改网关路径和 usage 排除。

## Decisions

### D1：结果字符串替代选项 id

- **选择**：`DeviceAppSurveySubmitReq` 删除 `optionIds`，增加 `optionResults`。控制器把该数组传给 `SurveySubmit`。请求里残留的 `optionIds` 不参与写库。
- **理由**：快照要保存的是用户看到的「原文 + 评价」，不是选项 id。
- **备选**：同时保留 id 和 rating → 客户端还要和维护中的 id 对齐，已明确改为只传结果字符串。

### D2：用启用选项原文加后缀验收

- **选择**：后缀只允许 `未使用`、`不好用`、`好用`，与原文直接连接，无分隔符。一条结果在 trim 后为空、与其它条重复、条数大于 40，或码位大于 203（选项上限 200 加最长后缀 3）时整单拒绝。否则必须至少能对上一条启用且文案非空的选项。多个选项都能对上时仍然接受，快照保存客户端原文。
- **理由**：挡住把意见或无关长文写进快照，同时允许罕见的后缀歧义，因为 Hub 展示的就是这条字符串。
- **备选**：不核对选项、来什么存什么 → 快照变成第二条意见通道，拒绝。

### D3：不再回表替换快照

- **选择**：校验通过后 `json.Marshal(optionResults)` 写入 `option_snapshot`。提交路径不再调用 `surveyOptionTexts` 生成快照。
- **理由**：回表会丢掉客户端拼上的评价。
- **备选**：服务端自己拼接 → 与「客户端拼、接口收结果字符串」相反，拒绝。

### D4：空结果与覆盖规则不变

- **选择**：`optionResults` 为空且意见 trim 后为空，拒绝且不改行。只有意见、或只有合法结果的子集，都允许。`valid=1` 的重复提交保持原行并成功。`valid=0` 时用本次结果覆盖并恢复有效。活动不可提交时拒绝插入和覆盖。
- **理由**：延续只写意见和无效后再答。Flutter 会在有选项时要求每行都评，服务端不代替客户端做这项必填。

## Risks / Trade-offs

- [旧 App 仍传 `optionIds`] → 这些 id 被忽略；只写意见会落成空快照。与 Flutter `survey-option-result-submit` 一起发布。
- [Hub 在用户打开 sheet 后改了选项文案] → 结果对不上，提交被拒绝，不留下半行。
- [已有快照没有评价后缀] → 不回填。Hub 继续按字符串展示。

## Migration Plan

无表结构变更。先部署能接受 `optionResults` 的 device-service，再发拼接评价的 App。回退提交字段时，已写入的带后缀快照保持原字符串。

## Open Questions

无。
