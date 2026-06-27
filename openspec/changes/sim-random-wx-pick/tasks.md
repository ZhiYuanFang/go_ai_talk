## 1. device random 接口

- [x] 1.1 `api/v1/device_sim_internal_http.go`：新增 `DeviceSimWxRandomReq` / `DeviceSimWxRandomRes` 与中文注释
- [x] 1.2 `sim_user.go`：实现 `PickRandomSimulatedWx`（MIN/MAX + 均匀 `id>=R LIMIT 1`，空洞回退 minId）
- [x] 1.3 `device_sim_internal.go`：`SimWxRandom` handler；路由经现有 `register_device_service` Bind

## 2. sim 客户端与任务

- [x] 2.1 `clients.go`：新增 `pickRandomSimWx`（单次 GET random）；`randomSimSession` 改为只用 random + login，删除重复 list 调用
- [x] 2.2 `clients.go`：删除或停用 random 路径上的 `listSimWxIDs` / `accountForWx`；`countSimUsers` 保留 list total
- [x] 2.3 `tasks.go` `RunFollowTask`：两次 random 取不同 wxId（重试有界）；`sessionForWx` 简化为已知 account 登录或内联

## 3. 校验

- [x] 3.1 `go build` device / simuser 相关包通过
- [x] 3.2 `openspec validate sim-random-wx-pick` 通过
