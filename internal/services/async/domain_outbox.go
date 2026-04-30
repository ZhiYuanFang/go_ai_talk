package async

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"hello/internal/dao"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"
	"hello/internal/services/contracts"
	"hello/internal/services/device"
	"hello/internal/services/history"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	outboxEnabledEnv      = "OUTBOX_RELAY_ENABLED"
	outboxWorkersEnv      = "OUTBOX_RELAY_WORKERS"
	outboxPollMsEnv       = "OUTBOX_RELAY_POLL_INTERVAL_MS"
	outboxMaxAttemptsEnv  = "OUTBOX_RELAY_MAX_ATTEMPTS"
	outboxTableName       = "domain_outbox"
	outboxStatusPending   = "pending"
	outboxStatusPublished = "published"
	outboxStatusFailed    = "failed"
	outboxProjectionDonePrefix = "outbox:projection:done:"
	projectionRepairEnabledEnv = "CACHE_PROJECTION_REPAIR_ENABLED"
	projectionRepairPollMsEnv  = "CACHE_PROJECTION_REPAIR_POLL_INTERVAL_MS"
	projectionRepairSampleEnv  = "CACHE_PROJECTION_REPAIR_SAMPLE_LIMIT"
)

type outboxStatus string

const (
	outboxStatusPendingEnum   outboxStatus = outboxStatusPending
	outboxStatusPublishedEnum outboxStatus = outboxStatusPublished
	outboxStatusFailedEnum    outboxStatus = outboxStatusFailed
)

var outboxRelayOnce sync.Once
var projectionRepairOnce sync.Once

type domainOutboxRow struct {
	EventID    string `json:"event_id"`
	Id         int64  `json:"id"`
	RoutingKey string `json:"routing_key"`
	Payload    string `json:"payload"`
	Attempts   int    `json:"attempts"`
}

func isOutboxRelayEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(outboxEnabledEnv)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// StartDomainOutboxRelay 启动 outbox 异步中继。
func StartDomainOutboxRelay(ctx context.Context) {
	if !isOutboxRelayEnabled() {
		return
	}
	outboxRelayOnce.Do(func() {
		workers := envIntOrDefault(outboxWorkersEnv, 1)
		if workers <= 0 {
			workers = 1
		}
		pollMs := envIntOrDefault(outboxPollMsEnv, 1500)
		if pollMs < 100 {
			pollMs = 100
		}
		for i := 0; i < workers; i++ {
			workerID := i + 1
			go runOutboxRelayWorker(ctx, workerID, time.Duration(pollMs)*time.Millisecond)
		}
		startProjectionRepairWorker(ctx)
	})
}

func runOutboxRelayWorker(ctx context.Context, workerID int, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := relayOneOutboxEvent(ctx); err != nil {
				glog.Warningf(ctx, "outbox relay worker %d failed: %v", workerID, err)
			}
		}
	}
}

func relayOneOutboxEvent(ctx context.Context) error {
	group := dao.History.Group()
	maxAttempts := envIntOrDefault(outboxMaxAttemptsEnv, 10)
	if maxAttempts <= 0 {
		maxAttempts = 10
	}
	return g.DB(group).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		row, err := tx.Model(outboxTableName).
			WhereIn("status", []string{outboxStatusPending, outboxStatusFailed}).
			Where("attempts<?", maxAttempts).
			OrderAsc("id").
			Limit(1).
			One()
		if err != nil {
			return err
		}
		if row == nil || row.IsEmpty() {
			return nil
		}
		item := domainOutboxRow{
			EventID:    row["event_id"].String(),
			Id:         row["id"].Int64(),
			RoutingKey: row["routing_key"].String(),
			Payload:    row["payload"].String(),
			Attempts:   row["attempts"].Int(),
		}
		routingKey, ok := contracts.ParseRoutingKey(ctx, item.RoutingKey, "outbox_relay")
		if !ok || strings.TrimSpace(item.Payload) == "" {
			_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
				"status":     outboxStatusFailedEnum,
				"attempts":   item.Attempts + 1,
				"last_error": "invalid outbox payload",
			}).Update()
			return nil
		}
		item.RoutingKey = routingKey.String()
		if err := applyCacheProjectionIfNeeded(ctx, item); err != nil {
			_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
				"status":       outboxStatusFailedEnum,
				"attempts":     item.Attempts + 1,
				"last_error":   truncateError(err.Error(), 512),
				"published_at": nil,
			}).Update()
			return nil
		}
		if err := publishMQEvent(ctx, item.RoutingKey, item.Payload); err != nil {
			_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
				"status":       outboxStatusFailedEnum,
				"attempts":     item.Attempts + 1,
				"last_error":   truncateError(err.Error(), 512),
				"published_at": nil,
			}).Update()
			return nil
		}
		_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
			"status":       outboxStatusPublishedEnum,
			"attempts":     item.Attempts + 1,
			"last_error":   "",
			"published_at": gdb.Raw("NOW()"),
		}).Update()
		return nil
	})
}

func applyCacheProjectionIfNeeded(ctx context.Context, item domainOutboxRow) error {
	route := strings.TrimSpace(item.RoutingKey)
	routingKey, ok := contracts.ParseRoutingKey(ctx, route, "cache_projection")
	if !ok {
		return fmt.Errorf("invalid routing key: %s", route)
	}
	cache := cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})
	doneKey := outboxProjectionDonePrefix + strings.TrimSpace(item.EventID)
	if done, err := cache.Exists(ctx, doneKey); err == nil && done {
		return nil
	}
	switch {
	case routingKey.HasPrefix(eventkit.RoutingPrefixHistoryRecord):
		if err := history.ApplyProjection(ctx, route, item.Payload); err != nil {
			_ = scheduleProjectionRepair(ctx, route, item.Payload)
			return err
		}
	case routingKey.HasPrefix(eventkit.RoutingPrefixDevice):
		if err := device.ApplyProjection(ctx, route, item.Payload); err != nil {
			_ = scheduleProjectionRepair(ctx, route, item.Payload)
			return err
		}
	case routingKey.HasPrefix(eventkit.RoutingPrefixVoiceTask):
		// voice.task.* 当前无缓存投影处理器，仅保留分组入口用于后续扩展。
		return nil
	default:
		glog.Warningf(ctx, "metric=unmapped_routing_prefix source=cache_projection routingKey=%s", route)
		return fmt.Errorf("unmapped routing prefix: %s", route)
	}
	return cache.SetEX(ctx, doneKey, "1", 24*time.Hour)
}

func startProjectionRepairWorker(ctx context.Context) {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(projectionRepairEnabledEnv)))
	enabled := v == "1" || v == "true" || v == "yes" || v == "on"
	if !enabled {
		return
	}
	projectionRepairOnce.Do(func() {
		pollMs := envIntOrDefault(projectionRepairPollMsEnv, 30000)
		if pollMs < 1000 {
			pollMs = 1000
		}
		ticker := time.NewTicker(time.Duration(pollMs) * time.Millisecond)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					limit := envIntOrDefault(projectionRepairSampleEnv, 20)
					if limit <= 0 {
						limit = 20
					}
					if err := runProjectionReconcileOnce(ctx, limit); err != nil {
						glog.Warningf(ctx, "projection reconcile failed: %v", err)
					}
				}
			}
		}()
	})
}

func runProjectionReconcileOnce(ctx context.Context, limit int) error {
	rows, err := g.DB(dao.History.Group()).Model(dao.History.Table()).
		Fields(dao.History.Columns().DeviceNo).
		Where(dao.History.Columns().DeviceNo+" <> ''").
		OrderDesc(dao.History.Columns().Id).
		Limit(limit).
		All()
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		deviceNo := strings.TrimSpace(row[dao.History.Columns().DeviceNo].String())
		if deviceNo == "" {
			continue
		}
		if _, ok := seen[deviceNo]; ok {
			continue
		}
		seen[deviceNo] = struct{}{}
		if err = history.RebuildHistoryCacheByDevice(ctx, deviceNo); err != nil {
			glog.Warningf(ctx, "history cache rebuild failed: deviceNo=%s err=%v", deviceNo, err)
		}
		if err = history.RebuildBirthdayCacheByDevice(ctx, deviceNo); err != nil {
			glog.Warningf(ctx, "birthday cache rebuild failed: deviceNo=%s err=%v", deviceNo, err)
		}
		if err = device.RebuildUserProfileCacheByDevice(ctx, deviceNo); err != nil {
			glog.Warningf(ctx, "profile cache rebuild failed: deviceNo=%s err=%v", deviceNo, err)
		}
	}
	_ = history.RebuildHistoryMetaCache(ctx)
	_ = device.RebuildEventCache(ctx)
	_ = device.RebuildActionCache(ctx)
	return nil
}

func scheduleProjectionRepair(ctx context.Context, routingKey, payload string) error {
	if _, err := contracts.ParseRoutingKeyCompat(ctx, routingKey, "projection_repair"); err != nil {
		return err
	}
	cache := cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})
	key, err := cachekit.Key(cachekit.DomainSystem, "projection", "repair", fmt.Sprintf("%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{
		"routing_key": strings.TrimSpace(routingKey),
		"payload":     strings.TrimSpace(payload),
	})
	if err = cache.SetEX(ctx, key, string(body), 10*time.Minute); err != nil {
		return err
	}
	return nil
}

func publishMQEvent(ctx context.Context, routingKey, payload string) error {
	pub, err := newObservedEventPublisher()
	if err != nil {
		return err
	}
	return pub.Publish(ctx, routingKey, json.RawMessage(payload))
}

func truncateError(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

func InsertOutboxEventTx(tx gdb.TX, routingKey eventkit.RouteKey, payload map[string]interface{}) error {
	if !routingKey.IsValid() {
		return fmt.Errorf("invalid routing key: %s", routingKey.String())
	}
	eventID := fmt.Sprintf("outbox-%d", time.Now().UnixNano())
	body, _ := json.Marshal(payload)
	_, err := tx.Model(outboxTableName).Data(g.Map{
		"event_id":    eventID,
		"event_type":  routingKey.String(),
		"routing_key": routingKey.String(),
		"payload":     string(body),
		"status":      outboxStatusPendingEnum,
		"attempts":    0,
		"last_error":  "",
	}).Insert()
	return err
}
