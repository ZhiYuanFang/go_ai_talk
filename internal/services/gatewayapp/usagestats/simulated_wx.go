package usagestats

import (
	"context"
	"sync"

	"hello/internal/platform/cachekit"
)

var (
	simWxLocalMu sync.RWMutex
	simWxLocal   = map[int64]bool{}
)

// IsSimulatedWx 判断 wxId 是否为模拟用户（Redis SET + 本地缓存 + device internal HTTP 回源）。
func IsSimulatedWx(ctx context.Context, wxID int64) bool {
	if wxID <= 0 {
		return false
	}
	if hit, ok := simWxLocalLookup(wxID); ok {
		return hit
	}
	member := cachekit.GatewayUsageSimWxMember(wxID)
	if ok, err := usageCache.SetIsMember(ctx, cachekit.GatewayUsageSimWxSetKey(), member); err == nil && ok {
		simWxLocalStore(wxID, true)
		return true
	}
	ok, err := fetchWxIsSimulatedHTTP(ctx, wxID)
	if err != nil {
		return false
	}
	if ok {
		_ = usageCache.SetAdd(ctx, cachekit.GatewayUsageSimWxSetKey(), member)
	}
	simWxLocalStore(wxID, ok)
	return ok
}

// MarkSimulatedWxCached 写入 Redis 与本地缓存（device 注册模拟用户后可调用）。
func MarkSimulatedWxCached(ctx context.Context, wxID int64) {
	if wxID <= 0 {
		return
	}
	_ = usageCache.SetAdd(ctx, cachekit.GatewayUsageSimWxSetKey(), cachekit.GatewayUsageSimWxMember(wxID))
	simWxLocalStore(wxID, true)
}

func simWxLocalLookup(wxID int64) (bool, bool) {
	simWxLocalMu.RLock()
	defer simWxLocalMu.RUnlock()
	v, ok := simWxLocal[wxID]
	return v, ok
}

func simWxLocalStore(wxID int64, simulated bool) {
	simWxLocalMu.Lock()
	simWxLocal[wxID] = simulated
	if len(simWxLocal) > 10000 {
		simWxLocal = map[int64]bool{}
	}
	simWxLocalMu.Unlock()
}
