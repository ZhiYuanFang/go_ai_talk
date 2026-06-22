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
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/os/glog"
)

const pieceListCacheTTL = 60 * time.Second

var pieceCache = cachekit.Default()

// ListHistoryPiece 按设备、事件与时间区间查询历史记录（startTime/endTime 为 Unix 秒，与库内 BIGINT 一致）。
func ListHistoryPiece(ctx context.Context, deviceNo string, eventID int64, startTimeUnixSec, endTimeUnixSec int64) ([]entity.History, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || eventID <= 0 || startTimeUnixSec <= 0 || endTimeUnixSec <= 0 {
		return nil, fmt.Errorf("deviceNo、eventId、startTime、endTime 均不能为空且须为有效 Unix 秒")
	}
	ver := pieceCacheEpoch(ctx, deviceNo)
	cacheKey := cachekit.HistoryPieceDataKey(deviceNo, eventID, startTimeUnixSec, endTimeUnixSec, ver)
	if raw, ok, err := pieceCache.Get(ctx, cacheKey); err == nil && ok && raw != "" {
		var cached []entity.History
		if err := json.Unmarshal([]byte(raw), &cached); err == nil {
			return cached, nil
		}
	}
	stCol := dao.History.Columns().StartTime
	rows, err := dao.History.Ctx(ctx).
		Fields(historyListSelectFields()...).
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
		out = append(out, historyRowToEntity(row))
	}
	if blob, err := json.Marshal(out); err == nil {
		if err2 := pieceCache.SetEX(ctx, cacheKey, string(blob), pieceListCacheTTL); err2 != nil {
			glog.Warningf(ctx, "[history-piece] 写缓存失败 key=%s err=%v", cacheKey, err2)
		}
	}
	return out, nil
}

func pieceCacheEpoch(ctx context.Context, deviceNo string) int64 {
	key := cachekit.HistoryPieceVerKey(deviceNo)
	raw, ok, err := pieceCache.Get(ctx, key)
	if err != nil || !ok {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return n
}
