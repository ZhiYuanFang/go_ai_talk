package simuser

import (
	"context"
	"sync"

	"hello/internal/services/aimodel"
)

// SimLLMLaneStore sim-user-service 侧 aimodel ProfileStore。
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
	p := aimodel.DefaultSeedProfile(lane)
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

// InitAIModel 注册 ProfileStore。
func InitAIModel() {
	aimodel.SetProfileStore(NewSimLLMLaneStore())
}
