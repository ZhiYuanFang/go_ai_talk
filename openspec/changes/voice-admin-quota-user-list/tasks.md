## 1. Device 身份契约（babyName）

- [x] 1.1 扩展 `AdminWxListItem` / `ListWxPage`（及 HTTP 客户端映射）返回 `babyName`（按 deviceNo 联查 user 画像；无则空串）
- [x] 1.2 确认 device admin `wx/list` 响应字段与契约一致（若对外 admin JSON 同步暴露 babyName）

## 2. Voice 列表 API

- [x] 2.1 在 `api/v1/voice_ai_quota_http.go` 新增 `GET /voice/admin/api/ai-quota/users` 请求/响应 DTO（page/pageSize/deviceNo；list 含 deviceNo、wxId、account、babyName、voiceAi/clinicAi 的 used/limit）
- [x] 2.2 实现 store/service：经 `DeviceAdmin().ListWxPage` 取页；批量解析有效 limit（default + override）与当月 used（cachekit 既有 usage 键）
- [x] 2.3 Controller 挂载 users 列表，复用 `X-Admin-Password` 校验；注册到 voice-service
- [x] 2.4 PUT user：当提交上限等于当前全局默认时清除该 feature override（与 design/spec 一致）

## 3. 运维 UI

- [x] 3.1 `voice-admin.html`：删除单人 override 模块；保留全局默认与 ai-model-admin 链接
- [x] 3.2 新增用户额度表：列顺序 deviceNo → 喂养已用/上限 → 胖宝已用/上限 → wxId → account → babyName；分页 + deviceNo 查询
- [x] 3.3 行内保存上限调用 `PUT /voice/admin/api/ai-quota/user`；保存后刷新列表
- [x] 3.4 页内说明列出喂养/胖宝受影响业务接口（与 proposal Impact 一致）

## 4. 校验与自检

- [x] 4.1 确认 `/voice/admin/api/*` 网关反代已覆盖新 path；无新增 App HTTP / 无需改 usage denylist
- [x] 4.2 `openspec validate voice-admin-quota-user-list --strict` 通过；评审 grep 无 voice 直查他域 DAO
