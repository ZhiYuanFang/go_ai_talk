## 1. 回退推送门禁（对齐 3.1.9）

- [x] 1.1 删除 voice `predict_imminent` 中对 `RemotePredictCapabilityAccess` 的调用
- [x] 1.2 删除 history `RefreshBabyLiveActivity` 中 access / `filterLiveActivityTimingOnly` 剥离逻辑，恢复直接 upsert/end
- [x] 1.3 删除 cash `predict_capability_access.go`、internal access API、`api/v1` 对应类型、controller 方法、`clients/cash` RemotePredictCapabilityAccess

## 2. 删除基础版，收敛 Pro 壳

- [x] 2.1 移除 `FeatureIDPredictCapabilityBasic` / basic SKU 常量与 `InviteOncePerUser`/`IsPredictCapabilityFeature` 中的 basic；Ensure 停止种子 basic，并对已有 basic 行 `status=0`
- [x] 2.2 `StartFeatureTrial` 仅允许 Pro；删除 basic 规则文案分支
- [x] 2.3 确认 Pro 支付/邀请/试用 + `predict_pro_entry` 校验 + eligibility API 仍可用；eligibility **未**写入 maintenance_skip

## 3. 保留域卫生（不回退）

- [x] 3.1 确认广告 Activate/Hub/catalog 清理仍在
- [x] 3.2 确认 catalog 仍返回 `status`（含下架）

## 4. 收尾

- [x] 4.1 `go build` 相关包；跑 service-import 检查
- [x] 4.2 回复重部署：`cash-service`、`voice-service`、`history-service`；gateway 不必；Hub Ctrl+F5
