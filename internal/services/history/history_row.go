package history

import (
	"context"
	"encoding/json"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

func historyListSelectFields() []interface{} {
	c := dao.History.Columns()
	return []interface{}{
		c.Id,
		c.DeviceNo,
		c.EventId,
		c.EventName,
		c.EventNumber,
		c.EventUnit,
		c.StartTime,
		c.EndTime,
		c.Remark,
		c.PostId,
		c.MediaType,
		c.ImageKeys,
		c.VideoKey,
	}
}

func encodeImageKeys(keys []string) string {
	raw, err := json.Marshal(keys)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func nullablePostId(v uint64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func nullableMediaType(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func nullableImageKeys(keys []string) interface{} {
	if len(keys) == 0 {
		return nil
	}
	return encodeImageKeys(keys)
}

func nullableVideoKey(v string) interface{} {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func decodeImageKeys(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var keys []string
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return nil
	}
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			out = append(out, k)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func nullableEventUnit(v string) interface{} {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func lookupEventUnit(ctx context.Context, eventID int64) string {
	return resolveEventUnitFromDevice(ctx, eventID)
}

// enrichHistoryEventUnit 写入前补全 event_unit：优先已有值，否则查事件主档。
func enrichHistoryEventUnit(ctx context.Context, item *entity.History) {
	if item == nil {
		return
	}
	if strings.TrimSpace(item.EventUnit) != "" {
		item.EventUnit = strings.TrimSpace(item.EventUnit)
		return
	}
	item.EventUnit = lookupEventUnit(ctx, item.EventId)
}

func historyRowToEntity(row gdb.Record) entity.History {
	c := dao.History.Columns()
	return entity.History{
		Id:          row[c.Id].Int64(),
		DeviceNo:    row[c.DeviceNo].String(),
		EventId:     row[c.EventId].Int64(),
		EventName:   row[c.EventName].String(),
		EventNumber: row[c.EventNumber].Int64(),
		EventUnit:   row[c.EventUnit].String(),
		StartTime:   row[c.StartTime].Int64(),
		EndTime:     row[c.EndTime].Int64(),
		Remark:      row[c.Remark].String(),
		PostId:      row[c.PostId].Uint64(),
		MediaType:   row[c.MediaType].Int(),
		ImageKeys:   strings.TrimSpace(row[c.ImageKeys].String()),
		VideoKey:    row[c.VideoKey].String(),
	}
}

func historyInsertData(item entity.History) map[string]interface{} {
	c := dao.History.Columns()
	return map[string]interface{}{
		c.DeviceNo:    item.DeviceNo,
		c.EventId:     item.EventId,
		c.EventName:   item.EventName,
		c.EventNumber: item.EventNumber,
		c.EventUnit:   nullableEventUnit(item.EventUnit),
		c.StartTime:   item.StartTime,
		c.EndTime:     item.EndTime,
		c.Remark:      item.Remark,
		c.PostId:      nullablePostId(item.PostId),
		c.MediaType:   nullableMediaType(item.MediaType),
		c.ImageKeys:   nullableImageKeys(decodeImageKeys(item.ImageKeys)),
		c.VideoKey:    nullableVideoKey(item.VideoKey),
	}
}

func historyUpdateData(item entity.History) map[string]interface{} {
	c := dao.History.Columns()
	return map[string]interface{}{
		c.EventId:     item.EventId,
		c.EventName:   item.EventName,
		c.EventNumber: item.EventNumber,
		c.EventUnit:   nullableEventUnit(item.EventUnit),
		c.StartTime:   item.StartTime,
		c.EndTime:     item.EndTime,
		c.Remark:      item.Remark,
		c.PostId:      nullablePostId(item.PostId),
		c.MediaType:   nullableMediaType(item.MediaType),
		c.ImageKeys:   nullableImageKeys(decodeImageKeys(item.ImageKeys)),
		c.VideoKey:    nullableVideoKey(item.VideoKey),
	}
}
