## ADDED Requirements

### Requirement: Hub 运维页统一近全宽

凡同时使用 `pangbao-theme.css` 与 `admin-pages.css` 的 Hub 运维静态页，内容壳 MUST 为近全宽（视口减去约 32px 边距），MUST NOT 再以约 1180px 或 1400px 硬封顶；原 `page-narrow` / `hub-page-wide` MUST 与默认 Hub 宽度一致。

#### Scenario: 宽屏无大块两侧留白

- **WHEN** 运维在宽屏打开任意 Hub 运维页（含统计、功能管理、向量、设备历史等）
- **THEN** 主内容区接近视口宽度，仅保留约 32px 级外边距

#### Scenario: 非 Hub 协议页不受影响

- **WHEN** 用户打开仅使用 theme、不加载 `admin-pages.css` 的协议/官网页
- **THEN** 仍使用 theme 原有 `.page` 宽度约定
