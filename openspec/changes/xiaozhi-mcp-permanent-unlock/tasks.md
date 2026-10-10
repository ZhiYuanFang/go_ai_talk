## 1. Schema 与领域常量

- [x] 1.1 cash `EnsureSchema`：创建 `xiaozhi_mcp_product`、`xiaozhi_mcp_entitlement`（含中文注释）；种子一笔永久 SKU（`product_code` 固定、`apple_product_id` 默认可空）
- [x] 1.2 定义常量（product_code、unlock_method、status）与读写/授予/撤销领域函数（永久权益，不走 `ActivateFeature`）

## 2. 支付类型分支

- [x] 2.1 App 建单：`POST /cash/app/api/xiaozhi-mcp/orders`（读专用 SKU，写入 `feature_order`，调起参数同现有通道）
- [x] 2.2 履约分流：`DispatchFulfillPaid` / `FulfillFeaturePaid`（或等价）识别 MCP `product_code` → `GrantXiaozhiMcpPermanent`；非 MCP 保持原逻辑
- [x] 2.3 Apple `productId` 反查与退款路径：命中 MCP SKU 时授/撤永久权益；收益统计排除或单独分类 MCP `product_code`

## 3. App / Internal 查询

- [x] 3.1 `GET /cash/app/api/xiaozhi-mcp/unlock`：返回 `unlocked` + 可售 SKU
- [x] 3.2 `GET /cash/internal/api/xiaozhi-mcp/entitlement`（内部密钥）：供 device 校验；已开通/未开通语义明确
- [ ] 3.3 向负责人确认上述 App 接口是否计入 usage；按结论处理（未确认不改 `maintenance_skip`）
- [x] 3.4 gateway-app：确认 cash App/Admin/internal 反代与 Bearer 策略覆盖新路径；`g.Meta` 登记完整

## 4. device Add 门禁

- [x] 4.1 `clients/cash` 增加开通态查询客户端
- [x] 4.2 `AddXiaozhiMcpBinding`（或 App Add 控制器）写库前校验；未开通/ cash 失败 fail-closed；List/Alias/Delete 不校验

## 5. 独立小 Admin（一页两区）

- [x] 5.1 Admin API：SKU 读/改；手工授；撤销最近手工授（鉴权同其它 cash Admin）
- [x] 5.2 静态页 `cash-xiaozhi-mcp-admin.html`：区 A SKU、区 B 手工授/撤销；登记 `admin_static_pages.go` + `admin-modules.js` 导航
- [x] 5.3 （SHOULD）`history.html` 小智区展示当前登录 wx 开通态，未开通禁用 Add 并提示

## 6. 自检

- [x] 6.1 跑 `hack/check-service-import`（device 不 import cash 实现包）
- [x] 6.2 确认开通功能管理 / `feature/catalog` 不出现 MCP SKU
- [ ] 6.3 手工走通：Admin 配价 → App 开通态 →（沙箱）支付或手工授 → Add 成功；未开通 Add 失败；List/Delete 仍可用
