## Why

App 要把每一项的「未使用 / 不好用 / 好用」写进选项快照。现在的提交只收选项 id，服务端按 id 回表抄原文，快照里带不走评价。改为由客户端拼好结果字符串，服务端校验后原样保存。

## What Changes

- **BREAKING**：`POST /device/app/api/survey/submit` 不再读取 `optionIds`。请求体改为 `campaignId`、`optionResults`（字符串数组）、`comment`。
- 每条结果必须等于某条启用中选项的原文，再直接接上「未使用」「不好用」或「好用」之一。通过后把这些字符串按提交顺序写入 `option_snapshot`，不再回表抄原文。
- 空串、重复、超过 40 条、单条超过 203 个 Unicode 码位，或对不上任一启用选项加合法后缀时，拒绝且不落库、不改写已有行。
- `optionResults` 与 trim 后的意见都空时仍拒绝。只写意见、或只提交一部分已启用选项的合法结果，仍然允许。不要求客户端把每一个启用选项都交上来。
- 有效标记、无效行覆盖、重复的有效提交保持原行、活动不可提交时拒绝，以及 Hub 展示快照字符串，都保持 `survey-response-validity` 的行为。GET current 的选项列表不变。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `device-satisfaction-survey`：提交从选项 id 改为选项结果字符串，快照保存客户端拼好的文案。

## Impact

- **device-service**：`api/v1/device_app_survey_http.go` 的提交结构；`SurveySubmit` 与控制器入参。`option_snapshot` 列类型不变。
- **gateway-app**：路径与 usage 排除不变。
- **Hub**：仍展示快照字符串，不解析评价后缀。
- **Flutter**：由 `flutter_ai_talk` 变更 `survey-option-result-submit` 改提交体；两边需要一起上线。
