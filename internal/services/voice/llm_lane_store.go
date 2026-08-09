package voice

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

const llmLaneConfigTable = "llm_lane_config"

// VoiceLLMLaneStore voice-service 侧 understanding + clinic + careAlert profile 源（DB > YAML > 种子）。
type VoiceLLMLaneStore struct {
	mu    sync.Mutex
	cache map[aimodel.Lane]aimodel.Profile
}

// NewVoiceLLMLaneStore 构造 voice lane 配置存储。
func NewVoiceLLMLaneStore() *VoiceLLMLaneStore {
	return &VoiceLLMLaneStore{cache: make(map[aimodel.Lane]aimodel.Profile)}
}

// Load 读取单条 lane 配置。
func (s *VoiceLLMLaneStore) Load(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, error) {
	s.mu.Lock()
	if p, ok := s.cache[lane]; ok {
		s.mu.Unlock()
		return p, nil
	}
	s.mu.Unlock()

	if err := EnsureLLMLaneSchema(ctx); err != nil {
		return aimodel.Profile{}, err
	}
	if err := EnsureLLMLaneDefaultRows(ctx); err != nil {
		return aimodel.Profile{}, err
	}
	p, err := loadLLMLaneProfile(ctx, lane)
	if err != nil {
		return aimodel.Profile{}, err
	}
	s.mu.Lock()
	s.cache[lane] = p
	s.mu.Unlock()
	return p, nil
}

// InvalidateCache 清空 store 内缓存；不得再调用 aimodel.InvalidateLaneCache（由调用方统一触发）。
func (s *VoiceLLMLaneStore) InvalidateCache() {
	s.mu.Lock()
	s.cache = make(map[aimodel.Lane]aimodel.Profile)
	s.mu.Unlock()
}

type llmLaneRow struct {
	Lane         string `json:"lane"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	FreeProvider string `json:"freeProvider"`
	FreeModel    string `json:"freeModel"`
	MaxInFlight  int    `json:"maxInFlight"`
	MaxWaiters   int    `json:"maxWaiters"`
	UpdatedAt    int64  `json:"updatedAt"`
	UpdatedBy    string `json:"updatedBy"`
}

// EnsureLLMLaneSchema 幂等补齐 free 列（重复执行忽略 Duplicate column）。
func EnsureLLMLaneSchema(ctx context.Context) error {
	alters := []string{
		`ALTER TABLE llm_lane_config ADD COLUMN free_provider VARCHAR(32) NOT NULL DEFAULT ''`,
		`ALTER TABLE llm_lane_config ADD COLUMN free_model VARCHAR(64) NOT NULL DEFAULT ''`,
	}
	for _, sql := range alters {
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "Duplicate column") && !strings.Contains(msg, "1060") {
				return err
			}
		}
	}
	return nil
}

var voiceLLMLanes = []aimodel.Lane{
	aimodel.LaneVoiceUnderstanding,
	aimodel.LaneClinic,
	aimodel.LaneCareAlert,
}

// EnsureLLMLaneDefaultRows 保证 llm_lane_config 表存在种子行；种子行在启动时按 env 刷新 provider/model。
func EnsureLLMLaneDefaultRows(ctx context.Context) error {
	if err := EnsureLLMLaneSchema(ctx); err != nil {
		return err
	}
	for _, lane := range voiceLLMLanes {
		yamlP, yamlOK := loadLLMLaneYAMLProfile(ctx, lane)
		cold := aimodel.MergeColdStartProfile(lane, yamlP, yamlOK)
		n, err := g.DB().Model(llmLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Count()
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		if n == 0 {
			_, err = g.DB().Model(llmLaneConfigTable).Ctx(ctx).Data(g.Map{
				"lane":          string(lane),
				"provider":      string(cold.Provider),
				"model":         cold.Model,
				"free_provider": "",
				"free_model":    "",
				"max_in_flight": cold.MaxInFlight,
				"max_waiters":   cold.MaxWaiters,
				"updated_at":    now,
				"updated_by":    "seed",
			}).Insert()
			if err != nil {
				return err
			}
			continue
		}
		var row llmLaneRow
		if scanErr := g.DB().Model(llmLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Scan(&row); scanErr != nil {
			continue
		}
		if !aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
			continue
		}
		_, _ = g.DB().Model(llmLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Data(g.Map{
			"provider":      string(cold.Provider),
			"model":         cold.Model,
			"max_in_flight": cold.MaxInFlight,
			"max_waiters":   cold.MaxWaiters,
			"updated_at":    now,
		}).Update()
	}
	return nil
}

func loadLLMLaneProfile(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, error) {
	var row llmLaneRow
	err := g.DB().Model(llmLaneConfigTable).Ctx(ctx).Where("lane", string(lane)).Scan(&row)
	if err == nil && strings.TrimSpace(row.Lane) != "" {
		if !aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
			return rowToProfile(lane, row), nil
		}
		yamlP, yamlOK := loadLLMLaneYAMLProfile(ctx, lane)
		cold := aimodel.MergeColdStartProfile(lane, yamlP, yamlOK)
		p := rowToProfile(lane, row)
		p.Provider = cold.Provider
		p.Model = cold.Model
		if aimodel.IsSeedUpdatedBy(row.UpdatedBy) {
			p.MaxInFlight = cold.MaxInFlight
			p.MaxWaiters = cold.MaxWaiters
			p.TimeoutSec = cold.TimeoutSec
		}
		return p, nil
	}
	yamlP, yamlOK := loadLLMLaneYAMLProfile(ctx, lane)
	return aimodel.MergeColdStartProfile(lane, yamlP, yamlOK), nil
}

func rowToProfile(lane aimodel.Lane, row llmLaneRow) aimodel.Profile {
	p := aimodel.Profile{
		Lane:         lane,
		Provider:     aimodel.Provider(strings.TrimSpace(row.Provider)),
		Model:        row.Model,
		FreeProvider: aimodel.Provider(strings.TrimSpace(row.FreeProvider)),
		FreeModel:    strings.TrimSpace(row.FreeModel),
		MaxInFlight:  row.MaxInFlight,
		MaxWaiters:   row.MaxWaiters,
		UpdatedAt:    row.UpdatedAt,
		UpdatedBy:    row.UpdatedBy,
	}
	if p.Provider == "" {
		p.Provider = aimodel.DefaultSeedProfile(lane).Provider
	}
	if strings.TrimSpace(p.Model) == "" {
		p.Model = aimodel.DefaultSeedProfile(lane).Model
	}
	if p.MaxInFlight <= 0 {
		p.MaxInFlight = aimodel.DefaultSeedProfile(lane).MaxInFlight
	}
	if p.MaxWaiters < 0 {
		p.MaxWaiters = aimodel.DefaultSeedProfile(lane).MaxWaiters
	}
	if p.TimeoutSec <= 0 {
		p.TimeoutSec = aimodel.DefaultSeedProfile(lane).TimeoutSec
	}
	return p
}

func loadLLMLaneYAMLProfile(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, bool) {
	rel := strings.TrimSpace(g.Cfg().MustGet(ctx, "voiceChat.llmLanesFile").String())
	if rel == "" {
		rel = defaultVoiceChatSharedRel
	}
	path, err := gfile.Search(rel)
	if err != nil || path == "" {
		return aimodel.Profile{}, false
	}
	j, err := gjson.Load(path)
	if err != nil {
		return aimodel.Profile{}, false
	}
	block := j.Get("voiceChat.llmLanes." + string(lane))
	if block == nil || block.IsNil() {
		return aimodel.Profile{}, false
	}
	var raw struct {
		Provider    string `json:"provider"`
		Model       string `json:"model"`
		MaxInFlight int    `json:"maxInFlight"`
		MaxWaiters  int    `json:"maxWaiters"`
		TimeoutSec  int    `json:"timeoutSec"`
	}
	if err := block.Scan(&raw); err != nil {
		return aimodel.Profile{}, false
	}
	if strings.TrimSpace(raw.Model) == "" {
		return aimodel.Profile{}, false
	}
	return aimodel.Profile{
		Lane:        lane,
		Provider:    aimodel.Provider(strings.TrimSpace(raw.Provider)),
		Model:       raw.Model,
		MaxInFlight: raw.MaxInFlight,
		MaxWaiters:  raw.MaxWaiters,
		TimeoutSec:  raw.TimeoutSec,
	}, true
}

// LLMLanesAdminDTO Admin GET 响应。
type LLMLanesAdminDTO struct {
	VoiceUnderstanding aimodel.LaneProfileDTO `json:"voiceUnderstanding"`
	Clinic             aimodel.LaneProfileDTO `json:"clinic"`
	CareAlert          aimodel.LaneProfileDTO `json:"careAlert"`
	Allowlist          map[string][]string    `json:"allowlist"`
}

// GetLLMLanesForAdmin 读取三条 voice lane 当前配置。
func GetLLMLanesForAdmin(ctx context.Context) (LLMLanesAdminDTO, error) {
	store := NewVoiceLLMLaneStore()
	vu, err := store.Load(ctx, aimodel.LaneVoiceUnderstanding)
	if err != nil {
		return LLMLanesAdminDTO{}, err
	}
	cl, err := store.Load(ctx, aimodel.LaneClinic)
	if err != nil {
		return LLMLanesAdminDTO{}, err
	}
	ca, err := store.Load(ctx, aimodel.LaneCareAlert)
	if err != nil {
		return LLMLanesAdminDTO{}, err
	}
	return LLMLanesAdminDTO{
		VoiceUnderstanding: profileToDTO(vu),
		Clinic:             profileToDTO(cl),
		CareAlert:          profileToDTO(ca),
		Allowlist:          buildProviderAllowlist(),
	}, nil
}

func profileToDTO(p aimodel.Profile) aimodel.LaneProfileDTO {
	return aimodel.LaneProfileDTO{
		Provider:     string(p.Provider),
		Model:        p.Model,
		FreeProvider: string(p.FreeProvider),
		FreeModel:    p.FreeModel,
		MaxInFlight:  p.MaxInFlight,
		MaxWaiters:   p.MaxWaiters,
		UpdatedAt:    p.UpdatedAt,
		UpdatedBy:    p.UpdatedBy,
	}
}

func buildProviderAllowlist() map[string][]string {
	out := make(map[string][]string, len(aimodel.ProviderModels))
	for p, models := range aimodel.ProviderModels {
		out[string(p)] = models
	}
	return out
}

// UpdateLLMLanesForAdmin 校验并持久化 Admin PUT（含 careAlert 与 free）。
func UpdateLLMLanesForAdmin(ctx context.Context, vu, clinic, careAlert aimodel.LaneProfileDTO, updatedBy string) error {
	if err := validateLaneDTO(aimodel.LaneVoiceUnderstanding, vu); err != nil {
		return err
	}
	if err := validateLaneDTO(aimodel.LaneClinic, clinic); err != nil {
		return err
	}
	if err := validateLaneDTO(aimodel.LaneCareAlert, careAlert); err != nil {
		return err
	}
	now := time.Now().Unix()
	operator := strings.TrimSpace(updatedBy)
	if operator == "" {
		operator = "admin"
	}
	if err := EnsureLLMLaneSchema(ctx); err != nil {
		return err
	}
	if err := upsertLLMLaneRow(ctx, aimodel.LaneVoiceUnderstanding, vu, now, operator); err != nil {
		return err
	}
	if err := upsertLLMLaneRow(ctx, aimodel.LaneClinic, clinic, now, operator); err != nil {
		return err
	}
	if err := upsertLLMLaneRow(ctx, aimodel.LaneCareAlert, careAlert, now, operator); err != nil {
		return err
	}
	NewVoiceLLMLaneStore().InvalidateCache()
	aimodel.InvalidateLaneCache()
	return nil
}

func validateLaneDTO(lane aimodel.Lane, dto aimodel.LaneProfileDTO) error {
	provider := aimodel.Provider(strings.TrimSpace(dto.Provider))
	if provider == "" {
		return fmt.Errorf("%s: provider 不能为空", lane)
	}
	if !aimodel.IsAllowedModel(provider, dto.Model) {
		return fmt.Errorf("%s: model 不在 allowlist", lane)
	}
	fp := strings.TrimSpace(dto.FreeProvider)
	fm := strings.TrimSpace(dto.FreeModel)
	if fp != "" || fm != "" {
		if fp == "" || fm == "" {
			return fmt.Errorf("%s: freeProvider 与 freeModel 须同时为空或同时非空", lane)
		}
		if !aimodel.IsAllowedModel(aimodel.Provider(fp), fm) {
			return fmt.Errorf("%s: free model 不在 allowlist", lane)
		}
	}
	if dto.MaxInFlight < 1 {
		return fmt.Errorf("%s: maxInFlight 须 >= 1", lane)
	}
	if dto.MaxWaiters < 0 {
		return fmt.Errorf("%s: maxWaiters 须 >= 0", lane)
	}
	return nil
}

func upsertLLMLaneRow(ctx context.Context, lane aimodel.Lane, dto aimodel.LaneProfileDTO, now int64, updatedBy string) error {
	_, err := g.DB().Model(llmLaneConfigTable).Ctx(ctx).Data(g.Map{
		"lane":          string(lane),
		"provider":      strings.TrimSpace(dto.Provider),
		"model":         strings.TrimSpace(dto.Model),
		"free_provider": strings.TrimSpace(dto.FreeProvider),
		"free_model":    strings.TrimSpace(dto.FreeModel),
		"max_in_flight": dto.MaxInFlight,
		"max_waiters":   dto.MaxWaiters,
		"updated_at":    now,
		"updated_by":    updatedBy,
	}).Save()
	return err
}

// InitVoiceLLMProfileStore 在 voice-service 启动时注册 aimodel ProfileStore。
func InitVoiceLLMProfileStore() {
	aimodel.SetProfileStore(NewVoiceLLMLaneStore())
}
