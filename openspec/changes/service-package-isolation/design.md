## Context

单体演进微服务后：`cmd/*-service` 与一库一服务已落地，跨库禁直查；但 Go 仍单 module，`internal/services/{cash,voice,device,…}` 可互相 import。典型污染：

- 出站客户端反挂在被调方（`cash/client.go` 的 `Remote*` 被 voice/ucg import，编译链带上 cash 实现包）。
- `controller` 约 74 文件平铺；voice 宿主文件却名 `device_care_alert_*`；cash controller 为解析 wxId import `voice`。
- 内部密钥头、Bearer/Admin 头、`ParseHeaderWxID`、`ConstantTimeEqual` 散落 device/gatewayapp/voice。

约束：App/Flutter 可见 path/method/字段 **冻结**；不新建测试文件；与功能变更解耦。

## Goals / Non-Goals

**Goals:**

- 全服务一次完成 B+C+D：中立 clients、controller 按进程分包、禁跨域业务 import。
- platform 承接无业务共享工具；门禁脚本锁住回归。
- 各 `cmd/*-service` 仍能 `go build`；对外 HTTP 行为不变。

**Non-Goals:**

- 多 module / 多仓；改 App URL；改业务语义或 DB schema。
- 把 tip/care-alert 从 voice 迁到 device（宿主可仍为 voice，仅纠正包/文件归属命名）。
- 重写 contracts 业务接口语义（可微调 import 位置）。

## Decisions

### D1 — `internal/clients/{target}` 中立出站客户端

- 目录：`internal/clients/cash`、`device`、`history`、`ucg`、`voice`（按被调服务命名）。
- 迁入：`RemoteIsVipByWxID`、`RemoteCareAlertAccess`；cash 侧 `device_client`/`history_client`/`ucg_client`；device 的 `admin_http_client`/`user_internal_http`/`voice_upstream_qa`/`ucg_upload`；ucg/gatewayapp/simuser 中域间 HTTP；等。
- **排除**：Python/Aliyun Green/LLM provider（非本仓域服务）可留原包或 `clients/python` 等非域名。
- 调用方：`services/voice` → `clients/cash`，**禁止**再 `import hello/internal/services/cash`。
- **备选**（调用方旁 `voice/client/cash`）：否决为默认——全做时 ucg+voice 共用 VIP client，中立更省。

### D2 — controller 按进程子包

```
internal/controller/
  cash/ voice/ device/ history/ ucg/
  gatewayapp/   # 含 route proxy、auth exempt、usage
  simuser/ notify/
```

- 各进程 `Register*ServiceHTTP` 只 Bind 本子包类型。
- 文件可改名（`CareAlertController`），`g.Meta path:"/device/api/..."` **逐字保留**。
- 共享仅限 gateway 反代/跨切面 → `controller/gatewayapp`（或 `controller/proxy`，一期并入 gatewayapp）。

### D3 — import 禁令（D）

允许：

| 导入方 | 可 import |
|--------|-----------|
| `services/{X}` | `platform/*`、`contracts`、`aimodel`（共享选模）、`clients/{Y}`（Y≠X 或 Y=任意被调） |
| `services/{X}` | 同包 `services/{X}` |
| `controller/{X}` | `services/{X}`、`clients/*`、`platform`、`contracts`、`api/v1` |

禁止：

- `services/voice` → `services/cash` / `device` / `history` / `ucg` / `gatewayapp`（业务实现）
- `services/ucg` → `services/cash` / `device`（改为 clients + platform 头）
- `controller/cash` → `services/voice`（ParseHeaderWxID 下沉）
- `controller/history|ucg` → `services/device` 仅因密钥校验（下沉 platform）

`aimodel` / `contracts` 视为共享基础设施，允许跨进程 import（与「业务线包」区分）。

### D4 — platform 下沉清单

至少迁出（名称以实现为准）：

- 内部密钥头名 + `Validate*InternalSecret*`
- `HeaderInternalWxId` / `DeviceNo` / `ClientIP` 等网关注入头常量
- `ParseHeaderWxID`
- `ConstantTimeEqual`（或等价）
- voice 内对 gatewayapp 头的别名删除，改为 platform

### D5 — 门禁

- `hack/check-service-import.sh`（或 `.ps1`/go run）：扫描 `internal/services/{cash,voice,device,history,ucg,simuser,mcpbridge,appstatus,gatewayapp}` 之间互引；允许列表仅 `contracts`/`aimodel`/`platform`/`clients`。
- `AGENTS.md` + `openspec/project.md` 增加「源码包边界」条款（与跨库 HTTP 条款并列）。

### D6 — 实施顺序（单 change 多 commit）

1. platform 抽公共工具，改调用点（行为不变）。
2. 建 `clients/*`，迁客户端，改 services import。
3. 解 domain_refs：voice 经 clients/contracts 取 DeviceAdmin，不再 import device/history 实现包。
4. controller 分子包 + register 改 import；清污染 import。
5. 门禁脚本 + 文档；全量 `go build` 各 cmd。
6. 自检：`g.Meta path` / App 路由字符串未改。

### D7 — App URL 冻结验收

- 禁止改 `api/v1` 与 controller 中对外 `path:"..."`（internal 路径可随包走但字符串建议也不改，减少网关/文档差）。
- 评审 checklist：Flutter 无需发版。

## Risks / Trade-offs

- [巨量改动难审] → 同一 OpenSpec，提交按 D6 分批；PR 描述附目录图与「零 URL 变更」声明。
- [搬文件漏改 Bind] → 每进程 build + 启动路由 `/api.json` 或既有冒烟。
- [假解耦：client 仍在 services/cash] → 门禁禁止 services 互引；Remote 必须在 clients。
- [GoFrame 子包类型重名] → 子包名隔离；必要时类型加进程前缀。
- [与功能 PR 冲突] → 先合并/归档 `feature-activation-care-alert` 再开干，或在隔离分支 rebase。

## Migration Plan

1. 合并本变更后仅需重建各服务镜像；无 DB 迁移、无 App 发版。
2. 回滚：整 PR revert（纯结构）；无数据回滚。
3. 若门禁先 warn 后 error：过渡期一天内收紧为 fail CI。

## Open Questions

- `gatewayapp/usagestats` 是否保持子包挂在 gatewayapp 下（建议是）。
- `aimodel` 是否继续允许被 voice/ucg/sim 直接 import（本设计允许）。
- controller 子包是否改 Go package 名为 `cashctrl` 等避免与 services 同名冲突（实现自定，import 路径清晰即可）。
