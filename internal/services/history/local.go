package history

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	contracts "hello/internal/services/contracts"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type localService struct{}

var historyCache = newCacheRepo()

func (s *localService) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	return ListDeviceHistory(ctx, deviceNo)
}

func (s *localService) ListHistoryPage(ctx context.Context, deviceNo string, page int, pageSize int) (contracts.HistoryPageResult, error) {
	return ListDeviceHistoryPage(ctx, deviceNo, page, pageSize)
}

func (s *localService) ListHistoryFilter(ctx context.Context, deviceNo string, eventIds []int64, startTime int64, endTime int64, limit int, remark string, ignoreTimeRange bool) ([]entity.History, error) {
	return ListDeviceHistoryFilter(ctx, deviceNo, eventIds, startTime, endTime, limit, remark, ignoreTimeRange)
}

func (s *localService) ListHistoryPageV2(ctx context.Context, deviceNo string, page int, pageSize int, startTime int64, endTime int64, limit int) (contracts.HistoryPageResult, error) {
	return ListDeviceHistoryPageV2(ctx, deviceNo, page, pageSize, startTime, endTime, limit)
}

func (s *localService) GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error) {
	return GetLatestDeviceHistory(ctx, deviceNo)
}

func (s *localService) EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTimeUnixSec int64, remark string) (bool, error) {
	return EndLatestDeviceHistoryIfMatch(ctx, deviceNo, eventID, endTimeUnixSec, remark)
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

func (s *localService) GetBirthday(ctx context.Context, deviceNo string) (string, int64, int, error) {
	return GetDeviceBirthday(ctx, deviceNo)
}

func (s *localService) SaveBirthday(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error {
	return SaveBirthday(ctx, deviceNo, babyName, birthdayUnixSec, sex)
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
		Fields(historyListSelectFields()...).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.History.Columns().Id).
		All()
	if err != nil {
		return nil, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyRowToEntity(row))
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
		Fields(historyListSelectFields()...).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.History.Columns().Id).
		Page(page, pageSize).
		All()
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyRowToEntity(row))
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

// EndLatestDeviceHistoryIfMatch 按 eventId 闭合该设备最近一条未结束历史。
//
// 业务语义（App end-latest 与语音 feeding+end 共用）：
//   - 匹配条件：device_no + event_id + end_time=0，按 id DESC 取一条；
//   - 不要求该行是设备「全局最新」一条，从而「开始睡眠 → 中间记其它事件 → 结束睡眠」仍能闭合原睡眠行；
//   - 无未闭合同 event 时返回 false（不改已结束行），由 voice 降级瞬时 AddHistory。
//
// 注意：禁止再使用「GetLatest 再比 EventId」——那会在中间夹杂其它事件时误判未匹配并导致重复新建。
func EndLatestDeviceHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTimeUnixSec int64, remark string) (bool, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || eventID <= 0 {
		return false, nil
	}
	// 0 表示调用方未显式传结束时间，落库为当前 Unix 秒。
	if endTimeUnixSec == 0 {
		endTimeUnixSec = time.Now().Unix()
	}
	remark = strings.TrimSpace(remark)

	// 权威查询：同 event 的最近未闭合行（end_time=0），而非全局最新 history。
	var open entity.History
	err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(dao.History.Columns().EventId, eventID).
		Where(dao.History.Columns().EndTime, 0).
		OrderDesc(dao.History.Columns().Id).
		Limit(1).
		Scan(&open)
	if err != nil {
		return false, err
	}
	if open.Id <= 0 {
		return false, nil
	}

	data := g.Map{dao.History.Columns().EndTime: endTimeUnixSec}
	if remark != "" {
		data[dao.History.Columns().Remark] = remark
		open.Remark = remark
	}
	_, err = dao.History.Ctx(ctx).
		Where(dao.History.Columns().Id, open.Id).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Data(data).
		Update()
	if err != nil {
		return false, err
	}
	open.EndTime = endTimeUnixSec
	// patchHistoryOnUpdate 仅在 latest.id==本行时改写 latest 缓存，闭合非最新未结束行是安全的。
	historyCache.patchHistoryOnUpdate(ctx, open)
	// 与 UpdateDeviceHistory 一致：改 endTime（及可选 remark）时也要递增 piece 版本并向 app:history:notify 广播，否则 App WS 收不到「结束事件」类更新。
	bumpPieceCacheEpoch(ctx, deviceNo)
	publishHistoryChange(ctx, deviceNo, "update", historyToNotifyPayload(ctx, open))
	return true, nil
}

// errDuplicateOpenHistoryStart 同 eventId 已存在未闭合行时拒绝再次 start（进行中 insert/update）。
const errDuplicateOpenHistoryStart = "已在进行中"

// hasOpenHistoryForEvent 查询设备下指定 eventId 是否存在未闭合（end_time=0）历史行。
// excludeID>0 时在查重中排除该行（用于 update 将 end_time 改回 0）。
// 与 EndLatestDeviceHistoryIfMatch 的匹配维度（device_no + event_id + end_time=0）对称。
func hasOpenHistoryForEvent(ctx context.Context, deviceNo string, eventID int64, excludeID int64) (bool, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || eventID <= 0 {
		return false, nil
	}
	m := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(dao.History.Columns().EventId, eventID).
		Where(dao.History.Columns().EndTime, 0)
	if excludeID > 0 {
		m = m.WhereNot(dao.History.Columns().Id, excludeID)
	}
	n, err := m.Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// guardDuplicateOpenHistoryStart 进行中（endTime=0）写入前校验同 event 是否已有未闭合行。
func guardDuplicateOpenHistoryStart(ctx context.Context, deviceNo string, eventID int64, endTime int64, excludeID int64) error {
	if eventID <= 0 || endTime != 0 {
		return nil
	}
	open, err := hasOpenHistoryForEvent(ctx, deviceNo, eventID, excludeID)
	if err != nil {
		return err
	}
	if open {
		return fmt.Errorf("%s", errDuplicateOpenHistoryStart)
	}
	return nil
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

func SaveBirthday(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error {
	deviceNo = strings.TrimSpace(deviceNo)
	babyName = strings.TrimSpace(babyName)
	if deviceNo == "" {
		return nil
	}
	if sex > 0 {
		sex = 1
	}
	err := delegateSaveProfile(ctx, deviceNo, babyName, birthdayUnixSec, sex)
	if err != nil {
		logDelegateFailure(ctx, "save_profile", err)
		return err
	}
	_ = historyCache.setBirthday(ctx, deviceNo, babyName, birthdayUnixSec, sex)
	return nil
}

func GetDeviceBirthday(ctx context.Context, deviceNo string) (string, int64, int, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", 0, 0, nil
	}
	if babyName, birthday, sex, ok, err := historyCache.getBirthday(ctx, deviceNo); err == nil && ok {
		return babyName, birthday, sex, nil
	}
	babyName, birthday, sex, err := delegateGetProfile(ctx, deviceNo)
	if err != nil {
		logDelegateFailure(ctx, "get_profile", err)
		return "", 0, 0, err
	}
	babyName = strings.TrimSpace(babyName)
	if sex > 0 {
		sex = 1
	}
	_ = historyCache.setBirthday(ctx, deviceNo, babyName, birthday, sex)
	return babyName, birthday, sex, nil
}

func AddDeviceHistory(ctx context.Context, item entity.History) (int64, error) {
	item.DeviceNo = strings.TrimSpace(item.DeviceNo)
	item.EventName = strings.TrimSpace(item.EventName)
	item.Remark = strings.TrimSpace(item.Remark)
	if item.DeviceNo == "" {
		return 0, fmt.Errorf("deviceNo 不能为空")
	}
	if err := guardDuplicateOpenHistoryStart(ctx, item.DeviceNo, item.EventId, item.EndTime, 0); err != nil {
		return 0, err
	}
	enrichHistoryEventUnit(ctx, &item)
	var id int64
	err := g.DB(dao.History.Group()).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		res, err := tx.Model(dao.History.Table()).Data(historyInsertData(item)).Insert()
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
		publishHistoryChange(ctx, item.DeviceNo, "create", historyToNotifyPayload(ctx, item))
	}
	return id, err
}

func GetDeviceHistoryByID(ctx context.Context, id int64, deviceNo string) (entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if id <= 0 || deviceNo == "" {
		return entity.History{}, fmt.Errorf("参数无效")
	}
	row, err := dao.History.Ctx(ctx).
		Fields(historyListSelectFields()...).
		Where(dao.History.Columns().Id, id).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		One()
	if err != nil {
		return entity.History{}, err
	}
	if row.IsEmpty() {
		return entity.History{}, fmt.Errorf("记录不存在")
	}
	return historyRowToEntity(row), nil
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
	if err := guardDuplicateOpenHistoryStart(ctx, item.DeviceNo, item.EventId, item.EndTime, item.Id); err != nil {
		return err
	}
	enrichHistoryEventUnit(ctx, &item)
	err := g.DB(dao.History.Group()).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(dao.History.Table()).
			Where(dao.History.Columns().Id, item.Id).
			Where(dao.History.Columns().DeviceNo, item.DeviceNo).
			Data(historyUpdateData(item)).Update()
		return err
	})
	if err == nil {
		historyCache.patchHistoryOnUpdate(ctx, item)
		bumpPieceCacheEpoch(ctx, item.DeviceNo)
		publishHistoryChange(ctx, item.DeviceNo, "update", historyToNotifyPayload(ctx, item))
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
	return err
}

// escapeLikeKeyword 转义 LIKE 通配符，避免用户输入 %/_ 扩大匹配。
func escapeLikeKeyword(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// ListDeviceHistoryFilter 多条件筛选历史记录。
// eventIds 非空时按事件ID列表过滤；ignoreTimeRange 为假且 startTime/endTime > 0 时按时间区间过滤。
// ignoreTimeRange 为真时强制忽略 startTime/endTime（即使非 0），用于时间窗不明确但仍要查「之前发生过什么」。
// remark 非空时 AND 备注模糊（排除 NULL/空串）；无 eventIds 时视为探针，limit 上限 20。
// 返回结果按 id 倒序，limit 默认 100，上限 500。
func ListDeviceHistoryFilter(ctx context.Context, deviceNo string, eventIds []int64, startTime int64, endTime int64, limit int, remark string, ignoreTimeRange bool) ([]entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return []entity.History{}, nil
	}
	remark = strings.TrimSpace(remark)
	// 仅备注、不定事件：探针路径，强制小 limit，避免扫全表再灌模型。
	if remark != "" && len(eventIds) == 0 {
		if limit <= 0 || limit > 20 {
			limit = 20
		}
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	m := dao.History.Ctx(ctx).
		Fields(historyListSelectFields()...).
		Where(dao.History.Columns().DeviceNo, deviceNo)
	if len(eventIds) > 0 {
		m = m.WhereIn(dao.History.Columns().EventId, eventIds)
	}
	// 显式忽略时间窗时不施加任何 start_time 条件，避免调用方猜测区间漏查。
	if !ignoreTimeRange {
		if startTime > 0 {
			m = m.WhereGTE(dao.History.Columns().StartTime, startTime)
		}
		if endTime > 0 {
			m = m.WhereLTE(dao.History.Columns().StartTime, endTime)
		}
	}
	// 可空备注不参与模糊命中；LIKE 前转义 % _
	if remark != "" {
		col := dao.History.Columns().Remark
		m = m.WhereNot(col, "").Where(col+" LIKE ?", "%"+escapeLikeKeyword(remark)+"%")
	}
	rows, err := m.
		OrderDesc(dao.History.Columns().Id).
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyRowToEntity(row))
	}
	return items, nil
}

// ListDeviceHistoryPageV2 分页查询设备历史记录 v2，支持时间范围和 limit 参数。
// limit > 0 时用 limit 替代 pageSize，page 忽略（固定为1）。
// startTime/endTime > 0 时按时间区间过滤。不传新参数时行为与 v1 完全一致。
func ListDeviceHistoryPageV2(ctx context.Context, deviceNo string, page int, pageSize int, startTime int64, endTime int64, limit int) (contracts.HistoryPageResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	effectivePage := page
	effectivePageSize := pageSize
	// limit > 0 时替代 pageSize，page 固定为 1
	if limit > 0 {
		effectivePage = 1
		effectivePageSize = limit
		if effectivePageSize > 500 {
			effectivePageSize = 500
		}
	} else {
		if effectivePage <= 0 {
			effectivePage = 1
		}
		if effectivePageSize <= 0 {
			effectivePageSize = 20
		}
		if effectivePageSize > 100 {
			effectivePageSize = 100
		}
	}
	if deviceNo == "" {
		return contracts.HistoryPageResult{List: []entity.History{}, Total: 0, Page: effectivePage, PageSize: effectivePageSize}, nil
	}
	countModel := dao.History.Ctx(ctx).Where(dao.History.Columns().DeviceNo, deviceNo)
	listModel := dao.History.Ctx(ctx).
		Fields(historyListSelectFields()...).
		Where(dao.History.Columns().DeviceNo, deviceNo)
	if startTime > 0 {
		countModel = countModel.WhereGTE(dao.History.Columns().StartTime, startTime)
		listModel = listModel.WhereGTE(dao.History.Columns().StartTime, startTime)
	}
	if endTime > 0 {
		countModel = countModel.WhereLTE(dao.History.Columns().StartTime, endTime)
		listModel = listModel.WhereLTE(dao.History.Columns().StartTime, endTime)
	}
	total, err := countModel.Count()
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	if total == 0 {
		return contracts.HistoryPageResult{List: []entity.History{}, Total: 0, Page: effectivePage, PageSize: effectivePageSize}, nil
	}
	rows, err := listModel.
		OrderDesc(dao.History.Columns().Id).
		Page(effectivePage, effectivePageSize).
		All()
	if err != nil {
		return contracts.HistoryPageResult{}, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyRowToEntity(row))
	}
	return contracts.HistoryPageResult{List: items, Total: total, Page: effectivePage, PageSize: effectivePageSize}, nil
}
