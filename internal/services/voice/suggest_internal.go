package voice

import (
	"context"
	"fmt"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"
)

// ListSuggestItems 查询成长建议列表（权威数据在 voice 库 suggest 表）。
func ListSuggestItems(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return []entity.Suggest{}, nil
	}
	rows := make([]entity.Suggest, 0)
	err := dao.Suggest.Ctx(ctx).
		Fields(dao.Suggest.Columns().Id, dao.Suggest.Columns().DeviceNo, dao.Suggest.Columns().Suggest, dao.Suggest.Columns().Time).
		Where(dao.Suggest.Columns().DeviceNo, deviceNo).
		OrderDesc(dao.Suggest.Columns().Id).
		Scan(&rows)
	return rows, err
}

// DeleteSuggestItem 删除一条成长建议。
func DeleteSuggestItem(ctx context.Context, id int64, deviceNo string) error {
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
