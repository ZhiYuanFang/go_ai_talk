## 1. DDL 与 device 域模型

- [x] 1.1 新增 `hack/ddl_wx_is_vip.sql`：`ALTER TABLE wx ADD COLUMN is_vip TINYINT NOT NULL DEFAULT 0`（及必要注释）
- [x] 1.2 同步 `entity.Wx` / `dao`（`internal/dao/internal/wx.go` columns）增加 `IsVip` / `is_vip`
- [x] 1.3 在 `docs/runbooks/release-deploy-and-run.md`（或现有 wx DDL 段落旁）补充执行该 DDL 的发布提示

## 2. device-service VIP 读契约

- [x] 2.1 新增 `api/v1` 内部接口：`GET /device/app/api/user/internal/vip-by-wx-id`（query `wxId`，内部密钥鉴权）
- [x] 2.2 device service/controller：按 `wxId` 查 `wx.is_vip`；无行返回 `isVip=false`；缺密钥拒绝
- [x] 2.3 确认路由挂在 device-service 注册路径上（与现有 user internal 一致）

## 3. voice 客户端与选模

- [x] 3.1 在 `device` 包增加 `RemoteIsVipByWxID`（经 `userInternalHTTP`，与 `RemoteWxIDByDeviceNo` 同风格）
- [x] 3.2 将 `isAccountVIP(ctx, wxID int64) bool` 改为调 Remote；失败/超时打 Warning 并返回 false（降级）
- [x] 3.3 `resolveCareAlertModelProfile` / 生成日志改为使用 `wxID`，不再传 `deviceNo` 作 VIP 键

## 4. care-alert 强制 wxId

- [x] 4.1 `DeviceCareAlertController`（或 service 入口）解析 `X-Internal-Wx-Id`；`wxId<=0` 拒绝三条 API
- [x] 4.2 `CareAlertDaily` 生成路径传入触发者 `wxId`；禁止 care-alert 使用 `ResolveVoiceWxID` 的 deviceNo fallback
- [x] 4.3 DELETE / feedback 同样强制 `wxId`（即使不读 VIP）

## 5. 契约文档与自检

- [x] 5.1 更新 `openspec/changes/llm-care-alert-daily/CONTRACT.md`：VIP=`wx.id`、强制 Header、查失败降级、去掉「恒 false」
- [x] 5.2 评审自检：`rg 'internal/dao' internal/services/voice` 无新增 wx 直查；care-alert 路径无 deviceNo→VIP
- [x] 5.3 gateway-app 自检：反代仍覆盖 `/device/api/care-alert/*`；care-alert **不**加入 Bearer 白名单（须登录）；usage 统计策略若未获负责人答复则不改 `maintenance_skip.go`
