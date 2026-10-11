## 1. Voice Admin API

- [x] 1.1 `api/v1`：`VoiceAdminPredictImminentPendingGetReq/Res`、`VoiceAdminPredictActiveEtaGetReq/Res`（路径见 design D1；复用既有 event/item 结构体）
- [x] 1.2 voice Admin controller：`VerifyVoiceAdminPassword` + 调用 `LoadPredictImminentPending` / `LoadPredictActiveEta`；映射完整字段；空列表语义
- [x] 1.3 `register_voice_service.go` 注册 Admin 控制器（非 InternalSecret 组）

## 2. Hub 静态页

- [x] 2.1 `history.html`：宝宝档案与历史事件之间增加「预测上报」一张卡两段表格（字段完整列）
- [x] 2.2 刷新 + 进入页自动 `AdminCommon.adminFetch` 双 GET；空列表提示；时间可读展示（可附 unix）

## 3. 验收

- [ ] 3.1 App PUT imminent / active-eta 后，Hub 卡片两段字段对齐；清空上报后两段为空提示（部署后手工）
- [x] 3.2 确认无新 env；部署清单含 voice-service + gateway-app（静态）
