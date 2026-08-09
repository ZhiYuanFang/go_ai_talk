package ucg

import (
	"context"
	"strings"
	"sync"
	"time"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/frame/g"
)

// UcgPolishProfileStore ucg-service 侧 polish lane profile 源。
type UcgPolishProfileStore struct {
	mu    sync.Mutex
	cache aimodel.Profile
}

// NewUcgPolishProfileStore 构造 polish profile 存储。
func NewUcgPolishProfileStore() *UcgPolishProfileStore {
	return &UcgPolishProfileStore{}
}

// Load 读取 polish lane 配置。
func (s *UcgPolishProfileStore) Load(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, error) {
	if lane != aimodel.LanePolish {
		return aimodel.Profile{}, aimodel.ErrProfileStoreUnset
	}
	s.mu.Lock()
	if s.cache.Lane != "" {
		p := s.cache
		s.mu.Unlock()
		return p, nil
	}
	s.mu.Unlock()

	if err := EnsureUcgAIConfigDefaultRow(ctx); err != nil {
		return aimodel.Profile{}, err
	}
	p, err := loadPolishProfileFresh(ctx)
	if err != nil {
		return aimodel.Profile{}, err
	}
	s.mu.Lock()
	s.cache = p
	s.mu.Unlock()
	return p, nil
}

// InvalidateCache Admin PUT 后失效 store 本地 cache；不得再调用 aimodel.InvalidateLaneCache（由调用方统一触发）。
func (s *UcgPolishProfileStore) InvalidateCache() {
	s.mu.Lock()
	s.cache = aimodel.Profile{}
	s.mu.Unlock()
}

type ucgAIConfigLaneRow struct {
	Id                  int    `json:"id"`
	VisionModel         string `json:"visionModel"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest"`
	Provider            string `json:"provider"`
	FreeProvider        string `json:"freeProvider"`
	FreeModel           string `json:"freeModel"`
	MaxInFlight         int    `json:"maxInFlight"`
	MaxWaiters          int    `json:"maxWaiters"`
	UpdatedAt           int64  `json:"updatedAt"`
	UpdatedBy           string `json:"updatedBy"`
}

// EnsureUcgAIConfigSchema 幂等补齐 free 列。
func EnsureUcgAIConfigSchema(ctx context.Context) error {
	for _, sql := range []string{
		`ALTER TABLE ucg_ai_config ADD COLUMN free_provider VARCHAR(32) NOT NULL DEFAULT ''`,
		`ALTER TABLE ucg_ai_config ADD COLUMN free_model VARCHAR(64) NOT NULL DEFAULT ''`,
	} {
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
				return err
			}
		}
	}
	return nil
}

// EnsureUcgAIConfigDefaultRow 保证 ucg_ai_config 单行存在（冷启动 env > 代码种子）。
func EnsureUcgAIConfigDefaultRow(ctx context.Context) error {
	if err := EnsureUcgAIConfigSchema(ctx); err != nil {
		return err
	}
	n, err := g.DB().Model("ucg_ai_config").Ctx(ctx).Where("id", aiConfigSingletonID).Count()
	if err != nil {
		return err
	}
	cold := aimodel.MergeColdStartProfile(aimodel.LanePolish, aimodel.Profile{}, false)
	now := time.Now().Unix()
	if n == 0 {
		_, err = g.DB().Model("ucg_ai_config").Ctx(ctx).Data(g.Map{
			"id":                     aiConfigSingletonID,
			"vision_model":           cold.Model,
			"max_images_per_request": 9,
			"provider":               string(cold.Provider),
			"max_in_flight":          cold.MaxInFlight,
			"max_waiters":            cold.MaxWaiters,
			"updated_at":             now,
			"updated_by":             "seed",
		}).Insert()
		return err
	}
	var row ucgAIConfigLaneRow
	if scanErr := g.DB().Model("ucg_ai_config").Ctx(ctx).Where("id", aiConfigSingletonID).Scan(&row); scanErr != nil {
		return nil
	}
	if !aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
		return nil
	}
	_, err = g.DB().Model("ucg_ai_config").Ctx(ctx).Where("id", aiConfigSingletonID).Data(g.Map{
		"provider":        string(cold.Provider),
		"vision_model":    cold.Model,
		"max_in_flight":   cold.MaxInFlight,
		"max_waiters":     cold.MaxWaiters,
		"updated_at":      now,
	}).Update()
	return err
}

func loadPolishProfileFresh(ctx context.Context) (aimodel.Profile, error) {
	cfg := LoadAIConfig(ctx)
	return aimodel.Profile{
		Lane:         aimodel.LanePolish,
		Provider:     aimodel.Provider(strings.TrimSpace(cfg.Provider)),
		Model:        cfg.VisionModel,
		FreeProvider: aimodel.Provider(strings.TrimSpace(cfg.FreeProvider)),
		FreeModel:    strings.TrimSpace(cfg.FreeModel),
		MaxInFlight:  cfg.MaxInFlight,
		MaxWaiters:   cfg.MaxWaiters,
		TimeoutSec:   cfg.TimeoutSeconds,
		UpdatedAt:    cfg.UpdatedAt,
		UpdatedBy:    cfg.UpdatedBy,
	}, nil
}

// InitUcgPolishProfileStore 在 ucg-service 启动时注册 aimodel ProfileStore。
func InitUcgPolishProfileStore() {
	aimodel.SetProfileStore(NewUcgPolishProfileStore())
}
