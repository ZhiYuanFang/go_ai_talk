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

const refreshPayloadWxZeroPrefix = "0:"

// refreshPayloadWxDeviceSep wx 会话 refresh 载荷中 wxId 与 device_no 的分隔（格式 `wxId:device_no`；旧载荷仅 wxId 仍兼容）。
const refreshPayloadWxDeviceSep = ":"

// refreshPayloadEpochSep 会话世代分隔：载荷末尾 `|e{epoch}`（改名/吊销时 bump epoch 使旧 refresh 失效）。
const refreshPayloadEpochSep = "|e"

var refreshCache = cachekit.Default()

func refreshTTL(ctx context.Context) time.Duration {
	ttl := g.Cfg().MustGet(ctx, "gatewayApp.refreshTtlSeconds").Int64()
	if ttl <= 0 {
		ttl = 604800
	}
	return time.Duration(ttl) * time.Second
}

// currentSessionEpoch 读取当前会话世代；键不存在视为 0。
func currentSessionEpoch(ctx context.Context, wxID int64, deviceNoCarry string) (int64, error) {
	var key string
	if wxID > 0 {
		key = cachekit.GatewaySessionEpochWxKey(wxID)
	} else {
		dn := strings.TrimSpace(deviceNoCarry)
		if dn == "" {
			return 0, fmt.Errorf("纯设备会话 epoch 需要 deviceNo")
		}
		key = cachekit.GatewaySessionEpochDeviceKey(dn)
	}
	raw, ok, err := refreshCache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if !ok || strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || n < 0 {
		return 0, nil
	}
	return n, nil
}

// appendEpoch 在 refresh 载荷末尾附加 `|e{epoch}`。
func appendEpoch(payload string, epoch int64) string {
	return strings.TrimSpace(payload) + refreshPayloadEpochSep + strconv.FormatInt(epoch, 10)
}

// splitEpoch 解析载荷末尾世代；无后缀则 epoch=0（兼容旧 refresh）。
func splitEpoch(raw string) (payload string, epoch int64, err error) {
	raw = strings.TrimSpace(raw)
	idx := strings.LastIndex(raw, refreshPayloadEpochSep)
	if idx < 0 {
		return raw, 0, nil
	}
	payload = strings.TrimSpace(raw[:idx])
	epStr := strings.TrimSpace(raw[idx+len(refreshPayloadEpochSep):])
	epoch, err = strconv.ParseInt(epStr, 10, 64)
	if err != nil || epoch < 0 {
		return "", 0, fmt.Errorf("refresh 世代无效")
	}
	return payload, epoch, nil
}

// IssueRefreshToken 生成 refresh 并写入 Redis，返回明文令牌（仅返回一次给客户端）。
// wxID>0 时载荷为十进制 wxId，若 deviceNoCarry 非空则为 `wxId:device_no`；wxID==0 时（仅设备会话）须传非空 deviceNoCarry，载荷为 "0:"+device_no。
// 载荷末尾附加当前 session epoch（`|eN`），吊销时 bump epoch 即可使旧 token 失效。
func IssueRefreshToken(ctx context.Context, wxID int64, deviceNoCarry string) (string, error) {
	dn0 := strings.TrimSpace(deviceNoCarry)
	if wxID == 0 && dn0 == "" {
		return "", fmt.Errorf("wxId 为 0 的会话签发 refresh 时必须携带 deviceNo")
	}
	if wxID < 0 {
		return "", fmt.Errorf("wxId 无效")
	}
	epoch, err := currentSessionEpoch(ctx, wxID, dn0)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	key := cachekit.GatewayRefreshTokenKey(token)
	val := strconv.FormatInt(wxID, 10)
	if wxID == 0 {
		val = refreshPayloadWxZeroPrefix + dn0
	} else if dn0 != "" {
		val = val + refreshPayloadWxDeviceSep + dn0
	}
	val = appendEpoch(val, epoch)
	if err := refreshCache.SetEX(ctx, key, val, refreshTTL(ctx)); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeRefreshToken 校验 refresh，返回 wxId 与设备号携带值（wxID==0 或新格式 wxId:device_no 时非空）；若 rotate 为 true 则删除旧键（单次使用策略）。
// 校验 session epoch：与 Redis 当前世代不一致则视为已吊销。
func ConsumeRefreshToken(ctx context.Context, refreshToken string, rotate bool) (wxID int64, deviceNoCarry string, err error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return 0, "", fmt.Errorf("refreshToken 不能为空")
	}
	key := cachekit.GatewayRefreshTokenKey(refreshToken)
	v, ok, err := refreshCache.Get(ctx, key)
	if err != nil {
		return 0, "", err
	}
	v = strings.TrimSpace(v)
	if !ok || v == "" {
		return 0, "", fmt.Errorf("refresh 无效或已过期")
	}
	payload, tokenEpoch, err := splitEpoch(v)
	if err != nil {
		return 0, "", err
	}
	v = payload

	if strings.HasPrefix(v, refreshPayloadWxZeroPrefix) {
		dn := strings.TrimSpace(v[len(refreshPayloadWxZeroPrefix):])
		if dn == "" {
			return 0, "", fmt.Errorf("refresh 载荷无效")
		}
		cur, e2 := currentSessionEpoch(ctx, 0, dn)
		if e2 != nil {
			return 0, "", e2
		}
		if tokenEpoch != cur {
			_ = refreshCache.Del(ctx, key)
			return 0, "", fmt.Errorf("refresh 已吊销")
		}
		if rotate {
			_ = refreshCache.Del(ctx, key)
		}
		return 0, dn, nil
	}
	if idx := strings.Index(v, refreshPayloadWxDeviceSep); idx > 0 {
		wxID, err = strconv.ParseInt(v[:idx], 10, 64)
		if err != nil || wxID <= 0 {
			return 0, "", fmt.Errorf("refresh 载荷无效")
		}
		deviceNoCarry = strings.TrimSpace(v[idx+len(refreshPayloadWxDeviceSep):])
	} else {
		wxID, err = strconv.ParseInt(v, 10, 64)
		if err != nil || wxID <= 0 {
			return 0, "", fmt.Errorf("refresh 载荷无效")
		}
	}
	cur, e2 := currentSessionEpoch(ctx, wxID, deviceNoCarry)
	if e2 != nil {
		return 0, "", e2
	}
	if tokenEpoch != cur {
		_ = refreshCache.Del(ctx, key)
		return 0, "", fmt.Errorf("refresh 已吊销")
	}
	if rotate {
		_ = refreshCache.Del(ctx, key)
	}
	return wxID, deviceNoCarry, nil
}

// BumpSessionEpochForWxIDs 吊销指定账号的全部 App refresh（INCR 会话世代）。
// 业务逻辑：宝宝改名后由 device 编排调用；旧 refresh 在 Consume 时因 epoch 不匹配失败。
func BumpSessionEpochForWxIDs(ctx context.Context, wxIDs []int64) error {
	for _, id := range wxIDs {
		if id <= 0 {
			continue
		}
		key := cachekit.GatewaySessionEpochWxKey(id)
		if _, err := refreshCache.Incr(ctx, key); err != nil {
			return fmt.Errorf("bump wx session epoch wxId=%d: %w", id, err)
		}
	}
	return nil
}

// BumpSessionEpochForDeviceNo 吊销纯设备会话（wxId=0）上该 device_no 的 refresh。
func BumpSessionEpochForDeviceNo(ctx context.Context, deviceNo string) error {
	dn := strings.TrimSpace(deviceNo)
	if dn == "" {
		return fmt.Errorf("deviceNo 不能为空")
	}
	key := cachekit.GatewaySessionEpochDeviceKey(dn)
	if _, err := refreshCache.Incr(ctx, key); err != nil {
		return fmt.Errorf("bump device session epoch: %w", err)
	}
	return nil
}
