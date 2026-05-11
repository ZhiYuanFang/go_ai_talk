package gatewayapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

const refreshKeyPrefix = "gw:app:rt:"

var refreshCache = cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})

// IssueRefreshToken 生成 refresh 并写入 Redis，返回明文令牌（仅返回一次给客户端）。
func IssueRefreshToken(ctx context.Context, wxID int64) (string, error) {
	ttl := g.Cfg().MustGet(ctx, "gatewayApp.refreshTtlSeconds").Int64()
	if ttl <= 0 {
		ttl = 604800
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	key := refreshKeyPrefix + token
	if err := refreshCache.SetEX(ctx, key, strconv.FormatInt(wxID, 10), time.Duration(ttl)*time.Second); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeRefreshToken 校验 refresh，返回 wxId；若 rotate 为 true 则删除旧键（单次使用策略）。
func ConsumeRefreshToken(ctx context.Context, refreshToken string, rotate bool) (wxID int64, err error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return 0, fmt.Errorf("refreshToken 不能为空")
	}
	key := refreshKeyPrefix + refreshToken
	v, ok, err := refreshCache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if !ok || strings.TrimSpace(v) == "" {
		return 0, fmt.Errorf("refresh 无效或已过期")
	}
	wxID, err = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil || wxID <= 0 {
		return 0, fmt.Errorf("refresh 载荷无效")
	}
	if rotate {
		_ = refreshCache.Del(ctx, key)
	}
	return wxID, nil
}
