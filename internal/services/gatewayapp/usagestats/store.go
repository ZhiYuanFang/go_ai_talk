package usagestats

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	dayKeyTTLSeconds = 90 * 24 * 3600
	crossFieldSep    = "\x1f"
)

var usageCache = cachekit.Default()

// APIListItem API 频率列表项。
type APIListItem struct {
	ApiKey  string `json:"apiKey"`
	Summary string `json:"summary"`
	Count   int64  `json:"count"`
	LastAt  int64  `json:"lastAt"`
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
		if wxId > 0 && IsSimulatedWx(ctx, wxId) {
			return
		}
		if err := record(ctx, wxId, apiKey, at); err != nil {
			glog.Warningf(ctx, "[usagestats] 写入 Redis 失败 apiKey=%s wxId=%d err=%v", apiKey, wxId, err)
		}
	}()
}

func record(ctx context.Context, wxId int64, apiKey string, at time.Time) error {
	day := at.Format("20060102")
	ts := at.Unix()
	ttl := time.Duration(dayKeyTTLSeconds) * time.Second

	globalKey := cachekit.GatewayUsageDayGlobalKey(day)
	if _, err := usageCache.HashIncrBy(ctx, globalKey, apiKey, 1); err != nil {
		return err
	}
	_ = usageCache.Expire(ctx, globalKey, ttl)

	if err := usageCache.HashSet(ctx, cachekit.GatewayUsageLastGlobalKey(), apiKey, strconv.FormatInt(ts, 10)); err != nil {
		return err
	}

	if wxId > 0 {
		wxKey := cachekit.GatewayUsageDayWxKey(day, wxId)
		if _, err := usageCache.HashIncrBy(ctx, wxKey, apiKey, 1); err != nil {
			return err
		}
		_ = usageCache.Expire(ctx, wxKey, ttl)

		crossKey := cachekit.GatewayUsageDayCrossKey(day)
		if _, err := usageCache.HashIncrBy(ctx, crossKey, crossField(apiKey, wxId), 1); err != nil {
			return err
		}
		_ = usageCache.Expire(ctx, crossKey, ttl)

		if err := usageCache.HashSet(ctx, cachekit.GatewayUsageLastWxKey(wxId), apiKey, strconv.FormatInt(ts, 10)); err != nil {
			return err
		}
	}
	return nil
}

// ListAPIs 聚合窗口内 API 频率；sortBy 为 count（默认）或 lastAt。
func ListAPIs(ctx context.Context, days int, sortBy string, summaryFn func(apiKey string) string) ([]APIListItem, error) {
	counts := make(map[string]int64)
	lastAt := make(map[string]int64)
	for _, day := range dayRange(days) {
		key := cachekit.GatewayUsageDayGlobalKey(day)
		all, err := usageCache.HashGetAll(ctx, key)
		if err != nil {
			return nil, err
		}
		for k, v := range all {
			n, _ := strconv.ParseInt(v, 10, 64)
			counts[k] += n
		}
	}
	allLast, err := usageCache.HashGetAll(ctx, cachekit.GatewayUsageLastGlobalKey())
	if err != nil {
		return nil, err
	}
	for k, v := range allLast {
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
	applyAPISort(out, ParseSortBy(sortBy))
	return out, nil
}

// ListUsersForAPI 某 API 在窗口内的 wxId 分布。
func ListUsersForAPI(ctx context.Context, days int, apiKey string, sortBy string) ([]UserListItem, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey 不能为空")
	}
	counts := make(map[int64]int64)
	lastAt := make(map[int64]int64)
	prefix := apiKey + crossFieldSep

	for _, day := range dayRange(days) {
		key := cachekit.GatewayUsageDayCrossKey(day)
		all, err := usageCache.HashGetAll(ctx, key)
		if err != nil {
			return nil, err
		}
		for field, val := range all {
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
	for wxId := range counts {
		raw, ok, err := usageCache.HashGet(ctx, cachekit.GatewayUsageLastWxKey(wxId), apiKey)
		if err != nil || !ok {
			continue
		}
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			lastAt[wxId] = n
		}
	}
	out := make([]UserListItem, 0, len(counts))
	for wxId, cnt := range counts {
		out = append(out, UserListItem{WxId: wxId, Count: cnt, LastAt: lastAt[wxId]})
	}
	applyUserSort(out, ParseSortBy(sortBy))
	return out, nil
}

// ListAPIsForUser 某 wxId 在窗口内的 API 分布。
func ListAPIsForUser(ctx context.Context, days int, wxId int64, sortBy string, summaryFn func(apiKey string) string) ([]UserAPIItem, error) {
	if wxId <= 0 {
		return nil, fmt.Errorf("wxId 须为正整数")
	}
	counts := make(map[string]int64)
	for _, day := range dayRange(days) {
		key := cachekit.GatewayUsageDayWxKey(day, wxId)
		all, err := usageCache.HashGetAll(ctx, key)
		if err != nil {
			return nil, err
		}
		for k, v := range all {
			n, _ := strconv.ParseInt(v, 10, 64)
			counts[k] += n
		}
	}
	lastMap := make(map[string]int64)
	allLast, err := usageCache.HashGetAll(ctx, cachekit.GatewayUsageLastWxKey(wxId))
	if err != nil {
		return nil, err
	}
	for k, v := range allLast {
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
	applyUserAPISort(out, ParseSortBy(sortBy))
	return out, nil
}

func crossField(apiKey string, wxId int64) string {
	return apiKey + crossFieldSep + strconv.FormatInt(wxId, 10)
}

func dayRange(days int) []string {
	now := time.Now()
	if days <= 0 {
		days = 90
	}
	out := make([]string, 0, days)
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -i)
		out = append(out, d.Format("20060102"))
	}
	return out
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

func sortAPIListByLastAt(list []APIListItem) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].LastAt > list[i].LastAt ||
				(list[j].LastAt == list[i].LastAt && list[j].Count > list[i].Count) ||
				(list[j].LastAt == list[i].LastAt && list[j].Count == list[i].Count && list[j].ApiKey < list[i].ApiKey) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func applyAPISort(list []APIListItem, sortBy string) {
	if sortBy == SortByLastAt {
		sortAPIListByLastAt(list)
		return
	}
	sortAPIList(list)
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

func sortUserListByLastAt(list []UserListItem) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].LastAt > list[i].LastAt ||
				(list[j].LastAt == list[i].LastAt && list[j].Count > list[i].Count) ||
				(list[j].LastAt == list[i].LastAt && list[j].Count == list[i].Count && list[j].WxId < list[i].WxId) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func applyUserSort(list []UserListItem, sortBy string) {
	if sortBy == SortByLastAt {
		sortUserListByLastAt(list)
		return
	}
	sortUserList(list)
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

func sortUserAPIListByLastAt(list []UserAPIItem) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].LastAt > list[i].LastAt ||
				(list[j].LastAt == list[i].LastAt && list[j].Count > list[i].Count) ||
				(list[j].LastAt == list[i].LastAt && list[j].Count == list[i].Count && list[j].ApiKey < list[i].ApiKey) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func applyUserAPISort(list []UserAPIItem, sortBy string) {
	if sortBy == SortByLastAt {
		sortUserAPIListByLastAt(list)
		return
	}
	sortUserAPIList(list)
}
