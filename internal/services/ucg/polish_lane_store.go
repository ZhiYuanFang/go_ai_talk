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

// InvalidateCache Admin PUT 后失效。
func (s *UcgPolishProfileStore) InvalidateCache() {
	s.mu.Lock()
	s.cache = aimodel.Profile{}
	s.mu.Unlock()
	aimodel.InvalidateLaneCache()
}

type ucgAIConfigLaneRow struct {
	Id                  int    `json:"id"`
	VisionModel         string `json:"visionModel"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest"`
	Provider            string `json:"provider"`
	MaxInFlight         int    `json:"maxInFlight"`
	MaxWaiters          int    `json:"maxWaiters"`
	UpdatedAt           int64  `json:"updatedAt"`
	UpdatedBy           string `json:"updatedBy"`
}

// EnsureUcgAIConfigDefaultRow 保证 ucg_ai_config 单行存在（种子 A polish）。
func EnsureUcgAIConfigDefaultRow(ctx context.Context) error {
	n, err := g.DB().Model("ucg_ai_config").Ctx(ctx).Where("id", aiConfigSingletonID).Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	seed := aimodel.DefaultSeedProfile(aimodel.LanePolish)
	now := time.Now().Unix()
	_, err = g.DB().Model("ucg_ai_config").Ctx(ctx).Data(g.Map{
		"id":                     aiConfigSingletonID,
		"vision_model":           seed.Model,
		"max_images_per_request": 9,
		"provider":               string(seed.Provider),
		"max_in_flight":          seed.MaxInFlight,
		"max_waiters":            seed.MaxWaiters,
		"updated_at":             now,
		"updated_by":             "seed",
	}).Insert()
	return err
}

func loadPolishProfileFresh(ctx context.Context) (aimodel.Profile, error) {
	cfg := LoadAIConfig(ctx)
	return aimodel.Profile{
		Lane:        aimodel.LanePolish,
		Provider:    aimodel.Provider(strings.TrimSpace(cfg.Provider)),
		Model:       cfg.VisionModel,
		MaxInFlight: cfg.MaxInFlight,
		MaxWaiters:  cfg.MaxWaiters,
		TimeoutSec:  cfg.TimeoutSeconds,
		UpdatedAt:   cfg.UpdatedAt,
		UpdatedBy:   cfg.UpdatedBy,
	}, nil
}

// InitUcgPolishProfileStore 在 ucg-service 启动时注册 aimodel ProfileStore。
func InitUcgPolishProfileStore() {
	aimodel.SetProfileStore(NewUcgPolishProfileStore())
}
