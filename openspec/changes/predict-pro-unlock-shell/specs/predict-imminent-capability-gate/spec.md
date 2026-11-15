## REMOVED Requirements

### Requirement: imminent 推送与 Live Activity 写入 MUST 校验预测能力开通或 VIP

**Reason**: 产品改为预测推送全员开放（对齐 3.1.9）；开通与推送解耦。

**Migration**: 删除 cash predict-capability access、clients Remote 及 voice/history 调用；改由 `predict-push-ungate` 约束。

### Requirement: Live Activity 清屏 MUST NOT 受预测能力开通门禁

**Reason**: 门禁已撤销，清屏豁免不再作为独立商业门禁需求（空投影 end 仍按原 LA 逻辑）。

**Migration**: 保持 Refresh 空投影 end；无需再为「无权益」单独分支。

### Requirement: timing Live Activity MUST NOT 受预测能力开通门禁

**Reason**: 门禁已撤销；timing 本就不被商业门禁约束，无需再并列声明。

**Migration**: 无。
