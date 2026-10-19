## 1. voice 代理

- [x] 1.1 新增 `GET /voice/admin/api/intent-cues/by-document`，管理员校验后转发 Python `/v1/admin/cues/by-document`；空白 document 直接拒绝
- [x] 1.2 中文业务注释；不把该 admin 路径计入 App usage（沿用 `/voice/admin/api/` 跳过即可，不改 denylist）

## 2. 门户

- [x] 2.1 `intent-agent-admin.html` 增加原文输入与查询；命中表一行一条，展示维度、id、payload
- [x] 2.2 按 dimension 调用已有 `*-vectors/{id}` 删除；未知维度不提供删除；空输入不请求
