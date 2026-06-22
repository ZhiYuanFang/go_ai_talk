package history

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"
	"hello/internal/shared/eventlogo"

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
		cache: cachekit.Default(),
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

func (r *cacheRepo) delHistoryListKeyBestEffort(ctx context.Context, deviceNo string) {
	key, err := cachekit.HistoryListKey(deviceNo)
	if err != nil {
		return
	}
	if err := r.cache.Del(ctx, key); err != nil {
		glog.Warningf(ctx, "history cache del list key failed: deviceNo=%s err=%v", deviceNo, err)
	}
}

func (r *cacheRepo) patchHistoryOnAdd(ctx context.Context, item entity.History) {
	items, ok, err := r.getHistoryList(ctx, item.DeviceNo)
	if err == nil && ok {
		items = append([]entity.History{item}, items...)
		if setErr := r.setHistoryList(ctx, item.DeviceNo, items); setErr != nil {
			glog.Warningf(ctx, "history cache add patch failed: deviceNo=%s err=%v", item.DeviceNo, setErr)
			r.delHistoryListKeyBestEffort(ctx, item.DeviceNo)
		}
	}
	// 列表 cache miss 时仍更新 latest，避免冷缓存下 GetLatestHistory 长期 miss。
	_ = r.setLatestHistory(ctx, item.DeviceNo, item)
}

func (r *cacheRepo) patchHistoryOnUpdate(ctx context.Context, item entity.History) {
	items, ok, err := r.getHistoryList(ctx, item.DeviceNo)
	if err == nil && ok {
		for i := range items {
			if items[i].Id == item.Id {
				items[i] = item
				break
			}
		}
		if setErr := r.setHistoryList(ctx, item.DeviceNo, items); setErr != nil {
			glog.Warningf(ctx, "history cache update patch failed: deviceNo=%s err=%v", item.DeviceNo, setErr)
			r.delHistoryListKeyBestEffort(ctx, item.DeviceNo)
		}
	}
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
	if setErr := r.setHistoryList(ctx, deviceNo, next); setErr != nil {
		glog.Warningf(ctx, "history cache delete patch failed: deviceNo=%s err=%v", deviceNo, setErr)
		r.delHistoryListKeyBestEffort(ctx, deviceNo)
		return
	}
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
	// 与 device 共用 Redis 键：仅持久化 objectKey，CDN 映射在 HTTP 边界完成。
	stored := eventlogo.MapEventsLogoStored(ctx, items)
	body, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), historyEventCacheTTL)
}

type birthdayCache struct {
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

func (r *cacheRepo) getBirthday(ctx context.Context, deviceNo string) (string, int64, int, bool, error) {
	key, err := cachekit.UserProfileKey(deviceNo)
	if err != nil {
		return "", 0, 0, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		if err != nil {
			glog.Warningf(ctx, "history cache read failed type=birthday deviceNo=%s err=%v", deviceNo, err)
		}
		return "", 0, 0, false, err
	}
	var out birthdayCache
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return "", 0, 0, false, uErr
	}
	return strings.TrimSpace(out.BabyName), out.Birthday, out.Sex, true, nil
}

func (r *cacheRepo) setBirthday(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error {
	key, err := cachekit.UserProfileKey(deviceNo)
	if err != nil {
		return err
	}
	body, err := json.Marshal(birthdayCache{
		BabyName: strings.TrimSpace(babyName),
		Birthday: birthdayUnixSec,
		Sex:      sex,
	})
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), historyBirthdayCacheTTL)
}
