## 1. 功能使用统计页

- [x] 1.1 `api-usage-stats.html`：用户维度增加 pageSize 下拉（默认 20；20/50/100）、页码输入+跳转、上一页/下一页与页码信息
- [x] 1.2 `loadWxList` 改为请求可变 `page`/`pageSize`，写入 `wxTotal`；改条数回第 1 页；序号保持页内 1…N

## 2. 客户端使用统计页

- [x] 2.1 `client-usage-stats.html`：与 1.1–1.2 同等分页交互（client-usage wx-list）

## 3. 自检

- [x] 3.1 两页：默认 20 条、下一页、改 pageSize、跳转越界夹紧；宝宝名筛/lastAt 排序仍仅本页
