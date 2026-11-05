## 1. history.html

- [x] 1.1 历史事件表头在「开始时间」前增加「预计开始」；空态 colspan +1
- [x] 1.2 行渲染读 `expectStartTime`，有值用既有时间格式化，否则 `-`

## 2. 使用统计嵌入历史 Tab

- [x] 2.1 `api-usage-stats.html`：同步列与渲染
- [x] 2.2 `client-usage-stats.html`：同步列与渲染

## 3. 验收

- [ ] 3.1 有 expect 的记录在 history 页与 usage 右侧历史 Tab 均可见「预计开始」
- [ ] 3.2 无 expect 显示 `-`；无新增后端接口变更
