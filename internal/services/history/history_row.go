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
		c.ExpectStartTime,
		c.ExpectEndTime,
		c.ExpectAmount,
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

// optionalInt64FromRow 将可空 BIGINT 读成 *int64；NULL 保持 nil，避免与 0 混淆。
func optionalInt64FromRow(v gdb.Value) *int64 {
	if v.IsNil() {
		return nil
	}
	n := v.Int64()
	return &n
}

// optionalFloat64FromRow 将可空 DOUBLE 读成 *float64；NULL 保持 nil。
func optionalFloat64FromRow(v gdb.Value) *float64 {
	if v.IsNil() {
		return nil
	}
	n := v.Float64()
	return &n
}

func historyRowToEntity(row gdb.Record) entity.History {
	c := dao.History.Columns()
	return entity.History{
		Id:              row[c.Id].Int64(),
		DeviceNo:        row[c.DeviceNo].String(),
		EventId:         row[c.EventId].Int64(),
		EventName:       row[c.EventName].String(),
		EventNumber:     row[c.EventNumber].Int64(),
		EventUnit:       row[c.EventUnit].String(),
		StartTime:       row[c.StartTime].Int64(),
		EndTime:         row[c.EndTime].Int64(),
		Remark:          row[c.Remark].String(),
		PostId:          row[c.PostId].Uint64(),
		MediaType:       row[c.MediaType].Int(),
		ImageKeys:       strings.TrimSpace(row[c.ImageKeys].String()),
		VideoKey:        row[c.VideoKey].String(),
		ExpectStartTime: optionalInt64FromRow(row[c.ExpectStartTime]),
		ExpectEndTime:   optionalInt64FromRow(row[c.ExpectEndTime]),
		ExpectAmount:    optionalFloat64FromRow(row[c.ExpectAmount]),
	}
}

func historyInsertData(item entity.History) map[string]interface{} {
	c := dao.History.Columns()
	data := map[string]interface{}{
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
	// 预计字段：仅非 nil 才写入；全缺则列保持 DEFAULT NULL，禁止用零值占位。
	if item.ExpectStartTime != nil {
		data[c.ExpectStartTime] = *item.ExpectStartTime
	}
	if item.ExpectEndTime != nil {
		data[c.ExpectEndTime] = *item.ExpectEndTime
	}
	if item.ExpectAmount != nil {
		data[c.ExpectAmount] = *item.ExpectAmount
	}
	return data
}

// historyExpectPatchData 构造 expect-batch 专用 UPDATE 列集合。
// 业务约束：仅包含非 nil 的 expect_*；全缺则返回空 map（调用方做归属校验后 no-op）。
// 与 historyUpdateData 隔离：普通 update/batch 冻结路径永不写 expect_*。
func historyExpectPatchData(expectStartTime, expectEndTime *int64, expectAmount *float64) map[string]interface{} {
	c := dao.History.Columns()
	data := make(map[string]interface{}, 3)
	if expectStartTime != nil {
		data[c.ExpectStartTime] = *expectStartTime
	}
	if expectEndTime != nil {
		data[c.ExpectEndTime] = *expectEndTime
	}
	if expectAmount != nil {
		data[c.ExpectAmount] = *expectAmount
	}
	return data
}

// historyUpdateData 构造 UPDATE 列集合。
// 业务约束：永远不包含 expect_*，避免停表/改名等路径用结构体零值抹掉已落库的预计字段。
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
