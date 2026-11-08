// 预测 Live Activity：用户级 1/min 限流仍用于普通推送；Imminent 单槽 start/end 已由
// history 统一 Refresh（clients/history.RefreshBabyLiveActivity）取代。
package voice

import (
	"context"
	"os"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	predictLiveActivityEnabledEnv = "PUSH_LIVE_ACTIVITY_ENABLED"
	predictImminentRateLimitTTL   = 60 * time.Second
)

// predictLiveActivityEnabled 与 push-service / history Refresh 共用 env；默认关。
func predictLiveActivityEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(predictLiveActivityEnabledEnv)))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

// tryOccupyImminentRateLimit 用户级 1/min 先到占坑；未过闸返回 false（用于普通可见推送）。
func tryOccupyImminentRateLimit(ctx context.Context, wxID int64) bool {
	key, err := cachekit.PredictImminentRateLimitKey(wxID)
	if err != nil {
		glog.Warningf(ctx, "[predict-la] rate_key_err wxId=%d err=%v", wxID, err)
		return false
	}
	ok, nxErr := predictImminentCache.SetNXEX(ctx, key, "1", predictImminentRateLimitTTL)
	if nxErr != nil {
		glog.Warningf(ctx, "[predict-la] rate_setnx_err wxId=%d err=%v", wxID, nxErr)
		return false
	}
	return ok
}
