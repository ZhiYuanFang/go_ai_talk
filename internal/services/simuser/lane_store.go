package simuser

import (
	"context"
	"sync"

	"hello/internal/services/aimodel"
)

// SimLLMLaneStore sim-user-service 侧 aimodel ProfileStore（env > 代码种子；无 DB/Admin）。
type SimLLMLaneStore struct {
	mu    sync.Mutex
	cache map[aimodel.Lane]aimodel.Profile
}

func NewSimLLMLaneStore() *SimLLMLaneStore {
	return &SimLLMLaneStore{cache: make(map[aimodel.Lane]aimodel.Profile)}
}

func (s *SimLLMLaneStore) Load(ctx context.Context, lane aimodel.Lane) (aimodel.Profile, error) {
	_ = ctx
	s.mu.Lock()
	if p, ok := s.cache[lane]; ok {
		s.mu.Unlock()
		return p, nil
	}
	s.mu.Unlock()
	p := aimodel.MergeColdStartProfile(lane, aimodel.Profile{}, false)
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

// LoadAllLaneProfiles 读取四条 sim lane 当前生效 profile（供 runtime 只读展示）。
func LoadAllLaneProfiles() map[string]aimodel.LaneProfileDTO {
	lanes := []aimodel.Lane{
		aimodel.LaneSimText, aimodel.LaneSimVision,
		aimodel.LaneSimImageGen, aimodel.LaneSimVideoGen,
	}
	out := make(map[string]aimodel.LaneProfileDTO, len(lanes))
	store := NewSimLLMLaneStore()
	ctx := context.Background()
	for _, lane := range lanes {
		p, err := store.Load(ctx, lane)
		if err != nil {
			continue
		}
		out[string(lane)] = aimodel.LaneProfileDTO{
			Provider:    string(p.Provider),
			Model:       p.Model,
			MaxInFlight: p.MaxInFlight,
			MaxWaiters:  p.MaxWaiters,
		}
	}
	return out
}

// InitAIModel 注册 ProfileStore。
func InitAIModel() {
	aimodel.SetProfileStore(NewSimLLMLaneStore())
}
