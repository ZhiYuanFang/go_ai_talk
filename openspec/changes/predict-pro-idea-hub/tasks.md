## 1. 数据模型

- [x] 1.1 EnsureSchema：`predict_pro_idea` + `predict_pro_idea_like`（含唯一键与软删）
- [x] 1.2 确认 `device-service` / `ai_voice_device` / `default` / `DEVICE_DB_LINK`；必要时同步 `.env.example` 注释

## 2. App API

- [x] 2.1 `api/v1`：list / submit / like / unlike（`g.Meta` 含 path/method/summary）
- [x] 2.2 submit：JWT、绑定、宝宝名快照、≤200 字、`display_text`=原文；**不**查 Pro 权益
- [x] 2.3 submit 日限额：Asia/Shanghai 自然日、同 wx_id 对 `predict_pro_idea` **COUNT**（默认上限 10；含软删计入）
- [x] 2.4 list：未删除、created_at DESC；返回 text/babyName/likeCount/likedByMe/isMine；不下发 wxId
- [x] 2.5 like/unlike 幂等，返回 likeCount + likedByMe（任意登录用户）
- [x] 2.6 controller 注册于 device-service

## 3. gateway-app 与 usage

- [x] 3.1 `device_route_proxy`：BindMiddleware `/device/app/api/predict-pro-ideas/*`
- [x] 3.2 Admin HTML：`admin_static_pages` + `gateway_app_auth_exempt` 登记 `/device/admin/predict-pro-ideas`；App 路径 **不**进 exempt
- [x] 3.3 usage：**计入**；确认 **未**写入 `maintenance_skip.go`；apiregistry 可加载 summary

## 4. Admin API 与 HUB

- [x] 4.1 Admin list/update/delete（口令；list 含 wxId、babyName、likeCount）
- [x] 4.2 `predict-pro-ideas.html`（admin-pages 近全宽）+ `admin.html` 入口
- [x] 4.3 手工验收：submit → list → like → Admin 改文案 → 删除下线；usage 有记录
  - 静态/编译已通过；完整链路待部署后冒烟

## 5. 联调与重部署

- [x] 5.1 通知 Flutter `predict-pro-idea-wall` 契约可联调
- [x] 5.2 apply 收尾写明重部署：`device-service`、`gateway-app-server`；不必重部署 cash/voice/ucg
