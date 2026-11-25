# Tasks

## 1. 数据模型与迁移

- [x] 1.1 新增 `merchant`、商户-宝宝归属、`wx_baby_member`、`merchant_entitlement`、商户订阅商品表（或等价）DDL，并为 `wx` 增加 `role`；用空库/测试库执行迁移脚本验证可重复执行
- [x] 1.2 编写回填：非空 `wx.device_no` → 成员表一行、`role=parent` 默认；在测试数据上跑回填后抽查成员数与焦点一致
- [x] 1.3 种子商户连续订阅商品默认 `price_fen=100000`（1000 元/月）；验证读价 API/查询返回该默认

## 2. 身份与多宝宝（device-service）

- [x] 2.1 实现商户注册/登录（店名+密码哈希）及商户会话所需数据访问；用接口测试验证注册冲突与登录成功
- [x] 2.2 实现 `username/register` 商户创建子账户路径：`role=staff`；验证 parent→staff 变更被拒
- [x] 2.3 实现成员表 CRUD：列表宝宝、加成员、切焦点（更新 `wx.device_no`）；验证 `detail`/登录仍返回标量焦点 `deviceNo`
- [x] 2.4 实现 parent 创建虚拟宝宝与绑定（preview+bind）多成员；验证旧式「已有成员再绑其它号」被拒
- [x] 2.5 实现商户创建虚拟宝宝（额度校验调用权益）、画像后台/App 双写；验证无实体设备可登记
- [x] 2.6 实现商户分发/解绑 staff↔宝宝；验证 App 解绑 API 拒绝；分发时可设 staff 默认焦点
- [x] 2.7 补充 device 域单测/接口测覆盖 role 互斥与成员焦点；确认关键用例通过

## 3. 商户权益与支付（cash-service）

- [x] 3.1 实现 `merchant_entitlement` 读写与 `IsMerchantEntitled(expire_at)`；验证过期/有效边界
- [x] 3.2 实现 Hub Admin API：按 `merchantId` 设置**绝对** `expireAt`；验证非法过去时间拒绝（或与规格一致）
- [x] 3.3 实现商户商品月费 Hub 更新与商户侧读价；验证改价后读到新分值
- [x] 3.4 实现支付宝网页连续订阅建单/回调履约：幂等更新同一 `expire_at`，不写 `vip_entitlement`；用模拟通知测首次与重复回调
- [x] 3.5 实现内部探活「device_no 是否因商户欠费禁 create」供 history 调用；验证商户欠费 true、普通宝宝 false

## 4. History 写闸

- [x] 4.1 在 create/batch create 路径接入欠费探活；欠费拒绝插入；update/end/list 不受影响；用集成测覆盖三条路径
- [x] 4.2 确认 voice/Python batch create 同样被闸；用一条 batch create 在欠费商户宝宝上验证失败

## 5. Gateway：商户 JWT 与 SSE

- [x] 5.1 签发/校验商户 JWT，商户路由注入 `X-Internal-Merchant-Id`；验证 wx 令牌不能访问商户 API、反之亦然
- [x] 5.2 实现商户多 device SSE（复用历史 Redis 通道+集合过滤+baby 标识）；验证名下 create 推送、名外不推；禁缓冲路径限定
- [x] 5.3 注册商户 SPA 静态/反代入口与 API 前缀；验证 SPA 壳可匿名加载、API 仍需商户 JWT

## 6. Hub 运维页

- [x] 6.1 新增商户运维静态页：列表（开通/来源/expire_at/欠费）、选商户绝对授期、月费编辑；用 AdminCommon；浏览器不带 Admin 密码头
- [x] 6.2 在 `admin-modules.js` / 静态注册表登记模块；登录 Hub 后可导航并完成一次授期冒烟

## 7. 商户 Vue SPA

- [x] 7.1 脚手架 Vue 商户台（登录/注册）；对接商户 JWT；验证错误店名密码提示
- [x] 7.2 主面板玻璃拟态宝宝卡片（名/最近记录/下一次预测）+ 商户 SSE 局部刷新；用 mock/联调验证单卡更新
- [x] 7.3 右侧详情：按事件行展示预测，准确率图↔趋势图切换（无事件 Tab）；对照 Flutter 预测/趋势可接受对齐
- [x] 7.4 子账户创建、宝宝创建、分发/解绑、画像编辑、订阅开通页（网页支付宝）；完成主路径手工验收清单

## 8. Flutter 同发（兄弟仓 flutter_ai_talk）

- [x] 8.1 宝宝列表与切换焦点 UI；切换后历史/SSE/预测跟焦点；验证多宝宝可切换
- [x] 8.2 parent：App 内创建虚拟宝宝 + 绑定流程（可多次加入）；确认无解绑入口
- [x] 8.3 staff：展示全部分发宝宝；与旧版「只见焦点」兼容说明写入发布说明
- [x] 8.4 回归旧路径：无宝宝→绑定一个仍可用；自动化或手工用例通过

## 9. 部署与集成验收

- [x] 9.1 更新 Compose/环境变量/网关路由/runbook；`openspec` 域表：归档前准备将 `merchant-console` 写入 `project.md` 域列表
- [ ] 9.2 端到端：商户注册→建 2 宝宝→建 staff→分发→App 记事件→商户 SSE 刷新；Hub 授期→第三宝宝；欠费 create 失败、end 成功
- [ ] 9.3 端到端：支付宝连续订阅沙箱（或录制夹具）延长 `expire_at`；与 Hub 授期共用字段抽查

## Workflow follow-up

- 实现完成后归档本 change：合并 delta 至主规格（含新建 `openspec/specs/merchant-console/spec.md`），更新 `project.md` 域表，删除 change 目录。
- Flutter 侧变更在兄弟仓单独 PR，与本仓同发窗口对齐。
