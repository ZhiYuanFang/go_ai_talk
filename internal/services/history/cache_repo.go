package history

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	historyListCacheTTL     = 60 * time.Second
	historyLatestCacheTTL   = 30 * time.Second
	historyEventCacheTTL    = 10 * time.Minute
	historyBirthdayCacheTTL = 10 * time.Minute
)

type cacheRepo struct {
	cache cachekit.Cache
}

func newCacheRepo() *cacheRepo {
	return &cacheRepo{
		cache: cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{}),
	}
}

func (r *cacheRepo) getHistoryList(ctx context.Context, deviceNo string) ([]entity.History, bool, error) {
	key, err := cachekit.HistoryListKey(deviceNo)
	if err != nil {
		return nil, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		if err != nil {
			glog.Warningf(ctx, "history cache read failed type=list deviceNo=%s err=%v", deviceNo, err)
		}
		return nil, false, err
	}
	out := make([]entity.History, 0)
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return nil, false, uErr
	}
	return out, true, nil
}

func (r *cacheRepo) setHistoryList(ctx context.Context, deviceNo string, items []entity.History) error {
	key, err := cachekit.HistoryListKey(deviceNo)
	if err != nil {
		return err
	}
	body, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), historyListCacheTTL)
}

func (r *cacheRepo) getLatestHistory(ctx context.Context, deviceNo string) (entity.History, bool, error) {
	key, err := cachekit.HistoryLatestKey(deviceNo)
	if err != nil {
		return entity.History{}, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		if err != nil {
			glog.Warningf(ctx, "history cache read failed type=latest deviceNo=%s err=%v", deviceNo, err)
		}
		return entity.History{}, false, err
	}
	var out entity.History
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return entity.History{}, false, uErr
	}
	return out, true, nil
}

func (r *cacheRepo) setLatestHistory(ctx context.Context, deviceNo string, item entity.History) error {
	key, err := cachekit.HistoryLatestKey(deviceNo)
	if err != nil {
		return err
	}
	body, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), historyLatestCacheTTL)
}

func (r *cacheRepo) patchHistoryOnAdd(ctx context.Context, item entity.History) {
	items, ok, err := r.getHistoryList(ctx, item.DeviceNo)
	if err != nil || !ok {
		return
	}
	items = append([]entity.History{item}, items...)
	if err = r.setHistoryList(ctx, item.DeviceNo, items); err != nil {
		glog.Warningf(ctx, "history cache add patch failed: deviceNo=%s err=%v", item.DeviceNo, err)
	}
	_ = r.setLatestHistory(ctx, item.DeviceNo, item)
}

func (r *cacheRepo) patchHistoryOnUpdate(ctx context.Context, item entity.History) {
	items, ok, err := r.getHistoryList(ctx, item.DeviceNo)
	if err != nil || !ok {
		return
	}
	for i := range items {
		if items[i].Id == item.Id {
			items[i] = item
			break
		}
	}
	_ = r.setHistoryList(ctx, item.DeviceNo, items)
	latest, ok, err := r.getLatestHistory(ctx, item.DeviceNo)
	if err == nil && ok && latest.Id == item.Id {
		_ = r.setLatestHistory(ctx, item.DeviceNo, item)
	}
}

func (r *cacheRepo) patchHistoryOnDelete(ctx context.Context, deviceNo string, id int64) {
	items, ok, err := r.getHistoryList(ctx, deviceNo)
	if err != nil || !ok {
		return
	}
	next := make([]entity.History, 0, len(items))
	for _, item := range items {
		if item.Id != id {
			next = append(next, item)
		}
	}
	_ = r.setHistoryList(ctx, deviceNo, next)
	if len(next) > 0 {
		_ = r.setLatestHistory(ctx, deviceNo, next[0])
	}
}

func (r *cacheRepo) currentVersion(ctx context.Context, deviceNo string) int64 {
	key, err := cachekit.HistoryVersionKey(deviceNo)
	if err != nil {
		return 0
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok || strings.TrimSpace(raw) == "" {
		return 0
	}
	var v int64
	_, _ = fmt.Sscanf(raw, "%d", &v)
	return v
}

func (r *cacheRepo) setVersion(ctx context.Context, deviceNo string, version int64) {
	if version <= 0 {
		return
	}
	key, err := cachekit.HistoryVersionKey(deviceNo)
	if err != nil {
		return
	}
	_ = r.cache.SetEX(ctx, key, fmt.Sprintf("%d", version), 24*time.Hour)
}

func (r *cacheRepo) getEventOptions(ctx context.Context) ([]entity.Event, bool, error) {
	key, err := cachekit.EventOptionsKey()
	if err != nil {
		return nil, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		if err != nil {
			glog.Warningf(ctx, "history cache read failed type=event_options err=%v", err)
		}
		return nil, false, err
	}
	out := make([]entity.Event, 0)
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return nil, false, uErr
	}
	return out, true, nil
}

func (r *cacheRepo) setEventOptions(ctx context.Context, items []entity.Event) error {
	key, err := cachekit.EventOptionsKey()
	if err != nil {
		return err
	}
	body, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), historyEventCacheTTL)
}

type birthdayCache struct {
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
	Nickname string `json:"nickname"`
}

func (r *cacheRepo) getBirthday(ctx context.Context, deviceNo string) (string, int64, int, string, bool, error) {
	key, err := cachekit.UserProfileKey(deviceNo)
	if err != nil {
		return "", 0, 0, "", false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		if err != nil {
			glog.Warningf(ctx, "history cache read failed type=birthday deviceNo=%s err=%v", deviceNo, err)
		}
		return "", 0, 0, "", false, err
	}
	var out birthdayCache
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return "", 0, 0, "", false, uErr
	}
	return strings.TrimSpace(out.BabyName), out.Birthday, out.Sex, strings.TrimSpace(out.Nickname), true, nil
}

func (r *cacheRepo) setBirthday(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int, nickname string) error {
	key, err := cachekit.UserProfileKey(deviceNo)
	if err != nil {
		return err
	}
	body, err := json.Marshal(birthdayCache{
		BabyName: strings.TrimSpace(babyName),
		Birthday: birthdayUnixSec,
		Sex:      sex,
		Nickname: strings.TrimSpace(nickname),
	})
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), historyBirthdayCacheTTL)
}

type historyProjectionEvent struct {
	EventID    string `json:"event_id"`
	Version    int64  `json:"version"`
	HistoryID  int64  `json:"history_id"`
	DeviceNo   string `json:"device_no"`
	EventIDRef int64  `json:"event_id_ref"`
	EventName  string `json:"event_name"`
	EventNum   int64  `json:"event_number"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	Remark     string `json:"remark"`
}

// ApplyProjection 处理 history.record.* 异步缓存投影。
func ApplyProjection(ctx context.Context, routingKey string, payload string) error {
	parsed, ok := eventkit.ParseRoutingKey(routingKey)
	if !ok {
		return nil
	}
	if !strings.HasPrefix(parsed.String(), "history.record.") {
		return nil
	}
	var evt historyProjectionEvent
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		return err
	}
	evt.DeviceNo = strings.TrimSpace(evt.DeviceNo)
	if evt.DeviceNo == "" {
		return nil
	}
	// 版本乱序保护：低版本事件直接跳过。
	current := historyCache.currentVersion(ctx, evt.DeviceNo)
	if evt.Version > 0 && evt.Version < current {
		glog.Warningf(ctx, "history cache projection skipped by stale version: deviceNo=%s current=%d incoming=%d", evt.DeviceNo, current, evt.Version)
		return nil
	}
	switch parsed {
	case eventkit.RoutingHistoryRecordCreated:
		historyCache.patchHistoryOnAdd(ctx, entity.History{
			Id:          evt.HistoryID,
			DeviceNo:    evt.DeviceNo,
			EventId:     evt.EventIDRef,
			EventName:   evt.EventName,
			EventNumber: evt.EventNum,
			StartTime:   evt.StartTime,
			EndTime:     evt.EndTime,
			Remark:      strings.TrimSpace(evt.Remark),
		})
	case eventkit.RoutingHistoryRecordUpdated:
		historyCache.patchHistoryOnUpdate(ctx, entity.History{
			Id:          evt.HistoryID,
			DeviceNo:    evt.DeviceNo,
			EventId:     evt.EventIDRef,
			EventName:   evt.EventName,
			EventNumber: evt.EventNum,
			StartTime:   evt.StartTime,
			EndTime:     evt.EndTime,
			Remark:      strings.TrimSpace(evt.Remark),
		})
	case eventkit.RoutingHistoryRecordDeleted:
		historyCache.patchHistoryOnDelete(ctx, evt.DeviceNo, evt.HistoryID)
	}
	historyCache.setVersion(ctx, evt.DeviceNo, evt.Version)
	return nil
}
