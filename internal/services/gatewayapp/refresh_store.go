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

// refreshPayloadWxZeroPrefix 设备号直连登录（JWT sub=0）时 Redis 载荷前缀，后接 device_no，与纯数字 wxId 字符串区分。
const refreshPayloadWxZeroPrefix = "0:"

var refreshCache = cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})

// IssueRefreshToken 生成 refresh 并写入 Redis，返回明文令牌（仅返回一次给客户端）。
// wxID>0 时载荷为十进制 wxId；wxID==0 时（仅设备会话）须传非空 deviceNoCarry，载荷为 "0:"+device_no，供刷新时写回 access 的 device_no claim。
func IssueRefreshToken(ctx context.Context, wxID int64, deviceNoCarry string) (string, error) {
	dn0 := strings.TrimSpace(deviceNoCarry)
	if wxID == 0 && dn0 == "" {
		return "", fmt.Errorf("wxId 为 0 的会话签发 refresh 时必须携带 deviceNo")
	}
	if wxID < 0 {
		return "", fmt.Errorf("wxId 无效")
	}
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
	val := strconv.FormatInt(wxID, 10)
	if wxID == 0 {
		val = refreshPayloadWxZeroPrefix + dn0
	}
	if err := refreshCache.SetEX(ctx, key, val, time.Duration(ttl)*time.Second); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeRefreshToken 校验 refresh，返回 wxId 与设备号携带值（仅 wxID==0 时非空）；若 rotate 为 true 则删除旧键（单次使用策略）。
func ConsumeRefreshToken(ctx context.Context, refreshToken string, rotate bool) (wxID int64, deviceNoCarry string, err error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return 0, "", fmt.Errorf("refreshToken 不能为空")
	}
	key := refreshKeyPrefix + refreshToken
	v, ok, err := refreshCache.Get(ctx, key)
	if err != nil {
		return 0, "", err
	}
	v = strings.TrimSpace(v)
	if !ok || v == "" {
		return 0, "", fmt.Errorf("refresh 无效或已过期")
	}
	if strings.HasPrefix(v, refreshPayloadWxZeroPrefix) {
		dn := strings.TrimSpace(v[len(refreshPayloadWxZeroPrefix):])
		if dn == "" {
			return 0, "", fmt.Errorf("refresh 载荷无效")
		}
		if rotate {
			_ = refreshCache.Del(ctx, key)
		}
		return 0, dn, nil
	}
	wxID, err = strconv.ParseInt(v, 10, 64)
	if err != nil || wxID <= 0 {
		return 0, "", fmt.Errorf("refresh 载荷无效")
	}
	if rotate {
		_ = refreshCache.Del(ctx, key)
	}
	return wxID, "", nil
}
