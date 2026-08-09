## 1. 请求体契约

- [x] 1.1 将 `CareAlertAnalyzeRequest` 改为 `Model *PythonModelCfg \`json:"model,omitempty"\``；移除 `model` 字符串字段与 `ModelCfg`/`model_cfg`
- [x] 1.2 更新结构体中文注释：说明与 `python_ai_talk` `CareAlertAnalyzeRequest.model` 对齐，nil=omit 走保底

## 2. 生成路径接线

- [x] 2.1 `careAlertDailyGenerate` 调用 `CareAlertAnalyze` 时赋值 `Model: modelCfg`（不再写 `ModelCfg`）
- [x] 2.2 确认日志仍能从 `modelCfg` 打印模型名；premium 路径下非空

## 3. 文档

- [x] 3.1 更新 `openspec/changes/llm-care-alert-daily/CONTRACT.md` Go→Python 小节：删除「model 简写 + model_cfg」，改为 `model` 对象或 omit

## 4. 联调验收（手工）

- [ ] 4.1 有额度/VIP 账号清当日 care-alert 日缓存后拉 daily，Python 日志 `provider`/`name` 非 `(fallback-only)`
- [ ] 4.2 无额度且 free 空时请求省略 `model`，分析仍成功（保底序）
