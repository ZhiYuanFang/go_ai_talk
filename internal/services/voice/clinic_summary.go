package voice

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"
)

const (
	clinicRemarkRecordMax     = 30
	clinicRemarkMaxRunes      = 200
	clinicSummaryEmptyJSON    = `{"by_event":[],"records_with_remark":[]}`
)

// clinicEventAgg 单 event 7 天内聚合统计。
type clinicEventAgg struct {
	EventID              int64  `json:"event_id"`
	EventName            string `json:"event_name"`
	Count                int    `json:"count"`
	TotalAmount          int64  `json:"total_amount,omitempty"`
	AmountUnit           string `json:"amount_unit,omitempty"`
	TotalDurationMinutes int    `json:"total_duration_minutes,omitempty"`
	LastAt               int64  `json:"last_at,omitempty"`
}

// clinicRemarkRecord 近 7 天内有用户备注的单条喂养记录（非全量 dump）。
type clinicRemarkRecord struct {
	EventName       string `json:"event_name"`
	StartTime       string `json:"start_time"`
	AmountValue     int64  `json:"amount_value,omitempty"`
	DurationMinutes int    `json:"duration_minutes,omitempty"`
	Remark          string `json:"remark"`
}

// clinicSummaryPayload 注入 LLM 的 7 天喂养摘要：聚合 + 有备注记录子集。
type clinicSummaryPayload struct {
	ByEvent           []clinicEventAgg     `json:"by_event"`
	RecordsWithRemark []clinicRemarkRecord `json:"records_with_remark"`
}

// clinicSummaryCache Redis voice:clinic:summary:{wxId}:{deviceNo} 持久化结构。
type clinicSummaryCache struct {
	Summary          string `json:"summary"`
	HistoryWatermark int64  `json:"historyWatermark"`
	ComputedAt       int64  `json:"computedAt"`
}

func clinicHistoryCutoffUnix() int64 {
	return time.Now().Add(-7 * 24 * time.Hour).Unix()
}

// historyWatermarkFromRows 从 history 列表取 max(start/end) 作为懒刷新 watermark。
func historyWatermarkFromRows(rows []entity.History) int64 {
	var wm int64
	for _, h := range rows {
		if h.StartTime > wm {
			wm = h.StartTime
		}
		if h.EndTime > wm {
			wm = h.EndTime
		}
	}
	return wm
}

func clinicEventDisplayName(h entity.History, eventNames map[int64]string) string {
	name := eventNames[h.EventId]
	if name == "" {
		name = strings.TrimSpace(h.EventName)
	}
	if name == "" {
		return "未知事件"
	}
	return name
}

func truncateClinicRemark(s string) string {
	if utf8.RuneCountInString(s) <= clinicRemarkMaxRunes {
		return s
	}
	rs := []rune(s)
	return string(rs[:clinicRemarkMaxRunes])
}

func clinicRecordSortTime(h entity.History) int64 {
	if h.StartTime > 0 {
		return h.StartTime
	}
	return h.EndTime
}

func isValidClinicSummaryJSON(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	var payload clinicSummaryPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return false
	}
	return payload.ByEvent != nil && payload.RecordsWithRemark != nil
}

// buildClinicHistorySummary 经 DeviceHistory HTTP 契约拉取 7 天数据：by_event 聚合 + 有备注记录（≤30）。
func buildClinicHistorySummary(ctx context.Context, deviceNo string) (summaryJSON string, watermark int64, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return clinicSummaryEmptyJSON, 0, nil
	}
	rows, err := DeviceHistory().ListHistory(ctx, deviceNo)
	if err != nil {
		return "", 0, err
	}
	cutoff := clinicHistoryCutoffUnix()
	eventNames := loadEventNameAndUnitByID(ctx)
	byEvent := make(map[int64]*clinicEventAgg)
	remarkCandidates := make([]entity.History, 0)
	for _, h := range rows {
		if h.StartTime < cutoff && h.EndTime < cutoff {
			continue
		}
		agg, ok := byEvent[h.EventId]
		if !ok {
			name := clinicEventDisplayName(h, eventNames)
			agg = &clinicEventAgg{EventID: h.EventId, EventName: name}
			if h.EventNumber > 0 {
				agg.AmountUnit = strings.TrimSpace(h.EventUnit)
			}
			byEvent[h.EventId] = agg
		}
		agg.Count++
		if h.EventNumber > 0 {
			agg.TotalAmount += h.EventNumber
		}
		if h.StartTime > 0 && h.EndTime > h.StartTime {
			agg.TotalDurationMinutes += int(math.Round(float64(h.EndTime-h.StartTime) / 60))
		}
		last := h.EndTime
		if h.StartTime > last {
			last = h.StartTime
		}
		if last > agg.LastAt {
			agg.LastAt = last
		}
		if strings.TrimSpace(h.Remark) != "" {
			remarkCandidates = append(remarkCandidates, h)
		}
	}
	byEventOut := make([]clinicEventAgg, 0, len(byEvent))
	for _, v := range byEvent {
		byEventOut = append(byEventOut, *v)
	}
	sort.Slice(remarkCandidates, func(i, j int) bool {
		return clinicRecordSortTime(remarkCandidates[i]) > clinicRecordSortTime(remarkCandidates[j])
	})
	if len(remarkCandidates) > clinicRemarkRecordMax {
		remarkCandidates = remarkCandidates[:clinicRemarkRecordMax]
	}
	recordsWithRemark := make([]clinicRemarkRecord, 0, len(remarkCandidates))
	for _, h := range remarkCandidates {
		rec := clinicRemarkRecord{
			EventName: clinicEventDisplayName(h, eventNames),
			StartTime: formatLocalDatetimeFromUnix(clinicRecordSortTime(h)),
			Remark:    truncateClinicRemark(strings.TrimSpace(h.Remark)),
		}
		if h.EventNumber > 0 {
			rec.AmountValue = h.EventNumber
		}
		if dm := suggestDurationMinutes(h.StartTime, h.EndTime); dm > 0 {
			rec.DurationMinutes = dm
		}
		recordsWithRemark = append(recordsWithRemark, rec)
	}
	payload := clinicSummaryPayload{
		ByEvent:           byEventOut,
		RecordsWithRemark: recordsWithRemark,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", 0, err
	}
	return string(raw), historyWatermarkFromRows(rows), nil
}

// ensureClinicSummary 懒刷新：对比 watermark，过期则重算并写 Redis（TTL 见 aiClinic.summaryTtlSeconds）。
func (s *ClinicService) ensureClinicSummary(ctx context.Context, wxID int64, deviceNo string) (string, error) {
	key := cachekit.VoiceClinicSummaryKey(wxID, deviceNo)
	v, ok, err := clinicCache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	var cached clinicSummaryCache
	if ok && v != "" {
		_ = json.Unmarshal([]byte(v), &cached)
	}
	// 拉 list 仅用于 watermark；与 build 内 list 重复但 MVP 可接受（design 允许后续聚合 API）。
	rows, err := DeviceHistory().ListHistory(ctx, deviceNo)
	if err != nil {
		if cached.Summary != "" && isValidClinicSummaryJSON(cached.Summary) {
			return cached.Summary, nil
		}
		return "", err
	}
	currentWM := historyWatermarkFromRows(rows)
	cacheValid := cached.Summary != "" && isValidClinicSummaryJSON(cached.Summary)
	if cacheValid && cached.HistoryWatermark >= currentWM && currentWM > 0 {
		return cached.Summary, nil
	}
	if cacheValid && currentWM == 0 && cached.HistoryWatermark == 0 {
		return cached.Summary, nil
	}
	summary, wm, err := buildClinicHistorySummary(ctx, deviceNo)
	if err != nil {
		if cacheValid {
			return cached.Summary, nil
		}
		return "", err
	}
	payload, _ := json.Marshal(clinicSummaryCache{
		Summary:          summary,
		HistoryWatermark: wm,
		ComputedAt:       time.Now().Unix(),
	})
	ttl := s.cfg.SummaryTTLSeconds
	if ttl <= 0 {
		ttl = 86400
	}
	_ = clinicCache.SetEX(ctx, key, string(payload), time.Duration(ttl)*time.Second)
	return summary, nil
}
