# OpenSpec 基线规格 v2.0.5

> 本文件由 `openspec/specs` 下全部 capability 规格于 **v2.0.5** 合并而成，作为该版本的确定性规则基线，便于按版本查阅。

> 后续新变更请基于本文件对照增量，或通过 OpenSpec 新建 change 再合并至下一版本规格。

## 目录

- [admin-event-inline-color-confirm](#admin-event-inline-color-confirm)
- [admin-qa-library](#admin-qa-library)
- [ai-monthly-quota](#ai-monthly-quota)
- [app-legal-docs](#app-legal-docs)
- [apple-sign-in-api](#apple-sign-in-api)
- [async-cache-projection-sync](#async-cache-projection-sync)
- [background-loop-task-governance](#background-loop-task-governance)
- [cache-and-messaging-hard-dependencies](#cache-and-messaging-hard-dependencies)
- [chinese-documentation-governance](#chinese-documentation-governance)
- [ci-acr-github-secrets](#ci-acr-github-secrets)
- [compose-container-resource-limits](#compose-container-resource-limits)
- [compose-host-root-asset-volumes](#compose-host-root-asset-volumes)
- [compose-mysql-endpoint-via-env](#compose-mysql-endpoint-via-env)
- [compose-mysql-test-seed-desensitization](#compose-mysql-test-seed-desensitization)
- [compose-prod-test-dual-stack](#compose-prod-test-dual-stack)
- [compose-redis-topology-2g](#compose-redis-topology-2g)
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
- [device-route-canary-management](#device-route-canary-management)
- [device-wx-profile-apis](#device-wx-profile-apis)
- [documentation-language-compliance](#documentation-language-compliance)
- [domain-package-boundary-enforcement](#domain-package-boundary-enforcement)
- [enum-adapter-compatibility](#enum-adapter-compatibility)
- [event-logo-oss-cdn](#event-logo-oss-cdn)
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
- [history-delegate-downstream-urls](#history-delegate-downstream-urls)
- [history-device-sync-cache-projection](#history-device-sync-cache-projection)
- [history-event-unit-denorm](#history-event-unit-denorm)
- [history-piece-and-realtime-notify](#history-piece-and-realtime-notify)
- [history-profile-nickname](#history-profile-nickname)
- [history-service-db-ownership](#history-service-db-ownership)
- [main-config-boundary-pruning](#main-config-boundary-pruning)
- [main-config-without-database](#main-config-without-database)
- [microservice-boundary-final-alignment](#microservice-boundary-final-alignment)
- [pangbao-ai-clinic](#pangbao-ai-clinic)
- [pangbao-ai-clinic-flutter](#pangbao-ai-clinic-flutter)
- [redis-disaster-recovery-runbook](#redis-disaster-recovery-runbook)
- [redis-read-model-cache](#redis-read-model-cache)
- [routing-key-governance](#routing-key-governance)
- [routing-key-governance-workflow](#routing-key-governance-workflow)
- [routing-key-prefix-registry](#routing-key-prefix-registry)
- [runtime-dependency-check](#runtime-dependency-check)
- [runtime-docs-centralization-and-governance](#runtime-docs-centralization-and-governance)
- [service-boundary-no-cross-db](#service-boundary-no-cross-db)
- [service-code-full-cutover](#service-code-full-cutover)
- [service-dedicated-config-loading](#service-dedicated-config-loading)
- [service-migration-safety-and-rollback](#service-migration-safety-and-rollback)
- [service-runtime-standardization](#service-runtime-standardization)
- [single-default-db-per-service](#single-default-db-per-service)
- [typed-domain-enums](#typed-domain-enums)
- [ucg-admin-post-moderation](#ucg-admin-post-moderation)
- [ucg-admin-profile-moderation](#ucg-admin-profile-moderation)
- [ucg-ai-quota](#ucg-ai-quota)
- [ucg-aliyun-secrets-env](#ucg-aliyun-secrets-env)
- [ucg-app-http-api](#ucg-app-http-api)
- [ucg-audit-mq](#ucg-audit-mq)
- [ucg-audit-mq-reliability](#ucg-audit-mq-reliability)
- [ucg-audit-publish-outbox](#ucg-audit-publish-outbox)
- [ucg-chat-mysql-persist](#ucg-chat-mysql-persist)
- [ucg-chat-ws](#ucg-chat-ws)
- [ucg-data-model](#ucg-data-model)
- [ucg-device-internal-api](#ucg-device-internal-api)
- [ucg-gateway-proxy](#ucg-gateway-proxy)
- [ucg-green-audit](#ucg-green-audit)
- [ucg-oss-presign](#ucg-oss-presign)
- [ucg-recommend-feed](#ucg-recommend-feed)
- [ucg-recommend-mq](#ucg-recommend-mq)
- [ucg-service-runtime](#ucg-service-runtime)
- [validated-prefix-dispatch](#validated-prefix-dispatch)
- [voice-admin-ui](#voice-admin-ui)
- [voice-ai-quota](#voice-ai-quota)
- [voice-and-device-service-decomposition](#voice-and-device-service-decomposition)
- [voice-device-domain-http-access](#voice-device-domain-http-access)
- [voice-device-profile-http-contract](#voice-device-profile-http-contract)
- [voice-event-child-disambiguation](#voice-event-child-disambiguation)
- [voice-history-http-contract](#voice-history-http-contract)
- [voice-realtime-asr-ws](#voice-realtime-asr-ws)
- [voice-route-canary-management](#voice-route-canary-management)
- [voice-textchat-resilience](#voice-textchat-resilience)
- [wechat-ios-universal-links](#wechat-ios-universal-links)
- [wechat-oauth-platform-config](#wechat-oauth-platform-config)
- [worker-dedicated-config-loading](#worker-dedicated-config-loading)
- [worker-exclusive-background-runtime](#worker-exclusive-background-runtime)
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

## ai-monthly-quota

<!-- source: openspec/specs/ai-monthly-quota/spec.md -->

# ai-monthly-quota Specification

## Purpose
TBD - created by archiving change ai-monthly-quota. Update Purpose after archive.
## Requirements
### Requirement: UCG polish SHALL pre-check quota and consume only on DashScope success

`POST /ucg/app/api/posts/polish` MUST 在调用 DashScope 前于 **ucg-service 进程内**执行 polish check（MUST NOT 经 device internal API）；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用 DashScope。DashScope 成功返回有效正文后 MUST 于 ucg-service 进程内 consume。参数错误、未配置 AI、DashScope 失败 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求 DashScope

#### Scenario: DashScope 失败不扣减

- **WHEN** check 通过但 DashScope 返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（母婴 DeepSeek 或 casual 流式）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device **user 域** internal API 由 deviceNo 反查，MUST NOT 经 ai-quota API）。wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`voice_ai`** 执行 check；`allowed=false` MUST 返回 WS 错误帧 code **40302** message **「本月额度已用完」**。LLM 成功完成后 MUST consume。模式切换、规则回复、LLM 失败兜底、纯 ASR MUST NOT check 或 consume。

#### Scenario: 未登录不可用喂养 AI

- **WHEN** WS 会话 wxId 解析为 0 且用户 utterance 将触发 LLM
- **THEN** 系统 SHALL 返回 40301 且 SHALL NOT 调用 LLM

#### Scenario: 喂养 AI 额度用尽

- **WHEN** check 得到 voice_ai used=limit
- **THEN** WS SHALL 返回 40302「本月额度已用完」且 SHALL NOT 调用 LLM

#### Scenario: 模式切换不扣减

- **WHEN** 用户发送模式切换指令且不触发 LLM
- **THEN** 系统 SHALL NOT 调用 check 或 consume

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**（含润笔、喂养 AI、胖宝 AI）。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。额度展示 MUST 从分域 API 获取：voice 域（voiceAi、clinicAi）与 ucg 域（polish），MUST NOT 依赖 `/device/app/api/ai-quota`。

#### Scenario: 润笔超额弹框

- **WHEN** polish API 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 喂养 AI 超额弹框

- **WHEN** voice WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 胖宝 AI 超额弹框

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

### Requirement: Voice clinic AI SHALL enforce clinic_ai quota locally

voice-service 在处理 `/voice/clinic/ws` 的 `question` 并即将调用 LLM 前 MUST 解析 wxId>0。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`clinic_ai`** 执行 check；`allowed=false` MUST 返回 WS 40302。LLM 流式成功完成后 MUST consume。摘要刷新失败、rate limit、参数校验失败 MUST NOT consume。

#### Scenario: 胖宝额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM

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

`resource/public/privacy-policy.html` MUST 修正既有 DashScope/DeepSeek 供应商描述不准确之处。文档 MUST 在 AI 相关章节说明 **胖宝诊疗** 功能（用户可见名称；内部实现仍为 clinic）：为生成回答，系统会读取用户近 **7 天喂养记录聚合摘要**（非完整原始记录全文），并 MAY 在 App 内展示 AI **思考过程**（thinking）供用户参考。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读胖宝诊疗相关说明

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 7 天喂养摘要与 thinking 展示说明，且用户向名称 SHALL 为「胖宝诊疗」
- **AND** DeepSeek/DashScope 等模型供应商描述 SHALL 与后端实际调用一致

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
- **THEN** `g.Redis()` SHALL 连接测试单机 Redis，**SHALL NOT** 连接生产 cluster 节点

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
系统在执行 DeepSeek 意图分析与对话补全前，历史上下文读取 SHALL 优先命中 Redis 历史读模型，并在不可用时回源到历史服务或数据库。

#### Scenario: 命中缓存快速构造上下文
- **WHEN** DeepSeek 请求前需要最近历史上下文且 Redis 命中
- **THEN** 系统 MUST 使用缓存数据完成上下文构造并减少对数据库/远程历史服务访问

#### Scenario: 未命中缓存自动回源
- **WHEN** DeepSeek 请求前历史缓存未命中
- **THEN** 系统 MUST 回源获取历史并回填缓存后继续调用 DeepSeek

### Requirement: 上下文读取一致性与降级可观测
系统 MUST 对 DeepSeek 上下文读取提供命中率、回源率与降级原因可观测性，并在异常时保持功能可用。

#### Scenario: 记录上下文读取指标
- **WHEN** 任一 DeepSeek 请求完成上下文装配
- **THEN** 系统 MUST 记录本次是否命中 Redis、是否发生回源与耗时指标

#### Scenario: Redis 异常时不阻断回复
- **WHEN** Redis 不可达或读取失败
- **THEN** 系统 MUST 降级回源并继续完成 DeepSeek 调用，同时输出结构化告警日志

### Requirement: 历史窗口语义一致
系统 SHALL 保证 Redis 历史读模型与权威源在“最近 N 小时”窗口语义上保持一致，避免因缓存截断导致模型上下文偏差。

#### Scenario: 缓存与权威窗口对齐
- **WHEN** 系统读取最近 N 小时历史用于 DeepSeek
- **THEN** 系统 MUST 返回与权威源相同窗口边界的数据集合（允许在异步延迟窗口内最终一致）

---

## device-admin

<!-- source: openspec/specs/device-admin/spec.md -->

# device-admin Specification

## Purpose
TBD - created by archiving change fix-api-usage-stats. Update Purpose after archive.
## Requirements
### Requirement: device admin SHALL provide paginated wx account list

`device-service` MUST 暴露 `GET /device/admin/api/wx/list`，鉴权与现有 device admin 一致（Header `X-Admin-Password`）。查询参数：`page`（默认 1）、`pageSize`（默认 20，最大 100）、可选 `q`（模糊匹配 wx.id、deviceNo、unionid、account）。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`（用户名账号若有）。

#### Scenario: 默认分页列表

- **WHEN** 管理员携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/wx/list` 且未传分页参数
- **THEN** 响应 SHALL 返回第一页 wx 记录且 `page=1`

#### Scenario: 关键字搜索

- **WHEN** 管理员请求 `q=138` 且存在 deviceNo 或 id 匹配的 wx 行
- **THEN** 响应 `list` SHALL 仅包含匹配项

#### Scenario: 经 gateway-app 反代可达

- **WHEN** 管理员经 gateway-app 携带 Admin JWT 请求 `/device/admin/api/wx/list`
- **THEN** gateway-app SHALL 反代至 device-service 并成功返回列表

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

`gateway-app-server` MUST 在 HTTP 响应确定后（状态码已写入客户端方向）评估是否写入使用统计。实现 MUST 覆盖经领域反代（device/ucg/history/voice）短路 `ExitAll` 的路径，不得仅依赖 `BindMiddleware("/*")` 在 `Next()` 之后记录。仅当响应状态码满足 `200 <= status < 300` 时 SHALL 计数一次。统计路径 MUST 为归一化后的 `METHOD /path`（不含 query）。下列请求 MUST NOT 写入统计：WebSocket 升级、`/device/internal/` 前缀、`/device/admin/api/` 前缀（含本变更读 API 自身）、静态资源与 HTML 壳页，以及维护型 App API（`POST /device/app/api/token/refresh`、`GET /device/app/api/version/check`、`GET /device/app/api/site/home`、`/device/app/api/version/admin/*` 前缀、**`GET /ucg/app/api/posts/{id}/comments`**）。登录、注册、绑定、POST 评论与各业务 App API SHALL 继续计入。写入 MUST 异步执行且 SHALL NOT 阻塞或改变业务响应。

#### Scenario: token 刷新不计入

- **WHEN** 经 gateway-app 的 `POST /device/app/api/token/refresh` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: GET 评论列表不计入

- **WHEN** 经 gateway-app 的 `GET /ucg/app/api/posts/123/comments` 返回 HTTP 200
- **THEN** 系统 SHALL NOT 增加该 apiKey 的统计计数

#### Scenario: POST 评论仍计入

- **WHEN** 经 gateway-app 的 `POST /ucg/app/api/posts/123/comments` 返回 HTTP 200
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 登录 API 仍计入

- **WHEN** 经 gateway-app 的 `POST /device/app/api/apple_login` 返回 HTTP 200
- **THEN** 对应 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 2xx 成功请求计入统计

- **WHEN** 经 gateway-app 的 `GET /ucg/app/api/feed/recommend` 返回 HTTP 200
- **THEN** 对应归一化 apiKey 的全局日计数 SHALL 增加 1

#### Scenario: 反代 ExitAll 后仍计入

- **WHEN** 经 gateway-app 反代至 device-service 的 `GET /device/app/api/user/get` 返回 HTTP 200
- **THEN** 全局日计数 SHALL 增加 1 且 SHALL NOT 因反代中间件 `ExitAll` 而跳过写入

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

`GET /device/admin/api/usage/list` MUST 要求有效 **Admin JWT**（`Authorization: Bearer`，`aud=gateway-admin`）。查询参数 `days` 默认 `7`；`days=0` 表示聚合 TTL 内全部日桶。响应 MUST 包含 `list` 数组，每项至少含 `apiKey`、`summary`、`count`（窗口内合计）、`lastAt`（Unix 秒）。查询参数 `sortBy` 默认 `count`；`sortBy=lastAt` 时列表 SHALL 按 `lastAt` 降序，否则 SHALL 按 `count` 降序。当 Redis 日桶 Hash 中存在 field 时，读路径 MUST 正确解析 GoFrame Redis 返回值（含 `HGETALL` 经 adapter 转为 flat `[]string` 的情形）并 SHALL NOT 因类型解析失败返回空列表。

#### Scenario: Redis 有数据时列表非空

- **WHEN** Redis 键 `gw:usage:d:{today}:g` 的 Hash 含至少一个 apiKey field，且管理员携带有效 Admin JWT 请求 `days=7`
- **THEN** 响应 `list` SHALL 包含对应 apiKey 且 `count > 0`

#### Scenario: GoFrame HGETALL flat []string 可读

- **WHEN** `g.Redis().Do("HGETALL", key)` 返回的 `*gvar.Var` 经 adapter 内部表示为 flat `[]string`（非 map）
- **THEN** `ListAPIs` 等读路径 SHALL 仍正确聚合 field 与计数，SHALL NOT 返回空 `list`

#### Scenario: 默认近 7 天

- **WHEN** 管理员携带有效 Admin JWT 请求列表且未传 `days`
- **THEN** 每项 `count` SHALL 为最近 7 个自然日 INCR 之和

#### Scenario: 无效或缺失 token

- **WHEN** 请求未携带有效 Admin JWT
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回统计数据

### Requirement: Admin usage detail API SHALL list wxId callers for one API

`GET /device/admin/api/usage/detail` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `apiKey`；可选 `days`（默认 7）。响应 `list` 每项 SHALL 至少包含 `wxId`、`count`、`lastAt`。仅 `wxId > 0` 的调用 SHALL 出现在列表中。

#### Scenario: 按 API 下钻 wxId

- **WHEN** 管理员携带 Admin JWT 查询 `apiKey=GET /ucg/app/api/feed/recommend` 且 `days=7`
- **THEN** 响应 SHALL 列出窗口内调用过该 API 的 wxId、次数与最近调用时间

### Requirement: Admin usage user API SHALL list APIs called by one wxId

`GET /device/admin/api/usage/user` MUST 要求有效 Admin JWT。查询参数 MUST 包含 `wxId`（正整数）；可选 `days`（默认 7）。响应 `list` 每项 SHALL 至少包含 `apiKey`、`summary`、`count`、`lastAt`。

#### Scenario: 按 wxId 查看 API 分布

- **WHEN** 管理员携带 Admin JWT 查询 `wxId=1001` 且 `days=7`
- **THEN** 响应 SHALL 列出该用户在窗口内成功调用过的 API、中文说明、次数与最近调用时间

### Requirement: Usage read APIs SHALL be served by gateway-app and excluded from device proxy

路径 `/device/admin/api/usage/*` MUST 由 gateway-app 本机处理，MUST NOT 被 `device-service` 反代吞掉。读 API 仅读取 gateway Redis 统计，MUST NOT 直连 device/history/voice/ucg 数据库。

#### Scenario: 读 API 不经 device-service

- **WHEN** 管理员请求 `GET /device/admin/api/usage/list`
- **THEN** gateway-app SHALL 本地响应且 SHALL NOT 将请求转发至 device-service

### Requirement: api-usage-stats.html SHALL provide standalone admin UI

静态页 `resource/public/api-usage-stats.html` MUST 通过 `/device/admin/api-usage-stats` 提供（**仅 App 网关**）。页面 MUST 使用 Hub Admin JWT（`admin-common.js`）调用读 API，MUST NOT 使用 `X-Admin-Password`。页面 SHALL 提供「按 API」「按用户」两个视图；默认时间窗口为近 7 天。页脚或说明区 SHALL 注明：用户维度仅统计 `wxId>0` 的已登录账号；HTTP 统计不含 WebSocket 与维护型接口（token 刷新、版本检查等）。页面 SHALL 提供排序选择，默认按调用次数降序，可选最近调用降序。

#### Scenario: 独立页在有效 token 下查看 API 列表

- **WHEN** 管理员已持有 Admin JWT 并打开 `api-usage-stats.html`
- **THEN** 页面 SHALL 展示 API 频率表且每项含中文 summary

#### Scenario: 按用户从 wx 列表点选

- **WHEN** 管理员切换到「按用户」视图
- **THEN** 页面 SHALL 展示 wx 账号列表；点选某 wxId 后 SHALL 展示该用户的 API 调用列表

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
官网页面 SHALL 以玻璃拟态风格展示品牌定位文案，并 SHALL 展示从数据库权威链路读取的事件列表。每个事件项 MUST 至少包含事件名与事件 logo；若 logo 为 path-only 资源，前端或聚合接口 MUST 能将其解析为当前站点可访问的同源地址。

#### Scenario: 官网首屏展示品牌定位
- **WHEN** 用户打开官网首页
- **THEN** 页面 SHALL 明确表达“专注母婴喂养方面的服务商”以及“更便捷、更轻松地照顾孩子”等核心信息

#### Scenario: 官网展示事件 logo 与事件名
- **WHEN** 官网聚合到至少一条事件数据
- **THEN** 页面 SHALL 为每条事件渲染事件 logo 与事件名，且 logo 地址 MUST 可被当前官网域名直接访问

### Requirement: 官网提供匿名只读聚合数据接口
系统 SHALL 提供一个适用于官网匿名访问的只读聚合接口，由 `gateway-app-server` 统一返回官网所需的事件展示数据、Android 下载信息与 iOS 下载说明。该接口 MUST 通过服务契约或本进程已有能力获取数据，MUST NOT 让前端直接调用受保护业务接口或跨服务直连数据库。

#### Scenario: 匿名读取官网数据
- **WHEN** 未登录用户请求官网聚合接口
- **THEN** 系统 SHALL 返回成功响应，其中包含事件列表、Android 下载展示信息与 iOS 下载说明

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

## pangbao-ai-clinic

<!-- source: openspec/specs/pangbao-ai-clinic/spec.md -->

# pangbao-ai-clinic Specification

## Purpose
TBD - created by archiving change add-pangbao-ai-clinic-room. Update Purpose after archive.
## Requirements
### Requirement: voice-service SHALL 提供胖宝 AI 诊室 WebSocket handler

`voice-service` MUST 在路径 `/voice/clinic/ws` 注册 WebSocket handler（`BindHandler`），处理经 gateway-app 透传而来的客户端文本提问并流式返回 DeepSeek 回答。用户可见功能名称为 **胖宝诊疗**；实现与配置仍使用 `clinic` / `clinic_ai` / `voice:clinic:*` 命名。该路径为**集群内业务端点**；App 对外入口 MUST 为 gateway-app-server 同路径透传（见 `gateway-ws-edge-proxy`）。实现 MUST NOT 将连接注册到 `VoiceWSManager`。实现 MUST NOT 提供 TTS 或音频上行能力（MVP 纯文本）。

#### Scenario: 经 gateway-app 透传后握手成功

- **WHEN** 客户端经 gateway-app 透传对 voice-service `/voice/clinic/ws` 完成 WebSocket Upgrade
- **THEN** voice-service SHALL 接受连接并等待首帧 `auth` JSON

### Requirement: Clinic WebSocket SHALL 以 wxId 为主键绑定身份

`/voice/clinic/ws` 的连接、会话、限流与额度维度 MUST 以 **`wx.id`（正整数）** 为主键，与 `/voice/chat/ws` 以 `deviceNo` 注册 `VoiceWSManager` 的行为 MUST 显式不同。实现 MUST NOT 使用 `deviceNo` 反查 wxId 作为 clinic 鉴权路径（即 MUST NOT 调用 `ResolveVoiceWxID` 的 deviceNo fallback）。`deviceNo` MAY 仅用于 history 摘要拉取，且 MUST 与 JWT `device_no` claim 一致。

#### Scenario: 首帧 auth 绑定 wxId

- **WHEN** 客户端握手后发送 `type=auth` 且 JWT 含 `sub=1001` 与有效 `device_no`
- **THEN** 服务端 SHALL 解析 `wxId=1001` 并返回 `auth_ok`
- **AND** 后续 session/限流/额度 MUST 使用 wxId=1001

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

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，服务端 MUST 下发 `type=session_sync`（见 `amend-pangbao-clinic-ux`）。`session_sync` 之后，客户端上行 MUST 支持 `type=question` 帧（含非空 `text` 与非空 **`turnId`** UUID）与 **`type=cancel`** 帧（含非空 **`turnId`**）。服务端下行 MUST 支持 `auth_ok`、`session_sync`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 七种 `type`。流式下行帧（`thinking_delta`、`answer_delta`、`answer_done`）**MUST** 含与当前 turn 一致的 **`turnId`**。`turn_cancelled` MUST 含 **`turnId`** 与 **`reason`**，取值为 **`superseded`**、**`cancelled`** 或 **`disconnected`** 之一。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

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
- **AND** MUST NOT consume `clinic_ai` 且 MUST NOT append session

### Requirement: Clinic SHALL 注入近 7 天喂养事件聚合摘要

每次处理 `question` 前，系统 MUST 取得该设备近 7 天喂养 history，并按 event 聚合为摘要（含 count、amount 合计、duration 合计等），注入 LLM system/context。摘要 MUST NOT 为原始 history 行全量 dump。history 数据 MUST 经 HTTP 契约（如 `DeviceHistory`）获取；voice-service MUST NOT 直连 history 库表。

#### Scenario: 摘要在 prompt 中可见

- **WHEN** 设备近 7 天有 3 次「母乳」记录
- **THEN** 注入 LLM 的上下文中 SHALL 包含按 event 聚合后的统计，且 token 量 SHALL 小于同等条数原始 JSON 行列表

#### Scenario: history 契约失败

- **WHEN** history HTTP 契约不可用或返回错误
- **THEN** 系统 SHALL 返回可诊断 `error` 帧且 MUST NOT 在无摘要时静默调用 LLM（除非 design 明确降级为空摘要并记录日志——本 spec 要求显式失败或空摘要+警告日志二选一，实现 MUST 在 design 中择一并在日志中可观测）

### Requirement: Clinic 摘要 SHALL 懒刷新

系统 MUST 在每次 `question` 前对比 history watermark 与已缓存摘要的 watermark；若缓存缺失或 watermark 过期，MUST 重新计算摘要并更新 Redis 缓存 `voice:clinic:summary:{wxId}:{deviceNo}`（或等价键）。

#### Scenario: history 未变化复用缓存

- **WHEN** 同一 session 内第二次提问且 history watermark 未变
- **THEN** 系统 SHALL 复用 Redis 中摘要且 MUST NOT 重复全量聚合计算

#### Scenario: 新喂养记录触发重算

- **WHEN** 用户新增喂养记录导致 history watermark 前进
- **THEN** 下一次 `question` MUST 重算摘要并更新缓存

### Requirement: Clinic 会话 SHALL 使用固定 12 小时 TTL（wxId 键）

Redis 键 **`voice:clinic:session:{wxId}`** MUST 在**首条**成功完成的 `question`（即首次 `answer_done` 后 append）时创建，TTL 为 12 小时自 `firstQuestionAt` 起算。后续提问 MUST NOT 滑动续期 TTL。进入胖宝页与 `auth_ok` MUST NOT 预创建 session。Session MUST 记录 auth 时锁定的 `deviceNo` 供摘要使用。**未完成**（cancelled / superseded / disconnected）的 turn **MUST NOT** append 至 session。

#### Scenario: 首问创建 session

- **WHEN** wxId=1001 用户首条 turn 以 `answer_done` 成功结束
- **THEN** Redis MUST 写入键 `voice:clinic:session:1001` 且 EX=12h

#### Scenario: 会话内多轮上下文

- **WHEN** 12h 内同一 wxId session 已有完成轮次且新 turn 以 `answer_done` 成功结束
- **THEN** LLM 上下文 SHALL 包含同 session 内先前已完成 Q&A

#### Scenario: TTL 过期后会话重置

- **WHEN** 首问后超过 12h 同一 wxId 再提问
- **THEN** 系统 MUST 创建新 session 且 prior Q&A 上下文 SHALL 为空

#### Scenario: 取消的 turn 不写入 session

- **WHEN** turn 被 cancel 或 supersede 且未产生 `answer_done`
- **THEN** Redis session MUST NOT 追加该 partial Q&A

### Requirement: Clinic LLM SHALL 使用 deepseek-v4-pro 并启用 thinking

Clinic LLM 调用 MUST 使用模型 `deepseek-v4-pro`，MUST 启用 thinking（`extra_body.thinking` 或等价配置，`reasoning_effort: high`）。LLM 请求超时 MUST 为 **120 秒**，配置源 MUST 为 `config.voice-service.yaml` 的 `aiClinic` 块，MUST NOT 依赖 `voice-chat.shared.yaml` 的 voice ball 超时。

#### Scenario: thinking 流映射

- **WHEN** DeepSeek 返回 reasoning 流
- **THEN** 服务端 MUST 将其映射为 `thinking_delta` 帧

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`；`wxId≤0` MUST 返回 `error` code **40301**。LLM 调用前 MUST 经 device internal 对 feature `clinic_ai` 以该 wxId 执行 check；`allowed=false` MUST 返回 code **40302** message **「本月额度已用完」** 且 MUST NOT 调用 LLM。**仅** 当 turn 以 **`answer_done` 成功结束** 时 MUST 以同一 wxId consume。**cancelled**、**superseded**、**disconnected** 或 LLM/摘要失败而中断的 turn **MUST NOT** consume。

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM

#### Scenario: 用户 cancel 不扣额度

- **WHEN** turn 在流式过程中被 `cancel` 或 `superseded` 结束且未收到 `answer_done`
- **THEN** 系统 MUST NOT 对该 turn 调用 consume `clinic_ai`

#### Scenario: 成功完成才扣额度

- **WHEN** turn 完整流式结束并下发 `answer_done`
- **THEN** 系统 MUST consume `clinic_ai` 一次

### Requirement: Clinic SHALL 实施 Redis 限流（per wxId）

系统 MUST 对 clinic 提问路径实施 Redis 限流，键 **`voice:clinic:rate:{wxId}`**。限流计数 **MUST** 在 **`answer_done` 成功** 后递增；**cancelled**、**superseded**、**disconnected** 或失败 turn **MUST NOT** 递增限流计数。处理新 `question` 前 MUST 检查当前窗口计数；超限时 MUST 返回 WS `error` code **42901** 且 MUST NOT 调用 LLM。

#### Scenario: 短时间频繁提问

- **WHEN** 同一 wxId 在限流窗口内已成功完成（`answer_done`）次数超过阈值
- **THEN** 下一次 `question` SHALL 返回 42901 且 MUST NOT 调用 LLM

#### Scenario: supersede 未完成 turn 不计入限流

- **WHEN** 用户在窗口内多次改问但均未产生 `answer_done`
- **THEN** 限流计数 MUST NOT 因 supersede 而额外递增

### Requirement: Clinic SHALL 在 auth_ok 后下发 session_sync

`auth_ok` 成功后，voice-service MUST 读取 Redis `voice:clinic:session:{wxId}` 并立即向客户端下发 `session_sync` 帧。每次 WebSocket 重连并完成 `auth` 后 MUST 重复下发。payload MUST 含 `turns` 数组与 `expiresAt`（Unix 秒）。`turns` 每项 MUST 仅含 `question` 与 `answer` 字符串；MUST NOT 含 `thinking`。仅 MUST 包含 question 与 answer 均非空的已完成轮次（与 session 写入 `appendClinicTurn` 语义一致）。无 session 时 MUST 下发 `turns: []` 且 `expiresAt` 为 0 或省略。

#### Scenario: 有历史会话时恢复轮次

- **WHEN** wxId=1001 在 12h 内已有 2 轮已完成 Q&A 且客户端 `auth` 成功
- **THEN** 服务端 SHALL 在 `auth_ok` 后下发 `session_sync` 且 `turns` 长度为 2
- **AND** 每轮 MUST 含对应 `question` 与 `answer`
- **AND** MUST NOT 含 thinking 字段

#### Scenario: 无 session 时空同步

- **WHEN** wxId=1001 尚未发送过 `question` 且客户端 `auth` 成功
- **THEN** 服务端 SHALL 下发 `session_sync` 且 `turns` 为空数组

#### Scenario: 重连重复同步

- **WHEN** 同一 wxId 断开 WS 后重连并再次 `auth` 成功
- **THEN** 服务端 SHALL 再次下发 `session_sync` 且内容与当前 Redis session 一致

#### Scenario: expiresAt 反映会话过期

- **WHEN** session 存在且 `firstQuestionAt` 已知
- **THEN** `session_sync.expiresAt` SHALL 为会话绝对过期 Unix 秒（首问时刻 + 12h 固定 TTL）

### Requirement: Clinic WS handler SHALL 非阻塞处理 question 并支持显式取消

`/voice/clinic/ws` 读循环 MUST NOT 同步阻塞在单条 `HandleQuestion` 直至 LLM 结束。每条合法 `question` MUST 在独立 goroutine 中处理，且连接 MUST 维护 at most one **active turn** 及其 `context.CancelFunc`。收到匹配 active turn 的 `cancel`、收到新 `question`（supersede）或连接关闭时 MUST 调用 cancel 中断 LLM 流。

#### Scenario: 读循环可并发接收 cancel

- **WHEN** LLM 流式进行中且读循环收到 `cancel` 帧
- **THEN** handler MUST 在不等待 LLM 自然结束的情况下处理 cancel

#### Scenario: HandleQuestion 尊重 turn context

- **WHEN** turn context 已被 cancel
- **THEN** `HandleQuestion` MUST 停止 LLM 读流且 MUST NOT 写 `answer_done`

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

`GET /ucg/admin/api/posts/list` MUST 支持查询参数 `page`（从 1 开始）、`pageSize`（默认 20，最大 100）、可选 `status`（0/1/2/3）。省略 `status` 时 SHALL 返回全部状态的帖子。列表 MUST 按 `updated_at` 降序排序。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；每项 SHALL 至少包含 `id`、`authorWxId`、`content`、`status`、`rejectReason`、`createdAt`、`updatedAt`、`publishedAt` 及媒体展示字段。`media` 数组 MUST 包含该帖 **全部** 媒体项（按 `sortOrder`）；每项 SHALL 含 `cdnUrl`、`mediaKind`（1=图片，2=视频），图片 SHALL 含 OSS 缩略图 `thumbnailUrl`，视频 SHALL 含 OSS 首帧截帧 `thumbnailUrl`（`BuildVideoSnapshotURL`），SHALL NOT 仅返回无 thumbnail 的 mp4 `cdnUrl` 供列表 `<img>` 直接使用。

#### Scenario: 按状态筛选待审帖

- **WHEN** 管理员请求 `status=1`
- **THEN** 响应 `list` 中每条 `status` SHALL 为 1（pending_audit）

#### Scenario: 分页默认值

- **WHEN** 管理员未传 `page` 与 `pageSize`
- **THEN** 系统 SHALL 使用 `page=1`、`pageSize=20` 并返回对应分页元数据

#### Scenario: 视频帖返回首帧 thumbnail

- **WHEN** 列表项首条媒体 `mediaKind=2` 且 `objectKey` 有效
- **THEN** 该项 `thumbnailUrl` SHALL 非空且 SHALL 可用于 `<img>` 展示；`cdnUrl` SHALL 为可播放的视频 URL

#### Scenario: 多图帖返回全量 media

- **WHEN** 帖子关联多条 `ucg_post_media`
- **THEN** 列表项 `media` 数组长度 SHALL 等于关联条数且顺序与 `sortOrder` 一致

### Requirement: Admin SHALL batch reject posts with shared reject semantics

`POST /ucg/admin/api/posts/reject` MUST 接受 JSON body：`postIds`（非空数组，最多 100 个 id）、可选 `reason`（空时 SHALL 使用默认文案「违规已下架」）。对每条目标帖子，若当前 `status` 已为 3（rejected）SHALL 计入 `skipped` 且不更新行；若 `status` 为 0、1 或 2，系统 MUST 将 `status` 置为 3、写入 `reject_reason`、更新 `updated_at` 为当前 unix 秒，并计入 `rejected`。DB 错误计入 `failed`。响应 MUST 包含 `rejected`、`skipped`、`failed` 三个 id 数组。本操作 SHALL NOT 向作者发送通知或站内信。

#### Scenario: 批量驳回已发布帖

- **WHEN** 管理员提交 `postIds` 含 `status=2` 的帖子且口令正确
- **THEN** 对应行 `status` SHALL 变为 3 且 `reject_reason` SHALL 非空；该帖 SHALL NOT 出现在推荐或关注 Feed（`status=published` 查询）

#### Scenario: 已驳回帖幂等跳过

- **WHEN** 管理员对 `status=3` 的帖子再次提交驳回
- **THEN** 该行 SHALL NOT 被修改且该 id SHALL 出现在 `skipped`

#### Scenario: 作者可见驳回原因无推送

- **WHEN** 帖子被管理端驳回
- **THEN** 作者请求「我的动态」SHALL 可见该帖 `status=3` 与 `reject_reason`；系统 SHALL NOT 因本次驳回创建通知或 WS 推送

### Requirement: ucg-admin.html SHALL provide post moderation tab with batch reject UI

静态页 `resource/public/ucg-admin.html` MUST 在现有 UCG Admin 登录态下提供「动态审查」模块（可与 AI 配置以 Tab 切换）。模块 SHALL 调用列表 API 展示表格，对 `status≠3` 的行提供 checkbox；SHALL 提供「全选本页可驳回项」与「批量驳回」按钮；批量驳回前 MUST 经用户确认。`status=3` 的行 checkbox SHALL 禁用或不可选。操作成功后 SHALL 刷新当前列表。表格「媒体」列 SHALL 展示每条动态 **全量** 媒体缩略图（非仅第一条）；图片 SHALL 支持点击后在 modal 中查看原图（`cdnUrl`）；视频 SHALL 展示首帧缩略图并支持点击后在 modal 内播放（`<video controls>` + `cdnUrl`）。

同一页面 MUST 提供与「动态审查」并列的 **「资料机审失败」** Tab（或与 AI 配置/动态审查同级 Tab 导航），用于展示 `ProfileJobStatusModerationFailed` 的 profile audit job 列表及人工操作入口（详见 `ucg-admin-profile-moderation` 规格）。

#### Scenario: 动态审查批量驳回

- **WHEN** 管理员在「动态审查」勾选多条未驳回帖子并确认批量驳回
- **THEN** 系统 SHALL 调用批量驳回 API 且刷新列表

#### Scenario: 管理页含资料机审 Tab

- **WHEN** 管理员打开 ucg-admin.html 并已通过口令登录
- **THEN** 页面 SHALL 展示「资料机审失败」Tab 入口，且 SHALL NOT 与「动态审查」互斥隐藏

### Requirement: device admin entry SHALL link to UCG management

`resource/public/admin.html` 中指向 `/device/admin/ucg-admin.html` 的入口链接文案 SHALL 为「UCG 管理」（或等价中文），以涵盖 AI 配置与动态审查。

#### Scenario: 设备管理页入口文案

- **WHEN** 管理员打开设备管理页
- **THEN** UCG 入口链接可见文案 SHALL 为「UCG 管理」

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

`POST /ucg/app/api/posts/polish` MUST 在调用 DashScope 前于 **ucg-service 进程内**执行 polish check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用 DashScope。DashScope 成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、DashScope 失败 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求 DashScope

#### Scenario: DashScope 失败不扣减

- **WHEN** check 通过但 DashScope 返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

### Requirement: ucg App quota read API SHALL expose polish only

`GET /ucg/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ polish: { used, limit } }`，对应当月上海时区桶。本接口 MUST NOT 返回 `voiceAi` 或 `clinicAi` 字段。

#### Scenario: 登录用户查询润笔额度

- **WHEN** wxId=1001 请求 `/ucg/app/api/ai-quota` 且当月润笔已用 2、上限 5
- **THEN** `polish.used` SHALL 为 2 且 `polish.limit` SHALL 为 5

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

ucg-service SHALL implement REST endpoints (also reachable via gateway proxy) including: profile get/update, feed recommend/following, posts CRUD, media presign, follow/unfollow, likes, comments, conversations and messages list.

除评论列表外，分页 MUST 使用 `page`（从 1 开始）与 `pageSize`（默认 20，最大 50），响应 MUST 包含 `{ list, total, page, pageSize }`。

评论列表 `GET /ucg/app/api/posts/{id}/comments` SHALL **不适用**上述分页契约：SHALL 单次返回该帖评论全量列表（见下条 Requirement）。`page`/`pageSize` 查询参数 SHALL 被忽略或废弃，MUST NOT 再驱动服务端 `OFFSET` 分页。

#### Scenario: 推荐 Feed 分页

- **WHEN** `GET /ucg/app/api/feed/recommend?page=1&pageSize=20`
- **THEN** 响应 SHALL 仅含 `status=2` 的帖子，且 SHALL 含分页字段 `list`、`total`、`page`、`pageSize`

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

#### Scenario: 未读计数
- **WHEN** 收件人收到新消息
- **THEN** 其 `unread_count` SHALL 递增直至调用 read API

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

---

## ucg-recommend-feed

<!-- source: openspec/specs/ucg-recommend-feed/spec.md -->

# ucg-recommend-feed Specification

## Purpose
TBD - created by archiving change add-ucg-module. Update Purpose after archive.
## Requirements
### Requirement: Recommend feed SHALL use mixed ranking algorithm

Recommend feed SHALL rank published posts using mixed score combining new-post weight and engagement decay (likes/comments age decay). Scores MUST be persisted in `ucg_post_recommend` (or equivalent materialized store). The platform MUST update scores via **MQ event-driven single-post recompute** and **hot-zone paginated incremental reconciler** only. The platform **MUST NOT** periodically refresh scores by loading all `status=2` posts without pagination (全表刷新).

Hot zone MUST be defined as posts with `published_at >= round_hot_cutoff` where `round_hot_cutoff` is fixed at the start of each hot reconciler round (MUST NOT be recomputed with `NOW()` during pagination within the same round). Cold zone posts MUST rely on MQ events only for score updates; cold-zone paginated reconciler MUST NOT run. MQ-driven scores MAY be briefly stale; hot-zone reconciler MUST provide eventual consistency including time decay.

#### Scenario: 新帖权重

- **WHEN** 两条帖子互动相同但新帖发布时间更近
- **THEN** 推荐排序 SHALL 倾向较新帖子（在衰减窗口内）

#### Scenario: 仅 published 入推荐

- **WHEN** 计算推荐候选集
- **THEN** 算法 SHALL 仅包含 `status=2` 帖子

#### Scenario: 禁止全表刷新

- **WHEN** 后台任务更新 `ucg_post_recommend`
- **THEN** 系统 MUST NOT 执行 `SELECT` 全部 published 帖且无 `LIMIT` 的批量重算

#### Scenario: 发帖过审后入推荐排序

- **WHEN** 帖子审核通过变为 published 且 `ucg.post.published` 已处理
- **THEN** `ucg_post_recommend` MUST 存在该 `postId` 的 score 行供 Feed JOIN 排序

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

- `published` / `liked` / `unliked` / `comment.published` / `comment.removed` → `RecomputeRecommendScore(postId)` UPSERT `ucg_post_recommend`
- `unpublished` → `DELETE FROM ucg_post_recommend WHERE post_id=postId`

MUST NOT 在 consumer 内对全部 published 帖做无 LIMIT 全表扫描。

#### Scenario: unpublished 删除推荐行且永远 Ack

- **WHEN** consumer 收到 `ucg.post.unpublished` 且 `postId` 合法
- **THEN** 系统 MUST 执行 DELETE；`RowsAffected=0` MUST 视为成功且 MUST NOT 报错；处理完成后 MUST Ack 且 MUST NOT Nack requeue

### Requirement: like 类事件 throttle MUST 仅限制重算频率且 SHALL 允许短期误差

对 `ucg.post.liked`、`ucg.post.unliked`、`ucg.comment.published`、`ucg.comment.removed`，consumer MUST 对 `postId` 使用 Redis 单 key `SET NX EX` throttle（默认 500ms）。窗口内 NX 失败 MUST 跳过本次重算并 Ack。throttle MUST 仅保证 500ms 内最多 1 次 `RecomputeRecommendScore`；MUST NOT 保证反映每一次 like/unlike 方向变化。Publisher MUST NOT 在发送侧合并事件。

#### Scenario: 500ms 内多次 like 只重算一次

- **WHEN** 同一 `postId` 在 500ms 内收到 3 条 `ucg.post.liked`
- **THEN** consumer MUST 最多执行 1 次 `RecomputeRecommendScore`，且 3 条消息 MUST 均 Ack

#### Scenario: like 后 unlike 可能被 throttle 跳过

- **WHEN** 同一 `postId` 先收到 `ucg.post.liked` 触发重算，500ms 内又收到 `ucg.post.unliked` 且 throttle NX 失败
- **THEN** consumer MAY 跳过 unlike 触发的重算且 MUST Ack；热区 reconciler MUST 在后续扫描中将 score 收敛到正确值

### Requirement: 热区 reconciler MUST 分页增量且轮首固定 hotCutoff

系统 MUST 使用 MySQL 表 `ucg_recommend_hot_scan_state`（含 `last_post_id` 与 `round_hot_cutoff`）驱动热区分页 reconciler。当 `last_post_id=0` 开始新轮时 MUST 计算并持久化 `round_hot_cutoff`；同一轮后续分页 MUST 使用已存 `round_hot_cutoff` 且 MUST NOT 在分页过程中用 `NOW()` 重新计算 cutoff。每 tick MUST 使用 `published_at >= round_hot_cutoff AND id > last_post_id ORDER BY id LIMIT pageSize`；MUST NOT 一次加载全部 published 帖。

#### Scenario: 热区一轮扫完重置 cursor

- **WHEN** 某 tick 返回行数小于 `pageSize`
- **THEN** reconciler MUST 将 `last_post_id` 置 0；下轮开始 MUST 重新计算 `round_hot_cutoff`

#### Scenario: 续扫同一轮不重算 cutoff

- **WHEN** `last_post_id > 0` 且 reconciler 执行下一页
- **THEN** 系统 MUST 继续使用表中 `round_hot_cutoff` 且 MUST NOT 更新为新的 `now - hotWindow`

### Requirement: 热区 reconciler MUST 周期性重算即使无互动

在冷区零兜底前提下，热区 reconciler MUST 对每个扫到的 published 热区帖执行 `RecomputeRecommendScore`，即使该帖在扫描周期内无任何 like/comment 事件。该行为 MUST 用于热区时间衰减与 throttle 误差的最终一致收敛。

#### Scenario: 无互动帖仍更新 score

- **WHEN** 热区 reconciler 扫描到 `postId` 且该帖在上一扫描周期内 like/comment 计数未变
- **THEN** 系统 MUST 仍执行 `RecomputeRecommendScore(postId)` 以反映 `exp(-age/τ)` 变化

### Requirement: 冷区 MUST NOT 运行分页 reconciler

`published_at < round_hot_cutoff`（冷区）的帖子 MUST 仅通过 MQ 事件更新推荐分；系统 MUST NOT 为冷区启动定时全量或分页扫表任务。

#### Scenario: 冷区帖仅靠互动更新

- **WHEN** 冷区帖收到 `ucg.post.liked` 且 throttle 允许重算
- **THEN** 系统 MUST 更新该帖 `ucg_post_recommend.score` 且 MUST NOT 依赖冷区 reconciler

### Requirement: ucg-service AMQP audit 与 recommend consumer SHALL 共用单 connection

`ucg-service` 内 UCG 审核队列 consumer 与推荐分 consumer MUST 共用 **一条** AMQP connection；每个消费队列 MUST 使用 **独立 channel** 与独立 prefetch 配置。连接断线 MUST 统一 backoff 重连并恢复全部 channel Consume。

#### Scenario: 单 connection 多 channel

- **WHEN** ucg-service 启动且 audit 与 recommend consumer 均 enabled
- **THEN** 进程 MUST 仅建立 1 条到 RabbitMQ 5672 的 AMQP connection，且 audit 4 队列与 recommend 1 队列各占用独立 channel

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

## voice-admin-ui

<!-- source: openspec/specs/voice-admin-ui/spec.md -->

# voice-admin-ui Specification

## Purpose
TBD - created by archiving change refactor-ai-quota-domain-ownership. Update Purpose after archive.
## Requirements
### Requirement: voice-admin HTML SHALL configure voice_ai and clinic_ai quota

系统 MUST 提供独立 Admin 页面 **`/device/admin/voice-admin.html`**（静态文件 `resource/public/voice-admin.html`），包含「AI 配置」功能区：全局默认表单（`voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`）与 per-wxId override 表单（query/body 含 `wxId`）。页面 MUST 调用 **`/voice/admin/api/ai-quota/default`** 与 **`/voice/admin/api/ai-quota/user`**（经 gateway-app 反代至 voice-service）。页面 MUST NOT 包含润笔（polish）字段。

#### Scenario: 管理员修改喂养 AI 全局默认

- **WHEN** 运维在 voice-admin 页提交 voiceAi=5、clinicAi=30
- **THEN** 页面 SHALL 调用 PUT `/voice/admin/api/ai-quota/default` 且 voice-service 配置 SHALL 更新

#### Scenario: 页面不含润笔字段

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 SHALL NOT 展示 `polishMonthlyLimit` 输入控件

### Requirement: Admin Hub SHALL link voice-admin alongside ucg-admin

`resource/public/admin-modules.js` MUST 增加 voice-admin 模块入口，模式对齐 ucg-admin：`id: voice-admin`、导航至 `/device/admin/voice-admin.html`。Hub 登录后 MUST 可点击进入 voice-admin。

#### Scenario: Hub 导航可见 voice-admin

- **WHEN** 管理员登录 Admin Hub
- **THEN** 模块列表 SHALL 包含 voice-admin 入口且链接至 `/device/admin/voice-admin.html`

### Requirement: ucg-admin SHALL remove voice and clinic quota fields

`resource/public/ucg-admin.html`「AI 配置」Tab MUST **移除** `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit` 相关表单与 API 字段，**仅保留** `polishMonthlyLimit` 全局默认与 per-wxId override。

#### Scenario: ucg-admin 仅润笔配置

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 SHALL 仅展示 polish 相关字段且 SHALL NOT 调用 voice/clinic 配额 API

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

voice-service MUST 提供 `GET/PUT /voice/admin/api/ai-quota/default` 与 `GET/PUT /voice/admin/api/ai-quota/user`（query/body 含 `wxId`），认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（gateway 经 `VOICE_ADMIN_PASSWORD` 注入）。voice-service MUST 本地持久化至 `ai_voice_voice`，MUST NOT 转发 device 或 ucg。PUT default MUST 接受 `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit`（正整数）。PUT user MUST 接受 optional 两字段；空值 SHALL 表示清除该 feature override。

#### Scenario: 管理员修改全局胖宝默认

- **WHEN** 管理员 PUT default 为 voiceAi=5、clinicAi=30
- **THEN** voice 权威配置 SHALL 更新且新用户 check clinic_ai SHALL 使用 limit=30

#### Scenario: voice admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

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

### Requirement: 命中非叶子事件时必须追问且不写 history

当匹配到的事件存在子节点（`EXISTS parent_id = 该 id`）时，voice SHALL NOT 调用 `AddHistory` / 等价写库；SHALL 设置该设备的 pending 子事件上下文；SHALL 回复列出**直接子节点名称**的选择问句；且 SHALL 令 `finishTalk=false` 等待用户下一轮输入。

#### Scenario: 仅说换尿布时追问

- **WHEN** 「换尿布」有子节点「大便」「小便」
- **AND** 用户说「换尿布了」且命中「换尿布」
- **THEN** 系统 SHALL NOT 写入 history
- **AND** 回复 SHALL 引导用户在「大便」「小便」中选择（语义等价即可）
- **AND** `finishTalk` SHALL 为 false

#### Scenario: 三级树下追问直接子节点

- **WHEN** 「换尿布」子节点含「排泄类」，「排泄类」下含「大便」「小便」
- **AND** 用户仅命中「换尿布」
- **THEN** 第一轮追问 SHALL 仅针对「换尿布」的直接子节点（如「排泄类」与其它同级子名）
- **AND** SHALL NOT 在第一轮直接询问「大便还是小便」

### Requirement: pending 期间仅在当前父的直接子节点中匹配

存在 pending 子事件上下文时，voice SHALL 仅在 `pending.ParentEventId` 的**直接**子节点集合中执行文本匹配；命中仍为非叶子则 SHALL 更新 pending 并继续追问；命中叶子则 SHALL 清除 pending 并按原动作类型写 history。

#### Scenario: 第二轮回答大便后落库

- **WHEN** pending 父为「排泄类」且子含「大便」「小便」
- **AND** 用户第二轮说「大便」
- **THEN** 系统 SHALL 清除 pending
- **AND** SHALL 以「大便」叶子 event id 写入 history（在动作流程允许写库时）

### Requirement: pending 为内存态且不跨会话恢复

pending 子事件上下文 SHALL 存储于 voice 进程内存（按 deviceNo 键）；SHALL NOT 写入 Redis 或与 session 同步持久化；会话 TTL 或进程重启后 pending 丢失时，后续输入 SHALL 按无 pending 的新轮次处理。

#### Scenario: 超时后大便按新对话处理

- **WHEN** 用户第一轮触发「换尿布」pending 后长时间无后续
- **AND** pending 已因会话过期或重启而丢失
- **AND** 用户再说「大便」
- **THEN** 系统 SHALL NOT 假定仍在「换尿布」追问上下文中
- **AND** MAY 将「大便」作为独立话术在全树中匹配

### Requirement: 仅叶子事件 id 可写入 history

voice 写入 `history.event_id` 时，所选事件 MUST 为无子节点的叶子；非叶子 id SHALL NOT 作为新 history 行的 event_id。

#### Scenario: 追问完成前无 history 行

- **WHEN** 用户仅命中非叶子「换尿布」且处于追问态
- **THEN** 在该轮及 pending 未清除前 SHALL NOT 产生以「换尿布」为 event_id 的新 history 行

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

