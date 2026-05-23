## 1. Voice 域

- [x] 1.1 `ListQaPage`（id 倒序、默认 pageSize 10、最大 100）
- [x] 1.2 `DeleteQa` 与 `VoiceQaInternalCtrl` List/Delete

## 2. Device 域与契约

- [x] 2.1 `QaPageResult`、`DeviceAdminContract.ListQAPage` / `DeleteQA`
- [x] 2.2 `voice_upstream_qa.go` 分页拉取与删除 HTTP
- [x] 2.3 `admin.go`、`admin_http_client.go` 实现

## 3. API 与控制器

- [x] 3.1 `device_admin_http.go`、`voice_qa_internal_http.go` 分页与删除类型
- [x] 3.2 `device_admin.go` QaList/QaDelete；`device_internal_handlers` 改用 `ListQAPage`

## 4. 前端与路由

- [x] 4.1 `admin.html`：标题「问答库」、预览 10 条、「展开更多」
- [x] 4.2 `qa-records.html` 全量页（分页 + 删除）
- [x] 4.3 `register.go`、`gateway_app_register.go` 静态路由

## 5. 验收

- [x] 5.1 `go build` controller / device / voice 通过
