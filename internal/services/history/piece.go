package history

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const pieceListCacheTTL = 60 * time.Second

// ListHistoryPiece 按设备、事件与时间区间查询历史记录（startTime/endTime 为 Unix 秒，与库内 BIGINT 一致）。
func ListHistoryPiece(ctx context.Context, deviceNo string, eventID int64, startTimeUnixSec, endTimeUnixSec int64) ([]entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || eventID <= 0 || startTimeUnixSec <= 0 || endTimeUnixSec <= 0 {
		return nil, fmt.Errorf("deviceNo、eventId、startTime、endTime 均不能为空且须为有效 Unix 秒")
	}
	ver := pieceCacheEpoch(ctx, deviceNo)
	cacheKey := pieceCacheKey(deviceNo, eventID, startTimeUnixSec, endTimeUnixSec, ver)
	if raw, err := g.Redis().Do(ctx, "GET", cacheKey); err == nil && raw != nil {
		s := strings.TrimSpace(raw.String())
		if s != "" {
			var cached []entity.History
			if err := json.Unmarshal([]byte(s), &cached); err == nil {
				return cached, nil
			}
		}
	}
	stCol := dao.History.Columns().StartTime
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
		Where(dao.History.Columns().EventId, eventID).
		Where(stCol+" >= ? AND "+stCol+" <= ?", startTimeUnixSec, endTimeUnixSec).
		OrderAsc(dao.History.Columns().Id).
		All()
	if err != nil {
		return nil, err
	}
	out := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.History{
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
	if blob, err := json.Marshal(out); err == nil {
		if _, err2 := g.Redis().Do(ctx, "SET", cacheKey, string(blob), "EX", int(pieceListCacheTTL.Seconds())); err2 != nil {
			glog.Warningf(ctx, "[history-piece] 写缓存失败 key=%s err=%v", cacheKey, err2)
		}
	}
	return out, nil
}

func pieceCacheEpoch(ctx context.Context, deviceNo string) int64 {
	key := redisKeyPieceVerPrefix + deviceNo
	raw, err := g.Redis().Do(ctx, "GET", key)
	if err != nil || raw == nil {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(raw.String()), 10, 64)
	return n
}
