# Spec Delta

## Purpose

商户 B 端控制台：独立登录、虚拟宝宝配额与店员分发、玻璃拟态主面板与多宝宝 SSE 实时刷新，使门店可运营喂养宝宝而不混入 C 端会话。

## ADDED Requirements

### Requirement: 商户 SHALL 以店名密码自由注册并仅登后台
系统 MUST 提供商户注册与登录（店名+密码）。成功登录 MUST 签发 **商户 JWT**（主体为商户主键，MUST NOT 使用 wx JWT）。持商户 JWT MUST 仅可访问商户后台 API/页面；MUST NOT 用商户 JWT 调用 C 端 App 受保护接口。

#### Scenario: 注册并登录后台
- **WHEN** 新店名与合法密码完成注册并登录
- **THEN** 系统 MUST 返回商户 access 令牌且后续商户 API 校验通过

#### Scenario: 商户令牌不可登 App
- **WHEN** 客户端以商户 JWT 调用需 wx 会话的 App API
- **THEN** 系统 MUST 拒绝（401/业务错误）且 MUST NOT 注入 `X-Internal-Wx-Id`

### Requirement: 商户默认两宝宝额度与订阅后不限量
未拥有有效商户权益时，商户 MUST 最多创建 **2** 个虚拟宝宝。有效权益（`expire_at` 未过期）期间 MUST 允许创建宝宝且 MUST NOT 施加数量上限。超额创建（无权益且已有 2 个）MUST 拒绝。

#### Scenario: 免费额满拒绝
- **WHEN** 商户无有效权益且已有 2 个宝宝并请求再创建
- **THEN** 系统 MUST 拒绝创建

#### Scenario: 有效权益不限量
- **WHEN** 商户 `expire_at` 晚于当前时间且已有多于 2 个宝宝后再次创建
- **THEN** 系统 MUST 允许创建

### Requirement: 商户主账号 SHALL 分发与解绑宝宝与子账户
系统 MUST 支持宝宝与子账户（`role=staff` 的 wx）多对多绑定。仅商户主账号会话 MUST 可建立或解除该绑定。客户端 App MUST NOT 提供解绑能力。重复绑定同一对关系 MUST 幂等成功。

#### Scenario: 主账号分发
- **WHEN** 商户将会员宝宝绑定到已有 staff 子账户
- **THEN** 系统 MUST 建立成员关系且该 staff 在成员列表中可见该宝宝

#### Scenario: 主账号解绑
- **WHEN** 商户解除某 staff 与某宝宝的绑定
- **THEN** 系统 MUST 删除该成员边且 MUST NOT 删除宝宝或 history

#### Scenario: 非主账号不可解绑
- **WHEN** staff 或 parent 的 App 会话请求解绑宝宝
- **THEN** 系统 MUST 拒绝

### Requirement: 商户主面板 SHALL 展示宝宝卡片并可局部 SSE 刷新
商户后台主面板 MUST 展示名下全部宝宝卡片，每卡至少含：宝宝名、最近一次记录摘要、最近下一次预计事件。系统 MUST 提供 **单条** 商户 SSE：鉴权后按该商户宝宝 `device_no` 集合订阅历史变更通知，事件 MUST 携带宝宝标识；前端 MUST 能仅刷新对应卡片数据。

#### Scenario: SSE 扇出到商户连接
- **WHEN** 某商户名下宝宝发生 history create 且商户已建立 SSE
- **THEN** 网关 MUST 向该连接推送含该宝宝标识的事件帧

#### Scenario: 集合外不推送
- **WHEN** 非该商户名下的 device 发生 history 变更
- **THEN** MUST NOT 向该商户 SSE 推送该事件

### Requirement: 宝宝详情 SHALL 仿预测页并支持准确率与趋势切换
主面板选中宝宝后，界面 MUST 在右侧（或等价缩进区）展示该宝宝各信息事件行：事件与对应预测。用户 MUST 能在准确率图与趋势图之间切换；每行一个事件，MUST NOT 要求先切换「当前事件 Tab」才能看其它事件趋势。视觉 MUST 对齐 Flutter 玻璃拟态/预测/趋势语义（非像素级强制）。

#### Scenario: 切换准确率与趋势
- **WHEN** 商户在详情中点击切换准确率图与趋势图
- **THEN** 界面 MUST 按行展示对应图表且无需另选事件类型 Tab

### Requirement: 商户后台 SHALL 为 Vue SPA
商户面向控制台 MUST 以独立 Vue（或等价现代 SPA）交付，MUST NOT 复用运维 Hub 的 Admin 口令静态页作为商户登录面。Hub 运维页与商户台 MUST 保持入口与鉴权分离。

#### Scenario: 商户入口独立于 Hub
- **WHEN** 商户打开商户台登录页
- **THEN** MUST NOT 要求 `X-Admin-Password` 或 Hub Admin 会话
