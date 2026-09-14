// Package clientusage 客户端功能使用统计（gateway-app 本机 Redis）。
//
// 业务说明：App 主动上报 featureId，与 App API 使用统计并存；仅 Redis、可丢。
// 设计：日桶聚合 + 每用户时间线；3 秒限流；模拟用户跳过；窗口/TTL ≈30 天；时间线 ≤1 万条。
package clientusage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
	"hello/internal/services/gatewayapp/usagestats"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// dayKeyTTLSeconds 聚合日桶与时间线 key TTL（约 30 天）。
	dayKeyTTLSeconds = 30 * 24 * 3600
	// maxWindowDays Admin 查询窗口上限。
	maxWindowDays = 30
	// maxTimelineLen 单用户时间线硬顶。
	maxTimelineLen = 10000
	// rateLimitTTL 同 wx 上报间隔。
	rateLimitTTL = 1 * time.Second
	// maxFeatureIDLen featureId 最大长度，防巨型 Hash field。
	maxFeatureIDLen = 128
	// maxDescriptionLen description 最大长度。
	maxDescriptionLen = 128
	crossFieldSep     = "\x1f"
	timelineSep       = "|"
)

var featCache = cachekit.Default()

// FeatureListItem 功能维度聚合项。
type FeatureListItem struct {
	FeatureId   string `json:"featureId"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	LastAt      int64  `json:"lastAt"`
}

// UserCountItem 某功能下的用户项。
type UserCountItem struct {
	WxId   int64 `json:"wxId"`
	Count  int64 `json:"count"`
	LastAt int64 `json:"lastAt"`
}

// UserFeatureItem 某用户的功能项。
type UserFeatureItem struct {
	FeatureId   string `json:"featureId"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	LastAt      int64  `json:"lastAt"`
}

// TimelineItem 时间线条目（事件名 + 描述 + 服务端时间）。
type TimelineItem struct {
	FeatureId   string `json:"featureId"`
	Description string `json:"description"`
	At          int64  `json:"at"`
}

// Report 登录用户上报功能事件。
//
// 业务：校验 featureId/description → sim 跳过 → 3s 限流 → 写聚合 + 时间线 + 描述 registry。
// Returns: 限流/参数错误；sim 成功且不写。
func Report(ctx context.Context, wxID int64, featureID, description string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeNotAuthorized, "须登录后上报")
	}
	featureID = strings.TrimSpace(featureID)
	description = strings.TrimSpace(description)
	if featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	if description == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "description 不能为空")
	}
	if len(featureID) > maxFeatureIDLen {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 过长")
	}
	if len(description) > maxDescriptionLen {
		return gerror.NewCode(gcode.CodeInvalidParameter, "description 过长")
	}
	// 禁止分隔符污染时间线/交叉 field。
	if strings.Contains(featureID, timelineSep) || strings.Contains(featureID, crossFieldSep) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 含非法字符")
	}
	if strings.Contains(description, timelineSep) || strings.Contains(description, crossFieldSep) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "description 含非法字符")
	}

	if usagestats.IsSimulatedWx(ctx, wxID) {
		glog.Debugf(ctx, "[clientusage] skip simulated wxId=%d feature=%s", wxID, featureID)
		return nil
	}

	ok, err := featCache.SetNXEX(ctx, cachekit.GatewayFeatUsageRateLimitKey(wxID), "1", rateLimitTTL)
	if err != nil {
		return err
	}
	if !ok {
		return gerror.NewCode(gcode.CodeInvalidOperation, "上报过于频繁，请稍后再试")
	}

	at := time.Now()
	if err := record(ctx, wxID, featureID, description, at); err != nil {
		glog.Warningf(ctx, "[clientusage] 写入失败 wxId=%d feature=%s err=%v", wxID, featureID, err)
		return err
	}
	return nil
}

func record(ctx context.Context, wxID int64, featureID, description string, at time.Time) error {
	day := at.Format("20060102")
	ts := at.Unix()
	ttl := time.Duration(dayKeyTTLSeconds) * time.Second

	globalKey := cachekit.GatewayFeatUsageDayGlobalKey(day)
	if _, err := featCache.HashIncrBy(ctx, globalKey, featureID, 1); err != nil {
		return err
	}
	_ = featCache.Expire(ctx, globalKey, ttl)
	_ = featCache.HashSet(ctx, cachekit.GatewayFeatUsageLastGlobalKey(), featureID, strconv.FormatInt(ts, 10))
	_ = featCache.HashSet(ctx, cachekit.GatewayFeatUsageDescGlobalKey(), featureID, description)
	_ = featCache.Expire(ctx, cachekit.GatewayFeatUsageDescGlobalKey(), ttl)

	wxKey := cachekit.GatewayFeatUsageDayWxKey(day, wxID)
	if _, err := featCache.HashIncrBy(ctx, wxKey, featureID, 1); err != nil {
		return err
	}
	_ = featCache.Expire(ctx, wxKey, ttl)

	crossKey := cachekit.GatewayFeatUsageDayCrossKey(day)
	if _, err := featCache.HashIncrBy(ctx, crossKey, crossField(featureID, wxID), 1); err != nil {
		return err
	}
	_ = featCache.Expire(ctx, crossKey, ttl)
	_ = featCache.HashSet(ctx, cachekit.GatewayFeatUsageLastWxKey(wxID), featureID, strconv.FormatInt(ts, 10))

	// 时间线：unix|featureId|description；最新在头；裁剪至 10000；TTL 30 天。
	tlKey := cachekit.GatewayFeatUsageTimelineKey(wxID)
	member := fmt.Sprintf("%d%s%s%s%s", ts, timelineSep, featureID, timelineSep, description)
	if err := featCache.ListLPush(ctx, tlKey, member); err != nil {
		return err
	}
	if err := featCache.ListTrim(ctx, tlKey, 0, maxTimelineLen-1); err != nil {
		return err
	}
	_ = featCache.Expire(ctx, tlKey, ttl)
	return nil
}

func crossField(featureID string, wxID int64) string {
	return featureID + crossFieldSep + strconv.FormatInt(wxID, 10)
}

// NormalizeDays 将查询天数规范到 [1,30]；≤0 视为 30。
func NormalizeDays(days int) int {
	if days <= 0 || days > maxWindowDays {
		return maxWindowDays
	}
	return days
}

func dayRange(days int) []string {
	days = NormalizeDays(days)
	now := time.Now()
	out := make([]string, 0, days)
	for i := 0; i < days; i++ {
		out = append(out, now.AddDate(0, 0, -i).Format("20060102"))
	}
	return out
}

// ListFeatures 按功能聚合。
func ListFeatures(ctx context.Context, days int, sortBy string) ([]FeatureListItem, error) {
	counts := make(map[string]int64)
	for _, day := range dayRange(days) {
		all, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageDayGlobalKey(day))
		if err != nil {
			return nil, err
		}
		for k, v := range all {
			n, _ := strconv.ParseInt(v, 10, 64)
			counts[k] += n
		}
	}
	lastAt := make(map[string]int64)
	allLast, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageLastGlobalKey())
	if err != nil {
		return nil, err
	}
	for k, v := range allLast {
		n, _ := strconv.ParseInt(v, 10, 64)
		lastAt[k] = n
	}
	descMap, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageDescGlobalKey())
	if err != nil {
		return nil, err
	}
	out := make([]FeatureListItem, 0, len(counts))
	for id, cnt := range counts {
		out = append(out, FeatureListItem{
			FeatureId: id, Description: descMap[id], Count: cnt, LastAt: lastAt[id],
		})
	}
	sortFeatureList(out, usagestats.ParseSortBy(sortBy))
	return out, nil
}

// ListUsersForFeature 某功能下的用户分布。
func ListUsersForFeature(ctx context.Context, days int, featureID, sortBy string) ([]UserCountItem, error) {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	counts := make(map[int64]int64)
	prefix := featureID + crossFieldSep
	for _, day := range dayRange(days) {
		all, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageDayCrossKey(day))
		if err != nil {
			return nil, err
		}
		for field, val := range all {
			if !strings.HasPrefix(field, prefix) {
				continue
			}
			wxPart := strings.TrimPrefix(field, prefix)
			wxID, e := strconv.ParseInt(wxPart, 10, 64)
			if e != nil || wxID <= 0 {
				continue
			}
			n, _ := strconv.ParseInt(val, 10, 64)
			counts[wxID] += n
		}
	}
	lastAt := make(map[int64]int64)
	for wxID := range counts {
		raw, ok, err := featCache.HashGet(ctx, cachekit.GatewayFeatUsageLastWxKey(wxID), featureID)
		if err != nil || !ok {
			continue
		}
		if n, e := strconv.ParseInt(raw, 10, 64); e == nil {
			lastAt[wxID] = n
		}
	}
	out := make([]UserCountItem, 0, len(counts))
	for wxID, cnt := range counts {
		out = append(out, UserCountItem{WxId: wxID, Count: cnt, LastAt: lastAt[wxID]})
	}
	sortUserCountList(out, usagestats.ParseSortBy(sortBy))
	return out, nil
}

// ListFeaturesForUser 某用户的功能分布。
func ListFeaturesForUser(ctx context.Context, days int, wxID int64, sortBy string) ([]UserFeatureItem, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 须为正整数")
	}
	counts := make(map[string]int64)
	for _, day := range dayRange(days) {
		all, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageDayWxKey(day, wxID))
		if err != nil {
			return nil, err
		}
		for k, v := range all {
			n, _ := strconv.ParseInt(v, 10, 64)
			counts[k] += n
		}
	}
	lastMap := make(map[string]int64)
	allLast, err := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageLastWxKey(wxID))
	if err != nil {
		return nil, err
	}
	for k, v := range allLast {
		n, _ := strconv.ParseInt(v, 10, 64)
		lastMap[k] = n
	}
	out := make([]UserFeatureItem, 0, len(counts))
	descMap, err2 := featCache.HashGetAll(ctx, cachekit.GatewayFeatUsageDescGlobalKey())
	if err2 != nil {
		return nil, err2
	}
	for id, cnt := range counts {
		out = append(out, UserFeatureItem{
			FeatureId: id, Description: descMap[id], Count: cnt, LastAt: lastMap[id],
		})
	}
	sortUserFeatureList(out, usagestats.ParseSortBy(sortBy))
	return out, nil
}

// ListTimeline 读取用户时间线（最新在前）；limit≤0 默认 200，最大 1000。
func ListTimeline(ctx context.Context, wxID int64, limit int64) ([]TimelineItem, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 须为正整数")
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	raw, err := featCache.ListRange(ctx, cachekit.GatewayFeatUsageTimelineKey(wxID), 0, limit-1)
	if err != nil {
		return nil, err
	}
	out := make([]TimelineItem, 0, len(raw))
	for _, s := range raw {
		at, fid, desc, ok := parseTimelineMember(s)
		if !ok {
			continue
		}
		out = append(out, TimelineItem{FeatureId: fid, Description: desc, At: at})
	}
	return out, nil
}

// parseTimelineMember 解析 unix|featureId 或 unix|featureId|description。
func parseTimelineMember(s string) (at int64, featureID, description string, ok bool) {
	i := strings.Index(s, timelineSep)
	if i <= 0 {
		return 0, "", "", false
	}
	n, err := strconv.ParseInt(s[:i], 10, 64)
	if err != nil {
		return 0, "", "", false
	}
	rest := s[i+1:]
	j := strings.Index(rest, timelineSep)
	if j < 0 {
		fid := strings.TrimSpace(rest)
		if fid == "" {
			return 0, "", "", false
		}
		return n, fid, "", true
	}
	fid := strings.TrimSpace(rest[:j])
	if fid == "" {
		return 0, "", "", false
	}
	return n, fid, strings.TrimSpace(rest[j+1:]), true
}

func sortFeatureList(list []FeatureListItem, sortBy string) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			swap := false
			if sortBy == usagestats.SortByLastAt {
				swap = list[j].LastAt > list[i].LastAt ||
					(list[j].LastAt == list[i].LastAt && list[j].Count > list[i].Count) ||
					(list[j].LastAt == list[i].LastAt && list[j].Count == list[i].Count && list[j].FeatureId < list[i].FeatureId)
			} else {
				swap = list[j].Count > list[i].Count ||
					(list[j].Count == list[i].Count && list[j].FeatureId < list[i].FeatureId)
			}
			if swap {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortUserCountList(list []UserCountItem, sortBy string) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			swap := false
			if sortBy == usagestats.SortByLastAt {
				swap = list[j].LastAt > list[i].LastAt ||
					(list[j].LastAt == list[i].LastAt && list[j].Count > list[i].Count)
			} else {
				swap = list[j].Count > list[i].Count ||
					(list[j].Count == list[i].Count && list[j].WxId < list[i].WxId)
			}
			if swap {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortUserFeatureList(list []UserFeatureItem, sortBy string) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			swap := false
			if sortBy == usagestats.SortByLastAt {
				swap = list[j].LastAt > list[i].LastAt ||
					(list[j].LastAt == list[i].LastAt && list[j].Count > list[i].Count)
			} else {
				swap = list[j].Count > list[i].Count ||
					(list[j].Count == list[i].Count && list[j].FeatureId < list[i].FeatureId)
			}
			if swap {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}
