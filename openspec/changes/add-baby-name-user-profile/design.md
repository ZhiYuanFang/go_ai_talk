## Context

当前 `ai_voice_device.user` 表已具备 `baby_name` 字段，但画像链路仍停留在“生日+性别”模型：
- 设备画像接口 `GET/POST /device/app/api/user/get|save` 未暴露 `babyName`。
- 自动画像接口 `POST /device/app/api/user/auto_save` 未接收 `babyName`。
- 历史页面实际调用的 `GET/POST /device/history/api/birthday|birthday/save` 仅处理 `birthday/sex`。
- 设备域与历史域缓存结构均仅缓存 `birthday/sex`。
- 网页 `resource/public/history.html` 仅显示与编辑性别，不显示宝宝名字。

该变更跨越 API 定义、控制器、服务契约、HTTP 适配器、缓存与前端页面，属于跨模块一致性改造。

## Goals / Non-Goals

**Goals:**
- 将用户画像统一为 `babyName + birthday + sex`，贯通 device-service 与 history-service 的画像读写链路。
- 保持服务边界：history 仍经 device HTTP 契约访问画像，不跨库直查。
- 页面可展示并修改宝宝名字，且与既有保存按钮共用一次提交。
- 对旧客户端保持兼容：`babyName` 可选，缺省为空串。

**Non-Goals:**
- 不引入新的数据库表或索引，不调整 `user` 表主键与唯一约束。
- 不改造微信登录主流程（`/login`、`/bindwx`）的认证语义。
- 不改造语音提示词策略（本次仅保证画像字段可被读取）。

## Decisions

### 决策 1：以 `babyName` 作为统一 JSON 字段名
- 选择：API 与缓存统一使用 `babyName`（驼峰），数据库列保持 `baby_name`。
- 原因：与现有 `deviceNo`、`lastTalkAsk` 风格一致，降低前端改造成本。
- 备选：使用 `baby_name` 直出；放弃原因是破坏现有接口命名风格并增加前端适配分支。

### 决策 2：扩展既有画像接口，而非新增独立“名字接口”
- 选择：在既有 `get/save/auto_save` 与 `birthday/birthday.save` 请求/响应中增加 `babyName`。
- 原因：避免多接口读写不一致，减少调用方接口切换。
- 备选：新增 `/baby-name` 专用接口；放弃原因是增加状态分裂与额外维护成本。

### 决策 3：缓存模型同步升级，写后立即刷新
- 选择：device 与 history 两侧画像缓存结构新增 `babyName` 字段，并在写成功后更新缓存。
- 原因：避免“数据库已更新但页面读取旧缓存”造成名字回退。
- 备选：仅删缓存不写缓存；放弃原因是会提升回源抖动，且与当前写后回填模式不一致。

### 决策 4：历史服务继续作为页面 BFF，不改变服务边界
- 选择：页面仍调用 history 接口；history 通过既有 device HTTP 契约透传 `babyName`。
- 原因：遵循当前网关/页面依赖关系，避免前端跨服务改造。
- 备选：页面改为直接调用 device 接口；放弃原因是会引入额外鉴权与前端路由耦合问题。

## Risks / Trade-offs

- [风险] DTO 字段漏改导致链路某一跳丢失 `babyName` → [缓解] 以“接口定义→控制器→契约→适配器→缓存→页面”顺序改造并做端到端自测。
- [风险] 旧缓存中无 `babyName` 字段，反序列化后空值覆盖前端展示 → [缓解] 读取时以空串作为兼容值，写入后覆盖新结构。
- [风险] 页面保存逻辑将空字符串写回，误清空名字 → [缓解] 前端显式 `trim`，并在文案中区分“已保存”与“已清空”。
- [取舍] 不新增专用名字接口，短期更快；长期若画像字段继续增长，可能需要统一画像对象版本化。

## Migration Plan

1. 发布后端接口兼容改造（新增可选 `babyName` 字段，保持旧请求可用）。
2. 发布 history-service 透传与缓存升级，确保页面读取可得到 `babyName`。
3. 发布网页 `history.html` UI 更新（展示并可编辑宝宝名字）。
4. 灰度验证：
   - 读取接口返回 `babyName`；
   - 保存生日/性别时不丢名字；
   - 单独改名字可成功；
   - 旧客户端不传 `babyName` 仍可正常使用。
5. 回滚策略：前端可先回滚；后端新增字段保持向后兼容，无需 DDL 回滚。

## Open Questions

- `babyName` 最大长度与字符集是否需要统一约束（例如 32 字、支持 emoji 与否）？
- `auto_save` 在未传 `babyName` 时是否保留原值（推荐）还是覆盖为空串（需产品确认）？
- 是否需要将 `babyName` 纳入语音提示词（本次非目标，后续可单独提案）？
