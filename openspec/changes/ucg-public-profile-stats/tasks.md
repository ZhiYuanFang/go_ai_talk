## 1. 实现

- [x] 1.1 `GetPublicProfile` 返回前调用 `enrichProfileStats(ctx, wxID, dto)`
- [x] 1.2 `go build ./internal/services/ucg/...` 通过

## 2. 验收

- [ ] 2.1 联调：sim 关注真人后 `GET /profile/{simWxId}` 的 `followingCount >= 1`
- [ ] 2.2 联调：App 他人主页「关注 N」与 API 一致
