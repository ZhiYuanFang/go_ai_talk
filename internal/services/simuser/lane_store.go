package simuser

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/frame/g"
)

const simLLMLaneConfigTable = "sim_llm_lane_config"

// SimLLMLaneStore sim-user-service 侧 aimodel ProfileStore（DB > env > 代码种子）。
type SimLLMLaneStore struct {
	mu    sync.Mutex
	cache map[aimodel.Lane]aimodel.Profile
}

func NewSimLLMLaneStore() *SimLLMLaneStore {
	return &SimLLMLaneStore{cache: make(map[aimodel.Lane]aimodel.Profile)}
}

func (s *SimLLMLaneStore) Load(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, error) {
	s.mu.Lock()
	if p, ok := s.cache[lane]; ok {
		s.mu.Unlock()
		return p, nil
	}
	s.mu.Unlock()

	if err := EnsureSimLLMLaneDefaultRows(ctx); err != nil {
		return aimodel.Profile{}, err
	}
	p, err := loadSimLLMLaneProfile(ctx, lane)
	if err != nil {
		return aimodel.Profile{}, err
	}
	s.mu.Lock()
	s.cache[lane] = p
	s.mu.Unlock()
	return p, nil
}

func (s *SimLLMLaneStore) InvalidateCache() {
	s.mu.Lock()
	s.cache = make(map[aimodel.Lane]aimodel.Profile)
	s.mu.Unlock()
}

type simLLMLaneRow struct {
	Lane        string `json:"lane"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	MaxInFlight int    `json:"maxInFlight"`
	MaxWaiters  int    `json:"maxWaiters"`
	TimeoutSec  int    `json:"timeoutSec"`
	UpdatedAt   int64  `json:"updatedAt"`
	UpdatedBy   string `json:"updatedBy"`
}

func allSimLLMLanes() []aimodel.Lane {
	return []aimodel.Lane{
		aimodel.LaneSimText, aimodel.LaneSimVision,
		aimodel.LaneSimImageGen, aimodel.LaneSimVideoGen,
	}
}

// EnsureSimLLMLaneDefaultRows 保证四 lane 种子行存在；seed 行启动时按 env 刷新 provider/model。
func EnsureSimLLMLaneDefaultRows(ctx context.Context) error {
	for _, lane := range allSimLLMLanes() {
		cold := aimodel.MergeColdStartProfile(lane, aimodel.Profile{}, false)
		n, err := g.DB().Model(simLLMLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Count()
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		if n == 0 {
			_, err = g.DB().Model(simLLMLaneConfigTable).Ctx(ctx).Data(g.Map{
				"lane":          string(lane),
				"provider":      string(cold.Provider),
				"model":         cold.Model,
				"max_in_flight": cold.MaxInFlight,
				"max_waiters":   cold.MaxWaiters,
				"timeout_sec":   cold.TimeoutSec,
				"updated_at":    now,
				"updated_by":    "seed",
			}).Insert()
			if err != nil {
				return err
			}
			continue
		}
		var row simLLMLaneRow
		if scanErr := g.DB().Model(simLLMLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Scan(&row); scanErr != nil {
			continue
		}
		if !aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
			continue
		}
		_, _ = g.DB().Model(simLLMLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Data(g.Map{
			"provider":      string(cold.Provider),
			"model":         cold.Model,
			"max_in_flight": cold.MaxInFlight,
			"max_waiters":   cold.MaxWaiters,
			"timeout_sec":   cold.TimeoutSec,
			"updated_at":    now,
		}).Update()
	}
	return nil
}

func loadSimLLMLaneProfile(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, error) {
	var row simLLMLaneRow
	err := g.DB().Model(simLLMLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Scan(&row)
	if err == nil && strings.TrimSpace(row.Lane) != "" {
		if !aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
			return simRowToProfile(lane, row), nil
		}
		cold := aimodel.MergeColdStartProfile(lane, aimodel.Profile{}, false)
		p := simRowToProfile(lane, row)
		p.Provider = cold.Provider
		p.Model = cold.Model
		if aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
			p.MaxInFlight = cold.MaxInFlight
			p.MaxWaiters = cold.MaxWaiters
			p.TimeoutSec = cold.TimeoutSec
		}
		return p, nil
	}
	return aimodel.MergeColdStartProfile(lane, aimodel.Profile{}, false), nil
}

func simRowToProfile(lane aimodel.Lane, row simLLMLaneRow) aimodel.Profile {
	p := aimodel.Profile{
		Lane:        lane,
		Provider:    aimodel.Provider(strings.TrimSpace(row.Provider)),
		Model:       row.Model,
		MaxInFlight: row.MaxInFlight,
		MaxWaiters:  row.MaxWaiters,
		TimeoutSec:  row.TimeoutSec,
		UpdatedAt:   row.UpdatedAt,
		UpdatedBy:   row.UpdatedBy,
	}
	def := aimodel.DefaultSeedProfile(lane)
	if p.Provider == "" {
		p.Provider = def.Provider
	}
	if strings.TrimSpace(p.Model) == "" {
		p.Model = def.Model
	}
	if p.MaxInFlight <= 0 {
		p.MaxInFlight = def.MaxInFlight
	}
	if p.MaxWaiters < 0 {
		p.MaxWaiters = def.MaxWaiters
	}
	if p.TimeoutSec <= 0 {
		p.TimeoutSec = def.TimeoutSec
	}
	return p
}

// SimLLMLanesAdminDTO Admin GET 响应。
type SimLLMLanesAdminDTO struct {
	SimText      aimodel.LaneProfileDTO `json:"simText"`
	SimVision    aimodel.LaneProfileDTO `json:"simVision"`
	SimImageGen  aimodel.LaneProfileDTO `json:"simImageGen"`
	SimVideoGen  aimodel.LaneProfileDTO `json:"simVideoGen"`
	Allowlist    map[string][]string    `json:"allowlist"`
}

// GetSimLLMLanesForAdmin 读取四 lane 当前配置。
func GetSimLLMLanesForAdmin(ctx context.Context) (SimLLMLanesAdminDTO, error) {
	store := NewSimLLMLaneStore()
	st, err := store.Load(ctx, aimodel.LaneSimText)
	if err != nil {
		return SimLLMLanesAdminDTO{}, err
	}
	sv, err := store.Load(ctx, aimodel.LaneSimVision)
	if err != nil {
		return SimLLMLanesAdminDTO{}, err
	}
	si, err := store.Load(ctx, aimodel.LaneSimImageGen)
	if err != nil {
		return SimLLMLanesAdminDTO{}, err
	}
	sv2, err := store.Load(ctx, aimodel.LaneSimVideoGen)
	if err != nil {
		return SimLLMLanesAdminDTO{}, err
	}
	return SimLLMLanesAdminDTO{
		SimText:     simProfileToDTO(st),
		SimVision:   simProfileToDTO(sv),
		SimImageGen: simProfileToDTO(si),
		SimVideoGen: simProfileToDTO(sv2),
		Allowlist:   simBuildProviderAllowlist(),
	}, nil
}

func simProfileToDTO(p aimodel.Profile) aimodel.LaneProfileDTO {
	return aimodel.LaneProfileDTO{
		Provider: string(p.Provider), Model: p.Model,
		MaxInFlight: p.MaxInFlight, MaxWaiters: p.MaxWaiters,
		UpdatedAt: p.UpdatedAt, UpdatedBy: p.UpdatedBy,
	}
}

func simBuildProviderAllowlist() map[string][]string {
	out := make(map[string][]string, len(aimodel.ProviderModels))
	for p, models := range aimodel.ProviderModels {
		out[string(p)] = models
	}
	return out
}

// UpdateSimLLMLanesForAdmin 校验并持久化 Admin PUT；不触发 scheduler reload。
func UpdateSimLLMLanesForAdmin(ctx context.Context, text, vision, imageGen, videoGen aimodel.LaneProfileDTO, updatedBy string) error {
	if err := validateSimLaneDTO(aimodel.LaneSimText, text); err != nil {
		return err
	}
	if err := validateSimLaneDTO(aimodel.LaneSimVision, vision); err != nil {
		return err
	}
	if err := validateSimLaneDTO(aimodel.LaneSimImageGen, imageGen); err != nil {
		return err
	}
	if err := validateSimLaneDTO(aimodel.LaneSimVideoGen, videoGen); err != nil {
		return err
	}
	now := time.Now().Unix()
	operator := strings.TrimSpace(updatedBy)
	if operator == "" {
		operator = "admin"
	}
	for lane, dto := range map[aimodel.Lane]aimodel.LaneProfileDTO{
		aimodel.LaneSimText: text, aimodel.LaneSimVision: vision,
		aimodel.LaneSimImageGen: imageGen, aimodel.LaneSimVideoGen: videoGen,
	} {
		if err := upsertSimLLMLaneRow(ctx, lane, dto, now, operator); err != nil {
			return err
		}
	}
	NewSimLLMLaneStore().InvalidateCache()
	aimodel.InvalidateLaneCache()
	return nil
}

func validateSimLaneDTO(lane aimodel.Lane, dto aimodel.LaneProfileDTO) error {
	provider := aimodel.Provider(strings.TrimSpace(dto.Provider))
	if provider == "" {
		return fmt.Errorf("%s: provider 不能为空", lane)
	}
	if !aimodel.IsAllowedModel(provider, dto.Model) {
		return fmt.Errorf("%s: model 不在 allowlist", lane)
	}
	if dto.MaxInFlight < 1 {
		return fmt.Errorf("%s: maxInFlight 须 >= 1", lane)
	}
	if dto.MaxWaiters < 0 {
		return fmt.Errorf("%s: maxWaiters 须 >= 0", lane)
	}
	return nil
}

func upsertSimLLMLaneRow(ctx context.Context, lane aimodel.Lane, dto aimodel.LaneProfileDTO, now int64, updatedBy string) error {
	_, err := g.DB().Model(simLLMLaneConfigTable).Ctx(ctx).Data(g.Map{
		"lane":          string(lane),
		"provider":      strings.TrimSpace(dto.Provider),
		"model":         strings.TrimSpace(dto.Model),
		"max_in_flight": dto.MaxInFlight,
		"max_waiters":   dto.MaxWaiters,
		"updated_at":    now,
		"updated_by":    updatedBy,
	}).Save()
	return err
}

// LoadAllLaneProfiles 读取四条 sim lane 当前生效 profile（供 runtime 展示）。
func LoadAllLaneProfiles() map[string]aimodel.LaneProfileDTO {
	out := make(map[string]aimodel.LaneProfileDTO, 4)
	store := NewSimLLMLaneStore()
	ctx := context.Background()
	for _, lane := range allSimLLMLanes() {
		p, err := store.Load(ctx, lane)
		if err != nil {
			continue
		}
		out[string(lane)] = simProfileToDTO(p)
	}
	return out
}

// InitAIModel 注册 ProfileStore。
func InitAIModel() {
	aimodel.SetProfileStore(NewSimLLMLaneStore())
}
