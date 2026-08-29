## Context

邀请码互推已落地；缺运营侧「加群获取邀请码」入口。gateway 已有 APK 落盘目录与 `GET /device/app/apk/*` 匿名下载。cash 合成 feature catalog。探索定稿：强制 `er_code.png`、页级展示、gateway 落盘 + cash 记有效期、过期 App 不返回但 Admin 可预览。

## Goals / Non-Goals

**Goals:**

- Admin 上传覆盖唯一 `er_code.png`；配置 `expires_at`。
- catalog 顶层条件返回 `inviteGroupQrUrl`。
- Admin 过期仍可预览。

**Non-Goals:**

- 不定时删文件；不支持多图/多码；不把 QR 挂到单个 feature 项上；不改邀请兑码规则。

## Decisions

### D1：文件与下载

- 磁盘名强制 `er_code.png`；上传非 png 则拒绝或转存为该名（实现选拒绝非 png，简单）。
- 落盘目录：`gatewayapp.ApkStorageDir()`。
- 对外路径：`/device/app/apk/er_code.png`；catalog URL 拼 `PublicBaseURL` + path + `?v={updatedAt}`。

### D2：职责拆分

- gateway：`POST` Admin 上传（multipart），鉴权同其它 Admin；只写文件，可回传相对 path。
- cash：表 `invite_group_qr`（singleton id=1）：`expires_at`、`updated_at`、`file_name`（恒 `er_code.png`）；Admin GET/POST 有效期；上传成功后 Admin 页再调 cash 刷新 `updated_at`（或 gateway 回调——一期由 Admin 串行：upload → cash touch）。

### D3：catalog

- `CashFeatureCatalogRes` 增加可选 `inviteGroupQrUrl string`。
- 条件：`expires_at > now && expires_at > 0`；`expires_at==0` 对 App 视为失效。
- 不检查磁盘是否真实存在（避免 cash 读 gateway 盘）；依赖运维先上传。

### D4：Admin 预览

- 预览 URL 始终可用（相对或绝对 apk 路径 + v=）；旁注是否已对 App 过期。

### D5：usage

- catalog 已在 maintenance_skip；Admin 上传不计 App usage。

## Risks / Trade-offs

- [文件在 gateway、元数据在 cash，不同步] → Admin UI 强制先上传再保存有效期；文档说明。
- [CDN/客户端缓存旧图] → `?v=updatedAt`。
- [publicBaseUrl 未配] → 可返回相对 path，客户端拼网关基址（与 APK 下载一致）。

## Migration Plan

1. 部署 gateway（上传）+ cash（表与 catalog）+ Admin HTML。
2. 运维上传 png 并设有效期。
3. 无历史数据迁移。

## Open Questions

- 无。
