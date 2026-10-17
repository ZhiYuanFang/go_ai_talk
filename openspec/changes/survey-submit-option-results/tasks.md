## 1. 提交入口

- [x] 1.1 `DeviceAppSurveySubmitReq` 改为 `optionResults`，控制器把该字符串数组传给 `SurveySubmit`，不再读取 `optionIds`

## 2. 校验与快照

- [x] 2.1 拒绝空串、重复、超过 40 条、单条超过 203 个码位，以及不能对上「启用选项原文 + 未使用/不好用/好用」的结果；多个选项都能对上时仍接受
- [x] 2.2 校验通过后按提交顺序把 `optionResults` 写入 `option_snapshot`，不再回表替换成不含后缀的原文。空结果加空意见仍拒绝；只写意见、合法子集、`valid=1` 重复提交保持原行、`valid=0` 覆盖并恢复有效，保持现有行为

## 3. 收尾

- [x] 3.1 更新已有 submit 调用和断言里的 `optionIds`，改为 `optionResults`；不改 current、Hub 页面和网关 usage 排除
