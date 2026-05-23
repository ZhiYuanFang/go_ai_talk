## 1. 前端浮层与样式

- [ ] 1.1 在 `admin.html` 增加 `#eventColorPopover` 结构：取色器、色值展示、确定/取消按钮及基础样式
- [ ] 1.2 浮层使用 `fixed` 或等价方式定位，避免表格 `overflow` 裁切；点击其它行色调时关闭已有浮层

## 2. 交互逻辑

- [ ] 2.1 色调列点击改为打开浮层并初始化 `pendingInlineEventRow` 与色值，移除「hidden color + change 即时提交」
- [ ] 2.2 **确定**：`normalizeEventColor` 后调用 `submitEventRowUpdate`；成功/失败关闭或保留浮层并写入 `eventMsg`
- [ ] 2.3 **取消**：关闭浮层且不请求 API；提交中与 `eventRowBusyId` 联动禁用按钮

## 3. 验收

- [ ] 3.1 手工验收：改色后不点确定 → 列表不变；点确定 → 列表与 DB 更新；点取消 → 不更新
