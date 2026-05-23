## 1. 前端浮层与样式

- [x] 1.1 在 `admin.html` 增加 `#eventColorPopover` 结构：取色器、色值展示、确定/取消按钮及基础样式
- [x] 1.2 浮层使用 `fixed` 或等价方式定位，避免表格 `overflow` 裁切；点击其它行色调时关闭已有浮层

## 2. 交互逻辑

- [x] 2.1 色调列点击改为打开浮层并初始化 `pendingInlineEventRow` 与色值，移除「hidden color + change 即时提交」
- [x] 2.2 **确定**：`normalizeEventColor` 后调用 `submitEventRowUpdate`；成功/失败关闭或保留浮层并写入 `eventMsg`
- [x] 2.3 **取消**：关闭浮层且不请求 API；提交中与 `eventRowBusyId` 联动禁用按钮

## 3. 验收

- [x] 3.1 手工验收：改色后不点确定 → 列表不变；点确定 → 列表与 DB 更新；点取消 → 不更新

## 4. 编辑弹框展示（用户补充）

- [x] 4.1 编辑/新增弹框增加「当前 Logo 与色调」预览区；编辑时展示库中 logo 图与色块/色值
- [x] 4.2 弹框内修改取色器或选择新 Logo 文件时同步更新预览
