// Package history 出站 history-service HTTP 契约客户端。
package history

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"
)

type httpDeviceHistoryClient struct {
	historyBase string
	targets     contracts.HTTPTargets
	client      *http.Client
}

func HTTPDeviceHistory() contracts.DeviceHistoryContract {
	historyBase := strings.TrimSpace(os.Getenv("HISTORY_SERVICE_URL"))
	return &httpDeviceHistoryClient{
		historyBase: strings.TrimRight(historyBase, "/"),
		targets:     contracts.ResolveHTTPTargets(),
		client:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *httpDeviceHistoryClient) notReady() error {
	if r.historyBase == "" {
		return fmt.Errorf("history remote adapter not configured: missing HISTORY_SERVICE_URL")
	}
	return nil
}

func (r *httpDeviceHistoryClient) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
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

func (r *httpDeviceHistoryClient) ListHistoryPage(ctx context.Context, deviceNo string, page int, pageSize int) (contracts.HistoryPageResult, error) {
	if err := r.notReady(); err != nil {
		return contracts.HistoryPageResult{}, err
	}
	var resp struct {
		List     []entity.History `json:"list"`
		Total    int              `json:"total"`
		Page     int              `json:"page"`
		PageSize int              `json:"pageSize"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, r.historyBase, t.HistoryListPath(), map[string]string{
		"deviceNo": strings.TrimSpace(deviceNo),
		"page":     strconv.Itoa(page),
		"pageSize": strconv.Itoa(pageSize),
	}, nil, &resp)
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	return contracts.HistoryPageResult{List: resp.List, Total: resp.Total, Page: resp.Page, PageSize: resp.PageSize}, nil
}

func (r *httpDeviceHistoryClient) ListHistoryFilter(ctx context.Context, deviceNo string, eventIds []int64, startTime int64, endTime int64, limit int, remark string, ignoreTimeRange bool) ([]entity.History, error) {
	if err := r.notReady(); err != nil {
		return nil, err
	}
	var resp struct {
		List []entity.History `json:"list"`
	}
	eventIdsStr := ""
	if len(eventIds) > 0 {
		parts := make([]string, 0, len(eventIds))
		for _, id := range eventIds {
			parts = append(parts, strconv.FormatInt(id, 10))
		}
		eventIdsStr = strings.Join(parts, ",")
	}
	query := map[string]string{
		"deviceNo": strings.TrimSpace(deviceNo),
		"eventIds": eventIdsStr,
	}
	// 忽略时间窗时仍可透传 start/end（对端会丢弃）；不传可减小噪音，但为真必须带上开关。
	if ignoreTimeRange {
		query["ignoreTimeRange"] = "true"
	} else {
		if startTime > 0 {
			query["startTime"] = strconv.FormatInt(startTime, 10)
		}
		if endTime > 0 {
			query["endTime"] = strconv.FormatInt(endTime, 10)
		}
	}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	if strings.TrimSpace(remark) != "" {
		query["remark"] = strings.TrimSpace(remark)
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, r.historyBase, t.HistoryFilterPath(), query, nil, &resp)
	return resp.List, err
}

func (r *httpDeviceHistoryClient) ListHistoryPageV2(ctx context.Context, deviceNo string, page int, pageSize int, startTime int64, endTime int64, limit int) (contracts.HistoryPageResult, error) {
	if err := r.notReady(); err != nil {
		return contracts.HistoryPageResult{}, err
	}
	var resp struct {
		List     []entity.History `json:"list"`
		Total    int              `json:"total"`
		Page     int              `json:"page"`
		PageSize int              `json:"pageSize"`
	}
	query := map[string]string{
		"deviceNo": strings.TrimSpace(deviceNo),
		"page":     strconv.Itoa(page),
		"pageSize": strconv.Itoa(pageSize),
	}
	if startTime > 0 {
		query["startTime"] = strconv.FormatInt(startTime, 10)
	}
	if endTime > 0 {
		query["endTime"] = strconv.FormatInt(endTime, 10)
	}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, r.historyBase, t.HistoryListV2Path(), query, nil, &resp)
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	return contracts.HistoryPageResult{List: resp.List, Total: resp.Total, Page: resp.Page, PageSize: resp.PageSize}, nil
}

func (r *httpDeviceHistoryClient) GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error) {
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

func (r *httpDeviceHistoryClient) EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTimeUnixSec int64, remark string) (bool, error) {
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
		"endTime":  endTimeUnixSec,
		"remark":   strings.TrimSpace(remark),
	}, &resp)
	return resp.Updated, err
}

func (r *httpDeviceHistoryClient) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
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

func (r *httpDeviceHistoryClient) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	if err := r.notReady(); err != nil {
		return err
	}
	t := r.targets
	return r.doJSON(ctx, http.MethodPost, t.VoiceBaseURL, t.VoiceSuggestDeletePath(), nil, map[string]interface{}{"id": id, "deviceNo": strings.TrimSpace(deviceNo)}, nil)
}

func (r *httpDeviceHistoryClient) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
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

func (r *httpDeviceHistoryClient) GetBirthday(ctx context.Context, deviceNo string) (string, int64, int, error) {
	if err := r.notReady(); err != nil {
		return "", 0, 0, err
	}
	var resp struct {
		BabyName string `json:"babyName"`
		Birthday int64  `json:"birthday"`
		Sex      int    `json:"sex"`
	}
	t := r.targets
	err := r.doJSON(ctx, http.MethodGet, t.DeviceBaseURL, t.DeviceProfileGetPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp)
	return strings.TrimSpace(resp.BabyName), resp.Birthday, resp.Sex, err
}

func (r *httpDeviceHistoryClient) SaveBirthday(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error {
	if err := r.notReady(); err != nil {
		return err
	}
	t := r.targets
	return r.doJSON(ctx, http.MethodPost, t.DeviceBaseURL, t.DeviceProfileSavePath(), nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"babyName": strings.TrimSpace(babyName),
		"birthday": birthdayUnixSec,
		"sex":      sex,
	}, nil)
}

func (r *httpDeviceHistoryClient) AddHistory(ctx context.Context, item entity.History) (int64, error) {
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
		"startTime":   item.StartTime,
		"endTime":     item.EndTime,
		"remark":      strings.TrimSpace(item.Remark),
	}, &resp)
	return resp.Id, err
}

func (r *httpDeviceHistoryClient) UpdateHistory(ctx context.Context, item entity.History) error {
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
		"startTime":   item.StartTime,
		"endTime":     item.EndTime,
		"remark":      strings.TrimSpace(item.Remark),
		"postId":      item.PostId,
		"mediaType":   item.MediaType,
		"imageKeys":   item.ImageKeys,
		"videoKey":    strings.TrimSpace(item.VideoKey),
	}, nil)
}

func (r *httpDeviceHistoryClient) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
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

func (r *httpDeviceHistoryClient) doJSON(ctx context.Context, method, baseURL, path string, query map[string]string, body interface{}, out interface{}) error {
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
