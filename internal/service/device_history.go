package service

import (
	"context"
	"fmt"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type sDeviceHistory struct{}

var insDeviceHistory sDeviceHistory

// DeviceHistory 返回设备历史/建议等查询实现。
func DeviceHistory() DeviceHistoryContract {
	return &insDeviceHistory
}

var _ DeviceHistoryContract = (*sDeviceHistory)(nil)

func (s *sDeviceHistory) ListHistory(ctx context.Context, deviceNo string) ([]DeviceHistoryItem, error) {
	return ListDeviceHistory(ctx, deviceNo)
}

func (s *sDeviceHistory) ListSuggest(ctx context.Context, deviceNo string) ([]DeviceSuggestItem, error) {
	return ListDeviceSuggest(ctx, deviceNo)
}

func (s *sDeviceHistory) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	return DeleteDeviceSuggest(ctx, id, deviceNo)
}

func (s *sDeviceHistory) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	return ListEventOptions(ctx)
}

func (s *sDeviceHistory) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	return GetDeviceBirthday(ctx, deviceNo)
}

func (s *sDeviceHistory) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	return SaveBirthday(ctx, deviceNo, birthday, sex)
}

func (s *sDeviceHistory) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	return AddDeviceHistory(ctx, item)
}

func (s *sDeviceHistory) UpdateHistory(ctx context.Context, item entity.History) error {
	return UpdateDeviceHistory(ctx, item)
}

func (s *sDeviceHistory) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	return DeleteDeviceHistory(ctx, id, deviceNo)
}

func ListDeviceHistory(ctx context.Context, deviceNo string) ([]DeviceHistoryItem, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return []DeviceHistoryItem{}, nil
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

	items := make([]DeviceHistoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, DeviceHistoryItem{
			Id:          row[dao.History.Columns().Id].Int64(),
			DeviceNo:    row[dao.History.Columns().DeviceNo].String(),
			EventId:     row[dao.History.Columns().EventId].Int64(),
			EventName:   row[dao.History.Columns().EventName].String(),
			EventNumber: row[dao.History.Columns().EventNumber].Int64(),
			StartTime:   row[dao.History.Columns().StartTime].String(),
			EndTime:     row[dao.History.Columns().EndTime].String(),
			Remark:      row[dao.History.Columns().Remark].String(),
		})
	}
	return items, nil
}

func ListDeviceSuggest(ctx context.Context, deviceNo string) ([]DeviceSuggestItem, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return []DeviceSuggestItem{}, nil
	}

	rows := make([]entity.Suggest, 0)
	err := dao.Suggest.Ctx(ctx).
		Fields(dao.Suggest.Columns().Id, dao.Suggest.Columns().DeviceNo, dao.Suggest.Columns().Suggest, dao.Suggest.Columns().Time).
		Where(dao.Suggest.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.Suggest.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func DeleteDeviceSuggest(ctx context.Context, id int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if id <= 0 || deviceNo == "" {
		return fmt.Errorf("参数无效")
	}
	_, err := dao.Suggest.Ctx(ctx).
		Where(dao.Suggest.Columns().Id, id).
		Where(dao.Suggest.Columns().DeviceNo, deviceNo).
		Delete()
	return err
}

func ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	rows := make([]entity.Event, 0)
	err := dao.Event.Ctx(ctx).
		Fields(dao.Event.Columns().Id, dao.Event.Columns().Name, dao.Event.Columns().NeedQuantity).
		OrderAsc(dao.Event.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	deviceNo = strings.TrimSpace(deviceNo)
	birthday = strings.TrimSpace(birthday)
	if deviceNo == "" {
		return nil
	}
	if sex > 0 {
		sex = 1
	}
	_, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Data(g.Map{
		dao.User.Columns().Birthday: birthday,
		dao.User.Columns().Sex:      sex,
	}).Update()
	return err
}

func GetDeviceBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", 0, nil
	}

	row, err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Birthday, dao.User.Columns().Sex).
		Where(dao.User.Columns().DeviceNo, deviceNo).
		Limit(1).
		One()
	if err != nil {
		return "", 0, err
	}
	if row == nil || row.IsEmpty() {
		return "", 0, nil
	}
	sex := row[dao.User.Columns().Sex].Int()
	if sex > 0 {
		sex = 1
	}
	return strings.TrimSpace(row[dao.User.Columns().Birthday].String()), sex, nil
}

func AddDeviceHistory(ctx context.Context, item entity.History) (int64, error) {
	item.DeviceNo = strings.TrimSpace(item.DeviceNo)
	item.EventName = strings.TrimSpace(item.EventName)
	item.StartTime = strings.TrimSpace(item.StartTime)
	item.EndTime = strings.TrimSpace(item.EndTime)
	item.Remark = strings.TrimSpace(item.Remark)
	if item.DeviceNo == "" {
		return 0, fmt.Errorf("deviceNo 不能为空")
	}
	res, err := dao.History.Ctx(ctx).Data(g.Map{
		dao.History.Columns().DeviceNo:    item.DeviceNo,
		dao.History.Columns().EventId:     item.EventId,
		dao.History.Columns().EventName:   item.EventName,
		dao.History.Columns().EventNumber: item.EventNumber,
		dao.History.Columns().StartTime:   item.StartTime,
		dao.History.Columns().EndTime:     item.EndTime,
		dao.History.Columns().Remark:      item.Remark,
	}).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func UpdateDeviceHistory(ctx context.Context, item entity.History) error {
	if item.Id <= 0 {
		return fmt.Errorf("id 无效")
	}
	item.DeviceNo = strings.TrimSpace(item.DeviceNo)
	item.EventName = strings.TrimSpace(item.EventName)
	item.StartTime = strings.TrimSpace(item.StartTime)
	item.EndTime = strings.TrimSpace(item.EndTime)
	item.Remark = strings.TrimSpace(item.Remark)
	if item.DeviceNo == "" {
		return fmt.Errorf("deviceNo 不能为空")
	}
	_, err := dao.History.Ctx(ctx).
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
}

func DeleteDeviceHistory(ctx context.Context, id int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if id <= 0 || deviceNo == "" {
		return fmt.Errorf("参数无效")
	}
	_, err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().Id, id).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Delete()
	return err
}
