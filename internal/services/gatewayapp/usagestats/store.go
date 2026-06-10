package usagestats

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	dayKeyTTLSeconds = 90 * 24 * 3600
	crossFieldSep    = "\x1f"
)

// APIListItem API 频率列表项。
type APIListItem struct {
	ApiKey   string `json:"apiKey"`
	Summary  string `json:"summary"`
	Count    int64  `json:"count"`
	LastAt   int64  `json:"lastAt"`
}

// UserListItem 某 API 的 wxId 调用项。
type UserListItem struct {
	WxId   int64 `json:"wxId"`
	Count  int64 `json:"count"`
	LastAt int64 `json:"lastAt"`
}

// UserAPIItem 某用户的 API 调用项。
type UserAPIItem struct {
	ApiKey  string `json:"apiKey"`
	Summary string `json:"summary"`
	Count   int64  `json:"count"`
	LastAt  int64  `json:"lastAt"`
}

// RecordAsync 异步写入成功调用统计；失败静默并打 warning。
func RecordAsync(wxId int64, apiKey string, at time.Time) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := record(ctx, wxId, apiKey, at); err != nil {
			glog.Warningf(ctx, "[usagestats] 写入 Redis 失败 apiKey=%s wxId=%d err=%v", apiKey, wxId, err)
		}
	}()
}

func record(ctx context.Context, wxId int64, apiKey string, at time.Time) error {
	day := at.Format("20060102")
	ts := at.Unix()

	globalKey := dayGlobalKey(day)
	if _, err := g.Redis().Do(ctx, "HINCRBY", globalKey, apiKey, 1); err != nil {
		return err
	}
	_, _ = g.Redis().Do(ctx, "EXPIRE", globalKey, dayKeyTTLSeconds)

	if _, err := g.Redis().Do(ctx, "HSET", keyLastGlobal(), apiKey, ts); err != nil {
		return err
	}

	if wxId > 0 {
		wxKey := dayWxKey(day, wxId)
		if _, err := g.Redis().Do(ctx, "HINCRBY", wxKey, apiKey, 1); err != nil {
			return err
		}
		_, _ = g.Redis().Do(ctx, "EXPIRE", wxKey, dayKeyTTLSeconds)

		crossKey := dayCrossKey(day)
		if _, err := g.Redis().Do(ctx, "HINCRBY", crossKey, crossField(apiKey, wxId), 1); err != nil {
			return err
		}
		_, _ = g.Redis().Do(ctx, "EXPIRE", crossKey, dayKeyTTLSeconds)

		if _, err := g.Redis().Do(ctx, "HSET", keyLastWx(wxId), apiKey, ts); err != nil {
			return err
		}
	}
	return nil
}

// ListAPIs 聚合窗口内 API 频率。
func ListAPIs(ctx context.Context, days int, summaryFn func(apiKey string) string) ([]APIListItem, error) {
	counts := make(map[string]int64)
	lastAt := make(map[string]int64)
	for _, day := range dayRange(days) {
		key := dayGlobalKey(day)
		all, err := g.Redis().Do(ctx, "HGETALL", key)
		if err != nil {
			return nil, err
		}
		for k, v := range redisHashToMap(all) {
			n, _ := strconv.ParseInt(v, 10, 64)
			counts[k] += n
		}
	}
	allLast, err := g.Redis().Do(ctx, "HGETALL", keyLastGlobal())
	if err != nil {
		return nil, err
	}
	for k, v := range redisHashToMap(allLast) {
		n, _ := strconv.ParseInt(v, 10, 64)
		lastAt[k] = n
	}
	out := make([]APIListItem, 0, len(counts))
	for apiKey, cnt := range counts {
		out = append(out, APIListItem{
			ApiKey:  apiKey,
			Summary: summaryFn(apiKey),
			Count:   cnt,
			LastAt:  lastAt[apiKey],
		})
	}
	sortAPIList(out)
	return out, nil
}

// ListUsersForAPI 某 API 在窗口内的 wxId 分布。
func ListUsersForAPI(ctx context.Context, days int, apiKey string) ([]UserListItem, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey 不能为空")
	}
	counts := make(map[int64]int64)
	lastAt := make(map[int64]int64)
	prefix := apiKey + crossFieldSep

	for _, day := range dayRange(days) {
		key := dayCrossKey(day)
		all, err := g.Redis().Do(ctx, "HGETALL", key)
		if err != nil {
			return nil, err
		}
		for field, val := range redisHashToMap(all) {
			if !strings.HasPrefix(field, prefix) {
				continue
			}
			wxPart := strings.TrimPrefix(field, prefix)
			wxId, err := strconv.ParseInt(wxPart, 10, 64)
			if err != nil || wxId <= 0 {
				continue
			}
			n, _ := strconv.ParseInt(val, 10, 64)
			counts[wxId] += n
		}
	}
	// 最近时间从 per-wx last hash 读取
	for wxId := range counts {
		raw, err := g.Redis().Do(ctx, "HGET", keyLastWx(wxId), apiKey)
		if err != nil {
			continue
		}
		if s := redisString(raw); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				lastAt[wxId] = n
			}
		}
	}
	out := make([]UserListItem, 0, len(counts))
	for wxId, cnt := range counts {
		out = append(out, UserListItem{WxId: wxId, Count: cnt, LastAt: lastAt[wxId]})
	}
	sortUserList(out)
	return out, nil
}

// ListAPIsForUser 某 wxId 在窗口内的 API 分布。
func ListAPIsForUser(ctx context.Context, days int, wxId int64, summaryFn func(apiKey string) string) ([]UserAPIItem, error) {
	if wxId <= 0 {
		return nil, fmt.Errorf("wxId 须为正整数")
	}
	counts := make(map[string]int64)
	for _, day := range dayRange(days) {
		key := dayWxKey(day, wxId)
		all, err := g.Redis().Do(ctx, "HGETALL", key)
		if err != nil {
			return nil, err
		}
		for k, v := range redisHashToMap(all) {
			n, _ := strconv.ParseInt(v, 10, 64)
			counts[k] += n
		}
	}
	lastMap := make(map[string]int64)
	allLast, err := g.Redis().Do(ctx, "HGETALL", keyLastWx(wxId))
	if err != nil {
		return nil, err
	}
	for k, v := range redisHashToMap(allLast) {
		n, _ := strconv.ParseInt(v, 10, 64)
		lastMap[k] = n
	}
	out := make([]UserAPIItem, 0, len(counts))
	for apiKey, cnt := range counts {
		out = append(out, UserAPIItem{
			ApiKey:  apiKey,
			Summary: summaryFn(apiKey),
			Count:   cnt,
			LastAt:  lastMap[apiKey],
		})
	}
	sortUserAPIList(out)
	return out, nil
}

func dayGlobalKey(day string) string  { return "gw:usage:d:" + day + ":g" }
func dayWxKey(day string, wxId int64) string {
	return fmt.Sprintf("gw:usage:d:%s:w:%d", day, wxId)
}
func dayCrossKey(day string) string { return "gw:usage:d:" + day + ":x" }
func keyLastGlobal() string         { return "gw:usage:last:g" }
func keyLastWx(wxId int64) string   { return fmt.Sprintf("gw:usage:last:w:%d", wxId) }

func crossField(apiKey string, wxId int64) string {
	return apiKey + crossFieldSep + strconv.FormatInt(wxId, 10)
}

func dayRange(days int) []string {
	now := time.Now()
	if days <= 0 {
		// 全部：在 TTL 内取 90 天
		days = 90
	}
	out := make([]string, 0, days)
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -i)
		out = append(out, d.Format("20060102"))
	}
	return out
}

func redisHashToMap(v interface{}) map[string]string {
	out := make(map[string]string)
	if v == nil {
		return out
	}
	if m, ok := v.(map[string]interface{}); ok {
		for k, val := range m {
			out[k] = redisString(val)
		}
		return out
	}
	if m, ok := v.(map[string]string); ok {
		return m
	}
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return out
	}
	for i := 0; i+1 < len(arr); i += 2 {
		k := redisString(arr[i])
		val := redisString(arr[i+1])
		if k != "" {
			out[k] = val
		}
	}
	return out
}

func redisString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(v)
	}
}

func sortAPIList(list []APIListItem) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Count > list[i].Count || (list[j].Count == list[i].Count && list[j].ApiKey < list[i].ApiKey) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortUserList(list []UserListItem) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Count > list[i].Count || (list[j].Count == list[i].Count && list[j].WxId < list[i].WxId) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortUserAPIList(list []UserAPIItem) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Count > list[i].Count || (list[j].Count == list[i].Count && list[j].ApiKey < list[i].ApiKey) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}
