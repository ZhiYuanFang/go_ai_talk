# Tasks

## 1. 改名迁移去 suggest

- [x] 1.1 从 `MigrateDeviceNo` 及内部 migrate API/注释中移除对 `suggest` 的 UPDATE；用 `rg "dao\.Suggest|migrate suggest"` 确认 history 迁移路径无命中
- [x] 1.2 修正未归档 `hub-rename-baby-device-no` 制品中「迁 suggest」表述（design/spec/tasks），与本变更一致；抽查该 change 的 history-feeding delta 已无 suggest 跟迁

## 2. 拆除 HTTP 与服务契约

- [x] 2.1 删除 voice suggest List/Delete API、controller、`suggest_internal` 与 register bind；`go build`（或等价包编译）通过且路由注册无 suggest
- [x] 2.2 删除 device history / admin suggest 请求类型与 controller 方法；契约接口去掉 `ListSuggest`/`DeleteSuggest`；相关 clients/adapter/delegate/HTTP path helpers 同步删除；全仓 `rg "ListSuggest|DeleteSuggest|VoiceSuggest"` 无业务命中

## 3. Hub UI 与遗留模型

- [x] 3.1 从 `resource/public/history.html` 移除成长建议区块、`loadSuggest` 及调用；页面加载指定设备时不再请求 suggest 管理口
- [x] 3.2 清理无引用的 `entity`/`do`/`dao` Suggest 与 hack dao 配置表项（保持可编译）；确认无 DROP 表脚本被误加

## 4. 意图字面与规格对齐收尾

- [x] 4.1 确认 `target_type=suggest` 无写库/无成长建议生成调用（枚举兼容可留）；`rg "成长建议|ActionTargetTypeSuggest"` 仅余注释或兼容解析
- [x] 4.2 列出本变更需重部署进程（至少 voice-service、history-service，及托管 Hub 静态页的进程）；在 apply 收尾说明中写明

## Workflow follow-up

- 实现完成后按项目流程归档本变更（合并 delta 进主规格后删除 change 目录）。
- Hub 自由填写 OLD→NEW 补迁入口仍属改名运维另案，不在本 tasks 跟踪。
