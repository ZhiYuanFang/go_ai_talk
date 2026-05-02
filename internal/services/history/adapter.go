package history

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

type Contract interface {
	ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error)
	GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error)
	EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTime string) (bool, error)
	ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error)
	DeleteSuggest(ctx context.Context, id int64, deviceNo string) error
	ListEventOptions(ctx context.Context) ([]entity.Event, error)
	GetBirthday(ctx context.Context, deviceNo string) (string, int, error)
	SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error
	AddHistory(ctx context.Context, item entity.History) (int64, error)
	UpdateHistory(ctx context.Context, item entity.History) error
	DeleteHistory(ctx context.Context, id int64, deviceNo string) error
}

const (
	historyServiceModeEnv         = "HISTORY_SERVICE_MODE"
	historyRemoteURLEnv           = "HISTORY_SERVICE_URL"
	historyCanaryPercentEnv       = "HISTORY_SERVICE_CANARY_PERCENT"
	historyRemoteFailoverLocalEnv = "HISTORY_SERVICE_REMOTE_FAILOVER_LOCAL"
	historyModeLocal              = "local"
	historyModeRemote             = "remote"
	historyModeCanary             = "canary"
)

type historyRemoteClient struct {
	historyBase string
	targets     contracts.HTTPTargets
	client      *http.Client
}

func newHistoryRemoteClient() Contract {
	historyBase := strings.TrimSpace(os.Getenv(historyRemoteURLEnv))
	return &historyRemoteClient{
		historyBase: strings.TrimRight(historyBase, "/"),
		targets:     contracts.ResolveHTTPTargets(),
		client:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *historyRemoteClient) notReady() error {
	if r.historyBase == "" {
		return fmt.Errorf("history remote adapter not configured: missing %s", historyRemoteURLEnv)
	}
	return nil
}

func (r *historyRemoteClient) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	if err := r.notReady(); err != nil {
		return nil, err
	}
	var resp struct {
		List []entity.History `json:"list"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, r.historyBase, t.HistoryListPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp)
	return resp.List, err
}

func (r *historyRemoteClient) GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error) {
	if err := r.notReady(); err != nil {
		return entity.History{}, err
	}
	var resp struct {
		Item entity.History `json:"item"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, r.historyBase, t.HistoryEventLatestPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp)
	return resp.Item, err
}

func (r *historyRemoteClient) EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTime string) (bool, error) {
	if err := r.notReady(); err != nil {
		return false, err
	}
	var resp struct {
		Updated bool `json:"updated"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodPost, r.historyBase, t.HistoryEventEndLatestPath(), nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"eventId":  eventID,
		"endTime":  strings.TrimSpace(endTime),
	}, &resp)
	return resp.Updated, err
}

func (r *historyRemoteClient) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	if err := r.notReady(); err != nil {
		return nil, err
	}
	var resp struct {
		List []entity.Suggest `json:"list"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, t.VoiceBaseURL, t.VoiceSuggestListPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp)
	return resp.List, err
}

func (r *historyRemoteClient) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	if err := r.notReady(); err != nil {
		return err
	}
	t := r.targets
	return r.doJSON(ctx, http.MethodPost, t.VoiceBaseURL, t.VoiceSuggestDeletePath(), nil, map[string]interface{}{"id": id, "deviceNo": strings.TrimSpace(deviceNo)}, nil)
}

func (r *historyRemoteClient) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	if err := r.notReady(); err != nil {
		return nil, err
	}
	var resp struct {
		List []entity.Event `json:"list"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, t.DeviceBaseURL, t.DeviceInternalEventOptionsPath(), nil, nil, &resp)
	return resp.List, err
}

func (r *historyRemoteClient) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	if err := r.notReady(); err != nil {
		return "", 0, err
	}
	var resp struct {
		Birthday string `json:"birthday"`
		Sex      int    `json:"sex"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, t.DeviceBaseURL, t.DeviceProfileGetPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp)
	return strings.TrimSpace(resp.Birthday), resp.Sex, err
}

func (r *historyRemoteClient) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	if err := r.notReady(); err != nil {
		return err
	}
	t := r.targets
	return r.doJSON(ctx, http.MethodPost, t.DeviceBaseURL, t.DeviceProfileSavePath(), nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"birthday": strings.TrimSpace(birthday),
		"sex":      sex,
	}, nil)
}

func (r *historyRemoteClient) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	if err := r.notReady(); err != nil {
		return 0, err
	}
	var resp struct {
		Id int64 `json:"id"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodPost, r.historyBase, t.HistoryEventAddPath(), nil, map[string]interface{}{
		"deviceNo":    strings.TrimSpace(item.DeviceNo),
		"eventId":     item.EventId,
		"eventName":   strings.TrimSpace(item.EventName),
		"eventNumber": item.EventNumber,
		"startTime":   strings.TrimSpace(item.StartTime),
		"endTime":     strings.TrimSpace(item.EndTime),
		"remark":      strings.TrimSpace(item.Remark),
	}, &resp)
	return resp.Id, err
}

func (r *historyRemoteClient) UpdateHistory(ctx context.Context, item entity.History) error {
	if err := r.notReady(); err != nil {
		return err
	}
	t := r.targets
	return r.doJSON(ctx, http.MethodPost, r.historyBase, t.HistoryEventUpdatePath(), nil, map[string]interface{}{
		"id":          item.Id,
		"deviceNo":    strings.TrimSpace(item.DeviceNo),
		"eventId":     item.EventId,
		"eventName":   strings.TrimSpace(item.EventName),
		"eventNumber": item.EventNumber,
		"startTime":   strings.TrimSpace(item.StartTime),
		"endTime":     strings.TrimSpace(item.EndTime),
		"remark":      strings.TrimSpace(item.Remark),
	}, nil)
}

func (r *historyRemoteClient) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	if err := r.notReady(); err != nil {
		return err
	}
	t := r.targets
	return r.doJSON(ctx, http.MethodPost, r.historyBase, t.HistoryEventDeletePath(), nil, map[string]interface{}{"id": id, "deviceNo": strings.TrimSpace(deviceNo)}, nil)
}

type responseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (r *historyRemoteClient) doJSON(ctx context.Context, method, baseURL, path string, query map[string]string, body interface{}, out interface{}) error {
	if strings.TrimSpace(baseURL) == "" {
		return fmt.Errorf("remote http base url is empty")
	}
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + path)
	if err != nil {
		return err
	}
	if len(query) > 0 {
		q := reqURL.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				q.Set(k, strings.TrimSpace(v))
			}
		}
		reqURL.RawQuery = q.Encode()
	}
	var bodyReader strings.Reader
	if body != nil {
		raw, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return marshalErr
		}
		bodyReader = *strings.NewReader(string(raw))
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), &bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env responseEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		if strings.TrimSpace(env.Message) == "" {
			return fmt.Errorf("history remote call failed: status=%d", resp.StatusCode)
		}
		return fmt.Errorf("history remote call failed: %s", strings.TrimSpace(env.Message))
	}
	if env.Code != 0 {
		return fmt.Errorf("history remote business failed: %s", strings.TrimSpace(env.Message))
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

type switchConfig struct {
	mode            string
	canaryPercent   int
	failoverToLocal bool
}

type switchAdapter struct {
	local  Contract
	remote Contract
	cfg    switchConfig
}

func loadSwitchConfigFromEnv() switchConfig {
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
	return switchConfig{mode: mode, canaryPercent: canaryPercent, failoverToLocal: failoverToLocal}
}

func (a *switchAdapter) shouldUseRemote(deviceNo string) bool {
	switch a.cfg.mode {
	case historyModeRemote:
		return true
	case historyModeCanary:
		if a.cfg.canaryPercent <= 0 {
			return false
		}
		if a.cfg.canaryPercent >= 100 {
			return true
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(strings.TrimSpace(deviceNo)))
		return int(h.Sum32()%100) < a.cfg.canaryPercent
	default:
		return false
	}
}

func (a *switchAdapter) run(useRemote bool, remoteFn func() error, localFn func() error) error {
	if !useRemote {
		return localFn()
	}
	err := remoteFn()
	if err != nil && a.cfg.failoverToLocal {
		return localFn()
	}
	return err
}

func (a *switchAdapter) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if cached, ok, err := historyCache.getHistoryList(ctx, deviceNo); err == nil && ok {
		glog.Debugf(ctx, "history cache hit type=list deviceNo=%s count=%d", deviceNo, len(cached))
		return cached, nil
	}
	glog.Debugf(ctx, "history cache miss type=list deviceNo=%s", deviceNo)
	if !a.shouldUseRemote(deviceNo) {
		items, err := a.local.ListHistory(ctx, deviceNo)
		if err == nil {
			_ = historyCache.setHistoryList(ctx, deviceNo, items)
			glog.Debugf(ctx, "history cache refill type=list source=local deviceNo=%s count=%d", deviceNo, len(items))
		}
		return items, err
	}
	items, err := a.remote.ListHistory(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		glog.Warningf(ctx, "history cache degrade type=list source=remote->local deviceNo=%s err=%v", deviceNo, err)
		items, err = a.local.ListHistory(ctx, deviceNo)
	}
	if err == nil {
		_ = historyCache.setHistoryList(ctx, deviceNo, items)
		glog.Debugf(ctx, "history cache refill type=list source=remote deviceNo=%s count=%d", deviceNo, len(items))
	}
	return items, err
}

func (a *switchAdapter) GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if item, ok, err := historyCache.getLatestHistory(ctx, deviceNo); err == nil && ok {
		glog.Debugf(ctx, "history cache hit type=latest deviceNo=%s id=%d", deviceNo, item.Id)
		return item, nil
	}
	glog.Debugf(ctx, "history cache miss type=latest deviceNo=%s", deviceNo)
	if !a.shouldUseRemote(deviceNo) {
		item, err := a.local.GetLatestHistory(ctx, deviceNo)
		if err == nil && item.Id > 0 {
			_ = historyCache.setLatestHistory(ctx, deviceNo, item)
			glog.Debugf(ctx, "history cache refill type=latest source=local deviceNo=%s id=%d", deviceNo, item.Id)
		}
		return item, err
	}
	item, err := a.remote.GetLatestHistory(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		glog.Warningf(ctx, "history cache degrade type=latest source=remote->local deviceNo=%s err=%v", deviceNo, err)
		item, err = a.local.GetLatestHistory(ctx, deviceNo)
	}
	if err == nil && item.Id > 0 {
		_ = historyCache.setLatestHistory(ctx, deviceNo, item)
		glog.Debugf(ctx, "history cache refill type=latest source=remote deviceNo=%s id=%d", deviceNo, item.Id)
	}
	return item, err
}

func (a *switchAdapter) EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTime string) (bool, error) {
	if !a.shouldUseRemote(deviceNo) {
		return a.local.EndLatestHistoryIfMatch(ctx, deviceNo, eventID, endTime)
	}
	updated, err := a.remote.EndLatestHistoryIfMatch(ctx, deviceNo, eventID, endTime)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.EndLatestHistoryIfMatch(ctx, deviceNo, eventID, endTime)
	}
	if err == nil && updated {
		item, lErr := a.GetLatestHistory(ctx, deviceNo)
		if lErr == nil && item.Id > 0 {
			item.EndTime = strings.TrimSpace(endTime)
			historyCache.patchHistoryOnUpdate(ctx, item)
		}
	}
	return updated, err
}

func (a *switchAdapter) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	if !a.shouldUseRemote(deviceNo) {
		return a.local.ListSuggest(ctx, deviceNo)
	}
	items, err := a.remote.ListSuggest(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.ListSuggest(ctx, deviceNo)
	}
	return items, err
}

func (a *switchAdapter) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	return a.run(a.shouldUseRemote(deviceNo), func() error { return a.remote.DeleteSuggest(ctx, id, deviceNo) }, func() error {
		return a.local.DeleteSuggest(ctx, id, deviceNo)
	})
}

func (a *switchAdapter) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	if items, ok, err := historyCache.getEventOptions(ctx); err == nil && ok {
		return items, nil
	}
	useRemote := a.cfg.mode == historyModeRemote
	if !useRemote {
		items, err := a.local.ListEventOptions(ctx)
		if err == nil {
			_ = historyCache.setEventOptions(ctx, items)
		}
		return items, err
	}
	items, err := a.remote.ListEventOptions(ctx)
	if err != nil && a.cfg.failoverToLocal {
		items, err = a.local.ListEventOptions(ctx)
	}
	if err == nil {
		_ = historyCache.setEventOptions(ctx, items)
	}
	return items, err
}

func (a *switchAdapter) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	if birthday, sex, ok, err := historyCache.getBirthday(ctx, deviceNo); err == nil && ok {
		return birthday, sex, nil
	}
	if !a.shouldUseRemote(deviceNo) {
		birthday, sex, err := a.local.GetBirthday(ctx, deviceNo)
		if err == nil {
			_ = historyCache.setBirthday(ctx, deviceNo, birthday, sex)
		}
		return birthday, sex, err
	}
	b, sex, err := a.remote.GetBirthday(ctx, deviceNo)
	if err != nil && a.cfg.failoverToLocal {
		b, sex, err = a.local.GetBirthday(ctx, deviceNo)
	}
	if err == nil {
		_ = historyCache.setBirthday(ctx, deviceNo, b, sex)
	}
	return b, sex, err
}

func (a *switchAdapter) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	err := a.run(a.shouldUseRemote(deviceNo), func() error { return a.remote.SaveBirthday(ctx, deviceNo, birthday, sex) }, func() error {
		return a.local.SaveBirthday(ctx, deviceNo, birthday, sex)
	})
	if err == nil {
		_ = historyCache.setBirthday(ctx, deviceNo, birthday, sex)
	}
	return err
}

func (a *switchAdapter) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	useRemote := a.shouldUseRemote(item.DeviceNo)
	if !useRemote {
		return a.local.AddHistory(ctx, item)
	}
	id, err := a.remote.AddHistory(ctx, item)
	if err != nil && a.cfg.failoverToLocal {
		return a.local.AddHistory(ctx, item)
	}
	if err == nil && id > 0 {
		item.Id = id
		historyCache.patchHistoryOnAdd(ctx, item)
	}
	return id, err
}

func (a *switchAdapter) UpdateHistory(ctx context.Context, item entity.History) error {
	err := a.run(a.shouldUseRemote(item.DeviceNo), func() error { return a.remote.UpdateHistory(ctx, item) }, func() error {
		return a.local.UpdateHistory(ctx, item)
	})
	if err == nil {
		historyCache.patchHistoryOnUpdate(ctx, item)
	}
	return err
}

func (a *switchAdapter) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	err := a.run(a.shouldUseRemote(deviceNo), func() error { return a.remote.DeleteHistory(ctx, id, deviceNo) }, func() error {
		return a.local.DeleteHistory(ctx, id, deviceNo)
	})
	if err == nil {
		historyCache.patchHistoryOnDelete(ctx, deviceNo, id)
	}
	return err
}

var (
	once sync.Once
	ins  Contract
)

// DeviceHistory 返回可切换的历史服务适配器。
func DeviceHistory() Contract {
	once.Do(func() {
		ins = &switchAdapter{
			local:  &localService{},
			remote: newHistoryRemoteClient(),
			cfg:    loadSwitchConfigFromEnv(),
		}
	})
	return ins
}
