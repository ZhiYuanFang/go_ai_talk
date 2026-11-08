## 1. Refresh 合并入口（history）

- [x] 1.1 实现同 `deviceNo` trailing 300ms schedule（mutex + timer map）；到期调用现有 `RefreshBabyLiveActivity`；中文注释声明非扫表 ticker
- [x] 1.2 internal HTTP Refresh 入口改为走 schedule（或文档约定 voice 只打该入口）
- [x] 1.3 本进程若需直接 Refresh，统一走 schedule；禁止旁路立即双发

## 2. 收敛触发点

- [x] 2.1 history C/U/D（含闭合）：仅 `invalidateImminentPending`，去掉 Refresh
- [x] 2.2 voice pending PUT（空/非空成功后）：schedule Refresh；保留 delayed MQ
- [x] 2.3 voice MQ fire 过闸：保留普通推送；schedule Refresh
- [x] 2.4 voice active-eta PUT/清空：删除 `refreshUnifiedLiveActivity` 调用

## 3. 校验

- [x] 3.1 `gofmt`；编译 history / voice 相关包
- [x] 3.2 自检：CUD 无 LA accepted；pending PUT / fire 有合并后的 Refresh；active-eta 无 LA；短窗连打只投影一次
