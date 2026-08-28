package cash

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

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// historyFeedingDayCount 与 history 内部契约字段对齐。
type historyFeedingDayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type historyFeedingDayStatsData struct {
	DeviceNo string                   `json:"deviceNo"`
	Days     []historyFeedingDayCount `json:"days"`
}

type historyEnvelope struct {
	Code    int                        `json:"code"`
	Message string                     `json:"message"`
	Data    historyFeedingDayStatsData `json:"data"`
}

// FetchFeedingDayStats 经 HISTORY_SERVICE_URL 拉取按日 history 条数；失败 fail-closed。
// Side Effects: 出站 HTTP；禁止直查 history 库。
func FetchFeedingDayStats(ctx context.Context, deviceNo string, days int) (*historyFeedingDayStatsData, error) {
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
		days = 14
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
	req.Header.Set(HeaderInternalSecret, secret)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[cash] history feeding-day-stats 调用失败 deviceNo=%s err=%v", deviceNo, err)
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "history 资格取数失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[cash] history feeding-day-stats HTTP=%d body=%s", resp.StatusCode, string(body))
		return nil, gerror.NewCode(gcode.CodeInternalError, "history 资格取数失败")
	}
	var env historyEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "history 响应解析失败")
	}
	if env.Code != 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "history 资格取数失败: "+env.Message)
	}
	return &env.Data, nil
}
