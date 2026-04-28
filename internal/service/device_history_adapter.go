package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"sync"

	"hello/internal/model/entity"
)

const (
	historyServiceModeEnv         = "HISTORY_SERVICE_MODE" // local | remote | canary
	historyRemoteURLEnv           = "HISTORY_SERVICE_URL"
	historyCanaryPercentEnv       = "HISTORY_SERVICE_CANARY_PERCENT"
	historyRemoteFailoverLocalEnv = "HISTORY_SERVICE_REMOTE_FAILOVER_LOCAL"

	historyModeLocal  = "local"
	historyModeRemote = "remote"
	historyModeCanary = "canary"
)

type historyRemoteClient struct {
	baseURL string
}

var _ DeviceHistoryContract = (*historyRemoteClient)(nil)

func newHistoryRemoteClient() DeviceHistoryContract {
	baseURL := strings.TrimSpace(os.Getenv(historyRemoteURLEnv))
	return &historyRemoteClient{baseURL: baseURL}
}

func (r *historyRemoteClient) notReady() error {
	if r.baseURL == "" {
		return fmt.Errorf("history remote adapter not configured: missing %s", historyRemoteURLEnv)
	}
	return fmt.Errorf("history remote adapter not implemented yet")
}

func (r *historyRemoteClient) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	return nil, r.notReady()
}

func (r *historyRemoteClient) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	return nil, r.notReady()
}

func (r *historyRemoteClient) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	return r.notReady()
}

func (r *historyRemoteClient) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	return nil, r.notReady()
}

func (r *historyRemoteClient) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	return "", 0, r.notReady()
}

func (r *historyRemoteClient) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	return r.notReady()
}

func (r *historyRemoteClient) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	return 0, r.notReady()
}

func (r *historyRemoteClient) UpdateHistory(ctx context.Context, item entity.History) error {
	return r.notReady()
}

func (r *historyRemoteClient) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	return r.notReady()
}

type historySwitchConfig struct {
	mode            string
	canaryPercent   int
	failoverToLocal bool
}

type historySwitchAdapter struct {
	local  DeviceHistoryContract
	remote DeviceHistoryContract
	cfg    historySwitchConfig
}

func loadHistorySwitchConfigFromEnv() historySwitchConfig {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(historyServiceModeEnv)))
	switch mode {
	case historyModeRemote, historyModeCanary:
	default:
		mode = historyModeLocal
	}

	canaryPercent, err := strconv.Atoi(strings.TrimSpace(os.Getenv(historyCanaryPercentEnv)))
	if err != nil {
		canaryPercent = 0
	}
	if canaryPercent < 0 {
		canaryPercent = 0
	}
	if canaryPercent > 100 {
		canaryPercent = 100
	}

	failoverRaw := strings.ToLower(strings.TrimSpace(os.Getenv(historyRemoteFailoverLocalEnv)))
	failoverToLocal := failoverRaw == "" || failoverRaw == "1" || failoverRaw == "true" || failoverRaw == "yes"

	return historySwitchConfig{
		mode:            mode,
		canaryPercent:   canaryPercent,
		failoverToLocal: failoverToLocal,
	}
}

func newHistorySwitchAdapter(local, remote DeviceHistoryContract, cfg historySwitchConfig) DeviceHistoryContract {
	return &historySwitchAdapter{
		local:  local,
		remote: remote,
		cfg:    cfg,
	}
}

func (a *historySwitchAdapter) shouldUseRemote(deviceNo string) bool {
	switch a.cfg.mode {
	case historyModeRemote:
		return true
	case historyModeCanary:
		// 金丝雀发布，根据配置百分比决定是否使用远程服务。
		if a.cfg.canaryPercent <= 0 {
			return false
		}
		if a.cfg.canaryPercent >= 100 {
			return true
		}
		// 金丝雀发布，根据设备号哈希值决定是否使用远程服务。
		h := fnv.New32a()
		_, _ = h.Write([]byte(strings.TrimSpace(deviceNo)))
		return int(h.Sum32()%100) < a.cfg.canaryPercent
	default:
		return false
	}
}

func (a *historySwitchAdapter) run(useRemote bool, remoteFn func() error, localFn func() error) error {
	if !useRemote {
		return localFn()
	}
	err := remoteFn()
	if err != nil && a.cfg.failoverToLocal {
		return localFn()
	}
	return err
}

func (a *historySwitchAdapter) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	if !a.shouldUseRemote(deviceNo) {
		return a.local.ListHistory(ctx, deviceNo)
	}
	items, err := a.remote.ListHistory(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.ListHistory(ctx, deviceNo)
	}
	return items, err
}

func (a *historySwitchAdapter) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	if !a.shouldUseRemote(deviceNo) {
		return a.local.ListSuggest(ctx, deviceNo)
	}
	items, err := a.remote.ListSuggest(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.ListSuggest(ctx, deviceNo)
	}
	return items, err
}

func (a *historySwitchAdapter) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	return a.run(
		a.shouldUseRemote(deviceNo),
		func() error { return a.remote.DeleteSuggest(ctx, id, deviceNo) },
		func() error { return a.local.DeleteSuggest(ctx, id, deviceNo) },
	)
}

func (a *historySwitchAdapter) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	useRemote := a.cfg.mode == historyModeRemote
	if !useRemote {
		return a.local.ListEventOptions(ctx)
	}
	items, err := a.remote.ListEventOptions(ctx)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.ListEventOptions(ctx)
	}
	return items, err
}

func (a *historySwitchAdapter) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	if !a.shouldUseRemote(deviceNo) {
		return a.local.GetBirthday(ctx, deviceNo)
	}
	b, sex, err := a.remote.GetBirthday(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.GetBirthday(ctx, deviceNo)
	}
	return b, sex, err
}

func (a *historySwitchAdapter) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	return a.run(
		a.shouldUseRemote(deviceNo),
		func() error { return a.remote.SaveBirthday(ctx, deviceNo, birthday, sex) },
		func() error { return a.local.SaveBirthday(ctx, deviceNo, birthday, sex) },
	)
}

func (a *historySwitchAdapter) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	useRemote := a.shouldUseRemote(item.DeviceNo)
	if !useRemote {
		return a.local.AddHistory(ctx, item)
	}
	id, err := a.remote.AddHistory(ctx, item)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.AddHistory(ctx, item)
	}
	return id, err
}

func (a *historySwitchAdapter) UpdateHistory(ctx context.Context, item entity.History) error {
	return a.run(
		a.shouldUseRemote(item.DeviceNo),
		func() error { return a.remote.UpdateHistory(ctx, item) },
		func() error { return a.local.UpdateHistory(ctx, item) },
	)
}

func (a *historySwitchAdapter) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	return a.run(
		a.shouldUseRemote(deviceNo),
		func() error { return a.remote.DeleteHistory(ctx, id, deviceNo) },
		func() error { return a.local.DeleteHistory(ctx, id, deviceNo) },
	)
}

var (
	historyAdapterOnce sync.Once
	historyAdapterIns  DeviceHistoryContract
)

// DeviceHistory 返回可切换的历史服务适配器；默认本地实现，支持后续远程化。
func DeviceHistory() DeviceHistoryContract {
	historyAdapterOnce.Do(func() {
		historyAdapterIns = newHistorySwitchAdapter(
			deviceHistoryLocal(),
			newHistoryRemoteClient(),
			loadHistorySwitchConfigFromEnv(),
		)
	})
	return historyAdapterIns
}
