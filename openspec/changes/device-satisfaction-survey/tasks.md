## 1. 表与领域模型

- [x] 1.1 device 启动 schema：幂等创建 `survey_campaign`、`survey_option`、`survey_response`（唯一键 campaign_id+wx_id），中文注释写明开关、unix 秒、快照语义
- [x] 1.2 生成或手写 DAO/entity/do，查询无行路径使用 `.One()` + `IsEmpty()`

## 2. App API

- [x] 2.1 `api/v1` 新增 `GET /device/app/api/survey/current` 与 `POST /device/app/api/survey/submit` 的 `g.Meta`，不改既有请求结构
- [x] 2.2 实现 current：只看最新活动；关或超出 `[start_at,end_at]` 或不存在或该 wxId 已提交则 `show=false`；否则返回说明与启用选项
- [x] 2.3 实现 submit：双空拒绝且不写行；快照选项文案、会话 deviceNo 的宝宝昵称与设备号；重复提交保持单行成功；活动不可展示时拒绝新插入
- [x] 2.4 controller 挂在 device 进程，中文业务注释；不 import 他域业务包，不直查他库

## 3. 网关与统计

- [x] 3.1 `installDeviceProxyMiddleware` 增加 `/device/app/api/survey/*`
- [x] 3.2 确认两条路径不进 Bearer 匿名白名单
- [x] 3.3 按已确认结论，在 `maintenance_skip.go` 精确排除 `GET /device/app/api/survey/current` 与 `POST /device/app/api/survey/submit`，注释写明负责人确认不统计

## 4. Hub

- [x] 4.1 admin API：读写最新活动（开关、上海时区起止、说明、选项行）、「开启新一轮」、按活动分页列出提交（wxId、宝宝昵称、设备号、选项快照、意见、时间）
- [x] 4.2 静态页 `survey-admin.html`：加载 `admin-pages.css`，宽屏 `.hub-md` 左编辑右结果；登记 `admin_static_pages.go` 与 `admin-modules.js` 导航
- [x] 4.3 自检：反代前缀、auth 不豁免、usage 排除、`g.Meta` summary、近全宽左右分栏
