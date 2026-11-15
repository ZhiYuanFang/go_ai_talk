## ADDED Requirements

### Requirement: Admin predict-pro-ideas list API SHALL paginate and expose wxId and baby name

`GET /device/admin/api/predict-pro-ideas/list` MUST require admin authentication (`X-Admin-Password` or documented equivalent). It SHALL accept `page` (default 1) and `pageSize` (default 20, max 100). Results MUST be ordered by `created_at` descending and include `list`, `total`, `page`, `pageSize`. Each item MUST expose at least `id`, `wxId`, `babyName`, `content`, `displayText`, `likeCount`, `createdAt`. Default listing SHOULD exclude soft-deleted rows. Admin 列表 **必须** 含 wxId 与宝宝名。

#### Scenario: Authenticated page

- **WHEN** 管理员请求 list 且口令正确
- **THEN** 系统 SHALL 返回分页 list，项内含 `wxId` 与 `babyName`

#### Scenario: Bad password

- **WHEN** 请求缺少有效 Admin 口令
- **THEN** 系统 MUST 拒绝

### Requirement: Admin SHALL update display text of an idea

`POST /device/admin/api/predict-pro-ideas/update` MUST accept `{ "id": <int64>, "displayText": "<text>" }`. After trim, `displayText` MUST be non-empty and MUST NOT exceed 200 characters. The server SHALL update `display_text` for a non-deleted row and MUST NOT change `content` by default. 管理员 **必须** 能改展示文案。

#### Scenario: Update display text

- **WHEN** Admin 对未删除想法提交合法 displayText
- **THEN** 系统 SHALL 更新 `display_text` 且 App list 随后返回新文案

### Requirement: Admin SHALL delete an idea

`POST /device/admin/api/predict-pro-ideas/delete` MUST accept `{ "id": <int64> }` and soft-delete the row. Subsequent App list MUST exclude it. 管理员 **必须** 能删除想法。

#### Scenario: Delete hides from App

- **WHEN** Admin 删除想法 id=N
- **THEN** App list MUST NOT 再包含 id=N

### Requirement: Admin predict-pro-ideas page SHALL mirror feedback-records pattern

The system MUST serve `resource/public/predict-pro-ideas.html` at `/device/admin/predict-pro-ideas`. The page MUST load `admin-pages.css` and follow Hub 近全宽布局；宽屏若同时存在列表与编辑区 MUST 左右分栏或等价避免整页长滚。The page MUST use admin password login, call Admin predict-pro-ideas APIs, show a paginated table including wxId and baby name, allow edit of display text and delete, and link back to `/device/admin`. `admin.html` MUST include a card entry. gateway-app MUST register the HTML in `admin_static_pages` and Bearer/auth exempt for the **HTML shell path only** (对齐 `feedback-records`). Admin **必须** 提供独立管理页与入口。

#### Scenario: Open management page

- **WHEN** 管理员打开管理页并输入正确口令
- **THEN** 页面 SHALL 展示想法表且可改文案与删除

### Requirement: This release SHALL NOT model adopted-idea Pro grants

The server MUST NOT add an「已采纳」persistence field or automatic one-month Pro entitlement grant for idea authors in this change. 本版 **不得** 实现采纳态或自动发 Pro。

#### Scenario: No adopt API

- **WHEN** 本变更交付
- **THEN** Admin/App API MUST NOT 提供采纳标记或自动履约接口作为本能力一部分
