## ADDED Requirements

### Requirement: 不得再暴露 clinic/tip HTTP 飞轮

voice-service MUST NOT 再 Bind 暴露 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`。`DeviceClinicFeedbackController`（或仅承载上述两接口的等价控制器）MUST 删除。gateway-app MUST NOT 再为 `/device/api/clinic/*`、`/device/api/tip/*` 配置指向 voice 的反代（若该 pattern 仅服务上述反馈）。

#### Scenario: clinic feedback 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/api/clinic/feedback`
- **THEN** 请求 MUST NOT 由 clinic feedback 控制器处理

#### Scenario: tip feedback 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/api/tip/feedback`
- **THEN** 请求 MUST NOT 由 tip feedback 控制器处理

#### Scenario: gateway 反代清理

- **WHEN** 审查 `installVoiceProxyMiddleware`（或等价）绑定的 path 列表
- **THEN** MUST NOT 包含仅服务于 tip SSE、tip feedback、clinic feedback 的 `/device/tip/*`、`/device/api/tip/*`、`/device/api/clinic/*` pattern

### Requirement: care-alert 飞轮必须保留

本变更 MUST NOT 删除或下线 `POST /device/api/care-alert/feedback`。care-alert daily / delete item / feedback 仍由 **voice-service** 宿主处理，gateway MUST 继续将 `/device/api/care-alert/*` 反代至 voice。

#### Scenario: care-alert feedback 仍注册

- **WHEN** 审查 voice-service HTTP 注册与 care-alert 控制器
- **THEN** `POST /device/api/care-alert/feedback` MUST 仍由 care-alert 控制器处理，且服务层 `CareAlertFeedback`（及 Python CareAlertFeedback 客户端若原先存在）MUST 仍可用

#### Scenario: care-alert 反代仍在

- **WHEN** 审查 gateway voice 域反代 path 列表
- **THEN** MUST 仍包含 `/device/api/care-alert/*`

### Requirement: clinic WS 不受本变更影响

本变更 MUST NOT 移除或改坏 `/voice/clinic/ws`（及等价 clinic 连续对话 WebSocket）注册与反代。

#### Scenario: clinic WS 仍可达

- **WHEN** 审查 gateway WS 反代与 voice clinic WS 控制器
- **THEN** clinic WebSocket 路径 MUST 仍注册且可被反代

### Requirement: usage skip 与已删路径一致

`maintenance_skip.go`（或等价）MUST NOT 再保留已删除的 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback` 排除项。MUST NOT 因本变更新增对 care-alert 路径的排除（除非负责人另行确认；本变更默认不动 care-alert usage 策略）。

#### Scenario: skip 列表无死路径

- **WHEN** 审查 usage 维护型排除列表
- **THEN** MUST NOT 出现已下线的 clinic/tip feedback 精确 path 条目

## REMOVED Requirements

### Requirement: voice-service SHALL Bind DeviceClinicFeedbackController 暴露 clinic/tip feedback（基线 v3.0.0）

**Reason**：喂养/诊疗反馈由连续对话消化；Flutter 已无 HTTP 飞轮调用。  
**Migration**：删除 HTTP 接口；无替代点赞 API。care-alert 飞轮不在本 REMOVED 范围内。

### Requirement: clinic/tip feedback MUST NOT 计入 usage 且路径保持（基线 v3.0.0 用法条款）

**Reason**：路径已删除，usage 排除项失去意义。  
**Migration**：从 skip 列表删除对应条目；无需新排除。
