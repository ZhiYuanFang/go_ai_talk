## 1. Platform 下沉（零行为）

- [x] 1.1 新建 platform 子包：内部密钥头与 Validate*、注入头常量（wxId/deviceNo/clientIP）、ParseHeaderWxID、ConstantTimeEqual
- [x] 1.2 替换 cash/voice/ucg/history/gatewayapp/device 等对旧符号的引用；删除业务包上的重复别名
- [x] 1.3 确认无对外 path 变更；相关进程可编译

## 2. B — clients 中立化

- [x] 2.1 创建 `internal/clients/cash`：迁入 RemoteIsVipByWxID、RemoteCareAlertAccess；从 `services/cash` 删除对外 Remote 出口
- [x] 2.2 创建 `internal/clients/device|history|ucg|voice`：迁入既有域间 `*_client` / HTTP 出站（cash/device/ucg/gatewayapp/simuser/voice 侧清单）
- [x] 2.3 改 voice/ucg 等调用方改 import 为 clients；去掉 `services/voice`→`services/cash|device|history` 业务 import
- [x] 2.4 改写 `domain_refs`：经 clients + contracts 装配，不再 import device/history 实现包
- [x] 2.5 排除项文档化：Python / Green / aimodel 客户端不迁入域名 clients（或单独非域名目录）

## 3. D — 隔离与门禁

- [x] 3.1 清理剩余跨域业务 import（含 controller 层：cash 不 import voice；history/ucg 不借 device 仅校验密钥）
- [x] 3.2 新增 `hack/check-service-import.sh`（或等价）：禁 services 业务互引；允许 platform/contracts/aimodel/clients
- [x] 3.3 更新 `AGENTS.md` 与 `openspec/project.md` 源码包边界条款

## 4. C — controller 按进程分包

- [x] 4.1 建 `controller/{cash,voice,device,history,ucg,gatewayapp,simuser,notify}` 并迁入对应文件
- [x] 4.2 修正误导命名文件归属（voice 的 care-alert/tip/clinic；history 的 device_history 等）；**禁止改 g.Meta path 字符串**
- [x] 4.3 更新各 `register_*_service` / gateway-app register 的 import 与 Bind
- [x] 4.4 gateway 反代与跨切面归入 gatewayapp（或 proxy）子包

## 5. 全量编译与冻结验收

- [x] 5.1 `go build` 全部 `cmd/*-service`（及 gateway-app、mcp 等）
- [x] 5.2 自检：对外 App path 字符串未改（care-alert/cash/ucg/history 等抽样 + 禁止改 path 的 diff 审查）
- [x] 5.3 跑 import 门禁为失败即退出；不新增 `*_test.go`
