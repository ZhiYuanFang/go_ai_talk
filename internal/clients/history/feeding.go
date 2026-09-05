// Package history 出站调用 history-service 的中立 HTTP 客户端。
package history

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// FeedingDayCount 与 history 内部契约字段对齐。
type FeedingDayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// FeedingDayStatsData Days[0]=上海昨天，向过去，不含今日；长度=请求 days。
type FeedingDayStatsData struct {
	DeviceNo string                   `json:"deviceNo"`
	Days     []FeedingDayCount `json:"days"`
}

type feedingEnvelope struct {
	Code    int                        `json:"code"`
	Message string                     `json:"message"`
	Data    FeedingDayStatsData `json:"data"`
}

// FetchFeedingDayStats 经 HISTORY_SERVICE_URL 拉取近 days 个上海已闭合日每日条数；失败 fail-closed。
//
// days 等于场景 requiredDays（窗口长度不变；锚点已由 history 平移至昨天起）。
// Side Effects: 出站 HTTP；禁止直查 history 库。
func FetchFeedingDayStats(ctx context.Context, deviceNo string, days int) (*FeedingDayStatsData, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("HISTORY_SERVICE_URL")), "/")
	if base == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "未配置 HISTORY_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	if days <= 0 {
		days = 7
	}
	u, err := url.Parse(base + "/history/internal/api/feeding-day-stats")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("deviceNo", deviceNo)
	q.Set("days", fmt.Sprintf("%d", days))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/history] history feeding-day-stats 调用失败 deviceNo=%s err=%v", deviceNo, err)
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "history 资格取数失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[clients/history] history feeding-day-stats HTTP=%d body=%s", resp.StatusCode, string(body))
		return nil, gerror.NewCode(gcode.CodeInternalError, "history 资格取数失败")
	}
	var env feedingEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "history 响应解析失败")
	}
	if env.Code != 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "history 资格取数失败: "+env.Message)
	}
	return &env.Data, nil
}
