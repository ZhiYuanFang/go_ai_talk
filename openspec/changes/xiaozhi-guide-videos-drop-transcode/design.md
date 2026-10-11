## Context

- UCG 视频：`UploadMediaObject` 对 `mediaKind=2` 现为三分支（v1 直传 / 可解码转码 v2 / 拒）；internal `upload-video` 强制 `NormalizeVideo`。同步 ffmpeg 占用 ucg CPU/内存，与 C 端共用进程。
- OSS：服务端内网 PutObject；客户端预签名已停用（caidoukeji 内网 endpoint）。
- 小智：开通 Admin 页（SKU / 手工授 / 已开通列表）已存在；绑定在 device；Flutter 在兄弟仓。引导视频属运营内容，元数据适合落 cash，与开通页同域运维。
- 约束：服务边界（cash→ucg 经 `clients/ucg`）；不为 Hub 单独抬视频上限；App GET 计入 usage；不默认加 Redis；接口版本不破坏已有 v1 结构（新增路由即可）。

## Goals / Non-Goals

**Goals:**

- 全局删除服务端视频转码兜底；不合规一律拒绝。
- 双固定槽引导视频：Hub 可改标题、上传/覆盖/清空；App 回 CDN；Flutter 双空藏入口。
- Hub 上传与 Flutter/UCG **同一验真与体积规则**（v1 + `MaxMediaUploadBytes`）。

**Non-Goals:**

- 不恢复客户端 OSS 预签名直传。
- 不新建异步转码队列/独立 worker（后续另案）。
- 不做多视频 CMS、多语言、审核流。
- 不把引导视频放进群二维码式本机盘（APK 目录）。
- 本仓不改 Flutter 工程源码（契约 + Impact 约束兄弟仓同迭代）。

## Decisions

### 1. 全局删转码：改原路径，不为 Hub 开旁路

- **选择**：`UploadMediaObject` 视频分支改为：`ValidateVideoBytes(v1)` 通过则 PUT，否则 4xx；删除 `ProbeVideoDecodable`→`UploadVideoTranscodedObject` 分支。`POST …/upload-video` 改为同一语义（验真 + 直传），响应 `transformVersion=v1`（及对应 contentHash）。删除或停用 `NormalizeVideo` 调用链。
- **理由**：用户明确「迟早要删、不为 Hub 单独路径」；避免两套语义。
- **备选**：仅 Hub 跳过转码、社区保留 B 分支 → 否决（仍会卡顿且双轨）。

### 2. ffprobe 保留，ffmpeg 转码移除

- **选择**：继续用 ffprobe 做 v1/v2 register 验真；镜像可保留 `ffmpeg` 二进制（Dockerfile 不必强制删），但上传路径 MUST NOT 调用转码。`EnsureVideoThumb` 仍走 OSS snapshot，不依赖本地转码。
- **理由**：验真仍需要 probe；瘦镜像可另案。

### 3. sim T4 对齐直传

- **选择**：T4 仍调 `upload-video`；成功后按响应 `transformVersion`（首版为 `v1`）register；源片须已 v1 合规。失败则 skip job（与现失败语义类似，文案从「转码失败」改为「上传/验真失败」）。
- **理由**：去掉对服务端出 v2 的依赖。

### 4. 引导视频元数据在 cash，文件在 ucg OSS

- **选择**：表 `xiaozhi_mcp_guide_video`（`slot` PK：`promo`|`tutorial`；`title`；`object_key`；`updated_at`）。空 `object_key` = 未配置。种子两行默认标题（如「能力介绍」「绑定教学」），Hub 可改。
- OSS 前缀建议 `xiaozhi-guide/`：ucg internal 增加 `kind=xiaozhi-guide`（或等价）→ v1 验真 + Put（**无转码**），与全局规则一致，仅前缀不同。
- cash Admin multipart → `clients/ucg` → 写回 `object_key`；覆盖时 MAY 忽略旧对象删除（运维低频；可选后续 GC）。
- **备选**：元数据放 device → 开通页跨域别扭；本机盘 → 不适合 CDN 点播。

### 5. Hub UI：开通页区 D

- **选择**：扩展 `cash-xiaozhi-mcp-admin.html`：区 D 展示两槽（标题输入、当前 CDN 预览/链接、上传、清空）。不新开导航项。
- Admin API：`GET/PUT …/guide-videos`、`POST …/guide-videos/{slot}/upload`、`DELETE …/guide-videos/{slot}`（清空 object_key）。

### 6. App 读模型与 usage

- **选择**：`GET /cash/app/api/xiaozhi-mcp/guide-videos`（Bearer）；仅返回 `object_key` 非空的槽：`slot`、`title`、`videoCdnUrl`、可选 `coverCdnUrl`（`BuildVideoThumbnailURL`）。
- **usage**：负责人确认 **计入** → MUST NOT 加入 `maintenance_skip`。
- 不进 auth exempt（须登录）。

### 7. Flutter

- 进入小智页（或可见时）拉 GET；`list.length==0` → 不渲染入口；否则按 `title` 展示，点击用系统/应用内播放器播 `videoCdnUrl`。
- 契约字段稳定后兄弟仓实现；本仓 tasks 含「契约联调检查项」，不勾选 Flutter 源码任务为本仓完成条件时可单列「兄弟仓跟进」并标注本迭代必做。

## Risks / Trade-offs

- [Risk] Web/随意片源发帖失败率上升 → **Mitigation**：错误文案明确「须为 H.264+AAC 的 mp4」；原生客户端本就偏合规。
- [Risk] sim T4 片源不合规导致任务常 skip → **Mitigation**：检查/更换 T4 素材为 v1 合规；日志可观测。
- [Risk] 大文件 multipart 穿 gateway/cash/ucg 超时 → **Mitigation**：与现网 UCG 视频同限（约 25MB），运维压片；不为 Hub 抬额。
- [Risk] 删转码后仍留死代码误用 → **Mitigation**：删除调用链与入口；specs REMOVED + 代码删除。
- [Risk] 引导片与社区片共用 ucg 验真路径抢资源 → **Mitigation**：运维低频；无转码尖峰已大幅下降。

## Migration Plan

1. 先发 **ucg**（去转码 + internal 直传）与 **sim**（register v1），观察社区上传/T4。
2. 再发 **cash** + Hub 区 D + App GET；运维上传两条合规片。
3. Flutter 发版消费 GET。
4. 回滚：恢复转码代码成本高，宜靠客户端合规 + 配置开关仅在设计中不引入半吊子开关；若紧急回滚需整版回退 ucg。

## Open Questions

- （无阻塞）旧 OSS 上已转码 v2 对象与 register 行为不变；仅新上传路径变更。
- 覆盖上传是否同步删旧 objectKey：首版可不删，避免误删被引用对象。
