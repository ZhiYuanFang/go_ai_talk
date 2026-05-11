package history

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/eventkit"
	contracts "hello/internal/services/contracts"
	"hello/internal/services/workeroutbox"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

type localService struct{}

var historyCache = newCacheRepo()

func (s *localService) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	return ListDeviceHistory(ctx, deviceNo)
}

func (s *localService) ListHistoryPage(ctx context.Context, deviceNo string, page int, pageSize int) (contracts.HistoryPageResult, error) {
	return ListDeviceHistoryPage(ctx, deviceNo, page, pageSize)
}

func (s *localService) GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error) {
	return GetLatestDeviceHistory(ctx, deviceNo)
}

func (s *localService) EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTimeUnixSec int64) (bool, error) {
	return EndLatestDeviceHistoryIfMatch(ctx, deviceNo, eventID, endTimeUnixSec)
}

func (s *localService) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	return ListDeviceSuggest(ctx, deviceNo)
}

func (s *localService) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	return DeleteDeviceSuggest(ctx, id, deviceNo)
}

func (s *localService) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	return ListEventOptions(ctx)
}

func (s *localService) GetBirthday(ctx context.Context, deviceNo string) (int64, int, error) {
	return GetDeviceBirthday(ctx, deviceNo)
}

func (s *localService) SaveBirthday(ctx context.Context, deviceNo string, birthdayUnixSec int64, sex int) error {
	return SaveBirthday(ctx, deviceNo, birthdayUnixSec, sex)
}

func (s *localService) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	return AddDeviceHistory(ctx, item)
}

func (s *localService) UpdateHistory(ctx context.Context, item entity.History) error {
	return UpdateDeviceHistory(ctx, item)
}

func (s *localService) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	return DeleteDeviceHistory(ctx, id, deviceNo)
}

func ListDeviceHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return []entity.History{}, nil
	}
	if cached, ok, err := historyCache.getHistoryList(ctx, deviceNo); err == nil && ok {
		return cached, nil
	}
	rows, err := dao.History.Ctx(ctx).
		Fields(
			dao.History.Columns().Id,
			dao.History.Columns().DeviceNo,
			dao.History.Columns().EventId,
			dao.History.Columns().EventName,
			dao.History.Columns().EventNumber,
			dao.History.Columns().StartTime,
			dao.History.Columns().EndTime,
			dao.History.Columns().Remark,
		).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.History.Columns().Id).
		All()
	if err != nil {
		return nil, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.History{
			Id:          row[dao.History.Columns().Id].Int64(),
			DeviceNo:    row[dao.History.Columns().DeviceNo].String(),
			EventId:     row[dao.History.Columns().EventId].Int64(),
			EventName:   row[dao.History.Columns().EventName].String(),
			EventNumber: row[dao.History.Columns().EventNumber].Int64(),
			StartTime:   row[dao.History.Columns().StartTime].Int64(),
			EndTime:     row[dao.History.Columns().EndTime].Int64(),
			Remark:      row[dao.History.Columns().Remark].String(),
		})
	}
	_ = historyCache.setHistoryList(ctx, deviceNo, items)
	return items, nil
}

// ListDeviceHistoryPage 分页查询设备历史记录。
// 外部列表页只需要当前页数据与总数，避免复用全量缓存返回过多数据。
func ListDeviceHistoryPage(ctx context.Context, deviceNo string, page int, pageSize int) (contracts.HistoryPageResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if deviceNo == "" {
		return contracts.HistoryPageResult{List: []entity.History{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	total, err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Count()
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	if total == 0 {
		return contracts.HistoryPageResult{List: []entity.History{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	rows, err := dao.History.Ctx(ctx).
		Fields(
			dao.History.Columns().Id,
			dao.History.Columns().DeviceNo,
			dao.History.Columns().EventId,
			dao.History.Columns().EventName,
			dao.History.Columns().EventNumber,
			dao.History.Columns().StartTime,
			dao.History.Columns().EndTime,
			dao.History.Columns().Remark,
		).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.History.Columns().Id).
		Page(page, pageSize).
		All()
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.History{
			Id:          row[dao.History.Columns().Id].Int64(),
			DeviceNo:    row[dao.History.Columns().DeviceNo].String(),
			EventId:     row[dao.History.Columns().EventId].Int64(),
			EventName:   row[dao.History.Columns().EventName].String(),
			EventNumber: row[dao.History.Columns().EventNumber].Int64(),
			StartTime:   row[dao.History.Columns().StartTime].Int64(),
			EndTime:     row[dao.History.Columns().EndTime].Int64(),
			Remark:      row[dao.History.Columns().Remark].String(),
		})
	}
	return contracts.HistoryPageResult{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func GetLatestDeviceHistory(ctx context.Context, deviceNo string) (entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return entity.History{}, nil
	}
	if item, ok, err := historyCache.getLatestHistory(ctx, deviceNo); err == nil && ok {
		return item, nil
	}
	var item entity.History
	err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.History.Columns().Id).
		Limit(1).
		Scan(&item)
	if err != nil {
		return entity.History{}, err
	}
	if item.Id > 0 {
		_ = historyCache.setLatestHistory(ctx, deviceNo, item)
	}
	return item, nil
}

func EndLatestDeviceHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTimeUnixSec int64) (bool, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || eventID <= 0 {
		return false, nil
	}
	// 0 表示调用方未显式传结束时间，落库为当前 Unix 秒。
	if endTimeUnixSec == 0 {
		endTimeUnixSec = time.Now().Unix()
	}
	last, err := GetLatestDeviceHistory(ctx, deviceNo)
	if err != nil {
		return false, err
	}
	if last.Id <= 0 || last.EventId != eventID {
		return false, nil
	}
	_, err = dao.History.Ctx(ctx).
		Where(dao.History.Columns().Id, last.Id).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Data(g.Map{dao.History.Columns().EndTime: endTimeUnixSec}).
		Update()
	if err != nil {
		return false, err
	}
	last.EndTime = endTimeUnixSec
	historyCache.patchHistoryOnUpdate(ctx, last)
	return true, nil
}

func ListDeviceSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return []entity.Suggest{}, nil
	}
	items, err := delegateListSuggest(ctx, deviceNo)
	logDelegateFailure(ctx, "list_suggest", err)
	return items, err
}

func DeleteDeviceSuggest(ctx context.Context, id int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if id <= 0 || deviceNo == "" {
		return fmt.Errorf("参数无效")
	}
	err := delegateDeleteSuggest(ctx, id, deviceNo)
	logDelegateFailure(ctx, "delete_suggest", err)
	return err
}

func ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	if cached, ok, err := historyCache.getEventOptions(ctx); err == nil && ok {
		return cached, nil
	}
	rows, err := delegateListEventOptions(ctx)
	if err != nil {
		return nil, err
	}
	_ = historyCache.setEventOptions(ctx, rows)
	return rows, nil
}

func SaveBirthday(ctx context.Context, deviceNo string, birthdayUnixSec int64, sex int) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil
	}
	if sex > 0 {
		sex = 1
	}
	err := delegateSaveProfile(ctx, deviceNo, birthdayUnixSec, sex)
	if err != nil {
		logDelegateFailure(ctx, "save_profile", err)
		return err
	}
	_ = historyCache.setBirthday(ctx, deviceNo, birthdayUnixSec, sex)
	return nil
}

func GetDeviceBirthday(ctx context.Context, deviceNo string) (int64, int, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, 0, nil
	}
	if birthday, sex, ok, err := historyCache.getBirthday(ctx, deviceNo); err == nil && ok {
		return birthday, sex, nil
	}
	birthday, sex, err := delegateGetProfile(ctx, deviceNo)
	if err != nil {
		logDelegateFailure(ctx, "get_profile", err)
		return 0, 0, err
	}
	if sex > 0 {
		sex = 1
	}
	_ = historyCache.setBirthday(ctx, deviceNo, birthday, sex)
	return birthday, sex, nil
}

func AddDeviceHistory(ctx context.Context, item entity.History) (int64, error) {
	item.DeviceNo = strings.TrimSpace(item.DeviceNo)
	item.EventName = strings.TrimSpace(item.EventName)
	item.Remark = strings.TrimSpace(item.Remark)
	if item.DeviceNo == "" {
		return 0, fmt.Errorf("deviceNo 不能为空")
	}
	var id int64
	err := g.DB(dao.History.Group()).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		res, err := tx.Model(dao.History.Table()).Data(g.Map{
			dao.History.Columns().DeviceNo:    item.DeviceNo,
			dao.History.Columns().EventId:     item.EventId,
			dao.History.Columns().EventName:   item.EventName,
			dao.History.Columns().EventNumber: item.EventNumber,
			dao.History.Columns().StartTime:   item.StartTime,
			dao.History.Columns().EndTime:     item.EndTime,
			dao.History.Columns().Remark:      item.Remark,
		}).Insert()
		if err != nil {
			return err
		}
		id, _ = res.LastInsertId()
		return nil
	})
	if err == nil && id > 0 {
		item.Id = id
		historyCache.patchHistoryOnAdd(ctx, item)
		bumpPieceCacheEpoch(ctx, item.DeviceNo)
		publishHistoryChange(ctx, item.DeviceNo, "create", historyToPayload(item))
	}
	if err == nil && id > 0 && isOutboxRelayEnabled() {
		version := time.Now().UnixNano()
		if e2 := workeroutbox.EnqueueDomainOutbox(ctx, eventkit.RoutingHistoryRecordCreated, map[string]interface{}{
			"event_id":     fmt.Sprintf("history-created-%d", time.Now().UnixNano()),
			"version":      version,
			"history_id":   id,
			"device_no":    item.DeviceNo,
			"event_id_ref": item.EventId,
			"event_name":   item.EventName,
			"event_number": item.EventNumber,
			"start_time":   item.StartTime,
			"end_time":     item.EndTime,
			"remark":       item.Remark,
			"occurred_at":  time.Now().Format(time.RFC3339Nano),
		}); e2 != nil {
			glog.Warningf(ctx, "[history] worker outbox enqueue failed after insert history_id=%d err=%v", id, e2)
		}
	}
	return id, err
}

func UpdateDeviceHistory(ctx context.Context, item entity.History) error {
	if item.Id <= 0 {
		return fmt.Errorf("id 无效")
	}
	item.DeviceNo = strings.TrimSpace(item.DeviceNo)
	item.EventName = strings.TrimSpace(item.EventName)
	item.Remark = strings.TrimSpace(item.Remark)
	if item.DeviceNo == "" {
		return fmt.Errorf("deviceNo 不能为空")
	}
	err := g.DB(dao.History.Group()).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.History.Table()).
			Where(dao.History.Columns().Id, item.Id).
			Where(dao.History.Columns().DeviceNo, item.DeviceNo).
			Data(g.Map{
				dao.History.Columns().EventId:     item.EventId,
				dao.History.Columns().EventName:   item.EventName,
				dao.History.Columns().EventNumber: item.EventNumber,
				dao.History.Columns().StartTime:   item.StartTime,
				dao.History.Columns().EndTime:     item.EndTime,
				dao.History.Columns().Remark:      item.Remark,
			}).Update()
		return err
	})
	if err == nil {
		historyCache.patchHistoryOnUpdate(ctx, item)
		bumpPieceCacheEpoch(ctx, item.DeviceNo)
		publishHistoryChange(ctx, item.DeviceNo, "update", historyToPayload(item))
	}
	if err == nil && isOutboxRelayEnabled() {
		version := time.Now().UnixNano()
		if e2 := workeroutbox.EnqueueDomainOutbox(ctx, eventkit.RoutingHistoryRecordUpdated, map[string]interface{}{
			"event_id":     fmt.Sprintf("history-updated-%d", time.Now().UnixNano()),
			"version":      version,
			"history_id":   item.Id,
			"device_no":    item.DeviceNo,
			"event_id_ref": item.EventId,
			"event_name":   item.EventName,
			"event_number": item.EventNumber,
			"start_time":   item.StartTime,
			"end_time":     item.EndTime,
			"remark":       item.Remark,
			"occurred_at":  time.Now().Format(time.RFC3339Nano),
		}); e2 != nil {
			glog.Warningf(ctx, "[history] worker outbox enqueue failed after update history_id=%d err=%v", item.Id, e2)
		}
	}
	return err
}

func DeleteDeviceHistory(ctx context.Context, id int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if id <= 0 || deviceNo == "" {
		return fmt.Errorf("参数无效")
	}
	err := g.DB(dao.History.Group()).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.History.Table()).
			Where(dao.History.Columns().Id, id).
			Where(dao.History.Columns().DeviceNo, deviceNo).
			Delete()
		return err
	})
	if err == nil {
		historyCache.patchHistoryOnDelete(ctx, deviceNo, id)
		bumpPieceCacheEpoch(ctx, deviceNo)
		publishHistoryChange(ctx, deviceNo, "delete", map[string]interface{}{
			"id":       id,
			"deviceNo": deviceNo,
		})
	}
	if err == nil && isOutboxRelayEnabled() {
		version := time.Now().UnixNano()
		if e2 := workeroutbox.EnqueueDomainOutbox(ctx, eventkit.RoutingHistoryRecordDeleted, map[string]interface{}{
			"event_id":    fmt.Sprintf("history-deleted-%d", time.Now().UnixNano()),
			"version":     version,
			"history_id":  id,
			"device_no":   deviceNo,
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		}); e2 != nil {
			glog.Warningf(ctx, "[history] worker outbox enqueue failed after delete history_id=%d err=%v", id, e2)
		}
	}
	return err
}

func isOutboxRelayEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
