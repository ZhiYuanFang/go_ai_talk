package device

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"
	"hello/internal/dao"
)

const (
	eventOptionsTTL = 10 * time.Minute
	actionOptionsTTL = 10 * time.Minute
	userProfileTTL = 10 * time.Minute
)

type deviceCacheRepo struct {
	cache cachekit.Cache
}

func newDeviceCacheRepo() *deviceCacheRepo {
	return &deviceCacheRepo{
		cache: cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{}),
	}
}

var deviceCache = newDeviceCacheRepo()

func (r *deviceCacheRepo) currentVersion(ctx context.Context, key string) int64 {
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok || strings.TrimSpace(raw) == "" {
		return 0
	}
	var v int64
	_, _ = fmt.Sscanf(raw, "%d", &v)
	return v
}

func (r *deviceCacheRepo) setVersion(ctx context.Context, key string, version int64) {
	if version <= 0 {
		return
	}
	_ = r.cache.SetEX(ctx, key, fmt.Sprintf("%d", version), 24*time.Hour)
}

func (r *deviceCacheRepo) getEventOptions(ctx context.Context) ([]entity.Event, bool, error) {
	key, err := cachekit.EventOptionsKey()
	if err != nil {
		return nil, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}
	out := make([]entity.Event, 0)
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return nil, false, uErr
	}
	return out, true, nil
}

func (r *deviceCacheRepo) setEventOptions(ctx context.Context, items []entity.Event) error {
	key, err := cachekit.EventOptionsKey()
	if err != nil {
		return err
	}
	body, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), eventOptionsTTL)
}

func (r *deviceCacheRepo) getActionOptions(ctx context.Context) ([]entity.Action, bool, error) {
	key, err := cachekit.ActionOptionsKey()
	if err != nil {
		return nil, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}
	out := make([]entity.Action, 0)
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return nil, false, uErr
	}
	return out, true, nil
}

func (r *deviceCacheRepo) setActionOptions(ctx context.Context, items []entity.Action) error {
	key, err := cachekit.ActionOptionsKey()
	if err != nil {
		return err
	}
	body, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), actionOptionsTTL)
}

type cachedUserProfile struct {
	DeviceNo string `json:"deviceNo"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

func (r *deviceCacheRepo) getUserProfile(ctx context.Context, deviceNo string) (cachedUserProfile, bool, error) {
	key, err := cachekit.UserProfileKey(deviceNo)
	if err != nil {
		return cachedUserProfile{}, false, err
	}
	raw, ok, err := r.cache.Get(ctx, key)
	if err != nil || !ok {
		return cachedUserProfile{}, false, err
	}
	var out cachedUserProfile
	if uErr := json.Unmarshal([]byte(raw), &out); uErr != nil {
		return cachedUserProfile{}, false, uErr
	}
	out.DeviceNo = strings.TrimSpace(out.DeviceNo)
	return out, true, nil
}

func (r *deviceCacheRepo) setUserProfile(ctx context.Context, profile cachedUserProfile) error {
	key, err := cachekit.UserProfileKey(profile.DeviceNo)
	if err != nil {
		return err
	}
	body, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return r.cache.SetEX(ctx, key, string(body), userProfileTTL)
}

type deviceProjectionEvent struct {
	EventID  string `json:"event_id"`
	Version  int64  `json:"version"`
	DeviceNo string `json:"device_no"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

// ApplyProjection 异步应用 device 侧缓存投影事件。
func ApplyProjection(ctx context.Context, routingKey, payload string) error {
	var evt deviceProjectionEvent
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		return err
	}
	parsed, ok := eventkit.ParseRoutingKey(routingKey)
	if !ok {
		return nil
	}
	switch parsed {
	case eventkit.RoutingDeviceEventChanged:
		verKey, _ := cachekit.DeviceEventVersionKey()
		if evt.Version > 0 && evt.Version < deviceCache.currentVersion(ctx, verKey) {
			return nil
		}
		rows := make([]entity.Event, 0)
		if err := dao.Event.Ctx(ctx).Fields(
			dao.Event.Columns().Id,
			dao.Event.Columns().Name,
			dao.Event.Columns().NeedQuantity,
			dao.Event.Columns().ExtraNames,
		).OrderAsc(dao.Event.Columns().Id).Scan(&rows); err != nil {
			return err
		}
		if err := deviceCache.setEventOptions(ctx, rows); err != nil {
			return err
		}
		deviceCache.setVersion(ctx, verKey, evt.Version)
	case eventkit.RoutingDeviceActionChanged:
		verKey, _ := cachekit.DeviceActionVersionKey()
		if evt.Version > 0 && evt.Version < deviceCache.currentVersion(ctx, verKey) {
			return nil
		}
		rows := make([]entity.Action, 0)
		if err := dao.Action.Ctx(ctx).Fields(
			dao.Action.Columns().Id,
			dao.Action.Columns().Name,
			dao.Action.Columns().TargetType,
		).OrderAsc(dao.Action.Columns().Id).Scan(&rows); err != nil {
			return err
		}
		if err := deviceCache.setActionOptions(ctx, rows); err != nil {
			return err
		}
		deviceCache.setVersion(ctx, verKey, evt.Version)
	case eventkit.RoutingDeviceUserProfileUpdated:
		evt.DeviceNo = strings.TrimSpace(evt.DeviceNo)
		if evt.DeviceNo == "" {
			return nil
		}
		verKey, _ := cachekit.UserProfileVersionKey(evt.DeviceNo)
		if evt.Version > 0 && evt.Version < deviceCache.currentVersion(ctx, verKey) {
			return nil
		}
		if err := deviceCache.setUserProfile(ctx, cachedUserProfile{
			DeviceNo: evt.DeviceNo,
			Birthday: evt.Birthday,
			Sex:      evt.Sex,
		}); err != nil {
			return err
		}
		deviceCache.setVersion(ctx, verKey, evt.Version)
	}
	return nil
}
