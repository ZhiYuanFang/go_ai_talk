## 1. Schema

- [x] 1.1 `appointment_next` ensure：幂等增加 `remark` 列（默认空串，VARCHAR(512)）；新建表 DDL 含该列；中文注释
- [x] 1.2 entity / do / dao columns 补齐 `remark`

## 2. API / Service

- [x] 2.1 `api/v1` App GET/PUT DTO 增加 `remark`；PUT 校验 `max-length:512`
- [x] 2.2 `api/v1` Internal GET 响应增加 `remark`
- [x] 2.3 device service：读返回 nextAt+remark；PUT 正 nextAt 持久化 remark；`nextAt=0` 同时清空 remark
- [x] 2.4 App / Internal controller 接线；确认 gateway 反代无需额外字段白名单

## 3. 校验

- [x] 3.1 `gofmt`；编译 device 相关包
- [x] 3.2 自检：App 读写 remark；nextAt=0 清空；Internal GET 同字段；无行空串
