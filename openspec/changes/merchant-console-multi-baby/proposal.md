# Proposal

## Why

产品需要触达 B 端门店：商户自主注册后台、管理虚拟宝宝并分发给店员子账户，同时用连续订阅（或 Hub 授期）解除默认两宝宝上限。并行地，C 端从「一账号一宝宝」演进为可创建/绑定多个虚拟宝宝，且须与商户台、新版 Flutter **同发**，同时保持旧版 App「无宝宝→绑定一个终生宝宝」可继续使用。

## What Changes

- 新增 **商户独立身份**（店名+密码、独立 JWT）；商户仅可登录 Vue 商户后台，不可登录 C 端 App。
- 商户默认可创建 **两个虚拟宝宝**（`device_no`，无实体硬件）；开通连续订阅或 Hub 授期后 **宝宝数量不限**。
- 商户主账号可将宝宝与 **子账户** 多对多分发/解绑；子账户为普通 `wx`，经 `username/register` 创建且 `role=staff`；**仅商户后台可解绑**，客户端无解绑入口。
- `wx` 增加账号级单值 **`role`（`parent` | `staff`）**，二者互斥；同一自然人（同一 `wx` 行）不可兼为家长与店员。
- **BREAKING（契约演进，旧客户端兼容）**：宝宝归属改为 `wx↔device_no` 成员表为真源；保留 `wx.device_no` 为焦点/兼容镜像。旧 App 仍走「无→一」；新 App 支持多宝宝切换，parent 可在 App 内 **创建/绑定虚拟宝宝**。
- 商户欠费/过期：名下全部宝宝 **禁止新增** history 事件行；**允许** update/end 与只读历史/预测。
- 商户主面板：玻璃拟态宝宝卡片 + 右侧详情（仿 Flutter 预测/准确率/趋势）；**一条 SSE** 按 merchant 订阅多 `device_no`，事件带 baby 标识并局部刷新卡片。
- Hub：商户列表（开通/付费/到期/欠费）；对选定商户 **设置绝对 `expire_at`**；与支付宝连续订阅 **共用同一条 `expire_at`**。
- 商户开通默认月费 **1000 元**（Hub 可改）；支付宝为 **网页连续订阅**（非现有 VIP 一次性 +30 天包）。
- 本变更为 **单一大 change**（后端 + Vue 商户台 + Flutter 多宝宝 + Hub），tasks 内分泳道，不拆多个 change。

## Capabilities

### New Capabilities

- `merchant-console`：商户注册/登录与 JWT 边界、虚拟宝宝配额与分发、商户台读模型与多 device SSE、Vue 玻璃拟态主面板/详情行为。

### Modified Capabilities

- `identity-device`：`wx.role`、成员表多宝宝、焦点 `device_no` 兼容、App 创建/绑定虚拟宝宝、staff 子账户创建约束。
- `care-commerce`：商户权益/`expire_at`、Hub 绝对授期、支付宝连续订阅 SKU 与履约、默认月费可配。
- `gateway-edge`：商户鉴权与路由、商户多 device SSE、商户台静态/SPA 托管或反代。
- `history-feeding`：商户欠费时 create 拒绝、update/end 放行的写闸。
- `admin-hub`：商户订阅运维页（列表、欠费状态、绝对到期授期、月费配置入口）。

## Impact

- **服务**：`device-service`（身份/成员/虚拟宝宝）、`cash-service`（商户权益与支付）、`history-service`（写闸）、`gateway-app-server`（商户 JWT/SSE/反代）；可能新增商户 API 面。
- **前端**：新 Vue 商户 SPA；兄弟仓 `flutter_ai_talk` 多宝宝 UI 与同发；Hub 静态页扩展。
- **数据**：`merchant`、成员关系、商户权益表；`wx` 增 `role`；迁移现有 `wx.device_no` → 成员表 + 镜像。
- **依赖**：支付宝周期扣款/连续订阅能力（相对现网一次性 VIP 为新轨）；Redis 历史通知通道复用、商户侧订阅索引新建。
- **兼容**：旧 Flutter 无再绑入口，仅保证「无→一」与标量 `deviceNo`；staff 多宝宝时旧 App 只见默认焦点宝宝（已接受）。
