# OpenSpec 基线规格 v3.0.3

> 本文件由 `openspec/specs` 下全部 capability 规格于 **v3.0.3** 合并而成，作为该版本的确定性规则基线，便于按版本查阅。

> 后续新变更请基于本文件对照增量，或通过 OpenSpec 新建 change 再合并至下一版本规格。

## 目录

- [admin-event-inline-color-confirm](#admin-event-inline-color-confirm)
- [admin-qa-library](#admin-qa-library)
- [ai-model-admin-ui](#ai-model-admin-ui)
- [ai-monthly-quota](#ai-monthly-quota)
- [aimodel-media-gen](#aimodel-media-gen)
- [aimodel-thinking-mode](#aimodel-thinking-mode)
- [app-ai-quota-degraded-ui](#app-ai-quota-degraded-ui)
- [app-legal-docs](#app-legal-docs)
- [app-status-banner-service](#app-status-banner-service)
- [apple-server-notifications](#apple-server-notifications)
- [apple-sign-in-api](#apple-sign-in-api)
- [async-cache-projection-sync](#async-cache-projection-sync)
- [background-loop-task-governance](#background-loop-task-governance)
- [cache-and-messaging-hard-dependencies](#cache-and-messaging-hard-dependencies)
- [cachekit-zrevrange-parse](#cachekit-zrevrange-parse)
- [care-alert-access-gate](#care-alert-access-gate)
- [care-alert-cash-vip](#care-alert-cash-vip)
- [care-alert-feature-unlock](#care-alert-feature-unlock)
- [care-alert-feeding-eligibility](#care-alert-feeding-eligibility)
- [care-alert-force-refresh](#care-alert-force-refresh)
- [care-alert-python-wire](#care-alert-python-wire)
- [care-alert-quota-lane](#care-alert-quota-lane)
- [care-alert-unlock-client](#care-alert-unlock-client)
- [care-alert-vip-selection](#care-alert-vip-selection)
- [cash-admin-grant-revoke](#cash-admin-grant-revoke)
- [cash-admin-hub-tiles](#cash-admin-hub-tiles)
- [cash-admin-manual-grant](#cash-admin-manual-grant)
- [cash-admin-revenue](#cash-admin-revenue)
- [cash-duration-days](#cash-duration-days)
- [cash-service-runtime](#cash-service-runtime)
- [cash-vip-admin-api](#cash-vip-admin-api)
- [cash-vip-admin-ui](#cash-vip-admin-ui)
- [catalog-activatable-total](#catalog-activatable-total)
- [chat-stream-intent-land](#chat-stream-intent-land)
- [chat-stream-reply-only](#chat-stream-reply-only)
- [chat-streaming](#chat-streaming)
- [chinese-documentation-governance](#chinese-documentation-governance)
- [ci-acr-github-secrets](#ci-acr-github-secrets)
- [client-feature-usage-stats](#client-feature-usage-stats)
- [clinic-tip-feedback-flutter](#clinic-tip-feedback-flutter)
- [clinic-tip-feedback-voice-host](#clinic-tip-feedback-voice-host)
- [company-site-legal](#company-site-legal)
- [compose-container-resource-limits](#compose-container-resource-limits)
- [compose-host-root-asset-volumes](#compose-host-root-asset-volumes)
- [compose-mysql-endpoint-via-env](#compose-mysql-endpoint-via-env)
- [compose-mysql-test-seed-desensitization](#compose-mysql-test-seed-desensitization)
- [compose-prod-test-dual-stack](#compose-prod-test-dual-stack)
- [compose-redis-topology-2g](#compose-redis-topology-2g)
- [controller-process-layout](#controller-process-layout)
- [dao-extension-layer-simplification](#dao-extension-layer-simplification)
- [database-unix-timestamp-storage](#database-unix-timestamp-storage)
- [deepseek-history-redis-prefer](#deepseek-history-redis-prefer)
- [device-admin](#device-admin)
- [device-admin-event-logo-color-ui](#device-admin-event-logo-color-ui)
- [device-admin-event-parent-picker-ui](#device-admin-event-parent-picker-ui)
- [device-admin-event-tree-ui](#device-admin-event-tree-ui)
- [device-admin-user-list](#device-admin-user-list)
- [device-app-device-login](#device-app-device-login)
- [device-event-cache-rebuild-on-mutate](#device-event-cache-rebuild-on-mutate)
- [device-event-hierarchy](#device-event-hierarchy)
- [device-event-logo-color](#device-event-logo-color)
- [device-event-type-field](#device-event-type-field)
- [device-event-update-parent-id](#device-event-update-parent-id)
- [device-ops-ai-debug](#device-ops-ai-debug)
- [device-route-canary-management](#device-route-canary-management)
- [device-sim-user](#device-sim-user)
- [device-wx-by-device-no-list](#device-wx-by-device-no-list)
- [device-wx-profile-apis](#device-wx-profile-apis)
- [docker-deploy-logging](#docker-deploy-logging)
- [documentation-language-compliance](#documentation-language-compliance)
- [domain-package-boundary-enforcement](#domain-package-boundary-enforcement)
- [end-open-history-by-event](#end-open-history-by-event)
- [enum-adapter-compatibility](#enum-adapter-compatibility)
- [event-logo-oss-cdn](#event-logo-oss-cdn)
- [feature-activation-subject](#feature-activation-subject)
- [feature-activation-toolkit](#feature-activation-toolkit)
- [feature-admin-activation-snapshot](#feature-admin-activation-snapshot)
- [feature-admin-activation-subject](#feature-admin-activation-subject)
- [feature-admin-api](#feature-admin-api)
- [feature-admin-contract-ux](#feature-admin-contract-ux)
- [feature-admin-rule-readonly](#feature-admin-rule-readonly)
- [feature-admin-semantic-layout](#feature-admin-semantic-layout)
- [feature-admin-ui](#feature-admin-ui)
- [feature-branding-client](#feature-branding-client)
- [feature-catalog-entitlement](#feature-catalog-entitlement)
- [feature-catalog-sku-embed](#feature-catalog-sku-embed)
- [feature-order-query](#feature-order-query)
- [feature-sku-update](#feature-sku-update)
- [feature-unlock-fulfillment](#feature-unlock-fulfillment)
- [feature-visual-branding](#feature-visual-branding)
- [feeding-chat-ws-io](#feeding-chat-ws-io)
- [feeding-effective-day-core](#feeding-effective-day-core)
- [feeding-eligibility-admin](#feeding-eligibility-admin)
- [feeding-llm-admin-simplify](#feeding-llm-admin-simplify)
- [feeding-shell-b-removal](#feeding-shell-b-removal)
- [gateway-admin-jwt](#gateway-admin-jwt)
- [gateway-app-api-usage-stats](#gateway-app-api-usage-stats)
- [gateway-app-cors](#gateway-app-cors)
- [gateway-app-cors-reverse-proxy](#gateway-app-cors-reverse-proxy)
- [gateway-app-device-login-device-no](#gateway-app-device-login-device-no)
- [gateway-app-jwt-device-no-header](#gateway-app-jwt-device-no-header)
- [gateway-app-official-site](#gateway-app-official-site)
- [gateway-app-path-only-assets](#gateway-app-path-only-assets)
- [gateway-app-server](#gateway-app-server)
- [gateway-app-version-admin](#gateway-app-version-admin)
- [gateway-app-version-admin-crud](#gateway-app-version-admin-crud)
- [gateway-app-version-check](#gateway-app-version-check)
- [gateway-no-business-workers](#gateway-no-business-workers)
- [gateway-policy-layer-consolidation](#gateway-policy-layer-consolidation)
- [gateway-route-middleware-domain-isolation](#gateway-route-middleware-domain-isolation)
- [gateway-voice-http-proxy](#gateway-voice-http-proxy)
- [gateway-ws-delegation-convergence](#gateway-ws-delegation-convergence)
- [gateway-ws-edge-proxy](#gateway-ws-edge-proxy)
- [growth-invite-per-user-only](#growth-invite-per-user-only)
- [growth-trajectory-user-gates](#growth-trajectory-user-gates)
- [history-delegate-downstream-urls](#history-delegate-downstream-urls)
- [history-device-sync-cache-projection](#history-device-sync-cache-projection)
- [history-event-unit-denorm](#history-event-unit-denorm)
- [history-filter-api](#history-filter-api)
- [history-list-v2-api](#history-list-v2-api)
- [history-piece-and-realtime-notify](#history-piece-and-realtime-notify)
- [history-profile-nickname](#history-profile-nickname)
- [history-service-db-ownership](#history-service-db-ownership)
- [history-voice-delegation](#history-voice-delegation)
- [hms-notification-payload](#hms-notification-payload)
- [intent-clarify-conversation-id](#intent-clarify-conversation-id)
- [invite-code-identity](#invite-code-identity)
- [invite-group-qr-admin](#invite-group-qr-admin)
- [invite-group-qr-catalog](#invite-group-qr-catalog)
- [lane-free-model](#lane-free-model)
- [llm-lane-admin](#llm-lane-admin)
- [llm-lane-gate](#llm-lane-gate)
- [logger-level-env](#logger-level-env)
- [main-config-boundary-pruning](#main-config-boundary-pruning)
- [main-config-without-database](#main-config-without-database)
- [mcp-feeding-ws-text](#mcp-feeding-ws-text)
- [microservice-boundary-final-alignment](#microservice-boundary-final-alignment)
- [microservice-ci-release](#microservice-ci-release)
- [notify-service-runtime](#notify-service-runtime)
- [pangbao-ai-clinic](#pangbao-ai-clinic)
- [pangbao-ai-clinic-flutter](#pangbao-ai-clinic-flutter)
- [predict-imminent-push](#predict-imminent-push)
- [predict-imminent-push-copy](#predict-imminent-push-copy)
- [predict-imminent-schedule](#predict-imminent-schedule)
- [prediction-allowed-count](#prediction-allowed-count)
- [prediction-count-grant](#prediction-count-grant)
- [prediction-effective-count](#prediction-effective-count)
- [privacy-policy-company](#privacy-policy-company)
- [prod-ecs-cutover](#prod-ecs-cutover)
- [push-app-and-internal-api](#push-app-and-internal-api)
- [push-caller-migration](#push-caller-migration)
- [push-device-admin](#push-device-admin)
- [push-dispatch-observability](#push-dispatch-observability)
- [push-service-runtime](#push-service-runtime)
- [push-token-uniqueness](#push-token-uniqueness)
- [python-intent-target-action-align](#python-intent-target-action-align)
- [redis-disaster-recovery-runbook](#redis-disaster-recovery-runbook)
- [redis-platform-access](#redis-platform-access)
- [redis-read-model-cache](#redis-read-model-cache)
- [remove-clinic-tip-feedback](#remove-clinic-tip-feedback)
- [remove-device-tip](#remove-device-tip)
- [remove-prediction-slot-unlock](#remove-prediction-slot-unlock)
- [routing-key-governance](#routing-key-governance)
- [routing-key-governance-workflow](#routing-key-governance-workflow)
- [routing-key-prefix-registry](#routing-key-prefix-registry)
- [runtime-dependency-check](#runtime-dependency-check)
- [runtime-docs-centralization-and-governance](#runtime-docs-centralization-and-governance)
- [service-boundary-no-cross-db](#service-boundary-no-cross-db)
- [service-code-full-cutover](#service-code-full-cutover)
- [service-dedicated-config-loading](#service-dedicated-config-loading)
- [service-import-isolation](#service-import-isolation)
- [service-migration-safety-and-rollback](#service-migration-safety-and-rollback)
- [service-outbound-clients](#service-outbound-clients)
- [service-runtime-standardization](#service-runtime-standardization)
- [sim-llm-lane-admin](#sim-llm-lane-admin)
- [sim-runtime-config](#sim-runtime-config)
- [sim-user-admin](#sim-user-admin)
- [sim-user-service](#sim-user-service)
- [single-default-db-per-service](#single-default-db-per-service)
- [tip-contract-slim](#tip-contract-slim)
- [tip-flutter-sse-go-dialect](#tip-flutter-sse-go-dialect)
- [tip-generate-voice-host](#tip-generate-voice-host)
- [tip-python-derived-context](#tip-python-derived-context)
- [tip-streaming](#tip-streaming)
- [typed-domain-enums](#typed-domain-enums)
- [ucg-admin-post-moderation](#ucg-admin-post-moderation)
- [ucg-admin-profile-moderation](#ucg-admin-profile-moderation)
- [ucg-ai-quota](#ucg-ai-quota)
- [ucg-aliyun-secrets-env](#ucg-aliyun-secrets-env)
- [ucg-app-http-api](#ucg-app-http-api)
- [ucg-app-profile](#ucg-app-profile)
- [ucg-audit-mq](#ucg-audit-mq)
- [ucg-audit-mq-reliability](#ucg-audit-mq-reliability)
- [ucg-audit-publish-outbox](#ucg-audit-publish-outbox)
- [ucg-chat-mysql-persist](#ucg-chat-mysql-persist)
- [ucg-chat-ws](#ucg-chat-ws)
- [ucg-data-model](#ucg-data-model)
- [ucg-device-internal-api](#ucg-device-internal-api)
- [ucg-entry-eligibility](#ucg-entry-eligibility)
- [ucg-feed-index-lazy-warm](#ucg-feed-index-lazy-warm)
- [ucg-feed-index-startup-heal](#ucg-feed-index-startup-heal)
- [ucg-feed-no-geo-zset-fallback](#ucg-feed-no-geo-zset-fallback)
- [ucg-feed-redis-store](#ucg-feed-redis-store)
- [ucg-following-feed](#ucg-following-feed)
- [ucg-force-first-grant](#ucg-force-first-grant)
- [ucg-force-ledger](#ucg-force-ledger)
- [ucg-gateway-proxy](#ucg-gateway-proxy)
- [ucg-green-audit](#ucg-green-audit)
- [ucg-image-thumb](#ucg-image-thumb)
- [ucg-internal-profile-batch](#ucg-internal-profile-batch)
- [ucg-media-server-upload-only](#ucg-media-server-upload-only)
- [ucg-oss-caidoukeji-config](#ucg-oss-caidoukeji-config)
- [ucg-oss-presign](#ucg-oss-presign)
- [ucg-recommend-feed](#ucg-recommend-feed)
- [ucg-recommend-mq](#ucg-recommend-mq)
- [ucg-service-runtime](#ucg-service-runtime)
- [ucg-sim-chat-internal](#ucg-sim-chat-internal)
- [ucg-sim-chat-unread-sample](#ucg-sim-chat-unread-sample)
- [ucg-sim-feed-sample](#ucg-sim-feed-sample)
- [ucg-video-thumb](#ucg-video-thumb)
- [ucg-video-transcode](#ucg-video-transcode)
- [ucg-video-validate](#ucg-video-validate)
- [user-agreement-company](#user-agreement-company)
- [validated-prefix-dispatch](#validated-prefix-dispatch)
- [vip-entitlement](#vip-entitlement)
- [vip-overlay-entitlement](#vip-overlay-entitlement)
- [vip-payment](#vip-payment)
- [vip-product-admin](#vip-product-admin)
- [vip-product-pricing](#vip-product-pricing)
- [vip-quota-entitlement](#vip-quota-entitlement)
- [voice-admin-ui](#voice-admin-ui)
- [voice-ai-quota](#voice-ai-quota)
- [voice-and-device-service-decomposition](#voice-and-device-service-decomposition)
- [voice-chat-dashscope-stt](#voice-chat-dashscope-stt)
- [voice-device-domain-http-access](#voice-device-domain-http-access)
- [voice-device-profile-http-contract](#voice-device-profile-http-contract)
- [voice-event-child-disambiguation](#voice-event-child-disambiguation)
- [voice-history-http-contract](#voice-history-http-contract)
- [voice-internal-text-chat](#voice-internal-text-chat)
- [voice-realtime-asr-ws](#voice-realtime-asr-ws)
- [voice-route-canary-management](#voice-route-canary-management)
- [voice-textchat-resilience](#voice-textchat-resilience)
- [wechat-ios-universal-links](#wechat-ios-universal-links)
- [wechat-oauth-platform-config](#wechat-oauth-platform-config)
- [worker-dedicated-config-loading](#worker-dedicated-config-loading)
- [worker-exclusive-background-runtime](#worker-exclusive-background-runtime)
- [wx-account-vip](#wx-account-vip)
- [wx-created-at](#wx-created-at)
- [wx-username-auth](#wx-username-auth)

---

## admin-event-inline-color-confirm

<!-- source: openspec/specs/admin-event-inline-color-confirm/spec.md -->

# admin-event-inline-color-confirm Specification

## Purpose
TBD - created by archiving change admin-event-inline-color-confirm. Update Purpose after archive.
## Requirements
### Requirement: 行内色调编辑须确认后保存

事件管理页在行内修改事件色调时，SHALL 提供明确的用户确认步骤；系统 SHALL NOT 在取色器 `change` 事件发生时自动调用更新接口。

#### Scenario: 打开色调编辑浮层

- **WHEN** 已登录管理员点击某行「色调」展示区域
- **THEN** 页面 SHALL 显示包含取色控件及 **确定**、**取消** 控件的编辑浮层
- **AND** 取色器 SHALL 初始化为该行当前 `color`（合法时）或默认色

#### Scenario: 点击确定后提交

- **WHEN** 用户在浮层中调整颜色并点击 **确定**
- **THEN** 客户端 SHALL 调用 `POST /device/admin/api/event/update` 并携带该行完整字段与新 `color`
- **AND** 成功后 SHALL 关闭浮层并刷新事件列表

#### Scenario: 点击取消不提交

- **WHEN** 用户在浮层中点击 **取消** 或等价取消操作
- **THEN** 系统 SHALL NOT 调用更新接口
- **AND** 浮层 SHALL 关闭且列表数据保持不变

#### Scenario: 提交中防重复

- **WHEN** 色调更新请求正在进行
- **THEN** 浮层内 **确定** 按钮 SHALL 处于不可用或加载状态直至请求结束

### Requirement: 弹窗编辑与其它行内能力不受影响

弹窗「编辑事件」中的色调与提交行为 SHALL 保持可用；行内 Logo 点击上传流程 SHALL 不因本变更而不可用。

#### Scenario: 弹窗编辑仍可用

- **WHEN** 用户点击行内「编辑」按钮
- **THEN** 页面 SHALL 仍打开含色调字段的编辑弹窗并按原逻辑提交

---

## admin-qa-library

<!-- source: openspec/specs/admin-qa-library/spec.md -->

# Spec: 管理端问答库

## REQ-QA-001 分页列表

管理端 MUST 通过分页接口获取问答库，默认 `pageSize=10`，按 `id` 降序。

#### Scenario: 首页预览

- **WHEN** 管理员登录设备管理页
- **THEN** 问答库卡片展示最多 10 条最新记录

#### Scenario: 分页参数

- **WHEN** 请求 `GET /device/admin/api/qa/list?page=2&pageSize=20`
- **THEN** 响应包含 `list`、`total`、`page`、`pageSize`

## REQ-QA-002 展开更多

当 `total > 10` 时，管理端首页 MUST 显示「展开更多」链接至全量页。

## REQ-QA-003 删除

全量页 MUST 支持按 `id` 删除问答库行；删除 MUST 经 voice 内部接口落库，device 不直连 `qa` 表。

#### Scenario: 删除成功

- **WHEN** 管理员确认删除并提交 `POST /device/admin/api/qa/delete`
- **THEN** 对应行从列表消失且 voice 库记录已删除

---

## ai-model-admin-ui

<!-- source: openspec/specs/ai-model-admin-ui/spec.md -->

# ai-model-admin-ui Specification

## Purpose
TBD - created by archiving change ai-model-admin-with-sim-runtime-db. Update Purpose after archive.
## Requirements
### Requirement: gateway-app SHALL provide unified AI model and concurrency admin page

系统 MUST 提供独立 Admin 页面 **`/device/admin/ai-model-admin.html`**（静态文件 `resource/public/ai-model-admin.html`），经 gateway-app 托管。页面 MUST 在一屏内展示并编辑 **7 条 LLM lane** 的 `provider`、`model`、`maxInFlight`、`maxWaiters`：

| 分组 | lane 标识 | 后端 API |
|------|-----------|----------|
| Voice | `voiceUnderstanding`、`clinic` | `GET/PUT /voice/admin/api/llm-lanes` |
| UCG 润笔 | polish（`visionModel` 作 model） | `GET/PUT /ucg/admin/api/ai-config` |
| Sim | `simText`、`simVision`、`simImageGen`、`simVideoGen` | `GET/PUT /sim/admin/api/llm-lanes` |

页面 MUST 并行加载三域配置；保存 MUST 分别调用对应 PUT（可 `Promise.all`），并在部分失败时分域展示错误。页面 MUST 展示 allowlist 驱动的 provider→model 下拉联动。页面 MUST 含简短说明：同 `provider+model` 的多 lane 共享 Redis 闸门池。

#### Scenario: 统一页加载七 lane

- **WHEN** 已鉴权管理员打开 ai-model-admin.html
- **THEN** 页面 MUST 展示 voice×2、ucg polish、sim×4 共七组模型与并发字段且值来自对应 GET API

#### Scenario: 统一页保存 voice lane

- **WHEN** 管理员修改 `clinic.maxInFlight` 并点击保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes` 且 MUST NOT 调用 sim scheduler reload

#### Scenario: 部分域保存失败

- **WHEN** voice PUT 成功而 sim PUT 返回 400
- **THEN** 页面 MUST 明确提示 sim 失败原因且 voice 成功状态 MUST 可见

### Requirement: Admin Hub SHALL link ai-model-admin

`resource/public/admin-modules.js` MUST 增加 `id: ai-model-admin` 模块入口，导航至 `/device/admin/ai-model-admin.html`，`showInNav: true`。Hub 登录后 MUST 可点击进入。

#### Scenario: Hub 导航可见 AI 模型与并发

- **WHEN** 管理员登录 Admin Hub
- **THEN** 模块列表 MUST 包含 ai-model-admin 入口

### Requirement: voice-admin ucg-admin sim-admin SHALL link to unified page instead of editing LLM

`voice-admin.html`、`ucg-admin.html`、`sim-admin.html` MUST **移除** LLM 模型/并发编辑 Tab 或表单控件。各页 MUST 保留指向 `/device/admin/ai-model-admin.html` 的可见链接（文案含「模型与并发」或等价）。各页原有非 LLM 职责（额度、Prompt、任务状态等）MUST 不变。

#### Scenario: voice-admin 无 LLM 编辑 Tab

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 MUST NOT 含「LLM 车道」编辑 Tab 或 maxInFlight 输入框，且 MUST 含跳转统一页链接

#### Scenario: ucg-admin 无 polish 并发编辑

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 MUST NOT 含 provider/maxInFlight/maxWaiters/visionModel 编辑控件，且 MUST 含跳转统一页链接

#### Scenario: sim-admin 无 LLM lane 编辑

- **WHEN** 运维打开 sim-admin.html
- **THEN** 页面 MUST NOT 含 sim LLM lane 编辑表单，且 MUST 含跳转统一页链接

### Requirement: ai-model-admin static page MUST NOT count toward App usage stats

`/device/admin/ai-model-admin.html` 及本页触发的 voice/ucg/sim Admin PUT 为运维型接口，MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: 统一页保存不计入 usage

- **WHEN** 管理员从 ai-model-admin 保存 llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增 App API 计数

### Requirement: AI 模型与并发页展示 careAlert 与 free

`/device/admin/ai-model-admin.html` MUST 在 Voice 区域展示 **三条** lane：`voiceUnderstanding`、`clinic`、`careAlert`；对上述三条及 UCG `polish` MUST 提供 freeProvider/freeModel 编辑控件（允许清空）。Sim 四条 MUST NOT 展示 free 编辑。保存时 MUST 将 free 字段一并提交至对应域 Admin API。页面仍为运维型，MUST NOT 计入 App usage 统计。

#### Scenario: 页面加载三条 Voice lane

- **WHEN** 已鉴权管理员打开 ai-model-admin.html
- **THEN** 页面 MUST 展示 careAlert 块及各业务 lane 的 free 字段

#### Scenario: 清空 free 可保存

- **WHEN** 管理员清空某业务 lane 的 free 并保存成功
- **THEN** 后续非 premium 路径 MUST 按 omit/Python 自选语义工作

---

## ai-monthly-quota

<!-- source: openspec/specs/ai-monthly-quota/spec.md -->

# ai-monthly-quota Specification

## Purpose
TBD - created by archiving change ai-monthly-quota. Update Purpose after archive.
## Requirements
### Requirement: UCG polish SHALL pre-check quota and consume only on DashScope success

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（经 **ucg-service** `LanePolish` / aimodel）前于本进程执行 polish 额度 check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用上游。额度 check 通过后、调用上游前 MUST 经 polish lane 闸门 `Acquire`；若队列满 MUST 返回 **50301**「当前队列已满，请稍后重试」且 MUST NOT 调用上游、MUST NOT consume。上游成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求上游

#### Scenario: 上游失败不扣减

- **WHEN** check 通过但上游返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: 队列满不扣减

- **WHEN** check 通过但 polish lane 闸门返回队列满
- **THEN** API SHALL 返回 50301 且 SHALL NOT 调用 consume

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（**voiceUnderstanding** lane，含母婴理解、casual 流式、成长建议、历史问答等全部喂养 voice LLM 路径）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device **user 域** internal API 由 deviceNo 反查，MUST NOT 经 ai-quota API）。wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`voice_ai`** 执行 check；`allowed=false` MUST 返回 WS 错误帧 code **40302** message **「本月额度已用完」**。额度 check 通过后、调用上游前 MUST 经 voiceUnderstanding lane 闸门 `Acquire`；队列满 MUST 返回 **50301** 且 MUST NOT 调用 LLM、MUST NOT consume。LLM 成功完成后 MUST consume。模式切换、规则回复、LLM 失败兜底、纯 ASR、50301 MUST NOT check 或 consume（50301 在额度 check 之后短路，不 consume）。

#### Scenario: 未登录不可用喂养 AI

- **WHEN** WS 会话 wxId 解析为 0 且用户 utterance 将触发 LLM
- **THEN** 系统 SHALL 返回 40301 且 SHALL NOT 调用 LLM

#### Scenario: 喂养 AI 额度用尽

- **WHEN** check 得到 voice_ai used=limit
- **THEN** WS SHALL 返回 40302「本月额度已用完」且 SHALL NOT 调用 LLM

#### Scenario: 模式切换不扣减

- **WHEN** 用户发送模式切换指令且不触发 LLM
- **THEN** 系统 SHALL NOT 调用 check 或 consume

#### Scenario: 队列满不扣减喂养额度

- **WHEN** voice_ai check 通过但 voiceUnderstanding 闸门队列满
- **THEN** WS SHALL 返回 50301 且 MUST NOT consume voice_ai

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**，**但 polish HTTP 与 clinic WS 在仅因月度额度用尽（degraded 路径）时 MUST NOT 再返回 40302**。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。额度展示 MUST 从分域 API 获取：voice 域（voiceAi、clinicAi）与 ucg 域（polish），MUST NOT 依赖 `/device/app/api/ai-quota`。当额度 API 返回 `degraded=true` 时，polish 与 clinicAi MUST 展示降速文案（见 `app-ai-quota-degraded-ui`），MUST NOT 仅展示「剩余 0 次」。

#### Scenario: 润笔 degraded 不弹 40302

- **WHEN** 用户额度用尽且 polish API 成功返回正文与 `quotaDegraded=true`
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」

#### Scenario: 喂养 AI 超额弹框

- **WHEN** voice WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 胖宝 degraded 不弹 40302

- **WHEN** 用户 clinic_ai 额度用尽且 clinic WS 正常流式返回答案
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」

### Requirement: Voice clinic AI SHALL enforce clinic_ai quota locally

voice-service 在处理 `/voice/clinic/ws` 的 `question` 并即将调用 LLM（**clinic** lane）前 MUST 解析 wxId>0。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`clinic_ai`** 执行 check；`allowed=false` MUST 返回 WS 40302。业务限流 42901 检查后、调用上游前 MUST 经 clinic lane 闸门 `Acquire`；队列满 MUST 返回 **50301** 且 MUST NOT 调用 LLM、MUST NOT consume。LLM 流式成功完成后 MUST consume。摘要刷新失败、rate limit、参数校验失败、50301 MUST NOT consume。

#### Scenario: 胖宝额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM

#### Scenario: 队列满不扣减胖宝额度

- **WHEN** clinic_ai check 与 42901 检查通过但 clinic 闸门队列满
- **THEN** WS SHALL 返回 50301 且 MUST NOT consume clinic_ai

---

## aimodel-media-gen

<!-- source: openspec/specs/aimodel-media-gen/spec.md -->

# aimodel-media-gen Specification

## Purpose
TBD - created by archiving change ucg-sim-user-service. Update Purpose after archive.
## Requirements
### Requirement: aimodel SHALL expose sim lanes sharing Redis gates by model

`internal/services/aimodel` MUST 新增 Lane：`simText`、`simVision`、`simImageGen`、`simVideoGen`。`sim-user-service` 启动时 MUST 注册 `ProfileStore` 提供上述 lane 的 `provider`、`model`、`maxInFlight`、`maxWaiters`。

默认 model MUST 为：

| Lane | Model |
|------|-------|
| simText | glm-4.7-flash |
| simVision | glm-4.6v-flash |
| simImageGen | cogview-3-flash |
| simVideoGen | cogvideox-flash |

`Acquire` MUST 使用与现有服务相同的 Redis 键 `ai:llm:gate:{normalizedModel}:*`，使同 model 跨进程共用并发池。

#### Scenario: Shared gate with voice

- **WHEN** voice-service 与 sim-service 同时调用 `glm-4.7-flash`
- **THEN** 二者 MUST 竞争同一 inflight 上限

### Requirement: aimodel SHALL support image and video generation for sim

包 MUST 导出：

- `GenerateImage(ctx, lane, prompt) (result, err)` — CogView-3-Flash，POST 时 `Acquire`，释放后返回可下载 URL 或字节
- `SubmitVideoGeneration(ctx, lane, prompt) (taskID, err)` — CogVideoX-Flash 提交，POST 时 `Acquire`
- `PollVideoGeneration(ctx, taskID) (VideoPollResult, err)` — GET `/paas/v4/async-result/{taskID}` 轮询，MUST NOT 占用 inflight 槽

`PollVideoGeneration` 解析 MUST 支持智谱 **AsyncVideoGenerationResponse**：

- 状态：`task_status`（`PROCESSING`、`SUCCESS`、`FAIL`，大小写不敏感）；MAY fallback 顶层 `status`
- 视频 URL MUST 优先 `video_result[0].url`；MAY fallback `video_url` 或 `output.video_url`
- `SUCCESS` 且 URL 非空 → `VideoStatusSuccess`；`SUCCESS` 无 URL → `VideoStatusProcessing`；`FAIL` → `VideoStatusFailed`；`PROCESSING`/其他 → `VideoStatusProcessing`

#### Scenario: Image generation acquires gate

- **WHEN** sim 调用 GenerateImage
- **THEN** 上游 HTTP 请求期间 MUST 持有 `cogview-3-flash` 槽位并在完成后释放

#### Scenario: Poll does not acquire gate

- **WHEN** sim 调用 PollVideoGeneration
- **THEN** MUST NOT 调用 `Acquire` inflight

#### Scenario: Official SUCCESS with video_result

- **WHEN** async-result 响应为 `{"task_status":"SUCCESS","video_result":[{"url":"https://example.com/a.mp4"}]}`
- **THEN** PollVideoGeneration MUST 返回 status=success 且 VideoURL 为上述 url

#### Scenario: PROCESSING without video_result

- **WHEN** 响应为 `{"task_status":"PROCESSING"}`
- **THEN** MUST 返回 status=processing 且 VideoURL 为空

#### Scenario: FAIL task

- **WHEN** 响应为 `{"task_status":"FAIL"}`
- **THEN** MUST 返回 status=failed

#### Scenario: Legacy video_url fallback

- **WHEN** 响应为 `{"task_status":"SUCCESS","video_url":"https://legacy.example/v.mp4"}` 且无 video_result
- **THEN** MUST 返回 status=success 且 VideoURL 为 legacy url

### Requirement: sim LLM text and vision SHALL use Invoke with sim lanes

昵称、文案、聊天、评论生成 MUST 经 `aimodel.Invoke`（或 `InvokeStream` 若需要）与对应 sim lane，MUST NOT 在 sim-service 内直连智谱 HTTP 绕过闸门。

#### Scenario: Comment uses simVision

- **WHEN** T2 生成评论
- **THEN** 调用 MUST 使用 `LaneSimVision` 且受 polish 同 model 池约束

### Requirement: aimodel ChatRequest SHALL support optional temperature for chat completions

`aimodel.ChatRequest` MUST 提供可选字段 `Temperature *float64`。当指针非 nil 时，`Invoke` / `InvokeStream`（及 `InvokeWithHeldProfile` 等等价路径）构建的上游 JSON MUST 包含 `temperature` 字段且值为所设浮点数。当指针为 nil 时，MUST NOT 向请求体写入 `temperature`（保持上游 provider 默认行为）。

#### Scenario: Temperature omitted by default

- **WHEN** 调用方构造 `ChatRequest` 且未设置 `Temperature`
- **THEN** 序列化后的请求体 MUST NOT 含 `temperature` 键

#### Scenario: Temperature explicitly set

- **WHEN** 调用方设置 `Temperature` 指向 `0.85`
- **THEN** 请求体 MUST 含 `"temperature": 0.85`

#### Scenario: Stream and non-stream parity

- **WHEN** 同一 `ChatRequest`（含 Temperature）分别用于流式与非流式
- **THEN** 二者 MUST 写入相同 `temperature` 值

---

## aimodel-thinking-mode

<!-- source: openspec/specs/aimodel-thinking-mode/spec.md -->

# aimodel-thinking-mode Specification

## Purpose
TBD - created by archiving change aimodel-thinking-disabled-default. Update Purpose after archive.
## Requirements
### Requirement: aimodel SHALL 对智谱请求显式声明 thinking 模式且默认 disabled

`internal/services/aimodel` 在 `provider=zhipu` 的 chat/completions 请求体中 MUST **始终**包含 `thinking.type` 字段。当 `ChatRequest.ThinkingEnabled` 为 `false` 或未设置（零值）时 MUST 发送 `thinking.type=disabled`。当 `ThinkingEnabled` 为 `true` 时 MUST 发送 `thinking.type=enabled`。MUST NOT 因未启用 thinking 而省略 `thinking` 字段（避免 GLM-4.7 等模型上游默认开启 thinking）。

#### Scenario: sim 短文案默认关闭 thinking

- **WHEN** sim-user 经 `aimodel.Invoke(LaneSimText, ChatRequest{MaxTokens: 64})` 调用且未设置 `ThinkingEnabled`
- **THEN** 发往智谱的请求体 MUST 含 `"thinking": {"type": "disabled"}`

#### Scenario: clinic opt-in 开启 thinking

- **WHEN** voice clinic 经 `aimodel.InvokeStream` 调用且 `ThinkingEnabled=true`
- **THEN** 发往智谱的请求体 MUST 含 `"thinking": {"type": "enabled"}`

### Requirement: ChatRequest.MaxTokens SHALL 文档化与 thinking 的预算关系

`aimodel.ChatRequest` 的 `MaxTokens` 字段注释 MUST 说明：该值为上游 completion token 上限；当 thinking enabled 时 reasoning 与 content **共用**该预算，调用方不得将其仅理解为最终正文长度上限。

#### Scenario: 调用方阅读契约

- **WHEN** 开发者查阅 `internal/services/aimodel/request.go` 中 `ChatRequest`
- **THEN** MUST 可见 thinking 默认 disabled 语义及 MaxTokens 与 reasoning 共预算的说明

### Requirement: DeepSeek adapter SHALL 仅在 opt-in 时启用 thinking

`provider=deepseek` 时，仅当 `ChatRequest.ThinkingEnabled=true` MUST 写入 `extra_body.thinking`（或等价配置）与可选 `reasoning_effort`。`ThinkingEnabled=false` MUST NOT 写入 thinking 启用字段。

#### Scenario: 喂养 voice 未 opt-in

- **WHEN** voice 经 aimodel 调用 DeepSeek provider 且 `ThinkingEnabled=false`
- **THEN** 请求体 MUST NOT 含 thinking enabled 配置

---

## app-ai-quota-degraded-ui

<!-- source: openspec/specs/app-ai-quota-degraded-ui/spec.md -->

# app-ai-quota-degraded-ui Specification

## Purpose
TBD - created by archiving change ai-quota-degraded-fallback. Update Purpose after archive.
## Requirements
### Requirement: Flutter SHALL 解析额度 API degraded 字段

`AiQuotaFeatureStatus`（`app/lib/data/ai_quota_models.dart`）MUST 扩展 **`degraded`** 布尔字段（JSON `degraded`，默认 false）。`VoiceAiQuotaStatus` 与 `PolishAiQuotaStatus` 的 fromJson MUST 解析各 feature 的 `degraded`。当 `degraded=true` 时，`remaining` 计算逻辑不变（仍为 `limit - used`）。

#### Scenario: 解析 polish degraded

- **WHEN** `/ucg/app/api/ai-quota` 返回 `polish: { used: 5, limit: 5, degraded: true }`
- **THEN** `PolishAiQuotaStatus.polish.degraded` SHALL 为 true

#### Scenario: 解析 clinic degraded

- **WHEN** `/voice/app/api/ai-quota` 返回 `clinicAi: { used: 10, limit: 10, degraded: true }`
- **THEN** `VoiceAiQuotaStatus.clinicAi.degraded` SHALL 为 true

### Requirement: AiQuotaRemainingHint SHALL 展示降速文案

`AiQuotaRemainingHint`（`app/lib/ui/widgets/ai_quota_remaining_hint.dart`）在 **`remaining=0` 且 `degraded=true`** 时 MUST 展示醒目降速文案，MUST NOT 仅展示「剩余 0 次」：

- **polish**：**「本月润笔额度已用完，已降速」**
- **clinicAi**：**「本月胖宝诊疗额度已用完，已降速」**
- **voiceAi**：保持现有 **「本月 AI 对话剩余 N 次」** 文案（`remaining=0` 时展示「剩余 0 次」）；degraded 字段不影响 voice hint 文案

降速文案 SHOULD 使用区别于普通 hint 的视觉权重（如略高对比度或 accent 色），具体样式由实现决定。

#### Scenario: 润笔降速 hint

- **WHEN** polish 额度 API 返回 used=5、limit=5、degraded=true
- **THEN** compose 页 hint SHALL 展示「本月润笔额度已用完，已降速」

#### Scenario: 胖宝降速 hint

- **WHEN** clinicAi 额度 API 返回 used=limit、degraded=true
- **THEN** 胖宝诊疗页 hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」

#### Scenario: voice_ai 额度用尽 hint 不变

- **WHEN** voiceAi used=limit、degraded=true
- **THEN** hint SHALL 仍展示「本月 AI 对话剩余 0 次」（或等价 remaining 文案）

### Requirement: Flutter SHALL 收窄 40302 弹框至 voice_ai

`ai_quota_errors.dart` 及全局 40302 处理 MUST：**polish HTTP** 与 **clinic WS** 在仅额度用尽场景不再收到 40302，MUST NOT 弹「本月额度已用完」。**voice_ai WS** 收到 40302 MUST 仍弹框 **「本月额度已用完」**。40301 行为不变。

#### Scenario: voice 喂养 40302 仍弹框

- **WHEN** voice chat WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: polish 成功 degraded 不弹框

- **WHEN** polish 返回 200 且 body 含 `quotaDegraded=true`
- **THEN** App SHALL NOT 弹 40302 额度用尽框

### Requirement: ucg_compose_screen SHALL 处理 quotaDegraded 可选提示

`ucg_compose_screen.dart` 在润笔成功响应 **`quotaDegraded=true`** 时 MAY 展示一次性 snackbar/toast（如「已切换至降速模式」）；`quotaDegraded=false` 或省略时 MUST NOT 展示该提示。该提示 MUST NOT 替代 `AiQuotaRemainingHint` 的持久降速文案。

#### Scenario: degraded 润笔可选 toast

- **WHEN** 用户触发润笔且响应 `quotaDegraded=true`
- **THEN** App MAY 展示一次性降速提示
- **AND** hint 区域 SHALL 同步反映 degraded 状态（刷新 quota provider 或本地标记）

### Requirement: pangbao_ai_screen SHALL 刷新 clinic 额度 degraded 状态

胖宝诊疗页（`pangbao_ai_screen.dart`）MUST 使用含 `degraded` 的额度 provider 驱动 `AiQuotaRemainingHintFeature.clinicAi`；额度用尽后用户仍可提问时，hint MUST 展示降速文案而非触发 40302 弹框。

#### Scenario: 额度用尽后仍可提问

- **WHEN** clinic_ai degraded=true 且用户发送 question 并成功收到 answer_done
- **THEN** UI SHALL NOT 展示 40302 弹框
- **AND** hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」

### Requirement: App 额度状态与 VIP 有额度语义一致

App 所依赖的 voice/ucg `ai-quota` API 在账号为 VIP 时，MUST 对相关 feature 返回 `allowed=true`，MUST NOT 再以 40302「本月额度已用完」阻断 VIP 用户的 premium 路径。非 VIP 用户在额度用尽走 free/omit 时，API 可继续返回 `degraded=true`（或等价），但文案/弹框策略 SHOULD 与「免费模型降速」而非「功能不可用」对齐；Flutter 完整文案改版可作为跨仓跟随，本仓 MUST 保证字段语义可供客户端区分 VIP 与降速。

#### Scenario: VIP 拉取 voice ai-quota

- **WHEN** VIP 用户 GET `/voice/app/api/ai-quota`
- **THEN** `voiceAi`/`clinicAi`（及若暴露的 `careAlert`）MUST `allowed=true`

#### Scenario: 非 VIP 额尽仍可 degraded

- **WHEN** 非 VIP 用户某 feature used 已达 limit
- **THEN** 快照 MUST 标明不可再走 premium（如 `allowed=false` 且 `degraded=true`），客户端 MUST NOT 将其理解为 VIP

---

## app-legal-docs

<!-- source: openspec/specs/app-legal-docs/spec.md -->

# app-legal-docs Specification

## Purpose
TBD - created by archiving change update-legal-docs-apple-sign-in. Update Purpose after archive.
## Requirements
### Requirement: 隐私政策 SHALL 披露 Apple 登录数据收集边界

`resource/public/privacy-policy.html` MUST 在「我们收集哪些信息」中说明：使用 Apple 登录时，为建立与识别账户，系统收集并存储 Apple 提供的匿名用户标识符；系统 SHALL NOT 在文案中声称存储 Apple 邮箱、姓名或登录凭证原文。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读隐私政策中的 Apple 登录说明

- **WHEN** 用户通过 App WebView 或浏览器打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 Apple 登录相关收集说明
- **AND** 说明 SHALL 与后端仅持久化 `apple_sub` 的行为一致

### Requirement: 用户协议 SHALL 披露 Apple 登录账号方式

`resource/public/user-agreement.html` MUST 在「账号注册与安全」中说明：iOS 用户可选用「通过 Apple 登录」建立账户。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读用户协议中的 Apple 登录说明

- **WHEN** 用户通过 App WebView 或浏览器打开 `/user-agreement.html`
- **THEN** 页面 SHALL 包含 Apple 登录作为可选登录方式的说明

### Requirement: 合规文档 URL SHALL 保持不变

gateway-app 暴露的合规文档路径 MUST 仍为 `/privacy-policy.html` 与 `/user-agreement.html`，客户端无需因本次修订修改加载 URL。

#### Scenario: 客户端加载合规文档

- **WHEN** 客户端请求既有隐私政策或用户协议 URL
- **THEN** gateway SHALL 返回更新后的 HTML 内容
- **AND** 路径 SHALL 与修订前相同

### Requirement: 隐私政策 SHALL 披露胖宝 AI 与模型供应商

`resource/public/privacy-policy.html` MUST 说明喂养语音 AI、胖宝诊疗与润笔等功能可能将用户输入或摘要发送至 **可配置的第三方大模型服务**（包括但不限于智谱 GLM、DeepSeek、阿里云 DashScope，以运维后台实际配置为准）。文档 MUST 在 AI 相关章节说明 **胖宝诊疗**：为生成回答，系统会读取用户近 **7 天喂养记录聚合摘要**（非完整原始记录全文），并 MAY 在 App 内展示 AI **思考过程**（thinking）供用户参考。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读胖宝诊疗相关说明

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 7 天喂养摘要与 thinking 展示说明，且用户向名称 SHALL 为「胖宝诊疗」

#### Scenario: 供应商描述与可配置 provider 一致

- **WHEN** 系统默认使用智谱模型且 Admin 可切回 DeepSeek/DashScope
- **THEN** 隐私政策 SHALL 表述为可配置多供应商，且 MUST NOT 声称仅使用单一固定供应商

---

## app-status-banner-service

<!-- source: openspec/specs/app-status-banner-service/spec.md -->

# app-status-banner-service Specification

## Purpose
TBD - created by archiving change app-status-banner-service. Update Purpose after archive.
## Requirements
### Requirement: app-status-service SHALL expose public banner API without auth

`app-status-service` MUST 提供 `GET /app/api/status/banner`，**无鉴权**。响应 MUST 使用 `{ code, message, data }` 信封。`data.active=false` 时 MUST 仅返回 `{ "active": false }`。`data.active=true` 时 MUST 含 `title`、`message`、`dismissible`、`updatedAt`，MAY 含 `expectedEndAt`（unix 秒）与 `contentKey`。响应 MUST 设置 `Cache-Control: public, max-age=30`。

#### Scenario: inactive banner

- **WHEN** 进程内 `active=false`
- **THEN** `GET /app/api/status/banner` 返回 `data.active=false` 且无 title/message

#### Scenario: active maintenance banner

- **WHEN** 运维 PUT `active=true` 且 `dismissible=false`
- **THEN** App 可读全字段且 `dismissible` 为 false

### Requirement: app-status-service SHALL provide admin banner CRUD with Hub credentials

同进程 MUST 提供 `POST /admin/api/login`（`GATEWAY_APP_ADMIN_USERNAME/PASSWORD`）、`GET /admin/api/banner`、`PUT /admin/api/banner`（Admin JWT）。状态 MUST 存于进程内存；重启 MUST 恢复默认 inactive。MUST NOT 使用 MySQL/Redis。

#### Scenario: admin saves banner

- **WHEN** 管理员 Bearer 有效并 PUT 新 title/message
- **THEN** 随后 GET public banner 立即反映新文案

### Requirement: app-status-service SHALL serve standalone admin page

MUST 托管 `resource/public/app-status-admin.html` 于 `GET /admin`，含登录、启用、文案、expectedEndAt、dismissible、保存、立即关闭、预览与 contentKey 快照提示。

#### Scenario: gateway down

- **WHEN** gateway-app 不可用但 status 服务存活
- **THEN** 运维仍可通过 status 域名打开 `/admin` 并修改 banner

---

## apple-server-notifications

<!-- source: openspec/specs/apple-server-notifications/spec.md -->

# apple-server-notifications Specification

## Purpose
TBD - created by archiving change add-apple-server-notifications. Update Purpose after archive.
## Requirements
### Requirement: cash-service MUST 提供 Apple ASN V2 异步通知入口

cash-service MUST 提供 `POST /cash/app/api/vip/apple/notifications`：gateway-app MUST 将该精确 path 列入 Bearer 匿名白名单（与支付宝 notify 同模式）。请求体 MUST 按 App Store Server Notifications V2 解析；系统 MUST 校验 `signedPayload` 密码学签名。验签失败 MUST NOT 履约。校验通过后 MUST 在同一请求路径内同步履约或处理退款（MUST NOT 依赖 RabbitMQ 或扫表）。对 Apple 的成功应答 MUST 为 HTTP 2xx，以便停止无意义重试；业务上无法映射订单但仍验签成功的可忽略类型 MUST 返回 2xx 并记日志，避免毒害重试队列。

#### Scenario: 合法购买通知开通

- **WHEN** Apple 投递验签通过的购买成功类通知，且 `appAccountToken` 对应未支付的 VIP 或功能订单
- **THEN** 系统 MUST 将订单置 `paid`（若尚未），并按该订单类型开通 VIP 或功能权益

#### Scenario: 验签失败

- **WHEN** `signedPayload` 签名非法
- **THEN** 系统 MUST NOT 开通权益，MUST NOT 将订单标为 paid

#### Scenario: 重复通知幂等

- **WHEN** 同一 Apple `transactionId`（渠道交易号）对已 paid 订单再次通知
- **THEN** 系统 MUST 幂等成功且 MUST NOT 错误地重复叠加权益时长或数量

### Requirement: Apple 建单 MUST 生成并返回 appAccountToken UUID

当 `channel=apple_iap` 创建 VIP 或功能订单时，系统 MUST 生成 UUID 写入订单的 `app_account_token`，并在建单响应中返回 `appAccountToken`。客户端 MUST 将该值作为 StoreKit 购买的 `appAccountToken`。ASN 履约 MUST 使用通知中的 `appAccountToken` 查找订单；缺失 token 或找不到订单时 MUST NOT 凭空开通（可记日志）。

#### Scenario: Apple 建单返回 token

- **WHEN** 用户以 `channel=apple_iap` 成功建单
- **THEN** 响应 MUST 含非空 UUID 格式的 `appAccountToken`，且库中对应订单行 MUST 存有相同值

#### Scenario: 无 token 无法 ASN 开通

- **WHEN** 通知验签通过但 `appAccountToken` 为空或无匹配订单
- **THEN** 系统 MUST NOT 开通权益

### Requirement: REFUND 通知 MUST 收回对应开通

当验签通过的通知类型为退款（`REFUND`）时，系统 MUST 将该渠道交易关联订单标记为退款语义，并撤销或失效该笔支付所授予的 VIP 续期或功能权益（实现须可测、幂等）。MUST NOT 因退款通知开通新权益。

#### Scenario: 退款收回 VIP

- **WHEN** 某已履约 VIP 订单的 `transactionId` 收到合法 `REFUND`
- **THEN** 该账号因该笔支付获得的权益 MUST 被撤销或调整为不再有效（按实现约定），且重复 REFUND MUST 幂等

### Requirement: gateway 与 usage 对 ASN 的约定

gateway-app MUST 反代 `/cash/app/api/*` 使 ASN path 可达。`POST /cash/app/api/vip/apple/notifications` MUST NOT 计入 App usage 统计，MUST 写入 `maintenance_skip`（或等价精确排除）。

#### Scenario: 匿名可达

- **WHEN** 无用户 Bearer 的 Apple 服务器 POST 该 notifications path
- **THEN** gateway MUST 放行至 cash-service（由 cash 验签）

#### Scenario: 不计入 usage

- **WHEN** ASN 回调成功处理
- **THEN** 系统 MUST NOT 将其记为需展示的 App 功能使用次数

---

## apple-sign-in-api

<!-- source: openspec/specs/apple-sign-in-api/spec.md -->

# apple-sign-in-api Specification

## Purpose
TBD - created by archiving change apple-sign-in-api. Update Purpose after archive.
## Requirements
### Requirement: wx 表 SHALL 持久化 Apple JWT sub

系统 MUST 在 `wx` 表新增 `apple_sub` 列，用于存储经校验的 Apple `identityToken` 内 `sub` 字段。系统 SHALL NOT 将 Apple 邮箱、姓名或 `identityToken` 原文写入业务库。`apple_sub` SHALL 建立唯一索引；允许多行 `apple_sub` 为 NULL（兼容既有微信/用户名记录）。

#### Scenario: 首次 Apple 登录创建 wx 行
- **WHEN** 校验通过的 `sub` 在 `wx` 表中不存在
- **THEN** 系统 SHALL 插入新 `wx` 行并仅设置 `apple_sub` 与 `platform`（及系统默认字段）
- **AND** 响应 `isNewUser` SHALL 为 `true`
- **AND** 系统 SHALL NOT 写入 `unionid` 或 Apple 邮箱

#### Scenario: 既有 Apple 用户再次登录
- **WHEN** 校验通过的 `sub` 已存在于某 `wx.apple_sub`
- **THEN** 系统 SHALL 返回该行的 `wxId` 与已绑定 `deviceNo`（若有）
- **AND** 响应 `isNewUser` SHALL 为 `false`

### Requirement: device-service SHALL 校验 Apple identityToken

device-service SHALL 提供 `POST /device/app/api/user/apple/login`（与网关聚合 `POST /device/app/api/apple_login` 区分），接受 JSON body 至少含 **`identityToken`**（字符串）与 **`platform`**（与客户端约定，iOS 为 `ios`）。系统 SHALL 使用 Apple JWKS（`https://appleid.apple.com/auth/keys`）验证 JWT 签名，并校验 **`iss`** 为 `https://appleid.apple.com`、**`aud`** 等于配置项 `apple.ios.bundleId`（`com.fzy.pangbao`）、**`exp`** 未过期。校验失败时 SHALL 返回明确业务错误且 SHALL NOT 创建或匹配用户行。body 中的 **`authorizationCode`** MAY 为可选字段；本能力不强制以其换票，但 SHALL 在 API 定义中保留以便将来收紧校验。

#### Scenario: 有效 identityToken 登录成功
- **WHEN** 客户端提交未过期且 `aud` 匹配的 `identityToken`
- **THEN** 系统 SHALL 提取 `sub` 并完成查/建 `wx` 行
- **AND** 响应 SHALL 包含 `wxId`、`isNewUser`、已绑定时的 `deviceNo`
- **AND** 响应 SHALL NOT 包含 `accessToken`、`refreshToken` 或 `apple_sub` 明文

#### Scenario: 无效或过期 token
- **WHEN** `identityToken` 签名校验失败、`aud` 不匹配或已过期
- **THEN** 系统 SHALL 返回业务校验失败语义
- **AND** SHALL NOT 创建或更新 `wx` 行

#### Scenario: identityToken 为空
- **WHEN** `identityToken` trim 后为空
- **THEN** 系统 SHALL 返回参数错误
- **AND** SHALL NOT 调用 Apple JWKS

### Requirement: gateway-app SHALL 聚合 Apple 登录并签发令牌

gateway-app-server SHALL 提供 **`POST /device/app/api/apple_login`**，将请求体（至少 **`identityToken`**、**`platform`**）转发至 device-service 的 **`POST /device/app/api/user/apple/login`**；当 device 返回成功且 `wxId > 0` 时，SHALL 签发 **`accessToken`**（JWT：`sub` 为 wx 主键，含 **`device_no` claim** 当 device 返回非空 `deviceNo`）与 **`refreshToken`**，响应字段 SHALL 与 `POST /device/app/api/login`、`POST /device/app/api/username_login` 对齐：`wxId`、`deviceNo`、`isNewUser`、`accessToken`、`refreshToken`。该路径 SHALL 列入 Bearer 鉴权白名单。

#### Scenario: 聚合登录成功签发 JWT
- **WHEN** 客户端调用 **`POST /device/app/api/apple_login`** 且 device 业务返回成功与有效 `wxId`
- **THEN** gateway SHALL 返回 `accessToken` 与 `refreshToken`
- **AND** 响应 `data` SHALL 含 `wxId`、`deviceNo`、`isNewUser`

#### Scenario: device 业务失败
- **WHEN** device 返回非零 `code` 或 `wxId` 无效
- **THEN** gateway SHALL 返回业务失败语义
- **AND** SHALL NOT 签发 JWT

### Requirement: Apple 与微信登录 SHALL 为独立查/建路径

Apple 登录 SHALL 以 `apple_sub` 查/建 `wx` 行；微信登录 SHALL 以 `unionid` 查/建。用户仅使用单一方式登录且未绑定第二方式时，SHALL 对应单条 `wx` 行。用户分别以两种方式**各独立登录一次且从未在同一行绑定**时，SHALL 产生两条独立 `wx` 行。

#### Scenario: 仅 Apple 登录未绑定微信
- **WHEN** 用户通过 `apple_login` 登录且当前 `wx` 行仅有 `apple_sub`
- **THEN** 系统 SHALL 维护单条 `wx` 记录
- **AND** 用户 MAY 继续使用而无需绑定微信

#### Scenario: 两次独立登录产生两行
- **WHEN** 同一自然人先后以 Apple、微信各登录一次，且每次登录时均未将第二方式绑定到当前行
- **THEN** 系统 SHALL 维护两条独立 `wx` 记录（各含对应标识）
- **AND** SHALL NOT 自动合并

### Requirement: 已登录用户 SHALL 可将第二登录方式绑定到当前 wx 行

系统 SHALL 提供 Bearer 绑定能力，向**当前**已登录 `wx` 行 UPDATE 第二标识符（`apple_sub` 或 `unionid`）。绑定 SHALL 在目标标识符未被**不同** `wxId` 占用时成功。系统 MUST NOT 合并两条已独立存在的完整 `wx` 行（各 `wxId` 已分别通过独立登录创建且含对应标识）。

#### Scenario: Apple 用户绑定微信成功
- **WHEN** 已登录 `wx` 行仅有 `apple_sub`、`unionid` 为空，且 `POST /device/app/api/user/wx/bindwx` 提交的 `jsCode` 换得 `unionid` 未被其他 `wxId` 占用
- **THEN** 系统 SHALL 将 `unionid` 写入当前 `wx` 行
- **AND** 当前会话 `wxId` SHALL 不变

#### Scenario: 微信用户绑定 Apple 成功
- **WHEN** 已登录 `wx` 行仅有 `unionid`、`apple_sub` 为空，且 `POST /device/app/api/user/apple/bind` 提交的 `identityToken` 校验通过且 `sub` 未被其他 `wxId` 占用
- **THEN** 系统 SHALL 将 `apple_sub` 写入当前 `wx` 行
- **AND** 当前会话 `wxId` SHALL 不变

#### Scenario: apple_sub 已被其他 wx 行占用
- **WHEN** 用户尝试 `apple/bind`，但校验得的 `sub` 已存在于 `wxId' != 当前 wxId` 的行
- **THEN** 系统 SHALL 返回 `ErrAppleSubTakenByOther`
- **AND** SHALL NOT 修改任何 `wx` 行

#### Scenario: unionid 已被其他 wx 行占用
- **WHEN** 用户尝试 `wx/bindwx`，但 `unionid` 已存在于 `wxId' != 当前 wxId` 的行
- **THEN** 系统 SHALL 返回 `ErrUnionIDTakenByOther`
- **AND** SHALL NOT 修改任何 `wx` 行

#### Scenario: 尝试合并两条已独立完整账号
- **WHEN** 用户曾分别以 Apple、微信各登录并各产生独立 `wx` 行，现以其中一行登录并尝试绑定另一行已占用的 `apple_sub` 或 `unionid`
- **THEN** 系统 SHALL 返回 `ErrAppleSubTakenByOther`、`ErrUnionIDTakenByOther` 或 `ErrAccountMergeConflict`
- **AND** SHALL NOT 合并、删除或迁移另一条 `wx` 行的数据

#### Scenario: 绑定端点须 Bearer 鉴权
- **WHEN** 客户端调用 `apple/bind` 或 `wx/bindwx` 且无有效 Bearer access token
- **THEN** 系统 SHALL 拒绝请求
- **AND** SHALL NOT 写入 `wx` 行

### Requirement: bind 端点契约

device-service SHALL 暴露 **`POST /device/app/api/user/apple/bind`**（body 至少含 **`identityToken`**、**`platform`**）与 **`POST /device/app/api/user/wx/bindwx`**（body 至少含 **`jsCode`**、**`platform`**）。两路径 SHALL 从请求头 **`X-Internal-Wx-Id`**（gateway 从 JWT `sub` 注入）定位当前 `wx` 行。`wx/bindwx` SHALL 泛化绑定语义，不限于用户名账号；既有 **`POST /device/app/api/user/username/bindwx`** SHALL 保留。

#### Scenario: apple/bind 请求校验
- **WHEN** 已登录用户提交有效 `identityToken` 且 `sub` 可绑定到当前行
- **THEN** 系统 SHALL 更新当前行 `apple_sub` 并返回成功
- **AND** SHALL NOT 返回新 JWT（会话 `wxId` 不变）

#### Scenario: wx/bindwx 供 Apple-only 账号使用
- **WHEN** 当前 `wx` 行仅有 `apple_sub`、无用户名密码，且 `jsCode` 有效
- **THEN** `wx/bindwx` SHALL 允许绑定微信
- **AND** SHALL NOT 要求先创建用户名密码

### Requirement: profile SHALL 可选暴露绑定状态

profile 读接口（如 `GET /device/app/api/user/detail` 或等价）SHALL 在响应中提供 **`isAppleBound`**（当且仅当 `apple_sub` 非空时为 `true`）与 **`authProviders`**（按当前行已配置身份派生的最小列表，如 `apple`、`wechat`、`username`）。SHALL NOT 在 profile 中暴露 `apple_sub` 或 `unionid` 明文。

#### Scenario: 双绑账号 profile
- **WHEN** 当前 `wx` 行同时含非空 `apple_sub` 与 `unionid`
- **THEN** 响应 SHALL 含 `isAppleBound=true`、`isWxBound=true`
- **AND** `authProviders` SHALL 同时包含 `apple` 与 `wechat`

### Requirement: 配置 SHALL 提供 iOS Bundle ID

device-service 配置 SHALL 包含 `apple.ios.bundleId`，默认值 SHALL 为 `com.fzy.pangbao`，供 `identityToken` 的 `aud` 校验使用。生产环境 SHALL 可通过配置文件或等价覆盖机制修改，且 SHALL NOT 将 Bundle ID 硬编码在多处业务逻辑中。

#### Scenario: aud 与配置一致时通过
- **WHEN** token 的 `aud` 等于当前配置的 `apple.ios.bundleId`
- **THEN** `aud` 校验 SHALL 通过

#### Scenario: aud 与配置不一致时拒绝
- **WHEN** token 的 `aud` 不等于配置的 `apple.ios.bundleId`
- **THEN** 登录 SHALL 失败并返回明确错误

### Requirement: 联调页 SHALL 支持 Apple 登录探测

`resource/public/gateway-app-integration-test.html` SHALL 提供用户可触发的操作，向当前配置的网关基址发起 **`POST /device/app/api/apple_login`**（`Content-Type: application/json`，body 含 **`identityToken`** 与 **`platform`**），并将响应中的 token 与业务字段展示在页面日志区（与现有微信/用户名登录区块并列或分区清晰）。

#### Scenario: 联调页发起 apple_login
- **WHEN** 运维在联调页填入有效 `identityToken` 并触发 Apple 登录
- **THEN** 页面 SHALL 展示 HTTP 状态与响应体中的 `accessToken`、`refreshToken`、`wxId`、`deviceNo`、`isNewUser`

---

## async-cache-projection-sync

<!-- source: openspec/specs/async-cache-projection-sync/spec.md -->

# async-cache-projection-sync Specification

## Purpose
TBD - created by archiving change async-redis-read-model-for-history-action-event-user. Update Purpose after archive.
## Requirements
### Requirement: 缓存投影 MUST 由写路径同步 patch 与读 miss 重建承担

`history`/`device` 读模型缓存更新 MUST 在写库成功后的同请求内同步 patch（见 `history-device-sync-cache-projection`），读路径 cache miss 时 MUST 回源 MySQL 并回填。系统 MUST NOT 依赖 worker-service、domain_outbox relay 或异步投影 consumer 作为唯一更新路径。

#### Scenario: 写后同步 patch 替代 outbox relay

- **WHEN** history 记录新增且事务提交成功
- **THEN** 系统 MUST 在同请求内 patch Redis 读模型，MUST NOT 仅 enqueue domain_outbox 等待 worker

#### Scenario: 无 worker 时读路径仍正确

- **WHEN** Redis list key 不存在且客户端请求 history 列表
- **THEN** 系统 MUST 从 MySQL 重建列表并回填 Redis

---

## background-loop-task-governance

<!-- source: openspec/specs/background-loop-task-governance/spec.md -->

# background-loop-task-governance Specification

## Purpose
TBD - created by archiving change remove-worker-simplify-cache. Update Purpose after archive.
## Requirements
### Requirement: 循环后台任务 MUST 经 OpenSpec 明确批准后方可引入

系统 MUST **默认禁止**在 `internal/services/**` 业务实现中新增循环后台任务，包括但不限于：`time.NewTicker` / `time.Tick` 轮询、`for { select {} }` 常驻 goroutine 扫描 MySQL/Redis/outbox 表、定时 reconciler、pending 业务表全表/分页扫描兜底。

新增上述任务 **MUST** 在 OpenSpec **proposal 与 design** 中写明：任务名称、宿主进程、周期/触发条件、环境开关、失败语义、关闭方式，以及 **为何不能** 在请求内同步完成或使用 AMQP push consumer 替代。

#### Scenario: 未批准的后台 ticker 被拒绝合入

- **WHEN** PR 在业务包内新增 `Start*Reconciler` 或 ticker 扫表且无对应 OpenSpec 变更引用
- **THEN** 评审 MUST 要求补充已批准的变更或删除该后台任务

#### Scenario: 已批准 UCG outbox relay 仍允许

- **WHEN** 变更引用已归档或进行中的 UCG OpenSpec（如 audit publish outbox、chat persist）且在 design 中声明 ticker 语义
- **THEN** 该后台任务 MAY 存在于 `ucg-service` 进程

### Requirement: AMQP push consumer 与 ticker 扫表 MUST 区分治理

经 RabbitMQ **broker push**（`autoAck=false`）的消息 consumer **不视为** ticker 扫表任务，但 **MUST** 在 OpenSpec 变更中声明队列名、routing key 与宿主进程。HTTP Management API Pull 轮询队列 **视为** 循环后台任务，适用批准流程。

#### Scenario: UCG AMQP consumer 合规

- **WHEN** `ucg-service` 启动 AMQP push consumer 消费 `ucg.audit.*` 或 `ucg.recommend.score.q`
- **THEN** 该 consumer MUST 有 OpenSpec 依据且 MUST NOT 部署在已删除的 worker-service 内

### Requirement: 角色边界变更 MUST 伴随文档更新

当后台任务执行角色或宿主进程发生调整时，运行文档与部署说明 MUST 同步更新，以确保运维与开发对角色边界认知一致。

#### Scenario: 移除 worker-service 后文档更新

- **WHEN** 完成 worker-service 删除与缓存同步简化
- **THEN** runbook 与部署清单 MUST 说明 history/device 不再依赖 worker；gateway MUST NOT 启动业务后台任务；UCG 后台任务宿主为 ucg-service

---

## cache-and-messaging-hard-dependencies

<!-- source: openspec/specs/cache-and-messaging-hard-dependencies/spec.md -->

# cache-and-messaging-hard-dependencies Specification

## Purpose
TBD - created by archiving change platform-hardening-redis-rabbitmq-service-split. Update Purpose after archive.
## Requirements
### Requirement: Redis 必须作为唯一缓存后端
系统 SHALL 将 Redis 作为 voice、device、history 相关缓存状态的唯一后端，并且 SHALL NOT 在请求处理路径执行内存缓存兜底逻辑。

#### Scenario: Redis 不可用时服务启动
- **WHEN** 服务进程启动且 Redis 连通性检查失败
- **THEN** 进程启动 SHALL 立即失败，并在启动日志中输出依赖失败原因

#### Scenario: 运行时缓存操作失败
- **WHEN** 请求处理中发生 Redis 缓存读写失败
- **THEN** 系统 SHALL 返回明确的依赖错误，且不得切换到内存兜底

### Requirement: RabbitMQ 必须作为唯一事件通道

系统 SHALL 将 RabbitMQ 作为跨服务事件通道。**UCG 审核/推荐等必需 consumer 流程** MUST 保持 RabbitMQ；history/device 的 `history.record.*` / `device.*` **fan-out 发布** 随 worker 删除而移除，不再作为必需流程。API 类进程 MAY 在 RabbitMQ 不可达时仍启动。**已删除 worker-service**，不再有「worker 启动 MUST 因 RabbitMQ 失败而失败」的进程角色。

#### Scenario: RabbitMQ 不可用时 API 服务启动

- **WHEN** API 类服务进程启动且 RabbitMQ 连通性检查失败
- **THEN** 进程 MAY 继续启动并记录 MQ 降级 Warning（ucg-service 若启用 consumer 则按 ucg 规格处理）

#### Scenario: 运行时必需 UCG 事件发布失败

- **WHEN** ucg-service 路径要求发布审核/推荐事件且 RabbitMQ 发布失败
- **THEN** 系统 MUST 按 UCG outbox/MQ 规格处理（与 worker 删除无关）

---

## cachekit-zrevrange-parse

<!-- source: openspec/specs/cachekit-zrevrange-parse/spec.md -->

# cachekit-zrevrange-parse Specification

## Purpose
TBD - created by archiving change ucg-feed-zrevrange-parse-fix. Update Purpose after archive.
## Requirements
### Requirement: ZREVRANGE WITHSCORES 解析兼容嵌套与扁平响应

`cachekit` 的 `SortedSetRevRangeWithScores`（及内部 helper）在解析 Redis `ZREVRANGE key start stop WITHSCORES` 结果时，MUST 同时支持：

- **嵌套形态**：`[[member, score], [member, score], ...]`（go-redis 9.x 经 GoFrame `Do` 常见返回）；
- **扁平形态**：`[member, score, member, score, ...]`（历史/其他 adapter 返回）。

解析输出的每个 `ZSetMemberScore.Member` MUST 为 Redis ZSET 的 member 字符串（如 `"1"`），MUST NOT 为整对 `[member,score]` 的 JSON 或数组字符串。`Score` MUST 为对应浮点分值。

#### Scenario: 嵌套响应解析为正确 member

- **WHEN** `Do(ZREVRANGE ... WITHSCORES)` 返回 `[[ "1", "1.718" ], [ "19", "1.195" ]]`
- **THEN** `SortedSetRevRangeWithScores` MUST 返回 `[{Member:"1", Score:1.718}, {Member:"19", Score:1.195}]`（score 允许浮点误差）

#### Scenario: 扁平响应仍兼容

- **WHEN** 响应为 `[ "1", "1.718", "19", "1.195" ]`
- **THEN** 解析结果 MUST 与嵌套形态等价

#### Scenario: 空 ZSET

- **WHEN** Redis 返回空数组或 nil
- **THEN** MUST 返回空 slice 且无 error

---

## care-alert-access-gate

<!-- source: openspec/specs/care-alert-access-gate/spec.md -->

# care-alert-access-gate Specification

## Purpose
TBD - created by archiving change feature-activation-care-alert. Update Purpose after archive.
## Requirements
### Requirement: cash-service MUST 提供值得留意可看合成内部契约

cash-service MUST 提供经内部密钥保护的只读接口（路径以实现为准，如 `GET /cash/internal/api/care-alert/access`），入参 MUST 含 `deviceNo` 与 `wxId`。响应 MUST 至少包含：`allowed`（bool）、`feedingQualified`（bool）、`featureActive`（bool）。合成规则 MUST 为：`feedingQualified` 取自与 App `care-alert/eligibility` 同构的 `care_alert_entry` 判定；`featureActive` 为该设备 `care_alert_smart_remind` 未过期权益 **或** 该 `wxId` 当前 VIP；`allowed = feedingQualified AND featureActive`。VIP / 功能开通 MUST NOT 将 `feedingQualified` 置 true。voice-service MUST NOT 直查 `ai_voice_cash`。

#### Scenario: 双门槛均满足

- **WHEN** 设备喂养资格合格且（有未过期值得留意权益或账号 VIP）
- **THEN** 响应 MUST `allowed=true`、`feedingQualified=true`、`featureActive=true`

#### Scenario: 仅喂养合格未开通非 VIP

- **WHEN** `feedingQualified=true` 且无有效权益且非 VIP
- **THEN** 响应 MUST `allowed=false`、`featureActive=false`

#### Scenario: 已开通但喂养不合格

- **WHEN** 设备有永久值得留意权益但喂养资格不合格
- **THEN** 响应 MUST `allowed=false`、`feedingQualified=false`、`featureActive=true`

#### Scenario: VIP 覆盖开通但不覆盖喂养

- **WHEN** 用户 VIP 有效、无设备权益、喂养不合格
- **THEN** 响应 MUST `feedingQualified=false`、`featureActive=true`、`allowed=false`

### Requirement: voice CareAlert 业务接口 MUST 经 access 双门禁

voice-service 在处理值得留意日列表（`CareAlertDaily`，含强制刷新与缓存命中路径）时 MUST 先调用 cash 值得留意 access 契约；仅当 `allowed=true` 时 MAY 返回缓存内容或触发生成。当 cash 不可达、超时或返回错误时，系统 MUST fail-closed 拒绝该次业务（MUST NOT 当作已开通放行），并记录 Warning。删除单条与反馈接口 MUST 同样调用 access 且仅在 `allowed=true` 时执行变更（与 Daily 一致，防止过期后改缓存）。

#### Scenario: 未开通拒绝 daily

- **WHEN** access 返回 `allowed=false`（例如合格未开通）
- **THEN** CareAlertDaily MUST 返回错误或空拒绝语义，MUST NOT 返回日列表内容，MUST NOT 调用 Python 生成

#### Scenario: 缓存命中仍校验

- **WHEN** 当日日缓存存在但 access 返回 `allowed=false`（如邀请已过期）
- **THEN** 系统 MUST NOT 向客户端返回该缓存内容

#### Scenario: cash 故障 fail-closed

- **WHEN** voice 调用 cash access 失败
- **THEN** CareAlertDaily MUST 失败关闭，MUST NOT 降级为跳过开通检查

---

## care-alert-cash-vip

<!-- source: openspec/specs/care-alert-cash-vip/spec.md -->

# care-alert-cash-vip Specification

## Purpose
TBD - created by archiving change vip-cash-service. Update Purpose after archive.
## Requirements
### Requirement: care-alert 经 cash-service 判定触发者 VIP

voice-service 在护理留意日缓存未命中、需要选模时，MUST 使用当前请求触发者的 `wxId`，经 HTTP 调用 cash-service 内部 VIP 接口判定是否 VIP：VIP MUST 选用 DeepSeek；非 VIP MUST 选用 Zhipu。MUST NOT 查询 `wx.is_vip`，MUST NOT 调用 device-service 的 VIP 接口，MUST NOT 使用 `deviceNo` 反查作为 VIP 或登录依据。日缓存键 MUST 仍为 `deviceNo + Asia/Shanghai` 自然日；缓存命中 MUST NOT 重跑 LLM。本路径 MUST NOT 扣减 clinic AI 配额。

#### Scenario: VIP 触发者首次生成

- **WHEN** 当日缓存未命中且 cash 返回该触发者 `isVip=true`
- **THEN** 服务 MUST 以 DeepSeek 配置调用 Python 分析并写入宝宝日缓存

#### Scenario: 非 VIP 触发者首次生成

- **WHEN** 当日缓存未命中且 cash 返回 `isVip=false`
- **THEN** 服务 MUST 以 Zhipu 配置调用 Python 分析并写入宝宝日缓存

#### Scenario: cash VIP 查询失败降级

- **WHEN** 调用 cash 内部 VIP 接口超时或失败
- **THEN** 服务 MUST 按非 VIP（Zhipu）继续生成（其它步骤成功时），MUST 打 Warning，MUST NOT 因 VIP 查询失败单独使整个 daily 失败

### Requirement: care-alert 必须携带有效 wxId

`GET/DELETE/POST` 之 `/device/api/care-alert/*` MUST 要求 `X-Internal-Wx-Id` 解析后 `wxId>0`；纯设备会话 MUST 被拒绝。

#### Scenario: 缺少 wxId

- **WHEN** care-alert 请求缺少有效 `X-Internal-Wx-Id`
- **THEN** 服务 MUST 返回错误且 MUST NOT 完成需登录的生成/删除/飞轮成功语义

### Requirement: 拆除 device 域 wx.is_vip 路径

本变更 MUST 移除（或停止注册）下列由 `care-alert-vip-by-wx` 引入的路径，使 VIP 真相源唯一落在 cash-service：`wx` 表 `is_vip` 模型字段与 DDL 执行指引、`GET /device/app/api/user/internal/vip-by-wx-id`、voice 经 device 的 `RemoteIsVipByWxID`。`hack/ddl_wx_is_vip.sql` MUST NOT 作为发布必执行项。

#### Scenario: device 不再提供 VIP 内部接口

- **WHEN** 客户端或 voice 请求已删除的 device `vip-by-wx-id` 路径
- **THEN** 该接口 MUST 不可用（404 或未注册），调用方 MUST 改用 cash internal VIP 接口

#### Scenario: 发布文档不再要求执行 is_vip DDL

- **WHEN** 按本变更更新后的 runbook 部署 VIP 能力
- **THEN** 文档 MUST 指引部署 `cash-service` / `ai_voice_cash`，MUST NOT 要求执行 `ddl_wx_is_vip.sql` 作为 VIP 前置

---

## care-alert-feature-unlock

<!-- source: openspec/specs/care-alert-feature-unlock/spec.md -->

# care-alert-feature-unlock Specification

## Purpose
TBD - created by archiving change feature-activation-care-alert. Update Purpose after archive.
## Requirements
### Requirement: 系统 MUST 提供值得留意智能提醒功能定义与设备维开通

cash-service MUST 种子（或等价确保存在）稳定功能编号 `care_alert_smart_remind`（标题面向「值得留意智能提醒」），启用开通方式 MUST 包含 `payment`、`invite_code`、`ad`。权益主体 MUST 为 `device_no`（全家共享）。付费履约 MUST 授予永久 entitlement（SKU `duration_days=0` 或等价）。邀请码与广告开通 MUST 按功能定义上的邀请/广告授予天数限时授予（种子默认 7，Admin 可改）。VIP 购买/续期 MUST NOT 写入该功能的 `feature_entitlement` 行。

#### Scenario: 付费永久开通

- **WHEN** 用户支付值得留意开通商品成功
- **THEN** 该 `device_no` 对应 `care_alert_smart_remind` entitlement 的 `expires_at` MUST 为 0（永久）或等价永久语义

#### Scenario: 邀请限时开通

- **WHEN** 用户用邀请码兑换 `care_alert_smart_remind` 且定义天数为 7
- **THEN** 该设备 entitlement MUST 为约 7 天有效，且 unlock_method 反映邀请码

#### Scenario: VIP 不写功能表

- **WHEN** 账号 VIP 履约成功
- **THEN** 系统 MUST NOT 仅为值得留意插入或更新 `feature_entitlement`

### Requirement: 值得留意邀请开通 MUST 按设备仅成功一次且原力记用户

对 `care_alert_smart_remind`，系统在邀请码兑换成功时 MUST 以 `(device_no, feature_id)` 约束该设备仅能成功邀请开通一次（任意邀请码、任意兑换账号）。重复兑码 MUST 拒绝且 MUST NOT 延长权益或再次发放获客原力。获客原力与兑换流水 MUST 仍按邀请码主人/兑换用户维度记录，MUST NOT 因设备主体而改记到「设备账号」。人×码×功能去重 MUST 继续生效。本约束 MUST NOT 应用于 `prediction_unlock`。

#### Scenario: 同设备第二次邀请拒绝

- **WHEN** 设备 D 已成功用任意码开通 `care_alert_smart_remind`，同设备另一账号再次用其它码兑换该功能
- **THEN** 系统 MUST 拒绝，MUST NOT 更新 entitlement，MUST NOT 再次给码主人加原力

#### Scenario: 预测不受设备一次限制

- **WHEN** 同一设备依次用两个不同好友码兑换 `prediction_unlock`（且同设备互兑规则通过）
- **THEN** 两次均可成功且永久条数各 +1

#### Scenario: 原力仍记用户

- **WHEN** 用户 U 兑码为设备 D 开通值得留意成功
- **THEN** 获客原力 MUST 记到码主人用户侧，兑换流水 MUST 含 U 的 `wx_id` 与 D 的 `device_no`

### Requirement: 目录与 VIP 覆盖语义 MUST 对齐预测槽模式

合成功能目录 MUST 返回 `care_alert_smart_remind` 的真实设备 `unlocked` / `expiresAt`，MUST NOT 因请求者 `isVip` 将 `unlocked` 改写为 true。业务使用态（含 access 合成）MUST 在「设备权益未过期」与「当前 isVip」之间按 OR 放行。UCG 与值得留意**喂养**资格 MUST 继续不受 isVip / 功能开通影响。

#### Scenario: VIP 使用态可看但 catalog 真实

- **WHEN** 用户 isVip=true 且设备无值得留意 entitlement
- **THEN** catalog 该项 `unlocked` MUST 为 false，但业务可看判定 MUST 为允许（在喂养合格前提下）

#### Scenario: 过期后非 VIP 不可用

- **WHEN** 邀请授予已过期且用户非 VIP
- **THEN** catalog `unlocked` MUST 为 false，业务可看判定中 featureActive MUST 为 false

---

## care-alert-feeding-eligibility

<!-- source: openspec/specs/care-alert-feeding-eligibility/spec.md -->

# care-alert-feeding-eligibility Specification

## Purpose
TBD - created by archiving change feeding-eligibility-admin-scenes. Update Purpose after archive.
## Requirements
### Requirement: 值得留意喂养资格 App API MUST 由 cash-service 提供且与 UCG 同构

**cash-service** MUST 提供登录+绑机可达的 `GET /cash/app/api/care-alert/eligibility`（或等价路径），按 `care_alert_entry` 配置在 cash 内合成并返回至少：`qualified`、`requiredDays`、`effectiveDays`、`remainingDays`，并 MAY 返回 `message`。`device_no` MUST 只信网关内部头。该路径 MUST NOT 加入 Bearer 匿名白名单。资格结果 MUST 可按日 Redis 缓存且 MUST NOT 落 MySQL 权威。history 失败 MUST fail-closed。device-service MUST NOT 作为该喂养天数资格的权威计算宿主。

#### Scenario: 连续满配置天数合格

- **WHEN** `care_alert_entry.requiredDays=2` 且连续有效日 ≥ 2
- **THEN** 响应 MUST `qualified=true`、`remainingDays=0`

#### Scenario: 不足返回剩余天数

- **WHEN** `requiredDays=2` 且 `effectiveDays=1`
- **THEN** 响应 MUST `qualified=false`、`remainingDays=1`

#### Scenario: 资格 API 归属 cash

- **WHEN** 客户端查询值得留意喂养是否达标
- **THEN** 请求 MUST 打到 cash App eligibility 路径，MUST NOT 依赖 device care-alert 接口返回喂养连续天数资格

### Requirement: 客户端 MUST 以 cash 服务端资格替代「昨日有发生」闸门

Flutter 值得留意展示流 MUST 先请求 **cash** care-alert eligibility。当 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 仍展示「值得留意」卡片，并 MUST 使用 eligibility 返回的 `effectiveDays` / `requiredDays` / `remainingDays` **由客户端拼接**进度文案，告知用户已累计的有效日（及/或剩余所需日），且 MUST NOT 调用值得留意生成或日列表数据接口。当 `qualified=true` 时，客户端 MUST 再按原逻辑请求值得留意接口并展示内容。客户端 MUST NOT 再以「上海昨日有发生」作为是否可拉取值得留意数据的前提，MUST NOT 用本地 history 自行判定权威 `qualified`。客户端 MUST NOT 被要求向用户声明「今日不计入」；服务端可选 `message` MUST NOT 作为进度数字的权威来源。本资格门槛的产品目的为促活 / 解锁更多体验，MUST NOT 被实现或文案绑定为「为保证值得留意数据准确而要求昨日有发生」。

#### Scenario: 未合格展示已累计有效日

- **WHEN** eligibility 返回 `qualified=false`、`effectiveDays=1`、`requiredDays=2`、`remainingDays=1`
- **THEN** UI MUST 展示值得留意卡片，并由客户端拼出反映已累计有效日与所需/剩余日的进度提示，且 MUST NOT 请求 care-alert daily/生成接口

#### Scenario: 合格后按原逻辑拉数

- **WHEN** eligibility 返回 `qualified=true`
- **THEN** 客户端 MUST 允许按既有逻辑请求值得留意数据接口并展示结果

#### Scenario: 资格失败 fail-closed

- **WHEN** eligibility HTTP 失败
- **THEN** 客户端 MUST NOT 当作已合格去请求生成接口

#### Scenario: 不强制解释今日排除

- **WHEN** 展示未合格进度文案
- **THEN** 文案 MUST NOT 被规格要求包含「今日不计入」或等价规则说明书表述

### Requirement: UCG 进度文案同样由客户端拼字段

UCG 入场资格展示（若展示未合格进度）MUST 使用 cash UCG eligibility 的 `effectiveDays` / `requiredDays` / `remainingDays` 由客户端拼接；MUST NOT 依赖服务端 `message` 作为已累计有效日数字的唯一来源；MUST NOT 强制向用户声明「今日不计入」。

#### Scenario: UCG 未合格展示有效日进度

- **WHEN** UCG eligibility 返回 `qualified=false` 且 `effectiveDays=3`、`requiredDays=7`
- **THEN** 客户端进度展示 MUST 能体现已累计有效日为 3（或等价「3/7」「还差 4」等由字段可推导的表述）

---

## care-alert-force-refresh

<!-- source: openspec/specs/care-alert-force-refresh/spec.md -->

# care-alert-force-refresh Specification

## Purpose
TBD - created by archiving change unify-feeding-chat-ws. Update Purpose after archive.
## Requirements
### Requirement: care-alert daily 支持 force 强刷

`GET /device/api/care-alert/daily` MUST 支持强制刷新参数（query 名实现为 `force=1` 或等价布尔约定）。当 force 为真时，在通过既有鉴权（App Bearer，网关注入 `X-Internal-Wx-Id>0`）之后，服务端 MUST 删除该 `deviceNo` 当日（Asia/Shanghai）care-alert 日缓存键（经 `cachekit` 既有 `CareAlertDailyKey` builder），再执行与未命中缓存相同的生成/single-flight 流程并写回缓存。无 force 或 force 为假时，行为 MUST 与现网一致（命中日缓存则直接返回且不重跑 LLM）。

#### Scenario: 无 force 命中缓存不重生

- **WHEN** 当日缓存已存在且请求未带 force
- **THEN** 响应 MUST 返回缓存 items，MUST NOT 再次调用 Python 生成

#### Scenario: force 清缓存后重生

- **WHEN** 当日缓存已存在且请求带 force=1（或等价真值），且 wxId>0
- **THEN** 服务端 MUST 删除当日日缓存后重新生成并返回新 items，MUST 写回日缓存

#### Scenario: force 仍要登录

- **WHEN** 请求带 force 但缺少有效 App 鉴权导致 wxId 无效
- **THEN** 服务端 MUST 拒绝请求，MUST NOT 因 force 而跳过鉴权

### Requirement: force 不引入鉴权旁路

care-alert 强刷 MUST NOT 依赖 Admin JWT、MUST NOT 新增仅 Admin 可调的「伪造 wx」接口作为唯一强刷路径。运维调试台 MUST 使用与 App 相同的 daily+force 契约。

#### Scenario: Admin token 不能代替 App 强刷

- **WHEN** 仅携带 Admin JWT 调用 care-alert daily（含 force）
- **THEN** 网关/服务 MUST 按 App 用户接口规则拒绝（与现网 Admin 不可访问 App 接口一致）

---

## care-alert-python-wire

<!-- source: openspec/specs/care-alert-python-wire/spec.md -->

# care-alert-python-wire Specification

## Purpose
TBD - created by archiving change care-alert-python-model-wire. Update Purpose after archive.
## Requirements
### Requirement: care-alert 分析请求以 model 字段传首选模型

voice-service 调用 Python `POST /v1/care-alert/analyze` 时，MUST 将 `ResolveLaneModel` 得到的非空模型配置序列化为 JSON 字段 **`model`**（对象形状：`provider`、`name`、`max_in_flight`，与 clinic/tip/intent 的 `PythonModelCfg` 一致）。MUST NOT 依赖 JSON 字段 `model_cfg` 作为 Python 可读契约。当选模结果为 omit（指针 nil）时，请求体 MUST NOT 出现 `model` 字段，以便 Python 走免费保底序。

#### Scenario: premium 或 free 有配置时传 model 对象

- **WHEN** 日缓存未命中且 `ResolveLaneModel` 返回非 nil 的 `*PythonModelCfg`
- **THEN** 发往 Python 的 JSON MUST 包含 `model` 对象，且其 `provider`/`name` 与该配置一致
- **AND** 请求体 MUST NOT 以 `model_cfg` 作为唯一模型载体（不得仅发 `model_cfg` 而省略 `model`）

#### Scenario: omit 时不传 model

- **WHEN** 日缓存未命中且选模结果为 nil（非 premium 且 lane free 为空）
- **THEN** 发往 Python 的 JSON MUST 省略 `model` 字段
- **AND** 服务 MUST 仍可调用分析接口（由 Python 保底序处理）

#### Scenario: 与兄弟仓字段名对齐

- **WHEN** 对照 `python_ai_talk` 的 `CareAlertAnalyzeRequest`
- **THEN** Go 序列化字段名 MUST 为 `model`（可选），MUST NOT 要求 Python 解析 `model_cfg`

---

## care-alert-quota-lane

<!-- source: openspec/specs/care-alert-quota-lane/spec.md -->

# care-alert-quota-lane Specification

## Purpose
TBD - created by archiving change vip-quota-joint-entitlement. Update Purpose after archive.
## Requirements
### Requirement: care-alert 独立额度 feature

voice-service MUST 提供独立 AI 额度 feature **`care_alert`**（与 `voice_ai` / `clinic_ai` 并列），含全局默认上限、per-wxId override、按月用量与 App/Admin 可读状态。care-alert 日生成路径 MUST 使用该 feature 参与统一 premium 判定，MUST NOT 再用 `clinic_ai` 代替 care-alert 额度，MUST NOT 仅根据 VIP 布尔硬切 DeepSeek/Zhipu。

#### Scenario: 非 VIP 有 care_alert 额度走正式模

- **WHEN** 日缓存未命中，触发者非 VIP，且 `care_alert` 额度 Allowed=true
- **THEN** 服务 MUST 按 `careAlert` lane 正式模型调用 Python 分析

#### Scenario: 非 VIP 且 care_alert 用尽走 free/omit

- **WHEN** 日缓存未命中，触发者非 VIP，且 `care_alert` 额度 Allowed=false
- **THEN** 服务 MUST 按 `careAlert` lane 的 free 配置选模；free 为空 MUST omit model 交 Python

#### Scenario: VIP 走正式模

- **WHEN** 日缓存未命中且触发者为 VIP
- **THEN** 服务 MUST 按 `careAlert` lane 正式模型调用 Python，MUST NOT 因 used 达旧 limit 改走 free

#### Scenario: Hub 可配置 care_alert 月度额度

- **WHEN** 管理员打开 `/device/admin/voice-admin.html` 或调用 voice Admin `ai-quota` API
- **THEN** 全局默认 MUST 可读/写 `careAlertMonthlyLimit`；用户列表 MUST 展示 `careAlert` 已用与上限，且 PUT 用户 override MUST 可提交 `careAlertMonthlyLimit`

### Requirement: care-alert 成功时非 VIP 计次

护理留意日分析在 Python 成功并写入（或准备写入）日缓存的成功路径上，若触发者 **非 VIP**，系统 MUST 对 `care_alert` 执行 consume；若触发者为 VIP，MUST NOT consume。缓存命中直接返回时 MUST NOT check/consume 额度，MUST NOT 重新选模。

#### Scenario: 非 VIP 首次生成成功扣次

- **WHEN** 非 VIP 触发者缓存未命中且生成成功
- **THEN** `care_alert` 月用量 MUST 增加 1（在 limit 允许的成功语义下）

#### Scenario: VIP 首次生成不计次

- **WHEN** VIP 触发者缓存未命中且生成成功
- **THEN** `care_alert` 月用量 MUST NOT 因本次成功增加

#### Scenario: 缓存命中不计次

- **WHEN** 当日缓存已存在
- **THEN** 服务 MUST 直接返回列表，MUST NOT consume `care_alert`

### Requirement: careAlert LLM lane 与并发

系统 MUST 提供独立 lane **`careAlert`**，具备与其它 voice lane 相同的 `provider`/`model`/`maxInFlight`/`maxWaiters` 及 free 字段，并经 aimodel 闸门 `Acquire` 控制并发（premium 或已配置 free 时按 design 约定 Acquire）。care-alert 生成 MUST 使用该 lane，MUST NOT 再借用 clinic 闸门数值并硬编码供应商切换。

#### Scenario: Admin 可配置 careAlert

- **WHEN** 管理员 GET/PUT `/voice/admin/api/llm-lanes`
- **THEN** 响应/请求 MUST 包含 `careAlert` 子对象（含正式模与 free 字段）

#### Scenario: 生成走 careAlert 闸门

- **WHEN** premium 路径调用 Python care-alert 分析
- **THEN** 服务 MUST 在调用前对该 lane 执行 Acquire（队列满语义与其它 lane 一致）

### Requirement: care-alert 仍要求有效 wxId

care-alert HTTP 接口 MUST 继续要求 `wxId>0`；MUST NOT 赋予硬件特权；MUST NOT 用 deviceNo 反查替代登录。

#### Scenario: 缺少 wxId

- **WHEN** care-alert 请求无有效 `X-Internal-Wx-Id`
- **THEN** 服务 MUST 拒绝，MUST NOT 生成或计次

---

## care-alert-unlock-client

<!-- source: openspec/specs/care-alert-unlock-client/spec.md -->

# care-alert-unlock-client Specification

## Purpose
TBD - created by archiving change feature-activation-care-alert. Update Purpose after archive.
## Requirements
### Requirement: Flutter 值得留意展示 MUST 区分喂养门槛与开通门槛

Flutter 值得留意卡片流 MUST 继续先请求 cash 喂养资格。当喂养 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 展示进度或失败提示，MUST NOT 请求值得留意 daily，MUST NOT 仅因未开通跳转开通中心。当喂养合格且设备该功能未有效开通且账号非 VIP 时，客户端 MUST 展示可点击的开通引导态，用户点击后 MUST 进入功能开通中心（宜定位到 `care_alert_smart_remind`）。当喂养合格且（目录显示已开通未过期 **或** 当前 isVip）时，客户端 MUST 按既有逻辑请求值得留意 daily 并展示。

#### Scenario: 未喂养合格不进开通中心

- **WHEN** eligibility `qualified=false`
- **THEN** UI MUST 展示喂养进度类文案，点击 MUST NOT 作为开通中心的主路径要求

#### Scenario: 合格未开通点击进开通中心

- **WHEN** eligibility `qualified=true` 且非 VIP 且 catalog 中 `care_alert_smart_remind` 未有效开通
- **THEN** 用户点击值得留意引导区域 MUST 导航至开通中心

#### Scenario: VIP 或已开通拉 daily

- **WHEN** eligibility `qualified=true` 且（isVip 或功能已开通未过期）
- **THEN** 客户端 MUST 允许请求 daily 并展示结果

### Requirement: 开通中心 MUST 展示值得留意项且 VIP 覆盖与预测模式一致

功能目录拉取后，开通中心 MUST 展示 `care_alert_smart_remind`（功能启用时）。有效开通展示 MUST 使用「设备 `unlocked` 未过期 **或** isVip」的 OR 语义（与非预测项既有 `isFeatureEffectivelyUnlocked` 一致）。catalog 的 `unlocked` MUST 仍反映设备真实权益。客户端 MUST NOT 用 isVip 绕过喂养资格。

#### Scenario: 非 VIP 未开通显示可开通

- **WHEN** 非 VIP 用户打开开通中心且设备未开通值得留意
- **THEN** 该项 MUST 展示为未开通并可走支付/邀请（及若启用的广告）开通

#### Scenario: VIP 覆盖显示

- **WHEN** VIP 用户打开开通中心且设备无值得留意权益行
- **THEN** 有效开通态 MUST 视为可用，且 MUST NOT 要求客户端把 catalog `unlocked` 改写成服务端已开通

---

## care-alert-vip-selection

<!-- source: openspec/specs/care-alert-vip-selection/spec.md -->

# care-alert-vip-selection Specification

## Purpose
TBD - created by archiving change care-alert-vip-by-wx. Update Purpose after archive.
## Requirements
### Requirement: care-alert 必须携带有效 wxId

voice-service 上经 gateway 暴露的 care-alert HTTP（`GET /device/api/care-alert/daily`、`DELETE /device/api/care-alert/daily/item`、`POST /device/api/care-alert/feedback`）MUST 要求请求携带有效 `X-Internal-Wx-Id`（解析后 `wxId > 0`）。纯设备会话（`wxId=0` 或缺失头）MUST 被拒绝。鉴权路径 MUST NOT 使用按 `deviceNo` 反查 `wxId` 作为登录替代。

#### Scenario: 缺少或无效 wxId

- **WHEN** 客户端调用任一 care-alert 接口且 `X-Internal-Wx-Id` 缺失、非正整数或为 `0`
- **THEN** 服务 MUST 返回错误且 MUST NOT 读写日缓存生成路径中的 LLM，也 MUST NOT 完成需登录的删除/飞轮成功语义

#### Scenario: 有效 wxId 与 deviceNo

- **WHEN** 请求同时具备 `wxId > 0` 与合法 `deviceNo`
- **THEN** 服务 MAY 继续执行宝宝维度的缓存/忽略/飞轮逻辑；VIP 判定主键 MUST 为该 `wxId`

### Requirement: 按触发者 VIP 选模且与日缓存维度分离

护理留意首次日生成（缓存未命中）时，voice-service MUST 按**当前请求触发者**的 `wxId` 查询账号 VIP：VIP 为真 MUST 选用 DeepSeek；非 VIP MUST 选用 Zhipu。日缓存键 MUST 仍为 `deviceNo + Asia/Shanghai` 自然日；缓存命中 MUST NOT 因读缓存再次查询 VIP 或重跑 LLM。本路径 MUST NOT 扣减 clinic AI 配额。

#### Scenario: VIP 触发者首次生成

- **WHEN** 某 `deviceNo` 当日缓存未命中，且触发者 `wxId` 经 device 契约判定为 VIP
- **THEN** 服务 MUST 以 DeepSeek 配置调用 Python 分析，并将结果写入该 `deviceNo` 当日缓存

#### Scenario: 非 VIP 触发者首次生成

- **WHEN** 某 `deviceNo` 当日缓存未命中，且触发者判定为非 VIP
- **THEN** 服务 MUST 以 Zhipu 配置调用 Python 分析并写入当日缓存

#### Scenario: 他看护命中缓存

- **WHEN** 同 `deviceNo` 同日已有缓存，另一 `wxId`（无论是否 VIP）再次 GET
- **THEN** 服务 MUST 直接返回缓存列表，MUST NOT 按后者 VIP 重新选模或重跑 LLM

### Requirement: VIP 查询失败降级为非 VIP

voice-service 在 care-alert 生成路径查询账号 VIP 时，若 device 契约超时、不可达、或返回不可用错误，MUST 将本次生成视为非 VIP（Zhipu），MUST NOT 因此使整个 daily 生成失败（设备校验与 Python 分析等其它错误仍按原语义失败）。实现 MUST 记录可观测 Warning。

#### Scenario: device VIP 接口失败

- **WHEN** 缓存未命中且查询 `vip-by-wx-id`（或等价契约）失败
- **THEN** 服务 MUST 选用 Zhipu 继续生成（若其它步骤成功），且 MUST 打出 Warning 级日志表明 VIP 降级

---

## cash-admin-grant-revoke

<!-- source: openspec/specs/cash-admin-grant-revoke/spec.md -->

# cash-admin-grant-revoke Specification

## Purpose
TBD - created by archiving change revoke-admin-grant-hub-tiles. Update Purpose after archive.
## Requirements
### Requirement: 系统仅在最近一笔为手工授且仍有效时允许撤销 VIP

cash-service MUST 提供 `POST /cash/admin/api/vip/entitlements/revoke`，鉴权 MUST 校验 `X-Admin-Password`。请求体 MUST 含 `wxId`（>0）。系统 MUST 仅在同时满足以下条件时撤销：

1. 该 `wxId` 的 `vip_entitlement.expire_at` 大于当前时间；
2. 该 `wxId` 最近一笔 `vip_order.status=paid` 的 `channel` 为 `admin`。

否则 MUST 拒绝，MUST NOT 修改权益或订单。

#### Scenario: 最近一笔为手工授且未过期

- **WHEN** 管理员对仍有效、且最近 paid 订单 `channel=admin` 的 `wxId` 提交撤销
- **THEN** 系统 MUST 将该 `expire_at` 设为当前时间，且该笔订单 `status` MUST 变为 `revoked`

#### Scenario: 最近一笔为付费渠道

- **WHEN** 最近 paid 订单 `channel` 为 `alipay` 或 `apple_iap`
- **THEN** 系统 MUST 拒绝撤销，MUST NOT 修改 `expire_at`

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝且 MUST NOT 写库

### Requirement: 撤销 VIP MUST 直接过期且 MUST NOT 按天数回退

撤销成功时系统 MUST 将 `vip_entitlement.expire_at` 设为不大于当前时间，使 `IsVip` 立即为 false。系统 MUST NOT 使用 `ShrinkEntitlement` 或按 `durationDays` 扣减。系统 MUST 使用订单状态 `revoked`，MUST NOT 把该订单标为 `refunded`。

#### Scenario: 付费剩余被一并清掉

- **WHEN** 账号先有付费剩余、随后手工续期，最近 paid 为 `admin`，管理员确认撤销
- **THEN** 系统 MUST 将到期设为当前时间，MUST NOT 只扣手工授的天数

#### Scenario: 重复撤销

- **WHEN** 目标 admin 订单已是 `revoked`，或权益已过期
- **THEN** 系统 MUST 返回错误，MUST NOT 再改写其它 paid 订单

### Requirement: 功能手工授撤销 MUST 按主体校验最近 admin 订单

cash-service MUST 提供 `POST /cash/admin/api/feature/grants/revoke`。请求 MUST 含 `featureId`，并按 `activation_subject` 要求 `wxId` 或 `deviceNo`。系统 MUST 找到该功能该主体最近一笔 `feature_order.status=paid`；仅当其 `channel=admin` 且对应权益仍有效时，MUST 将权益到期设为当前时间，并将该订单标为 `revoked`。`prediction_unlock` 与数量累加类 MUST NOT 提供此撤销。

#### Scenario: 撤销 user 主体功能

- **WHEN** 管理员对仍有效、最近 paid 为 `admin` 的账号维功能提交 `wxId` 与 `featureId`
- **THEN** 系统 MUST 使该账号该功能立即过期，并标记该 admin 订单 `revoked`

#### Scenario: 主体标识缺失

- **WHEN** 功能为 `device` 主体且请求没有 `deviceNo`
- **THEN** 系统 MUST 拒绝，MUST NOT 改权益

---

## cash-admin-hub-tiles

<!-- source: openspec/specs/cash-admin-hub-tiles/spec.md -->

# cash-admin-hub-tiles Specification

## Purpose
TBD - created by archiving change revoke-admin-grant-hub-tiles. Update Purpose after archive.
## Requirements
### Requirement: 开通管理页顶部 MUST 平铺 VIP、上架功能与群二维码

静态页 `cash-feature-admin.html` MUST 在顶部平铺可点选的项：固定的 VIP、`feature_def.status=1` 的各功能、以及群二维码。`prediction_unlock` MUST NOT 出现在平铺中。同一时刻 MUST 只展示当前选中项的相关内容，MUST NOT 在未选中时同时展开全部功能定义表、全部 SKU 表与 VIP 套餐表。

#### Scenario: 点选一项后只看该项

- **WHEN** 已登录运维点击「值得留意」
- **THEN** 页面 MUST 展示该功能的定义、该功能的 SKU、开通快照与手工授，且 MUST NOT 同时展示 VIP 套餐表或其它功能的编辑表

#### Scenario: 预测槽位不在平铺中

- **WHEN** 页面加载功能平铺
- **THEN** 平铺 MUST NOT 包含 `prediction_unlock`

### Requirement: VIP 面板 MUST 收纳套餐、权益列表、授予与撤销

选中 VIP 时，页面 MUST 提供现有 VIP 套餐编辑，以及权益列表（含 `channel`、`grantReason`）、手工授表单。列表中 `channel=admin` 且仍有效的行 MUST 提供撤销，提交 MUST 调用 VIP 撤销 API。成功后 MUST 刷新列表。浏览器 MUST NOT 发送 `X-Admin-Password`。`cash-vip-admin.html` MUST NOT 再作为与本面板并行的第二套可写入口（可改为跳转到本页并选中 VIP）。

#### Scenario: 从 VIP 行撤销手工授

- **WHEN** 运维在 VIP 列表对 `channel=admin` 且有效的行确认撤销
- **THEN** 页面 MUST 调用撤销 API，刷新后该行 MUST 显示为已过期或不再提供撤销

#### Scenario: 付费行没有撤销

- **WHEN** 列表行渠道为支付宝或 Apple
- **THEN** 该行 MUST NOT 提供撤销按钮

### Requirement: 功能面板 MUST 支持对该功能手工授与撤销

选中某一上架功能时，手工授表单的功能编号 MUST 固定为该项，MUST NOT 再让运维从全量下拉里改选。快照中 `unlock_method=admin` 且仍有效的行 MUST 提供撤销，并调用功能撤销 API。群二维码 MUST 仅在其自身面板中编辑，MUST NOT 作为某个功能的子表单。

#### Scenario: 在功能面板授功能

- **WHEN** 运维在已选功能面板填写主体标识、期限与非空理由并提交
- **THEN** 系统 MUST 以该平铺项的 `featureId` 调用授功能 API，并刷新该功能快照

#### Scenario: 群二维码独立面板

- **WHEN** 运维点选群二维码
- **THEN** 页面 MUST 展示上传与有效期，且 MUST NOT 要求选择 `featureId`

---

## cash-admin-manual-grant

<!-- source: openspec/specs/cash-admin-manual-grant/spec.md -->

# cash-admin-manual-grant Specification

## Purpose
TBD - created by archiving change add-admin-manual-grant. Update Purpose after archive.
## Requirements
### Requirement: vip_order 与 feature_order SHALL 具备 grant_reason 列

`ai_voice_cash` 表 `vip_order` 与 `feature_order` MUST 各具备 `grant_reason` 列（字符串，建议 `VARCHAR(256) NOT NULL DEFAULT ''`）。支付渠道创建的订单 MUST 将该列置为空串。Admin 手工授予写入的订单 MUST 写入非空授权理由。cash-service 启动 `EnsureSchema`（或等价迁移）MUST 幂等补齐该列。

#### Scenario: 新库建表含列

- **WHEN** 空库执行 EnsureSchema
- **THEN** `vip_order` 与 `feature_order` MUST 包含 `grant_reason` 列

#### Scenario: 已有库补列

- **WHEN** 旧库缺少该列且服务启动 EnsureSchema
- **THEN** 系统 MUST 添加该列且 MUST NOT 因重复添加而 Fatal（幂等）

### Requirement: cash-service SHALL 提供 Admin 手工授 VIP API

cash-service MUST 提供 `POST /cash/admin/api/vip/entitlements/grant`，鉴权 MUST 校验网关注入的 `X-Admin-Password`。请求体 MUST 含 `wxId`（>0）、`durationDays`（整数 ≥1）、`reason`（trim 后非空，长度不得超过列上限）。成功时 MUST：

1. 插入一行 `vip_order`：`channel=admin`、`amount_fen=0`、`status=paid`、`paid_at` 为当前时间、`grant_reason` 为理由、`product_code` 为默认 VIP 商品码（与一期商品一致，如 `vip_monthly_19`）；
2. 按与支付相同的续期语义调用权益续期：`new_expire = max(now, current_expire) + durationDays * 86400`。

订单插入与权益写入 MUST 在同一成功路径内完成（推荐同一事务）；任一步失败 MUST 向调用方返回错误且 MUST NOT 声称授予成功。外部支付宝/Apple 回调 MUST NOT 被设计为创建 `channel=admin` 订单的正常路径。

#### Scenario: 合法授 VIP

- **WHEN** 管理员以正确口令提交有效 `wxId`、`durationDays=30`、非空 `reason`
- **THEN** 系统 MUST 写入带 `grant_reason` 的 admin `vip_order`，且该 `wxId` 的 `vip_entitlement.expire_at` MUST 按续期公式延长

#### Scenario: 理由为空拒绝

- **WHEN** 请求 `reason` 为空或仅空白
- **THEN** 系统 MUST 拒绝授予，MUST NOT 插入订单，MUST NOT 修改权益

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** 系统 MUST 拒绝且 MUST NOT 写入订单或权益

### Requirement: cash-service SHALL 提供 Admin 手工授功能 API

cash-service MUST 提供 `POST /cash/admin/api/feature/grants`（路径以实现为准，须落在 `/cash/admin/api/` 下），鉴权同 VIP Admin。请求 MUST 含 `featureId`、非空 `reason`，并按该功能 `activation_subject`：

- `user`：MUST 要求 `wxId` >0，写入账号维权益（或该功能既定 Grant 路径）；
- `device`：MUST 要求非空 `deviceNo`，写入设备维权益或预测数量权威。

`durationDays` 与 `grantQuantity` 按该功能 `grant_kind` / 产品语义使用：权益类 MUST 支持续期天数（与支付 Grant 一致；`durationDays=0` 表示永久仅当该功能既有 Grant 允许且请求显式传递——默认实现可要求权益类 `durationDays≥1`，预测 `allowed_count_delta` 则 MUST 要求 `grantQuantity≥1`）。成功时 MUST 插入 `feature_order`：`channel=admin`、`amount_fen=0`、`status=paid`、`grant_reason` 非空，并完成对应 Grant。

#### Scenario: 授 user 主体功能

- **WHEN** 管理员对 `activation_subject=user` 的功能提交有效 `wxId`、`durationDays≥1`、非空 `reason`
- **THEN** 系统 MUST 写入 admin `feature_order` 与账号维权益续期/授予

#### Scenario: 授 device 主体功能缺 deviceNo

- **WHEN** 功能为 `device` 主体且请求未提供 `deviceNo`
- **THEN** 系统 MUST 拒绝，MUST NOT 落库订单或权益

#### Scenario: 功能授理由必填

- **WHEN** `reason` 为空
- **THEN** 系统 MUST 拒绝授予

### Requirement: Hub VIP 权益页 SHALL 支持手工授 VIP

静态页 `cash-vip-admin.html` MUST 提供手工授 VIP 表单：至少 `wxId`、`durationDays`、`reason`（必填）。提交 MUST 调用授 VIP Admin API（经 `AdminCommon.adminFetch`）。成功后 MUST 刷新权益列表。页面 MUST NOT 在浏览器发送 `X-Admin-Password`。

#### Scenario: 运维从 VIP 页授 VIP

- **WHEN** 已登录 Hub 的运维填写有效表单并提交
- **THEN** 系统 MUST 完成授予，且刷新后列表 MUST 能反映该 `wxId` 的有效权益（或更新后的到期时间）

### Requirement: Hub 开通功能页 SHALL 支持手工授功能

静态页 `cash-feature-admin.html` MUST 提供手工授功能表单：至少 `featureId`、按主体的目标标识、期限或数量、`reason`（必填）。提交 MUST 调用授功能 Admin API。成功后 MUST 可刷新开通快照（若页内已有快照区）。

#### Scenario: 运维从开通功能页授功能

- **WHEN** 已登录运维提交合法功能授表单
- **THEN** 系统 MUST 完成授予且 MUST 落库带 `grant_reason` 的 admin `feature_order`

### Requirement: Admin 手工授 API SHALL 不计入 App usage 统计

`/cash/admin/api/vip/entitlements/grant` 与功能授 Admin 路径 MUST 视为运维管理通道，MUST NOT 按 App 对外接口计入 usage；实现 MUST NOT 仅为这些路径修改 App 侧 `maintenance_skip` 假定。

#### Scenario: 管理端手工授

- **WHEN** 运维通过 Hub 调用手工授 API
- **THEN** 系统 MUST NOT 将其记为需计入的 App 功能使用

---

## cash-admin-revenue

<!-- source: openspec/specs/cash-admin-revenue/spec.md -->

# cash-admin-revenue Specification

## Purpose
TBD - created by archiving change cash-admin-revenue-stats. Update Purpose after archive.
## Requirements
### Requirement: 管理端 MUST 提供手填进账与出账

cash-service MUST 在与 `vip_order` 同库的新表中保存手填流水。每条 MUST 包含方向（`in` 进账或 `out` 出账）、非空名称、大于等于 1 的金额（分）和发生时间（unix 秒）。名称 MUST NOT 要求唯一。Admin MUST 能新增、按 id 修改上述字段、按 id 物理删除、并按发生时间倒序列出。方向、空名称或金额小于 1 MUST 拒绝且 MUST NOT 写库。未传发生时间时 MUST 使用写入时刻。

#### Scenario: 新增一笔出账

- **WHEN** 管理员提交方向 `out`、名称「模型费用」、金额 5000 分
- **THEN** 系统 MUST 写入该行，且列表中能按名称与金额读回

#### Scenario: 金额为 0

- **WHEN** 管理员提交金额 0
- **THEN** 系统 MUST 拒绝，且 MUST NOT 插入行

### Requirement: 收益合计 MUST 只统计实付并按功能名拆开

管理端收益合计 MUST 只把 `status=paid` 且 `channel` 为 `alipay` 或 `apple_iap` 的订单 `amount_fen` 算作付费。`channel=admin`、`refunded` 以及其他非 paid 状态 MUST NOT 计入。VIP 付费 MUST 为一个合计。功能付费 MUST 按功能拆开：展示名为功能标题，标题为空时 MUST 用功能编号；对不上功能商品的订单 MUST 单独成组并以商品编码为名称。合计 MUST 在读取时对订单求和，MUST NOT 把订单金额写入手填表。

#### Scenario: 功能付费按名称列出

- **WHEN** 两个不同功能各有一笔已支付的支付宝或 Apple 订单
- **THEN** 合计 MUST 返回两行，每行含该功能名称与对应金额之和

#### Scenario: 手工授与退款不计入

- **WHEN** 存在 `channel=admin` 的 paid 订单，以及 `status=refunded` 的订单
- **THEN** 这两类订单的金额 MUST NOT 出现在 VIP 或功能付费合计中

### Requirement: 差额 MUST 用手填与付费一起算出

在同一时间范围内，系统 MUST 计算差额 = VIP 付费 + 各功能付费之和 + 手填进账之和 − 手填出账之和。查询参数 `start`、`end`（unix 秒）均可省略。省略两者时 MUST 统计全部。给出 `start` 时订单 MUST 满足 `paid_at >= start`、手填 MUST 满足发生时间 `>= start`。给出 `end` 时对应时间 MUST `< end`。订单过滤与手填过滤 MUST 使用同一组参数。

#### Scenario: 对比成本与收益

- **WHEN** VIP 实付 1900 分、某功能实付 600 分、手填进账 0、手填出账 5000 分，且未传时间范围
- **THEN** 差额 MUST 为 1900 + 600 − 5000

#### Scenario: 时间范围同时作用于两边

- **WHEN** 管理员传入 `start` 与 `end`
- **THEN** 付费合计 MUST 只含该窗口内 `paid_at` 的实付，手填合计 MUST 只含该窗口内发生的流水

### Requirement: 开通功能管理 MUST 提供收益统计面板

`cash-feature-admin.html` 顶栏 MUST 在现有入口之外提供「收益统计」。选中后 MUST 只显示该面板，并展示 VIP 合计、按功能名的付费列表、手填进账合计、手填出账合计、差额，以及手填流水的增删改查。页面 MUST 说明付费金额是用户支付的标价，渠道抽成与运营成本须记为出账。金额输入 MUST 以分为单位。

#### Scenario: 打开收益统计

- **WHEN** 管理员在开通功能管理选中「收益统计」
- **THEN** 页面 MUST 展示上述合计与手填列表，且 MUST NOT 改动订单或权益数据

---

## cash-duration-days

<!-- source: openspec/specs/cash-duration-days/spec.md -->

# cash-duration-days Specification

## Purpose
TBD - created by archiving change no-permanent-duration. Update Purpose after archive.
## Requirements
### Requirement: 启动种子 MUST NOT 覆盖已有 VIP 天数

cash-service `EnsureSchema` 在 `vip_product` 行已存在时 MUST NOT 更新 `duration_days`。首次插入该商品时 MUST 仍写入默认 30。`product_code` MUST 保持 `vip_monthly_19`。

#### Scenario: 重启不改运维天数

- **WHEN** `vip_product.duration_days` 已被改为 7，且 cash-service 再次执行 EnsureSchema
- **THEN** 该列 MUST 仍为 7

#### Scenario: 空库首次种子

- **WHEN** 不存在 `vip_monthly_19` 且执行 EnsureSchema
- **THEN** 系统 MUST 插入 `duration_days=30` 的该商品行

### Requirement: VIP 天数 MUST 大于等于 1

Admin 更新 VIP 商品时 `durationDays` MUST ≥1，否则 MUST 拒绝且 MUST NOT 写库。支付履约调用续期时，若商品天数 `<1`，系统 MUST 拒绝续期，MUST NOT 把天数替换为 30，MUST NOT 把 `expire_at` 写成 0。

#### Scenario: 后台保存 0 天

- **WHEN** 管理员将 VIP 有效天数提交为 0
- **THEN** 系统 MUST 拒绝，且库中原天数 MUST 不变

#### Scenario: 履约遇到非正天数

- **WHEN** 一笔 VIP 订单支付成功，但商品 `duration_days` 小于 1
- **THEN** 系统 MUST NOT 延长 `vip_entitlement`，MUST NOT 把该次续期当成 30 天

### Requirement: 功能授予天数 MUST 大于等于 1

功能 SKU 的 `duration_days`、功能定义的邀请授予天数与广告授予天数，在 Admin 保存时 MUST 均 ≥1。支付履约、邀请码兑换、广告开通 MUST NOT 在天数为 0 时写入永久权益（`expires_at=0`）。天数 `<1` 时 MUST 拒绝开通。

#### Scenario: 保存功能套餐天数为 0

- **WHEN** 管理员把某功能 SKU 的有效天数设为 0
- **THEN** 系统 MUST 拒绝保存

#### Scenario: 邀请天数配置为 0 时兑码

- **WHEN** 功能定义的邀请授予天数为 0，用户兑换邀请码
- **THEN** 系统 MUST 拒绝开通，MUST NOT 把权益 `expires_at` 写成 0

### Requirement: Hub MUST NOT 再把 0 标成永久

`cash-feature-admin.html` 中 VIP 有效天数、功能 SKU 有效天数、邀请/广告授予天数的输入 MUST 要求至少 1。页面 MUST NOT 再显示「0=永久」。

#### Scenario: 打开 VIP 天数输入

- **WHEN** 运维打开 VIP 套餐面板
- **THEN** 有效天数控件 MUST 不允许小于 1，且 MUST NOT 出现「0=永久」说明

### Requirement: 历史永久行 MUST NOT 被本变更批量改写

本变更 MUST NOT 执行把已有 `expires_at=0` 批量改为当前时间的迁移。已存在的此类权益行 MUST 仍按变更前的读规则判断是否有效。

#### Scenario: 上线时已有永久权益

- **WHEN** 某账号功能权益在本变更部署前 `expires_at=0`
- **THEN** 部署本身 MUST NOT 把该行改写为已过期

---

## cash-service-runtime

<!-- source: openspec/specs/cash-service-runtime/spec.md -->

# cash-service-runtime Specification

## Purpose
TBD - created by archiving change vip-cash-service. Update Purpose after archive.
## Requirements
### Requirement: cash-service 独立进程与配置

系统 MUST 提供独立进程 `cash-service`，默认配置文件为 `manifest/config/config.cash-service.yaml`（或等价），MUST NOT 将现金域专属配置回流到主配置 `manifest/config/config.yaml`。进程启动 MUST 能通过环境变量 **`CASH_DB_LINK`** 覆盖其 GoFrame `database.default` 连接。

#### Scenario: 默认配置指向本域

- **WHEN** 未设置 `GF_GCFG_FILE` 启动 `cmd/cash-service`
- **THEN** 运行时 MUST 使用 cash-service 专用配置文件，且 MUST NOT 依赖共享主配置承载支付/VIP 业务项

#### Scenario: 数据库连接可被 CASH_DB_LINK 覆盖

- **WHEN** 部署环境设置 `CASH_DB_LINK` 为 `ai_voice_cash` 的 DSN
- **THEN** cash-service MUST 使用该连接访问本域库，MUST NOT 直连 `ai_voice_device` / `ai_voice_voice` 等他域库承载 VIP 表

### Requirement: 部署清单包含 cash-service

微服务 Compose、`.env.example` 与发布 runbook MUST 登记 `cash-service`（镜像/服务名、`CASH_DB_LINK`、`CASH_SERVICE_ADDR`、他服所需的 `CASH_SERVICE_URL`）。

#### Scenario: Compose 可拉起 cash-service

- **WHEN** 按微服务 compose 启动含 cash-service 的栈且已配置 `CASH_DB_LINK`
- **THEN** 容器 MUST 监听配置地址（建议 `:9806`）并提供可探测的 HTTP 健康/API 入口（与现网其它 `*-service` 惯例一致）

---

## cash-vip-admin-api

<!-- source: openspec/specs/cash-vip-admin-api/spec.md -->

# cash-vip-admin-api Specification

## Purpose
TBD - created by archiving change vip-admin-console. Update Purpose after archive.
## Requirements
### Requirement: cash-service SHALL 提供分页只读 VIP 权益 Admin API

cash-service MUST 提供 `GET /cash/admin/api/vip/entitlements`，鉴权 MUST 校验网关注入的 `X-Admin-Password`（与 cash Admin 口令配置一致）。该接口 MUST 为只读，MUST NOT 修改 `vip_entitlement`、`vip_order` 或 `vip_product`。Query MUST 支持 `page`（默认 1）、`pageSize`（默认 20，最大 200），以及可选精确过滤 `wxId`。响应 MUST 含 `list`、`page`、`pageSize`、`total`。

#### Scenario: 合法管理员分页拉取

- **WHEN** 请求携带正确 `X-Admin-Password` 且未指定 `wxId`
- **THEN** cash-service MUST 返回分页后的权益列表与 `total`

#### Scenario: 口令错误

- **WHEN** 请求未携带或携带错误的 `X-Admin-Password`
- **THEN** cash-service MUST 拒绝该请求且 MUST NOT 返回权益数据

#### Scenario: 按 wxId 过滤

- **WHEN** 请求携带合法正整数 `wxId` 与正确口令
- **THEN** 响应 `list` MUST 仅包含该 `wxId` 的权益行（0 或 1 条）

### Requirement: 权益列表范围 SHALL 包含已过期 VIP

列表主数据源 MUST 为 `vip_entitlement` 的全部行，MUST NOT 默认过滤为仅 `expire_at > now`。每行 MUST 提供当前是否仍有效的布尔字段（如 `isVip`：当且仅当 `expire_at` 晚于当前时间），以及 `expireAt`（unix 秒）与可派生的剩余秒数（如 `remainingSeconds`，过期时可为 0 或负数）。

#### Scenario: 已过期权益仍出现在列表

- **WHEN** 某 `wx_id` 存在权益且 `expire_at` 已不晚于当前时间
- **THEN** 该行 MUST 仍可出现在列表中，且有效布尔字段 MUST 为 false

#### Scenario: 未过期权益标记有效

- **WHEN** 某权益 `expire_at` 晚于当前时间
- **THEN** 该行有效布尔字段 MUST 为 true

### Requirement: 激活金额 SHALL 取最近一次已支付订单金额

每条权益行的激活金额（如 `lastPaidAmountFen`）MUST 取该 `wx_id` 在 `vip_order` 中 `status=paid` 的订单，按 `paid_at DESC`（相等时可按 `id DESC`）的最近一条之 `amount_fen`。同条 MUST 附带该订单的 `channel` 与 `paidAt`（字段名以实现为准，语义一致；无 paid 订单时渠道与支付时间 MUST 为空或 0）。MUST NOT 使用未支付订单或商品现价冒充激活金额。若无任何 paid 订单，金额 MUST 为 0 或空语义，且 MUST NOT 伪造渠道支付时间。

#### Scenario: 多次支付取最近一次

- **WHEN** 同一 `wx_id` 存在多笔 `status=paid` 订单且 `paid_at` 不同
- **THEN** 激活金额与渠道、支付时间 MUST 对应 `paid_at` 最新的那一笔

#### Scenario: 无已支付订单

- **WHEN** 某权益行对应 `wx_id` 不存在 `status=paid` 的订单
- **THEN** 激活金额 MUST 不以未支付订单填充，响应 MUST 可区分「无实付」

### Requirement: gateway-app SHALL 反代并鉴权 cash Admin API

gateway-app-server MUST 将 `/cash/admin/api/*` 反代至 `CASH_SERVICE_URL` 指向的 cash-service。`IsGatewayAdminAPIPath` MUST 将 `/cash/admin/api/` 前缀视为须 Admin JWT 的管理 API。通过 Admin JWT 后，网关 MUST 向下游注入 `X-Admin-Password`（优先 `CASH_ADMIN_PASSWORD`，否则回退 `GATEWAY_APP_ADMIN_PASSWORD`）。浏览器客户端 MUST NOT 被要求自带 `X-Admin-Password`。device-service 及其它非 cash 进程 MUST NOT 直查 `ai_voice_cash` 以提供本列表。

#### Scenario: Hub 已登录调用 Admin API

- **WHEN** 运维持有效 Admin JWT 请求 `GET /cash/admin/api/vip/entitlements`
- **THEN** gateway-app MUST 反代至 cash-service 并注入口令头，cash-service MUST 可完成鉴权

#### Scenario: 无 Admin JWT

- **WHEN** 未携带有效 Admin JWT 请求 `/cash/admin/api/vip/entitlements`
- **THEN** gateway-app MUST 拒绝该请求（不得当作普通 App Bearer 用户接口放行）

### Requirement: cash Admin API SHALL 不计入 App usage 统计

`/cash/admin/api/*` MUST 视为运维管理通道，MUST NOT 按 App 对外接口计入 usage 统计策略；实现 MUST NOT 仅为该 Admin 路径修改 App 侧 `maintenance_skip` 假定。

#### Scenario: 管理端拉取列表

- **WHEN** 运维通过 Hub 调用 VIP 权益 Admin API
- **THEN** 系统 MUST NOT 将其记为需计入的 App 功能使用（与既有 voice/sim admin 管理通道一致）

---

## cash-vip-admin-ui

<!-- source: openspec/specs/cash-vip-admin-ui/spec.md -->

# cash-vip-admin-ui Specification

## Purpose
TBD - created by archiving change vip-admin-console. Update Purpose after archive.
## Requirements
### Requirement: Hub SHALL 提供 VIP 权益只读管理页

系统 MUST 在 `gateway-app-server` 注册静态页 `/device/admin/cash-vip-admin.html`（静态文件 `resource/public/cash-vip-admin.html`）。页面 MUST 使用 `resource/public/admin-common.js` 的 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch`（或等价封装）初始化。页面 MUST 为只读：MUST NOT 提供改价、退款、手工开通或编辑权益的表单/按钮。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 已登录 Hub 打开页面

- **WHEN** 管理员已在运维 Hub 登录并访问 `/device/admin/cash-vip-admin.html`
- **THEN** 主内容区 MUST 可见且 MUST 能发起对 `/cash/admin/api/vip/entitlements` 的加载请求

#### Scenario: 未登录访问

- **WHEN** 管理员未登录即打开该静态页
- **THEN** 页面 MUST 按 admin-common 惯例引导登录或隐藏需鉴权的主内容，MUST NOT 在浏览器侧携带运维口令头

### Requirement: Admin Hub SHALL 登记 cash-vip-admin 模块

`resource/public/admin-modules.js` MUST 保留模块登记：`id: cash-vip-admin`（或等价稳定 id）、`pagePath` 为 `/device/admin/cash-vip-admin.html`、`showInNav: false`。Hub 主导航 MUST NOT 展示「VIP 权益」独立入口。`RegisterAdminStaticPages`（或等价单点注册表）MUST 仍包含该 `pagePath`。PR MUST NOT 仅在主网关注册该页而 App 网关不可见。权益只读页的运维入口 MUST 为开通功能管理页文案为「VIP列表」的按钮（或等价控件）并导航至上述 `pagePath`。

#### Scenario: Hub 导航不展示 VIP 模块

- **WHEN** 运维已登录 `/device/admin` Hub
- **THEN** 模块列表 MUST NOT 以独立导航项展示「VIP 权益」

#### Scenario: 静态页仍可直达

- **WHEN** 已登录管理员打开 `/device/admin/cash-vip-admin.html`
- **THEN** 主内容区 MUST 可见且 MUST 能发起对 `/cash/admin/api/vip/entitlements` 的加载请求

### Requirement: VIP 管理页 SHALL 展示权益与激活金额列

`cash-vip-admin.html` MUST 将 `GET /cash/admin/api/vip/entitlements` 结果以表格或等价列表展示，列至少含：`wxId`、是否仍有效、到期时间、剩余时间（过期时 MUST 展示「已过期」或等价文案，而非伪装为正剩余）、激活金额、支付渠道、最近支付时间。页面 MUST 支持分页；宜提供按 `wxId` 查询。激活金额展示 MUST 与 API 的最近一次 paid `amount_fen` 语义一致（页内须能识别单位为分或换算为元）。

#### Scenario: 列表含已过期行

- **WHEN** API 返回 `isVip=false` 且剩余秒数 ≤0 的行
- **THEN** UI MUST 仍展示该行，且剩余时间文案 MUST 表明已过期

#### Scenario: 展示最近实付金额

- **WHEN** API 返回某行 `lastPaidAmountFen` 为非零
- **THEN** UI MUST 在激活金额列展示对应该值的金额（不得展示商品现价替代）

---

## catalog-activatable-total

<!-- source: openspec/specs/catalog-activatable-total/spec.md -->

# catalog-activatable-total Specification

## Purpose
TBD - created by archiving change invite-peer-force-ucg. Update Purpose after archive.
## Requirements
### Requirement: catalog 聚合 totalActivatableCount

cash 合成功能目录时，对 `prediction_unlock` 项 MUST 调用 device **一级根**事件计数并写入响应字段 `totalActivatableCount`。永久合成 `allowedCount` MUST 仍为 defaultFree+permanentDelta（MUST NOT 因 VIP 或邀请改为 -1）。当计数契约失败时，系统 MUST NOT 返回误导性的正数天花板冒充成功（宜省略或 0），并记录日志。客户端预测锁槽位与「已全部激活」判断 MUST 与一级根集合对齐（本仓外 Flutter 同步）。

#### Scenario: catalog 含一级根天花板

- **WHEN** App 拉取 feature catalog 且 device 一级根计数成功返回 N
- **THEN** `prediction_unlock.totalActivatableCount` MUST 等于 N

#### Scenario: 无子一级根计入天花板

- **WHEN** 字典含无子一级根「洗澡」且 device 计数成功
- **THEN** 「洗澡」MUST 反映在 `totalActivatableCount` 中（即计入 N）

### Requirement: device 提供一级根事件计数

device 服务 MUST 提供可被 cash 调用的契约，返回事件字典中**一级根**事件数量（`parent_id = 0` 的事件行数，MUST 包含无子节点的一级根）。计数权威 MUST 为 device 事件字典，MUST NOT 要求 cash 直查 device 库表。该计数 MUST NOT 再等同于「至少有一个子事件的节点数」。

#### Scenario: 统计一级根含无子根

- **WHEN** 事件字典存在 8 个 `parent_id = 0` 的事件，其中部分无子节点
- **THEN** 返回的 count MUST 为 8

#### Scenario: 子事件不计入

- **WHEN** 某一级根下存在若干子事件
- **THEN** 这些子事件 MUST NOT 增加 count；仅该一级根本身计 1

---

## chat-stream-intent-land

<!-- source: openspec/specs/chat-stream-intent-land/spec.md -->

# chat-stream-intent-land Specification

## Purpose
TBD - created by archiving change fix-chat-stream-intent-land. Update Purpose after archive.
## Requirements
### Requirement: 流式意图结果直接落地且禁止二次非流式意图调用

VoiceService 的 `HandleTranscriptForIntentStream` 在 Python 流式意图分析成功结束后，MUST 使用本次 Stream 解析出的意图结果执行与 `chatWithResult` 相同的 confirm / 回复透传逻辑；喂养 history 写库 MUST 由 Python 在意图分析阶段经 history `event/batch` 完成，Go MUST NOT 在主路径二次写库。MUST NOT 再调用非流式 `AnalyzeIntent` 或 `callDeepSeekUnifiedIntent`。

#### Scenario: Stream 成功后透传回复

- **WHEN** 调用方调用 `HandleTranscriptForIntentStream` 且 `AnalyzeIntentStream` 成功返回可解析的意图结果
- **THEN** voice SHALL 将该结果映射为统一意图结构（与非流式 `mapPythonRespToIntent` 等价）
- **AND** SHALL 经 `applyUnifiedIntentResult` 处理 NeedConfirm / exit / 自然语言回复
- **AND** MUST NOT 对该次 transcript 再次调用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`
- **AND** MUST NOT 调用已删除的 Go 侧 `handleUnifiedIntentAction` 写库

#### Scenario: Stream 失败不得回退非流式意图分析

- **WHEN** `AnalyzeIntentStream` 失败或结果无法解析（如 Result 为空）
- **THEN** voice SHALL 返回错误或降级话术
- **AND** MUST NOT 回退调用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent` 以补救落库

### Requirement: 保留 NeedConfirm 与 pending confirm 行为矩阵
流式落地路径 MUST 保留与 `confirm-ws-adaptation` / 现网 `chatWithResult` 一致的确认语义：`NeedConfirm` 置 pending、下一轮 confirm/reject 走 `ConfirmIntent`，且 pending 轮次不得无谓发起新的意图 Stream。

#### Scenario: Stream 返回 NeedConfirm
- **WHEN** 流式意图结果 `NeedConfirm=true`
- **THEN** voice SHALL 保存 pending confirm（含 `conversation_id` 等）
- **AND** SHALL 向调用方返回确认话术（优先 `ConfirmMessage`，否则 `Reply`）
- **AND** SHALL 不执行喂养事件落库

#### Scenario: 存在 pending confirm 时优先处理用户反馈
- **WHEN** 设备已有 pending confirm，且用户输入可解析为 confirm 或 reject
- **THEN** voice SHALL 调用 `ConfirmIntent` 恢复图执行并走统一落库/回复逻辑
- **AND** MUST NOT 再对该轮输入调用 `AnalyzeIntentStream` 或非流式 `AnalyzeIntent`

#### Scenario: pending 反馈无法识别则清理后进入常规流式意图
- **WHEN** 设备已有 pending confirm，但用户输入无法解析为 confirm/reject
- **THEN** voice SHALL 清理 pending
- **AND** SHALL 进入常规流式意图分析（仅一次 Stream）后落地

### Requirement: 保留 quota 单次 consume 与 multi-event / handleUnifiedIntentAction

流式落地路径 MUST 与非流式核心路径共享 quota 语义：常规意图分析（非澄清续聊）成功取得可落地意图后 MUST 执行一次 `consumeVoiceAIQuotaOnSuccess`；multi-event 与普通喂养动作的 history 写库 MUST 均由 Python batch 在意图阶段完成，Go 仅透传 `content` 回复。

#### Scenario: quota 单次 consume

- **WHEN** 常规流式意图分析（非 pending confirm 恢复）成功取得可落地意图
- **THEN** voice SHALL 在成功后执行一次 `consumeVoiceAIQuotaOnSuccess`
- **AND** MUST NOT 因同轮重复意图调用导致双倍 consume

#### Scenario: multi-event 由 Python batch 落库

- **WHEN** 流式或非流式意图结果含多条 `events[]` 且 `need_confirm=false`
- **THEN** Python MUST 经 batch 按各子项 op/action 写库
- **AND** Go voice MUST 透传 Python 返回的回复文本，MUST NOT 在 Go 侧逐条 `AddHistory`

### Requirement: TTS 非流式路径不受影响
现有 TTS / 非流式对话入口 MUST 继续可走非流式意图分析，行为与本变更前对调用方可见效果一致（落库语义与流式落地共享 apply 逻辑时除外，对外仍一次意图调用）。

#### Scenario: HandleTranscriptForStreaming 仍可非流式意图分析
- **WHEN** 调用现有 `HandleTranscriptForStreaming`（或其它走 `chatWithResult` 常规意图路径的入口）
- **THEN** voice MAY 继续使用非流式 `AnalyzeIntent` / `callDeepSeekUnifiedIntent`
- **AND** 落库/confirm 行为矩阵 SHALL 与流式落地路径一致（共享公共 apply 逻辑）

### Requirement: 范围与服务边界

本能力约束 Go `internal/services/voice` 内流式意图落地编排；history 写库守卫在 `history-service`；voice MUST NOT 直连他域 DAO。

#### Scenario: 变更边界

- **WHEN** 实现本变更
- **THEN** voice 删码 SHALL 限于 dead 写库/pending 路径
- **AND** 设备/历史读模型访问 SHALL 继续经既有契约（如 `DeviceAdmin()`、`DeviceHistory().ListHistory`）

---

## chat-stream-reply-only

<!-- source: openspec/specs/chat-stream-reply-only/spec.md -->

# chat-stream-reply-only Specification

## Purpose
TBD - created by archiving change chat-stream-reply-only. Update Purpose after archive.
## Requirements
### Requirement: Chat stream pushes thinking then business reply only

文本对话流式路径（voice internal chat/stream 及经 history 委派的对外 chat/stream）MUST 在思考阶段仅推送 `event: thinking`；MUST 在意图落地（或 preamble 短路）后推送业务话术为 `event: answer`。MUST NOT 将 Python 意图分析 JSON 增量作为 `event: answer` 推送给客户端。

#### Scenario: Successful stream land with thinking

- **WHEN** 调用方发起流式文本对话且 Python `AnalyzeIntentStream` 产出 thinking 增量并成功落地得到业务 Reply
- **THEN** 服务端 SHALL 将 thinking 增量以 SSE `event: thinking` 推送
- **AND** SHALL 在落地完成后以 SSE `event: answer` 推送业务 Reply
- **AND** MUST NOT 推送意图结果 JSON 作为 answer

#### Scenario: Preamble short-circuit without Python stream

- **WHEN** preamble 短路（如 pending child）直接得到业务 Reply 且未调用 AnalyzeIntentStream
- **THEN** 服务端 MAY 不推送 thinking
- **AND** SHALL 以 SSE `event: answer` 推送业务 Reply

### Requirement: Intent stream callback has no OnAnswer

`contracts.IntentStreamCallback` MUST NOT 提供用于外泄意图 JSON 的 `OnAnswer` 回调；流式过程对外仅允许 `OnThinking`。Python 客户端内部累积意图 JSON 供解析 MUST 仍可用，但 MUST NOT 再转发到 `IntentStreamCallback`。

#### Scenario: Intent JSON not forwarded via callback

- **WHEN** `HandleTranscriptForIntentStream` 处理 AnalyzeIntentStream 的 answer（意图 JSON）事件
- **THEN** 系统 SHALL 在内部累积并解析意图结果
- **AND** MUST NOT 通过 `IntentStreamCallback` 将 JSON 增量回调给调用方

### Requirement: HandleTranscriptForIntentStream returns chat result

`HandleTranscriptForIntentStream` 成功或可降级返回时，返回结构 MUST 对齐聊天结果语义，至少包含 Ask、Reply、Exit、FinishTalk；MUST NOT 再要求调用方依赖 Thinking 或 AnswerJSON 字段消费 UI 内容（UI 内容由 thinking 回调与 Reply 承担）。

#### Scenario: Controller writes reply from return value

- **WHEN** voice internal ChatStream 调用 `HandleTranscriptForIntentStream` 结束且 Reply 非空
- **THEN** 控制器 SHALL 使用返回值中的 Reply 写入 SSE `event: answer`

#### Scenario: Stream failure surfaces error event

- **WHEN** 流式意图分析失败或业务错误返回
- **THEN** 控制器 SHALL 写入 SSE `event: error`
- **AND** 若返回结构仍含降级 Reply，MAY 同时写入 `event: answer`

---

## chat-streaming

<!-- source: openspec/specs/chat-streaming/spec.md -->

# chat-streaming Specification

## Purpose
TBD - created by archiving change tip-chat-streaming. Update Purpose after archive.
## Requirements
### Requirement: Chat streaming API
Go 服务端 SHALL 暴露 `POST /device/history/api/chat/stream` 接口，以 SSE 方式流式返回文本对话的思考过程与最终回复。

#### Scenario: Flutter client requests chat via SSE
- **WHEN** Flutter 客户端向 `/device/history/api/chat/stream` 发送 POST 请求（body 包含 deviceNo、transcript）
- **THEN** Go 服务端 SHALL 立即返回 SSE 响应头（Content-Type: text/event-stream、Cache-Control: no-cache）
- **AND** 按序推送 `thinking`、`answer` 两类 SSE 事件
- **AND** 最终以 `data: [DONE]` 结束流式响应

### Requirement: Chat thinking events streamed
Go 服务端 SHALL 将 AI 思考过程增量实时推送给客户端。

#### Scenario: Chat thinking delta pushed to client
- **WHEN** Go 服务端从 voice 服务流式意图分析收到 thinking delta 回调
- **THEN** Go 服务端 SHALL 立即将该 thinking delta 以 SSE `event: thinking` 帧转发给客户端
- **AND** 不得缓冲或延迟超过 100ms

### Requirement: Chat answer events streamed
Go 服务端 SHALL 将 AI 最终回复内容增量实时推送给客户端。

#### Scenario: Chat answer delta pushed to client
- **WHEN** Go 服务端从 voice 服务流式意图分析收到 answer delta 回调
- **THEN** Go 服务端 SHALL 立即将该 answer delta 以 SSE `event: answer` 帧转发给客户端

### Requirement: Chat streaming via history delegate
history 服务 SHALL 以 HTTP SSE 客户端方式委派流式对话请求到 voice 服务。

#### Scenario: History service delegates streaming chat to voice
- **WHEN** history 服务的 `DelegateTextChatStream()` 被调用
- **THEN** history 服务 SHALL 向 voice 服务的 internal 接口 `/voice/internal/api/text/chat/stream` 发起 SSE 请求
- **AND** 将 voice 服务返回的 SSE 事件逐帧透传给调用方

### Requirement: Voice internal streaming chat API
voice 服务 SHALL 暴露 internal SSE 接口 `/voice/internal/api/text/chat/stream`，供跨进程委派使用。

#### Scenario: Voice internal chat streaming endpoint
- **WHEN** 经内部认证的请求到达 `/voice/internal/api/text/chat/stream`
- **THEN** voice 服务 SHALL 调用 `HandleTranscriptForIntentStream()` 执行流式意图分析
- **AND** 将 OnThinking/OnAnswer 回调以 SSE 帧写入 HTTP 响应
- **AND** 流式结束后推送 `data: [DONE]`

### Requirement: Chat stream preserves sync API compatibility
原同步接口 `/device/history/api/chat` SHALL 保持完全不变，继续提供一次性返回 reply 的能力。

#### Scenario: Sync chat API still works
- **WHEN** 客户端调用原同步接口 `/device/history/api/chat`
- **THEN** Go 服务端 SHALL 以与变更前完全一致的同步方式返回 `{reply: "..."}`

### Requirement: Chat API error handling
当 AI 服务不可用时，Go 服务端 SHALL 返回固定错误提示。

#### Scenario: AI unavailable during streaming chat
- **WHEN** 流式对话过程中 AI 服务调用失败
- **THEN** Go 服务端 SHALL 以 SSE `event: error` 帧推送错误消息："AI服务暂时不可用，请稍后再试"
- **AND** 紧接着推送 `data: [DONE]` 结束连接

### Requirement: Flutter sendCommand uses streaming
Flutter 端 `sendCommand()` SHALL 改为 SSE 流式调用 `/chat/stream` 接口。

#### Scenario: Flutter streaming chat display
- **WHEN** Flutter 调用 `sendCommand(text)`
- **THEN** Flutter SHALL 建立 SSE 连接到 `/device/history/api/chat/stream`
- **AND** 收到 thinking 事件时更新 `_chatThinking` 状态并在消息条展示
- **AND** 收到 answer 事件时累积 `_chatReply` 并切换展示最终回复
- **AND** 收到 `data: [DONE]` 时关闭 SSE 连接

---

## chinese-documentation-governance

<!-- source: openspec/specs/chinese-documentation-governance/spec.md -->

# chinese-documentation-governance Specification

## Purpose
TBD - created by archiving change enforce-chinese-documentation. Update Purpose after archive.
## Requirements
### Requirement: OpenSpec 工件默认使用中文
系统在创建或更新 OpenSpec 变更工件时，说明性文本 SHALL 使用中文撰写，包括 proposal、design、specs、tasks。

#### Scenario: 创建新变更工件
- **WHEN** 用户通过 OpenSpec 工作流生成新的 proposal/design/specs/tasks
- **THEN** 生成内容中的说明性文本 SHALL 为中文

### Requirement: 必要技术标识允许保留英文
系统在文档中文化过程中 SHALL 允许保留必要英文技术标识，包括环境变量、路径、接口、协议和代码符号。

#### Scenario: 文档包含技术标识
- **WHEN** 文档中出现环境变量名、API 路径或代码符号
- **THEN** 这些标识 SHALL 保持英文原文，不做强制翻译

---

## ci-acr-github-secrets

<!-- source: openspec/specs/ci-acr-github-secrets/spec.md -->

# ci-acr-github-secrets Specification

## Purpose
TBD - created by archiving change ci-acr-github-secrets. Update Purpose after archive.
## Requirements
### Requirement: CI ACR 凭证来自 GitHub Secrets 而非仓库 .env 文件

`docker-acr` workflow MUST NOT 依赖 git 仓库内存在的 `manifest/docker/.env.test` 或 `manifest/docker/.env.prod` 来获取 ACR 登录信息。构建 push 所需的 `REGISTRY`、`ACR_USERNAME`、`ACR_PASSWORD` MUST 从 GitHub Actions Secrets 加载。仓库 `.gitignore` SHALL 继续忽略 `**/.env.*`（`.env.example` 除外），且 MUST NOT 要求将含真实 ACR 密码的 `.env.test|prod` 提交至 git。

#### Scenario: tag push 触发测试环境构建

- **WHEN** 开发者 push git tag `v2.0.0-beta.1` 且 GitHub Environment `test` 已配置 `REGISTRY`、仓库 Secrets 已配置 `ACR_USERNAME` 与 `ACR_PASSWORD`
- **THEN** workflow SHALL 选择 `test` 环境、从 Secrets 读取 ACR 配置、成功 login 并向对应命名空间 push 七微服务镜像（含 tag 与 git sha）

#### Scenario: 缺少 Secrets 时明确失败

- **WHEN** workflow 运行且目标环境的 `REGISTRY` 或 `ACR_USERNAME` 或 `ACR_PASSWORD` 未配置
- **THEN** workflow SHALL 在 ACR login 之前失败，并输出可操作的错误信息（指明缺失项与 runbook 配置章节）

#### Scenario: checkout 无 .env 文件仍可构建

- **WHEN** 仓库 checkout 后不存在 `manifest/docker/.env.test` 与 `.env.prod`
- **THEN** `docker-acr` workflow SHALL 仍能完成镜像 build 与 push（不因「缺少环境文件」失败）

### Requirement: CI push 地址从 REGISTRY 推导公网域名

workflow SHALL 从 Secrets 中的 `REGISTRY`（允许含 `-vpc` 专线域名）推导 CI push 用的公网 registry 地址（去掉 host 中的 `-vpc` 段）。若推导结果仍含 `-vpc` 或缺少命名空间路径，workflow MUST 失败并给出格式说明。

#### Scenario: REGISTRY 含 vpc 域名

- **WHEN** Environment `test` 的 `REGISTRY` 为 `crpi-xxx-vpc.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test`
- **THEN** workflow SHALL 使用 `crpi-xxx.cn-hangzhou.personal.cr.aliyuncs.com/pangbao-test` 作为 push 地址

#### Scenario: REGISTRY 格式非法

- **WHEN** `REGISTRY` 不含 `/` 命名空间段
- **THEN** workflow MUST 失败并提示格式须为 `<域名>/<命名空间>`

### Requirement: tag 路由规则保持不变

workflow SHALL 保留现有环境选择规则：`workflow_dispatch` 使用输入 `target_env`；**tag push 时以 base tag（`+` 之前）判定**：`vMAJOR.MINOR.PATCH`（无预发布后缀）→ `prod`；其余 `v*` 预发布 base tag → `test`。git tag 含 `+` 构建后缀 MUST NOT 改变上述路由逻辑。

#### Scenario: 正式 semver tag 路由生产

- **WHEN** push tag `v2.0.3`
- **THEN** workflow SHALL 使用 GitHub Environment `prod` 的 Secrets

#### Scenario: 预发布 tag 路由测试

- **WHEN** push tag `v2.0.3-beta.2`
- **THEN** workflow SHALL 使用 GitHub Environment `test` 的 Secrets

#### Scenario: 带构建后缀的预发布 tag 仍路由测试

- **WHEN** push tag `v2.0.3-beta.2+ucg`
- **THEN** workflow SHALL 使用 GitHub Environment `test` 的 Secrets

### Requirement: Runbook 区分 CI Secrets 与 ECS .env

`docs/runbooks/release-deploy-and-run.md` SHALL 文档化：GitHub Actions 使用 Environments/Secrets 配置 ACR；ECS 部署使用本地 `manifest/docker/.env.test|prod`（从 `.env.example` 复制，不上传 git）。文档 MUST 列出所需 Secret 名称与各 Environment 的 `REGISTRY` 配置方式，并 MUST NOT 再声明「无需 GitHub Secrets」。

#### Scenario: 运维按 runbook 配置 CI

- **WHEN** 运维阅读 runbook「ACR 与 CI 凭证」章节
- **THEN** 文档 SHALL 提供 `ACR_USERNAME`、`ACR_PASSWORD`、`REGISTRY`（test/prod 分环境）的配置步骤与验证方式（如 workflow_dispatch）

#### Scenario: 运维按 runbook 配置 ECS

- **WHEN** 运维在 ECS 部署测试栈
- **THEN** runbook SHALL 说明 `.env.test` 仍须含 `REGISTRY`、`IMAGE_TAG`、`ACR_*` 等完整字段，且该文件仅存在于服务器、不进 git

### Requirement: ACR 密码不得出现在 workflow 日志

workflow MUST 对 `ACR_PASSWORD` 使用 GitHub `::add-mask::`（或等价机制），防止明文密码写入 Actions 日志。

#### Scenario: 成功 login 后日志无密码

- **WHEN** workflow 完成 ACR login
- **THEN** Actions 日志 SHALL NOT 包含 `ACR_PASSWORD` 明文

### Requirement: docker-acr SHALL support tag build-scope suffix for selective service builds

`docker-acr` workflow SHALL 支持 git tag **`+` 构建范围后缀**：`+` **之前** 的片段为 **base tag**（用作 ACR 主 tag、`primary_tag`、与服务器 `IMAGE_TAG` 一致）；`+` **之后** 为逗号分隔的服务别名列表，workflow **仅** build 并 push 列表对应的服务镜像。无 `+` 后缀时 SHALL 构建全部 6 个微服务（与变更前行为一致）。

workflow MUST NOT 对未纳入构建范围的服务执行 retag 或 manifest 复制。未构建的服务在 ACR 上 **MAY** 不存在 `:${base_tag}` 镜像；该状态 SHALL 视为预期，而非 workflow 失败条件。

服务别名 MUST 至少支持：`gateway`、`gateway-app`、`history`/`history-service`、`voice`/`voice-service`、`device`/`device-service`、`ucg`/`ucg-service`；`all` MUST 表示全量 6 服务。非法别名 MUST 导致 workflow 在 build 前失败并输出可操作的错误信息。

test/prod 环境路由 MUST 基于 **base tag**（去掉 `+` 后缀后）应用现有规则，MUST NOT 因 `+ucg` 等后缀改变命名空间选择。

#### Scenario: 预发布 tag 仅构建 ucg

- **WHEN** 开发者 push git tag `v1.0.0-rc.4+ucg` 且 Secrets 配置正确
- **THEN** workflow SHALL 仅 build/push `ucg-service` 至 `:${v1.0.0-rc.4}` 与 `:${git_sha}`，且 SHALL NOT push 其他五服务镜像至 `v1.0.0-rc.4`

#### Scenario: 无后缀全量构建保持不变

- **WHEN** 开发者 push git tag `v1.0.0-rc.4`（无 `+`）
- **THEN** workflow SHALL build/push 全部 6 个微服务至 `:${v1.0.0-rc.4}` 与 `:${git_sha}`

#### Scenario: 非法服务别名失败

- **WHEN** push tag `v1.0.0-rc.4+unknown-svc`
- **THEN** workflow MUST 失败且 MUST NOT push 任意镜像

#### Scenario: base tag 环境路由不受后缀影响

- **WHEN** push tag `v2.0.0-rc.1+ucg`
- **THEN** workflow SHALL 使用 GitHub Environment `test`（因 base tag 为预发布），且 push 主 tag MUST 为 `v2.0.0-rc.1`（不含 `+ucg`）

### Requirement: workflow_dispatch SHALL accept optional services scope

手动触发 `docker-acr` 时，workflow SHALL 支持可选输入 `services`（逗号分隔别名，空=全量 6 服务），语义 MUST 与 tag `+` 后缀一致。`image_tag` 输入 MUST 为 base tag（不含 `+` 后缀）。

#### Scenario: 手动仅构建 ucg

- **WHEN** 运维 workflow_dispatch 选择 `target_env=test`、`image_tag=v1.0.0-rc.4`、`services=ucg`
- **THEN** workflow SHALL 仅 build/push `ucg-service` 至 `:v1.0.0-rc.4`

---

## client-feature-usage-stats

<!-- source: openspec/specs/client-feature-usage-stats/spec.md -->

# client-feature-usage-stats Specification

## Purpose
TBD - created by archiving change add-client-feature-usage-stats. Update Purpose after archive.
## Requirements
### Requirement: App client-usage report MUST accept logged-in feature events

`gateway-app-server` MUST 提供 `POST /device/app/api/client-usage/report`（或 design 锁定的等价本机 App 路径）。请求体 MUST 包含非空字符串 `featureId`（客户端自由字符串；服务端 MUST NOT 强制枚举白名单）。系统 MUST 从登录态解析 `wxId`；当 `wxId <= 0` 时 MUST 拒绝写入且 MUST NOT 记入 Redis。当 `featureId` 为空或超过实现设定的最大长度时 MUST 拒绝。成功时 MUST 使用**服务端时间**记录事件，MUST NOT 信任客户端时间戳。

#### Scenario: 登录用户上报成功

- **WHEN** 已登录用户提交合法 `featureId` 且未触发限流且非模拟用户
- **THEN** 系统 SHALL 返回成功，并 SHALL 更新聚合与该用户时间线

#### Scenario: 未登录拒绝

- **WHEN** 请求无法解析出正整数 `wxId`
- **THEN** 系统 MUST 拒绝，且 MUST NOT 写入任何 client-usage Redis 键

### Requirement: Simulated users MUST NOT be recorded

当关联 `wxId` 被判定为模拟用户（与 App API usage 同源判定）时，client-usage report MUST NOT 写入聚合或时间线。

#### Scenario: 模拟用户上报不落库

- **WHEN** 模拟用户调用 report 且请求其余字段合法
- **THEN** 系统 SHALL NOT 增加聚合计数，SHALL NOT 追加时间线（响应可为成功空结果以免客户端重试）

### Requirement: Per-user rate limit of one successful report per 3 seconds

同一 `wxId` 在任意 3 秒窗口内 MUST 最多完成一次成功的 client-usage 写入。超出限制时 MUST 拒绝本次写入（例如 HTTP 429 或明确业务错误），MUST NOT 更新聚合或时间线。

#### Scenario: 3 秒内第二次上报被拒

- **WHEN** 同一用户在成功上报后 3 秒内再次调用 report
- **THEN** 系统 MUST 拒绝第二次写入，且 Redis 聚合/时间线 MUST NOT 因第二次请求增加

### Requirement: Redis aggregation window MUST be at most 30 days

系统 MUST 以日桶等形式维护按 `featureId`、按用户、以及功能×用户交叉的聚合计数，查询窗口 MUST 支持至多 **30 天**（例如 `days=7` 与 `days=30`）。聚合数据 MUST 仅存 Redis，MUST 经 `cachekit` 访问，键 MUST 由 platform builder 登记。产品接受 Redis 丢失导致统计丢失。

#### Scenario: 按功能查询近 7 天

- **WHEN** 管理员请求功能维度列表且 `days=7`
- **THEN** 系统 SHALL 返回窗口内各 `featureId` 的合计次数（及实现约定的最近时间字段，若有）

### Requirement: Per-user timeline MUST store event name and server time only

每个真实用户 MUST 可查询上报时间线。每条时间线条目 MUST 仅包含：`featureId`（事件名）与服务端 Unix 时间。MUST NOT 要求或持久化客户端自定义 `extra` 载荷作为一期范围。

#### Scenario: 用户时间线字段

- **WHEN** 管理员查询某 `wxId` 的时间线
- **THEN** 每条记录 SHALL 含 `featureId` 与服务端时间，且 SHALL NOT 依赖客户端提交的时间戳

### Requirement: Per-user timeline MUST cap at 10000 entries and retain about 30 days

单用户时间线 MUST 最多保留 **10000** 条；超出时 MUST 丢弃最旧条目。时间线保留策略 MUST 与约 **30 天** 有效期对齐（key TTL 与/或按日过期）。条数上限与时间窗口 MUST 同时生效。

#### Scenario: 超过一万条裁剪

- **WHEN** 某用户时间线已有 10000 条且再次成功上报
- **THEN** 系统 SHALL 追加新事件并 SHALL 移除最旧事件，使条数不超过 10000

### Requirement: Report API MUST NOT count toward App API usage stats

`POST /device/app/api/client-usage/report`（精确 METHOD+path）MUST 登记于 `maintenance_skip`（或等价 denylist），MUST NOT 计入 App API 使用统计。负责人已确认本接口不统计。

#### Scenario: report 不计入 API usage

- **WHEN** 用户成功调用 client-usage report
- **THEN** App API 使用统计对应计数 MUST NOT 因该请求递增

### Requirement: Admin client-usage APIs and Hub page

`gateway-app` MUST 提供需 Admin JWT 的读接口，至少支持：按功能聚合列表、按用户列表/下钻、单用户时间线。路径 MUST 由 gateway-app 本机处理，MUST NOT 直连他域数据库。Hub MUST 增加独立入口「客户端使用统计」，交互对齐现有 App API 使用统计页（功能/用户维度；时间范围不超过 30 天）。页面 MUST 提示数据仅 Redis、可丢失，且时间线可能因 10000 条上限短于次数合计。

#### Scenario: Hub 按用户查看时间线

- **WHEN** 管理员打开「客户端使用统计」并进入某用户详情
- **THEN** 系统 SHALL 展示该用户聚合摘要与时间线（事件名 + 服务端时间）

---

## clinic-tip-feedback-flutter

<!-- source: openspec/specs/clinic-tip-feedback-flutter/spec.md -->

# clinic-tip-feedback-flutter Specification

## Purpose
TBD - created by archiving change close-clinic-tip-feedback. Update Purpose after archive.
## Requirements
### Requirement: tip 反馈字段与 URL 对齐 Go API
Flutter tip 客户端 SHALL 以 `POST /device/api/tip/feedback` 提交反馈，JSON Body 使用 `answerId`（string）与 `feedback`（int：1 或 -1），MUST NOT 继续使用 `tipId` / `feedbackResult`。

#### Scenario: tip submitFeedback 对齐
- **WHEN** 用户在首页小贴士面板点击 thumbs up/down 且本地已有非空 `answerId`
- **THEN** 客户端 SHALL 向 `/device/api/tip/feedback` 发送 `{ "answerId": "<id>", "feedback": 1|-1 }`
- **AND** 已反馈后 MUST NOT 允许再次更改

#### Scenario: 无 answerId 不提交
- **WHEN** 本地 `answerId` 为空或缺失
- **THEN** 客户端 MUST NOT 发起 tip feedback HTTP 请求

### Requirement: tip 流式 done 写入 answerId
Flutter tip 流式接收完成后 SHALL 将服务端返回的 `answerId` 写入 tip 状态，供反馈提交使用。该能力依赖包 B tip SSE；若 B 未完成，本变更 SHALL 补最小接线或在实现报告中标明 tip 反馈 blocker。

#### Scenario: done 携带 answerId
- **WHEN** tip SSE 流结束且 done 帧含 `answerId`
- **THEN** tip 状态 SHALL 保存该字符串，反馈按钮可用

#### Scenario: 包 B 阻塞时的降级
- **WHEN** 包 B 未完成导致 tip 无法获得 `answerId`
- **THEN** 实现报告 SHALL 明确 tip 反馈 blocker
- **AND** clinic 反馈与 Go Bind/反代/skip MUST 仍可完成

### Requirement: clinic 反馈 UI 与 HTTP
胖宝诊疗屏（`pangbao_ai_screen`）SHALL 在助手回答完成且存在 `answerId` 时展示 thumbs up/down，并经 HTTP `POST /device/api/clinic/feedback` 提交 `{answerId, feedback}`。

#### Scenario: clinic 回答完成后可反馈
- **WHEN** 某条助手消息已 `answer_done` 且 `answerId` 非空，且尚未反馈
- **THEN** UI SHALL 展示反馈按钮
- **AND** 用户点击后 SHALL POST `/device/api/clinic/feedback` 并标记该条已反馈

#### Scenario: 无 answerId 或已反馈
- **WHEN** `answerId` 为空，或该条 `feedbackGiven` 已为 true
- **THEN** UI MUST NOT 再次提交 clinic feedback

---

## clinic-tip-feedback-voice-host

<!-- source: openspec/specs/clinic-tip-feedback-voice-host/spec.md -->

# clinic-tip-feedback-voice-host Specification

## Purpose
TBD - created by archiving change close-clinic-tip-feedback. Update Purpose after archive.
## Requirements
### Requirement: voice-service 注册 clinic/tip 反馈控制器
voice-service SHALL 在 HTTP 注册中 Bind `DeviceClinicFeedbackController`（或等价控制器），对外暴露已有 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`。

#### Scenario: clinic feedback 可达
- **WHEN** 已鉴权客户端向 voice-service（经 gateway）发送 `POST /device/api/clinic/feedback`，Body 含合法 `answerId` 与 `feedback`（1 或 -1）
- **THEN** 控制器 SHALL 调用 `PythonAIClient.ClinicFeedback` 并将 Python 结果映射为 `{code, message, data}` 响应

#### Scenario: tip feedback 可达
- **WHEN** 已鉴权客户端向 voice-service（经 gateway）发送 `POST /device/api/tip/feedback`，Body 含合法 `answerId` 与 `feedback`（1 或 -1）
- **THEN** 控制器 SHALL 调用 `PythonAIClient.TipFeedback` 并将 Python 结果映射为 `{code, message, data}` 响应

#### Scenario: 非法 feedback 值
- **WHEN** 请求 Body 中 `feedback` 不是 1 或 -1
- **THEN** 控制器 SHALL 返回业务错误（如 code=400）且 MUST NOT 调用 Python

### Requirement: gateway 反代 clinic/tip API 到 VOICE
gateway-app SHALL 将 `/device/api/clinic/*` 与 `/device/api/tip/*` 反代至 `VOICE_API_PROXY`（voice 域路由模式），MUST NOT 反代至 device-service。

#### Scenario: clinic/tip API 走 voice 上游
- **WHEN** gateway-app 收到匹配 `/device/api/clinic/*` 或 `/device/api/tip/*` 的请求且 voice 代理已启用
- **THEN** 请求 SHALL 经 voice 反代中间件转发到 voice-service
- **AND** MUST NOT 进入 device 域 `/device/app/api/feedback/*` 反代路径

#### Scenario: 与 device 用户反馈路径隔离
- **WHEN** 客户端请求 `/device/app/api/feedback/*`
- **THEN** 行为 SHALL 保持既有 device 域反代，不受本变更影响

### Requirement: feedback 接口不计入 usage 统计
负责人已确认：`POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback` MUST NOT 计入 gateway-app App API 使用统计。实现 SHALL 在 `maintenance_skip.go` 以精确 `METHOD + path` 排除二者。

#### Scenario: clinic feedback 被 skip
- **WHEN** gateway-app 处理 `POST /device/api/clinic/feedback`
- **THEN** usagestats SHALL 将其视为维护型 API 并跳过计数

#### Scenario: tip feedback 被 skip
- **WHEN** gateway-app 处理 `POST /device/api/tip/feedback`
- **THEN** usagestats SHALL 将其视为维护型 API 并跳过计数

#### Scenario: tip generate 不在本包改统计策略
- **WHEN** 评估 `POST /device/tip/generate` 的 usage 策略
- **THEN** 本变更 MUST NOT 将其写入 `maintenance_skip.go`（统计属包 B）

### Requirement: 反馈路径与鉴权保持既有契约
反馈 HTTP 路径 MUST 保持为 `/device/api/clinic/feedback` 与 `/device/api/tip/feedback`；经 gateway-app 暴露时 MUST 要求登录（不加入 Bearer exempt）。

#### Scenario: 路径不变
- **WHEN** 查阅 `api/v1` 的 `g.Meta` 或对外文档
- **THEN** clinic/tip 反馈 path 仍为上述两路径，无破坏性改名

#### Scenario: 未登录被拒
- **WHEN** 无有效 Bearer 的客户端请求上述反馈 POST
- **THEN** gateway-app SHALL 按既有鉴权规则拒绝（401），不得因本变更加入白名单

---

## company-site-legal

<!-- source: openspec/specs/company-site-legal/spec.md -->

# company-site-legal Specification

## Purpose
TBD - created by archiving change company-branding-site-legal. Update Purpose after archive.
## Requirements
### Requirement: 官网页脚 MUST 展示公司主体与备案信息

`resource/public/pangbao-home.html` 页脚 MUST 展示运营主体中文全称「杭州彩逗科技有限公司」、英文名「Hangzhou Caidou Technology Co., Ltd.」、ICP 备案号「浙ICP备2024101366号」（可点击链至工信部备案查询站）、客服邮箱 `tou_zy@foxmail.com`，以及隐私政策与用户协议链接。页脚 MUST NOT 再展示「官网事件 logo 与 Android 下载链路均来自系统权威数据」类运维旁白。

#### Scenario: 用户查看官网底部

- **WHEN** 用户打开官网首页并滚动至页脚
- **THEN** 页面 SHALL 可见公司中英文全称、浙ICP备2024101366号与客服邮箱，且可进入隐私政策与用户协议

### Requirement: 官网 iOS 下载 MUST 提供 App Store 直达链接

官网 iOS 下载区 MUST 提供可点击链接，目标为  
`https://apps.apple.com/cn/app/%E8%83%96%E5%AE%9D/id6774418472`（或等价已编码的 App Store 胖宝应用页）。MAY 保留「亦可在 App Store 搜索胖宝」作为辅助说明，但 MUST NOT 仅依赖「打开 App Store 首页」而无应用直达。

#### Scenario: 用户点击 iOS 下载

- **WHEN** 用户在官网点击 iOS 下载/打开 App Store 主按钮
- **THEN** 浏览器 SHALL 打开胖宝应用的 App Store 详情页（应用 id 6774418472）

### Requirement: 官网 MUST 简要说明产品由公司运营

官网 MUST 在页脚或独立短段中说明「胖宝」由杭州彩逗科技有限公司运营（文案可简短）。产品能力展示 MUST 至少覆盖喂养记录、同月龄社区、AI 陪伴，并 SHOULD 体现预测/提醒或诊疗等现网能力中的至少一项，且 MUST 保留 AI 不能替代医疗诊断的免责提示。

#### Scenario: 用户阅读产品能力

- **WHEN** 用户查看官网产品能力区块
- **THEN** 页面 SHALL 展示多类产品能力要点，并 SHALL 含 AI 免责相关提示

---

## compose-container-resource-limits

<!-- source: openspec/specs/compose-container-resource-limits/spec.md -->

# compose-container-resource-limits Specification

## Purpose
TBD - created by archiving change compose-2g-redis-limits. Update Purpose after archive.
## Requirements
### Requirement: 生产与测试 Compose SHALL 定义容器 CPU 与内存上限

仓库 SHALL 提供 **`manifest/docker/docker-compose.resources.prod.yml`** 与 **`manifest/docker/docker-compose.resources.test.yml`**（或后继等价 overlay），为下列组件定义 **`mem_limit`** 与 **`cpus`**（或 compose 规范中等价、在非 Swarm 模式下对 `docker compose up` 生效的字段）：

- 生产/测试 **全部** 微服务（gateway、gateway-app、history、voice、device、worker、ucg）
- 生产 Redis、测试 Redis、生产/测试 RabbitMQ

runbook SHALL 文档化默认配额表及「2G ECS survival profile」说明。`voice-service` 测试实例 SHALL 拥有 **不低于** 其它微服务的 memory limit（ documented 起步值 **512M**）。

#### Scenario: 启动命令叠加 resources overlay

- **WHEN** 运维按 runbook 启动生产微服务并叠加 `-f docker-compose.resources.prod.yml`
- **THEN** `docker inspect` 或 `docker stats` SHALL 显示对应容器配置了 memory/cpu 上限

#### Scenario: 本地开发不受 prod/test limits 强制约束

- **WHEN** 开发者仅使用基线 `microservices.yml` + `microservices.local.yml` 且 **不** 叠加 `resources.*.yml`
- **THEN** 本地容器 **MAY** 无 cgroup 上限（便于调试）

### Requirement: limits SHALL 防止单容器耗尽宿主机

资源上限的配置意图 SHALL 在 runbook 中说明：当某容器内存超过 `mem_limit` 时，内核 **MAY** OOM kill 该容器，**SHALL NOT** 无限制占用同机其它栈（含 MySQL 宿主机进程）的全部物理内存。runbook SHALL 包含 OOM 排查步骤（`dmesg`、`docker stats`、调高 voice-test limit 等）。

#### Scenario: 文档化 OOM 语义

- **WHEN** 运维查阅 `release-deploy-and-run.md` 资源 limits 章节
- **THEN** 文档 SHALL 说明 limits 与宿主机 2G 物理内存的关系，以及 ASR 验收时优先保障 test voice 的建议

---

## compose-host-root-asset-volumes

<!-- source: openspec/specs/compose-host-root-asset-volumes/spec.md -->

# compose-host-root-asset-volumes Specification

## Purpose
TBD - created by archiving change compose-host-root-asset-volumes. Update Purpose after archive.
## Requirements
### Requirement: device-service 事件 logo 持久化到宿主机根目录

在 Docker Compose 部署下，device-service 容器 SHALL 通过 bind mount 将 **`/ai_talk_images`** 映射到 Linux 宿主机同路径 **`/ai_talk_images`**，使 `SaveEventLogo` 写入的文件出现在宿主机上。

#### Scenario: 上传 logo 后宿主机可见

- **WHEN** 管理员通过 API 上传事件 logo 且 device-service 使用默认 `eventImageStorageDir` `/ai_talk_images/`
- **THEN** 宿主机路径 `/ai_talk_images/` 下 SHALL 存在对应文件
- **AND** 容器内同路径 SHALL 可读取该文件

#### Scenario: 容器重建后文件保留

- **WHEN** 宿主机 `/ai_talk_images` 已存在 logo 文件且运维对 device-service 执行 `docker compose up --force-recreate`
- **THEN** 重建后容器 SHALL 仍能读取宿主机挂载目录中的同名文件

### Requirement: gateway-app APK 持久化到宿主机根目录

在 Docker Compose 部署下，gateway-app 容器 SHALL 通过 bind mount 将 **`/apk/ai_talk`** 映射到 Linux 宿主机同路径 **`/apk/ai_talk`**，使版本管理上传的 APK 出现在宿主机上。

#### Scenario: 上传 APK 后宿主机可见

- **WHEN** 管理员通过版本管理接口上传 APK 且 gateway-app 使用默认 `apkStorageDir` `/apk/ai_talk/`
- **THEN** 宿主机路径 `/apk/ai_talk/` 下 SHALL 存在对应 `.apk` 文件

#### Scenario: 容器重建后 APK 保留

- **WHEN** 宿主机 `/apk/ai_talk` 已存在 APK 且运维对 gateway-app 执行 `docker compose up --force-recreate`
- **THEN** 重建后 gateway-app SHALL 仍能通过 `GET /device/app/apk/` 提供该文件

### Requirement: 挂载路径与配置默认一致

Compose 卷挂载点 SHALL 与代码/配置默认存储目录一致：`/ai_talk_images`（device）、`/apk/ai_talk`（gateway-app）。未通过环境变量修改存储路径时，SHALL NOT 要求额外配置即可满足本需求。

#### Scenario: 默认配置下路径一致

- **WHEN** 未设置 `DEVICE_EVENT_IMAGE_STORAGE_DIR` 与 `GATEWAY_APP_APK_STORAGE_DIR`
- **THEN** 写盘路径与 bind mount 目标路径 SHALL 均为上述宿主机根下目录

### Requirement: 部署文档说明宿主机准备

项目 runbook 或等价部署文档 SHALL 说明：Linux Docker **生产**部署前建议执行 `mkdir -p /ai_talk_images /apk/ai_talk`；**测试**部署前建议执行 `mkdir -p /ai_talk_images_test /apk/ai_talk_test`；并给出验证宿主机与容器内文件一致的示例命令。

#### Scenario: 运维可按文档验收生产静态目录

- **WHEN** 运维按文档创建生产目录并启动 prod compose 后上传 logo 与 APK
- **THEN** 文档中的 `ls` 或 `docker exec` 验收步骤 SHALL 能确认宿主机 `/ai_talk_images` 与 `/apk/ai_talk` 非空

#### Scenario: 运维可按文档验收测试静态目录

- **WHEN** 运维按文档创建测试目录并启动 test compose 后上传 logo 与 APK
- **THEN** 文档中的验收步骤 SHALL 能确认宿主机 `/ai_talk_images_test` 与 `/apk/ai_talk_test` 非空

### Requirement: 测试栈事件 logo 持久化到独立宿主机目录

在 Docker Compose **测试** overlay 部署下，device-service 容器 SHALL 通过 bind mount 将容器内 **`/ai_talk_images`** 映射到 Linux 宿主机 **`/ai_talk_images_test`**，与生产目录 `/ai_talk_images` 隔离。

#### Scenario: 测试上传 logo 不写入生产目录

- **WHEN** 管理员在测试环境上传事件 logo 且 device-service 使用默认 `eventImageStorageDir` `/ai_talk_images/`
- **THEN** 文件 SHALL 出现在宿主机 `/ai_talk_images_test/` 下
- **AND** 宿主机 `/ai_talk_images/`（生产）SHALL NOT 因该上传而新增或修改同名文件

### Requirement: 测试栈 APK 持久化到独立宿主机目录

在 Docker Compose **测试** overlay 部署下，gateway-app 容器 SHALL 通过 bind mount 将容器内 **`/apk/ai_talk`** 映射到 Linux 宿主机 **`/apk/ai_talk_test`**，与生产目录 `/apk/ai_talk` 隔离。

#### Scenario: 测试上传 APK 不写入生产目录

- **WHEN** 管理员在测试环境版本管理页上传 APK
- **THEN** 文件 SHALL 出现在宿主机 `/apk/ai_talk_test/` 下
- **AND** 宿主机 `/apk/ai_talk/`（生产）SHALL NOT 因该上传而新增或修改同名文件

---

## compose-mysql-endpoint-via-env

<!-- source: openspec/specs/compose-mysql-endpoint-via-env/spec.md -->

# compose-mysql-endpoint-via-env Specification

## Purpose
TBD - created by archiving change compose-mysql-host-env-and-docker-host. Update Purpose after archive.
## Requirements
### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入各服务 MySQL 连接串

`manifest/docker/docker-compose.microservices.yml`（或其后继官方等价文件）SHALL 允许在**不修改已提交 YAML 内口令占位**的前提下，通过环境变量为 `history-service`、`device-service`、`voice-service`、`worker`、`ucg-service` 及 `gateway-app`（`APP_DB_LINK`）注入数据库连接：其中 history/device/voice/ucg SHALL 分别支持 `HISTORY_DB_LINK`、`DEVICE_DB_LINK`、`VOICE_DB_LINK`、`UCG_DB_LINK` 覆盖默认库；worker SHALL 支持 `WORKER_OUTBOX_DB_LINK`（由 cmd 写入 `GF_DATABASE_OUTBOX_LINK`）；gateway-app SHALL 支持 `APP_DB_LINK` 覆盖 `database.app`。prod/test 分环境 `.env` 文件 SHALL 分别注入对应库名。

#### Scenario: device 使用注入的 link 启动

- **WHEN** 部署者在启动 Compose 前设置 `DEVICE_DB_LINK` 为合法 MySQL DSN
- **THEN** `device-service` 进程 SHALL 使用该 DSN 作为 `GF_DATABASE_DEFAULT_LINK`，而不依赖仅写在镜像内配置文件中的占位地址

#### Scenario: 测试 worker 使用 test outbox 库

- **WHEN** 测试栈设置 `WORKER_OUTBOX_DB_LINK` 指向 `ai_voice_worker_test`
- **THEN** worker-service SHALL 使用该 DSN 作为 outbox 库连接，SHALL NOT 写入生产 `ai_voice_worker`

### Requirement: 参考 Compose MUST 为访问宿主机 MySQL 提供 host.docker.internal 解析

当 MySQL 监听在 **运行 Docker 的宿主机** 上且业务容器使用 bridge 网络时，参考 Compose 中需访问该 MySQL 的服务 SHALL 配置 `extra_hosts`，使主机名 `host.docker.internal` 解析到宿主机（例如 `host-gateway` 语义），以便连接串中使用 `tcp(host.docker.internal:3306)` 等地址时行为可预期。

#### Scenario: Linux 上 compose up 后容器解析 host.docker.internal

- **WHEN** 在支持 `host-gateway` 的 Docker Engine 上执行 `docker compose up` 使用该参考文件
- **THEN** 业务容器内 SHALL 能将 `host.docker.internal` 解析到宿主机侧地址，从而可与宿主机上监听的 MySQL 建立 TCP 连接（在 DSN 已正确配置且 mysqld 对 Docker 网桥来源放行时）

### Requirement: 仓库 MUST 提供无密钥的 Compose 数据库环境样例

仓库 SHALL 提供一份可复制为本地 `.env` 的示例文件（例如 `manifest/docker/.env.example`），其中 SHALL 用中文或英文注释说明：**MySQL 与 Docker 同机**时推荐将主机设为 `host.docker.internal`；**MySQL 在其它机器**时将主机设为从容器网络可达的 DNS 或 IP（如 RDS、内网 IP），且 SHALL NOT 包含真实生产口令。

#### Scenario: 新成员首次接 Compose 栈

- **WHEN** 开发者复制示例为 `.env` 并按注释填写自己的 MySQL 拓扑
- **THEN** 其 SHALL 能区分同机与异机两种填法，且无需从 git 历史中寻找口令

### Requirement: 参考微服务 Compose MUST 支持通过环境变量注入 Redis 地址

`manifest/docker/docker-compose.microservices.yml`（或其后继官方等价文件）SHALL 允许通过环境变量 **`GF_REDIS_DEFAULT_ADDRESS`** 覆盖 GoFrame `redis.default.address`，作用于 **所有** 依赖 Redis 的微服务（含 gateway、gateway-app、history、voice、device、worker、ucg）。当变量为空或未设置时，SHALL 回退镜像内 yaml 默认地址（cluster 三主种子）。`.env.test.example` SHALL 文档化测试单机地址 `redis-test:6379`；`.env.prod.example` **SHALL NOT** 要求填写该变量（生产使用 yaml 默认 cluster 种子）。

#### Scenario: 测试栈注入单机 Redis 地址

- **WHEN** `.env.test` 设置 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379` 且启动测试微服务栈
- **THEN** 各服务容器环境 SHALL 包含该变量，且 Redis 客户端 SHALL 连接 `redis-test:6379`

#### Scenario: 生产栈不注入时沿用 yaml cluster 种子

- **WHEN** 生产 `.env.prod` 未设置 `GF_REDIS_DEFAULT_ADDRESS`
- **THEN** 微服务 SHALL 使用 config yaml 中的 `redis-node-1:7001,redis-node-2:7002,redis-node-3:7003`

### Requirement: 仓库 MUST 提供分环境的 Compose 数据库环境样例

除现有 `manifest/docker/.env.example` 外，仓库 SHALL 提供 `manifest/docker/.env.prod.example` 与 `manifest/docker/.env.test.example`。prod 示例 SHALL 说明各 `*_DB_LINK` 指向生产库名（无 `_test` 后缀）及 `IMAGE_TAG` 为 semver。test 示例 SHALL 说明各 `*_DB_LINK` 指向 `ai_voice_*_test` 库、`IMAGE_TAG=develop`、`GATEWAY_APP_PUBLIC_BASE_URL=https://test.pangbao.cuplay.top:9702`，且 SHALL NOT 包含真实生产口令。

#### Scenario: 新运维区分 prod 与 test env 文件

- **WHEN** 运维复制 `.env.test.example` 为 `.env.test` 并按注释填写
- **THEN** 其 SHALL 能将 `HISTORY_DB_LINK` 指向 `ai_voice_history_test` 且将 `IMAGE_TAG` 设为 `develop`，且 SHALL NOT 误用生产库 DSN

### Requirement: 测试栈 Compose MUST 支持 Redis 地址环境注入

测试 overlay 或 `.env.test.example` SHALL 文档化并通过 compose `environment` 注入 `GF_REDIS_DEFAULT_ADDRESS`，指向测试 Redis cluster 三主种子（测试网络内服务名与端口，与 prod 物理隔离）。

#### Scenario: 测试 gateway-app 使用 test Redis

- **WHEN** 测试栈 gateway-app 启动且 `GF_REDIS_DEFAULT_ADDRESS` 已按 `.env.test.example` 配置为 test cluster 种子
- **THEN** 版本检查等 Redis 缓存 SHALL 读写 test cluster，SHALL NOT 依赖 prod cluster 的节点地址

---

## compose-mysql-test-seed-desensitization

<!-- source: openspec/specs/compose-mysql-test-seed-desensitization/spec.md -->

# compose-mysql-test-seed-desensitization Specification

## Purpose
TBD - created by archiving change compose-prod-test-dual-stack. Update Purpose after archive.
## Requirements
### Requirement: 测试 MySQL 库 SHALL 与生产库名隔离

测试环境各业务库 SHALL 使用与生产对应、带 `_test` 后缀的库名（至少包含 `ai_voice_history_test`、`ai_voice_device_test`、`ai_voice_voice_test`、`ai_voice_worker_test`、`ai_voice_app_test`、`ai_voice_ucg_test`）。`.env.test` 中各 `*_DB_LINK` SHALL 指向上述测试库，MUST NOT 指向生产库名。

#### Scenario: 测试 device-service 连接测试库

- **WHEN** 测试栈 device-service 启动且 `DEVICE_DB_LINK` 已按 `.env.test.example` 配置
- **THEN** 进程 SHALL 连接 `ai_voice_device_test`（或 documented 等价名），SHALL NOT 连接 `ai_voice_device`

### Requirement: 仓库 SHALL 提供生产到测试的脱敏种子流程

仓库 SHALL 在 runbook 和/或 `hack/mask-seed-data.sh` 中描述可重复流程：从生产 `ai_voice_*` 导出 → 脱敏 → 导入 `ai_voice_*_test`。脱敏 SHALL 至少处理：用户手机号、微信 openid/unionid、refresh token/session、设备标识（替换或前缀化）。导入 SHALL 覆盖测试库既有数据（运维须在 runbook 中警告）。

#### Scenario: 脱敏后测试库无原始手机号

- **WHEN** 运维按文档完成脱敏 import
- **THEN** 测试库 user 相关表中 SHALL NOT 保留与生产 export 完全相同的手机号明文

#### Scenario: 发版前刷新测试种子

- **WHEN** 准备发布新的 release candidate
- **THEN** runbook SHALL 要求（或 recommend 作为 checklist 必项）在测试验收前执行一次脱敏种子刷新

### Requirement: 脱敏种子 SHALL 同步测试静态资源

当生产种子包含 `/ai_talk_images/` 路径引用时，运维 SHALL 将对应 logo 文件同步至宿主机 `/ai_talk_images_test/`（或 test overlay documented 路径），使测试管理页与 App 静态读链路可验收。

#### Scenario: 测试环境 logo 可读

- **WHEN** 测试库 event 行引用 `/ai_talk_images/<file>` 且文件已同步至测试静态目录
- **THEN** 经 test gateway 或 gateway-app 反代的静态请求 SHALL 返回 200

---

## compose-prod-test-dual-stack

<!-- source: openspec/specs/compose-prod-test-dual-stack/spec.md -->

# compose-prod-test-dual-stack Specification

## Purpose
TBD - created by archiving change compose-prod-test-dual-stack. Update Purpose after archive.
## Requirements
### Requirement: 仓库 SHALL 提供生产与测试双栈 Compose overlay

仓库 SHALL 在 `manifest/docker/` 提供 `docker-compose.microservices.prod.yml` 与 `docker-compose.microservices.test.yml`，与基线 `docker-compose.microservices.yml` 组合使用。prod/test overlay SHALL 使用 `${REGISTRY}/<service>:${IMAGE_TAG}` 引用镜像仓库（如 `${REGISTRY}/gateway:${IMAGE_TAG}`，无 `go-ai-talk/` 路径前缀，以适配阿里云 ACR 等单段仓库名），且 SHALL NOT 包含 `build` 段。基线文件 MAY 保留 `build` 与 `:local` 供本机开发。

#### Scenario: 测试栈从 registry pull 启动

- **WHEN** 运维设置 `REGISTRY`、 `IMAGE_TAG=develop` 并执行 `docker compose -f ...microservices.yml -f ...microservices.test.yml pull && up -d --no-build`
- **THEN** 各业务容器 SHALL 使用 registry 中 `:develop` 镜像启动，且 SHALL NOT 在宿主机执行源码 build

#### Scenario: 生产栈使用 semver tag

- **WHEN** 运维在 `.env.prod` 设置 `IMAGE_TAG=v1.0.0` 并 pull + up
- **THEN** 生产容器 SHALL 使用 `:v1.0.0` 镜像，且 SHALL NOT 使用 `:develop` 或 `:local`

### Requirement: 生产与测试 SHALL 使用独立 Docker 网络完全隔离

生产栈与测试栈 SHALL 分别仅加入独立的 external Docker 网络（约定名 `go-ai-talk-prod-net` 与 `go-ai-talk-test-net`）。同一宿主机上 prod 与 test 的中间件与微服务 SHALL NOT 共用同一 bridge 网络的 DNS 解析。

#### Scenario: test 网络内 rabbitmq 不可被 prod 容器解析

- **WHEN** prod 与 test 栈同时运行且各自 RabbitMQ 仅加入对应网络
- **THEN** prod 容器内 SHALL NOT 通过服务名 `rabbitmq` 解析到 test 的 RabbitMQ 实例

### Requirement: 测试栈 SHALL 独立 Redis Cluster 与 RabbitMQ

仓库 SHALL 提供 `docker-compose.redis-cluster.test.yml` 与 `docker-compose.rabbitmq.test.yml`。测试 Redis cluster 宿主机映射端口 SHALL 使用 17001–17006（或与 prod 7001–7006 不冲突的 documented 端口段）。测试 RabbitMQ 宿主机映射 SHALL 使用 5673/15673（或与 prod 5672/15672 不冲突的 documented 端口段）。测试微服务 SHALL 通过环境变量 `GF_REDIS_DEFAULT_ADDRESS` 与 `MQ_HTTP_API_BASE` 指向 test 网络内中间件。

#### Scenario: 测试 history 与 worker 使用 test RabbitMQ

- **WHEN** 测试栈 `history-service` 与 `worker` 已启动且 `OUTBOX_RELAY_ENABLED`/`MQ_CONSUMER_ENABLED` 为 true
- **THEN** 二者 SHALL 仅与 test 网络内 RabbitMQ 通信，且 prod worker SHALL NOT 消费 test 队列中的消息

### Requirement: 测试栈后端端口 SHALL 与生产错开

测试栈微服务宿主机端口映射 SHALL 为：gateway 19701、gateway-app 19702、history 19801、voice 19802、device 19803、ucg 19804、worker 19901（或与 runbook  documented 表一致且不与 prod 9701–9901 冲突）。测试栈 container_name SHALL 与 prod 不同（例如带 `-test` 后缀或使用 `COMPOSE_PROJECT_NAME` 前缀）。

#### Scenario: 同机 prod 与 test 同时监听

- **WHEN** prod 与 test 栈同时 up
- **THEN** 宿主机 SHALL 可同时访问 `127.0.0.1:9701`（prod gateway）与 `127.0.0.1:19701`（test gateway）且无端口绑定冲突

### Requirement: 测试对外访问形态 SHALL 与生产一致

测试环境对外入口 SHALL 为 `https://test.pangbao.cuplay.top:9701`（主网关）与 `https://test.pangbao.cuplay.top:9702`（App 网关），由 Nginx（或等价反代）转发至测试后端 19701/19702。测试栈 SHALL 设置 `GATEWAY_APP_PUBLIC_BASE_URL` 为 `https://test.pangbao.cuplay.top:9702`（或 runbook documented 等价 HTTPS 基址）。

#### Scenario: 客户端仅换域名访问测试 App 网关

- **WHEN** 客户端将生产基址 `www.pangbao.cuplay.top:9702` 换为 `test.pangbao.cuplay.top:9702` 且路径不变（如 `/device/app/api/version/check`）
- **THEN** 请求 SHALL 到达测试 gateway-app 且 API 路径语义与生产一致

### Requirement: 镜像 tag 策略 SHALL 区分测试浮动与生产钉死

测试默认 `IMAGE_TAG=develop`（CI 覆盖的浮动 tag）。生产 MUST 使用 semver release tag（如 `v1.0.0`），MUST NOT 在生产 `.env.prod` 中使用 `develop` 或 `latest`。CI SHOULD 同时 push 不可变 `:<git-sha>` tag 供排错。

#### Scenario: 生产 env 拒绝 develop

- **WHEN** 运维检查生产部署配置
- **THEN** `.env.prod` 中 `IMAGE_TAG` SHALL 为 semver 形式且 SHALL NOT 等于 `develop`

### Requirement: 部分 CI 构建发版 SHALL 支持按服务 pull 与 up

当 CI 使用带 `+` 构建后缀的 tag（如 `v1.0.0-rc.4+ucg`）仅 push 部分服务镜像时，运维 MUST NOT 假设 ACR 上存在全部六服务 `:${base_tag}` 镜像。部署 MUST 通过 **按服务** `docker compose pull <service>` 与 `up -d --no-build <service>` 更新变更服务；对未构建服务 **MAY** 继续运行已拉取的旧 tag 本地镜像，直至下次无后缀全量 tag 发版。

`docs/runbooks/release-deploy-and-run.md` MUST 文档化：git tag 全名与 `.env` 中 `IMAGE_TAG`（base tag）的关系、`+ucg` 等后缀含义、按服务 pull/up 示例命令，以及全栈 `compose pull` 在部分构建 tag 下 **预期失败** 的说明（表示构建范围与部署操作不匹配）。

#### Scenario: 部分构建后仅更新 ucg

- **WHEN** CI 已对 tag `v1.0.0-rc.4+ucg` 仅 push `ucg-service:v1.0.0-rc.4`，且运维将 `.env.test` 中 `IMAGE_TAG` 设为 `v1.0.0-rc.4`
- **THEN** 运维 SHALL 执行 `pull`/`up` 仅针对 `ucg-service`；其他服务容器 MAY 保持上一版本镜像运行

#### Scenario: 全栈 pull 在部分 tag 下失败为预期

- **WHEN** ACR 不存在 `gateway:v1.0.0-rc.4`（因本次 CI 为 `+ucg` 部分构建）且运维执行全栈 `docker compose pull`
- **THEN** pull MUST 对缺失镜像报错；运维 SHALL 改为按服务 pull 或先打无后缀全量 tag 构建

---

## compose-redis-topology-2g

<!-- source: openspec/specs/compose-redis-topology-2g/spec.md -->

# compose-redis-topology-2g Specification

## Purpose
TBD - created by archiving change compose-2g-redis-limits. Update Purpose after archive.
## Requirements
### Requirement: 生产 Redis Cluster SHALL 为 3 主 0 从

`manifest/docker/docker-compose.redis-cluster.yml` SHALL 仅定义 **3** 个 Redis 服务（`redis-node-1`..`redis-node-3`），端口 **7001–7003**。仓库 runbook SHALL 文档化初始化命令：`redis-cli --cluster create` 仅包含上述三节点，且 **`--cluster-replicas 0`**。应用 config 中三主种子地址 `redis-node-1:7001,redis-node-2:7002,redis-node-3:7003` SHALL 与拓扑一致，**无需**为缩容修改 Go 代码。

#### Scenario: 生产 cluster 初始化成功

- **WHEN** 运维在空 volume 上启动 3 节点 compose 并执行 documented `cluster create`
- **THEN** `CLUSTER INFO` SHALL 报告 `cluster_state:ok`，且 `CLUSTER NODES` SHALL 显示 3 个 master、0 个 replica

#### Scenario: 生产微服务连接 Redis

- **WHEN** 生产微服务在 `go-ai-talk-net` 上启动且未设置 `GF_REDIS_DEFAULT_ADDRESS`
- **THEN** 进程 SHALL 通过 yaml 默认三主种子连接生产 3 节点 cluster

### Requirement: 测试 Redis SHALL 为单机 standalone

仓库 SHALL 提供 `manifest/docker/docker-compose.redis-standalone.test.yml`（或后继等价文件），定义 **单** Redis 服务（约定服务名 **`redis-test`**，容器端口 **6379**），且 **仅** 加入 `go-ai-talk-test-net`。**SHALL NOT** 要求测试栈执行 `redis-cli --cluster create`。测试栈 **SHALL NOT** 依赖 `docker-compose.redis-cluster.test.yml` 六节点拓扑作为默认路径。

#### Scenario: 测试 Redis 启动无需 cluster create

- **WHEN** 运维 `up -d` 测试 standalone Redis compose
- **THEN** 容器 running 后 SHALL 可直接 `redis-cli PING` 返回 `PONG`，且 **无需** cluster 初始化步骤

#### Scenario: 测试与生产 Redis 网络隔离

- **WHEN** 生产与测试栈同时运行
- **THEN** 测试 Redis 容器 SHALL 不在 `go-ai-talk-net` 上，生产 Redis 容器 SHALL 不在 `go-ai-talk-test-net` 上

### Requirement: 测试 Redis 地址 SHALL 经环境变量注入单机地址

测试部署 MUST 通过 `manifest/docker/.env.test` 设置 **`GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`**（或 runbook documented 等价单地址）。基线 `docker-compose.microservices.yml` SHALL 为需 Redis 的服务提供 `${GF_REDIS_DEFAULT_ADDRESS:-}` 注入；未设置时 SHALL 回退 yaml 默认 cluster 种子（供生产/local cluster 使用）。

#### Scenario: 测试微服务读写 test 单机 Redis

- **WHEN** 测试栈微服务启动且 `GF_REDIS_DEFAULT_ADDRESS=redis-test:6379`
- **THEN** 经 `internal/platform/cachekit`（如启动探活 Ping）或 `internal/platform/redismsgkit` 建立的 Redis 连接 SHALL 指向测试单机 `redis-test:6379`，**SHALL NOT** 连接生产 cluster 节点

---

## controller-process-layout

<!-- source: openspec/specs/controller-process-layout/spec.md -->

# controller-process-layout Specification

## Purpose
TBD - created by archiving change service-package-isolation. Update Purpose after archive.
## Requirements
### Requirement: HTTP controller MUST 按宿主进程分子包

`internal/controller` 下 MUST 按进程划分目录（至少包含：cash、voice、device、history、ucg、gatewayapp、simuser、notify）。各进程的 `Register*ServiceHTTP`（或等价注册入口）MUST 仅绑定本进程子包中的控制器类型。误导性文件名（例如实际由 voice 宿主注册却以 `device_` 为前缀的 care-alert/tip/clinic 控制器）MUST 迁入正确进程子包，并可重命名类型；gateway 反代与跨切面 MUST 归入 gatewayapp（或明确的 proxy 子包并由 gateway-app 注册）。

#### Scenario: voice 注册 care-alert

- **WHEN** 构建并注册 voice-service HTTP
- **THEN** care-alert 控制器 MUST 位于 voice 子包且由 voice 注册入口 Bind，MUST NOT 再与 device-service 注册入口混绑

#### Scenario: cash 仅绑 cash 控制器

- **WHEN** 注册 cash-service
- **THEN** Bind 集合 MUST 仅含 cash 子包控制器（VIP/功能等）

### Requirement: 对外 App 路由字符串 MUST 冻结

本变更中，面向 App/网关已暴露的路由 `path` 字符串（含 `g.Meta` 中的 path 与既有反代匹配模式所依赖的对外前缀）MUST NOT 被修改。允许调整 Go 包路径、类型名与文件名。internal 服务间 path 宜保持不变；若必须调整 MUST 同步所有调用方 clients，且仍 MUST NOT 影响 App 路径。

#### Scenario: care-alert App path 不变

- **WHEN** 完成本变更后的代码审查或自动化检查
- **THEN** App 使用的 `/device/api/care-alert/*`（及同类既有对外 path）MUST 与变更前字符串一致

#### Scenario: 重命名控制器不影响 path

- **WHEN** 将 `DeviceCareAlertController` 重命名并移入 `controller/voice`
- **THEN** 其 `g.Meta` path 字段 MUST 仍为原对外路径

---

## dao-extension-layer-simplification

<!-- source: openspec/specs/dao-extension-layer-simplification/spec.md -->

# dao-extension-layer-simplification Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: DAO extension files SHALL follow minimum-necessary rule
DAO `*_ext.go` files MUST be retained only when they provide business-meaningful extensions beyond generated DAO wrappers.

#### Scenario: Ext file has no added behavior
- **WHEN** an ext file only duplicates generated DAO behavior without business logic
- **THEN** the file MUST be merged away or removed

#### Scenario: Ext file provides service-specific query semantics
- **WHEN** an ext file includes domain query composition or behavior not present in generated DAO
- **THEN** the file MAY be retained with explicit comment/documented rationale

---

## database-unix-timestamp-storage

<!-- source: openspec/specs/database-unix-timestamp-storage/spec.md -->

# database-unix-timestamp-storage Specification

## Purpose
TBD - created by archiving change database-unix-timestamp-storage. Update Purpose after archive.
## Requirements
### Requirement: 时间类字段落库形态

除经架构评审明确豁免的「纯日历日期」字段外，凡表示**时刻**（事件发生时间、创建/更新时间、最后活跃时间等）的数据库列 **MUST** 以 **Unix 时间戳秒** 数值存储，MySQL 类型 **MUST** 为可表达该范围的整数类型（推荐 `BIGINT`）。**MUST NOT** 将本地墙钟格式化的日期时间字符串作为权威落库值。

#### Scenario: 新表创建

- **WHEN** 新建包含「时刻」语义的表或列
- **THEN** 该列类型为整数型时间戳秒且注释标明 UTC 纪元秒，且应用写入路径使用 UTC 纪元秒（如 `time.Time.Unix()`）而非格式化字符串

### Requirement: API 与 JSON 契约

对外 HTTP JSON 中代表「时刻」的字段 **MUST** 使用数字类型（Unix 秒），与数据库存储单位一致；字段文档或 OpenAPI **MUST** 标明单位为秒。若迁移期需兼容旧客户端，**MUST** 在变更说明中定义弃用截止条件，且服务端权威值仍为数字戳。

#### Scenario: 客户端解析

- **WHEN** 客户端接收代表事件发生时刻的字段
- **THEN** 该值为 JSON number（Unix 秒），客户端在展示给用户时自行按目标时区转换，不依赖服务端返回本地日历字符串作为权威

### Requirement: 迁移与数据完整性

对已有「非数字时刻」列的迁移 **MUST** 提供可重复执行的回填策略，并在切换读写前完成行数一致性与抽样校验。**MUST** 定义 NULL/非法旧值的处置规则（拒绝写入、置 0 或置哨兵值须文档化且经评审）。

#### Scenario: 回填后校验

- **WHEN** 执行从旧列到新秒级列的回填脚本
- **THEN** 存在自动化或清单式校验（行数、非 NULL 比例、时间范围合理性）且通过后应用才切换为只读新列

### Requirement: 服务边界

各服务 **MUST** 仅修改本服务拥有库内的表与 DAO；跨服务时间语义通过契约（HTTP/RPC/事件）传递，传递值 **MUST** 为 Unix 秒或与契约显式声明的单位一致。

#### Scenario: history 与 device 分库

- **WHEN** 在 history-service 所属库中迁移时间列
- **THEN** 不修改 device-service 所属库的表结构于同一提交中混写；各自变更独立可发布

---

## deepseek-history-redis-prefer

<!-- source: openspec/specs/deepseek-history-redis-prefer/spec.md -->

# deepseek-history-redis-prefer Specification

## Purpose
TBD - created by archiving change async-redis-read-model-for-history-action-event-user. Update Purpose after archive.
## Requirements
### Requirement: DeepSeek 历史读取 Redis 优先

系统在执行 **voiceUnderstanding lane** 所触发的 LLM 意图分析与对话补全前，历史上下文读取 SHALL 优先命中 Redis 历史读模型，并在不可用时回源到历史服务或数据库。

#### Scenario: LLM 请求前 Redis 命中

- **WHEN** voiceUnderstanding 路径请求前需要最近历史上下文且 Redis 命中
- **THEN** 系统 MUST 使用缓存历史组装 prompt

#### Scenario: 缓存未命中回源

- **WHEN** voiceUnderstanding 路径请求前历史缓存未命中
- **THEN** 系统 MUST 回源获取历史并回填缓存后继续调用 LLM

### Requirement: 上下文读取一致性与降级可观测

系统 MUST 对 **voiceUnderstanding** LLM 上下文读取提供命中率、回源率与降级原因可观测性，并在异常时保持功能可用。

#### Scenario: 请求完成可观测

- **WHEN** 任一 voiceUnderstanding LLM 请求完成上下文装配
- **THEN** 系统 SHOULD 输出结构化日志含缓存命中/回源标记

#### Scenario: Redis 不可用降级

- **WHEN** Redis 读模型暂时不可用
- **THEN** 系统 MUST 降级回源并继续完成 LLM 调用，同时输出结构化告警日志

### Requirement: 历史窗口语义一致

系统读取最近 N 小时历史用于 **voiceUnderstanding** LLM 时 MUST 遵守既有 history 读模型契约，MUST NOT 新增跨库直查他域表。

#### Scenario: 历史问答使用读模型

- **WHEN** 历史问答等价路径加载 12 小时历史
- **THEN** MUST 经 Redis 优先读模型路径

---

## device-admin

<!-- source: openspec/specs/device-admin/spec.md -->

# device-admin Specification

## Purpose
TBD - created by archiving change fix-api-usage-stats. Update Purpose after archive.
## Requirements
### Requirement: device admin SHALL provide paginated wx account list

`device-service` MUST 暴露 `GET /device/admin/api/wx/list`，鉴权与现有 device admin 一致（Header `X-Admin-Password`）。查询参数：`page`（默认 1）、`pageSize`（默认 20，最大 100）、可选 `q`（模糊匹配 wx.id、deviceNo、unionid、account）。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`（用户名账号若有）、`createdAt`（wx 账号创建 Unix 秒，0 表示未知）。列表 MUST **仅**包含 `is_simulated=0` 的真实用户；`is_simulated=1` 的行 MUST NOT 出现在 `list` 或 `total` 计数中。

#### Scenario: 默认分页列表

- **WHEN** 管理员携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/wx/list` 且未传分页参数
- **THEN** 响应 SHALL 返回第一页 wx 记录且 `page=1`
- **AND** `list` MUST NOT 含 `is_simulated=1` 的 wxId

#### Scenario: 关键字搜索

- **WHEN** 管理员请求 `q=138` 且存在 deviceNo 或 id 匹配的非模拟 wx 行
- **THEN** 响应 `list` SHALL 仅包含匹配的非模拟项

#### Scenario: 经 gateway-app 反代可达

- **WHEN** 管理员经 gateway-app 携带 Admin JWT 请求 `/device/admin/api/wx/list`
- **THEN** gateway-app SHALL 反代至 device-service 并成功返回列表

#### Scenario: createdAt on new account

- **WHEN** 列表项对应迁移后新注册的 wx 账号
- **THEN** 该项 `createdAt` MUST 大于 0

#### Scenario: createdAt zero displays as unknown

- **WHEN** 列表项 `createdAt` 为 0
- **THEN** 消费方展示层 MAY 显示「—」

---

## device-admin-event-logo-color-ui

<!-- source: openspec/specs/device-admin-event-logo-color-ui/spec.md -->

# device-admin-event-logo-color-ui Specification

## Purpose
TBD - created by archiving change device-admin-event-logo-color-ui. Update Purpose after archive.
## Requirements
### Requirement: 事件管理列表展示 Logo 与色调

设备管理页（`admin.html` 或等价路由）在登录并加载事件列表后，SHALL 在**树形**表格中展示 **Logo** 与 **色调** 列；每行（含根、中间与叶子节点）SHALL 根据 `GET /device/admin/api/event/list` 返回的 `logo`、`color`、`parentId` 渲染预览与层级缩进。

#### Scenario: 列表含 logo 与 color 字段时展示预览

- **WHEN** 事件列表项包含 `logo` 路径与有效 `color` 色值
- **THEN** 页面 SHALL 在 Logo 列显示可识别的缩略图
- **AND** 色调列 SHALL 显示与 `color` 一致的色块及可读色值文本

#### Scenario: 无 logo 时展示可点击占位

- **WHEN** 事件项 `logo` 为空
- **THEN** Logo 列 SHALL 显示明确占位（如「点击上传」）
- **AND** 占位区域 SHALL 可触发 logo 更换流程

#### Scenario: 子节点独立展示父级同级列

- **WHEN** 子事件在树中缩进展示
- **THEN** 该行 SHALL 仍含 Logo 与色调列且使用**该子事件自身**的 `logo`/`color` 字段

### Requirement: 管理页 Logo 预览使用同源 URL

管理页用于 `<img src>` 的 logo 地址 SHALL 直接使用 API 返回的 CDN 绝对 URL；SHALL NOT 再拼接当前页面 origin 与 `/ai_talk_images/` path。

#### Scenario: API 返回 CDN logo 时展示

- **WHEN** `GET /device/admin/api/event/list` 返回 `logo` 为 `https://resorce.cuplay.top/event/...`
- **THEN** 页面 SHALL 将该 URL 直接用于 `<img src>`

#### Scenario: 无 logo 时展示可点击占位

- **WHEN** 事件项 `logo` 为空
- **THEN** Logo 列 SHALL 显示明确占位（如「点击上传」）
- **AND** 占位区域 SHALL 可触发 logo 更换流程

### Requirement: 主网关提供同源事件图片访问

gateway-service（管理页常用入口，如 :9701）SHALL 注册 `GET /ai_talk_images/*`（及 HEAD），将请求反代或等价转发至 device-service 的同名静态读能力，使管理页同源 URL 可成功返回图片。

#### Scenario: 经主网关读取已上传 logo

- **WHEN** 客户端请求 `GET https://<gateway-host>/ai_talk_images/<安全文件名>` 且 device-service 上文件存在
- **THEN** gateway-service SHALL 返回对应图片内容且 SHALL NOT 要求 App 网关 Bearer

### Requirement: 点击色调即可更新 color

管理页 SHALL 允许用户通过点击列表行中的色调展示区域修改该事件的 `color`，并在成功后刷新列表。

#### Scenario: 点击色块修改 color

- **WHEN** 用户点击某行色调区域并选择新色值后确认提交
- **THEN** 客户端 SHALL 调用 `POST /device/admin/api/event/update`（multipart）并携带该行 `id`、`name`、`needQuantity`、`extraNames` 及新 `color`
- **AND** 未选择新 logo 文件时 SHALL NOT 清除原有 `logo`

#### Scenario: 更新成功后列表反映新色值

- **WHEN** 更新接口返回成功
- **THEN** 页面 SHALL 刷新事件列表且该行色调展示与新 `color` 一致

### Requirement: 点击 Logo 即可更新 logo 文件

管理页 SHALL 允许用户通过点击列表行中的 Logo 缩略图或占位触发文件选择，上传新图并更新该事件。

#### Scenario: 点击 Logo 上传新图

- **WHEN** 用户点击 Logo 区域并选择合法图片文件（如 png/jpeg/webp）
- **THEN** 客户端 SHALL 以 multipart 调用 `POST /device/admin/api/event/update`，包含 `logo` 文件及该行完整文本字段
- **AND** 成功后服务端 `event.logo` SHALL 更新为 path-only 新路径

#### Scenario: 更新成功后列表展示新缩略图

- **WHEN** logo 更新成功且列表重新加载
- **THEN** 该行 Logo 列 SHALL 使用同源 URL 展示新图

### Requirement: 行内编辑与弹窗编辑并存

名称、事件扩展、是否需要计数等字段 SHALL 仍可通过既有「编辑」弹窗修改；行内交互仅负责 **logo** 与 **color**，SHALL NOT 要求用户仅为改色/改图打开完整弹窗。

#### Scenario: 编辑按钮仍打开完整表单

- **WHEN** 用户点击行内「编辑」按钮
- **THEN** 页面 SHALL 打开包含名称与其它字段的编辑弹窗（行为与变更前一致）

### Requirement: 树形列表每行可新增子事件

除顶部「新增事件」外，事件管理页每一行 SHALL 提供「新增子事件」操作，打开创建表单并携带该行 id 作为 `parentId`。

#### Scenario: 父行展示新增子事件按钮

- **WHEN** 用户查看事件树中任意节点行
- **THEN** 操作列 SHALL 含「新增子事件」入口

#### Scenario: 新增子事件成功后树刷新

- **WHEN** 用户通过「新增子事件」成功创建记录
- **THEN** 页面 SHALL 重新加载列表且新节点出现在对应父节点下

---

## device-admin-event-parent-picker-ui

<!-- source: openspec/specs/device-admin-event-parent-picker-ui/spec.md -->

# device-admin-event-parent-picker-ui Specification

## Purpose
TBD - created by archiving change device-event-update-parent-id. Update Purpose after archive.
## Requirements
### Requirement: 编辑事件时 SHALL 可选择父事件

设备管理页在**编辑**已有事件时，SHALL 提供父事件选择控件（含「无 / 根」选项，对应 `parentId=0`）。提交 `POST /device/admin/api/event/update` 时 SHALL 在 multipart 表单中包含 **`parentId`** 字段。

#### Scenario: 打开编辑弹窗默认选中当前父

- **WHEN** 管理员编辑 `parentId=5` 的事件
- **THEN** 父事件选择器 SHALL 默认选中 id=5 的项（或等价展示父名称）

#### Scenario: 提交修改父节点

- **WHEN** 管理员将父改为根并保存
- **THEN** 请求 SHALL 包含 `parentId=0`
- **AND** 成功后列表树形结构 SHALL 反映该节点位于根层

### Requirement: 父事件选择器 SHALL 排除非法选项

选择器 SHALL NOT 提供**当前事件自身**及其**全部后代**作为父选项，以避免必然触发后端成环校验失败。

#### Scenario: 编辑叶子事件时不出现自身为父

- **WHEN** 编辑 id=20 的叶子事件
- **THEN** 父事件下拉 SHALL NOT 包含 id=20

#### Scenario: 编辑有子节点时不出现其子孙为父

- **WHEN** 编辑 id=10 且存在 `parent_id=10` 的子事件 20
- **THEN** 父事件下拉 SHALL NOT 包含 id=20

---

## device-admin-event-tree-ui

<!-- source: openspec/specs/device-admin-event-tree-ui/spec.md -->

# device-admin-event-tree-ui Specification

## Purpose
TBD - created by archiving change device-event-hierarchy. Update Purpose after archive.
## Requirements
### Requirement: 事件管理页树形展示层级

设备管理页事件模块 SHALL 根据 `ListEvents` 返回的扁平数组（含 `parentId`）渲染**树形**列表：根节点按 id 或现有排序规则排列，子节点缩进展示在其父节点之下；深度 SHALL 支持通用树（不限两级）。

#### Scenario: 换尿布与子事件分级可见

- **WHEN** 列表含 `换尿布(parentId=0)` 与 `大便(parentId=换尿布.id)`
- **THEN** 页面 SHALL 将「大便」行展示在「换尿布」之下并带可视缩进

#### Scenario: 多级中间节点可展开式展示

- **WHEN** 存在根 → 中间 → 叶子三级关系
- **THEN** 页面 SHALL 按 parentId 递归嵌套展示全部层级

### Requirement: 支持新增根事件与新增子事件

页面 SHALL 提供「新增事件」创建根节点；每一行（含中间节点）SHALL 提供「新增子事件」入口，提交时携带 `parentId` 为该行 id。

#### Scenario: 从换尿布行新增子事件

- **WHEN** 用户点击「换尿布」行的「新增子事件」
- **AND** 填写名称「小便」并提交
- **THEN** 客户端 SHALL `POST /device/admin/api/event/add` 且表单含 `parentId=<换尿布.id>`
- **AND** 成功后列表 SHALL 在「换尿布」下展示「小便」

#### Scenario: 新增根事件不带 parentId

- **WHEN** 用户点击顶部「新增事件」并提交
- **THEN** 请求 SHALL NOT 携带非零 `parentId`（或显式 `parentId=0`）

### Requirement: 树形列表保留 Logo 与色调行内编辑

树形结构中每一节点 SHALL 独立展示并支持行内 **Logo**、**色调** 编辑（行为与扁平列表时期一致）；子节点 SHALL NOT 因存在父节点而隐藏 logo/color 列。

#### Scenario: 中间节点可上传独立 Logo

- **WHEN** 用户为中间节点「排泄类」点击 Logo 上传新图
- **THEN** 仅该节点 `logo` SHALL 更新
- **AND** 父节点「换尿布」的 `logo` SHALL 保持不变

### Requirement: 子事件创建表单不预填父 logo 与 color

打开「新增子事件」弹窗时，色调与 Logo SHALL 使用与「新增根事件」相同的默认空状态，SHALL NOT 预填父节点当前 `color` 或 `logo` 预览为默认值。

#### Scenario: 子事件弹窗色值非父色

- **WHEN** 父节点 color 为 `#FF0000`
- **AND** 用户打开该父下的「新增子事件」弹窗
- **THEN** 颜色选择器 SHALL NOT 因父色而默认选中 `#FF0000`（除非用户手动选择）

---

## device-admin-user-list

<!-- source: openspec/specs/device-admin-user-list/spec.md -->

# device-admin-user-list Specification

## Purpose
TBD - created by archiving change device-admin-user-list-pagination. Update Purpose after archive.
## Requirements
### Requirement: 管理端设备记录分页列表

device-service SHALL 提供 `GET /device/admin/api/user/list`，要求 Header `X-Admin-Password` 有效。查询参数 `page`（默认 1）、`pageSize`（默认 5，最大 100）、可选 `q`（`device_no` 模糊包含，大小写不敏感以库排序规则为准）。响应 MUST 为 `{ list, total, page, pageSize }`，`list` 每项含 `deviceNo`、`activeTime`、`lastTalkTime`、`lastTalkAsk`、`lastTalkAnswer`、`lastApiPath`、`lastApiAt`。

#### Scenario: 默认分页

- **WHEN** 管理员请求 `GET /device/admin/api/user/list` 且未传 `pageSize`
- **THEN** 返回最多 5 条记录且 `pageSize` 字段为 5

#### Scenario: 模糊搜索

- **WHEN** 管理员请求带 `q=abc`
- **THEN** 仅返回 `device_no` 包含子串 `abc` 的设备

### Requirement: 最近 HTTP 接口落库

对任意经网关处理的 HTTP 请求，若可解析出非空 `deviceNo` 且路径不是 WebSocket、不以 `/device/internal/` 开头，系统 SHALL 异步更新该设备 `last_api_path`（`METHOD /path`）与 `last_api_at`（Unix 秒）。WebSocket 升级请求 MUST NOT 触发更新。

#### Scenario: 带 query 的 history 列表

- **WHEN** 客户端请求 `GET /device/history/api/list?deviceNo=d1&page=1`
- **THEN** 设备 `d1` 的 `last_api_path` 更新为 `GET /device/history/api/list`

### Requirement: 管理端设备号跳转历史页

`admin.html` 设备记录表中 `device_no` MUST 为指向 `/device/history/{deviceNo}` 的链接（URL 编码 deviceNo）。

#### Scenario: 点击设备号

- **WHEN** 管理员点击列表中某行的设备号链接
- **THEN** 浏览器导航至同源的 `/device/history/{deviceNo}` 历史管理页

---

## device-app-device-login

<!-- source: openspec/specs/device-app-device-login/spec.md -->

# device-app-device-login Specification

## Purpose
TBD - created by archiving change device-app-device-login. Update Purpose after archive.
## Requirements
### Requirement: device-service 提供设备号业务登录

device-service SHALL 提供 **`POST /device/app/api/user/device_login`**，从 JSON body 读取 **`deviceNo`**（字符串，trim 后非空）。系统 SHALL 校验该设备号已在设备域注册表中注册。若存在 **`wx` 表行**其 **`device_no`** 与该值一致，响应 **`data.wxId`** SHALL 为该 wx 主键；若无绑定 wx 行，**`wxId` SHALL 为 `0`**（仍返回 **`deviceNo`**）。**`isNewUser`** 在设备号登录场景 SHALL 为 `false`。响应 SHALL NOT 包含由 gateway-app 签发的 **`accessToken`/`refreshToken`**。

#### Scenario: 已注册且已绑定 wx 的设备登录成功

- **WHEN** `deviceNo` 对应已注册设备且 wx 行已绑定该 `device_no`
- **THEN** 系统 SHALL 返回 `code=0` 且 `data` 含非零 **`wxId`**、**`deviceNo`**，且 **`isNewUser` 为 false**

#### Scenario: 已注册但未绑定 wx

- **WHEN** 设备已注册但无 wx 行绑定该 `device_no`
- **THEN** 系统 SHALL 返回 `code=0` 且 **`wxId` 为 0**、**`deviceNo`** 正确、**`isNewUser` 为 false**

#### Scenario: 设备不存在

- **WHEN** `deviceNo` 在设备注册表中不存在
- **THEN** 系统 SHALL 返回非 0 业务码及明确错误语义，且 SHALL NOT 返回 token

### Requirement: gateway-app 聚合设备号登录并签发令牌

gateway-app-server SHALL 提供 **`POST /device/app/api/device_login`**，将请求体（至少 **`deviceNo`**）转发至 device-service 的 **`POST /device/app/api/user/device_login`**；当 device 返回成功时，SHALL 签发 **`accessToken`**（JWT：`sub` 为 wx 主键，**无 wx 时 `sub` 为 `"0"`**，且 MUST 含 **`device_no` claim**）与 **`refresh_token`**（wx 会话载荷为纯数字 wxId；**`sub`=0 的会话** SHALL 在 refresh 侧携带 **`device_no`** 以便旋转 refresh 时恢复 claim）。该路径 SHALL 列入 **Bearer 鉴权白名单**。

#### Scenario: 聚合登录成功（含 wxId=0）

- **WHEN** 客户端调用 **`POST /device/app/api/device_login`** 且 body 中 `deviceNo` 在 device 侧校验通过
- **THEN** 响应 SHALL 包含 **`accessToken`/`refreshToken`** 及与 device 返回一致的 **`wxId`、`deviceNo`、`isNewUser`**

#### Scenario: device 业务失败

- **WHEN** device 返回非 0 或缺少 **`deviceNo`**
- **THEN** 网关 SHALL NOT 签发 token，且 SHALL 向客户端返回明确错误语义

### Requirement: 联调页提供设备号登录调试

`resource/public/gateway-app-integration-test.html` SHALL 提供用户可触发的操作，向当前配置的网关基址发起 **`POST /device/app/api/device_login`**（`Content-Type: application/json`，body 含 **`deviceNo`**），并将响应中的 token 与业务字段展示在页面日志区（与现有登录区块并列或分区清晰）。

#### Scenario: 用户点击设备登录

- **WHEN** 用户填写 `deviceNo` 并触发设备登录操作
- **THEN** 页面 SHALL 发起上述 HTTP 请求并 SHALL 展示成功或失败的可读结果

---

## device-event-cache-rebuild-on-mutate

<!-- source: openspec/specs/device-event-cache-rebuild-on-mutate/spec.md -->

# device-event-cache-rebuild-on-mutate Specification

## Purpose
TBD - created by archiving change device-event-cache-rebuild-on-mutate. Update Purpose after archive.
## Requirements
### Requirement: 事件表变更后 Redis 缓存必须从数据库重建

当 `device-service` 成功执行对 `event` 表的插入、更新或删除后，系统 SHALL 使用数据库当前全量事件行（含 `logo`、`color` 与 **`parent_id`**）重建 Redis 中的事件选项缓存，且 SHALL NOT 通过先调用 `ListEvents`（可能仅返回变更前缓存）再写回的方式刷新缓存。

#### Scenario: 更新事件 color 后缓存含新色值

- **WHEN** 管理员通过 API 成功更新某事件的 `color`
- **THEN** 随后对 `ListEvents` 或等价读路径的调用在缓存命中时 SHALL 返回包含新 `color` 的该事件行

#### Scenario: 更新事件 logo 后缓存含新 path

- **WHEN** 管理员成功上传并更新某事件的 `logo` 路径
- **THEN** Redis 事件选项快照中该事件的 `logo` 字段 SHALL 与数据库一致

#### Scenario: 新增事件后缓存含新行

- **WHEN** 管理员成功新增一条事件记录
- **THEN** 随后缓存命中时 SHALL 包含该新事件

#### Scenario: 删除事件后缓存不含已删行

- **WHEN** 管理员成功删除一条事件记录
- **THEN** 随后缓存命中时 SHALL NOT 包含已删除的事件 id

#### Scenario: 新增子事件后缓存含 parentId

- **WHEN** 管理员成功新增 `parent_id=5` 的子事件
- **THEN** 随后缓存命中时该事件行 SHALL 包含 `parentId=5`

### Requirement: 写后刷新失败可观测

若重建 Redis 缓存失败，系统 SHALL 记录警告级别日志且 SHALL NOT 将写库事务回滚（写库已成功）。

#### Scenario: Redis 不可用时的行为

- **WHEN** 数据库写入成功但 `RebuildEventCache` 因 Redis 错误失败
- **THEN** 系统 SHALL 记录可观测警告日志
- **AND** API 仍可对客户端返回写库成功（与现网语义一致）

---

## device-event-hierarchy

<!-- source: openspec/specs/device-event-hierarchy/spec.md -->

# device-event-hierarchy Specification

## Purpose
TBD - created by archiving change device-event-hierarchy. Update Purpose after archive.
## Requirements
### Requirement: 事件表以 parent_id 表达通用树

`device-service` 持久化的 `event` 行 SHALL 使用 `parent_id` 表示层级：`0`（或等价空值约定）为根节点；非零值 MUST 指向已存在的父事件 id。系统 SHALL NOT 在业务逻辑中读写 `child_ids` 列。

#### Scenario: 新增子事件写入 parent_id

- **WHEN** 管理员提交 `POST /device/admin/api/event/add` 且表单 `parentId=5`
- **THEN** 新行 `parent_id` SHALL 为 `5`
- **AND** 父行 SHALL NOT 依赖 `child_ids` 维护子列表

#### Scenario: 根事件 parent_id 为零

- **WHEN** 管理员新增根事件且未提交 `parentId` 或 `parentId=0`
- **THEN** 新行 `parent_id` SHALL 为 `0`

### Requirement: 同父下事件名唯一

创建或更新事件时，系统 SHALL 在相同 `parent_id` 下保证 `name` 唯一；不同 `parent_id` 下 MAY 存在相同 `name`。

#### Scenario: 同父重复名称被拒绝

- **WHEN** 父 id=5 下已存在名为「大便」的事件
- **AND** 客户端在同一 `parentId=5` 下再次提交 `name=大便`
- **THEN** API SHALL 返回业务错误且 SHALL NOT 插入

#### Scenario: 不同父允许同名

- **WHEN** 父 id=5 下已存在「其他」
- **AND** 客户端在 `parentId=10` 下提交 `name=其他`
- **THEN** API SHALL 允许创建

### Requirement: 有子节点的事件不可删除

`DeleteEvent` SHALL 在存在任意 `parent_id` 等于待删 id 的行时拒绝删除。

#### Scenario: 删除有子的父事件失败

- **WHEN** 事件 id=5 存在 `parent_id=5` 的子行
- **AND** 客户端请求删除 id=5
- **THEN** API SHALL 返回可识别业务错误
- **AND** 数据库 SHALL 保留 id=5 行

#### Scenario: 删除叶子事件成功

- **WHEN** 事件 id=12 无子行
- **THEN** 删除 SHALL 成功且 SHALL 触发事件缓存重建

### Requirement: ListEvents 返回 parentId

`GET /device/admin/api/event/list` 及内部 `ListEvents` 契约 SHALL 在每条事件记录中包含 `parentId` 字段。

#### Scenario: 列表含 parentId

- **WHEN** 客户端请求事件列表
- **THEN** 每项 SHALL 包含与数据库 `parent_id` 一致的 `parentId`

### Requirement: 新增子事件不继承父 logo 与 color

带非零 `parentId` 创建事件时，系统 SHALL NOT 从父行复制 `logo` 或 `color`；新行视觉字段 SHALL 仅来自本次提交或系统默认值。

#### Scenario: 子事件使用表单色值而非父色

- **WHEN** 父事件 `color=#FF0000`
- **AND** 子事件创建表单提交 `color=#4A90D9` 与 `parentId=5`
- **THEN** 新行 `color` SHALL 为 `#4A90D9`
- **AND** SHALL NOT 自动设为 `#FF0000`

---

## device-event-logo-color

<!-- source: openspec/specs/device-event-logo-color/spec.md -->

# device-event-logo-color Specification

## Purpose
TBD - created by archiving change device-event-logo-and-path-only-assets. Update Purpose after archive.
## Requirements
### Requirement: 事件 logo 与 color SHALL 可配置且列表可见

device-service 事件字典 MUST 支持 `logo`（OSS objectKey，前缀 `event/`）与 `color` 的持久化；所有返回事件字典列表的 HTTP 接口 MUST 在 JSON 中将 `logo` 序列化为 CDN 绝对 URL（无 logo 时为空串），并 MUST 包含 `color` 字段。Redis 事件选项缓存（`device:event:options:all`）MAY 仅持久化 objectKey；**任何** HTTP 响应边界 MUST 在返回前完成 CDN 映射，不得因缓存命中而跳过。

#### Scenario: 管理端事件列表返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/admin/api/event/list`
- **THEN** 响应 `list[]` 中每项 MUST 包含 `logo` 与 `color` 字段
- **AND** `logo` 若有值 MUST 为 CDN 绝对 URL（`https://` 开头）

#### Scenario: 历史与内部事件选项返回 logo 与 color

- **WHEN** 客户端请求 `GET /device/history/api/event/options` 或 `GET /device/internal/api/event/options`
- **THEN** 响应 `list[]` MUST 同样包含 CDN 形式的 `logo` 与 `color`

#### Scenario: 管理端更新 logo 后 history 缓存命中仍返回 CDN

- **WHEN** 管理员成功更新某事件 logo 且 device-service 已重建 Redis 事件选项缓存（缓存内 logo 为 objectKey）
- **AND** App 或客户端请求 `GET /device/history/api/event/options` 且 history-service 命中该 Redis 键
- **THEN** 响应 `list[]` 中每项 `logo` 若有值 MUST 仍为 CDN 绝对 URL（`https://` 开头）
- **AND** MUST NOT 返回裸 objectKey（如 `event/2026/06/xxx.png`）

### Requirement: 事件新增与更新 SHALL 支持 multipart 上传 logo

`POST /device/admin/api/event/add` 与 `POST /device/admin/api/event/update` MUST accept `multipart/form-data`，至少包含表单字段 `name`、`eventType`、`extraNames`、`color`；`update` MUST 包含 `id`。可选文件字段名 MUST 为 `logo`。

#### Scenario: 新增事件并上传 logo

- **WHEN** 客户端 multipart 提交有效 `name` 与合法图片 `logo`
- **THEN** 服务端 MUST 经 ucg internal 上传至 OSS
- **AND** MUST 将 `event.logo` 设为 `event/...` objectKey
- **AND** MUST 将 `color` 写入 `event.color`（若提供）

#### Scenario: 更新事件未传 logo 保留原值

- **WHEN** 客户端对已有事件 multipart 更新且未包含 `logo` 文件
- **THEN** 服务端 MUST 保留原 `event.logo` objectKey
- **AND** MAY 更新 `color` 与其它文本字段

### Requirement: 事件 Redis 缓存投影 SHALL 含 logo 与 color

`ListEvents` 使用的 Redis 事件选项缓存 MUST 与数据库查询一致，包含 `logo` 与 `color`，以便缓存命中时列表仍返回完整字段。

#### Scenario: 缓存命中仍含 logo

- **WHEN** `ListEvents` 从 Redis 命中事件选项
- **THEN** 返回的 `[]Event` MUST 含非省略的 `logo`、`color` 字段

### Requirement: history 写 Redis 事件选项缓存 SHALL 仅持久化 objectKey

history-service 向 `device:event:options:all` 写入事件选项快照时，SHALL 将每项 `logo` 规范为 OSS objectKey（无前缀域名）；SHALL NOT 将 CDN 绝对 URL 写入共享 Redis 键，以避免与 device-service 重建缓存的 objectKey 格式混写。

#### Scenario: history 回源 internal 后写缓存

- **WHEN** history-service cache miss 后从 device internal 拉取事件列表（响应 logo 已为 CDN URL）
- **AND** history-service 将结果写入 Redis 事件选项键
- **THEN** 写入 Redis 的 JSON 中 `logo` MUST 为 objectKey 形式
- **AND** 后续 history HTTP 响应 MUST 仍通过 CDN 映射返回绝对 URL

---

## device-event-type-field

<!-- source: openspec/specs/device-event-type-field/spec.md -->

# device-event-type-field Specification

## Purpose
TBD - created by archiving change event-type-replace-need-quantity. Update Purpose after archive.
## Requirements
### Requirement: 事件主档必须持久化有效的 event_type

`device-service` 在创建或更新 `event` 表记录时，SHALL 接受并持久化 `event_type`，其值 MUST 为 `number`、`time` 或 `one` 之一。系统 SHALL NOT 再读写 `need_quantity` 列或 API 字段 `needQuantity`。

#### Scenario: 管理端新增事件带 eventType

- **WHEN** 客户端 `POST /device/admin/api/event/add` 提交合法 `eventType=number`
- **THEN** 数据库新行 `event_type` SHALL 为 `number`
- **AND** 随后 `ListEvents` 或缓存命中 SHALL 返回该 `eventType`

#### Scenario: 非法 eventType 被拒绝

- **WHEN** 客户端提交 `eventType` 为空或不在 `number|time|one`
- **THEN** API SHALL 返回参数错误且 SHALL NOT 插入或更新行

### Requirement: 事件选项 Redis 快照含 event_type

写库成功后，系统 SHALL 通过从数据库全量扫描（含 `event_type` 列）重建 Redis 事件 options，且 SHALL NOT 依赖可能过期的 `ListEvents` 缓存读回后写回。

#### Scenario: 更新事件后缓存含新类型

- **WHEN** 管理员成功更新某事件的 `eventType` 为 `one`
- **THEN** 重建后的 Redis 快照中该事件 SHALL 带有 `eventType` 为 `one`

### Requirement: 匹配已有事件时不改 event_type

当仅合并别名（`extra_names`）或命中已有事件名时，系统 SHALL NOT 更新该事件行的 `event_type`。

#### Scenario: DeepSeek 仅追加别名

- **WHEN** 抽取结果匹配已存在事件名且仅合并 `extraNames`
- **THEN** 该事件 `event_type` 列 SHALL 保持不变

### Requirement: 对话新建事件时由模型提供 event_type

经 voice 调用的 `InsertOrGetEventByNeedle` 或 DeepSeek 落库插入新事件时，系统 SHALL 将模型给出的 `event_type` 写入新行；若模型未给出合法值，SHALL 使用规范化默认值（`time`）并仍可观测。

#### Scenario: 语音路径插入新事件

- **WHEN** 用户话术导致新建事件且 DeepSeek 返回 `event_type` 为 `number`
- **THEN** 新插入的 `event` 行 `event_type` SHALL 为 `number`

---

## device-event-update-parent-id

<!-- source: openspec/specs/device-event-update-parent-id/spec.md -->

# device-event-update-parent-id Specification

## Purpose
TBD - created by archiving change device-event-update-parent-id. Update Purpose after archive.
## Requirements
### Requirement: UpdateEvent SHALL 支持修改 parent_id

`device-service` 的 `UpdateEvent`（及 `POST /device/admin/api/event/update`）SHALL 接受 **`parentId`**（非负整数，`0` 表示根）。当请求携带有效 `parentId` 且与库内当前值不同时，系统 SHALL 更新该行的 `parent_id` 字段，并 SHALL 在成功写库后触发与现有事件变更一致的 Redis 事件选项缓存重建。

#### Scenario: 将事件挂到新的父节点下

- **WHEN** 管理员提交 `id=10`、`parentId=5`，且 id=5 存在、不构成环
- **THEN** id=10 行的 `parent_id` SHALL 变为 `5`
- **AND** 随后 `ListEvents` / 缓存中该项的 `parentId` SHALL 为 `5`

#### Scenario: 将事件提升为根节点

- **WHEN** 管理员提交 `id=10`、`parentId=0`
- **THEN** id=10 行的 `parent_id` SHALL 为 `0`

#### Scenario: 未变更父节点时仅更新其他字段

- **WHEN** 管理员仅修改 `name` 且提交的 `parentId` 与库内一致
- **THEN** 系统 SHALL 仅更新非层级字段，且 SHALL NOT 产生无效的父节点写操作错误

### Requirement: 修改 parent_id 须校验父存在且无环

当 `parentId > 0` 时，系统 SHALL 校验对应父事件行存在。系统 SHALL 拒绝 `parentId` 等于待更新事件自身 id，SHALL 拒绝将父设为其**任意后代**（防止环）。违反时 SHALL 返回业务错误且**不得**部分更新。

#### Scenario: 父节点不存在

- **WHEN** 提交 `parentId=99999` 且库中无 id=99999
- **THEN** 请求 SHALL 失败并返回可识别的错误信息

#### Scenario: 父节点为自身

- **WHEN** 提交 `id=10`、`parentId=10`
- **THEN** 请求 SHALL 失败

#### Scenario: 父节点为子孙造成环

- **WHEN** 事件 10 拥有后代 20（`parent_id=10`）
- **AND** 提交 `id=10`、`parentId=20`
- **THEN** 请求 SHALL 失败

### Requirement: 修改 parent_id 后同父 name 唯一

更新名称或父节点后，系统 SHALL 在**目标父** `parent_id` 下保证 `name` 唯一（排除自身 id），规则与 `device-event-hierarchy` 中 AddEvent 一致。

#### Scenario: 移动后与兄弟同名冲突

- **WHEN** 父 5 下已存在 `name=大便` 的事件
- **AND** 将另一事件移至 `parentId=5` 且 `name=大便`
- **THEN** 请求 SHALL 失败并返回与「事件已存在」一致的业务错误

### Requirement: 有子节点的事件 MAY 修改 parent_id

存在 `parent_id = 待更新 id` 的子行时，系统 **MAY** 允许修改该事件的 `parent_id`；子行的 `parent_id` SHALL 仍指向原 id，除非另有删除/移动子树需求。

#### Scenario: 中间节点更换父级而子节点仍挂在其下

- **WHEN** 事件 10 有子事件 20（`parent_id=10`）
- **AND** 成功将事件 10 的 `parent_id` 改为 5
- **THEN** 事件 20 的 `parent_id` SHALL 仍为 `10`

---

## device-ops-ai-debug

<!-- source: openspec/specs/device-ops-ai-debug/spec.md -->

# device-ops-ai-debug Specification

## Purpose
TBD - created by archiving change unify-feeding-chat-ws. Update Purpose after archive.
## Requirements
### Requirement: 运维页支持 App 用户名登录

设备数据运维单设备页（`/device/admin/history/{deviceNo}` 所载 `history.html`）MUST 提供 App 用户名密码登录入口，调用既有 `POST /device/app/api/username_login`，并将返回的 `accessToken` 存于与 Admin JWT **不同**的存储键。Admin JWT 继续用于既有 admin API；tip 与 care-alert 调试请求 MUST 使用 App `accessToken` 作为 Bearer。

#### Scenario: 未 App 登录不可调 tip/care-alert 调试

- **WHEN** 运维未完成 App 用户名登录即点击小贴士生成或值得留意拉取
- **THEN** 页面 MUST 提示需先登录，MUST NOT 使用 Admin JWT 调用上述 App 接口

#### Scenario: Admin 与 App token 分钥

- **WHEN** 运维同时持有 Admin JWT 与 App accessToken
- **THEN** 历史 CRUD MUST 仍带 Admin Bearer；tip/care-alert MUST 带 App Bearer

### Requirement: 运维文字对话调试走 chat WS 文模式

运维页 MUST 提供文字对话调试面板：连接 `/voice/chat/ws`，使用 text 输入与 text 输出模态，展示服务端 `thinking_delta` 与 `answer`。MUST NOT 为此面板新增 gateway 鉴权豁免或 Admin 伪造身份接口。

#### Scenario: 发送文本可见思考与回答

- **WHEN** 运维在已 start（文模式）的连接上发送非空文本
- **THEN** 面板 MUST 展示思考增量与最终 answer（允许忽略音频帧）

### Requirement: 运维小贴士调试走正式 tip SSE

运维页 MUST 通过 `POST /device/tip/generate`（body 含当前 `deviceNo` 及可选 event 字段）发起 SSE，展示 `thinking`/`answer`/`done`（或等价事件）。MUST NOT 新增 tip 的 Admin 旁路或鉴权豁免。重复点击生成即视为强刷（服务端无日缓存门禁时仍走正式额度语义）。

#### Scenario: tip SSE 展示思考过程

- **WHEN** 运维已 App 登录并对当前设备触发小贴士生成
- **THEN** 页面 MUST 以流式方式展示思考与最终建议内容

### Requirement: 运维值得留意调试走正式 care-alert API

运维页 MUST 通过 `GET /device/api/care-alert/daily` 拉取当日列表并展示；MUST 提供强刷操作，请求带 force 语义（见 `care-alert-force-refresh`）。MUST NOT 使用 Admin JWT 或伪造 `X-Internal-Wx-Id` 调用该接口。

#### Scenario: 拉取当日值得留意

- **WHEN** 运维已 App 登录并点击拉取值得留意
- **THEN** 页面 MUST 展示返回的 `day` 与 `items`（可空列表）

#### Scenario: 强刷按钮带 force

- **WHEN** 运维点击值得留意强刷
- **THEN** 页面 MUST 使用带 force 的 daily 请求（仍携带 App Bearer）

---

## device-route-canary-management

<!-- source: openspec/specs/device-route-canary-management/spec.md -->

# device-route-canary-management Specification

## Purpose
TBD - created by archiving change voice-device-canary-route-split. Update Purpose after archive.
## Requirements
### Requirement: Gateway MUST 为 device 路由提供独立可配置代理能力
gateway MUST 以独立中间件管理 `/device/admin/api/*` 路由，并支持 `local|proxy|canary` 三态。

#### Scenario: device 路由进入 local 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=local`
- **THEN** gateway MUST 执行本地处理链路，且 MUST NOT 将请求转发到 device-service

#### Scenario: device 路由进入 proxy 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=proxy` 且 `DEVICE_API_PROXY_URL` 可用
- **THEN** gateway MUST 将 `/device/admin/api/*` 请求全量转发到 device-service

#### Scenario: device 路由进入 canary 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=canary` 且配置了 `DEVICE_API_PROXY_CANARY_PERCENT`
- **THEN** gateway MUST 按稳定分流键执行百分比转发，其余请求保持本地处理

### Requirement: device canary 分流 MUST 保持同键稳定
gateway MUST 采用稳定分流键（如 deviceNo 或请求头标识）对 canary 流量做无状态一致性计算。

#### Scenario: 同一分流键连续请求
- **WHEN** 同一设备在 canary 模式下发起多次 `/device/admin/api/*` 请求
- **THEN** 请求 MUST 稳定命中同一流量路径（proxy 或 local）

---

## device-sim-user

<!-- source: openspec/specs/device-sim-user/spec.md -->

# device-sim-user Specification

## Purpose
TBD - created by archiving change ucg-sim-user-service. Update Purpose after archive.
## Requirements
### Requirement: wx table SHALL store simulated user flag

`device-service` 权威库表 `wx` MUST 新增列 `is_simulated`（TINYINT NOT NULL DEFAULT 0）。`1` 表示模拟用户；公开 App 注册路径 MUST NOT 允许客户端自行设置为 1。

#### Scenario: Default not simulated

- **WHEN** 用户经公开 `username/register` 注册
- **THEN** 新行 `is_simulated` MUST 为 0

### Requirement: device internal sim register SHALL wrap username register with flag

`POST /device/internal/api/sim/username/register` MUST 要求有效 `X-Device-Gateway-Internal-Secret`（或兼容 `X-Gateway-Internal-Secret`）。请求体 MUST 含 `account`、`password`。服务 MUST 调用 `WxUsernameRegister` 并在同一注册流程内将 `is_simulated` 设为 1。成功响应 MUST 含 `wxId`、`account`。账号已存在 MUST 返回业务错误。

调用方（sim-user-service T1）MUST 传入随机生成的 account/password，MUST NOT 依赖固定 `ptest{N}` 或固定默认密码作为长期约定。

#### Scenario: Internal sim register with random account

- **WHEN** sim-service 携带有效密钥 POST `{account:"s8k2m9xq4n",password:"Kp9#mX2vLq4n"}`
- **THEN** 响应 MUST 含 `wxId>0` 且该 wx 行 `is_simulated=1`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim register
- **THEN** HTTP MUST 为 403

### Requirement: device internal SHALL list and batch-query simulated users

device-service MUST 提供：

- `GET /device/internal/api/sim/wx/list` — 分页返回 `is_simulated=1` 的 wxId 与 account 列表（供 Admin、计数 total 等列举场景）
- `GET /device/internal/api/sim/wx/random` — 返回单条随机 simulated 用户（供 sim 任务随机选取）
- `GET /device/internal/api/sim/wx/ids` — 返回**全量** simulated wxId（供 T5 未读抽样等 MUST 覆盖全库 sim 的场景）
- sim 批量查询 MUST 在现有 `POST /device/internal/api/ucg/wx/batch` 响应项中增加 `isSimulated` 字段，或提供等价的 sim 专用 batch 接口

#### Scenario: List sim users

- **WHEN** sim-service 请求 sim wx list 分页
- **THEN** 返回列表 MUST 仅含 `is_simulated=1` 的 wxId

#### Scenario: Batch includes flag

- **WHEN** UCG 或 gateway 批量查询 wxId 展示字段
- **THEN** 每项 MUST 含 `isSimulated` 布尔值

#### Scenario: List not required for random pick

- **WHEN** sim-service 需随机选取一个模拟用户执行任务
- **THEN** MUST 使用 random 接口，MUST NOT 依赖 list 分页结果在客户端 `rand` 选 wxId

#### Scenario: T5 must not use list first page as full sim set

- **WHEN** sim-service T5 需要全库 sim wxId 集合
- **THEN** MUST 使用 ids 接口，MUST NOT 使用 `sim/wx/list?page=1&pageSize=200` 代替全集

### Requirement: device internal SHALL provide random single simulated wx pick

device-service MUST 提供 `GET /device/internal/api/sim/wx/random`，要求有效内部密钥（与 sim wx list 相同）。服务 MUST 在 `is_simulated=1` 集合上通过有界 ID 探测返回 **0 或 1** 条 `{wxId, account}`，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 覆盖全库 simulated 用户（非仅第一页）。锚点 MUST 在 `[minId, maxId]` 上 **均匀** 生成：`R = minId + floor((maxId - minId) * U)`（`U` 为 `(0,1)` 均匀随机），随后 `WHERE is_simulated=1 AND id >= R ORDER BY id ASC LIMIT 1`；锚点落空且 eligible 存在时 MUST 回退 `minId` 一条。MUST NOT 对 high-id / 新注册用户做幂次偏置。

#### Scenario: Random returns one sim user

- **WHEN** sim-service 携带有效密钥 GET random 且存在至少一条 `is_simulated=1`
- **THEN** 响应 MUST 含 `wxId>0` 与非空 `account`

#### Scenario: Random empty when no sim users

- **WHEN** 无 `is_simulated=1` 行
- **THEN** 响应 MUST 表示无结果（空或 found=false），且 MUST NOT 500

#### Scenario: Bounded SQL only

- **WHEN** 代码评审 random 实现
- **THEN** MUST 为 MIN/MAX 聚合 + `LIMIT 1` 探测，MUST NOT 全表加载或 `ORDER BY RAND()`

### Requirement: device internal SHALL provide full simulated wx id list

device-service MUST 提供 `GET /device/internal/api/sim/wx/ids`，要求有效内部密钥（与 sim wx list 相同）。服务 MUST 在单条有界 SQL 内返回 **全部** `is_simulated=1` 的 wxId 列表，响应 `{ ids: int64[], total: int }`。MUST NOT 分页截断（不得仅返回前 200 条）。当 `total` 超过 **10000** 时 MUST 返回 4xx 且 MUST NOT 返回部分 ids。

#### Scenario: Ids returns all simulated users

- **WHEN** 库中存在 350 条 `is_simulated=1` 且请求携带有效密钥
- **THEN** 响应 `total` MUST 为 350 且 `ids` 长度 MUST 为 350

#### Scenario: Ids empty when no sim users

- **WHEN** 无 `is_simulated=1` 行
- **THEN** 响应 MUST 为 `{ ids: [], total: 0 }` 且 MUST NOT 500

#### Scenario: Over limit rejected

- **WHEN** `is_simulated=1` 计数超过 10000
- **THEN** MUST 返回错误且 MUST NOT 返回 ids 数组

### Requirement: device internal SHALL deactivate simulated wx by wxId

`device-service` MUST 提供 `POST /device/internal/api/sim/wx/{wxId}/deactivate`，要求有效 `X-Device-Gateway-Internal-Secret`（与 sim register 相同）。

服务 MUST 校验目标 wx 行存在且 `is_simulated=1`；否则 MUST 返回 4xx 业务错误且 MUST NOT 删除。

校验通过后 MUST 调用 `WxDeactivateByID` 删除 wx 单行，并 MUST 失效 wx 相关 cachekit 键；MUST 从 `usage:sim_wx_ids` SET 移除该 wxId member（经 `cachekit.GatewayUsageSimWxSetKey` / `GatewayUsageSimWxMember`）。

MUST NOT 删除 ucg 或 user 域数据。

#### Scenario: Deactivate sim wx success

- **WHEN** sim-service 携带有效密钥 POST deactivate 且 wxId 为 `is_simulated=1`
- **THEN** wx 行 MUST 删除且 HTTP MUST 成功

#### Scenario: Reject real user deactivate

- **WHEN** wxId 对应 `is_simulated=0`
- **THEN** MUST 返回 4xx 且 wx 行 MUST 保持不变

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim deactivate
- **THEN** HTTP MUST 为 403

---

## device-wx-by-device-no-list

<!-- source: openspec/specs/device-wx-by-device-no-list/spec.md -->

# device-wx-by-device-no-list Specification

## Purpose
TBD - created by archiving change predict-imminent-offline-push. Update Purpose after archive.
## Requirements
### Requirement: device MUST 能按 device_no 列出全部绑定 wxId

device-service MUST 提供内部（或服务间）契约：输入非空 `deviceNo`，返回 `wx` 表中 `device_no` 等于该值的全部用户主键 `wxId` 列表（顺序不限）。MUST NOT 仅返回 `LIMIT 1` 单条作为「全部绑定用户」的语义。实现 MUST 对单次返回数量设上限（配置或常量）；超过上限时 MUST 截断并记录告警日志，或提供分页（一期可选截断+告警）。voice 等调用方 MUST 经 `clients/device` 访问，MUST NOT 直查 device 库。

#### Scenario: 同设备多账号全部返回

- **WHEN** 两个不同 `wx.id` 的行均绑定同一 `device_no`
- **THEN** 列表接口 MUST 返回包含这两个 `wxId` 的结果集

#### Scenario: 无绑定时返回空列表

- **WHEN** 没有任何 `wx` 行绑定该 `deviceNo`
- **THEN** 接口 MUST 成功返回空列表且 MUST NOT 报错为系统故障

---

## device-wx-profile-apis

<!-- source: openspec/specs/device-wx-profile-apis/spec.md -->

# device-wx-profile-apis Specification

## Purpose
TBD - created by archiving change gateway-app-server-wx-auth-history-ws. Update Purpose after archive.
## Requirements
### Requirement: 微信登录仅返回业务字段

device-service SHALL 提供 `POST /device/app/api/user/login`（设备 wx 业务登录，与网关聚合 `POST /device/app/api/login` 区分），接受 **`jsCode`**（微信开放平台授权临时 `code`）与 **`platform`**（与 device 配置平台键一致）。

系统 SHALL 使用服务端持有的微信凭据换取 `unionid` 并按 `unionid` 查找或创建 `wx` 行。对于微信登录路径，若微信响应中 `unionid` 为空，系统 SHALL 返回明确业务错误且 SHALL NOT 创建或匹配用户行。系统同时 SHALL 支持同表中的用户名账号记录（其 `unionid` MAY 为空），且 SHALL NOT 因存在 `unionid` 为空的用户名记录影响微信登录判定。

响应 SHALL 包含至少 `wxId`、`isNewUser`、已绑定时的 `deviceNo`；响应 SHALL NOT 包含 gateway 签发令牌，也 SHALL NOT 返回 `unionid`、`openid`、微信令牌明文。

#### Scenario: 新微信用户登录成功
- **WHEN** 首次出现的 `unionid` 调用登录接口
- **THEN** 系统 SHALL 创建 wx 行并返回 `isNewUser=true`

#### Scenario: 既有微信用户登录成功
- **WHEN** `unionid` 已存在于 `wx` 表
- **THEN** 系统 SHALL 返回已有 `wxId` 与已绑定 `deviceNo`（若有）

### Requirement: 绑定设备与 wx

device-service SHALL 提供 `POST /device/app/api/user/bindwx`，从请求头读取 **`X-Internal-Wx-Id`**（值为 `wx.id`，由 gateway 从 access token `sub` 注入），从 JSON body 读取 `deviceNo`，并将设备号绑定到对应 `wx` 行。

#### Scenario: 绑定成功
- **WHEN** 头部包含有效 `X-Internal-Wx-Id` 且 `deviceNo` 合法并已注册
- **THEN** 系统 SHALL 持久化绑定关系并返回成功语义

#### Scenario: 头部无效
- **WHEN** 缺失或提供非法 `X-Internal-Wx-Id`
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 更新绑定关系

### Requirement: 自动保存画像

device-service SHALL 提供 `POST /device/app/api/user/auto_save`，从请求头读取 **`X-Internal-Wx-Id`**，从 body 读取 `birthday` 与 `sex`，并 SHALL 返回 `device_no`。当目标 `wx` 行尚未绑定设备时，系统 SHALL 生成全局唯一、6 位大写字母 `device_no`，完成设备注册与绑定后写入画像；当已绑定时，系统 SHALL 仅更新画像并返回原 `device_no`。

#### Scenario: 无设备号时创建并绑定
- **WHEN** `wxId` 有效且当前 wx 行未绑定 `device_no`
- **THEN** 系统 SHALL 生成并绑定唯一 `device_no`，保存画像后返回该值

#### Scenario: 已绑定设备仅更新画像
- **WHEN** `wxId` 有效且 wx 已绑定 `device_no`
- **THEN** 系统 SHALL 仅更新画像并返回原 `device_no`

#### Scenario: 候选设备号冲突重试
- **WHEN** 随机候选 `device_no` 与现有数据冲突
- **THEN** 系统 SHALL 重试生成直到成功或达到最大重试上限

### Requirement: 按 unionid 查询设备号

device-service SHALL 提供 `GET /device/app/api/user/detail`，并以 **`X-Internal-Wx-Id`** 识别当前账号主体，返回该主体绑定的 `device_no`（未绑定时返回约定空值或错误语义）。

#### Scenario: 已绑定返回设备号
- **WHEN** `X-Internal-Wx-Id` 对应记录已绑定 `device_no`
- **THEN** 响应 SHALL 包含该 `device_no`

#### Scenario: 未绑定返回空语义
- **WHEN** `X-Internal-Wx-Id` 对应记录未绑定设备
- **THEN** 响应 SHALL 返回空 `device_no` 或约定未绑定语义

### Requirement: 按主键 id 解析 unionid（内部）

device-service SHALL 提供仅供内网或网关调用的只读接口（例如 `GET /device/app/api/user/internal/by-id`），根据 wx 表主键 id 返回对应 **`union_id`（响应字段 unionId）**，以便 gateway-app 在仅持有 access 内 id 时解析 **unionid** 并写入 **`X-Internal-Wx-Union-Id`**；该接口 SHALL 不对外网匿名开放（依赖部署网络或额外共享密钥策略）。

#### Scenario: 有效 id

- **WHEN** 网关使用有效 id 调用内部解析接口
- **THEN** 响应 SHALL 包含与该 id 对应的 unionId

#### Scenario: 无效 id

- **WHEN** id 不存在或非法
- **THEN** 系统 SHALL 返回明确错误且 SHALL NOT 泄露其他行信息

### Requirement: Redis 缓存与失效

device-service 对高频读路径（含 `wxId -> unionid`、`wxId -> deviceNo`）SHALL 可选使用 Redis 缓存；在绑定设备、注销、或任何影响映射关系的写操作成功后，系统 SHALL 失效相关缓存键，确保后续读取一致。

#### Scenario: 写后缓存一致性
- **WHEN** bindwx、auto_save 或 deactivate 成功完成
- **THEN** 与该 `wxId` 相关缓存 SHALL 被删除或失效

### Requirement: 设备画像读接口 SHALL 返回宝宝名字
系统在读取设备画像时 MUST 同时返回 `babyName`、`birthday`、`sex` 三个字段；其中 `babyName` 为可选字符串，未设置时返回空串。
该要求适用于 device 画像接口与历史页面画像接口的统一读取语义。

#### Scenario: 读取画像返回完整字段
- **WHEN** 调用方使用有效 `deviceNo` 请求画像读取接口
- **THEN** 响应 SHALL 包含 `babyName`、`birthday`、`sex`
- **AND** 当数据库中 `baby_name` 为空时，`babyName` SHALL 返回空串

### Requirement: 设备画像写接口 SHALL 支持宝宝名字更新
系统在保存设备画像时 MUST 接受 `babyName` 字段，并与 `birthday`、`sex` 一并持久化到 `user` 表画像字段集合。
该要求适用于 `/device/app/api/user/save`、`/device/app/api/user/auto_save` 以及历史页面画像保存链路。

#### Scenario: 仅修改宝宝名字
- **WHEN** 调用方提交合法 `deviceNo` 与 `babyName`，且未变更生日/性别
- **THEN** 系统 SHALL 更新 `user.baby_name`
- **AND** 系统 SHALL 保持 `birthday`、`sex` 原值不变

#### Scenario: 同时修改名字与性别生日
- **WHEN** 调用方提交 `babyName`、`birthday`、`sex`
- **THEN** 系统 SHALL 在一次保存语义内持久化三项画像字段

### Requirement: 账号注销删除 wx 记录
系统 MUST 提供 `POST /device/app/api/user/deactivate`。接口 SHALL 从请求头读取 `X-Internal-Wx-Id`，并按该主键删除 `wx` 表中的对应单条记录。删除成功后，系统 SHALL 使该 `wxId` 相关缓存映射失效，避免后续读取命中陈旧数据。

#### Scenario: 注销成功删除单条记录
- **WHEN** 请求头包含有效的 `X-Internal-Wx-Id` 且该 `wx` 记录存在
- **THEN** 系统 SHALL 删除该主键对应的一条 `wx` 记录并返回成功语义

#### Scenario: 请求头缺失或无效
- **WHEN** `X-Internal-Wx-Id` 缺失、非整数或小于等于 0
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 执行删除

#### Scenario: 目标记录不存在
- **WHEN** `X-Internal-Wx-Id` 合法但对应 `wx` 记录不存在
- **THEN** 系统 SHALL 返回明确业务错误语义（已注销或不存在），且 SHALL NOT 影响其他记录

### Requirement: 查询当前账号 profile

device-service SHALL 提供 `GET /device/app/api/user/profile`。接口 SHALL NOT 要求额外 query 或 body 入参；SHALL 从请求头读取 **`X-Internal-Wx-Id`**（值为 `wx.id`，由 gateway 从 access token `sub` 注入）定位当前 `wx` 行，并返回账号状态字段。

响应 SHALL 包含：
- **`isWxBound`**（bool，始终返回）：当且仅当该行 `unionid` 非空时为 `true`；
- **`account`**（string）：该行用户名账号；当账号为空时，响应 SHALL 省略该字段（JSON `omitempty`）；
- **`deviceNo`**（string，始终返回）：该行已绑定设备号；未绑定时 SHALL 返回空字符串。

响应 SHALL NOT 包含 `unionid`、`password`、`openid` 或微信令牌明文。

#### Scenario: 纯微信用户已绑设备
- **WHEN** `X-Internal-Wx-Id` 有效且对应 `wx` 行 `unionid` 非空、`account` 为空、已绑定 `device_no`
- **THEN** 响应 SHALL 包含 `isWxBound=true`、`deviceNo` 为已绑定值，且 SHALL NOT 包含 `account` 字段

#### Scenario: 纯用户名用户未绑微信
- **WHEN** `X-Internal-Wx-Id` 有效且对应 `wx` 行 `account` 非空、`unionid` 为空
- **THEN** 响应 SHALL 包含 `isWxBound=false`、`account` 为对应用户名，以及 `deviceNo`（未绑设备时为 `""`）

#### Scenario: 用户名与微信均已绑定
- **WHEN** `X-Internal-Wx-Id` 有效且对应 `wx` 行 `unionid` 与 `account` 均非空
- **THEN** 响应 SHALL 包含 `isWxBound=true`、`account` 与 `deviceNo`

#### Scenario: 请求头缺失或无效
- **WHEN** `X-Internal-Wx-Id` 缺失、非整数或小于等于 0
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 返回 profile 数据

#### Scenario: wx 记录不存在
- **WHEN** `X-Internal-Wx-Id` 合法但对应 `wx` 记录不存在
- **THEN** 系统 SHALL 返回明确错误语义（如 404），且 SHALL NOT 泄露其他行信息

---

## docker-deploy-logging

<!-- source: openspec/specs/docker-deploy-logging/spec.md -->

# docker-deploy-logging Specification

## Purpose
TBD - created by archiving change docker-container-log-limits. Update Purpose after archive.
## Requirements
### Requirement: 生产与测试 Compose 栈 MUST 限制容器 json-file 日志保留量

`manifest/docker` 下用于 **生产** 与 **测试** 长期运行的 Compose 服务（微服务基线六件套、RabbitMQ prod/test、Redis prod cluster / test standalone）MUST 配置 `logging.driver=json-file`，且 MUST 设置 `max-size` 与 `max-file` 轮转选项。微服务与 Redis 默认 MUST 为 `max-size=10m`、`max-file=3`；RabbitMQ MUST 为 `max-size=20m`、`max-file=3`。策略 MUST 在 prod 与 test 对齐（同一 compose 源或等价 anchor）。

#### Scenario: 微服务容器日志有上限

- **WHEN** 运维在 ECS 上对测试或生产微服务栈执行 `docker compose up -d --force-recreate`
- **THEN** 各微服务容器 inspect 的 LogConfig MUST 含 `max-size` 10m 与 `max-file` 3

#### Scenario: RabbitMQ 容器日志有上限

- **WHEN** 运维 recreate 生产或测试 RabbitMQ compose 栈
- **THEN** Rabbit 容器 LogConfig MUST 含 `max-size` 20m 与 `max-file` 3

### Requirement: RabbitMQ MUST 降低 stdout 日志级别至 warning

生产与测试 RabbitMQ compose MUST 挂载仓库内 `manifest/docker/rabbitmq/rabbitmq.conf`，且该文件 MUST 将 `log.console.level`、`log.connection.level`、`log.channel.level` 设为 `warning`。Management 插件与 AMQP/HTTP 业务行为 MUST NOT 因该配置而关闭。

#### Scenario: Rabbit 正常收发时 stdout 更安静

- **WHEN** 客户端 HTTP Publish 与 AMQP consume 正常运行
- **THEN** Rabbit 容器 `docker logs` MUST NOT 以 info 级别刷屏连接/channel 常规行（warning 及以上仍可见）

### Requirement: runbook MUST 说明日志策略生效与磁盘清理

`docs/runbooks/release-deploy-and-run.md` MUST 说明：logging 变更需 recreate 容器；已有巨型 `*-json.log` 的清理方式（删容器或 truncate）；可用 `docker system df -v` 验收。

#### Scenario: 运维按 runbook 部署后验收

- **WHEN** 运维完成 logging 变更的 recreate
- **THEN** runbook MUST 提供检查 LogConfig 与磁盘占用的命令示例

---

## documentation-language-compliance

<!-- source: openspec/specs/documentation-language-compliance/spec.md -->

# documentation-language-compliance Specification

## Purpose
TBD - created by archiving change enforce-chinese-documentation. Update Purpose after archive.
## Requirements
### Requirement: 变更文档需要通过语言合规检查
系统在变更进入实施阶段前 SHALL 完成文档语言合规检查，若说明性文本以英文为主则不得作为可实施输入。

#### Scenario: 变更进入 apply 前校验
- **WHEN** 变更已生成 proposal、design、specs、tasks 并准备进入实施阶段
- **THEN** 文档语言合规检查 SHALL 确认说明性文本为中文，否则应阻止进入实施

### Requirement: 语言规则适用于增量更新
系统在对已有变更进行增量更新时 SHALL 继续遵循中文文档规则，不因历史内容存在英文而豁免。

#### Scenario: 更新已有变更工件
- **WHEN** 用户对现有变更工件进行追加或修订
- **THEN** 新增与修改内容 SHALL 使用中文说明并保持术语一致

---

## domain-package-boundary-enforcement

<!-- source: openspec/specs/domain-package-boundary-enforcement/spec.md -->

# domain-package-boundary-enforcement Specification

## Purpose
TBD - created by archiving change migrate-service-to-services-full-cutover. Update Purpose after archive.
## Requirements
### Requirement: 领域包边界 MUST 与服务边界一致
迁移后代码 MUST 按领域归属放置在 `internal/services/voice`、`internal/services/device`、`internal/services/history` 等目录，且包语义必须与目录一致。

#### Scenario: 包语义审查
- **WHEN** 审查迁移后的领域目录
- **THEN** 目录内代码包语义 MUST 体现对应领域职责，不得继续使用统一 `service` 包承载多域逻辑

### Requirement: 共享目录准入 MUST 可审计
`internal/shared` MUST 仅容纳无领域语义的通用能力；含领域流程或领域模型耦合的实现 MUST 禁止进入共享目录。

#### Scenario: 共享目录准入检查
- **WHEN** 有文件计划迁入 `internal/shared`
- **THEN** 评审 MUST 给出“无领域语义”依据，否则该文件 MUST 回到对应领域目录

### Requirement: 新增代码 MUST 禁止回流到 `internal/service`
迁移完成后，新增实现文件 MUST 不得再放入 `internal/service`。

#### Scenario: 新增文件路径检查
- **WHEN** 提交包含新增实现文件
- **THEN** 若目标路径为 `internal/service`，该提交 MUST 视为不符合边界规范

---

## end-open-history-by-event

<!-- source: openspec/specs/end-open-history-by-event/spec.md -->

# end-open-history-by-event Specification

## Purpose
TBD - created by archiving change end-open-history-by-event. Update Purpose after archive.
## Requirements
### Requirement: EndLatest 按 eventId 闭合最近未结束记录

`history-service` 的 `EndLatestHistoryIfMatch`（及内部 HTTP `POST /device/history/api/event/end-latest`）MUST 在指定 `deviceNo` 与 `eventId` 下，查找 `end_time=0`（未闭合）的历史行，按 `id` 降序取最近一条并更新其结束时间；MUST NOT 仅因「全局最新一条 history 的 eventId 不等于目标」而返回未匹配。当 remark 非空时 MUST 同时覆盖该行备注；remark 为空串时 MUST NOT 修改原备注。

#### Scenario: 中间夹其它事件仍能结束睡眠

- **WHEN** 设备已有未闭合睡眠记录（`eventId=E` 且 `end_time=0`），其后又写入其它事件记录，使全局最新一条不再是睡眠
- **AND** 调用方对同一 `deviceNo` 请求结束 `eventId=E`（语音 `feeding+end` 或 App `end-latest`）
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=true`
- **AND** SHALL 更新该未闭合睡眠行的 `end_time`（而非新建一条睡眠历史）

#### Scenario: 无未闭合同 event 时返回未匹配

- **WHEN** 该 `deviceNo` 下不存在 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方请求结束 `eventId=E`
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=false`
- **AND** SHALL NOT 修改任何已结束（`end_time!=0`）的同 `eventId` 行

#### Scenario: 多条未闭合同 event 只闭合最近一条

- **WHEN** 同一 `deviceNo` 下存在多条 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方请求结束 `eventId=E`
- **THEN** 系统 SHALL 仅更新其中 `id` 最大的一条的 `end_time`
- **AND** 其余未闭合行保持 `end_time=0`

### Requirement: 语音 end 优先闭合未结束记录再降级新建

喂养事件 CRUD（含 start/end/one/multi）MUST 由 `python-ai-talk` 在意图落地阶段经 `POST /device/history/api/event/batch` 写入 history；Go `voice-service` MUST NOT 在 `applyUnifiedIntentResult` 或等价主路径内调用 `DeviceHistory().AddHistory` / `UpdateHistory` 执行喂养写库。当 Python 处理 `action=end` 且目标 event 为 E 时，batch MUST 优先使用 `op=end` 或等价语义经 `EndLatestHistoryIfMatch` 闭合；仅当无未闭合同 E 记录时 MAY 降级 `create` 瞬时结束行（`start_time=end_time`）。MUST NOT 在仍存在未闭合 E 记录时新建一条 E 的历史行作为「结束」手段。

#### Scenario: 孩子醒了结束已开始的睡眠

- **WHEN** Python 返回结束睡眠意图且该设备存在未闭合睡眠历史（同 eventId）
- **THEN** Python batch MUST 闭合该未闭合睡眠行（经 end-latest 或 batch op=end）
- **AND** MUST NOT 再新增一条睡眠历史行作为结束

#### Scenario: 从未开始睡眠时 end 可降级新建

- **WHEN** Python 返回结束睡眠意图，且该设备不存在未闭合的同 eventId 历史
- **THEN** batch MAY 写入瞬时结束记录
- **AND** Go voice MAY 透传 Python 成功话术

#### Scenario: 重复 start 被 history 拒绝

- **WHEN** Python batch 提交同 eventId 的 start（`endTime=0`）且该 event 已在进行中
- **THEN** history-service MUST 拒绝该 create
- **AND** batch 对应 item MUST 返回失败 reason

### Requirement: 禁止同 eventId 重复 start 进行中记录

`history-service` 在 `AddDeviceHistory`（及经其的 `POST /device/history/api/event/add`、`POST /device/history/api/event/batch` 之 `op=create`）插入新行时，若请求满足 `eventId>0` 且 `endTime=0`（进行中），MUST 先查询该 `deviceNo` 下是否已存在同 `eventId` 且 `end_time=0` 的历史行；若存在 MUST 拒绝插入并返回可读错误（语义等价：`该事件已在进行中，请先结束后再开始`）。`UpdateDeviceHistory`（及 `event/update`、batch `op=update`）将某行 `endTime` 改为 `0` 时 MUST 应用相同规则（排除当前行 `id`）。本守卫 MUST NOT 限制不同 `eventId` 的并行进行中； MUST NOT 拦截 `endTime!=0` 的瞬时记录（含 `startTime=endTime` 的 one 类记录）。

#### Scenario: 同 event 已有进行中时拒绝再次 start

- **WHEN** 设备 `deviceNo=D` 已存在 `eventId=E` 且 `end_time=0` 的历史行
- **AND** 调用方经 add 或 batch create 请求插入 `eventId=E`、`endTime=0` 的新行
- **THEN** history-service MUST 拒绝该插入
- **AND** MUST NOT 新增第二条同 E 的未闭合行

#### Scenario: 不同 event 允许并行进行中

- **WHEN** 设备已有 `eventId=E1` 且 `end_time=0` 的睡眠记录
- **AND** 调用方插入 `eventId=E2`（E2≠E1）且 `endTime=0` 的计时记录
- **THEN** history-service MUST 允许插入

#### Scenario: 瞬时 one 记录不受限

- **WHEN** 设备已有 `eventId=E` 且 `end_time=0` 的进行中行
- **AND** 调用方插入同 E 但 `endTime>0` 且 `endTime>=startTime` 的瞬时记录
- **THEN** history-service MUST 允许插入

#### Scenario: batch 单条失败不整单回滚

- **WHEN** Python 经 `event/batch` 提交多条 items，其中一条因重复 start 失败
- **THEN** 该 item MUST 在 results 中 `ok=false` 且 `reason` 含可读拒绝语义
- **AND** 其它 items MUST 仍按既有 batch 语义独立执行

---

## enum-adapter-compatibility

<!-- source: openspec/specs/enum-adapter-compatibility/spec.md -->

# enum-adapter-compatibility Specification

## Purpose
定义字符串到枚举迁移期兼容策略，保证旧入口可用并可验证关键路径完成枚举化收敛。

## Requirements
### Requirement: 渐进迁移兼容层
系统 MUST 在迁移期间保留字符串入口的兼容适配层，并通过统一适配函数将旧字符串路径映射到新枚举实现。

#### Scenario: 旧入口继续可用
- **WHEN** 调用方仍传入历史字符串值
- **THEN** 系统 MUST 通过兼容适配层完成转换并保持行为一致

#### Scenario: 兼容层输出弃用提示
- **WHEN** 旧入口被调用
- **THEN** 系统 SHOULD 输出弃用告警日志，提示迁移到枚举入口

### Requirement: 枚举化迁移可验证
系统 SHALL 提供可验证迁移清单，确保关键模块不再新增裸字符串匹配。

#### Scenario: 核心模块迁移完成检查
- **WHEN** 执行迁移验收
- **THEN** 系统 MUST 能确认 outbox、consumer、voice 关键路径已使用枚举匹配

---

## event-logo-oss-cdn

<!-- source: openspec/specs/event-logo-oss-cdn/spec.md -->

# event-logo-oss-cdn Specification

## Purpose
TBD - created by archiving change event-logo-oss-cdn. Update Purpose after archive.
## Requirements
### Requirement: 事件 logo SHALL 存储于 OSS event 前缀

device-service 在管理端上传或迁移脚本写入时，MUST 将 `event.logo` 设为 OSS objectKey，前缀 MUST 为 `event/`，MUST NOT 含 scheme 或域名。

#### Scenario: 管理端上传新 logo

- **WHEN** 客户端 multipart 提交合法图片至 event add/update
- **THEN** device-service MUST 经 HTTP 调用 ucg-service 内部上传接口
- **AND** MUST 将返回的 objectKey 写入 `event.logo`

#### Scenario: 迁移脚本回填

- **WHEN** 迁移脚本成功上传历史 logo
- **THEN** `event.logo` MUST 为 `event/{id}/logo.{ext}` 形式

### Requirement: 对外 API logo 字段 SHALL 返回 CDN 绝对 URL

所有返回事件字典列表的 HTTP 接口（含 admin list、history/internal event options、gateway site home 的 logoUrl）MUST 将 logo 序列化为 `https://{cdnHost}/{objectKey}`，MUST NOT 返回 `/ai_talk_images/` path-only。

#### Scenario: history event options 返回 CDN

- **WHEN** 客户端请求 `GET /device/history/api/event/options`
- **THEN** 每项 `logo` 若有值 MUST 以 `https://` 开头且指向 CDN 域名

### Requirement: ucg-service SHALL 提供内部 OSS 上传供 device 调用

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload`，接受 multipart 文件，鉴权 MUST 使用网关内部密钥；响应 MUST 含 `objectKey` 与 `cdnUrl`。

#### Scenario: device 内部上传成功

- **WHEN** device-service 携带有效内部密钥上传 png
- **THEN** ucg-service MUST 写入 OSS 并返回 objectKey 与 cdnUrl

### Requirement: 迁移完成后 SHALL 下线 ai_talk_images 静态链路

系统 MUST NOT 再注册 `GET /ai_talk_images/*` 静态读或网关反代；Docker compose MUST NOT 再挂载宿主机 `/ai_talk_images` 目录。

#### Scenario: 旧静态 URL 不可用

- **WHEN** 迁移与部署完成后请求 `GET /ai_talk_images/event_1.png`
- **THEN** 网关或 device-service MUST NOT 再提供该静态文件

---

## feature-activation-subject

<!-- source: openspec/specs/feature-activation-subject/spec.md -->

# feature-activation-subject Specification

## Purpose
TBD - created by archiving change growth-trajectory-user-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 功能定义 MUST 声明开通主体 device 或 user

cash-service MUST 在 `feature_def` 上持久化 `activation_subject`，取值仅为 `device` 或 `user`，缺省 MUST 为 `device`。EnsureSchema MUST 幂等加列并对已有行回填 `device`。种子 `growth_trajectory_predict` MUST 为 `user`；`care_alert_smart_remind` 与 `prediction_unlock` MUST 为 `device`。

#### Scenario: 成长轨迹种子为人

- **WHEN** cash EnsureSchema 首次或幂等执行且写入/确认成长轨迹功能定义
- **THEN** 该功能 `activation_subject` MUST 为 `user`

#### Scenario: 值得留意仍为机

- **WHEN** 读取 `care_alert_smart_remind` 功能定义
- **THEN** `activation_subject` MUST 为 `device`

### Requirement: ActivateFeature MUST 按主体写入对应权益表

系统 MUST 支持 `ActivationSubjectDevice` 与 `ActivationSubjectUser`。当主体为 `device` 时 MUST 写入（或续期）`feature_entitlement`（按 `device_no`）。当主体为 `user` 时 MUST 写入（或续期）`feature_user_entitlement`（按 `wx_id`），MUST NOT 为该次履约写入成长轨迹的设备权益行。`prediction_unlock` 与 `allowed_count_delta` MUST 仅允许 device 主体。

#### Scenario: 用户主体支付履约

- **WHEN** 成长轨迹付费订单支付成功且功能定义为 `user`
- **THEN** 系统 MUST 以订单 `wx_id` 为键续期 `feature_user_entitlement`，时长取 SKU `duration_days`（种子 30）

#### Scenario: 拒绝预测写到账号维

- **WHEN** 调用 Activate 且 feature 为预测条数类且 SubjectType=user
- **THEN** 系统 MUST 拒绝且 MUST NOT 修改 `feature_allowed_count`

### Requirement: 支付邀请广告 MUST 按 activation_subject 选择 SubjectKey

支付履约、邀请码兑换、广告完成在授予非预测功能时 MUST 读取功能定义的 `activation_subject`：为 `user` 时 SubjectKey MUST 为触发账号 `wx_id`；为 `device` 时 MUST 为请求/订单中的 `device_no`。广告完成在主体为 `user` 时 MUST 要求有效 `wxId`，否则 MUST 拒绝。

#### Scenario: 邀请开通成长轨迹写人

- **WHEN** 用户兑码开通 `growth_trajectory_predict` 成功
- **THEN** 权益 MUST 落在兑换人 `wx_id` 的 user 权益表，且邀请天数取自 `feature_def.invite_duration_days`

### Requirement: 目录合成 MUST 按主体合并开通态

`GetFeatureCatalog`（或等价 App 目录）MUST 接受设备号，并在调用方提供登录 `wxId` 时：对 `activation_subject=device` 的功能用设备权益判断 `unlocked`；对 `user` 用该 `wxId` 的 user 权益判断。未提供有效 `wxId` 时，user 类功能 MUST 视为未解锁（`unlocked=false`），MUST NOT 误用设备权益冒充。

#### Scenario: 同设备另一账号目录未解锁

- **WHEN** 设备 D 上账号 A 已开通成长轨迹，账号 B（非 VIP）请求目录且绑定 D
- **THEN** 成长轨迹项对 B MUST 为未解锁

---

## feature-activation-toolkit

<!-- source: openspec/specs/feature-activation-toolkit/spec.md -->

# feature-activation-toolkit Specification

## Purpose
TBD - created by archiving change feature-activation-care-alert. Update Purpose after archive.
## Requirements
### Requirement: 系统 MUST 提供按主体与通道入参驱动的功能开通原子入口

cash-service MUST 提供统一开通入口（名称以实现为准），调用方 MUST 传入至少：`featureId`、激活主体类型（`device` 或 `user`）、主体键、开通通道（`payment`、`invite_code` 或 `ad`）、通道引用与操作者 `wxId`（邀请场景）。支付回调、邀请码兑换成功路径、广告完成开通 MUST 经该入口写入权益或数量，MUST NOT 旁路直接 upsert 表。一期对主体类型 `user` MUST 拒绝或返回明确错误（本变更不落地用户维权益表）。预测与值得留意等设备共享功能 MUST 使用主体 `device`。

#### Scenario: 支付汇入原子入口

- **WHEN** 功能订单支付履约成功
- **THEN** 系统 MUST 调用原子入口且 Channel 为 `payment`，按 SKU 授予效果写入对应设备权益或条数

#### Scenario: 邀请汇入原子入口

- **WHEN** 邀请码兑换某功能成功校验通过后授予
- **THEN** 系统 MUST 调用原子入口且 Channel 为 `invite_code`，MUST NOT 在兑换函数内绕过入口直接写 `feature_entitlement` / `feature_allowed_count`

#### Scenario: 广告汇入原子入口

- **WHEN** 广告完成开通申报通过校验
- **THEN** 系统 MUST 调用原子入口且 Channel 为 `ad`

#### Scenario: user 主体一期拒绝

- **WHEN** 调用方以 SubjectType=`user` 请求开通
- **THEN** 系统 MUST 拒绝，MUST NOT 写入设备权益行冒充成功

### Requirement: 邀请码与广告的授予效果 MUST 同源解析

对同一 `featureId`，通道 `invite_code` 与 `ad` MUST 使用同一套效果解析规则。对权益型功能，授予天数 MUST 取自该功能定义上的邀请/广告授予天数配置（`feature_def.duration_days` 或等价字段）：`0` 表示永久，`>0` 表示自授予时刻起对应自然日秒数。对预测数量类功能，邀请与广告 MUST 继续按永久条数增量语义授予，MUST NOT 因该天数配置改变条数语义。支付通道 MUST 继续以商品 SKU 的 `duration_days` / `grant_kind` 为准，MUST NOT 强制与邀请天数相同。

#### Scenario: 权益型邀请与广告同天数

- **WHEN** 某权益型功能配置邀请/广告授予天数为 7，用户分别用邀请码与广告开通该功能
- **THEN** 两次授予的到期语义 MUST 同为约 7 天（续期规则与既有 entitlement upsert 一致）

#### Scenario: 预测邀请广告仍为条数

- **WHEN** 用户对 `prediction_unlock` 完成邀请或广告开通
- **THEN** 系统 MUST 永久 `allowed_count` +1（或等价增量），MUST NOT 仅因定义天数改为限时全开

### Requirement: Admin MUST 可配置功能的邀请/广告授予天数

开通功能管理中的功能定义编辑 MUST 提供「邀请/广告授予天数」表单字段并持久化到功能定义。`0` MUST 表示永久。该字段 MUST NOT 被解释为支付套餐时长（支付仍以 SKU 为准）。预测类功能可展示该字段但运行时条数授予 MUST 忽略之。

#### Scenario: 运维修改邀请天数后兑码

- **WHEN** 管理员将某权益型功能的邀请/广告天数从 7 改为 3 并保存，之后用户首次邀请开通该功能
- **THEN** 新授予 MUST 按 3 天计算到期

#### Scenario: 支付套餐独立

- **WHEN** 功能定义邀请天数为 7 且某支付 SKU `duration_days=0`
- **THEN** 支付开通 MUST 为永久，MUST NOT 变成 7 天

---

## feature-admin-activation-snapshot

<!-- source: openspec/specs/feature-admin-activation-snapshot/spec.md -->

# feature-admin-activation-snapshot Specification

## Purpose
TBD - created by archiving change feature-ops-invite-force-admin. Update Purpose after archive.
## Requirements
### Requirement: Admin 按功能查看当前开通快照

cash Admin MUST 提供按 `featureId` 查询的开通快照接口（及管理页入口），返回该功能当前持久化开通主体列表（可分页），每项至少包含：主体标识（`deviceNo` 或 `wxId`）、最近开通方式（若适用）、是否未过期有效、到期时间或剩余时效（`expires_at=0` 表示永久；预测条数类可用条数代替时效）。快照 MUST 来自当前权益/条数权威表（方案 A），MUST NOT 要求新建事件账本。页面 MUST 注明不含 VIP 旁路覆盖、非完整历史事件流。

#### Scenario: 成长轨迹列出账号权益

- **WHEN** 运维打开 `growth_trajectory_predict` 详情快照
- **THEN** 列表 MUST 来自 `feature_user_entitlement`（或等价），含 wx、方式、是否有效与到期/永久

#### Scenario: 值得留意列出设备权益

- **WHEN** 运维打开 `care_alert_smart_remind` 详情快照
- **THEN** 列表 MUST 来自 `feature_entitlement`，含 deviceNo、方式、是否有效与到期/永久

#### Scenario: 预测列出条数

- **WHEN** 运维打开 `prediction_unlock` 详情快照
- **THEN** 列表 MUST 展示设备与永久条数（或等价字段），MUST NOT 伪造权益型到期字段为必需

#### Scenario: 已过期仍可见

- **WHEN** 某主体权益 `expires_at` 已早于当前时间且行仍存在
- **THEN** 快照 MUST 仍可列出并标记为未在有效期内

---

## feature-admin-activation-subject

<!-- source: openspec/specs/feature-admin-activation-subject/spec.md -->

# feature-admin-activation-subject Specification

## Purpose
TBD - created by archiving change growth-trajectory-user-entitlement. Update Purpose after archive.
## Requirements
### Requirement: Admin 功能定义 API MUST 暴露开通主体

`GET /cash/admin/api/feature/defs` 返回的每一项 MUST 包含 `activationSubject`（`device`|`user`）。`POST /cash/admin/api/feature/defs` MUST 接受可选 `activationSubject`；若提供则 MUST 校验取值并持久化。将 `prediction_unlock`（或其它条数类）更新为 `user` MUST 被拒绝。未传该字段时 MUST 保持库中原值不变。

#### Scenario: 列表返回人对机

- **WHEN** 管理员拉取功能定义列表
- **THEN** 每条 MUST 含 `activationSubject`，成长轨迹为 `user`，值得留意为 `device`

#### Scenario: 禁止预测改为人

- **WHEN** Admin 将预测功能 `activationSubject` 设为 `user` 并保存
- **THEN** 系统 MUST 拒绝且 MUST NOT 修改该行主体

### Requirement: 开通功能管理页 MUST 标明对人或对机

`resource/public/cash-feature-admin.html`（经 gateway 运维 Hub 入口）MUST 在功能定义列表中展示每条功能的开放主体，文案 MUST 区分「对人」与「对机」（或等价清晰中文），MUST NOT 仅依赖英文枚举让运维猜测。编辑区 MUST 展示当前主体；若允许修改，MUST 提供明确控件并在保存后与列表一致。

#### Scenario: 列表一眼可辨成长轨迹对人

- **WHEN** 运维打开开通功能管理页并加载定义列表
- **THEN** `growth_trajectory_predict` 行 MUST 显示「对人」（或含「人/账号」语义的同等文案）

#### Scenario: 值得留意显示对机

- **WHEN** 列表展示 `care_alert_smart_remind`
- **THEN** 该行 MUST 显示「对机」（或含「设备」语义的同等文案）

---

## feature-admin-api

<!-- source: openspec/specs/feature-admin-api/spec.md -->

# feature-admin-api Specification

## Purpose
TBD - created by archiving change commercial-feature-entitlement. Update Purpose after archive.
## Requirements
### Requirement: cash-service MUST 提供付费功能管理 Admin CRUD API

cash-service MUST 在 `/cash/admin/api/feature/**`（或等价前缀）提供管理端 API，支持功能定义（`feature_def`）、解锁方式、功能 SKU（`feature_product`）、预测数量授予相关配置的创建/读取/更新/停用。写操作 MUST 校验 Admin 鉴权（与现网 cash Admin 一致）。Admin API MUST NOT 改变 App usage 统计 denylist 的默认假设（管理端前缀）。

#### Scenario: 创建功能定义

- **WHEN** 管理员提交合法的新功能定义
- **THEN** 系统 MUST 持久化到 `ai_voice_cash` 并在后续目录/管理列表中可见

#### Scenario: 配置功能 SKU

- **WHEN** 管理员创建或更新功能 `product_code` 与价格/授予语义
- **THEN** 系统 MUST 持久化，且支付建单 MUST 能引用该 SKU

#### Scenario: 未鉴权拒绝

- **WHEN** 请求未携带有效 Admin 凭证
- **THEN** 系统 MUST 拒绝写操作

### Requirement: cash-service MUST 提供邀请码管理 Admin API 含兑换明细

系统 MUST 提供邀请码 Admin API（建议 `/cash/admin/api/invite-code/**`）：支持创建/读取/更新/停用邀请码；创建时 MUST 绑定 `owner_wx_id`；一期 MUST 强制一 `owner_wx_id` 仅一条有效码；MUST 可配置有效期与可开通功能集合。MUST 提供按码查询兑换明细：至少含 `redeemer_wx_id`、`device_no`、`featureId`、`redeemed_at`，供追踪销售推广。

#### Scenario: 创建邀请码绑定 owner

- **WHEN** 管理员为某 `owner_wx_id` 创建邀请码并指定有效期与可开功能
- **THEN** 系统 MUST 保存码记录；若该 owner 已有码则 MUST 拒绝或按产品约定停用旧码（一期一人一码）

#### Scenario: 查看某码被哪些人使用

- **WHEN** 管理员查询某邀请码的兑换明细
- **THEN** 系统 MUST 返回使用该码成功兑换的记录列表，含兑换人 `wx_id` 与时间

### Requirement: Admin API MUST 经 gateway-app 反代且禁止他域直查

`/cash/admin/api/feature/*` 与邀请码 Admin 路径 MUST 由 gateway-app cash 反代到达 cash-service。其他非 cash 进程 MUST NOT 直查 `ai_voice_cash` 功能表。

#### Scenario: 反代可达

- **WHEN** 已登录 Hub 的管理员调用功能或邀请码 Admin API
- **THEN** 请求 MUST 经 gateway 到达 cash-service 并成功鉴权（与 VIP Admin 同模式）

---

## feature-admin-contract-ux

<!-- source: openspec/specs/feature-admin-contract-ux/spec.md -->

# feature-admin-contract-ux Specification

## Purpose
TBD - created by archiving change prediction-default-invite-admin-ux. Update Purpose after archive.
## Requirements
### Requirement: 功能编号在 Admin 中 MUST 只读且不可随意新建

Admin「开通功能」界面与 API MUST NOT 允许运维修改已有功能的 `featureId`，MUST NOT 允许通过管理页创建任意新 `featureId`（新功能仅种子/发版写入）。Admin MUST 允许编辑标题、简介、开通方式、上架、排序及预测默认开通数等非编号字段。App catalog MUST 继续返回服务端 `title`/`description`，供客户端直接展示。

#### Scenario: 编辑文案不改编号

- **WHEN** 运维在管理页修改某已有功能的标题与简介并保存
- **THEN** 系统 MUST 持久化新文案，且 `featureId` MUST 保持不变；后续 catalog MUST 返回新文案

#### Scenario: 拒绝新建未知功能编号

- **WHEN** 运维尝试保存一个库中不存在的 `featureId`
- **THEN** 系统 MUST 拒绝该写入

### Requirement: 售卖套餐商品编码 MUST 可自动生成且所属功能为选择

创建功能 SKU 时，若未提供 `productCode`，服务端 MUST 自动生成全局唯一商品编码；Admin UI MUST 以只读方式展示已生成编码，MUST NOT 要求运维手填编码作为新建前提。所属功能 MUST 从已有功能定义中选择（下拉），MUST NOT 依赖自由文本输入功能编号。

#### Scenario: 新建套餐自动生成编码

- **WHEN** 运维选择所属功能、填写价格与授予参数并保存，且未填写商品编码
- **THEN** 系统 MUST 生成唯一 `productCode` 并落库，且列表可展示该编码

#### Scenario: 所属功能来自已有定义

- **WHEN** 运维打开售卖套餐表单
- **THEN** 所属功能控件 MUST 为已有 `feature_def` 的选择列表，而非任意字符串输入框

### Requirement: 开通方式 MUST 为枚举多选

Admin 配置功能允许的开通方式时，MUST 提供固定枚举（至少 `payment`、`invite_code`、`ad`）的多选控件，MUST NOT 以「逗号拼接自由字符串」作为唯一录入方式。持久化与 App 响应可仍使用逗号分隔字符串以兼容现网。

#### Scenario: 多选保存为兼容串

- **WHEN** 运维勾选「付费」与「邀请码」并保存
- **THEN** 持久化的开通方式 MUST 同时包含 `payment` 与 `invite_code`（顺序可实现自定），且 MUST NOT 依赖运维手输逗号串

### Requirement: Admin 操作按钮 MUST 与表单纵向对齐

开通功能管理页与邀请码管理页的主操作按钮（保存/创建/刷新等）MUST 与表单字段采用清晰的纵向对齐布局（同一表单流内垂直堆叠或独立操作行），避免与输入框在同一 flex 行末尾错位导致难用。

#### Scenario: 开通功能页按钮位置

- **WHEN** 运维打开开通功能管理页
- **THEN** 「保存功能定义」「保存套餐」等按钮 MUST 与对应表单区域纵向对齐，而非仅挤在字段行尾且与标签基线错乱

---

## feature-admin-rule-readonly

<!-- source: openspec/specs/feature-admin-rule-readonly/spec.md -->

# feature-admin-rule-readonly Specification

## Purpose
TBD - created by archiving change feature-ops-invite-force-admin. Update Purpose after archive.
## Requirements
### Requirement: 功能定义展示只读开通规则

开通功能管理（Admin）在展示某一功能定义时 MUST 提供只读开通规则说明（服务端按 `feature_id` 派生或常量映射），涵盖：开放主体（对人/对机）、邀请开通约束与效果口径、付费主体口径，以及共用「同宝宝禁兑」说明。运维 MUST NOT 通过该字段编辑并改写真实闸门逻辑。

#### Scenario: 成长轨迹规则文案

- **WHEN** 运维查看 `growth_trajectory_predict` 的功能定义/详情
- **THEN** 只读规则 MUST 表明对人开通、同机可兑不同邀请码、每人邀请仅一次，且提及同宝宝账号互兑禁止

#### Scenario: 值得留意规则文案

- **WHEN** 运维查看 `care_alert_smart_remind`
- **THEN** 只读规则 MUST 表明对机开通且同宝宝邀请开通仅一次

### Requirement: 同宝宝禁兑继续生效

系统兑换邀请码时 MUST 拒绝码主人当前绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换；MUST NOT 因此写入开通或获客原力。拒绝提示 MUST 明确不可使用同一宝宝下其他账号的邀请码（或同等语义）。

#### Scenario: 同宝宝拒兑

- **WHEN** 兑换者 `deviceNo` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换

---

## feature-admin-semantic-layout

<!-- source: openspec/specs/feature-admin-semantic-layout/spec.md -->

# feature-admin-semantic-layout Specification

## Purpose
TBD - created by archiving change feature-branding-logo-color. Update Purpose after archive.
## Requirements
### Requirement: 开通功能管理页 MUST 按语义分组展示功能定义表单

`cash-feature-admin.html` 的功能定义编辑区 MUST 分为至少三个语义分组：**文案**、**开通策略**、**视觉**。文案组 MUST 含功能编号（只读）、显示名称、简介。开通策略组 MUST 含开通方式、邀请/广告授予天数、默认开通条数、开放主体、是否上架、排序。视觉组 MUST 含 Logo 预览/上传与主色选择。

#### Scenario: 分组标题可见

- **WHEN** 管理员打开开通功能管理页并载入某功能
- **THEN** 表单 MUST 能区分文案、开通策略、视觉三组字段

### Requirement: 功能定义表单 MUST 适当横向排列以降低页长

短输入字段（如天数、条数、主体、上架、排序、主色与 Logo 控件）MUST 在宽屏下横向排列（两列或等价 grid），MUST NOT 将全部字段单列纵向堆叠。简介与开通方式多选可跨列全宽。保存/刷新按钮 MUST 横向排列。

#### Scenario: 宽屏下短字段并排

- **WHEN** 管理员在桌面宽度浏览功能定义表单
- **THEN** 开通策略中的短字段 MUST 至少两列并排，页面纵向长度相对原单列布局明显缩短

### Requirement: 视觉组 MUST 支持 Logo 与主色编辑

视觉组 MUST 提供：当前 Logo 预览（无则占位）、文件选择上传、`<input type="color">` 或等价色板与 hex 展示。保存功能定义时 MUST 将所选 color 与（若已上传）logo 经 Admin API 持久化。交互风格 SHOULD 对齐事件管理中的 Logo/色调编辑习惯。

#### Scenario: 选择主色并保存

- **WHEN** 管理员在视觉组选择主色并点击保存功能定义
- **THEN** 后续刷新列表/载入编辑 MUST 显示该主色

#### Scenario: 上传 Logo 后预览更新

- **WHEN** 管理员成功上传 Logo
- **THEN** 视觉组预览 MUST 更新为新图（本地 object URL 或 CDN URL）

### Requirement: 售卖套餐表单 MAY 横排短字段且 MUST NOT 配置功能视觉

售卖套餐编辑区可将价格、天数、授予数量等短字段横向排列。SKU 表单 MUST NOT 提供独立于功能定义的 logo/color 配置。

#### Scenario: SKU 无视觉字段

- **WHEN** 管理员编辑售卖套餐
- **THEN** 表单 MUST NOT 出现功能级 Logo/主色控件

---

## feature-admin-ui

<!-- source: openspec/specs/feature-admin-ui/spec.md -->

# feature-admin-ui Specification

## Purpose
TBD - created by archiving change commercial-feature-entitlement. Update Purpose after archive.
## Requirements
### Requirement: Hub MUST 提供「开通功能管理」静态页

系统 MUST 在 `gateway-app-server` 注册付费功能管理静态页（建议 `/device/admin/cash-feature-admin.html`）。页面 MUST 使用 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch` 访问功能 Admin API，支持功能定义、解锁方式、功能 SKU 等 CRUD。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 登录后可打开功能管理页

- **WHEN** 管理员已登录 Hub 并访问开通功能管理页
- **THEN** 页面 MUST 完成 Admin 校验并可通过 adminFetch 加载功能/SKU 列表

### Requirement: Hub MUST 提供独立「邀请码管理」静态页

系统 MUST 注册独立邀请码管理静态页（建议 `/device/admin/cash-invite-code-admin.html`）。页面 MUST 支持邀请码 CRUD（有效期、owner_wx_id、可开功能）及按码查看兑换明细（`wx_id` / 设备 / 功能 / 时间）。MUST 使用 AdminCommon，MUST NOT 在浏览器附带 `X-Admin-Password`。

#### Scenario: 可查看兑换明细

- **WHEN** 管理员在邀请码管理页打开某码的使用记录
- **THEN** 页面 MUST 展示该码成功兑换的用户与时间等信息

### Requirement: Admin Hub SHALL 登记两个模块入口

`admin-modules.js` MUST 增加至少两个模块：开通功能管理、邀请码管理（稳定 id、中文标题、`pagePath`、`showInNav: true`）。`RegisterAdminStaticPages` 与静态页 Bearer 白名单 MUST 包含上述 `pagePath`。

#### Scenario: 导航可见两个入口

- **WHEN** 管理员打开运维 Hub 导航
- **THEN** 模块列表 MUST 同时包含开通功能管理与邀请码管理入口

### Requirement: 管理页交互 MUST 对齐现有 cash-vip-admin 模式且可写

两页的登录、同源 API、错误提示与退出行为 MUST 对齐 Hub 惯例；MUST 为可写 CRUD，MUST NOT 误用 VIP 只读页的禁止写约束。

#### Scenario: 使用 AdminCommon 而非裸口令头

- **WHEN** 页面发起 Admin API 请求
- **THEN** 请求 MUST 经 Admin JWT / adminFetch，MUST NOT 在浏览器侧附带 `X-Admin-Password`

### Requirement: 开通功能管理页 MUST 提供 VIP列表入口

`cash-feature-admin.html` MUST 提供按钮或等价控件，可见文案 MUST 为「VIP列表」，激活后 MUST 导航至 `/device/admin/cash-vip-admin.html`。该入口 MUST 替代 Hub 主导航上的「VIP 权益」模块作为常规运维路径。

#### Scenario: 按钮文案与跳转

- **WHEN** 已登录管理员打开开通功能管理页
- **THEN** 页面 MUST 展示文案为「VIP列表」的入口，激活后 MUST 进入 VIP 权益只读页

---

## feature-branding-client

<!-- source: openspec/specs/feature-branding-client/spec.md -->

# feature-branding-client Specification

## Purpose
TBD - created by archiving change feature-branding-logo-color. Update Purpose after archive.
## Requirements
### Requirement: 客户端 MUST 从功能目录解析 logo 与 color

Flutter App MUST 在功能 catalog 模型（如 `FeatureCatalogItem`）中解析并保存每项的 `logo`（CDN URL 字符串）与 `color`（hex 字符串）。解析失败或缺字段时，`logo`/`color` MUST 视为空串，MUST NOT 导致目录整表解析失败。

#### Scenario: 解析含视觉字段的目录项

- **WHEN** catalog 返回某项含非空 `logo` 与 `color`
- **THEN** 客户端模型 MUST 暴露对应值供 UI 使用

### Requirement: 指定界面 MUST 在标题左侧展示功能 logo

下列界面在展示对应功能标题时，MUST 在标题左侧展示该功能 logo（网络图）；当 `logo` 为空时 MUST 使用本地占位，MUST NOT 留白导致标题错位：

1. 功能开通界面（各功能卡）
2. AI 分析界面（喂养分析入口与成长轨迹入口各自对应功能）
3. 功能开通弹窗
4. 喂养记录分析详情页
5. 成长轨迹详情页

#### Scenario: 开通中心卡片标题带 logo

- **WHEN** 用户打开功能开通界面且某功能已配置 logo
- **THEN** 该功能卡标题左侧 MUST 显示其 logo

#### Scenario: 无 logo 时占位

- **WHEN** 某功能 `logo` 为空
- **THEN** 上述界面 MUST 显示占位图/图标且标题仍可读

### Requirement: 指定界面 MUST 使用功能 color 作为玻璃拟态主色

上述界面中，对应功能所在的页面或对话框背景/玻璃面板 MUST 使用该功能 `color` 作为 accent（混入方式对齐喂养事件 sheet：`HistoryEditGlassPanel` / `AppModalGlassPanel` / `panelGlassGradient` 等既有 accent 混合）。AI 分析界面若同时展示喂养分析与成长轨迹入口，MUST 分别使用各自功能的 color，MUST NOT 整页强制单一功能色。

当 `color` 为空时，MUST 回退应用主题 primary（或等价），MUST NOT 崩溃。

#### Scenario: 开通弹窗随功能着色

- **WHEN** 用户打开某功能的开通弹窗且该功能 color 为非空 hex
- **THEN** 对话框玻璃拟态 MUST 呈现该主色倾向

#### Scenario: 成长轨迹详情随功能着色

- **WHEN** 用户进入成长轨迹详情页且 `growth_trajectory_predict` 的 color 非空
- **THEN** 页面/玻璃背景 MUST 使用该 color 作为 accent

#### Scenario: AI 分析双卡分色

- **WHEN** 用户打开 AI 分析界面
- **THEN** 喂养分析入口与成长轨迹入口 MUST 分别使用 care 与 growth 功能的 color（若已配置）

---

## feature-catalog-entitlement

<!-- source: openspec/specs/feature-catalog-entitlement/spec.md -->

# feature-catalog-entitlement Specification

## Purpose
TBD - created by archiving change commercial-feature-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 系统 MUST 提供绑机可读的合成功能目录 API

cash-service MUST 提供 `GET /cash/app/api/feature/catalog`：须登录且 `X-Internal-Device-No` 非空；`device_no` MUST 仅取自该内部头，MUST NOT 采信 query/body 覆盖。响应 MUST 一次返回全部启用功能项（无分页）。每项 MUST 含稳定 `featureId`、展示标题、解锁方式提示，以及该设备维度的 `unlocked`（bool）。已开通项 MUST 含 `unlockMethod` 与 `expiresAt`（unix 秒；`0` 表示永久）。UCG 入场资格 MUST NOT 作为目录项出现。该接口 MUST NOT 计入 App usage 统计。该路径 MUST NOT 加入 Bearer 匿名白名单。

#### Scenario: 一次返回完整目录且含开通态

- **WHEN** 已登录且已绑机的客户端请求功能目录
- **THEN** 响应 MUST 包含全部启用功能项，每项含 `unlocked`，且 MUST NOT 要求 page/pageSize

#### Scenario: 未绑机拒绝

- **WHEN** 请求缺少有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝，MUST NOT 返回目录数据

#### Scenario: 停用功能不出现在目录

- **WHEN** 某功能在 Admin 被停用
- **THEN** 目录 API MUST NOT 再返回该项（缓存失效后或 TTL 内一致语义以实现为准）

#### Scenario: UCG 不在目录中

- **WHEN** 客户端请求功能目录
- **THEN** 响应 MUST NOT 包含 UCG 入场资格项或用 `unlocked` 表示喂养资格

### Requirement: 过期权益 MUST NOT 视为已开通

当某权益 `expiresAt > 0` 且已早于当前时间，对应目录项 `unlocked` MUST 为 false（或等价不可用）。

#### Scenario: 过期后 unlocked 为 false

- **WHEN** 设备曾开通某功能但权益已过期
- **THEN** 目录中该项 `unlocked` MUST 为 false

### Requirement: 功能权益权威数据 MUST 落 MySQL 且按 device_no 维度

功能定义与设备权益 MUST 持久化在 `ai_voice_cash`。权益主体 MUST 为 `device_no`。其他服务 MUST NOT 直查该库。VIP 账号权益 MUST 使用独立 VIP 表，MUST NOT 因 VIP 购买写入功能权益行。

#### Scenario: 开通写入本域库

- **WHEN** 支付/邀请码/广告成功开通某功能
- **THEN** cash-service MUST 在 `ai_voice_cash` 写入或更新对应 `device_no` 权益行（或更新 `allowedCount`）

### Requirement: App MUST NOT 依赖独立 entitlements 列表接口作为主读模型

本变更 App 主读模型 MUST 为合成 catalog。系统 MUST NOT 将独立 `GET .../feature/entitlements` 作为客户端必调接口（可省略该 App 路径）。

#### Scenario: 客户端仅调 catalog 即可展示开通态

- **WHEN** 客户端已调用合成 catalog
- **THEN** 响应 MUST 足以按列表展示各功能是否已开通，无需再调独立权益列表

---

## feature-catalog-sku-embed

<!-- source: openspec/specs/feature-catalog-sku-embed/spec.md -->

# feature-catalog-sku-embed Specification

## Purpose
TBD - created by archiving change feature-catalog-embed-products. Update Purpose after archive.
## Requirements
### Requirement: App 合成功能目录项 MUST 嵌套可售 products 列表

`GET /cash/app/api/feature/catalog` 返回的每一功能项 MUST 包含 `products` 数组（可为空）。数组元素 MUST 为该 `featureId` 下状态启用的功能 SKU，且至少含：`productCode`、`priceFen`、`originalPriceFen`、`durationDays`、`grantKind`、`grantQuantity`；MAY 含 `appleProductId`。客户端 MUST 能仅凭本接口完成支付展示，并用所选 `productCode` 调用功能建单接口，MUST NOT 依赖硬编码商品码。系统 MUST NOT 为此新增独立的 App `GET /cash/app/api/feature/products` 作为必调接口。

#### Scenario: 启用 SKU 出现在对应功能项下

- **WHEN** Admin 为 `prediction_unlock` 配置了启用中的功能 SKU，且客户端请求合成 catalog
- **THEN** 该项 `products` MUST 含该 SKU 的 `productCode` 与 `priceFen` 等字段

#### Scenario: 无 payment 或无启用 SKU 时 products 为空

- **WHEN** 某功能无任何启用中的 `feature_product`
- **THEN** 该项 `products` MUST 为空数组（或等价空列表），MUST NOT 省略导致客户端解析失败（实现可用 `[]`）

#### Scenario: 停用 SKU 不下发

- **WHEN** 某 SKU `status` 非启用
- **THEN** catalog 的 `products` MUST NOT 包含该 SKU

#### Scenario: 不新增独立 App products 接口为必调

- **WHEN** 客户端需要展示功能价格并建单
- **THEN** 合成 catalog MUST 足够，MUST NOT 要求再调独立 App products 列表接口

---

## feature-order-query

<!-- source: openspec/specs/feature-order-query/spec.md -->

# feature-order-query Specification

## Purpose
TBD - created by archiving change fix-feature-alipay-fulfill-scan. Update Purpose after archive.
## Requirements
### Requirement: 功能订单按交易号查空 MUST 不阻断履约

`FulfillFeaturePaid` 在带有渠道交易号时 MUST 先查是否已有同渠道同交易号的已支付单。查无行时 MUST 视为尚未履约，MUST 继续按 `order_no` 加载订单并履约。MUST NOT 因该次空查询向调用方返回 `sql.ErrNoRows` 或导致支付宝回调失败。

#### Scenario: 首次支付宝回调功能单

- **WHEN** 功能订单状态为 `created`，支付宝以新的 `trade_no` 回调，且库中尚无该 `channel_txn_id`
- **THEN** 系统 MUST 将该单置为 `paid` 并完成功能开通，MUST NOT 因「按交易号查无」失败

#### Scenario: 同交易号重复回调

- **WHEN** 同一 `channel` 与 `channel_txn_id` 已对应一笔 `paid` 功能订单
- **THEN** 系统 MUST 幂等成功且 MUST NOT 重复开通

### Requirement: 功能订单与 SKU 读空 MUST 使用明确语义

按 `order_no` 查功能订单无行时，系统 MUST 返回明确的业务错误（订单不存在），MUST NOT 返回裸的 `sql.ErrNoRows`。按 `app_account_token` 查无行时 MUST 视为未找到（与 VIP 同语义）。按商品编码或 Apple 商品 ID 查启用 SKU 无行时，系统 MUST 返回明确的业务错误（商品不存在或未映射），MUST NOT 返回裸的 `sql.ErrNoRows`。

#### Scenario: 不存在的功能订单号

- **WHEN** 履约请求的 `order_no` 在 `feature_order` 中不存在
- **THEN** 错误信息 MUST 表明订单不存在，且 MUST NOT 仅为 `sql: no rows in result set`

---

## feature-sku-update

<!-- source: openspec/specs/feature-sku-update/spec.md -->

# feature-sku-update Specification

## Purpose
TBD - created by archiving change feature-sku-update-only. Update Purpose after archive.
## Requirements
### Requirement: 保存功能套餐 MUST NOT 新建商品编码

Admin 更新功能 SKU 时，商品编码去空白后为空，或 `feature_product` 中不存在该编码，系统 MUST 拒绝且 MUST NOT 插入新行。编码已存在时 MUST 只更新该行的价格、天数、Apple 商品 ID、上下架及其他非主键字段，MUST NOT 修改 `product_code`。

#### Scenario: 空编码保存

- **WHEN** 管理员提交的商品编码为空
- **THEN** 系统 MUST 拒绝，且 `feature_product` 行数 MUST 不变

#### Scenario: 已有编码改价

- **WHEN** 管理员提交已存在的商品编码，并把价格改为新的分值
- **THEN** 该编码的价格 MUST 更新，且 MUST NOT 出现新的商品编码

### Requirement: 打开功能时 MUST 填入已有套餐

开通功能管理选中某个功能时，售卖套餐表单 MUST 填入该功能已有套餐的商品编码。若存在上架且编码不以 `fp_` 开头的行，MUST 优先填入该行；否则填入上架行；再否则填入列表中的第一条。该功能没有任何套餐时，编码 MUST 保持为空，且页面 MUST NOT 生成新编码。页面 MUST NOT 再提供清空编码以新建套餐的操作。

#### Scenario: 打开已有种子套餐的功能

- **WHEN** 管理员打开一个已有非 `fp_` 上架套餐的功能
- **THEN** 表单商品编码 MUST 为该套餐编码

#### Scenario: 没有新建入口

- **WHEN** 管理员查看售卖套餐区域
- **THEN** 页面 MUST NOT 显示「新建套餐」

---

## feature-unlock-fulfillment

<!-- source: openspec/specs/feature-unlock-fulfillment/spec.md -->

# feature-unlock-fulfillment Specification

## Purpose
TBD - created by archiving change commercial-feature-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 功能支付履约 MUST 授予 device_no 权益或递增 allowedCount

cash-service MUST 支持独立功能 SKU（`feature_product`）：建单与支付宝/Apple IAP 验签 MUST 与 VIP 支付栈一致（服务端验签）。支付成功后 MUST 按商品配置向订单关联的主体授予功能权益或递增 `allowedCount`（主体按功能 `activation_subject`：device 或 user）。功能订单 MUST 使用独立 `feature_order` 表。支付宝异步通知、Apple ASN 与 Apple 客户端验单入口 MUST 与 VIP **共用分流**（按订单表区分），MUST NOT 污染 VIP 订单语义。`channel=apple_iap` 的功能建单 MUST 生成并返回 UUID `appAccountToken`，供 StoreKit 与 ASN 映射。

#### Scenario: 功能 SKU 支付成功写权益

- **WHEN** 某功能商品支付回调/验单成功且 `grant_kind=entitlement`
- **THEN** 系统 MUST 幂等为订单主体写入/续期对应 `feature_id` 权益，`unlockMethod` MUST 为 `payment`

#### Scenario: 功能 SKU 支付成功增加 allowedCount

- **WHEN** 某功能商品支付成功且 `grant_kind` 为预测数量增量
- **THEN** 系统 MUST 幂等将该主体的 `allowedCount` 增加配置的 `grant_quantity`

#### Scenario: 共用回调按订单分流

- **WHEN** 支付宝通知或 Apple ASN 到达且订单号/token 属于 `feature_order`
- **THEN** 系统 MUST 执行功能履约且 MUST NOT 续期 `vip_entitlement`

#### Scenario: 重复回调不重复授予

- **WHEN** 同一渠道交易号或已 paid 订单再次通知
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误时长或数量

#### Scenario: 功能 Apple 建单返回 appAccountToken

- **WHEN** 用户以 `channel=apple_iap` 创建功能订单
- **THEN** 响应 MUST 含 UUID `appAccountToken` 且落库可被 ASN 反查

### Requirement: VIP 购买 MUST NOT 写入功能权益

VIP 商品履约 MUST 仅续期 `vip_entitlement`（`wx_id`）。MUST NOT 为任何 `device_no` 写入 `feature_entitlement`，MUST NOT 修改 `feature_allowed_count`。既有 VIP status API MUST 保持可用，供客户端 V1 覆盖功能列表（不得覆盖 UCG 入场）。

#### Scenario: VIP 履约不影响功能表

- **WHEN** 用户成功购买或续期月 VIP
- **THEN** 系统 MUST 更新 VIP 到期时间，且 MUST NOT 新增或更新任何功能权益行 / allowedCount

### Requirement: 邀请码兑换 MUST 按单功能开通并遵守一家绑定与人维去重

cash-service MUST 提供 `POST /cash/app/api/feature/invite-codes/redeem`（计入 usage）。请求 MUST 含邀请码与目标 `featureId`。主体：`device_no` 与 `wx_id` 均只信网关注入头；`wx_id` MUST > 0。

规则 MUST 包括：

- 一码可覆盖其配置的（或所有支持邀请码的）多个功能，但单次请求 MUST 只开通所请求的一个 `featureId`，MUST NOT 自动开通其余功能。
- 同一 `wx_id` 对同一 `featureId` 仅能通过邀请码开通一次。
- 同一 `wx_id` 仅能使用一家 `owner_wx_id` 的邀请码；**仅当成功开通某一功能后**才写入该绑定；失败兑换 MUST NOT 绑定。
- `redeemer_wx_id` MUST NOT 等于码的 `owner_wx_id`（不可自用）。
- 成功后 MUST 写设备权益或 `allowedCount`，`unlockMethod` MUST 为 `invite_code`，并写入兑换流水（含 owner/redeemer/device/feature/time）。

#### Scenario: 指定功能兑换成功

- **WHEN** 合法用户提交有效码与尚未码开的 `featureId` 且未违反一家/自用规则
- **THEN** 系统 MUST 仅开通该功能（或增加对应数量），并占用相应兑换记录

#### Scenario: 同码再次开通另一功能

- **WHEN** 用户已用某码成功开通功能 A，再次提交同一码与功能 B（B 尚未码开且码支持 B）
- **THEN** 系统 MUST 允许开通 B，且 MUST NOT 因已开 A 而自动开其它功能

#### Scenario: 换一家邀请码被拒绝

- **WHEN** 用户已成功使用 owner=X 的码开通过任一功能，再提交 owner=Y 的码
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 失败兑换不绑定一家

- **WHEN** 用户首次兑码因码无效/过期/功能不支持等原因失败
- **THEN** 系统 MUST NOT 写入 redeemer→owner 绑定

#### Scenario: 不可自用

- **WHEN** 兑换人 `wx_id` 等于码的 `owner_wx_id`
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 同一功能不可再次码开

- **WHEN** 某 `wx_id` 已通过邀请码开通过某 `featureId`
- **THEN** 再次用任意邀请码开通同一 `featureId` MUST 被拒绝

### Requirement: 广告完成开通 MVP MUST 信任客户端申报

cash-service MUST 提供 `POST /cash/app/api/feature/ad/complete`（计入 usage）。MVP MUST 接受客户端完成申报并为 `device_no` 授予对应功能或数量，`unlockMethod` MUST 为 `ad`。MUST 提供基础防刷（短窗幂等或设备日限额）。MUST NOT 强制接入第三方广告服务端验真。

#### Scenario: 客户端申报广告完成获得开通

- **WHEN** 合法请求申报某 `featureId` 广告完成且未触发防刷拒绝
- **THEN** 系统 MUST 授予对应权益或数量增量

#### Scenario: 重复申报幂等

- **WHEN** 同一幂等键在短窗内重复提交广告完成
- **THEN** 系统 MUST NOT 重复叠加授予

---

## feature-visual-branding

<!-- source: openspec/specs/feature-visual-branding/spec.md -->

# feature-visual-branding Specification

## Purpose
TBD - created by archiving change feature-branding-logo-color. Update Purpose after archive.
## Requirements
### Requirement: feature_def MUST 持久化 logo 与 color

`ai_voice_cash.feature_def` MUST 包含 `logo` 与 `color` 列。`logo` MUST 存储 OSS objectKey（建议前缀 `feature/`），MUST NOT 在库内持久化完整 CDN URL 作为权威值。`color` MUST 为 `#RGB` 或 `#RRGGBB`（大小写不敏感）或空串。cash-service 以外进程 MUST NOT 直查该表。

#### Scenario: 写入 objectKey 与 hex

- **WHEN** 管理员保存某功能的 logo 与主色
- **THEN** 系统 MUST 将 objectKey 与校验通过的 color 持久化到该 `feature_id` 行

#### Scenario: 非法 color 拒绝

- **WHEN** 提交的 color 不是空串且不符合 `#RGB`/`#RRGGBB`
- **THEN** 系统 MUST 拒绝保存并返回错误

### Requirement: 系统 MUST 为三个内置功能种子默认 color

EnsureSchema/种子 MUST 为下列功能写入默认 `color`（logo 默认为空串）：

| feature_id | color |
|------------|-------|
| `prediction_unlock` | `#3B82F6` |
| `care_alert_smart_remind` | `#0D9488` |
| `growth_trajectory_predict` | `#EA580C` |

当行已存在且 `color` 非空时，种子 MUST NOT 覆盖该 color。

#### Scenario: 新库种子带默认色

- **WHEN** 全新库执行功能定义种子
- **THEN** 上述三个 `feature_id` 的 `color` MUST 分别为表中默认值且 `logo` 为空

#### Scenario: 运维已改色不被种子覆盖

- **WHEN** 某功能 `color` 已被运维改为非空自定义值后再次跑种子
- **THEN** 该行 `color` MUST 保持运维值

### Requirement: App 功能目录 MUST 下发 logo 与 color

`GET /cash/app/api/feature/catalog` 每项 MUST 包含 `logo` 与 `color`。HTTP 边界上 `logo` MUST 为可访问的 CDN 绝对 URL；若库内 objectKey 为空，则 `logo` MUST 为空串。`color` MUST 原样下发库内值（种子后内置功能通常非空）。该扩展属于 v1 结构变更；本能力不要求保留无字段旧响应。

#### Scenario: 已配置视觉的目录项

- **WHEN** 某启用功能已配置 logo objectKey 与 color，客户端请求 catalog
- **THEN** 该项 MUST 含非空 CDN `logo` 与对应 `color`，且仍含既有 `unlocked` 等开通态字段

#### Scenario: 未上传 logo

- **WHEN** 某功能 `logo` 库内为空
- **THEN** catalog 该项 `logo` MUST 为空串，`color` 仍 MUST 返回（含种子默认色）

### Requirement: Admin 功能定义 API MUST 读写 logo 与 color

功能定义列表与更新 API MUST 支持 `logo` 与 `color`。列表/读出时 `logo` MUST 映射为 CDN URL（与 App 一致）。更新时客户端可提交 objectKey（上传接口返回值）与 color；未更换 logo 时 MUST 允许保留原 objectKey。

#### Scenario: 更新主色

- **WHEN** 管理员仅修改 color 并保存
- **THEN** 系统 MUST 更新 color，且 MUST NOT 清空已有 logo

### Requirement: 系统 MUST 支持功能 logo 上传至 OSS

系统 MUST 提供 Admin 可调用的功能 logo 上传能力（建议 `POST /cash/admin/api/feature/defs/logo` 或等价），经 ucg 契约写入 OSS，返回 objectKey（及可选 CDN URL）。objectKey MUST 使用 `feature/` 前缀（或与实现约定的等价 feature 命名空间）。实现 MUST 经 `clients/ucg`（或 platform 封装），MUST NOT 从 cash 业务包 import device 服务实现。

#### Scenario: 上传成功返回 objectKey

- **WHEN** 管理员上传合法 png/jpg/jpeg/webp logo
- **THEN** 系统 MUST 返回可写入 `feature_def.logo` 的 objectKey

#### Scenario: 超大或非法类型拒绝

- **WHEN** 文件超过大小上限或扩展名不支持
- **THEN** 系统 MUST 拒绝上传且 MUST NOT 写入 `feature_def`

---

## feeding-chat-ws-io

<!-- source: openspec/specs/feeding-chat-ws-io/spec.md -->

# feeding-chat-ws-io Specification

## Purpose
TBD - created by archiving change unify-feeding-chat-ws. Update Purpose after archive.
## Requirements
### Requirement: chat WS 支持输入输出模态组合

`/voice/chat/ws` MUST 在保留现有音频流式能力的基础上，支持自然语言喂养的四种 I/O 组合：输入 `audio|text`，输出 `audio|text`。模态 MUST 由客户端在 `start` 帧中声明（字段名实现可选 `inputModality`/`outputModality` 或等价约定）；缺省 MUST 为 `audio`/`audio`，以保持横屏语音兼容。

#### Scenario: 默认音入音出兼容横屏

- **WHEN** 客户端发送既有横屏 `start`（未声明模态或显式 audio/audio）并上传 PCM
- **THEN** 服务端 MUST 按现网行为进行 ASR、意图流式、TTS，并下发 `thinking_delta`/`answer`/`audio_*`（及既有 asr 帧）

#### Scenario: 文入文出

- **WHEN** 客户端 `start` 声明 text 输入与 text 输出，随后发送 text 上行帧（非空文本）
- **THEN** 服务端 MUST NOT 依赖 PCM/ASR 完成该轮，MUST 下发 `thinking_delta` 与 `answer`，MUST NOT 下发 TTS `audio_chunk`/`audio_end`

#### Scenario: 文入音出

- **WHEN** 客户端声明 text 输入与 audio 输出并发送 text 帧
- **THEN** 服务端 MUST 跳过 ASR，MUST 在意图完成后执行 TTS 并下发音频帧

#### Scenario: 音入文出

- **WHEN** 客户端声明 audio 输入与 text 输出并完成一轮 ASR commit
- **THEN** 服务端 MUST 下发思考与 answer，MUST NOT 下发 TTS 音频帧

### Requirement: 文模式 start 仍校验音频元数据

纯 text 输入时，`start` MUST 仍要求与现网一致的音频元数据字段（含 `sampleRate`、`bits`、`channels`、`length`、`mode=stream` 等既有必填项）校验通过；服务端 MUST NOT 因文模式放宽或省略这些字段的必填校验。客户端 MAY 传入与横屏一致的占位数值。

#### Scenario: 缺 sampleRate 的文模式 start 被拒

- **WHEN** 客户端声明 text 输入但 `start` 缺少合法 `sampleRate`（或既有必填音频字段）
- **THEN** 服务端 MUST 拒绝该 `start`（错误帧或等价失败），MUST NOT 进入 `started` 可对话状态

### Requirement: 喂养对话经硬件特权与 VU 正式模

经 `/voice/chat/ws` 触发的喂养意图 LLM MUST 标记硬件特权，MUST 使用 `voiceUnderstanding` lane 的正式（非 free）模型配置，MUST 经该 lane 并发闸门 `Acquire`。该路径 MUST NOT 因 `voice_ai` 月度额度用尽而拒绝对话或改走 free 模型。

#### Scenario: 不计 voice_ai 次

- **WHEN** 一轮 WS 喂养意图 LLM 成功完成（含文模式）
- **THEN** 系统 MUST NOT 对该轮执行 `voice_ai` 成功计次（consume 为 no-op 或不调用）

---

## feeding-effective-day-core

<!-- source: openspec/specs/feeding-effective-day-core/spec.md -->

# feeding-effective-day-core Specification

## Purpose
TBD - created by archiving change feeding-eligibility-admin-scenes. Update Purpose after archive.
## Requirements
### Requirement: 有效喂养日判定 MUST 独立于场景门槛且仅由 cash-service 执行

系统 MUST 在 **cash-service** 提供与场景解耦的有效日判定：使用 `Asia/Shanghai` 日历日；某日有效当且仅当该 `device_no` 在该日 history 记录数 ≥ 该次计算使用的 `minRecordsPerDay`；连续有效日 MUST 以 **上海昨日** 为锚向前连续累计（`days[0]` MUST 为昨天），遇无效日中断；**请求当日（今日）MUST NOT** 计入有效日或连续 streak。资格计算 MUST 经 history-service HTTP 按日计数，MUST NOT 由 cash 直查 history 库。UCG 入场与值得留意喂养资格的天数判断 MUST NOT 实现在 device-service、voice-service、history-service 或客户端权威路径。

#### Scenario: 日条数达门槛计有效

- **WHEN** 某上海已闭合日 history 条数等于该次 `minRecordsPerDay`
- **THEN** 该日 MUST 计为有效喂养日

#### Scenario: 不足门槛中断连续

- **WHEN** 从昨天往回扫描时某日条数小于 `minRecordsPerDay`
- **THEN** 该日 MUST NOT 计入，且连续计数 MUST 在该日中断

#### Scenario: 今日不计入

- **WHEN** 请求时刻为上海某日且当日 history 条数已 ≥ `minRecordsPerDay`
- **THEN** 该「今日」MUST NOT 增加 `effectiveDays`，连续扫描 MUST 自昨天起算

#### Scenario: 跨日立即合格

- **WHEN** 上海日切换后，以昨天为锚的连续有效日 ≥ 该场景 `requiredDays`
- **THEN** 当日资格请求 MUST 可返回 `qualified=true`（按请求日缓存键 miss 后重算即可），MUST NOT 依赖额外后台 ticker

#### Scenario: 资格合成宿主为 cash

- **WHEN** 计算 UCG 或值得留意喂养资格
- **THEN** 连续天数与 `qualified` 合成 MUST 在 cash-service 完成，history MUST 仅返回按日条数

### Requirement: 取数窗口 MUST 等于所需连续天数

调用 history 按日统计时，请求天数 MUST 等于本次合成所需的 `requiredDays`（单场景）；窗口 MUST 覆盖自 **上海昨日** 起往前共 `requiredDays` 个已闭合日历日，MUST NOT 包含今日；MUST NOT 无必要请求大于该阈值的固定窗口（例如仅为缓冲而固定 14）。若一次请求需服务多个场景，窗口 MUST 为各场景 `requiredDays` 的最大值且 MUST NOT 更大。history 实现 MUST 在数据库侧按上海日聚合 `COUNT`（或等价），MUST NOT 为资格统计拉取窗口内全部 history 行的 `start_time` 列表再在应用层逐条累加。

#### Scenario: UCG 仅拉 requiredDays 已闭合日

- **WHEN** 计算 `ucg_entry` 且配置 `requiredDays=7`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `7`，且返回序列 `days[0]` MUST 为上海昨天

#### Scenario: 值得留意仅拉 requiredDays 已闭合日

- **WHEN** 计算 `care_alert_entry` 且配置 `requiredDays=2`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `2`，且窗口实质为上海昨天与前天（不含今日）

#### Scenario: 按日 COUNT 返回

- **WHEN** history 处理 `feeding-day-stats` 且窗口内某日有多条记录
- **THEN** 响应该日 MUST 为聚合后的单条 `{date, count}`，MUST NOT 要求调用方再按原始行重算

### Requirement: 各场景 MUST 使用各自配置的连续天数与日条数门槛

场景 `ucg_entry` 与 `care_alert_entry` MUST 分别持久化 `requiredDays` 与 `minRecordsPerDay`；合成某一场景资格时 MUST 只使用该场景配置，MUST NOT 强制两场景共用同一日条数门槛。VIP 与功能权益 MUST NOT 改变任一场景的 `qualified`。

#### Scenario: 两场景门槛可不同

- **WHEN** `ucg_entry` 的 `minRecordsPerDay=10` 且 `care_alert_entry` 的 `minRecordsPerDay=5`
- **THEN** 同一设备同一日的 history 条数=6 时，该日对值得留意 MUST 计有效，对 UCG MUST NOT 计有效

#### Scenario: UCG 合格判据读配置

- **WHEN** `ucg_entry.requiredDays=7` 且连续有效日 ≥ 7
- **THEN** UCG eligibility MUST `qualified=true` 且 `requiredDays=7`

### Requirement: 各场景门槛语义不变且 care_alert 为促活门槛

场景 `ucg_entry` 与 `care_alert_entry` MUST 分别持久化并使用各自的 `requiredDays` 与 `minRecordsPerDay`；合成某一场景资格时 MUST 只使用该场景配置，MUST NOT 强制两场景共用同一日条数门槛。VIP 与功能权益 MUST NOT 改变任一场景的 `qualified`。`care_alert_entry` 的连续有效日门槛 MUST 视为功能解锁 / 促活门槛，MUST NOT 被定义为「为值得留意数据准确性而要求昨日有发生」的继承语义；排除今日后 `requiredDays=2` 几何上对应「昨天+前天」连续有效，属阈值结果而非旧客户端闸门的等价替换说明义务。

#### Scenario: 两场景门槛可不同

- **WHEN** `ucg_entry` 的 `minRecordsPerDay=10` 且 `care_alert_entry` 的 `minRecordsPerDay=5`
- **THEN** 同一设备同一已闭合日的 history 条数=6 时，该日对值得留意 MUST 计有效，对 UCG MUST NOT 计有效

#### Scenario: UCG 合格判据读配置

- **WHEN** `ucg_entry.requiredDays=7` 且以昨天为锚的连续有效日 ≥ 7
- **THEN** UCG eligibility MUST `qualified=true` 且 `requiredDays=7`

#### Scenario: care_alert 满两闭合日合格

- **WHEN** `care_alert_entry.requiredDays=2` 且上海昨天与前天均为有效喂养日
- **THEN** care-alert eligibility MUST `qualified=true`

---

## feeding-eligibility-admin

<!-- source: openspec/specs/feeding-eligibility-admin/spec.md -->

# feeding-eligibility-admin Specification

## Purpose
TBD - created by archiving change feeding-eligibility-admin-scenes. Update Purpose after archive.
## Requirements
### Requirement: Admin MUST 可配置喂养资格场景阈值

cash-service MUST 提供管理端 API（及 Hub 静态页入口）读取与更新已登记场景的 `requiredDays`、`minRecordsPerDay`。场景键 `ucg_entry`、`care_alert_entry` MUST 由种子/发版约定，Admin MUST NOT 创建任意新 `scene_key`。写操作 MUST 校验 Admin 鉴权；成功后 MUST 使相关资格缓存失效（或等价使旧缓存不可命中）。

#### Scenario: 更新 UCG 连续天数

- **WHEN** 运维将 `ucg_entry.requiredDays` 从 7 改为 5 并保存
- **THEN** 后续 UCG eligibility 的 `requiredDays` MUST 为 5，且缓存 MUST NOT 长期返回改前结果

#### Scenario: 拒绝未知场景新建

- **WHEN** 运维尝试写入未登记的 `scene_key`
- **THEN** 系统 MUST 拒绝该写入

### Requirement: 场景配置 MUST 有安全默认种子

EnsureSchema（或等价迁移）MUST 种子 `ucg_entry`（默认 requiredDays=7、minRecordsPerDay=10）与 `care_alert_entry`（默认 requiredDays=2、minRecordsPerDay=10），幂等且 MUST NOT 在重复启动时用种子覆盖运维已改值（或仅 INSERT 忽略已存在）。

#### Scenario: 首次部署有两行默认

- **WHEN** 空库执行 EnsureSchema
- **THEN** 配置表 MUST 含上述两场景默认行

---

## feeding-llm-admin-simplify

<!-- source: openspec/specs/feeding-llm-admin-simplify/spec.md -->

# feeding-llm-admin-simplify Specification

## Purpose
TBD - created by archiving change unify-feeding-chat-ws. Update Purpose after archive.
## Requirements
### Requirement: voiceUnderstanding 不再配置 free 模型

喂养 lane `voiceUnderstanding` 的 Admin 配置与 API MUST 仅包含正式模字段：`provider`、`model`、`maxInFlight`、`maxWaiters`（及既有审计字段）。MUST NOT 再要求或展示 `freeProvider`/`freeModel`。PUT 若仍收到 free 字段，MUST 忽略或拒绝，且 MUST NOT 再将其作为额尽降级选模真相源。

#### Scenario: Admin 保存 VU 无 free 控件

- **WHEN** 运维打开 AI 模型与并发页的 voiceUnderstanding（喂养默认智能体）区块并保存
- **THEN** UI MUST NOT 提供 free 模型下拉；保存体 MUST NOT 依赖 free 字段才能成功

#### Scenario: 硬件喂养只用正式模

- **WHEN** `/voice/chat/ws` 触发喂养 LLM
- **THEN** 选模 MUST 使用 VU 正式 provider/model，MUST NOT 读取 VU free 配置

### Requirement: 额度 Admin 隐藏喂养月度项

Voice AI 额度管理界面 MUST 隐藏喂养（`voice_ai` / `voiceAiMonthlyLimit`）全局默认与 per-user 覆盖的展示与编辑入口。clinic / care-alert 等其他 feature 的额度配置 MUST 保持可用（若该页已有）。

#### Scenario: 额度页无喂养表单项

- **WHEN** 管理员打开 voice-admin 额度功能区
- **THEN** 页面 MUST NOT 展示可编辑的喂养月度额度字段

### Requirement: 喂养对话与 voice_ai 额度解耦

自然语言喂养对话路径 MUST NOT 再以 `voice_ai` 月度 used/limit 作为拒绝或 free 降级条件。`voice_ai` 存储与内部 API MAY 短期保留，但 MUST NOT 作为 chat WS 对话门禁。

#### Scenario: 无 40302 挡喂养 WS

- **WHEN** 某 wx 的 `voice_ai` 已用尽且客户端经 `/voice/chat/ws` 发起喂养对话
- **THEN** 服务端 MUST NOT 因该额度用尽返回阻断对话的 40302（或等价拒绝）；MUST 仍可走 VU 正式模（受并发闸门约束）

---

## feeding-shell-b-removal

<!-- source: openspec/specs/feeding-shell-b-removal/spec.md -->

# feeding-shell-b-removal Specification

## Purpose
TBD - created by archiving change unify-feeding-chat-ws. Update Purpose after archive.
## Requirements
### Requirement: 删除对外 history 文本喂养 API

系统 MUST NOT 再提供 `POST /device/history/api/chat` 与 `POST /device/history/api/chat/stream` 作为自然语言喂养入口。本变更 MUST 直接删除（或取消注册）上述路由与 `HistoryCtrl` 对应处理方法，MUST NOT 保留过渡期 200 兼容实现。

#### Scenario: chat 路径不可用

- **WHEN** 客户端请求 `POST /device/history/api/chat` 或 `.../chat/stream`
- **THEN** 网关/history MUST 返回未找到或未注册路由的失败语义（非成功 SSE/JSON 对话）

### Requirement: 删除 voice internal 文本喂养 API

voice-service MUST NOT 再提供 `POST /voice/internal/api/text/chat` 与 `POST /voice/internal/api/text/chat/stream`。history-service MUST 删除 `DelegateTextChat` 与 `DelegateTextChatStream` 及其调用点。若 `POST /voice/text/chat` 仅服务于已删除文本喂养调试路径，MUST 一并删除。

#### Scenario: internal text chat 不可用

- **WHEN** 带内部密钥请求 `/voice/internal/api/text/chat` 或 `.../stream`
- **THEN** voice-service MUST 不再成功执行文本喂养对话

### Requirement: 自然语言喂养仅经 chat WS

自然语言喂养（意图理解与业务落地）对外 MUST 仅经 `/voice/chat/ws`。系统 MUST NOT 再维护第二条 HTTP 文本喂养外壳。

#### Scenario: 文档与 Admin 说明一致

- **WHEN** 运维查看 voice-admin 或相关 runbook 中的喂养 AI 入口列表
- **THEN** 列表 MUST 以 `/voice/chat/ws` 为喂养对话入口，MUST NOT 再将已删的 history/internal text chat 路径列为可用入口

---

## gateway-admin-jwt

<!-- source: openspec/specs/gateway-admin-jwt/spec.md -->

# gateway-admin-jwt Specification

## Purpose
TBD - created by archiving change admin-consolidate-gateway-app-jwt. Update Purpose after archive.
## Requirements
### Requirement: gateway-app SHALL provide admin login issuing Admin JWT

`gateway-app-server` MUST 提供 `POST /device/admin/api/login`（`Content-Type: application/json`），请求体含 `username` 与 `password`。校验 MUST 对照环境变量 `GATEWAY_APP_ADMIN_USERNAME`（默认 `admin`）与 `GATEWAY_APP_ADMIN_PASSWORD`。成功时响应 MUST 为 `{ code: 0, data: { accessToken, expiresIn } }`，其中 `accessToken` 为 HS256 JWT，`aud` MUST 为 `gateway-admin`，`iss` MUST 为 `gateway-app-server`。未配置 `GATEWAY_APP_ADMIN_PASSWORD` 时 MUST 返回 503 且 SHALL NOT 签发 token。登录接口 MUST 列入 Bearer 白名单。

#### Scenario: 正确账号密码登录

- **WHEN** 客户端提交与 env 一致的 username/password
- **THEN** 系统 SHALL 返回 `code=0` 及非空 `accessToken`

#### Scenario: 密码错误

- **WHEN** 客户端提交错误 password
- **THEN** 系统 SHALL 返回 401 且 SHALL NOT 签发 token

#### Scenario: 未配置 admin 密码

- **WHEN** `GATEWAY_APP_ADMIN_PASSWORD` 为空且客户端请求 login
- **THEN** 系统 SHALL 返回 503 语义（管理未启用）

### Requirement: Admin JWT and user access JWT SHALL be mutually isolated

Bearer 中间件 MUST 区分 Admin JWT（`aud=gateway-admin`）与用户 access JWT。管理 API 路径（见下条）MUST 要求有效 Admin JWT。App/UCG 用户 API MUST 要求用户 JWT 且 MUST 拒绝仅含 Admin JWT 的请求。用户 JWT MUST NOT 访问管理 API（返回 403）。

#### Scenario: Admin JWT 访问设备管理 API

- **WHEN** 客户端对 `GET /device/admin/api/user/list` 携带有效 Admin JWT
- **THEN** 请求 SHALL 通过 gateway-app Bearer 校验并进入下游或本机 handler

#### Scenario: Admin JWT 访问用户 profile API

- **WHEN** 客户端对受保护 App API 仅携带 Admin JWT
- **THEN** gateway-app SHALL 返回 403

#### Scenario: 用户 JWT 访问管理 API

- **WHEN** 客户端对 `GET /device/admin/api/event/list` 仅携带用户 access JWT
- **THEN** gateway-app SHALL 返回 403

### Requirement: gateway-app SHALL inject downstream admin passwords server-side

校验 Admin JWT 成功后，gateway-app MUST 在转发或本机处理前注入：`/device/admin/api/*` 与 `/device/app/api/version/admin/*` 注入 `X-Admin-Password` 等于 `DEVICE_ADMIN_PASSWORD`；`/ucg/admin/api/*` 注入 `X-Admin-Password` 等于 `UCG_ADMIN_PASSWORD`。Hook 入口 MUST 删除客户端传入的 `X-Admin-Password`，防止伪造。

#### Scenario: 浏览器不传 X-Admin-Password 仍可调用 device 管理 API

- **WHEN** 管理员仅携带 Admin JWT 请求 `GET /device/admin/api/user/list`
- **THEN** gateway-app 反代至 device-service 的请求 SHALL 含服务端注入的有效 `X-Admin-Password` 且 device-service SHALL 返回业务数据

#### Scenario: 客户端伪造 X-Admin-Password 无效

- **WHEN** 客户端携带错误 `X-Admin-Password` 与有效 Admin JWT
- **THEN** 下游 SHALL 仍使用网关注入口令且 SHALL NOT 使用客户端伪造值

### Requirement: Web admin pages SHALL be served only by gateway-app

下列静态路由 MUST 仅由 `gateway-app-server` 注册：`/device/admin`、`/device/admin/qa-records`、`/device/admin/feedback-records`、`/device/admin/api-usage-stats`、`/device/admin/ucg-admin.html`、`/device/app/version-admin.html`、`/device/history/*deviceNo`（history 壳页）。`gateway-service` MUST NOT 再 `ServeFile` 上述路径。

#### Scenario: 9702 可打开设备管理页

- **WHEN** 客户端 `GET /device/admin` 访问 App 网关
- **THEN** 系统 SHALL 返回 `admin.html`

#### Scenario: 9701 不直接提供 admin 静态页

- **WHEN** 客户端 `GET /device/admin` 访问主网关且迁移已完成
- **THEN** 系统 SHALL NOT 返回 admin 静态 HTML 作为 200 正文（SHALL 302 或 runbook  documented 等价行为）

### Requirement: gateway-service SHALL redirect legacy admin URLs to App gateway

主网关 MUST 对 `/device/admin` 及子路径（静态 admin 页）、以及 `/device/history/*` 壳页（不含 `/device/history/api/*`）返回 **302**，`Location` 为 `GATEWAY_APP_PUBLIC_BASE_URL` 与请求路径拼接；env 未配置时 MAY fallback 至同 host 的 App 网关端口（如 `:9702`）。

#### Scenario: 旧 bookmark 跳转

- **WHEN** 用户访问 `https://example.com:9701/device/admin`
- **THEN** 浏览器 SHALL 被重定向至 App 网关等价路径

### Requirement: admin front-end SHALL use shared Bearer client

仓库 MUST 提供 `resource/public/admin-common.js`（登录、`Authorization: Bearer` fetch、logout）与 `resource/public/admin-modules.js`（模块登记）。所有管理静态页 MUST 通过 admin-common 调用 API，MUST NOT 在浏览器侧发送 `X-Admin-Password`。`admin.html` MUST 从 admin-modules 渲染导航链接，MUST NOT 使用 `gatewayAppBase()` 端口配对。

#### Scenario: Hub 登录后子页共用 token

- **WHEN** 管理员在 `/device/admin` 登录成功
- **THEN** 打开 `/device/admin/qa-records` SHALL 使用同一 Admin JWT 加载数据而无需再次输入口令

### Requirement: New admin modules SHALL be registered in admin-modules.js

新增 Web 管理模块时 MUST 在 `admin-modules.js` 增加条目（至少含 `id`、`title`、`pagePath`、`apiPrefixes`），并在 gateway-app 静态路由单点注册表中增加 `pagePath`。PR MUST NOT 新增仅主网关可见的 admin 静态路由。

#### Scenario: 登记与路由一致

- **WHEN** admin-modules 声明 `pagePath=/device/admin/foo`
- **THEN** gateway-app MUST 注册该路径对应的静态文件 handler

---

## gateway-app-api-usage-stats

<!-- source: openspec/specs/gateway-app-api-usage-stats/spec.md -->

# gateway-app-api-usage-stats Specification

## Purpose
TBD - created by archiving change gateway-app-api-usage-stats. Update Purpose after archive.
## Requirements
### Requirement: gateway-app SHALL record successful App HTTP API usage after response

`gateway-app-server` MUST 在 HTTP 响应确定后（状态码已写入客户端方向）评估是否写入使用统计。实现 MUST 覆盖经领域反代（device/ucg/history/voice）短路 `ExitAll` 的路径，不得仅依赖 `BindMiddleware("/*")` 在 `Next()` 之后记录。仅当响应状态码满足 `200 <= status < 300` 时 SHALL 计数一次。统计路径 MUST 为归一化后的 `METHOD /path`（不含 query）。下列请求 MUST NOT 写入统计：WebSocket 升级、`/device/internal/` 前缀、`/device/admin/api/` 前缀（含本变更读 API 自身）、静态资源与 HTML 壳页，以及维护型 App API（`POST /device/app/api/token/refresh`、`GET /device/app/api/version/check`、`GET /device/app/api/site/home`、`/device/app/api/version/admin/*` 前缀、**`GET /ucg/app/api/posts/{id}/comments`**）。**此外，当请求关联的 `wxId > 0` 且该 wxId 在 device 域标记为 `is_simulated=1` 时，MUST NOT 写入任何 usage 统计（全局日计数、per-wxId、交叉维度均跳过）。** 登录、注册、绑定、POST 评论与各业务 App API 对**非模拟**用户 SHALL 继续计入。写入 MUST 异步执行且 SHALL NOT 阻塞或改变业务响应。

#### Scenario: token 刷新不计入

- **WHEN** 经 gateway-app 的 `POST /device/app/api/token/refresh` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: GET 评论列表不计入

- **WHEN** 经 gateway-app 的 `GET /ucg/app/api/posts/123/comments` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: POST 评论仍计入

- **WHEN** 经 gateway-app 的 `POST /ucg/app/api/posts/123/comments` 返回 HTTP 200 且调用方 wxId 非模拟用户
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: Simulated user API call skipped

- **WHEN** 模拟用户 wxId=1001 经 gateway 调用 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
- **THEN** 系统 MUST NOT 增加全局、wxId=1001 或交叉维度计数

#### Scenario: 登录 API 仍计入 for real users

- **WHEN** 经 gateway-app 的 `POST /device/app/api/apple_login` 返回 HTTP 200 且为新真实用户
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 2xx 成功请求计入统计

- **WHEN** 真实用户经 gateway-app 的 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
- **THEN** 对应归一化 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 4xx 鉴权失败不计入

- **WHEN** 经 gateway-app 的请求返回 HTTP 401
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: WebSocket 升级不计入

- **WHEN** 客户端对 `/voice/chat/ws` 发起 WebSocket 升级
- **THEN** 系统 SHALL NOT 写入 HTTP 使用统计

### Requirement: API paths SHALL be normalized and annotated with Chinese summary

系统 MUST 在启动时自 `api/v1` 路由元数据构建注册表，将动态路径段归一化为与 `g.Meta path` 一致的模板（如 `/ucg/app/api/posts/{id}`）。管理端展示的 `summary` MUST 取自注册表中文说明；未命中注册表时 SHALL 显示「未登记」并保留归一化或原始 apiKey。

#### Scenario: 动态帖子 ID 归一化

- **WHEN** 统计 `GET /ucg/app/api/posts/42` 的成功调用
- **THEN** 聚合键 SHALL 为 `GET /ucg/app/api/posts/{id}` 且 summary SHALL 为「获取单帖」或注册表中等价中文

#### Scenario: 未登记路径

- **WHEN** 某成功请求路径无法匹配任何 `api/v1` 模板
- **THEN** 列表项 summary SHALL 为「未登记」且仍 SHALL 展示 apiKey

### Requirement: Usage counters SHALL be stored in Redis with daily buckets

统计存储 MUST 使用 gateway-app 可访问的 Redis。全局 API 计数 MUST 按日分桶（`YYYYMMDD`）使用 INCR；当且仅当 `wxId > 0` 时 MUST 额外写入用户维度日计数。系统 MUST 维护全局与用户维度的 `last_at`（最近一次成功调用的 Unix 秒）。日桶 key MUST 设置 TTL（不少于 90 天）以支持近 30 天查询。

#### Scenario: 登录用户双维度计数

- **WHEN** `wxId=1001` 的用户成功调用 `POST /ucg/app/api/posts`
- **THEN** 全局 apiKey 日计数与用户 `wx:1001` 维度日计数 SHALL 各增加 1

#### Scenario: 无 wxId 仅全局计数

- **WHEN** 请求可解析 `deviceNo` 但 `wxId=0` 且返回 2xx
- **THEN** 全局 apiKey 日计数 SHALL 增加 1 且 SHALL NOT 写入用户维度键

### Requirement: Admin usage list API SHALL return API frequency for a time window

`GET /device/admin/api/usage/list` MUST 要求有效 **Admin JWT**（`Authorization: Bearer`，`aud=gateway-admin`）。查询参数 `days` 默认 `7`；`days=0` 表示聚合 TTL 内全部日桶。响应 MUST 包含 `list` 数组，每项至少含 `apiKey`、`summary`、`count`（窗口内合计）、`lastAt`（Unix 秒）。查询参数 `sortBy` 默认 `count`；`sortBy=lastAt` 时列表 SHALL 按 `lastAt` 降序，否则 SHALL 按 `count` 降序。当 Redis 日桶 Hash 中存在 field 时，读路径 MUST 经 `cachekit.HashGetAll` 正确解析 GoFrame Redis 返回值（含 `HGETALL` 经 adapter 转为 flat `[]string` 的情形）并 SHALL NOT 因类型解析失败返回空列表。

#### Scenario: Redis 有数据时列表非空

- **WHEN** Redis 键 `gw:usage:d:{today}:g` 的 Hash 含至少一个 apiKey field，且管理员携带有效 Admin JWT 请求 `days=7`
- **THEN** 响应 `list` SHALL 包含对应 apiKey 且 `count > 0`

#### Scenario: GoFrame HGETALL flat []string 可读

- **WHEN** `cachekit.HashGetAll` 底层收到 GoFrame adapter 表示为 flat `[]string`（非 map）的 `HGETALL` 结果
- **THEN** `ListAPIs` 等读路径 SHALL 仍正确聚合 field 与计数，SHALL NOT 返回空 `list`

#### Scenario: 默认近 7 天

- **WHEN** 管理员携带有效 Admin JWT 请求列表且未传 `days`
- **THEN** 每项 `count` SHALL 为最近 7 个自然日 INCR 之和

#### Scenario: 无效或缺失 token

- **WHEN** 请求未携带有效 Admin JWT
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回统计数据

### Requirement: Admin usage detail API SHALL list wxId callers for one API

`GET /device/admin/api/usage/detail` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `apiKey`；可选 `days`（默认 7）、`sortBy`（`count` 或 `lastAt`）。响应 `list` 每项 SHALL 至少包含 `wxId`、`nickname`、`count`、`lastAt`。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: 按 API 下钻 wxId

- **WHEN** 管理员携带 Admin JWT 查询 `apiKey=GET /ucg/app/api/feed/recommend` 且 `days=7`
- **THEN** 响应 SHALL 列出窗口内调用过该 API 的 wxId、展示昵称、次数与最近调用时间

### Requirement: Admin usage user API SHALL list APIs called by one wxId

`GET /device/admin/api/usage/user` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `wxId`（正整数）；可选 `days`（默认 7）。响应 `list` 每项 SHALL 至少包含 `apiKey`、`summary`、`count`、`lastAt`。

#### Scenario: 按 wxId 查看 API 分布

- **WHEN** 管理员携带 Admin JWT 查询 `wxId=1001` 且 `days=7`
- **THEN** 响应 SHALL 列出该用户在窗口内成功调用过的 API、中文说明、次数与最近调用时间

### Requirement: Usage read APIs SHALL be served by gateway-app and excluded from device proxy

路径 `/device/admin/api/usage/*` MUST 由 gateway-app 本机处理，MUST NOT 被 `device-service` 反代吞掉。统计计数读路径（`usage/list`、`usage/detail`、`usage/user`）MUST 从 gateway Redis 读取，MUST NOT 直连 device/history/voice/ucg **数据库**。展示字段 enrich（`usage/wx-list` 的 wx 列表与昵称、`usage/detail` 的 nickname）MAY 经 **HTTP 内部契约**调用 device-service 与 ucg-service，MUST NOT 直连他域 DAO 或库表。

#### Scenario: 读 API 不经 device-service 反代

- **WHEN** 管理员请求 `GET /device/admin/api/usage/list`
- **THEN** gateway-app SHALL 本地响应且 SHALL NOT 将请求转发至 device-service

#### Scenario: Wx-list orchestrates via HTTP contracts

- **WHEN** 管理员请求 `GET /device/admin/api/usage/wx-list`
- **THEN** gateway-app SHALL 本地处理该请求
- **AND** MUST 经 HTTP 契约获取 wx 列表与 nickname，MUST NOT 直连 MySQL

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。页面 MUST 使用 Hub Admin JWT（`admin-common.js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号；HTTP 统计不含 WebSocket 与维护型接口（token 刷新、版本检查等）。页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。「按用户」视图 MUST 调用 `GET /device/admin/api/usage/wx-list`（MUST NOT 直接调用 `wx/list`）。「按用户」wx 表格 MUST 展示列：wxId、UCG 昵称、注册时间（`createdAt=0` 显示「—」）、设备号、unionid、平台、用户名账号。API 下钻用户表 MUST 展示列：wxId、UCG 昵称、次数、最近调用。

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户从 wx 列表点选

- **WHEN** 管理员切换到「按用户」视图
- **THEN** 页面 SHALL 展示非模拟 wx 账号列表且含 UCG 昵称与注册时间
- **AND** 点选某 wxId 后 SHALL 展示该用户的 API 调用列表

#### Scenario: API 下钻展示昵称

- **WHEN** 管理员在「按 API」视图对某 API 点击下钻
- **THEN** 下钻用户表 MUST 展示每行 wxId 的 UCG 昵称

### Requirement: device admin SHALL link to API usage stats from device record card

`resource/public/admin.html` 的**设备记录**卡片头部 `card-actions` MUST 包含指向 `/device/admin/api-usage-stats` 的链接，文案为「功能使用统计」（或等价中文）。链接在管理员登录成功后 SHALL 可见（与问答库/反馈「展开更多」一致的显示时机）。

#### Scenario: 设备记录区入口

- **WHEN** 管理员登录设备管理页并进入主界面
- **THEN** 设备记录卡片 `card-actions` SHALL 显示「功能使用统计」链接

### Requirement: 新增 App HTTP 接口 MUST 经负责人确认是否计入 usage 统计

当 OpenSpec 变更或实现工作 **新增** 经 gateway-app 对外的 App HTTP 路由（`api/v1` 的 `g.Meta` 或 gateway-app `BindHandler`，不含已结构性 skip 的 internal/admin/static/WS）时，执行方（含 AI）**MUST 向产品负责人询问**该接口是否计入 App API 使用统计；负责人未明确答复前 **MUST NOT** 擅自将其加入或移出 `maintenance_skip.go` denylist。proposal 或 tasks **SHALL** 记录确认结论。

#### Scenario: 新增业务 API 且负责人要求统计

- **WHEN** 变更新增 `POST /ucg/app/api/foo` 且负责人确认需要统计
- **THEN** 实现 SHALL 在 `api/v1` 登记路由且 SHALL NOT 写入 `maintenance_skip.go`

#### Scenario: 新增维护型 API 且负责人要求不统计

- **WHEN** 变更新增维护型接口且负责人确认不统计
- **THEN** 实现 SHALL 在 `maintenance_skip.go` 增加对应排除规则并在 proposal/tasks 中说明

### Requirement: Admin usage wx-list API SHALL return enriched wx rows for usage stats UI

`GET /device/admin/api/usage/wx-list` MUST 要求有效 Admin JWT。查询参数 MUST 与 `GET /device/admin/api/wx/list` 对齐：`page`、`pageSize`、`q`（均可选，默认分页语义一致）。gateway-app MUST 经 device 契约取得非模拟 wx 分页列表（含 `createdAt`），并对当前页 wxIds 经 ucg internal `profiles/batch` 补充 `nickname`。响应 MUST 含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`、`createdAt`、`nickname`。

#### Scenario: Usage wx-list returns nickname and createdAt

- **WHEN** 管理员携带 Admin JWT 请求 `GET /device/admin/api/usage/wx-list?page=1&pageSize=20`
- **THEN** 响应 `list` 每项 MUST 含 `nickname` 与 `createdAt`
- **AND** MUST NOT 含 `is_simulated=1` 的 wxId

#### Scenario: Usage wx-list excludes simulated from total

- **WHEN** 库中存在模拟用户且存在真实用户
- **THEN** `total` MUST 仅统计 `is_simulated=0` 的行数

### Requirement: Admin usage detail list items SHALL include display nickname

`GET /device/admin/api/usage/detail` 响应 `list` 每项除 `wxId`、`count`、`lastAt` 外 MUST 含 `nickname`（string）。`nickname` MUST 经 ucg internal `profiles/batch` 取得，语义与 batch 变更一致（含无 profile 时的推导默认昵称）。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: Detail includes nickname per wxId

- **WHEN** 管理员携带 Admin JWT 查询某 `apiKey` 且下钻列表含 wxId=1001
- **THEN** 对应项 MUST 含非空 `nickname` 或推导默认昵称（如「家长」）

---

## gateway-app-cors

<!-- source: openspec/specs/gateway-app-cors/spec.md -->

# gateway-app-cors Specification

## Purpose
TBD - created by archiving change gateway-app-cors-ip-allowlist. Update Purpose after archive.
## Requirements
### Requirement: App 网关按主机白名单回显 CORS Origin

`gateway-app-server` 对浏览器跨域请求 SHALL 在响应中包含 CORS 头。当且仅当请求头 `Origin` 解析成功且其主机（不含端口比较 IP 字面量，含端口时取 hostname）等于 `192.168.0.131` 或 `120.55.50.105`，且 scheme 为 `http` 或 `https` 时，SHALL 将 `Access-Control-Allow-Origin` 设为该 `Origin` 的完整原始值（回显），从而允许该主机上任意端口的 Web 来源。

#### Scenario: 匹配内网 IP 任意端口

- **WHEN** 请求包含 `Origin: http://192.168.0.131:5173` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应包含 `Access-Control-Allow-Origin: http://192.168.0.131:5173`

#### Scenario: 匹配公网 IP 任意端口

- **WHEN** 请求包含 `Origin: https://120.55.50.105:8443` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应包含 `Access-Control-Allow-Origin: https://120.55.50.105:8443`

#### Scenario: 非白名单主机不回显

- **WHEN** 请求包含 `Origin: https://evil.example` 且请求到达 `gateway-app-server` 的 HTTP 处理链
- **THEN** 响应 SHALL NOT 设置 `Access-Control-Allow-Origin`（或不得回显该 Origin）

### Requirement: CORS 方法与请求头

`gateway-app-server` SHALL 在 CORS 响应中声明允许方法包含 `GET`、`POST`、`OPTIONS`，并 SHALL 在 `Access-Control-Allow-Headers`（或对预检的等价响应）中允许 `Content-Type` 与 `Authorization`，以满足常见 JSON 与 Bearer 联调。

#### Scenario: 预检请求获得方法与头

- **WHEN** 浏览器发送 `OPTIONS` 预检，且 `Origin` 通过主机白名单校验，且带有 `Access-Control-Request-Method: POST` 与 `Access-Control-Request-Headers: content-type, authorization`
- **THEN** 响应状态码为成功（2xx），且包含允许上述方法与头的 CORS 响应头（具体头名大小写遵循实现，语义须满足浏览器识别）

### Requirement: 预检不破坏既有鉴权豁免

对 `OPTIONS` 请求的 Bearer 豁免行为 SHALL 保持与变更前一致：预检请求 MUST NOT 因缺少 Bearer 被拒绝（例如 401）。

#### Scenario: OPTIONS 无 Bearer 仍成功预检

- **WHEN** `OPTIONS` 请求指向需鉴权的 API 路径，且无 `Authorization` 头，但 `Origin` 通过白名单
- **THEN** 响应 SHALL NOT 仅因缺少 Bearer 而返回 401（允许返回 204 或其它 2xx 并完成 CORS 头）

---

## gateway-app-cors-reverse-proxy

<!-- source: openspec/specs/gateway-app-cors-reverse-proxy/spec.md -->

# gateway-app-cors-reverse-proxy Specification

## Purpose
TBD - created by archiving change gateway-app-cors-proxy-history-api. Update Purpose after archive.
## Requirements
### Requirement: 反向代理响应在允许来源下须带齐 CORS 头

对 `gateway-app-server` 上经 `httputil.ReverseProxy` 转发至下游的 **`/device/history/api/*`** 请求：当请求头 `Origin` 经 `ReflectGatewayAppCORSOrigin` 判定为允许（`ok == true`）时，返回给客户端的最终响应 **MUST** 包含 `Access-Control-Allow-Origin`（值为该 Origin 的回显）、`Access-Control-Allow-Methods`、`Access-Control-Allow-Headers`、`Access-Control-Max-Age`，其语义 **MUST** 与同一进程内直连 `gateway_app_cors` 中间件对同源校验通过请求所写入的头一致（若实现抽取为共享函数，则以该函数为准）。

#### Scenario: 带 Authorization 的 GET 列表在代理命中时可通过浏览器 CORS

- **WHEN** 客户端为浏览器且发送 `GET /device/history/api/list?...`，带 `Origin: http://localhost:58912`（或任一当前白名单允许的 Origin），带 `Authorization: Bearer <token>`，且该请求在网关内 **命中** history 反向代理并成功自下游取得 2xx 或业务约定的 HTTP 状态
- **THEN** 最终响应 **MUST** 包含 `Access-Control-Allow-Origin` 且值等于请求 `Origin`（回显），并 **MUST** 包含与直连 API 一致的 `Access-Control-Allow-Methods` 与 `Access-Control-Allow-Headers`（至少涵盖 `GET, POST, OPTIONS` 与 `Content-Type, Authorization` 语义）

#### Scenario: 不允许的 Origin 不在代理响应中伪造 CORS

- **WHEN** 请求命中 history 反向代理且 `Origin` 未通过 `ReflectGatewayAppCORSOrigin`
- **THEN** 网关 **MUST NOT** 为通过该策略而添加 `Access-Control-Allow-Origin`（避免对非白名单来源误放行）

### Requirement: voice/device 代理与 history 行为一致（若共用构建函数）

若 `voice`、`device` 领域 HTTP 代理与 history 共用同一 `buildReverseProxy` 或同一套 CORS 注入扩展点，则对上述代理路径在相同 Origin 规则下 **MUST** 适用与 history 相同的 CORS 注入语义，除非规格或设计文档显式排除某路径。

#### Scenario: 共用构建函数时代理路径一致

- **WHEN** 实现将 CORS 注入接在共用的 ReverseProxy 构建路径上，且请求命中该代理
- **THEN** 允许来源下的 CORS 响应头行为 **MUST** 与 history 场景等效，避免出现「仅 app 直连有 CORS、代理无 CORS」的分裂

---

## gateway-app-device-login-device-no

<!-- source: openspec/specs/gateway-app-device-login-device-no/spec.md -->

# gateway-app-device-login-device-no Specification

## Purpose
TBD - created by archiving change gateway-app-device-login-return-device-no. Update Purpose after archive.
## Requirements
### Requirement: 设备号业务登录响应须含非空 deviceNo

对 **POST `/device/app/api/user/device_login`**（device-service，业务成功 `code=0`），响应 `data` **MUST** 包含 JSON 字段 **`deviceNo`**，且为 **trim 后非空字符串**，表示本次完成登录校验的设备号（与请求入参在规范化后一致，或与库内绑定到该会话的权威 `device_no` 一致）。

#### Scenario: 纯设备会话（无 wx 行绑定）

- **WHEN** 客户端提交已注册设备号且 wx 表无对应行，业务校验成功
- **THEN** 响应 `data.deviceNo` **MUST** 为非空字符串，且与本次登录所用设备号一致

### Requirement: 网关聚合设备号登录响应须含非空 deviceNo

对 **POST `/device/app/api/device_login`**（gateway-app-server，业务成功 `code=0`），响应 `data` **MUST** 包含 **`deviceNo`**，且为 **trim 后非空字符串**。若下游 device 返回的 `data.deviceNo` 为空或缺失，网关 **SHALL** 使用本次请求体中的 `deviceNo`（trim 后）作为回包与 JWT 签发所用设备号；若兜底后仍为空，**MUST** 拒绝成功语义（沿用现有参数/内部错误路径）。

#### Scenario: 下游 data 缺 deviceNo 时网关兜底

- **WHEN** device `device_login` 返回 `code=0` 但 `data` 中无 `deviceNo` 或值为空白，且请求体 `deviceNo` 经 trim 后非空
- **THEN** 网关返回的聚合响应 `data.deviceNo` **MUST** 等于该 trim 后的请求 `deviceNo`，且签发的 access/refresh 所绑定的设备号与该值一致

#### Scenario: 请求与下游均无可用设备号

- **WHEN** 请求体 `deviceNo` trim 后为空，或业务失败
- **THEN** 网关 **MUST NOT** 返回 `code=0` 且带非空 `deviceNo` 的成功形态（保持现有错误语义）

---

## gateway-app-jwt-device-no-header

<!-- source: openspec/specs/gateway-app-jwt-device-no-header/spec.md -->

# gateway-app-jwt-device-no-header Specification

## Purpose
TBD - created by archiving change gateway-app-jwt-device-no-header. Update Purpose after archive.
## Requirements
### Requirement: access JWT SHALL 同时携带 wx 主键与 device_no 声明

gateway-app-server 签发的 **access_token（JWT）SHALL** 使用标准 **`sub`** claim 承载 **`wx` 表主键 id**（十进制字符串，与现网 refresh 语义一致）；并 **SHALL** 包含 **`device_no` 私有声明**（与 `ai_voice_device.wx.device_no` 语义一致）。当用户尚未绑定设备时，`device_no` 声明 **MAY** 为空或省略，其实现策略 **MUST** 在实现与评审中保持唯一且文档化。

#### Scenario: 已绑定设备用户登录后拿到 access

- **WHEN** device 在 `POST /device/app/api/user/login` 返回的 `deviceNo` 非空且网关签发 access
- **THEN** JWT **MUST** 可被解析为包含非空的 **`device_no` 声明** 与与 `wxId` 一致的 **`sub`**

### Requirement: Bearer 中间件 SHALL 注入 Wx-Id 与可选 Device-No 头且不再拉取 unionid

gateway-app-server 对非白名单 HTTP 请求在校验 access JWT 成功后，**SHALL** 设置 **`X-Internal-Wx-Id`** 为 **`sub`** 所表示的整数 wx 主键（字符串形式与头规范在实现中固定）；**SHALL** 在 **`device_no` 声明非空** 时设置 **`X-Internal-Device-No`** 为该值。**MUST NOT** 为完成上述注入而调用 device-service 的 **`GET /device/app/api/user/internal/by-id`**（即 **禁止** 将「id→unionid」作为网关热路径依赖）。

#### Scenario: 受保护 HTTP 请求鉴权通过

- **WHEN** 客户端携带合法 Bearer access JWT
- **THEN** 发往 device/history/voice 等下游的代理请求 **MUST** 携带 **`X-Internal-Wx-Id`**；且当 JWT 含非空 **`device_no` 声明** 时 **MUST** 携带 **`X-Internal-Device-No`**

#### Scenario: 对外 HTTP 契约保持不变

- **WHEN** App 调用 `POST /device/app/api/login` 或 `POST /device/app/api/token/refresh`
- **THEN** 请求与响应 JSON **MUST** 保持与变更前一致的字段名与客户端可见语义（客户端 **MUST NOT** 需要解析 JWT 载荷即可集成）

### Requirement: device-service 用户域 SHALL 以 X-Internal-Wx-Id 识别 wx 行

device-service 对 **`POST /device/app/api/user/bindwx`**、**`POST /device/app/api/user/auto_save`**、**`GET /device/app/api/user/detail`** 等依赖「当前登录 wx」的接口，**SHALL** 从请求头 **`X-Internal-Wx-Id`** 读取 wx 主键并定位 `wx` 行；**MUST NOT** 将 **`X-Internal-Wx-Union-Id`** 作为网关受信任路径的必需依赖（若保留兼容，**MUST** 在部署文档中声明过渡期与移除时间）。

#### Scenario: bindwx 成功

- **WHEN** 请求携带合法 **`X-Internal-Wx-Id`** 且 body 中 `deviceNo` 合法
- **THEN** 系统 SHALL 完成绑定并返回成功语义

### Requirement: 历史 WebSocket SHALL 使用 JWT device_no 声明与首帧 device_no 校验

gateway-app-server 的历史 WebSocket 在首帧 `auth` 后，**SHALL** 校验 access JWT 的 **`device_no` 声明** 与首帧 JSON 中的 **`device_no`（或 `deviceNo`，以实现为准且单一）** 一致（在声明非空时）；**MUST NOT** 依赖「unionid → detail 拉 device_no」链完成该校验。

#### Scenario: 认证成功

- **WHEN** JWT 有效且 **`device_no` 声明** 与首帧设备号一致
- **THEN** 连接 SHALL 注册到对应 `device_no` 的推送组

#### Scenario: 认证失败

- **WHEN** JWT 有效但设备号不一致或声明缺失导致无法满足校验策略
- **THEN** 服务端 SHALL 拒绝订阅并 SHALL NOT 将连接加入推送组

### Requirement: refresh 重新签发的 access SHALL 同步 device_no 声明

gateway-app-server 在处理 **`POST /device/app/api/token/refresh`** 时，**SHALL** 在签发新 access JWT 时写入 **与当前 wx 会话权威一致的 `device_no` 声明**（以 device-service 返回或网关侧明确规则为准），以避免换绑后长期持有错误 `device_no` claim 的策略 **MUST** 在 design 的 D5 中落地为单一实现。

#### Scenario: 刷新成功

- **WHEN** refresh_token 有效且旋转策略允许签发新 access
- **THEN** 新 access JWT **MUST** 包含更新后的 **`device_no` 声明**（若设备域当前已绑定）

---

## gateway-app-official-site

<!-- source: openspec/specs/gateway-app-official-site/spec.md -->

# gateway-app-official-site Specification

## Purpose
TBD - created by archiving change pangbao-official-site-homepage. Update Purpose after archive.
## Requirements
### Requirement: Gateway-app 根路径承载胖宝官网
`gateway-app-server` SHALL 在根路径 `/` 返回“胖宝”官网 HTML，而不是当前纯文本“智能语音 App 网关”。该官网路由变更 MUST 仅作用于 `gateway-app-server` 进程，MUST NOT 改变主网关或其他微服务进程的根路径行为。

#### Scenario: 访问 gateway-app 根路径
- **WHEN** 浏览器对 `gateway-app-server` 发起 `GET /`
- **THEN** 系统 SHALL 返回官网 HTML 页面，页面标题与主视觉 SHALL 展示品牌名“胖宝”

#### Scenario: 官网替换不扩散到主网关
- **WHEN** 本次变更部署完成
- **THEN** 系统 SHALL 仅修改 `gateway-app-server` 的根路径处理逻辑，主网关进程的路由与代理行为 MUST 保持不变

### Requirement: 官网展示母婴喂养定位与事件卡片
官网页面 SHALL 以玻璃拟态风格展示品牌定位文案，并 SHALL 展示从数据库权威链路读取的事件列表。每个事件项 MUST 至少包含事件名与事件 logo；若 logo 为 path-only 资源，前端或聚合接口 MUST 能将其解析为当前站点可访问的同源地址。除喂养事件卡片外，页面 SHALL 包含独立的产品能力区块（见「官网展示三大产品能力柱」），首屏定位 MUST 覆盖 0–18 月新手妈妈的喂养、轻社交与 AI 价值。

#### Scenario: 官网首屏展示品牌定位
- **WHEN** 用户打开官网首页
- **THEN** 页面 SHALL 明确表达面向 0–18 月新手妈妈（或等价阶段化表述），以及「喂养记录更便捷、同月龄轻社交、AI 智能陪伴」等核心信息，而不仅限于单一喂养工具定位

#### Scenario: 官网展示事件 logo 与事件名
- **WHEN** 官网聚合到至少一条事件数据
- **THEN** 页面 SHALL 为每条事件渲染事件 logo 与事件名，且 logo 地址 MUST 可被当前官网域名直接访问

#### Scenario: 喂养事件区说明对齐小月龄场景
- **WHEN** 用户查看 `#events` 喂养事件区块
- **THEN** 区块说明文案 SHOULD 提及小月龄高频照护场景（如吃奶、睡觉、换尿布）及语音记录等等价表述

### Requirement: 官网提供匿名只读聚合数据接口
系统 SHALL 提供一个适用于官网匿名访问的只读聚合接口，由 `gateway-app-server` 统一返回官网所需的事件展示数据、Android 下载信息与 iOS 下载说明。该接口 MUST 通过服务契约或本进程已有能力获取数据，MUST NOT 让前端直接调用受保护业务接口或跨服务直连数据库。接口返回的 `heroTitle`、`heroSubtitle`、`serviceSummary` MUST 与官网 0–18 月新手妈妈及「喂养 + 轻社交 + AI」三角定位一致。

#### Scenario: 匿名读取官网数据
- **WHEN** 未登录用户请求官网聚合接口
- **THEN** 系统 SHALL 返回成功响应，其中包含事件列表、Android 下载展示信息、iOS 下载说明，以及已更新语义的 Hero 文案字段

#### Scenario: 官网数据来源遵守服务边界
- **WHEN** `gateway-app-server` 组装官网响应
- **THEN** 系统 MUST 通过现有服务契约读取事件权威数据，并复用本进程版本信息读取能力，MUST NOT 新增跨服务直连他域库表行为

### Requirement: 官网展示 Android 下载二维码与 iOS 指引
官网 SHALL 提供独立的应用下载区块。Android 下载区 MUST 基于数据库中的最新下载链接生成二维码并展示可点击下载入口；iOS 下载区 MUST 提示用户前往 App Store 搜索“胖宝”下载。

#### Scenario: Android 存在可下载版本
- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 可归一化为有效路径
- **THEN** 官网聚合接口 SHALL 返回官网可直接使用的 Android 下载地址，页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: Android 暂无可下载版本
- **WHEN** 版本表没有可用的 Android 下载记录
- **THEN** 页面 SHALL 不展示失效二维码，并 SHALL 展示明确的“Android 下载暂未开放”或等价提示

#### Scenario: iOS 下载说明固定展示
- **WHEN** 用户查看官网下载区
- **THEN** 页面 SHALL 展示“前往 App Store 搜索‘胖宝’下载”的文案，而不要求数据库提供 iOS 下载链接

### Requirement: 官网展示三大产品能力柱
官网页面 SHALL 在 Hero 与喂养事件区块之间展示独立的「产品能力」区块（锚点 `#capabilities`），以玻璃拟态卡片呈现三大能力柱：**智能喂养记录**、**同月龄轻社交**、**AI 智能陪伴**。每柱 MUST 至少包含标题与 2 条场景化说明；移动端 MUST 可正常阅读（单列或等价布局）。

#### Scenario: 用户浏览产品能力区块
- **WHEN** 用户打开官网首页并滚动至 `#capabilities`
- **THEN** 页面 SHALL 展示上述三柱，且文案 MUST 涵盖语音记事件、同阶段妈妈社区（发帖/关注/私信）、AI 语音助手与智能喂养问答等等价能力描述

#### Scenario: 导航可跳转至产品能力
- **WHEN** 用户点击导航中的「产品能力」或等价链接
- **THEN** 页面 SHALL 滚动至 `#capabilities` 锚点

### Requirement: 官网面向 0-18 月新手妈妈定位
官网首屏（Hero）SHALL 明确面向 **0–18 月新手妈妈** 或等价表述（如 kicker 含「0–18 月」）。Hero MUST 同时传达「喂养记录」「同月龄连接」「AI 问答」三类价值，而非仅喂养工具定位。

#### Scenario: 首屏展示阶段化定位
- **WHEN** 用户打开官网首页
- **THEN** Hero 区域 SHALL 包含 0–18 月或新手妈妈阶段化定位文案，且副标题或 metric 卡片 MUST 提及社交与 AI 能力

#### Scenario: SEO meta 覆盖核心关键词
- **WHEN** 爬虫或浏览器读取页面 `<head>`
- **THEN** `<title>` 与 `<meta name="description">` MUST 包含喂养记录相关表述，并 SHOULD 包含 0–18 月、AI、同月龄或社区等关键词中的至少两项

### Requirement: 官网 AI 能力合规表述
官网展示 AI 能力时 SHALL 使用「智能喂养问答」或等价非医疗措辞，MUST NOT 单独使用「诊疗」作为官网主宣传语。页面 MUST 包含简短免责声明，说明 AI 回答仅供参考、不能替代医生诊断。

#### Scenario: AI 能力柱不含诊疗主宣传
- **WHEN** 用户阅读 `#capabilities` 中 AI 相关文案
- **THEN** 页面 SHALL NOT 以「胖宝诊疗」作为唯一或主标题对外展示，SHALL 使用智能问答/AI 助手等表述

#### Scenario: 展示 AI 免责声明
- **WHEN** 用户浏览含 AI 描述的区块或页脚
- **THEN** 页面 SHALL 展示等价于「AI 回答仅供参考，不能替代医生诊断」的免责声明

### Requirement: 官网 Footer 法律页链接
官网页脚 SHALL 提供指向隐私政策与用户协议的可见链接，分别对应 `/privacy-policy.html` 与 `/user-agreement.html`（或 gateway-app 注册的等价路由）。

#### Scenario: 页脚可访问法律文档
- **WHEN** 用户查看官网页脚
- **THEN** 页面 SHALL 展示可点击的「隐私政策」与「用户协议」链接，且链接 MUST 在同源 gateway-app 域名下可访问

---

## gateway-app-path-only-assets

<!-- source: openspec/specs/gateway-app-path-only-assets/spec.md -->

# gateway-app-path-only-assets Specification

## Purpose
TBD - created by archiving change device-event-logo-and-path-only-assets. Update Purpose after archive.
## Requirements
### Requirement: APK download_url SHALL 仅存应用内路径

gateway-app-server 在 APK 上传写库时，`app_version.download_url` MUST 仅存以 `/` 开头的路径，格式为 `/device/app/apk/<filename>.apk`；MUST NOT 将 `publicBaseUrl` 或任何域名写入该列（新写入）。

#### Scenario: 上传 APK 写库为路径

- **WHEN** 管理员成功上传 APK 并完成数据库 Insert
- **THEN** `download_url` MUST 等于 `/device/app/apk/` 加安全文件名
- **AND** 上传接口 JSON 响应中的 `downloadUrl` MUST 为同一路径（非绝对 URL）

#### Scenario: 上传不再因缺少 publicBaseUrl 拒绝写库

- **WHEN** `gatewayApp.publicBaseUrl` 未配置
- **AND** 上传文件与其它表单字段合法
- **THEN** 服务端 MUST 仍能完成落盘与数据库写入（路径存库）

### Requirement: 版本检查接口 SHALL 返回 path 型 downloadUrl

`GET /device/app/api/version/check` 响应中的 `downloadUrl` MUST 为应用内路径（新数据）；若库内仍为历史绝对 URL，服务端 MUST 在返回前归一化为路径（仅保留 path 部分）。

#### Scenario: 版本检查返回路径供客户端拼接

- **WHEN** 版本表存在最新行且 `download_url` 已按 path 存储
- **THEN** `downloadUrl` MUST 形如 `/device/app/apk/xxx.apk`
- **AND** MUST NOT 以 `http://` 或 `https://` 开头

### Requirement: gateway-app SHALL 代理事件 logo 静态路径

gateway-app-server MUST 注册 `GET /ai_talk_images/*`，将请求转发至 device-service 同路径（或等价静态源），使客户端可通过 gateway-app 端口（如 `:9702`）访问 `https://<host>:9702/ai_talk_images/...` 而无需直连 device 端口。

#### Scenario: 经 gateway-app 访问事件 logo

- **WHEN** 客户端请求 gateway-app 的 `GET /ai_talk_images/event_1.png`
- **AND** device-service 上对应文件存在
- **THEN** gateway-app MUST 返回成功图片响应（经反代或共享存储）

### Requirement: APK 下载路径契约保持不变

既有 `GET /device/app/apk/*filename` 下载处理器 MUST 继续从 `apkStorageDir` 提供文件；与 path-only `download_url` 组合后，客户端完整下载地址为 `<gateway-app-base>` + `downloadUrl`。

#### Scenario: path 与下载路由一致

- **WHEN** `download_url` 为 `/device/app/apk/foo.apk`
- **THEN** 对 gateway-app 发起 `GET /device/app/apk/foo.apk` MUST 可下载该文件

---

## gateway-app-server

<!-- source: openspec/specs/gateway-app-server/spec.md -->

# gateway-app-server Specification

## Purpose
TBD - created by archiving change gateway-app-server-wx-auth-history-ws. Update Purpose after archive.
## Requirements
### Requirement: App 网关进程独立运行

系统 SHALL 提供名为 gateway-app-server 的独立 HTTP 服务进程，具备与现有 gateway 相当的静态资源与领域反向代理能力，并额外承载 App 鉴权、令牌、版本检查、历史 WebSocket，以及 **UCG HTTP 反向代理**（`/ucg/app/api/*` → ucg-service）、**UCG 聊天 WebSocket 升级代理**（`/ucg/app/ws/chat` → ucg-service `/ws/chat`），以及 **voice HTTP 反向代理**（`/voice/app/api/*`、`/voice/admin/api/*` → voice-service）。App 对外 UCG 与 voice App/Admin 流量 MUST 经本进程暴露，与现有 App API 同域。

#### Scenario: 进程启动与配置隔离

- **WHEN** 使用 gateway-app-server 专用配置文件启动进程
- **THEN** 服务 SHALL 仅加载该进程所需的数据库分组（含 ai_voice_app）与下游 URL 配置（含 `UCG_SERVICE_BASE_URL`、`UCG_WS_PROXY_URL`、**`VOICE_API_PROXY_URL`**），且 SHALL NOT 将 voiceChat 等业务配置错误合并到错误进程的权威配置源中（遵循仓库既有配置边界约定）

#### Scenario: UCG HTTP 代理可用

- **WHEN** 配置 `UCG_SERVICE_BASE_URL` 且 ucg-service 健康
- **THEN** 对 `/ucg/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 ucg-service

#### Scenario: voice HTTP 代理可用

- **WHEN** 配置 `VOICE_API_PROXY_URL` 且 voice-service 健康
- **THEN** 对 `/voice/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 voice-service
- **AND** 对 `/voice/admin/api/*` 的请求 SHALL 经 Admin JWT 校验与 `VOICE_ADMIN_PASSWORD` 注入后转发至 voice-service

#### Scenario: UCG 聊天 WS 升级代理可用

- **WHEN** 配置 `UCG_WS_ROUTE_MODE=proxy` 且 `UCG_WS_PROXY_URL` 指向可达的 ucg-service `/ws/chat`
- **AND** 客户端对 `/ucg/app/ws/chat` 发起 WebSocket Upgrade
- **THEN** gateway-app SHALL 将握手与后续双向帧透传至 ucg-service，行为与 `ws_route_proxy.go` voice WS 透传一致

### Requirement: Bearer 鉴权与内部头注入
系统 SHALL 对除白名单外的受保护 HTTP 路径校验 `Authorization: Bearer <access_token>`，其中 `access_token` MUST 为合法 JWT。系统 SHALL 在校验签名与过期时间后，从 `sub` 解析 `wx.id`（允许 `0` 表示纯设备会话）并向下游注入 **`X-Internal-Wx-Id`**；当 access 含 `device_no` 声明时，系统 SHALL 同步注入 **`X-Internal-Device-No`**。

#### Scenario: 鉴权通过并注入头
- **WHEN** Bearer 为合法未过期 JWT，且 `sub` 与 `device_no` 组合满足会话规则
- **THEN** 网关 SHALL 设置 `X-Internal-Wx-Id`，并在有值时设置 `X-Internal-Device-No`

#### Scenario: 鉴权失败
- **WHEN** Bearer 缺失、签名错误、已过期或会话字段非法
- **THEN** 网关 SHALL 返回未授权错误，且 SHALL NOT 注入内部头

### Requirement: 登录与令牌仅由 gateway-app 签发
系统 SHALL 在 gateway-app-server 提供并维护两类聚合登录：
1. `POST /device/app/api/login`（微信聚合登录，转发 device 微信业务登录）
2. 用户名聚合登录接口（路径位于 `/device/app/api/` 前缀下，转发 device 用户名登录业务接口）

两类聚合登录在成功后 SHALL 统一由 gateway 签发 access/refresh；access MUST 为 JWT，`sub` MUST 等于目标 `wx.id`；refresh SHALL 为不透明随机串并绑定 Redis 会话。

#### Scenario: 用户名聚合登录成功
- **WHEN** 客户端调用用户名聚合登录且 device 返回有效 `wxId`
- **THEN** 网关 SHALL 返回 accessToken 与 refreshToken，且 access `sub` SHALL 等于该 `wxId`

#### Scenario: 微信聚合登录保持兼容
- **WHEN** 客户端调用既有微信聚合登录
- **THEN** 网关 SHALL 按现有语义返回 token 与业务字段，不因新增用户名能力破坏兼容性

### Requirement: 刷新令牌接口

系统 SHALL 在 gateway-app-server 提供刷新 access 的 HTTP 接口（路径位于 `/device/app/api/` 前缀下），使用 Redis 校验 refresh 后签发新的 **JWT** 形态 access_token（`sub`/`iat`/`exp` 规则与登录接口一致），并可按产品策略旋转 refresh_token。

#### Scenario: 刷新成功

- **WHEN** 客户端提交有效 refresh_token
- **THEN** 系统 SHALL 返回新的 access_token 且该 token SHALL 为合法 JWT，且旧 refresh 的处理策略（保留至过期或立即失效）SHALL 与设计文档一致并在实现中单一实现

### Requirement: 版本检查 API

系统 SHALL 在 gateway-app-server 提供 `GET /device/app/api/version/check`，从查询参数读取 `currentVersion`，读取 ai_voice_app.version 表（或经缓存的等价数据）并返回 needUpdate、latestVersion、releaseNotes、downloadUrl、forceUpdate。

#### Scenario: 返回版本信息

- **WHEN** 客户端携带合法 currentVersion 调用版本检查接口
- **THEN** 响应 SHALL 包含布尔 needUpdate 及 latestVersion、releaseNotes、downloadUrl、forceUpdate 字段，且 MAY 使用 Redis 缓存版本行以降低数据库压力

### Requirement: 历史 WebSocket 与首帧认证

系统 SHALL 在 gateway-app-server 提供 WebSocket 端点；连接建立后首条文本帧 MUST 为 JSON，包含 `type` 为 `auth`、`access_token`（snake_case 键名，值为 **JWT 字符串**）与 `device_no`；服务端 MUST 按与 HTTP Bearer 相同的规则校验 JWT 后，再校验 `sub` 对应 wx 身份与该 device_no 的绑定关系，通过后才将连接注册到按 device_no 分组的推送集合。

#### Scenario: 认证成功并订阅

- **WHEN** 客户端发送合法 auth 帧且 access_token 为有效 JWT、device_no 与该 token 身份匹配
- **THEN** 连接 SHALL 保持打开并能够接收后续由 Redis 通知触发的历史变更消息

#### Scenario: 认证失败

- **WHEN** auth 帧缺失、字段不合法或 device_no 与身份不匹配
- **THEN** 服务端 SHALL 拒绝订阅（关闭连接或发送错误文本帧）且 SHALL NOT 将该连接加入任何 device_no 推送组

### Requirement: Redis Pub/Sub 消费与下行

系统 SHALL 在 gateway-app-server 进程内维护对约定 Redis channel 的订阅；当收到 history-service 发布的消息时，SHALL 向所有已认证且匹配 `device_no` 的 WebSocket 连接推送 JSON 业务消息。

#### Scenario: 收到发布并推送

- **WHEN** Redis 收到一条包含已知 device_no 与历史载荷的合法通知
- **THEN** 网关 SHALL 向该 device_no 下已注册且仍存活的连接广播该消息体

### Requirement: 鉴权白名单
系统 SHALL 将无需 Bearer 的入口纳入白名单，至少包含：微信聚合登录、用户名聚合登录、refresh、公开版本检查（若启用）及 WebSocket 握手路径。

#### Scenario: 无令牌访问用户名登录
- **WHEN** 客户端无 Authorization 头调用用户名聚合登录接口
- **THEN** 请求 SHALL 进入对应处理器且 SHALL NOT 被 Bearer 中间件拦截

---

## gateway-app-version-admin

<!-- source: openspec/specs/gateway-app-version-admin/spec.md -->

# gateway-app-version-admin Specification

## Purpose
TBD - created by archiving change gateway-app-version-admin-apk-upload. Update Purpose after archive.
## Requirements
### Requirement: 版本管理页访问控制

gateway-app-server SHALL 提供「版本管理」相关 UI 与 API；在未通过 **Admin JWT**（`aud=gateway-admin`，由 Hub `POST /device/admin/api/login` 签发）鉴权前，SHALL NOT 暴露 APK 上传与写库能力。系统 MUST NOT 再使用独立版本管理口令、`POST /device/app/api/version/admin/login` 或 `gw_ver_admin` Cookie 会话作为鉴权手段（B1）。

#### Scenario: 无 Admin JWT 拒绝管理操作

- **WHEN** 客户端在未携带有效 Admin JWT 的情况下调用上传或写库接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 写入磁盘或数据库

#### Scenario: Admin JWT 校验通过允许操作

- **WHEN** 客户端携带由 Hub 登录签发的有效 Admin JWT
- **THEN** 系统 SHALL 允许后续受保护操作（上传、写库、列表等）

### Requirement: Android APK 落盘路径

系统 SHALL 将成功接收的 Android APK 保存至 **`/apk/ai_talk/`** 目录下；若该目录（或父路径中本进程可创建的部分）不存在，SHALL 使用等价于 `MkdirAll` 的方式创建后再写入。

#### Scenario: 目录不存在时自动创建

- **WHEN** 管理员已鉴权且上传合法 APK，且 `/apk/ai_talk/` 尚不存在但进程具备在路径上创建目录的权限
- **THEN** 系统 SHALL 创建目录并完成文件保存

### Requirement: 下载 URL 与数据库一致性

上传成功后，系统 SHALL 根据配置的对外 **`publicBaseUrl`**（或等价项）与约定的 **HTTP GET 下载路径规则** 生成 **完整绝对 URL**，并 SHALL 将该 URL 写入 `ai_voice_app.version` 表中对应新记录的 **`download_url`** 字段，且该 URL SHALL 指向已保存的 APK 文件。

#### Scenario: 客户端可下载

- **WHEN** 任意客户端使用版本表中记录的 `download_url` 发起 GET 请求
- **THEN** 系统 SHALL 返回对应 APK 内容（`application/vnd.android.package-archive` 或等价二进制流）且 SHALL NOT 允许访问约定目录之外的文件

### Requirement: 上传与文件名校验

系统 SHALL 仅接受扩展名为 **`.apk`**（大小写不敏感可统一规范）的上传；SHALL 拒绝路径分隔符、空文件名等非法文件名；SHALL 可配置单文件大小上限并在超限时拒绝。

#### Scenario: 非 APK 扩展名拒绝

- **WHEN** 管理员上传文件扩展名不是 `.apk`
- **THEN** 系统 SHALL 拒绝保存且不更新 `download_url`

### Requirement: 与现有版本检查行为兼容

新增写库记录后，`GET /device/app/api/version/check` 所依据的「最新版本行」语义 SHALL 与现有实现一致（按主键或约定排序取最新一条），且返回的 **`downloadUrl`** SHALL 与库中 `download_url` 一致。

#### Scenario: 新插入行成为最新发版

- **WHEN** 管理员通过本功能插入一条包含 `latest_version` 与 `download_url` 的新 `version` 行且其排序上为最新
- **THEN** 客户端调用版本检查接口时 SHALL 收到该行的 `latestVersion` 与 `downloadUrl`（在 semver/比较规则允许的前提下与现有版本检查逻辑一致）

---

## gateway-app-version-admin-crud

<!-- source: openspec/specs/gateway-app-version-admin-crud/spec.md -->

# gateway-app-version-admin-crud Specification

## Purpose
TBD - created by archiving change gateway-app-version-admin-crud. Update Purpose after archive.
## Requirements
### Requirement: 版本管理历史列表

gateway-app-server SHALL 向已通过 **Admin JWT** 鉴权的客户端提供历史发版列表接口，返回 `ai_voice_app.version` 表中的记录，默认按主键 `id` 降序排列。

#### Scenario: 已登录管理员获取列表

- **WHEN** 客户端携带有效 Admin JWT 并请求列表接口且 Hub admin 登录功能已启用
- **THEN** 系统 SHALL 返回 `code=0` 及包含版本记录的列表（含 `id`、`latestVersion`、`releaseDate`、`releaseNotes`、`downloadUrl`、`forceUpdate`、`minVersion`）

#### Scenario: 未登录拒绝列表

- **WHEN** 客户端未携带有效 Admin JWT 即请求列表接口
- **THEN** 系统 SHALL 返回未授权响应且 SHALL NOT 返回版本行数据

#### Scenario: 分页参数生效

- **WHEN** 客户端传入合法的 `limit` 与 `offset`
- **THEN** 系统 SHALL 按 `id` 降序返回不超过 `limit` 条记录（`limit` 不得超过约定上限）

### Requirement: 按 id 查询单条版本

系统 SHALL 支持已携带有效 Admin JWT 的管理员按主键 `id` 查询单条 `version` 记录。

#### Scenario: 存在记录时返回详情

- **WHEN** 客户端携带有效 Admin JWT 请求存在的 `id`
- **THEN** 系统 SHALL 返回该行的完整发版字段

#### Scenario: 不存在记录

- **WHEN** 客户端携带有效 Admin JWT 请求不存在的 `id`
- **THEN** 系统 SHALL 返回未找到响应且 SHALL NOT 返回伪造数据

### Requirement: 更新版本元数据

系统 SHALL 允许已鉴权管理员更新已有 `version` 行的元数据字段，且 SHALL NOT 通过本接口修改 `download_url`。

#### Scenario: 成功更新可编辑字段

- **WHEN** 已鉴权管理员提交有效 `id` 及至少一个允许字段（如 `latestVersion`、`releaseNotes`、`forceUpdate`、`minVersion`、`releaseDate`）
- **THEN** 系统 SHALL 持久化更新并返回成功响应

#### Scenario: 更新后版本检查缓存失效

- **WHEN** 更新操作成功提交
- **THEN** 系统 SHALL 删除或失效用于 `GET /device/app/api/version/check` 的最新行 Redis 缓存键 `gw:app:version:latest`

#### Scenario: 未鉴权拒绝更新

- **WHEN** 未持有有效管理会话的客户端调用更新接口
- **THEN** 系统 SHALL 拒绝写库

### Requirement: 删除版本记录

系统 SHALL 允许已鉴权管理员按 `id` 删除 `version` 表记录，并在安全条件下清理关联 APK 文件。

#### Scenario: 成功删除数据库行

- **WHEN** 已鉴权管理员删除存在的 `id`
- **THEN** 系统 SHALL 从 `version` 表移除该行并返回成功响应

#### Scenario: 删除后失效最新版本缓存

- **WHEN** 删除操作成功
- **THEN** 系统 SHALL 失效 `gw:app:version:latest` 缓存，使 `version/check` 按剩余行中最大 `id` 重新加载

#### Scenario: 尽力删除磁盘 APK

- **WHEN** 被删行的 `download_url` 为约定 path-only 形式（前缀 `/device/app/apk/`）且文件名为安全 APK 名且在存储目录内存在对应文件
- **THEN** 系统 SHALL 尝试删除该文件；若删除失败 SHALL 仍完成删行并记录可观测警告日志

#### Scenario: 非法路径不删盘外文件

- **WHEN** `download_url` 不含约定前缀或文件名未通过安全校验
- **THEN** 系统 SHALL 仅删除数据库行且 SHALL NOT 删除存储目录外或路径穿越目标

### Requirement: 新增发版与现有上传兼容

「新增（Create）」SHALL 继续通过现有 `POST /device/app/api/version/admin/upload` multipart 接口完成（APK 校验、落盘、插入新行、写入 path-only `download_url`），行为与变更前一致；调用 MUST 携带有效 Admin JWT。

#### Scenario: 上传成功仍插入新行

- **WHEN** 客户端携带有效 Admin JWT 上传合法 APK 及 `latestVersion`
- **THEN** 系统 SHALL 插入新的 `version` 行并失效最新版本缓存，且 `download_url` SHALL 为可经 `GET /device/app/apk/` 下载的 path-only 值

### Requirement: 与版本检查接口语义一致

增删改及上传写库后，`GET /device/app/api/version/check` 所依据的「最新版本行」SHALL 仍为 `version` 表中 **`id` 最大** 且 `latest_version` 非空的一行；返回的 `downloadUrl` SHALL 经现有 path 归一化后与库中 `download_url` 一致。

#### Scenario: 删除当前最大 id 后检查回落

- **WHEN** 表中仍存在其它发版行且管理员删除了当前 `id` 最大的行
- **THEN** 客户端调用 `version/check` 时 SHALL 使用剩余行中 `id` 最大者作为最新发版

### Requirement: 管理页展示与操作

`gateway-app-version-admin.html`（或等价路由页面）SHALL 在持有有效 Admin JWT 时展示历史版本列表，并 SHALL 提供触发列表刷新、编辑元数据、删除记录及上传新版本 APK 的交互；所有写操作请求 SHALL 携带 `Authorization: Bearer` Admin JWT，MUST NOT 依赖版本管理独立 Cookie。

#### Scenario: 登录后可见历史表格

- **WHEN** 管理员已在 Hub 取得 Admin JWT 并打开版本管理页
- **THEN** 页面 SHALL 请求列表接口并展示历史版本行

#### Scenario: 操作后刷新列表

- **WHEN** 上传、更新或删除任一操作成功
- **THEN** 页面 SHALL 刷新列表以反映当前表状态

### Requirement: 版本管理未启用时的错误语义

当网关进程未配置 `GATEWAY_APP_ADMIN_PASSWORD`（Hub 登录不可用）时，受保护的管理接口（含列表、查询、更新、删除、上传）SHALL 与 Hub login 一致返回服务不可用语义，且 SHALL NOT 暴露写库能力。

#### Scenario: 未配置 admin 密码拒绝受保护接口

- **WHEN** `GATEWAY_APP_ADMIN_PASSWORD` 为空且客户端调用受保护版本管理接口
- **THEN** 系统 SHALL 返回与「管理未启用」一致的不可用响应

---

## gateway-app-version-check

<!-- source: openspec/specs/gateway-app-version-check/spec.md -->

# gateway-app-version-check Specification

## Purpose
TBD - created by archiving change gateway-app-version-check-empty-no-error. Update Purpose after archive.
## Requirements
### Requirement: 版本表无数据时版本检查须成功且无需更新

gateway-app-server 对 **`GET /device/app/api/version/check`** SHALL 在版本配置表（如 `app_version`）中**不存在任何可用版本行**时，仍返回 **`code=0`**。响应 **`data.needUpdate`** SHALL 为 **`false`**。响应 SHALL NOT 因「结果集无行」或等价空表语义返回非 0 业务码。

#### Scenario: 表无任何记录

- **WHEN** 版本表为空或查询不到任何版本行
- **THEN** HTTP 业务包装 SHALL 为成功（`code=0`）且 **`needUpdate` 为 false**，且 SHALL NOT 将空结果集作为错误返回给客户端

#### Scenario: 有版本记录时行为不变

- **WHEN** 存在至少一条版本记录且 `latestVersion` 可解析
- **THEN** 系统 SHALL 继续按现有规则比较 `currentVersion` 与 `latestVersion` 并设置 **`needUpdate`**

### Requirement: 区分空表与数据库基础设施故障

当版本表查询因**连接、权限、语法等**失败时，系统 MAY 返回非 0 业务码或错误信息以便运维定位。系统 SHALL NOT 将**仅无匹配行**与上述基础设施错误等同为「统一失败」而掩盖空表成功语义。

#### Scenario: 真实读库错误

- **WHEN** 数据库返回非「无行」类错误（如连接失败）
- **THEN** 系统 MAY 返回错误响应且 SHOULD NOT 冒充「无需更新」的成功语义

---

## gateway-no-business-workers

<!-- source: openspec/specs/gateway-no-business-workers/spec.md -->

# gateway-no-business-workers Specification

## Purpose
TBD - created by archiving change worker-exclusive-background-tasks. Update Purpose after archive.
## Requirements
### Requirement: Gateway MUST 保持无业务后台任务职责

`gateway-service` MUST 仅承担请求入口、路由转发与横切能力，MUST NOT 在进程启动阶段承担消息消费、事件中继、domain outbox relay 等业务后台任务职责。

#### Scenario: 网关处理请求

- **WHEN** gateway 接收 HTTP/WS 请求
- **THEN** gateway MUST 仅执行入口与代理逻辑，不应存在后台任务消费副作用

#### Scenario: 部署配置审查

- **WHEN** 审查 gateway 部署配置与启动流程
- **THEN** 必须能够确认 gateway 未启动 ticker 扫表或 MQ 消费；history/device 缓存由各自服务同步 patch；UCG 后台任务仅在 ucg-service

---

## gateway-policy-layer-consolidation

<!-- source: openspec/specs/gateway-policy-layer-consolidation/spec.md -->

# gateway-policy-layer-consolidation Specification

## Purpose
TBD - created by archiving change platform-hardening-redis-rabbitmq-service-split. Update Purpose after archive.
## Requirements
### Requirement: Gateway 必须收敛为流量与策略层
`gateway-service` SHALL 仅提供边缘层能力，包括鉴权、路由、策略执行、元数据透传和流量控制。

#### Scenario: 请求进入 gateway
- **WHEN** 客户端请求到达 `gateway-service`
- **THEN** gateway SHALL 执行边缘策略并转发至对应领域服务，不得执行领域业务规则

### Requirement: Gateway 在委派领域执行时必须保持外部契约稳定
`gateway-service` SHALL 在服务拆分过程中保持对外 API 契约稳定，并 SHALL 将领域业务执行委派给下游领域服务。

#### Scenario: 拆分后的既有外部 API 调用
- **WHEN** 客户端调用既有公开 API 端点
- **THEN** gateway SHALL 在调用下游领域服务处理业务的同时返回契约兼容响应

---

## gateway-route-middleware-domain-isolation

<!-- source: openspec/specs/gateway-route-middleware-domain-isolation/spec.md -->

# gateway-route-middleware-domain-isolation Specification

## Purpose
TBD - created by archiving change voice-device-canary-route-split. Update Purpose after archive.
## Requirements
### Requirement: Gateway 路由中间件 MUST 按领域拆分管理
gateway MUST 将 voice 与 device 路由代理逻辑拆分为独立中间件与配置读取路径，不得在同一中间件实现中混合管理两个领域。

#### Scenario: 修改 voice 路由逻辑
- **WHEN** 开发者调整 voice 路由代理策略
- **THEN** 变更 MUST 限定在 voice 独立中间件实现内，且不应直接影响 device 路由行为

#### Scenario: 修改 device 路由逻辑
- **WHEN** 开发者调整 device 路由代理策略
- **THEN** 变更 MUST 限定在 device 独立中间件实现内，且不应直接影响 voice 路由行为

### Requirement: 领域路由配置 MUST 互相隔离
voice 与 device 的路由模式、目标地址、canary 百分比配置 MUST 分别独立，禁止共享配置键。

#### Scenario: 仅调整 voice canary 百分比
- **WHEN** 运维仅修改 voice 的 canary 百分比配置
- **THEN** device 路由行为 MUST 保持不变

---

## gateway-voice-http-proxy

<!-- source: openspec/specs/gateway-voice-http-proxy/spec.md -->

# gateway-voice-http-proxy Specification

## Purpose
TBD - created by archiving change refactor-ai-quota-domain-ownership. Update Purpose after archive.
## Requirements
### Requirement: gateway-app-server SHALL proxy voice App HTTP APIs

`gateway-app-server` MUST 将 **`/voice/app/api/*`** 注册为 HTTP 反向代理至 voice-service（环境变量 `VOICE_API_PROXY_URL`、`VOICE_API_ROUTE_MODE`，模式对齐 `device_route_proxy.go` / `ucg_route_proxy.go`）。对受保护路径 MUST 经 Bearer 鉴权并注入 `X-Internal-Wx-Id`（及有值时的 `X-Internal-Device-No`）后转发。gateway-app MUST NOT 在本地实现 voice App 业务逻辑。

#### Scenario: App 查询 voice 域额度经 gateway 反代

- **WHEN** Flutter 携带有效 Bearer 请求 `GET /voice/app/api/ai-quota`
- **THEN** gateway-app MUST 注入内部头并转发至 voice-service 同路径
- **AND** gateway-app MUST NOT 本地聚合 polish 或 clinic 数据

#### Scenario: 反代目标不可达

- **WHEN** `VOICE_API_ROUTE_MODE=proxy` 且 voice-service 不可达
- **THEN** gateway-app MUST 返回可诊断的代理错误

### Requirement: gateway-app-server SHALL proxy voice Admin HTTP APIs with password injection

`gateway-app-server` MUST 将 **`/voice/admin/api/*`** 反代至 voice-service。Admin JWT 校验通过后，`InjectAdminDownstreamPassword` MUST 对 `/voice/admin/api/` 前缀注入 `X-Admin-Password`（值来自 **`VOICE_ADMIN_PASSWORD`** env / `voice.admin.password` 配置）。

#### Scenario: voice admin API 口令注入

- **WHEN** 已登录 Admin Hub 的用户请求 PUT `/voice/admin/api/ai-quota/default`
- **THEN** gateway-app MUST 注入 `X-Admin-Password` 并转发至 voice-service

#### Scenario: 未登录 Admin 拒绝

- **WHEN** 请求 `/voice/admin/api/*` 且无有效 Admin JWT
- **THEN** gateway-app SHALL 返回未授权且 SHALL NOT 转发

### Requirement: gateway-app-server SHALL remove device ai-quota App proxy

`gateway-app-server` MUST **移除** `device_route_proxy.go` 中对 **`/device/app/api/ai-quota`** 的反代登记。该路径 MUST NOT 再可达 device-service ai-quota 读 API。

#### Scenario: 旧 App 读路径不可用

- **WHEN** 客户端请求 `GET /device/app/api/ai-quota`
- **THEN** gateway-app SHALL NOT 反代至 device ai-quota（返回 404 或由网关本机明确拒绝）

### Requirement: gateway-service SHALL sync voice HTTP proxy paths

`gateway-service` MUST 同步注册 `/voice/app/api/*` 与 `/voice/admin/api/*` 反代，行为与 gateway-app 对齐，以便管理/通用网关与 App 网关路径一致。

#### Scenario: gateway-service 反代 voice App API

- **WHEN** 客户端经 gateway-service 请求 `/voice/app/api/ai-quota` 且 proxy 配置正确
- **THEN** gateway-service MUST 转发至 voice-service

---

## gateway-ws-delegation-convergence

<!-- source: openspec/specs/gateway-ws-delegation-convergence/spec.md -->

# gateway-ws-delegation-convergence Specification

## Purpose
TBD - created by archiving change gateway-ws-proxy-and-remove-local-ws. Update Purpose after archive.
## Requirements
### Requirement: Gateway MUST 移除本地 voice WS 领域执行
`gateway-service` MUST 不再在 `/voice/chat/ws` 执行本地语音对话业务逻辑，领域处理必须由 `voice-service` 承担。

#### Scenario: Gateway 收到 voice WS 请求
- **WHEN** 客户端连接 `/voice/chat/ws`
- **THEN** gateway MUST 仅执行边缘层职责（路由、策略、元数据透传），并将领域执行委派给 `voice-service`

### Requirement: Gateway SHALL 保持对外 WS 入口契约稳定
迁移到委派模式时，gateway MUST 保持外部 WS 路径与接入方式稳定，避免要求前端同步改地址。

#### Scenario: 前端继续使用既有 WS 地址
- **WHEN** 前端仍连接 gateway 既有 `/voice/chat/ws` 地址
- **THEN** 系统 MUST 可完成握手与消息收发，且业务执行由下游 `voice-service` 完成

---

## gateway-ws-edge-proxy

<!-- source: openspec/specs/gateway-ws-edge-proxy/spec.md -->

# gateway-ws-edge-proxy Specification

## Purpose
TBD - created by archiving change gateway-ws-proxy-and-remove-local-ws. Update Purpose after archive.
## Requirements
### Requirement: Gateway SHALL 支持 voice WebSocket 边缘透传
`gateway-service` MUST 在 `/voice/chat/ws` 提供可配置透传能力，将 WebSocket 连接转发到 `voice-service` 目标地址。

#### Scenario: WS 透传启用且目标可达
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为可达地址
- **THEN** gateway MUST 将 `/voice/chat/ws` 的握手与后续双向消息透传至目标服务

#### Scenario: WS 透传启用但目标不可达
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **THEN** gateway MUST 返回明确的握手/代理失败错误，且 MUST NOT 回退本地业务执行

### Requirement: Gateway MUST 提供 WS 透传配置约束
gateway MUST 通过环境变量控制 WS 路由行为，并对非法配置执行可预测回退。

#### Scenario: 路由模式非法
- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **THEN** gateway MUST 将 WS 路由模式视为 `local`

#### Scenario: 代理地址为空或非法
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为空或非法
- **THEN** gateway MUST 视为未启用可用代理目标并返回可诊断错误，不得出现静默成功

### Requirement: Gateway SHALL 支持听写 WebSocket 边缘透传

`gateway-service` 与 `gateway-app-server` MUST 在 `/voice/asr/ws` 提供与 `/voice/chat/ws` 相同的可配置 WebSocket 透传能力，将连接转发至 `voice-service`（与 `VOICE_WS_PROXY_URL` 同一目标基址）。

#### Scenario: WS 透传启用且目标可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为可达地址
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 将该路径的握手与后续双向消息透传至 voice-service，行为与 `/voice/chat/ws` 一致

#### Scenario: WS 透传启用但目标不可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 返回明确的握手/代理失败错误，且 MUST NOT 在 gateway 本地执行听写业务逻辑

#### Scenario: 路由模式非 proxy

- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 返回可诊断的配置错误（与 chat WS 一致），且 MUST NOT 静默成功

### Requirement: App 网关 SHALL 将听写 WS 纳入 Bearer 白名单

`gateway-app-server` MUST 将 `GET /voice/asr/ws`（WebSocket Upgrade）列入 Bearer 鉴权豁免路径，与 `/voice/chat/ws` 策略一致。

#### Scenario: 无 Bearer 的 Upgrade 请求

- **WHEN** App 客户端对 `/voice/asr/ws` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app MUST 允许进入透传或 voice-service 处理链，不得仅因缺少 Bearer 拒绝 Upgrade

### Requirement: gateway-app-server SHALL 注册胖宝 clinic WebSocket 为 App 对外入口

`gateway-app-server` MUST 将 `GET /voice/clinic/ws` 注册为 App 客户端的**唯一对外 WebSocket 入口**（与 `/voice/chat/ws`、`/voice/asr/ws` 同 `apiBaseUrl` 主机、同 `installVoiceWSProxyMiddleware` 透传链）。实现 MUST 将 `/voice/clinic/ws` 加入 `internal/controller/ws_route_proxy.go` 的 `voiceWSProxyPaths`，由 `RegisterGatewayAppHTTP` 已挂载的 `installVoiceWSProxyMiddleware` 将握手与双向消息透传至 `voice-service`（`VOICE_WS_PROXY_URL` 同一目标基址）。App 客户端 MUST NOT 被要求或配置为直连 `voice-service` 内网地址。

#### Scenario: App 经 gateway-app 连接 clinic WS

- **WHEN** Flutter 使用 `wss://{apiBaseUrl host}/voice/clinic/ws` 发起 WebSocket Upgrade
- **AND** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 可达
- **THEN** gateway-app-server MUST 透传握手与后续双向帧至 voice-service `/voice/clinic/ws`
- **AND** gateway-app MUST NOT 在本地执行 clinic 业务逻辑

#### Scenario: WS 透传启用但目标不可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **AND** 客户端连接 gateway-app `/voice/clinic/ws`
- **THEN** gateway-app MUST 返回明确的握手/代理失败错误，且 MUST NOT 在 gateway 本地执行 clinic 业务逻辑

#### Scenario: 路由模式非 proxy

- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **AND** 客户端连接 gateway-app `/voice/clinic/ws`
- **THEN** gateway-app MUST 返回可诊断的配置错误，且 MUST NOT 静默成功

### Requirement: gateway-service MUST 同步 clinic WebSocket 透传路径

`gateway-service` MUST 将 `/voice/clinic/ws` 加入同一 `voiceWSProxyPaths` 列表，行为与 `/voice/asr/ws` 一致，以便管理/通用网关与 App 网关路径对齐；**App 主入口仍为 gateway-app-server**。

#### Scenario: gateway-service 透传 clinic WS

- **WHEN** 客户端连接 gateway-service `/voice/clinic/ws` 且 proxy 模式配置正确
- **THEN** gateway-service MUST 透传至 voice-service，行为与 chat/ASR WS 一致

### Requirement: gateway-app-server SHALL 将 clinic WS 纳入 Bearer 白名单

`gateway-app-server` MUST 将 `GET /voice/clinic/ws`（WebSocket Upgrade）列入 `gateway_app_auth_exempt.go` 的 `gatewayAppAuthExemptExactGET`，与 `/voice/asr/ws` 策略一致：Upgrade 不要求 HTTP 层 Bearer；若客户端仍携带 Bearer，`HookBeforeServe` MAY 注入 `X-Internal-Wx-Id`，但 clinic 身份校验 MUST 由 voice-service 首帧 `auth` 完成。

#### Scenario: 无 Bearer 的 Upgrade 请求

- **WHEN** App 客户端对 gateway-app `/voice/clinic/ws` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app SHALL 允许请求进入 WS 透传链（由 voice-service 首帧 `auth` 校验 wxId；**非** deviceNo 反查）

#### Scenario: 可选 Bearer 仍注入内部头

- **WHEN** App 对 `/voice/clinic/ws` Upgrade 且携带有效 App access Bearer
- **THEN** gateway-app MAY 经 `InjectAccessHeadersFromBearer` 注入 `X-Internal-Wx-Id`
- **AND** voice-service clinic handler MUST NOT 仅以该头作为鉴权依据（首帧 JWT 为准）

---

## growth-invite-per-user-only

<!-- source: openspec/specs/growth-invite-per-user-only/spec.md -->

# growth-invite-per-user-only Specification

## Purpose
TBD - created by archiving change feature-ops-invite-force-admin. Update Purpose after archive.
## Requirements
### Requirement: 成长轨迹邀请按人一次且允许同机多码

对 `growth_trajectory_predict`，系统在邀请码兑换时 MUST NOT 因同一 `device_no` 已对本功能邀请开通而拒绝（MUST NOT 再适用 `InviteOncePerDevice` / MUST NOT 读写本功能的 `feature_invite_device_grant` 作为拒绝依据）。系统 MUST 继续：若同一 `redeemer_wx_id` 已对任意邀请码成功开通过本功能，则拒绝再次邀请开通；MUST 保留人×码×功能去重、自用禁兑与同宝宝禁兑。`care_alert_smart_remind` 的设备邀请一次约束 MUST 保持不变。

#### Scenario: 同宝宝第二人兑不同码

- **WHEN** 家长 A 已用码 X 邀请开通成长轨迹，家长 B（不同 wx、同一 `device_no`）使用另一好友码 Y 兑换成长轨迹且 B 从未邀请开通过本功能
- **THEN** 系统 MUST 允许兑换并写入 B 的 user 权益

#### Scenario: 同一人第二次邀请拒绝

- **WHEN** 同一 `redeemer_wx_id` 已对成长轨迹邀请开通成功后再次用任意码兑换本功能
- **THEN** 系统 MUST 拒绝

#### Scenario: 值得留意仍同机一次

- **WHEN** 某 `device_no` 已邀请开通 `care_alert_smart_remind` 后，同机另一账号用另一码再兑该功能
- **THEN** 系统 MUST 拒绝

---

## growth-trajectory-user-gates

<!-- source: openspec/specs/growth-trajectory-user-gates/spec.md -->

# growth-trajectory-user-gates Specification

## Purpose
TBD - created by archiving change growth-trajectory-user-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 成长轨迹智能分析 MUST 校验账号开通或 VIP

`POST /device/api/growth-trajectory/turn`（及 voice 内等价入口）在开流前 MUST 要求有效登录 `wxId`，并 MUST 经 cash 合成：该 `wxId` 对 `growth_trajectory_predict` 存在未过期 user 权益 **或** 该账号为 VIP。未满足 MUST 拒绝且 MUST NOT 调用 Python turn。合成 MUST NOT 仅因同 `device_no` 上他人已开通而放行。

#### Scenario: 未开通非 VIP 拒绝 turn

- **WHEN** 已登录用户无成长轨迹 user 权益且非 VIP 请求 turn
- **THEN** 系统 MUST 返回未开通类错误且不开 SSE 业务流

#### Scenario: VIP 免开通可 turn

- **WHEN** 用户无 user 权益但是 VIP 请求 turn
- **THEN** 系统 MUST 允许进入分析流（仍受日限等其它闸门约束）

### Requirement: 成长轨迹最新结果 MUST 按设备读取且免开通

`GET /device/api/growth-trajectory/latest` MUST 要求登录，MUST 按请求 `deviceNo` 返回该设备最新一条结果（无则空字段），MUST NOT 校验功能开通或 VIP。存储仍为按设备最新一条，MUST NOT 本变更引入多版本历史列表。

#### Scenario: 未开通用户可读最新报告

- **WHEN** 已登录但未开通成长轨迹的用户请求 latest 且该 deviceNo 已有结果
- **THEN** 响应 MUST 返回已有 `resultMarkdown`（或等价字段）且 MUST NOT 因未开通失败

#### Scenario: 未登录拒绝 latest

- **WHEN** 缺少有效登录 wxId 请求 latest
- **THEN** 系统 MUST 拒绝

### Requirement: 成长轨迹日限 MUST 按用户计数

成长轨迹每日成功结果次数上限 MUST 以登录 `wxId` + 上海日历日为键计数（经 cachekit 键构建器），MUST NOT 再以 `deviceNo` 作为日限主键。达上限时 turn MUST 在开流前拒绝。

#### Scenario: 同人换设备共享日限

- **WHEN** 同一 wxId 在设备 D1 已用尽当日次数后在 D2 再请求 turn
- **THEN** 系统 MUST 因日限拒绝

### Requirement: 成长轨迹邀请 MUST 同时人一次与设备一次

对 `growth_trajectory_predict` 邀请开通，系统 MUST 在既有人×码×功能去重之外：若该 `redeemer_wx_id` 已对任意邀请码成功开通过本功能，MUST 拒绝再次邀请开通；若该 `device_no` 已对本功能邀请开通过，MUST 拒绝。设备一次闸 MUST 继续写入/校验 `feature_invite_device_grant`（或等价）。值得留意等其它 InviteOncePerDevice 功能 MUST NOT 因本变更被迫加上「人一次」除非另行配置。

#### Scenario: 同一人第二次兑码成长轨迹

- **WHEN** 用户已用码1开通成长轨迹后用码2再兑同一功能
- **THEN** 系统 MUST 拒绝且 MUST NOT 延长 user 权益

#### Scenario: 同设备第二人兑码成长轨迹

- **WHEN** 设备 D 已有任意账号邀请开通过成长轨迹后另一账号再兑
- **THEN** 系统 MUST 因设备已开通过而拒绝

### Requirement: cash 成长轨迹 access MUST 以账号权益为准

内部 `GET /cash/internal/api/growth-trajectory/access` MUST 以 `wxId` 的 user 权益 ∨ VIP 合成 `allowed`/`featureActive`，MUST NOT 将设备维 `feature_entitlement` 上的成长轨迹行视为有效开通（本变更后）。

#### Scenario: 仅有历史 device 行不算开通

- **WHEN** 某 device 上仍有旧成长轨迹 device entitlement，但当前 wx 无 user 权益且非 VIP
- **THEN** access MUST 返回不允许

---

## history-delegate-downstream-urls

<!-- source: openspec/specs/history-delegate-downstream-urls/spec.md -->

# history-delegate-downstream-urls Specification

## Purpose
TBD - created by archiving change docker-history-cross-service-urls. Update Purpose after archive.
## Requirements
### Requirement: history-service 在隔离网络运行时 MUST 使用可路由的下游 HTTP 基址

当 `history-service` 进程运行在与其他领域服务**不同的网络命名空间**（例如 Docker 独立容器、Kubernetes Pod）且需要通过 HTTP 委派访问 `voice-service` 或 `device-service` 时，其 MUST 通过环境变量 `VOICE_SERVICE_URL` 与 `DEVICE_SERVICE_URL` 配置为可在该命名空间内解析并路由到目标服务的基址（例如同一编排系统中的服务 DNS 名），MUST NOT 依赖指向本容器 loopback 的默认基址（如 `http://127.0.0.1:9802`）作为跨容器访问手段。

#### Scenario: Docker Compose 微服务栈中 history 访问 device

- **WHEN** `history-service` 与 `device-service` 作为不同容器加入同一用户定义 bridge/overlay 网络，且 history 需调用 device 的 HTTP API
- **THEN** 部署配置 SHALL 为 history 设置 `DEVICE_SERVICE_URL`（例如 `http://device-service:9803`），使得 TCP 连接目标为 device 容器而非 history 容器自身

#### Scenario: Docker Compose 微服务栈中 history 访问 voice

- **WHEN** `history-service` 与 `voice-service` 作为不同容器在同一编排网络内，且 history 需调用 voice 的 HTTP API
- **THEN** 部署配置 SHALL 为 history 设置 `VOICE_SERVICE_URL`（例如 `http://voice-service:9802`），使得请求到达 voice 服务监听端口

### Requirement: 仓库参考 Compose 中 history 段落 SHALL 与 voice 下游配置语义一致

`manifest/docker/docker-compose.microservices.yml`（或其后继等价的官方微服务 Compose 参考）中，若同时定义 `history-service` 与 `voice-service`、`device-service`，则 `history-service` 的环境变量段落 SHALL 包含与同文件内其他服务一致的、基于服务名的 `VOICE_SERVICE_URL` 与 `DEVICE_SERVICE_URL`，以避免开箱即用栈中出现 history 误连 `127.0.0.1` 的失败。

#### Scenario: 审查者对比 voice 与 history 环境块

- **WHEN** 审查者检查微服务 Compose 文件中 voice 已配置 `DEVICE_SERVICE_URL` 指向 `device-service`
- **THEN** 其 SHALL 能在 history 段落找到对 voice、device 的显式 URL 配置，且主机部分为 compose 服务名而非 `127.0.0.1`

---

## history-device-sync-cache-projection

<!-- source: openspec/specs/history-device-sync-cache-projection/spec.md -->

# history-device-sync-cache-projection Specification

## Purpose
TBD - created by archiving change remove-worker-simplify-cache. Update Purpose after archive.
## Requirements
### Requirement: history 写操作 MUST 同步更新 Redis 读模型

`history-service`（及 voice adapter 本地 patch 路径）在 `history` 表 insert/update/delete **成功提交后**，MUST 在同请求内调用读模型 patch（`patchHistoryOnAdd` / `patchHistoryOnUpdate` / `patchHistoryOnDelete`）或等价逻辑更新 `history:record:list:*` 与 `history:record:latest:*`；MUST NOT 依赖 worker-service 或 domain_outbox 异步投影作为唯一更新路径。

#### Scenario: AddHistory 成功后列表与 latest 可读

- **WHEN** `AddDeviceHistory` 事务提交成功
- **THEN** 系统 MUST 尝试同步 patch Redis；后续 `GetLatestHistory` 在 cache miss 或 patch 成功时 MUST 返回含新记录的正确数据（以 MySQL 为准）

#### Scenario: 冷缓存新增记录仍更新 latest

- **WHEN** 设备 history 列表 Redis key 不存在且新增一条记录
- **THEN** 系统 MUST 至少写入 `history:record:latest:{deviceNo}`；列表 MAY 在首次 `ListHistory` miss 时从 MySQL 全量回填

### Requirement: list patch 失败 MUST 避免长期 stale hit

当 `setHistoryList` 失败时，系统 SHOULD best-effort 删除对应 list key，使下次读取走 MySQL 回源；MUST NOT 假设 worker 会异步修复。

#### Scenario: Redis 写入失败后读路径自愈

- **WHEN** prepend 列表缓存时 Redis 返回错误
- **THEN** 系统 MUST 记录 warning；SHOULD 删除 list key；读路径在 miss 时 MUST 从 MySQL 重建列表

### Requirement: device 域字典与画像 MUST 在写路径同步重建缓存

device admin 变更 `event`/`action`/`user` 后，MUST 在写库成功后调用 `refreshEventOptionsCacheAfterMutate`、`RebuildActionCache` 或 `setUserProfile` 等同步重建；MUST NOT 依赖 worker HTTP 回调作为唯一刷新路径。

#### Scenario: 后台新增事件后主数据可读

- **WHEN** 管理端新增事件并提交成功
- **THEN** 下一次 `ListEvents` MUST 能读到新事件（Redis hit 或 miss 回源 MySQL 后回填）

### Requirement: 读路径 MUST 在 cache miss 时回源 MySQL 并回填

`ListHistory`、`GetLatestHistory`、device `ListEvents` 等在 Redis miss 或不可用时，MUST 查询权威 MySQL 并回填 Redis；数据 correctness MUST 以 MySQL 为准。

#### Scenario: 列表 cache miss

- **WHEN** `getHistoryList` 未命中
- **THEN** adapter MUST 查库并 `setHistoryList` 回填

---

## history-event-unit-denorm

<!-- source: openspec/specs/history-event-unit-denorm/spec.md -->

# history-event-unit-denorm Specification

## Purpose
TBD - created by archiving change history-event-unit-denorm. Update Purpose after archive.
## Requirements
### Requirement: History 写入 MUST 反规范化事件单位

当新增或更新 `history` 记录且关联的 `eventId` 在 device 事件主档中存在非空 `unit` 时，history-service MUST 在持久化前将单位写入 `history.event_unit`（写入时刻快照）。若请求体已携带非空 `eventUnit`，MUST 优先使用该值；否则 MUST 经 device-service HTTP 契约解析单位。history-service MUST NOT 通过本进程 `default` 数据库连接直查 `event` 表。

#### Scenario: 微服务分库下新增历史带单位

- **WHEN** device 库中 `event.id=10` 的 `unit` 为 `ml`，客户端经 history HTTP 新增一条 `eventId=10` 的历史且请求体未传 `eventUnit`
- **THEN** history-service MUST 经 device 契约解析到 `ml` 并成功 INSERT，`history.event_unit` SHALL 为 `ml`

#### Scenario: 请求体显式携带单位

- **WHEN** 客户端 POST 新增历史且 `eventUnit` 为 `次`
- **THEN** history-service MUST 持久化 `history.event_unit=次`，且 MUST NOT 被主档单位覆盖

#### Scenario: 事件主档无单位

- **WHEN** 关联 `event.unit` 为空且请求体未传 `eventUnit`
- **THEN** history-service MUST 持久化 `history.event_unit` 为 NULL 或等效空值，且 MUST NOT 报错

#### Scenario: 禁止跨库直查 event 表

- **WHEN** history-service 运行于仅配置 `HISTORY_DB_LINK` 的环境
- **THEN** 补全 `event_unit` 的实现 MUST NOT 调用 `dao.Event` 或等价 history 库内 event 表访问

### Requirement: History 读路径 MUST 暴露 eventUnit

`GET /device/history/api/list`、`GET /device/history/api/piece` 及单条查询响应中的 history 实体 MUST 包含 `eventUnit` 字段（JSON camelCase），值与数据库 `event_unit` 一致。

#### Scenario: 列表返回单位

- **WHEN** 数据库行 `event_unit=ml`
- **THEN** 列表 JSON 中对应记录的 `eventUnit` SHALL 为 `ml`

### Requirement: 历史管理页 MUST 展示计数单位

`resource/public/history.html` 对 `eventType=number` 的记录，展示数量时 MUST 在数字后附加单位：优先使用记录 `eventUnit`；若为空且所选事件 option 含 unit，MAY 回退展示 option 单位。

#### Scenario: 有 eventUnit 的计数记录

- **WHEN** 列表项 `eventNumber=120` 且 `eventUnit=ml`
- **THEN** 页面数量列 SHALL 展示含 `ml` 的可读文本（如 `120ml` 或 `120 ml`，实现 MUST 文档化一种格式）

---

## history-filter-api

<!-- source: openspec/specs/history-filter-api/spec.md -->

# history-filter-api Specification

## Purpose
TBD - created by archiving change add-history-filter-v2-list. Update Purpose after archive.
## Requirements
### Requirement: 历史记录筛选 API

系统 SHALL 提供 `GET /device/history/api/filter` 接口，支持按设备号、事件ID列表、时间范围和返回条数上限筛选历史记录。

#### Scenario: 正常筛选 - 所有参数都传
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1,2,5`、`startTime=1234567890`、`endTime=1234567990`、`limit=50` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在时间范围内、事件ID为1/2/5的最多50条历史记录
- **AND** 结果 SHALL 按 id 倒序排列（最新在前）

#### Scenario: eventIds 为空时跳过事件过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=`（空串）、`startTime=1234567890`、`endTime=1234567990` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在时间范围内的所有事件类型的历史记录（不按事件ID过滤）

#### Scenario: startTime 为空时跳过开始时间条件
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1`、`startTime=0`、`endTime=1234567990` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在 endTime 之前、事件ID为1的历史记录（不限制开始时间）

#### Scenario: endTime 为空时跳过结束时间条件
- **WHEN** 客户端以 `deviceNo=XXX`、`eventIds=1`、`startTime=1234567890`、`endTime=0` 调用 filter 接口
- **THEN** 系统 SHALL 返回该设备在 startTime 之后、事件ID为1的历史记录（不限制结束时间）

#### Scenario: limit 默认值为 100
- **WHEN** 客户端以 `deviceNo=XXX`、`limit=0`（或不传 limit）调用 filter 接口
- **THEN** 系统 SHALL 最多返回 100 条历史记录

#### Scenario: limit 上限为 500
- **WHEN** 客户端以 `deviceNo=XXX`、`limit=1000` 调用 filter 接口
- **THEN** 系统 SHALL 将 limit 限制为 500，最多返回 500 条历史记录

#### Scenario: deviceNo 为空时返回空列表
- **WHEN** 客户端以 `deviceNo=`（空串）调用 filter 接口
- **THEN** 系统 SHALL 返回空列表（不报错）

#### Scenario: 响应格式正确
- **WHEN** filter 接口成功返回
- **THEN** 响应 JSON SHALL 为 `{ "list": [ { "id": ..., "deviceNo": ..., "eventId": ..., ... } ] }` 格式
- **AND** 每条记录 SHALL 包含与 v1 list 接口相同的字段

#### Scenario: 支持 local/remote/canary 服务模式
- **WHEN** 服务运行在 local 模式
- **THEN** filter 接口 SHALL 直连数据库执行查询
- **WHEN** 服务运行在 remote 模式
- **THEN** filter 接口 SHALL 通过 HTTP 调用远程 history-service
- **WHEN** 服务运行在 canary 模式
- **THEN** filter 接口 SHALL 按设备号一致性分流到本地或远程

### Requirement: filter 支持 ignoreTimeRange 强制忽略时间窗

系统 SHALL 在既有 `GET /device/history/api/filter` 上接受可选查询参数 `ignoreTimeRange`。未传或为假时，时间筛选语义 MUST 与现网一致（`startTime`/`endTime` > 0 才施加对应条件，=0 跳过）。为真时，系统 MUST **完全忽略**请求中的 `startTime` 与 `endTime`（即使均为正数），MUST NOT 对其施加任何时间 WHERE 条件；`deviceNo`、`eventIds`、`remark`、`limit` 与按 id 倒序等既有规则 MUST 保持不变。本参数为 additive 扩展，MUST NOT 要求客户端升级后才能调用原筛选行为。

#### Scenario: 默认不忽略时间（兼容旧客户端）

- **WHEN** 客户端调用 filter 且未传 `ignoreTimeRange`（或显式为假），并传入 `startTime`/`endTime` 正数
- **THEN** 系统 SHALL 按现网规则对 `start_time` 施加对应上下界过滤

#### Scenario: ignoreTimeRange 为真时忽略已填时间

- **WHEN** 客户端以 `ignoreTimeRange` 为真，且同时传入非零 `startTime` 与/或 `endTime` 调用 filter
- **THEN** 系统 SHALL 返回结果时 MUST NOT 使用上述时间值作为过滤条件
- **AND** 其他条件（如 `eventIds`、`remark`、`limit`）SHALL 仍按既有规则生效

#### Scenario: ignoreTimeRange 为真且无事件/备注时仍按 limit 收口

- **WHEN** 客户端以 `ignoreTimeRange` 为真、`eventIds` 为空、`remark` 为空调用 filter
- **THEN** 系统 SHALL 在无时间条件下按设备维度返回最多 `limit` 条（默认/上限规则与现网一致）的历史记录（id 倒序）

#### Scenario: local/remote/canary 透传一致

- **WHEN** history 以 remote 或 canary 模式调用 filter
- **THEN** `ignoreTimeRange` 为真时 MUST 经 HTTP 契约透传至 history-service，本地与远程语义 MUST 一致

---

## history-list-v2-api

<!-- source: openspec/specs/history-list-v2-api/spec.md -->

# history-list-v2-api Specification

## Purpose
TBD - created by archiving change add-history-filter-v2-list. Update Purpose after archive.
## Requirements
### Requirement: v2 历史列表 API

系统 SHALL 提供 `GET /device/history/api/v2/list` 接口，在 v1 分页基础上扩展 startTime、endTime、limit 可选参数。

#### Scenario: 不传新参数时行为与 v1 完全一致
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20` 调用 v2 list 接口（不传 startTime、endTime、limit）
- **THEN** 系统 SHALL 返回与 v1 list 接口完全相同的结果（分页、排序、字段完全一致）
- **AND** v1 list 接口 SHALL 保持完全不变

#### Scenario: 传了 limit 时优先使用 limit 替代 pageSize
- **WHEN** 客户端以 `deviceNo=XXX`、`page=2`、`pageSize=20`、`limit=100` 调用 v2 list 接口
- **THEN** 系统 SHALL 使用 limit=100 作为每页条数，page 忽略（固定为1）
- **AND** total 字段 SHALL 仍然返回符合条件的总记录数

#### Scenario: 传 startTime 时按开始时间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20`、`startTime=1234567890` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time >= 1234567890 的记录
- **AND** total 字段 SHALL 仅统计符合时间条件的记录数

#### Scenario: 传 endTime 时按结束时间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`page=1`、`pageSize=20`、`endTime=1234567990` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time <= 1234567990 的记录

#### Scenario: 同时传 startTime 和 endTime 时按时间区间过滤
- **WHEN** 客户端以 `deviceNo=XXX`、`startTime=1234567890`、`endTime=1234567990` 调用 v2 list 接口
- **THEN** 系统 SHALL 仅返回 start_time 在 [1234567890, 1234567990] 区间内的记录

#### Scenario: pageSize 上限仍为 100（与 v1 一致）
- **WHEN** 客户端以 `deviceNo=XXX`、`pageSize=200` 调用 v2 list 接口（未传 limit）
- **THEN** 系统 SHALL 将 pageSize 限制为 100

#### Scenario: 排序使用 id 倒序
- **WHEN** v2 list 接口返回结果
- **THEN** 结果 SHALL 按 id 倒序排列（与 v1 一致，最新在前）

#### Scenario: 时间单位为 Unix 秒
- **WHEN** 客户端传 `startTime` 或 `endTime`
- **THEN** 系统 SHALL 按 Unix 秒解释该值（与 piece 接口一致）
- **AND** 值为 0 时 SHALL 视为"不限制"，跳过对应时间条件

#### Scenario: deviceNo 为空时返回空分页
- **WHEN** 客户端以 `deviceNo=`（空串）调用 v2 list 接口
- **THEN** 系统 SHALL 返回空列表和 total=0（不报错）

#### Scenario: 响应格式与 v1 一致
- **WHEN** v2 list 接口成功返回
- **THEN** 响应 JSON SHALL 为 `{ "list": [...], "total": N, "page": N, "pageSize": N }` 格式（与 v1 完全一致）

#### Scenario: 支持 local/remote/canary 服务模式
- **WHEN** 服务运行在 local 模式
- **THEN** v2 list 接口 SHALL 直连数据库执行查询
- **WHEN** 服务运行在 remote 模式
- **THEN** v2 list 接口 SHALL 通过 HTTP 调用远程 history-service
- **WHEN** 服务运行在 canary 模式
- **THEN** v2 list 接口 SHALL 按设备号一致性分流到本地或远程

---

## history-piece-and-realtime-notify

<!-- source: openspec/specs/history-piece-and-realtime-notify/spec.md -->

# history-piece-and-realtime-notify Specification

## Purpose
TBD - created by archiving change gateway-app-server-wx-auth-history-ws. Update Purpose after archive.
## Requirements
### Requirement: 事件区间查询 piece 接口

history-service SHALL 提供 `GET /device/history/api/piece`，接受查询参数 `eventId`、`startTime`、`endTime`、`deviceNo`，并 SHALL 返回该设备在指定时间区间内、指定事件类型下的历史记录集合（字段与现有 history 列表语义一致或可被子集化但 MUST 文档化）。

#### Scenario: 有数据区间

- **WHEN** 参数合法且数据库存在匹配记录
- **THEN** 响应 SHALL 包含记录列表且顺序与产品设计一致（例如按时间升序）

#### Scenario: 无数据

- **WHEN** 区间内无匹配记录
- **THEN** 响应 SHALL 返回空列表而非错误，除非参数非法

### Requirement: piece 结果 Redis 缓存

history-service SHALL 对 piece 查询结果使用 Redis 缓存以降低数据库压力；缓存键 MUST 能区分 eventId、startTime、endTime、deviceNo 的组合。

#### Scenario: 缓存命中

- **WHEN** 相同查询在 TTL 内重复到达
- **THEN** 服务 MAY 从 Redis 返回缓存结果且结果与数据库一致

### Requirement: 历史 CUD 后发布 Redis 通知

history-service SHALL 在任何导致 history 表新增、更新或删除成功的业务路径完成后，向约定 Redis channel 发布一条消息；消息体 MUST 包含 `device_no`、操作类型 `action`（create、update、delete 之一）以及供前端更新的历史记录载荷。当记录存在非空 `event_unit` 时，载荷 MUST 包含 `eventUnit` 字段且值与数据库一致。

#### Scenario: 新增历史

- **WHEN** 新增一条 history 成功提交且 `event_unit=ml`
- **THEN** 系统 SHALL PUBLISH 一条 action 为 create 的消息且包含新记录标识、展示所需字段及 `eventUnit=ml`

#### Scenario: 更新或删除历史

- **WHEN** 更新或删除 history 成功提交
- **THEN** 系统 SHALL PUBLISH 对应 update 或 delete 的消息且包含受影响记录的主键或 event 关联信息；update 时 MUST 包含最新 `eventUnit`

### Requirement: CUD 后失效 piece 缓存

history-service SHALL 在 history 表发生增删改并成功提交后，使与该 device_no（及必要时 eventId）相关的 piece 缓存失效，以保证后续 piece 查询不返回陈旧数据。

#### Scenario: 写入后查询一致

- **WHEN** 同一 device_no 在写入后立刻发起 piece 查询
- **THEN** 查询结果 SHALL 反映刚写入的数据（通过失效缓存或直接读库达成）

### Requirement: History 投影事件 MUST 携带 event_unit

`history.record.created` / `history.record.updated` 域 outbox 与 Redis 缓存投影消费载荷 MUST 包含 `event_unit`（或等价 JSON 字段），与权威库 `history.event_unit` 一致；投影写入 Redis 读模型时 MUST 保留 `EventUnit`。

#### Scenario: 创建投影含单位

- **WHEN** outbox 发布 `history.record.created` 且对应行 `event_unit=次`
- **THEN** 投影载荷 MUST 含 `event_unit=次`，且 Redis 列表缓存中该条记录的 `eventUnit` SHALL 为 `次`

### Requirement: 语音结束事件 MUST 与 API 结束等价推送 WS

当用户通过语音链路触发「结束」计时类事件（`ActionTargetTypeEnd` / Python `target_type=feeding` 且 `action=end`）且系统向用户播报结束成功时，voice-service MUST 经 `DeviceHistory` 契约完成 history 写库，且该写库 MUST 在 history-service 侧触发与 App 调用 `POST /device/history/api/event/end-latest`（`updated=true`）或等价 `event/update` 相同的 Redis 通知：向 `app:history:notify` PUBLISH，`action` 为 `update` 或 `create`（与实落库操作一致），载荷含 `deviceNo`、记录主键及展示字段。

`EndLatestHistoryIfMatch` / `end-latest` 的匹配权威 MUST 为：该 `deviceNo` 下指定 `eventId` 且 `end_time=0` 的最近一条记录（按 `id` 降序）；MUST NOT 要求该行同时是设备全局最新一条 history。

voice-service MUST NOT 在 `EndLatestHistoryIfMatch` 返回 `updated=false` 且未执行成功降级写库的情况下向用户播报结束成功。

#### Scenario: 结束最近一条同 eventId 未闭合计时记录

- **WHEN** 用户语音或 App 结束事件 E，且 history-service 存在 `eventId=E` 且 `end_time=0` 的记录（取 `id` 最大者），即使全局最新一条 history 的 `eventId` 不等于 E
- **THEN** `EndLatestHistoryIfMatch` SHALL 返回 `updated=true`
- **AND** 系统 SHALL PUBLISH `action=update` 的 WS 通知
- **AND** 语音路径在写库成功后 MUST 表示结束成功

#### Scenario: EndLatest 未匹配时降级写库仍推送

- **WHEN** 用户语音结束事件 E，且该设备不存在 `eventId=E` 且 `end_time=0` 的记录，因而 `EndLatestHistoryIfMatch` 对 E 返回 `updated=false`
- **THEN** voice-service MUST 执行降级写库（至少 `AddHistory` 写入 E 的瞬时结束记录）
- **AND** 降级写库成功后系统 SHALL PUBLISH `action=create` 或 `update` 的 WS 通知
- **AND** 语音回复 MUST 仅在至少一次写库成功后可表示结束成功

#### Scenario: 与 App end-latest 行为一致

- **WHEN** 同一 `deviceNo` 下 App 调用 `event/end-latest` 成功（`updated=true`）与语音结束同一未闭合 event 在 history 库产生相同终态
- **THEN** 两种路径触发的 WS 载荷在 `action`、记录 `id`、`eventId`、`endTime` 语义上 MUST 一致（允许 `remark` 来源不同）

---

## history-profile-nickname

<!-- source: openspec/specs/history-profile-nickname/spec.md -->

# history-profile-nickname Specification

## Purpose
TBD - created by archiving change wx-username-auth-and-history-nickname. Update Purpose after archive.
## Requirements
### Requirement: 历史画像接口返回昵称
系统 SHALL 扩展历史画像读取接口（`GET /device/history/api/birthday`）返回 `nickname` 字段；该字段 MUST 通过 device 画像契约获取，history-service SHALL NOT 直连查询 device 域库表。

#### Scenario: 已有昵称
- **WHEN** 目标设备存在可用昵称
- **THEN** 响应 SHALL 包含非空 `nickname`

#### Scenario: 无昵称
- **WHEN** 目标设备当前无昵称
- **THEN** 响应 SHALL 返回 `nickname` 为空串，且接口 SHALL 保持成功响应

### Requirement: 历史页面展示昵称
系统 SHALL 在历史记录页面展示 `nickname`，并与既有性别展示共存；接口返回为空时页面 MUST 显示空态文案而非报错。

#### Scenario: 页面加载成功展示昵称
- **WHEN** 页面加载到包含 `nickname` 的画像数据
- **THEN** 页面 SHALL 显示昵称文本并维持原有性别主题逻辑

#### Scenario: 昵称为空时降级展示
- **WHEN** 接口返回 `nickname` 为空
- **THEN** 页面 SHALL 显示“未设置昵称”或等价占位，并 SHALL NOT 阻断其他历史数据渲染

---

## history-service-db-ownership

<!-- source: openspec/specs/history-service-db-ownership/spec.md -->

# history-service-db-ownership Specification

## Purpose
TBD - created by archiving change align-service-db-boundaries-history-voice-device. Update Purpose after archive.
## Requirements
### Requirement: history-service 进程 MUST 仅直连本域持久化表

`history-service`（及其独立配置所连接的默认数据库）MUST 仅对 `history` 与 `domain_outbox` 表执行 DAO/SQL 读写（不含只读副本或显式配置的跨库迁移工具）。MUST NOT 在 history 进程内对 `user`、`event`、`action`、`qa`、`suggest` 等他域业务表执行直连访问。

#### Scenario: 独立部署 history 库仅含 history 与 outbox

- **WHEN** 运行 `history-service` 且数据库中仅存在 `history` 与 `domain_outbox` 业务表
- **THEN** 服务 MUST 能完成历史记录与 outbox 相关功能，且 MUST NOT 因缺少他域表而依赖本地 DAO 回退直查

#### Scenario: 代码评审检查 history 包 import

- **WHEN** 评审 `internal/services/history` 或 history 进程入口的变更
- **THEN** MUST NOT 引入对 `dao.User`、`dao.Event`、`dao.Suggest`、`dao.Qa`、`dao.Action` 等他域 DAO 的直连依赖用于业务读写

### Requirement: 跨域数据 MUST 通过契约获取

当 history 域逻辑需要设备画像、事件字典或语音建议等非 history 表数据时，MUST 通过 **device-service / voice-service** 的 HTTP（或已批准的事件契约）获取，MUST NOT 在同一进程内直查他域表。

#### Scenario: 元数据或画像由其他服务提供

- **WHEN** 上层仍通过统一 `Contract` 需要「事件选项」或「生日」等能力
- **THEN** 实现 MUST 路由到对应服务客户端，而非 `history/local.go` 内对 `dao.Event` 或 `dao.User` 的查询

---

## history-voice-delegation

<!-- source: openspec/specs/history-voice-delegation/spec.md -->

# history-voice-delegation Specification

## Purpose
TBD - created by archiving change history-chat-delegate-voice. Update Purpose after archive.
## Requirements
### Requirement: history 文本 chat MUST HTTP 委派 voice-service

history-service 处理 `POST /device/history/api/chat` 时 MUST 经 HTTP 调用 voice-service internal 文本 chat 契约，MUST NOT 在 history 进程内 import 或调用 `voice.TextChat` / `voice.Voice()` 执行业务。

history-service MUST NOT 配置 `VOICE_DB_LINK` 或访问 voice 库表（含 `ai_quota_default`、`qa`、`suggest`、`llm_lane_config`）。

对外路径、请求体字段（`deviceNo`、`transcript`）与成功响应结构 MUST 保持不变。

#### Scenario: App chat 经委派完成

- **WHEN** App 调用 `POST /device/history/api/chat` 且 voice-service 可达
- **THEN** history-service MUST 向 voice-service 发起 internal text chat HTTP 请求并返回等价 `reply`

#### Scenario: 额度错误透传

- **WHEN** voice-service internal chat 返回 40301 或 40302
- **THEN** history-service 对外响应 MUST 携带相同 business code 与 message 语义

#### Scenario: 无 voice 库配置

- **WHEN** history-service 进程未配置 voice 数据库连接
- **THEN** chat 路径 MUST 仍可工作（依赖 voice-service HTTP，不依赖本地 voice 库）

---

## hms-notification-payload

<!-- source: openspec/specs/hms-notification-payload/spec.md -->

# hms-notification-payload Specification

## Purpose
TBD - created by archiving change fix-hms-notification-payload. Update Purpose after archive.
## Requirements
### Requirement: HMS 可见通知消息载荷

当 push-service 经 HMS channel 发送**非静默**且 `alert` 非空的推送时，请求体 MUST 包含顶层 `message.notification`（至少含 `title` 与 `body`），并且 MUST 在 `message.android.notification` 中提供合法的 `click_action`（默认打开应用：`type` 为 `3`；若配置了自定义 intent，则 `type` 为 `1` 且带 `intent`）。

系统 MUST NOT 仅依赖「只有 `android.data`、无顶层 notification」的形态作为可见通知的主路径。

#### Scenario: 预测临近可见推送

- **WHEN** `bizType=predict_imminent`（或其它非静默+非空 alert）经 HMS 下发
- **THEN** 发往华为的 JSON MUST 含 `message.notification.title` 与 `message.notification.body`，且 `message.android.notification.click_action` 合法

#### Scenario: 静默不伪造通知正文

- **WHEN** 推送为 silent 或 alert 为空
- **THEN** 系统 MUST NOT 以空/占位正文冒充完整用户可见通知消息（可仅 badge/data）

### Requirement: HMS 业务成功码判定

HMS `messages:send` 在 HTTP 2xx 时，系统 MUST 解析响应体业务字段 `code`：仅当成功码为 `80000000`（字符串或与之等价的表示）时，方可视为发送成功并记 `send_ok`；否则 MUST 视为失败并记 `send_failed`（日志含 `code` 与截断错误信息，MUST NOT 含完整 device token）。

#### Scenario: HTTP 200 但业务失败

- **WHEN** 华为返回 HTTP 200 且 `code` 非 `80000000`
- **THEN** dispatcher 路径 MUST 走失败日志（非 `send_ok`）

#### Scenario: 业务成功

- **WHEN** 华为返回 HTTP 2xx 且 `code` 为 `80000000`
- **THEN** 系统 MAY 记 `send_ok`

---

## intent-clarify-conversation-id

<!-- source: openspec/specs/intent-clarify-conversation-id/spec.md -->

# intent-clarify-conversation-id Specification

## Purpose
TBD - created by archiving change intent-clarify-conversation-id. Update Purpose after archive.
## Requirements
### Requirement: 意图请求可携带 conversation_id 续聊

voice 调用 Python `/v1/analyze/intent` 与 `/v1/analyze/intent/stream` 时，若该设备存在本地保存的澄清 `conversation_id`，MUST 在请求体中附带该字段与用户原文；若不存在则 MUST NOT 伪造。流式与非流式 MUST 使用同一请求字段契约。

#### Scenario: 有 pending cid 时附带续聊

- **WHEN** 设备本地存有上一轮 `need_confirm` 返回的 `conversation_id`
- **AND** 用户再次提交文本进入统一意图路径（流式或非流式）
- **THEN** voice SHALL 调用对应 intent 接口且请求包含该 `conversation_id` 与用户文本
- **AND** SHALL NOT 调用 `/v1/analyze/intent/confirm`

#### Scenario: 无 pending cid 时冷启动

- **WHEN** 设备本地无 `conversation_id`
- **THEN** voice SHALL 发起不含 `conversation_id`（或为空省略）的意图分析请求

### Requirement: need_confirm 仅保存 cid 并透传自然语言

当意图结果 `need_confirm=true` 时，voice MUST 按设备保存返回的 `conversation_id`，MUST 将 Python 返回的自然语言（优先 `confirm_message`，否则 `content`）作为对用户回复，MUST NOT 执行喂养事件 CRUD，MUST NOT 在 Go 侧将用户话术解析为 `confirm|reject`。

#### Scenario: 澄清问句直接回传

- **WHEN** Python 意图结果 `need_confirm=true` 且带有澄清话术与 `conversation_id`
- **THEN** voice SHALL 保存该 `conversation_id`
- **AND** SHALL 以该自然语言作为 Reply 返回给调用方
- **AND** SHALL NOT 写入喂养 history/事件落库路径

### Requirement: 非确认结果清理 cid 后走统一落库或闲聊

当意图结果 `need_confirm=false`（或不需要确认）时，voice MUST 清除该设备本地 `conversation_id`，随后 MUST：`target_type=conversation` 或空 target 仅透传 Python 自然语言回复；喂养/多事件写库 MUST 已由 Python 在同轮意图分析中经 history `event/batch` 完成，Go `applyUnifiedIntentResult` MUST NOT 再执行 Go 侧事件 CRUD。

#### Scenario: 澄清后落地喂养

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且 `target_type=feeding` 且 batch 已成功写库
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 透传 Python 返回的回复文本

#### Scenario: 澄清后闲聊回复

- **WHEN** 续聊轮 Python 返回 `need_confirm=false` 且目标为 conversation 或空 target
- **THEN** voice SHALL 清除本地 `conversation_id`
- **AND** SHALL 仅回复自然语言，不走 Go 侧喂养落库

### Requirement: 删除 ConfirmIntent 与 parseConfirmFeedback 主路径

voice MUST NOT 再提供或调用 `ConfirmIntent`（含对 `/v1/analyze/intent/confirm` 的 HTTP 封装），MUST NOT 使用 `parseConfirmFeedback` 将用户输入映射为 `confirm|reject` 以驱动澄清。`prepareChatPreamble` MUST NOT 因 pending 澄清而短路并绕过统一 intent 调用（`pending child` 本域逻辑除外）。

#### Scenario: 代码路径无 confirm 接口

- **WHEN** 审查 voice 意图澄清相关实现
- **THEN** 不存在对 `/v1/analyze/intent/confirm` 的调用
- **AND** 不存在将用户文本解析为仅 `confirm|reject` 后调用专用恢复接口的主路径

### Requirement: 流式与非流式行为矩阵一致

`chatWithResult`（非流式）与 `HandleTranscriptForIntentStream`（流式落地）MUST 共享同一澄清语义：同一 cid 读写规则、同一 `need_confirm` 透传与清 cid、同一 CRUD 边界；澄清轮 MUST 发起对应的 AnalyzeIntent / AnalyzeIntentStream（而非 ConfirmIntent）。

#### Scenario: 非流式澄清续聊

- **WHEN** 非流式路径上一轮 `need_confirm=true` 后用户再次输入
- **THEN** voice SHALL 经非流式 `AnalyzeIntent` 附带 `conversation_id` 续聊并按结果落地

#### Scenario: 流式澄清续聊

- **WHEN** 流式落地路径上一轮 `need_confirm=true` 后用户再次输入
- **THEN** voice SHALL 经 `AnalyzeIntentStream` 附带 `conversation_id` 续聊并按结果落地
- **AND** SHALL NOT 为澄清轮回退到 ConfirmIntent

### Requirement: 独立意图调用不得串带喂养澄清 cid

成长建议、历史问答等不经过统一 chat preamble 的独立 `AnalyzeIntent` 调用 MUST NOT 自动附带喂养澄清用的本地 `conversation_id`。

#### Scenario: 成长建议不携带澄清 cid

- **WHEN** 调用成长建议类独立意图分析
- **THEN** 请求 SHALL NOT 附带喂养澄清 pending 的 `conversation_id`

### Requirement: 澄清续聊轮免计 AI quota

当统一意图路径（流式或非流式）在发起 `AnalyzeIntent` / `AnalyzeIntentStream` **之前** 本地已存在有效澄清 `conversation_id` 时，voice MUST NOT 对该轮执行喂养 AI 额度预检扣减门槛失败拦截所依赖的计次语义，MUST NOT 在意图成功后对该轮执行额度 consume。冷启动（发起前无有效 cid）MUST 仍走常规 guard 与成功 consume（含结果为 `need_confirm=true` 的提问轮）。流式与非流式 MUST 使用同一免计判定。

#### Scenario: 带 cid 续聊免计次

- **WHEN** 设备本地存有有效澄清 `conversation_id`
- **AND** 用户再次进入统一意图路径并成功完成意图调用
- **THEN** voice SHALL NOT 对该轮执行 `consumeVoiceAIQuotaOnSuccess`（或等价扣减）
- **AND** SHALL 仍发起带 `conversation_id` 的意图分析并按结果落地

#### Scenario: 冷启动含首次澄清提问仍计次

- **WHEN** 设备本地无有效澄清 `conversation_id`
- **AND** 意图结果为 `need_confirm=true`
- **THEN** voice SHALL 按冷启动规则对该轮执行成功后的额度 consume（与普通意图一致）

#### Scenario: 额度用尽仍可澄清续聊

- **WHEN** 用户 AI 额度已用尽
- **AND** 设备本地仍有有效澄清 `conversation_id`
- **THEN** voice SHALL 允许该续聊轮继续调用意图接口（不得因额度预检阻断澄清完成）

---

## invite-code-identity

<!-- source: openspec/specs/invite-code-identity/spec.md -->

# invite-code-identity Specification

## Purpose
TBD - created by archiving change invite-peer-force-ucg. Update Purpose after archive.
## Requirements
### Requirement: 用户邀请码由 cash Ensure 持有

系统 MUST 在 cash 库为每个有效 `owner_wx_id` 最多保留一条邀请码（一用户一码）。系统 MUST 在用户注册完成后触发 Ensure，或在 App 首次读取「我的邀请码」时懒创建。邀请码 MUST NOT 写入 device `wx` 表。邀请码 MUST NOT 要求有效期、码级可开功能范围或 grant 时长才能用于兑换。

#### Scenario: 首次读取生成

- **WHEN** 某 wx 尚无邀请码且请求我的邀请码
- **THEN** 系统创建唯一码并返回，且 `redeemedCount` 为 0

#### Scenario: 重复 Ensure 幂等

- **WHEN** 同一 wx 再次 Ensure
- **THEN** 系统返回同一码且 MUST NOT 插入第二行

### Requirement: 兑码去重为人×码×功能

系统兑换邀请码时 MUST 拒绝自用（码主人 `owner_wx_id` 与兑换者 `redeemer_wx_id` 相同）。系统 MUST 拒绝码主人**当前**绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换（同一宝宝下不同账号互兑）。系统 MUST 以 `(redeemer_wx_id, code, feature_id)` 唯一约束防止重复兑换。系统 MUST 允许同一兑换者使用不同 owner 的码（且双方设备号不同时）。系统 MUST NOT 再执行「一家锁定」（redeemer 绑定单一 owner）。系统 MUST 仅当目标 `feature_def` 启用且 `unlock_methods` 含 `invite_code` 时允许兑换。码主人设备号 MUST 经 device-service 契约按 `owner_wx_id` 查询，MUST NOT 由 cash 直查 device 库。当 device 查询失败时，系统 MUST fail-closed 拒绝本次兑换。当查询成功但主人为空 `device_no`（未绑机）时，系统 MUST NOT 仅因同设备规则拒绝。

#### Scenario: 多好友码累加预测

- **WHEN** 用户依次兑换好友 A、B 的码且 featureId 为 `prediction_unlock`，且兑换者与 A、B 当前绑定设备号均不同
- **THEN** 两次均成功且该设备预测永久条数各 +1

#### Scenario: 不同设备互兑

- **WHEN** A 兑 B 的码开通某功能且 B 兑 A 的码开通同一功能，且双方当前绑定 `device_no` 不同
- **THEN** 两次均成功

#### Scenario: 同设备拒绝兑码

- **WHEN** 兑换者当前 `device_no` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换，MUST NOT 写入开通或获客原力

#### Scenario: 主人未绑机可兑

- **WHEN** 码主人经 device 查询得到空 `device_no`，且非自用、其它校验通过
- **THEN** 系统 MUST NOT 因同设备规则拒绝

#### Scenario: device 查询失败 fail-closed

- **WHEN** cash 调用 device 按主人 wxId 取 `device_no` 失败
- **THEN** 系统 MUST 拒绝本次兑换

#### Scenario: 同码同功能重复

- **WHEN** 同一用户对同一码同一 featureId 再次兑换
- **THEN** 系统拒绝

### Requirement: App 可读我的码与邀请列表

系统 MUST 提供 App 接口返回当前用户邀请码与获客数量。系统 MUST 提供 App 接口返回成功使用该码的用户列表（含 UCG 展示昵称与兑换时间）。系统 MUST NOT 再提供面向运营的邀请码造码 Admin 页面。

#### Scenario: 邀请列表

- **WHEN** owner 查询 invitees
- **THEN** 每条含昵称与 `redeemedAt`，且仅含成功兑换记录

---

## invite-group-qr-admin

<!-- source: openspec/specs/invite-group-qr-admin/spec.md -->

# invite-group-qr-admin Specification

## Purpose
TBD - created by archiving change invite-group-qr. Update Purpose after archive.
## Requirements
### Requirement: Admin can upload er_code.png to apk storage

The system MUST accept an Admin multipart upload that saves the image as exactly `er_code.png` under the gateway APK storage directory, overwriting any previous file. The system MUST reject uploads that are not a PNG when the implementation chooses reject-non-png. The download MUST remain available at `/device/app/apk/er_code.png` using the existing apk download route. The upload path MUST be classified as a Gateway Admin API path (Admin JWT accepted; not treated as an App user API).

#### Scenario: Overwrite

- **WHEN** Admin uploads a new PNG as the group QR
- **THEN** the storage path contains only `er_code.png` as the group QR asset (prior content replaced)

#### Scenario: Admin JWT accepted on upload

- **WHEN** Admin calls the group QR upload endpoint with a valid Admin JWT
- **THEN** the request MUST NOT be rejected as “admin token 不能访问 App 用户接口”

### Requirement: Admin configures expiry and can preview after expiry

The system MUST allow Admin to set `expires_at` for the group QR in cash. When `expires_at` is 0 or in the past, App catalog MUST NOT expose the URL, but Admin preview MUST still be able to load the current `er_code.png`.

#### Scenario: Expired still previewable

- **WHEN** `expires_at` is in the past
- **THEN** Admin UI can still preview the image and see the expiry, and App catalog omits `inviteGroupQrUrl`

---

## invite-group-qr-catalog

<!-- source: openspec/specs/invite-group-qr-catalog/spec.md -->

# invite-group-qr-catalog Specification

## Purpose
TBD - created by archiving change invite-group-qr. Update Purpose after archive.
## Requirements
### Requirement: Catalog returns inviteGroupQrUrl only while valid

When synthesizing `GET /cash/app/api/feature/catalog`, the system MUST include top-level `inviteGroupQrUrl` only when `expires_at > 0` and `expires_at > now`. The URL MUST point at the apk download path for `er_code.png` and SHOULD include a cache-busting query derived from `updated_at`. When invalid, the field MUST be omitted or empty.

#### Scenario: Valid window

- **WHEN** expiry is in the future and metadata exists
- **THEN** catalog data contains non-empty `inviteGroupQrUrl` referencing `er_code.png`

#### Scenario: Expired or zero

- **WHEN** `expires_at` is 0 or not after now
- **THEN** catalog MUST NOT present a usable invite group QR URL to the App

---

## lane-free-model

<!-- source: openspec/specs/lane-free-model/spec.md -->

# lane-free-model Specification

## Purpose
TBD - created by archiving change vip-quota-joint-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 业务 lane 可配置额度不足模型

对 lane `voiceUnderstanding`、`clinic`、`careAlert`、`polish`，Admin 配置 MUST 支持可选的免费/额度不足模型字段（`freeProvider`、`freeModel`，可为空）。非 premium 选模 MUST 读取上述字段：非空则作为下游模型；为空则 Go MUST omit 模型配置并由 Python（或 polish 侧约定的无模型覆盖语义）自行决定免费模型。Sim 四条 lane（`simText`/`simVision`/`simImageGen`/`simVideoGen`）MUST NOT 要求或持久化 free 字段。

#### Scenario: 保存 clinic free 模型

- **WHEN** 管理员 PUT llm-lanes 且 `clinic.freeProvider`/`freeModel` 为允许列表内有效值
- **THEN** 系统 MUST 持久化该 free 配置，供后续非 premium 诊所/tip 路径使用

#### Scenario: 清空 free 表示交 Python

- **WHEN** 管理员将某业务 lane 的 freeProvider 与 freeModel 存为空
- **THEN** 非 premium 且走 Python 的请求 MUST 省略 model 字段

#### Scenario: Sim 无 free

- **WHEN** 管理员 GET/PUT sim llm-lanes
- **THEN** API MUST NOT 要求客户端提交 free 字段；即便误传也 MUST NOT 改变 Sim 运行时选模契约（本变更不引入 Sim 额度策略）

### Requirement: 正式模型与 free 模型分离

lane 的正式 `provider`/`model` MUST 仅用于 premium 路径；MUST NOT 在非 premium 路径回退使用正式模型（除非 free 显式配置为相同值）。硬编码 `DegradedVoiceUnderstandingProfile` / `DegradedClinicProfile` / `DegradedPolishProfile` MUST NOT 再作为生产选模真相源。

#### Scenario: 额尽不再使用代码种子智谱

- **WHEN** 非 VIP 用户某 feature 额度用尽且该 lane free 配置指向运维设定的模型 A
- **THEN** 下游 MUST 使用模型 A，MUST NOT 忽略 free 而改用代码内 Degraded* 常量

---

## llm-lane-admin

<!-- source: openspec/specs/llm-lane-admin/spec.md -->

# llm-lane-admin Specification

## Purpose
TBD - created by archiving change llm-lane-gate-admin. Update Purpose after archive.
## Requirements
### Requirement: voice-service SHALL 提供 LLM lane Admin API

voice-service MUST 提供 `GET /voice/admin/api/llm-lanes` 与 `PUT /voice/admin/api/llm-lanes`，认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（经 gateway-app 反代）。响应与请求 MUST 包含 `voiceUnderstanding` 与 `clinic` 两个子对象，各含 `provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。GET MUST 返回 provider→model allowlist 供 Admin UI 下拉联动。PUT MUST 校验 allowlist 与正整数边界（`maxInFlight>=1`，`maxWaiters>=0`）。

#### Scenario: 管理员读取 lane 配置

- **WHEN** 运维携带正确口令 GET `/voice/admin/api/llm-lanes`
- **THEN** 响应 SHALL 含两 lane 当前 DB 配置与 allowlist

#### Scenario: 管理员更新 clinic 并发

- **WHEN** 运维 PUT `clinic.maxInFlight=5` 且 model 在 allowlist 内
- **THEN** voice-service MUST 持久化至 `ai_voice_voice` 且 MUST 失效 lane 缓存

#### Scenario: 口令错误拒绝

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 MUST 返回未授权且 MUST NOT 修改配置

### Requirement: ucg-service SHALL 扩展 AI 配置 Admin 以包含 polish lane 闸门

`GET/PUT /ucg/admin/api/ai-config` MUST 扩展支持 `provider`、`maxInFlight`、`maxWaiters`（与现有 `visionModel` / `maxImagesPerRequest` 并存或语义对齐为 polish lane 的 `model`）。PUT 成功后 MUST 失效 AI/lane 缓存（与现有 `InvalidateAIConfigCache` 一致）。

#### Scenario: 管理员更新润笔模型与缓冲池

- **WHEN** 运维 PUT `provider=zhipu`、`visionModel=glm-4.6v-flash`、`maxInFlight=1`、`maxWaiters=15`
- **THEN** ucg-service MUST 持久化且下一笔润笔 MUST 使用新 profile

### Requirement: DB 种子 MUST 默认为智谱三模型（方案 A）

首次 EnsureDefaultRows（或 migration 种子）时：若 lane 配置行不存在，系统 MUST 写入：`voiceUnderstanding` → `zhipu` / `glm-4.7-flash` / `maxInFlight=1` / `maxWaiters=20`；`clinic` → `zhipu` / `glm-4.1v-thinking-flash` / `maxInFlight=1` / `maxWaiters=10`；`polish` → `zhipu` / `glm-4.6v-flash` / `maxInFlight=1` / `maxWaiters=15`。YAML 中 DeepSeek/DashScope 默认值 MUST 保留作 DB 缺失时的冷启动兜底。

#### Scenario: 新环境首次启动

- **WHEN** voice 与 ucg 库尚无 lane 配置行
- **THEN** EnsureDefaultRows 后 Admin GET MUST 返回上述智谱默认

### Requirement: voice-admin 与 ucg-admin SHALL 提供 LLM lane 配置 UI

LLM lane 的模型与并发 Admin UI MUST 集中在 **`/device/admin/ai-model-admin.html`**（见 `ai-model-admin-ui`）。`voice-admin.html` 与 `ucg-admin.html` MUST NOT 再提供 LLM lane 编辑 Tab 或表单；MUST 仅链至统一页。后端 API（`GET/PUT /voice/admin/api/llm-lanes`、`GET/PUT /ucg/admin/api/ai-config` 之 polish 字段）MUST 不变，由统一页调用。

#### Scenario: voice LLM 经统一页保存

- **WHEN** 运维在 ai-model-admin 修改 voiceUnderstanding 并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes` 且 voice-admin MUST NOT 含重复编辑 UI

#### Scenario: ucg polish 经统一页保存

- **WHEN** 运维在 ai-model-admin 修改 polish maxWaiters 并保存
- **THEN** 页面 MUST 调用 PUT `/ucg/admin/api/ai-config` 且 ucg-admin MUST NOT 含重复编辑 UI

### Requirement: LLM lane Admin API MUST NOT 计入 App usage 统计

`/voice/admin/api/llm-lanes` 与扩展后的 `/ucg/admin/api/ai-config` 为运维型 Admin API，MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: Admin 保存不计入 usage

- **WHEN** 管理员 PUT llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增该路径计数

### Requirement: Admin PUT 失效 lane 缓存 MUST NOT 导致进程崩溃

Admin 成功 PUT `/ucg/admin/api/ai-config` 或 `/voice/admin/api/llm-lanes` 后，系统 MUST 失效进程内 lane/profile 缓存以使下一笔 LLM 调用读取新配置；该失效路径 MUST NOT 因 `InvalidateLaneCache` 与 `ProfileStore.InvalidateCache` 互相调用而导致 stack overflow 或进程退出。

#### Scenario: ucg PUT ai-config 后进程保持运行

- **WHEN** 运维携带有效口令 PUT `/ucg/admin/api/ai-config` 且持久化成功
- **THEN** ucg-service 进程 MUST 保持运行且后续 GET 同一接口 MUST 返回更新后的配置

#### Scenario: voice PUT llm-lanes 后进程保持运行

- **WHEN** 运维携带有效口令 PUT `/voice/admin/api/llm-lanes` 且持久化成功
- **THEN** voice-service 进程 MUST 保持运行且后续 GET MUST 返回更新后的 lane 配置

### Requirement: voice llm-lanes 包含 careAlert 与 free 字段

`GET/PUT /voice/admin/api/llm-lanes` MUST 包含三个业务子对象：`voiceUnderstanding`、`clinic`、`careAlert`。每一子对象 MUST 含正式 `provider`/`model`/`maxInFlight`/`maxWaiters` 以及可选 `freeProvider`/`freeModel`（允许空字符串）。PUT MUST 校验正式模型 allowlist；当 free 非空时 MUST 同样校验 allowlist；free 全空 MUST 合法。

#### Scenario: PUT 写入 careAlert 与 free

- **WHEN** 管理员提交含 `careAlert` 正式模与空 free 的合法 PUT
- **THEN** 系统 MUST 持久化成功，后续 GET MUST 读出一致字段

#### Scenario: 非法 free 模型拒绝

- **WHEN** freeProvider/freeModel 非空但不在 allowlist
- **THEN** PUT MUST 失败且 MUST NOT 部分写入该 lane

### Requirement: ucg polish lane 支持 free 字段

ucg 侧 polish lane Admin 读写 MUST 支持与 voice 业务 lane 同语义的 `freeProvider`/`freeModel`（可空）。

#### Scenario: 保存 polish free

- **WHEN** 管理员保存 polish free 为允许列表内模型
- **THEN** 后续非 premium 润笔 MUST 能读到该配置

---

## llm-lane-gate

<!-- source: openspec/specs/llm-lane-gate/spec.md -->

# llm-lane-gate Specification

## Purpose
TBD - created by archiving change aimodel-thinking-disabled-default. Update Purpose after archive.
## Requirements
### Requirement: aimodel 统一入口 SHALL 默认关闭上游 thinking

凡经 `aimodel.Invoke` / `InvokeStream` / `InvokeWithHeldProfile` / `InvokeStreamWithHeldProfile` 发起的 chat/completions 请求，若调用方未将 `ChatRequest.ThinkingEnabled` 设为 `true`，aimodel MUST 按 provider 规则关闭 thinking（智谱 MUST 显式 `disabled`）。该默认语义 MUST 适用于所有 lane（含 `voiceUnderstanding`、`polish`、`simText`、`simVision` 等），MUST NOT 依赖上游模型默认 thinking 行为。

#### Scenario: 润笔未 opt-in

- **WHEN** `PolishPostText` 调用 `aimodel.Invoke(LanePolish, ...)` 且未设置 `ThinkingEnabled`
- **THEN** 智谱请求 MUST 显式 `thinking.type=disabled`

#### Scenario: voiceUnderstanding 闲聊

- **WHEN** 喂养语音 LLM 经 `LaneVoiceUnderstanding` 调用且未设置 `ThinkingEnabled`
- **THEN** 对智谱 provider 的请求 MUST 显式 `thinking.type=disabled`

### Requirement: 系统 SHALL 按 model 维度实施 Redis LLM 并发闸门

系统 MUST 为每条 LLM lane 的 **当前 `profile.model`** 维护独立 Redis 并发闸门。闸门 MUST 支持可配置 `maxInFlight`（同时占用上游 API 的槽位数，默认 1）与 `maxWaiters`（允许排队等待的请求数，缓冲池）。`maxInFlight` MUST 为不小于 1 的整数且 MUST NOT 在代码中写死为 1。闸门键 MUST 使用规范化 model 名（小写、trim），格式为 `ai:llm:gate:{model}:inflight` 与 `ai:llm:gate:{model}:waiting`（或 design 批准的等价原子语义）。不同 model MUST NOT 共用同一闸门池。

#### Scenario: 同 model 并发互斥

- **WHEN** `voiceUnderstanding` lane 配置 `model=glm-4.7-flash` 且 `maxInFlight=1`，且已有一条请求占用槽位
- **THEN** 第二条同 model 请求 MUST 进入等待队列（若 `waiting < maxWaiters`）且 MUST NOT 在上一条释放前发起上游 HTTP

#### Scenario: 不同 model 可并行

- **WHEN** `glm-4.7-flash` 槽位被 voiceUnderstanding 占用，且 clinic lane 配置 `model=glm-4.1v-thinking-flash`
- **THEN** clinic 请求 MUST 使用独立闸门池且 MAY 并行调用上游

#### Scenario: 换 model 使用新池

- **WHEN** Admin 将 `voiceUnderstanding.model` 从 `glm-4.7-flash` 改为 `glm-4.7-flashx`
- **THEN** 新请求 MUST 使用 `glm-4.7-flashx` 闸门键且 MUST NOT 与旧 model 池共享 inflight 计数

### Requirement: 等待队列满时 MUST 立即拒绝且不得调用上游

当 `waiting >= maxWaiters` 时，系统 MUST 在业务入口立即返回错误，message **「当前队列已满，请稍后重试」**，code **50301**。该路径 MUST NOT 调用上游 LLM API，且 MUST NOT consume 月度 AI 额度（`voice_ai` / `clinic_ai` / `polish`）。

#### Scenario: 润笔队列满

- **WHEN** `polish` lane 的 `glm-4.6v-flash` 缓冲池已满且用户请求 `POST /ucg/app/api/posts/polish`
- **THEN** API MUST 返回 50301 且 MUST NOT 请求上游

#### Scenario: 喂养语音队列满

- **WHEN** `voiceUnderstanding` lane 缓冲池已满且用户 commit 后将触发 LLM
- **THEN** voice WS MUST 返回 `error` 帧 code 50301 且 MUST NOT 调用 LLM

#### Scenario: 胖宝队列满

- **WHEN** `clinic` lane 缓冲池已满且客户端发送合法 `question`
- **THEN** clinic WS MUST 返回 `error` 帧 code 50301 且 MUST NOT 调用 LLM

### Requirement: aimodel 包 SHALL 提供 Lane 统一调用入口

`internal/services/aimodel` MUST 导出 lane 枚举 `voiceUnderstanding`、`clinic`、`polish` 及 `Invoke` / `InvokeStream`。业务代码 MUST 通过 lane 调用上游，MUST NOT 在业务层硬编码 provider endpoint 或 model 字符串。`Acquire` 成功至上游调用完全结束（含流式读毕或连接关闭）期间 MUST 持有闸门槽位，并在 `defer` 或等价路径释放。

#### Scenario: 闲聊流式持槽至流结束

- **WHEN** `streamCasualReplyWithBaiduTTS` 经 `InvokeStream(LaneVoiceUnderstanding)` 调用上游
- **THEN** 闸门槽位 MUST 从首个上游 HTTP 发起持有至 SSE 读取结束或 context 取消

### Requirement: 系统 SHALL 支持多 provider 适配器

aimodel MUST 支持至少三种 provider：`zhipu`、`deepseek`、`dashscope`。lane profile MUST 含 `provider` 字段；API Key MUST 仅从环境变量或进程配置读取，MUST NOT 存于 Admin DB。切换 provider MUST 仅需更新 lane profile（及对应 env 已配置），MUST NOT 修改业务 lane 调用点。

#### Scenario: Admin 切回 DeepSeek 喂养模型

- **WHEN** `voiceUnderstanding` profile 改为 `provider=deepseek`、`model=deepseek-chat` 且 `DEEPSEEK_API_KEY` 已配置
- **THEN** 下一笔喂养 LLM 请求 MUST 调用 DeepSeek endpoint 且 MUST 使用 `deepseek-chat` 闸门池

#### Scenario: provider 对应 key 缺失

- **WHEN** profile 为 `provider=zhipu` 但 `GLM_API_KEY` 未配置
- **THEN** 系统 MUST 返回明确配置错误且 MUST NOT 调用上游

### Requirement: voiceUnderstanding lane SHALL 覆盖喂养语音全部 LLM 路径

下列路径 MUST 经 `LaneVoiceUnderstanding` 调用上游：统一意图解析、实体/动作抽取、闲聊直答、闲聊流式 LLM 段、成长建议、历史问答，以及 `event_child_pending` 中的实体抽取。纯 ASR、纯 TTS、规则回复与模式切换（不触发 LLM）MUST NOT 经过该 lane。

#### Scenario: 成长建议走 voiceUnderstanding

- **WHEN** 用户触发成长建议且将调用 LLM
- **THEN** 系统 MUST 使用 `voiceUnderstanding` lane profile 的 model 与闸门

#### Scenario: 纯 ASR 不占用 LLM 闸门

- **WHEN** 用户仅进行语音转写且无 LLM 调用
- **THEN** 系统 MUST NOT 调用 `aimodel.Acquire`

### Requirement: careAlert 车道参与并发闸门

`careAlert` lane MUST 注册到与其它 lane 相同的 aimodel 并发闸门机制；`maxInFlight`/`maxWaiters` MUST 生效。同一 `provider+model` 跨 lane 共享池的既有语义 MUST 保持（含 free 模型若指向相同 provider+model）。

#### Scenario: careAlert 队列满

- **WHEN** careAlert 闸门等待队列已满且新的 premium 生成请求需要 Acquire
- **THEN** 服务 MUST 按既有 50301（或该路径等价错误）拒绝，MUST NOT 调用 Python 分析，MUST NOT consume `care_alert`

---

## logger-level-env

<!-- source: openspec/specs/logger-level-env/spec.md -->

# logger-level-env Specification

## Purpose
TBD - created by archiving change fix-logger-level-env-push-warn-logs. Update Purpose after archive.
## Requirements
### Requirement: GF_LOGGER_LEVEL 覆盖默认 logger

系统 MUST 在长期运行服务启动时，于 yaml logger 配置已应用到默认 logger 之后，读取环境变量 `GF_LOGGER_LEVEL`：若 trim 后非空，则对该默认 logger 调用与 GoFrame 一致的级别字符串设置（如 `SetLevelStr`），使其覆盖 yaml 中的 `logger.level`；若为空，则 MUST 保持 yaml 级别不变。

#### Scenario: 生产设为 prod

- **WHEN** 进程环境 `GF_LOGGER_LEVEL=prod` 且服务已调用 `loggercfg.ApplyFromEnv`
- **THEN** 默认 logger 仅输出 WARN / ERRO / CRIT（及框架强制保留级别），`Infof`/`Debugf` 不再出现在该进程日志中

#### Scenario: 留空沿用 yaml

- **WHEN** `GF_LOGGER_LEVEL` 未设置或为空字符串
- **THEN** 默认 logger 级别保持配置文件中的 `logger.level`（当前服务配置多为 `all`）

#### Scenario: 生效可观测

- **WHEN** 非空 `GF_LOGGER_LEVEL` 成功应用
- **THEN** 进程 MUST 打出一条 Warning 级日志，标明服务名与所应用级别字符串

#### Scenario: 非法级别

- **WHEN** `GF_LOGGER_LEVEL` 为 GoFrame 不识别的字符串
- **THEN** 系统 MUST 记录 Error 且 MUST NOT 静默成功；logger 级别 MUST 保持应用前状态（yaml 已加载的级别）

---

## main-config-boundary-pruning

<!-- source: openspec/specs/main-config-boundary-pruning/spec.md -->

# main-config-boundary-pruning Specification

## Purpose
TBD - created by archiving change service-dedicated-config-final-boundary. Update Purpose after archive.
## Requirements
### Requirement: 主配置 MUST 仅包含网关与全局公共配置
`manifest/config/config.yaml` MUST 仅保留 gateway 与全局共享配置项，不应包含仅属于 `voice-service`、`device-service`、`history-service` 的专属业务配置字段。

#### Scenario: 主配置检查
- **WHEN** 审查 `config.yaml` 字段归属
- **THEN** 所有服务专属字段 MUST 已迁移到对应服务专属配置文件

### Requirement: 删除主配置服务专属项 MUST 保持服务可启动
当主配置移除服务专属字段后，各服务 MUST 仍可通过其专属配置文件独立启动并完成依赖加载。

#### Scenario: 删除 voice 专属段后启动 voice-service
- **WHEN** 主配置不再包含 voice 专属业务配置项
- **THEN** `voice-service` MUST 通过自身配置文件正常启动且功能不缺失

---

## main-config-without-database

<!-- source: openspec/specs/main-config-without-database/spec.md -->

# main-config-without-database Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Main config SHALL not carry database settings
`manifest/config/config.yaml` MUST NOT contain `database.*` fields after worker dedicated configuration is introduced.

#### Scenario: Review main config fields
- **WHEN** auditing `manifest/config/config.yaml`
- **THEN** no database connection configuration MUST exist in the file

#### Scenario: Gateway runtime without DB dependency
- **WHEN** `gateway-service` starts with main config
- **THEN** gateway MUST run without requiring database fields from main config

---

## mcp-feeding-ws-text

<!-- source: openspec/specs/mcp-feeding-ws-text/spec.md -->

# mcp-feeding-ws-text Specification

## Purpose
TBD - created by archiving change unify-feeding-chat-ws. Update Purpose after archive.
## Requirements
### Requirement: MCP 喂养工具改经 chat WS 文模式

mcp-service 的喂养对话工具（现名 `baby_feeding_advisor` 或等价）MUST 通过 `/voice/chat/ws` 完成一轮对话：使用 text 输入与 text 输出模态，发送工具入参文本，读取服务端 `answer`（及可选 `thinking_delta`）作为工具返回。MUST NOT 再调用 `DelegateTextChat` 或 `/voice/internal/api/text/chat`。

#### Scenario: 工具调用成功返回回复

- **WHEN** 小智侧调用 MCP 喂养工具并传入非空 `transcript`
- **THEN** mcp-service MUST 对配置的设备建立（或复用）chat WS，以文模式提交该文本，并在收到业务 `answer` 后将其作为工具结果返回

#### Scenario: 空 transcript 仍本地校验

- **WHEN** 工具收到空或仅空白的 transcript
- **THEN** mcp-service MUST 在本地返回错误，MUST NOT 无意义地占用 WS 意图轮次

### Requirement: MCP 不再依赖 history 文本委派

mcp-service MUST NOT import 或调用 `histsvc.DelegateTextChat` / `DelegateTextChatStream` 作为喂养实现。进程依赖与 runbook MUST 改为描述 WS 连接所需基址/密钥（若有），而非 internal text HTTP。

#### Scenario: 无 DelegateTextChat 调用点

- **WHEN** 检索 mcpbridge / mcp-service 喂养工具实现
- **THEN** MUST NOT 存在对 `DelegateTextChat` 的调用

---

## microservice-boundary-final-alignment

<!-- source: openspec/specs/microservice-boundary-final-alignment/spec.md -->

# microservice-boundary-final-alignment Specification

## Purpose
TBD - created by archiving change service-dedicated-config-final-boundary. Update Purpose after archive.
## Requirements
### Requirement: 配置边界 MUST 与服务边界一致
系统 MUST 保证“服务职责边界、配置归属边界、运行入口边界”一致；任何服务不得通过共享主配置承担他域职责或访问路径。

#### Scenario: gateway 运行角色审查
- **WHEN** 审查 gateway 启动入口与配置项
- **THEN** gateway MUST 仅包含流量与策略层配置，不得加载他域业务执行配置

#### Scenario: voice 跨服务访问
- **WHEN** voice 需要获取 history/device 领域数据
- **THEN** voice MUST 通过服务契约访问，不得通过主配置回流到跨库直查实现

### Requirement: 最终形态迁移 MUST 包含可回滚路径
面向最终微服务形态的配置与边界收敛 MUST 提供清晰的分阶段切换与回滚策略，避免一次性切换导致生产不可用。

#### Scenario: canary 切换失败
- **WHEN** 配置切换到 canary/remote 后出现异常
- **THEN** 系统 MUST 支持按服务维度快速回滚到 local/上一版本配置

---

## microservice-ci-release

<!-- source: openspec/specs/microservice-ci-release/spec.md -->

# microservice-ci-release Specification

## Purpose
TBD - created by archiving change cash-service-acr-ci. Update Purpose after archive.
## Requirements
### Requirement: 新增微服务 MUST 同步 CI 与发版清单

凡变更引入或首次对外部署新的微服务进程（含 `cmd/*-service` 与对应 `Dockerfile.*`），OpenSpec proposal/design/tasks 与实现 MUST 同步更新下列发版清单，不得仅完成基线 Compose / 本地 Dockerfile：

1. **GitHub Actions**：至少 `.github/workflows/docker-acr.yml`（服务别名、`dockerfile_for`、`ALL_SERVICES`、全量服务数注释与错误提示）。
2. **环境 overlay**：`manifest/docker/docker-compose.microservices.test.yml` 与 `docker-compose.microservices.prod.yml` 中的 ACR 镜像、`container_name`（若适用）、ports（若服务暴露 HTTP）、networks。
3. **发布 runbook**：`docs/runbooks/release-deploy-and-run.md` 中全量服务数、构建范围别名表与相关说明。
4. **ACR 仓库**：提醒在目标命名空间创建与 matrix `id` 一致的仓库（运维操作，可记在 tasks）。

全局约定 MUST 写入 `openspec/project.md`（微服务 CI / ACR 发版约定），并作为评审硬性检查项。

#### Scenario: propose 新微服务时列出 CI 任务

- **WHEN** 某变更提案新增独立微服务进程
- **THEN** tasks MUST 含更新 `docker-acr.yml`、test/prod overlay 与 runbook 别名/服务数的可勾选项，且 `openspec/project.md` MUST 已存在对应强制约定可供引用

#### Scenario: 仅改基线 Compose 视为不完整

- **WHEN** 评审发现新服务仅出现在 `docker-compose.microservices.yml` 而未出现在 `docker-acr.yml` 的 `ALL_SERVICES`
- **THEN** 评审 MUST 要求补齐后再合并（或等价记录明确豁免理由）

### Requirement: cash-service 纳入 ACR 构建矩阵

`.github/workflows/docker-acr.yml` MUST 将 `cash-service` 列入全量构建列表，MUST 支持别名 `cash` 与 `cash-service` 解析为 matrix id `cash-service`，MUST 将 dockerfile 映射为 `manifest/docker/Dockerfile.cash-service`。全量构建服务数文案 MUST 与列表长度一致（本变更后为 10）。

#### Scenario: 全量 tag 构建含 cash-service

- **WHEN** 推送不含 `+` 后缀的预发布或正式 tag（或 workflow_dispatch 且 services 留空）
- **THEN** 构建 matrix MUST 包含 id `cash-service` 且使用 `Dockerfile.cash-service`

#### Scenario: 部分构建仅 cash

- **WHEN** git tag 带 `+cash` 后缀，或 workflow_dispatch 的 `services` 为 `cash`
- **THEN** CI MUST 仅 build/push `cash-service`（不强制 retag 其他服务）

### Requirement: cash-service 纳入 test/prod Compose overlay

测试与生产 overlay MUST 声明服务 `cash-service`，镜像为 `${REGISTRY}/cash-service:${IMAGE_TAG}`，并加入对应外部网络。test overlay MUST 映射宿主机端口 `19807`→容器 `9807`；prod overlay MUST 映射 `9807`→`9807`。业务环境变量仍由基线 compose + env-file 注入，overlay MUST NOT 重复书写业务 env（与现有 overlay 惯例一致）。

#### Scenario: 测试栈可 pull cash 镜像

- **WHEN** 使用 `.env.test` 与 microservices + test overlay 执行 `compose pull`
- **THEN** MUST 能解析并拉取 `${REGISTRY}/cash-service:${IMAGE_TAG}`（在 ACR 已存在该 tag 的前提下）

#### Scenario: 生产栈端口不与测试冲突

- **WHEN** 同机分别使用 test / prod overlay 启动
- **THEN** cash-service 宿主机端口 MUST 分别为 `19807` 与 `9807`，MUST NOT 互相占用

---

## notify-service-runtime

<!-- source: openspec/specs/notify-service-runtime/spec.md -->

# notify-service-runtime Specification

## Purpose
TBD - created by archiving change notify-service-rename-ci. Update Purpose after archive.
## Requirements
### Requirement: notify-service SHALL run as dedicated microservice

平台 SHALL 提供 **`notify-service`** 进程（原 app-status 维护通知能力），监听 `NOTIFY_SERVICE_ADDR` 默认 `:9806`，加载 `manifest/config/config.notify-service.yaml`。进程 **MUST NOT** 依赖 MySQL、Redis 或 RabbitMQ 启动探活。Docker 镜像名与 ACR 仓库名 **MUST** 为 `notify-service`。

#### Scenario: 启动与配置隔离

- **WHEN** 启动 notify-service 且未设置 `GF_GCFG_FILE`
- **THEN** 进程 SHALL 加载 `config.notify-service.yaml`，且配置 **MUST NOT** 含 `database` 段

#### Scenario: HTTP 契约保持不变

- **WHEN** 客户端 `GET /app/api/status/banner`
- **THEN** 路径与响应语义 SHALL 与 rename 前一致

### Requirement: docker-acr workflow SHALL build notify-service image

`.github/workflows/docker-acr.yml` MUST 将 `notify-service` 纳入全量构建矩阵（与 gateway、ucg 等并列，共 8 服务）。构建别名 **MUST** 接受 `notify` 与 `notify-service`，映射 Dockerfile `manifest/docker/Dockerfile.notify-service`。Tag `vX.Y.Z+notify` **MUST** 仅构建并 push `notify-service` 镜像。

#### Scenario: 全量 tag 含 notify-service

- **WHEN** push tag `v1.0.0-rc.1`（无 `+` 后缀）
- **THEN** CI matrix **MUST** 包含 `notify-service` 且 push `${REGISTRY}/notify-service:${tag}`

#### Scenario: 单服务 notify 构建

- **WHEN** push tag `v1.0.0-rc.2+notify` 或 workflow_dispatch `services=notify`
- **THEN** CI **MUST** 仅 build/push `notify-service` 镜像

---

## pangbao-ai-clinic

<!-- source: openspec/specs/pangbao-ai-clinic/spec.md -->

# pangbao-ai-clinic Specification

## Purpose
TBD - created by archiving change add-pangbao-ai-clinic-room. Update Purpose after archive.
## Requirements
### Requirement: voice-service SHALL 提供胖宝 AI 诊室 WebSocket handler

`voice-service` MUST 在路径 `/voice/clinic/ws` 注册 WebSocket handler（`BindHandler`），处理经 gateway-app 透传而来的客户端文本提问并流式返回 **clinic lane 配置的上游 LLM** 回答（默认种子为智谱 `glm-4.1v-thinking-flash`，Admin 可切回 DeepSeek 等）。用户可见功能名称为 **胖宝诊疗**；实现与配置仍使用 `clinic` / `clinic_ai` / `voice:clinic:*` 命名。该路径为**集群内业务端点**；App 对外入口 MUST 为 gateway-app-server 同路径透传（见 `gateway-ws-edge-proxy`）。实现 MUST NOT 将连接注册到 `VoiceWSManager`。实现 MUST NOT 提供 TTS 或音频上行能力（MVP 纯文本）。

#### Scenario: 经 gateway-app 透传后握手成功

- **WHEN** 客户端经 gateway-app 透传对 voice-service `/voice/clinic/ws` 完成 WebSocket Upgrade
- **THEN** voice-service SHALL 接受连接并等待首帧 `auth` JSON

### Requirement: Clinic WebSocket SHALL 以 wxId 为主键绑定身份

`/voice/clinic/ws` 的连接、限流与额度维度 MUST 以 **`wx.id`（正整数）** 为主键，与 `/voice/chat/ws` 以 `deviceNo` 注册 `VoiceWSManager` 的行为 MUST 显式不同。实现 MUST NOT 使用 `deviceNo` 反查 wxId 作为 clinic 鉴权路径（即 MUST NOT 调用 `ResolveVoiceWxID` 的 deviceNo fallback）。`deviceNo` MUST 与 JWT `device_no` claim 一致，并 MUST 透传给 Python Clinic 请求的 `device_no`；MUST NOT 再用于 Go 侧 history 摘要或宝宝画像拉取。

#### Scenario: 首帧 auth 绑定 wxId

- **WHEN** 客户端握手后发送 `type=auth` 且 JWT 含 `sub=1001` 与有效 `device_no`
- **THEN** 服务端 SHALL 解析 `wxId=1001` 并返回 `auth_ok`
- **AND** 后续限流/额度 MUST 使用 wxId=1001

#### Scenario: 未登录拒绝

- **WHEN** 客户端 `auth` 帧 JWT 无效或 `sub≤0`
- **THEN** 服务端 SHALL 返回 `error` code **40301** 且 MUST NOT 进入 `question` 处理

#### Scenario: deviceNo 与 JWT 不一致

- **WHEN** `auth` 帧 `deviceNo` 与 JWT `device_no` claim 不一致
- **THEN** 服务端 SHALL 返回参数错误 `error` 帧且 MUST NOT 返回 `auth_ok`

#### Scenario: 与 voice ball 连接互不踢线

- **WHEN** 同一 `deviceNo` 已建立 `/voice/chat/ws` 且同一 `wxId` 另建 `/voice/clinic/ws`
- **THEN** 两条连接 SHALL 同时保持，且 clinic 连接 MUST NOT 触发 `VoiceWSManager` 替换逻辑

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，客户端上行 MUST 支持 `type=question` 帧（含非空 `text` 与非空 **`turnId`** UUID）与 **`type=cancel`** 帧（含非空 **`turnId`**）。服务端下行 MUST 支持 `auth_ok`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 六种 `type`，MUST NOT 下发 `session_sync`。流式下行帧（`thinking_delta`、`answer_delta`、`answer_done`）**MUST** 含与当前 turn 一致的 **`turnId`**。`turn_cancelled` MUST 含 **`turnId`** 与 **`reason`**，取值为 **`superseded`**、**`cancelled`** 或 **`disconnected`** 之一。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

#### Scenario: 流式 thinking 与 answer 携带 turnId

- **WHEN** 客户端发送 `question` 且 `turnId=uuid-A` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 推送的 `thinking_delta`、`answer_delta` 与最终 `answer_done` 均 MUST 含 `turnId=uuid-A`

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题或缺少 turnId 拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白，或缺少/空 `turnId`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: cancel 中断当前 turn

- **WHEN** 客户端在 turn `uuid-A` 流式进行中发送 `cancel` 且 `turnId=uuid-A`
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=cancelled`
- **AND** MUST NOT 下发 `answer_done` 且 MUST NOT consume `clinic_ai`

#### Scenario: 新 question supersede 上一 turn

- **WHEN** turn `uuid-A` 仍在流式进行中且客户端发送新 `question` 且 `turnId=uuid-B`
- **THEN** 服务端 MUST 取消 turn A 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=superseded`
- **AND** MUST 开始处理 turn B

#### Scenario: WS 断开取消 active turn

- **WHEN** 连接存在 active LLM turn 且 WebSocket 读循环因关闭或错误退出
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST NOT consume `clinic_ai`

### Requirement: Clinic LLM SHALL 使用 deepseek-v4-pro 并启用 thinking

Clinic LLM 调用 MUST 使用模型 `deepseek-v4-pro`，MUST 启用 thinking（`extra_body.thinking` 或等价配置，`reasoning_effort: high`）。LLM 请求超时 MUST 为 **120 秒**，配置源 MUST 为 `config.voice-service.yaml` 的 `aiClinic` 块，MUST NOT 依赖 `voice-chat.shared.yaml` 的 voice ball 超时。

#### Scenario: thinking 流映射

- **WHEN** DeepSeek 返回 reasoning 流
- **THEN** 服务端 MUST 将其映射为 `thinking_delta` 帧

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`；`wxId≤0` MUST 返回 `error` code **40301**。LLM 调用前 MUST 对 feature `clinic_ai` 以该 wxId 执行 check。若 `allowed=true`，MUST 经 `LoadProfile(LaneClinic)` 调用 LLM；**仅** turn 以 **`answer_done` 成功结束** 时 MUST 以同一 wxId consume。**若 `allowed=false`（`used >= limit`）**，MUST **NOT** 返回 code **40302**；MUST 经 **degraded** 路径调用 LLM，强制 `DefaultSeedProfile(LaneClinic)`（智谱 **`glm-4.1v-thinking-flash`**），且 **`answer_done` 成功时 MUST NOT consume** `clinic_ai`。**cancelled**、**superseded**、**disconnected** 或 LLM 失败而中断的 turn **MUST NOT** consume（含 degraded 路径）。

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽 degraded 问答

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径流式返回答案
- **AND** `answer_done` 后 MUST NOT consume clinic_ai

#### Scenario: 额度内成功完成扣减

- **WHEN** check 得到 used < limit 且 turn 完整流式结束并下发 `answer_done`
- **THEN** 系统 MUST consume `clinic_ai` 一次

#### Scenario: 用户 cancel 不扣额度

- **WHEN** turn 在流式过程中被 `cancel` 或 `superseded` 结束且未收到 `answer_done`
- **THEN** 系统 MUST NOT 对该 turn 调用 consume `clinic_ai`（含 degraded 路径）

### Requirement: Clinic SHALL 实施 Redis 限流（per wxId）

系统 MUST 对 clinic 提问路径实施 Redis 限流，键 **`voice:clinic:rate:{wxId}`**。限流计数 **MUST** 在 **`answer_done` 成功** 后递增；**cancelled**、**superseded**、**disconnected** 或失败 turn **MUST NOT** 递增限流计数。处理新 `question` 前 MUST 检查当前窗口计数；超限时 MUST 返回 WS `error` code **42901** 且 MUST NOT 调用 LLM。42901 与 **50301**（LLM lane 队列满）为不同语义：50301 MUST 在额度与 42901 检查通过后、调用上游前返回。

#### Scenario: 短时间频繁提问

- **WHEN** 同一 wxId 在限流窗口内已成功完成（`answer_done`）次数超过阈值
- **THEN** 下一次 `question` SHALL 返回 42901 且 MUST NOT 调用 LLM

#### Scenario: supersede 未完成 turn 不计入限流

- **WHEN** 用户在窗口内多次改问但均未产生 `answer_done`
- **THEN** 限流计数 MUST NOT 因 supersede 而额外递增

#### Scenario: 队列满返回 50301

- **WHEN** 42901 与 clinic_ai 检查通过但 clinic lane 闸门队列满
- **THEN** WS MUST 返回 50301 且 MUST NOT 调用 LLM

### Requirement: Clinic WS handler SHALL 非阻塞处理 question 并支持显式取消

`/voice/clinic/ws` 读循环 MUST NOT 同步阻塞在单条 `HandleQuestion` 直至 LLM 结束。每条合法 `question` MUST 在独立 goroutine 中处理，且连接 MUST 维护 at most one **active turn** 及其 `context.CancelFunc`。收到匹配 active turn 的 `cancel`、收到新 `question`（supersede）或连接关闭时 MUST 调用 cancel 中断 LLM/Python 流。用户显式 `cancel` 与 supersede MUST 分别对应下行 `turn_cancelled` 的 `reason=cancelled` 与 `reason=superseded`；连接关闭 MUST 仅取消上下文且 MUST NOT 依赖下行 `turn_cancelled`。

#### Scenario: 读循环可并发接收 cancel

- **WHEN** LLM 流式进行中且读循环收到 `cancel` 帧
- **THEN** handler MUST 在不等待 LLM 自然结束的情况下处理 cancel

#### Scenario: HandleQuestion 尊重 turn context

- **WHEN** turn context 已被 cancel
- **THEN** `HandleQuestion` MUST 停止 LLM 读流且 MUST NOT 写 `answer_done`

#### Scenario: cancel 在 thinking 阶段即时生效

- **WHEN** 服务端已开始下发 `thinking_delta` 且尚未进入 `answer_delta`，客户端发送匹配的 `cancel`
- **THEN** 服务端 MUST 取消 turn context 并下发 `turn_cancelled` 且 `reason=cancelled`

#### Scenario: cancel 在 answer 阶段即时生效

- **WHEN** 服务端已开始下发 `answer_delta` 且尚未 `answer_done`，客户端发送匹配的 `cancel`
- **THEN** 服务端 MUST 取消 turn context 并下发 `turn_cancelled` 且 `reason=cancelled`
- **AND** MUST NOT 随后下发该 turn 的 `answer_done`

### Requirement: Clinic LLM SHALL 经 ChatRequest 显式 opt-in thinking

胖宝诊疗 LLM 调用（`streamClinicLLM` / `streamClinicLLMHeld`）在构造 `aimodel.ChatRequest` 时 MUST 设置 `ThinkingEnabled=true`（或后续等价的显式 enabled 字段），以确保经 aimodel 层发送 `thinking: enabled` 而非继承默认 disabled。Clinic WS 下行 `thinking_delta` / `answer_delta` 协议 MUST 保持不变。

#### Scenario: clinic 流式仍返回 thinking

- **WHEN** 客户端发送合法 `question` 且 clinic lane 上游返回 reasoning 流
- **THEN** voice-service MUST 继续映射为 `thinking_delta` 帧且上游请求 MUST 为 thinking enabled

#### Scenario: clinic 不受 aimodel 默认 disabled 影响

- **WHEN** aimodel 全局默认 thinking 为 disabled
- **THEN** clinic 路径 MUST 仍为 enabled 且 MUST NOT 因零值误关 thinking

### Requirement: Clinic App quota read API SHALL expose clinic_ai degraded flag

`GET /voice/app/api/ai-quota` 响应中 `clinicAi` 对象 MUST 含 **`degraded`** 布尔字段；当 `clinic_ai` 的 `used >= limit` 时 MUST 为 `true`。`voiceAi` 对象 MAY 含同名字段（`used >= limit` 时为 true），但 voice_ai 用尽时 WS 仍 MUST 返回 40302，行为不变。

#### Scenario: clinic 额度用尽 API 标记

- **WHEN** wxId=1001 的 clinic_ai used=10、limit=10
- **THEN** `clinicAi.degraded` SHALL 为 true

#### Scenario: clinic 额度内

- **WHEN** wxId=1001 的 clinic_ai used=3、limit=10
- **THEN** `clinicAi.degraded` SHALL 为 false

### Requirement: Clinic LLM SHALL 经 clinic lane 与可配置 provider 调用

胖宝 `question` 处理中的 LLM MUST 经 `aimodel.InvokeStream(LaneClinic)` 调用；provider 与 model MUST 来自 Admin/DB profile，MUST NOT 硬编码 DeepSeek。thinking 流式下行语义（`thinking_delta` / `answer_delta`）MUST 保持与现有 WS 协议一致。

#### Scenario: 默认种子为智谱 thinking 模型

- **WHEN** 新部署且 DB 为种子 A 默认值
- **THEN** clinic lane MUST 使用 `provider=zhipu` 且 `model=glm-4.1v-thinking-flash`

#### Scenario: Admin 切回 DeepSeek

- **WHEN** Admin 将 clinic profile 改为 `provider=deepseek`、`model=deepseek-v4-pro`
- **THEN** 下一笔 `question` MUST 调用 DeepSeek 适配器且 MUST 使用 `deepseek-v4-pro` 闸门池

### Requirement: Clinic LLM stream SHALL emit answer_delta after thinking phase

当 Clinic 调用上游且 `ThinkingEnabled=true` 时，aimodel 流式层在已接收至少一个 reasoning/thinking 分片后，对上游单独到达的 `content` 分片 MUST 路由为 `answer`（触发 `OnAnswerDelta` 与 `answer_delta` WS 帧），MUST NOT 仅写入未订阅的 content 通道。

#### Scenario: reasoning 结束后 content 分片到达

- **WHEN** 上游先流式 `reasoning_content` 再流式纯 `content` 分片
- **THEN** voice-service MUST 向客户端发送 `answer_delta` 帧
- **AND** `answer_done` 的 `answer` 字段 MUST 非空（与上游正文一致）

#### Scenario: 非 thinking 闲聊路径不变

- **WHEN** `ThinkingEnabled=false` 且上游仅发送 `content`
- **THEN** 流式层 MUST 仍将 content 路由至 `OnContentDelta`（MUST NOT 误发 answer_delta）

### Requirement: Clinic Go path SHALL NOT cache or sync conversation history

voice-service Clinic WS MUST NOT 在 Redis 或其他 Go 进程内存储多轮对话 turns，MUST NOT 向客户端下发用于恢复 UI 历史的会话同步帧。对话 UI 历史由客户端本地负责；agent 多轮上下文由 Python 智能服务负责。Go Clinic 路径在成功 `auth_ok` 后 MUST 仅处理 `question`/`cancel` 与流式下行（`thinking_delta`/`answer_delta`/`answer_done`/`turn_cancelled`/`error`），并对 Python 调用仅透传提问所需字段（至少含 `question`、`device_no`、模型配置）。

#### Scenario: auth_ok 后无会话同步帧

- **WHEN** 客户端完成 `auth` 并收到 `auth_ok`
- **THEN** 服务端 MUST NOT 下发 `type=session_sync`
- **AND** 客户端 MAY 立即发送 `question`

#### Scenario: answer_done 不写 Go 会话缓存

- **WHEN** turn 以 `answer_done` 成功结束
- **THEN** voice-service MUST NOT 写入 `voice:clinic:session:{wxId}` 或等价 Go 侧对话缓存键

### Requirement: Clinic Go path SHALL NOT build feeding summary or baby profile for LLM

voice-service 在处理 Clinic `question` 时 MUST NOT 经 `DeviceHistory` 聚合近 7 天喂养摘要，MUST NOT 读写 `voice:clinic:summary:*`，MUST NOT 经 `DeviceProfile` 拉取宝宝画像并注入 Go 侧 LLM system context。喂养上下文与宝宝画像若需要，MUST 由 Python Clinic/companion 路径自行获取。

#### Scenario: question 不因摘要失败而阻断

- **WHEN** history 服务不可用
- **THEN** Clinic `question` 路径 MUST NOT 仅因「喂养摘要加载失败」返回 error（Go 不再依赖该摘要）

#### Scenario: question 不拉 DeviceProfile

- **WHEN** 客户端发送合法 `question`
- **THEN** voice-service Clinic 路径 MUST NOT 为拼装 LLM prompt 调用 DeviceProfile（本路径）

---

## pangbao-ai-clinic-flutter

<!-- source: openspec/specs/pangbao-ai-clinic-flutter/spec.md -->

# pangbao-ai-clinic-flutter Specification

## Purpose
TBD - created by archiving change add-pangbao-ai-clinic-room. Update Purpose after archive.
## Requirements
### Requirement: Flutter SHALL 提供胖宝 AI 诊室入口与页面

App（`flutter_ai_talk`）MUST 在首页 `home_immersive_header.dart` **新增** **胖宝** 入口（`Icons.pets`，tooltip「胖宝」），跳转至新页面 `pangbao_ai_screen.dart`。诊疗页 AppBar 标题 MUST 为 **「胖宝诊疗」**。首页品牌标题与 tooltip **MUST NOT** 改为「胖宝诊疗」（保持「胖宝」）。页面 MUST 支持文本输入提问并通过 WebSocket 展示流式 thinking 与 answer。首页 **MUST 保留** 原 **趋势** 入口（`Icons.insights`，tooltip「趋势」→ `/trends`），胖宝入口 **MUST NOT** 替换或隐藏趋势入口。

#### Scenario: 从首页进入胖宝诊疗

- **WHEN** 用户在首页点击胖宝入口（tooltip「胖宝」）
- **THEN** App SHALL 导航至胖宝诊疗页面且 AppBar 显示「胖宝诊疗」

#### Scenario: 从首页进入趋势

- **WHEN** 用户在首页点击趋势入口
- **THEN** App SHALL 导航至趋势图表页面（`/trends`）

### Requirement: Flutter SHALL 使用 ClinicWsClient 经 gateway-app 连接 WebSocket

App MUST 实现 `clinic_ws_client.dart`。连接 URL MUST 使用 `wsClinicUrl` / `wsClinicUrlEffective`，默认由 `apiBaseUrl`（gateway-app-server 主机）推导为 `wss://{host}/voice/clinic/ws`（对齐 `wsVoiceAsrUrlEffective` 模式）。客户端 **MUST NOT** 配置或连接 voice-service 内网地址。连接成功后 MUST 先发送首帧 `type=auth`（`accessToken` + `deviceNo`，与 history WS / UCG chat 一致），收到 `auth_ok` 后方可发送 `question`。每条 `question` MUST 含客户端生成的 UUID **`turnId`**。客户端 MUST 支持发送 **`type=cancel`**（含 `turnId`）。WS 生命周期：App 进入后台或离开诊疗页时 MUST **先** 对 active turn 发送 `cancel`（best-effort）**再** disconnect；回前台 MAY 重连且 MUST 重新 `auth`。客户端 MUST 解析 `auth_ok`、`session_sync`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 帧；流式帧 **MUST** 按 **`turnId`** 与当前 active turn 对齐，不匹配 **MUST** 丢弃。

#### Scenario: 连接 gateway-app 而非 voice-service

- **WHEN** App 建立胖宝诊疗 WebSocket
- **THEN** 连接目标主机 MUST 与 `apiBaseUrl` 一致（gateway-app-server）
- **AND** 路径 MUST 为 `/voice/clinic/ws`（或含 `apiBaseUrl` path 前缀的等价路径）

#### Scenario: 首帧 auth

- **WHEN** clinic WS 握手成功
- **THEN** 客户端 MUST 发送 `auth` 帧且 MUST NOT 在收到 `auth_ok` 前发送 `question`

#### Scenario: question 携带 turnId

- **WHEN** 用户发送新问题
- **THEN** 客户端 MUST 生成新 UUID 作为 `turnId` 并随 `question` 上行

#### Scenario: 未登录不发 question

- **WHEN** 用户未登录（无有效 accessToken）
- **THEN** 客户端 MUST NOT 建立可提问的 clinic WS 连接；若服务端返回 40301 MUST 引导登录

#### Scenario: 离开页面显式 cancel

- **WHEN** 用户离开胖宝诊疗页或 App 进入后台且存在 active 流式 turn
- **THEN** 客户端 MUST 发送 `cancel`（含 active `turnId`）后再断开 WebSocket

#### Scenario: 流式展示回答

- **WHEN** 收到与 active `turnId` 一致的 `answer_delta` 序列后以 `answer_done` 结束
- **THEN** UI SHALL 逐字/逐段更新回答区域

#### Scenario: session_sync 恢复历史

- **WHEN** 收到 `session_sync` 且 `turns` 非空
- **THEN** App SHALL 将已完成轮次填充至聊天 `_items`（user 问 + assistant 答 + 免责声明）
- **AND** MUST NOT 为历史轮次渲染 thinking（服务端未提供）

#### Scenario: 丢弃 stale turnId 帧

- **WHEN** 收到 `thinking_delta` 或 `answer_delta` 且 `turnId` 不等于当前 active turn
- **THEN** 客户端 MUST 丢弃该帧且 MUST NOT 更新 UI

#### Scenario: turn_cancelled 清理进行中 UI

- **WHEN** 收到 `turn_cancelled` 且 `turnId` 为当前或刚结束的 active turn
- **THEN** UI MUST 清除该 turn 的进行中 thinking/answer 流式状态

### Requirement: Flutter SHALL 实现 thinking 展示交互规范

胖宝诊疗页 thinking 区域 MUST：默认最多可见 **5 行**，不足 5 行时弹性高度；流式过程中 MUST 自动滚动至**最新** thinking 行（折叠视口 **底对齐** 最新内容，对齐 `home_voice_message_strip` 内层 `jumpTo(maxScrollExtent)` 模式）；超过 5 行 MUST 折叠，用户点击可展开全部或在折叠区内局部滚动；折叠态 MUST 使用内层 `ScrollController` 且 **MUST NOT** 使用 `NeverScrollableScrollPhysics` 阻止底对齐滚动；用户手动上滑固定（pin scroll）后 MUST 停止自动滚动直至用户回到底部或点击「跟随最新」。

#### Scenario: 流式自动滚动至最新行

- **WHEN** thinking 流式追加且用户未 pin scroll
- **THEN** 折叠视口 SHALL 展示最新 thinking 行（非顶部旧行）
- **AND** 内层 scroll offset SHALL 跳转至 `maxScrollExtent`

#### Scenario: 用户 pin 后停止跟随

- **WHEN** 用户上滑 thinking 区域内层 scroll 以查看较早内容
- **THEN** 后续 `thinking_delta` MUST NOT 强制跳回底部，直至用户点击「跟随最新」或滚回底部

#### Scenario: 跟随最新恢复

- **WHEN** 用户已 pin 且点击「跟随最新」
- **THEN** thinking 区域 SHALL 跳至最新内容并恢复流式 auto-scroll

### Requirement: Flutter SHALL 展示免责声明

每条 AI 回答（`answer_done` 后）UI MUST 展示文案：**「本回答仅供参考，不能替代医生诊断」**。

#### Scenario: 回答完成后展示免责

- **WHEN** 一次问答流式完成
- **THEN** 该条回答下方 SHALL 显示上述免责声明

### Requirement: Flutter SHALL 使用独立胖宝 consent

胖宝功能 MUST 使用独立 consent 键 `pangbao_ai_consent_v1`，与首页喂养 AI consent 分离。首次进入胖宝页且未 consent 时 MUST 展示同意流程；未同意 MUST NOT 发送 `question`。

#### Scenario: 首次进入需 consent

- **WHEN** 用户首次打开胖宝页且本地无 `pangbao_ai_consent_v1`
- **THEN** App SHALL 展示 consent 对话框且 MUST NOT 发送 `auth`/`question` 直至用户同意

### Requirement: Flutter SHALL 展示 clinic_ai 额度

`ai_quota_models` MUST 扩展 `clinicAi` 字段。诊疗页额度 hint MUST 展示 **「本月胖宝诊疗剩余 N 次」**（或等价 `AiQuotaRemainingHintFeature.clinicAi` 文案）。收到 WS `error` code **40302** 或 HTTP 40302 MUST 弹框 **「本月额度已用完」**。code **40301** MUST 引导登录。

#### Scenario: 胖宝诊疗额度展示

- **WHEN** 用户进入胖宝诊疗页且额度 API 返回 clinicAi remaining=5
- **THEN** UI SHALL 展示「本月胖宝诊疗剩余 5 次」

#### Scenario: 胖宝额度用尽

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

### Requirement: Flutter SHALL 允许流式过程中停止或改问

胖宝诊疗页在 LLM 流式（thinking 或 answer）进行中 **MUST NOT** 全局锁定文本输入。用户 **MUST** 能够：（a）发送新问题，以新 `turnId` supersede 当前 turn；和/或（b）通过停止控件发送 `cancel` 中断当前 turn。改问 **MUST** supersede 服务端上一 turn（由服务端下发 `turn_cancelled` reason=superseded）。

#### Scenario: 流式期间发送新问题

- **WHEN** thinking 或 answer 流式进行中且用户输入并发送新问题
- **THEN** 客户端 MUST 分配新 `turnId` 并发送 `question`
- **AND** UI MUST 展示新问题的 user 气泡并等待新 turn 的流式回复

#### Scenario: 流式期间点击停止

- **WHEN** 流式进行中且用户点击停止
- **THEN** 客户端 MUST 发送 `cancel` 且 `turnId` 为当前 active turn
- **AND** 收到 `turn_cancelled` 后 MUST 结束进行中 assistant 流式 UI

#### Scenario: 可选编辑用户气泡改问

- **WHEN** 实现 tap 用户气泡编辑（optional）
- **THEN** 预填问题文本后发送 MUST 使用新 `turnId` 并 supersede 当前 turn

---

## predict-imminent-push

<!-- source: openspec/specs/predict-imminent-push/spec.md -->

# predict-imminent-push Specification

## Purpose
TBD - created by archiving change predict-imminent-offline-push. Update Purpose after archive.
## Requirements
### Requirement: 系统 MUST 提供可按业务类型发送的公共可见推送

系统 MUST 提供跨域可调用的公共推送契约（HTTP internal 或等价），入参至少包含接收方 `wxId`、业务类型 `bizType`、可见文案（及可选深链/data）。当 `bizType` 为预测临近（实现常量名 MAY 为 `predict_imminent`）时，系统 MUST 向该用户已注册的推送 token（APNs/HMS/MiPush）发送可见通知，且 MUST NOT 使用 UCG 社区未读总和作为该通知的角标语义（MAY 省略 badge 或使用与社区无关的固定策略）。调用方（含 voice）MUST 经 `clients/*` 契约调用，MUST NOT 直接 import `internal/services/ucg` 业务包或直查 ucg 库表。

#### Scenario: 预测临近推送不套用社区角标

- **WHEN** voice 以预测临近 `bizType` 请求向某 `wxId` 发送可见推送
- **THEN** 下发载荷 MUST 为可见提醒，且 MUST NOT 把 UCG 会话/通知未读之和写入该推送的业务角标语义

### Requirement: 预测临近推送 MUST 扇出到宝宝全部绑定用户

当预测临近叫醒校验通过后，系统 MUST 解析该 `deviceNo` 下全部绑定 `wxId`，并对每一用户尝试公共推送。单个用户无有效 token 或发送失败 MUST NOT 阻止对其余用户的尝试；整体流程 MUST 仍按调度规格 Ack 消息。

#### Scenario: 两名看护均收到尝试

- **WHEN** 某 `deviceNo` 绑定两个 `wxId` 且均有有效推送 token，且叫醒校验通过
- **THEN** 系统 MUST 对两个 `wxId` 均发起可见推送调用

#### Scenario: 无 token 用户跳过

- **WHEN** 某绑定 `wxId` 没有可用推送 token
- **THEN** 系统 MUST 跳过该用户且 MUST 继续处理其他绑定用户

---

## predict-imminent-push-copy

<!-- source: openspec/specs/predict-imminent-push-copy/spec.md -->

# predict-imminent-push-copy Specification

## Purpose
TBD - created by archiving change predict-imminent-push-copy. Update Purpose after archive.
## Requirements
### Requirement: 预测临近可见推送正文 MUST 使用昵称与 title 模板

当 `voice-service` 在预测临近叫醒路径上准备向绑定用户发送 `bizType=predict_imminent` 的**可见**系统推送时，通知正文（传入推送契约的 `alert`）MUST 为：`{宝宝昵称}要{事件名}了`，其中：

- **事件名** MUST 为该叫醒载荷中的 `title`（trim 后的非空字符串），MUST NOT 由服务端展开事件树选取子事件名，MUST NOT 使用「事件将在约 N 分钟内发生」类时间回落文案作为事件名。
- **宝宝昵称** MUST 在发送前经 device 画像契约按 `deviceNo` 读取；当返回的 `babyName` 为空、或画像读取失败时，昵称 MUST 回落为字面量「宝宝」。
- 各厂商通知**标题**（如「胖宝」）本需求 MUST NOT 要求修改。

#### Scenario: 有昵称与 title

- **WHEN** 画像 `babyName` 为「小宝」且叫醒载荷 `title` 为「换尿布」，且其它推送闸均通过
- **THEN** 发往推送契约的 `alert` MUST 等于「小宝要换尿布了」

#### Scenario: 昵称为空回落

- **WHEN** 画像 `babyName` 为空串（或画像读取失败）且 `title` 为「吃奶」，且其它推送闸均通过
- **THEN** `alert` MUST 等于「宝宝要吃奶了」

#### Scenario: 不使用旧前缀

- **WHEN** 校验通过并发送可见预测临近推送
- **THEN** `alert` MUST NOT 以「宝宝提醒：」开头

### Requirement: 空 title 时 MUST NOT 发送可见推送

当叫醒载荷中的 `title` 经 trim 后为空时，系统 MUST NOT 调用推送发送（MUST NOT 向任何绑定 wx 下发该次可见 `predict_imminent` 通知），MUST 记录可观测日志，且消费路径 MUST 仍按现有约定 Ack（MUST NOT 因空 title 而 Nack requeue 或再次 Publish 延时重试）。

#### Scenario: 缺 title 跳过

- **WHEN** Redis 匹配与其它闸均可通过，但载荷 `title` 为空或仅空白
- **THEN** 系统 MUST NOT 发送该次可见推送，且 MUST Ack 该叫醒消息

#### Scenario: 有 title 才发送

- **WHEN** 载荷 `title` 为非空「睡觉」且其它闸通过
- **THEN** 系统 MUST 按模板拼装正文并尝试向绑定用户发送可见推送

---

## predict-imminent-schedule

<!-- source: openspec/specs/predict-imminent-schedule/spec.md -->

# predict-imminent-schedule Specification

## Purpose
TBD - created by archiving change predict-imminent-offline-push. Update Purpose after archive.
## Requirements
### Requirement: 客户端 MUST 能全量替换宝宝待发生预测节点

系统 MUST 提供经 gateway-app 暴露的 App HTTP 接口，供已登录且绑定设备的客户端在预测更新后提交该宝宝（`deviceNo`）的待发生事件全量列表。每条事件 MUST 包含 `eventId` 与 `nextAt`（unix 秒）。服务端 MUST 以该列表**整表替换**该宝宝在 Redis 中的待办（MUST NOT 写入业务 MySQL 作为权威）。待办键 MUST 经 `cachekit` 登记且 MUST NOT 依赖 TTL 过期清理（靠后续替换维护）。后一次成功写入 MUST 覆盖前一次（最后写入赢）。空列表 MUST 清空该宝宝待办。

#### Scenario: 预测刷新后覆盖旧列表

- **WHEN** 客户端先后两次成功提交同一 `deviceNo` 的不同 `events` 列表
- **THEN** Redis 中该宝宝待办 MUST 仅等于第二次提交内容，且 MUST NOT 残留仅存在于第一次列表中的事件节点

#### Scenario: 空列表清空

- **WHEN** 客户端提交空 `events` 且请求成功
- **THEN** 该宝宝 Redis 待办 MUST 为空，后续基于旧节点的叫醒校验 MUST 失败并跳过推送

### Requirement: 同步成功后 MUST 按提前量投递延时叫醒消息

对提交列表中每一条有效事件，系统 MUST 计算 `fireAt = nextAt - 300`（秒），并以 `x-delay = max(0, fireAt - now) * 1000` 毫秒发布到已启用 `rabbitmq_delayed_message_exchange` 的延时交换机，routing 到 voice 预测临近队列。消息体 MUST 足以在消费时定位 `deviceNo`、`eventId`、`nextAt`。当 `fireAt <= now` 时 MUST 仍投递（delay 为 0 或等价即时投递），不得因「已过理想叫醒点」而拒绝入队。

#### Scenario: 五分钟前提醒入队

- **WHEN** 当前时间为 T0，某事件 `nextAt = T0 + 600`
- **THEN** 系统 MUST 发布延时消息且延迟约 300 秒量级（允许实现取整误差），消息携带该 `deviceNo`、`eventId`、`nextAt`

#### Scenario: 已进入窗口仍入队

- **WHEN** 某事件 `nextAt - 300 <= now`
- **THEN** 系统 MUST 仍发布叫醒消息（零延迟或即时），MUST NOT 静默丢弃该事件的叫醒

### Requirement: 消费叫醒时 MUST 以 Redis 校验为准且业务失败 MUST Ack

`voice-service` MUST 以 AMQP push consumer（`autoAck=false`）消费预测临近队列。处理单条 delivery 时 MUST 先读取该宝宝当前 Redis 待办：若不存在匹配的 `eventId` 与相同 `nextAt`，MUST 跳过推送并对该 delivery **Ack**。下列情况 MUST **Ack** 且 MUST NOT `Nack(requeue=true)`，MUST NOT 在消费路径再次 Publish 延时重试消息：载荷非法、Redis 不匹配、推送去重命中、历史闸命中、列绑定用户失败、系统推送发送失败。进程崩溃导致未 Ack 时 broker 可重投；应用层 MUST 依赖去重键限制重复可见推送。

#### Scenario: 列表已替换则旧消息作废

- **WHEN** 延时消息携带的 `nextAt` 与 Redis 当前该 `eventId` 的 `nextAt` 不一致或不存在该事件
- **THEN** consumer MUST NOT 发送系统推送，且 MUST Ack 该消息

#### Scenario: 推送失败不回队

- **WHEN** 校验通过但向厂商推送返回错误
- **THEN** consumer MUST 记录可观测日志、MUST Ack，且 MUST NOT Nack requeue，MUST NOT 再发布一条延时消息作为重试

### Requirement: 推送前 MUST 做五分钟去重与三十分钟历史闸

对同一 `(deviceNo, eventId)`，成功进入推送发送路径前 MUST 检查推送去重状态：若五分钟内已推送过，MUST 跳过并 Ack。若未去重命中，MUST 查询 history：该宝宝在过去 1800 秒内是否存在同类事件记录（`start_time` 落在窗口内）；若存在 MUST 跳过推送并 Ack。当上报 `eventId` 为事件树一级根时，历史查询 MUST 包含该根及其后代叶子 `eventId` 集合（经 device 事件目录解析），MUST NOT 仅用根 id 查询导致漏判。通过闸后 MUST 写入五分钟去重标记（成功发送或「已尝试发送」的语义由实现固定为「进入发送前占位」或「发送成功后写入」，但 MUST 保证五分钟内不会对用户重复可见推送风暴）。

#### Scenario: 用户已记录喂养则不推

- **WHEN** Redis 仍匹配某喂养类预测节点，但 history 在近 30 分钟内已有对应叶子事件
- **THEN** 系统 MUST NOT 发送该节点的临近推送

#### Scenario: 五分钟内不重复推同一事件

- **WHEN** 同一 `(deviceNo, eventId)` 在五分钟内第二次通过 Redis 匹配进入推送逻辑
- **THEN** 系统 MUST 跳过第二次可见推送

### Requirement: 叫醒消费宿主与拓扑 MUST 声明

本能力的 AMQP consumer MUST 运行于 `voice-service`。系统 MUST 使用独立于即时 `voice.events` 业务的延时交换机（启用 delayed message 插件），并声明专用队列与绑定。变更与部署文档 MUST 说明插件启用方式；插件不可用时同步路径 MUST 失败可观测，MUST NOT 静默假装已调度。

#### Scenario: voice 进程订阅专用队列

- **WHEN** `voice-service` 启动且 MQ 配置可用
- **THEN** 系统 MUST 订阅预测临近专用队列以消费延时到期消息

---

## prediction-allowed-count

<!-- source: openspec/specs/prediction-allowed-count/spec.md -->

# prediction-allowed-count Specification

## Purpose
TBD - created by archiving change commercial-feature-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 预测开通数量 MUST 经功能列表暴露且不得下发事件 ID

对预测类功能，合成 catalog 对应项 MUST 暴露非负整数 `allowedCount`（无记录则为 0）。响应 MUST NOT 包含由服务端挑选的预测事件 ID 列表。客户端负责按本地预测列表顺序解锁前 N 项。系统 MUST NOT 将独立 `GET .../prediction/allowed-count` 作为客户端必调接口（可省略）。

#### Scenario: 列表项含 allowedCount 而非事件 ID

- **WHEN** 客户端请求合成功能目录且存在预测类功能
- **THEN** 该项 MUST 含 `allowedCount`，且 MUST NOT 含事件 ID 数组作为开通清单

#### Scenario: 无开通记录时为 0

- **WHEN** 设备从未获得预测开通数量增量
- **THEN** 该项 `allowedCount` MUST 为 `0`

### Requirement: Admin 对预测功能 MUST 只配置开通数量语义

与预测解锁相关的可配置授予物 MUST 表达为开通数量（增量），MUST NOT 要求运营配置具体事件 ID。

#### Scenario: 授予数量后列表可读到

- **WHEN** 经支付/邀请码/广告等路径为某 `device_no` 增加预测开通数量 N
- **THEN** 随后合成目录中该项 `allowedCount` MUST 反映累加后的权威值（无扣减前提下 ≥ 原值 + N）

### Requirement: allowedCount 权威在 MySQL 且热读经 Redis

`allowedCount` 权威 MUST 存于 MySQL（`feature_allowed_count`）。读路径 MUST 可经 Redis 缓存加速（产品已确认）；履约或 Admin 变更后 MUST 失效或写穿。Redis 访问 MUST 经 `cachekit` 与 platform 键 builder。

#### Scenario: 履约后数量可见

- **WHEN** 功能 SKU 支付成功且 `grant_kind` 为预测数量增量
- **THEN** 该 `device_no` 的 `allowedCount` MUST 增加对应 `grant_quantity`，且后续 catalog 读 MUST 看到更新后的值（缓存失效后）

---

## prediction-count-grant

<!-- source: openspec/specs/prediction-count-grant/spec.md -->

# prediction-count-grant Specification

## Purpose
TBD - created by archiving change invite-peer-force-ucg. Update Purpose after archive.
## Requirements
### Requirement: 预测开通三通道每次永久 +1

对 `prediction_unlock`，系统在支付履约、邀请码兑换成功、广告完成开通时 MUST 各将设备永久 `allowed_count` 增量 +1（或 SKU/`grantQuantity` 为 1 的等价增量）。预测邀请兑换 MUST NOT 再写入临时/永久全开（`full_access` / catalog `allowedCount=-1`）。预测支付商品的 `grantQuantity` MUST 为 1。

#### Scenario: 支付一次

- **WHEN** 用户支付预测开通商品成功
- **THEN** 该设备永久预测条数 +1，且 MUST NOT 仅因本次支付变为全开哨兵

#### Scenario: 广告一次

- **WHEN** 用户完成预测广告开通
- **THEN** 该设备永久预测条数 +1

#### Scenario: 邀请一次

- **WHEN** 用户用好友邀请码兑换 `prediction_unlock`
- **THEN** 该设备永久预测条数 +1

### Requirement: 非预测邀请开通为权益一次

当非 `prediction_unlock` 的功能启用 `invite_code` 且用户兑换成功时，系统 MUST 授予 entitlement 一次（按功能定义），MUST NOT 默认按条数累加。

#### Scenario: 其它功能兑码

- **WHEN** 用户对支持邀请码的非预测功能兑码成功
- **THEN** 系统写入对应 entitlement，且该人×码×功能不可再兑

---

## prediction-effective-count

<!-- source: openspec/specs/prediction-effective-count/spec.md -->

# prediction-effective-count Specification

## Purpose
TBD - created by archiving change prediction-default-invite-admin-ux. Update Purpose after archive.
## Requirements
### Requirement: 预测有效条数 MUST 为默认加永久累加

对 `prediction_unlock`（及约定的预测数量类功能），合成 catalog 暴露的有效可看条数（非临时全开时）MUST 等于 `defaultFree + permanentDelta`，其中 `defaultFree` 来自功能定义可配置默认开通数（非负），`permanentDelta` 来自设备永久累加权威（付费/广告等，非邀请码临时全开）。无设备永久行时 `permanentDelta` MUST 为 0。系统 MUST NOT 在无临时全开时仅因缺行而固定返回 0 并忽略已配置的 `defaultFree`。

#### Scenario: 仅配置默认时返回默认条数

- **WHEN** 某设备无永久累加记录，且该功能 `defaultFree=3`，且无有效临时全开
- **THEN** catalog 该项 `allowedCount` MUST 为 `3`

#### Scenario: 默认加永久累加

- **WHEN** `defaultFree=3` 且该设备 `permanentDelta=5`，且无有效临时全开
- **THEN** catalog 该项 `allowedCount` MUST 为 `8`

### Requirement: 邀请码预测临时全开 MUST 使用哨兵 -1

当设备对预测功能存在有效临时全开（邀请码授予且未过期，或永久全开标志生效）时，catalog 该项 `allowedCount` MUST 为 `-1`（哨兵，表示全部可看）。有限期临时全开 MUST 暴露到期时间（`expiresAt` 为 unix 秒）。到期后 MUST 回落为「默认 + 永久累加」，且 MUST NOT 清除既有永久累加。邀请码兑换预测 MUST NOT 再将 `grant_quantity` 累加进永久 `allowed_count`。

#### Scenario: 期限内返回 -1

- **WHEN** 用户用邀请码成功开通 `prediction_unlock` 且授予有效天数 > 0，且当前时间未超过到期
- **THEN** catalog 该项 `allowedCount` MUST 为 `-1`，且 MUST 含对应 `expiresAt`

#### Scenario: 到期后回落

- **WHEN** 临时全开已过期，且 `defaultFree=2`、`permanentDelta=4`
- **THEN** catalog 该项 `allowedCount` MUST 为 `6`，且 MUST NOT 为 `-1`

#### Scenario: 邀请码不增加永久条数

- **WHEN** 邀请码兑换预测成功
- **THEN** 该设备永久累加权威值 MUST 与兑换前相同（仅更新临时全开状态）

### Requirement: 默认开通数 MUST 可经 Admin 配置

功能定义 MUST 支持配置预测类默认开通数；Admin 更新后，后续无临时全开的 catalog 读 MUST 反映新默认（缓存失效后）。功能编号仍为与客户端约定的稳定 ID，本要求不授权 Admin 修改 `featureId`。

#### Scenario: Admin 修改默认开通数

- **WHEN** 运维将 `prediction_unlock` 的默认开通数从 0 改为 2 并保存
- **THEN** 无永久累加且无临时全开的设备再次拉 catalog 时该项 `allowedCount` MUST 为 `2`

---

## privacy-policy-company

<!-- source: openspec/specs/privacy-policy-company/spec.md -->

# privacy-policy-company Specification

## Purpose
TBD - created by archiving change company-branding-site-legal. Update Purpose after archive.
## Requirements
### Requirement: 隐私政策 MUST 披露运营主体与联系方式

`resource/public/privacy-policy.html` MUST 在显著位置（开篇或独立「运营者」节）说明：本应用「胖宝」由**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）运营；注册地址表述 MUST 为「以市场主体公示为准」或等价；联系邮箱 MUST 含 `tou_zy@foxmail.com`。文档 MUST 更新生效或修订日期。路径 MUST 仍为 `/privacy-policy.html`。

#### Scenario: 用户阅读主体信息

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 可见公司中文全称与客服邮箱，且 SHALL NOT 将运营者表述为未具名的「个人开发者」或「胖宝团队」作为唯一主体

### Requirement: 隐私政策 MUST 按现网能力披露收集与处理范围

隐私政策 MUST 在既有登录、宝宝档案、喂养记录、社区、位置（可选）、AI 与内容审核披露基础上，补充或明确至少下列现网相关项（表述可与实现细节对齐，但 MUST NOT 隐瞒已上线能力）：

- **推送通知**：为向用户发送离线提醒等，可能收集并使用设备推送标识（如 APNs / 华为 / 小米等厂商 token）；用户可在系统设置中关闭通知。
- **预测临近提醒**：客户端可将待发生喂养相关时间点同步至服务端，用于在约定提前量发送提醒。
- **商业化功能**：用户开通 VIP 或付费/邀请等功能时，系统处理与订单、权益状态相关的必要信息。
- **AI 相关**：喂养 AI、胖宝诊疗、社区润笔及同类智能功能可能将用户输入或为生成回答所需的上下文发送至可配置的第三方大模型（包括但不限于智谱 GLM、DeepSeek、阿里云 DashScope，以实际配置为准）；MUST NOT 声称仅使用单一固定供应商。若 App 展示 AI 思考过程，MUST 说明该展示性质。
- **监护人**：宝宝昵称、性别、出生日期等由监护人提供并用于月龄与个性化功能。

既有 Apple 登录「仅匿名用户标识符、不存储 Apple 邮箱/姓名/凭证原文」要求 MUST 保持。

#### Scenario: 用户查阅推送与预测说明

- **WHEN** 用户阅读「我们收集哪些信息」或等价章节
- **THEN** 文案 SHALL 提及推送标识用于提醒，以及预测相关时间点可能同步服务端

#### Scenario: 用户查阅 AI 与主体

- **WHEN** 用户阅读 AI / 第三方章节与运营者信息
- **THEN** 文案 SHALL 含多供应商可配置表述，并 SHALL 含公司主体全称

---

## prod-ecs-cutover

<!-- source: openspec/specs/prod-ecs-cutover/spec.md -->

# prod-ecs-cutover Specification

## Purpose
TBD - created by archiving change migrate-prod-to-4c8g-ecs. Update Purpose after archive.
## Requirements
### Requirement: 生产与测试主机职责分离

切流完成后，系统部署 MUST 满足：生产 Docker 栈与生产 MySQL 仅运行在新 4C8G ECS；测试 Docker 栈与 `ai_voice_*_test` 库仅留在旧 ECS。旧 ECS MUST NOT 再运行生产微服务 compose（可短期保留生产库只读回滚）。

#### Scenario: 切流后拓扑验收

- **WHEN** 运维完成 DNS/反代切流并执行验收
- **THEN** 生产域名请求解析到的服务 MUST 连接新机 MySQL 生产库（库名无 `_test` 后缀），且测试域名请求 MUST 仍连接旧机 `_test` 库

#### Scenario: 旧机停止生产栈

- **WHEN** 生产流量已切至新机
- **THEN** 旧机生产微服务 compose MUST 处于 down，以避免容器继续写入旧机 `ai_voice_*`

### Requirement: 生产库迁移范围与改址方式

生产切流 MUST 仅迁移 MySQL 生产库 `ai_voice_*`（含 history/device/voice/app/ucg/sim 等域约定库）。测试库 `ai_voice_*_test` MUST NOT 迁往新机。应用进程改连新库 MUST 通过更新 `.env.prod` 的 `MYSQL_TCP_HOST` 完成；`*_DB_LINK` MUST 继续使用 `mysql-host` 占位符与生产库名。

#### Scenario: 仅改 MYSQL_TCP_HOST

- **WHEN** 新机生产服务启动且 `MYSQL_TCP_HOST` 已设为新机可达地址
- **THEN** 启动日志 MUST 显示各 `database.*` 主机为该地址，且库名为对应 `ai_voice_*`（非 `_test`）

#### Scenario: 测试 env 不跟迁

- **WHEN** 生产已迁至新机
- **THEN** 旧机 `.env.test` 的 `MYSQL_TCP_HOST` MUST 仍指向旧机 MySQL，且库名 MUST 保持 `_test` 后缀

### Requirement: 新机 MySQL buffer 首日规格

新机生产 MySQL 在首日上线时 MUST 将 `innodb_buffer_pool_size` 配置为 **1G**。旧机（测试）MySQL MUST NOT 跟随调整到 1G，MUST 维持适合约 2G 主机的较小 buffer（例如约 256M 量级）。

#### Scenario: 新机 buffer 验收

- **WHEN** 新机 MySQL 完成配置并重启后执行 `SHOW VARIABLES LIKE 'innodb_buffer_pool_size'`
- **THEN** 返回值 MUST 为 1G（1073741824 字节）或运维文档等效写法对应的 1G

#### Scenario: 旧机 buffer 不跟调

- **WHEN** 生产切流完成
- **THEN** 旧机 MySQL `innodb_buffer_pool_size` MUST 仍为切流前测试适用规格，不得因新机升配而被改为 1G

### Requirement: 生产 Redis 空集群冷启

新机生产 Redis Cluster MUST 以空 volume 冷启：首次 `up` 后执行一次三主 `cluster create`，达到 `cluster_state:ok` 后再启动依赖 Redis 的生产微服务。MUST NOT 将旧机生产 Redis AOF/volume 拷贝为新机初始数据。切流 runbook MUST 声明冷启的用户可感知后果：App refresh token 失效需重新登录；语音/诊所会话上下文丢失；AI 月额度 Redis 计数归零；Feed 索引/快照需经读路径或写路径重建；UCG 私信正文以 MySQL 为权威不因冷启丢失。

#### Scenario: 新集群首次创建

- **WHEN** 新机首次初始化生产 Redis Cluster
- **THEN** 运维 MUST 在空节点上执行一次 `cluster create`，且 `CLUSTER INFO` MUST 显示 `cluster_state:ok`

#### Scenario: 冷启后私信仍可读

- **WHEN** Redis 为空且用户拉取已持久化到 MySQL 的会话历史
- **THEN** 系统 MUST 仍能返回私信正文（MySQL fallback / warm）

#### Scenario: 冷启后需重新登录

- **WHEN** 用户持有切流前签发的 refresh token 访问新机生产环境
- **THEN** 该 refresh token MUST 不再有效，用户 MUST 通过登录流程获取新凭证

### Requirement: 换机 runbook 与档位文档

仓库 MUST 提供可执行的生产换机切流说明（含准备、dump/restore、Redis 冷启、停写、DNS 切流、验收、回滚、旧库保留后再清理），并 MUST 在内存档位文档中记录 **4C8G / 单 prod + 本机 MySQL** 的建议，其中 MySQL buffer 起点为 1G。

#### Scenario: runbook 可独立执行

- **WHEN** 运维按仓库 runbook 执行生产换机
- **THEN** 文档 MUST 覆盖新机 MySQL buffer=1G、仅迁生产库、`MYSQL_TCP_HOST` 改址、Redis 空集群冷启、旧机 test 留守与验收命令

#### Scenario: 档位文档含 4C8G

- **WHEN** 查阅 `memory-sizing-guide`（或等价档位文档）
- **THEN** 文档 MUST 包含 4C8G prod-only 建议，并写明新机 `innodb_buffer_pool_size` 首日 1G

---

## push-app-and-internal-api

<!-- source: openspec/specs/push-app-and-internal-api/spec.md -->

# push-app-and-internal-api Specification

## Purpose
TBD - created by archiving change extract-push-service. Update Purpose after archive.
## Requirements
### Requirement: App MUST 能注册与注销推送 token

系统 MUST 经 gateway-app 对外提供中性路径注册/注销接口（如 `POST /app/api/push/register` 与 `POST /app/api/push/unregister`），请求体含 `channel`、`token`、`deviceKey`（注销可按既有字段）。接口 MUST 使用登录用户 `wxId`（网关注入内部头），MUST NOT 要求微信 unionid 绑定作为前提。系统 MUST NOT 再提供 `/ucg/app/api/push/register` 与 `/ucg/app/api/push/unregister` 作为有效推送注册入口。

#### Scenario: 登录用户注册成功

- **WHEN** 已登录用户以合法 channel/token/deviceKey 调用注册接口
- **THEN** push-service MUST 持久化该 token，且后续按该 wxId 发送可见推送时可投递到对应通道

#### Scenario: 旧 UCG 注册路径移除

- **WHEN** 客户端请求已删除的 `/ucg/app/api/push/register`
- **THEN** 系统 MUST NOT 再将其作为推送注册成功路径（404 或网关未反代至 push 实现）

### Requirement: Internal MUST 支持按 bizType 发送且 badge 由调用方提供

push-service MUST 提供经内部密钥鉴权的发送接口，入参至少含 `wxId`、`bizType`、可见文案（或静默标志）与可选 `badge`、`data`。当请求携带 `badge` 时，下发载荷 MUST 使用该值；未携带时 MUST 默认 `0`。push-service MUST NOT 为填充 badge 去查询 UCG 未读。无有效 token 时 MUST 跳过发送且 MUST NOT 视为调用方硬失败（可返回成功空投递或等价 best-effort 语义，并在实现中固定）。

#### Scenario: UCG 传入未读角标

- **WHEN** 调用方传入 `badge=5` 与非空 alert
- **THEN** 厂商载荷中的角标 MUST 为 5

#### Scenario: 预测类默认零角标

- **WHEN** 调用方不传 badge 或显式传 0
- **THEN** 厂商载荷角标 MUST 为 0

### Requirement: push 注册注销 MUST 不计入 App usage 统计

经 gateway-app 暴露的 push 注册与注销 App 接口 MUST 列入 usage 维护型排除（`maintenance_skip` 精确 METHOD+path）。负责人结论：不统计。旧 ucg push 排除项 MUST 在路由删除后移除或替换为新 path，避免残留无效键。

#### Scenario: 注册不写入使用统计

- **WHEN** 客户端成功或失败调用 `POST /app/api/push/register`
- **THEN** 该次调用 MUST NOT 计入 App API 使用统计报表

---

## push-caller-migration

<!-- source: openspec/specs/push-caller-migration/spec.md -->

# push-caller-migration Specification

## Purpose
TBD - created by archiving change extract-push-service. Update Purpose after archive.
## Requirements
### Requirement: 跨域调用 MUST 经 clients/push

voice、ucg 及其他域服务发送系统推送时 MUST 通过 `internal/clients/push`（或同等 clients 包）出站 HTTP 调用 push-service，MUST NOT import `internal/services/push` 业务实现包，MUST NOT 再 import `internal/services/ucg` 仅为发送推送。UCG 社区可见/静默角标推送 MUST 在调用前计算未读（若需要）并将 `badge` 放入请求。

#### Scenario: voice 预测不依赖 ucg 推送包

- **WHEN** voice 预测临近校验通过并发送提醒
- **THEN** 出站目标 MUST 为 push-service，MUST NOT 调用已删除的 ucg internal push 路径作为唯一实现

### Requirement: UCG 进程 MUST 移除厂商推送实现

`ucg-service` MUST 删除（或停止编译）原 `push_*.go` 厂商发送与 `ucg_push_device` 写路径权威，MUST 删除 App/Internal 推送注册与 by-biz-type 路由。社区私信/评论等原推送触发点 MUST 改为经 `clients/push` 发送。

#### Scenario: ucg 不再直连 APNs 配置作为推送宿主

- **WHEN** 审查 ucg-service 运行时配置
- **THEN** 推送厂商证书/密钥 MUST NOT 再作为 ucg-service 发送推送的必需依赖（可整段移除 `UCG_APNS_*` 推送注入）

---

## push-device-admin

<!-- source: openspec/specs/push-device-admin/spec.md -->

# push-device-admin Specification

## Purpose
TBD - created by archiving change push-token-unique-takeover. Update Purpose after archive.
## Requirements
### Requirement: 运维 MUST 能分页查看并删除 push_device

系统 MUST 在 push-service 提供经 Admin 鉴权的管理 API，用于分页列出 `push_device` 行（至少含 `id`、`wxId`、`channel`、`deviceKey`、`token`、`updatedAt`），并支持按行 `id` 删除。列表 MUST 支持合理分页；MAY 支持按 `wxId`、`channel` 或 token 前缀筛选。删除成功后该行 MUST 不再参与后续按 wx 扇出推送。API MUST 经 gateway-app 反代暴露（建议前缀 `/push/admin/api/`），MUST 使用与现网一致的 Admin JWT + 下游 `X-Admin-Password` 注入模式；浏览器 MUST NOT 直接携带运维口令头。

#### Scenario: 列表可见

- **WHEN** 已登录运维人员请求设备列表接口且鉴权通过
- **THEN** 响应 MUST 返回当前 `push_device` 分页数据（字段至少含 id 与 wxId、channel、token、updatedAt）

#### Scenario: 手动删除

- **WHEN** 运维人员对某 `id` 调用删除且鉴权通过
- **THEN** 该行 MUST 从 `push_device` 移除，后续对该原 `wxId` 的推送若无其它 token 则 MUST 跳过该设备投递

### Requirement: 运维 Hub MUST 提供推送设备管理入口与静态页

系统 MUST 在 gateway-app 注册推送设备管理静态页（建议 `/device/admin/push-device-admin.html`），并在运维 Hub 模块列表（`admin-modules`）中提供可点击入口。页面 MUST 使用 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch` 访问上述 Admin API，支持查看列表与确认后删除。页面展示 token 时 MUST 截断或脱敏，MUST NOT 以明文完整 token 为默认展示（MAY 提供展开，但默认截断）。

#### Scenario: Hub 可进入

- **WHEN** 运维人员打开运维 Hub 并已登录
- **THEN** MUST 能通过登记入口打开推送设备管理页并加载列表（在 API 可用前提下）

---

## push-dispatch-observability

<!-- source: openspec/specs/push-dispatch-observability/spec.md -->

# push-dispatch-observability Specification

## Purpose
TBD - created by archiving change fix-logger-level-env-push-warn-logs. Update Purpose after archive.
## Requirements
### Requirement: Internal by-biz-type 受理可观测

`POST /push/internal/api/by-biz-type` 处理路径 MUST 在以下情况输出 **Warning** 级日志（前缀建议 `[push]`），且 MUST NOT 打印完整 device token 或完整 alert 正文：

- 内部密钥校验失败
- 请求体/必填字段校验失败
- 校验通过并受理发送（含 `wxId`、`bizType`、`badge`、`silent`、`alertLen`）

#### Scenario: 受理成功有日志

- **WHEN** 合法内部请求到达且校验通过
- **THEN** push-service 日志中 MUST 出现一条 Warning，表明已 `accepted` 该 `wxId` 与 `bizType`

#### Scenario: 鉴权失败有日志

- **WHEN** 内部密钥无效或缺失
- **THEN** 在返回 403 的同时 MUST 打 Warning（`auth_denied` 或等价 reason）

### Requirement: Dispatch 与厂商发送可观测

异步 dispatch 路径 MUST：

- 在查询到待发设备列表后打 Warning，含 `wxId`、`bizType`、`deviceCount`（`deviceCount=0` 时沿用或等价于现有 `no_device` skip）
- 某 channel 因凭证未配置而跳过时打 **Warning**（不得仅 Debug）
- 某 token 发送成功时打 Warning（`send_ok` 或等价）
- 异步 goroutine panic 时打 Error/Warning，不得空 recover 吞掉

#### Scenario: 无注册设备

- **WHEN** 受理后该 `wxId` 在 `push_device` 无可用 token
- **THEN** 日志 MUST 以 Warning 标明 skip（如 `no_device`），且在 `GF_LOGGER_LEVEL=prod` 下仍可见

#### Scenario: 凭证未配置

- **WHEN** 存在设备记录但对应 channel 厂商凭证未配置
- **THEN** 日志 MUST 以 Warning 标明该 channel 因凭证未配置而跳过（`GF_LOGGER_LEVEL=prod` 下仍可见）

#### Scenario: 发送成功

- **WHEN** 某 channel sender 返回成功（无 error）
- **THEN** 日志 MUST 以 Warning 标明 `send_ok`（或等价），含 `channel`、`wxId`、`bizType`

---

## push-service-runtime

<!-- source: openspec/specs/push-service-runtime/spec.md -->

# push-service-runtime Specification

## Purpose
TBD - created by archiving change extract-push-service. Update Purpose after archive.
## Requirements
### Requirement: 系统 MUST 提供独立 push-service 进程

系统 MUST 以独立进程 `push-service` 承载厂商推送（APNs/HMS/MiPush）与设备 token 存储。该进程 MUST NOT 复用 `notify-service`。进程 MUST 使用独立配置文件（如 `config.push-service.yaml`）与可配置监听地址（默认建议 `:9808`）。数据库 MUST 使用同 MySQL 实例上的独立 schema `ai_voice_push`，token 表 MUST NOT 再作为 UCG 业务表写入权威。

#### Scenario: 进程可独立启动

- **WHEN** 运维仅启动 push-service 且数据库 `ai_voice_push` 可用
- **THEN** 服务 MUST 能接受注册与 internal 发送请求，且 MUST NOT 依赖 ucg-service 进程内存活才能连接厂商

### Requirement: Token 存储 MUST 在 ai_voice_push

push-service MUST 将设备推送 token 持久化于 `ai_voice_push` 中的设备表（实现表名可定为 `push_device`），并以 `(wx_id, device_key, channel)` 唯一约束支撑 upsert。channel MUST 限于 `apns`、`hms`、`mipush`。

#### Scenario: 同设备同通道覆盖 token

- **WHEN** 同一 `wxId`+`deviceKey`+`channel` 再次注册新 token
- **THEN** 系统 MUST 更新为新 token，MUST NOT 插入重复唯一键行

### Requirement: 部署清单 MUST 包含 push-service

CI ACR 工作流与 `docker-compose.microservices.yml`（及必要 overlay/env 示例）MUST 纳入 `push-service` 镜像构建/推送与本地编排；推送厂商相关环境变量 MUST 注入 push-service 容器，MUST NOT 再作为 ucg-service 推送实现的必需配置。

#### Scenario: ACR 可构建 push-service

- **WHEN** 工作流以全量或指定 `push-service` 触发构建
- **THEN** MUST 使用对应 Dockerfile 产出 push-service 镜像标签

---

## push-token-uniqueness

<!-- source: openspec/specs/push-token-uniqueness/spec.md -->

# push-token-uniqueness Specification

## Purpose
TBD - created by archiving change push-token-unique-takeover. Update Purpose after archive.
## Requirements
### Requirement: 推送 token MUST 全局唯一且注册后来顶上

系统在持久化推送设备（`push_device`）时，厂商 `token` 字符串 MUST 在全表唯一（MUST NOT 仅按 `wx_id`+`device_key`+`channel` 唯一而允许多行同 token）。当已登录用户调用注册接口提交某 `token`，若库中已存在相同 `token` 的任意行（无论原 `wx_id`、`channel`、`device_key` 是否相同），系统 MUST 删除这些既有行（或等价保证最终仅保留当前注册结果），再为当前 `wxId` 写入/更新对应 `(device_key, channel)` 记录。同一用户使用不同手机产生不同 `token` 时，系统 MUST 允许并存多行。系统 MUST 在库层提供 `token` 唯一约束；启动迁移 MUST 清理存量重复 token（保留最近更新的一行）后再添加约束。

#### Scenario: 同机换号接管 token

- **WHEN** 账号 A 已注册 token `T`，随后账号 B 在同一设备以相同 token `T` 调用注册成功
- **THEN** 库中 MUST 不再存在账号 A 对该 `T` 的行，且 MUST 存在账号 B 持有 `T` 的行

#### Scenario: 同用户两部手机

- **WHEN** 同一 `wxId` 先后注册两个不同的 token `T1`、`T2`（不同设备）
- **THEN** 库中 MUST 同时保留 `T1` 与 `T2` 两行（在各自 deviceKey/channel 语义下）

#### Scenario: 同账号刷新注册

- **WHEN** 同一 `wxId`、同一 `deviceKey`、同一 `channel` 再次注册（token 可变或不变）
- **THEN** 系统 MUST 成功更新该键上的记录，且 MUST NOT 因本需求导致该用户失去该设备推送能力

---

## python-intent-target-action-align

<!-- source: openspec/specs/python-intent-target-action-align/spec.md -->

# python-intent-target-action-align Specification

## Purpose
TBD - created by archiving change align-python-intent-target-action. Update Purpose after archive.
## Requirements
### Requirement: 意图落地按 Python target_type 与 action 两轴分支

voice 统一意图落地（`applyUnifiedIntentResult` 及流式/非流式共用路径）MUST 以 Python `target_type` 为领域、`action` 为喂养动作，MUST NOT 将 `target_type` 直接传入 `ParseActionTargetType` 作为 CRUD 开关。权威枚举与兄弟仓 `IntentResponse` 一致：`target_type` 为 feeding|history|suggest|conversation|exit；喂养 `action` 为 start|end|one|multi|disambiguate。喂养 CRUD MUST 由 Python 经 `POST /device/history/api/event/batch` 完成；Go voice MUST NOT 调用 `DeviceHistory().AddHistory` / `UpdateHistory` / `DeleteHistory` 执行喂养写库。

#### Scenario: feeding + one 由 Python batch 落库

- **WHEN** 意图结果 `need_confirm=false` 且 `target_type=feeding` 且 `action=one`
- **AND** 事件可解析成功
- **THEN** Python MUST 经 batch 写入一次性 history 记录
- **AND** Go voice SHALL 透传 Python 回复，MUST NOT 在 Go 侧 `AddHistory`

#### Scenario: feeding + start/end 由 Python batch 落库

- **WHEN** `need_confirm=false` 且 `target_type=feeding` 且 `action` 为 `start` 或 `end`
- **THEN** Python MUST 经 batch 执行开始或结束（end 优先 `EndLatestHistoryIfMatch` 语义）
- **AND** Go voice SHALL 透传回复

#### Scenario: feeding + multi 由 Python batch 落库

- **WHEN** `need_confirm=false` 且 `target_type=feeding` 且 `action=multi` 且 `events` 非空
- **THEN** Python MUST 经 batch 按各子项 op/action 顺序写库
- **AND** Go voice MUST NOT 保留 Go 侧 multi-event 写库处理器

### Requirement: 非喂养领域按 target_type 分流

当 `need_confirm=false` 时：`target_type=history` MUST 走历史查询/search 类处理（Python 回复或读 history，不在 Go 写库）；`target_type=suggest` MUST 走成长建议类处理；`target_type=exit` MUST 走退出；`target_type=conversation` 或空 MUST 仅返回自然语言而不做喂养 CRUD。

#### Scenario: history 不误入 conversation 早退

- **WHEN** `target_type=history` 且 `need_confirm=false`
- **THEN** voice SHALL 进入历史/search 处理路径（透传 Python 回复）

#### Scenario: conversation 仅回复

- **WHEN** `target_type=conversation` 或空
- **THEN** voice SHALL 仅回复内容，SHALL NOT 写入喂养 history

### Requirement: disambiguate 不落库

当 `action=disambiguate` 且未走 `need_confirm` 澄清早退时，Python MUST NOT 执行喂养 batch 写库；Go voice MUST 透传自然语言回复。

#### Scenario: disambiguate 防御

- **WHEN** `target_type=feeding` 且 `action=disambiguate` 且 `need_confirm=false`
- **THEN** Python MUST NOT 提交喂养 batch create/start
- **AND** Go voice MUST NOT 调用 history 写库

---

## redis-disaster-recovery-runbook

<!-- source: openspec/specs/redis-disaster-recovery-runbook/spec.md -->

# redis-disaster-recovery-runbook Specification

## Purpose
TBD - created by archiving change ucg-chat-mysql-persist. Update Purpose after archive.
## Requirements
### Requirement: 项目 SHALL 提供 Redis 容灾 runbook 文档

仓库 SHALL 包含 `docs/runbooks/redis-disaster-recovery.md`，说明本项目 Redis 在 Docker 环境下的重启、卷保留、AOF 持久化与数据丢失场景下的恢复步骤。文档 SHALL 使用中文说明性文本；命令、路径、环境变量名可保留英文。

#### Scenario: 运维查阅容器重启恢复

- **WHEN** 运维人员 Redis 容器 stop/start 或 `docker compose up --force-recreate`（未使用 `down -v`）
- **THEN** runbook SHALL 提供验证步骤（PING、CLUSTER INFO、DBSIZE）
- **AND** SHALL 说明无需重复 `cluster create`（生产 cluster 卷已有元数据时）

#### Scenario: 运维查阅 volume 丢失影响

- **WHEN** 运维人员执行 `down -v` 或 volume 损坏
- **THEN** runbook SHALL 列出按数据类型的可恢复性（MySQL 权威 vs 仅 Redis vs 可 lazy warm）
- **AND** SHALL 说明 UCG 私信在方案 A 下可从 MySQL 读时回填 Redis

### Requirement: runbook SHALL 说明本项目 Redis 持久化配置

文档 SHALL 描述测试 standalone 与生产 cluster 的 compose 文件路径、`--appendonly yes`、volume 名称（如 `redis-test-data`、`redis-node-*-data`），以及 AOF 文件位于容器 `/data` 目录。

#### Scenario: 查阅 AOF 与 volume 位置

- **WHEN** 运维需要确认持久化是否启用
- **THEN** runbook SHALL 指向 `manifest/docker/docker-compose.redis-standalone.test.yml` 与 `docker-compose.redis-cluster.yml` 中的相关配置

### Requirement: runbook SHALL 提供 volume 备份与还原指引

文档 SHALL 包含 Docker volume 备份示例（如 `tar` 打包 `/data`）与还原步骤，并注明备份前应尽量降低写入或接受 point-in-time 一致性限制。

#### Scenario: 计划内维护前备份

- **WHEN** 运维在维护窗口前备份 Redis 数据
- **THEN** runbook SHALL 提供可复制的备份命令示例（测试 standalone 与生产 cluster 节点）

### Requirement: runbook SHALL 与 release-deploy-and-run 交叉引用

`docs/runbooks/redis-disaster-recovery.md` SHALL 链接至 `docs/runbooks/release-deploy-and-run.md` 中 Redis 日常重启、cluster create 误报等章节，避免重复维护冲突步骤。

#### Scenario: 生产 cluster 报 Node is not empty

- **WHEN** 运维误执行 cluster create
- **THEN** runbook SHALL 说明先查 `cluster_state:ok` 并引用 release runbook 对应排障节

---

## redis-platform-access

<!-- source: openspec/specs/redis-platform-access/spec.md -->

# redis-platform-access Specification

## Purpose
TBD - created by archiving change enforce-redis-platform-access. Update Purpose after archive.
## Requirements
### Requirement: 业务代码 SHALL 经 cachekit WithObserver 访问 Redis KV

`internal/services/**` 与 `internal/controller/**` 中的 Redis 键值读写（含 String、Hash、List、Set、Sorted Set、INCR/DECR、EXPIRE/PERSIST、EVAL 等）MUST 经 `cachekit.Cache` 接口执行，且 MUST 使用 `cachekit.WithObserver(..., cachekit.LoggingObserver{})`（或等价的 `cachekit.Default()`）包装。**SHALL NOT** 直接调用 `g.Redis().Do(...)` 或 `g.Redis()`。唯一允许直连 `g.Redis()` 的 Go 源码 MUST 位于 `internal/platform/cachekit/**` 与 `internal/platform/rediscfg/**`。

#### Scenario: 业务包无 g.Redis 直连

- **WHEN** 对 `internal/services` 与 `internal/controller` 执行 `rg 'g\.Redis\(\)' --glob '*.go'`
- **THEN** 匹配数 SHALL 为 0

#### Scenario: cachekit 操作可观测

- **WHEN** 业务经 `cachekit.WithObserver` 执行 Redis GET 且 Redis 可用
- **THEN** 观测器 SHALL 收到带 `operation`、`key`、`duration` 的回调；失败时 SHALL 以 warning 级别记录

#### Scenario: Redis 不可用错误语义一致

- **WHEN** 经 cachekit 执行操作且 Redis 返回连接/协议错误
- **THEN** 调用方 SHALL 收到 wrapping `cachekit.ErrUnavailable` 的错误

### Requirement: Redis Pub/Sub SHALL 经 redismsgkit WithObserver

Redis **PUBLISH** 与 **SUBSCRIBE**（及进程级订阅 goroutine）MUST 经 `internal/platform/redismsgkit` 抽象，且 MUST 使用 `redismsgkit.WithObserver(..., redismsgkit.LoggingObserver{})`。**SHALL NOT** 在业务或 controller 层直接 `g.Redis().Do("PUBLISH", ...)` 或使用 `github.com/redis/go-redis/v9` 创建客户端。

#### Scenario: 历史变更通知发布经 redismsgkit

- **WHEN** history-service 广播 App 历史增删改
- **THEN** MUST 调用 `redismsgkit` Publisher 向频道 `app:history:notify` 发布，且 SHALL NOT 直接 PUBLISH

#### Scenario: gateway-app 订阅经 redismsgkit

- **WHEN** gateway-app-server 启动历史 WS 通知订阅
- **THEN** MUST 经 `redismsgkit` 订阅 `app:history:notify`，且业务包 SHALL NOT 含 `redis.NewClient` / `redis.NewClusterClient`

### Requirement: Redis 键与 Pub/Sub 频道 SHALL 仅经 platform builder 构造

除 `internal/platform/cachekit/**` 与 `internal/platform/redismsgkit/**` 外，**SHALL NOT** 出现 Redis 键或 Pub/Sub 频道名字面量拼接（含 `fmt.Sprintf("domain:...")`）。既有线上键字符串 MUST 通过 platform builder 返回且与本变更前一致（策略 A：不重命名键空间）。

#### Scenario: ucg 聊天键经 keys_ucg builder

- **WHEN** ucg-service 读写会话消息 List
- **THEN** 键 MUST 由 `cachekit` 域 builder（如 `UCGChatMsgListKey(convID)`）生成，且返回值 SHALL 等于 `ucg:chat:conv:{convID}:msgs`

#### Scenario: 跨服务 ai 配额键单点定义

- **WHEN** voice-service 或 ucg-service 读写 AI 月度配额
- **THEN** 键 MUST 由 `cachekit` 的 `AIQuotaUsageKey`（或等价）生成，且 voice/ucg SHALL NOT 各自维护前缀常量

#### Scenario: App 历史通知频道单点定义

- **WHEN** 任意模块引用 App 历史 WS 通知频道名
- **THEN** MUST 使用 `redismsgkit.ChannelAppHistoryNotify`（或后继等价常量），且字符串 SHALL 为 `app:history:notify`

### Requirement: cachekit SHALL 提供全仓 typed 方法且无 Raw Do 后门

`cachekit.Cache` MUST 暴露本仓库业务所需的 typed 方法（含 `HashGetAll` 正确解析 GoFrame adapter 返回的 flat `[]string`），且 observed 包装 MUST 覆盖全部方法。**SHALL NOT** 向业务暴露通用 `Do(cmd, args...)` 逃逸口。

#### Scenario: HashGetAll 解析 flat []string

- **WHEN** 底层 `HGETALL` 经 GoFrame 返回 flat `[]string`（非 map）
- **THEN** `cachekit.HashGetAll` SHALL 仍返回正确的 `map[string]string`，SHALL NOT 因类型解析失败返回空 map

#### Scenario: 新增 Redis 命令须扩展 typed 接口

- **WHEN** 业务需要本变更未列出的 Redis 命令
- **THEN** MUST 先在 `cachekit` 或 `redismsgkit` 增加 typed 方法与观测，SHALL NOT 在业务层临时直连

### Requirement: 仓库级 AI 与代码评审 SHALL 检查 Redis platform 合规

`AGENTS.md` MUST 包含与 `openspec/project.md` 一致的 Redis 访问与键命名强制条款。PR 评审 MUST 包含对业务/controller 层 `g.Redis()` 与 Redis 键字面量的检查。

#### Scenario: AGENTS.md 含 Redis 强制节

- **WHEN** 查阅仓库根 `AGENTS.md`
- **THEN** SHALL 存在独立的 Redis 访问与 Redis 键命名强制说明，并引用 `cachekit` / `redismsgkit`

---

## redis-read-model-cache

<!-- source: openspec/specs/redis-read-model-cache/spec.md -->

# redis-read-model-cache Specification

## Purpose
TBD - created by archiving change async-redis-read-model-for-history-action-event-user. Update Purpose after archive.
## Requirements
### Requirement: Redis 优先读取历史与元数据
系统对 `history/action/event/user` 的读取 SHALL 默认优先从 Redis 读模型获取；当缓存缺失、反序列化失败或依赖异常时，系统 MUST 回源权威数据源并在成功后回填 Redis。

#### Scenario: 缓存命中直接返回
- **WHEN** 读取 `history/action/event/user` 请求命中 Redis 且数据有效
- **THEN** 系统 MUST 直接返回缓存结果且不访问数据库

#### Scenario: 缓存缺失触发回源回填
- **WHEN** 读取请求未命中 Redis
- **THEN** 系统 MUST 回源数据库或契约服务获取数据并回填到 Redis 后返回

#### Scenario: 缓存损坏自动降级
- **WHEN** Redis 中对应 key 数据格式错误或反序列化失败
- **THEN** 系统 MUST 降级为回源读取并覆盖修复该缓存 key

### Requirement: 统一缓存键空间与版本语义
系统 SHALL 为 `history/action/event/user` 定义统一域内 key 规则与版本键规则，并 MUST 在读取时识别版本语义以支持后续乱序保护和修复。

#### Scenario: 键命名符合域规范
- **WHEN** 任一模块写入或读取缓存 key
- **THEN** key MUST 满足统一格式（domain:module:kind:identifier）且可由领域缓存仓储一致生成

#### Scenario: 版本键可用于一致性判断
- **WHEN** 读取方发现实体数据键与版本键不一致
- **THEN** 系统 MUST 触发回源修复或异步修复流程并避免返回明显过期快照

### Requirement: DeepSeek 历史上下文读取复用读模型
语音链路在构造 DeepSeek prompt 时，历史与画像读取 MUST 复用 Redis 读模型优先路径，不得绕过读模型直接形成新的 DB 热点通道。

#### Scenario: 构造 prompt 时命中 Redis 历史
- **WHEN** 语音链路需要读取最近历史记录
- **THEN** 系统 MUST 优先读取 Redis 中历史读模型并用于 prompt 构造

#### Scenario: Redis 不可用时语音链路可用
- **WHEN** Redis 短时不可用
- **THEN** 系统 MUST 回源获取历史并继续完成 prompt 构造，同时记录降级日志与指标

---

## remove-clinic-tip-feedback

<!-- source: openspec/specs/remove-clinic-tip-feedback/spec.md -->

# remove-clinic-tip-feedback Specification

## Purpose
TBD - created by archiving change remove-tip-and-clinic-feedback. Update Purpose after archive.
## Requirements
### Requirement: 不得再暴露 clinic/tip HTTP 飞轮

voice-service MUST NOT 再 Bind 暴露 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`。`DeviceClinicFeedbackController`（或仅承载上述两接口的等价控制器）MUST 删除。gateway-app MUST NOT 再为 `/device/api/clinic/*`、`/device/api/tip/*` 配置指向 voice 的反代（若该 pattern 仅服务上述反馈）。

#### Scenario: clinic feedback 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/api/clinic/feedback`
- **THEN** 请求 MUST NOT 由 clinic feedback 控制器处理

#### Scenario: tip feedback 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/api/tip/feedback`
- **THEN** 请求 MUST NOT 由 tip feedback 控制器处理

#### Scenario: gateway 反代清理

- **WHEN** 审查 `installVoiceProxyMiddleware`（或等价）绑定的 path 列表
- **THEN** MUST NOT 包含仅服务于 tip SSE、tip feedback、clinic feedback 的 `/device/tip/*`、`/device/api/tip/*`、`/device/api/clinic/*` pattern

### Requirement: care-alert 飞轮必须保留

本变更 MUST NOT 删除或下线 `POST /device/api/care-alert/feedback`。care-alert daily / delete item / feedback 仍由 **voice-service** 宿主处理，gateway MUST 继续将 `/device/api/care-alert/*` 反代至 voice。

#### Scenario: care-alert feedback 仍注册

- **WHEN** 审查 voice-service HTTP 注册与 care-alert 控制器
- **THEN** `POST /device/api/care-alert/feedback` MUST 仍由 care-alert 控制器处理，且服务层 `CareAlertFeedback`（及 Python CareAlertFeedback 客户端若原先存在）MUST 仍可用

#### Scenario: care-alert 反代仍在

- **WHEN** 审查 gateway voice 域反代 path 列表
- **THEN** MUST 仍包含 `/device/api/care-alert/*`

### Requirement: clinic WS 不受本变更影响

本变更 MUST NOT 移除或改坏 `/voice/clinic/ws`（及等价 clinic 连续对话 WebSocket）注册与反代。

#### Scenario: clinic WS 仍可达

- **WHEN** 审查 gateway WS 反代与 voice clinic WS 控制器
- **THEN** clinic WebSocket 路径 MUST 仍注册且可被反代

### Requirement: usage skip 与已删路径一致

`maintenance_skip.go`（或等价）MUST NOT 再保留已删除的 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback` 排除项。MUST NOT 因本变更新增对 care-alert 路径的排除（除非负责人另行确认；本变更默认不动 care-alert usage 策略）。

#### Scenario: skip 列表无死路径

- **WHEN** 审查 usage 维护型排除列表
- **THEN** MUST NOT 出现已下线的 clinic/tip feedback 精确 path 条目

---

## remove-device-tip

<!-- source: openspec/specs/remove-device-tip/spec.md -->

# remove-device-tip Specification

## Purpose
TBD - created by archiving change remove-tip-and-clinic-feedback. Update Purpose after archive.
## Requirements
### Requirement: 服务端不得再暴露小贴士 SSE

voice-service 与 gateway-app MUST NOT 再对外提供 `POST /device/tip/generate`（及 `/device/tip/*` 反代）。`TipCtrl`（或等价）MUST NOT 再注册；`VoiceService.TipStream` 与仅服务于 tip 的 Python 客户端方法 MUST 删除或不可达。

#### Scenario: tip generate 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/tip/generate`
- **THEN** 请求 MUST NOT 由 tip SSE 控制器处理（典型为 404 或未绑定路由），且 voice-service HTTP Bind 列表 MUST NOT 包含 `NewTipCtrl()`（或等价）

#### Scenario: TipStream 从运行时移除

- **WHEN** 检索 voice 服务实现与 contracts 中的 tip 流式入口
- **THEN** MUST NOT 存在可被 HTTP 控制器调用的 `TipStream` 业务路径

### Requirement: 运维页不得再调用 tip generate

`resource/public/history.html`（及本仓内其它运维静态页若存在同等调用）MUST NOT 再请求 `/device/tip/generate`，相关调试 UI MUST 删除或改为不可用状态。

#### Scenario: history 运维无 tip 请求

- **WHEN** 审查 `resource/public/history.html` 源码
- **THEN** MUST NOT 出现对 `/device/tip/generate` 的 `fetch`/XHR 调用

---

## remove-prediction-slot-unlock

<!-- source: openspec/specs/remove-prediction-slot-unlock/spec.md -->

# remove-prediction-slot-unlock Specification

## Purpose
TBD - created by archiving change revoke-admin-grant-hub-tiles. Update Purpose after archive.
## Requirements
### Requirement: 系统 MUST NOT 再种子或特判 prediction_unlock

cash-service `EnsureSchema` MUST NOT 再插入或每次启动写回 `feature_def.feature_id=prediction_unlock`。履约、下单、Apple 通知、Admin 规则与 catalog MUST NOT 再为该 `feature_id` 走「预测条数 / `allowed_count_delta`」专用分支。若请求仍携带该 `feature_id`，系统 MUST 拒绝开通并记录日志，MUST NOT 把它当成普通权益授予。

#### Scenario: 启动不再写预测种子

- **WHEN** cash-service 执行 EnsureSchema
- **THEN** 系统 MUST NOT 执行针对 `prediction_unlock` 的 INSERT 或将其 `status` 写回的 UPDATE

#### Scenario: 旧请求被拒绝

- **WHEN** 支付履约或 Admin 授功能的 `featureId` 为 `prediction_unlock`
- **THEN** 系统 MUST 拒绝，MUST NOT 增加 `feature_allowed_count`

### Requirement: Admin MUST NOT 再提供预测条数表单项

`cash-feature-admin.html` MUST NOT 再展示「增加可看数量」、默认开通条数，或手工授里的授予数量字段。SKU 的 `grant_kind` 表单 MUST NOT 再提供 `allowed_count_delta`。

#### Scenario: 打开功能 SKU 表单

- **WHEN** 运维编辑任一上架功能的套餐
- **THEN** 表单 MUST NOT 出现 `allowed_count_delta` 选项

### Requirement: 删除预测代码 MUST NOT 删除历史表

实现 MUST NOT DROP `feature_allowed_count`，MUST NOT 删除已有 `feature_order` / `feature_product` 历史行。其它功能（值得留意、成长轨迹、邀请码、群二维码、VIP）的履约路径 MUST 保持可用。

#### Scenario: 历史表仍在

- **WHEN** 本变更部署完成
- **THEN** `feature_allowed_count` 表 MUST 仍然存在，且 care / growth 的支付与邀请开通 MUST 仍可成功

---

## routing-key-governance

<!-- source: openspec/specs/routing-key-governance/spec.md -->

# routing-key-governance Specification

## Purpose
定义路由键治理约束，确保事件发布链路仅允许已注册路由键并集中管理路由定义。

## Requirements
### Requirement: 路由键白名单校验
系统 MUST 对 outbox 写入与事件发布链路中的 `routing_key` 执行白名单校验，未注册路由键不得进入发布流程。

#### Scenario: 合法路由键正常发布
- **WHEN** 业务写入或发布已注册的 `routing_key`
- **THEN** 系统 MUST 允许进入 outbox 与发布流程

#### Scenario: 非法路由键被拒绝
- **WHEN** 业务使用未注册的 `routing_key`
- **THEN** 系统 MUST 拒绝该请求并输出结构化错误日志

### Requirement: 路由键集中定义
系统 SHALL 提供集中路由键定义与查询接口，避免在多个模块重复硬编码路由字符串。

#### Scenario: 新增路由键需要注册
- **WHEN** 开发新增事件路由键
- **THEN** 该路由键 MUST 先在集中注册处声明后才能被调用方引用

---

## routing-key-governance-workflow

<!-- source: openspec/specs/routing-key-governance-workflow/spec.md -->

# routing-key-governance-workflow Specification

## Purpose
TBD - created by archiving change routing-key-prefix-governance-and-dispatch. Update Purpose after archive.
## Requirements
### Requirement: 新增路由键必须遵循注册流程
系统 MUST 定义并执行新增路由键的标准流程：前缀确认、路由键注册、分发映射、观测校验、文档更新。

#### Scenario: 开发者新增路由键
- **WHEN** 开发者新增一个路由键用于新事件
- **THEN** 必须同时完成注册、分发映射和文档更新，缺任一项视为未完成迁移

### Requirement: 迁移验收必须禁止新增核心裸字符串匹配
系统 SHALL 在迁移验收清单中明确要求：核心分发模块不得新增针对 `routing_key` 的裸字符串匹配分支。

#### Scenario: 代码评审检查分发逻辑
- **WHEN** 评审者检查 outbox 与投影分发相关改动
- **THEN** 若发现新增裸字符串匹配而未使用统一前缀/枚举入口，必须拒绝合并

---

## routing-key-prefix-registry

<!-- source: openspec/specs/routing-key-prefix-registry/spec.md -->

# routing-key-prefix-registry Specification

## Purpose
TBD - created by archiving change routing-key-prefix-governance-and-dispatch. Update Purpose after archive.
## Requirements
### Requirement: 路由键前缀必须集中注册
系统 MUST 在统一注册入口维护路由键前缀常量，并将前缀作为领域分组的唯一来源，禁止在业务模块重复定义核心前缀字面量。

#### Scenario: 新增前缀时集中维护
- **WHEN** 开发者需要新增一个事件族前缀
- **THEN** 必须在统一注册入口新增前缀常量并在路由键定义处复用该常量

### Requirement: 路由键定义必须采用前缀与后缀组合
系统 SHALL 通过“前缀常量 + 后缀常量/字面量”生成路由键枚举值，保证同一事件族的命名一致性。

#### Scenario: 定义 history 事件路由键
- **WHEN** 定义 `history.record.created`、`history.record.updated`、`history.record.deleted`
- **THEN** 这些路由键必须共享 `history.record.` 前缀常量并仅通过后缀区分

---

## runtime-dependency-check

<!-- source: openspec/specs/runtime-dependency-check/spec.md -->

# runtime-dependency-check Specification

## Purpose
TBD - created by archiving change rabbitmq-optional-startup-check. Update Purpose after archive.
## Requirements
### Requirement: API 类进程启动 SHALL 不因 RabbitMQ 不可达而失败

gateway-service、gateway-app-server、device-service、history-service、voice-service、ucg-service 在启动前 SHALL 校验 Redis 连通性；RabbitMQ 管理 API 探活 MAY 失败且不 SHALL 阻断进程启动。探活失败时 SHALL 记录 Warning 级别日志。

#### Scenario: RabbitMQ 宕机时 device-service 启动

- **WHEN** device-service 启动且 Redis 可达但 RabbitMQ 管理 API 不可达
- **THEN** 进程 SHALL 成功启动并监听 HTTP

#### Scenario: Redis 不可达时 API 进程启动

- **WHEN** 任一 API 类进程启动且 Redis 不可达
- **THEN** 进程启动 SHALL 失败

### Requirement: worker-service 启动 SHALL 保持 RabbitMQ 强依赖

worker-service 在启动前 SHALL 同时校验 Redis 与 RabbitMQ；任一失败 SHALL 导致进程启动失败。

#### Scenario: RabbitMQ 宕机时 worker-service 启动

- **WHEN** worker-service 启动且 RabbitMQ 管理 API 不可达
- **THEN** 进程启动 SHALL 失败

---

## runtime-docs-centralization-and-governance

<!-- source: openspec/specs/runtime-docs-centralization-and-governance/spec.md -->

# runtime-docs-centralization-and-governance Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Runtime docs SHALL be centralized and governed

`dao-sync-by-domain.md` and `release-deploy-and-run.md` MUST be maintained in a dedicated runtime-docs directory, and change governance MUST require synchronized updates when runtime behavior changes. Changes that introduce or alter Compose prod/test dual-stack deployment, registry image tagging, or test seed desensitization MUST update `release-deploy-and-run.md` in the same change.

#### Scenario: Docs location is consolidated

- **WHEN** checking runtime operation documents
- **THEN** both target documents MUST exist under one dedicated new folder

#### Scenario: Governance requires synchronized update

- **WHEN** project runtime/deployment/database boundary rules change
- **THEN** project governance (`openspec/project.md`) MUST require updating both runtime docs

#### Scenario: Dual-stack change updates runbook

- **WHEN** a change adds prod/test Compose overlays or test seed procedures
- **THEN** `release-deploy-and-run.md` MUST be updated in that change before merge

### Requirement: Runbook SHALL 文档化 2G 双栈 survival 配置

`docs/runbooks/release-deploy-and-run.md` SHALL 包含 **2G ECS**（或 documented 低内存同机双栈）专节，至少包括：

- 生产 Redis **3 主 0 从** 与测试 Redis **单机** 拓扑及迁移步骤（含 `down -v` 与数据丢失说明）
- prod/test 启动命令叠加 **`docker-compose.resources.{prod,test}.yml`**
- 默认 **mem_limit / cpus** 对照表与 `docker stats` 验收
- MySQL 同机 **`innodb_buffer_pool_size`** 建议（如 256M 级）
- ASR 验收约定：生产微服务保持 Up，**仅 test 域名** 进行语音压测，避免 prod 并发 ASR
- OOM / swap 排错

#### Scenario: 运维按 2G 文档完成测试 Redis 迁移

- **WHEN** 运维阅读 runbook 2G 专节并按步骤从六节点 test cluster 迁到 standalone
- **THEN** 其 SHALL 能完成 standalone Redis 启动、`.env.test` 更新与微服务 recreate，且 **无需** `cluster create`

#### Scenario: 运维按文档叠加资源 limits 启动双栈

- **WHEN** 运维按 runbook 生产/测试启动命令启动双栈
- **THEN** 命令示例 SHALL 包含 `-f docker-compose.resources.prod.yml` 或 test 等价文件

### Requirement: release-deploy-and-run SHALL 文档化 Compose 镜像版本控制

`docs/runbooks/release-deploy-and-run.md` SHALL 包含章节说明：`docker-compose.microservices.yml` 与 prod/test overlay 的关系；镜像仓库 `${REGISTRY}` 与 `${IMAGE_TAG}` 用法；测试默认 `develop`、生产 semver 钉扎；registry `pull` + `--no-build` 部署；禁止生产使用 `:local`/`develop`；按服务镜像 tag 回滚步骤。

#### Scenario: 运维按 runbook 回滚生产镜像

- **WHEN** 生产发版后需回滚至上一 semver
- **THEN** runbook SHALL 提供将 `.env.prod` 中 `IMAGE_TAG` 改回上一版本并 `pull` + `force-recreate` 的步骤

### Requirement: release-deploy-and-run SHALL 文档化生产测试双栈部署

`docs/runbooks/release-deploy-and-run.md` SHALL 包含生产/测试双栈对照表（网络、端口、库名、静态目录、中间件端口）、测试栈启动顺序（test 网络 → test Redis cluster → test Rabbit 初始化 → microservices test）、Nginx 反代 `test.pangbao.cuplay.top:9701/9702`、健康检查 URL（对外形态与生产一致仅域名不同）、脱敏种子刷新与发版前 checklist。

#### Scenario: 运维首次搭建测试栈

- **WHEN** 运维仅阅读 `release-deploy-and-run.md`
- **THEN** 其 SHALL 能按文档顺序完成测试中间件与微服务启动，并验证 `https://test.pangbao.cuplay.top:9702/api.json`（或 documented 等价 HTTPS 探活）

---

## service-boundary-no-cross-db

<!-- source: openspec/specs/service-boundary-no-cross-db/spec.md -->

# service-boundary-no-cross-db Specification

## Purpose
TBD - created by archiving change enforce-service-boundary-no-cross-db. Update Purpose after archive.
## Requirements
### Requirement: 服务边界 MUST 与数据库边界一致
每个微服务 MUST 仅访问其所属数据库/表；跨服务数据访问 MUST 通过服务契约（HTTP/RPC/事件）进行，MUST NOT 通过 DAO 或 SQL 直连他域数据表。

#### Scenario: Voice 访问 history 数据
- **WHEN** voice 需要读取或更新 history 领域数据
- **THEN** voice MUST 通过 history 服务契约调用完成，代码中 MUST NOT 出现对 history DAO 的直接查询或更新

#### Scenario: Voice 访问 device/user 画像数据
- **WHEN** voice 需要读取设备资料、生日、性别或注册状态
- **THEN** voice MUST 通过 device 服务契约调用完成，代码中 MUST NOT 出现对 `user/device` 领域 DAO 的直接查询

#### Scenario: 代码评审发现跨库直查
- **WHEN** 新增代码出现跨服务数据库直连访问
- **THEN** 该变更 MUST 被视为违反架构约束并在合入前整改

### Requirement: 迁移期分流 MUST 支持可控回退
服务边界治理迁移期 MUST 支持 `local|remote|canary` 切换，并对同一分流键保持稳定命中，确保可渐进放量与快速回滚。

#### Scenario: Canary 分流验证
- **WHEN** 开启 canary 模式并设置百分比
- **THEN** 同一设备标识 MUST 稳定命中同一路径，避免在 local 与 remote 之间抖动

#### Scenario: 远程路径故障回退
- **WHEN** remote 路径连续失败且开启 failover 配置
- **THEN** 调用方 MUST 回退到 local 路径并记录可观测日志

### Requirement: 单服务单库部署下禁止进程内跨域 DAO

当进程配置的 `database.default` 仅包含本服务所属库时，业务代码 MUST NOT 通过 import 他域服务包并调用其基于 DAO 的实现来访问他域表；跨服务数据 MUST 经 HTTP/RPC/消息。禁止依赖「同一代码仓库、不同包」造成可连他域表的假象。

#### Scenario: Voice 进程仅配置 voice 库

- **WHEN** `voice-service` 的配置仅连接 `qa`/`suggest` 所在库
- **THEN** 代码路径 MUST NOT 执行对 `user`/`event`/`action` 等表的 DAO；必须通过 device-service 契约

#### Scenario: 评审发现 voice 包引用他域 DAO

- **WHEN** 代码评审发现 `internal/services/voice` 直接或间接触发他域 `dao` 访问
- **THEN** 该变更 MUST 拒绝合入，直至改为 HTTP 客户端或经批准的同进程例外（文档化且仅限非生产）

### Requirement: Device 进程内 outbox 写入 MUST 使用显式 history 库连接

`domain_outbox` 若仅存在于 history 库，device-service MUST 使用独立配置的 `history_relay`（或等价）连接组写入，MUST NOT 误用 `default` 连接组指向 device 库写 outbox。

#### Scenario: 分库部署

- **WHEN** device 与 history 为不同数据库实例
- **THEN** 未正确配置 relay 时 MUST 跳过或失败可观测，MUST NOT 静默写入错误库

### Requirement: 表归属 MUST 与部署库一致

在分库部署下，`history` 表与 `domain_outbox` MUST 仅由可连接 history 库的进程写入；`user`、`event`、`action` MUST 仅由可连接 device 库的进程写入；`qa`、`suggest` MUST 仅由可连接 voice 库的进程写入。禁止因历史单体代码路径而使用错误默认库连接组访问上述表。

#### Scenario: device 进程不写 history 库中的 outbox（除非显式配置）

- **WHEN** `domain_outbox` 表仅存在于 history 服务数据库中
- **THEN** device-service MUST NOT 使用 `user` 表所在连接组对 `domain_outbox` 执行 Insert，除非运维显式配置为同一物理库且经架构评审

#### Scenario: voice 进程不写 event/action 表

- **WHEN** voice-service 需要新增或查询事件字典、动作记录
- **THEN** voice MUST 通过 device 服务契约完成，MUST NOT 使用 `dao.Event` 或 `dao.Action` 直连 device 库表

### Requirement: history 服务 MUST NOT 冒充他域数据权威

对外 HTTP 或内部契约 MUST NOT 将「生日、事件选项、语音建议」等响应伪装为 history 数据库本地查询结果；若经网关聚合， MUST 在实现上分别调用 device/voice 权威服务，且错误语义可追溯至真实下游。

#### Scenario: 拆分后的 API 归属

- **WHEN** 客户端请求事件选项或用户画像
- **THEN** 响应数据 MUST 来源于 device 域存储与接口，而非 history 进程内对 `event`/`user` 表的 DAO 查询

### Requirement: ucg-service MUST NOT cross-read device database for wx data

ucg-service MUST treat `wx` table as device-domain data. All wx validation, batch profile display fields, and baby_name for default nickname MUST be fetched via device-service internal HTTP APIs with `DEVICE_GATEWAY_INTERNAL_SECRET`. ucg-service MUST NOT import device DAO or execute SQL against device database.

#### Scenario: ucg 读取 wx 展示名
- **WHEN** ucg-service 需要渲染帖子作者昵称且 ucg_profile 缺失
- **THEN** 服务 MUST 调用 device internal batch API，且 MUST NOT 查询 device 库 `wx` 表

#### Scenario: 评审发现 ucg 跨库 DAO
- **WHEN** 代码评审发现 ucg-service 直连 `wx` 表
- **THEN** 变更 MUST 拒绝合入

### Requirement: ucg 表 MUST 仅由 ucg-service 写入 ai_voice_ucg

Tables `ucg_*` MUST reside in database `ai_voice_ucg` and MUST only be written by ucg-service default connection. Other services MUST NOT insert/update ucg tables via cross-DB SQL.

#### Scenario: gateway 不写 ucg 表
- **WHEN** gateway-app 代理 UCG HTTP 请求
- **THEN** gateway MUST NOT 直接写入 `ucg_post` 或任何 ucg 表

---

## service-code-full-cutover

<!-- source: openspec/specs/service-code-full-cutover/spec.md -->

# service-code-full-cutover Specification

## Purpose
TBD - created by archiving change migrate-service-to-services-full-cutover. Update Purpose after archive.
## Requirements
### Requirement: `internal/service` 实现文件 MUST 全量迁移
系统 MUST 将 `internal/service` 中实现文件按领域归属迁移到 `internal/services/*` 或 `internal/shared/*`，迁移完成后不得遗留可编译业务实现文件。

#### Scenario: 全量迁移完成
- **WHEN** 执行迁移收口检查
- **THEN** `internal/service` 中不得再存在业务实现文件，且对应实现已在目标目录可追踪

### Requirement: 迁移后调用路径 MUST 指向新目录
所有服务入口、控制器和内部调用方 MUST 使用迁移后的包路径，不得继续依赖旧 `internal/service` 路径。

#### Scenario: 调用路径校验
- **WHEN** 对迁移范围执行 import 路径审查
- **THEN** 迁移后的调用引用 MUST 全部指向 `internal/services/*` 或 `internal/shared/*`

---

## service-dedicated-config-loading

<!-- source: openspec/specs/service-dedicated-config-loading/spec.md -->

# service-dedicated-config-loading Specification

## Purpose
TBD - created by archiving change service-dedicated-config-final-boundary. Update Purpose after archive.
## Requirements
### Requirement: 服务配置 MUST 按服务进程独立加载
`voice-service`、`device-service`、`history-service`、`gateway-service` MUST 具备独立默认配置文件，服务启动时 MUST 优先使用本服务默认配置，并允许通过 `GF_GCFG_FILE` 显式覆盖。

#### Scenario: voice-service 启动未指定 GF_GCFG_FILE
- **WHEN** 启动 `voice-service` 且环境变量未设置 `GF_GCFG_FILE`
- **THEN** 系统 MUST 加载 `voice-service` 专属默认配置文件

#### Scenario: device-service 启动指定 GF_GCFG_FILE
- **WHEN** 启动 `device-service` 且设置 `GF_GCFG_FILE`
- **THEN** 系统 MUST 使用指定配置文件并覆盖默认路径

### Requirement: 服务级覆盖变量 MUST 仅影响本服务
服务级环境变量覆盖（如数据库连接、监听地址）MUST 仅影响当前服务实例，不得通过同名变量隐式影响其他服务配置行为。

#### Scenario: 设置 VOICE_DB_LINK
- **WHEN** 部署仅设置 `VOICE_DB_LINK`
- **THEN** 系统 MUST 只影响 `voice-service` 数据库连接，不得改变 `history-service` 与 `device-service` 连接

---

## service-import-isolation

<!-- source: openspec/specs/service-import-isolation/spec.md -->

# service-import-isolation Specification

## Purpose
TBD - created by archiving change service-package-isolation. Update Purpose after archive.
## Requirements
### Requirement: 业务服务包 MUST NOT 互相 import 实现

`internal/services` 下各业务域包（至少：cash、voice、device、history、ucg、gatewayapp、simuser、mcpbridge、appstatus）MUST NOT 直接 import 另一业务域包中的实现代码。跨域协作 MUST 通过 `internal/clients/*` 出站调用、消息事件或 `contracts` 接口（实现由本域或 clients 装配）。共享无业务工具 MUST 放在 `internal/platform/**`（或同等 platform 包）。`aimodel` 与 `contracts` MAY 被各业务包 import，视为共享基础设施而非业务线包。

#### Scenario: voice 不再 import services/cash

- **WHEN** 检查 `internal/services/voice` 的 Go import
- **THEN** MUST NOT 出现 `hello/internal/services/cash`

#### Scenario: ucg 不再 import services/device 业务实现

- **WHEN** 检查 `internal/services/ucg` 的 Go import
- **THEN** MUST NOT 出现对 `hello/internal/services/device` 的依赖（设备 HTTP 经 clients；类型/密钥头经 platform 或 clients DTO）

#### Scenario: 允许 platform 与 contracts

- **WHEN** voice 需要解析内部 wxId 头或调用 DeviceAdmin 契约
- **THEN** MAY import `platform` 与 `contracts`，MUST NOT 因此 import device 业务实现包

### Requirement: 伪公共工具 MUST 下沉 platform

下列能力 MUST 从业务域包迁出至 `internal/platform`（或等价 platform 子包），供多进程共用：内部网关密钥头与校验、App/内部注入头常量（如 wxId/deviceNo/clientIP）、从请求头解析 wxId、常量时间比较等。迁出后，原业务包 MUST NOT 再作为他域获取这些工具的唯一来源。

#### Scenario: cash controller 不 import voice 解析 wxId

- **WHEN** cash 控制器需要解析 `X-Internal-Wx-Id`
- **THEN** MUST 使用 platform（或等价）解析函数，MUST NOT import `internal/services/voice`

#### Scenario: history/ucg 不借 device 包做密钥校验

- **WHEN** history 或 ucg 内部接口校验网关内部密钥
- **THEN** MUST 使用 platform 校验，MUST NOT 仅为校验而 import device 业务包

### Requirement: 仓库 MUST 提供跨域 import 门禁

仓库 MUST 提供可运行的检查脚本（路径以实现为准，如 `hack/check-service-import.sh`），用于检测业务服务包之间的禁止 import；检查失败时 MUST 以非零退出码表示（可先 warn 后在同一变更内收紧为 fail）。`AGENTS.md` 与 `openspec/project.md` MUST 记载源码包边界约定，与「跨库须走 HTTP」并列。

#### Scenario: 门禁检出 voice→cash

- **WHEN** 有人重新引入 `services/voice` → `services/cash` 的 import 并运行门禁
- **THEN** 检查 MUST 失败（收紧后）或明确告警（过渡配置下）

#### Scenario: clients 与 platform 不误杀

- **WHEN** `services/voice` 仅 import `clients/cash` 与 `platform`
- **THEN** 门禁 MUST 通过

---

## service-migration-safety-and-rollback

<!-- source: openspec/specs/service-migration-safety-and-rollback/spec.md -->

# service-migration-safety-and-rollback Specification

## Purpose
TBD - created by archiving change migrate-service-to-services-full-cutover. Update Purpose after archive.
## Requirements
### Requirement: 迁移执行 MUST 分批且可验证
全量迁移 MUST 以可回滚批次执行，每批完成后必须通过编译校验与关键服务启动校验后方可进入下一批。

#### Scenario: 批次完成校验
- **WHEN** 单个迁移批次完成
- **THEN** 系统 MUST 通过既定编译检查与启动健康检查

### Requirement: 迁移异常 MUST 支持按服务维度回滚
迁移引发异常时，系统 MUST 支持按受影响服务维度回退代码与配置，不要求全局回滚。

#### Scenario: 单服务回滚
- **WHEN** `voice-service` 批次迁移后出现运行异常
- **THEN** 团队 MUST 可仅回滚 `voice-service` 相关迁移批次并恢复可用

### Requirement: 收口验收 MUST 覆盖关键链路无回归
全量迁移收口时 MUST 验证 gateway/voice/device/history 关键链路，确保外部行为与迁移前一致。

#### Scenario: 收口链路验收
- **WHEN** 所有批次迁移完成并准备收口
- **THEN** 关键业务链路 MUST 通过无回归验证，且迁移结果可被文档化追踪

---

## service-outbound-clients

<!-- source: openspec/specs/service-outbound-clients/spec.md -->

# service-outbound-clients Specification

## Purpose
TBD - created by archiving change service-package-isolation. Update Purpose after archive.
## Requirements
### Requirement: 域间出站客户端 MUST 位于中立 clients 包

系统 MUST 将调用他域服务的 HTTP/WS 出站客户端放置于 `internal/clients/{target}/`（`target` 为被调服务域名，如 `cash`、`device`、`history`、`ucg`、`voice`）。被调方业务包 `internal/services/{target}` MUST NOT 再向其他服务进程导出 `Remote*` 或等价「调本域 internal API」的客户端符号供其 `import`。调用方业务包 MUST 通过 `clients/{target}` 发起跨进程调用，MUST NOT 为调用他域而 `import` 他域 `internal/services/{other}` 业务实现包。

#### Scenario: voice 调 cash VIP/access

- **WHEN** voice-service 需要查询 VIP 或值得留意 access
- **THEN** 源码 MUST `import` `internal/clients/cash`（或等价 clients 路径），MUST NOT `import hello/internal/services/cash`

#### Scenario: cash 调 device/history

- **WHEN** cash-service 需要设备号或喂养日统计
- **THEN** 出站客户端 MUST 位于 `internal/clients/device` 与 `internal/clients/history`（或从原 `cash/*_client.go` 迁入后由此引用），MUST NOT 要求 device/history 业务包反向依赖 cash

#### Scenario: 非本仓域依赖可例外

- **WHEN** 出站目标为 Python AI、阿里云内容安全或外部 LLM
- **THEN** 客户端 MAY 保留在原业务包或非域名 clients 子目录，MUST NOT 因此允许域业务包互相 import

### Requirement: 出站客户端迁入 MUST NOT 改变内部 API 契约语义

迁入 clients 时，请求的 path、方法、内部密钥头与响应解析语义 MUST 与迁入前一致（允许包路径与符号名变更）。MUST NOT 借机修改 App 对外路径。

#### Scenario: CareAlert access 路径不变

- **WHEN** voice 经 clients/cash 调用值得留意 access
- **THEN** 请求 URL path MUST 仍为既有 `/cash/internal/api/care-alert/access`（或迁入前实际使用的同一 path）

---

## service-runtime-standardization

<!-- source: openspec/specs/service-runtime-standardization/spec.md -->

# service-runtime-standardization Specification

## Purpose
TBD - created by archiving change platform-hardening-redis-rabbitmq-service-split. Update Purpose after archive.
## Requirements
### Requirement: 缓存键与 TTL 规范必须统一
所有服务 SHALL 遵循统一的 Redis key 命名空间规范和已文档化的 TTL 规则，覆盖缓存、守卫与幂等状态。

#### Scenario: 任一服务引入新缓存键
- **WHEN** 某服务为运行时状态新增 Redis key
- **THEN** 该 key SHALL 符合统一命名规范与 TTL 策略

### Requirement: 事件命名与投递语义必须统一
所有跨服务事件 SHALL 使用统一的 exchange/routing-key 命名规范，并遵循明确的投递失败语义。

#### Scenario: 服务发布跨服务事件
- **WHEN** 某服务发出跨服务处理所需的领域事件
- **THEN** 其 SHALL 通过 RabbitMQ 按统一 exchange/routing-key 规范发布，并执行既定发布失败行为

### Requirement: 本迁移阶段执行禁测文件策略
代码库 SHALL 删除现有 Go 测试文件，并在本迁移阶段 SHALL NOT 新增 Go 测试文件。

#### Scenario: 迁移阶段引入新代码
- **WHEN** 开发者在本迁移范围内新增或重构代码
- **THEN** 其 SHALL 通过运行时核验脚本与运维检查进行验证，而不是新增 `*_test.go` 文件

---

## sim-llm-lane-admin

<!-- source: openspec/specs/sim-llm-lane-admin/spec.md -->

# sim-llm-lane-admin Specification

## Purpose
TBD - created by archiving change ai-model-admin-with-sim-runtime-db. Update Purpose after archive.
## Requirements
### Requirement: sim-user-service SHALL persist sim LLM lanes in database

`sim-user-service` MUST 在 `SIM_DB_LINK` 对应库维护表 **`sim_llm_lane_config`**（或语义等价名），主键 `lane`，取值 `simText`、`simVision`、`simImageGen`、`simVideoGen`。每行 MUST 含 `provider`、`model`、`max_in_flight`、`max_waiters`、`timeout_sec`（可选）、`updated_at`、`updated_by`。

`SimLLMLaneStore.Load(lane)` MUST 按优先级：**DB 行** > **环境变量**（`SIM_LLM_*`，仅 seed/迁移）> **代码默认种子**。EnsureSchema MUST 在无行时写入默认种子（与当前 env 默认语义一致）。

#### Scenario: 新环境 DB 种子

- **WHEN** sim 库尚无 `sim_llm_lane_config` 行且进程首次 EnsureSchema
- **THEN** 四 lane MUST 存在默认行且 Admin GET MUST 可读

#### Scenario: DB 覆盖 env

- **WHEN** DB 行与 `SIM_LLM_*` env 冲突
- **THEN** 运行时 MUST 使用 DB 值

### Requirement: sim-user-service SHALL provide sim LLM lane Admin API

sim-user-service MUST 提供 `GET /sim/admin/api/llm-lanes` 与 `PUT /sim/admin/api/llm-lanes`，鉴权 MUST 与现有 sim-admin 一致（Header `X-Admin-Password`）。响应与请求 MUST 含四 lane 子对象，各含 `provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。GET MUST 返回 provider→model allowlist（含 zhipu 生图/生视频 model）。PUT MUST 校验 allowlist 与边界（`maxInFlight>=1`，`maxWaiters>=0`）。PUT 成功后 MUST 调用 `aimodel.InvalidateLaneCache()` 且 MUST NOT 触发 scheduler reload。

#### Scenario: 管理员读取 sim lanes

- **WHEN** 已鉴权 GET `/sim/admin/api/llm-lanes`
- **THEN** 响应 MUST 含四 lane 配置与 allowlist

#### Scenario: 管理员更新 simText 并发

- **WHEN** PUT `simText.maxInFlight=2` 且 model 在 allowlist
- **THEN** sim-service MUST 持久化且 MUST 失效 lane 缓存

#### Scenario: sim LLM PUT 不 reload scheduler

- **WHEN** 仅 PUT llm-lanes 成功
- **THEN** scheduler goroutine MUST NOT 因该 PUT 而 Stop/Start

### Requirement: aimodel allowlist SHALL include sim image and video generation models

`internal/services/aimodel` 的 `ProviderModels`（zhipu）MUST 包含 `cogview-3-flash`、`cogvideox-flash`（及 normalize 大小写变体），供 simImageGen/simVideoGen Admin PUT 校验通过。

#### Scenario: Admin 配置 sim 生图 model

- **WHEN** PUT simImageGen model=`cogview-3-flash`
- **THEN** 校验 MUST 通过且 MUST 持久化

### Requirement: sim LLM lane Admin API MUST NOT count toward App usage stats

`/sim/admin/api/llm-lanes` MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: sim llm-lanes PUT 不计 usage

- **WHEN** 管理员 PUT llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增

---

## sim-runtime-config

<!-- source: openspec/specs/sim-runtime-config/spec.md -->

# sim-runtime-config Specification

## Purpose
TBD - created by archiving change ai-model-admin-with-sim-runtime-db. Update Purpose after archive.
## Requirements
### Requirement: sim-user-service SHALL persist runtime task and interval configuration in database

除既有 `sim_config.enabled`、`max_sim_users` 外，sim 库 MUST 持久化以下运行时项（列或 JSON 字段，语义等价即可）：

- **taskSwitches**：`register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`（bool）；**MUST NOT** 含 `videoPoll`
- **intervals**（秒或 duration）：`register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`；**MUST NOT** 含 `videoPollIdle`、`videoPollActive`
- **rateLimit**：`ucgRateLimitRps`、`ucgRateLimitBurst`

`LoadRuntimeFromDB(ctx)` MUST 组装进程内 `RuntimeFlags`；**DB 优先**，env 仅作迁移期兜底。进程级 **`SIM_USER_SERVICE_ENABLED=false`** MUST 仍为硬闸。

读取旧 `runtime_json` 时：若缺失 `intervalPostVideoPollSec` 且存在 `intervalVideoPollActiveSec`，MAY 将其作为 poll interval 初值；`videoPoll` 与 idle 字段 MUST 忽略。缺失 `taskPostDebate` / `taskDebateComment` / 对应 interval 时 MUST 回退 env 默认（T7/T8 默认开启、12h/1h）。

#### Scenario: DB 运行时覆盖 env

- **WHEN** DB 中 `interval_post_debate=43200` 且 env 兜底为 12h
- **THEN** 新启动的 T7 goroutine MUST 使用 DB 值

#### Scenario: 进程总闸仍读 env

- **WHEN** `SIM_USER_SERVICE_ENABLED=false` 且 DB `enabled=true`
- **THEN** scheduler MUST NOT 启动

### Requirement: sim-user-service SHALL support scheduler reload on runtime config save

`PUT /sim/admin/api/config` 持久化后：

- 变更 **taskSwitches**、**T4 调度周期**（`postVideo`）、**rateLimit**、ephemeral、**enabled** → MUST Reload scheduler
- 变更 **仅** `postVideoPollInterval` / `postVideoPollMaxWait` → MUST NOT 触发 scheduler Reload（进行中的 poll 使用启动时快照）
- 变更 **仅** `maxSimUsers` → MAY 跳过 Reload
- LLM lane PUT MUST NOT 触发 Reload

#### Scenario: 修改 postVideoPollMaxWait 不 reload

- **WHEN** PUT 仅变更 `postVideoPollMaxWait`
- **THEN** 响应 MUST 含 `scheduleReloaded=false`

#### Scenario: 修改 postVideo 调度周期触发 reload

- **WHEN** PUT 变更 `interval_post_video`
- **THEN** 响应 MUST 含 `scheduleReloaded=true`

### Requirement: sim config PUT response SHALL describe save effects

`PUT /sim/admin/api/config` 响应 MUST 扩展：

- `scheduleReloaded`（bool）
- `effects[]`：`kind`（如 `scheduler_reloaded`、`task_interval_changed`、`ephemeral_may_continue`）、可选 `task`、`message`
- `taskSchedule[]`：每任务 `name`、`enabled`、`intervalSec`、`lastRunAt`（来自 `sim_task_run` 若存在）、`nextRunHint`（人类可读，如「约 3h 后」或「保存后立即进入新周期等待」）

系统 MUST NOT 保证强杀进行中的 LLM 调用或已 spawn 的 E1 聊天 goroutine；`effects` MUST 可表达「进行中任务可能跑完旧配置」。

#### Scenario: 关闭 chat 任务的效果提示

- **WHEN** PUT 将 `task_chat=false`
- **THEN** 响应 `effects` MUST 含 chat 相关提示且 `taskSchedule` 中 chat MUST 反映 disabled

#### Scenario: 修改 interval 的下一跑提示

- **WHEN** PUT 缩短 postImage 周期且该任务有 `lastRunAt`
- **THEN** `taskSchedule` 中 postImage 的 `nextRunHint` MUST 基于 `lastRunAt + 新 interval` 估算

### Requirement: docker compose env SHALL omit DB-backed sim runtime and LLM variables

`manifest/docker/docker-compose.microservices.yml` 与 `.env.example` MUST 移除已 DB 化的 sim 变量，至少包括：`SIM_LLM_*`、`SIM_TASK_*`、`SIM_INTERVAL_*`、`SIM_EPHEMERAL_*`、`SIM_STARTUP_STAGGER_MAX`、`SIM_UCG_RATE_LIMIT_*`、`SIM_VIDEO_POLL_ENABLED`。MUST 保留：`SIM_DB_LINK`、`SIM_USER_SERVICE_ENABLED`、`SIM_ADMIN_PASSWORD`、API Key 类变量。

#### Scenario: compose 无 SIM_LLM env

- **WHEN** 运维查看 microservices compose 中 sim-user-service environment
- **THEN** MUST NOT 含 `SIM_LLM_TEXT_PROVIDER` 等 LLM lane env 块

---

## sim-user-admin

<!-- source: openspec/specs/sim-user-admin/spec.md -->

# sim-user-admin Specification

## Purpose
TBD - created by archiving change sim-admin-task-ai-display. Update Purpose after archive.
## Requirements
### Requirement: sim admin API SHALL expose per-task AI model catalog

`GET /sim/admin/api/status` 与 `GET /sim/admin/api/runtime` 响应 MUST 含 `taskAiModels`（object）：键为调度任务名（至少含 `register`、`comment`、`post_image`、`post_video_submit`、`post_debate`、`debate_comment`、`chat_scan`、`follow`），值为数组。**MUST NOT** 含 `video_poll` 键。

数组元素 MUST 含 `laneKey`、`usage`（可选）、`provider`、`model`。`post_video_submit` MUST 含 simText 与 simVideoGen 相关条目。`post_debate` 与 `debate_comment` MUST 含 simText 条目。

#### Scenario: Status includes T7 T8 models

- **WHEN** GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_debate` 与 `taskAiModels.debate_comment` MUST 存在且含 simText

#### Scenario: Status returns AI for post video task

- **WHEN** 已鉴权 GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_video_submit` MUST 含 simText 与 simVideoGen 信息且 MUST NOT 存在 `video_poll` 键

### Requirement: sim admin UI SHALL display AI models per scheduled task

`sim-admin.html` 任务状态表 MUST 增加「AI 模型」列。每行 MUST 根据 `taskAiModels[taskName]` 展示：

- 用途说明（若有 `usage`）
- lane 键（如 `simText`）
- `provider/model`；未配置时 MUST 可读提示（如「未配置」）

无 LLM 的任务 MUST 显示「—」。页面 MUST 说明修改模型须至 ai-model-admin，本页 MUST NOT 提供 lane 编辑。

#### Scenario: Task table shows AI column

- **WHEN** 管理员打开 sim-admin 并加载状态
- **THEN** 表格 MUST 含「AI 模型」列且 T1 行展示 simText 与 simImageGen 信息

#### Scenario: Refresh reflects lane change

- **WHEN** 管理员在 ai-model-admin 修改 simText model 后点击 sim-admin「刷新状态」
- **THEN** 任务表中依赖 simText 的行 MUST 展示新 model

#### Scenario: No lane editor on status table

- **WHEN** 管理员查看任务状态表
- **THEN** UI MUST NOT 提供修改 provider/model 的输入控件

### Requirement: sim admin web SHALL be served from gateway-app

系统 MUST 在 `gateway-app-server` 注册静态页 `/device/admin/sim-admin.html`（`resource/public/sim-admin.html`）。页面 MUST 要求管理员已在运维 Hub 登录（与 ucg-admin 一致的鉴权模式）。

#### Scenario: Admin page reachable

- **WHEN** 管理员已登录 Hub 并访问 `/device/admin/sim-admin.html`
- **THEN** 浏览器 MUST 加载模拟管理界面

### Requirement: sim admin API SHALL expose config and prompts

`sim-user-service` MUST 提供 Admin HTTP API（经 gateway 反代或直连，鉴权与 ucg-admin 对齐）：

- `GET/PUT /sim/admin/api/config` — 字段至少含 `enabled`（bool）、`maxSimUsers`（int，默认 100）
- `GET/PUT /sim/admin/api/prompts/{taskType}` — `taskType` 至少含：`register_nickname`、`register_avatar`、`comment`、`post_image_text`、`post_video_text`、`post_debate_text`、`debate_comment`、`chat_reply`；每项含 `systemPrompt`、`userPromptTemplate`
- `GET /sim/admin/api/status` — 各任务上次运行时间、成功/失败计数、pending video job 数
- `GET /sim/admin/api/users` — 分页模拟用户列表（见 ADDED Requirement）
- `POST /sim/admin/api/users/{wxId}/deactivate` — 注销单个模拟用户（见 ADDED Requirement）

Prompt 变更 MUST 在下一任务 tick 生效，无需重启进程（读 DB + 短 TTL 缓存可接受）。

#### Scenario: Debate prompts editable

- **WHEN** GET `/sim/admin/api/prompts/debate_comment`
- **THEN** MUST 返回 userPromptTemplate 且保存后下一次 T8 MUST 使用新模板

#### Scenario: Update max sim users

- **WHEN** 管理员 PUT `maxSimUsers=50`
- **THEN** 下一次 T1 MUST 在 sim 用户数 ≥50 时停止注册

#### Scenario: Update comment prompt

- **WHEN** 管理员修改 `comment` 的 `userPromptTemplate`
- **THEN** 下一次 T2 MUST 使用新模板渲染变量

### Requirement: sim admin API SHALL expose read-only runtime snapshot

`sim-user-service` MUST 提供 `GET /sim/admin/api/runtime`（鉴权与现有 sim-admin API 一致）。响应 MUST 包含字段，反映**当前 DB 生效值**（非 env 只读镜像）：

- `serviceEnabled`（bool）— 对应 `SIM_USER_SERVICE_ENABLED`（env 硬闸）
- `dbEnabled`（bool）— 对应 `sim_config.enabled`
- `databaseName`（string）— 自 `SIM_DB_LINK` 解析的库名，MUST NOT 含账号密码或 host
- `simUserCount`（int）— 当前 `is_simulated=1` 用户数；拉取失败时可为 `-1`
- `maxSimUsers`（int）— 来自 `sim_config`
- `taskSwitches`（object）— 键至少含：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPoll`
- `intervals`（object）— 键至少含：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPollIdle`、`videoPollActive`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`；值为字符串时长
- `rateLimitRps`（number）、`rateLimitBurst`（int）

响应 MUST NOT 含 DSN 凭据、`SIM_ADMIN_PASSWORD`、`GLM_API_KEY` 或默认登录密码明文。

#### Scenario: Admin reads runtime from DB

- **WHEN** 已鉴权管理员 GET `/sim/admin/api/runtime` 且 DB 中 comment 周期为 3h
- **THEN** 响应 `intervals.comment` MUST 为 3h 且 MUST NOT 依赖 env 覆盖

#### Scenario: Runtime excludes secrets

- **WHEN** 管理员 GET `/sim/admin/api/runtime`
- **THEN** 响应 body MUST NOT 包含 `@tcp(` 连接串或 `password` 字段

### Requirement: sim admin UI SHALL show structured task status

`sim-admin.html` MUST 将 `GET /sim/admin/api/status` 结果以结构化形式展示（表格或等价列表），字段至少含：任务名、上次运行时间、成功次数、失败次数、最近错误；并单独展示 `pendingVideoJobs`。

原始 JSON dump  alone MUST NOT 作为唯一展示方式（可保留「查看原始 JSON」折叠为可选）。

#### Scenario: Task status table

- **WHEN** 管理员加载页面或点击刷新状态
- **THEN** 各 `status.tasks` 条目 MUST 以可读表格行展示而非仅 `JSON.stringify` 整块输出

### Requirement: sim admin UI SHALL edit runtime configuration with save effect feedback

`sim-admin.html` MUST 提供可编辑运行时配置表单，字段覆盖 `taskSwitches`、`intervals`、`rateLimitRps`、`rateLimitBurst`、`ephemeralChatLoop`、`ephemeralChatWindow`（及既有 `enabled`、`maxSimUsers`）。保存 MUST 调用扩展后的 **`PUT /sim/admin/api/config`**。保存成功后 MUST 展示 API 返回的 **`effects`** 与 **`taskSchedule`**（含各任务「立即生效 / 预计下一跑」提示）。页面 MUST 区分 `serviceEnabled`（env，说明须改 env 并 recreate）与 `dbEnabled`（可在线保存）。页面 MUST NOT 提供 sim LLM lane 编辑（链至 ai-model-admin）。`GET /sim/admin/api/status` 结构化任务状态展示 MUST 保留。

#### Scenario: 可编辑任务开关

- **WHEN** 管理员在 sim-admin 取消勾选 chat 并保存
- **THEN** 页面 MUST 调用 PUT config 且 MUST 展示保存结果面板含 schedule 提示

#### Scenario: 保存后展示下一跑提示

- **WHEN** PUT config 返回 `taskSchedule` 含 postImage 的 `nextRunHint`
- **THEN** sim-admin MUST 在保存结果区展示该提示

#### Scenario: Interval inputs are editable

- **WHEN** 管理员查看运行配置区块
- **THEN** UI MUST 提供周期间隔输入控件（非只读文本）

### Requirement: sim config PUT response taskSchedule SHALL reflect effective scheduling state

`PUT /sim/admin/api/config` 响应中的 `taskSchedule[]` 每一项 MUST 含：

- `configEnabled`（bool）— `runtime_json` 中对应任务开关（配置层）
- `enabled`（bool）— **自动调度实际是否启用**：MUST 等于 `configEnabled && dbEnabled && serviceEnabled`，其中 `dbEnabled` 为保存后的 `sim_config.enabled`，`serviceEnabled` 为当前进程 `SIM_USER_SERVICE_ENABLED`
- `nextRunHint`（string）— 当 `enabled=false` 时 MUST 说明阻塞原因（任务配置关 / 业务总闸关 / 进程总闸关），MUST NOT 在总闸关闭时给出「约 X 后」类下一跑时间

`name`、`label`、`intervalSec`、`lastRunAt` 语义不变。

#### Scenario: DB off and T4 config on shows not effectively enabled

- **WHEN** PUT 保存 `sim_config.enabled=false` 且 `taskPostVideo=true`
- **THEN** `taskSchedule` 中 `post_video_submit` MUST 含 `configEnabled=true` 且 `enabled=false`，且 `nextRunHint` MUST 表明业务总闸关闭

#### Scenario: Env off and task config on

- **WHEN** 进程 `SIM_USER_SERVICE_ENABLED=false` 且 PUT 保存某任务 `configEnabled=true` 与 `sim_config.enabled=true`
- **THEN** 该任务 `enabled=false` 且 `nextRunHint` MUST 表明进程总闸关闭

#### Scenario: All gates on and task on

- **WHEN** `serviceEnabled=true`、`dbEnabled=true`、任务 `configEnabled=true`
- **THEN** `enabled=true` 且 `nextRunHint` MAY 基于 `lastRunAt` 与周期估算下一跑

### Requirement: sim config PUT effects SHALL note scheduler blocked when gates closed

当保存后存在 `configEnabled=true` 的任务且（`dbEnabled=false` 或 `serviceEnabled=false`）时，`effects[]` MUST 含可读提示：任务开关已保存但自动调度未启动，并 MAY 说明可手动执行任务。

#### Scenario: Effects on DB disabled save

- **WHEN** PUT 将 `enabled=false` 且至少一任务配置开关为 true
- **THEN** `effects` MUST 含业务总闸相关提示

### Requirement: sim admin save result UI SHALL distinguish config vs effective task state

`sim-admin.html` 保存结果中的 `taskSchedule` 表格 MUST 区分 **配置开关**（`configEnabled`）与 **自动调度是否生效**（`enabled`）。当 `enabled=false` 时 UI MUST NOT 仅展示「开」而暗示任务已在运行。

#### Scenario: Save result shows effective off for T4

- **WHEN** 管理员保存 DB 关、T4 配置开
- **THEN** 保存结果表 MUST 显示 T4 配置为开、自动调度为关（或等价文案）

### Requirement: sim-admin SHALL expose editable runtime configuration aligned with DB schema

`sim-admin.html` MUST 提供可编辑运行时配置表单，字段覆盖 `taskSwitches`、`intervals`、`rateLimitRps`、`rateLimitBurst`（及既有 `enabled`、`maxSimUsers`）。**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow` 或 E1 相关文案。保存 MUST 调用 **`PUT /sim/admin/api/config`**。保存成功后 MUST 展示 API 返回的 **`effects`** 与 **`taskSchedule`**。页面 MUST 区分 `serviceEnabled`（env）与 `dbEnabled`（可在线保存）。`GET /sim/admin/api/status` 结构化任务状态 MUST 保留；`taskAiModels.chat_scan` usage MUST 为「未读回复」（非「E1 回复」）。

#### Scenario: No E1 fields in admin form

- **WHEN** 管理员打开 sim-admin 运行配置区
- **THEN** MUST NOT 展示 E1 循环/窗口输入框

#### Scenario: Save without ephemeral effects

- **WHEN** PUT 保存 task/interval 变更
- **THEN** 响应 `effects` MUST NOT 含 `ephemeral_may_continue`

#### Scenario: Chat scan task AI label

- **WHEN** GET status 或 runtime 含 `taskAiModels.chat_scan`
- **THEN** lane usage MUST 为「未读回复」

### Requirement: sim-admin runtime panel SHALL display effective task schedule

`GET /sim/admin/api/runtime` 与 status 中 `intervals` 键 MUST 含 `register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`；**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow`、`videoPollIdle`、`videoPollActive`。

`taskSchedule[]`（config PUT 响应与 admin UI）MUST 含 `post_debate`（T7 辩论发帖）与 `debate_comment`（T8 辩论论点）项，含 `configEnabled`、`enabled`、`intervalSec`、`nextRunHint`。

#### Scenario: Admin shows T7 T8 schedule

- **WHEN** GET config 或 PUT 保存后返回 `taskSchedule`
- **THEN** MUST 含 name=`post_debate` 与 name=`debate_comment`

#### Scenario: Runtime intervals without ephemeral

- **WHEN** GET runtime
- **THEN** `intervals` MUST NOT 含 E1 键

### Requirement: sim admin runtime intervals SHALL include T4 inline poll parameters

`GET /sim/admin/api/config` 与 `GET /sim/admin/api/runtime` 的 `intervals` MUST 含：

- `postVideoPollInterval` — T4 提交后 async-result 轮询间隔（字符串 duration）
- `postVideoPollMaxWait` — T4 视频发布最大等待（字符串 duration）

MUST NOT 再返回 `videoPollIdle`、`videoPollActive`。`taskSwitches` MUST NOT 含 `videoPoll`。

`PUT /sim/admin/api/config` MUST 接受并持久化上述字段。

#### Scenario: Config get includes poll intervals

- **WHEN** 管理员 GET `/sim/admin/api/config`
- **THEN** `intervals.postVideoPollInterval` 与 `intervals.postVideoPollMaxWait` MUST 为可读 duration 字符串

### Requirement: sim admin UI SHALL reflect T4 inline video flow

`sim-admin.html` MUST：

- 移除 P1 任务开关、P1 周期输入、P1 手动执行行及 runtime 只读区中的 P1 项
- 在 T4 相关区域提供可编辑 `postVideoPollInterval`、`postVideoPollMaxWait`（保存走 `PUT /sim/admin/api/config`）
- T4 手动「执行」按钮：自点击至 `sim_task_run.post_video_submit.lastRunAt` 更新前 MUST 显示「执行中…」且 disabled；**MUST NOT** 依赖新增 status 字段
- status 轮询对 `post_video_submit` MUST 足够长以覆盖 `postVideoPollMaxWait`（避免过早恢复按钮）

#### Scenario: T4 button stays busy until task run updates

- **WHEN** 管理员手动执行 T4 且视频轮询需 5 分钟完成
- **THEN** 5 分钟内按钮 MUST 保持「执行中…」直至 status 中 lastRunAt 变化

#### Scenario: P1 removed from admin form

- **WHEN** 管理员打开 sim-admin 配置区
- **THEN** MUST NOT 显示 P1 开关或 P1 idle/active 输入

### Requirement: sim-user-admin SHALL expose paginated simulated user list API

`sim-user-service` MUST 提供 `GET /sim/admin/api/users`（鉴权与现有 sim-admin 一致）。Query 参数 MUST 支持 `page`（默认 1）、`pageSize`（默认 20，最大 200）。

响应 MUST 含 `list`、`total`、`page`、`pageSize`。`list` 每项 MUST 含：

- `wxId`（int64）
- `account`（string）
- `nickname`（string，无 UCG profile 时为空字符串）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）
- `createdAt`（int64 Unix 秒；无 credential 记录时为 0）
- `passwordPlain`（string；无 credential 时 MUST 为 `123456` 且 `passwordPlainLegacy` MUST 为 true）
- `passwordPlainLegacy`（bool，可选；true 表示历史用户未持久化明文）

服务 MUST 经 device internal `sim/wx/list`、ucg internal `profiles/batch` 与 sim 库 `sim_wx_credential` 合并结果，MUST NOT 直查 device/ucg 库。

#### Scenario: Admin lists sim users with profile

- **WHEN** 已鉴权 GET `/sim/admin/api/users?page=1&pageSize=20` 且存在带 UCG profile 的 sim 用户
- **THEN** 响应 `list` MUST 含对应 `nickname` 与非空 `avatarUrl` 或 `avatarThumbnailUrl`（若 profile 有头像）

#### Scenario: Legacy user password fallback

- **WHEN** 列表项 wxId 在 `sim_wx_credential` 无记录（历史 `ptest*` 用户）
- **THEN** 该项 `passwordPlain` MUST 为 `123456` 且 `passwordPlainLegacy` MUST 为 true

#### Scenario: CreatedAt from credential

- **WHEN** T1 注册后写入 `sim_wx_credential`
- **THEN** 列表该项 `createdAt` MUST 等于 credential 的 `created_at`

### Requirement: sim-user-admin SHALL expose simulated user deactivate API

`sim-user-service` MUST 提供 `POST /sim/admin/api/users/{wxId}/deactivate`（鉴权与现有 sim-admin 一致）。路径参数 `wxId` MUST 为正整数。

成功时 MUST：调用 device internal sim deactivate 删除 `wx` 行；删除 sim 库 `sim_wx_credential` 对应行；将该 wxId 的 `sim_video_job` 中 `pending`/`processing` 行标为 `skipped`。

MUST NOT 调用 ucg 或 gateway App API 删除帖子/profile。注销语义 MUST 与 App `POST /device/app/api/user/deactivate` 一致（仅删 wx）。

#### Scenario: Deactivate simulated user success

- **WHEN** 已鉴权 POST `/sim/admin/api/users/1001/deactivate` 且 wxId=1001 为 `is_simulated=1`
- **THEN** HTTP MUST 成功且该 wx 行 MUST 自 device 库删除且 credential 行 MUST 删除

#### Scenario: Reject non-sim wxId

- **WHEN** wxId 存在但 `is_simulated=0`
- **THEN** MUST 返回 4xx 业务错误且 MUST NOT 删除 wx

#### Scenario: Reject invalid wxId

- **WHEN** wxId 不存在或已注销
- **THEN** MUST 返回明确业务错误（已注销或不存在）

### Requirement: sim-admin UI SHALL display simulated user table with deactivate

`sim-admin.html` MUST 在现有页面内嵌入「模拟用户」区块（MUST NOT 要求单独 Hub 模块页）。表格 MUST 展示：头像（优先 `avatarThumbnailUrl`）、UCG 昵称、账号、wxId、注册时间、密码、注销操作。

- 注册时间：`createdAt=0` 时 MUST 显示「—」
- 密码：`passwordPlainLegacy=true` 时 MUST 标注「默认密码（历史）」
- 注销：点击前 MUST `confirm`；成功后 MUST 刷新列表并更新 runtime 区 `simUserCount`

分页 MUST 调用 `GET /sim/admin/api/users`。页面 MUST NOT 提供 LLM lane 编辑（既有约束不变）。

#### Scenario: Admin sees user row

- **WHEN** 管理员已登录 Hub 并打开 sim-admin 且存在 sim 用户
- **THEN** 表格 MUST 展示至少一行含昵称/账号/wxId/密码列

#### Scenario: Deactivate refreshes list

- **WHEN** 管理员确认注销某一 sim 用户且 API 成功
- **THEN** 该行 MUST 从表格消失且 runtime 模拟用户数 MUST 减少

---

## sim-user-service

<!-- source: openspec/specs/sim-user-service/spec.md -->

# sim-user-service Specification

## Purpose
TBD - created by archiving change ucg-sim-user-service. Update Purpose after archive.
## Requirements
### Requirement: sim-user-service SHALL run as an independent process with a master enable switch

系统 MUST 提供独立进程 `sim-user-service`（`cmd/sim-user-service`）。环境变量 `SIM_USER_SERVICE_ENABLED` 为 `false` 时，进程 MAY 启动健康检查但 MUST NOT 启动任何周期 ticker 或视频轮询。为 `true` 时 MUST 注册全部已声明背景任务。

#### Scenario: Service disabled

- **WHEN** `SIM_USER_SERVICE_ENABLED=false`
- **THEN** 进程 MUST NOT 执行 T1–T6 与 P1 任务

#### Scenario: Service enabled

- **WHEN** `SIM_USER_SERVICE_ENABLED=true` 且依赖服务可达
- **THEN** 进程 MUST 按各任务独立开关启动对应 ticker

### Requirement: T1 register task SHALL create simulated users every 24 hours

每 24 小时（±10% jitter）MUST 执行注册任务（`SIM_TASK_REGISTER_ENABLED`）。当当前 sim 用户数 `< maxSimUsers` 时，MUST 生成**随机**账号与**随机**密码，调用 device internal sim 注册，经 `simText` 生成昵称、`simImageGen` 生成头像，完成 UCG profile 与 media 上传并更新 nickname/avatarKey；**仅在 profile 写入成功后** MUST 写入 `sim_wx_credential`。profile 完成前任一步失败 MUST 回滚 wx（见 ADDED Requirement）。当已达 `maxSimUsers` 时 MUST 跳过本次执行。

MUST NOT 再分配 `ptest{N}` 序号账号。`sim_account_seq` MUST NOT 参与 T1 账号生成。

#### Scenario: Register under cap with random credentials

- **WHEN** sim 用户数为 99 且 `maxSimUsers=100` 且 T1 全流程成功
- **THEN** 系统 MUST 注册新 sim 用户并标记 `is_simulated=1`，且 MUST 写入 credential

#### Scenario: Skip at cap

- **WHEN** sim 用户数已达 `maxSimUsers`
- **THEN** T1 MUST NOT 调用注册接口

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，且 **`excludeMediaTypes` MUST 含 `2`（视频）**，以获取单条非视频已发布帖（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`）。取得帖子后 MUST 经单次 `LaneSimVision` 生成评论并 `POST` 发表。T2 MUST NOT 对 `mediaType=2` 的帖子发表评论；若 sample 仍返回视频帖（例如 ucg 未升级），MUST skip 并记任务失败或 warning，且 MUST NOT POST 评论。

允许评论的帖子类型 MUST 为 `mediaType=0`（纯文字）或 `mediaType=1`（图文）。Green 审核 MUST 走正常 UCG 路径。

#### Scenario: Sample request excludes video

- **WHEN** 代码评审 `RunCommentTask` 的 sample 请求体
- **THEN** MUST 含 `"excludeMediaTypes": [2]`（或等价语义）

#### Scenario: No comment on video post

- **WHEN** sample 返回帖的 `mediaType` 为 `2`
- **THEN** T2 MUST NOT 调用 LLM 或 POST 评论

#### Scenario: Text and image posts still commented

- **WHEN** sample 返回 `mediaType=0` 或 `1` 且 LLM 可用
- **THEN** T2 MUST 按现有多模态/纯文本规则生成并发表评论

#### Scenario: No eligible post after exclude

- **WHEN** sample 因 exclude 返回空 `list`
- **THEN** T2 MUST 记录失败（如「无已发布帖」）且 MUST NOT 发表评论

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** MUST NOT 请求 `/ucg/app/api/feed/recommend`

### Requirement: T3 image post task SHALL publish image posts every 3.5 hours

每 3.5 小时（±10% jitter）MUST 随机 sim 用户经 `simText` 生成母婴文案、`simImageGen` 生成配图，完成 OSS media 链路后以 `submit=true` 创建图文动态。

#### Scenario: Image post submitted

- **WHEN** 生图与上传成功
- **THEN** 帖子 MUST 进入 `pending_audit` 并触发正常审核 MQ

### Requirement: T4 and P1 SHALL submit and poll video posts without retry

`RunPostVideoSubmitTask`（调度与手动共用）在 `SubmitVideoGeneration` 成功且 `InsertVideoJob` 写入 `pending` 后 MUST 启动视频结果轮询。轮询 MUST 调用智谱 `GET /paas/v4/async-result/{task_id}`（经 `aimodel.PollVideoGeneration`）。轮询 MUST 使用 submit 阶段已获得的 `loginSession`，MUST NOT 再经分页 list 线性查 account。

- **success**：下载视频 → UCG media 上传 → `POST /ucg/app/api/posts`（`submit=true`）→ `sim_video_job=done` → `RecordTaskRun("post_video_submit", true, ...)`
- **failed**（上游明确失败）：`sim_video_job=skipped` → `RecordTaskRun(..., false, ...)`
- **processing / pending**：在 `postVideoPollInterval` 后重试，直至 `now >= submitTime + postVideoPollMaxWait` → 超时视为发布失败 → `skipped` + `RecordTaskRun(..., false, ...)`

MUST NOT 在 submit 成功时单独写 `RecordTaskRun` success。

#### Scenario: Poll success posts video

- **WHEN** T4 提交后 async-result 返回 success 且上传发帖 OK
- **THEN** job MUST 为 `done` 且 `sim_task_run` MUST 记 post_video_submit 成功

#### Scenario: Poll timeout fails task

- **WHEN** 自 submit 起超过 `postVideoPollMaxWait` 仍为 processing
- **THEN** job MUST 为 `skipped` 且 `sim_task_run` MUST 记 post_video_submit 失败

### Requirement: T5 and E1 SHALL handle chat with real users

每 `intervals.chat`（默认 1h，± jitter）T5（`chat_scan`）MUST 执行：

1. `GET /device/internal/api/sim/wx/ids` 取得**全量** sim wxId 列表
2. `POST /ucg/internal/api/chat/sim-unread-sample`，请求体含完整 `simWxIds`
3. 若无 eligible 未读（`found=false`）→ MUST `RecordTaskRun(chat_scan, false, "无未读会话")` 并结束
4. 若有命中 → 结合 sample 返回的 `messages` 经 `simText`（`chat_reply` prompt）生成 **一条**回复，并 `POST /ucg/internal/api/chat/send`

每 tick MUST 最多回复 **一条**会话。MUST NOT spawn detached goroutine 或临时聊天窗口循环。MUST NOT 使用 App `GET /conversations` 扫描全量未读。peer MUST 为真人（由 ucg sample 的 `peer NOT IN simWxIds` 保证）。

#### Scenario: Single reply per tick

- **WHEN** T5 tick 且 ucg sample 返回一条 eligible 未读
- **THEN** 系统 MUST 发送恰好一条 chat 消息并 MUST NOT 启动后台聊天 goroutine

#### Scenario: Skip when no unread

- **WHEN** 全量 simWxIds 非空但 ucg sample 返回 `found=false`
- **THEN** MUST 记录失败或 skip 语义（非 success 空消息）且 MUST NOT 调用 LLM

#### Scenario: Real peer only via sample

- **WHEN** 未读会话对端 wxId 属于 simWxIds
- **THEN** ucg sample MUST NOT 返回该会话，T5 MUST NOT 回复

#### Scenario: Full sim coverage

- **WHEN** sim 用户总数超过 200
- **THEN** T5 MUST 仍通过 ids 接口覆盖全部 sim 用户，MUST NOT 仅使用前 200 个 id

### Requirement: T6 follow task SHALL follow sim to sim every 7 hours

每 7 小时（±10% jitter）MUST 随机选取一个 sim 用户 A，经 gateway 登录；调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，body **`excludeAuthorWxIds`** MUST 为 device internal 拉取的全量 sim wxId 列表；从返回帖取得 **`authorWxId`** 作为关注目标 B，并对 B 调用 `POST /ucg/app/api/follow/{wxId}`。`authorWxId` MUST NOT 等于 A 的 wxId（不等则重试抽样，有界次数后仍失败则记 task 失败）。MUST NOT 执行 sim→sim 关注；MUST NOT 使用 `pickTwoDistinctSimWx` 或两次 `sim/wx/random` 互关。已关注 MUST 幂等跳过。无 eligible 真人 author 时 MUST 记 task 失败（如「无真人作者」），MUST NOT 假 success。

#### Scenario: Sim follows real author

- **WHEN** 存在 published 帖且 author 非 sim，A 未关注 B
- **THEN** `POST /ucg/app/api/follow/{B}` MUST 成功或幂等

#### Scenario: No sim to sim follow

- **WHEN** T6 tick 执行
- **THEN** MUST NOT 选取两个 sim 用户互相关注

#### Scenario: No real author available

- **WHEN** 所有 published 帖作者均在 `excludeAuthorWxIds` 中
- **THEN** MUST 记 follow task 失败且 MUST NOT POST follow

### Requirement: sim-user-service MUST use HTTP contracts only

sim-user-service MUST NOT import 或查询 device/ucg 域 DAO。跨服务数据 MUST 经 gateway App API、device internal API、ucg internal API 与 `aimodel` 完成。

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审检索 `internal/dao` import
- **THEN** sim-user-service 包路径下 MUST NOT 出现 device/ucg 业务表 DAO

### Requirement: AI queue full SHALL abort current task tick

当 `aimodel.Acquire` 返回队列满或调用超时时，当前任务 tick MUST 提前结束并记录日志，MUST NOT 阻塞其他任务 goroutine。

#### Scenario: Queue full

- **WHEN** `ErrQueueFull` 在评论任务中发生
- **THEN** 该次 T2 执行 MUST 结束且不重试至下一周期

### Requirement: sim task intervals SHALL be overridable via environment variables

各背景任务周期 MUST 支持环境变量覆盖，未设置或非法值时 MUST 回退下列默认值：`SIM_INTERVAL_REGISTER=24h`、`SIM_INTERVAL_COMMENT=6h`、`SIM_INTERVAL_POST_IMAGE=3h30m`、`SIM_INTERVAL_POST_VIDEO=6h30m`、`SIM_INTERVAL_POST_DEBATE=12h`、`SIM_INTERVAL_DEBATE_COMMENT=1h`、`SIM_INTERVAL_CHAT=1h`、`SIM_INTERVAL_FOLLOW=7h`。周期执行 MUST 保留 ±10% jitter。

#### Scenario: Default intervals preserved

- **WHEN** 未设置任何 `SIM_INTERVAL_*` 环境变量
- **THEN** T1–T8 名义周期 MUST 与上表一致

#### Scenario: Custom debate comment interval

- **WHEN** `SIM_INTERVAL_DEBATE_COMMENT=30m` 且服务已启动
- **THEN** T8 两次成功执行间隔 MUST 约为 30m（含 jitter）

### Requirement: P1 video poll SHALL use adaptive idle and active intervals

P1 MUST 根据是否存在 `pending`/`processing` 的 `sim_video_job` 选择下一等待间隔：`SIM_INTERVAL_VIDEO_POLL_IDLE`（默认 `10m`）与 `SIM_INTERVAL_VIDEO_POLL_ACTIVE`（默认 `2m`）。无 pending job 时 MUST NOT 调用智谱轮询或 UCG 发帖。

#### Scenario: Idle backoff

- **WHEN** `sim_video_job` 无 pending/processing 行
- **THEN** 下一次 P1 唤醒间隔 MUST 使用 idle 间隔（默认 10m）

#### Scenario: Active polling

- **WHEN** 存在至少一条 pending/processing job
- **THEN** 下一次 P1 唤醒间隔 MUST 使用 active 间隔（默认 2m）且 MUST 执行智谱 `PollVideoGeneration`

### Requirement: sim outbound App HTTP SHALL be globally rate limited

经 gateway-app 的 `appGet`/`appPost`/`appPut` MUST 经全局限速器；默认 `SIM_UCG_RATE_LIMIT_RPS=2`（token bucket，burst 默认 4）。限速 MUST 在发起 HTTP 前阻塞等待，MUST NOT 静默丢弃请求。

#### Scenario: Burst within limit

- **WHEN** 2 秒内发起不超过 4 次 App API 调用且 RPS=2
- **THEN** 调用 MUST 成功发出（不因限速失败）

#### Scenario: Over limit waits

- **WHEN** 持续超过 RPS 限额发起调用
- **THEN** 超额调用 MUST 等待至许可可用后再发送

### Requirement: scheduler SHALL stagger first task ticks after startup

`SIM_USER_SERVICE_ENABLED=true` 启动 scheduler 时，各任务 goroutine 在首次执行前 MUST 额外等待 `0` 至 `SIM_STARTUP_STAGGER_MAX`（默认 `30m`）的均匀随机延迟，以避免多任务同时首次齐射。

#### Scenario: Staggered startup

- **WHEN** 服务启动且启用全部任务
- **THEN** 各任务首次 tick 时间 MUST NOT 全部相同（在 stagger 窗口内随机分布）

### Requirement: sim chat temperature constants SHALL be documented for redeploy tuning

任务 temperature 常量 MUST 附中文注释说明业务用途；变更常量 MUST 通过 redeploy `sim-user-service` 生效。首期 MUST NOT 依赖 env 或 sim-admin 配置项。

#### Scenario: Constants live in one file

- **WHEN** 代码评审 sim temperature 定义
- **THEN** 全部任务常量 MUST 位于同一 Go 源文件（如 `task_llm_temp.go`）

### Requirement: sim tasks SHALL pick random simulated user via device random API

各调度任务（T2 评论、T3 图文、T4 视频、E1 聊天、T6 关注等）当需要随机模拟用户时，sim-user-service MUST 经 device internal **`GET /device/internal/api/sim/wx/random`** 取得单条 `{wxId, account}` 并完成 gateway 登录；MUST NOT 调用 `sim/wx/list` 拉取分页列表后在内存随机；MUST NOT 对同一选取流程重复请求 list。T6 关注 MUST 选取两个不同 wxId（在仅 1 个 sim 用户时 MUST 失败「sim 用户不足」）。

#### Scenario: Comment task uses random pick once

- **WHEN** `RunCommentTask` 需要随机 sim 用户
- **THEN** MUST 仅一次 random 调用取得 account 并登录，MUST NOT 先 list 再 random index

#### Scenario: Follow picks two distinct users

- **WHEN** `RunFollowTask` 且 simulated 用户数 ≥ 2
- **THEN** MUST 经 random（或等价 device 有界选取）得到两个不同 wxId 并完成关注

#### Scenario: No sim users

- **WHEN** random 返回无用户
- **THEN** 任务 MUST 失败且错误语义与现网「无模拟用户」一致

#### Scenario: Count still uses list total

- **WHEN** T1 注册判断 `maxSimUsers` 上限
- **THEN** MAY 继续使用 list `pageSize=1` 读取 `total`，MUST NOT 为此拉取 pageSize=200 全量

### Requirement: sim T4 video upload SHALL use ucg internal transcode API

sim-user-service 在上传 T4 视频至 UCG 时 MUST 经 HTTP 调用 ucg-service `POST /ucg/internal/api/media/upload-video`（与 device internal 鉴权方式一致），MUST 使用返回的 `objectKey` 与 `contentHash` 调用 App API `RegisterMedia`（`mediaKind=2`、`transformVersion=v2`、`dedupHit=false`）。

MUST NOT 再使用 `transformVersion=sim-raw` 或 presign 直传未验真/未转码字节。

#### Scenario: T4 registers v2 after internal upload

- **WHEN** internal 转码上传成功
- **THEN** 后续 register MUST 使用 `transformVersion=v2` 且 MUST 通过 ucg 侧 v2 验真

#### Scenario: Internal transcode failure skips job

- **WHEN** internal 转码上传失败
- **THEN** T4 流水线 MUST 将 job 标为 skipped 且 MUST NOT 使用 presign 回退直传

### Requirement: T4 video pipeline SHALL be globally single-flight

任意时刻 sim-user-service 进程内 MUST 最多一条进行中的 T4 视频流水线（submit + poll + post）。`videoPostInFlight`（或等价机制）为 true 时：

- 调度 tick MUST skip 新 submit（MAY 记 success + 说明「video poll in progress」）
- 手动 `POST .../tasks/post_video_submit/run` MUST 拒绝或返回「任务正在执行中」（与 `manualBusy` 语义一致）

流水线结束（成功/失败/超时）后 MUST 清除 inFlight。

#### Scenario: Scheduler skips while poll active

- **WHEN** 上一 T4 流水线仍在轮询且调度 tick 到达
- **THEN** MUST NOT 再次 SubmitVideoGeneration

#### Scenario: Manual run rejected while poll active

- **WHEN** 管理员在流水线进行中再次点击 T4 手动执行
- **THEN** API MUST 返回任务忙错误

### Requirement: sim-user-service SHALL discard pending video jobs on startup

进程 scheduler 启动前（或等效启动钩子）MUST 将 `sim_video_job` 中 `status IN ('pending','processing')` 更新为 `skipped`。MUST NOT 为遗留 job 恢复轮询 goroutine。

#### Scenario: Startup clears stale jobs

- **WHEN** sim-user-service 重启且 DB 存在 pending job
- **THEN** 启动后这些 job MUST 为 skipped 且无恢复轮询

### Requirement: sim-user-service SHALL persist simulated user credentials for admin and task login

`sim-user-service` MUST 在 `SIM_DB_LINK` 库维护表 `sim_wx_credential`（或语义等价名），字段至少含：

- `wx_id`（BIGINT PRIMARY KEY）
- `account`（VARCHAR，非空）
- `password_plain`（VARCHAR，非空）
- `created_at`（BIGINT Unix 秒，非空）

T1 MUST 在 **UCG profile PUT 成功之后** INSERT credential（`created_at` 为 profile 写入完成时刻）。MUST NOT 在 profile 完成前写入 credential。`EnsureSchema` MUST 幂等创建该表。

注销 sim 用户时 MUST DELETE 对应 `wx_id` 行。MUST NOT 在响应日志中打印 `password_plain`。

#### Scenario: Credential written after profile success

- **WHEN** T1 profile PUT 成功完成
- **THEN** `sim_wx_credential` MUST 含对应 wx_id 与非空 `password_plain`

#### Scenario: Credential not written on profile failure

- **WHEN** T1 在 profile PUT 之前失败并已回滚 wx
- **THEN** MUST NOT 存在该 wx_id 的 credential 行

#### Scenario: Credential removed on admin deactivate

- **WHEN** admin deactivate 成功删除 wxId=2001
- **THEN** credential 行 MUST 不存在

### Requirement: sim task random user login SHALL use per-wxId credential password

当 T2、T3、T4、T5、T6 或手动任务需经 gateway `username_login` 登录某一 sim 用户时，sim-user-service MUST 按该用户 `wxId` 从 `sim_wx_credential` 读取 `password_plain` 作为登录密码。

若 credential 无记录（历史用户），MUST fallback 到运行时 `defaultPassword`（yaml/env `simUser.defaultPassword`，空则 `123456`）。

MUST NOT 假定所有 sim 用户共用同一注册密码（除上述历史 fallback）。

#### Scenario: Task login uses stored password

- **WHEN** T2 随机选中 wxId=2001 且 credential 存在
- **THEN** `username_login` MUST 使用 credential 中的 `password_plain`

#### Scenario: Legacy sim login fallback

- **WHEN** 随机选中历史 wxId 无 credential 行
- **THEN** login MUST 使用 defaultPassword fallback

### Requirement: T1 register SHALL rollback sim wx on profile setup failure

T1（`RunRegisterTask`，含 sim-admin 手动触发）在 device internal sim 注册成功获得 `wxId` 后，若 **未完成** UCG profile 写入（`PUT /ucg/app/api/profile/me` 含 nickname 与 avatarKey），则 MUST 视为注册失败。

下列任一步失败时 MUST 调用 device internal `POST /device/internal/api/sim/wx/{wxId}/deactivate` 回滚 wx（须 `is_simulated=1`），并 MUST 删除 sim 库 `sim_wx_credential` 中对应行（若存在）：

- `usernameLogin` 失败
- Prompt 加载失败
- 昵称 `simText` 调用失败，或生成 nickname trim 后为空
- 头像 `simImageGen` 调用失败，或返回 URL 为空
- 头像 upload / media 链路失败
- profile `PUT` 失败

回滚后 MUST `RecordTaskRun(register, false, ...)`，且该 wx MUST NOT 计入 sim 用户数。回滚 device 调用失败时 MUST 记录 warning 日志，任务仍记失败。

手动 T1 与 scheduler T1 MUST 共用上述语义。

#### Scenario: Nickname AI failure rolls back wx

- **WHEN** T1 已 simRegister 成功但 `simText` 昵称生成返回错误
- **THEN** 系统 MUST 注销该 wxId 且 MUST NOT 保留 `sim_wx_credential` 行

#### Scenario: Avatar AI failure rolls back wx

- **WHEN** T1 昵称生成成功但 `simImageGen` 失败
- **THEN** 系统 MUST 注销该 wxId

#### Scenario: Empty nickname rolls back wx

- **WHEN** T1 昵称 AI 返回仅空白字符
- **THEN** 系统 MUST 视为失败并 MUST 注销该 wxId

#### Scenario: Profile PUT success commits registration

- **WHEN** T1 profile PUT 成功且 nickname 非空、avatarKey 非空
- **THEN** 系统 MUST 写入 `sim_wx_credential` 且 MUST NOT 回滚 wx

#### Scenario: Failed registration does not consume cap slot

- **WHEN** T1 因头像 AI 失败回滚 wx 且当前 sim 用户数为 99、`maxSimUsers=100`
- **THEN** 下一次 T1 MUST 仍可尝试注册新用户

### Requirement: T7 post_debate task SHALL publish debate posts every 12 hours

每 12 小时（±10% jitter，可由 `SIM_INTERVAL_POST_DEBATE` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，经 `LaneSimText`（prompt `post_debate_text`）生成 JSON：`content`（话题正文）、`debateLeft`、`debateRight`（各 ≤5 字，与 UCG `validateDebateLabel` 一致），并 `POST /ucg/app/api/posts`，body MUST 含 `content`、`debateLeft`、`debateRight`、`mediaType=0`、`submit=true`，且 MUST NOT 含 `media`。标签校验失败或 LLM 失败 MUST 记 `post_debate` task 失败。

#### Scenario: Debate post submitted for audit

- **WHEN** LLM 输出合法且 POST 成功
- **THEN** 帖子 MUST 为 `type=debate`（或等价双填标签）且 MUST 进入 `pending_audit`

#### Scenario: Invalid debate labels fail task

- **WHEN** LLM 输出任一方立场标签超过 5 字或为空
- **THEN** MUST 记 `post_debate` 失败且 MUST NOT POST

#### Scenario: Default interval twelve hours

- **WHEN** 未设置 `SIM_INTERVAL_POST_DEBATE` 且 DB 无覆盖
- **THEN** T7 名义周期 MUST 为 12h（含 jitter）

### Requirement: T8 debate_comment task SHALL vote and post argument every 1 hour

每 1 小时（±10% jitter，可由 `SIM_INTERVAL_DEBATE_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，**`onlyDebate` MUST 为 `true`**，取得单条辩论帖（含 `debateLeft`、`debateRight`）。若 `authorWxId` 等于当前 sim 用户 wxId，MUST 记失败（如「不可评自己的帖」）且 MUST NOT 投票或评论。否则 MUST 经 `LaneSimText`（prompt `debate_comment`）生成 JSON：`side`（`left` 或 `right`）、`argument`（utf8 长度 ≤10），MUST 先 `POST /ucg/app/api/posts/{id}/vote` 再 `POST /ucg/app/api/posts/{id}/comments`。argument 超长 MUST 截断至 10 字或记失败（实现 MUST 在 POST 前保证 ≤10 字）。无 eligible 辩论帖时 MUST 记失败（如「无已发布辩论帖」）。

#### Scenario: Vote before comment

- **WHEN** T8 tick 且 sample 命中他人辩论帖
- **THEN** MUST 先 POST vote 成功再 POST comment

#### Scenario: Skip own debate post

- **WHEN** sample 返回帖 `authorWxId` 等于当前 sim wxId
- **THEN** MUST NOT POST vote 或 comment

#### Scenario: Argument length bound

- **WHEN** LLM 输出 argument 超过 10 字
- **THEN** 实现 MUST 截断至 10 字或记失败，且 POST 的 comment MUST NOT 超过 10 字

#### Scenario: Default interval one hour

- **WHEN** 未设置 `SIM_INTERVAL_DEBATE_COMMENT` 且 DB 无覆盖
- **THEN** T8 名义周期 MUST 为 1h（含 jitter）

#### Scenario: No debate posts available

- **WHEN** `onlyDebate=true` 且 sample 返回空 `list`
- **THEN** MUST 记 `debate_comment` 失败且 MUST NOT 调用 LLM

### Requirement: sim task intervals SHALL include T7 and T8 defaults

`SIM_INTERVAL_POST_DEBATE` 默认 MUST 为 `12h`；`SIM_INTERVAL_DEBATE_COMMENT` 默认 MUST 为 `1h`。对应 env 开关 `SIM_TASK_POST_DEBATE_ENABLED`、`SIM_TASK_DEBATE_COMMENT_ENABLED` 迁移期默认 SHOULD 为 true（与 DB runtime_json 合并后生效）。

#### Scenario: T7 T8 env defaults

- **WHEN** 未设置 T7/T8 interval env
- **THEN** LoadRuntimeFlags MUST 回退 12h 与 1h

---

## single-default-db-per-service

<!-- source: openspec/specs/single-default-db-per-service/spec.md -->

# single-default-db-per-service Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Service DAO access SHALL use only default database
Service processes MUST access database via their own `database.default` connection and MUST NOT rely on multi-group routing fallback logic.

#### Scenario: Domain DB group resolver removed
- **WHEN** checking DAO infrastructure files
- **THEN** `internal/dao/domain_db.go` multi-group resolver MUST be removed

#### Scenario: Service reads only local default connection
- **WHEN** a service executes DAO operations
- **THEN** the resolved DB connection MUST come from the service-local `database.default` config

---

## tip-contract-slim

<!-- source: openspec/specs/tip-contract-slim/spec.md -->

# tip-contract-slim Specification

## Purpose
TBD - created by archiving change tip-derive-baby-age. Update Purpose after archive.
## Requirements
### Requirement: Go App Tip API 请求体去掉月龄与当前时间

`POST /device/tip/generate` 的请求体（`DeviceTipGenerateReq` 或接线期等价结构）SHALL NOT 再包含 `babyAgeMonths`、`currentTime` 字段。相关 dc/注释 MUST NOT 再声称「0 则服务端按生日推算」或「0 则使用服务端当前时间」。

#### Scenario: 精简后的 App 请求体

- **WHEN** Flutter 或其它 App 客户端调用 `POST /device/tip/generate`
- **THEN** 请求 JSON SHALL 可仅含设备与事件相关字段（如 `deviceNo`、`eventId`、`eventName`）及既有其它非月龄/时间必填项
- **AND** 服务端 SHALL NOT 要求 `babyAgeMonths` 或 `currentTime`

#### Scenario: 假注释消除

- **WHEN** 评审 `api/v1/device_tip_http.go`（及控制器/服务注释）
- **THEN** SHALL NOT 存在「月龄为 0 由服务端推算」且实际未实现的误导描述

### Requirement: Go TipStream 与 PythonAIClient 去掉对应参数

`VoiceService.TipStream` 与 `PythonAIClient.TipStreamRequest` SHALL 删除 `babyAgeMonths`/`BabyAgeMonths`、`currentTime`/`CurrentTime`（及 JSON `baby_age_months`、`current_time`）。Go MUST NOT 再为 tip 补全 `currentTime` 后传给 Python。

#### Scenario: TipStream 签名瘦身

- **WHEN** 控制器调用 `TipStream` 发起小贴士流
- **THEN** 调用签名 SHALL 不再接收月龄与当前时间参数
- **AND** 发往 Python 的 JSON body SHALL 不含 `baby_age_months`、`current_time`

### Requirement: Python TipRequest 删除或废弃入站月龄与时间字段

`TipRequest` SHALL 删除 `baby_age_months`、`current_time` 字段，或将其标为废弃且非必填并忽略；权威月龄与时间上下文 MUST 来自服务端派生（见 `tip-python-derived-context`）。

#### Scenario: Tip 入站无月龄时间字段可成功

- **WHEN** Go 以 snake_case 发送 tip 请求且不含 `baby_age_months`、`current_time`
- **THEN** Python SHALL 接受该请求（其它必填合法时）
- **AND** SHALL 使用派生月龄与派生当前时间生成提示词

### Requirement: Flutter tip 调用去掉月龄与 currentTime

Flutter `tip_repository` / `tip_provider` / `home_screen` SHALL 不再计算或传递 `babyAgeMonths`，且请求体 SHALL NOT 包含 `currentTime`。

#### Scenario: home 触发 tip 不再传月龄

- **WHEN** `home_screen` 因历史 create 事件触发 tip 流式生成
- **THEN** 调用链 SHALL NOT 再传入 `babyAgeMonths`
- **AND** POST body SHALL NOT 包含 `babyAgeMonths` 或 `currentTime`

---

## tip-flutter-sse-go-dialect

<!-- source: openspec/specs/tip-flutter-sse-go-dialect/spec.md -->

# tip-flutter-sse-go-dialect Specification

## Purpose
TBD - created by archiving change wire-tip-generate-on-voice. Update Purpose after archive.
## Requirements
### Requirement: tip 请求路径为 /device/tip/generate

Flutter tip 客户端 SHALL 向网关基址下的 `/device/tip/generate` 发起 POST SSE 请求（不得使用缺少 `/device` 前缀的 `/tip/generate`）。

#### Scenario: URL 含 device tip 前缀

- **WHEN** `TipRepository.streamTip` 构建请求 URL
- **THEN** 路径 MUST 以 `/device/tip/generate` 结尾（相对 `apiBaseUrl`）

### Requirement: tip SSE 解析对齐 Go 方言

Flutter tip SSE 解析 SHALL 读取 `event:` 确定事件类型，`data:` 为纯文本内容；MUST 支持 thinking/answer/done/error，并在流结束识别 `data: [DONE]`。

#### Scenario: thinking/answer 增量

- **WHEN** 收到 `event: thinking`（或 `answer`）及随后 `data: <text>`
- **THEN** 客户端 MUST 将 data 纯文本作为该类型增量内容累积

#### Scenario: done 解析 answerId

- **WHEN** 收到 `event: done` 且 data 为含 `answerId` 的 JSON
- **THEN** 客户端 MUST 解析并保存 `answerId` 字符串，供后续反馈字段对齐使用

### Requirement: 状态模型使用 answerId

tip models / provider SHALL 使用 `String? answerId` 标识本次生成结果，MUST NOT 依赖从 SSE 流写入错误的 `tipId int`。

#### Scenario: 流结束后持有 answerId

- **WHEN** SSE done 事件成功解析 `answerId`
- **THEN** `TipContent`（或等价状态）MUST 更新为该 `answerId`，反馈按钮可用性可依据其非空判断

#### Scenario: 反馈 submit 字段对齐（可不完整飞轮）

- **WHEN** 调用方提交 tip 反馈
- **THEN** 请求体 SHOULD 使用 `answerId` 字符串字段对齐；完整 UI/后端反馈闭环属包 C，本包不要求反馈接口一定可用

---

## tip-generate-voice-host

<!-- source: openspec/specs/tip-generate-voice-host/spec.md -->

# tip-generate-voice-host Specification

## Purpose
TBD - created by archiving change wire-tip-generate-on-voice. Update Purpose after archive.
## Requirements
### Requirement: TipCtrl 绑定到 voice-service

系统 SHALL 在 voice-service HTTP 注册路径中绑定 `NewTipCtrl()`（或等价），使 `POST /device/tip/generate` 由 voice 进程处理。

#### Scenario: voice 进程可路由 tip generate

- **WHEN** voice-service 启动并完成 `RegisterVoiceServiceHTTP`
- **THEN** `POST /device/tip/generate` MUST 由 `TipCtrl.Generate` 处理，且调用 `voice.TipStream` 写出 SSE

#### Scenario: 不绑定 feedback 控制器

- **WHEN** 本变更完成 voice 路由注册
- **THEN** 系统 MUST NOT Bind clinic tip feedback 相关控制器（留给包 C）

### Requirement: gateway 反代 /device/tip 到 voice

gateway-app SHALL 将 `/device/tip/*` 反向代理到 VOICE_API_PROXY（与现有 voice 域反代同一配置族）。

#### Scenario: App 请求经网关到达 voice

- **WHEN** 客户端向 gateway 发起 `POST /device/tip/generate`
- **THEN** 请求 MUST 被代理到 voice-service 对应路径，而非 history/device 本域处理

### Requirement: tip generate 需登录且计入 usage

`POST /device/tip/generate` SHALL 要求 Bearer 登录，且 MUST 计入 App API usage 统计。

#### Scenario: 不进入鉴权豁免

- **WHEN** 维护 `gateway_app_auth_exempt` 列表
- **THEN** `/device/tip/generate` MUST NOT 出现在 Bearer 豁免路径中

#### Scenario: 不进入 usage 维护排除

- **WHEN** 维护 `usagestats/maintenance_skip`
- **THEN** `/device/tip/generate` MUST NOT 被写入排除列表（即保持统计）

### Requirement: SSE 方言保持 Go 风格

`TipCtrl.Generate` SHALL 继续以 `event:` + data 纯文本写出 thinking/answer/done，done 的 data MUST 含 `answerId` 字段；最终 MUST 发送 `data: [DONE]`。

#### Scenario: 流式帧格式

- **WHEN** tip 流成功推送 thinking/answer/done
- **THEN** 每帧 MUST 含 `event: <name>` 与对应 `data:` 行，且 done 的 data 可解析出 `answerId`

---

## tip-python-derived-context

<!-- source: openspec/specs/tip-python-derived-context/spec.md -->

# tip-python-derived-context Specification

## Purpose
TBD - created by archiving change tip-derive-baby-age. Update Purpose after archive.
## Requirements
### Requirement: Tip 月龄由 Python 根据生日自算

在 tip 生成图中，系统 SHALL 在 `fetch_baby_profile`（或等价画像拉取）完成之后，根据宝宝 `birthday` 与当前日期（时区 `Asia/Shanghai`）计算月龄，供提示词及依赖月龄的检索使用。调用方（Go/Flutter）MUST NOT 被要求提供 `baby_age_months` / `babyAgeMonths`。

#### Scenario: 有生日时写入计算出的月龄

- **WHEN** tip 图成功获取到可解析的 `birthday`
- **THEN** 系统 SHALL 按日历月差计算非负整数月龄（未来生日钳为 0）
- **AND** 提示词 SHALL 包含该月龄（例如「宝宝月龄：{n} 个月」）

#### Scenario: 无生日时提示「未知」而非 0 个月

- **WHEN** 画像缺失、无 `birthday` 或生日无法解析
- **THEN** 系统 SHALL NOT 将月龄表示为 `0` 个月以表示「未知」
- **AND** 提示词 SHALL 使用「宝宝月龄：未知」（或等价明确未知文案）
- **AND** 若知识检索依赖月龄，SHALL 跳过月龄过滤或使用不限月龄策略，MUST NOT 用 0 冒充新生儿月龄过滤

### Requirement: Tip 当前时间由 Python 在写提示词时生成

系统 SHALL 在构建 tip 提示词时生成当前时间上下文，时区 MUST 为 `Asia/Shanghai`。Tip 入站请求 MUST NOT 将 `current_time` / `currentTime` 作为必填字段。

#### Scenario: 提示词含上海时区当前时间

- **WHEN** tip 流式生成进入写提示词阶段
- **THEN** 提示词 SHALL 包含基于 `Asia/Shanghai` 的当前时间（可读本地时间；MAY 同时附带 Unix 秒）
- **AND** 该时间 SHALL 来自 Python 运行时时钟，而非请求体字段

#### Scenario: 请求体不再要求 current_time

- **WHEN** 调用方 POST tip stream 且 body **不含** `current_time` / `currentTime`
- **THEN** 系统 SHALL 仍能完成 tip 生成（在其它必填字段合法时）
- **AND** SHALL NOT 因缺少该字段返回 422

---

## tip-streaming

<!-- source: openspec/specs/tip-streaming/spec.md -->

# tip-streaming Specification

## Purpose
TBD - created by archiving change tip-chat-streaming. Update Purpose after archive.
## Requirements
### Requirement: Tip streaming generate API
Go 服务端 SHALL 暴露 `POST /device/tip/generate` 接口，以 SSE 方式流式返回小贴士生成的思考过程与最终建议。

#### Scenario: Flutter client requests tip via SSE
- **WHEN** Flutter 客户端向 `/device/tip/generate` 发送 POST 请求（body 包含 deviceNo、eventId、eventName；月龄与当前时间由服务端派生，不再要求 babyAgeMonths / currentTime）
- **THEN** Go 服务端 SHALL 立即返回 SSE 响应头（Content-Type: text/event-stream、Cache-Control: no-cache）
- **AND** 按序推送 `thinking`、`answer`、`done` 三类 SSE 事件
- **AND** 最终以 `data: [DONE]` 结束流式响应

### Requirement: Tip thinking events streamed
Go 服务端 SHALL 将 Python AI 返回的思考过程增量实时推送给客户端，不得等全部完成后一次性返回。

#### Scenario: Thinking delta pushed to client
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `thinking` 类型的 SSE 事件
- **THEN** Go 服务端 SHALL 立即将该 thinking delta 以 SSE `event: thinking` 帧转发给 Flutter 客户端
- **AND** 不得缓冲或延迟超过 100ms

### Requirement: Tip answer events streamed
Go 服务端 SHALL 将 Python AI 返回的建议内容增量实时推送给客户端。

#### Scenario: Answer delta pushed to client
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `answer` 类型的 SSE 事件
- **THEN** Go 服务端 SHALL 立即将该 answer delta 以 SSE `event: answer` 帧转发给 Flutter 客户端

### Requirement: Tip done event with answer_id
Go 服务端 SHALL 在流式结束时返回 answer_id，供后续反馈使用。

#### Scenario: Done event includes answer_id
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `done` 事件（含 answer_id）
- **THEN** Go 服务端 SHALL 向客户端推送 SSE `event: done` 帧，body 含 answer_id
- **AND** 紧接着推送 `data: [DONE]` 结束整个 SSE 连接

### Requirement: Tip API error handling
当 Python AI 服务不可用时，Go 服务端 SHALL 返回固定错误提示给客户端。

#### Scenario: Python AI unavailable during tip generate
- **WHEN** Go 服务端调用 Python `/v1/tip/stream` 失败（连接超时、5xx 错误等）
- **THEN** Go 服务端 SHALL 以 SSE `event: error` 帧向客户端推送错误消息："AI服务暂时不可用，请稍后再试"
- **AND** 紧接着推送 `data: [DONE]` 结束连接

---

## typed-domain-enums

<!-- source: openspec/specs/typed-domain-enums/spec.md -->

# typed-domain-enums Specification

## Purpose
定义核心领域值的类型化枚举契约，消除关键路径裸字符串匹配并保持协议兼容。

## Requirements
### Requirement: 核心领域值类型化枚举
系统 SHALL 为 `target_type`、`mode`、状态机状态、`event type` 提供类型化枚举定义，并通过统一常量与解析函数替代散落裸字符串匹配。

#### Scenario: 调用层使用枚举分支
- **WHEN** 业务代码需要按 `target_type` 或 `mode` 分支处理
- **THEN** 代码 MUST 使用枚举类型与常量进行判断，而不是直接比较裸字符串

#### Scenario: 非法值解析失败
- **WHEN** 输入字符串无法映射为合法枚举值
- **THEN** 系统 MUST 返回明确错误并记录可观测日志

### Requirement: 枚举与字符串双向兼容
系统 SHALL 提供枚举到字符串、字符串到枚举的双向转换能力，保证现有 DB 与消息协议字符串格式兼容。

#### Scenario: 入站字符串转换为枚举
- **WHEN** 系统从 DB/MQ/HTTP 读取字符串字段
- **THEN** 系统 MUST 通过统一 Parse 方法转换为枚举值后参与业务判断

#### Scenario: 出站枚举保持原协议字符串
- **WHEN** 系统写入 DB 或发布消息
- **THEN** 系统 MUST 输出与历史协议兼容的字符串值

---

## ucg-admin-post-moderation

<!-- source: openspec/specs/ucg-admin-post-moderation/spec.md -->

# ucg-admin-post-moderation Specification

## Purpose
TBD - created by archiving change ucg-admin-post-batch-reject. Update Purpose after archive.
## Requirements
### Requirement: UCG admin SHALL authenticate post moderation APIs with X-Admin-Password

`ucg-service` Admin 动态审查接口 MUST 与现有 UCG Admin 共用认证：请求 Header `X-Admin-Password` MUST 等于配置项 `ucg.admin.password`；校验失败 MUST 返回未授权，且 SHALL NOT 返回帖子数据或执行驳回。

#### Scenario: 口令正确访问列表

- **WHEN** 管理员携带正确 `X-Admin-Password` 请求 `GET /ucg/admin/api/posts/list`
- **THEN** 系统 SHALL 返回分页帖子列表

#### Scenario: 口令错误拒绝

- **WHEN** 管理员携带错误或未携带 `X-Admin-Password` 请求任一动态审查 Admin API
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改 `ucg_post`

### Requirement: Admin SHALL list all posts with optional status filter

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（0/1/2/3）。省略 `status` 时 SHALL 返回全部状态的帖子。列表 MUST 按 `updated_at` 降序排序。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段。`media` 数组 MUST 包含该帖 **全部** 媒体项（按 `sortOrder`）；每项 SHALL 含 `cdnUrl`、`mediaKind`（1=图片，2=视频），图片 SHALL 含物理缩略图 `thumbnailUrl`（`BuildImageThumbnailURL`，path 为 `{stem}_thumb.{ext}`，MUST NOT 含 `x-oss-process`），视频 SHALL 含物理首帧缩略图 `thumbnailUrl`（`BuildVideoThumbnailURL`，path 为 `{stem}_thumb.jpg`，MUST NOT 含 `x-oss-process`），SHALL NOT 仅返回无 thumbnail 的 mp4 `cdnUrl` 供列表 `<img>` 直接使用。

#### Scenario: 按状态筛选待审帖

- **WHEN** 管理员请求 `status=1`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 1（pending_audit）

#### Scenario: 分页默认值

- **WHEN** 管理员未传 `page` 与 `pageSize`
- **THEN** 系统 SHALL 使用 `page=1`、`pageSize=20` 并返回对应分页元数据

#### Scenario: 视频帖返回首帧 thumbnail

- **WHEN** 列表项首条媒体 `mediaKind=2` 且 `objectKey` 有效且 OSS 存在对应 `_thumb.jpg`
- **THEN** 该项 `thumbnailUrl` SHALL 非空且 SHALL 可用于 `<img>` 展示；`cdnUrl` SHALL 为可播放的视频 URL；`thumbnailUrl` SHALL NOT 含 `x-oss-process`

#### Scenario: 多图帖返回全量 media

- **WHEN** 帖子关联多条 `ucg_post_media`
- **THEN** 列表项 `media` 数组长度 SHALL 等于关联条数且顺序与 `sortOrder` 一致

#### Scenario: 视频 thumbnail 为物理 jpg path

- **WHEN** 视频 objectKey 为 `social/2026/06/a.mp4` 且 thumb 已 materialize
- **THEN** 该项 `thumbnailUrl` SHALL 指向 `.../a_thumb.jpg` 且 SHALL NOT 包含 `x-oss-process`

### Requirement: Admin SHALL batch reject posts with shared reject semantics

`POST /ucg/admin/api/posts/reject` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）、**`reason`（必填，trim 后非空）**。`reason` 为空或仅空白时 MUST 返回参数错误且 SHALL NOT 修改任何帖子。对每条目标帖子，若当前 `status` 已为 3（rejected）SHALL 计入 `skipped` 且不更新行；若 `status` 为 0、1、2、4 或 5，系统 MUST 将 `status` 置为 3、写入 `reject_reason`（管理员提供的 reason）、更新 `updated_at` 为当前 unix 秒，并计入 `rejected`。若原帖为 `status=2`（published），MUST 从推荐/Feed 索引移除。DB 错误计入 `failed`。响应 MUST 包含 `rejected`、`skipped`、`failed` 三个 id 数组。本操作 SHALL NOT 向作者发送通知或站内信。

#### Scenario: 批量驳回已发布帖

- **WHEN** 管理员提交 `postIds` 含 `status=2` 的帖子、非空 `reason` 且口令正确
- **THEN** 对应行 `status` SHALL 变为 3 且 `reject_reason` SHALL 等于提交的 reason；该帖 SHALL NOT 出现在推荐或关注 Feed

#### Scenario: 驳回缺少 reason 拒绝

- **WHEN** 管理员提交 reject 且 `reason` 为空或仅空白
- **THEN** API SHALL 返回参数错误且 MUST NOT 修改 `ucg_post`

#### Scenario: 已驳回帖幂等跳过

- **WHEN** 管理员对 `status=3` 的帖子再次提交驳回
- **THEN** 该行 SHALL NOT 被修改且该 id SHALL 出现在 `skipped`

#### Scenario: 作者可见驳回原因无推送

- **WHEN** 帖子被管理端驳回且 reason 为「含不当用语」
- **THEN** 作者请求「我的动态」SHALL 可见该帖 `status=3` 与 `rejectReason=含不当用语`；系统 SHALL NOT 因本次驳回创建通知或 WS 推送

### Requirement: ucg-admin.html SHALL provide post moderation tab with batch reject UI

静态页 `resource/public/ucg-admin.html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox；SHALL 提供「全选本页可驳回项」、**「批量通过」**与「批量驳回」按钮。批量驳回前 MUST 弹出理由输入且 MUST NOT 在理由为空时提交；批量通过前 MUST 经用户确认。**批量驳回 MUST 在请求 body 中携带非空 `reason`。** `status=3` 的行 checkbox SHALL 禁用或不可选。操作成功后 SHALL 刷新当前列表。表格 SHALL 含 **驳回原因** 列（展示 `rejectReason`）。状态筛选 SHALL 含 0–5（含「发布失败(4)」「机审失败(5)」）。表格「媒体」列 SHALL 展示每条动态 **全量** 媒体缩略图；图片 SHALL 支持 modal 原图；视频 SHALL 展示首帧缩略图并支持 modal 播放。

动态审查 Tab 内 **工具栏行**（状态筛选、刷新、批量通过、批量驳回、已选提示）SHALL 使用 flex 布局且 **`align-items: center`**，使 label、select、button、hint 文本在同一行内纵向居中对齐；SHALL NOT 因全局 `.row { align-items: flex-start }` 导致工具栏元素顶对齐错位。样式 SHOULD  scoped 至 `#panelPosts`（或等价 class），MUST NOT 改变其它 Tab 的 `.row` 布局。

同一页面 MUST 提供与「动态审查」并列的 **「资料机审失败」** Tab。

#### Scenario: 动态审查批量驳回须填理由

- **WHEN** 管理员勾选帖子并点击批量驳回且在 prompt 中输入非空理由
- **THEN** 系统 SHALL 调用 reject API 且 body 含该 reason，并刷新列表

#### Scenario: 动态审查批量通过

- **WHEN** 管理员勾选含待审/机审失败/发布失败帖并确认批量通过
- **THEN** 系统 SHALL 调用 approve API 且刷新列表

#### Scenario: 驳回理由为空不提交

- **WHEN** 管理员在驳回 prompt 留空或取消
- **THEN** 页面 SHALL NOT 调用 reject API

#### Scenario: 管理页含资料机审 Tab

- **WHEN** 管理员打开 ucg-admin.html 并已通过口令登录
- **THEN** 页面 SHALL 展示「资料机审失败」Tab 入口

#### Scenario: 动态审查工具栏纵向居中

- **WHEN** 管理员打开「动态审查」Tab 且工具栏含状态 select 与批量操作按钮
- **THEN** 工具栏行内各控件 SHALL 纵向居中对齐（视觉同一水平中线），且 AI 配置等其它 Tab 的 `.row` 布局 SHALL 保持不变

### Requirement: device admin entry SHALL link to UCG management

`resource/public/admin.html` 中指向 `/device/admin/ucg-admin.html` 的入口链接文案 SHALL 为「UCG 管理」（或等价中文），以涵盖 AI 配置与动态审查。

#### Scenario: 设备管理页入口文案

- **WHEN** 管理员打开设备管理页
- **THEN** UCG 入口链接可见文案 SHALL 为「UCG 管理」

### Requirement: Admin SHALL batch approve posts for human publish

`POST /ucg/admin/api/posts/approve` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）。对每条目标帖子：

- 若 `status=2`（published）SHALL 计入 `skipped` 且不更新行。
- 若 `status=1`（pending_audit）、`4`（apply_failed）或 `5`（moderation_failed），系统 MUST 在不调用 Green 的前提下将帖子发布：`status` 置为 `2`，写入 `published_at` 与 `updated_at`，清空面向作者的 `reject_reason`（及 apply 失败相关字段）；`status=5` 时 MUST 写入 `moderation_verdict=pass` 与 `moderation_at`；`status=1` 且 `moderation_verdict=0` 时 MUST 写入 `moderation_verdict=pass`。成功后 MUST 调用与 MQ publish 等价的 Feed/Redis 同步（`syncPublishedPostRedis`）。计入 `approved`。
- 若 `status=0`（draft）或 `3`（rejected），SHALL 计入 `failed` 且不更新行。
- DB/CAS 错误计入 `failed`。

响应 MUST 包含 `approved`、`skipped`、`failed` 三个 id 数组。本操作 SHALL NOT 向作者发送通知或站内信。

#### Scenario: 批量通过待审帖

- **WHEN** 管理员提交含 `status=1` 的 `postIds` 且口令正确
- **THEN** 对应行 `status` SHALL 变为 `2` 且 `published_at` SHALL 非空；该帖 SHALL 可出现在推荐/关注 Feed

#### Scenario: 批量通过机审失败帖

- **WHEN** 管理员提交含 `status=5` 的 postId
- **THEN** 行 SHALL 变为 `published` 且 `moderation_verdict` SHALL 为 pass；`reject_reason` SHALL 清空

#### Scenario: 已发布帖幂等跳过

- **WHEN** 管理员对 `status=2` 的帖子提交 approve
- **THEN** 行 SHALL NOT 被修改且 id SHALL 出现在 `skipped`

#### Scenario: 已驳回帖不可批准

- **WHEN** 管理员对 `status=3` 的帖子提交 approve
- **THEN** 行 SHALL NOT 被修改且 id SHALL 出现在 `failed`

---

## ucg-admin-profile-moderation

<!-- source: openspec/specs/ucg-admin-profile-moderation/spec.md -->

# ucg-admin-profile-moderation Specification

## Purpose
TBD - created by archiving change ucg-audit-green-once-admin. Update Purpose after archive.
## Requirements
### Requirement: UCG Admin SHALL list profile audit jobs in moderation_failed state

ucg-service MUST 提供 admin HTTP API（Header `X-Admin-Password` 与现有 UCG admin 一致）：

- `GET /ucg/admin/api/profile-audit-jobs/list` — 分页查询 `ucg_profile_audit_job`；默认筛选 `status=ProfileJobStatusModerationFailed(5)`；可选 query 覆盖 status/page/pageSize。

响应 MUST 含：jobId、wxId、auditVersion、nickname/avatarKey/bio（job patch 字段）、rejectReason（机审失败日志）、createdAt、updatedAt。

#### Scenario: 列表默认仅机审失败

- **WHEN** 管理员 GET list 且未传 status
- **THEN** 响应 SHALL 仅包含 `status=5` 的 job 行

#### Scenario: 未授权拒绝

- **WHEN** 请求缺少或错误 `X-Admin-Password`
- **THEN** API SHALL 返回未授权错误且 MUST NOT 返回 job 数据

### Requirement: UCG Admin SHALL manually resolve moderation_failed profile jobs

`POST /ucg/admin/api/profile-audit-jobs/resolve` MUST 接受 `jobId`（必填）、`action`（`approve` | `reject`）、`reason`（reject 时必填）。

- **approve**：CAS 将 job 从 `moderation_failed(5)` 转为 `approved(2)`；MUST 按 job 非空 patch 字段更新 `ucg_profile`（与 `approveProfileJobCAS` 语义一致）；SHOULD 补写 `moderation_verdict=pass` 与 `moderation_at`。
- **reject**：CAS 将 job 转为 `rejected(3)` 并写入 `reject_reason`；MUST NOT 更新已发布 profile。

操作 MUST 使用 CAS（`status=5` + 匹配 `audit_version`），并发重复请求 MUST 幂等（0 行 affected 返回成功或明确 skip）。

#### Scenario: 人工通过应用 patch

- **WHEN** 管理员对 status=5 的 job 执行 approve 且 job.bio 非空
- **THEN** job SHALL 变为 approved，且 `ucg_profile.bio` SHALL 更新为 job 中 bio

#### Scenario: 人工驳回

- **WHEN** 管理员对 status=5 的 job 执行 reject 并填写 reason
- **THEN** job SHALL 变为 rejected 且 reject_reason SHALL 等于所填 reason，且已发布 profile MUST 不变

#### Scenario: 非 moderation_failed 拒绝

- **WHEN** 管理员对 status≠5 的 job 调用 resolve
- **THEN** API SHALL 返回参数/状态错误且 MUST NOT 变更 job 或 profile

### Requirement: ucg-admin.html SHALL provide profile moderation_failed review UI

静态页 MUST 在「资料机审失败」Tab 内：

- 调用 list API 展示表格（wxId、patch 摘要、失败原因、时间）；
- 每行提供「通过」「驳回」按钮；驳回 MUST 弹窗收集 reason；
- 操作成功后 MUST 刷新当前页列表。

#### Scenario: 通过后列表减少

- **WHEN** 管理员点击某行「通过」且 API 成功
- **THEN** 该行 MUST 从当前列表消失（或刷新后不再出现）

---

## ucg-ai-quota

<!-- source: openspec/specs/ucg-ai-quota/spec.md -->

# ucg-ai-quota Specification

## Purpose
TBD - created by archiving change refactor-ai-quota-domain-ownership. Update Purpose after archive.
## Requirements
### Requirement: ucg-service SHALL be authoritative for polish quota configuration and usage

`ucg-service` MUST 在 **`ai_voice_ucg`** 库（GoFrame `database.default`，连接 `UCG_DB_LINK`）维护 **`polish`** feature 的 AI 月度额度全局默认与 per-wxId override。全局默认 MUST 包含 `polishMonthlyLimit`（初始 **5**）；Admin 可修改。per-wxId override MAY 覆盖 polish；未 override MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 **`ai:usage:polish:{wxId}:{YYYYMM}`**，`YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。ucg-service MUST NOT 将 polish 配额配置或用量写入 device 或 voice 库表；MUST NOT 再转发 device internal ai-quota API。

#### Scenario: 全局润笔默认配置

- **WHEN** 管理员将全局 `polishMonthlyLimit` 设为 10
- **THEN** 无 override 的用户润笔上限 SHALL 为 10

#### Scenario: 单人 override 覆盖润笔

- **WHEN** wxId=1001 的 override 为 `polishMonthlyLimit=20`
- **THEN** 该用户润笔上限 SHALL 为 20

### Requirement: ucg polish SHALL pre-check and consume quota locally

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（`LanePolish`）前于 **ucg-service 进程内**执行 polish check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用上游。check 通过后 MUST 经 polish lane 闸门；队列满 MUST 返回 **50301**。上游成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求上游

#### Scenario: 上游失败不扣减

- **WHEN** check 通过但上游返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: 队列满

- **WHEN** check 通过但 polish 闸门队列满
- **THEN** API SHALL 返回 50301 且 MUST NOT consume

### Requirement: ucg App quota read API SHALL expose polish only

`GET /ucg/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ polish: { used, limit, degraded } }`，对应当月上海时区桶；**`degraded` MUST 为 `used >= limit`**。本接口 MUST NOT 返回 `voiceAi` 或 `clinicAi` 字段。

#### Scenario: 登录用户查询润笔额度

- **WHEN** wxId=1001 请求 `/ucg/app/api/ai-quota` 且当月润笔已用 2、上限 5
- **THEN** `polish.used` SHALL 为 2、`polish.limit` SHALL 为 5 且 `polish.degraded` SHALL 为 false

#### Scenario: 额度用尽 degraded 标记

- **WHEN** wxId=1001 当月润笔已用 5、上限 5
- **THEN** `polish.degraded` SHALL 为 true 且 `polish.used` SHALL 为 5

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: ucg admin SHALL configure polish quota locally only

ucg-service MUST 提供 `GET/PUT /ucg/admin/api/ai-quota/default` 与 `GET/PUT /ucg/admin/api/ai-quota/user`（query/body 含 `wxId`），认证 MUST 为 Header `X-Admin-Password` 等于 `ucg.admin.password`。ucg-service MUST **本地**持久化 polish 配置至 `ai_voice_ucg`，MUST NOT 转发 device。PUT default MUST 仅接受 `polishMonthlyLimit`（正整数）。PUT user MUST 仅接受 optional `polishMonthlyLimit`；空值 SHALL 表示清除 override。Admin API MUST NOT 接受或返回 `voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`。

#### Scenario: 管理员修改全局润笔默认

- **WHEN** 管理员 PUT default 为 polish=10
- **THEN** ucg 权威配置 SHALL 更新且新用户 check SHALL 使用 limit=10

#### Scenario: ucg admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

### Requirement: polish 额度与 VIP 共同策略

ucg 润笔（`polish` feature）MUST 使用与 voice 相同的 VIP∪额度共同策略：VIP 时 Allowed=true 且成功不计次；非 VIP 额度内使用 `polish` lane 正式模型并在成功后 consume；非 VIP 额尽时 MUST 使用该 lane 的 free 配置，free 为空时 MUST NOT 使用硬编码 `DegradedPolishProfile` 作为真相源（按 design 以无覆盖/默认上游语义处理，且 MUST 可观测）。

#### Scenario: VIP 润笔不计次

- **WHEN** VIP 用户润笔成功返回正文
- **THEN** `polish` 月用量 MUST NOT 增加，响应 MUST NOT 再以额度用尽为由失败

#### Scenario: 非 VIP 额尽走 free

- **WHEN** 非 VIP 且 polish 额度用尽，且 Admin 配置了 polish free 模型
- **THEN** 润笔 MUST 使用 free 模型完成（成功时可不计次），MUST NOT 返回 40302 阻断

---

## ucg-aliyun-secrets-env

<!-- source: openspec/specs/ucg-aliyun-secrets-env/spec.md -->

# ucg-aliyun-secrets-env Specification

## Purpose
TBD - created by archiving change ucg-aliyun-secrets-env. Update Purpose after archive.
## Requirements
### Requirement: ucg-service 阿里云凭证 MUST 经环境变量注入且 yaml 不得含明文

`ucg-service` 运行时使用的阿里云 OSS AccessKey ID/Secret 与 DashScope API Key MUST 来自容器环境变量，不得依赖 `manifest/config/config.ucg-service.yaml` 中的明文值。仓库内该 yaml 的 `ucg.oss.accessKeyId`、`ucg.oss.accessKeySecret`、`ucg.ai.dashscope_api_key` MUST 留空字符串。Green 内容审核 MUST 复用 OSS 凭证（`LoadGreenConfig` fallback），MUST NOT 要求独立的 Green AccessKey 环境变量。

环境变量名 MUST 为：

| 用途 | 环境变量 |
|------|----------|
| OSS AccessKey ID | `UCG_OSS_ACCESS_KEY_ID` |
| OSS AccessKey Secret | `UCG_OSS_ACCESS_KEY_SECRET` |
| DashScope API Key | `UCG_DASHSCOPE_API_KEY` |

#### Scenario: Compose 启动 ucg-service 时注入 OSS 凭证

- **WHEN** 运维使用 `docker compose --env-file manifest/docker/.env.prod` 启动含 `ucg-service` 的栈，且 `.env.prod` 含 `UCG_OSS_ACCESS_KEY_ID` 与 `UCG_OSS_ACCESS_KEY_SECRET`
- **THEN** ucg-service 容器环境 MUST 可见上述变量且 presign 接口 SHALL 可成功返回 uploadUrl

#### Scenario: yaml 无明文时仍可通过 env 润笔

- **WHEN** `config.ucg-service.yaml` 中 `dashscope_api_key` 为空且容器 env 设置 `UCG_DASHSCOPE_API_KEY`
- **THEN** AI 润笔运行时配置 MUST 使用该 key 且 SHALL NOT 因 yaml 为空而单独失败

#### Scenario: Green 复用 OSS env 凭证

- **WHEN** `ucg.green.enabled` 为 true 且 Green yaml 中 accessKey 为空，容器 env 已设置 `UCG_OSS_ACCESS_KEY_*`
- **THEN** Green 客户端 MUST 使用与 OSS 相同的 AccessKey 发起审核

### Requirement: Docker Compose 基线 MUST pass-through ucg 阿里云 env

`manifest/docker/docker-compose.microservices.yml` 的 `ucg-service.environment` MUST 包含 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 的 `${VAR:-}` 引用，使 local/test/prod overlay 共享同一注入点。

#### Scenario: 基线 compose 引用 env 文件变量

- **WHEN** 开发者查看 `docker-compose.microservices.yml` 中 `ucg-service` 段
- **THEN** MUST 可见上述三个 environment 条目且格式与其它 secrets pass-through 一致

### Requirement: 部署 env 文件 MUST 含真实凭证且 example 仅含占位符

`manifest/docker/.env.local`、`.env.test`、`.env.prod` MUST 包含真实值的 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY`（与本变更实施时 yaml 中既有凭证一致，由实施者填入）。仓库 MUST 提供 `manifest/docker/.env.example`，列出相同 key 名与注释说明，MUST NOT 含真实密钥。

#### Scenario: example 文件可供新环境复制

- **WHEN** 新成员复制 `.env.example` 为本地 `.env` 并填入凭证
- **THEN** key 名 MUST 与 compose pass-through 及 Go 代码读取的 env 名完全一致

### Requirement: runbook MUST 文档化 ucg 阿里云 env 约定

`docs/runbooks/release-deploy-and-run.md` MUST 说明 ucg-service 部署时必需的三项阿里云相关 env、与 yaml 留空的关系，并指向 `.env.example`。

#### Scenario: runbook 检索 env 名

- **WHEN** 运维查阅 release runbook 中 ucg-service 或 secrets 相关章节
- **THEN** MUST 能找到 `UCG_OSS_ACCESS_KEY_ID`、`UCG_OSS_ACCESS_KEY_SECRET`、`UCG_DASHSCOPE_API_KEY` 的说明

---

## ucg-app-http-api

<!-- source: openspec/specs/ucg-app-http-api/spec.md -->

# ucg-app-http-api Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: App HTTP API SHALL expose UCG REST under /ucg/app/api

ucg-service SHALL implement REST endpoints (also reachable via gateway proxy) including: profile get/update, feed recommend/following, posts CRUD, media presign, follow/unfollow, likes, comments, conversations and messages list。

**推荐 Feed**（`GET /ucg/app/api/feed/recommend`）MUST 使用 **cursor 分页**（见 ADDED Requirement），**不适用** `{ total, page, pageSize }` 契约。

除推荐 Feed 与评论列表外，其他列表分页 MUST 仍使用 `page`（从 1 开始）与 `pageSize`（默认 20，最大 50），响应 MUST 包含 `{ list, total, page, pageSize }`。

评论列表 `GET /ucg/app/api/posts/{id}/comments` SHALL **不适用**上述分页契约：SHALL 单次返回该帖评论全量列表。`page`/`pageSize` 查询参数 SHALL 被忽略或废弃，MUST NOT 再驱动服务端 `OFFSET` 分页。

#### Scenario: 推荐 Feed cursor 分页

- **WHEN** `GET /ucg/app/api/feed/recommend?pageSize=20&lat=31.2&lng=121.5`
- **THEN** 响应 SHALL 仅含 `status=2` 的帖子，且 SHALL 含 `list`、`hasMore`，MAY 含 `nextCursor`；MUST NOT 含 `total`

#### Scenario: 我的动态含全状态

- **WHEN** 作者 `GET /ucg/app/api/posts/mine`
- **THEN** 响应 SHALL 含 draft/pending/rejected/published 本人帖子

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误

### Requirement: Profile API SHALL auto-create default nickname from baby name

On first access, ucg-service SHALL create `ucg_profile` with nickname `{babyName}的家长` fetched via device internal API when profile missing.

#### Scenario: 首次进入 UCG
- **WHEN** 已登录用户首次请求 `/profile/me` 且无 profile 行
- **THEN** 服务 SHALL 创建 profile 且 nickname SHALL 使用 device 返回的 baby_name 拼接

### Requirement: 评论列表 SHALL 单次返回升序全量并批量填充作者 profile

`GET /ucg/app/api/posts/{id}/comments` MUST 仅对已发布（`status=2`）帖子返回评论。评论 MUST 按 `created_at` **升序**排序（最早在前、最新在列表底部）。响应 MUST 为 JSON：

- `list`：评论数组，每项 SHALL 含 `id`、`postId`、`authorWxId`、`content`、`createdAt`，且 SHOULD 含 `author`（公开 profile，与帖子作者 profile 同为实时 `ucg_profile` 语义，MUST NOT 依赖评论表快照列）
- `total`：SHALL 优先为帖子 `comment_count`；若不可用 SHALL 回退为 `len(list)`
- `truncated`：布尔；当评论数超过配置上限且仅返回部分行时为 `true`，否则为 `false`

服务 MUST NOT 对评论列表执行独立 `COUNT(*)` 查询。服务 MUST 通过单次 `ucg_profile` 批量查询（如 `wx_id IN (...)`）填充列表内所有 `author`，MUST NOT 在列表循环中逐条查询 profile（N+1）。评论列表读路径 MUST NOT 使用 Redis 缓存。

当配置 `ucg.comments.listMax`（默认 500，0 表示不限制）大于 0 且 `comment_count` 超过该值时，SHALL 仅返回按 `created_at ASC` 的前 `listMax` 条，并设 `truncated=true`；`total` SHALL 仍为完整 `comment_count`。

#### Scenario: 常规模帖单次拉取全量评论

- **WHEN** 客户端 `GET /ucg/app/api/posts/{id}/comments` 且该帖已发布、评论数不超过 `listMax`
- **THEN** 响应 SHALL 含全部评论且 `list` 按 `created_at` 升序
- **AND** `truncated` SHALL 为 `false`
- **AND** `total` SHALL 等于帖子 `comment_count`
- **AND** 每条评论的 `author` SHALL 经批量 profile 查询填充

#### Scenario: 超长帖评论截断

- **WHEN** 帖子 `comment_count` 为 600 且 `listMax=500`
- **THEN** 响应 `list` SHALL 含 500 条（最早 500 条）
- **AND** `truncated` SHALL 为 `true`
- **AND** `total` SHALL 为 600

#### Scenario: 发表评论响应可供乐观追加

- **WHEN** 客户端 `POST /ucg/app/api/posts/{id}/comments` 成功
- **THEN** 响应 SHALL 含完整评论字段及 `author`
- **AND** 客户端 SHALL 可将该条追加至本地 `list` 末尾而无需再次 GET 全列表

#### Scenario: 评论列表不使用 Redis

- **WHEN** ucg-service 处理评论列表读请求
- **THEN** 系统 MUST NOT 读写 Redis 中的评论列表或 profile 缓存键

### Requirement: 评论列表 GET SHALL 不计入 App API 使用统计

负责人已确认：`GET /ucg/app/api/posts/{id}/comments` MUST NOT 计入 gateway-app App API 使用统计。gateway-app SHALL 在 `maintenance_skip.go` 排除该 GET 路径（原始 `req.URL.Path` 为 `/ucg/app/api/posts/<postId>/comments`；归一化 apiKey 为 `GET /ucg/app/api/posts/{id}/comments`）。`POST /ucg/app/api/posts/{id}/comments`（发表评论）SHALL 仍计入统计。

#### Scenario: 成功 GET 评论列表不写入 usage

- **WHEN** 客户端成功调用 `GET /ucg/app/api/posts/{id}/comments` 经 gateway-app
- **THEN** gateway-app MUST NOT 将该请求计入 App API 使用统计

#### Scenario: 发表评论仍计入 usage

- **WHEN** 客户端成功调用 `POST /ucg/app/api/posts/{id}/comments` 经 gateway-app
- **THEN** gateway-app SHALL 将该请求计入 App API 使用统计

### Requirement: Comments API SHALL reflect audit pending and rejection

`POST .../posts/{id}/comments` 在 Green 完成前 MUST 返回 `status=1`（或等价 pending 字段）。`GET .../posts/{id}/comments` MUST 仅返回 `status=2`（published）评论给 **非作者** 视角；作者 MAY 看到自身 pending/rejected 评论及 reject_reason。

#### Scenario: 他人看不到待审评论

- **WHEN** 用户 A 发表评论且未审过，用户 B 拉取评论列表
- **THEN** 列表 MUST NOT 包含 A 的该条评论

#### Scenario: 作者见违规评论

- **WHEN** 用户 A 评论 Green fail
- **THEN** A 拉取评论或评论详情 MUST 可见 reject_reason

### Requirement: Profile me API SHALL read pending patch from audit job

`GET /ucg/app/api/profile/me` 对作者 MUST 合并 MySQL 最新 pending `ucg_profile_audit_job`（`auditPending=true`、预览 nickname/avatar/bio），MUST NOT 依赖 Redis profile pending 键作为长期权威。待审预览与 MQ/CAS 的版本语义 MUST 以 job 表 `audit_version` 为准，MUST NOT 从 `ucg_profile` 或 Redis 读取审核轮次。

#### Scenario: 待审头像预览

- **WHEN** 用户提交新 avatar 且 job pending
- **THEN** profile/me MUST 返回 `auditPending=true` 且 avatar 预览为新 key

### Requirement: Post and feed media DTOs SHALL expose physical video thumbnailUrl

App HTTP API 返回的帖子/动态媒体项（含推荐 Feed、关注 Feed、帖子详情、我的动态等经 `loadPostMedia` 或等价路径组装的 `media` 数组）中，当 `mediaKind=2`（视频）且 `objectKey` 非空时，MUST 填充 `thumbnailUrl` 为 `BuildVideoThumbnailURL(objectKey)` 物理首帧 jpg CDN URL。该 URL MUST NOT 含 `x-oss-process` query。`cdnUrl` MUST 仍为可播放的 mp4 CDN URL。

本要求 MUST NOT 改变既有 v1/v2 响应 JSON 字段名或结构，仅变更 `thumbnailUrl` 的 URL 形态。

#### Scenario: Feed 中视频帖含物理 thumbnailUrl

- **WHEN** 客户端 `GET /ucg/app/api/feed/recommend` 且列表含已发布视频帖（OSS 已存在 `_thumb.jpg`）
- **THEN** 对应 `media` 项 `thumbnailUrl` SHALL 以 `_thumb.jpg` 结尾且 SHALL NOT 含 `x-oss-process`

#### Scenario: 视频 media 仍保留可播放 cdnUrl

- **WHEN** 帖子含视频媒体
- **THEN** 该项 `cdnUrl` SHALL 指向 mp4 原片且 `thumbnailUrl` SHALL 与 `cdnUrl` path 不同

### Requirement: v2 创建帖 API SHALL 支持可选坐标

系统 MUST 提供 `POST /ucg/app/api/v2/posts`，请求体 MUST 兼容 v1 创建帖字段并 MAY 含可选 `lat`、`lng`（WGS84 十进制度）。成功响应 MUST 与 v1 创建帖相同结构（`UcgPostItem`）。该 endpoint MUST 计入 gateway-app App API 使用统计。

#### Scenario: v2 创建带坐标

- **WHEN** 客户端 `POST /ucg/app/api/v2/posts` 含合法 body 与 `lat`/`lng`
- **THEN** 服务 MUST 持久化坐标（MySQL 或等价）且在 publish 后 GEOADD 与 snapshot 含坐标

#### Scenario: v2 创建无坐标

- **WHEN** 客户端 `POST /ucg/app/api/v2/posts` 不含 lat/lng
- **THEN** 服务 MUST 成功创建且帖 MUST NOT 进入 GEO 索引

#### Scenario: v2 创建计入 usage

- **WHEN** 经 gateway-app 的 `POST /ucg/app/api/v2/posts` 返回 HTTP 2xx
- **THEN** gateway-app MUST 将该请求计入 App API 使用统计

### Requirement: 帖子读写 API SHALL 可选接受 viewer 坐标并返回 distanceMeters

下列接口 MUST 支持可选 viewer 坐标；当 viewer 与目标帖均有坐标时，响应 MUST 含 `distanceMeters` 字符串（米，纯数字字符串，如 `"1234"`），MUST NOT 含距离文案标签；否则 MUST 省略该字段：

- `GET /ucg/app/api/feed/recommend` — query `lat`、`lng`（可选）；**例外**：当 viewer 已登录且为帖子作者时，该 item MUST omit `distanceMeters`（见 `ucg-recommend-feed` 本人帖场景）
- `GET /ucg/app/api/posts/{id}` — query `lat`、`lng`（可选）
- `PUT /ucg/app/api/posts/{id}` — body 可选 `lat`、`lng`（更新帖坐标）

#### Scenario: Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且帖有坐标且 viewer 非该帖作者
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 推荐 Feed 本人帖 omit 距离

- **WHEN** `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且帖有坐标且 viewer 为该帖作者
- **THEN** 该帖 JSON MUST NOT 含 `distanceMeters`

#### Scenario: 无坐标省略距离

- **WHEN** `GET /ucg/app/api/posts/{id}` 无 lat/lng query 或帖无坐标
- **THEN** 响应 MUST NOT 含 `distanceMeters` 字段

#### Scenario: 更新帖坐标

- **WHEN** `PUT /ucg/app/api/posts/{id}` body 含新 lat/lng
- **THEN** 服务 MUST 更新存储坐标并同步 Redis GEO 与 post snapshot

### Requirement: 推荐 Feed cursor 参数 SHALL 冻结 session 上下文

`GET /ucg/app/api/feed/recommend` MUST 接受 query：

- `cursor`（可选，opaque）
- `pageSize`（可选，默认 20，最大 50）
- `lat`、`lng`（可选，仅首屏无 cursor 时生效）

有 `cursor` 时 MUST 忽略新的 lat/lng。响应 `nextCursor` MUST 在 `hasMore=true` 时存在。

#### Scenario: 翻页携带 cursor

- **WHEN** 客户端用上一页 `nextCursor` 请求第二页
- **THEN** 系统 MUST 使用 cursor 内冻结坐标与 session 且 MUST 返回不重复 postId

#### Scenario: 下拉刷新无 cursor

- **WHEN** 客户端不传 `cursor` 请求 Feed
- **THEN** 系统 MUST 创建新 feed session 且 MAY 使用本次 lat/lng

### Requirement: 关注 Feed API SHALL 保持 page/total 分页并可选接受 viewer 坐标

`GET /ucg/app/api/feed/following` MUST 继续要求有效 `X-Internal-Wx-Id`。MUST 接受 query `page`（从 1）、`pageSize`（默认 20，最大 50）及可选 `lat`、`lng`。响应 MUST 为 `{ list, total, page, pageSize }`（`UcgPageRes`），**MUST NOT** 含 `nextCursor` 或 `hasMore`。本 endpoint 为既有路由的 query 扩展，**MUST NOT** 作为新 App API 计入 usage 统计。

#### Scenario: 关注 Feed 分页契约不变

- **WHEN** `GET /ucg/app/api/feed/following?page=2&pageSize=20`
- **THEN** 响应 MUST 含 `list`、`total`、`page`、`pageSize`
- **AND** MUST NOT 含 `nextCursor` 或 `hasMore`

#### Scenario: 关注 Feed 可选坐标 query

- **WHEN** `GET /ucg/app/api/feed/following?page=1&lat=31.2&lng=121.5`
- **THEN** 服务 MUST 接受 lat/lng 并用于列表 item 距离计算
- **AND** 分页字段 MUST 仍为 `{ list, total, page, pageSize }`

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误

---

## ucg-app-profile

<!-- source: openspec/specs/ucg-app-profile/spec.md -->

# ucg-app-profile Specification

## Purpose
TBD - created by archiving change ucg-public-profile-stats. Update Purpose after archive.
## Requirements
### Requirement: 公开 profile 返回社交统计

`GET /ucg/app/api/profile/{wxId}` MUST 在响应中返回与 `GET /ucg/app/api/profile/me` 一致的 `followingCount`、`followerCount`、`postCount`，数值 MUST 经 ucg 库内实时 COUNT（`ucg_follow`、`ucg_post`），与 `enrichProfileStats` 语义相同。

#### Scenario: 他人主页展示关注数

- **WHEN** 客户端请求 `GET /ucg/app/api/profile/{wxId}` 且该 wx 存在 profile 行
- **THEN** 响应 MUST 包含 `followingCount` 等于 `COUNT(ucg_follow WHERE follower_wx_id = wxId)`
- **AND** 响应 MUST 包含 `followerCount` 等于 `COUNT(ucg_follow WHERE followee_wx_id = wxId)`
- **AND** 响应 MUST 包含 `postCount` 等于 `COUNT(ucg_post WHERE author_wx_id = wxId)`

#### Scenario: profile 不存在

- **WHEN** 请求的 wxId 无 `ucg_profile` 行
- **THEN** MUST 返回 404，行为与变更前一致

---

## ucg-audit-mq

<!-- source: openspec/specs/ucg-audit-mq/spec.md -->

# ucg-audit-mq Specification

## Purpose
TBD - created by archiving change ucg-mq-green-audit. Update Purpose after archive.
## Requirements
### Requirement: UCG 审核事件 SHALL 使用已注册的 RabbitMQ routing keys

`ucg-service` 发布 UCG 审核事件时 MUST 使用下列 routing key 之一，且 MUST 经 `eventkit` 注册校验通过：

- `ucg.post.created`

- `ucg.comment.created`

- `ucg.profile.patch.submitted`

- `ucg.chat.msg.created`

#### Scenario: 发帖后发布事件

- **WHEN** 用户 submit 创建帖子且数据库事务提交成功

- **THEN** 系统 MUST Publish `ucg.post.created`，载荷至少含 `postId` 与 `auditVersion`（等于 `ucg_post.audit_version` 当前值）

#### Scenario: 未注册 routing key 拒绝发布

- **WHEN** 代码尝试 Publish 未在 eventkit 注册的 ucg 路由键

- **THEN** Publish MUST 失败且 MUST NOT 静默丢弃

### Requirement: 所有 UCG 审核 MQ 载荷 MUST 携带 auditVersion

四类审核事件的 JSON 载荷 MUST 含非空 `auditVersion`（INT），且 MUST 等于 **入队 outbox 时刻**对应权威表列的当前值（冻结在 outbox `payload` 内）：

- 帖子：`ucg_post.audit_version`
- 评论：`ucg_post_comment.audit_version`
- 资料：`ucg_profile_audit_job.audit_version`（载荷 MUST 含 `jobId`）
- 私信：`ucg_chat_message.audit_version`（载荷 MUST 含 `messageId` 与 `conversationId`）

relay worker Publish MUST 使用 outbox 内冻结的 `payload`，MUST NOT 在重试时从业务表重新读取版本（避免与用户再提审后的新版本混淆）。

#### Scenario: 资料审核载荷含 job 版本

- **WHEN** 用户提交资料变更且 job 行 `audit_version=2` 入队 outbox
- **THEN** outbox `payload` MUST 含 `jobId` 与 `auditVersion=2`

#### Scenario: 评论审核载荷含版本

- **WHEN** 用户发表评论且评论行 `audit_version=1` 入队 outbox
- **THEN** outbox `payload` MUST 含 `commentId` 与 `auditVersion=1`

### Requirement: RabbitMQ 拓扑 SHALL 为 UCG 审核队列绑定 topic exchange

仓库 `hack/rabbitmq-init` 脚本 MUST 为上述四个 routing key 创建 durable 队列并完成与 `voice.events` topic exchange 的 binding。runbook MUST 文档化队列名与初始化步骤。

#### Scenario: 本地/测试环境 init 后队列存在

- **WHEN** 运维执行 rabbitmq init 脚本

- **THEN** 管理台 SHALL 可见 `ucg.post.created.q` 等队列且 binding 正确

### Requirement: UCG 审核消费者 SHALL 驻留 ucg-service 并按 routing key 分发

`ucg-service` MUST 启动 UCG 审核 **AMQP push consumer**（TCP 5672，`autoAck=false`），从四个 UCG 审核 durable 队列接收 broker 推送的消息并调用 Green 审核逻辑。consumer MUST NOT 部署在 worker-service 内直连 ucg 数据库。消费者 MUST 从载荷读取 `auditVersion` 用于 CAS。每个 UCG 审核队列 MUST 运行 **一个** AMQP consumer goroutine；并发 MUST 由 channel **`prefetch`**（环境变量 `UCG_AUDIT_MQ_PREFETCH`，默认 5）控制。MUST NOT 使用 HTTP Management API `/queues/.../get` 轮询作为 UCG 审核消费路径。

#### Scenario: 消费帖子审核消息

- **WHEN** 队列 `ucg.post.created.q` 收到合法 `ucg.post.created` 载荷（含 `postId`、`auditVersion`）
- **THEN** consumer MUST 执行 Green 审核并在成功后 CAS 更新帖子状态（条件含 `status` 与 `audit_version`），且 MUST 在处理完成后 **manual Ack**

#### Scenario: 四队列并行消费

- **WHEN** ucg-service 启动且 `UCG_AUDIT_MQ_CONSUMER_ENABLED=true`
- **THEN** 系统 MUST 对 `ucg.post.created.q`、`ucg.comment.created.q`、`ucg.profile.patch.submitted.q`、`ucg.chat.msg.created.q` 各建立 AMQP Consume，且 MUST NOT 依赖 ticker 轮询依次拉取四队列

### Requirement: 审核消费者 MUST 通过 audit_version CAS 更新且过期消息 SHALL 优雅跳过

所有 MQ 审核消费者在写审态时 MUST 使用条件更新：`WHERE id=? AND status=? AND audit_version=?`（私信为 `audit_status` + `audit_version`）。CAS 成功时 MUST NOT 递增 `audit_version`。`RowsAffected=0` MUST 视为重复投递或过期版本（如用户已再提审 bump 版本），MUST ACK 且 MUST NOT 无限重试，MUST NOT 覆盖新轮次状态。

#### Scenario: 重复消费同一 post 事件

- **WHEN** 同一 `postId` 与 `auditVersion` 的事件被投递两次且首次已 CAS 成功（status 已非 pending）

- **THEN** 第二次 CAS MUST 影响 0 行且 MUST NOT 将已 published 帖改回 pending

#### Scenario: 再提审后旧版本消息过期

- **WHEN** 用户再提审使 `ucg_post.audit_version` 从 2 递增为 3，队列中仍存在 `auditVersion=2` 的消息

- **THEN** consumer CAS `status=1 AND audit_version=2` MUST 影响 0 行且 MUST ACK，当前帖 MUST 保持 version=3 的 pending 状态

#### Scenario: 资料 CAS 仅针对 job 表版本

- **WHEN** consumer 处理 `ucg.profile.patch.submitted` 且载荷 `auditVersion=1`

- **THEN** UPDATE MUST 针对 `ucg_profile_audit_job` 且 MUST 使用 `status=1 AND audit_version=1`；MUST NOT 以 Redis 或 `ucg_profile` 列作为 CAS 版本

### Requirement: MQ Publish 失败 SHALL 可补发且 MUST NOT 阻塞用户主路径（聊天除外可选）

帖/评/资料 submit API 在事务与 outbox 入队成功但 HTTP Publish 失败时 MUST 记录 warning/error 日志；系统 MUST 通过 **`ucg_audit_publish_outbox` relay worker** 自动重试 Publish（使用 outbox 冻结载荷）。系统 **MUST NOT** 运行定时扫描 pending 审态业务表的 reconciler 作为恢复机制。聊天 WS 在 Redis 投递成功且 outbox 入队成功但 Publish 失败时 MUST 同样由 relay worker 恢复。

#### Scenario: 发帖 Publish 失败仍返回成功

- **WHEN** 帖子已写入 `status=1` 且 outbox 已 commit 但 RabbitMQ 暂不可用
- **THEN** HTTP 创建接口 MAY 仍返回 200 与帖子 DTO；relay worker MUST 在 MQ 恢复后 Publish 并成功标记 outbox `done`

### Requirement: UCG AMQP 审核 consumer MUST 使用 manual ack 且 SHALL 按处理结果 Ack 或 Nack

UCG 审核 AMQP consumer MUST 以 `autoAck=false` 订阅队列。消息 MUST 在 Green + CAS 业务处理完成且 handler 返回成功后 **Ack**。下列情况 MUST **Ack** 且 MUST NOT requeue：JSON 非法、缺 `auditVersion`、实体不存在、CAS `RowsAffected=0`（过期/重复）。Green 或数据库 **可重试** 错误 MUST **Nack(requeue=true)**。进程在处理完成前崩溃时，broker MUST 能重投未 Ack 的消息。

#### Scenario: 处理成功后 Ack

- **WHEN** consumer 收到合法帖子审核消息且 Green 通过、CAS 更新成功
- **THEN** consumer MUST 向 broker 发送 Ack，该消息 MUST NOT 再次投递给同一 consumer

#### Scenario: 可重试错误 Nack requeue

- **WHEN** Green API 返回临时错误或 MySQL 更新失败且非 CAS 0 行跳过
- **THEN** consumer MUST Nack(requeue=true)，消息 MUST 可再次被消费

#### Scenario: 毒消息 Ack 丢弃

- **WHEN** 消息体非合法 JSON 或缺少 `auditVersion`
- **THEN** consumer MUST 记录 warning 日志并 Ack，MUST NOT 无限 requeue

### Requirement: UCG AMQP consumer SHALL 支持连接断线重连

当 RabbitMQ AMQP 连接或 channel 异常关闭时，ucg-service 中的 UCG 审核 consumer MUST 以 backoff 策略自动重连并恢复四队列 Consume。AMQP 连接失败 MUST NOT 导致 ucg-service HTTP API 进程退出；MUST 记录可观测 warning/error 日志。

#### Scenario: RabbitMQ 短暂不可用后恢复

- **WHEN** ucg-service 运行中 RabbitMQ 重启导致 AMQP 断开
- **THEN** consumer MUST 在连接恢复后继续消费队列消息，reconciler MUST 仍可补发 pending 条目

### Requirement: UCG 审核 Publisher MAY 保持 HTTP Management API

本变更 MUST NOT 要求 UCG 审核事件 Publisher 改为 AMQP。`ucg-service` MAY 继续通过 HTTP `MQ_HTTP_API_BASE` 向 `voice.events` exchange 发布审核事件；AMQP consumer 与 HTTP Publisher 并存 MUST 为受支持的部署形态。

#### Scenario: HTTP 发布 AMQP 消费

- **WHEN** 发帖后 HTTP Publisher 成功发布 `ucg.post.created` 至 exchange
- **THEN** 绑定队列中的消息 MUST 由 AMQP push consumer 接收并处理

---

## ucg-audit-mq-reliability

<!-- source: openspec/specs/ucg-audit-mq-reliability/spec.md -->

# ucg-audit-mq-reliability Specification

## Purpose
TBD - created by archiving change fix-ucg-audit-mq-green-retry-storm. Update Purpose after archive.
## Requirements
### Requirement: UCG audit MQ consumer SHALL bound apply retries and stop infinite requeue

`ucg-service` 内 UCG 审核 AMQP consumer（含 `ucg.profile.patch.submitted.q`、`ucg.post.created.q` 及本变更纳入的其它审核队列）在 handler 处理单条 delivery 时 MUST 区分 **机审阶段**与 **apply 阶段**。当 apply 阶段失败且实体 `apply_attempts` 未达配置上限时，handler MAY 返回 error 触发 `Nack(requeue=true)`。当 `apply_attempts` 达到上限 `UCG_AUDIT_APPLY_MAX_ATTEMPTS`（默认 5，可通过环境变量覆盖）时，系统 MUST 将实体标记为 apply 失败终态、记录可观测告警日志，且 handler MUST 返回 nil 以 **Ack** 该 delivery，MUST NOT 无限 requeue。

#### Scenario: apply 失败未达上限时 requeue

- **WHEN** 资料 job 机审 verdict 已为 pass，但 `approveProfileJobCAS` 因 DB 错误失败，且 `apply_attempts` 为 2、上限为 5
- **THEN** 系统 MUST 将 `apply_attempts` 递增为 3，handler MUST 返回非 nil error，且该 MQ 消息 MUST 被 Nack requeue

#### Scenario: apply 失败达上限时 Ack 停止风暴

- **WHEN** 同上失败场景且递增后 `apply_attempts` 等于上限 5
- **THEN** 系统 MUST 标记 job 为 apply 失败终态（如 `status=apply_failed` 或等价字段组合），MUST 输出含 jobId 与 auditVersion 的 error 级日志，且 handler MUST 返回 nil 使消息 Ack，后续同 payload 重投 MUST NOT 再次调用 Green

#### Scenario: 队列 redelivery 不触发无限 Nack 环

- **WHEN** 同一 delivery 因历史缺陷已被 requeue 超过 apply 上限，且 DB 已记录 apply 失败终态
- **THEN** handler MUST 返回 nil Ack，队列 `messages_unacknowledged` MUST NOT 因该 job 无限增长

### Requirement: UCG audit handler SHALL treat persisted moderation verdict as Green skip signal

当 `ucg_profile_audit_job` 或 `ucg_post`（及纳入本能力的同类审核实体）在对应 `audit_version` 下已持久化 `moderation_verdict`（非 0）时，MQ consumer handler 在 Phase 2 apply 重试路径 MUST NOT 再次调用 `GreenModerator` 的 `ModerateText` / `ModerateImageURL` / `ModerateVideoURL`。

#### Scenario: MQ 重投跳过 Green 仅重试 apply

- **WHEN** profile job `id=100`、`audit_version=2` 的 `moderation_verdict=1`（pass），`status` 仍为 pending，且 MQ 消息因先前 apply 失败被 redeliver
- **THEN** handler MUST NOT 调用 Green API，MUST 仅执行 apply 阶段（`approveProfileJobCAS` 或等价逻辑）

#### Scenario: 首次消费执行 Green 并落库 verdict

- **WHEN** 新提交 profile job `moderation_verdict=0` 且收到首次 audit MQ 消息
- **THEN** handler MUST 调用 Green 完成机审，且 MUST 在 apply 之前将 `moderation_verdict` 与 `moderation_reason`（若 reject）持久化到 MySQL

#### Scenario: 并发双消费避免重复 Green

- **WHEN** 两条相同 job 的 delivery 几乎同时被不同 consumer 处理，且 `moderation_verdict` 仍为 0
- **THEN** Phase 1 持久化 MUST 使用带 `moderation_verdict=0` 条件的 CAS/UPDATE，至多一条成功写入 verdict；另一条 MUST 读取已写入 verdict 后跳过 Green 进入 apply

### Requirement: UCG audit runbook SHALL document stuck queue remediation

`docs/runbooks/release-deploy-and-run.md` MUST 包含 UCG 审核 MQ 卡死与 apply 失败的人工处理步骤，至少覆盖：暂停 consumer（`UCG_AUDIT_MQ_CONSUMER_ENABLED`）、核对 DB 中 `moderation_verdict` 与 `status`、在安全条件下 purge 队列积压、对 `moderation_verdict=pass` 且长期 pending 的 job 的修复指引。

#### Scenario: 运维按 runbook 处理 profile 队列积压

- **WHEN** `ucg.profile.patch.submitted.q` ready 消息数异常升高，且 DB 显示多条 job 已 `moderation_verdict=1` 但 `status=pending`
- **THEN** 运维人员 MUST 能按 runbook 暂停 consumer、确认无需重复 Green 后清理或重放消息，且 MUST NOT 仅无限重启 consumer 而不查 DB

---

## ucg-audit-publish-outbox

<!-- source: openspec/specs/ucg-audit-publish-outbox/spec.md -->

# ucg-audit-publish-outbox Specification

## Purpose
TBD - created by archiving change ucg-audit-publish-outbox. Update Purpose after archive.
## Requirements
### Requirement: 审核 Publish MUST 经 transactional outbox 持久化

`ucg-service` 在触发四类 UCG 审核 MQ 事件（`ucg.post.created`、`ucg.comment.created`、`ucg.profile.patch.submitted`、`ucg.chat.msg.created`）时，MUST 在与对应业务写库**同一数据库事务**内 INSERT 一行至 `ucg_audit_publish_outbox`（`status=pending`）。outbox 行 MUST 含 `routing_key` 与 JSON `payload`，且 `payload.auditVersion` MUST 等于入队时刻权威表上的当前 `audit_version`（冻结快照，非 relay 时再读库）。

#### Scenario: 发帖 submit 同事务入队

- **WHEN** 用户 submit 创建帖子且事务 INSERT `ucg_post` 成功
- **THEN** 同一事务 MUST INSERT outbox 行，`routing_key=ucg.post.created`，载荷含 `postId` 与 `auditVersion`

#### Scenario: 再提审 bump 版本后入队

- **WHEN** 用户再提审使 `ucg_post.audit_version` 递增为 3 且事务提交
- **THEN** outbox 新行 MUST 含 `auditVersion=3`（非旧版本 2）

### Requirement: Audit Publish Relay Worker MUST 仅轮询 outbox 表

`ucg-service` MUST 运行 `StartAuditPublishRelayWorker`，按配置间隔从 `ucg_audit_publish_outbox` 选取 `status IN (pending, failed)` 且 `attempts < maxAttempts` 的行（`ORDER BY id ASC LIMIT 1`），经 HTTP 向 `voice.events` Publish。成功 MUST 将行标记为 `done`；失败 MUST 递增 `attempts`、记录 `last_error` 并标记 `failed`（未达 maxAttempts 时 MUST 可被后续 tick 重试）。Worker MUST NOT 扫描 `ucg_post`、`ucg_post_comment`、`ucg_profile_audit_job`、`ucg_chat_message` 的 pending 审态以发现漏发事件。

#### Scenario: RabbitMQ 短暂不可用后恢复

- **WHEN** outbox 行 Publish 连续失败且 RabbitMQ 恢复
- **THEN** relay worker MUST 重试该行直至 Publish 成功并标记 `done`

#### Scenario: Worker 不扫 pending 帖表

- **WHEN** 存在 `ucg_post.status=1` 但无对应 outbox 行（历史脏数据）
- **THEN** relay worker MUST NOT 因扫表而补发；恢复 MUST 依赖一次性运维 seed 或新业务路径

### Requirement: 事务提交后 SHOULD best-effort 即时 Publish

业务事务成功提交后，系统 SHOULD 尝试对刚写入的 outbox 行执行一次即时 HTTP Publish；若成功 MUST 将该行标记为 `done` 且 relay worker MUST NOT 重复 Publish 已成功行。

#### Scenario: 即时 Publish 成功

- **WHEN** 发帖事务提交且 RabbitMQ 可用
- **THEN** outbox 行 SHOULD 在 relay worker 介入前变为 `done`

### Requirement: Publish 失败 MUST NOT 阻塞用户主路径

帖/评/资料 submit API 与聊天 WS 投递在 outbox 入队成功但即时 Publish 失败时 MUST 仍完成用户可见的成功路径（与现网一致）；恢复 MUST 由 relay worker 承担，MUST NOT 依赖 pending 业务表 reconciler。

#### Scenario: 发帖 Publish 失败仍返回成功

- **WHEN** 帖子与 outbox 已 commit 但 RabbitMQ 暂不可用
- **THEN** HTTP 创建接口 MAY 仍返回 200；outbox 行 MUST 保持 `pending`/`failed` 直至 relay 成功

---

## ucg-chat-mysql-persist

<!-- source: openspec/specs/ucg-chat-mysql-persist/spec.md -->

# ucg-chat-mysql-persist Specification

## Purpose
TBD - created by archiving change ucg-chat-mysql-persist. Update Purpose after archive.
## Requirements
### Requirement: 私信发送 SHALL 同步写入 Redis 与 MySQL outbox

`DeliverChatMessage` 在 Green 审核通过后，SHALL 先执行 Redis `INCR seq` 与 `RPUSH` 消息 JSON 至 `ucg:chat:conv:{conversationId}:msgs`；随后 SHALL 同步 INSERT 一行至 `ucg_chat_message_outbox`（payload 含完整消息字段，status=pending）。Redis 写入失败 SHALL 中止发送；outbox 写入失败 SHALL 记录 Error 日志且 SHALL NOT 阻止 WS `message_delivered` 推送。

#### Scenario: 正常发送双写

- **WHEN** 用户发送私信且 Green 审核通过
- **THEN** Redis LIST 中 SHALL 包含该消息 JSON
- **AND** MySQL `ucg_chat_message_outbox` SHALL 存在对应 pending 行

#### Scenario: Redis 写入失败

- **WHEN** Redis RPUSH 失败
- **THEN** 系统 SHALL NOT 写入 outbox
- **AND** 系统 SHALL NOT 向接收方推送 message_delivered

### Requirement: persist worker SHALL 异步将 outbox flush 至 ucg_chat_message

ucg-service SHALL 运行 `StartChatPersistWorker`，轮询 `ucg_chat_message_outbox` 中 status=pending 或 failed（未超最大 attempts）的记录；SHALL INSERT 至 `ucg_chat_message`（utf8mb4），幂等处理重复 `(conversation_id, id)`；成功后 SHALL 将 outbox 行标记为 done。

#### Scenario: outbox 成功落库

- **WHEN** worker 处理 pending outbox 行且 MySQL 可用
- **THEN** `ucg_chat_message` SHALL 存在与 Redis msg id 一致的记录
- **AND** outbox 行 status SHALL 变为 done

#### Scenario: 重复 flush 幂等

- **WHEN** worker 对同一 outbox 行重试 INSERT
- **THEN** `ucg_chat_message` SHALL NOT 出现重复 `(conversation_id, id)` 行

### Requirement: 读消息 SHALL Redis 优先并在 miss 时回源 MySQL

`listChatMessages` SHALL 优先从 Redis LRANGE 分页返回。当 Redis `LLEN` 为 0 且 MySQL `ucg_chat_message` 对该会话 `COUNT(*) > 0` 时，SHALL 从 MySQL 按 id 升序分页查询并返回等效 `ChatMessage` 结构。

#### Scenario: Redis 有缓存

- **WHEN** Redis LIST 非空
- **THEN** 系统 SHALL 从 Redis 返回消息，SHALL NOT 查询 MySQL

#### Scenario: Redis 空且 MySQL 有历史

- **WHEN** Redis LIST 为空且 MySQL 存在该会话消息
- **THEN** 系统 SHALL 从 MySQL 返回消息列表

### Requirement: MySQL 回源读 SHALL 按需 lazy warm Redis 并对齐 seq

从 MySQL fallback 读取时，系统 SHALL 将读到的消息 JSON RPUSH 回 Redis（至少当前页；可配置整会话阈值）；SHALL 将 `ucg:chat:conv:{id}:seq` 设为 MySQL 该会话 `MAX(id)`（当 Redis seq 缺失或小于该值时）。lazy warm SHOULD 使用短 TTL 分布式锁避免并发重复重建。

#### Scenario: Redis 丢数据后首次读会话

- **WHEN** Redis 消息 LIST 为空但 MySQL 有消息且用户拉取历史
- **THEN** 用户 SHALL 看到 MySQL 中的消息
- **AND** Redis seq SHALL 对齐至 MySQL MAX(id)

### Requirement: ucg_chat_message 表 SHALL 使用 utf8mb4 存储正文

`ucg_chat_message.content` 及表默认字符集 SHALL 为 utf8mb4，以支持 emoji 等四字节 Unicode 入库。

#### Scenario: 含 emoji 的私信持久化

- **WHEN** 用户发送含 emoji 的私信且 worker flush 完成
- **THEN** MySQL 中该消息 content SHALL 原样保留 emoji

---

## ucg-chat-ws

<!-- source: openspec/specs/ucg-chat-ws/spec.md -->

# ucg-chat-ws Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: Chat SHALL use Redis for durable message storage and WebSocket delivery

ucg-service SHALL persist chat messages in Redis without TTL (forever retention in MVP), expose internal `GET /ws/chat` WebSocket with JWT auth first frame, and push real-time events to conversation members after audit pass. App clients SHALL connect via gateway-app external path `/ucg/app/ws/chat` (upgrade proxy to internal `/ws/chat`).

#### Scenario: WS 首帧认证
- **WHEN** 客户端经 gateway 连接后首帧 JSON 含合法 JWT
- **THEN** ucg-service SHALL 保持连接并注册 wxId；非法 JWT SHALL 关闭连接

#### Scenario: Redis 永久保留
- **WHEN** 消息审核通过并投递
- **THEN** 消息 SHALL 写入 Redis 键空间且 SHALL NOT 设置过期淘汰（MVP）

#### Scenario: 内部 WS 不经公网暴露
- **WHEN** 部署 ucg-service
- **THEN** `/ws/chat` MAY 仅集群内可达；App 对外入口 MUST 为 gateway `/ucg/app/ws/chat`

### Requirement: Conversation list SHALL support unread counts and pin/delete flags

API SHALL return conversations with unread_count, pinned, last message preview; member soft-delete via `deleted_at` on `ucg_conversation_member`. List ordering SHALL use `pinned DESC, updated_at DESC` from `ucg_conversation_member`.

`GET /ucg/app/api/conversations` MUST NOT 因单条会话的对方成员缺失或对方账号已注销而返回整页错误。对每条当前用户未软删的会话，响应 MUST 仍包含 `id`、`unreadCount`、`lastPreview`、`pinned`、`updatedAt`。

当 `ucg_conversation_member` 中仍存在对方成员行时，列表项 `peerWxId` MUST 为该行的历史 `wx_id`。当无法解析对方成员行时，`peerWxId` MAY 为 `0`。

当对方成员行不存在，或 device internal `ValidateWx` 表明该 `peerWxId` 对应 wx 不存在时，列表项 MUST NOT 填充 `peerNickname`、`peerAvatarKey`、`peerAvatarUrl`、`peerAvatarThumbnailUrl`（省略或空串）。发消息、创建会话等写路径 MUST 保持原有严格校验，不在本 Requirement 中放宽。

#### Scenario: 未读计数

- **WHEN** 收件人收到新消息
- **THEN** 其 `unread_count` SHALL 递增直至调用 read API

#### Scenario: 对方成员行缺失时会话仍出现在列表

- **WHEN** 用户请求 `GET /ucg/app/api/conversations` 且某 direct 会话中除本人外无其他 `ucg_conversation_member` 行
- **THEN** 该会话 MUST 仍出现在 `list` 中且接口 MUST 返回成功
- **AND** 该项 `peerWxId` MAY 为 `0`
- **AND** `peerNickname` 与 avatar 相关字段 MUST 为空

#### Scenario: 对方已注销时保留 peerWxId 且展示为空

- **WHEN** 用户请求会话列表且某会话对方 `ucg_conversation_member` 行仍存在 `wx_id=W`，但 `ValidateWx(W)` 返回 `exists=false`
- **THEN** 该项 `peerWxId` MUST 为 `W`
- **AND** `peerNickname` 与 avatar 相关字段 MUST 为空
- **AND** 接口 MUST 返回成功

#### Scenario: 发消息路径仍要求有效对方

- **WHEN** 客户端经 WebSocket 向会话发送消息且对方成员不存在
- **THEN** 系统 MUST 拒绝发送且 MUST NOT 因本变更而静默成功

### Requirement: Outbound chat messages SHALL be processed via WebSocket handler
`ucg-service` WebSocket handler MUST 在收到 `type=message` 后：校验成员关系；向发送方发送 `message_ack`；**立即**写入 Redis 消息（`audit_status=pending`，`audit_version` 与 `ucg_chat_message` 权威列一致）、增加收件人未读、向收件方发送 `message_delivered`；Publish `ucg.chat.msg.created`（载荷含 `auditVersion`）。**MUST NOT** 在 handler 内同步调用 Green 阻塞投递。

#### Scenario: 发送方先收到 ack

- **WHEN** 客户端发送合法 message 帧
- **THEN** 发送方 MUST 先收到 `message_ack` 再收到异步审核结果

#### Scenario: 收件方先于审核收到消息

- **WHEN** 消息已写入 Redis
- **THEN** 收件方 MUST 收到 `message_delivered`，且 body MUST 含 pending 审态标识与 `audit_version`

### Requirement: WebSocket SHALL emit post-audit rejection events
Green 事后驳回时，系统 MUST：

- 向发送方 MUST 推送 `audit_failed`（含 `clientMsgId`、`reason`）
- 向在线收件方 MUST 推送 `msg_hidden`（含 `conversationId`、`messageId`）

仅当 CAS（`audit_status='pending' AND audit_version=?`）成功将消息置 rejected 后 MUST 推送上述事件；CAS 0 行（过期/重复消息）MUST NOT 推送。

#### Scenario: 在线收件人隐藏违规消息

- **WHEN** 消息 CAS 为 rejected 且收件方 WS 在线
- **THEN** 系统 MUST 向收件方推送 `msg_hidden`

#### Scenario: 重复 reject 消息不重复推送

- **WHEN** 同一 `auditVersion` 的 reject 事件重复消费且 CAS 影响 0 行
- **THEN** 系统 MUST NOT 再次推送 `audit_failed` 或 `msg_hidden`

### Requirement: Chat messages with video SHALL include mediaThumbnailUrl as physical thumb CDN URL

ucg-service 在 enrich 聊天消息媒体（含 HTTP 消息列表、WS 实时 `message_delivered` 及审核通过后推送）时，当消息含非空 `videoKey` 时 MUST 设置 `mediaThumbnailUrl` 为 `BuildVideoThumbnailURL(videoKey)`。该 URL MUST NOT 含 `x-oss-process`。`mediaCdnUrl` MUST 仍为视频原片 CDN URL。

本要求 MUST 在 Chat WebSocket v1 首版即满足，MUST NOT 推迟至后续 API 版本。MUST NOT 改变 v1 消息 JSON 字段结构，仅填充/修正 `mediaThumbnailUrl` 值。

#### Scenario: WS 实时视频消息含 mediaThumbnailUrl

- **WHEN** 收件方经 WS 收到含 `videoKey` 的 `message_delivered` 且 OSS 存在对应 `_thumb.jpg`
- **THEN** 消息 body SHALL 含非空 `mediaThumbnailUrl` 且 URL SHALL NOT 含 `x-oss-process`

#### Scenario: HTTP 消息历史与 WS 语义一致

- **WHEN** 客户端 `GET` 会话消息列表且行含 `videoKey`
- **THEN** 每项 SHALL 含与 WS 相同规则的 `mediaThumbnailUrl`

#### Scenario: 图片消息行为不变

- **WHEN** 消息仅含 `imageKey`
- **THEN** `mediaThumbnailUrl` SHALL 仍为 `BuildImageThumbnailURL(imageKey)` 物理缩略图 URL

---

## ucg-data-model

<!-- source: openspec/specs/ucg-data-model/spec.md -->

# ucg-data-model Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: UCG data SHALL live in ai_voice_ucg with defined post status enum

Database `ai_voice_ucg` SHALL contain tables: `ucg_profile`, `ucg_post`, `ucg_post_media`, `ucg_follow`, `ucg_post_like`, `ucg_post_comment`, `ucg_conversation`, `ucg_conversation_member`, and MAY contain `ucg_post_recommend`. Post `status` MUST use: 0=draft, 1=pending_audit, 2=published, 3=rejected.

#### Scenario: 创建待审核帖
- **WHEN** 用户提交帖子
- **THEN** `ucg_post.status` SHALL 为 1（pending_audit），且 SHALL NOT 为 2 直至 Green 通过

#### Scenario: 拒绝态记录原因
- **WHEN** Green 审核失败
- **THEN** `ucg_post.status` SHALL 为 3 且 `reject_reason` SHALL 非空

### Requirement: Timestamps SHALL use unix seconds

All `created_at`/`updated_at`/`published_at` columns MUST store unix seconds consistent with `database-unix-timestamp-storage` baseline.

#### Scenario: 写入创建时间
- **WHEN** 插入新 post
- **THEN** `created_at` SHALL 为 unix 秒级整数

### Requirement: Conversation member list SHALL be sortable by pin and last activity

`ucg_conversation_member` MUST include `updated_at` (unix seconds) maintained on new messages or pin changes; index `idx_wx_list (wx_id, pinned, updated_at)` SHALL support per-user conversation list ordering.

#### Scenario: 新消息刷新排序
- **WHEN** 会话成员收到审核通过的新消息
- **THEN** 各成员行的 `updated_at` SHALL 更新为当前 unix 秒

### Requirement: ucg_post SHALL include audit_version for CAS

表 `ucg_post` MUST 新增列 `audit_version INT NOT NULL DEFAULT 1`。新建 pending 帖 MUST 从 1 起始。用户再提审进入 pending 时 MUST `audit_version++`。消费者 CAS 审态变更成功 MUST NOT 递增该列。

#### Scenario: 新帖默认版本

- **WHEN** 用户 submit 创建帖子

- **THEN** 插入行 MUST 含 `audit_version=1` 且 `status=1`

#### Scenario: 再提审递增版本

- **WHEN** 作者对已发布或驳回帖 submit 再提审

- **THEN** 行 MUST `audit_version` 递增且 `status=1`；MUST NOT 回滚帖文正文等业务字段

### Requirement: ucg_post_comment SHALL include audit status columns

表 `ucg_post_comment` MUST 新增：

- `status TINYINT NOT NULL` — 0 draft（保留）、1 pending_audit、2 published、3 rejected

- `audit_version INT NOT NULL DEFAULT 1`

- `reject_reason VARCHAR`（可空）

首评 `audit_version=1`；再提审 MUST 递增。消费者 CAS MUST 使用 `status` + `audit_version`。

#### Scenario: 新评论 pending

- **WHEN** 用户发表评论

- **THEN** 插入 MUST `status=1`、`audit_version=1`

### Requirement: ucg_profile_audit_job SHALL store pending profile patches with audit_version

数据库 MUST 提供 `ucg_profile_audit_job`（或 design 等价命名）存储待审资料 patch，至少含：`wx_id`、`nickname`、`avatar_key`、`bio`、`status`（1/2/3）、**`audit_version`**、`reject_reason`、时间戳。

`audit_version` 为该 job 的 **唯一**审核轮次源；资料 CAS 与 MQ 载荷 MUST 读此列。MUST NOT 用 Redis 或 `ucg_profile` 表列作为审核版本。

#### Scenario: 资料提交创建 job

- **WHEN** 用户 PUT profile 变更

- **THEN** MUST 插入 pending job 行（`audit_version=1` 或再提审递增）而非仅写 Redis

### Requirement: ucg_chat_message SHALL include audit_status and audit_version

表 `ucg_chat_message` MUST 新增（或规范既有 `status` 字符串）：

- `audit_status`：`pending` | `approved` | `rejected`

- `audit_version INT NOT NULL DEFAULT 1`

- `reject_reason`（可空）

Redis 消息 JSON MUST 镜像 `audit_status` 与 `audit_version`，且与 MySQL 列语义一致。私信 CAS 与 MQ 载荷 MUST 以 `ucg_chat_message.audit_version` 为权威。

#### Scenario: 聊天消息初始 pending

- **WHEN** WS 发送消息并投递 Redis

- **THEN** `ucg_chat_message`（或等价 outbox 镜像）MUST `audit_status=pending` 且 `audit_version=1`；Redis JSON MUST 含相同 `audit_version`

---

## ucg-device-internal-api

<!-- source: openspec/specs/ucg-device-internal-api/spec.md -->

# ucg-device-internal-api Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: device-service SHALL expose internal HTTP for ucg wx and baby name

device-service SHALL provide internal endpoints callable only with header `X-Device-Gateway-Internal-Secret` matching `DEVICE_GATEWAY_INTERNAL_SECRET`: validate wx id, batch fetch display fields, and get baby_name for default nickname. ucg-service MUST use these APIs and MUST NOT query `wx` table directly.

#### Scenario: 校验 wxId
- **WHEN** ucg-service 内部调用 validate 且 secret 正确
- **THEN** device-service SHALL 返回 wx 是否存在及必要展示字段

#### Scenario: 错误 secret 拒绝
- **WHEN** internal 请求 secret 不匹配
- **THEN** device-service SHALL 返回 403 且 SHALL NOT 返回 wx 数据

#### Scenario: ucg 禁止直连 device 库
- **WHEN** 代码评审发现 ucg-service import device DAO 查询 wx
- **THEN** 变更 MUST 拒绝合入

---

## ucg-entry-eligibility

<!-- source: openspec/specs/ucg-entry-eligibility/spec.md -->

# ucg-entry-eligibility Specification

## Purpose
TBD - created by archiving change commercial-feature-entitlement. Update Purpose after archive.
## Requirements
### Requirement: UCG 入场资格 API SHALL 按 device_no 返回连续有效喂养日结果

cash-service MUST 提供 `GET /cash/app/api/ucg/eligibility`，按网关注入的 `X-Internal-Device-No` 返回至少：`qualified`（bool）、`requiredDays`（固定 7）、`effectiveDays`（int）、`remainingDays`（int），并 MAY 返回可选文案字段。请求 MUST 经登录到达；`device_no` 为空时 MUST 返回错误，MUST NOT 伪造合格。该接口 MUST NOT 计入 App usage 统计（维护型 skip）。

#### Scenario: 连续满 7 有效日判定合格

- **WHEN** 某 `device_no` 按规则计算出的 `effectiveDays >= 7`
- **THEN** 响应 MUST 含 `qualified=true`、`requiredDays=7`、`remainingDays=0`

#### Scenario: 不足 7 日返回剩余天数

- **WHEN** `effectiveDays = 3`
- **THEN** 响应 MUST 含 `qualified=false`、`requiredDays=7`、`remainingDays=4`

#### Scenario: 缺少 device_no

- **WHEN** 请求未提供有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝并返回错误，MUST NOT 返回 `qualified=true`

### Requirement: 有效喂养日 MUST 为上海日历日且连续统计且当日 history 不少于 10

有效喂养日判定 MUST 使用 `Asia/Shanghai`。某日有效当且仅当该 `device_no` 在该日内 history 记录数 ≥ 10。连续有效日 MUST 以请求当日为锚向前连续统计。资格计算 MUST 经 history-service HTTP 契约取数，MUST NOT 由 cash-service 直查 history 库。

#### Scenario: 单日满 10 条计为有效日

- **WHEN** 某上海日该设备 history 行数 = 10
- **THEN** 该日 MUST 计入有效喂养日

#### Scenario: 单日不足 10 条不计有效且中断连续

- **WHEN** 某上海日该设备 history 行数 = 9
- **THEN** 该日 MUST NOT 计入有效喂养日，且连续计数在该日中断

#### Scenario: 禁止跨库直查

- **WHEN** 实现资格计算
- **THEN** cash-service MUST 仅通过 history 内部 HTTP 获取按日计数，MUST NOT 连接 history 域数据库

### Requirement: UCG 入场资格 MUST NOT 受 VIP 或功能权益影响

资格计算与响应 MUST 仅基于喂养有效日规则。账号 VIP 状态、功能权益、邀请码/支付/广告开通 MUST NOT 将 `qualified` 置为 true 或缩短 `remainingDays`。客户端 MUST NOT 被本规格要求以 `isVip` 绕过入场门（服务端亦不提供 VIP 短路字段）。

#### Scenario: VIP 用户仍按喂养日计算

- **WHEN** 请求方账号为 VIP 但 `effectiveDays < 7`
- **THEN** 响应 MUST 为 `qualified=false`，且结果与非 VIP 在相同喂养数据下一致

### Requirement: 资格结果 MUST 按日缓存于 Redis 且不得落库

同一 `device_no` 在同一上海日内的资格结果 MUST 在缓存命中时直接返回，并 MUST 写入 Redis（经 `cachekit` + platform 键 builder）。MUST NOT 将资格结果作为权威数据写入 MySQL。MUST NOT 为此新增后台 ticker；miss 时在请求路径同步计算。history 调用失败时 MUST fail-closed（返回错误）。

#### Scenario: 同日第二次请求读缓存

- **WHEN** 某设备当日已成功计算并写入 Redis 后再次请求资格 API
- **THEN** 系统 MUST 返回与缓存一致的结果，且 MUST NOT 再次全量重算（除非缓存丢失）

#### Scenario: 资格不落 MySQL

- **WHEN** 资格计算完成
- **THEN** 系统 MUST NOT 将资格结果写入 MySQL 业务表

#### Scenario: Redis 经 platform

- **WHEN** 读写资格缓存
- **THEN** 代码 MUST 经 `cachekit`（含 Observer），键 MUST 来自 platform builder，MUST NOT 业务层直连 `g.Redis()`

---

## ucg-feed-index-lazy-warm

<!-- source: openspec/specs/ucg-feed-index-lazy-warm/spec.md -->

# ucg-feed-index-lazy-warm Specification

## Purpose
TBD - created by archiving change ucg-feed-index-lazy-warm. Update Purpose after archive.
## Requirements
### Requirement: 推荐 Feed 索引冷启动 SHALL 按需从 MySQL warm Redis

当推荐索引需要 warm（见下方冷/短缺判据）且 `ucg.feed.indexAutoWarmEnabled` 为 true 时，`ListRecommendFeed` MUST 在组装候选集之前执行有界 warm：分页读取 published 帖并调用与 `cmd/ucg-feed-backfill` 等价的 `syncPublishedPostRedis`（ZADD score、有坐标则 GEOADD、写 post/profile snapshot）。warm 完成后 MUST 继续现有 GEO+ZSET+cursor 读路径，**MUST NOT** 改用 MySQL 排序降级。

**需要 warm 的判据** MUST 为：MySQL 存在 `status=published` 帖子，且（`ZCARD ucg:recommend:score` 为 0，**或** `publishedCount - zcard >=` 配置阈值 `indexHealGapThreshold`（默认 **50**））。仅 `ZCARD > 0` 但相对 published **无明显短缺**时 MUST NOT 触发全量/有界 warm。

warm MUST 使用分布式锁 `ucg:feed:index:warm:lock`（cachekit 登记键）防止并发惊群；单请求 warm 帖数 MUST 不超过配置的 `indexWarmMaxPosts`（默认 2000）；单批分页大小 MUST 可配置（默认 200）。单帖 warm 失败 MUST 记日志并继续后续帖（best-effort）。

#### Scenario: ZSET 空且 DB 有 published 帖

- **WHEN** `ZCARD ucg:recommend:score` 为 0 且 MySQL published 计数 > 0，且 auto warm 开启
- **THEN** 系统 MUST warm 至少一批帖至 Redis 并重试 Feed 候选收集，响应 SHOULD 含帖子（在 cap 覆盖范围内）

#### Scenario: ZSET 非空但相对 published 明显短缺

- **WHEN** `ZCARD` 为 20、published 为 500、缺口不低于阈值，且 auto warm 开启
- **THEN** 系统 MUST 触发有界 warm 补齐索引（受 `indexWarmMaxPosts` 限制），MUST NOT 因 ZCARD>0 而跳过

#### Scenario: 索引充足不 warm

- **WHEN** `ZCARD` 与 published 的差低于阈值（含大致齐平）
- **THEN** 系统 MUST NOT 触发有界 warm

#### Scenario: auto warm 关闭

- **WHEN** `indexAutoWarmEnabled=false`
- **THEN** 系统 MUST 保持空/短缺 ZSET 时不在请求路径 warm，依赖启动 heal 或运维 backfill

#### Scenario: 并发 Feed 请求

- **WHEN** 多个请求同时检测到需要 warm
- **THEN** 仅一个请求 MUST 持有 warm 锁执行灌库；其他请求 MUST 等待或短退避后读 ZCARD，不得无界阻塞

#### Scenario: warm 与 publish 写路径一致

- **WHEN** warm 处理某 published 帖
- **THEN** Redis MUST 写入 `ucg:recommend:score` 与（若有坐标）`ucg:feed:geo` 及 snapshot，语义与 `publishPostCAS` 后同步一致

---

## ucg-feed-index-startup-heal

<!-- source: openspec/specs/ucg-feed-index-startup-heal/spec.md -->

# ucg-feed-index-startup-heal Specification

## Purpose
TBD - created by archiving change ucg-feed-index-startup-heal. Update Purpose after archive.
## Requirements
### Requirement: ucg-service SHALL self-check and auto-heal recommend feed index on startup

`ucg-service` 进程启动时（依赖检查通过后、HTTP 服务可接受请求的路径上）MUST 执行一次推荐 Feed 索引自检（任务名 **FeedIndexStartupHeal** 的检测阶段），比较：

- MySQL `ucg_post` 中 `status=published` 的计数 `publishedCount`
- Redis `ZCARD` of `ucg:recommend:score`（经 cachekit / `UCGRecommendScoreKey`）得 `zcard`

当 `publishedCount > 0` 且（`zcard == 0` 或 `publishedCount - zcard >=` 配置阈值 `indexHealGapThreshold`，默认 **50**）时，系统 MUST 打 **ERROR** 级别可观测日志（含 `publishedCount`、`zcard`、gap）。

当自检开关开启且自动补齐开关开启时，系统 MUST **异步**（MUST NOT 阻塞 HTTP `Run` 启动）执行有界 heal：分页读取 published 帖并调用与 `cmd/ucg-feed-backfill` / `syncPublishedPostRedis` 等价的写入（ZADD `ucg:recommend:score`、有坐标则 GEOADD `ucg:feed:geo`、写 post/profile snapshot）。heal MUST 使用独立分布式锁键（cachekit 登记，如 `ucg:feed:index:startup-heal:lock`），与请求路径 warm 锁分离；未获锁 MUST 跳过 heal 并 Warning，MUST NOT Fatal 进程。单帖失败 MUST 记日志并继续。heal 处理帖数 MUST 不超过配置上限 `indexStartupHealMaxPosts`（默认 **10000**）。

环境/配置 MUST 支持关闭自检或仅自检不 heal（例如 `UCG_FEED_INDEX_STARTUP_CHECK_ENABLED`、`UCG_FEED_INDEX_STARTUP_HEAL_ENABLED`；默认均开启）。关闭自检时 MUST NOT 启动 heal。本任务 MUST NOT 实现为 `time.Ticker` 周期扫表。

#### Scenario: 启动发现缺口并异步补齐

- **WHEN** ucg-service 启动且 publishedCount=500、zcard=10、阈值=50、check 与 heal 均开启
- **THEN** 系统 MUST 打 ERROR 日志含两计数，且 MUST 在后台尝试 heal 灌入索引（获锁成功时），HTTP 监听 MUST 仍可启动

#### Scenario: 无缺口不 heal

- **WHEN** publishedCount=100、zcard=100（或 gap 低于阈值）
- **THEN** 系统 MUST NOT 因启动自检触发 heal 灌库

#### Scenario: 仅检查不补齐

- **WHEN** check 开启且 heal 关闭，且存在缺口
- **THEN** 系统 MUST 打 ERROR 日志且 MUST NOT 自动 ZADD 补齐

#### Scenario: 多副本争锁

- **WHEN** 两副本同时启动且均检出缺口
- **THEN** 至多一个副本 MUST 持有 startup-heal 锁执行灌库；另一副本 MUST 跳过或等待锁策略后跳过，MUST NOT 无锁双全量狂写

#### Scenario: 自检关闭

- **WHEN** startup check 开关为 false
- **THEN** 启动 MUST NOT 执行缺口比较与 heal

---

## ucg-feed-no-geo-zset-fallback

<!-- source: openspec/specs/ucg-feed-no-geo-zset-fallback/spec.md -->

# ucg-feed-no-geo-zset-fallback Specification

## Purpose
TBD - created by archiving change ucg-feed-no-geo-zset-fallback. Update Purpose after archive.
## Requirements
### Requirement: viewer 有坐标时 unlimited 半径步 SHALL 从 ZSET 补全无 GEO 帖

当 viewer 请求含有效 lat/lng 且 GEO 半径阶梯已执行至 `radiusKm=0`（unlimited）时，Feed 候选收集 MUST 从 `ucg:recommend:score` ZSET 分页扫描成员并加入候选集，**MUST NOT** 在 unlimited 步仍仅使用 GEOSEARCH。ZSET 中 **不在** `ucg:feed:geo` 的 published 帖（无坐标）MUST 可进入 Feed，按 `baseScore` 参与 `finalScore`（`distanceTerm=0`），响应 MUST NOT 含 `distanceMeters`。

已在较小 GEO 半径（50–500km）或 pool 中的帖 MUST 由 pool/session 去重，不得重复下发。

#### Scenario: 帖仅在 ZSET 不在 GEO，viewer 带坐标

- **WHEN** published 帖在 `ucg:recommend:score` 有 member 但 `GEOPOS ucg:feed:geo` 为空，且 `GET /feed/recommend` 首屏含有效 lat/lng
- **THEN** 响应 `list` MUST 含该帖（在 pageSize 与候选 batch 覆盖范围内）

#### Scenario: unlimited 步不使用 GEO 替代 ZSET

- **WHEN** viewer 带坐标且当前半径阶梯为 `radiusKm=0`
- **THEN** 候选收集 MUST 使用 ZSET 读路径，MUST NOT 以 GEO 20000km 搜索作为 unlimited 步的唯一来源

#### Scenario: 无 viewer 坐标行为不变

- **WHEN** `GET /feed/recommend` 不含有效 lat/lng
- **THEN** 候选收集 MUST 仍按 ZSET/baseScore 排序，与变更前一致

---

## ucg-feed-redis-store

<!-- source: openspec/specs/ucg-feed-redis-store/spec.md -->

# ucg-feed-redis-store Specification

## Purpose
TBD - created by archiving change ucg-feed-geo-composite-score. Update Purpose after archive.
## Requirements
### Requirement: Feed Redis 键 SHALL 经 cachekit builder 登记且禁止业务层字面量

UCG Feed 读写的 Redis 键 MUST 在 `internal/platform/cachekit/keys_ucg.go` 登记下列 builder，业务与 controller MUST 经 `cachekit.Cache`（含 `WithObserver`）访问，MUST NOT 使用 `g.Redis()` 或键字面量拼接：

- `UCGFeedGeoKey()` → `ucg:feed:geo`（GEO）
- `UCGRecommendScoreKey()` → `ucg:recommend:score`（ZSET）
- `UCGPostSnapshotKey(postId)` → `ucg:post:snapshot:{postId}`（STRING JSON）
- `UCGProfileSnapshotKey(wxId)` → `ucg:profile:snapshot:{wxId}`（STRING JSON）
- `UCGUserLikedPostsKey(wxId)` → `ucg:user:{wxId}:liked-posts`（SET）
- `UCGFeedSessionKey(sessionId)` → `ucg:feed:session:{sessionId}`（SET，TTL 30min）

#### Scenario: 新增键仅经 platform 登记

- **WHEN** 实现需读写 Feed GEO 索引
- **THEN** 代码 MUST 调用 `UCGFeedGeoKey()` 且 MUST NOT 在 `internal/services/**` 出现 `ucg:feed:geo` 字面量

### Requirement: 帖子 snapshot SHALL 缓存 Feed 展示所需字段及 server-only 坐标

`ucg:post:snapshot:{postId}` JSON MUST 含客户端 Feed 所需字段：`id`、`content`、媒体 CDN URL、`authorWxId`、`likeCount`、`ipLocation`、`publishedAt`、`mediaType`。MAY 含 server-only `lat`、`lng` 供距离计算；该坐标 MUST NOT 出现在 App HTTP 响应体。

帖子 published 或更新时 MUST 写入/刷新 snapshot；unpublished/delete MUST DEL 键。

#### Scenario: snapshot 命中时 Feed 不查帖表

- **WHEN** `ListRecommendFeed` 组装 postId 列表且 snapshot 存在
- **THEN** 系统 MUST 从 Redis JSON 填充 `UcgPostItem` 字段且 MUST NOT 对该帖执行 `ucg_post` 单条 SELECT

#### Scenario: 坐标仅服务端使用

- **WHEN** snapshot 含 `lat`/`lng` 且 API 返回帖子
- **THEN** 响应 JSON MUST NOT 含 `lat`/`lng` 字段

### Requirement: 作者 profile snapshot SHALL 缓存公开 profile 字段

`ucg:profile:snapshot:{wxId}` JSON MUST 含 `wxId`、`nickname`、`bio`、`avatarUrl`、`avatarThumbnailUrl`。帖子 publish/update 时 MUST 刷新作者 snapshot；Feed 读路径 MUST 批量 GET snapshot 填充 `author`。

#### Scenario: Feed 作者信息来自 snapshot

- **WHEN** Feed 返回帖含 `authorWxId`
- **THEN** `author` 字段 MUST 优先来自 `ucg:profile:snapshot:{wxId}` 且 MUST NOT 在循环内逐条查 `ucg_profile`

### Requirement: 用户点赞 SET SHALL 支撑 likedByMe 批量判定

`ucg:user:{wxId}:liked-posts` SET member MUST 为 `postId` 字符串。like MUST SADD；unlike MUST SREM。Feed 读路径对当前页 postIds MUST 批量 SISMEMBER（pipeline）填充 `likedByMe`，MUST NOT 查 `ucg_post_like` 表。

#### Scenario: Feed likedByMe 走 Redis

- **WHEN** 已登录用户请求推荐 Feed 且 page 含 20 帖
- **THEN** 系统 MUST 经一次 pipeline SISMEMBER 判定 liked 状态且 MUST NOT 对每帖查询 MySQL like 表

### Requirement: Feed session SET SHALL 防止 cursor 分页重复下发

`ucg:feed:session:{sessionId}` MUST 为 SET，member 为已返回 `postId`；TTL MUST 为 30min（可配置）。每次 Feed 页返回前 MUST SADD 本页 postIds；候选集 MUST 排除 session 中已见 postId。

#### Scenario: 同 session 翻页无重复

- **WHEN** 客户端用合法 `nextCursor` 连续请求两页且 session 未过期
- **THEN** 两页 `list` 中 postId MUST 互斥（无交集）

#### Scenario: 刷新生成新 session

- **WHEN** 客户端请求 Feed 且不传 `cursor`
- **THEN** 系统 MUST 生成新 `sessionId` 且 MUST NOT 复用旧 session 的 seen SET

---

## ucg-following-feed

<!-- source: openspec/specs/ucg-following-feed/spec.md -->

# ucg-following-feed Specification

## Purpose
TBD - created by archiving change ucg-following-feed-snapshot. Update Purpose after archive.
## Requirements
### Requirement: 关注 Feed SHALL 经 MySQL 分页取 postId 并由 Redis snapshot 组装展示

关注 Feed 读路径 MUST 按下列顺序执行：

1. 从 MySQL `ucg_follow` 读取当前用户 followee `wxId` 列表；无关注时 MUST 返回空 `list` 与 `total=0`。
2. 从 MySQL `ucg_post` 查询 `status=2`（published）且 `author_wx_id IN (followees)` 的帖子，**MUST** 按 `published_at DESC` 排序，**MUST** 使用 `page`/`pageSize` 分页并返回准确 `total`。
3. DB 查询 MUST 仅获取组装所需 postId（及排序字段），**MUST NOT** 在关注 Feed 读路径 JOIN profile、like 或媒体表。
4. 对当前页 postId 列表 MUST 调用与推荐 Feed 共用的 snapshot 组装逻辑（post snapshot、profile snapshot、liked SET batch），**MUST NOT** 使用 `postsFromResult` 全量 MySQL 组装。
5. snapshot miss MUST best-effort 调用 `backfillPostSnapshot` 回源 MySQL 并回填 Redis（与推荐 Feed 相同语义）。
6. 本读路径 **MUST NOT** 新增 Redis 键（不得创建 author ZSET、followees SET 或 following session）。
7. 本读路径 **MUST NOT** 使用 cursor、`nextCursor`、`hasMore` 替代 `{ total, page, pageSize }`。
8. 本读路径 **MUST NOT** 使用 GEO 半径筛选或 composite score 改变关注 Feed 排序；排序权威来源 MUST 仍为 MySQL `published_at DESC`。

#### Scenario: 有关注时分页返回 published 帖

- **WHEN** 已登录用户 `GET /ucg/app/api/feed/following?page=1&pageSize=20` 且其关注的人中有 published 帖
- **THEN** 响应 MUST 含 `{ list, total, page, pageSize }`
- **AND** `list` 中每条 MUST 为 `status=2` 且作者 MUST 在当前用户 followee 集合内
- **AND** `list` 顺序 MUST 按 `published_at` 降序

#### Scenario: 组装走 Redis snapshot 而非帖表 N+1

- **WHEN** 关注 Feed 返回非空 `list` 且对应 post snapshot 已存在于 Redis
- **THEN** 系统 MUST 从 post/profile snapshot 与 liked SET 填充 item 字段
- **AND** MUST NOT 对该页帖子逐条 SELECT `ucg_post` 完整行用于展示组装

#### Scenario: snapshot miss 回源

- **WHEN** 关注 Feed 当前页某 postId 在 Redis 无 post snapshot
- **THEN** 系统 MUST 尝试 `backfillPostSnapshot` 回源并回填
- **AND** 回填成功时该帖 SHOULD 出现在 `list` 中

#### Scenario: 无关注时空列表

- **WHEN** 已登录用户未关注任何人
- **THEN** 响应 MUST 为 `{ list: [], total: 0, page, pageSize }`
- **AND** MUST NOT 查询 `ucg_post` 全表

#### Scenario: 不使用 Redis 关注 timeline

- **WHEN** 实现关注 Feed 读路径
- **THEN** 系统 MUST NOT 使用 ZUNIONSTORE 或 per-author 帖 ZSET 作为分页来源
- **AND** MUST NOT 新增 followees SET 或 author timeline 键

### Requirement: 关注 Feed SHALL 可选计算并返回 distanceMeters

当 viewer 请求含有效 `lat`/`lng` 且帖子 snapshot（或回源后）含坐标时，关注 Feed item MUST 含 `distanceMeters` 字符串（米，纯数字字符串）；否则 MUST 省略该字段。距离计算 MUST 复用推荐 Feed 的 haversine 逻辑，**MUST NOT** 将 post 坐标下发客户端。

#### Scenario: 关注 Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/following?lat=31.2&lng=121.5&page=1` 且某帖 snapshot 含坐标
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 无 viewer 坐标省略距离

- **WHEN** `GET /ucg/app/api/feed/following?page=1` 不含 lat/lng
- **THEN** 响应 item MUST NOT 含 `distanceMeters`

#### Scenario: 帖无坐标省略距离

- **WHEN** viewer 含 lat/lng 但帖无坐标
- **THEN** 该 item MUST NOT 含 `distanceMeters`

---

## ucg-force-first-grant

<!-- source: openspec/specs/ucg-force-first-grant/spec.md -->

# ucg-force-first-grant Specification

## Purpose
TBD - created by archiving change feature-ops-invite-force-admin. Update Purpose after archive.
## Requirements
### Requirement: 原力首次增量可写入空账户

当 `ucg_user_force` 尚无该 `wx_id` 行时，系统执行原力增量（含获客 `invite_acquisition` +100 与辩论自投 +1）MUST 成功插入余额行并写入 `ucg_force_ledger`，MUST NOT 因空结果集 `Scan`/`sql.ErrNoRows` 失败。已有行时 MUST 在事务内累加并写流水。

#### Scenario: 首次获客加分

- **WHEN** cash 通知 ucg 对无原力行的码主人执行获客加分且入参合法
- **THEN** 该用户 `force_value` 为 100（或既有 0+100），且 ledger 存在 `invite_acquisition` 记录

#### Scenario: 已有余额再加分

- **WHEN** 用户已有 `force_value=100` 再次获客成功加分
- **THEN** `force_value` 变为 200 且新增一条 ledger

### Requirement: 我的资料返回原力值

`GET` 当前用户 profile（如 `/ucg/app/api/profile/me`）成功时 MUST 填充与本域 `ucg_user_force` 一致的 `forceValue`（无行视为 0），MUST NOT 因未 enrich 而长期省略真实非零原力。

#### Scenario: 有原力时 me 可见

- **WHEN** 用户 `ucg_user_force.force_value=100` 且请求我的资料
- **THEN** 响应 MUST 含 `forceValue` 为 100

---

## ucg-force-ledger

<!-- source: openspec/specs/ucg-force-ledger/spec.md -->

# ucg-force-ledger Specification

## Purpose
TBD - created by archiving change invite-peer-force-ucg. Update Purpose after archive.
## Requirements
### Requirement: 原力存储在 ucg 域

系统 MUST 在 ucg 数据库维护每用户原力当前值与积分流水。系统 MUST NOT 再以 device `wx.force_value` 作为原力权威。作者对自己辩论文章投票成功时，系统 MUST 原力 +1 并写入 reason=`debate_self_vote` 的流水。系统 MUST NOT 回填本变更上线前的历史投票流水。

#### Scenario: 辩论加分

- **WHEN** 作者自投成功
- **THEN** ucg 原力 +1 且 ledger 新增一条 `debate_self_vote`

### Requirement: 获客加原力

当邀请码被他人成功兑换一次时，系统 MUST 将码主人获客计数 +1，并经 cash→ucg 契约为码主人原力 +100，ledger reason=`invite_acquisition`。原力加分失败 MUST NOT 撤销已成功的功能开通；系统 MUST 记录可观测错误以便补偿。

#### Scenario: 兑码触发获客积分

- **WHEN** 用户 B 成功使用用户 A 的邀请码
- **THEN** A 的 `redeemed_count` +1，且 A 的原力尝试 +100 并在成功时写入 `invite_acquisition` 流水

### Requirement: App 可读原力与流水

系统 MUST 在 UCG App 可读路径提供当前 `forceValue`（可经 profile 或专用接口）与积分流水分页。系统 MUST NOT 要求返回「距下一档所需积分」；档位展示由客户端计算。

#### Scenario: 流水列表

- **WHEN** 用户请求 force ledger
- **THEN** 返回含 reason、delta、时间的记录，新产生的获客与辩论分均可见

---

## ucg-gateway-proxy

<!-- source: openspec/specs/ucg-gateway-proxy/spec.md -->

# ucg-gateway-proxy Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: gateway-app SHALL HTTP-proxy /ucg/app/api to ucg-service

gateway-app-server SHALL register reverse proxy for path prefix `/ucg/app/api/*` to configured `UCG_SERVICE_BASE_URL`, applying existing Bearer JWT validation and injecting `X-Internal-Wx-Id` from JWT `sub` before forwarding. CORS behavior SHALL match other domain proxies.

#### Scenario: 鉴权后转发
- **WHEN** App 带合法 Bearer 请求 `/ucg/app/api/profile/me`
- **THEN** gateway SHALL 转发至 ucg-service 且 SHALL 设置 `X-Internal-Wx-Id`

#### Scenario: 推荐接口匿名可读
- **WHEN** 产品配置推荐 Feed 为匿名可读且请求在白名单内
- **THEN** gateway SHALL 允许无 Bearer 转发 `/ucg/app/api/feed/recommend`（若实现匿名策略）

### Requirement: gateway-app SHALL WebSocket-proxy /ucg/app/ws/chat to ucg-service

gateway-app-server SHALL register WebSocket upgrade reverse proxy for exact path `/ucg/app/ws/chat` to ucg-service internal endpoint `/ws/chat`, using the same `httputil.ReverseProxy` pattern as `ws_route_proxy.go` / voice WS edge proxy. Configuration SHALL use `UCG_WS_ROUTE_MODE` and `UCG_WS_PROXY_URL`. App clients MUST NOT connect directly to ucg-service for chat.

#### Scenario: WS 经网关同域
- **WHEN** 客户端连接 `wss://{apiBaseUrl host}/ucg/app/ws/chat`
- **THEN** gateway SHALL 透传至 ucg-service `/ws/chat`，且 SHALL NOT 要求 App 配置独立 ucg-service 公网 WS 域名

#### Scenario: WS 代理目标不可达
- **WHEN** `UCG_WS_ROUTE_MODE=proxy` 且 ucg-service WS 不可达或握手失败
- **THEN** gateway SHALL 返回可诊断的 `ws_proxy` 阶段错误，且 SHALL NOT 静默成功

---

## ucg-green-audit

<!-- source: openspec/specs/ucg-green-audit/spec.md -->

# ucg-green-audit Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: Post and profile content SHALL use Green async audit Option A

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter pending visibility: ONLY author MAY see content in feeds/profile until Green pass sets published or profile active state; on fail content SHALL be rejected with reason visible to author as「违规已下架」或等价文案。**触发方式 MUST 为事务提交后写入 `ucg_audit_publish_outbox` 并经 relay HTTP Publish RabbitMQ 事件（载荷含冻结 `auditVersion`），由 ucg-service AMQP consumer 异步执行 Green；MUST NOT 依赖定时扫 pending 业务表 worker 或 reconciler 发现漏审条目。** 全链路 MUST 遵循各实体 `audit_version` 权威列与 CAS 规则（见 data-model / audit-mq 规格）。

#### Scenario: 发帖不再依赖 audit_worker 或 pending reconciler

- **WHEN** 用户 submit 发帖且 outbox 与 relay 正常
- **THEN** 系统 MUST Publish `ucg.post.created`（含 `auditVersion`）且 MUST NOT 依赖 `StartUcgAuditReconciler` 或 audit_worker 扫表触发首次审核

#### Scenario: Publish 失败由 outbox relay 恢复

- **WHEN** 帖子已 pending 且 outbox 行 Publish  initially 失败
- **THEN** relay worker MUST 重试 Publish；MUST NOT 通过扫描 `ucg_post.status=1` 补发

### Requirement: Green text rejection reason SHALL expose human-readable riskTips

当 Green **文本**机审判定违规（`Labels` 非空且非 `nonLabel`）时，系统 MUST 将 `AuditVerdict.Reason` 设为用户可读文案：**若 API `Data.Reason` 为 JSON 字符串，MUST 解析并取 `riskTips` 字段**；若 `Reason` 为纯文本，MUST 使用该纯文本；若解析后仍为空，MUST 使用默认文案「违规已下架」（或代码中等价常量）。系统 MUST NOT 将 `Data.Reason` 的原始 JSON 字符串直接写入 `reject_reason` 或经 App/WS 暴露给作者。

该规则 MUST 适用于所有经 `ModerateText` / `parseTextModeration` 的实体（帖子、评论、资料 job、私信）。图片/视频机审路径不在本 Requirement 范围内。

#### Scenario: Reason 为 JSON 时展示 riskTips

- **WHEN** Green 文本审核返回 `Labels` 命中且 `Data.Reason` 为 `{"riskTips":"命中违禁内容","riskLevel":"high"}`
- **THEN** 机审 consumer 写入的 `reject_reason`（及 WS `audit_failed` 的 reason）MUST 为 `命中违禁内容`，MUST NOT 为完整 JSON 字符串

#### Scenario: Reason 为纯文本时原样使用

- **WHEN** Green 返回 `Data.Reason` 为纯文本 `内容不合规`
- **THEN** `AuditVerdict.Reason` MUST 为 `内容不合规`

#### Scenario: JSON 无 riskTips 时回退默认文案

- **WHEN** Green 返回 `Data.Reason` 为 `{"riskLevel":"high"}` 且无有效 `riskTips`
- **THEN** `AuditVerdict.Reason` MUST 为「违规已下架」（或等价默认文案）

#### Scenario: 帖子作者可见修正后的驳回原因

- **WHEN** 用户帖子文本机审失败且 CAS 驳回成功
- **THEN** 作者「我的动态」接口返回的 `rejectReason` MUST 为上述可读文案，MUST NOT 含 JSON 花括号包裹的原始 Reason

### Requirement: Green image and video moderation dataId SHALL comply with Alibaba constraints

当 `ucg.green.enabled` 为 true 且系统调用 `GreenModerator.ModerateImageURL` 或 `ModerateVideoURL` 时，若向阿里云 Green 传入 `ServiceParameters.dataId`，该值 MUST 满足：长度不超过 64 字符；字符集仅含大小写英文字母、数字、下划线（`_`）、短划线（`-`）、英文句号（`.`）。系统 MUST NOT 将完整 HTTP(S) URL（含 scheme、`:`、`/` 等）作为 `dataId`。

系统 SHOULD 从媒体 CDN URL 的 path 部分推导合规 `dataId`（例如将 objectKey 中的 `/` 规范为 `_`）。若无法推导合规值，MUST 省略 `dataId` 字段而非传入非法值。

当 Green 返回 `body.Code != 200` 或 HTTP 非 200 时，解析层 MUST 在 error 中包含 business `Code`（及 `Msg` 若 API 提供），以便运维区分参数错误与额度/限流等故障。

#### Scenario: 帖子图片审核使用合规 dataId

- **WHEN** 用户提交带 `social/` 前缀 objectKey 的图片帖且 Phase1 调用 `ModerateImageURL`
- **THEN** Green `ImageModeration` 请求的 `ServiceParameters` MUST NOT 含完整 CDN URL 作为 `dataId`；若含 `dataId` 则 MUST 为规范化 object path（如 `social_2026_06_xxx.jpg`）且长度 ≤64

#### Scenario: 资料头像与私信媒体同步约束

- **WHEN** 资料审核 job 调用头像 `ModerateImageURL`，或私信调用 `ModerateImageURL` / `ModerateVideoURL`
- **THEN** `dataId` 约束 MUST 与帖子媒体相同，MUST NOT 使用完整 URL

#### Scenario: Green 参数错误可观测

- **WHEN** Green 图片 API 因参数校验返回 `body.Code != 200`
- **THEN** ucg-service 日志中的 error MUST 包含 `green image: code <n>` 及 Msg（若存在），MUST NOT 仅返回无 code 的泛化文案

#### Scenario: 合规 dataId 下图片 Phase1 可完成

- **WHEN** 新发图片帖且 Green 配置有效、CDN URL 公网可访问
- **THEN** Phase1 MUST 成功发起 `baselineCheck` 并持久化 `moderation_verdict`（pass 或 reject），MUST NOT 因非法 `dataId` 直接进入 `moderation_failed`（status=5）

### Requirement: Profile patch SHALL use MySQL audit job and MQ instead of Redis pending queue

资料变更（nickname/avatar/bio）MUST 写入 MySQL `ucg_profile_audit_job` 并 Publish `ucg.profile.patch.submitted`。入队前 MUST 将请求字段与 **`ucg_profile` 已发布行**对比，**仅**将相对已发布值发生变化的非空字段写入 job；若 nickname、avatarKey、bio 均无实质变更，MUST NOT enqueue 且 MUST NOT 触发 Green。Consumer 机审时 SHOULD 跳过 job 中与已发布 profile 相同的字段（兼容历史全量 job 消息）。

#### Scenario: 仅 bio 变更

- **WHEN** 作者 PUT profile 携带未改 nickname 与新 bio
- **THEN** job MUST 仅含 bio 非空；Green MUST NOT 调用 `nickname_detection` 审该 nickname

#### Scenario: 全量相同无变更

- **WHEN** PUT 三字段均与已发布 profile 相同
- **THEN** MUST NOT 创建新 audit MQ 消息；HTTP MAY 返回 200 与当前作者 profile

### Requirement: Post status updates SHALL use CAS with audit_version

对 `ucg_post` 的机审与管理端审态变更（publish/reject）MUST 使用条件更新：`WHERE id=? AND status=? AND audit_version=?`。CAS 成功时 MUST 更新 `status`（及 `reject_reason` 等），**MUST NOT** 递增 `audit_version`。MUST NOT 存在无条件 `UPDATE status` 的机审路径。

#### Scenario: 机审通过 CAS

- **WHEN** consumer 审核 postId=1，载荷 `auditVersion=2`，且当前 `status=1`、`audit_version=2`
- **THEN** UPDATE MUST 使用 `status=1 AND audit_version=2` 条件；成功后 `status=2`，`audit_version` MUST 仍为 2

#### Scenario: 再提审递增版本

- **WHEN** 作者对已发布或驳回帖 submit 再提审
- **THEN** 系统 MUST 将 `status` 置 1 且 `audit_version` MUST 递增，并 Publish 新 `auditVersion`；帖文业务字段 MUST NOT 因再提审而 reset

#### Scenario: 过期 post 消息跳过

- **WHEN** 再提审后行内 `audit_version=3`，consumer 收到 `auditVersion=2` 的旧消息
- **THEN** CAS MUST 影响 0 行且 MUST ACK，MUST NOT 覆盖 version=3 的 pending 状态

### Requirement: Comments SHALL use Green async audit before public visibility

用户发表评论时 MUST 写入 `status=1`（pending_audit）与 `audit_version`（首评 1），MUST Publish `ucg.comment.created`（含 `auditVersion`）；Green 通过且 CAS（`WHERE status=1 AND audit_version=?`）成功后 status=2，评论 SHALL 出现在评论列表且 MAY 触发 `comment_count` 递增与通知；失败则 status=3，仅作者可见违规信息。CAS 成功 MUST NOT 递增 `audit_version`。

#### Scenario: 评论待审不对公众展示

- **WHEN** 用户发表评论且 Green 未完成
- **THEN** `GET .../comments` 响应 MUST NOT 含该条（其他用户视角）；作者 MAY 在响应或单独字段看到审核中

#### Scenario: 评论审核通过后计数

- **WHEN** Green pass 且 CAS 将评论置 published
- **THEN** 帖子 `comment_count` MUST 递增且 MAY 触发评论通知

### Requirement: Chat messages SHALL use post-delivery async Green audit (Mode A)

私信发送后 MUST 立即向收件人 WS 投递消息（`audit_status=pending`），MUST 写入 Redis 并增加收件人未读；MUST Publish `ucg.chat.msg.created`（含 `auditVersion`，权威为 `ucg_chat_message.audit_version`）异步 Green。pending 期间收件人 MUST 可见该消息。Green pass MUST CAS `audit_status` 从 pending 到 approved（`WHERE audit_status='pending' AND audit_version=?`）；Green fail MUST CAS 为 rejected，且 MUST 仅发送方可见并含违规提示，收件人 MUST NOT 在历史与列表中看到 rejected 消息。CAS 成功 MUST NOT 递增 `audit_version`。收件人未读在投递时 +1，reject MUST NOT 回滚未读。

#### Scenario: pending 期间收件人可见

- **WHEN** 用户发送聊天消息且 Green 未完成
- **THEN** 收件人 WS MUST 已收到 `message_delivered` 且拉取历史 MUST 含该条（非 rejected）

#### Scenario: 事后驳回仅发送方可见

- **WHEN** Green fail 且 CAS 为 rejected
- **THEN** 发送方 MUST 收到含 reason 的 `audit_failed`（或等价事件）且列表仍可见该条；收件方 MUST NOT 见该条，在线时 SHOULD 收到 `msg_hidden`

#### Scenario: 未读不回滚

- **WHEN** 消息已投递并已对收件人 unread+1，随后 Green reject
- **THEN** 收件人 unread_count MUST NOT 因 reject 减少

#### Scenario: 异步审核经 MQ

- **WHEN** 聊天消息已写入 Redis
- **THEN** 系统 MUST Publish `ucg.chat.msg.created`（含 `auditVersion`）且 MUST NOT 在 WS handler 内同步阻塞 Green

#### Scenario: 过期 chat 消息 CAS 跳过

- **WHEN** 消息已 CAS 为 approved 或 rejected，重复 MQ 消息携带旧 `auditVersion` 到达
- **THEN** CAS MUST 影响 0 行且 MUST ACK，MUST NOT 翻转审态

---

## ucg-image-thumb

<!-- source: openspec/specs/ucg-image-thumb/spec.md -->

# ucg-image-thumb Specification

## Purpose
TBD - created by archiving change ucg-image-thumb-physical. Update Purpose after archive.
## Requirements
### Requirement: Image thumb suffix SHALL be globally defined as _thumb before extension

平台 MUST 在 `internal/shared/mediacdn` 定义缩略图后缀常量（值为 `_thumb`）及 `ThumbObjectKey` helper。任意服务与脚本 MUST 经该 helper 派生缩略图 objectKey，MUST NOT 在业务代码中散落 `_thumb` 字面量。

对原图 objectKey `{path}/{stem}.{ext}`，缩略图 objectKey MUST 为 `{path}/{stem}_thumb.{ext}`，且 `{ext}` MUST 与原图扩展名一致（如 `abc.jpg` → `abc_thumb.jpg`，`abc.png` → `abc_thumb.png`）。

`IsThumbObjectKey` MUST 识别已为缩略图的 key；对这类 key 调用 `ThumbObjectKey` MUST 原样返回，MUST NOT 产生 `_thumb_thumb` 双后缀。

#### Scenario: JPG 原图派生 thumb key

- **WHEN** 原图 objectKey 为 `social/2026/06/xyz.jpg`
- **THEN** `ThumbObjectKey` SHALL 返回 `social/2026/06/xyz_thumb.jpg`

#### Scenario: PNG 扩展名保持一致

- **WHEN** 原图 objectKey 为 `social/2026/06/xyz.png`
- **THEN** 缩略图 objectKey SHALL 为 `social/2026/06/xyz_thumb.png`

### Requirement: EnsureImageThumb SHALL create idempotent physical thumb objects via OSS process

ucg-service MUST 提供 `EnsureImageThumb(ctx, objectKey)`：对图片原图 objectKey，若缩略图对象不存在，MUST 经 OSS `GetObject` 携带图片处理参数（`auto-orient,1/resize,m_lfit,w_200/quality,q_90`，并按扩展名保持输出格式）获取字节后 `PutObject` 至 `ThumbObjectKey(objectKey)`；若缩略图已存在 MUST 跳过（幂等）。原图不存在时 MUST 返回明确错误。

处理参数 MUST NOT 在返回给客户端的 CDN URL 中出现；仅用于服务端生成物理对象。

#### Scenario: 首次生成 thumb

- **WHEN** 原图 `social/.../a.jpg` 存在于 OSS 且 `social/.../a_thumb.jpg` 不存在
- **THEN** `EnsureImageThumb` SHALL 上传 `a_thumb.jpg` 且后续 `HEAD` 可命中

#### Scenario: 重复调用幂等

- **WHEN** `a_thumb.jpg` 已存在
- **THEN** `EnsureImageThumb` SHALL 成功返回且 MUST NOT 覆盖已有对象

### Requirement: BuildImageThumbnailURL SHALL return physical thumb CDN URL without x-oss-process

`BuildImageThumbnailURL(objectKey)` MUST 返回 `BuildCdnURL(ThumbObjectKey(objectKey))`。图片缩略图 CDN URL MUST NOT 包含 `x-oss-process` query 参数。

视频首帧 `BuildVideoSnapshotURL` 本 capability 不修改，MAY 仍使用 `x-oss-process`。

#### Scenario: 图片列表缩略图 URL 无 query

- **WHEN** 服务为图片 objectKey 拼装 `thumbnailUrl`
- **THEN** URL path SHALL 以 `_thumb.{ext}` 结尾且 SHALL NOT 含 `x-oss-process`

### Requirement: Media deletion SHALL remove paired thumb objects

当 ucg-service 删除用户拥有的图片 OSS 原图对象时，MUST 同时尝试删除 `ThumbObjectKey(原图 key)`；thumb 对象不存在时 MUST NOT 导致整次删除失败。

#### Scenario: 删除原图同时删除 thumb

- **WHEN** `DeleteOwnedMedia` 删除 `social/.../a.jpg` 且 blob 允许删 OSS
- **THEN** OSS 上 `social/.../a_thumb.jpg` SHALL 被删除或已不存在

### Requirement: Backfill CLI SHALL populate historical image thumbs before read-path cutover

仓库 MUST 提供 `cmd/ucg-image-thumb-backfill`：从 UCG 数据库收集去重图片 objectKey（至少含 `ucg_media_blob` media_kind=1、`ucg_post_media` 图片、`ucg_profile.avatar_key`、`ucg_chat_message.image_key`），对每条调用与线上一致的 `EnsureImageThumb`（或等价逻辑）。

CLI MUST 支持 `--dry-run`（仅打印将处理 key）、`--limit`、`--concurrency`；执行结束 MUST 输出 ok/skipped/missing/failed 汇总。运维 runbook MUST 规定：各环境 backfill 验证通过后再部署读路径切换。

#### Scenario: dry-run 不写 OSS

- **WHEN** 运维执行 `go run ./cmd/ucg-image-thumb-backfill --dry-run`
- **THEN** 脚本 SHALL 列出将处理的 key 且 MUST NOT 调用 `PutObject`

#### Scenario: 漏网 key 可重跑

- **WHEN** 首次 backfill 部分失败
- **THEN** 再次执行脚本 SHALL 跳过已成功 thumb 并仅处理剩余失败项

---

## ucg-internal-profile-batch

<!-- source: openspec/specs/ucg-internal-profile-batch/spec.md -->

# ucg-internal-profile-batch Specification

## Purpose
TBD - created by archiving change sim-admin-user-list-deregister. Update Purpose after archive.
## Requirements
### Requirement: ucg internal SHALL batch query public profiles by wxIds

`ucg-service` MUST 提供 `POST /ucg/internal/api/profiles/batch`，鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。

请求体 MUST 含 `wxIds`（int64 数组，允许空）。响应 MUST 含 `list` 数组；对请求体中**每个有效 wxId（>0）** MUST 在 `list` 中返回**恰好一条**记录。每项 MUST 含：

- `wxId`（uint64 或 int64）
- `nickname`（string，展示用昵称）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）

当 wxId 存在 `ucg_profile` 行时，`nickname` MUST 与 `GetPublicProfilesByWxIDs` 语义一致（含默认昵称刷新逻辑）。当 wxId **不存在** `ucg_profile` 行时，实现 MUST 经 device 契约取得 `babyName` 并 MUST 使用与 App 一致的推导规则：`babyName` 为空时 `nickname` 为「家长」，否则为 `{babyName}的家长`；**MUST NOT** 为此路径写入 `ucg_profile`。MUST NOT 返回 unionid 等敏感字段。

#### Scenario: Batch returns profile nickname for existing profile

- **WHEN** 受信调用方 POST `{ "wxIds": [1001] }` 且 1001 有 ucg_profile
- **THEN** 响应 `list` MUST 含 wxId=1001 的 nickname

#### Scenario: Batch synthesizes nickname without profile row

- **WHEN** 受信调用方 POST `{ "wxIds": [2002] }` 且 2002 无 ucg_profile 但 device 返回 babyName=`"小宝"`
- **THEN** 响应 `list` MUST 含 wxId=2002 且 nickname=`"小宝的家长"`

#### Scenario: Batch synthesizes default parent nickname

- **WHEN** 受信调用方 POST `{ "wxIds": [2003] }` 且 2003 无 ucg_profile 且 babyName 为空
- **THEN** 响应 `list` MUST 含 wxId=2003 且 nickname=`"家长"`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 profiles batch
- **THEN** HTTP MUST 为 403

#### Scenario: Empty wxIds

- **WHEN** POST `{ "wxIds": [] }`
- **THEN** 响应 `list` MUST 为空数组且 MUST NOT 500

---

## ucg-media-server-upload-only

<!-- source: openspec/specs/ucg-media-server-upload-only/spec.md -->

# ucg-media-server-upload-only Specification

## Purpose
TBD - created by archiving change ucg-oss-caidoukeji-server-upload. Update Purpose after archive.
## Requirements
### Requirement: All app clients MUST upload media via server multipart API

UCG 媒体上传主路径 MUST 为服务端 multipart 上传（`POST /ucg/app/api/media/upload` 或经 gateway 同域代理的等价路径）。iOS/Android 客户端 MUST 与 Web 一样走该路径，MUST NOT 再依赖「预签名 URL + 客户端直传 OSS」作为主上传流程。

#### Scenario: 原生上传经服务端

- **WHEN** 原生 App 上传图片或视频且 resolve 未命中去重
- **THEN** 客户端 SHALL 将文件以 multipart 提交至服务端上传接口，SHALL NOT PUT 至 OSS 预签名 URL

### Requirement: Media presign MUST NOT issue client-usable upload URLs

`POST /ucg/app/api/media/presign`（若路由仍保留）MUST 拒绝签发可供客户端直传的预签名上传 URL（返回明确错误）。服务端 MUST NOT 向客户端返回指向 OSS 内网域名的上传地址。

#### Scenario: 调用 presign 被拒绝

- **WHEN** 客户端请求 media presign
- **THEN** 接口 SHALL 失败并提示改用服务端 upload，且 SHALL NOT 返回可用的 uploadUrl

### Requirement: Historical pang-bao objects are out of scope for dual-read

系统 MUST NOT 实现运行时回退读取旧桶 `pang-bao`。历史可见性依赖运维将对象以相同 objectKey 拷贝到 `caidoukeji` 且 CDN 已指向新桶。

#### Scenario: 同 objectKey 在新桶可读

- **WHEN** 数据库中已有 objectKey 且该对象已按相同 key 存在于 `caidoukeji`，CDN 指向新桶
- **THEN** 既有展示 URL（cdnBaseUrl + objectKey）SHALL 可访问，无需改库

---

## ucg-oss-caidoukeji-config

<!-- source: openspec/specs/ucg-oss-caidoukeji-config/spec.md -->

# ucg-oss-caidoukeji-config Specification

## Purpose
TBD - created by archiving change ucg-oss-caidoukeji-server-upload. Update Purpose after archive.
## Requirements
### Requirement: UCG OSS MUST use caidoukeji bucket on Hangzhou internal endpoint

`ucg-service` 的 OSS 配置 MUST 使用 bucket `caidoukeji`、region `cn-hangzhou`。服务端 SDK endpoint MUST 为 `oss-cn-hangzhou-internal.aliyuncs.com`（不含 bucket 虚拟主机前缀）。AccessKey MUST 继续经 `UCG_OSS_ACCESS_KEY_*` 注入，yaml 明文 MUST 留空。展示 URL MUST 继续基于配置的 `cdnBaseUrl` 与 objectKey 拼接。系统 MUST NOT 再将 `pang-bao` 或 `oss-cn-beijing.aliyuncs.com` 作为默认生产配置。

#### Scenario: 默认配置指向新桶内网

- **WHEN** 使用仓库默认 `config.ucg-service.yaml` 且未设置 endpoint 覆盖 env
- **THEN** `LoadOSSConfig`（或等价）SHALL 得到 bucket=`caidoukeji` 且 endpoint 为杭州内网主机

#### Scenario: 非 VPC 可用 env 覆盖 endpoint

- **WHEN** 开发环境设置了约定的 OSS endpoint 覆盖变量为公网杭州 endpoint
- **THEN** 服务端 OSS 客户端 SHALL 使用该覆盖值连接（便于本地联调）

### Requirement: Server-side OSS operations MUST use configured endpoint only

服务端 `PutObject`、删除、缩略图、转码读写、校验下载等 MUST 使用上述配置的 endpoint/bucket，MUST NOT 硬编码旧桶名。

#### Scenario: 媒体服务端上传写入新桶

- **WHEN** 调用 `POST /ucg/app/api/media/upload`（经网关）成功上传文件
- **THEN** 对象 SHALL 写入 `caidoukeji`，响应 SHALL 含 objectKey 与基于 cdnBaseUrl 的 cdnUrl

---

## ucg-oss-presign

<!-- source: openspec/specs/ucg-oss-presign/spec.md -->

# ucg-oss-presign Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: OSS presign SHALL use pang-bao bucket with social/ prefix

ucg-service SHALL provide presigned upload for bucket `pang-bao`, region `cn-beijing`, endpoint `oss-cn-beijing.aliyuncs.com`, generating objectKey with prefix `social/`. Database and API DTOs MUST store objectKey only; CDN display URL is `https://resorce.cuplay.top/{objectKey}`.

OSS AccessKey credentials MUST NOT be stored as plaintext in `manifest/config/config.ucg-service.yaml`. The yaml fields `ucg.oss.accessKeyId` and `ucg.oss.accessKeySecret` MUST be empty in the repository. At runtime, credentials MUST be supplied via environment variables `UCG_OSS_ACCESS_KEY_ID` and `UCG_OSS_ACCESS_KEY_SECRET` (typically injected through `manifest/docker/.env.*` and Docker Compose).

#### Scenario: 获取 presign

- **WHEN** 客户端 `POST /ucg/app/api/media/presign` with media kind and extension，且 ucg-service 已通过 env 配置有效 OSS 凭证
- **THEN** 响应 SHALL 含 uploadUrl、objectKey（以 `social/` 开头），且 SHALL NOT 要求客户端自定义 bucket

#### Scenario: DB 仅存 objectKey

- **WHEN** 帖子媒体写入 `ucg_post_media`
- **THEN** 行 SHALL 仅保存 objectKey 字段，且 SHALL NOT 保存完整 CDN URL

#### Scenario: 未配置 OSS 凭证时 presign 失败

- **WHEN** 容器 env 与 yaml 均未提供有效 `UCG_OSS_ACCESS_KEY_*`
- **THEN** presign 接口 SHALL 返回明确错误且 SHALL NOT 使用硬编码或 yaml 明文 fallback

### Requirement: Image register and server upload SHALL ensure physical thumb exists

当 `RegisterMedia` 成功登记**新**图片 blob（原图已在 OSS）或 `putOSSObject` 成功上传图片（`mediaKind=1`）后，ucg-service MUST 同步调用 `EnsureImageThumb` 生成物理缩略图对象。`EnsureImageThumb` 失败时 register/直传 MUST 返回错误，MUST NOT 仅登记原图而无 thumb。

`PresignUpload` MUST NOT 生成缩略图（客户端尚未完成 PUT）。

dedup hit 路径 MAY 依赖已有 thumb；`EnsureImageThumb` 幂等调用 MUST 可接受。

#### Scenario: 新图 register 成功后 OSS 有成对对象

- **WHEN** 客户端完成原图 PUT 且 `RegisterMedia` 成功登记新图片 blob
- **THEN** OSS MUST 同时存在原图 objectKey 与对应 `ThumbObjectKey`

#### Scenario: 服务端直传图片后生成 thumb

- **WHEN** `UploadMediaObject` 成功上传 `mediaKind=1` 文件
- **THEN** OSS MUST 存在对应 `_thumb.{ext}` 对象

#### Scenario: thumb 生成失败阻止 register

- **WHEN** 原图存在但 `EnsureImageThumb` 失败
- **THEN** `RegisterMedia` SHALL 返回错误且 MUST NOT 仅完成 blob 登记而无 thumb

### Requirement: Video register and server upload SHALL ensure physical first-frame thumb exists

当 `RegisterMedia` 成功登记**新**视频 blob（原视频已在 OSS，`mediaKind=2`）或 `putOSSObject` 成功上传视频（`mediaKind=2`）后，ucg-service MUST 同步调用 `EnsureVideoThumb` 生成物理首帧缩略图对象（`{stem}_thumb.jpg`）。`EnsureVideoThumb` 失败时 register/直传 MUST 返回错误，MUST NOT 仅登记原视频而无 thumb。

`PresignUpload` MUST NOT 生成缩略图（客户端尚未完成 PUT）。

dedup hit 路径 MAY 依赖已有 thumb；`EnsureVideoThumb` 幂等调用 MUST 可接受。

#### Scenario: 新视频 register 成功后 OSS 有成对对象

- **WHEN** 客户端完成 mp4 PUT 且 `RegisterMedia` 成功登记新视频 blob
- **THEN** OSS MUST 同时存在原视频 objectKey 与对应 `{stem}_thumb.jpg`

#### Scenario: 服务端直传视频后生成 thumb

- **WHEN** 服务端成功上传 `mediaKind=2` 视频文件
- **THEN** OSS MUST 存在对应 `_thumb.jpg` 对象

#### Scenario: thumb 生成失败阻止 register

- **WHEN** 原视频存在但 `EnsureVideoThumb` 失败
- **THEN** `RegisterMedia` SHALL 返回错误且 MUST NOT 仅完成 blob 登记而无 thumb

### Requirement: Video media SHALL use transformVersion v1 or v2 with validation gate

视频 blob 索引 MUST 使用 `transformVersion` 区分管线产出：

- `v2`：canonical（H.264 Main + AAC + faststart）；原生客户端转码与 sim 服务端转码产出 MUST 使用 `v2` register
- `v1`：Web Phase 1 宽验真直传；**非 canonical**（可 Baseline、可无 faststart），但 **必须有 AAC 音轨**

`contentHash` MUST 为 OSS 上最终对象字节的 SHA-256 hex 小写。`v1` 与 `v2` MUST NOT 跨版本 dedup 命中（键为 `(contentHash, transformVersion)`）。

新写入 MUST NOT 使用 `sim-raw` 或其他未登记版本。

#### Scenario: v1 and v2 blobs do not dedup each other

- **WHEN** 同一源文件分别作为 v1 直传与 v2 转码后 contentHash 不同或 version 不同
- **THEN** `media/resolve` MUST NOT 跨 v1/v2 返回 hit

### Requirement: Web video upload response SHALL include contentHash for register

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 且上传成功时，响应 `data` MUST 含：

- `contentHash`：64 位小写 hex SHA-256，对 **OSS 上最终对象字节** 计算（v1 直传为原始字节；服务端转码为转码后字节）
- `transformVersion`：字符串 `v1` 或 `v2`，指示客户端 `RegisterMedia` 应使用的版本（v1 直传路径为 `v1`；服务端转码路径为 `v2`）

Web 客户端与 Flutter Web（经 gateway 同域 upload）MUST 使用响应中的 `transformVersion` 与 `contentHash` 配对 register；MUST NOT 在转码路径仍 register `v1`。

#### Scenario: Web direct upload returns v1 hint

- **WHEN** Web 成功上传 v1 合规视频（直传路径）
- **THEN** JSON `data` MUST 含 `contentHash`（长度 64）且 `transformVersion` MUST 为 `v1`

#### Scenario: Web transcode upload returns v2 hint

- **WHEN** Web 上传 v1 不合规但服务端转码成功
- **THEN** JSON `data` MUST 含 `contentHash` 且 `transformVersion` MUST 为 `v2`

#### Scenario: Gateway forwards transformVersion unchanged

- **WHEN** 请求经 gateway-app 反向代理至 ucg-service
- **THEN** 响应 JSON MUST 透传 `transformVersion` 字段（无裁剪）

---

## ucg-recommend-feed

<!-- source: openspec/specs/ucg-recommend-feed/spec.md -->

# ucg-recommend-feed Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: Recommend feed SHALL use mixed ranking algorithm

Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay (likes/comments age decay). Implementation MAY persist scores in `ucg_post_recommend` refreshed by热区 reconciler（默认 1h tick）与冷区 MQ 互动重算。

对 **尚无** `ucg_post_recommend` 行的 published 帖，Feed MUST 在排序中优先于已有 score 的帖（置顶区），置顶区内 MUST 按 `published_at` 降序；已有 score 的帖 MUST 按 `score` 降序（同分则 `published_at`、`id` 降序）。该置顶 MUST 持续直至 reconciler 或冷区 MQ 路径首次写入 recommend 行。

#### Scenario: Unscored new post appears before scored posts

- **WHEN** 用户请求 `GET /ucg/app/api/feed/recommend` 且存在 published 帖尚无 `ucg_post_recommend` 行
- **THEN** 响应 `list` 中该类帖 MUST 排在已有 score 帖之前（同置顶区内按 `published_at` 新者优先）

#### Scenario: After reconciler scores post leaves pin tier

- **WHEN** reconciler 已为帖 UPSERT `ucg_post_recommend.score`
- **THEN** 该帖 MUST 按 score 与全站其他已算分帖一起排序，不再仅因「无 recommend 行」置顶

### Requirement: 推荐 Feed 读路径 SHALL 在索引冷启动时回填 Redis 再返回

在 `ucg-feed-geo-composite-score` 已采用的 Redis 复合分 Feed 读路径之上，当推荐索引冷启动（ZSET 空且 MySQL 有 published 帖）时，Feed 读路径 MUST NOT 直接返回空列表；MUST 先执行有界索引 warm（见 `ucg-feed-index-lazy-warm`），再继续 GEO/ZSET/cursor 分页。snapshot miss 的 `backfillPostSnapshot` 语义不变。

#### Scenario: 未 backfill 环境首次打开推荐

- **WHEN** 用户请求 `GET /ucg/app/api/feed/recommend` 且无 cursor，Redis 尚无 recommend score，MySQL 有 published 帖
- **THEN** 响应 `list` MUST NOT 因索引缺失而恒为空（在 warm cap 内应有帖）

#### Scenario: 已有索引行为不变

- **WHEN** Redis ZSET 已有成员
- **THEN** Feed 排序、cursor、distance 行为 MUST 与 `ucg-feed-geo-composite-score` 一致，且无额外 warm 开销

### Requirement: 推荐 Feed 对作者本人帖 SHALL omit distanceMeters

当已登录 viewer 请求 `GET /ucg/app/api/feed/recommend` 且请求含有效 `lat`/`lng`、帖子 snapshot 含坐标时，若帖子 `authorWxId` 等于 viewer 的 `wxId`，该 item **MUST NOT** 含 JSON 字段 `distanceMeters`（即使 viewer 与帖坐标可计算 haversine）。本要求 **MUST NOT** 改变该帖的 `finalScore` 或 composite 排序语义。

#### Scenario: 本人帖 omit 距离

- **WHEN** 已登录用户 `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且列表含其本人发布的 published 帖（帖含坐标）
- **THEN** 该 item JSON MUST NOT 含 `distanceMeters`
- **AND** 同页他人帖在 viewer 与帖均有坐标时 MUST 仍含 `distanceMeters`

#### Scenario: 本人帖排序不变

- **WHEN** viewer 本人帖与他人帖参与同一推荐页 composite 排序
- **THEN** 系统 MUST 仍按既有 `finalScore = baseScore + distanceTerm` 排序
- **AND** omit `distanceMeters` MUST NOT 单独改变该帖的 `finalScore` 计算

#### Scenario: 未登录无本人概念

- **WHEN** 匿名或未带 `X-Internal-Wx-Id` 的有效登录上下文请求推荐 Feed
- **THEN** 距离字段行为 MUST 与变更前一致（有坐标则按他人帖规则返回 `distanceMeters`）

---

## ucg-recommend-mq

<!-- source: openspec/specs/ucg-recommend-mq/spec.md -->

# ucg-recommend-mq Specification

## Purpose
TBD - created by archiving change ucg-recommend-mq-incremental. Update Purpose after archive.
## Requirements
### Requirement: 推荐分更新事件 SHALL 使用已注册的 ucg.recommend routing keys

`ucg-service` 在推荐分需更新时 MUST Publish 下列 routing key 之一（经 `eventkit` 注册校验）：

- `ucg.post.published`
- `ucg.post.unpublished`
- `ucg.post.liked`
- `ucg.post.unliked`
- `ucg.comment.published`
- `ucg.comment.removed`

载荷 MUST 至少含 `postId`；comment 类事件 MUST 含 `commentId`。

#### Scenario: 审核通过后发布推荐更新

- **WHEN** 帖子 Green 审核 CAS 成功且 `status=published`
- **THEN** 系统 MUST Publish `ucg.post.published`，载荷含 `postId`

#### Scenario: 删帖或下架统一 unpublish

- **WHEN** 作者删除帖子或帖子从 published 变为 rejected/unpublished
- **THEN** 系统 MUST Publish `ucg.post.unpublished`，载荷含 `postId`

### Requirement: RabbitMQ 拓扑 SHALL 为推荐分队列绑定 topic exchange

仓库 `hack/rabbitmq-init` MUST 声明 durable 队列 `ucg.recommend.score.q` 并与 `voice.events` exchange 完成 binding（覆盖上述 routing key 或 `ucg.recommend.#`）。

#### Scenario: init 后队列可见

- **WHEN** 运维执行 rabbitmq init
- **THEN** 管理台 SHALL 可见 `ucg.recommend.score.q` 且 binding 正确

### Requirement: 推荐分 AMQP consumer MUST 驻留 ucg-service 且 SHALL 单帖更新

`ucg-service` MUST 经 AMQP push（`autoAck=false`）消费 `ucg.recommend.score.q`，按 routing key 分发：

- `liked` / `unliked` / `comment.published` / `comment.removed` → 若帖处于**冷区**（`published_at` 早于 `now - hotWindowHours`，与热区 reconciler 窗口一致），MUST `RecomputeRecommendScore(postId)` UPSERT `ucg_post_recommend`；若帖处于**热区**，MUST Ack 跳过本次重算（不得 Nack 重试）。
- `published` → MUST Ack 跳过且不 UPSERT（新帖 score 由热区 reconciler 写入；曝光由 Feed 未算分置顶保证）。
- `unpublished` → `DELETE FROM ucg_post_recommend WHERE post_id=postId`（或写路径已同步删除时 Ack 成功即可）。

MUST NOT 在 consumer 内对全部 published 帖做无 LIMIT 全表扫描。

#### Scenario: Cold zone like triggers recompute

- **WHEN** consumer 收到 `ucg.post.liked` 且 `postId` 对应帖 `published_at` 早于热区 cutoff
- **THEN** 系统 MUST 执行 `RecomputeRecommendScore` 且 MUST Ack

#### Scenario: Hot zone like skips recompute

- **WHEN** consumer 收到 `ucg.post.liked` 且帖 `published_at` 在热区窗口内
- **THEN** 系统 MUST NOT 调用 `RecomputeRecommendScore` 且 MUST Ack

#### Scenario: unpublished 删除推荐行且永远 Ack

- **WHEN** consumer 收到 `ucg.post.unpublished` 且 `postId` 合法
- **THEN** 系统 MUST 执行 DELETE；`RowsAffected=0` MUST 视为成功且 MUST NOT 报错；处理完成后 MUST Ack

### Requirement: like 类事件 throttle MUST 仅限制重算频率且 SHALL 允许短期误差

对 **冷区** `ucg.post.liked`、`ucg.post.unliked`、`ucg.comment.published`、`ucg.comment.removed`，consumer MUST 对 `postId` 使用 Redis 单 key `SET NX EX` throttle（默认 500ms）。窗口内 NX 失败 MUST 跳过本次重算并 Ack。throttle MUST 仅保证 500ms 内最多 1 次 `RecomputeRecommendScore`；MUST NOT 保证反映每一次 like/unlike 方向变化。Publisher MUST NOT 在发送侧合并事件。

热区事件因跳过 Recompute，MUST NOT 依赖 throttle 收敛（由 reconciler 负责）。

#### Scenario: Cold zone 500ms 内多次 like 只重算一次

- **WHEN** 冷区帖 `postId` 在 500ms 内收到 3 条 `ucg.post.liked`
- **THEN** consumer MUST 最多执行 1 次 `RecomputeRecommendScore`，且 3 条消息 MUST 均 Ack

### Requirement: 热区 reconciler MUST 分页增量且轮首固定 hotCutoff

系统 MUST 使用 MySQL 表 `ucg_recommend_hot_scan_state`（含 `last_post_id` 与 `round_hot_cutoff`）驱动热区分页 reconciler。当 `last_post_id=0` 开始新轮时 MUST 计算并持久化 `round_hot_cutoff`；同一轮后续分页 MUST 使用已存 `round_hot_cutoff` 且 MUST NOT 在分页过程中用 `NOW()` 重新计算 cutoff。每 tick MUST 使用 `published_at >= round_hot_cutoff AND id > last_post_id ORDER BY id LIMIT pageSize`；MUST NOT 一次加载全部 published 帖。

默认 tick 间隔 MUST 为 **3600s**（1 小时），可由 env `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 覆盖。

#### Scenario: 热区一轮扫完重置 cursor

- **WHEN** 某 tick 返回行数小于 `pageSize`
- **THEN** reconciler MUST 将 `last_post_id` 置 0；下轮开始 MUST 重新计算 `round_hot_cutoff`

#### Scenario: Default interval one hour

- **WHEN** 未设置 `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 且 yaml 为变更后默认值
- **THEN** reconciler 日志 MUST 显示 interval 为 1h 量级

### Requirement: 热区 reconciler MUST 周期性重算即使无互动

在冷区零兜底前提下，热区 reconciler MUST 对每个扫到的 published 热区帖执行 `RecomputeRecommendScore`，即使该帖在扫描周期内无任何 like/comment 事件。该行为 MUST 用于热区时间衰减与热区互动改由 reconciler 收敛后的 score 更新。

#### Scenario: 无互动热区帖仍更新 score

- **WHEN** 热区 reconciler 扫描到 `postId` 且该帖在上一扫描周期内 like/comment 计数未变
- **THEN** 系统 MUST 仍执行 `RecomputeRecommendScore(postId)` 以反映 `exp(-age/τ)` 变化

### Requirement: 冷区 MUST NOT 运行分页 reconciler

`published_at < round_hot_cutoff`（冷区）的帖子 MUST 通过 MQ 互动事件更新推荐分（见上文冷区分流）；系统 MUST NOT 为冷区启动定时全量或分页扫表任务。

#### Scenario: 冷区帖靠互动翻红

- **WHEN** 冷区帖收到 `ucg.post.liked` 且 throttle 允许重算
- **THEN** 系统 MUST 更新该帖 `ucg_post_recommend.score` 且 Feed MUST 可将该帖按 score 前排展示

### Requirement: ucg-service AMQP audit 与 recommend consumer SHALL 共用单 connection

`ucg-service` 内 UCG 审核队列 consumer 与推荐分 consumer MUST 共用 **一条** AMQP connection；每个消费队列 MUST 使用 **独立 channel** 与独立 prefetch 配置。连接断线 MUST 统一 backoff 重连并恢复全部 channel Consume。

#### Scenario: 单 connection 多 channel

- **WHEN** ucg-service 启动且 audit 与 recommend consumer 均 enabled
- **THEN** 进程 MUST 仅建立 1 条到 RabbitMQ 5672 的 AMQP connection，且 audit 4 队列与 recommend 1 队列各占用独立 channel

### Requirement: publishPostCAS 成功路径 SHALL 同步 Redis 推荐索引

帖子 Green 审核 CAS 成功变为 published 时，写路径 MUST 同步：

1. ZADD `ucg:recommend:score`（initial baseScore）
2. 若帖含 lat/lng → GEOADD `ucg:feed:geo`
3. SET post snapshot 与 author profile snapshot

MUST NOT 写入 `ucg_post_recommend`。

#### Scenario: 过审发帖同步 GEO 与 snapshot

- **WHEN** 帖 publish 成功且含 lat/lng
- **THEN** 系统 MUST GEOADD、`ZADD` 且写入 post/profile snapshot 于同一写路径事务语义内（best-effort 顺序：MySQL commit 后 Redis）

### Requirement: ucg-service 热区 reconciler SHALL 独立于 recommend MQ consumer 开关

`StartRecommendHotReconciler` MUST 在 `ucg-service` 启动时运行（与 `UCG_RECOMMEND_MQ_CONSUMER_ENABLED` 无关）。`UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` MUST 仅停止订阅 `ucg.recommend.score.q`，MUST NOT 停止热区 reconciler。

#### Scenario: Consumer disabled reconciler still runs

- **WHEN** `UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` 且进程启动
- **THEN** 日志 MUST 含 `[ucg-recommend-hot] started` 且 MUST NOT 订阅 recommend 队列

### Requirement: 发帖过审 MUST NOT 发布 recommend post.published 事件

帖 `pending_audit → published` 成功后，系统 MUST NOT 调用 `PublishPostPublished` / 向 `ucg.recommend.score.q` 发送 `published` 路由事件。该帖首次 score MUST 由热区 reconciler 写入。

#### Scenario: Publish CAS does not emit recommend published

- **WHEN** `publishPostCAS` 成功将帖设为 published
- **THEN** MUST NOT 向 recommend MQ 发送 `post.published` 事件

---

## ucg-service-runtime

<!-- source: openspec/specs/ucg-service-runtime/spec.md -->

# ucg-service-runtime Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: ucg-service SHALL run as dedicated microservice

The platform SHALL provide `ucg-service` process listening on `UCG_SERVICE_ADDR` default `:9804`, loading `manifest/config/config.ucg-service.yaml` when `GF_GCFG_FILE` is unset, mirroring `history-service` startup pattern.

#### Scenario: 启动与配置隔离
- **WHEN** 启动 ucg-service 且未设置 `GF_GCFG_FILE`
- **THEN** 进程 SHALL 加载 `config.ucg-service.yaml`，且 default DB SHALL 指向 `ai_voice_ucg`

#### Scenario: 依赖检查失败不监听
- **WHEN** MySQL 或 Redis 不可用且 fail-fast 启用
- **THEN** 进程 SHALL 退出且 SHALL NOT 进入监听态

---

## ucg-sim-chat-internal

<!-- source: openspec/specs/ucg-sim-chat-internal/spec.md -->

# ucg-sim-chat-internal Specification

## Purpose
TBD - created by archiving change ucg-sim-user-service. Update Purpose after archive.
## Requirements
### Requirement: ucg internal chat send SHALL allow simulated users only

`ucg-service` MUST 提供 `POST /ucg/internal/api/chat/send`，要求有效内部网关密钥（与 device ucg internal 一致）。请求体 MUST 含 `senderWxId`、`conversationId`、`clientMsgId`，以及 `content` 或 `imageKey`/`videoKey` 之一（互斥规则与 App WS 一致）。

发送前 MUST 经 device internal 确认 `senderWxId` 对应 `is_simulated=1`。否则 MUST 返回 403。发送方 MUST 为会话成员。成功 MUST 调用 `ProcessOutboundChatMessage`（含 Green 异步审核与真人 peer push）。**消息投递成功后**，MUST 对 `senderWxId` 调用 `MarkConversationRead`，将其在该会话的 `unread_count` 置 0（T5 未读闭环）；`last_read_msg_id` SHOULD 为本次投递消息 id。

#### Scenario: Sim user sends message

- **WHEN** `senderWxId` 为 sim 且为会话成员且 content 非空
- **THEN** 消息 MUST 持久化并进入正常聊天审核流程

#### Scenario: Sim sender marked read after send

- **WHEN** internal send 成功且发送前 sim 侧 `unread_count > 0`
- **THEN** 发送完成后 sim 侧 `unread_count` MUST 为 0

#### Scenario: Real user rejected

- **WHEN** `senderWxId` 的 `is_simulated=0`
- **THEN** API MUST 返回 403 且 MUST NOT 发送消息

#### Scenario: Not a member

- **WHEN** sender 非 conversation 成员
- **THEN** API MUST 返回 404 或等价业务错误

### Requirement: sim-user-service MUST NOT use WebSocket for outbound chat

模拟用户出站聊天 MUST 经上述 internal HTTP 契约；sim-user-service MUST NOT 建立 `/ucg/app/ws/chat` 长连接。

#### Scenario: No WS from sim service

- **WHEN** sim 用户回复私聊
- **THEN** 实现 MUST 使用 `POST /ucg/internal/api/chat/send` 而非 WebSocket

---

## ucg-sim-chat-unread-sample

<!-- source: openspec/specs/ucg-sim-chat-unread-sample/spec.md -->

# ucg-sim-chat-unread-sample Specification

## Purpose
TBD - created by archiving change sim-t5-unread-sample. Update Purpose after archive.
## Requirements
### Requirement: ucg-service SHALL expose internal sim unread conversation sample API

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/chat/sim-unread-sample`，供 `sim-user-service` 等受信内部调用方抽取 **一条** eligible 未读会话样本。鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。

请求体 MUST 含 `simWxIds`（`int64` 数组，非空）。MAY 含 `messageLimit`（默认 20，最大 50）。

**eligible 定义**（direct 1:1 会话）：

- sim 侧成员 `m`：`m.wx_id IN simWxIds` 且 `m.unread_count > 0` 且 `m.deleted_at = 0`
- 对端成员 `peer`：同 `conversation_id` 且 `peer.wx_id != m.wx_id` 且 `peer.wx_id NOT IN simWxIds` 且 `peer.deleted_at = 0`

响应 MUST NOT 调用 App 层 author 富化或 recommend 路径。命中时 MUST 返回 `conversationId`、`simWxId`、`peerWxId`、`unreadCount` 及 `messages[]`（最近消息，元素含 `senderWxId`、`content`）。无 eligible 时 MUST 返回 `found=false`（或等价空结果）与 HTTP 200（code=0），MUST NOT 500。

#### Scenario: Sample returns one eligible unread conversation

- **WHEN** 存在 sim 用户 S∈simWxIds 对真人 P 的未读会话，且请求含完整 simWxIds
- **THEN** 响应 MUST 含 `found=true`、`simWxId=S`、`peerWxId=P`、`conversationId>0` 及非空 `messages`

#### Scenario: No eligible unread

- **WHEN** simWxIds 非空但无任何 eligible 未读会话
- **THEN** 响应 MUST 含 `found=false` 且 MUST NOT 返回会话 id

#### Scenario: Invalid secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Sim-sim unread excluded

- **WHEN** 仅存在 sim↔sim 双方均在 simWxIds 内的未读会话
- **THEN** 响应 MUST 为 `found=false`

### Requirement: sim unread sample MUST use bounded ID probe without retry loop

抽样 MUST 在 eligible 集合上对 `ucg_conversation_member.id` 使用 **MIN/MAX + 均匀锚点 + `id >= anchor ORDER BY id LIMIT 1`**（空洞回退 minId），固定 **2 条**有界 SQL。MUST NOT 使用 `ORDER BY RAND()`。MUST NOT 对 `simWxIds` 内单个 wx 循环重试或分页扫描会话列表。

消息加载 MUST 为单条有界查询（`LIMIT messageLimit`），MUST NOT 调用 device HTTP。

#### Scenario: Bounded SQL on eligible subset

- **WHEN** 代码评审 sample 实现
- **THEN** MUST 为 eligible 子集上的 MIN/MAX + LIMIT 1 探测，MUST NOT 全表 `ORDER BY RAND()`

#### Scenario: No cross-domain DAO in ucg

- **WHEN** 代码评审 ucg internal sample 实现
- **THEN** MUST NOT import device 域 DAO；sim 身份过滤 MUST 依赖请求体 `simWxIds`

#### Scenario: Message limit clamp

- **WHEN** 请求 `messageLimit` 为 100
- **THEN** 实际返回 MUST 最多 50 条消息

---

## ucg-sim-feed-sample

<!-- source: openspec/specs/ucg-sim-feed-sample/spec.md -->

# ucg-sim-feed-sample Specification

## Purpose
TBD - created by archiving change sim-gentle-polling. Update Purpose after archive.
## Requirements
### Requirement: ucg-service SHALL expose internal published post sample API for sim workloads

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/posts/sample`，供 `sim-user-service` 等受信内部调用方抽取已发布帖子样本。鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。响应 MUST 仅包含评论任务所需最小字段，MUST NOT 触发 recommend Feed 的 author/media 富化或 device HTTP 调用。

每条样本 MUST 含 `postId`、`content`、`mediaType`。图文帖（`mediaType=1`）且存在首图 `coverObjectKey` 时 MUST 含 `coverCdnUrl`（全分辨率 CDN URL）。视频帖（`mediaType=2`）且存在封面 media key 时 MUST 含 `coverCdnUrl`（视频首帧 snapshot URL，MUST NOT 为 mp4 直链）。纯文字帖（`mediaType=0`）或无有效 media key 时 MAY 省略 `coverCdnUrl`。`coverObjectKey` MAY 继续返回供调试；LLM 输入 MUST 使用 `coverCdnUrl`。

#### Scenario: Image post sample includes cover CDN URL

- **WHEN** sample 返回 `mediaType=1` 且首条 media 有 objectKey
- **THEN** 响应项 MUST 含非空 `coverCdnUrl`，且 MUST 由 ucg OSS CDN 配置拼装

#### Scenario: Video post sample includes snapshot URL

- **WHEN** sample 返回 `mediaType=2` 且首条 media 有 objectKey
- **THEN** `coverCdnUrl` MUST 为首帧 snapshot URL（含 `x-oss-process=video/snapshot` 语义），MUST NOT 为原始视频文件 URL

#### Scenario: Text-only post omits cover CDN URL

- **WHEN** sample 返回 `mediaType=0`
- **THEN** `coverCdnUrl` MUST 为空或省略

#### Scenario: Invalid or missing secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Empty plaza

- **WHEN** 无 `status=published` 帖子
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

### Requirement: sample API MUST use bounded single-query read pattern

抽样读路径 MUST 在 ucg 库内完成，MUST NOT 调用 `postsFromResult`、`GetPublicProfile` 或 `Device().BatchWx`。

- **latest 模式**（缺省）：单条有界 SQL，`LIMIT` ≤ 50；`limit` 默认 20，超出 MUST 截断为 50；排序 `published_at DESC`。
- **random 模式**：MUST 使用有界读模式：一次 `MIN(id)/MAX(id)` 聚合（`WHERE status=published` 及可选 author 排除）加一次 `id >= R LIMIT 1` 探测（共 2 次 SQL）；每次 MUST 带 `LIMIT 1` 或聚合有界，MUST NOT 全表加载。
- 当请求 body 含非空 `excludeAuthorWxIds` 时，latest 与 random 模式的 published 帖查询 MUST 附加 `author_wx_id NOT IN (...)`（与 T5 `simWxIds` 排除 sim peer 语义对称）。

#### Scenario: Limit clamp in latest mode

- **WHEN** 请求 `limit` 为 100 且未指定 random 模式
- **THEN** 实际查询 MUST 最多返回 50 条

#### Scenario: Random mode bounded queries

- **WHEN** 请求 `mode=random`
- **THEN** 读路径 MUST 最多 2 次 SQL且最终 MUST NOT 返回超过 1 条

#### Scenario: Exclude sim authors in random mode

- **WHEN** 请求 `mode=random` 且 `excludeAuthorWxIds` 含全部 sim wxId，且存在至少一条 published 帖其 author 不在该集合
- **THEN** 响应 `list[0].authorWxId` MUST 不在 `excludeAuthorWxIds` 中

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审 ucg internal sample 实现（含 random 分支）
- **THEN** MUST NOT import device 域 DAO 或直连 device 库表

### Requirement: sample API SHALL support random mode via ID probe with mild recency bias

当请求 body `mode` 为 `random` 时，系统 MUST 在 ucg 库内通过有界 ID 探测返回 **0 或 1** 条满足 **status 与 excludeMediaTypes（若提供）** 条件的已发布帖，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 在过滤后的 published 集合上覆盖全库（非仅最新 N 条）。锚点 MUST 在 eligible 帖的 `[minId, maxId]` 上按 `R = minId + floor((maxId - minId) * U^α)` 生成（`U` 均匀随机，默认 `α = 0.65`），随后 `WHERE … AND id >= R ORDER BY id ASC LIMIT 1`。响应字段 MUST 含 `postId`、`content`、`mediaType`、可选 `coverObjectKey`/`coverCdnUrl`。

#### Scenario: Random mode returns one published post

- **WHEN** 有效 internal 密钥、body `{ "mode": "random" }`（或无 exclude），且存在 eligible published 帖
- **THEN** 响应 `list` MUST 含 1 条 published 帖

#### Scenario: Random mode empty plaza

- **WHEN** 无 eligible published 帖（含 exclude 后为空）
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Latest mode unchanged without exclude

- **WHEN** `mode=latest`（或缺省）、`limit=20`、无 `excludeMediaTypes`
- **THEN** MUST 仍按 `published_at DESC` 返回最多 20 条

### Requirement: sample API SHALL support excludeMediaTypes filter

`POST /ucg/internal/api/posts/sample` 请求 body MAY 含 `excludeMediaTypes`（整型数组）。当数组非空时，抽样 MUST 仅返回 `media_type` **不在**该集合中的已发布帖。`mode=random` 的 ID 探测（MIN/MAX 与 `id>=R` probe）MUST 应用与列表查询相同的 `excludeMediaTypes` filter。未传或空数组时 MUST 保持变更前行为（含视频帖）。

#### Scenario: Random sample excludes video

- **WHEN** 请求为 `{ "mode": "random", "excludeMediaTypes": [2] }` 且存在 published 非视频帖
- **THEN** 返回的每条样本 `mediaType` MUST NOT 为 `2`

#### Scenario: Random bounds respect exclude filter

- **WHEN** 请求 `mode=random` 且 `excludeMediaTypes` 为 `[2]`
- **THEN** MIN/MAX id 聚合 MUST 仅在 `media_type NOT IN (2)` 的 published 帖上计算

#### Scenario: No eligible posts after exclude

- **WHEN** 所有 published 帖均为 `media_type=2` 且 `excludeMediaTypes` 含 `2`
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Backward compatible without exclude

- **WHEN** 请求未含 `excludeMediaTypes`
- **THEN** random/latest 行为 MUST 与引入本字段前一致

### Requirement: sample API response SHALL include authorWxId for internal consumers

`POST /ucg/internal/api/posts/sample` 响应 `list` 每项 MUST 含 `authorWxId`（`ucg_post.author_wx_id`）。T2 等既有调用方 MAY 忽略该字段。

#### Scenario: Author field present

- **WHEN** sample 返回非空 `list`
- **THEN** 每条 MUST 含 `authorWxId` 且 MUST 大于 0

### Requirement: sample API SHALL support excludeDebate and onlyDebate filters

`POST /ucg/internal/api/posts/sample` 请求 body MAY 含 `excludeDebate`（bool）与 `onlyDebate`（bool）。二者 MUST NOT 同时为 true。判定 MUST 与 UCG `isDebatePost` 一致：`(debate_left_label != '' AND debate_right_label != '') OR type = 'debate'`。

- `excludeDebate=true`：MUST 仅返回非辩论帖；random 的 MIN/MAX 与 probe MUST 应用相同 filter。
- `onlyDebate=true`：MUST 仅返回辩论帖；random 的 MIN/MAX 与 probe MUST 应用相同 filter。
- 均未传或为 false：MUST 保持变更前行为（含辩论帖）。

响应 `list` 每项 MUST 含 `debateLeft`、`debateRight`（字符串；非辩论帖 MUST 为空字符串或省略）。T2 MAY 忽略；T8 MUST 使用。

#### Scenario: T2 exclude debate in SQL

- **WHEN** 请求 `{ "mode": "random", "excludeDebate": true }` 且存在 published 非辩论帖
- **THEN** 返回样本 MUST NOT 为辩论帖

#### Scenario: T8 only debate in SQL

- **WHEN** 请求 `{ "mode": "random", "onlyDebate": true }` 且存在 published 辩论帖
- **THEN** 返回样本 MUST 为辩论帖且 MUST 含非空 `debateLeft` 与 `debateRight`

#### Scenario: Mutually exclusive flags rejected

- **WHEN** 请求同时 `excludeDebate=true` 且 `onlyDebate=true`
- **THEN** MUST 返回 400 且 MUST NOT 查询

#### Scenario: Empty when only debate and none exist

- **WHEN** `onlyDebate=true` 且无 published 辩论帖
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Debate fields on debate sample

- **WHEN** sample 返回辩论帖
- **THEN** `debateLeft` 与 `debateRight` MUST 各 ≤5 字且非空

---

## ucg-video-thumb

<!-- source: openspec/specs/ucg-video-thumb/spec.md -->

# ucg-video-thumb Specification

## Purpose
TBD - created by archiving change ucg-video-thumb-physical. Update Purpose after archive.
## Requirements
### Requirement: Video thumb objectKey SHALL be stem_thumb.jpg regardless of video extension

平台 MUST 在 `internal/shared/mediacdn` 提供 `VideoThumbObjectKey(videoObjectKey)` helper（及 `VideoThumbExt = "jpg"` 常量）。对视频原 objectKey `{path}/{stem}.{videoExt}`，缩略图 objectKey MUST 为 `{path}/{stem}_thumb.jpg`，且 MUST NOT 使用与视频相同的扩展名（如 `xyz.mp4` → `xyz_thumb.jpg`，NOT `xyz_thumb.mp4`）。

业务代码 MUST 经该 helper 派生视频 thumb key，MUST NOT 散落 `_thumb.jpg` 字面量拼接。

对已是视频 thumb 的 key（stem 以 `_thumb` 结尾且 ext 为 jpg），`VideoThumbObjectKey` MUST 原样返回。

#### Scenario: MP4 原视频派生 thumb key

- **WHEN** 原视频 objectKey 为 `social/2026/06/xyz.mp4`
- **THEN** `VideoThumbObjectKey` SHALL 返回 `social/2026/06/xyz_thumb.jpg`

#### Scenario: thumb key 幂等

- **WHEN** 输入已为 `social/2026/06/xyz_thumb.jpg`
- **THEN** `VideoThumbObjectKey` SHALL 原样返回该 key

### Requirement: EnsureVideoThumb SHALL create idempotent physical first-frame objects via OSS snapshot

ucg-service MUST 提供 `EnsureVideoThumb(ctx, videoObjectKey)`：对视频原 objectKey，若 `{stem}_thumb.jpg` 不存在，MUST 经 OSS `GetObject` 携带 `video/snapshot,t_0` 获取首帧字节后 `PutObject` 至 `VideoThumbObjectKey(videoObjectKey)`，`Content-Type` MUST 为 `image/jpeg`；若 thumb 已存在 MUST 跳过（幂等）。原视频不存在时 MUST 返回明确错误。

OSS `x-oss-process` MUST 仅用于服务端生成阶段，MUST NOT 出现在返回客户端的 CDN URL 中。

#### Scenario: 首次生成视频 thumb

- **WHEN** 原视频 `social/.../a.mp4` 存在于 OSS 且 `social/.../a_thumb.jpg` 不存在
- **THEN** `EnsureVideoThumb` SHALL 上传 `a_thumb.jpg` 且后续 `HEAD` 可命中

#### Scenario: 重复调用幂等

- **WHEN** `a_thumb.jpg` 已存在
- **THEN** `EnsureVideoThumb` SHALL 成功返回且 MUST NOT 覆盖已有对象

### Requirement: BuildVideoThumbnailURL SHALL return physical thumb CDN URL without x-oss-process

`BuildVideoThumbnailURL(videoObjectKey)` MUST 返回 `BuildCdnURL(VideoThumbObjectKey(videoObjectKey))`。视频缩略图 CDN URL MUST NOT 包含 `x-oss-process` query 参数。

#### Scenario: 视频列表缩略图 URL 无 query

- **WHEN** 服务为视频 objectKey 拼装 `thumbnailUrl` 或 `mediaThumbnailUrl`
- **THEN** URL path SHALL 以 `_thumb.jpg` 结尾且 SHALL NOT 含 `x-oss-process`

### Requirement: Video media deletion SHALL remove paired _thumb.jpg objects

当 ucg-service 删除用户拥有的视频 OSS 原对象时，MUST 同时尝试删除 `VideoThumbObjectKey(原视频 key)`；thumb 对象不存在时 MUST NOT 导致整次删除失败。

#### Scenario: 删除 mp4 同时删除 thumb jpg

- **WHEN** `DeleteOwnedMedia` 删除 `social/.../a.mp4` 且 blob 允许删 OSS
- **THEN** OSS 上 `social/.../a_thumb.jpg` SHALL 被删除或已不存在

### Requirement: Historical videos without physical thumbs SHALL NOT be backfilled by platform CLI

平台 MUST NOT 提供批量 backfill CLI 为历史视频生成 thumb。无物理 thumb 的历史视频读路径 MAY 返回 404 thumb URL；补救 MUST 由用户通过重编辑或重上传触发写路径 `EnsureVideoThumb`。

#### Scenario: 无 backfill 命令

- **WHEN** 运维检索仓库 `cmd/` 与 runbook
- **THEN** MUST NOT 存在 `ucg-video-thumb-backfill` 或等价批量补全工具

---

## ucg-video-transcode

<!-- source: openspec/specs/ucg-video-transcode/spec.md -->

# ucg-video-transcode Specification

## Purpose
TBD - created by archiving change ucg-video-normalize-phase1. Update Purpose after archive.
## Requirements
### Requirement: NormalizeVideo SHALL produce v2 canonical mp4

ucg-service MUST 提供 `NormalizeVideo`（或等价导出函数），将任意可解码输入转码为 **v2 canonical** mp4：

- 视频：libx264，profile **main**，pix_fmt **yuv420p**
- 音频：**aac**；若输入无音轨 MUST 补 **静音 AAC** 轨并与视频时长对齐
- 容器：mp4，**movflags +faststart**
- 输出 MUST 通过 `v2` 验真规则

转码 MUST 使用进程内 `ffmpeg`/`ffprobe`（非 OSS 侧处理）。失败 MUST 返回错误且 MUST NOT 上传半成品至 OSS。

#### Scenario: Transcode silent video adds AAC

- **WHEN** 输入文件无音轨且调用 `NormalizeVideo`
- **THEN** 输出 MUST 含 AAC 音轨且 MUST 通过 v2 验真

#### Scenario: Transcode output is faststart mp4

- **WHEN** 输入为 moov 后置的 h264+AAC mp4
- **THEN** 输出 MUST 满足 v2 faststart 要求

### Requirement: Internal upload-video SHALL transcode before OSS

ucg-service MUST 注册 `POST /ucg/internal/api/media/upload-video`，接受 multipart 视频文件，鉴权 MUST 与现有 `POST /ucg/internal/api/media/upload` 一致（内部网关密钥）。

处理流程 MUST 为：读取 body（上限与 `MaxMediaUploadBytes` 一致）→ `NormalizeVideo` → PUT OSS（`video/mp4`）→ 响应 `objectKey`、`cdnUrl`、`contentHash`（SHA-256 hex 小写，对 **OSS 上最终字节** 计算）。

本接口 MUST NOT 自动 `RegisterMedia`；调用方 MUST 自行 register（含 `transformVersion=v2`）。

#### Scenario: Internal upload returns canonical object

- **WHEN** 内部密钥有效且上传可解码视频
- **THEN** 响应 MUST 含 objectKey 与 contentHash，且 OSS 对象 MUST 通过 v2 验真

#### Scenario: Internal upload unauthorized

- **WHEN** 未提供有效内部密钥
- **THEN** MUST 返回 403

#### Scenario: Internal upload transcode failure

- **WHEN** ffmpeg 无法解码输入
- **THEN** MUST 返回 5xx 或 4xx 明确错误且 MUST NOT 留下 OSS 对象

### Requirement: ucg-service container SHALL include ffmpeg

`manifest/docker/Dockerfile.ucg-service` 构建的镜像 MUST 包含可执行的 `ffmpeg` 与 `ffprobe`，供验真与转码使用。部署 ucg-service  MUST 使用含 ffmpeg 的镜像方可启用本能力。

#### Scenario: Transcode available in container

- **WHEN** ucg-service 容器启动且配置启用视频转码
- **THEN** 进程 MUST 能成功执行 `ffmpeg -version` 与 `ffprobe -version`

---

## ucg-video-validate

<!-- source: openspec/specs/ucg-video-validate/spec.md -->

# ucg-video-validate Specification

## Purpose
TBD - created by archiving change ucg-video-normalize-phase1. Update Purpose after archive.
## Requirements
### Requirement: ucg-service SHALL validate video by transformVersion with forked rules

ucg-service MUST 提供按 `transformVersion` 分叉的视频 ffprobe 验真能力。`v1` 与 `v2` 规则 **部分对齐、刻意分叉**：通过 `v1` 验真 **不** 等价于 canonical 合规。

**v2（canonical）** MUST 满足：

- 容器 format 为 mp4
- 视频轨 codec 为 h264，profile 为 **Main**，pix_fmt 为 yuv420p
- **必须有音轨且 codec 为 aac**（静音 AAC 合法）
- **faststart**：moov MUST 位于 mdat 之前（可播放渐进下载）
- 大小 ≤ ucg-service 配置的单文件上传上限

**v1（Web Phase 1 宽规）** MUST 满足：

- 容器 format 为 mp4
- 视频轨 codec 为 h264，profile 为 **Main 或 Baseline**，pix_fmt 为 yuv420p
- **必须有音轨且 codec 为 aac**（静音 AAC 合法）；**无音轨 MUST 拒绝**（Phase 1 不补轨）
- faststart **不** 强制
- 大小 ≤ 单文件上传上限

验真 MUST 支持对内存字节与 OSS objectKey（Range 读取 + ffprobe）执行。未列出的 `transformVersion` MUST NOT 套用 v1/v2 规则。

#### Scenario: v2 rejects Baseline profile

- **WHEN** OSS 对象为 h264 Baseline + AAC + faststart 的 mp4 且 register 请求 `transformVersion=v2`
- **THEN** 验真 MUST 失败且 register MUST 返回 4xx

#### Scenario: v1 accepts Baseline with AAC

- **WHEN** Web 上传 h264 Baseline + AAC 的 mp4（可无 faststart）且验真版本为 v1
- **THEN** 验真 MUST 通过

#### Scenario: v1 rejects missing audio track

- **WHEN** 视频仅有视频轨、无音轨，且验真版本为 v1
- **THEN** 验真 MUST 失败且 MUST NOT 直传 OSS（Web upload）或 MUST 拒绝 register

#### Scenario: v2 rejects missing AAC

- **WHEN** 视频含非 AAC 音轨（如 mp3）且验真版本为 v2
- **THEN** 验真 MUST 失败

#### Scenario: v2 rejects missing faststart

- **WHEN** mp4 满足 h264 Main + AAC 但 moov 不在 mdat 前且验真版本为 v2
- **THEN** 验真 MUST 失败

### Requirement: Web video proxy upload SHALL validate v1 before OSS PUT

`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 时 MUST 对上传字节执行 **v1 验真优先** 的分支处理：

1. **v1 通过**：MUST 将 **原始上传字节** 直传 OSS（与 Phase 1 一致）；`contentHash` MUST 为原始字节 SHA-256；响应 MUST 提示 `transformVersion=v1`。
2. **v1 未通过且可解码**：MUST 调用 `NormalizeVideo` 转码为 v2 canonical mp4 后 PUT OSS；`contentHash` MUST 为 **转码后 OSS 字节** SHA-256；响应 MUST 提示 `transformVersion=v2`；register MUST 使用 `v2`（非 `v1`）。
3. **v1 未通过且不可解码**：MUST 返回 4xx 且 MUST NOT 在 OSS 上创建对象。

v1 已合规但缺少 faststart 的对象 MUST 走路径 1（直传 `v1`），MUST NOT 仅为 faststart 触发 remux 或转码。

#### Scenario: Web compliant video uploads direct v1

- **WHEN** Web 上传满足 v1 规则的 mp4（可无 faststart）
- **THEN** 响应 MUST 含 `objectKey`、`cdnUrl`、`contentHash`，OSS MUST 存在与上传字节一致的对象，且 MUST 含 `transformVersion=v1`

#### Scenario: Web non-compliant but decodable transcodes to v2

- **WHEN** Web 上传 HEVC mp4 或无 AAC 音轨但 ffprobe 可解析且含视频轨
- **THEN** API MUST 成功返回且 OSS 对象 MUST 通过 v2 验真，响应 MUST 含 `transformVersion=v2` 与转码后 `contentHash`

#### Scenario: Web undecodable video rejected

- **WHEN** Web 上传损坏或非视频文件
- **THEN** API MUST 返回 4xx 且 OSS MUST NOT 存在新对象

#### Scenario: v1 compliant without faststart stays direct

- **WHEN** Web 上传 h264 Main/Baseline + AAC 的 mp4 但 moov 不在 mdat 前（v1 允许）
- **THEN** MUST 直传原始字节且 `transformVersion` MUST 为 `v1`（MUST NOT 转码）

### Requirement: Video RegisterMedia SHALL validate OSS object by transformVersion

`RegisterMedia` 在 `mediaKind=2` 时 MUST：

- 仅接受 `transformVersion` 为 `v1` 或 `v2`；其他值（含 `sim-raw`）MUST 返回 400
- 在登记 blob/ownership 之前 MUST 对 `objectKey` 对应 OSS 对象执行与请求 `transformVersion` 匹配的验真
- 验真失败 MUST 返回 4xx 且 MUST NOT 完成 blob 登记

`mediaKind=1` 行为 MUST 不变（图片 thumb 逻辑不受本 requirement 影响）。

#### Scenario: Native v2 register succeeds

- **WHEN** 客户端 PUT 满足 v2 规则的 mp4 且 `RegisterMedia` 带 `transformVersion=v2` 与正确 contentHash
- **THEN** 登记 MUST 成功

#### Scenario: sim-raw register rejected

- **WHEN** `RegisterMedia` 请求 `transformVersion=sim-raw` 且 `mediaKind=2`
- **THEN** MUST 返回 400 且 MUST NOT 登记

#### Scenario: v1 register validates relaxed rules

- **WHEN** OSS 对象为 v1 合规（Baseline + AAC、无 faststart）且 register `transformVersion=v1`
- **THEN** 登记 MUST 成功

#### Scenario: Register v2 on v1-only object fails

- **WHEN** OSS 对象仅满足 v1（如无 faststart）但 register `transformVersion=v2`
- **THEN** MUST 返回 4xx

### Requirement: ProbeVideoDecodable SHALL distinguish transcodable from corrupt uploads

ucg-service MUST 提供对上传字节的 **可解码探测**（如 `ProbeVideoDecodable`），用于 Web 视频代理上传在 **v1 验真失败** 后判断是否进入服务端转码：

- ffprobe MUST 能解析容器与 streams
- streams MUST 含至少一条 `codec_type=video` 的视频轨

不满足上述条件 MUST 视为 **不可解码**，Web upload MUST 返回 4xx 且 MUST NOT 调用 `NormalizeVideo`。

本探测 MUST NOT 替代 v1/v2 验真；仅用于 B 分支（转码兜底）门禁。

#### Scenario: HEVC mp4 is decodable for transcode fallback

- **WHEN** 上传 mp4 含 h265/hevc 视频轨，ffprobe 成功解析
- **THEN** `ProbeVideoDecodable` MUST 成功（即使 v1 验真因非 h264 失败）

#### Scenario: Corrupt file is not decodable

- **WHEN** 上传字节无法被 ffprobe 解析
- **THEN** `ProbeVideoDecodable` MUST 失败且 Web upload MUST NOT PUT OSS

---

## user-agreement-company

<!-- source: openspec/specs/user-agreement-company/spec.md -->

# user-agreement-company Specification

## Purpose
TBD - created by archiving change company-branding-site-legal. Update Purpose after archive.
## Requirements
### Requirement: 用户协议 MUST 以公司为运营与缔约主体

`resource/public/user-agreement.html` MUST 明确本协议由用户与**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）订立；「胖宝」为该公司运营的应用程序。注册地址表述 MUST 为「以市场主体公示为准」或等价。联系方式 MUST 提供客服邮箱 `tou_zy@foxmail.com`（或指向与隐私政策一致的联系渠道）。路径 MUST 仍为 `/user-agreement.html`。文档 MUST 更新生效或修订日期。

#### Scenario: 用户阅读协议主体

- **WHEN** 用户打开 `/user-agreement.html`
- **THEN** 页面 SHALL 可见公司中文全称作为运营/缔约主体，且 MUST NOT 将知识产权或运营归属表述为单独的「胖宝团队」而无公司主体

### Requirement: 用户协议 MUST 覆盖知识产权、付费与 AI 免责要点

用户协议 MUST：

- 将 App UI、LOGO、软件技术等知识产权归属表述为归杭州彩逗科技有限公司所有（或该公司享有相应权利）；用户宝宝记录数据归用户所有的既有精神 MUST 保留。
- 说明付费或会员/功能开通相关服务以 App 内展示与支付渠道规则为准；依法可退款情形遵循适用法律与平台规则（表述简洁即可）。
- 说明 AI 生成内容仅供参考，不能替代专业医疗诊断；保留社区行为规范与账号安全等既有合理条款。
- 保留微信登录与 Apple 登录相关账号说明的既有正确边界。

#### Scenario: 用户阅读知识产权与 AI

- **WHEN** 用户阅读知识产权与服务说明相关章节
- **THEN** 文案 SHALL 将 IP 归属指向公司，并 SHALL 含 AI 不能替代医疗诊断的免责要点

---

## validated-prefix-dispatch

<!-- source: openspec/specs/validated-prefix-dispatch/spec.md -->

# validated-prefix-dispatch Specification

## Purpose
TBD - created by archiving change routing-key-prefix-governance-and-dispatch. Update Purpose after archive.
## Requirements
### Requirement: 分发前必须执行路由键白名单校验
系统 MUST 在 outbox 发布与投影分发入口先执行路由键合法性校验；未注册路由键必须被拒绝并记录拒绝来源。

#### Scenario: 收到未注册路由键
- **WHEN** outbox 处理到不在注册表中的 `routing_key`
- **THEN** 系统必须拒绝该事件并输出包含来源模块的告警日志

### Requirement: 校验通过后必须按前缀分组分发
系统 SHALL 在路由键通过合法性校验后，基于前缀常量将事件分发到对应领域处理器，而非依赖逐项路由键枚举分支。

#### Scenario: 处理 history.record 前缀事件
- **WHEN** 事件路由键为 `history.record.*` 且已通过白名单校验
- **THEN** 系统必须将该事件分发给 history 投影处理器

### Requirement: 必须提供未知前缀默认保护
系统 MUST 为已注册但未映射分发处理器的前缀保留默认保护分支，避免静默忽略。

#### Scenario: 路由键合法但前缀未绑定处理器
- **WHEN** 事件通过白名单校验但其前缀没有配置分发处理器
- **THEN** 系统必须输出告警并按既定失败语义处理

---

## vip-entitlement

<!-- source: openspec/specs/vip-entitlement/spec.md -->

# vip-entitlement Specification

## Purpose
TBD - created by archiving change vip-cash-service. Update Purpose after archive.
## Requirements
### Requirement: VIP 权益以 ai_voice_cash 表为准

账号是否 VIP MUST 由 `ai_voice_cash` 中的权益记录判定，主键为 `wx.id`（`wx_id`）。MUST NOT 使用 `ai_voice_device.wx.is_vip` 或 `device_no` 作为 VIP 真相源。当且仅当权益 `expire_at` 晚于当前时间时，该账号 MUST 被视为 VIP。

#### Scenario: 无权益记录

- **WHEN** 查询某 `wxId` 且不存在权益行或 `expire_at` 已过期
- **THEN** 系统 MUST 判定 `isVip=false`

#### Scenario: 未过期权益

- **WHEN** 查询某 `wxId` 且权益 `expire_at` 大于当前时间
- **THEN** 系统 MUST 判定 `isVip=true`

### Requirement: 内部按 wxId 查询 VIP

cash-service MUST 提供经内部密钥鉴权的 HTTP 接口 `GET /cash/internal/api/vip/by-wx-id`（query `wxId`），响应至少包含 `wxId` 与 `isVip`（宜含 `expireAt`）。voice-service 等跨域调用方 MUST 经此契约读取 VIP，MUST NOT 直查 `ai_voice_cash`。

#### Scenario: 合法内部查询

- **WHEN** 调用方携带正确内部密钥请求正整数 `wxId`
- **THEN** cash-service MUST 返回该账号当前 VIP 布尔结果

#### Scenario: 缺少内部密钥

- **WHEN** 请求未携带正确内部密钥
- **THEN** cash-service MUST 拒绝该请求

#### Scenario: 无效 wxId

- **WHEN** `wxId` 缺失或 ≤0
- **THEN** cash-service MUST 返回参数错误，MUST NOT 视为 VIP

---

## vip-overlay-entitlement

<!-- source: openspec/specs/vip-overlay-entitlement/spec.md -->

# vip-overlay-entitlement Specification

## Purpose
TBD - created by archiving change invite-peer-force-ucg. Update Purpose after archive.
## Requirements
### Requirement: VIP 与永久单事件开通写路径分离

VIP 支付/续期履约 MUST 只更新 VIP 权益（含 `expireAt`），MUST NOT 修改 `feature_allowed_count` 或把 VIP 写成永久 feature entitlement 来「买断」预测条数。单事件支付/邀请/广告 MUST 只写入对应永久条数或 entitlement，MUST NOT 修改 VIP `expireAt`。

#### Scenario: 开通月卡

- **WHEN** 用户 VIP 履约成功
- **THEN** VIP 有效期延长且预测永久 `allowed_count` 不变

#### Scenario: 单次预测支付

- **WHEN** 用户支付预测 +1 成功
- **THEN** 永久条数 +1 且 VIP 状态不变

### Requirement: 读模型叠加使用权

对需开通的功能（含预测锁），业务使用态 MUST 在「永久已开通/条数内」与「当前 isVip」之间按 OR 放行。catalog 返回的预测 `allowedCount` MUST 反映永久合成值，MUST NOT 仅因 isVip 改为全开哨兵。UCG 入场资格 MUST 继续不受 isVip 影响。

#### Scenario: VIP 未买断条数

- **WHEN** 用户 isVip=true 且永久预测条数小于 totalActivatableCount
- **THEN** 使用态可查看预测事项，但 catalog `allowedCount` 仍为永久值

---

## vip-payment

<!-- source: openspec/specs/vip-payment/spec.md -->

# vip-payment Specification

## Purpose
TBD - created by archiving change vip-cash-service. Update Purpose after archive.
## Requirements
### Requirement: 一期单档月会员商品

cash-service MUST 提供一期唯一可售 VIP 商品：`product_code=vip_monthly_19`，标价 **1900 分人民币**，时长 **30 天**（续期按日叠加）。App MUST 能通过 `GET /cash/app/api/vip/product`（需登录，`X-Internal-Wx-Id>0`）读取该商品信息。

#### Scenario: 查询一期商品

- **WHEN** 已登录用户请求 VIP 商品接口
- **THEN** 响应 MUST 包含 `vip_monthly_19`、月费语义与 19 元标价（或等价分单位字段）

### Requirement: 创建 VIP 订单

cash-service MUST 提供 `POST /cash/app/api/vip/orders`：从 Header 读取 `X-Internal-Wx-Id`（>0），body 含 `productCode` 与 `channel`（`alipay` 或 `apple_iap`）。成功时 MUST 创建 `created` 状态订单并返回客户端调起支付所需参数（支付宝为调起串/参数；Apple 为 orderNo、预期 `appleProductId`，以及 UUID 格式的 `appAccountToken`）。金额 MUST 以服务端商品表为准。

#### Scenario: 支付宝建单

- **WHEN** 合法 `wxId` 以 `channel=alipay` 且 `productCode=vip_monthly_19` 建单
- **THEN** 系统 MUST 落库订单且状态为 `created`，并返回可用于调起支付宝的参数

#### Scenario: Apple IAP 建单

- **WHEN** 合法 `wxId` 以 `channel=apple_iap` 建单
- **THEN** 系统 MUST 落库订单且返回与 ASC 商品映射一致的 `appleProductId`（或等价字段）、订单号，以及非空 UUID `appAccountToken`

#### Scenario: 未登录建单

- **WHEN** 缺少有效 `X-Internal-Wx-Id`
- **THEN** 系统 MUST 拒绝建单

### Requirement: 支付宝异步通知开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/alipay/notify`：网关可匿名到达，但 MUST 校验支付宝签章与订单金额/商户订单号。校验通过后 MUST 将订单置为 `paid`（若尚未支付），并按商品时长续期该 `wx_id` 的 `vip_entitlement`（`new_expire = max(now, current_expire) + 30d`）。重复通知 MUST 幂等，MUST NOT 重复叠加错误时长。

#### Scenario: 首次合法 notify

- **WHEN** 支付宝对某 `created` 订单首次投递合法成功通知
- **THEN** 订单 MUST 变为 `paid`，且对应账号 VIP `expire_at` MUST 相对支付前至少延长 30 天

#### Scenario: 重复 notify

- **WHEN** 同一渠道交易对已 `paid` 订单再次通知
- **THEN** 系统 MUST 返回成功应答语义且 MUST NOT 再次延长权益（或仅保持幂等结果）

#### Scenario: 验签失败

- **WHEN** notify 签章非法
- **THEN** 系统 MUST NOT 将订单标为 paid，MUST NOT 开通权益

### Requirement: Apple IAP 验单开通权益

cash-service MUST 提供 `POST /cash/app/api/vip/apple/verify`（需登录）：接受 App 提交的 IAP 交易凭证（JWS 或团队选定的等价载荷）。该路径 MAY 用于支付后即时开通（加速 UX），但 **Apple 扣款后的权威开通来源 MUST 为 App Store Server Notifications**（见能力 `apple-server-notifications`）。verify 与 ASN 履约 MUST 共用幂等键（`channel=apple_iap` 与 `channel_txn_id=transactionId`），MUST NOT 因双路径重复叠加权益。校验通过且 `productId` 映射正确时，MUST 幂等将关联订单置 `paid`（或绑定已有订单）并续期 entitlement。本路径 MUST 使用 **Apple IAP**，MUST NOT 使用 Apple Pay（PassKit）作为数字会员开通方式。

#### Scenario: 合法 IAP 验单

- **WHEN** 已登录用户提交可验证的 IAP 交易且商品映射正确
- **THEN** 系统 MUST 开通或续期该 `wxId` 的月会员权益

#### Scenario: 重复提交同一 transaction

- **WHEN** 同一 Apple `transactionId`（或等价渠道交易号）再次 verify
- **THEN** 系统 MUST 幂等成功，MUST NOT 重复叠加错误的时长

#### Scenario: ASN 已履约后再 verify

- **WHEN** 同一 `transactionId` 已由 ASN 履约成功，用户再次调用 verify
- **THEN** 系统 MUST 幂等成功且 MUST NOT 再次延长权益

### Requirement: 查询当前 VIP 状态

cash-service MUST 提供 `GET /cash/app/api/vip/status`（需登录），返回当前 Header `wxId` 的 `isVip` 与 `expireAt`（无权益时 `isVip=false`）。

#### Scenario: 会员中查询状态

- **WHEN** 权益未过期的用户请求 status
- **THEN** 响应 MUST 含 `isVip=true` 与未来的 `expireAt`

---

## vip-product-admin

<!-- source: openspec/specs/vip-product-admin/spec.md -->

# vip-product-admin Specification

## Purpose
TBD - created by archiving change add-vip-product-admin. Update Purpose after archive.
## Requirements
### Requirement: cash-service MUST 提供 VIP 商品 Admin 读与写

cash-service MUST 提供 `GET /cash/admin/api/vip/product` 与 `POST /cash/admin/api/vip/product`。鉴权 MUST 与既有 cash Admin 一致（网关注入 `X-Admin-Password`）。GET MUST 返回一期 VIP 商品（`productCode` 为 `vip_monthly_19`）的 `title`、`priceFen`、`originalPriceFen`、`durationDays`、`appleProductId`、`status`。POST MUST 更新该行对应字段并持久化到 `ai_voice_cash.vip_product`；MUST NOT 创建第二行 VIP SKU，MUST NOT 修改 `product_code`。`priceFen` MUST 为非负整数（分）；`originalPriceFen=0` 表示无划线价。gateway-app MUST 经既有 `/cash/admin/api/*` 反代到达，MUST NOT 新增顶层前缀。本路径为运维通道，MUST NOT 计入 App usage 统计，MUST NOT 为此修改 `maintenance_skip.go`。

#### Scenario: 管理员读取 VIP 商品

- **WHEN** 已鉴权管理员请求 `GET /cash/admin/api/vip/product`
- **THEN** 系统 MUST 返回 `vip_monthly_19` 的现价、原价、时长与 Apple 商品 ID

#### Scenario: 管理员更新现价与 Apple 商品 ID

- **WHEN** 已鉴权管理员 POST 合法的 `priceFen` 与 `appleProductId`
- **THEN** 系统 MUST 写入 `vip_product`，且后续 App 读价与建单 MUST 使用新现价，Apple 建单/验单 MUST 使用新 `appleProductId`

#### Scenario: 未鉴权拒绝写

- **WHEN** 请求未携带有效 Admin 凭证
- **THEN** 系统 MUST 拒绝 POST 且 MUST NOT 修改 `vip_product`

### Requirement: VIP 与功能建单 MUST 只认库内 Apple 商品 ID

`GetActiveProduct`、VIP Apple 验单以及 EnsureSchema 种子 MUST NOT 读取 `CASH_APPLE_PRODUCT_ID` 或 yaml `cash.appleProductId`。VIP 的 Apple 商品 ID 真相源 MUST 仅为 `vip_product.apple_product_id`。当该字段为空时，VIP Apple 建单 MUST NOT 伪造 productId；响应可带说明需在开通功能管理配置的 `payTip`。部署模板（`.env.example`、cash-service compose 环境注入、cash-service 配置注释）MUST 删除 `CASH_APPLE_PRODUCT_ID`。`CASH_APPLE_BUNDLE_ID` MUST 保留（JWS bundle 校验）。功能 SKU 的 Apple ID 仍在 `feature_product`，MUST NOT 改回 env。

#### Scenario: 库中已配置 Apple 商品 ID

- **WHEN** `vip_product.apple_product_id` 为非空且进程环境未设置 `CASH_APPLE_PRODUCT_ID`
- **THEN** `GET /cash/app/api/vip/product` 与 VIP Apple 建单 MUST 返回该库值

#### Scenario: 库中未配置 Apple 商品 ID

- **WHEN** `vip_product.apple_product_id` 为空
- **THEN** 系统 MUST NOT 从环境变量填入 Apple 商品 ID

#### Scenario: 部署模板无 CASH_APPLE_PRODUCT_ID

- **WHEN** 运维按 `.env.example` 与 compose 部署 cash-service
- **THEN** 模板与 compose 注入 MUST NOT 再出现 `CASH_APPLE_PRODUCT_ID`

### Requirement: 开通功能管理页 MUST 可编辑 VIP 套餐

`cash-feature-admin.html` MUST 提供独立于功能定义列表的 VIP 套餐表单，字段至少含现价（分）、原价（分）、时长、Apple 商品 ID；保存 MUST 调用 VIP 商品 Admin 写接口。页面 MUST NOT 把 VIP 渲染为一条可点选的 `feature_def`。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 保存 VIP 价格

- **WHEN** 管理员在开通功能管理填写 VIP 现价并保存
- **THEN** 页面 MUST 经 adminFetch POST Admin VIP 商品接口且成功后主内容反映新值

#### Scenario: VIP 不出现在功能定义列表

- **WHEN** 管理员查看开通功能管理的功能定义列表
- **THEN** 列表 MUST NOT 将 `vip` 或 `vip_monthly_19` 作为功能编号行展示

---

## vip-product-pricing

<!-- source: openspec/specs/vip-product-pricing/spec.md -->

# vip-product-pricing Specification

## Purpose
TBD - created by archiving change vip-price-db. Update Purpose after archive.
## Requirements
### Requirement: VIP 商品承载现价与原价

`ai_voice_cash.vip_product` MUST 提供现价字段 `price_fen` 与原价字段 `original_price_fen`（单位：分）。`original_price_fen=0` 表示无划线原价。改价 MUST 通过更新该表完成；写路径 MUST 为开通功能管理所调用的 VIP 商品 Admin API（及等价手工 SQL）。系统 MUST NOT 使用环境变量或 yaml 作为定价或 Apple 商品 ID 的真相源。

#### Scenario: 原价为零

- **WHEN** 商品行 `original_price_fen` 为 0
- **THEN** 读价 API MUST 返回 `originalPriceFen=0`（客户端可不展示划线价）

#### Scenario: 原价与现价均可读

- **WHEN** 商品行 `price_fen=1900` 且 `original_price_fen=9900`
- **THEN** 读价 API MUST 分别返回对应分值

#### Scenario: Admin 改价后读价一致

- **WHEN** 管理员经 VIP 商品 Admin API 将 `price_fen` 更新为新值
- **THEN** `GET /cash/app/api/vip/product` 与后续建单 `amount_fen` MUST 使用该新值

### Requirement: 匿名读取 VIP 商品价格

cash-service MUST 提供 `GET /cash/app/api/vip/product`，响应至少包含 `productCode`、`priceFen`、`originalPriceFen`、`durationDays`。该接口 MUST 允许未携带 Bearer / 未注入有效 `X-Internal-Wx-Id` 的调用方成功读取。gateway-app MUST 将该路径列入 Bearer 鉴权白名单（精确 GET）。支付宝等建单路径 MUST 仍要求登录，且建单金额 MUST 使用库内 `price_fen`。

#### Scenario: 未登录拉价成功

- **WHEN** 客户端不带 Authorization 请求 `GET /cash/app/api/vip/product`
- **THEN** gateway MUST 放行且 cash-service MUST 返回上架商品的现价与原价

#### Scenario: 建单仍用库现价

- **WHEN** 已登录用户创建支付宝 VIP 订单
- **THEN** 订单 `amount_fen` MUST 等于当时 `vip_product.price_fen`，MUST NOT 使用请求体中的任意标价覆盖

---

## vip-quota-entitlement

<!-- source: openspec/specs/vip-quota-entitlement/spec.md -->

# vip-quota-entitlement Specification

## Purpose
TBD - created by archiving change vip-quota-joint-entitlement. Update Purpose after archive.
## Requirements
### Requirement: 统一 premium 权益判定

系统 MUST 通过公共方法（或等价单一出口）计算业务 LLM 是否走 premium 路径：`isPremium = isVip(wxId) OR hardwarePrivilege OR quota.Allowed(feature)`。VIP 判定 MUST 经 cash-service 内部契约按 `wxId` 读取；查询失败 MUST 视为非 VIP 并记录 Warning，MUST NOT 因此单独使无关主流程硬失败（调用方仍可走非 premium）。硬件特权 MUST 仅由显式入口标记赋予，MUST NOT 仅因 `wxId=0` 推断为硬件特权。

#### Scenario: VIP 账号为 premium

- **WHEN** cash 返回该 `wxId` 为 VIP 且 privilege 为 Account
- **THEN** `isPremium` MUST 为 true，无论该 feature 月度 used 是否已达 limit

#### Scenario: 非 VIP 但额度允许

- **WHEN** 非 VIP、非硬件特权，且对应 feature 的额度快照 `Allowed=true`
- **THEN** `isPremium` MUST 为 true

#### Scenario: 非 VIP 且额度用尽

- **WHEN** 非 VIP、非硬件特权，且额度 `Allowed=false`
- **THEN** `isPremium` MUST 为 false

#### Scenario: 硬件特权为 premium

- **WHEN** 请求携带硬件特权标记（如 `/voice/chat/ws` 或 MCP/internal text chat 入口所置）
- **THEN** `isPremium` MUST 为 true，MUST NOT 要求 `wxId>0` 或 cash VIP

#### Scenario: 未登录无硬件特权

- **WHEN** privilege 为 Account 且 `wxId<=0`（无硬件标记）
- **THEN** `isPremium` MUST 为 false（放行主流程时 MUST 走非 premium 选模），MUST NOT 仅因未登录返回 40301 而阻止已约定可匿名/设备通道以外的、本策略允许继续的 LLM 路径（各业务原有强制登录要求仍按其自身 Requirement）

### Requirement: VIP 与硬件不计次

当本次请求因 VIP 或硬件特权而 premium 时，系统 MUST NOT 对该 feature 执行成功计次（consume MUST 为 no-op）。额度读接口对 VIP 用户 MUST 将该 feature 视为有额度（`Allowed=true`，且 MUST NOT 因 used>=limit 单独返回 40302 阻断）。

#### Scenario: VIP 成功调用不扣次

- **WHEN** VIP 用户完成一次需计次的 LLM 成功路径
- **THEN** 系统 MUST NOT 增加该 feature 的月用量

#### Scenario: 硬件特权成功不扣次

- **WHEN** 带硬件特权的语音对话成功完成 LLM
- **THEN** 系统 MUST NOT 因该次成功增加 `voice_ai`（或其它）月用量

#### Scenario: VIP 读额度 API

- **WHEN** VIP 用户请求 App `ai-quota`（voice 或 ucg 域）
- **THEN** 响应中对应 feature MUST `allowed=true`（即使 used 已达历史 limit）

### Requirement: 原子选模出口

所有向 Python 传递模型配置的业务路径，以及 Go 直调 aimodel 且受本策略约束的路径（含 polish），MUST 经公共选模出口决策：premium 时使用该 lane 正式 `provider/model`；非 premium 时使用该 lane 的 free 配置；free 为空时 MUST omit model（Python 路径不出现 model/model_cfg 字段；Go 直调路径 MUST NOT 使用已废除的硬编码 Degraded* 作为真相源）。调用方 MUST NOT 在出口之外自行根据 VIP 或额度拼装另一套模型。

#### Scenario: premium 传正式模

- **WHEN** `isPremium=true` 且 lane 已配置正式模型
- **THEN** 下游请求 MUST 携带该正式 provider/model（或等价 ModelCfg）

#### Scenario: 非 premium 且 free 已配置

- **WHEN** `isPremium=false` 且 lane freeProvider/freeModel 均非空（或约定的有效 free 配置）
- **THEN** 下游 MUST 使用 free 配置模型

#### Scenario: 非 premium 且 free 为空

- **WHEN** `isPremium=false` 且 free 配置为空
- **THEN** Go→Python 请求 MUST 省略 model（及 care-alert 的 model_cfg 中的模型字段），由 Python 自行选择免费模型

---

## voice-admin-ui

<!-- source: openspec/specs/voice-admin-ui/spec.md -->

# voice-admin-ui Specification

## Purpose
TBD - created by archiving change refactor-ai-quota-domain-ownership. Update Purpose after archive.
## Requirements
### Requirement: voice-admin HTML SHALL configure voice_ai and clinic_ai quota

系统 MUST 提供独立 Admin 页面 **`/device/admin/voice-admin.html`**（静态文件 `resource/public/voice-admin.html`），包含「AI 额度」功能区：

1. **全局默认**表单（`voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`），调用 **`GET/PUT /voice/admin/api/ai-quota/default`**。
2. **用户额度分页表**（替代原「单人 override（wxId）」表单模块）：调用 **`GET /voice/admin/api/ai-quota/users`** 分页加载；行内修改上限 MUST 调用既有 **`PUT /voice/admin/api/ai-quota/user`**。页面 MUST **移除**手输 wxId 加载/保存的单人 override 模块（含清除 checkbox UI）。

表格列顺序 MUST 为（左→右）：`deviceNo`、喂养已用、喂养上限、胖宝已用、胖宝上限、`wxId`、`account`、`babyName`。已用列只读；上限列可编辑。一行 MUST 同时展示喂养（`voice_ai`）与胖宝（`clinic_ai`）两额度。页面 MUST 提供按 **`deviceNo`** 查询并触发列表刷新。

页面 MUST NOT 包含润笔（polish）字段。页面 MUST 保留指向 **`/device/admin/ai-model-admin.html`** 的可见链接（文案含「模型与并发」或等价），MUST NOT 在本页恢复 LLM lane 编辑 Tab。页面 MUST 使用 `resource/public/admin-common.js` 的 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch`（或等价封装）初始化；Hub 登录后主内容区 MUST 可见且 MUST 加载全局默认与用户额度第一页。

页面说明区 MUST 列出额度变更影响的业务入口摘要：

- **喂养 AI**：`/voice/chat/ws`；`POST /voice/internal/api/text/chat`；`POST /voice/internal/api/text/chat/stream`；`POST /device/history/api/chat`；`POST /device/history/api/chat/stream`；`POST /voice/text/chat`。
- **胖宝 AI**：`/voice/clinic/ws`。

#### Scenario: 管理员修改喂养 AI 全局默认

- **WHEN** 运维在 voice-admin 页提交 voiceAi=5、clinicAi=30
- **THEN** 页面 SHALL 调用 PUT `/voice/admin/api/ai-quota/default` 且 voice-service 配置 SHALL 更新

#### Scenario: 页面不含润笔字段

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示 `polishMonthlyLimit` 输入控件

#### Scenario: 无单人 override 模块

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示手输 wxId 的「单人 override」加载/保存表单

#### Scenario: 额度表列顺序与双额度

- **WHEN** 列表返回至少一行
- **THEN** 表头第一列 SHALL 为 deviceNo，且同一行 SHALL 含喂养已用/上限与胖宝已用/上限

#### Scenario: 按 deviceNo 查询

- **WHEN** 运维输入 deviceNo 并查询
- **THEN** 页面 MUST 请求 `/voice/admin/api/ai-quota/users` 且携带该 deviceNo 过滤参数并刷新表格

#### Scenario: 行内修改上限

- **WHEN** 运维修改某行喂养或胖宝上限并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/ai-quota/user`（含对应 wxId 与限额字段）

#### Scenario: Hub 登录后主面板可见

- **WHEN** 运维已在 `/device/admin` Hub 登录并打开 voice-admin.html
- **THEN** 页面 MUST 展示全局默认与用户额度表（非仅页头标题）

#### Scenario: 链至统一模型页

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 MUST 展示指向 ai-model-admin 的链接且 MUST NOT 含 LLM lane 编辑表单

### Requirement: Admin Hub SHALL link voice-admin alongside ucg-admin

`resource/public/admin-modules.js` MUST 增加 voice-admin 模块入口，模式对齐 ucg-admin：`id: voice-admin`、导航至 `/device/admin/voice-admin.html`。Hub 登录后 MUST 可点击进入 voice-admin。

#### Scenario: Hub 导航可见 voice-admin

- **WHEN** 管理员登录 Admin Hub
- **THEN** 模块列表 SHALL 包含 voice-admin 入口且链接至 `/device/admin/voice-admin.html`

### Requirement: ucg-admin SHALL remove voice and clinic quota fields

`resource/public/ucg-admin.html`「AI 配置」Tab MUST **移除** `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit` 相关表单与 API 字段，**仅保留** `polishMonthlyLimit` 全局默认与 per-wxId override。同一 Tab MUST 扩展 **润笔 lane** 的 `provider`、`maxInFlight`、`maxWaiters` 配置（`visionModel` 作为 polish 的 model 选择），并调用扩展后的 **`/ucg/admin/api/ai-config`**。

#### Scenario: ucg-admin 仅润笔配置

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 SHALL 展示 polish 相关字段（含模型、并发、缓冲池）且 SHALL NOT 调用 voice/clinic 配额 API

---

## voice-ai-quota

<!-- source: openspec/specs/voice-ai-quota/spec.md -->

# voice-ai-quota Specification

## Purpose
TBD - created by archiving change refactor-ai-quota-domain-ownership. Update Purpose after archive.
## Requirements
### Requirement: voice-service SHALL be authoritative for voice_ai and clinic_ai quota configuration and usage

`voice-service` MUST 在 **`ai_voice_voice`** 库（GoFrame `database.default`，连接 `VOICE_DB_LINK`）维护 AI 月度额度全局默认与 per-wxId override，并 MUST 为 **`voice_ai`** 与 **`clinic_ai`** 两个 feature 独立计数。全局默认 MUST 包含 `voiceAiMonthlyLimit`（初始 **5**）与 `clinicAiMonthlyLimit`（初始 **30**）；Admin 可独立修改。per-wxId override MAY 单独覆盖任一 feature；未 override 的 feature MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 **`ai:usage:voice_ai:{wxId}:{YYYYMM}`** 与 **`ai:usage:clinic_ai:{wxId}:{YYYYMM}`**，其中 `YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。voice-service MUST NOT 将 voice/clinic 配额配置或用量写入 device 或 ucg 库表。

#### Scenario: 全局默认独立配置

- **WHEN** 管理员将全局 `voiceAiMonthlyLimit` 设为 5 且 `clinicAiMonthlyLimit` 设为 30
- **THEN** 无 override 的用户喂养 AI 上限 SHALL 为 5、胖宝 AI 上限 SHALL 为 30

#### Scenario: 单人 override 覆盖单 feature

- **WHEN** wxId=1001 的 override 为 `clinicAiMonthlyLimit=50` 且未设 `voiceAiMonthlyLimit`
- **THEN** 该用户胖宝 AI 上限 SHALL 为 50 且喂养 AI SHALL 使用全局默认

### Requirement: voice internal quota check and consume APIs SHALL enforce wxId and feature semantics

`POST /voice/internal/api/ai-quota/check` MUST 要求有效 voice internal 密钥（与 voice-service 其它 internal API 一致）。body MUST 含 `wxId`（正整数）与 `feature`（`voice_ai` | `clinic_ai`）。响应 MUST 含 `allowed`（boolean）、`used`、`limit`。`check` MUST NOT 修改用量。

`POST /voice/internal/api/ai-quota/consume` MUST 在 AI 成功返回后由 voice 本进程调用；成功时 MUST `INCR` 当月用量并返回 `{ used, limit }`。若扣减后 `used > limit`，系统 MUST 回滚该次 INCR 并返回超额错误。

#### Scenario: check 不扣减

- **WHEN** voice 在胖宝提问前调用 check feature=clinic_ai 且 used=29、limit=30
- **THEN** 响应 `allowed=true` 且 Redis 计数 SHALL 仍为 29

#### Scenario: consume 成功扣减

- **WHEN** voice 在 LLM 成功后调用 consume feature=voice_ai 且 used=4、limit=5
- **THEN** used SHALL 变为 5 且响应 used=5

#### Scenario: wxId 无效拒绝

- **WHEN** internal API 收到 wxId=0
- **THEN** 系统 SHALL 返回错误且 MUST NOT 读写的用量

### Requirement: voice App quota read API SHALL expose voiceAi and clinicAi

`GET /voice/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ voiceAi: { used, limit }, clinicAi: { used, limit } }`，对应当月上海时区桶。本接口 MUST NOT 返回 `polish` 字段。

#### Scenario: 登录用户查询 voice 域额度

- **WHEN** wxId=1001 请求 `/voice/app/api/ai-quota` 且当月胖宝已用 5、上限 30
- **THEN** `clinicAi.used` SHALL 为 5 且 `clinicAi.limit` SHALL 为 30

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: voice admin SHALL configure global default and per-user override locally

voice-service MUST 提供 `GET/PUT /voice/admin/api/ai-quota/default` 与 `GET/PUT /voice/admin/api/ai-quota/user`（query/body 含 `wxId`），以及 **`GET /voice/admin/api/ai-quota/users`**（用户额度分页列表，见 ADDED 要求）。认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（gateway 经 `VOICE_ADMIN_PASSWORD` 注入）。voice-service MUST 本地持久化至 `ai_voice_voice`，MUST NOT 转发 device 或 ucg 写配额。PUT default MUST 接受 `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit`（正整数）。PUT user MUST 接受 optional 两字段；空值 SHALL 表示清除该 feature override。运维浏览全量用户额度 MUST 使用 users 列表 API，而非依赖仅按 wxId 点查的 UI 作为唯一入口。

#### Scenario: 管理员修改全局胖宝默认

- **WHEN** 管理员 PUT default 为 voiceAi=5、clinicAi=30
- **THEN** voice 权威配置 SHALL 更新且新用户 check clinic_ai SHALL 使用 limit=30

#### Scenario: voice admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

#### Scenario: 列表 API 与单人写并存

- **WHEN** 管理员先 GET users 再对其中一行 PUT user 修改 voiceAiMonthlyLimit
- **THEN** 后续 GET users 同 wx 行的 `voiceAi.limit` SHALL 反映新有效上限

### Requirement: voice feeding LLM SHALL 经 voiceUnderstanding lane 闸门

voice-service 对 feature `voice_ai` 的额度 check 通过后、调用任何喂养 voice LLM 前 MUST 经 `LaneVoiceUnderstanding` 闸门。队列满 MUST 返回 WS code **50301**。本要求与 `ai-monthly-quota` 中喂养 AI 条款一致，强调全部 LLM 路径（含 casual 流式与成长建议）无遗漏。

#### Scenario: 闲聊流式占用 voiceUnderstanding 闸门

- **WHEN** commit 后进入 `StreamCasualReplyWithBaiduTTS` 的 LLM 段
- **THEN** MUST 使用 voiceUnderstanding profile 的 model 闸门

### Requirement: voice admin SHALL list per-wx effective quota with identity fields

voice-service MUST 提供 `GET /voice/admin/api/ai-quota/users`，认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（gateway 经 `VOICE_ADMIN_PASSWORD` 注入）。

Query MUST 支持：`page`（从 1）、`pageSize`（默认 20，最大 100）、**`deviceNo`**（可选，按设备号过滤）。实现 MUST 经 **`DeviceAdmin().ListWxPage`**（或等价 device 契约）取得真实 wx 分页全集（排除模拟号），MUST NOT 在 voice 进程直查 device/history 库表。

响应 MUST 为分页结构 `{ list, total, page, pageSize }`。`list[]` 每项 MUST 含：

- `deviceNo`（string）
- `wxId`（int64，对应 wx.id）
- `account`（string，对应 wx.account）
- `babyName`（string，无档案时为空串）
- `voiceAi`：`{ used, limit }`（当月上海时区桶；`limit` 为有效上限 = 正数 override ∨ 全局默认）
- `clinicAi`：`{ used, limit }`（同上）

`used` MUST 来自既有 Redis usage 键（经 `cachekit`）；键不存在时 MUST 为 0。本接口 MUST NOT 修改用量或 override。本接口 MUST NOT 返回 polish 字段。

改上限 MUST 继续使用既有 `PUT /voice/admin/api/ai-quota/user`；当提交的某 feature 上限等于当前全局默认时，该 feature MUST 清除 override（写空/NULL），使该用户后续跟随全局默认。

#### Scenario: 分页返回身份与双额度

- **WHEN** 管理员请求 `/voice/admin/api/ai-quota/users?page=1&pageSize=20` 且 device 域存在真实 wx
- **THEN** 响应 `list` 每项 SHALL 含 deviceNo、wxId、account、babyName 以及 voiceAi/clinicAi 的 used 与 limit

#### Scenario: 按 deviceNo 过滤

- **WHEN** 管理员请求 `deviceNo` 等于某已绑定设备号
- **THEN** 返回列表 SHALL 仅包含该设备号匹配的 wx 行（在契约过滤语义下）

#### Scenario: 无 override 时 limit 为全局默认

- **WHEN** 某 wx 无 clinic override 且全局 clinicAiMonthlyLimit=30
- **THEN** 该行 `clinicAi.limit` SHALL 为 30

#### Scenario: 有 override 时 limit 为覆盖值

- **WHEN** wxId=1001 的 voice override 为 10 且当月 voice 已用 3
- **THEN** 该行 `voiceAi.limit` SHALL 为 10 且 `voiceAi.used` SHALL 为 3

#### Scenario: 口令错误拒绝

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回用户额度列表

#### Scenario: 保存等于默认则清除 override

- **WHEN** 管理员 PUT user 将某 wx 的 clinicAiMonthlyLimit 设为与当前全局 clinic 默认相同的正整数
- **THEN** 该 wx 的 clinic override SHALL 被清除（后续 limit 跟随全局默认）

### Requirement: voice 域额度与 VIP 共同策略

`voice_ai` 与 `clinic_ai` 的 check/快照/App 状态 MUST 与 VIP 共同决策：VIP 时 MUST `Allowed=true` 且成功路径不计次。非 VIP 时保持月度 used/limit 语义；用尽后 MUST 允许继续走非 premium 选模（free/omit），MUST NOT 再依赖硬编码 Degraded 智谱种子作为唯一降级模型。

#### Scenario: 非 VIP 额度内喂养

- **WHEN** 非 VIP 且 `voice_ai` Allowed，完成喂养 LLM 成功
- **THEN** 系统 MUST 使用 `voiceUnderstanding` 正式模型，并 consume `voice_ai`

#### Scenario: 非 VIP 喂养额尽

- **WHEN** 非 VIP 且 `voice_ai` 用尽
- **THEN** 对话 MUST 仍可继续，选模 MUST 走 `voiceUnderstanding` free（或 omit），且该次成功 MUST NOT consume

### Requirement: tip 挂靠 clinic 额度与 lane

小贴士（TipStream）MUST 使用与 polish/诊疗相同的 VIP∪额度策略，feature 为 **`clinic_ai`**，lane 为 **`clinic`**（含 free）。MUST NOT 在无权益判定的情况下无脑传正式模型。

#### Scenario: tip 与诊疗共享 clinic_ai

- **WHEN** 非 VIP 用户 `clinic_ai` 仍有额度并成功生成 tip
- **THEN** 系统 MUST 按 clinic 正式模调用 Python，并 consume `clinic_ai`

#### Scenario: tip 在 clinic_ai 用尽时

- **WHEN** 非 VIP 且 `clinic_ai` 用尽后请求 tip
- **THEN** 系统 MUST 按 clinic free/omit 调用，MUST NOT 因 40302 单独拒绝 tip（与 clinic degraded 放行策略对齐）

### Requirement: 硬件语音通道特权

经 gateway 设备直连的 `/voice/chat/ws` 以及 MCP/internal text chat 设备通道，MUST 标记硬件特权并按 premium 选模（`voiceUnderstanding` 正式模），MUST NOT 计次。该特权 MUST NOT 自动授予 clinic WS、care-alert、tip、polish。

#### Scenario: MCP 文本对话走正式模不计次

- **WHEN** mcp-service 以配置 deviceNo 调用 internal text chat 成功
- **THEN** 意图分析 MUST 使用 `voiceUnderstanding` 正式模型，MUST NOT consume `voice_ai`

---

## voice-and-device-service-decomposition

<!-- source: openspec/specs/voice-and-device-service-decomposition/spec.md -->

# voice-and-device-service-decomposition Specification

## Purpose
TBD - created by archiving change platform-hardening-redis-rabbitmq-service-split. Update Purpose after archive.
## Requirements
### Requirement: Voice 与 Device 领域逻辑必须运行在独立服务中
系统 SHALL 将 voice 领域与 device 领域的业务逻辑部署到独立可部署服务，并 SHALL 定义 gateway 调用所需的明确内部服务契约。

#### Scenario: Voice 请求路由
- **WHEN** gateway 收到 voice 领域请求
- **THEN** gateway SHALL 按既定内部契约将请求路由到 `voice-service`，而不是在本地执行 voice 业务逻辑

#### Scenario: Device 请求路由
- **WHEN** gateway 收到 device 领域请求
- **THEN** gateway SHALL 按既定内部契约将请求路由到 `device-service`，而不是在本地执行 device 业务逻辑

### Requirement: 服务边界遵循领域数据归属
系统 SHALL 按当前数据库/领域归属划分服务边界，并 SHALL 通过显式服务接口处理跨领域访问。

#### Scenario: Voice 流程需要 Device 领域数据
- **WHEN** `voice-service` 需要访问 device 领域数据
- **THEN** `voice-service` SHALL 通过契约化内部 API 或事件交互获取数据，而不是直接嵌入 device 领域实现

---

## voice-chat-dashscope-stt

<!-- source: openspec/specs/voice-chat-dashscope-stt/spec.md -->

# voice-chat-dashscope-stt Specification

## Purpose
TBD - created by archiving change voice-chat-dashscope-stt. Update Purpose after archive.
## Requirements
### Requirement: 对话 WebSocket SHALL 使用百炼流式 STT（chat profile）

`voice-service` 在处理 `/voice/chat/ws` 且输入模态为 `audio` 时，流式 STT MUST 使用 **chat profile** 配置（`sttChat`），provider MUST 为 `dashscope`，默认模型 MUST 为 `qwen-audio-3.0-asr-flash-streaming`。MUST NOT 在该路径默认使用百度 `dev_pid=1537` 流式 ASR。

#### Scenario: 对话 WS 建连后创建 DashScope 流式 ASR

- **WHEN** 客户端对 `/voice/chat/ws` 发送合法 `start` 且 `inputModality=audio`（或省略默认为 audio）
- **AND** `sttChat.provider=dashscope` 且 `sttChat.streamEnabled=true`
- **AND** 有效 DashScope API Key 与 Workspace ID 已配置
- **THEN** 服务端 MUST 通过百炼 WebSocket 建立流式 ASR 会话
- **AND** 服务端 MUST NOT 为该会话调用百度流式 ASR 建连

#### Scenario: DashScope 凭证缺失

- **WHEN** 客户端对 `/voice/chat/ws` 开始 audio 会话
- **AND** chat profile 的 API Key 或 Workspace ID 未配置
- **THEN** 服务端 MUST 返回 `error` 帧且 `stage=stt`
- **AND** MUST NOT 静默降级为无 ASR 继续对话

### Requirement: 听写 WebSocket SHALL 继续使用百度 STT（dictation profile）

`/voice/asr/ws` 流式 STT MUST 使用 **dictation profile** 配置（`sttDictation` 或与现有 `stt` 块等价），provider MUST 保持 `baidu`。本变更 MUST NOT 改变听写端点的 WS 协议与事件类型。

#### Scenario: 听写 WS 仍走百度

- **WHEN** 客户端对 `/voice/asr/ws` 发送合法 `start`
- **AND** `sttDictation.provider=baidu` 且 `streamEnabled=true`
- **THEN** 服务端 MUST 使用百度流式 ASR（`CreateStreamASRSession` dictation profile）
- **AND** 下行 MUST 仍支持 `asr_partial`、`asr_final`（由 client commit/end 触发定稿）

### Requirement: CreateStreamASRSession SHALL 支持 profile 分流

`Voice().CreateStreamASRSession`（及 `VoiceContract` 等价接口）MUST 接受 `profile` 参数，取值 `chat` 或 `dictation`，并 MUST 根据 profile 选择 `sttChat` 或 `sttDictation` 配置块。

#### Scenario: chat 与 dictation 调用不同 provider

- **WHEN** `voice_ws.go` 调用 `CreateStreamASRSession(..., profile=chat, ...)`
- **THEN** 实现 MUST 读取 `sttChat` 并创建 DashScope 会话
- **WHEN** `voice_asr_ws.go` 调用 `CreateStreamASRSession(..., profile=dictation, ...)`
- **THEN** 实现 MUST 读取 `sttDictation` 并创建百度会话

### Requirement: DashScope 流式 ASR SHALL 兼容现有 PCM 参数

百炼流式 ASR 实现 MUST 接受与现有链路一致的音频元数据：`sampleRate=16000`（或 `start` 声明值）、`bits=16`、`channels=1`、PCM 小端二进制分片。`WriteAudio` / `Commit` / `Close` 语义 MUST 与 `StreamASRSession` 契约一致，以便 `/voice/chat/ws` 无需修改客户端。

#### Scenario: 对话 WS 收到 asr_partial

- **WHEN** DashScope 返回中间识别结果且文本非空
- **THEN** `/voice/chat/ws` MUST 向客户端发送 `{"type":"asr_partial","code":0,"text":"..."}`（与现有行为一致）

#### Scenario: 对话 WS commit 后 asr_final

- **WHEN** 服务端对 DashScope 会话执行 `Commit` 且得到有效转写文本
- **THEN** `/voice/chat/ws` MUST 继续现有对话链路（LLM/TTS），且 MUST 下发 `asr_final` 或等价最终文本事件（与现实现一致）

### Requirement: 对话 STT 配置 SHALL 位于 voice-chat.shared.yaml

对话 STT 配置 MUST 位于 `manifest/config/voice-chat.shared.yaml` 的 `voiceChat.sttChat`（或经 `GF_VOICE_CHAT_FILE` 加载的等价文件）。MUST NOT 将 DashScope STT 专属字段回流到 `manifest/config/config.yaml` 主网关配置。

#### Scenario: voice-service 加载 sttChat

- **WHEN** `voice-service` 启动并加载 `voice-chat.shared.yaml`
- **THEN** `sttChat.provider`、`sttChat.model`、`sttChat.streamEnabled` MUST 可供 `CreateStreamASRSession(chat)` 读取

### Requirement: 环境变量 SHALL 支持 DashScope 凭证注入

部署 MUST 支持经环境变量注入百炼凭证：`VOICE_DASHSCOPE_API_KEY`（优先）或复用 `UCG_DASHSCOPE_API_KEY`；`DASHSCOPE_WORKSPACE_ID` 用于 WebSocket 端点。`manifest/docker/.env.example` MUST 文档化上述变量。

#### Scenario: 仅配置 UCG 密钥

- **WHEN** `VOICE_DASHSCOPE_API_KEY` 未设置且 `UCG_DASHSCOPE_API_KEY` 已设置
- **AND** `sttChat.apiKey` 配置为空
- **THEN** chat profile STT MUST 使用 `UCG_DASHSCOPE_API_KEY` 鉴权

---

## voice-device-domain-http-access

<!-- source: openspec/specs/voice-device-domain-http-access/spec.md -->

# voice-device-domain-http-access Specification

## Purpose
TBD - created by archiving change enforce-http-only-cross-service-no-foreign-dao. Update Purpose after archive.
## Requirements
### Requirement: Voice MUST 经 HTTP 访问 device 领域持久化数据

在 **voice-service** 进程内，凡涉及 `user`、`event`、`action` 等他域表的读取或写入（含语音意图、DeepSeek 实体抽取、动作词典维护等），MUST 通过 **device-service** 暴露的 HTTP 接口完成；MUST NOT 在 `voice` 包或 voice 进程内调用 `device` 包中会触发他域 `dao.User`、`dao.Event`、`dao.Action` 的实现路径。

#### Scenario: 语音链路查询事件列表

- **WHEN** voice 需要加载事件字典以匹配用户说法
- **THEN** voice MUST 向 device-service 发起 HTTP 请求获取列表，MUST NOT 使用本进程 default 数据库连接访问 `event` 表

#### Scenario: 语音链路写入动作或事件

- **WHEN** voice 需要将新动作或事件变更持久化
- **THEN** voice MUST 调用 device-service HTTP 接口完成写入，MUST NOT 在 voice 进程内执行 `dao.Action` 或 `dao.Event` 的 Insert/Update

### Requirement: 迁移期 local 路径 MUST 仍为 HTTP 到 device 入口

若配置为「本地」模式以简化联调，其语义 MUST 为调用 **本机或可解析的 device-service 基址**（如 `http://127.0.0.1:9803`）的 HTTP，MUST NOT 解释为在同一进程内直接执行他域 DAO。

#### Scenario: 开发单机多端口

- **WHEN** 开发者在同一主机分别启动 voice 与 device 监听不同端口
- **THEN** voice 的 local 配置 MUST 仍指向 device HTTP 基址，而非共享同一 ORM 连接访问 device 库

---

## voice-device-profile-http-contract

<!-- source: openspec/specs/voice-device-profile-http-contract/spec.md -->

# voice-device-profile-http-contract Specification

## Purpose
TBD - created by archiving change enforce-service-boundary-no-cross-db. Update Purpose after archive.
## Requirements
### Requirement: Voice MUST 通过 Device 服务契约获取设备画像数据
`voice-service` 在涉及设备信息、用户生日、性别、注册状态等 device/profile 领域数据时 MUST 通过 `device-service` 暴露的内部 HTTP 接口获取，MUST NOT 直接访问 `user/device` 领域数据库表。

#### Scenario: 通用问答需要设备画像
- **WHEN** voice 在生成通用问答提示词时需要生日或性别等画像信息
- **THEN** voice MUST 调用 device 内部接口获取画像数据，并将结果用于提示词构建

#### Scenario: 设备信息接口不可达
- **WHEN** voice 调用 device 内部画像接口超时或网络失败
- **THEN** voice MUST 返回可观测的错误语义，并按配置决定是否执行迁移期兜底

### Requirement: Device 内部画像接口 MUST 提供一致错误结构
device 内部画像接口 MUST 对参数错误、设备不存在、服务异常返回统一错误结构，供 voice 侧做稳定错误映射。

#### Scenario: 设备不存在
- **WHEN** voice 传入的 `deviceNo` 在 device 服务中不存在
- **THEN** device MUST 返回可区分的业务错误码，voice MUST 返回可理解的失败信息

#### Scenario: 参数缺失
- **WHEN** voice 调用画像接口时缺失关键参数
- **THEN** device MUST 返回参数错误结构，voice MUST 记录请求参数异常日志

### Requirement: 画像链路的「本地」实现 MUST 不依赖进程内 user DAO

即使存在 `localDeviceProfileAdapter` 类实现，`voice-service` 在生产单库模式下 MUST 使用 **HTTP 远程实现**（或指向 device 基址的 HTTP local）；MUST NOT 依赖在 voice 进程内对 `dao.User` 的查询作为获取画像的主路径。

#### Scenario: 生产 voice 配置

- **WHEN** 部署 `voice-service` 且 `database.default` 仅含 voice 域表
- **THEN** 设备画像 MUST 通过 `device-service` 的 HTTP 接口获取；若误配为进程内 local 适配器，MUST 视为配置错误并暴露启动或运行期检查（若实现）

---

## voice-event-child-disambiguation

<!-- source: openspec/specs/voice-event-child-disambiguation/spec.md -->

# voice-event-child-disambiguation Specification

## Purpose
TBD - created by archiving change device-event-hierarchy. Update Purpose after archive.
## Requirements
### Requirement: 事件匹配时子节点优先于父节点

voice-service 在本地文本匹配事件时，SHALL 对候选事件按**深度降序、名称长度降序**排序后再匹配，使叶子或深层子节点优先于浅层父节点命中。

#### Scenario: 同时提及父名与子名时命中叶子

- **WHEN** 事件树含「换尿布(父)」与「大便(子)」
- **AND** 用户说「换尿布，拉了大便」
- **THEN** 匹配结果 SHALL 为「大便」事件 id
- **AND** SHALL NOT 进入父节点追问流程

---

## voice-history-http-contract

<!-- source: openspec/specs/voice-history-http-contract/spec.md -->

# voice-history-http-contract Specification

## Purpose
TBD - created by archiving change enforce-service-boundary-no-cross-db. Update Purpose after archive.
## Requirements
### Requirement: Voice MUST 通过 History HTTP 契约获取历史域数据

`voice-service` 在涉及 **历史记录**（由 `history` 表承载的会话/事件时间线数据）的查询或写入时 MUST 通过 `history-service` 暴露的内部 HTTP 接口完成，MUST NOT 直接访问 history 领域数据库表。用户画像（生日、性别等）、事件类型字典、动作记录等 **非 history 表** 数据 MUST NOT 通过本需求所述的 history 接口冒充权威来源，MUST 分别遵循 device 与 voice 域契约。

#### Scenario: 查询历史记录用于对话生成

- **WHEN** voice 处理“查询历史记录”或需要最近历史上下文的请求
- **THEN** voice MUST 调用 history 内部查询接口获取数据，并使用返回结果生成回复

#### Scenario: History 服务不可达

- **WHEN** voice 调用 history 内部接口超时或网络失败
- **THEN** voice MUST 返回可观测的错误语义，并按照配置决定是否执行本地兜底（仅迁移期允许）

### Requirement: History 内部接口 MUST 提供稳定错误语义
history 内部接口 MUST 对参数错误、资源不存在、服务异常返回可区分的错误结构，供 voice 做一致错误处理与日志分类。

#### Scenario: 参数不合法
- **WHEN** voice 传入缺失 `deviceNo` 或非法参数
- **THEN** history MUST 返回明确的参数错误码与错误信息，voice MUST 将其映射为调用方可理解的失败结果

#### Scenario: 服务端内部异常
- **WHEN** history 在处理请求时发生内部错误
- **THEN** history MUST 返回统一错误结构，voice MUST 记录失败原因并输出统一告警日志

### Requirement: Voice MUST 通过 Device HTTP 契约访问用户画像与事件字典

`voice-service` 需要读取或更新 **设备用户画像**（如生日、性别）或 **事件/动作** 相关持久化数据时 MUST 通过 `device-service` 暴露的 HTTP 契约完成，MUST NOT 使用 `dao.User`、`dao.Event`、`dao.Action` 直连 device 库表。

#### Scenario: 事件抽取结果落库

- **WHEN** voice 在理解流程中创建或解析事件实体并需持久化
- **THEN** voice MUST 调用 device 服务接口（或已批准的适配层），MUST NOT 在 voice 进程内对 `event` 表执行 DAO Insert
- **AND** 新建事件时 MUST 向 device 传递合法 `eventType`（`number` | `time` | `one`）

#### Scenario: 读取事件选项列表

- **WHEN** voice 需要事件字典列表或 `eventType` 等元数据
- **THEN** voice MUST 从 device 服务获取，MUST NOT 依赖 history 服务返回的 `event` 表投影作为权威来源
- **AND** 响应项 SHALL 含 `eventType`，SHALL NOT 含 `needQuantity`

### Requirement: Voice MUST 在本域处理 suggest 表

`voice-service` 对 **`suggest` 表** 的读写 MUST 仅在 voice 进程内通过本域 DAO 或本域服务接口完成；history-service MUST NOT 作为 suggest 数据的权威存储进程。

#### Scenario: 写入每日建议

- **WHEN** voice 生成并保存建议文案
- **THEN** 持久化 MUST 发生在 voice 库 `suggest` 表路径上，MUST NOT 由 history 进程执行 `dao.Suggest` 写入

### Requirement: Voice 写入 history SHOULD 传递已知 eventUnit

当 voice-service 经 device HTTP 契约已解析出 `entity.Event` 且 `Unit` 非空时，调用 history 新增/更新接口的请求体 SHOULD 携带 `eventUnit` 字段。history-service MUST 仍支持未携带时经 device 契约补全。

#### Scenario: 语音记录奶量事件

- **WHEN** voice 解析到事件 `unit=ml` 并调用 history 新增一条计数记录
- **THEN** HTTP 请求体 SHOULD 包含 `eventUnit=ml`，且 history 持久化后 `event_unit` SHALL 为 `ml`

#### Scenario: Voice 未传 eventUnit 时由 history 补全

- **WHEN** voice 调用 history 新增记录未传 `eventUnit` 但 `eventId` 对应主档 `unit=ml`
- **THEN** history-service MUST 经 device 契约补全并成功写入 `event_unit=ml`

---

## voice-internal-text-chat

<!-- source: openspec/specs/voice-internal-text-chat/spec.md -->

# voice-internal-text-chat Specification

## Purpose
TBD - created by archiving change history-chat-delegate-voice. Update Purpose after archive.
## Requirements
### Requirement: voice-service 提供 internal 文本对话 API

voice-service SHALL 提供 `POST /voice/internal/api/text/chat`，接受 JSON body `deviceNo`、`transcript`；调用方 MUST 在 Header 携带 `X-Device-Gateway-Internal-Secret`（与 `DEVICE_GATEWAY_INTERNAL_SECRET` 一致）。若需喂养 AI 额度预检，调用方 SHOULD 携带 `X-Internal-Wx-Id`（正整数 wxId）。

成功时响应 data MUST 含 `reply` 字符串。失败时 MUST 使用与 App 网关一致的 business code（含 40301 未登录、40302 额度用尽）。

对话能力 MUST 统一走母婴喂养模式，MUST NOT 存在闲聊模式切换。AI 推理 MUST 统一调用 Python 微服务接口（`/v1/analyze/intent`），MUST NOT 在 Go 侧保留 DeepSeek/LLM 直连兜底。Python 服务不可用时，MUST 返回固定降级提示语"AI 服务暂时不可用，请稍后再试"。

#### Scenario: 合法 internal 请求成功

- **WHEN** secret 正确且 `deviceNo`、`transcript` 合法且额度允许
- **THEN** 接口 MUST 返回 `code=0` 且 `reply` 为非空或允许的空串回复

#### Scenario: 未登录 wxId

- **WHEN** 未携带有效 `X-Internal-Wx-Id` 且链路需要额度预检
- **THEN** 接口 MUST 返回 business code 40301

#### Scenario: 额度用尽

- **WHEN** wxId 有效且当月 voice_ai 额度已用尽
- **THEN** 接口 MUST 返回 business code 40302

#### Scenario: secret 无效

- **WHEN** `X-Device-Gateway-Internal-Secret` 缺失或错误
- **THEN** 接口 MUST 拒绝请求且 MUST NOT 执行 TextChat

#### Scenario: Python 微服务不可用

- **WHEN** Python 微服务（`/v1/analyze/intent`）调用失败（超时、连接失败、5xx 等）
- **THEN** 接口 MUST 返回固定降级提示语"AI 服务暂时不可用，请稍后再试"
- **AND** MUST NOT 回退到 DeepSeek/LLM 直连

#### Scenario: 闲聊模式切换命令

- **WHEN** 用户输入包含"切换到闲聊"等闲聊模式切换语义
- **THEN** 系统 MUST 按母婴喂养模式处理
- **AND** MUST NOT 切换到闲聊模式
- **AND** MUST NOT 返回闲聊相关回复

---

## voice-realtime-asr-ws

<!-- source: openspec/specs/voice-realtime-asr-ws/spec.md -->

# voice-realtime-asr-ws Specification

## Purpose
TBD - created by archiving change voice-realtime-asr-ws. Update Purpose after archive.
## Requirements
### Requirement: Voice-service SHALL 提供实时听写 WebSocket 端点

`voice-service` MUST 在路径 `/voice/asr/ws` 提供 WebSocket 服务，将客户端上行的 PCM 音频流送入已配置的流式 STT（当前为百度流式 ASR），并将识别出的中文文本实时返回给客户端。

#### Scenario: 握手成功并开始会话

- **WHEN** 客户端对 `/voice/asr/ws` 发起 WebSocket Upgrade 且握手成功
- **AND** 客户端发送合法 `start` 文本帧（含非空 `deviceNo` 与有效 `sampleRate`/`bits`/`channels`）
- **THEN** 服务端 MUST 回复 `{"type":"started","code":0,"mode":"stream"}`（或等价字段集）
- **AND** 服务端 MUST NOT 调用对话 LLM、TTS 或设备最近对话落库接口

#### Scenario: 流式 STT 未启用或不可用

- **WHEN** `start` 成功但流式 STT 配置不可用（如 `stt.streamEnabled=false` 或 provider 不支持）
- **THEN** 服务端 MUST 发送 `{"type":"error","code":1,"stage":"stt",...}` 且 MUST NOT 假装听写成功

### Requirement: 上行协议 SHALL 限定为听写所需子集

听写 WebSocket MUST 接受以下上行消息类型：

- Text JSON：`type` 为 `start`（开始会话）
- Binary：16-bit 小端 PCM 音频分片，参数与 `start` 声明一致
- Text JSON：`type` 为 `commit`（**一句听写结束**，触发 finalize 并下发 `asr_final`）
- Text JSON：`type` 为 `end`（结束当前听写 WebSocket 会话）
- Text：心跳 `ping`（服务端回复 `pong`）

#### Scenario: 客户端发送 commit

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"commit"}`
- **AND** 当前已有流式 ASR 会话或已接收过非空音频缓冲
- **THEN** 服务端 MUST 对当前句执行 STT finalize 并下发 `{"type":"asr_final",...,"source":"client"}`
- **AND** 服务端 MUST NOT 因此进入对话/TTS 链路

#### Scenario: commit 前无音频

- **WHEN** 客户端发送 `{"type":"commit"}` 且当前无 ASR 会话且无已缓冲音频
- **THEN** 服务端 MUST 返回 `error`（如 `stage=validate`）

#### Scenario: 客户端发送 end

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"end"}`
- **THEN** 服务端 MUST 关闭当前流式 ASR 会话并回复 `{"type":"ended","code":0}`
- **AND** 若关闭前仍有未 finalize 的音频，服务端 MAY 先执行 finalize 并下发 `asr_final`（`source` 为 `end`）

### Requirement: 下行协议 SHALL 仅包含听写相关事件

服务端下行 Text JSON MUST 以听写为主，至少包含：

- `asr_partial`：非空中间识别文本；**亦包含**流式 STT 引擎级 final 回调转发（与中间 partial 同为该类型，供客户端覆盖预览）
- `asr_final`：一句听写定稿文本（**仅**由客户端 `commit` 或 `end` 触发的 finalize 产生）
- `asr_no_result`：finalize 后无有效文本时
- `error`、`started`、`ended`

服务端 MUST NOT 在该端点下发 `audio_chunk`、`chat_delta`、`exit` 或 TTS 相关字段。

#### Scenario: 收到 ASR 中间结果

- **WHEN** 流式 STT 产生新的中间文本且与上次 partial 不同
- **THEN** 服务端 MUST 发送 `{"type":"asr_partial","code":0,"text":"<识别文本>"}`

#### Scenario: 收到 ASR 最终结果

- **WHEN** 客户端发送 `commit` 或 `end` 导致服务端执行 finalize 且得到有效转写文本
- **THEN** 服务端 MUST 发送 `{"type":"asr_final","code":0,"text":"<识别文本>"}` 且 `source` MUST 为 `client` 或 `end`

### Requirement: 听写连接 SHALL 与对话连接隔离

`voice-service` 在处理 `/voice/asr/ws` 时 MUST NOT 将连接注册到用于 `/voice/chat/ws` 的「单设备单连接」替换管理器（`VoiceWSManager`），以避免听写页与对话页互相踢连接。

#### Scenario: 同一 deviceNo 同时存在 chat 与 asr 连接

- **WHEN** 设备 `device-001` 已建立 `/voice/chat/ws` 连接且另建 `/voice/asr/ws` 连接
- **THEN** 两条连接 MUST 均可保持，直至各自关闭

### Requirement: Voice 域边界 SHALL 保持不变

听写实现 MUST 仅使用 voice 域已有 STT 能力与配置（`voice-chat.shared.yaml` / `Voice().CreateStreamASRSession`），且 MUST NOT 在 voice-service 内直接访问 device/history/user 等他域数据库表。

#### Scenario: 听写会话不查 device 库

- **WHEN** 客户端在 `start` 中提供 `deviceNo`
- **THEN** 服务端 MAY 将其用于日志与限流键，且 MUST NOT 为听写路径新增对 device 表 DAO 的依赖

### Requirement: 听写 WS MUST NOT 服务端主动截句

`/voice/asr/ws` 实现 MUST NOT 基于服务端静音计时、STT 回调间隔或无回调时长自动调用 STT finalize 或向客户端发送 `asr_final`。

#### Scenario: 长时间静音但未发送 commit

- **WHEN** 客户端已 `start` 并持续发送二进制音频或保持连接
- **AND** 客户端在超过 2 秒内未发送 `commit` 或 `end`
- **AND** 流式 STT 未产生新的 partial 或产生引擎 final 回调
- **THEN** 服务端 MUST NOT 因静音或超时自动下发 `asr_final`
- **AND** 服务端 MUST NOT 因上述原因自动关闭并重建 ASR 会话

#### Scenario: 引擎 onFinal 回调

- **WHEN** 流式 STT 提供商推送引擎级 final 结果（如百度 `FIN_TEXT`）且文本非空
- **THEN** 服务端 MUST NOT 将该结果作为 `asr_final` 下发给客户端
- **AND** 服务端 MUST NOT 仅因该回调而关闭当前 ASR 会话
- **AND** 若该文本与上次已下发的 `asr_partial` 的 `text` 不同，服务端 MUST 再发送一条 `{"type":"asr_partial","code":0,"text":"<识别文本>"}` 以供客户端更新预览

#### Scenario: 禁止的 asr_final source

- **WHEN** 服务端在听写 WS 上产生 `asr_final`
- **THEN** `source` MUST NOT 为 `silence`、`auto_commit` 或 `asr_callback`

---

## voice-route-canary-management

<!-- source: openspec/specs/voice-route-canary-management/spec.md -->

# voice-route-canary-management Specification

## Purpose
TBD - created by archiving change voice-device-canary-route-split. Update Purpose after archive.
## Requirements
### Requirement: Gateway MUST 为 voice 路由提供独立可配置代理能力
gateway MUST 以独立中间件管理 `/voice/text/*` 路由，并支持 `local|proxy|canary` 三态。

#### Scenario: voice 路由进入 local 模式
- **WHEN** `VOICE_API_ROUTE_MODE=local`
- **THEN** gateway MUST 执行本地处理链路，且 MUST NOT 将请求转发到 voice-service

#### Scenario: voice 路由进入 proxy 模式
- **WHEN** `VOICE_API_ROUTE_MODE=proxy` 且 `VOICE_API_PROXY_URL` 可用
- **THEN** gateway MUST 将 `/voice/text/*` 请求全量转发到 voice-service

#### Scenario: voice 路由进入 canary 模式
- **WHEN** `VOICE_API_ROUTE_MODE=canary` 且配置了 `VOICE_API_PROXY_CANARY_PERCENT`
- **THEN** gateway MUST 按稳定分流键执行百分比转发，其余请求保持本地处理

### Requirement: voice canary 分流 MUST 保持同键稳定
gateway MUST 采用稳定分流键（如 deviceNo）对 canary 流量做无状态一致性计算。

#### Scenario: 同一分流键连续请求
- **WHEN** 同一设备在 canary 模式下发起多次 `/voice/text/*` 请求
- **THEN** 请求 MUST 稳定命中同一流量路径（proxy 或 local）

---

## voice-textchat-resilience

<!-- source: openspec/specs/voice-textchat-resilience/spec.md -->

# voice-textchat-resilience Specification

## Purpose
TBD - created by archiving change textchat-mq-publish-best-effort. Update Purpose after archive.
## Requirements
### Requirement: TextChat SHALL 不依赖 voice.task 事件链路

TextChat 对话完成路径 MUST NOT 发布或依赖 `voice.task.requested` 事件；worker-service consumer 已删除，对话持久化与业务逻辑 MUST 在 voice-service 请求内完成。

#### Scenario: 对话完成无 task 发布

- **WHEN** TextChat 会话正常结束
- **THEN** 系统 MUST NOT 调用 `publishTaskRequested` 或等价 MQ 发布

---

## wechat-ios-universal-links

<!-- source: openspec/specs/wechat-ios-universal-links/spec.md -->

# wechat-ios-universal-links Specification

## Purpose
TBD - created by archiving change wechat-ios-universal-links. Update Purpose after archive.
## Requirements
### Requirement: Apple AASA 文件 SHALL 在 `www.pangbao.cuplay.top` 主机根路径可访问
系统 SHALL 为胖宝 iOS 应用提供 Apple `apple-app-site-association` 文件，并且 MUST 同时支持 `GET /apple-app-site-association` 与 `GET /.well-known/apple-app-site-association`。两条路径返回的内容 MUST 等价、响应状态 MUST 为 `200`、传输协议 MUST 为 HTTPS，且响应不得被改写到任何其它业务路径后才可获取。

#### Scenario: Apple 从根路径获取 AASA 文件
- **WHEN** 运维或校验程序请求 `https://www.pangbao.cuplay.top/apple-app-site-association`
- **THEN** 系统返回 `200` 和可解析的 AASA JSON 内容，且不发生 301/302 到其它路径

#### Scenario: Apple 从 well-known 路径获取 AASA 文件
- **WHEN** 运维或校验程序请求 `https://www.pangbao.cuplay.top/.well-known/apple-app-site-association`
- **THEN** 系统返回与根路径等价的 AASA JSON 内容，并使用适合 JSON/AASA 的响应头

### Requirement: AASA 内容 SHALL 与微信 Universal Links 前缀保持一致
AASA 内容 MUST 使用 `appIDs = ["<TEAM_ID>.com.fzy.pangbao"]` 的结构，其中 `com.fzy.pangbao` 为固定 Bundle ID，`<TEAM_ID>` 为部署时注入的真实 Apple Team ID。AASA `components` 或等价路径约束 MUST 放行 `https://www.pangbao.cuplay.top/wx/ulink/` 对应的 `/wx/ulink/*` 路径，使微信开放平台填写值、iOS 客户端 `universalLink` 和服务端声明保持一致。

#### Scenario: Team ID 已配置时生成正式 AASA 内容
- **WHEN** 部署环境已经提供真实 Team ID
- **THEN** AASA 中的 `appIDs` 使用 `<真实TeamID>.com.fzy.pangbao`，且放行路径覆盖 `/wx/ulink/*`

#### Scenario: 微信后台使用推荐的 Universal Links 前缀
- **WHEN** 接入人员在微信开放平台填写 Universal Links
- **THEN** 文档与服务端约束均指向 `https://www.pangbao.cuplay.top/wx/ulink/`

### Requirement: Team ID 缺失时系统 SHALL 提供显式待配置语义
在 Team ID 尚未提供的阶段，仓库 MUST 保留明确的 AASA 模板或配置占位说明；正式对外端点在未配置 Team ID 时 MUST 返回显式不可验证语义或运维可识别的失败提示，而不是伪造一个看似可用的生产 `appIDs`。

#### Scenario: 生产配置缺少 Team ID
- **WHEN** AASA 端点所在环境未设置 Team ID
- **THEN** 系统返回显式错误或不可用提示，并在日志/文档中指向需要补充的配置项

#### Scenario: 仓库中保留待补位模板
- **WHEN** 开发人员阅读仓库内 Universal Links 相关资源
- **THEN** 可以看到 Team ID 待补位规则，以及 `com.fzy.pangbao` 已固定、仅 Team ID 需要在部署前补齐

### Requirement: 仓库 SHALL 提供 GitHub 打包上架的 Universal Links 操作文档
仓库 MUST 提供面向 GitHub 打包链路的 runbook，明确 iOS 工程需要开启 `Associated Domains`、加入 `applinks:www.pangbao.cuplay.top`、保证 Provisioning Profile 启用该能力，并在微信 SDK 注册配置中使用与 AASA 一致的 `https://www.pangbao.cuplay.top/wx/ulink/`。文档 MUST 说明该流程适用于 GitHub Actions / CI 打包，不要求人工在本地 Xcode 界面逐步操作才能理解。文档 MUST 同时明确 `http://www.pangbao.cuplay.top/` 不能作为 Universal Links 或 AASA 校验地址。

#### Scenario: GitHub Actions 打包配置指引可读
- **WHEN** 维护者按照 runbook 配置 GitHub 打包环境
- **THEN** 可以明确知道需要准备哪些证书/描述文件/Secrets、如何确认 entitlements 被正确签入产物

#### Scenario: 发布后可执行 Universal Links 验证
- **WHEN** 维护者完成部署与打包
- **THEN** runbook 提供 `curl`、Apple/微信侧检查项或真机验证步骤，以确认 Universal Links 已生效

---

## wechat-oauth-platform-config

<!-- source: openspec/specs/wechat-oauth-platform-config/spec.md -->

# wechat-oauth-platform-config Specification

## Purpose
TBD - created by archiving change wechat-app-oauth-login. Update Purpose after archive.
## Requirements
### Requirement: 按 platform 加载微信开放平台凭据

device-service SHALL 从配置 `wechat.platforms` 读取各 `platform` 键对应的 `appId` 与 `appSecret`。系统 SHALL 至少支持以下键名：`ios`、`android`、`web`。当请求中的 `platform` 在配置中不存在或 `appId`/`appSecret` 任一为空时，SHALL 返回明确配置错误且 SHALL NOT 调用微信 API。

`ios` 与 `android` SHALL 映射到**同一微信开放平台移动应用**的 `appId`/`appSecret`（部署时两键配置相同值）。`web` SHALL 映射到**微信开放平台网站应用**的独立 `appId`/`appSecret`。

生产环境 SHALL 通过环境变量或挂载配置覆盖 `appSecret`，且 SHALL NOT 将真实密钥提交到版本库。

#### Scenario: 移动应用 platform 解析凭据

- **WHEN** 登录请求 `platform` 为 `ios` 或 `android` 且对应配置项已填写有效 `appId`/`appSecret`
- **THEN** 系统 SHALL 使用该移动应用凭据调用微信 OAuth 换票 API

#### Scenario: 网站应用 platform 解析凭据

- **WHEN** 登录请求 `platform` 为 `web` 且 `wechat.platforms.web` 已填写有效 `appId`/`appSecret`
- **THEN** 系统 SHALL 使用该网站应用凭据调用微信 OAuth 换票 API

#### Scenario: 未配置的 platform

- **WHEN** 登录请求 `platform` 在 `wechat.platforms` 中不存在或凭据不完整
- **THEN** 系统 SHALL 返回明确错误且 SHALL NOT 创建或匹配 wx 行

---

## worker-dedicated-config-loading

<!-- source: openspec/specs/worker-dedicated-config-loading/spec.md -->

# worker-dedicated-config-loading Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Worker-service SHALL use dedicated configuration
`worker-service` MUST have a dedicated default config file and MUST load it when `GF_GCFG_FILE` is not explicitly provided.

#### Scenario: Worker starts without GF_GCFG_FILE
- **WHEN** `worker-service` starts and `GF_GCFG_FILE` is empty
- **THEN** runtime MUST default to `manifest/config/config.worker-service.yaml`

#### Scenario: Deployment manifest uses worker dedicated config
- **WHEN** compose/kustomize/dockerfile defines worker runtime env
- **THEN** worker `GF_GCFG_FILE` MUST point to `manifest/config/config.worker-service.yaml`

---

## worker-exclusive-background-runtime

<!-- source: openspec/specs/worker-exclusive-background-runtime/spec.md -->

# worker-exclusive-background-runtime Specification

## Purpose
TBD - created by archiving change worker-exclusive-background-tasks. Update Purpose after archive.
## Requirements
### Requirement: worker-service 后台任务角色 MUST 已废止

`worker-service` 进程 MUST NOT 部署；原由其独占的 domain outbox relay、voice task consumer 等后台任务 MUST NOT 存在。各域经 OpenSpec 批准的后台任务 MUST 在对应业务进程内启动（见 `background-loop-task-governance`）。

#### Scenario: 部署清单无 worker-service

- **WHEN** 审查 docker-compose 或 K8s manifest
- **THEN** MUST NOT 包含 `worker-service` 容器或等价部署单元

---

## wx-account-vip

<!-- source: openspec/specs/wx-account-vip/spec.md -->

# wx-account-vip Specification

## Purpose
TBD - created by archiving change care-alert-vip-by-wx. Update Purpose after archive.
## Requirements
### Requirement: wx 表承载账号 VIP 标志

device-service 所属库 `ai_voice_device` 的 `wx` 表 MUST 提供账号 VIP 列 `is_vip`（`TINYINT NOT NULL DEFAULT 0`：`0` 表示非 VIP，`1` 表示 VIP）。该列以 `wx.id` 为账号主键语义，MUST NOT 以 `device_no` 作为 VIP 归属键。

#### Scenario: 默认非 VIP

- **WHEN** 新建或历史迁移后的 `wx` 行未显式设置 VIP
- **THEN** `is_vip` MUST 为 `0`（非 VIP）

#### Scenario: VIP 归属账号而非设备

- **WHEN** 同一 `device_no` 关联的账号上下文需要判断 VIP
- **THEN** 系统 MUST 按该请求对应的 `wx.id` 行上的 `is_vip` 判定，MUST NOT 用设备号反推「设备级 VIP」

### Requirement: device 内部按 wxId 查询 VIP

device-service MUST 提供经内部密钥鉴权的 HTTP 接口，按正整数 `wxId` 返回该账号是否 VIP。voice-service 及其他跨域调用方 MUST 经此契约读取，MUST NOT 在非 device 进程内直查 `wx` 表。

#### Scenario: 存在账号且为 VIP

- **WHEN** 内部调用方以合法密钥请求 `wxId` 对应行且 `is_vip=1`
- **THEN** 响应 MUST 表示 `isVip=true`，并包含该 `wxId`

#### Scenario: 存在账号且非 VIP 或行不存在

- **WHEN** 内部调用方请求的 `wxId` 对应行 `is_vip=0`，或该主键不存在
- **THEN** 响应 MUST 表示 `isVip=false`（或等价：调用方视为非 VIP），MUST NOT 要求调用方因此中断无关主业务（由调用方定义降级）

#### Scenario: 缺少内部密钥

- **WHEN** 请求未携带正确的 device 内部密钥
- **THEN** device-service MUST 拒绝该请求

---

## wx-created-at

<!-- source: openspec/specs/wx-created-at/spec.md -->

# wx-created-at Specification

## Purpose
TBD - created by archiving change enhance-usage-stats-user-display. Update Purpose after archive.
## Requirements
### Requirement: wx table SHALL store account creation timestamp

`device-service` 权威库表 `wx` MUST 新增列 `created_at`（`BIGINT NOT NULL DEFAULT 0`，Unix 秒）。该字段 MUST 表示 wx 账号行**首次创建**时刻。历史行在迁移后 `created_at` MAY 为 0；展示层对 0 MUST 显示「—」。

#### Scenario: New wx row gets created_at on insert

- **WHEN** 用户经微信 OAuth、Apple 登录或用户名注册**首次**创建 wx 行
- **THEN** 插入时 `created_at` MUST 为当前 Unix 秒且大于 0

#### Scenario: Legacy row defaults to zero

- **WHEN** 迁移脚本执行后存在迁移前已创建的 wx 行且未回填
- **THEN** 该行 `created_at` MUST 为 0

### Requirement: wx creation write paths SHALL set created_at

下列 device-service 写路径在 **Insert** 新 wx 行时 MUST 写入 `created_at`：

- `WxLogin`（unionid 首次出现）
- Apple 首次登录 Insert
- `WxUsernameRegister`

模拟用户经 `SimUsernameRegister` 创建时 MUST 同样写入 `created_at`（与真实用户相同写路径）。

#### Scenario: WeChat first login

- **WHEN** `WxLogin` 为新的 unionid 插入 wx 行
- **THEN** 新行 `created_at` MUST 大于 0

#### Scenario: Username register

- **WHEN** `WxUsernameRegister` 成功插入新账号
- **THEN** 新行 `created_at` MUST 大于 0

---

## wx-username-auth

<!-- source: openspec/specs/wx-username-auth/spec.md -->

# wx-username-auth Specification

## Purpose
TBD - created by archiving change wx-username-auth-and-history-nickname. Update Purpose after archive.
## Requirements
### Requirement: 用户名注册写入 wx 账号
系统 SHALL 提供用户名注册接口，并在 `ai_voice_device.wx` 新建账号行；该行以 `wx.id` 作为账号主键，`unionid` MAY 为空，`user_name` MUST 全局唯一，`password` MUST 以不可逆哈希密文保存。

#### Scenario: 注册成功
- **WHEN** 客户端提交合法且未占用的 `userName` 与 `password`
- **THEN** 系统 SHALL 新建一条 `wx` 记录并返回 `wxId`，且数据库中的 `password` SHALL NOT 为明文

#### Scenario: 用户名冲突
- **WHEN** 客户端提交的 `userName` 已被其他 `wx` 记录占用
- **THEN** 系统 SHALL 返回“用户名已存在”冲突错误，且 SHALL NOT 新建记录

### Requirement: 用户名密码登录
系统 SHALL 提供用户名登录接口，按 `user_name` 定位 `wx` 记录并校验哈希密码；校验通过后 SHALL 返回 `wxId` 与账号业务信息供网关签发令牌。

#### Scenario: 登录成功
- **WHEN** `userName` 存在且密码校验通过
- **THEN** 系统 SHALL 返回对应 `wxId`，并标识登录成功

#### Scenario: 登录失败
- **WHEN** `userName` 不存在或密码校验失败
- **THEN** 系统 SHALL 返回认证失败错误，且 SHALL NOT 泄露是“用户名不存在”还是“密码错误”的内部细节

### Requirement: 用户名账号绑定微信
系统 SHALL 提供用户名账号绑定微信接口，将微信 `unionid` 绑定到指定 `wx.id` 账号；同一 `unionid` MUST NOT 同时绑定多个账号。

#### Scenario: 绑定成功
- **WHEN** 当前账号未绑定微信，且目标 `unionid` 未被其他账号占用
- **THEN** 系统 SHALL 将该 `unionid` 写入当前账号并返回成功

#### Scenario: 微信已被占用
- **WHEN** 目标 `unionid` 已绑定在其他 `wx.id`
- **THEN** 系统 SHALL 返回“微信已绑定其他账号”错误，且 SHALL NOT 覆盖原绑定

### Requirement: 用户名账号绑定设备号
系统 SHALL 提供用户名账号绑定设备号接口，绑定前 MUST 校验设备号已在设备域注册，绑定后 SHALL 维护 `wx.device_no` 一致性并失效相关缓存。

#### Scenario: 绑定成功
- **WHEN** `deviceNo` 已注册且请求主体账号合法
- **THEN** 系统 SHALL 更新 `wx.device_no` 并返回成功

#### Scenario: 设备号未注册
- **WHEN** 提交的 `deviceNo` 未在设备域注册
- **THEN** 系统 SHALL 返回业务校验失败，且 SHALL NOT 更新绑定关系

### Requirement: 修改用户名密码
系统 SHALL 提供修改密码接口；调用方 MUST 提供旧密码并通过校验后方可写入新密码哈希。

#### Scenario: 改密成功
- **WHEN** 旧密码校验通过且新密码满足格式策略
- **THEN** 系统 SHALL 将 `password` 更新为新哈希并返回成功

#### Scenario: 旧密码错误
- **WHEN** 旧密码校验失败
- **THEN** 系统 SHALL 返回认证失败错误，且 SHALL NOT 修改数据库密码

### Requirement: 微信账号下创建用户名密码
系统 SHALL 提供“微信账号创建用户名密码”接口，使已存在微信账号（`unionid` 已绑定）补齐 `user_name` 与 `password`；若账号已存在用户名，系统 MUST 拒绝重复创建。

#### Scenario: 创建成功
- **WHEN** 微信账号存在且尚未设置 `user_name`
- **THEN** 系统 SHALL 写入唯一用户名与密码哈希，并返回成功

#### Scenario: 已存在用户名
- **WHEN** 当前微信账号已设置 `user_name`
- **THEN** 系统 SHALL 返回“账号已存在用户名密码”错误，且 SHALL NOT 覆盖原值

---

