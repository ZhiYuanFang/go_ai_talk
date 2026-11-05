## 1. 回滚服务端（若曾改）

- [x] 1.1 恢复 `DeviceAdminHistoryListReq` 无 startTime/endTime
- [x] 1.2 `HistoryAdminCtrl.List` 恢复走原 `inner.List`（无 ListV2 时间窗）

## 2. api-usage-stats 静态页

- [x] 2.1 左栏默认：宝宝名=`yes`、每页=`100`、`daysSelect`=`3`
- [x] 2.2 按用户右侧 Tab「历史事件」（默认）/「使用明细」；历史 list 仅 deviceNo/page/pageSize=50，无时间参数
- [x] 2.3 历史表列对齐 history.html（ID/事件名/数量/时长/开始/结束/备注，只读）；无设备号空态；链到完整运维

## 3. client-usage-stats 静态页

- [x] 3.1 同 2.1 左栏三默认
- [x] 3.2 同结构 Tab；使用明细保留功能汇总 + 时间线；历史同 2.2/2.3

## 4. 验收

- [ ] 4.1 两页默认有宝宝名/100/近3天；点行右侧默认历史、每页50、无时间筛选
- [ ] 4.2 切「使用明细」可见原内容；history-service 无本变更 API  diff
- [ ] 4.3 无 deviceNo 空态；表样式接近运维历史表
