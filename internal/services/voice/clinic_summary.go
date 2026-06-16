package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
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

// clinicSummaryCache Redis voice:clinic:summary:{wxId}:{deviceNo} 持久化结构。
type clinicSummaryCache struct {
	Summary           string `json:"summary"`
	HistoryWatermark  int64  `json:"historyWatermark"`
	ComputedAt        int64  `json:"computedAt"`
}

func clinicHistoryCutoffUnix() int64 {
	return time.Now().Add(-7 * 24 * time.Hour).Unix()
}

func clinicSummaryRedisKey(wxID int64, deviceNo string) string {
	return fmt.Sprintf("%s%d:%s", clinicSummaryKeyPrefix, wxID, strings.TrimSpace(deviceNo))
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

// buildClinicHistorySummary 经 DeviceHistory HTTP 契约拉取 7 天数据并按 event 聚合。
func buildClinicHistorySummary(ctx context.Context, deviceNo string) (summaryJSON string, watermark int64, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "[]", 0, nil
	}
	rows, err := DeviceHistory().ListHistory(ctx, deviceNo)
	if err != nil {
		return "", 0, err
	}
	cutoff := clinicHistoryCutoffUnix()
	eventNames := loadEventNameAndUnitByID(ctx)
	byEvent := make(map[int64]*clinicEventAgg)
	for _, h := range rows {
		if h.StartTime < cutoff && h.EndTime < cutoff {
			continue
		}
		agg, ok := byEvent[h.EventId]
		if !ok {
			name := eventNames[h.EventId]
			if name == "" {
				name = strings.TrimSpace(h.EventName)
			}
			if name == "" {
				name = "未知事件"
			}
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
	}
	out := make([]clinicEventAgg, 0, len(byEvent))
	for _, v := range byEvent {
		out = append(out, *v)
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return "", 0, err
	}
	return string(raw), historyWatermarkFromRows(rows), nil
}

// ensureClinicSummary 懒刷新：对比 watermark，过期则重算并写 Redis（TTL 见 aiClinic.summaryTtlSeconds）。
func (s *ClinicService) ensureClinicSummary(ctx context.Context, wxID int64, deviceNo string) (string, error) {
	key := clinicSummaryRedisKey(wxID, deviceNo)
	v, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return "", err
	}
	var cached clinicSummaryCache
	if !v.IsNil() && !v.IsEmpty() {
		_ = json.Unmarshal(v.Bytes(), &cached)
	}
	// 拉 list 仅用于 watermark；与 build 内 list 重复但 MVP 可接受（design 允许后续聚合 API）。
	rows, err := DeviceHistory().ListHistory(ctx, deviceNo)
	if err != nil {
		if cached.Summary != "" {
			return cached.Summary, nil
		}
		return "", err
	}
	currentWM := historyWatermarkFromRows(rows)
	if cached.Summary != "" && cached.HistoryWatermark >= currentWM && currentWM > 0 {
		return cached.Summary, nil
	}
	if cached.Summary != "" && currentWM == 0 && cached.HistoryWatermark == 0 {
		return cached.Summary, nil
	}
	summary, wm, err := buildClinicHistorySummary(ctx, deviceNo)
	if err != nil {
		if cached.Summary != "" {
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
	_, _ = g.Redis().Do(ctx, "SET", key, string(payload), "EX", ttl)
	return summary, nil
}
